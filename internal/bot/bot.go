package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
	"github.com/warmans/audio-search-bot/internal/model"
	"github.com/warmans/audio-search-bot/internal/search"
	"github.com/warmans/audio-search-bot/internal/searchterms"
	"github.com/warmans/audio-search-bot/internal/store"
	"github.com/warmans/audio-search-bot/internal/util"
	"io"
	"log"
	"log/slog"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

var punctuation = regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
var spaces = regexp.MustCompile(`[\s]{2,}`)
var metaWhitespace = regexp.MustCompile(`[\n\r\t]+`)

func NewBot(
	logger *slog.Logger,
	session *discordgo.Session,
	guildID string,
	searcher search.Searcher,
	srtStore *store.SRTStore,
	mediaPath string,
) *Bot {
	bot := &Bot{
		logger:    logger,
		session:   session,
		guildID:   guildID,
		searcher:  searcher,
		mediaPath: mediaPath,
		srtStore:  srtStore,
		commands: []*discordgo.ApplicationCommand{
			{
				Name:        "supertalk",
				Description: "Search with confirmation",
				Type:        discordgo.ChatApplicationCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:         "query",
						Description:  "enter a partial quote",
						Type:         discordgo.ApplicationCommandOptionString,
						Required:     true,
						Autocomplete: true,
					},
				},
			},
		},
	}
	bot.commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"supertalk": bot.queryBegin,
	}
	bot.buttonHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, suffix string){
		"cfm": bot.queryComplete,
		"up":  bot.updatePreview,
	}
	bot.modalHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, suffix string){}

	return bot
}

type Bot struct {
	logger          *slog.Logger
	session         *discordgo.Session
	searcher        search.Searcher
	mediaPath       string
	guildID         string
	srtStore        *store.SRTStore
	commands        []*discordgo.ApplicationCommand
	commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
	buttonHandlers  map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, customIdPayload string)
	modalHandlers   map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, customIdPayload string)
	createdCommands []*discordgo.ApplicationCommand
}

func (b *Bot) Start() error {
	b.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	b.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			// exact match
			if h, ok := b.commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionApplicationCommandAutocomplete:
			// exact match
			if h, ok := b.commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionModalSubmit:
			// prefix match buttons to allow additional data in the customID
			for k, h := range b.modalHandlers {
				actionPrefix := fmt.Sprintf("%s:", k)
				if strings.HasPrefix(i.ModalSubmitData().CustomID, actionPrefix) {
					h(s, i, strings.TrimPrefix(i.ModalSubmitData().CustomID, actionPrefix))
					return
				}
			}
			b.respondError(s, i, fmt.Errorf("unknown customID format: %s", i.MessageComponentData().CustomID))
			return
		case discordgo.InteractionMessageComponent:
			// prefix match buttons to allow additional data in the customID
			for k, h := range b.buttonHandlers {
				actionPrefix := fmt.Sprintf("%s:", k)
				if strings.HasPrefix(i.MessageComponentData().CustomID, actionPrefix) {
					h(s, i, strings.TrimPrefix(i.MessageComponentData().CustomID, actionPrefix))
					return
				}
			}
			b.respondError(s, i, fmt.Errorf("unknown customID format: %s", i.MessageComponentData().CustomID))
			return
		}
	})
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("failed to open session: %w", err)
	}
	var err error
	b.createdCommands, err = b.session.ApplicationCommandBulkOverwrite(b.session.State.User.ID, b.guildID, b.commands)
	if err != nil {
		return fmt.Errorf("cannot register commands: %w", err)
	}
	return nil
}

func (b *Bot) Close() error {
	// cleanup commands
	for _, cmd := range b.createdCommands {
		err := b.session.ApplicationCommandDelete(b.session.State.User.ID, b.guildID, cmd.ID)
		if err != nil {
			return fmt.Errorf("cannot delete %s command: %w", cmd.Name, err)
		}
	}
	return b.session.Close()
}

func (b *Bot) queryBegin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		selection := i.ApplicationCommandData().Options[0].StringValue()
		if selection == "" {
			return
		}

		customID, err := decodeCustomIDPayload(selection)
		if err != nil {
			b.respondError(s, i, err)
			return
		}
		if err := b.sendPreview(s, i, customID, false); err != nil {
			b.respondError(s, i, err)
			return
		}
		return
	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()

		rawTerms := strings.TrimSpace(data.Options[0].StringValue())

		terms, err := searchterms.Parse(rawTerms)
		if err != nil {
			return
		}
		if len(terms) == 0 {
			b.logger.Warn("No terms were given")
			if err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionApplicationCommandAutocompleteResult,
				Data: &discordgo.InteractionResponseData{
					Choices: []*discordgo.ApplicationCommandOptionChoice{},
				},
			}); err != nil {
				b.logger.Error("Failed to respond with autocomplete options", slog.String("err", err.Error()))
			}
			return
		}

		res, err := b.searcher.Search(context.Background(), terms)
		if err != nil {
			b.logger.Error("Failed to fetch autocomplete options", slog.String("err", err.Error()))
			return
		}
		var choices []*discordgo.ApplicationCommandOptionChoice
		for _, v := range res {
			payload, err := json.Marshal(CustomID{
				MediaID:   v.MediaID,
				StartLine: v.Pos,
				EndLine:   v.Pos,
			})
			if err != nil {
				b.logger.Error("failed to marshal result", slog.String("err", err.Error()))
				continue
			}
			name := fmt.Sprintf("[%s] %s", v.MediaID, v.Content)
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  util.TrimToN(name, 100),
				Value: string(payload),
			})
		}
		if err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionApplicationCommandAutocompleteResult,
			Data: &discordgo.InteractionResponseData{
				Choices: choices,
			},
		}); err != nil {
			b.logger.Error("Failed to respond with autocomplete options", slog.String("err", err.Error()))
		}
		return
	}
	b.respondError(s, i, fmt.Errorf("unknown command type"))
}

func (b *Bot) updatePreview(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	customIDPayload string,
) {
	customID, err := decodeCustomIDPayload(customIDPayload)
	if err != nil {
		b.respondError(s, i, err)
		return
	}

	if err := b.sendPreview(s, i, customID, true); err != nil {
		b.respondError(s, i, err)
		return
	}
}

func (b *Bot) sendPreview(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	customID CustomID,
	update bool,
) error {

	if !update {
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "⏳ Fetching...",
			},
		}); err != nil {
			return err
		}
	} else {
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Flags:       discordgo.MessageFlagsEphemeral,
				Content:     "⏳ Updating...",
				Files:       []*discordgo.File{},
				Components:  []discordgo.MessageComponent{},
				Attachments: util.ToPtr([]*discordgo.MessageAttachment{}),
			},
		}); err != nil {
			return err
		}
	}

	username := "unknown"
	if i.Member != nil {
		username = i.Member.DisplayName()
	}

	interactionResponse, err := b.audioFileResponse(customID, username)
	if err != nil {
		b.respondError(s, i, err)
		return err
	}

	interactionResponse.Components = b.buttons(customID)
	interactionResponse.Flags = discordgo.MessageFlagsEphemeral

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    util.ToPtrOrNil(interactionResponse.Content),
		Files:      interactionResponse.Files,
		Components: util.ToPtr(interactionResponse.Components),
	})

	return err
}

func (b *Bot) queryComplete(s *discordgo.Session, i *discordgo.InteractionCreate, customIDPayload string) {

	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	if customIDPayload == "" {
		b.respondError(s, i, fmt.Errorf("missing customID"))
		return
	}
	customID, err := decodeCustomIDPayload(customIDPayload)
	if err != nil {
		b.respondError(s, i, fmt.Errorf("failed to decode customID: %w", err))
		return
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{},
	}); err != nil {
		b.respondError(s, i, fmt.Errorf("failed to begin interaction: %w", err))
		return
	}

	username := "unknown"
	if i.Member != nil {
		username = i.Member.DisplayName()
	}

	// respond audio
	interactionResponse, err := b.audioFileResponse(customID, username)
	if err != nil {
		b.respondError(s, i, err)
		return
	}
	if _, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Content: interactionResponse.Content,
		Files:   interactionResponse.Files,
	}); err != nil {
		b.respondError(s, i, err)
		return
	}
}

func (b *Bot) buttons(customID CustomID) []discordgo.MessageComponent {

	audioButton := discordgo.Button{
		// Label is what the user will see on the button.
		Label: "Enable Audio",
		Emoji: &discordgo.ComponentEmoji{
			Name: "🔊",
		},
		// Style provides coloring of the button. There are not so many styles tho.
		Style: discordgo.SecondaryButton,
		// CustomID is a thing telling Discord which data to send when this button will be pressed.
		CustomID: encodeCustomIDForAction("up", customID.withOption(withModifier(ContentModifierNone))),
	}
	if customID.ContentModifier == ContentModifierNone {
		audioButton.Label = "Disable Audio"
		audioButton.Emoji = &discordgo.ComponentEmoji{
			Name: "🔇",
		}
		audioButton.CustomID = encodeCustomIDForAction("up", customID.withOption(withModifier(ContentModifierTextOnly)))
	}

	editRow1 := []discordgo.MessageComponent{}
	if customID.StartLine > 0 {
		editRow1 = append(editRow1, discordgo.Button{
			// Label is what the user will see on the button.
			Label: "Shift Dialog Backwards",
			Emoji: &discordgo.ComponentEmoji{
				Name: "⏪",
			},
			// Style provides coloring of the button. There are not so many styles tho.
			Style: discordgo.SecondaryButton,
			// CustomID is a thing telling Discord which data to send when this button will be pressed.
			CustomID: encodeCustomIDForAction(
				"up",
				customID.withOption(
					withStartLine(customID.StartLine-1),
					withEndLine(customID.EndLine-1),
				),
			),
		})
	}
	editRow1 = append(editRow1, discordgo.Button{
		// Label is what the user will see on the button.
		Label: "Shift Dialog Forward",
		Emoji: &discordgo.ComponentEmoji{
			Name: "⏩",
		},
		// Style provides coloring of the button. There are not so many styles tho.
		Style: discordgo.SecondaryButton,
		// CustomID is a thing telling Discord which data to send when this button will be pressed.
		CustomID: encodeCustomIDForAction(
			"up",
			customID.withOption(
				withStartLine(customID.StartLine+1),
				withEndLine(customID.EndLine+1),
			),
		),
	})
	if customID.EndLine-customID.StartLine < 25 {
		if customID.StartLine > 0 {
			editRow1 = append(editRow1, discordgo.Button{
				// Label is what the user will see on the button.
				Label: "Add Leading Dialog",
				Emoji: &discordgo.ComponentEmoji{
					Name: "➕",
				},
				// Style provides coloring of the button. There are not so many styles tho.
				Style: discordgo.SecondaryButton,
				// CustomID is a thing telling Discord which data to send when this button will be pressed.
				CustomID: encodeCustomIDForAction(
					"up",
					customID.withOption(
						withStartLine(customID.StartLine-1),
					),
				),
			})
		}
		editRow1 = append(editRow1, discordgo.Button{
			// Label is what the user will see on the button.
			Label: "Add Following Dialog",
			Emoji: &discordgo.ComponentEmoji{
				Name: "➕",
			},
			// Style provides coloring of the button. There are not so many styles tho.
			Style: discordgo.SecondaryButton,
			// CustomID is a thing telling Discord which data to send when this button will be pressed.
			CustomID: encodeCustomIDForAction(
				"up",
				customID.withOption(
					withEndLine(customID.EndLine+1),
				),
			),
		})
	}

	editRow2 := []discordgo.MessageComponent{}
	if customID.EndLine-customID.StartLine > 0 {
		editRow2 = append(editRow2, discordgo.Button{
			// Label is what the user will see on the button.
			Label: "Trim Leading Dialog",
			Emoji: &discordgo.ComponentEmoji{
				Name: "✂",
			},
			// Style provides coloring of the button. There are not so many styles tho.
			Style: discordgo.SecondaryButton,
			// CustomID is a thing telling Discord which data to send when this button will be pressed.
			CustomID: encodeCustomIDForAction(
				"up",
				customID.withOption(
					withStartLine(customID.StartLine+1),
				),
			),
		})
		editRow2 = append(editRow2, discordgo.Button{
			// Label is what the user will see on the button.
			Label: "Trim Trailing Dialog",
			Emoji: &discordgo.ComponentEmoji{
				Name: "✂",
			},
			// Style provides coloring of the button. There are not so many styles tho.
			Style: discordgo.SecondaryButton,
			// CustomID is a thing telling Discord which data to send when this button will be pressed.
			CustomID: encodeCustomIDForAction(
				"up",
				customID.withOption(
					withEndLine(customID.EndLine-1),
				),
			),
		})
	}

	buttons := []discordgo.MessageComponent{}
	if len(editRow1) > 0 {
		buttons = append(buttons, discordgo.ActionsRow{
			Components: editRow1,
		})
	}
	if len(editRow2) > 0 {
		buttons = append(buttons, discordgo.ActionsRow{
			Components: editRow2,
		})
	}
	buttons = append(buttons, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Post",
				Style:    discordgo.SuccessButton,
				CustomID: encodeCustomIDForAction("cfm", customID),
			},
			audioButton,
		},
	})

	return buttons
}

func (b *Bot) audioFileResponse(
	customID CustomID,
	username string,
) (*discordgo.InteractionResponseData, error) {

	dialog, err := b.srtStore.GetDialogRange(customID.MediaID, customID.StartLine, customID.EndLine)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch selected lines: %s", customID.String())
	}
	if len(dialog) == 0 {
		return nil, fmt.Errorf("no dialog was selected")
	}

	dialogFormatted := strings.Builder{}
	for _, d := range dialog {
		dialogFormatted.WriteString(fmt.Sprintf("\n> %s", d.Content))
	}

	var files []*discordgo.File

	if customID.ContentModifier != ContentModifierTextOnly {
		mp3, err := b.createMp3(dialog[0].MediaFileName, dialog[0].StartTimestamp, dialog[len(dialog)-1].EndTimestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to create mp3: %w", err)
		}
		files = append(files, &discordgo.File{
			Name:        createFileName(dialog, "mp3"),
			ContentType: "audio/mpeg",
			Reader:      mp3,
		})
	}
	var content string
	if customID.ContentModifier != ContentModifierAudioOnly {
		content = fmt.Sprintf(
			"%s\n\n %s",
			dialogFormatted.String(),
			fmt.Sprintf(
				"`%s` @ `%s - %s` | Posted by %s",
				customID.MediaID,
				dialog[0].StartTimestamp.String(),
				dialog[len(dialog)-1].EndTimestamp.String(),
				username,
			),
		)
	} else {
		content = fmt.Sprintf("Posted by %s", username)
	}
	return &discordgo.InteractionResponseData{
		Content:     content,
		Files:       files,
		Attachments: util.ToPtr([]*discordgo.MessageAttachment{}),
	}, nil
}

func (b *Bot) respondError(s *discordgo.Session, i *discordgo.InteractionCreate, err error, logCtx ...any) {
	b.logger.Error("Error response was sent: "+err.Error(), logCtx...)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Request failed with error: %s", err.Error()),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		b.logger.Error("failed to respond", slog.String("err", err.Error()))
		return
	}
}

func (b *Bot) createMp3(
	mediaFileName string,
	startTimestamp time.Duration,
	endTimestamp time.Duration,
) (io.Reader, error) {
	buff := &bytes.Buffer{}
	err := ffmpeg_go.
		Input(path.Join(b.mediaPath, mediaFileName),
			ffmpeg_go.KwArgs{
				"ss": fmt.Sprintf("%0.2f", startTimestamp.Seconds()),
				"to": fmt.Sprintf("%0.2f", endTimestamp.Seconds()),
			}).
		Output("pipe:",
			ffmpeg_go.KwArgs{
				"format": "mp3",
			},
		).WithOutput(buff, os.Stderr).Run()
	if err != nil {
		b.logger.Error("ffmpeg failed", slog.String("err", err.Error()))
		return nil, err
	}

	return buff, nil
}

func createFileName(dialog []model.Dialog, suffix string) string {
	raw := []string{}
	for _, v := range dialog {
		raw = append(raw, strings.Split(v.Content, " ")...)
	}
	return fmt.Sprintf("%s.%s", contentToFilename(strings.Join(raw, " ")), suffix)
}

func contentToFilename(rawContent string) string {
	rawContent = punctuation.ReplaceAllString(rawContent, "")
	rawContent = metaWhitespace.ReplaceAllString(rawContent, " ")
	rawContent = spaces.ReplaceAllString(rawContent, " ")
	rawContent = strings.ToLower(strings.TrimSpace(rawContent))

	split := strings.Split(rawContent, " ")
	if len(split) > 9 {
		split = split[:8]
	}
	return strings.Join(split, "-")
}
