package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var punctuation = regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
var spaces = regexp.MustCompile(`[\s]{2,}`)
var metaWhitespace = regexp.MustCompile(`[\n\r\t]+`)

func TrimToN(line string, maxLength int) string {
	if len(line) <= maxLength {
		return line
	}
	return line[:maxLength-4] + "..."
}

func ToPtr[T any](v T) *T {
	return &v
}

func ToPtrOrNil[T comparable](v T) *T {
	var empty T
	if v == empty {
		return nil
	}
	return &v
}

func FromPtr[T any](v *T) T {
	if v == nil {
		var empty T
		return empty
	}
	return *v
}

func ParseSeriesAndEpisodeFromFileName(filePatternRegex *regexp.Regexp, filename string) (int64, int64, error) {

	match := filePatternRegex.FindStringSubmatch(filename)

	result := make(map[string]string)
	for i, name := range filePatternRegex.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	var err error
	var seriesInt int64
	if seriesStr, ok := result["series"]; ok && seriesStr != "" {
		seriesInt, err = strconv.ParseInt(strings.TrimLeft(seriesStr, "0"), 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to parse matched series int %s: %w", seriesStr, err)
		}
	} else {
		return 0, 0, fmt.Errorf("file pattern did not match series in : %s", filename)
	}
	var episodeInt int64
	if episodeStr, ok := result["episode"]; ok && episodeStr != "" {
		episodeInt, err = strconv.ParseInt(strings.TrimLeft(episodeStr, "0"), 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to parse matched episode int %s: %w", episodeStr, err)
		}
	} else {
		return 0, 0, fmt.Errorf("file pattern did not match [episode]")
	}
	return seriesInt, episodeInt, nil
}

// ExtractSeriesAndEpisode e.g. S1E01
func ExtractSeriesAndEpisode(raw string) (int32, int32, error) {
	raw = strings.TrimPrefix(raw, "S")
	parts := strings.Split(raw, "E")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("name was in wrong format: %s", raw)
	}
	series, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("series %s was not parsable: %w", parts[0], err)
	}
	episode, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("episode %s was not parsable: %w", parts[1], err)
	}
	return int32(series), int32(episode), nil
}

func FormatSeriesAndEpisode(series int32, episode int32) string {
	return fmt.Sprintf("S%02dE%02d", series, episode)
}

func InStrings(s string, ss ...string) bool {
	for _, v := range ss {
		if s == v {
			return true
		}
	}
	return false
}

func ContentToFilename(rawContent string) string {
	rawContent = punctuation.ReplaceAllString(rawContent, "")
	rawContent = spaces.ReplaceAllString(rawContent, " ")
	rawContent = metaWhitespace.ReplaceAllString(rawContent, " ")
	rawContent = strings.ToLower(strings.TrimSpace(rawContent))
	split := strings.Split(rawContent, " ")
	if len(split) > 9 {
		split = split[:8]
	}
	return strings.Join(split, "-")
}
