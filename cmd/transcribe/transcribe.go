package transcribe

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/warmans/audio-search-bot/internal/assemblyai"
	"log/slog"
	"os"
)

func NewRootCommand(logger *slog.Logger) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "transcribe",
		Short: "",
	}

	cmd.AddCommand(NewMP3Command(logger))

	return cmd
}

// NewMP3Command
// for this you must extract the mp3 from the video file e.g.
// ffmpeg -i foo.mp4 foo.mp3
func NewMP3Command(logger *slog.Logger) *cobra.Command {
	var (
		mp3Path       string
		outputSRTPath string
	)
	cmd := &cobra.Command{
		Use:   "mp3",
		Short: "extract a correctly formatted episode name from stdin",
		RunE: func(cmd *cobra.Command, args []string) error {

			assemblyAiKey := os.Getenv("ASSEMBLY_AI_ACCESS_TOKEN")
			if assemblyAiKey == "" {
				return fmt.Errorf("ASSEMBLY_AI_ACCESS_TOKEN not set")
			}

			client := assemblyai.NewClient(logger, assemblyAiKey)
			return client.Transcribe(context.Background(), mp3Path, outputSRTPath)
		},
	}

	cmd.Flags().StringVar(&mp3Path, "i", "", "path to input MP3")
	cmd.Flags().StringVar(&outputSRTPath, "o", fmt.Sprintf("%s.srt", mp3Path), "path to dump SRT")

	return cmd
}
