package searchterms

import (
	"github.com/warmans/audio-search-bot/internal/util"
	"reflect"
	"testing"
)

func Test_extractOffset(t *testing.T) {
	tests := []struct {
		name  string
		terms []Term
		want  []Term
		want1 *int64
	}{
		{
			name:  "empty terms returns empty, nil",
			terms: make([]Term, 0),
			want:  make([]Term, 0),
			want1: nil,
		},
		{
			name: "no offset returns original terms",
			terms: []Term{
				{Field: "actor", Value: String("steve"), Op: CompOpEq},
				{Field: "publication", Value: String("xfm"), Op: CompOpEq},
				{Field: "series", Value: Int(1), Op: CompOpEq},
			},
			want: []Term{
				{Field: "actor", Value: String("steve"), Op: CompOpEq},
				{Field: "publication", Value: String("xfm"), Op: CompOpEq},
				{Field: "series", Value: Int(1), Op: CompOpEq},
			},
			want1: nil,
		}, {
			name: "no offset returns original terms",
			terms: []Term{
				{Field: "actor", Value: String("steve"), Op: CompOpEq},
				{Field: "publication", Value: String("xfm"), Op: CompOpEq},
				{Field: "series", Value: Int(1), Op: CompOpEq},
			},
			want: []Term{
				{Field: "actor", Value: String("steve"), Op: CompOpEq},
				{Field: "publication", Value: String("xfm"), Op: CompOpEq},
				{Field: "series", Value: Int(1), Op: CompOpEq},
			},
			want1: nil,
		}, {
			name: "offset is extracted from last position",
			terms: []Term{
				{Field: "actor", Value: String("steve"), Op: CompOpEq},
				{Field: "publication", Value: String("xfm"), Op: CompOpEq},
				{Field: "offset", Value: Int(10), Op: CompOpEq},
			},
			want: []Term{
				{Field: "actor", Value: String("steve"), Op: CompOpEq},
				{Field: "publication", Value: String("xfm"), Op: CompOpEq},
			},
			want1: util.ToPtr(int64(10)),
		}, {
			name: "offset is extracted from first position",
			terms: []Term{
				{Field: "offset", Value: Int(10), Op: CompOpEq},
				{Field: "actor", Value: String("steve"), Op: CompOpEq},
				{Field: "publication", Value: String("xfm"), Op: CompOpEq},
			},
			want: []Term{
				{Field: "actor", Value: String("steve"), Op: CompOpEq},
				{Field: "publication", Value: String("xfm"), Op: CompOpEq},
			},
			want1: util.ToPtr(int64(10)),
		}, {
			name: "offset is extracted from middle position",
			terms: []Term{
				{Field: "actor", Value: String("steve"), Op: CompOpEq},
				{Field: "offset", Value: Int(10), Op: CompOpEq},
				{Field: "publication", Value: String("xfm"), Op: CompOpEq},
			},
			want: []Term{
				{Field: "actor", Value: String("steve"), Op: CompOpEq},
				{Field: "publication", Value: String("xfm"), Op: CompOpEq},
			},
			want1: util.ToPtr(int64(10)),
		}, {
			name: "offset is only filter",
			terms: []Term{
				{Field: "offset", Value: Int(10), Op: CompOpEq},
			},
			want:  []Term{},
			want1: util.ToPtr(int64(10)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ExtractOffset(tt.terms)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractOffset() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("extractOffset() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
