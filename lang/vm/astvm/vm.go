package astvm

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"

	"runtime"

	"strings"

	"regexp"

	"github.com/aybabtme/streamql/lang/ast"
	"github.com/aybabtme/streamql/lang/msg"
	"github.com/aybabtme/streamql/lang/vm"
)

const T bool = false

var indent int

func trace() func() {
	if !T {
		return func() {}
	}
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	lastSlash := strings.LastIndex(fn.Name(), "/") + 1
	pkgi := strings.Index(fn.Name()[lastSlash:], ".") + 1
	name := fn.Name()[lastSlash+pkgi:]

	log.Printf("%s<%s>", strings.Repeat("|  ", indent), name)
	indent++
	return func() {
		indent--
		log.Printf("%s</%s>", strings.Repeat("|  ", indent), name)
	}
}

type ASTInterpreter struct {
	opts *vm.Options
	tree *ast.AST
}

func Interpreter(tree *ast.AST, opts *vm.Options) vm.VM {
	return &ASTInterpreter{
		opts: opts,
		tree: tree,
	}
}

func (vm *ASTInterpreter) Run(build msg.Builder, src msg.Source, sink msg.Sink) error {
	defer trace()()
	if vm.tree.Expr == nil {
		for {
			msg, more, err := src()
			if err != nil {
				return err
			}
			if !more {
				return nil
			}
			err = sink(msg)
			switch err.(type) {
			case nil:
			case skipableError:
				if vm.opts.Strict {
					return err
				}
			default:
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
		err = vm.evalExpr(build, msg, vm.tree.Expr, sink)
		switch err.(type) {
		case nil:
		case skipableError:
			if vm.opts.Strict {
				return err
			}
		default:
			return err
		}
	}
}

type skipableError interface {
	IsSkipable()
}

type skipable struct{ error }

func (*skipable) IsSkipable() {}

func (vm *ASTInterpreter) skipEvalWrongType(action string, got msg.Type, want ...msg.Type) error {
	defer trace()()
	str := fmt.Sprintf("%s is not defined on %v (can be done on ", action, got)
	if len(want) >= 1 {
		str += fmt.Sprintf("%v", want[0])
	}
	for _, w := range want[1:] {
		str += fmt.Sprintf(" or %v", w)
	}
	return &skipable{errors.New(str + ")")}
}

func (vm *ASTInterpreter) skipEvalWrongArgType(action string, target msg.Type, arg msg.Type, want ...msg.Type) error {
	defer trace()()
	str := fmt.Sprintf("%s by %v is not defined on %v (can be done by ", action, arg, target)
	if len(want) >= 1 {
		str += fmt.Sprintf("%v", want[0])
	}
	for _, w := range want[1:] {
		str += fmt.Sprintf(" or %v", w)
	}
	return &skipable{errors.New(str + ")")}
}

func (vm *ASTInterpreter) skipEvalWrongArgValue(action string, arg msg.Type, problem string) error {
	defer trace()()
	return &skipable{fmt.Errorf("%s with given %v is impossible: %s", action, arg, problem)}
}

func (vm *ASTInterpreter) evalExpr(build msg.Builder, m msg.Msg, expr *ast.Expr, sink msg.Sink) error {
	defer trace()()

	if expr.Next != nil {
		oldSink := sink
		sink = func(m msg.Msg) error {
			return vm.evalExpr(build, m, expr.Next, oldSink)
		}

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

func (vm *ASTInterpreter) evalLiteral(build msg.Builder, m msg.Msg, l *ast.Literal, sink msg.Sink) error {
	defer trace()()
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

func (vm *ASTInterpreter) evalSelector(build msg.Builder, m msg.Msg, s *ast.Selector, sink msg.Sink) error {
	defer trace()()
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

func (vm *ASTInterpreter) evalMemberSelector(build msg.Builder, m msg.Msg, sel *ast.MemberSelector, sink msg.Sink) error {
	defer trace()()
	if sel.Child != nil {
		oldSink := sink
		sink = func(m msg.Msg) error { return vm.evalSelector(build, m, sel.Child, oldSink) }
	}

	// the meaning of an index depends on the type of message
	switch m.Type() {

	case msg.TypeObject:
		member, ok, err := vm.evalExprToMsgType(build, m, sel.Index, "index", msg.TypeString)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		msg, ok := m.Member(member.StringVal())
		if !ok {
			return nil
		}
		return sink(msg)

	case msg.TypeArray:
		pos, ok, err := vm.evalExprToMsgType(build, m, sel.Index, "index", msg.TypeInt)
		if err != nil {
			return err
		}
		if !ok {
			return nil
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

func (vm *ASTInterpreter) evalSliceSelector(build msg.Builder, m msg.Msg, s *ast.SliceSelector, sink msg.Sink) error {
	defer trace()()
	if s.Child != nil {
		oldSink := sink
		sink = func(m msg.Msg) error { return vm.evalSelector(build, m, s.Child, oldSink) }
	}

	if m.Type() != msg.TypeArray {
		return vm.skipEvalWrongType("slice", m.Type(), msg.TypeObject, msg.TypeArray)
	}

	var (
		n    = m.Len()
		from = int64(0)
		to   = n
	)

	if s.From != nil {
		fromMsg, ok, err := vm.evalExprToMsgType(build, m, s.From, "slice from", msg.TypeInt)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		from = fromMsg.IntVal()
	}
	if s.To != nil {
		toMsg, ok, err := vm.evalExprToMsgType(build, m, s.To, "slice to", msg.TypeInt)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		to = toMsg.IntVal()
	}

	if from >= n {
		return nil
	}
	if to > n {
		to = n
	}
	if to-from <= 0 {
		return nil
	}

	src := m.Slice(from, to)
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

func (vm *ASTInterpreter) evalOperator(build msg.Builder, m msg.Msg, o *ast.Operator, sink msg.Sink) error {
	defer trace()()

	// bool operators
	switch {
	case o.LogNot != nil:
		arg, ok, err := vm.evalExprToMsgType(build, m, o.LogNot.Arg, "logical not", msg.TypeBool)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		v, err := build.Bool(!arg.BoolVal())
		if err != nil {
			return err
		}
		return sink(v)

	case o.LogAnd != nil:
		lhs, ok, err := vm.evalExprToMsgType(build, m, o.LogAnd.LHS, "left of logical and", msg.TypeBool)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		rhs, ok, err := vm.evalExprToMsgType(build, m, o.LogAnd.RHS, "right of logical and", msg.TypeBool)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		v, err := build.Bool(lhs.BoolVal() && rhs.BoolVal())
		if err != nil {
			return err
		}
		return sink(v)

	case o.LogOr != nil:
		lhs, ok, err := vm.evalExprToMsgType(build, m, o.LogOr.LHS, "left of logical and", msg.TypeBool)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		rhs, ok, err := vm.evalExprToMsgType(build, m, o.LogOr.RHS, "right of logical and", msg.TypeBool)
		if err != nil {
			return err
		}
		if !ok {
			return nil
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
		lhs, ok, err := vm.evalExprToMsgType(build, m, o.NumAdd.LHS, "left of an addition", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		rhs, ok, err := vm.evalExprToMsgType(build, m, o.NumAdd.RHS, "right of an addition", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		if !ok {
			return nil
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
		lhs, ok, err := vm.evalExprToMsgType(build, m, o.NumSub.LHS, "left of a subtraction", msg.TypeInt, msg.TypeFloat)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		rhs, ok, err := vm.evalExprToMsgType(build, m, o.NumSub.RHS, "right of a subtraction", msg.TypeInt, msg.TypeFloat)
		if err != nil {
			return err
		}
		if !ok {
			return nil
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
		lhs, ok, err := vm.evalExprToMsgType(build, m, o.NumDiv.LHS, "left of a division", msg.TypeInt, msg.TypeFloat)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		rhs, ok, err := vm.evalExprToMsgType(build, m, o.NumDiv.RHS, "right of a division", msg.TypeInt, msg.TypeFloat)
		if err != nil {
			return err
		}
		if !ok {
			return nil
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
		lhs, ok, err := vm.evalExprToMsgType(build, m, o.NumMul.LHS, "left of a multiplication", msg.TypeInt, msg.TypeFloat)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		rhs, ok, err := vm.evalExprToMsgType(build, m, o.NumMul.RHS, "right of a multiplication", msg.TypeInt, msg.TypeFloat)
		if err != nil {
			return err
		}
		if !ok {
			return nil
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
		switch {
		case lhs.Type() == msg.TypeInt && rhs.Type() == msg.TypeFloat:
			return float64(lhs.IntVal()) == rhs.FloatVal()
		case lhs.Type() == msg.TypeFloat && rhs.Type() == msg.TypeInt:
			return lhs.FloatVal() == float64(rhs.IntVal())
		default:
			if lhs.Type() != rhs.Type() {
				return false
			}
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
				lhsV, lhsOk := lhs.Member(lhsKey)
				rhsV, rhsOk := rhs.Member(rhsKey)
				if lhsOk != rhsOk {
					return false
				}
				if !checkEq(lhsV, rhsV) {
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
		switch {
		case lhs.Type() == msg.TypeInt && rhs.Type() == msg.TypeFloat:
			return float64(lhs.IntVal()) < rhs.FloatVal(), nil
		case lhs.Type() == msg.TypeFloat && rhs.Type() == msg.TypeInt:
			return lhs.FloatVal() < float64(rhs.IntVal()), nil
		default:
			if lhs.Type() != rhs.Type() {
				return false, vm.skipEvalWrongType("comparison", rhs.Type(), lhs.Type())
			}

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
		lhs, ok, err := vm.evalExprToMsgType(build, m, o.CmpEq.LHS, "left of an equality", msg.TypeInt, msg.TypeFloat, msg.TypeBool, msg.TypeString, msg.TypeArray, msg.TypeObject)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		rhs, ok, err := vm.evalExprToMsgType(build, m, o.CmpEq.RHS, "right of an equality", msg.TypeInt, msg.TypeFloat, msg.TypeBool, msg.TypeString, msg.TypeArray, msg.TypeObject)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		v, err := build.Bool(checkEq(lhs, rhs))
		if err != nil {
			return err
		}
		return sink(v)

	case o.CmpNotEq != nil:
		lhs, ok, err := vm.evalExprToMsgType(build, m, o.CmpNotEq.LHS, "left of a non-equality", msg.TypeInt, msg.TypeFloat, msg.TypeBool, msg.TypeString, msg.TypeArray, msg.TypeObject)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		rhs, ok, err := vm.evalExprToMsgType(build, m, o.CmpNotEq.RHS, "right of a non-equality", msg.TypeInt, msg.TypeFloat, msg.TypeBool, msg.TypeString, msg.TypeArray, msg.TypeObject)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		v, err := build.Bool(!checkEq(lhs, rhs))
		if err != nil {
			return err
		}
		return sink(v)

	case o.CmpGt != nil:
		lhs, ok, err := vm.evalExprToMsgType(build, m, o.CmpGt.LHS, "left of a greater-than comparison", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		rhs, ok, err := vm.evalExprToMsgType(build, m, o.CmpGt.RHS, "right of a greater-than comparison", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		isLess, err := checkLess(lhs, rhs)
		if err != nil {
			return err
		}
		striclyGreater := !isLess && !checkEq(lhs, rhs)
		v, err := build.Bool(striclyGreater)
		if err != nil {
			return err
		}
		return sink(v)

	case o.CmpGtOrEq != nil:
		lhs, ok, err := vm.evalExprToMsgType(build, m, o.CmpGtOrEq.LHS, "left of a greater-than-or-equal comparison", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		rhs, ok, err := vm.evalExprToMsgType(build, m, o.CmpGtOrEq.RHS, "right of a greater-than-or-equal comparison", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		if !ok {
			return nil
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
		lhs, ok, err := vm.evalExprToMsgType(build, m, o.CmpLs.LHS, "left of a less-than comparison", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		rhs, ok, err := vm.evalExprToMsgType(build, m, o.CmpLs.RHS, "right of a less-than comparison", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		isLess, err := checkLess(lhs, rhs)
		if err != nil {
			return err
		}
		strictlyLess := isLess && !checkEq(lhs, rhs)
		v, err := build.Bool(strictlyLess)
		if err != nil {
			return err
		}
		return sink(v)

	case o.CmpLsOrEq != nil:
		lhs, ok, err := vm.evalExprToMsgType(build, m, o.CmpLsOrEq.LHS, "left of a less-than-or-equal comparison", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		rhs, ok, err := vm.evalExprToMsgType(build, m, o.CmpLsOrEq.RHS, "right of a less-than-or-equal comparison", msg.TypeInt, msg.TypeFloat, msg.TypeString)
		if err != nil {
			return err
		}
		if !ok {
			return nil
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

func (vm *ASTInterpreter) evalFuncCall(build msg.Builder, m msg.Msg, f *ast.FuncCall, sink msg.Sink) error {
	defer trace()()
	arities, fn := vm.lookupFuncs(f.Name)
	if fn == nil {
		return fmt.Errorf("unknown function %q", f.Name)
	}
	for _, arity := range arities {
		if arity == len(f.Args) {
			return fn(build, m, f.Args, sink)
		}
	}
	arityString := strconv.Itoa(arities[0])
	for i, arity := range arities[1:] {
		if i == len(arities)-1 {
			arityString += " or "
		} else {
			arityString += ", "
		}
		arityString += strconv.Itoa(arity)
	}
	return fmt.Errorf("function %q requires %s arguments, %d were given", f.Name, arityString, len(f.Args))
}

type evalFunc func(build msg.Builder, m msg.Msg, args []*ast.Expr, sink msg.Sink) error

func (vm *ASTInterpreter) lookupFuncs(name string) ([]int, evalFunc) {
	defer trace()()
	switch name {
	// not implicit unary func
	case "select":
		return []int{1, 2}, vm.evalFuncSelect

	// implicit unary func
	case "length":
		return []int{0, 1}, vm.evalFuncLength
	case "keys":
		return []int{0, 1}, vm.evalFuncKeys

		// not implicit binary func
	case "regexp":
		return []int{2}, vm.evalFuncRegexp
	case "contains":
		return []int{2}, vm.evalFuncContains

		// implicit binary func
	case "has":
		return []int{1, 2}, vm.evalFuncHas

	}
	return nil, nil
}

// == select(bool) -> msg.Msg ==
// Emits the current message if the given expression evaluates to true.
func (vm *ASTInterpreter) evalFuncSelect(build msg.Builder, m msg.Msg, args []*ast.Expr, sink msg.Sink) error {
	defer trace()()

	args, toemit, ok, err := vm.implicitArgOrEvalExpr(build, "function select", m, 1, args, msg.TypeString, msg.TypeBool, msg.TypeInt, msg.TypeFloat, msg.TypeNull, msg.TypeObject, msg.TypeArray)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	cond, ok, err := vm.evalExprToMsgType(build, m, args[0], "function select", msg.TypeBool)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if cond.BoolVal() {
		return sink(toemit)
	}
	return nil
}

// == length(string|object|array) -> int ==
// Emits an integer representing the length of a string, the number of elements in an array or the number of keys in an object.
func (vm *ASTInterpreter) evalFuncLength(build msg.Builder, m msg.Msg, args []*ast.Expr, sink msg.Sink) error {
	defer trace()()

	_, arg, ok, err := vm.implicitArgOrEvalExpr(build, "function length", m, 0, args, msg.TypeString, msg.TypeObject, msg.TypeArray)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	var l int64
	switch arg.Type() {
	case msg.TypeString:
		l = int64(len(arg.StringVal()))
	case msg.TypeObject:
		l = int64(len(arg.Keys()))
	case msg.TypeArray:
		l = arg.Len()
	}
	lmsg, err := build.Int(l)
	if err != nil {
		return err
	}
	return sink(lmsg)
}

// == keys(object|array) -> array ==
// Emits an array representing the keys an object, or the indices of an array.
func (vm *ASTInterpreter) evalFuncKeys(build msg.Builder, m msg.Msg, args []*ast.Expr, sink msg.Sink) error {
	defer trace()()

	_, arg, ok, err := vm.implicitArgOrEvalExpr(build, "function keys", m, 0, args, msg.TypeObject, msg.TypeArray)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	var abFunc func(msg.ArrayBuilder) error

	switch arg.Type() {
	case msg.TypeObject:
		abFunc = func(ab msg.ArrayBuilder) error {
			keys := make([]string, len(arg.Keys()))
			copy(keys, arg.Keys())
			sort.Strings(keys)
			for _, k := range keys {
				err := ab.AddElem(func(b msg.Builder) (msg.Msg, error) {
					return b.String(k)
				})
				if err != nil {
					return err
				}
			}
			return nil
		}

	case msg.TypeArray:
		abFunc = func(ab msg.ArrayBuilder) error {
			for i := int64(0); i < arg.Len(); i++ {
				err := ab.AddElem(func(b msg.Builder) (msg.Msg, error) {
					return b.Int(i)
				})
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	keys, err := build.Array(abFunc)
	if err != nil {
		return err
	}
	return sink(keys)
}

// == has(object, string) -> bool ==
// Emits a bool representing whether the object has a key.
func (vm *ASTInterpreter) evalFuncHas(build msg.Builder, m msg.Msg, args []*ast.Expr, sink msg.Sink) error {
	defer trace()()

	args, obj, ok, err := vm.implicitArgOrEvalExpr(build, "function has", m, 1, args, msg.TypeObject, msg.TypeArray)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	key, ok, err := vm.evalExprToMsgType(build, m, args[0], "function has", msg.TypeString)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	_, has := obj.Member(key.StringVal())
	hasmsg, err := build.Bool(has)
	if err != nil {
		return err
	}
	return sink(hasmsg)
}

// == regexp(s, pattern string) -> bool ==
// Emits a boolean: if the given regexp matches the expression.
func (vm *ASTInterpreter) evalFuncRegexp(build msg.Builder, m msg.Msg, args []*ast.Expr, sink msg.Sink) error {
	defer trace()()

	s, ok, err := vm.evalExprToMsgType(build, m, args[0], "function regexp", msg.TypeString)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	pattern, ok, err := vm.evalExprToMsgType(build, m, args[1], "function regexp", msg.TypeString)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	// figure a way to cache the compiled regexp.... if the argument is a compile time literal, for instance
	re, err := regexp.Compile(pattern.StringVal())
	if err != nil {
		return vm.skipEvalWrongArgValue("function regexp", pattern.Type(), "invalid regexp: "+err.Error())
	}
	match, err := build.Bool(re.MatchString(s.StringVal()))
	if err != nil {
		return err
	}
	return sink(match)
}

// == contains(s, substring string) -> bool ==
// Emits a boolean: if the given substring is found in the expression.
func (vm *ASTInterpreter) evalFuncContains(build msg.Builder, m msg.Msg, args []*ast.Expr, sink msg.Sink) error {
	defer trace()()

	s, ok, err := vm.evalExprToMsgType(build, m, args[0], "function contains", msg.TypeString)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	substring, ok, err := vm.evalExprToMsgType(build, m, args[1], "function contains", msg.TypeString)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	contains := strings.Contains(s.StringVal(), substring.StringVal())

	match, err := build.Bool(contains)
	if err != nil {
		return err
	}
	return sink(match)
}

// helper

// evalExprToMsg evaluates an expression's result and verifies that it is of the requested type.
func (vm *ASTInterpreter) evalExprToMsg(build msg.Builder, m msg.Msg, expr *ast.Expr) (msg.Msg, bool, error) {
	defer trace()()
	var (
		evaled msg.Msg
		found  bool
	)
	err := vm.evalExpr(build, m, expr, func(got msg.Msg) error {
		evaled = got
		found = true
		return nil

	})
	return evaled, found, err
}

// evalExprToMsgType evaluates an expression's result and verifies that it is of the requested type.
func (vm *ASTInterpreter) evalExprToMsgType(build msg.Builder, m msg.Msg, expr *ast.Expr, action string, want ...msg.Type) (msg.Msg, bool, error) {
	defer trace()()
	var (
		evaled msg.Msg
		found  bool
	)
	err := vm.evalExpr(build, m, expr, func(got msg.Msg) error {
		for _, w := range want {
			if got.Type() != w {
				continue
			}
			evaled = got
			found = true
			return nil
		}
		return vm.skipEvalWrongArgType(action, m.Type(), got.Type(), want...)

	})
	return evaled, found, err
}

// implicitArgOrEvalExpr uses an implicit argument (the current message context) if no expression is given. otherwise it evals the
// expression the usual way.
func (vm *ASTInterpreter) implicitArgOrEvalExpr(build msg.Builder, action string, m msg.Msg, implIfLen int, args []*ast.Expr, want ...msg.Type) ([]*ast.Expr, msg.Msg, bool, error) {
	if len(args) != implIfLen {
		m, ok, err := vm.evalExprToMsgType(build, m, args[0], action, want...)
		return args[1:], m, ok, err
	}
	for _, t := range want {
		if m.Type() == t {
			return args, m, true, nil
		}
	}
	return nil, nil, false, vm.skipEvalWrongType(action, m.Type(), msg.TypeString, msg.TypeObject, msg.TypeArray)
}
