package spec

import (
	"fmt"
)
import (
	"bufio"
	"io"
	"strings"
)

type frame struct {
	i            int
	s            string
	line, column int
}
type Lexer struct {
	// The lexer runs in its own goroutine, and communicates via channel 'ch'.
	ch chan frame
	// We record the level of nesting because the action could return, and a
	// subsequent call expects to pick up where it left off. In other words,
	// we're simulating a coroutine.
	// TODO: Support a channel-based variant that compatible with Go's yacc.
	stack []frame
	stale bool

	// The 'l' and 'c' fields were added for
	// https://github.com/wagerlabs/docker/blob/65694e801a7b80930961d70c69cba9f2465459be/buildfile.nex
	// Since then, I introduced the built-in Line() and Column() functions.
	l, c int

	parseResult interface{}

	// The following line makes it easy for scripts to insert fields in the
	// generated code.
	// [NEX_END_OF_LEXER_STRUCT]
}

// NewLexerWithInit creates a new Lexer object, runs the given callback on it,
// then returns it.
func NewLexerWithInit(in io.Reader, initFun func(*Lexer)) *Lexer {
	yylex := new(Lexer)
	if initFun != nil {
		initFun(yylex)
	}
	yylex.ch = make(chan frame)
	var scan func(in *bufio.Reader, ch chan frame, family []dfa, line, column int)
	scan = func(in *bufio.Reader, ch chan frame, family []dfa, line, column int) {
		// Index of DFA and length of highest-precedence match so far.
		matchi, matchn := 0, -1
		var buf []rune
		n := 0
		checkAccept := func(i int, st int) bool {
			// Higher precedence match? DFAs are run in parallel, so matchn is at most len(buf), hence we may omit the length equality check.
			if family[i].acc[st] && (matchn < n || matchi > i) {
				matchi, matchn = i, n
				return true
			}
			return false
		}
		var state [][2]int
		for i := 0; i < len(family); i++ {
			mark := make([]bool, len(family[i].startf))
			// Every DFA starts at state 0.
			st := 0
			for {
				state = append(state, [2]int{i, st})
				mark[st] = true
				// As we're at the start of input, follow all ^ transitions and append to our list of start states.
				st = family[i].startf[st]
				if -1 == st || mark[st] {
					break
				}
				// We only check for a match after at least one transition.
				checkAccept(i, st)
			}
		}
		atEOF := false
		for {
			if n == len(buf) && !atEOF {
				r, _, err := in.ReadRune()
				switch err {
				case io.EOF:
					atEOF = true
				case nil:
					buf = append(buf, r)
				default:
					panic(err)
				}
			}
			if !atEOF {
				r := buf[n]
				n++
				var nextState [][2]int
				for _, x := range state {
					x[1] = family[x[0]].f[x[1]](r)
					if -1 == x[1] {
						continue
					}
					nextState = append(nextState, x)
					checkAccept(x[0], x[1])
				}
				state = nextState
			} else {
			dollar: // Handle $.
				for _, x := range state {
					mark := make([]bool, len(family[x[0]].endf))
					for {
						mark[x[1]] = true
						x[1] = family[x[0]].endf[x[1]]
						if -1 == x[1] || mark[x[1]] {
							break
						}
						if checkAccept(x[0], x[1]) {
							// Unlike before, we can break off the search. Now that we're at the end, there's no need to maintain the state of each DFA.
							break dollar
						}
					}
				}
				state = nil
			}

			if state == nil {
				lcUpdate := func(r rune) {
					if r == '\n' {
						line++
						column = 0
					} else {
						column++
					}
				}
				// All DFAs stuck. Return last match if it exists, otherwise advance by one rune and restart all DFAs.
				if matchn == -1 {
					if len(buf) == 0 { // This can only happen at the end of input.
						break
					}
					lcUpdate(buf[0])
					buf = buf[1:]
				} else {
					text := string(buf[:matchn])
					buf = buf[matchn:]
					matchn = -1
					ch <- frame{matchi, text, line, column}
					if len(family[matchi].nest) > 0 {
						scan(bufio.NewReader(strings.NewReader(text)), ch, family[matchi].nest, line, column)
					}
					if atEOF {
						break
					}
					for _, r := range text {
						lcUpdate(r)
					}
				}
				n = 0
				for i := 0; i < len(family); i++ {
					state = append(state, [2]int{i, 0})
				}
			}
		}
		ch <- frame{-1, "", line, column}
	}
	go scan(bufio.NewReader(in), yylex.ch, dfas, 0, 0)
	return yylex
}

type dfa struct {
	acc          []bool           // Accepting states.
	f            []func(rune) int // Transitions.
	startf, endf []int            // Transitions at start and end of input.
	nest         []dfa
}

var dfas = []dfa{
	// [.]
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 46:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 46:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// \[
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 91:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 91:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// \]
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 93:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 93:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// [(]
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 40:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 40:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// [)]
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 41:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 41:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// [:]
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 58:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 58:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// [|]
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 124:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 124:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// [!]
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 33:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 33:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// [&][&]
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 38:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 38:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 38:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// [|][|]
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 124:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 124:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 124:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// \+
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 43:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 43:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// \-
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 45:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 45:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// \*
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 42:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 42:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// \/
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 47:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 47:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// [=][=]
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 61:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 61:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 61:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// [!][=]
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 33:
				return 1
			case 61:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 33:
				return -1
			case 61:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 33:
				return -1
			case 61:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// [>]
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 62:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 62:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// [>][=]
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 61:
				return -1
			case 62:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 61:
				return 2
			case 62:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 61:
				return -1
			case 62:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// [<]
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 60:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 60:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// [<][=]
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 60:
				return 1
			case 61:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 60:
				return -1
			case 61:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 60:
				return -1
			case 61:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// [ \n\t\r]*
	{[]bool{true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 9:
				return 0
			case 10:
				return 0
			case 13:
				return 0
			case 32:
				return 0
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1}, []int{ /* End-of-input transitions */ -1}, nil},

	// true|false
	{[]bool{false, false, false, false, false, true, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 102:
				return 1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 2
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return 6
			case 101:
				return -1
			case 102:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 102:
				return -1
			case 108:
				return -1
			case 114:
				return 3
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 102:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return 4
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return 5
			case 102:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 102:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 102:
				return -1
			case 108:
				return 7
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 102:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return 8
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return 9
			case 102:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 102:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// null
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 108:
				return -1
			case 110:
				return 1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 108:
				return -1
			case 110:
				return -1
			case 117:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 108:
				return 3
			case 110:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 108:
				return 4
			case 110:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 108:
				return -1
			case 110:
				return -1
			case 117:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// [a-zA-Z_][a-zA-Z0-9_]*
	{[]bool{false, true, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 95:
				return 1
			}
			switch {
			case 48 <= r && r <= 57:
				return -1
			case 65 <= r && r <= 90:
				return 1
			case 97 <= r && r <= 122:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 95:
				return 2
			}
			switch {
			case 48 <= r && r <= 57:
				return 2
			case 65 <= r && r <= 90:
				return 2
			case 97 <= r && r <= 122:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 95:
				return 2
			}
			switch {
			case 48 <= r && r <= 57:
				return 2
			case 65 <= r && r <= 90:
				return 2
			case 97 <= r && r <= 122:
				return 2
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// \-?(0|[1-9][0-9]*)\.[0-9]+
	{[]bool{false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 45:
				return 1
			case 46:
				return -1
			case 48:
				return 2
			}
			switch {
			case 48 <= r && r <= 48:
				return -1
			case 49 <= r && r <= 57:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 45:
				return -1
			case 46:
				return -1
			case 48:
				return 2
			}
			switch {
			case 48 <= r && r <= 48:
				return -1
			case 49 <= r && r <= 57:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 45:
				return -1
			case 46:
				return 4
			case 48:
				return -1
			}
			switch {
			case 48 <= r && r <= 48:
				return -1
			case 49 <= r && r <= 57:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 45:
				return -1
			case 46:
				return 4
			case 48:
				return 5
			}
			switch {
			case 48 <= r && r <= 48:
				return 5
			case 49 <= r && r <= 57:
				return 5
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 45:
				return -1
			case 46:
				return -1
			case 48:
				return 6
			}
			switch {
			case 48 <= r && r <= 48:
				return 6
			case 49 <= r && r <= 57:
				return 6
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 45:
				return -1
			case 46:
				return 4
			case 48:
				return 5
			}
			switch {
			case 48 <= r && r <= 48:
				return 5
			case 49 <= r && r <= 57:
				return 5
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 45:
				return -1
			case 46:
				return -1
			case 48:
				return 6
			}
			switch {
			case 48 <= r && r <= 48:
				return 6
			case 49 <= r && r <= 57:
				return 6
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, nil},

	// \-?(0|[1-9][0-9]*)
	{[]bool{false, false, true, true, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 45:
				return 1
			case 48:
				return 2
			}
			switch {
			case 48 <= r && r <= 48:
				return -1
			case 49 <= r && r <= 57:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 45:
				return -1
			case 48:
				return 2
			}
			switch {
			case 48 <= r && r <= 48:
				return -1
			case 49 <= r && r <= 57:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 45:
				return -1
			case 48:
				return -1
			}
			switch {
			case 48 <= r && r <= 48:
				return -1
			case 49 <= r && r <= 57:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 45:
				return -1
			case 48:
				return 4
			}
			switch {
			case 48 <= r && r <= 48:
				return 4
			case 49 <= r && r <= 57:
				return 4
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 45:
				return -1
			case 48:
				return 4
			}
			switch {
			case 48 <= r && r <= 48:
				return 4
			case 49 <= r && r <= 57:
				return 4
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// ["]([^\\\"]|\\(a|b|f|n|r|t|v|\\|\'|"|x[0-9A-Fa-f][0-9A-Fa-f]|u[0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f]|U[0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f]))*["]
	{[]bool{false, false, true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 34:
				return 1
			case 39:
				return -1
			case 85:
				return -1
			case 92:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 102:
				return -1
			case 110:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 120:
				return -1
			}
			switch {
			case 48 <= r && r <= 57:
				return -1
			case 65 <= r && r <= 70:
				return -1
			case 97 <= r && r <= 102:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			case 39:
				return 3
			case 85:
				return 3
			case 92:
				return 4
			case 97:
				return 3
			case 98:
				return 3
			case 102:
				return 3
			case 110:
				return 3
			case 114:
				return 3
			case 116:
				return 3
			case 117:
				return 3
			case 118:
				return 3
			case 120:
				return 3
			}
			switch {
			case 48 <= r && r <= 57:
				return 3
			case 65 <= r && r <= 70:
				return 3
			case 97 <= r && r <= 102:
				return 3
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			case 39:
				return -1
			case 85:
				return -1
			case 92:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 102:
				return -1
			case 110:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 120:
				return -1
			}
			switch {
			case 48 <= r && r <= 57:
				return -1
			case 65 <= r && r <= 70:
				return -1
			case 97 <= r && r <= 102:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			case 39:
				return 3
			case 85:
				return 3
			case 92:
				return 4
			case 97:
				return 3
			case 98:
				return 3
			case 102:
				return 3
			case 110:
				return 3
			case 114:
				return 3
			case 116:
				return 3
			case 117:
				return 3
			case 118:
				return 3
			case 120:
				return 3
			}
			switch {
			case 48 <= r && r <= 57:
				return 3
			case 65 <= r && r <= 70:
				return 3
			case 97 <= r && r <= 102:
				return 3
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 34:
				return 5
			case 39:
				return 6
			case 85:
				return 7
			case 92:
				return 8
			case 97:
				return 9
			case 98:
				return 10
			case 102:
				return 11
			case 110:
				return 12
			case 114:
				return 13
			case 116:
				return 14
			case 117:
				return 15
			case 118:
				return 16
			case 120:
				return 17
			}
			switch {
			case 48 <= r && r <= 57:
				return -1
			case 65 <= r && r <= 70:
				return -1
			case 97 <= r && r <= 102:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			case 39:
				return 3
			case 85:
				return 3
			case 92:
				return 4
			case 97:
				return 3
			case 98:
				return 3
			case 102:
				return 3
			case 110:
				return 3
			case 114:
				return 3
			case 116:
				return 3
			case 117:
				return 3
			case 118:
				return 3
			case 120:
				return 3
			}
			switch {
			case 48 <= r && r <= 57:
				return 3
			case 65 <= r && r <= 70:
				return 3
			case 97 <= r && r <= 102:
				return 3
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			case 39:
				return 3
			case 85:
				return 3
			case 92:
				return 4
			case 97:
				return 3
			case 98:
				return 3
			case 102:
				return 3
			case 110:
				return 3
			case 114:
				return 3
			case 116:
				return 3
			case 117:
				return 3
			case 118:
				return 3
			case 120:
				return 3
			}
			switch {
			case 48 <= r && r <= 57:
				return 3
			case 65 <= r && r <= 70:
				return 3
			case 97 <= r && r <= 102:
				return 3
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			case 39:
				return -1
			case 85:
				return -1
			case 92:
				return -1
			case 97:
				return 24
			case 98:
				return 24
			case 102:
				return 24
			case 110:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 120:
				return -1
			}
			switch {
			case 48 <= r && r <= 57:
				return 24
			case 65 <= r && r <= 70:
				return 24
			case 97 <= r && r <= 102:
				return 24
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			case 39:
				return 3
			case 85:
				return 3
			case 92:
				return 4
			case 97:
				return 3
			case 98:
				return 3
			case 102:
				return 3
			case 110:
				return 3
			case 114:
				return 3
			case 116:
				return 3
			case 117:
				return 3
			case 118:
				return 3
			case 120:
				return 3
			}
			switch {
			case 48 <= r && r <= 57:
				return 3
			case 65 <= r && r <= 70:
				return 3
			case 97 <= r && r <= 102:
				return 3
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			case 39:
				return 3
			case 85:
				return 3
			case 92:
				return 4
			case 97:
				return 3
			case 98:
				return 3
			case 102:
				return 3
			case 110:
				return 3
			case 114:
				return 3
			case 116:
				return 3
			case 117:
				return 3
			case 118:
				return 3
			case 120:
				return 3
			}
			switch {
			case 48 <= r && r <= 57:
				return 3
			case 65 <= r && r <= 70:
				return 3
			case 97 <= r && r <= 102:
				return 3
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			case 39:
				return 3
			case 85:
				return 3
			case 92:
				return 4
			case 97:
				return 3
			case 98:
				return 3
			case 102:
				return 3
			case 110:
				return 3
			case 114:
				return 3
			case 116:
				return 3
			case 117:
				return 3
			case 118:
				return 3
			case 120:
				return 3
			}
			switch {
			case 48 <= r && r <= 57:
				return 3
			case 65 <= r && r <= 70:
				return 3
			case 97 <= r && r <= 102:
				return 3
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			case 39:
				return 3
			case 85:
				return 3
			case 92:
				return 4
			case 97:
				return 3
			case 98:
				return 3
			case 102:
				return 3
			case 110:
				return 3
			case 114:
				return 3
			case 116:
				return 3
			case 117:
				return 3
			case 118:
				return 3
			case 120:
				return 3
			}
			switch {
			case 48 <= r && r <= 57:
				return 3
			case 65 <= r && r <= 70:
				return 3
			case 97 <= r && r <= 102:
				return 3
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			case 39:
				return 3
			case 85:
				return 3
			case 92:
				return 4
			case 97:
				return 3
			case 98:
				return 3
			case 102:
				return 3
			case 110:
				return 3
			case 114:
				return 3
			case 116:
				return 3
			case 117:
				return 3
			case 118:
				return 3
			case 120:
				return 3
			}
			switch {
			case 48 <= r && r <= 57:
				return 3
			case 65 <= r && r <= 70:
				return 3
			case 97 <= r && r <= 102:
				return 3
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			case 39:
				return 3
			case 85:
				return 3
			case 92:
				return 4
			case 97:
				return 3
			case 98:
				return 3
			case 102:
				return 3
			case 110:
				return 3
			case 114:
				return 3
			case 116:
				return 3
			case 117:
				return 3
			case 118:
				return 3
			case 120:
				return 3
			}
			switch {
			case 48 <= r && r <= 57:
				return 3
			case 65 <= r && r <= 70:
				return 3
			case 97 <= r && r <= 102:
				return 3
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			case 39:
				return 3
			case 85:
				return 3
			case 92:
				return 4
			case 97:
				return 3
			case 98:
				return 3
			case 102:
				return 3
			case 110:
				return 3
			case 114:
				return 3
			case 116:
				return 3
			case 117:
				return 3
			case 118:
				return 3
			case 120:
				return 3
			}
			switch {
			case 48 <= r && r <= 57:
				return 3
			case 65 <= r && r <= 70:
				return 3
			case 97 <= r && r <= 102:
				return 3
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			case 39:
				return -1
			case 85:
				return -1
			case 92:
				return -1
			case 97:
				return 20
			case 98:
				return 20
			case 102:
				return 20
			case 110:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 120:
				return -1
			}
			switch {
			case 48 <= r && r <= 57:
				return 20
			case 65 <= r && r <= 70:
				return 20
			case 97 <= r && r <= 102:
				return 20
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			case 39:
				return 3
			case 85:
				return 3
			case 92:
				return 4
			case 97:
				return 3
			case 98:
				return 3
			case 102:
				return 3
			case 110:
				return 3
			case 114:
				return 3
			case 116:
				return 3
			case 117:
				return 3
			case 118:
				return 3
			case 120:
				return 3
			}
			switch {
			case 48 <= r && r <= 57:
				return 3
			case 65 <= r && r <= 70:
				return 3
			case 97 <= r && r <= 102:
				return 3
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			case 39:
				return -1
			case 85:
				return -1
			case 92:
				return -1
			case 97:
				return 18
			case 98:
				return 18
			case 102:
				return 18
			case 110:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 120:
				return -1
			}
			switch {
			case 48 <= r && r <= 57:
				return 18
			case 65 <= r && r <= 70:
				return 18
			case 97 <= r && r <= 102:
				return 18
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			case 39:
				return -1
			case 85:
				return -1
			case 92:
				return -1
			case 97:
				return 19
			case 98:
				return 19
			case 102:
				return 19
			case 110:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 120:
				return -1
			}
			switch {
			case 48 <= r && r <= 57:
				return 19
			case 65 <= r && r <= 70:
				return 19
			case 97 <= r && r <= 102:
				return 19
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			case 39:
				return 3
			case 85:
				return 3
			case 92:
				return 4
			case 97:
				return 3
			case 98:
				return 3
			case 102:
				return 3
			case 110:
				return 3
			case 114:
				return 3
			case 116:
				return 3
			case 117:
				return 3
			case 118:
				return 3
			case 120:
				return 3
			}
			switch {
			case 48 <= r && r <= 57:
				return 3
			case 65 <= r && r <= 70:
				return 3
			case 97 <= r && r <= 102:
				return 3
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			case 39:
				return -1
			case 85:
				return -1
			case 92:
				return -1
			case 97:
				return 21
			case 98:
				return 21
			case 102:
				return 21
			case 110:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 120:
				return -1
			}
			switch {
			case 48 <= r && r <= 57:
				return 21
			case 65 <= r && r <= 70:
				return 21
			case 97 <= r && r <= 102:
				return 21
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			case 39:
				return -1
			case 85:
				return -1
			case 92:
				return -1
			case 97:
				return 22
			case 98:
				return 22
			case 102:
				return 22
			case 110:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 120:
				return -1
			}
			switch {
			case 48 <= r && r <= 57:
				return 22
			case 65 <= r && r <= 70:
				return 22
			case 97 <= r && r <= 102:
				return 22
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			case 39:
				return -1
			case 85:
				return -1
			case 92:
				return -1
			case 97:
				return 23
			case 98:
				return 23
			case 102:
				return 23
			case 110:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 120:
				return -1
			}
			switch {
			case 48 <= r && r <= 57:
				return 23
			case 65 <= r && r <= 70:
				return 23
			case 97 <= r && r <= 102:
				return 23
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			case 39:
				return 3
			case 85:
				return 3
			case 92:
				return 4
			case 97:
				return 3
			case 98:
				return 3
			case 102:
				return 3
			case 110:
				return 3
			case 114:
				return 3
			case 116:
				return 3
			case 117:
				return 3
			case 118:
				return 3
			case 120:
				return 3
			}
			switch {
			case 48 <= r && r <= 57:
				return 3
			case 65 <= r && r <= 70:
				return 3
			case 97 <= r && r <= 102:
				return 3
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			case 39:
				return -1
			case 85:
				return -1
			case 92:
				return -1
			case 97:
				return 25
			case 98:
				return 25
			case 102:
				return 25
			case 110:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 120:
				return -1
			}
			switch {
			case 48 <= r && r <= 57:
				return 25
			case 65 <= r && r <= 70:
				return 25
			case 97 <= r && r <= 102:
				return 25
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			case 39:
				return -1
			case 85:
				return -1
			case 92:
				return -1
			case 97:
				return 26
			case 98:
				return 26
			case 102:
				return 26
			case 110:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 120:
				return -1
			}
			switch {
			case 48 <= r && r <= 57:
				return 26
			case 65 <= r && r <= 70:
				return 26
			case 97 <= r && r <= 102:
				return 26
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			case 39:
				return -1
			case 85:
				return -1
			case 92:
				return -1
			case 97:
				return 27
			case 98:
				return 27
			case 102:
				return 27
			case 110:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 120:
				return -1
			}
			switch {
			case 48 <= r && r <= 57:
				return 27
			case 65 <= r && r <= 70:
				return 27
			case 97 <= r && r <= 102:
				return 27
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			case 39:
				return -1
			case 85:
				return -1
			case 92:
				return -1
			case 97:
				return 28
			case 98:
				return 28
			case 102:
				return 28
			case 110:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 120:
				return -1
			}
			switch {
			case 48 <= r && r <= 57:
				return 28
			case 65 <= r && r <= 70:
				return 28
			case 97 <= r && r <= 102:
				return 28
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			case 39:
				return -1
			case 85:
				return -1
			case 92:
				return -1
			case 97:
				return 29
			case 98:
				return 29
			case 102:
				return 29
			case 110:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 120:
				return -1
			}
			switch {
			case 48 <= r && r <= 57:
				return 29
			case 65 <= r && r <= 70:
				return 29
			case 97 <= r && r <= 102:
				return 29
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			case 39:
				return -1
			case 85:
				return -1
			case 92:
				return -1
			case 97:
				return 30
			case 98:
				return 30
			case 102:
				return 30
			case 110:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 120:
				return -1
			}
			switch {
			case 48 <= r && r <= 57:
				return 30
			case 65 <= r && r <= 70:
				return 30
			case 97 <= r && r <= 102:
				return 30
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			case 39:
				return -1
			case 85:
				return -1
			case 92:
				return -1
			case 97:
				return 31
			case 98:
				return 31
			case 102:
				return 31
			case 110:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 120:
				return -1
			}
			switch {
			case 48 <= r && r <= 57:
				return 31
			case 65 <= r && r <= 70:
				return 31
			case 97 <= r && r <= 102:
				return 31
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			case 39:
				return 3
			case 85:
				return 3
			case 92:
				return 4
			case 97:
				return 3
			case 98:
				return 3
			case 102:
				return 3
			case 110:
				return 3
			case 114:
				return 3
			case 116:
				return 3
			case 117:
				return 3
			case 118:
				return 3
			case 120:
				return 3
			}
			switch {
			case 48 <= r && r <= 57:
				return 3
			case 65 <= r && r <= 70:
				return 3
			case 97 <= r && r <= 102:
				return 3
			}
			return 3
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// .
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			return 1
		},
		func(r rune) int {
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},
}

func NewLexer(in io.Reader) *Lexer {
	return NewLexerWithInit(in, nil)
}

// Text returns the matched text.
func (yylex *Lexer) Text() string {
	return yylex.stack[len(yylex.stack)-1].s
}

// Line returns the current line number.
// The first line is 0.
func (yylex *Lexer) Line() int {
	if len(yylex.stack) == 0 {
		return 0
	}
	return yylex.stack[len(yylex.stack)-1].line
}

// Column returns the current column number.
// The first column is 0.
func (yylex *Lexer) Column() int {
	if len(yylex.stack) == 0 {
		return 0
	}
	return yylex.stack[len(yylex.stack)-1].column
}

func (yylex *Lexer) next(lvl int) int {
	if lvl == len(yylex.stack) {
		l, c := 0, 0
		if lvl > 0 {
			l, c = yylex.stack[lvl-1].line, yylex.stack[lvl-1].column
		}
		yylex.stack = append(yylex.stack, frame{0, "", l, c})
	}
	if lvl == len(yylex.stack)-1 {
		p := &yylex.stack[lvl]
		*p = <-yylex.ch
		yylex.stale = false
	} else {
		yylex.stale = true
	}
	return yylex.stack[lvl].i
}
func (yylex *Lexer) pop() {
	yylex.stack = yylex.stack[:len(yylex.stack)-1]
}
func (yylex Lexer) Error(e string) {
	panic(e)
}

// Lex runs the lexer. Always returns 0.
// When the -s option is given, this function is not generated;
// instead, the NN_FUN macro runs the lexer.
func (yylex *Lexer) Lex(lval *yySymType) int {
OUTER0:
	for {
		switch yylex.next(0) {
		case 0:
			{
				lval.emit(tokDot)
			}
		case 1:
			{
				lval.emit(tokLeftBracket)
			}
		case 2:
			{
				lval.emit(tokRightBracket)
			}
		case 3:
			{
				lval.emit(tokLeftParens)
			}
		case 4:
			{
				lval.emit(tokRightParens)
			}
		case 5:
			{
				lval.emit(tokColon)
			}
		case 6:
			{
				lval.emit(tokPipe)
			}
		case 7:
			{
				lval.emit(tokLogNot)
			}
		case 8:
			{
				lval.emit(tokLogAnd)
			}
		case 9:
			{
				lval.emit(tokLogOr)
			}
		case 10:
			{
				lval.emit(tokNumAdd)
			}
		case 11:
			{
				lval.emit(tokNumSub)
			}
		case 12:
			{
				lval.emit(tokNumMul)
			}
		case 13:
			{
				lval.emit(tokNumDiv)
			}
		case 14:
			{
				lval.emit(tokCmpEq)
			}
		case 15:
			{
				lval.emit(tokCmpNotEq)
			}
		case 16:
			{
				lval.emit(tokCmpGt)
			}
		case 17:
			{
				lval.emit(tokCmpGtOrEq)
			}
		case 18:
			{
				lval.emit(tokCmpLs)
			}
		case 19:
			{
				lval.emit(tokCmpLsOrEq)
			}
		case 20:
			{
				lval.emit(tokWS)
			}
		case 21:
			{
				lval.emit(tokBool)
			}
		case 22:
			{
				lval.emit(tokNull)
			}
		case 23:
			{
				lval.emit(tokIdentifier)
			}
		case 24:
			{
				lval.emit(tokFloat)
			}
		case 25:
			{
				lval.emit(tokInt)
			}
		case 26:
			{
				lval.emit(tokString)
			}
		case 27:
			{
				lval.setError()
			}
		default:
			break OUTER0
		}
		continue
	}
	yylex.pop()

	return 0
}

type yySymType struct {
	lex    *Lexer
	tokens []tok
	err    error
}

func (yy *yySymType) emit(id string) {

	yy.tokens = append(yy.tokens, tok{id: id, lit: yy.lex.Text()})
}

func (yy *yySymType) setError() {
	yy.err = fmt.Errorf("%d:%d invalid argument after %q", yy.lex.Line(), yy.lex.Column(), yy.lex.Text())
}

func Tokenize(r io.Reader) ([]tok, error) {
	lex := NewLexer(r)
	v := &yySymType{lex: lex}
	lex.Lex(v)
	return v.tokens, v.err
}
