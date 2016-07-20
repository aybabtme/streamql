package scanner

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/aybabtme/streamql/lang/token"
)

func TestScan(t *testing.T) {

	type literal struct {
		tok token.Token
		lit string
	}

	tests := []struct {
		input string
		want  []literal
	}{
		{
			input: ".",
			want: []literal{
				{token.Dot, "."},
			},
		},
		{
			input: "42",
			want: []literal{
				{token.Integer, "42"},
			},
		},
		{
			input: "hello",
			want: []literal{
				{token.InlineString, "hello"},
			},
		},
		{
			input: ".hello",
			want: []literal{
				{token.Dot, "."},
				{token.InlineString, "hello"},
			},
		},
		{
			input: `.[].id`,
			want: []literal{
				{token.Dot, "."},
				{token.LeftBracket, "["},
				{token.RightBracket, "]"},
				{token.Dot, "."},
				{token.InlineString, "id"},
			},
		},
		{
			input: `.[].id | .[].id, .[].id`,
			want: []literal{
				{token.Dot, "."},
				{token.LeftBracket, "["},
				{token.RightBracket, "]"},
				{token.Dot, "."},
				{token.InlineString, "id"},
				{token.Whitespace, " "},
				{token.Pipe, "|"},
				{token.Whitespace, " "},
				{token.Dot, "."},
				{token.LeftBracket, "["},
				{token.RightBracket, "]"},
				{token.Dot, "."},
				{token.InlineString, "id"},
				{token.Comma, ","},
				{token.Whitespace, " "},
				{token.Dot, "."},
				{token.LeftBracket, "["},
				{token.RightBracket, "]"},
				{token.Dot, "."},
				{token.InlineString, "id"},
			},
		},
		{
			input: ".hello    ",
			want: []literal{
				{token.Dot, "."},
				{token.InlineString, "hello"},
				{token.Whitespace, "    "},
			},
		},
		{
			input: `.hello\ \ \ \ `,
			want: []literal{
				{token.Dot, "."},
				{token.InlineString, "hello    "},
			},
		},
		{
			input: ".[]",
			want: []literal{
				{token.Dot, "."},
				{token.LeftBracket, "["},
				{token.RightBracket, "]"},
			},
		},
		{
			input: ".hello[1:2]",
			want: []literal{
				{token.Dot, "."},
				{token.InlineString, "hello"},
				{token.LeftBracket, "["},
				{token.Integer, "1"},
				{token.Colon, ":"},
				{token.Integer, "2"},
				{token.RightBracket, "]"},
			},
		},
		{
			input: ".hello[1:2][42][].world",
			want: []literal{
				{token.Dot, "."},
				{token.InlineString, "hello"},
				{token.LeftBracket, "["},
				{token.Integer, "1"},
				{token.Colon, ":"},
				{token.Integer, "2"},
				{token.RightBracket, "]"},
				{token.LeftBracket, "["},
				{token.Integer, "42"},
				{token.RightBracket, "]"},
				{token.LeftBracket, "["},
				{token.RightBracket, "]"},
				{token.Dot, "."},
				{token.InlineString, "world"},
			},
		},
		{
			input: `\\\.\ \|\:\,\[\]\	\` + "\n",
			want: []literal{
				{token.InlineString, `\. |:,[]` + "\t\n"},
			},
		},
	}

	for n, tt := range tests {
		t.Logf("test #%d, input %q", n, tt.input)

		scan := NewScanner(strings.NewReader(tt.input))

		var got []literal
		for {
			tok, lit, err := scan.Scan()
			if err != nil && err != io.EOF {
				t.Fatal(err)
			}
			if err == io.EOF {
				break
			}
			got = append(got, literal{
				tok: tok, lit: lit,
			})
		}

		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("want=%#v", tt.want)
			t.Errorf(" got=%#v", got)
		}
	}
}
