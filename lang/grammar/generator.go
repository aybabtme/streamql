package grammar

import (
	"fmt"
	"strconv"

	"github.com/aybabtme/streamql/lang/ast"
)

func oneOfExpr(v interface{}) *ast.Expr {
	switch t := v.(type) {
	case *ast.Literal:
		return &ast.Expr{Literal: t}
	case *ast.Selector:
		return &ast.Expr{Selector: t}
	case *ast.Operator:
		return &ast.Expr{Operator: t}
	case *ast.FuncCall:
		return &ast.Expr{FuncCall: t}
	case *ast.Expr:
		return t
	default:
		panic(fmt.Sprintf("invalid expression: %T", t))
		return nil
	}
}

func oneOfSelector(v interface{}, child *ast.Selector) *ast.Selector {
	switch t := v.(type) {
	case *ast.NoopSelector:
		return &ast.Selector{Noop: t}
	case *ast.MemberSelector:
		if child != nil {
			t.Child = child
		}
		return &ast.Selector{Member: t}
	case *ast.SliceSelector:
		if child != nil {
			t.Child = child
		}
		return &ast.Selector{Slice: t}
	case *ast.Selector:
		return t
	default:
		panic(fmt.Sprintf("invalid expression for selection: %T", t))
		return nil
	}
}

func oneOfOperator(v interface{}) *ast.Operator {
	switch t := v.(type) {
	case *ast.OperandLogNot:
		return &ast.Operator{LogNot: t}
	case *ast.OperandLogAnd:
		return &ast.Operator{LogAnd: t}
	case *ast.OperandLogOr:
		return &ast.Operator{LogOr: t}
	case *ast.OperandNumAdd:
		return &ast.Operator{NumAdd: t}
	case *ast.OperandNumSub:
		return &ast.Operator{NumSub: t}
	case *ast.OperandNumDiv:
		return &ast.Operator{NumDiv: t}
	case *ast.OperandNumMul:
		return &ast.Operator{NumMul: t}
	case *ast.OperandCmpEq:
		return &ast.Operator{CmpEq: t}
	case *ast.OperandCmpNotEq:
		return &ast.Operator{CmpNotEq: t}
	case *ast.OperandCmpGt:
		return &ast.Operator{CmpGt: t}
	case *ast.OperandCmpGtOrEq:
		return &ast.Operator{CmpGtOrEq: t}
	case *ast.OperandCmpLs:
		return &ast.Operator{CmpLs: t}
	case *ast.OperandCmpLsOrEq:
		return &ast.Operator{CmpLsOrEq: t}
	default:
		panic(fmt.Sprintf("invalid expression for operator: %T", t))
		return nil
	}
}

func expr(sym yySymType) *ast.Expr {
	return oneOfExpr(sym.node)
}

func literal(sym yySymType) yySymType {
	switch t := sym.node.(type) {
	case *bool:
		return yySymType{node: &ast.Literal{Bool: t}}
	case *string:
		return yySymType{node: &ast.Literal{String: t}}
	case *int64:
		return yySymType{node: &ast.Literal{Int: t}}
	case *float64:
		return yySymType{node: &ast.Literal{Float: t}}
	case *struct{}:
		return yySymType{node: &ast.Literal{Null: t}}
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
	return yySymType{node: &ast.NoopSelector{}}
}

func emitMemberSelector(indexSym, subSelSym yySymType) yySymType {
	var (
		index *ast.Expr
		child *ast.Selector
	)
	switch indexSym.curID {
	case Identifier:
		index = &ast.Expr{Literal: &ast.Literal{String: &indexSym.cur.lit}}
	default:
		index = expr(indexSym)
	}
	if subSelSym.node != nil {
		child = oneOfSelector(subSelSym.node, nil)
	}
	return yySymType{node: &ast.MemberSelector{Index: index, Child: child}}
}

func emitSliceSelectorEach(subSelSym yySymType) yySymType {
	var child *ast.Selector
	if subSelSym.node != nil {
		child = oneOfSelector(subSelSym.node, nil)
	}
	return yySymType{node: &ast.SliceSelector{Child: child}}
}

func emitSliceSelector(fromSym, toSym yySymType, subSelSym yySymType) yySymType {
	var (
		from  *ast.Expr
		to    *ast.Expr
		child *ast.Selector
	)
	switch fromSym.curID {
	case Int:
		fromVal, err := strconv.ParseInt(fromSym.cur.lit, 10, 64)
		if err != nil {
			panic(err)
		}
		from = &ast.Expr{Literal: &ast.Literal{Int: &fromVal}}
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
		to = &ast.Expr{Literal: &ast.Literal{Int: &toVal}}
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
	return yySymType{node: &ast.SliceSelector{From: from, To: to, Child: child}}
}

func emitOpNot(arg yySymType) yySymType {
	return yySymType{node: &ast.OperandLogNot{Arg: expr(arg)}}
}
func emitOpAnd(lhs, rhs yySymType) yySymType {
	return yySymType{node: &ast.OperandLogAnd{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpOr(lhs, rhs yySymType) yySymType {
	return yySymType{node: &ast.OperandLogOr{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpNeg(arg yySymType) yySymType {
	z := int64(0)
	zero := &ast.Expr{Literal: &ast.Literal{Int: &z}}
	return yySymType{node: &ast.OperandNumSub{LHS: zero, RHS: expr(arg)}}
}
func emitOpAdd(lhs, rhs yySymType) yySymType {
	return yySymType{node: &ast.OperandNumAdd{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpSub(lhs, rhs yySymType) yySymType {
	return yySymType{node: &ast.OperandNumSub{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpDiv(lhs, rhs yySymType) yySymType {
	return yySymType{node: &ast.OperandNumDiv{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpMul(lhs, rhs yySymType) yySymType {
	return yySymType{node: &ast.OperandNumMul{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpEq(lhs, rhs yySymType) yySymType {
	return yySymType{node: &ast.OperandCmpEq{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpNotEq(lhs, rhs yySymType) yySymType {
	return yySymType{node: &ast.OperandCmpNotEq{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpGt(lhs, rhs yySymType) yySymType {
	return yySymType{node: &ast.OperandCmpGt{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpGtOrEq(lhs, rhs yySymType) yySymType {
	return yySymType{node: &ast.OperandCmpGtOrEq{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpLs(lhs, rhs yySymType) yySymType {
	return yySymType{node: &ast.OperandCmpLs{LHS: expr(lhs), RHS: expr(rhs)}}
}
func emitOpLsOrEq(lhs, rhs yySymType) yySymType {
	return yySymType{node: &ast.OperandCmpLsOrEq{LHS: expr(lhs), RHS: expr(rhs)}}
}

func emitFuncCall(arg0, arg1 yySymType) yySymType {
	var (
		name string
		args []*ast.Expr
	)
	if arg0.curID != Identifier {
		panic(fmt.Sprintf("invalid function name: %v", arg0.curID))
	} else {
		name = arg0.cur.lit
	}
	switch t := arg1.node.(type) {
	case []*ast.Expr:
		args = t
	case *ast.Expr:
		args = []*ast.Expr{t}
	case nil:
	default:
		panic(fmt.Sprintf("invalid function arguments: %#v", t))
	}

	return yySymType{node: &ast.FuncCall{Name: name, Args: args}}
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
	return yySymType{node: &ast.FuncCall{Name: name}}
}

func emitArg(arg0 yySymType) yySymType {
	var expr = expr(arg0)
	return yySymType{node: expr}
}

func emitArgs(arg0, arg1 yySymType) yySymType {
	var (
		expr = expr(arg0)
		prev []*ast.Expr
	)
	switch t := arg1.node.(type) {
	case []*ast.Expr:
		prev = t
	case *ast.Expr:
		prev = []*ast.Expr{t}
	case nil:
	default:
		panic(fmt.Sprintf("invalid function argument, %T", t))
	}
	return yySymType{node: append([]*ast.Expr{expr}, prev...)}
}
