package cmd

import (
	"github.com/spf13/cobra"
	"github.com/warmans/audio-search-bot/cmd/bot"
	"github.com/warmans/audio-search-bot/cmd/transcribe"
	"log/slog"
)

var (
	rootCmd = &cobra.Command{
		Use:   "tvgif",
		Short: "Discord bot for posting TV show gifs",
	}
)

func init() {

}

// Execute executes the root command.
func Execute(logger *slog.Logger) error {
	rootCmd.AddCommand(bot.NewBotCommand(logger))
	rootCmd.AddCommand(transcribe.NewRootCommand(logger))
	return rootCmd.Execute()
}
