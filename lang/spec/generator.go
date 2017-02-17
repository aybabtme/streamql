package spec

import (
	"fmt"
	"strconv"
)

func oneOfExpr(v interface{}) *Expr {
	switch t := v.(type) {
	case *Literal:
		return &Expr{Literal: t}
	case *Selector:
		return &Expr{Selector: t}
	case *Operator:
		return &Expr{Operator: t}
	case *FuncCall:
		return &Expr{FuncCall: t}
	case *Expr:
		return t
	default:
		panic(fmt.Sprintf("invalid expression: %T", t))
		return nil
	}
}

func oneOfSelector(v interface{}, child *Selector) *Selector {
	switch t := v.(type) {
	case *NoopSelector:
		return &Selector{Noop: t}
	case *MemberSelector:
		if child != nil {
			t.Child = child
		}
		return &Selector{Member: t}
	case *SliceSelector:
		if child != nil {
			t.Child = child
		}
		return &Selector{Slice: t}
	case *Selector:
		return t
	default:
		panic(fmt.Sprintf("invalid expression for selection: %T", t))
		return nil
	}
}

func oneOfOperator(v interface{}) *Operator {
	switch t := v.(type) {
	case *OperandLogNot:
		return &Operator{LogNot: t}
	case *OperandLogAnd:
		return &Operator{LogAnd: t}
	case *OperandLogOr:
		return &Operator{LogOr: t}
	case *OperandNumAdd:
		return &Operator{NumAdd: t}
	case *OperandNumSub:
		return &Operator{NumSub: t}
	case *OperandNumDiv:
		return &Operator{NumDiv: t}
	case *OperandNumMul:
		return &Operator{NumMul: t}
	case *OperandCmpEq:
		return &Operator{CmpEq: t}
	case *OperandCmpNotEq:
		return &Operator{CmpNotEq: t}
	case *OperandCmpGt:
		return &Operator{CmpGt: t}
	case *OperandCmpGtOrEq:
		return &Operator{CmpGtOrEq: t}
	case *OperandCmpLs:
		return &Operator{CmpLs: t}
	case *OperandCmpLsOrEq:
		return &Operator{CmpLsOrEq: t}
	default:
		panic(fmt.Sprintf("invalid expression for operator: %T", t))
		return nil
	}
}

func expr(sym yySymType) *Expr {
	return oneOfExpr(sym.node)
}

func literal(sym yySymType) yySymType {
	switch t := sym.node.(type) {
	case *bool:
		return yySymType{node: &Literal{Bool: t}}
	case *string:
		return yySymType{node: &Literal{String: t}}
	case *int64:
		return yySymType{node: &Literal{Int: t}}
	case *float64:
		return yySymType{node: &Literal{Float: t}}
	case *struct{}:
		return yySymType{node: &Literal{Null: t}}
	default:
		panic("invalid literal")
	}
}

func selector(sym yySymType) yySymType {
	return yySymType{node: oneOfSelector(sym.node, nil)}
}

func operator(sym yySymType) yySymType {
	return yySymType{node: oneOfOperator(sym.node)}
}

func funcCall(sym yySymType) yySymType {
	return sym
}

func pipe(lhs, rhs yySymType) yySymType {
	lhsExpr, rhsExpr := oneOfExpr(lhs.node), oneOfExpr(rhs.node)
	lhsExpr.Next = rhsExpr
	return yySymType{node: lhsExpr}
}

func emitBool(arg0 yySymType) yySymType {
	v, err := strconv.ParseBool(arg0.cur.lit)
	if err != nil {
		panic(err)
	}
	return yySymType{node: &v}
}

func emitString(arg0 yySymType) yySymType {
	v, err := strconv.Unquote(arg0.cur.lit)
	if err != nil {
		panic(err)
	}
	return yySymType{node: &v}
}

func emitInt(arg0 yySymType) yySymType {
	v, err := strconv.ParseInt(arg0.cur.lit, 10, 64)
	if err != nil {
		panic(err)
	}
	return yySymType{node: &v}
}

func emitFloat(arg0 yySymType) yySymType {
	v, err := strconv.ParseFloat(arg0.cur.lit, 64)
	if err != nil {
		panic(err)
	}
	return yySymType{node: &v}
}

func emitNull(arg0 yySymType) yySymType {
	switch arg0.cur.lit {
	case "null":
	default:
		panic(fmt.Sprintf("invalid literal for a null value: %q", arg0.cur.lit))
	}
	return yySymType{node: new(struct{})}
}

func emitNopSelector() yySymType {
	return yySymType{node: &NoopSelector{}}
}

func emitMemberSelector(indexSym, subSelSym yySymType) yySymType {
	var (
		index *Expr
		child *Selector
	)
	switch indexSym.curID {
	case Identifier:
		index = &Expr{Literal: &Literal{String: &indexSym.cur.lit}}
	default:
		index = expr(indexSym)
	}
	if subSelSym.node != nil {
		child = oneOfSelector(subSelSym.node, nil)
	}
	return yySymType{node: &MemberSelector{Index: index, Child: child}}
}

func emitSliceSelectorEach(subSelSym yySymType) yySymType {
	var child *Selector
	if subSelSym.node != nil {
		child = oneOfSelector(subSelSym.node, nil)
	}
	return yySymType{node: &SliceSelector{Child: child}}
}

func emitSliceSelector(fromSym, toSym yySymType, subSelSym yySymType) yySymType {
	var (
		from  *Expr
		to    *Expr
		child *Selector
	)
	switch fromSym.curID {
	case Int:
		fromVal, err := strconv.ParseInt(fromSym.cur.lit, 10, 64)
		if err != nil {
			panic(err)
		}
		from = &Expr{Literal: &Literal{Int: &fromVal}}
	default:
		if fromSym.node == implicitSliceIdx {
			from = nil
		} else {
			from = expr(fromSym)
		}
	}
	switch toSym.curID {
	case Int:
		toVal, err := strconv.ParseInt(toSym.cur.lit, 10, 64)
		if err != nil {
			panic(err)
		}
		to = &Expr{Literal: &Literal{Int: &toVal}}
	default:
		if toSym.node == implicitSliceIdx {
			to = nil
		} else {
			to = expr(toSym)
		}
	}

	if subSelSym.node != nil {
		child = oneOfSelector(subSelSym.node, nil)
	}
	return yySymType{node: &SliceSelector{From: from, To: to, Child: child}}
}

func emitOpNot(arg yySymType) yySymType {
	return yySymType{node: &OperandLogNot{Arg: expr(arg)}}
}
func emitOpAnd(lhs, rhs yySymType) yySymType {
	return yySymType{node: &OperandLogAnd{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpOr(lhs, rhs yySymType) yySymType {
	return yySymType{node: &OperandLogOr{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpNeg(arg yySymType) yySymType {
	z := int64(0)
	zero := &Expr{Literal: &Literal{Int: &z}}
	return yySymType{node: &OperandNumSub{LHS: zero, RHS: expr(arg)}}
}
func emitOpAdd(lhs, rhs yySymType) yySymType {
	return yySymType{node: &OperandNumAdd{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpSub(lhs, rhs yySymType) yySymType {
	return yySymType{node: &OperandNumSub{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpDiv(lhs, rhs yySymType) yySymType {
	return yySymType{node: &OperandNumDiv{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpMul(lhs, rhs yySymType) yySymType {
	return yySymType{node: &OperandNumMul{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpEq(lhs, rhs yySymType) yySymType {
	return yySymType{node: &OperandCmpEq{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpNotEq(lhs, rhs yySymType) yySymType {
	return yySymType{node: &OperandCmpNotEq{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpGt(lhs, rhs yySymType) yySymType {
	return yySymType{node: &OperandCmpGt{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpGtOrEq(lhs, rhs yySymType) yySymType {
	return yySymType{node: &OperandCmpGtOrEq{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpLs(lhs, rhs yySymType) yySymType {
	return yySymType{node: &OperandCmpLs{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpLsOrEq(lhs, rhs yySymType) yySymType {
	return yySymType{node: &OperandCmpLsOrEq{LHS: expr(lhs), RHS: expr(rhs)}}
}

func emitFuncCall(arg0, arg1 yySymType) yySymType {
	var (
		name string
		args []*Expr
	)
	if arg0.curID != Identifier {
		panic(fmt.Sprintf("invalid function name: %v", arg0.curID))
	} else {
		name = arg0.cur.lit
	}
	switch t := arg1.node.(type) {
	case []*Expr:
		args = t
	case *Expr:
		args = []*Expr{t}
	case nil:
	default:
		panic(fmt.Sprintf("invalid function arguments: %#v", t))
	}

	return yySymType{node: &FuncCall{Name: name, Args: args}}
}

func emitImplicitFuncCall(arg0 yySymType) yySymType {
	var (
		name string
	)
	if arg0.curID != Identifier {
		panic(fmt.Sprintf("invalid function name: %v", arg0.curID))
	} else {
		name = arg0.cur.lit
	}
	return yySymType{node: &FuncCall{Name: name}}
}

func emitArg(arg0 yySymType) yySymType {
	var expr = expr(arg0)
	return yySymType{node: expr}
}

func emitArgs(arg0, arg1 yySymType) yySymType {
	var (
		expr = expr(arg0)
		prev []*Expr
	)
	switch t := arg1.node.(type) {
	case []*Expr:
		prev = t
	case *Expr:
		prev = []*Expr{t}
	case nil:
	default:
		panic(fmt.Sprintf("invalid function argument, %T", t))
	}
	return yySymType{node: append([]*Expr{expr}, prev...)}
}
