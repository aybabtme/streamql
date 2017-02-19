package grammar

import (
	"reflect"
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {

	tests := []struct {
		name string
		args string
		want []tok
	}{
		{
			name: `tokDot`,
			args: `.`,
			want: []tok{{tokDot, `.`}},
		},
		{
			name: `tokComma`,
			args: `,`,
			want: []tok{{tokComma, `,`}},
		},
		{
			name: `tokLeftBracket`,
			args: `[`,
			want: []tok{{tokLeftBracket, `[`}},
		},
		{
			name: `tokRightBracket`,
			args: `]`,
			want: []tok{{tokRightBracket, `]`}},
		},
		{
			name: `tokLeftParens`,
			args: `(`,
			want: []tok{{tokLeftParens, `(`}},
		},
		{
			name: `tokRightParens`,
			args: `)`,
			want: []tok{{tokRightParens, `)`}},
		},
		{
			name: `tokColon`,
			args: `:`,
			want: []tok{{tokColon, `:`}},
		},
		{
			name: `tokPipe`,
			args: `|`,
			want: []tok{{tokPipe, `|`}},
		},
		{
			name: `tokLogNot`,
			args: `!`,
			want: []tok{{tokLogNot, `!`}},
		},
		{
			name: `tokLogAnd`,
			args: `&&`,
			want: []tok{{tokLogAnd, `&&`}},
		},
		{
			name: `tokLogOr`,
			args: `||`,
			want: []tok{{tokLogOr, `||`}},
		},
		{
			name: `tokNumAdd`,
			args: `+`,
			want: []tok{{tokNumAdd, `+`}},
		},
		{
			name: `tokNumSub`,
			args: `-`,
			want: []tok{{tokNumSub, `-`}},
		},
		{
			name: `tokNumMul`,
			args: `*`,
			want: []tok{{tokNumMul, `*`}},
		},
		{
			name: `tokNumDiv`,
			args: `/`,
			want: []tok{{tokNumDiv, `/`}},
		},
		{
			name: `tokCmpEq`,
			args: `==`,
			want: []tok{{tokCmpEq, `==`}},
		},
		{
			name: `tokCmpNotEq`,
			args: `!=`,
			want: []tok{{tokCmpNotEq, `!=`}},
		},
		{
			name: `tokCmpGt`,
			args: `>`,
			want: []tok{{tokCmpGt, `>`}},
		},
		{
			name: `tokCmpGtOrEq`,
			args: `>=`,
			want: []tok{{tokCmpGtOrEq, `>=`}},
		},
		{
			name: `tokCmpLs`,
			args: `<`,
			want: []tok{{tokCmpLs, `<`}},
		},
		{
			name: `tokCmpLsOrEq`,
			args: `<=`,
			want: []tok{{tokCmpLsOrEq, `<=`}},
		},

		{`tokIdentifier`, `hello`, []tok{{tokIdentifier, `hello`}}},
		{`tokIdentifier`, `helloHello`, []tok{{tokIdentifier, `helloHello`}}},
		{`tokIdentifier`, `hello01`, []tok{{tokIdentifier, `hello01`}}},
		{`tokIdentifier`, `hello`, []tok{{tokIdentifier, `hello`}}},
		{`tokIdentifier`, `hello_hello`, []tok{{tokIdentifier, `hello_hello`}}},
		{`tokIdentifier`, `_hello01`, []tok{{tokIdentifier, `_hello01`}}},
		{
			name: `true bool`,
			args: `true`,
			want: []tok{{tokBool, `true`}},
		},
		{
			name: `false bool`,
			args: `false`,
			want: []tok{{tokBool, `false`}},
		},
		{
			name: `null`,
			args: `null`,
			want: []tok{{tokNull, `null`}},
		},

		{`0`, "0", []tok{{tokInt, "0"}}},
		{`1`, "1", []tok{{tokInt, "1"}}},
		{`42`, "42", []tok{{tokInt, "42"}}},
		{`-1`, "-1", []tok{{tokNumSub, "-"}, {tokInt, "1"}}},
		{`-42`, "-42", []tok{{tokNumSub, "-"}, {tokInt, "42"}}},

		{`1-1`, "1-1", []tok{
			{tokInt, "1"},
			{tokNumSub, "-"},
			{tokInt, "1"},
		}},

		{`0.0`, "0.0", []tok{{tokFloat, "0.0"}}},
		{`1.0`, "1.0", []tok{{tokFloat, "1.0"}}},
		{`42.42`, "42.42", []tok{{tokFloat, "42.42"}}},
		{`-1.0`, "-1.0", []tok{{tokNumSub, "-"}, {tokFloat, "1.0"}}},
		{`-42.42`, "-42.42", []tok{{tokNumSub, "-"}, {tokFloat, "42.42"}}},

		{`1.0-1.0`, "1.0-1.0", []tok{
			{tokFloat, "1.0"},
			{tokNumSub, "-"},
			{tokFloat, "1.0"},
		}},

		{
			name: `string`,
			args: `""`,
			want: []tok{{tokString, `""`}},
		},
		{
			name: `string`,
			args: `"hello world"`,
			want: []tok{{tokString, `"hello world"`}},
		},
		{
			name: `string`,
			args: `"hello \"world"`,
			want: []tok{{tokString, `"hello \"world"`}},
		},
		{
			name: `string`,
			args: `"hello \\world"`,
			want: []tok{{tokString, `"hello \\world"`}},
		},
		{
			name: `string`,
			args: `"hello \"world\","`,
			want: []tok{{tokString, `"hello \"world\","`}},
		},
		{
			name: `string`,
			args: `"hello \"world\",\r\x00\u12af\U12afAF12"`,
			want: []tok{{tokString, `"hello \"world\",\r\x00\u12af\U12afAF12"`}},
		},

		{
			name: `actual query`,
			args: `.hello[0:1] | select(.is_red && .size == "large")`,
			want: []tok{
				{tokDot, "."},
				{tokIdentifier, "hello"},
				{tokLeftBracket, "["},
				{tokInt, "0"},
				{tokColon, ":"},
				{tokInt, "1"},
				{tokRightBracket, "]"},
				{tokPipe, "|"},
				{tokIdentifier, "select"},
				{tokLeftParens, "("},
				{tokDot, "."},
				{tokIdentifier, "is_red"},
				{tokLogAnd, "&&"},
				{tokDot, "."},
				{tokIdentifier, "size"},
				{tokCmpEq, "=="},
				{tokString, `"large"`},
				{tokRightParens, ")"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Tokenize(strings.NewReader(tt.args))
			if err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("want = %v", tt.want)
				t.Errorf(" got = %v", got)
			}
		})
	}
}
