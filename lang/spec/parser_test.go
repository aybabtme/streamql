package spec

import (
	"encoding/json"
	"log"
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	// yyDebug = 3
	yyErrorVerbose = true

	var (
		ast       = func(expr *Expr) *AST { return &AST{Expr: expr} }
		pipe      = func(lhs, rhs *Expr) *Expr { lhs.Next = rhs; return lhs }
		exprSel   = func(s *Selector) *Expr { return &Expr{Selector: s} }
		exprLit   = func(s *Literal) *Expr { return &Expr{Literal: s} }
		exprOp    = func(s *Operator) *Expr { return &Expr{Operator: s} }
		exprFn    = func(s *FuncCall) *Expr { return &Expr{FuncCall: s} }
		selNoop   = func() *Selector { return &Selector{Noop: &NoopSelector{}} }
		selMember = func(expr *Expr, child *Selector) *Selector {
			return &Selector{Member: &MemberSelector{Index: expr, Child: child}}
		}
		selSlice = func(from, to *Expr, child *Selector) *Selector {
			return &Selector{Slice: &SliceSelector{From: from, To: to, Child: child}}
		}
		opNot = func(lhs *Expr) *Operator { return &Operator{LogNot: &OperandLogNot{Arg: lhs}} }
		opAnd = func(lhs, rhs *Expr) *Operator { return &Operator{LogAnd: &OperandLogAnd{LHS: lhs, RHS: rhs}} }
		opOr  = func(lhs, rhs *Expr) *Operator { return &Operator{LogOr: &OperandLogOr{LHS: lhs, RHS: rhs}} }
		opAdd = func(lhs, rhs *Expr) *Operator { return &Operator{NumAdd: &OperandNumAdd{LHS: lhs, RHS: rhs}} }
		opSub = func(lhs, rhs *Expr) *Operator { return &Operator{NumSub: &OperandNumSub{LHS: lhs, RHS: rhs}} }
		opMul = func(lhs, rhs *Expr) *Operator { return &Operator{NumMul: &OperandNumMul{LHS: lhs, RHS: rhs}} }
		opDiv = func(lhs, rhs *Expr) *Operator { return &Operator{NumDiv: &OperandNumDiv{LHS: lhs, RHS: rhs}} }
		opEq  = func(lhs, rhs *Expr) *Operator { return &Operator{CmpEq: &OperandCmpEq{LHS: lhs, RHS: rhs}} }
		opGt  = func(lhs, rhs *Expr) *Operator { return &Operator{CmpGt: &OperandCmpGt{LHS: lhs, RHS: rhs}} }

		fn = func(name string, args ...*Expr) *FuncCall { return &FuncCall{Name: name, Args: args} }

		litBool   = func(v bool) *Literal { return &Literal{Bool: &v} }
		litString = func(v string) *Literal { return &Literal{String: &v} }
		litInt    = func(v int64) *Literal { return &Literal{Int: &v} }
		litFloat  = func(v float64) *Literal { return &Literal{Float: &v} }
		litNull   = func() *Literal { return &Literal{Null: &struct{}{}} }

		_ = ast
		_ = pipe
		_ = exprSel
		_ = exprLit
		_ = exprOp
		_ = exprFn
		_ = selNoop
		_ = selMember
		_ = selSlice
		_ = opNot
		_ = opAnd
		_ = opOr
		_ = opAdd
		_ = opSub
		_ = opMul
		_ = opDiv
		_ = opEq
		_ = opGt
		_ = fn
		_ = litBool
		_ = litString
		_ = litInt
		_ = litFloat
		_ = litNull
	)

	tests := []struct {
		name    string
		args    string
		want    *AST
		wantErr bool
	}{

		{args: "", want: ast(nil)},
		{args: ".", want: ast(exprSel(selNoop()))},
		{args: ". | .", want: ast(pipe(exprSel(selNoop()), exprSel(selNoop())))},
		{args: ". | . | .", want: ast(
			pipe(
				exprSel(selNoop()),
				pipe(
					exprSel(selNoop()),
					exprSel(selNoop()),
				),
			),
		)},
		{args: ". | . | . | .", want: ast(
			pipe(
				exprSel(selNoop()),
				pipe(
					exprSel(selNoop()),
					pipe(
						exprSel(selNoop()),
						exprSel(selNoop()),
					),
				),
			),
		)},
		{args: "true", want: ast(exprLit(litBool(true)))},
		{args: `""`, want: ast(exprLit(litString("")))},
		{args: `1`, want: ast(exprLit(litInt(1)))},
		{args: `1.0`, want: ast(exprLit(litFloat(1)))},
		{args: `null`, want: ast(exprLit(litNull()))},
		{args: ".hello", want: ast(exprSel(
			selMember(exprLit(litString("hello")), nil),
		))},
		{args: ".hello.bye", want: ast(exprSel(
			selMember(exprLit(litString("hello")),
				selMember(exprLit(litString("bye")), nil),
			),
		))},
		{args: ".[]", want: ast(
			exprSel(
				selSlice(nil, nil, nil),
			),
		)},
		{args: ".[1]", want: ast(
			exprSel(
				selMember(exprLit(litInt(1)), nil),
			),
		)},

		{args: ".[1:2]", want: ast(
			exprSel(
				selSlice(
					exprLit(litInt(1)),
					exprLit(litInt(2)),
					nil,
				),
			),
		)},
		{args: ".hello[]", want: ast(
			exprSel(
				selMember(
					exprLit(litString("hello")),
					selSlice(nil, nil, nil),
				),
			),
		)},
		{args: ".hello[1]", want: ast(
			exprSel(
				selMember(
					exprLit(litString("hello")),
					selMember(exprLit(litInt(1)), nil),
				),
			),
		)},
		{args: ".hello[1:2]", want: ast(
			exprSel(
				selMember(
					exprLit(litString("hello")),
					selSlice(
						exprLit(litInt(1)),
						exprLit(litInt(2)),
						nil,
					),
				),
			),
		)},
		{args: ".[][]", want: ast(
			exprSel(
				selSlice(nil, nil,
					selSlice(nil, nil, nil),
				),
			),
		)},
		{args: ".[1][1]", want: ast(
			exprSel(
				selMember(
					exprLit(litInt(1)),
					selMember(exprLit(litInt(1)), nil),
				),
			),
		)},
		{args: ".[1:2][1:2]", want: ast(
			exprSel(
				selSlice(
					exprLit(litInt(1)),
					exprLit(litInt(2)),
					selSlice(
						exprLit(litInt(1)),
						exprLit(litInt(2)),
						nil,
					),
				),
			),
		)},

		{args: ".hello[].bye", want: ast(
			exprSel(
				selMember(
					exprLit(litString("hello")),
					selSlice(nil, nil,
						selMember(exprLit(litString("bye")), nil),
					),
				),
			),
		)},
		{args: ".hello[1].bye", want: ast(
			exprSel(
				selMember(
					exprLit(litString("hello")),
					selMember(exprLit(litInt(1)),
						selMember(exprLit(litString("bye")), nil),
					),
				),
			),
		)},
		{args: ".hello[1:2].bye", want: ast(
			exprSel(
				selMember(
					exprLit(litString("hello")),
					selSlice(
						exprLit(litInt(1)),
						exprLit(litInt(2)),
						selMember(exprLit(litString("bye")), nil),
					),
				),
			),
		)},

		{args: `!.`, want: ast(
			exprOp(opNot(exprSel(selNoop()))),
		)},
		{args: `. && .`, want: ast(
			exprOp(opAnd(
				exprSel(selNoop()),
				exprSel(selNoop()),
			)),
		)},
		{args: `. || .`, want: ast(
			exprOp(opOr(
				exprSel(selNoop()),
				exprSel(selNoop()),
			)),
		)},
		{args: `. && . && .`, want: ast(
			exprOp(opAnd(
				exprOp(opAnd(
					exprSel(selNoop()),
					exprSel(selNoop()),
				)),
				exprSel(selNoop()),
			)),
		)},
		{args: `. || . || .`, want: ast(
			exprOp(opOr(
				exprOp(opOr(
					exprSel(selNoop()),
					exprSel(selNoop()),
				)),
				exprSel(selNoop()),
			)),
		)},
		{args: `!. || !. || !.`, want: ast(
			exprOp(opOr(
				exprOp(opOr(
					exprOp(opNot(exprSel(selNoop()))),
					exprOp(opNot(exprSel(selNoop()))),
				)),
				exprOp(opNot(exprSel(selNoop()))),
			)),
		)},
		{args: `!(. || .) || !.`, want: ast(
			exprOp(opOr(
				exprOp(opNot(
					exprOp(opOr(
						exprSel(selNoop()),
						exprSel(selNoop()),
					)),
				)),
				exprOp(opNot(exprSel(selNoop()))),
			)),
		)},
		{args: `!. || !(. || .)`, want: ast(
			exprOp(opOr(
				exprOp(opNot(exprSel(selNoop()))),
				exprOp(opNot(
					exprOp(opOr(
						exprSel(selNoop()),
						exprSel(selNoop()),
					)),
				)),
			)),
		)},

		{args: `. + .`, want: ast(
			exprOp(opAdd(
				exprSel(selNoop()),
				exprSel(selNoop()),
			)),
		)},
		{args: `. - .`, want: ast(
			exprOp(opSub(
				exprSel(selNoop()),
				exprSel(selNoop()),
			)),
		)},
		{args: `. / .`, want: ast(
			exprOp(opDiv(
				exprSel(selNoop()),
				exprSel(selNoop()),
			)),
		)},
		{args: `. * .`, want: ast(
			exprOp(opMul(
				exprSel(selNoop()),
				exprSel(selNoop()),
			)),
		)},
		{args: `. + . + .`, want: ast(
			exprOp(opAdd(
				exprOp(opAdd(
					exprSel(selNoop()),
					exprSel(selNoop()),
				)),
				exprSel(selNoop()),
			)),
		)},
		{args: `. - . - .`, want: ast(
			exprOp(opSub(
				exprOp(opSub(
					exprSel(selNoop()),
					exprSel(selNoop()),
				)),
				exprSel(selNoop()),
			)),
		)},
		{args: `. / . / .`, want: ast(
			exprOp(opDiv(
				exprOp(opDiv(
					exprSel(selNoop()),
					exprSel(selNoop()),
				)),
				exprSel(selNoop()),
			)),
		)},
		{args: `. * . * .`, want: ast(
			exprOp(opMul(
				exprOp(opMul(
					exprSel(selNoop()),
					exprSel(selNoop()),
				)),
				exprSel(selNoop()),
			)),
		)},

		{args: `1 + 2 - 3`, want: ast(
			exprOp(opAdd(
				exprLit(litInt(1)),
				exprOp(opSub(
					exprLit(litInt(2)),
					exprLit(litInt(3)),
				)),
			)),
		)},
		{args: `1 - 2 + 3`, want: ast(
			exprOp(opAdd(
				exprOp(opSub(
					exprLit(litInt(1)),
					exprLit(litInt(2)),
				)),
				exprLit(litInt(3)),
			)),
		)},
		{args: `1 * 2 / 3`, want: ast(
			exprOp(opMul(
				exprLit(litInt(1)),
				exprOp(opDiv(
					exprLit(litInt(2)),
					exprLit(litInt(3)),
				)),
			)),
		)},
		{args: `1 / 2 * 3`, want: ast(
			exprOp(opMul(
				exprOp(opDiv(
					exprLit(litInt(1)),
					exprLit(litInt(2)),
				)),
				exprLit(litInt(3)),
			)),
		)},

		{args: `1 - (2 + 3)`, want: ast(
			exprOp(opSub(
				exprLit(litInt(1)),
				exprOp(opAdd(
					exprLit(litInt(2)),
					exprLit(litInt(3)),
				)),
			)),
		)},
		{args: `1 / (2 * 3)`, want: ast(
			exprOp(opDiv(
				exprLit(litInt(1)),
				exprOp(opMul(
					exprLit(litInt(2)),
					exprLit(litInt(3)),
				)),
			)),
		)},

		{args: "select(true)", want: ast(
			exprFn(fn("select", exprLit(litBool(true)))),
		)},
		{args: "select(true && true)", want: ast(
			exprFn(fn("select",
				exprOp(opAnd(
					exprLit(litBool(true)),
					exprLit(litBool(true)),
				)),
			)),
		)},
		{args: "select(.)", want: ast(
			exprFn(fn("select", exprSel(selNoop()))),
		)},
		{args: "select(.lol)", want: ast(
			exprFn(fn("select", exprSel(
				selMember(exprLit(litString("lol")), nil),
			))),
		)},
		{args: "select(bool(.lol) && true)", want: ast(
			exprFn(fn("select", exprOp(
				opAnd(
					exprFn(fn("bool", exprSel(
						selMember(exprLit(litString("lol")), nil),
					))),
					exprLit(litBool(true)),
				),
			))),
		)},
		{args: "select(.lol && true)", want: ast(
			exprFn(fn("select", exprOp(
				opAnd(
					exprSel(selMember(exprLit(litString("lol")), nil)),
					exprLit(litBool(true)),
				),
			))),
		)},

		{args: `1+1>2 && 2+2 == 4 || 1+1>2 && !(2+2 == 4)`, want: ast(
			exprOp(opOr(
				// 1+1>2 && 2+2 == 4
				exprOp(opAnd(
					exprOp(opGt(
						exprOp(opAdd(
							exprLit(litInt(1)),
							exprLit(litInt(1)),
						)),
						exprLit(litInt(2)),
					)),
					exprOp(opEq(
						exprOp(opAdd(
							exprLit(litInt(2)),
							exprLit(litInt(2)),
						)),
						exprLit(litInt(4)),
					)),
				)),
				// 1+1>2 && !2+2 == 4
				exprOp(opAnd(
					exprOp(opGt(
						exprOp(opAdd(
							exprLit(litInt(1)),
							exprLit(litInt(1)),
						)),
						exprLit(litInt(2)),
					)),
					exprOp(opNot(
						exprOp(opEq(
							exprOp(opAdd(
								exprLit(litInt(2)),
								exprLit(litInt(2)),
							)),
							exprLit(litInt(4)),
						)),
					)),
				)),
			)),
		)},

		{args: `.lol[0:1] | select(.is_red && string(.size) == "large")`, want: ast(
			pipe(
				exprSel(selMember(
					exprLit(litString("lol")),
					selSlice(
						exprLit(litInt(0)),
						exprLit(litInt(1)),
						nil,
					),
				)),
				exprFn(fn(
					"select",
					exprOp(opAnd(
						exprSel(selMember(exprLit(litString("is_red")), nil)),
						exprOp(opEq(
							exprFn(fn(
								"string",
								exprSel(selMember(exprLit(litString("size")), nil)),
							)),
							exprLit(litString("large")),
						)),
					)),
				)),
			),
		)},
		{args: `.lol[0:1] | select(.is_red && string(.size) == "large") | select(.)`, want: ast(
			pipe(
				exprSel(selMember(
					exprLit(litString("lol")),
					selSlice(
						exprLit(litInt(0)),
						exprLit(litInt(1)),
						nil,
					),
				)),
				pipe(
					exprFn(fn(
						"select",
						exprOp(opAnd(
							exprSel(selMember(exprLit(litString("is_red")), nil)),
							exprOp(opEq(
								exprFn(fn(
									"string",
									exprSel(selMember(exprLit(litString("size")), nil)),
								)),
								exprLit(litString("large")),
							)),
						)),
					)),
					exprFn(fn(
						"select",
						exprSel(selNoop()),
					),
					),
				),
			),
		)},

		{args: "select(.)", want: ast(
			exprFn(fn("select", exprSel(
				selNoop(),
			))),
		)},
		{args: "select(.size > 1)", want: ast(
			exprFn(fn("select", exprOp(
				opGt(
					exprSel(selMember(exprLit(litString("size")), nil)),
					exprLit(litInt(1)),
				),
			))),
		)},
		{args: "select(.size > 1) / 1000", want: ast(
			exprOp(opDiv(
				exprFn(fn("select", exprOp(
					opGt(
						exprSel(selMember(exprLit(litString("size")), nil)),
						exprLit(litInt(1)),
					),
				))),
				exprLit(litInt(1000)),
			)),
		)},
		{args: "reduce(select(.size > 1) / 1000)", want: ast(
			exprFn(fn(
				"reduce",
				exprOp(opDiv(
					exprFn(fn("select", exprOp(
						opGt(
							exprSel(selMember(exprLit(litString("size")), nil)),
							exprLit(litInt(1)),
						),
					))),
					exprLit(litInt(1000)),
				)),
			)),
		)},
		{args: "reduce(select(.size > 1) | . / 1000)", want: ast(
			exprFn(fn(
				"reduce",
				pipe(
					exprFn(fn("select", exprOp(
						opGt(
							exprSel(selMember(exprLit(litString("size")), nil)),
							exprLit(litInt(1)),
						),
					))),
					exprOp(opDiv(
						exprSel(selNoop()),
						exprLit(litInt(1000)),
					)),
				),
			)),
		)},
		{args: "select(.cond.keep) | .name", want: ast(
			pipe(
				exprFn(fn(
					"select",
					exprSel(selMember(
						exprLit(litString("cond")),
						selMember(
							exprLit(litString("keep")),
							nil,
						),
					)),
				)),
				exprSel(selMember(
					exprLit(litString("name")),
					nil,
				)),
			),
		)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("args=%s", tt.args)
			got, err := Parse(strings.NewReader(tt.args))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				wantAst, _ := json.MarshalIndent(tt.want, "", "  ")
				gotAst, _ := json.MarshalIndent(got, "", "  ")
				t.Errorf("want=%s", string(wantAst))
				t.Errorf("got=%s", string(gotAst))
			}
		})
	}
}

func TestParseOnly(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	// yyDebug = 3
	yyErrorVerbose = true

	tests := []struct {
		name    string
		args    string
		wantErr bool
	}{

		{args: `bool(.)`},
		{args: `bool("true")`},
		{args: `bool("false")`},
		{args: `true && true`},
		{args: `true || true`},
		{args: `true && true && false`},
		{args: `true || true || false`},
		{args: `true && true || false`},
		{args: `true || true && false`},
		{args: `true && bool(.)`},
		{args: `true || bool(.)`},
		{args: `true && .`},
		{args: `true || .`},
		{args: `bool(.) && bool(.)`},
		{args: `bool(.) || bool(.)`},
		{args: `bool(.) && .`},
		{args: `bool(.) || .`},
		{args: `bool(.) && bool(.) && bool(.)`},
		{args: `bool(.) || bool(.) || bool(.)`},
		{args: `(true && true)`},
		{args: `true && (true || false)`},
		{args: `(true && true) || false`},

		{args: `int(.)`},
		{args: `int(4.2)`},
		{args: `int("42")`},
		{args: `42 + 42`},
		{args: `42 - 42`},
		{args: `42 / 42`},
		{args: `42 * 42`},
		{args: `42 + .`},
		{args: `42 - .`},
		{args: `42 / .`},
		{args: `42 * .`},
		{args: `. + 42`},
		{args: `. - 42`},
		{args: `. / 42`},
		{args: `. * 42`},
		{args: `. + 42 + 42`},
		{args: `. - 42 - 42`},
		{args: `. / 42 / 42`},
		{args: `. * 42 * 42`},
		{args: `. + 42 - 42`},
		{args: `. + 42 - 42`},
		{args: `. * 42 / 42`},
		{args: `. / 42 * 42`},
		{args: `42 + . + 42`},
		{args: `42 - . - 42`},
		{args: `42 / . / 42`},
		{args: `42 * . * 42`},
		{args: `42 + . - 42`},
		{args: `42 + . - 42`},
		{args: `42 * . / 42`},
		{args: `42 / . * 42`},
		{args: `42 + 42 + .`},
		{args: `42 - 42 - .`},
		{args: `42 / 42 / .`},
		{args: `42 * 42 * .`},
		{args: `42 + 42 - .`},
		{args: `42 + 42 - .`},
		{args: `42 * 42 / .`},
		{args: `42 / 42 * .`},
		{args: `42 + int(.)`},
		{args: `42 - int(.)`},
		{args: `42 / int(.)`},
		{args: `42 * int(.)`},
		{args: `int(.) + 42`},
		{args: `int(.) - 42`},
		{args: `int(.) / 42`},
		{args: `int(.) * 42`},
		{args: `int(.) + 42 + 42`},
		{args: `int(.) - 42 - 42`},
		{args: `int(.) / 42 / 42`},
		{args: `int(.) * 42 * 42`},
		{args: `int(.) + 42 - 42`},
		{args: `int(.) + 42 - 42`},
		{args: `int(.) * 42 / 42`},
		{args: `int(.) / 42 * 42`},
		{args: `42 + int(.) + 42`},
		{args: `42 - int(.) - 42`},
		{args: `42 / int(.) / 42`},
		{args: `42 * int(.) * 42`},
		{args: `42 + int(.) - 42`},
		{args: `42 + int(.) - 42`},
		{args: `42 * int(.) / 42`},
		{args: `42 / int(.) * 42`},
		{args: `42 + 42 + int(.)`},
		{args: `42 - 42 - int(.)`},
		{args: `42 / 42 / int(.)`},
		{args: `42 * 42 * int(.)`},
		{args: `42 + 42 - int(.)`},
		{args: `42 + 42 - int(.)`},
		{args: `42 * 42 / int(.)`},
		{args: `42 / 42 * int(.)`},

		{args: `float(.)`},
		{args: `float(42)`},
		{args: `float("4.2")`},
		{args: `4.2 + 4.2`},
		{args: `4.2 - 4.2`},
		{args: `4.2 / 4.2`},
		{args: `4.2 * 4.2`},
		{args: `4.2 + .`},
		{args: `4.2 - .`},
		{args: `4.2 / .`},
		{args: `4.2 * .`},
		{args: `. + 4.2`},
		{args: `. - 4.2`},
		{args: `. / 4.2`},
		{args: `. * 4.2`},
		{args: `. + 4.2 + 4.2`},
		{args: `. - 4.2 - 4.2`},
		{args: `. / 4.2 / 4.2`},
		{args: `. * 4.2 * 4.2`},
		{args: `. + 4.2 - 4.2`},
		{args: `. + 4.2 - 4.2`},
		{args: `. * 4.2 / 4.2`},
		{args: `. / 4.2 * 4.2`},
		{args: `4.2 + . + 4.2`},
		{args: `4.2 - . - 4.2`},
		{args: `4.2 / . / 4.2`},
		{args: `4.2 * . * 4.2`},
		{args: `4.2 + . - 4.2`},
		{args: `4.2 + . - 4.2`},
		{args: `4.2 * . / 4.2`},
		{args: `4.2 / . * 4.2`},
		{args: `4.2 + 4.2 + .`},
		{args: `4.2 - 4.2 - .`},
		{args: `4.2 / 4.2 / .`},
		{args: `4.2 * 4.2 * .`},
		{args: `4.2 + 4.2 - .`},
		{args: `4.2 + 4.2 - .`},
		{args: `4.2 * 4.2 / .`},
		{args: `4.2 / 4.2 * .`},
		{args: `4.2 + float(.)`},
		{args: `4.2 - float(.)`},
		{args: `4.2 / float(.)`},
		{args: `4.2 * float(.)`},
		{args: `float(.) + 4.2`},
		{args: `float(.) - 4.2`},
		{args: `float(.) / 4.2`},
		{args: `float(.) * 4.2`},
		{args: `float(.) + 4.2 + 4.2`},
		{args: `float(.) - 4.2 - 4.2`},
		{args: `float(.) / 4.2 / 4.2`},
		{args: `float(.) * 4.2 * 4.2`},
		{args: `float(.) + 4.2 - 4.2`},
		{args: `float(.) + 4.2 - 4.2`},
		{args: `float(.) * 4.2 / 4.2`},
		{args: `float(.) / 4.2 * 4.2`},
		{args: `4.2 + float(.) + 4.2`},
		{args: `4.2 - float(.) - 4.2`},
		{args: `4.2 / float(.) / 4.2`},
		{args: `4.2 * float(.) * 4.2`},
		{args: `4.2 + float(.) - 4.2`},
		{args: `4.2 + float(.) - 4.2`},
		{args: `4.2 * float(.) / 4.2`},
		{args: `4.2 / float(.) * 4.2`},
		{args: `4.2 + 4.2 + float(.)`},
		{args: `4.2 - 4.2 - float(.)`},
		{args: `4.2 / 4.2 / float(.)`},
		{args: `4.2 * 4.2 * float(.)`},
		{args: `4.2 + 4.2 - float(.)`},
		{args: `4.2 + 4.2 - float(.)`},
		{args: `4.2 * 4.2 / float(.)`},
		{args: `4.2 / 4.2 * float(.)`},

		{args: `string(.)`},
		{args: `string(42)`},
		{args: `string(true)`},
		{args: `string(4.2)`},
		{args: `"hello" + "hello"`},
		{args: `"hello" + .`},
		{args: `. + "hello"`},
		{args: `. + "hello" + "hello"`},
		{args: `"hello" + . + "hello"`},
		{args: `"hello" + "hello" + .`},
		{args: `"hello" + string(.)`},
		{args: `string(.) + "hello"`},
		{args: `string(.) + "hello" + "hello"`},
		{args: `"hello" + string(.) + "hello"`},
		{args: `"hello" + "hello" + string(.)`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("args=%s", tt.args)
			_, err := Parse(strings.NewReader(tt.args))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
