package grammar

import "fmt"

type token string

const (
	memberString = token(`member`)
	closeArray   = token(`]`)
	integer      = token(`42`)
	rangeSep     = token(`:`)
	epsilon      = token(``)
	openArray    = token(`[`)
	thisSymbol   = token(`.`)
	pipeSep      = token(` | `)
	filterSep    = token(`, `)
)

type sentence struct {
	Pop  func() (token, bool)
	Push func(token)
	Peek func() (token, bool)
}

func parser(input []token) error {
	idx := 0
	str := &sentence{
		Pop: func() (token, bool) {
			if len(input) >= idx {
				return token(""), false
			}
			i := idx
			idx++
			return input[i], true
		},
		Push: func(t token) {
			idx--
			input[idx] = t
		},
		Peek: func() (token, bool) {
			if len(input) >= idx {
				return token(""), false
			}
			return input[idx], true
		},
	}
	return start(str)
}

// Start -> Filters.
func start(str *sentence) error {
	return or(
		"Start",
		and("-> Filters", filters),
	)(str)
}

// Filters -> Filter FilterChain.
func filters(str *sentence) error {
	return or(
		"Filters",
		and("-> Filter FilterChain", filter, filterChain),
	)(str)
}

// FilterChain -> filter_sep Filters
//              | epsilon.
func filterChain(str *sentence) error {
	return or(
		"FilterChain",
		and("-> filter_sep Filters", terminal(filterSep), filters),
		and("-> epsilon", terminal(epsilon)),
	)(str)
}

// Filter -> Selector SelectorChain.
func filter(str *sentence) error {
	return or(
		"Filter",
		and("-> Selector SelectorChain", selector, selectorChain),
	)(str)
}

// SelectorChain -> pipe_sep Filter
//                | epsilon.
func selectorChain(str *sentence) error {
	return or(
		"SelectorChain",
		and("-> pipe_sep Filter", terminal(pipeSep), filter),
		and("-> epsilon", terminal(epsilon)),
	)(str)
}

// Selector -> this_symbol TypedSelector.
func selector(str *sentence) error {
	return or(
		"Selector",
		and("-> this_symbol TypedSelector", terminal(thisSymbol), typedSelector),
	)(str)
}

// TypedSelector -> member_string
//                | open_array ArraySelector
//                | epsilon.
func typedSelector(str *sentence) error {
	return or(
		"TypedSelector",
		and("-> member_string", terminal(memberString)),
		and("-> open_array ArraySelector", terminal(openArray), arraySelector),
		and("-> epsilon", terminal(epsilon)),
	)(str)
}

// ArraySelector -> close_array
//                | integer TypedArraySelector.
func arraySelector(str *sentence) error {
	return or(
		"ArraySelector",
		and("-> close_array", terminal(closeArray)),
		and("-> integer TypedArraySelector", terminal(integer), typedArraySelector),
	)(str)
}

// TypedArraySelector -> range_sep integer close_array
//                     | close_array.
func typedArraySelector(str *sentence) error {
	return or(
		"TypedArraySelector",
		and("-> range_sep integer close_array", terminal(rangeSep), terminal(integer), terminal(closeArray)),
		and("-> close_array", terminal(closeArray)),
	)(str)
}

// scaffolding

type rule func(*sentence) error

func and(name string, rules ...rule) rule {
	return func(str *sentence) error {
		for _, rule := range rules {
			if err := rule(str); err != nil {
				return fmt.Errorf("%s\n-> %v", name, err)
			}
		}
		return nil
	}
}

func or(name string, rules ...rule) rule {
	return func(str *sentence) error {
		for _, rule := range rules {
			if err := rule(str); err != nil {
				continue
			}
		}
		cur, ok := str.Peek()
		if !ok {
			return fmt.Errorf("%s\nreached end of stream but production rules are not satisfied yet", name)
		}
		return fmt.Errorf("%s\nno rule can be matched for %#v", name, cur)
	}
}

func terminal(t token) rule {
	return func(str *sentence) error {
		cur, ok := str.Peek()
		if !ok && t == epsilon {
			return nil
		}
		if !ok {
			return fmt.Errorf("expect token %#v, reached end of stream", t)
		}
		if cur != t {
			return fmt.Errorf("expect token %#v, got %#v", t, cur)
		}
		return nil
	}
}
