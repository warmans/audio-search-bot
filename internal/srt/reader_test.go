package srt

import (
	"github.com/stretchr/testify/require"
	"github.com/warmans/audio-search-bot/internal/limits"
	"github.com/warmans/audio-search-bot/internal/model"
	"strings"
	"testing"
	"time"
)

func TestRead(t *testing.T) {
	type args struct {
		source string
	}
	tests := []struct {
		name    string
		args    args
		want    []model.Dialog
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:    "empty reader returns empty result",
			args:    args{source: ""},
			want:    []model.Dialog{},
			wantErr: require.NoError,
		},
		{
			name: "single result",
			args: args{source: "1\n00:00:00,498 --> 00:00:02,827\nHere's what I love most\nabout food and diet."},
			want: []model.Dialog{
				{
					Pos:            1,
					StartTimestamp: time.Millisecond * 498,
					EndTimestamp:   time.Second*2 + time.Millisecond*827,
					Content:        "Here's what I love most\nabout food and diet.",
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "Slightly malformed millisecond specifications are permitted",
			args: args{source: "1\n00:00:00,4 --> 00:00:02,82\nHere's what I love most\nabout food and diet."},
			want: []model.Dialog{
				{
					Pos:            1,
					StartTimestamp: time.Millisecond * 4,
					EndTimestamp:   time.Second*2 + time.Millisecond*82,
					Content:        "Here's what I love most\nabout food and diet.",
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "single result above limit",
			args: args{source: "1\n00:00:01,000 --> 00:00:40,000\nHere's what I love most\nabout food and diet."},
			want: []model.Dialog{
				{
					Pos:            1,
					StartTimestamp: time.Second,
					EndTimestamp:   time.Second * 31,
					Content:        "Here's what I love most\nabout food and diet.",
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "multiple results result",
			args: args{source: "1\n00:00:00,498 --> 00:00:02,827\nHere's what I love most\nabout food and diet.\n\n2\n00:00:02,827 --> 00:00:06,383\nWe all eat several times a day,\nand we're totally in charge\n\n3\n00:00:06,383 --> 00:00:09,427\nof what goes on our plate\nand what stays off.\n\n"},
			want: []model.Dialog{
				{
					Pos:            1,
					StartTimestamp: time.Millisecond * 498,
					EndTimestamp:   time.Second*2 + time.Millisecond*827,
					Content:        "Here's what I love most\nabout food and diet.",
				},
				{
					Pos:            2,
					StartTimestamp: time.Second*2 + time.Millisecond*827,
					EndTimestamp:   time.Second*6 + time.Millisecond*383,
					Content:        "We all eat several times a day,\nand we're totally in charge",
				},
				{
					Pos:            3,
					StartTimestamp: time.Second*6 + time.Millisecond*383,
					EndTimestamp:   time.Second*9 + time.Millisecond*427,
					Content:        "of what goes on our plate\nand what stays off.",
				},
			},
			wantErr: require.NoError,
		},
		{
			name: "gaps are filled",
			args: args{source: "1\n00:00:00,498 --> 00:00:02,827\nHere's what I love most\nabout food and diet.\n\n2\n00:00:02,827 --> 00:00:04,383\nWe all eat several times a day,\nand we're totally in charge\n\n3\n00:00:06,383 --> 00:00:09,427\nof what goes on our plate\nand what stays off.\n\n"},
			want: []model.Dialog{
				{
					Pos:            1,
					StartTimestamp: time.Millisecond * 498,
					EndTimestamp:   time.Second*2 + time.Millisecond*827,
					Content:        "Here's what I love most\nabout food and diet.",
				},
				{
					Pos:            2,
					StartTimestamp: time.Second*2 + time.Millisecond*827,
					EndTimestamp:   time.Second*6 + time.Millisecond*383,
					Content:        "We all eat several times a day,\nand we're totally in charge",
				},
				{
					Pos:            3,
					StartTimestamp: time.Second*6 + time.Millisecond*383,
					EndTimestamp:   time.Second*9 + time.Millisecond*427,
					Content:        "of what goes on our plate\nand what stays off.",
				},
			},
			wantErr: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Read(strings.NewReader(tt.args.source), true, limits.MaxMediaDuration)
			if tt.wantErr != nil {
				tt.wantErr(t, err)
			}
			require.EqualValues(t, tt.want, got)
		})
	}
}
