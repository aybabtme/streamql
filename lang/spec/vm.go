package spec

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
)

type ASTInterpreter struct {
	tree *AST
}

func (vm *ASTInterpreter) Run(build Builder, src Source, sink Sink) error {
	if vm.tree.Expr == nil {
		for {
			msg, more, err := src()
			if err != nil {
				return err
			}
			if !more {
				return nil
			}
			if err := sink(msg); err != nil {
				return err
			}
		}
	}

	for {
		msg, more, err := src()
		if err != nil {
			return err
		}
		if !more {
			return nil
		}
		if err := vm.evalExpr(build, msg, vm.tree.Expr, sink); err != nil {
			return err
		}
	}
}

func (vm *ASTInterpreter) skipEvalWrongType(action string, got MsgT, want ...MsgT) error {
	str := fmt.Sprintf("%s is not defined on %v (can be done on ", action, got)
	if len(want) >= 1 {
		str += fmt.Sprintf("%v", want[1])
	}
	for _, w := range want[1:] {
		str += fmt.Sprintf(" or %v", w)
	}
	return errors.New(str + ")")
}

func (vm *ASTInterpreter) skipEvalWrongArgType(action string, target MsgT, arg MsgT, want ...MsgT) error {
	str := fmt.Sprintf("%s by %v is not defined on %v (can be done by ", action, arg, target)
	if len(want) >= 1 {
		str += fmt.Sprintf("%v", want[1])
	}
	for _, w := range want[1:] {
		str += fmt.Sprintf(" or %v", w)
	}
	return errors.New(str + ")")
}

func (vm *ASTInterpreter) skipEvalWrongArgValue(action string, arg MsgT, problem string) error {
	return fmt.Errorf("%s with given %v is impossible: %s", action, arg, problem)
}

func (vm *ASTInterpreter) evalExpr(build Builder, msg Msg, expr *Expr, sink Sink) error {

	if expr.Next != nil {
		sink = func(msg Msg) error { return vm.evalExpr(build, msg, expr.Next, sink) }
	}

	switch {
	case expr.Literal != nil:
		return vm.evalLiteral(build, msg, expr.Literal, sink)
	case expr.Selector != nil:
		return vm.evalSelector(build, msg, expr.Selector, sink)
	case expr.Operator != nil:
		return vm.evalOperator(build, msg, expr.Operator, sink)
	case expr.FuncCall != nil:
		return vm.evalFuncCall(build, msg, expr.FuncCall, sink)
	default:
		panic("invalid expression in AST has no possible evaluation branches")
	}
}

func (vm *ASTInterpreter) evalLiteral(build Builder, msg Msg, l *Literal, sink Sink) error {
	switch {
	case l.Bool != nil:
		return sink(build.Bool(*l.Bool))
	case l.String != nil:
		return sink(build.String(*l.String))
	case l.Int != nil:
		return sink(build.Int(*l.Int))
	case l.Float != nil:
		return sink(build.Float(*l.Float))
	case l.Null != nil:
		return sink(build.Null())
	default:
		panic("invalid literal in AST has no possible evaluation branches")
	}
}

func (vm *ASTInterpreter) evalSelector(build Builder, msg Msg, s *Selector, sink Sink) error {
	switch {
	case s.Member != nil:
		return vm.evalMemberSelector(build, msg, s.Member, sink)
	case s.Slice != nil:
		return vm.evalSliceSelector(build, msg, s.Slice, sink)
	case s.Noop != nil:
		return sink(msg)
	default:
		panic("invalid selector in AST has no possible evaluation branches")
	}
}

func (vm *ASTInterpreter) evalMemberSelector(build Builder, msg Msg, m *MemberSelector, sink Sink) error {
	if m.Child != nil {
		sink = func(msg Msg) error { return vm.evalSelector(build, msg, m.Child, sink) }
	}

	// the meaning of an index depends on the type of message
	switch msg.Type() {

	case MsgTObject:
		member, err := vm.evalExprToMsgType(build, msg, m.Index, "index", MsgTString)
		if err != nil {
			return err
		}
		return sink(msg.Member(member.String()))

	case MsgTArray:
		pos, err := vm.evalExprToMsgType(build, msg, m.Index, "index", MsgTInt)
		if err != nil {
			return err
		}
		idx := pos.Int()
		if idx < 0 || idx > msg.Len() {
			return vm.skipEvalWrongArgValue("index", pos.Type(), "index is out of range")
		}
		return sink(msg.Index(idx))

	default:
		return vm.skipEvalWrongType("index", msg.Type(), MsgTObject, MsgTArray)
	}
}

func (vm *ASTInterpreter) evalSliceSelector(build Builder, msg Msg, s *SliceSelector, sink Sink) error {
	if s.Child != nil {
		sink = func(msg Msg) error { return vm.evalSelector(build, msg, s.Child, sink) }
	}

	if msg.Type() != MsgTArray {
		return vm.skipEvalWrongType("index", msg.Type(), MsgTObject, MsgTArray)
	}

	from, err := vm.evalExprToMsgType(build, msg, s.From, "slice from", MsgTInt)
	if err != nil {
		return err
	}
	to, err := vm.evalExprToMsgType(build, msg, s.To, "slice to", MsgTInt)
	if err != nil {
		return err
	}

	src := msg.Slice(from.Int(), to.Int())
	for {
		msg, more, err := src()
		if err != nil {
			return err
		}
		if !more {
			return nil
		}
		if err := sink(msg); err != nil {
			return err
		}
	}
}

func (vm *ASTInterpreter) evalOperator(build Builder, msg Msg, o *Operator, sink Sink) error {

	// bool operators
	switch {
	case o.LogNot != nil:
		arg, err := vm.evalExprToMsgType(build, msg, o.LogNot.Arg, "logical not", MsgTBool)
		if err != nil {
			return err
		}
		return sink(build.Bool(!arg.Bool()))

	case o.LogAnd != nil:
		lhs, err := vm.evalExprToMsgType(build, msg, o.LogAnd.LHS, "left of logical and", MsgTBool)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, msg, o.LogAnd.RHS, "right of logical and", MsgTBool)
		if err != nil {
			return err
		}
		return sink(build.Bool(lhs.Bool() && rhs.Bool()))

	case o.LogOr != nil:
		lhs, err := vm.evalExprToMsgType(build, msg, o.LogOr.LHS, "left of logical and", MsgTBool)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, msg, o.LogOr.RHS, "right of logical and", MsgTBool)
		if err != nil {
			return err
		}
		return sink(build.Bool(lhs.Bool() || rhs.Bool()))
	}

	// numerical operators
	switch {
	case o.NumAdd != nil:
		lhs, err := vm.evalExprToMsgType(build, msg, o.NumAdd.LHS, "left of an addition", MsgTInt, MsgTFloat, MsgTString)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, msg, o.NumAdd.RHS, "right of an addition", MsgTInt, MsgTFloat, MsgTString)
		if err != nil {
			return err
		}
		switch lhs.Type() {
		case MsgTInt:
			switch rhs.Type() {
			case MsgTInt: // Int + Int
				return sink(build.Int(lhs.Int() + rhs.Int()))
			case MsgTFloat: // Int + Float, promote Int
				return sink(build.Float(float64(lhs.Int()) + rhs.Float()))
			case MsgTString: // Int + String, promote Int
				return sink(build.String(strconv.FormatInt(lhs.Int(), 10) + rhs.String()))
			}
		case MsgTFloat:
			switch rhs.Type() {
			case MsgTInt: // Float + Int, promote Int
				return sink(build.Float(lhs.Float() + float64(rhs.Int())))
			case MsgTFloat: // Float + Float
				return sink(build.Float(lhs.Float() + rhs.Float()))
			case MsgTString: // Float + String, promote Float
				return sink(build.String(
					strconv.FormatFloat(lhs.Float(), 'g', -1, 64) + rhs.String(),
				))
			}
		case MsgTString:
			switch rhs.Type() {
			case MsgTInt: // String + Int, promote Int
				return sink(build.String(
					lhs.String() + strconv.FormatInt(rhs.Int(), 10),
				))
			case MsgTFloat: // String + Float, promote Float
				return sink(build.String(
					lhs.String() + strconv.FormatFloat(rhs.Float(), 'g', -1, 64),
				))
			case MsgTString: // String + String
				return sink(build.String(lhs.String() + rhs.String()))
			}
		}
		panic("missing case")

	case o.NumSub != nil:
		lhs, err := vm.evalExprToMsgType(build, msg, o.NumSub.LHS, "left of a subtraction", MsgTInt, MsgTFloat)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, msg, o.NumSub.RHS, "right of a subtraction", MsgTInt, MsgTFloat)
		if err != nil {
			return err
		}
		switch lhs.Type() {
		case MsgTInt:
			switch rhs.Type() {
			case MsgTInt: // Int - Int
				return sink(build.Int(lhs.Int() - rhs.Int()))
			case MsgTFloat: // Int - Float, promote Int
				return sink(build.Float(float64(lhs.Int()) - rhs.Float()))
			}
		case MsgTFloat:
			switch rhs.Type() {
			case MsgTInt: // Float - Int, promote Int
				return sink(build.Float(lhs.Float() - float64(rhs.Int())))
			case MsgTFloat: // Float - Float
				return sink(build.Float(lhs.Float() - rhs.Float()))
			}
			panic("missing case")
		}

	case o.NumDiv != nil:
		lhs, err := vm.evalExprToMsgType(build, msg, o.NumDiv.LHS, "left of a division", MsgTInt, MsgTFloat)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, msg, o.NumDiv.RHS, "right of a division", MsgTInt, MsgTFloat)
		if err != nil {
			return err
		}
		switch {
		case rhs.Type() == MsgTFloat && rhs.Float() == 0:
			return vm.skipEvalWrongArgValue("division", rhs.Type(), "can't divide by zero")
		case rhs.Type() == MsgTInt && rhs.Int() == 0:
			return vm.skipEvalWrongArgValue("division", rhs.Type(), "can't divide by zero")
		}
		switch lhs.Type() {
		case MsgTInt:
			switch rhs.Type() {
			case MsgTInt: // Int / Int
				return sink(build.Int(lhs.Int() / rhs.Int()))
			case MsgTFloat: // Int / Float, promote Int
				return sink(build.Float(float64(lhs.Int()) / rhs.Float()))
			}
		case MsgTFloat:
			switch rhs.Type() {
			case MsgTInt: // Float / Int, promote Int
				return sink(build.Float(lhs.Float() / float64(rhs.Int())))
			case MsgTFloat: // Float / Float
				return sink(build.Float(lhs.Float() / rhs.Float()))
			}
			panic("missing case")
		}

	case o.NumMul != nil:
		lhs, err := vm.evalExprToMsgType(build, msg, o.NumMul.LHS, "left of a multiplication", MsgTInt, MsgTFloat)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, msg, o.NumMul.RHS, "right of a multiplication", MsgTInt, MsgTFloat)
		if err != nil {
			return err
		}
		switch lhs.Type() {
		case MsgTInt:
			switch rhs.Type() {
			case MsgTInt: // Int * Int
				return sink(build.Int(lhs.Int() * rhs.Int()))
			case MsgTFloat: // Int * Float, promote Int
				return sink(build.Float(float64(lhs.Int()) * rhs.Float()))
			}
		case MsgTFloat:
			switch rhs.Type() {
			case MsgTInt: // Float * Int, promote Int
				return sink(build.Float(lhs.Float() * float64(rhs.Int())))
			case MsgTFloat: // Float * Float
				return sink(build.Float(lhs.Float() * rhs.Float()))
			}
			panic("missing case")
		}
	}

	var checkEq func(lhs, rhs Msg) bool
	checkEq = func(lhs, rhs Msg) bool {
		if lhs.Type() != rhs.Type() {
			return false
		}
		switch lhs.Type() {
		case MsgTObject:
			lhsKeys := lhs.Keys()
			rhsKeys := rhs.Keys()
			if len(lhsKeys) != len(rhsKeys) {
				return false
			}
			sort.Strings(lhsKeys)
			sort.Strings(rhsKeys)
			for i, lhsKey := range lhsKeys {
				rhsKey := rhsKeys[i]
				if lhsKey != rhsKey {
					return false
				}
				if !checkEq(lhs.Member(lhsKey), rhs.Member(rhsKey)) {
					return false
				}
			}
			return true

		case MsgTArray:
			if lhs.Len() != rhs.Len() {
				return false
			}
			for i := int64(0); i < lhs.Len(); i++ {
				if !checkEq(lhs.Index(i), rhs.Index(i)) {
					return false
				}
			}
			return true

		case MsgTString:
			return lhs.String() == rhs.String()
		case MsgTInt:
			return lhs.Int() == rhs.Int()
		case MsgTFloat:
			return lhs.Float() == rhs.Float()
		case MsgTBool:
			return lhs.Bool() == rhs.Bool()
		case MsgTNull:
			return lhs.IsNull() == rhs.IsNull()
		default:
			panic("missing case")
		}
	}

	var checkLess func(lhs, rhs Msg) (bool, error)
	checkLess = func(lhs, rhs Msg) (bool, error) {
		if lhs.Type() != rhs.Type() {
			return false, vm.skipEvalWrongType("comparison", rhs.Type(), lhs.Type())
		}
		switch lhs.Type() {
		case MsgTString:
			return lhs.String() < rhs.String(), nil
		case MsgTInt:
			return lhs.Int() < rhs.Int(), nil
		case MsgTFloat:
			return lhs.Float() < rhs.Float(), nil
		default:
			panic("missing case")
		}
	}

	// comparators
	switch {
	case o.CmpEq != nil:
		lhs, err := vm.evalExprToMsgType(build, msg, o.CmpEq.LHS, "left of an equality", MsgTInt, MsgTFloat, MsgTBool, MsgTString, MsgTArray, MsgTObject)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, msg, o.CmpEq.RHS, "right of an equality", MsgTInt, MsgTFloat, MsgTBool, MsgTString, MsgTArray, MsgTObject)
		if err != nil {
			return err
		}
		return sink(build.Bool(checkEq(lhs, rhs)))

	case o.CmpNotEq != nil:
		lhs, err := vm.evalExprToMsgType(build, msg, o.CmpNotEq.LHS, "left of a non-equality", MsgTInt, MsgTFloat, MsgTBool, MsgTString, MsgTArray, MsgTObject)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, msg, o.CmpNotEq.RHS, "right of a non-equality", MsgTInt, MsgTFloat, MsgTBool, MsgTString, MsgTArray, MsgTObject)
		if err != nil {
			return err
		}
		return sink(build.Bool(!checkEq(lhs, rhs)))

	case o.CmpGt != nil:
		lhs, err := vm.evalExprToMsgType(build, msg, o.CmpGt.LHS, "left of a greater-than comparison", MsgTInt, MsgTFloat, MsgTString)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, msg, o.CmpGt.RHS, "right of a greater-than comparison", MsgTInt, MsgTFloat, MsgTString)
		if err != nil {
			return err
		}
		isLess, err := checkLess(lhs, rhs)
		if err != nil {
			return err
		}
		return sink(build.Bool(!isLess))

	case o.CmpGtOrEq != nil:
		lhs, err := vm.evalExprToMsgType(build, msg, o.CmpGtOrEq.LHS, "left of a greater-than-or-equal comparison", MsgTInt, MsgTFloat, MsgTString)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, msg, o.CmpGtOrEq.RHS, "right of a greater-than-or-equal comparison", MsgTInt, MsgTFloat, MsgTString)
		if err != nil {
			return err
		}
		isLess, err := checkLess(lhs, rhs)
		if err != nil {
			return err
		}
		return sink(build.Bool(checkEq(lhs, rhs) || !isLess))

	case o.CmpLs != nil:
		lhs, err := vm.evalExprToMsgType(build, msg, o.CmpLs.LHS, "left of a less-than comparison", MsgTInt, MsgTFloat, MsgTString)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, msg, o.CmpLs.RHS, "right of a less-than comparison", MsgTInt, MsgTFloat, MsgTString)
		if err != nil {
			return err
		}
		isLess, err := checkLess(lhs, rhs)
		if err != nil {
			return err
		}
		return sink(build.Bool(isLess))

	case o.CmpLsOrEq != nil:
		lhs, err := vm.evalExprToMsgType(build, msg, o.CmpLsOrEq.LHS, "left of a less-than-or-equal comparison", MsgTInt, MsgTFloat, MsgTString)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, msg, o.CmpLsOrEq.RHS, "right of a less-than-or-equal comparison", MsgTInt, MsgTFloat, MsgTString)
		if err != nil {
			return err
		}
		isLess, err := checkLess(lhs, rhs)
		if err != nil {
			return err
		}
		return sink(build.Bool(checkEq(lhs, rhs) || isLess))

	default:
		panic("invalid operator in AST has no possible evaluation branches")
	}

}

func (vm *ASTInterpreter) evalFuncCall(build Builder, msg Msg, f *FuncCall, sink Sink) error {
	arity, fn := vm.lookupFuncs(f.Name)
	if fn == nil {
		return fmt.Errorf("unknown function %q", f.Name)
	}
	if arity != len(f.Args) {
		return fmt.Errorf("function %q requires %d arguments, %d were given", f.Name, arity, len(f.Args))
	}
	return fn(build, msg, f.Args, sink)
}

type evalFunc func(build Builder, msg Msg, args []*Expr, sink Sink) error

func (vm *ASTInterpreter) lookupFuncs(name string) (int, evalFunc) {
	switch name {
	case "select":
		return 1, vm.evalFuncSelect
	}
	return 0, nil
}

// == select(bool) -> Msg ==
// Emits the current message if the given expression evaluates to true.
func (vm *ASTInterpreter) evalFuncSelect(build Builder, msg Msg, args []*Expr, sink Sink) error {
	cond, err := vm.evalExprToMsgType(build, msg, args[0], "function select", MsgTBool)
	if err != nil {
		return err
	}
	if !cond.Bool() {
		return sink(msg)
	}
	return nil
}

// helper

// evalExprToMsg evaluates an expression's result and verifies that it is of the requested type.
func (vm *ASTInterpreter) evalExprToMsg(build Builder, msg Msg, expr *Expr) (Msg, error) {
	var evaled Msg
	err := vm.evalExpr(build, msg, expr, func(got Msg) error {
		evaled = got
		return nil

	})
	return evaled, err
}

// evalExprToMsgType evaluates an expression's result and verifies that it is of the requested type.
func (vm *ASTInterpreter) evalExprToMsgType(build Builder, msg Msg, expr *Expr, action string, want ...MsgT) (Msg, error) {
	var evaled Msg
	err := vm.evalExpr(build, msg, expr, func(got Msg) error {
		for _, w := range want {
			if got.Type() != w {
				continue
			}
			evaled = got
			return nil
		}
		return vm.skipEvalWrongArgType(action, msg.Type(), got.Type(), want...)

	})
	return evaled, err
}
