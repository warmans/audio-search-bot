package util

import (
	"regexp"
	"testing"
)

func TestParseSeriesAndEpisodeFromFileName(t *testing.T) {
	type args struct {
		filePatternRegex *regexp.Regexp
		filename         string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		want1   int64
		wantErr bool
	}{
		{
			name: "series number above 10",
			args: args{
				filePatternRegex: ShortSeasonAndEpisode,
				filename:         "The Simpsons - 10x01 - Lard of the Dance.mkv",
			},
			want:    10,
			want1:   1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ParseSeriesAndEpisodeFromFileName(tt.args.filePatternRegex, tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSeriesAndEpisodeFromFileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseSeriesAndEpisodeFromFileName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseSeriesAndEpisodeFromFileName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
