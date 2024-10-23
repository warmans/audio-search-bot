package searchterms

import (
	"reflect"
	"testing"
)

func TestScan(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    []token
		wantErr bool
	}{
		{
			name: "scan word",
			args: args{
				str: "foo",
			},
			want:    []token{{tag: tagWord, lexeme: "foo"}, {tag: tagEOF}},
			wantErr: false,
		},
		{
			name: "scan number",
			args: args{
				str: "123",
			},
			want:    []token{{tag: tagInt, lexeme: "123"}, {tag: tagEOF}},
			wantErr: false,
		},
		{
			name: "scan words",
			args: args{
				str: "foo bar baz",
			},
			want: []token{
				{tag: tagWord, lexeme: "foo"},
				{tag: tagWord, lexeme: "bar"},
				{tag: tagWord, lexeme: "baz"},
				{tag: tagEOF},
			},
			wantErr: false,
		},
		{
			name: "scan quoted string",
			args: args{
				str: `"foo bar"`,
			},
			want:    []token{{tag: tagQuotedString, lexeme: "foo bar"}, {tag: tagEOF}},
			wantErr: false,
		},
		{
			name: "scan mention",
			args: args{
				str: `@steve`,
			},
			want:    []token{{tag: tagMention, lexeme: "@"}, {tag: tagWord, lexeme: "steve"}, {tag: tagEOF}},
			wantErr: false,
		},
		{
			name: "scan publication",
			args: args{
				str: `~xfm`,
			},
			want:    []token{{tag: tagPublication, lexeme: "~"}, {tag: tagWord, lexeme: "xfm"}, {tag: tagEOF}},
			wantErr: false,
		},
		{
			name: "scan timestamp",
			args: args{
				str: `+10m`,
			},
			want:    []token{{tag: tagTimestamp, lexeme: "+"}, {tag: tagInt, lexeme: "10"}, {tag: tagWord, lexeme: "m"}, {tag: tagEOF}},
			wantErr: false,
		},
		{
			name: "scan offset",
			args: args{
				str: `>10`,
			},
			want:    []token{{tag: tagOffset, lexeme: ">"}, {tag: tagInt, lexeme: "10"}, {tag: tagEOF}},
			wantErr: false,
		},
		{
			name: "scan everything",
			args: args{
				str: `"man alive" @steve ~xfm #s1 foo`,
			},
			want: []token{
				{tag: tagQuotedString, lexeme: "man alive"},
				{tag: tagMention, lexeme: "@"},
				{tag: tagWord, lexeme: "steve"},
				{tag: tagPublication, lexeme: "~"},
				{tag: tagWord, lexeme: "xfm"},
				{tag: tagId, lexeme: "#"},
				{tag: tagWord, lexeme: "s1"},
				{tag: tagWord, lexeme: "foo"},
				{tag: tagEOF}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Scan(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Scan() got = %v, want %v", got, tt.want)
			}
		})
	}
}
