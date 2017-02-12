package spec

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/aybabtme/streamql/lang/spec/msg"
)

type ASTInterpreter struct {
	tree *AST
}

func (vm *ASTInterpreter) Run(build msg.Builder, src msg.Source, sink msg.Sink) error {
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

func (vm *ASTInterpreter) skipEvalWrongType(action string, got msg.Type, want ...msg.Type) error {
	str := fmt.Sprintf("%s is not defined on %v (can be done on ", action, got)
	if len(want) >= 1 {
		str += fmt.Sprintf("%v", want[1])
	}
	for _, w := range want[1:] {
		str += fmt.Sprintf(" or %v", w)
	}
	return errors.New(str + ")")
}

func (vm *ASTInterpreter) skipEvalWrongArgType(action string, target msg.Type, arg msg.Type, want ...msg.Type) error {
	str := fmt.Sprintf("%s by %v is not defined on %v (can be done by ", action, arg, target)
	if len(want) >= 1 {
		str += fmt.Sprintf("%v", want[1])
	}
	for _, w := range want[1:] {
		str += fmt.Sprintf(" or %v", w)
	}
	return errors.New(str + ")")
}

func (vm *ASTInterpreter) skipEvalWrongArgValue(action string, arg msg.Type, problem string) error {
	return fmt.Errorf("%s with given %v is impossible: %s", action, arg, problem)
}

func (vm *ASTInterpreter) evalExpr(build msg.Builder, m msg.Msg, expr *Expr, sink msg.Sink) error {

	if expr.Next != nil {
		sink = func(m msg.Msg) error { return vm.evalExpr(build, m, expr.Next, sink) }
	}

	switch {
	case expr.Literal != nil:
		return vm.evalLiteral(build, m, expr.Literal, sink)
	case expr.Selector != nil:
		return vm.evalSelector(build, m, expr.Selector, sink)
	case expr.Operator != nil:
		return vm.evalOperator(build, m, expr.Operator, sink)
	case expr.FuncCall != nil:
		return vm.evalFuncCall(build, m, expr.FuncCall, sink)
	default:
		panic("invalid expression in AST has no possible evaluation branches")
	}
}

func (vm *ASTInterpreter) evalLiteral(build msg.Builder, m msg.Msg, l *Literal, sink msg.Sink) error {
	switch {
	case l.Bool != nil:
		v, err := build.Bool(*l.Bool)
		if err != nil {
			return err
		}
		return sink(v)
	case l.String != nil:
		v, err := build.String(*l.String)
		if err != nil {
			return err
		}
		return sink(v)
	case l.Int != nil:
		v, err := build.Int(*l.Int)
		if err != nil {
			return err
		}
		return sink(v)
	case l.Float != nil:
		v, err := build.Float(*l.Float)
		if err != nil {
			return err
		}
		return sink(v)
	case l.Null != nil:
		v, err := build.Null()
		if err != nil {
			return err
		}
		return sink(v)
	default:
		panic("invalid literal in AST has no possible evaluation branches")
	}
}

func (vm *ASTInterpreter) evalSelector(build msg.Builder, m msg.Msg, s *Selector, sink msg.Sink) error {
	switch {
	case s.Member != nil:
		return vm.evalMemberSelector(build, m, s.Member, sink)
	case s.Slice != nil:
		return vm.evalSliceSelector(build, m, s.Slice, sink)
	case s.Noop != nil:
		return sink(m)
	default:
		panic("invalid selector in AST has no possible evaluation branches")
	}
}

func (vm *ASTInterpreter) evalMemberSelector(build msg.Builder, m msg.Msg, sel *MemberSelector, sink msg.Sink) error {
	if sel.Child != nil {
		sink = func(m msg.Msg) error { return vm.evalSelector(build, m, sel.Child, sink) }
	}

	// the meaning of an index depends on the type of message
	switch m.Type() {

	case msg.TypeObject:
		member, err := vm.evalExprToMsgType(build, m, sel.Index, "index", msg.TypeString)
		if err != nil {
			return err
		}
		return sink(m.Member(member.StringVal()))

	case msg.TypeArray:
		pos, err := vm.evalExprToMsgType(build, m, sel.Index, "index", msg.TypeInt)
		if err != nil {
			return err
		}
		idx := pos.IntVal()
		if idx < 0 || idx > m.Len() {
			return vm.skipEvalWrongArgValue("index", pos.Type(), "index is out of range")
		}
		return sink(m.Index(idx))

	default:
		return vm.skipEvalWrongType("index", m.Type(), msg.TypeObject, msg.TypeArray)
	}
}

func (vm *ASTInterpreter) evalSliceSelector(build msg.Builder, m msg.Msg, s *SliceSelector, sink msg.Sink) error {
	if s.Child != nil {
		sink = func(m msg.Msg) error { return vm.evalSelector(build, m, s.Child, sink) }
	}

	if m.Type() != msg.TypeArray {
		return vm.skipEvalWrongType("index", m.Type(), msg.TypeObject, msg.TypeArray)
	}

	from, err := vm.evalExprToMsgType(build, m, s.From, "slice from", msg.TypeInt)
	if err != nil {
		return err
	}
	to, err := vm.evalExprToMsgType(build, m, s.To, "slice to", msg.TypeInt)
	if err != nil {
		return err
	}

	src := m.Slice(from.IntVal(), to.IntVal())
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

func (vm *ASTInterpreter) evalOperator(build msg.Builder, m msg.Msg, o *Operator, sink msg.Sink) error {

	// bool operators
	switch {
	case o.LogNot != nil:
		arg, err := vm.evalExprToMsgType(build, m, o.LogNot.Arg, "logical not", msg.TypeBool)
		if err != nil {
			return err
		}
		v, err := build.Bool(!arg.BoolVal())
		if err != nil {
			return err
		}
		return sink(v)

	case o.LogAnd != nil:
		lhs, err := vm.evalExprToMsgType(build, m, o.LogAnd.LHS, "left of logical and", msg.TypeBool)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, m, o.LogAnd.RHS, "right of logical and", msg.TypeBool)
		if err != nil {
			return err
		}
		v, err := build.Bool(lhs.BoolVal() && rhs.BoolVal())
		if err != nil {
			return err
		}
		return sink(v)

	case o.LogOr != nil:
		lhs, err := vm.evalExprToMsgType(build, m, o.LogOr.LHS, "left of logical and", msg.TypeBool)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, m, o.LogOr.RHS, "right of logical and", msg.TypeBool)
		if err != nil {
			return err
		}
		v, err := build.Bool(lhs.BoolVal() || rhs.BoolVal())
		if err != nil {
			return err
		}
		return sink(v)
	}

	// numerical operators
	switch {
	case o.NumAdd != nil:
		lhs, err := vm.evalExprToMsgType(build, m, o.NumAdd.LHS, "left of an addition", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, m, o.NumAdd.RHS, "right of an addition", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		switch lhs.Type() {
		case msg.TypeInt:
			switch rhs.Type() {
			case msg.TypeInt: // Int + Int
				v, err := build.Int(lhs.IntVal() + rhs.IntVal())
				if err != nil {
					return err
				}
				return sink(v)
			case msg.TypeFloat: // Int + Float, promote Int
				v, err := build.Float(float64(lhs.IntVal()) + rhs.FloatVal())
				if err != nil {
					return err
				}
				return sink(v)
			case msg.TypeString: // Int + String, promote Int
				v, err := build.String(strconv.FormatInt(lhs.IntVal(), 10) + rhs.StringVal())
				if err != nil {
					return err
				}
				return sink(v)
			}
		case msg.TypeFloat:
			switch rhs.Type() {
			case msg.TypeInt: // Float + Int, promote Int
				v, err := build.Float(lhs.FloatVal() + float64(rhs.IntVal()))
				if err != nil {
					return err
				}
				return sink(v)
			case msg.TypeFloat: // Float + Float
				v, err := build.Float(lhs.FloatVal() + rhs.FloatVal())
				if err != nil {
					return err
				}
				return sink(v)
			case msg.TypeString: // Float + String, promote Float
				v, err := build.String(
					strconv.FormatFloat(lhs.FloatVal(), 'g', -1, 64) + rhs.StringVal(),
				)
				if err != nil {
					return err
				}
				return sink(v)
			}
		case msg.TypeString:
			switch rhs.Type() {
			case msg.TypeInt: // String + Int, promote Int
				v, err := build.String(
					lhs.StringVal() + strconv.FormatInt(rhs.IntVal(), 10),
				)
				if err != nil {
					return err
				}
				return sink(v)
			case msg.TypeFloat: // String + Float, promote Float
				v, err := build.String(
					lhs.StringVal() + strconv.FormatFloat(rhs.FloatVal(), 'g', -1, 64),
				)
				if err != nil {
					return err
				}
				return sink(v)
			case msg.TypeString: // String + String
				v, err := build.String(lhs.StringVal() + rhs.StringVal())
				if err != nil {
					return err
				}
				return sink(v)
			}
		}
		panic("missing case")

	case o.NumSub != nil:
		lhs, err := vm.evalExprToMsgType(build, m, o.NumSub.LHS, "left of a subtraction", msg.TypeInt, msg.TypeFloat)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, m, o.NumSub.RHS, "right of a subtraction", msg.TypeInt, msg.TypeFloat)
		if err != nil {
			return err
		}
		switch lhs.Type() {
		case msg.TypeInt:
			switch rhs.Type() {
			case msg.TypeInt: // Int - Int
				v, err := build.Int(lhs.IntVal() - rhs.IntVal())
				if err != nil {
					return err
				}
				return sink(v)
			case msg.TypeFloat: // Int - Float, promote Int
				v, err := build.Float(float64(lhs.IntVal()) - rhs.FloatVal())
				if err != nil {
					return err
				}
				return sink(v)
			}
		case msg.TypeFloat:
			switch rhs.Type() {
			case msg.TypeInt: // Float - Int, promote Int
				v, err := build.Float(lhs.FloatVal() - float64(rhs.IntVal()))
				if err != nil {
					return err
				}
				return sink(v)
			case msg.TypeFloat: // Float - Float
				v, err := build.Float(lhs.FloatVal() - rhs.FloatVal())
				if err != nil {
					return err
				}
				return sink(v)
			}
			panic("missing case")
		}

	case o.NumDiv != nil:
		lhs, err := vm.evalExprToMsgType(build, m, o.NumDiv.LHS, "left of a division", msg.TypeInt, msg.TypeFloat)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, m, o.NumDiv.RHS, "right of a division", msg.TypeInt, msg.TypeFloat)
		if err != nil {
			return err
		}
		switch {
		case rhs.Type() == msg.TypeFloat && rhs.FloatVal() == 0:
			return vm.skipEvalWrongArgValue("division", rhs.Type(), "can't divide by zero")
		case rhs.Type() == msg.TypeInt && rhs.IntVal() == 0:
			return vm.skipEvalWrongArgValue("division", rhs.Type(), "can't divide by zero")
		}
		switch lhs.Type() {
		case msg.TypeInt:
			switch rhs.Type() {
			case msg.TypeInt: // Int / Int
				v, err := build.Int(lhs.IntVal() / rhs.IntVal())
				if err != nil {
					return err
				}
				return sink(v)
			case msg.TypeFloat: // Int / Float, promote Int
				v, err := build.Float(float64(lhs.IntVal()) / rhs.FloatVal())
				if err != nil {
					return err
				}
				return sink(v)
			}
		case msg.TypeFloat:
			switch rhs.Type() {
			case msg.TypeInt: // Float / Int, promote Int
				v, err := build.Float(lhs.FloatVal() / float64(rhs.IntVal()))
				if err != nil {
					return err
				}
				return sink(v)
			case msg.TypeFloat: // Float / Float
				v, err := build.Float(lhs.FloatVal() / rhs.FloatVal())
				if err != nil {
					return err
				}
				return sink(v)
			}
			panic("missing case")
		}

	case o.NumMul != nil:
		lhs, err := vm.evalExprToMsgType(build, m, o.NumMul.LHS, "left of a multiplication", msg.TypeInt, msg.TypeFloat)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, m, o.NumMul.RHS, "right of a multiplication", msg.TypeInt, msg.TypeFloat)
		if err != nil {
			return err
		}
		switch lhs.Type() {
		case msg.TypeInt:
			switch rhs.Type() {
			case msg.TypeInt: // Int * Int
				v, err := build.Int(lhs.IntVal() * rhs.IntVal())
				if err != nil {
					return err
				}
				return sink(v)
			case msg.TypeFloat: // Int * Float, promote Int
				v, err := build.Float(float64(lhs.IntVal()) * rhs.FloatVal())
				if err != nil {
					return err
				}
				return sink(v)
			}
		case msg.TypeFloat:
			switch rhs.Type() {
			case msg.TypeInt: // Float * Int, promote Int
				v, err := build.Float(lhs.FloatVal() * float64(rhs.IntVal()))
				if err != nil {
					return err
				}
				return sink(v)
			case msg.TypeFloat: // Float * Float
				v, err := build.Float(lhs.FloatVal() * rhs.FloatVal())
				if err != nil {
					return err
				}
				return sink(v)
			}
			panic("missing case")
		}
	}

	var checkEq func(lhs, rhs msg.Msg) bool
	checkEq = func(lhs, rhs msg.Msg) bool {
		if lhs.Type() != rhs.Type() {
			return false
		}
		switch lhs.Type() {
		case msg.TypeObject:
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

		case msg.TypeArray:
			if lhs.Len() != rhs.Len() {
				return false
			}
			for i := int64(0); i < lhs.Len(); i++ {
				if !checkEq(lhs.Index(i), rhs.Index(i)) {
					return false
				}
			}
			return true

		case msg.TypeString:
			return lhs.StringVal() == rhs.StringVal()
		case msg.TypeInt:
			return lhs.IntVal() == rhs.IntVal()
		case msg.TypeFloat:
			return lhs.FloatVal() == rhs.FloatVal()
		case msg.TypeBool:
			return lhs.BoolVal() == rhs.BoolVal()
		case msg.TypeNull:
			return lhs.IsNull() == rhs.IsNull()
		default:
			panic("missing case")
		}
	}

	var checkLess func(lhs, rhs msg.Msg) (bool, error)
	checkLess = func(lhs, rhs msg.Msg) (bool, error) {
		if lhs.Type() != rhs.Type() {
			return false, vm.skipEvalWrongType("comparison", rhs.Type(), lhs.Type())
		}
		switch lhs.Type() {
		case msg.TypeString:
			return lhs.StringVal() < rhs.StringVal(), nil
		case msg.TypeInt:
			return lhs.IntVal() < rhs.IntVal(), nil
		case msg.TypeFloat:
			return lhs.FloatVal() < rhs.FloatVal(), nil
		default:
			panic("missing case")
		}
	}

	// comparators
	switch {
	case o.CmpEq != nil:
		lhs, err := vm.evalExprToMsgType(build, m, o.CmpEq.LHS, "left of an equality", msg.TypeInt, msg.TypeFloat, msg.TypeBool, msg.TypeString, msg.TypeArray, msg.TypeObject)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, m, o.CmpEq.RHS, "right of an equality", msg.TypeInt, msg.TypeFloat, msg.TypeBool, msg.TypeString, msg.TypeArray, msg.TypeObject)
		if err != nil {
			return err
		}
		v, err := build.Bool(checkEq(lhs, rhs))
		if err != nil {
			return err
		}
		return sink(v)

	case o.CmpNotEq != nil:
		lhs, err := vm.evalExprToMsgType(build, m, o.CmpNotEq.LHS, "left of a non-equality", msg.TypeInt, msg.TypeFloat, msg.TypeBool, msg.TypeString, msg.TypeArray, msg.TypeObject)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, m, o.CmpNotEq.RHS, "right of a non-equality", msg.TypeInt, msg.TypeFloat, msg.TypeBool, msg.TypeString, msg.TypeArray, msg.TypeObject)
		if err != nil {
			return err
		}
		v, err := build.Bool(!checkEq(lhs, rhs))
		if err != nil {
			return err
		}
		return sink(v)

	case o.CmpGt != nil:
		lhs, err := vm.evalExprToMsgType(build, m, o.CmpGt.LHS, "left of a greater-than comparison", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, m, o.CmpGt.RHS, "right of a greater-than comparison", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		isLess, err := checkLess(lhs, rhs)
		if err != nil {
			return err
		}
		v, err := build.Bool(!isLess)
		if err != nil {
			return err
		}
		return sink(v)

	case o.CmpGtOrEq != nil:
		lhs, err := vm.evalExprToMsgType(build, m, o.CmpGtOrEq.LHS, "left of a greater-than-or-equal comparison", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, m, o.CmpGtOrEq.RHS, "right of a greater-than-or-equal comparison", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		isLess, err := checkLess(lhs, rhs)
		if err != nil {
			return err
		}
		v, err := build.Bool(checkEq(lhs, rhs) || !isLess)
		if err != nil {
			return err
		}
		return sink(v)

	case o.CmpLs != nil:
		lhs, err := vm.evalExprToMsgType(build, m, o.CmpLs.LHS, "left of a less-than comparison", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, m, o.CmpLs.RHS, "right of a less-than comparison", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		isLess, err := checkLess(lhs, rhs)
		if err != nil {
			return err
		}
		v, err := build.Bool(isLess)
		if err != nil {
			return err
		}
		return sink(v)

	case o.CmpLsOrEq != nil:
		lhs, err := vm.evalExprToMsgType(build, m, o.CmpLsOrEq.LHS, "left of a less-than-or-equal comparison", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		rhs, err := vm.evalExprToMsgType(build, m, o.CmpLsOrEq.RHS, "right of a less-than-or-equal comparison", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		isLess, err := checkLess(lhs, rhs)
		if err != nil {
			return err
		}
		v, err := build.Bool(checkEq(lhs, rhs) || isLess)
		if err != nil {
			return err
		}
		return sink(v)

	default:
		panic("invalid operator in AST has no possible evaluation branches")
	}

}

func (vm *ASTInterpreter) evalFuncCall(build msg.Builder, m msg.Msg, f *FuncCall, sink msg.Sink) error {
	arity, fn := vm.lookupFuncs(f.Name)
	if fn == nil {
		return fmt.Errorf("unknown function %q", f.Name)
	}
	if arity != len(f.Args) {
		return fmt.Errorf("function %q requires %d arguments, %d were given", f.Name, arity, len(f.Args))
	}
	return fn(build, m, f.Args, sink)
}

type evalFunc func(build msg.Builder, m msg.Msg, args []*Expr, sink msg.Sink) error

func (vm *ASTInterpreter) lookupFuncs(name string) (int, evalFunc) {
	switch name {
	case "select":
		return 1, vm.evalFuncSelect
	}
	return 0, nil
}

// == select(bool) -> msg.Msg ==
// Emits the current message if the given expression evaluates to true.
func (vm *ASTInterpreter) evalFuncSelect(build msg.Builder, m msg.Msg, args []*Expr, sink msg.Sink) error {
	cond, err := vm.evalExprToMsgType(build, m, args[0], "function select", msg.TypeBool)
	if err != nil {
		return err
	}
	if !cond.BoolVal() {
		return sink(m)
	}
	return nil
}

// helper

// evalExprToMsg evaluates an expression's result and verifies that it is of the requested type.
func (vm *ASTInterpreter) evalExprToMsg(build msg.Builder, m msg.Msg, expr *Expr) (msg.Msg, error) {
	var evaled msg.Msg
	err := vm.evalExpr(build, m, expr, func(got msg.Msg) error {
		evaled = got
		return nil

	})
	return evaled, err
}

// evalExprToMsgType evaluates an expression's result and verifies that it is of the requested type.
func (vm *ASTInterpreter) evalExprToMsgType(build msg.Builder, m msg.Msg, expr *Expr, action string, want ...msg.Type) (msg.Msg, error) {
	var evaled msg.Msg
	err := vm.evalExpr(build, m, expr, func(got msg.Msg) error {
		for _, w := range want {
			if got.Type() != w {
				continue
			}
			evaled = got
			return nil
		}
		return vm.skipEvalWrongArgType(action, m.Type(), got.Type(), want...)

	})
	return evaled, err
}
