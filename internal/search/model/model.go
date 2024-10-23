package model

import (
	"github.com/blugelabs/bluge"
	"github.com/warmans/audio-search-bot/internal/search/mapping"
	"time"
)

type DialogDocument struct {
	ID             string `json:"id"`
	Pos            int32  `json:"pos"`
	MediaID        string `json:"media_id"`
	Publication    string `json:"publication"`
	Series         int32  `json:"series"`
	Episode        int32  `json:"episode"`
	StartTimestamp int64  `json:"start_timestamp"`
	EndTimestamp   int64  `json:"end_timestamp"`
	MediaFileName  string `json:"video_file_name"`
	Content        string `json:"content"`
}

func (d *DialogDocument) FieldMapping() map[string]mapping.FieldType {
	return map[string]mapping.FieldType{
		"_id":             mapping.FieldTypeKeyword,
		"pos":             mapping.FieldTypeNumber,
		"media_id":        mapping.FieldTypeKeyword,
		"publication":     mapping.FieldTypeKeyword,
		"series":          mapping.FieldTypeNumber,
		"episode":         mapping.FieldTypeNumber,
		"start_timestamp": mapping.FieldTypeNumber,
		"end_timestamp":   mapping.FieldTypeNumber,
		"media_file_name": mapping.FieldTypeText,
		"content":         mapping.FieldTypeText,
	}
}

func (d *DialogDocument) Duration() time.Duration {
	return (time.Millisecond * time.Duration(d.EndTimestamp)) - (time.Millisecond * time.Duration(d.StartTimestamp))
}

func (d *DialogDocument) GetNamedField(name string) any {
	switch name {
	case "_id":
		return d.ID
	case "pos":
		return d.Pos
	case "media_id":
		return d.MediaID
	case "publication":
		return d.Publication
	case "series":
		return d.Series
	case "episode":
		return d.Episode
	case "start_timestamp":
		return d.StartTimestamp
	case "end_timestamp":
		return d.EndTimestamp
	case "media_file_name":
		return d.MediaFileName
	case "content":
		return d.Content
	}
	return ""
}

func (d *DialogDocument) SetNamedField(name string, value any) {
	switch name {
	case "_id":
		d.ID = string(value.([]byte))
	case "pos":
		d.Pos = int32(bytesToFloatOrZero(value))
	case "media_id":
		d.MediaID = string(value.([]byte))
	case "publication":
		d.Publication = string(value.([]byte))
	case "series":
		d.Series = int32(bytesToFloatOrZero(value))
	case "episode":
		d.Episode = int32(bytesToFloatOrZero(value))
	case "start_timestamp":
		d.StartTimestamp = int64(bytesToFloatOrZero(value))
	case "end_timestamp":
		d.EndTimestamp = int64(bytesToFloatOrZero(value))
	case "media_file_name":
		d.MediaFileName = string(value.([]byte))
	case "content":
		d.Content = string(value.([]byte))
	}
}

func bytesToFloatOrZero(val any) float64 {
	bytes := val.([]byte)
	float, err := bluge.DecodeNumericFloat64(bytes)
	if err != nil {
		return 0
	}
	return float
}
