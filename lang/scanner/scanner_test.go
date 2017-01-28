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
		name  string
		input string
		want  []literal
	}{
		{
			name:  "whitespace",
			input: "   \t  \n    \t\n \r \n    ",
			want: []literal{
				{token.Whitespace, "   \t  \n    \t\n \r \n    "},
			},
		},
		{
			name:  "simple inline string",
			input: "hello",
			want: []literal{
				{token.InlineString, "hello"},
			},
		},
		{
			name:  "simple string",
			input: `"hello"`,
			want: []literal{
				{token.String, `"hello"`},
			},
		},
		{
			name:  "string with escaped chars",
			input: `"he\"llo"`,
			want: []literal{
				{token.String, `"he\"llo"`},
			},
		},
		{
			name:  "string with escaped escape char",
			input: `"\\"`,
			want: []literal{
				{token.String, `"\\"`},
			},
		},
		{
			name:  "empty string",
			input: `""`,
			want: []literal{
				{token.String, `""`},
			},
		},
		{
			name:  "multi digit integer",
			input: "42",
			want: []literal{
				{token.Integer, "42"},
			},
		},
		{
			name:  "single digit integer",
			input: "1",
			want: []literal{
				{token.Integer, "1"},
			},
		},
		{
			name:  "single digit negative integer",
			input: "-1",
			want: []literal{
				{token.Integer, "-1"},
			},
		},
		{
			name:  "zero integer",
			input: "0",
			want: []literal{
				{token.Integer, "0"},
			},
		},
		{
			name:  "multi digit integer part float",
			input: "42.0",
			want: []literal{
				{token.Float, "42.0"},
			},
		},
		{
			name:  "multi digit decimal part float",
			input: "1.42",
			want: []literal{
				{token.Float, "1.42"},
			},
		},
		{
			name:  "single digit integer and decimal part",
			input: "1.1",
			want: []literal{
				{token.Float, "1.1"},
			},
		},
		{
			name:  "negative single digit integer and decimal part",
			input: "-1.1",
			want: []literal{
				{token.Float, "-1.1"},
			},
		},
		{
			name:  "zero float",
			input: "0.0",
			want: []literal{
				{token.Float, "0.0"},
			},
		},
		{
			name:  "multi digit integer part scientific float",
			input: "10e6",
			want: []literal{
				{token.Float, "10e6"},
			},
		},
		{
			name:  "multi digit integer part and scientific float",
			input: "10e60",
			want: []literal{
				{token.Float, "10e60"},
			},
		},
		{
			name:  "negative multi digit integer part and scientific float",
			input: "-10e60",
			want: []literal{
				{token.Float, "-10e60"},
			},
		},
		{
			name:  "single digit integer, decimal and scientific part float",
			input: "1.1e6",
			want: []literal{
				{token.Float, "1.1e6"},
			},
		},
		{
			name:  "negative single digit integer, decimal and scientific part float",
			input: "-1.1e6",
			want: []literal{
				{token.Float, "-1.1e6"},
			},
		},
		{
			name:  "negative multi digit decimal, single digit integer and scientific part float",
			input: "-1.10e6",
			want: []literal{
				{token.Float, "-1.10e6"},
			},
		},
		{
			name:  "only an exponent part is an inline string",
			input: "e10",
			want:  []literal{{token.InlineString, "e10"}},
		},
		{
			name:  "comma",
			input: ",",
			want: []literal{
				{token.Comma, ","},
			},
		},
		{
			name:  "pipe",
			input: "|",
			want: []literal{
				{token.Pipe, "|"},
			},
		},
		{
			name:  "dot",
			input: ".",
			want: []literal{
				{token.Dot, "."},
			},
		},
		{
			name:  "left bracket",
			input: "[",
			want: []literal{
				{token.LeftBracket, "["},
			},
		},
		{
			name:  "right bracket",
			input: "]",
			want: []literal{
				{token.RightBracket, "]"},
			},
		},
		{
			name:  "left parens",
			input: "(",
			want: []literal{
				{token.LeftParens, "("},
			},
		},
		{
			name:  "right parens",
			input: ")",
			want: []literal{
				{token.RightParens, ")"},
			},
		},
		{
			name:  "colon",
			input: `:`,
			want: []literal{
				{token.Colon, `:`},
			},
		},
		{
			name:  "dot and inline string",
			input: ".hello",
			want: []literal{
				{token.Dot, "."},
				{token.InlineString, "hello"},
			},
		},
		{
			name:  "dot, array brackets, dot and inline string",
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
			name:  "nested selectors with whitespace and pipes",
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
			name:  "dot, inline string follow by whitespace",
			input: ".hello    ",
			want: []literal{
				{token.Dot, "."},
				{token.InlineString, "hello"},
				{token.Whitespace, "    "},
			},
		},
		{
			name:  "dot, inline string that contains escaped whitespace",
			input: `.hello\ \ \ \ `,
			want: []literal{
				{token.Dot, "."},
				{token.InlineString, `hello\ \ \ \ `},
			},
		},
		{
			name:  "dot and brackets",
			input: ".[]",
			want: []literal{
				{token.Dot, "."},
				{token.LeftBracket, "["},
				{token.RightBracket, "]"},
			},
		},
		{
			name:  "dot, inline string and brackets with integer indices",
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
			name: "inline string of escaped characters",
			input: `\\\.\ \|\:\,\[\]\	\` + "\n",
			want: []literal{
				{token.InlineString, `\\\.\ \|\:\,\[\]\	\` + "\n"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

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
		})
	}
}

func TestScanIllegals(t *testing.T) {

	type literal struct {
		tok token.Token
		lit string
	}

	illegal := func(lit string) []literal {
		return []literal{{token.Illegal, lit}}
	}
	inline := func(lit string) []literal {
		return []literal{{token.InlineString, lit}}
	}

	tests := []struct {
		name  string
		input string
		want  []literal
	}{
		{"unclosed string", `"`, illegal(`"`)},

		{"single digit integer with invalid octal exponent part", `1e01`, illegal(`1e01`)},
		{"multi digit integer with invalid octal exponent part", `10e01`, illegal(`10e01`)},
		{"single digit float with invalid octal exponent part", `1.1e01`, illegal(`1.1e01`)},
		{"multi digit float with invalid octal exponent part", `1.10e01`, illegal(`1.10e01`)},
		{"single digit integer with invalid exponent part", `1eAAA`, append(illegal(`1eA`), inline("AA")...)},
		{"multi digit integer with invalid exponent part", `10eAAA`, append(illegal(`10eA`), inline("AA")...)},
		{"single digit float with invalid exponent part", `1.1eAAA`, append(illegal(`1.1eA`), inline("AA")...)},
		{"multi digit float with invalid exponent part", `1.10eAAA`, append(illegal(`1.10eA`), inline("AA")...)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

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
		})
	}
}
