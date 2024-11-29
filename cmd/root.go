package cmd

import (
	"github.com/spf13/cobra"
	"github.com/warmans/audio-search-bot/cmd/bot"
	"github.com/warmans/audio-search-bot/cmd/meta"
	"github.com/warmans/audio-search-bot/cmd/transcribe"
	"log/slog"
)

var (
	rootCmd = &cobra.Command{
		Use:   "audio-search-bot",
		Short: "Discord bot for searching audio files",
	}
)

func init() {

}

// Execute executes the root command.
func Execute(logger *slog.Logger) error {
	rootCmd.AddCommand(bot.NewBotCommand(logger))
	rootCmd.AddCommand(transcribe.NewRootCommand(logger))
	rootCmd.AddCommand(meta.NewRootCommand(logger))

	return rootCmd.Execute()
}
