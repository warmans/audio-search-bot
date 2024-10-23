package main

import (
	"github.com/warmans/audio-search-bot/cmd"
	"log/slog"
	"os"
)

func main() {
	logger := slog.Default()
	if err := cmd.Execute(logger); err != nil {
		logger.Error("Command failed", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
