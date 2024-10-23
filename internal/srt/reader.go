package srt

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/warmans/audio-search-bot/internal/limits"
	"github.com/warmans/audio-search-bot/internal/model"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var htmlTag = regexp.MustCompile(`<[^<>]+>`)

type srtEntity string

const (
	srtEntryPos    = srtEntity("pos")
	srtEntryTs     = srtEntity("ts")
	srtEntryDialog = srtEntity("dialog")
)

func Read(source io.Reader, eliminateSpeechGaps bool, limitDialogDuration time.Duration) ([]model.Dialog, error) {

	dialog := []model.Dialog{}
	currentDialog := model.Dialog{}
	wantNext := srtEntryPos

	scanner := bufio.NewScanner(source)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		// strip random BOMs
		line := strings.Replace(strings.TrimSpace(scanner.Text()), "\ufeff", "", -1)
		if line == "" {
			if currentDialog != (model.Dialog{}) {
				dialog = append(dialog, currentDialog)
				currentDialog = model.Dialog{}
			}
			wantNext = srtEntryPos
			continue
		}

		var err error
		switch wantNext {
		case srtEntryPos:
			if currentDialog.Pos, err = scanPos(line); err != nil {
				return nil, fmt.Errorf("failed to scan position: %w", err)
			}
			wantNext = srtEntryTs
		case srtEntryTs:
			startTimesamp, endTimestamp, err := scanTimestamps(line)
			if err != nil {
				return nil, fmt.Errorf("failed to scan timestamps: %w", err)
			}
			currentDialog.StartTimestamp = startTimesamp
			currentDialog.EndTimestamp = limitDuration(startTimesamp, endTimestamp, limitDialogDuration)
			wantNext = srtEntryDialog
		case srtEntryDialog:
			line = htmlTag.ReplaceAllString(line, "")
			// just keep adding content until a blank line is encountered
			if currentDialog.Content == "" {
				currentDialog.Content = line
			} else {
				currentDialog.Content += "\n" + line
			}
		}
	}
	if currentDialog != (model.Dialog{}) {
		dialog = append(dialog, currentDialog)
	}

	// override the end time of a line of dialog with the following line's start time
	if eliminateSpeechGaps {
		dialog = eliminateGaps(dialog)
	}

	return dialog, nil
}

// make the end timestamp of dialog equal to the start of the next line, unless it exceeds the max duration
func eliminateGaps(dialog []model.Dialog) []model.Dialog {
	fixed := make([]model.Dialog, len(dialog))
	for k, v := range dialog {
		if k == len(dialog)-1 {
			fixed[k] = v
			break
		}
		v.EndTimestamp = limitDuration(v.StartTimestamp, max(dialog[k+1].StartTimestamp, v.EndTimestamp), limits.MaxMediaDuration)
		fixed[k] = v
	}
	return fixed
}

func limitDuration(startTimestamp, endTimestamp, maxDuration time.Duration) time.Duration {
	if endTimestamp-startTimestamp > maxDuration {
		return startTimestamp + maxDuration
	}
	return endTimestamp
}

func scanPos(line string) (int32, error) {
	intVal, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return 0, fmt.Errorf("failed to parse position '%s': %w", line, err)
	}
	return int32(intVal), nil
}

func scanTimestamps(line string) (time.Duration, time.Duration, error) {
	times := strings.Split(line, "-->")
	if len(times) < 2 {
		return time.Duration(0), time.Duration(0), fmt.Errorf("invalid timestamp line: '%s'", line)
	}

	startTime, err := parseTime(times[0])
	if err != nil {
		return time.Duration(0), time.Duration(0), fmt.Errorf("invalid start timestamp '%s': %w", times[0], err)
	}
	endTime, err := parseTime(times[1])
	if err != nil {
		return time.Duration(0), time.Duration(0), fmt.Errorf("invalid end timestamp '%s': %w", times[1], err)
	}

	return startTime, endTime, nil
}

func parseTime(input string) (time.Duration, error) {
	regex := regexp.MustCompile(`(\d{2}):(\d{2}):(\d{2}),(\d+)`)
	matches := regex.FindStringSubmatch(input)

	if len(matches) < 4 {
		return time.Duration(0), fmt.Errorf("invalid time format:%s", input)
	}

	hour, err := strconv.Atoi(matches[1])
	if err != nil {
		return time.Duration(0), err
	}
	minute, err := strconv.Atoi(matches[2])
	if err != nil {
		return time.Duration(0), err
	}
	second, err := strconv.Atoi(matches[3])
	if err != nil {
		return time.Duration(0), err
	}
	millisecond, err := strconv.Atoi(matches[4])
	if err != nil {
		return time.Duration(0), err
	}

	return time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute + time.Duration(second)*time.Second + time.Duration(millisecond)*time.Millisecond, nil
}
