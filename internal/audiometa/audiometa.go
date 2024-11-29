package audiometa

import (
	"encoding/json"
	"errors"
	"fmt"
	ffmpeg_go "github.com/warmans/ffmpeg-go"
	"os"
	"path"
	"strings"
	"time"
)

type ProbeStream struct {
	Index          int    `json:"index"`
	CodecName      string `json:"codec_name"`
	CodecLongName  string `json:"codec_long_name"`
	CodecType      string `json:"codec_type"`
	CodecTagString string `json:"codec_tag_string"`
	CodecTag       string `json:"codec_tag"`
	SampleFmt      string `json:"sample_fmt"`
	SampleRate     string `json:"sample_rate"`
	Channels       int    `json:"channels"`
	ChannelLayout  string `json:"channel_layout"`
	BitsPerSample  int    `json:"bits_per_sample"`
	InitialPadding int    `json:"initial_padding"`
	RFrameRate     string `json:"r_frame_rate"`
	AvgFrameRate   string `json:"avg_frame_rate"`
	TimeBase       string `json:"time_base"`
	StartPts       int    `json:"start_pts"`
	StartTime      string `json:"start_time"`
	DurationTs     int64  `json:"duration_ts"`
	Duration       string `json:"duration"`
	BitRate        string `json:"bit_rate"`
}

type ProbeTags struct {
	Track               string `json:"track"`
	ReplaygainTrackGain string `json:"replaygain_track_gain"`
	ReplaygainTrackPeak string `json:"replaygain_track_peak"`
	Title               string `json:"title"`
	Album               string `json:"album"`
	AlbumArtist         string `json:"album_artist"`
	Artist              string `json:"artist"`
	Genre               string `json:"genre"`
	Date                string `json:"date"`
}

type ProbeFormat struct {
	Filename       string    `json:"filename"`
	NbStreams      int       `json:"nb_streams"`
	NbPrograms     int       `json:"nb_programs"`
	FormatName     string    `json:"format_name"`
	FormatLongName string    `json:"format_long_name"`
	StartTime      string    `json:"start_time"`
	Duration       string    `json:"duration"`
	Size           string    `json:"size"`
	BitRate        string    `json:"bit_rate"`
	ProbeScore     int       `json:"probe_score"`
	Tags           ProbeTags `json:"tags"`
}

type ProbeResult struct {
	Streams []ProbeStream `json:"streams"`
	Format  ProbeFormat   `json:"format"`
}

func DumpMeta(audioFilePath string) error {
	result, err := ExtractMeta(audioFilePath)
	if err != nil {
		return err
	}
	for _, v := range result.Streams {
		if v.CodecName == "png" {
			imageFileName := fmt.Sprintf("%s.png", strings.TrimSuffix(audioFilePath, path.Ext(audioFilePath)))
			err := DumpImage(audioFilePath, imageFileName)
			if err != nil {
				return err
			}
			break
		}
	}

	outputFileName := fmt.Sprintf("%s.meta.json", strings.TrimSuffix(audioFilePath, path.Ext(audioFilePath)))
	f, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(result.Format.Tags)
}

func DumpImage(audioFilePath string, outputImagePath string) error {
	if err := os.Remove(outputImagePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to remove old file")
		}
	}
	return ffmpeg_go.
		Input(audioFilePath, ffmpeg_go.KwArgs{}).
		Output(outputImagePath).
		Run()
}

func ExtractMeta(audioFilePath string) (*ProbeResult, error) {

	meta, err := ffmpeg_go.ProbeWithTimeout(audioFilePath, time.Second, nil)
	if err != nil {
		return nil, err
	}

	result := &ProbeResult{}
	json.NewDecoder(strings.NewReader(meta)).Decode(result)

	return result, nil
}
