package vm

import (
	"github.com/aybabtme/streamql/lang/ast"
	"github.com/aybabtme/streamql/lang/vm/msg"
)

func ASTInterpreter(tree *ast.FilterStmt) Engine {
	return &astEngine{tree: tree}
}

type astEngine struct {
	tree *ast.FilterStmt
}

func (vm *astEngine) Filter(in Source, sink Sink) {
	for {
		msg, more := in()
		if !more {
			return
		}
		for _, out := range vm.filter(msg, vm.tree) {
			if !sink(out) {
				return
			}
		}
	}
}

func (vm *astEngine) filter(in msg.Message, f *ast.FilterStmt) (out []msg.Message) {
	intermediary := []msg.Message{in}

	for _, s := range f.Selectors {
		var next []msg.Message
		for _, inter := range intermediary {
			next = append(next, vm.selector(inter, s)...)
		}
		intermediary = next

	}
	return intermediary
}

func (vm *astEngine) selector(in msg.Message, s *ast.SelectorStmt) (out []msg.Message) {
	if s.Object != nil {
		return vm.objectSelector(in, s.Object)
	}
	if s.Array != nil {
		return vm.arraySelector(in, s.Array)
	}
	return []msg.Message{in}
}

func (vm *astEngine) objectSelector(in msg.Message, obj *ast.ObjectSelectorStmt) (out []msg.Message) {
	child, ok := in.Member(obj.Member)
	if !ok {
		return nil
	}
	if obj.Child != nil {
		return vm.selector(child, obj.Child)
	}
	return []msg.Message{child}
}

func (vm *astEngine) arraySelector(in msg.Message, arr *ast.ArraySelectorStmt) []msg.Message {
	var (
		childs []msg.Message
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
		childs = []msg.Message{child}
	}
	var out []msg.Message
	if arr.Child != nil {
		for _, child := range childs {
			out = append(out, vm.selector(child, arr.Child)...)
		}
	} else {
		out = childs
	}
	return out
}
