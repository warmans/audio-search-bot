package assemblyai

import (
	"context"
	"encoding/json"
	"fmt"
	aai "github.com/AssemblyAI/assemblyai-go-sdk"
	"github.com/pkg/errors"
	"log/slog"
	"os"
)

func NewClient(logger *slog.Logger, apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		logger: logger,
	}
}

type Client struct {
	apiKey string
	logger *slog.Logger
}

func (c *Client) Transcribe(ctx context.Context, mp3Path string, outputPath string) error {

	client := aai.NewClient(c.apiKey)

	// transcript parameters where speaker_labels has been enabled
	params := &aai.TranscriptOptionalParams{
		SpeakerLabels: aai.Bool(true),
	}

	mp3, err := os.Open(mp3Path)
	if err != nil {
		return fmt.Errorf("failed to open mp3: %w", err)
	}
	defer mp3.Close()

	var cachePath = fmt.Sprintf("%s.json", mp3Path)
	var transcript *aai.Transcript
	if transcript, err = c.getCached(cachePath); err != nil {
		return err
	}

	if transcript == nil {
		c.logger.Info("No Cache, submitting job...", slog.String("i", mp3Path), slog.String("cache_path", cachePath))
		newTranscript, err := client.Transcripts.TranscribeFromReader(ctx, mp3, params)
		if err != nil {
			return fmt.Errorf("transcription failed: %w", err)
		}
		transcript = &newTranscript

		if err := c.dumpCache(mp3Path, transcript); err != nil {
			return err
		}
	}

	c.logger.Info("Converting result to SRT...", slog.String("o", outputPath))

	outputSRT, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputSRT.Close()

	return ToSrt(*transcript, outputSRT)
}
func (c *Client) getCached(cachePath string) (*aai.Transcript, error) {
	f, err := os.Open(cachePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var transcript *aai.Transcript
	if err := json.NewDecoder(f).Decode(&transcript); err != nil {
		return nil, err
	}

	return transcript, nil
}

func (c *Client) dumpCache(mp3Path string, transcript *aai.Transcript) error {
	f, err := os.Create(fmt.Sprintf("%s.json", mp3Path))
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(transcript)
}
