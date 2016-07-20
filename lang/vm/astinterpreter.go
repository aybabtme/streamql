package vm

import (
	"github.com/aybabtme/streamql/lang/ast"
)

func ASTInterpreter(tree *ast.FiltersStmt) Engine {
	return &astEngine{tree: tree}
}

type astEngine struct {
	tree *ast.FiltersStmt
}

func (vm *astEngine) Filter(in []Message) (out [][]Message) {
	for _, filter := range vm.tree.Filters {
		var filterOut []Message
		for _, msg := range in {
			filterOut = append(filterOut, vm.filter(msg, filter)...)
		}
		out = append(out, filterOut)
	}
	return out
}

func (vm *astEngine) filter(in Message, f *ast.FilterStmt) (out []Message) {
	intermediary := []Message{in}

	for _, s := range f.Selectors {
		var next []Message
		for _, inter := range intermediary {
			next = append(next, vm.selector(inter, s)...)
		}
		intermediary = next

	}
	return intermediary
}

func (vm *astEngine) selector(in Message, s *ast.SelectorStmt) (out []Message) {
	if s.Object != nil {
		return vm.objectSelector(in, s.Object)
	}
	if s.Array != nil {
		return vm.arraySelector(in, s.Array)
	}
	return []Message{in}
}

func (vm *astEngine) objectSelector(in Message, obj *ast.ObjectSelectorStmt) (out []Message) {
	child, ok := in.Member(obj.Member)
	if !ok {
		return nil
	}
	if obj.Child != nil {
		return vm.selector(child, obj.Child)
	}
	return []Message{child}
}

func (vm *astEngine) arraySelector(in Message, arr *ast.ArraySelectorStmt) []Message {
	var (
		childs []Message
		ok     bool
	)
	switch {
	case arr.Each != nil:
		childs, ok = in.Each()
		if !ok {
			return nil
		}
	case arr.RangeEach != nil:
		childs, ok = in.Range(arr.RangeEach.From, arr.RangeEach.To)
		if !ok {
			return nil
		}
	case arr.Index != nil:
		child, ok := in.Index(arr.Index.Index)
		if !ok {
			return nil
		}
		childs = []Message{child}
	}
	var out []Message
	if arr.Child != nil {
		for _, child := range childs {
			out = append(out, vm.selector(child, arr.Child)...)
		}
	} else {
		out = childs
	}
	return out
}
