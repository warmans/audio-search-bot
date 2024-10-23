package model

import (
	"fmt"
	"github.com/warmans/audio-search-bot/internal/util"
	"time"
)

type AudioMeta struct {
	SourceSRTName    string    `json:"source_srt_name"`
	SourceSRTModTime time.Time `json:"source_srt_mod_time"`
	ImportedIndex    bool      `json:"imported_index"`
	ImportedDB       bool      `json:"imported_db"`
}

type Dialog struct {
	Pos            int32         `json:"pos" db:"pos"`
	StartTimestamp time.Duration `json:"start_timestamp" db:"start_timestamp"`
	EndTimestamp   time.Duration `json:"end_timestamp" db:"end_timestamp"`
	Content        string        `json:"content" db:"content"`
	MediaFileName  string        `json:"media_file_name" db:"media_file_name"`
}

func (e *Dialog) ID(episodeID string) string {
	return fmt.Sprintf("%s-%d", episodeID, e.Pos)
}

type Audio struct {
	SRTFile     string    `json:"srt_file"`
	SRTModTime  time.Time `json:"srt_mod_time"`
	MediaFile   string    `json:"media_file"`
	Publication string    `json:"publication"`
	Series      int32     `json:"season"`
	Episode     int32     `json:"episode"`
	Dialog      []Dialog  `json:"dialog"`
}

func (a *Audio) ID() string {
	return fmt.Sprintf("%s-%s", a.Publication, util.FormatSeriesAndEpisode(a.Series, a.Episode))
}

type Publication struct {
	Name   string   `json:"name"`
	Series []string `json:"series"`
}
