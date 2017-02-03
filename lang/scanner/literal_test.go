package scanner

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aybabtme/streamql/lang/token"
)

func TestParseInlineString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "inline string",
			input: `hello`,
			want:  "hello",
		},
		{
			name:  "inline string that contains escaped whitespace",
			input: `hello\ \ \ \ `,
			want:  "hello    ",
		},
		{
			name: "inline string of escaped characters",
			input: `\\\.\ \|\:\,\[\]\	\` + "\n",
			want: `\. |:,[]` + "\t\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			sc := NewScanner(strings.NewReader(tt.input))
			tok, lit, err := sc.Scan()
			if err != nil != tt.wantErr {
				t.Fatalf("scanning: %v", err)
			}
			if tok != token.InlineString {
				t.Fatalf("got a %v", tok)
			}
			if lit != tt.input {
				t.Errorf("want literal=%q", tt.input)
				t.Errorf(" got literal=%q", lit)
			}

			got, err := ParseInlineString(lit)
			if err != nil != tt.wantErr {
				t.Fatalf("ParseInlineString: %v", err)
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("want=%q", tt.want)
				t.Errorf(" got=%q", got)
			}
		})
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "string",
			input: `"hello"`,
			want:  "hello",
		},
		{
			name:  "string that contains whitespace",
			input: `"hello    "`,
			want:  "hello    ",
		},
		{
			name:  "string of escaped characters",
			input: `"\\\"\n\t\r"`,
			want:  `\"` + "\n\t\r",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			sc := NewScanner(strings.NewReader(tt.input))
			tok, lit, err := sc.Scan()
			if err != nil != tt.wantErr {
				t.Fatalf("scanning: %v", err)
			}
			if tok != token.String {
				t.Fatalf("got a %v (lit=%q)", tok, lit)
			}
			if lit != tt.input {
				t.Errorf("want literal=%q", tt.input)
				t.Errorf(" got literal=%q", lit)
			}

			got, err := ParseString(lit)
			if err != nil != tt.wantErr {
				t.Fatalf("ParseString: %v", err)
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("want=%q", tt.want)
				t.Errorf(" got=%q", got)
			}
		})
	}
}

func TestParseNumber(t *testing.T) {

	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		// regular integers
		{"", `0`, 0, false},
		{"", `-0`, -0, false},
		{"", `1`, 1, false},
		{"", `-1`, -1, false},
		{"", `12`, 12, false},
		{"", `-12`, -12, false},
		{"", `1234567890`, 1234567890, false},
		{"", `-1234567890`, -1234567890, false},

		// decimal values
		{"", `0.0`, 0.0, false},
		{"", `-0.0`, -0.0, false},
		{"", `0.1`, 0.1, false},
		{"", `-0.1`, -0.1, false},
		{"", `0.12`, 0.12, false},
		{"", `-0.12`, -0.12, false},
		{"", `0.1234567890`, 0.1234567890, false},
		{"", `-0.1234567890`, -0.1234567890, false},
		{"", `0.0001`, 0.0001, false},
		{"", `1.2`, 1.2, false},
		{"", `-1.2`, -1.2, false},
		{"", `12.34`, 12.34, false},
		{"", `-12.34`, -12.34, false},

		// exponential values
		{"", `0e0`, 0e0, false},
		{"", `-0e0`, -0, false},
		{"", `0e1`, 0, false},
		{"", `-0e1`, -0, false},
		{"", `0e3`, 0, false},
		{"", `-0e3`, -0e3, false},

		{"", `1e2`, 1e2, false},
		{"", `-1e2`, -1e2, false},
		{"", `1e-2`, 1e-2, false},
		{"", `-1e-2`, -1e-2, false},
		{"", `12e34`, 12e34, false},
		{"", `-12e34`, -12e34, false},

		{"", `0.1e0`, 0.1e0, false},
		{"", `-0.1e0`, -0.1e0, false},
		{"", `0.1e1`, 0.1e1, false},
		{"", `-0.1e1`, -0.1e1, false},
		{"", `0.1e3`, 0.1e3, false},
		{"", `-0.1e3`, -0.1e3, false},

		{"", `1.1e2`, 1.1e2, false},
		{"", `-1.1e2`, -1.1e2, false},
		{"", `1.1e-2`, 1.1e-2, false},
		{"", `-1.1e-2`, -1.1e-2, false},
		{"", `12.1e34`, 12.1e34, false},
		{"", `-12.1e34`, -12.1e34, false},

		{"", `0.0e0`, 0.0e0, false},
		{"", `-0.0e0`, -0.0e0, false},
		{"", `0.0e1`, 0.0e1, false},
		{"", `-0.0e1`, -0.0e1, false},
		{"", `0.0e3`, 0.0e3, false},
		{"", `-0.0e3`, -0.0e3, false},

		{"", `1.0e2`, 1.0e2, false},
		{"", `-1.0e2`, -1.0e2, false},
		{"", `1.0e-2`, 1.0e-2, false},
		{"", `-1.0e-2`, -1.0e-2, false},
		{"", `12.0e34`, 12.0e34, false},
		{"", `-12.0e34`, -12.0e34, false},

		{"", `0.1234e0`, 0.1234e0, false},
		{"", `-0.1234e0`, -0.1234e0, false},
		{"", `0.1234e1`, 0.1234e1, false},
		{"", `-0.1234e1`, -0.1234e1, false},
		{"", `0.1234e3`, 0.1234e3, false},
		{"", `-0.1234e3`, -0.1234e3, false},

		{"", `1.1234e2`, 1.1234e2, false},
		{"", `-1.1234e2`, -1.1234e2, false},
		{"", `1.1234e-2`, 1.1234e-2, false},
		{"", `-1.1234e-2`, -1.1234e-2, false},
		{"", `12.1234e34`, 12.1234e34, false},
		{"", `-12.1234e34`, -12.1234e34, false},

		{"", `0.01234e0`, 0.01234e0, false},
		{"", `-0.01234e0`, -0.01234e0, false},
		{"", `0.01234e1`, 0.01234e1, false},
		{"", `-0.01234e1`, -0.01234e1, false},
		{"", `0.01234e3`, 0.01234e3, false},
		{"", `-0.01234e3`, -0.01234e3, false},

		{"", `1.01234e2`, 1.01234e2, false},
		{"", `-1.01234e2`, -1.01234e2, false},
		{"", `1.01234e-2`, 1.01234e-2, false},
		{"", `-1.01234e-2`, -1.01234e-2, false},
		{"", `12.01234e34`, 12.01234e34, false},
		{"", `-12.01234e34`, -12.01234e34, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Logf("input=%q", tt.input)
			sc := NewScanner(strings.NewReader(tt.input))
			tok, lit, err := sc.Scan()
			if err != nil != tt.wantErr {
				t.Fatalf("scanning: %v", err)
			}
			if tok != token.Float && tok != token.Integer {
				t.Fatalf("got a %v", tok)
			}
			if lit != tt.input {
				t.Errorf("want literal=%q", tt.input)
				t.Errorf(" got literal=%q", lit)
			}

			got, err := ParseNumber(lit)
			if err != nil != tt.wantErr {
				t.Fatalf("ParseNumber: %v", err)
			}

			acceptRoundError := 0.000000000000001
			if (tt.want > got && acceptRoundError < (tt.want-got)/tt.want) ||
				(tt.want < got && acceptRoundError < (got-tt.want)/got) {
				t.Errorf("want=%#v", tt.want)
				t.Errorf(" got=%#v", got)
			}
		})
	}
}

func TestParseInteger(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{"zero", `0`, 0, false},
		{"negative zero", `-0`, -0, false},
		{"single digit", `1`, 1, false},
		{"negative single digit", `-1`, -1, false},
		{"multiple digit", `12`, 12, false},
		{"negative multiple digit", `-12`, -12, false},
		{"all digits", `1234567890`, 1234567890, false},
		{"negative all digits", `-1234567890`, -1234567890, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			sc := NewScanner(strings.NewReader(tt.input))
			tok, lit, err := sc.Scan()
			if err != nil != tt.wantErr {
				t.Fatalf("scanning: %v", err)
			}
			if tok != token.Integer {
				t.Fatalf("got a %v", tok)
			}
			if lit != tt.input {
				t.Errorf("want literal=%q", tt.input)
				t.Errorf(" got literal=%q", lit)
			}

			got, err := ParseInteger(lit)
			if err != nil != tt.wantErr {
				t.Fatalf("ParseInteger: %v", err)
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("want=%#v", tt.want)
				t.Errorf(" got=%#v", got)
			}
		})
	}
}

func TestParseBoolean(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    bool
		wantTok token.Token
		wantErr bool
	}{
		{"true", "true", true, token.InlineString, false},
		{"false", "false", false, token.InlineString, false},
		{"other things", "1", false, token.Integer, true},
		{"other things", "0", false, token.Integer, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			sc := NewScanner(strings.NewReader(tt.input))
			tok, lit, err := sc.Scan()
			if tok != tt.wantTok {
				t.Fatalf("want a %v", tt.wantTok)
				t.Fatalf(" got a %v", tok)
			}
			if lit != tt.input {
				t.Errorf("want literal=%q", tt.input)
				t.Errorf(" got literal=%q", lit)
			}

			got, err := ParseBoolean(lit)
			if err != nil != tt.wantErr {
				t.Fatalf("ParseBoolean: %v", err)
			} else if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("want=%#v", tt.want)
				t.Errorf(" got=%#v", got)
			}
		})
	}
}
