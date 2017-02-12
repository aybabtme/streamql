package spec

import "testing"
import "strings"

func TestVM(t *testing.T) {
	r := strings.NewReader(`.`)

	ast, err := Parse(r)
	if err != nil {
		t.Fatal(err)
	}
	vm := &ASTInterpreter{tree: ast}
	_ = vm
}
