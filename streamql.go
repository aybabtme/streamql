package streamql

import (
	"strings"

	"github.com/aybabtme/streamql/lang/grammar"
	"github.com/aybabtme/streamql/lang/vm"
	"github.com/aybabtme/streamql/lang/vm/astvm"
)

// MustCompile is like Compile but panics if the query is
// invalid.
func MustCompile(query string) vm.VM {
	eng, err := Compile(query)
	if err != nil {
		panic(err)
	}
	return eng
}

// Compile parses the query and returns an error
// if the syntax is invalid or if the query doesn't
// have exactly 1 filter. It returns a VM that can
// execute the query.
func Compile(query string) (vm.VM, error) {
	tree, err := grammar.Parse(strings.NewReader(query))
	if err != nil {
		return nil, err
	}
	return astvm.Interpreter(tree, &vm.Options{}), nil
}
