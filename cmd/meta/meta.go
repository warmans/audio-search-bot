package meta

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/warmans/audio-search-bot/internal/audiometa"
	"log/slog"
)

func NewRootCommand(logger *slog.Logger) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "meta",
		Short: "",
	}

	cmd.AddCommand(NewDumpCommand())

	return cmd
}

func NewDumpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "dump",
		Short: "extract a correctly formatted episode name from stdin",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("first argument must be the media file path")
			}
			return audiometa.DumpMeta(args[0])
		},
	}
}
