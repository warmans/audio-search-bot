package searchterms

import (
	"reflect"
	"testing"
	"time"
)

func TestMustParse(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []Term
	}{
		{
			name: "parse word",
			args: args{s: "foo"},
			want: []Term{{Field: "content", Value: String("foo"), Op: CompOpFuzzyLike}},
		},
		{
			name: "parse words",
			args: args{s: "foo bar baz"},
			want: []Term{{Field: "content", Value: String("foo bar baz"), Op: CompOpFuzzyLike}},
		},
		{
			name: "parse quoted string",
			args: args{s: `"foo bar"`},
			want: []Term{{Field: "content", Value: String("foo bar"), Op: CompOpEq}},
		},
		{
			name: "parse quoted strings",
			args: args{s: `"foo bar" "baz"`},
			want: []Term{
				{Field: "content", Value: String("foo bar"), Op: CompOpEq},
				{Field: "content", Value: String("baz"), Op: CompOpEq},
			},
		},
		{
			name: "parse publication",
			args: args{s: `~xfm`},
			want: []Term{
				{Field: "publication", Value: String("xfm"), Op: CompOpEq},
			},
		},
		{
			name: "parse mention",
			args: args{s: `@steve`},
			want: []Term{
				{Field: "actor", Value: String("steve"), Op: CompOpEq},
			},
		},
		{
			name: "parse id",
			args: args{s: `#s01e05`},
			want: []Term{
				{Field: "series", Value: Int(1), Op: CompOpEq},
				{Field: "episode", Value: Int(5), Op: CompOpEq},
			},
		},
		{
			name: "parse id",
			args: args{s: `#E05`},
			want: []Term{
				{Field: "episode", Value: Int(5), Op: CompOpEq},
			},
		},
		{
			name: "parse id",
			args: args{s: `#S2`},
			want: []Term{
				{Field: "series", Value: Int(2), Op: CompOpEq},
			},
		},
		{
			name: "parse timestamp",
			args: args{s: `+10m30s`},
			want: []Term{
				{Field: "start_timestamp", Value: Duration(time.Minute*10 + time.Second*30), Op: CompOpGe},
			},
		},
		{
			name: "parse offset",
			args: args{s: `>20`},
			want: []Term{
				{Field: "offset", Value: Int(20), Op: CompOpEq},
			},
		},
		{
			name: "parse all",
			args: args{s: `@steve ~xfm #s1 +30m "man alive" karl >10`},
			want: []Term{
				{Field: "actor", Value: String("steve"), Op: CompOpEq},
				{Field: "publication", Value: String("xfm"), Op: CompOpEq},
				{Field: "series", Value: Int(1), Op: CompOpEq},
				{Field: "start_timestamp", Value: Duration(time.Minute * 30), Op: CompOpGe},
				{Field: "content", Value: String("man alive"), Op: CompOpEq},
				{Field: "content", Value: String("karl"), Op: CompOpFuzzyLike},
				{Field: "offset", Value: Int(10), Op: CompOpEq},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustParse(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MustParse() = %v, want %v", got, tt.want)
			}
		})
	}
}
