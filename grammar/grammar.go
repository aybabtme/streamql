package grammar

import "fmt"

type token string

const (
	memberString = token("member")
	closeArray   = token("]")
	integer      = token("42")
	rangeSep     = token(":")
	epsilon      = token("")
	openArray    = token("[")
	thisSymbol   = token(".")
	pipeSep      = token(" | ")
	filterSep    = token(", ")
)

func parser(input <-chan token) error {
	select {
	case cur := <-input:
		return start(cur, input)
	default:
		return nil
	}
}

// Start -> Filters.
func start(cur token, input <-chan token) error {
	return or(
		"Start",
		and("-> Filters", filters),
	)(cur, input)
}

// Filters -> Filter FilterChain.
func filters(cur token, input <-chan token) error {
	return or(
		"Filters",
		and("-> Filter FilterChain", filter, filterChain),
	)(cur, input)
}

// FilterChain -> filter_sep Filters
//              | epsilon.
func filterChain(cur token, input <-chan token) error {
	return or(
		"FilterChain",
		and("-> filter_sep Filters", terminal(filterSep), filters),
		and("-> epsilon", terminal(epsilon)),
	)(cur, input)
}

// Filter -> Selector SelectorChain.
func filter(cur token, input <-chan token) error {
	return or(
		"Filter",
		and("-> Selector SelectorChain", selector, selectorChain),
	)(cur, input)
}

// SelectorChain -> pipe_sep Filter
//                | epsilon.
func selectorChain(cur token, input <-chan token) error {
	return or(
		"SelectorChain",
		and("-> pipe_sep Filter", terminal(pipeSep), filter),
		and("-> epsilon", terminal(epsilon)),
	)(cur, input)
}

// Selector -> this_symbol TypedSelector.
func selector(cur token, input <-chan token) error {
	return or(
		"Selector",
		and("-> this_symbol TypedSelector", terminal(thisSymbol), typedSelector),
	)(cur, input)
}

// TypedSelector -> member_string
//                | open_array ArraySelector
//                | epsilon.
func typedSelector(cur token, input <-chan token) error {
	return or(
		"TypedSelector",
		and("-> member_string", terminal(memberString)),
		and("-> open_array ArraySelector", terminal(openArray), arraySelector),
		and("-> epsilon", terminal(epsilon)),
	)(cur, input)
}

// ArraySelector -> close_array
//                | integer TypedArraySelector.
func arraySelector(cur token, input <-chan token) error {
	return or(
		"ArraySelector",
		and("-> close_array", terminal(closeArray)),
		and("-> integer TypedArraySelector", terminal(integer), typedArraySelector),
	)(cur, input)
}

// TypedArraySelector -> range_sep integer close_array
//                     | close_array.
func typedArraySelector(cur token, input <-chan token) error {
	return or(
		"TypedArraySelector",
		and("-> range_sep integer close_array", terminal(rangeSep), terminal(integer), terminal(closeArray)),
		and("-> close_array", terminal(closeArray)),
	)(cur, input)
}

// scaffolding

type rule func(token, <-chan token) error

func and(name string, rules ...rule) rule {
	return func(cur token, input <-chan token) error {
		for _, rule := range rules {
			if err := rule(cur, input); err != nil {
				return fmt.Errorf("%s\n-> %v", name, err)
			}
		}
		return nil
	}
}

func or(name string, rules ...rule) rule {
	return func(cur token, input <-chan token) error {
		for _, rule := range rules {
			if err := rule(cur, input); err != nil {
				continue
			}
		}
		return fmt.Errorf("%s\nno rule can be matched for %#v", name, cur)
	}
}

func terminal(t token) rule {
	return func(cur token, input <-chan token) error {

		if cur != t {
			return fmt.Errorf("expect token %#v, got %#v", t, cur)
		}
		return nil
	}
}
