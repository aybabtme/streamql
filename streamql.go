package streamql

import (
	"fmt"
	"strings"

	"github.com/aybabtme/streamql/lang/parser"
	"github.com/aybabtme/streamql/lang/vm"
)

// MustCompile is like Compile but panics if the query is
// invalid.
func MustCompile(query string) vm.Engine {
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
func Compile(query string) (vm.Engine, error) {
	tree, err := parser.NewParser(strings.NewReader(query)).Parse()
	if err != nil {
		return nil, err
	}
	if len(tree.Filters) != 1 {
		return nil, fmt.Errorf("expect exactly 1 query, got %d", len(tree.Filters))
	}
	return vm.ASTInterpreter(tree.Filters[0]), nil
}
