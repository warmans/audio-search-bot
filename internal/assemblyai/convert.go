package assemblyai

import (
	"fmt"
	aai "github.com/AssemblyAI/assemblyai-go-sdk"
	"github.com/warmans/audio-search-bot/internal/util"
	"io"
	"strings"
	"time"
)

const minLineDuration = time.Second * 1
const maxLineDuration = time.Second * 5

// ToSrt not tested
func ToSrt(rawData aai.Transcript, outputWriter io.Writer) error {
	var currentLine []string
	var firstWordStartTimestamp time.Duration
	var subtitleIdx = 1

	for k, word := range rawData.Words {

		wordText := util.FromPtr(word.Text)

		if len(currentLine) == 0 {
			firstWordStartTimestamp = wordStart(word)
		}
		currentLine = append(currentLine, wordText)
		lineDuration := wordEnd(word) - firstWordStartTimestamp
		if (!isSentenceEnd(wordText) || lineDuration < minLineDuration) &&
			lineDuration < maxLineDuration &&
			k < len(rawData.Words)-1 {
			continue
		}

		if _, err := fmt.Fprintf(outputWriter, "%d\n", subtitleIdx); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(
			outputWriter,
			"%s --> %s\n",
			formatDurationAsSrtTimestamp(firstWordStartTimestamp),
			formatDurationAsSrtTimestamp(wordEnd(word)),
		); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(outputWriter, "%s\n", strings.Join(currentLine, " ")); err != nil {
			return err
		}
		if _, err := fmt.Fprint(outputWriter, "\n"); err != nil {
			return err
		}

		currentLine = []string{}
		subtitleIdx++
	}
	return nil
}

func formatDurationAsSrtTimestamp(dur time.Duration) string {
	return time.Unix(0, 0).UTC().Add(dur).Format("15:04:05,000")
}

func isSentenceEnd(word string) bool {
	for _, v := range []string{".", "?", "!"} {
		if strings.HasSuffix(word, v) {
			return true
		}
	}
	return false
}

func wordStart(word aai.TranscriptWord) time.Duration {
	return time.Duration(util.FromPtr(word.Start)) * time.Millisecond
}

func wordEnd(word aai.TranscriptWord) time.Duration {
	return time.Duration(util.FromPtr(word.End)) * time.Millisecond
}
