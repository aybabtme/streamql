package grammar

import (
	"encoding/json"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/aybabtme/streamql/lang/ast"
)

func TestParse(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	// yyDebug = 3
	yyErrorVerbose = true

	var (
		mkAST     = func(expr *ast.Expr) *ast.AST { return &ast.AST{Expr: expr} }
		pipe      = func(lhs, rhs *ast.Expr) *ast.Expr { lhs.Next = rhs; return lhs }
		exprSel   = func(s *ast.Selector) *ast.Expr { return &ast.Expr{Selector: s} }
		exprLit   = func(s *ast.Literal) *ast.Expr { return &ast.Expr{Literal: s} }
		exprUnOp  = func(s *ast.UnaryOperator) *ast.Expr { return &ast.Expr{UnaryOperator: s} }
		exprBinOp = func(s *ast.BinaryOperator) *ast.Expr { return &ast.Expr{BinaryOperator: s} }
		exprFn    = func(s *ast.FuncCall) *ast.Expr { return &ast.Expr{FuncCall: s} }
		selNoop   = func() *ast.Selector { return &ast.Selector{Noop: &ast.NoopSelector{}} }
		selMember = func(expr *ast.Expr, child *ast.Selector) *ast.Selector {
			return &ast.Selector{Member: &ast.MemberSelector{Index: expr, Child: child}}
		}
		selSlice = func(from, to *ast.Expr, child *ast.Selector) *ast.Selector {
			return &ast.Selector{Slice: &ast.SliceSelector{From: from, To: to, Child: child}}
		}
		opNot = func(lhs *ast.Expr) *ast.UnaryOperator {
			return &ast.UnaryOperator{Arg: lhs, LogNot: &ast.OpLogNot{}}
		}
		opAnd = func(lhs, rhs *ast.Expr) *ast.BinaryOperator {
			return &ast.BinaryOperator{LHS: lhs, RHS: rhs, LogAnd: &ast.OpLogAnd{}}
		}
		opOr = func(lhs, rhs *ast.Expr) *ast.BinaryOperator {
			return &ast.BinaryOperator{LHS: lhs, RHS: rhs, LogOr: &ast.OpLogOr{}}
		}
		opAdd = func(lhs, rhs *ast.Expr) *ast.BinaryOperator {
			return &ast.BinaryOperator{LHS: lhs, RHS: rhs, NumAdd: &ast.OpNumAdd{}}
		}
		opSub = func(lhs, rhs *ast.Expr) *ast.BinaryOperator {
			return &ast.BinaryOperator{LHS: lhs, RHS: rhs, NumSub: &ast.OpNumSub{}}
		}
		opMul = func(lhs, rhs *ast.Expr) *ast.BinaryOperator {
			return &ast.BinaryOperator{LHS: lhs, RHS: rhs, NumMul: &ast.OpNumMul{}}
		}
		opDiv = func(lhs, rhs *ast.Expr) *ast.BinaryOperator {
			return &ast.BinaryOperator{LHS: lhs, RHS: rhs, NumDiv: &ast.OpNumDiv{}}
		}
		opEq = func(lhs, rhs *ast.Expr) *ast.BinaryOperator {
			return &ast.BinaryOperator{LHS: lhs, RHS: rhs, CmpEq: &ast.OpCmpEq{}}
		}
		opGt = func(lhs, rhs *ast.Expr) *ast.BinaryOperator {
			return &ast.BinaryOperator{LHS: lhs, RHS: rhs, CmpGt: &ast.OpCmpGt{}}
		}

		fn = func(name string, args ...*ast.Expr) *ast.FuncCall { return &ast.FuncCall{Name: name, Args: args} }

		litBool   = func(v bool) *ast.Literal { return &ast.Literal{Bool: &v} }
		litString = func(v string) *ast.Literal { return &ast.Literal{String: &v} }
		litInt    = func(v int64) *ast.Literal { return &ast.Literal{Int: &v} }
		litFloat  = func(v float64) *ast.Literal { return &ast.Literal{Float: &v} }
		litNull   = func() *ast.Literal { return &ast.Literal{Null: &struct{}{}} }

		_ = mkAST
		_ = pipe
		_ = exprSel
		_ = exprLit
		_ = exprUnOp
		_ = exprBinOp
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
		want    *ast.AST
		wantErr bool
	}{

		{args: "", want: mkAST(nil)},
		{args: ".", want: mkAST(exprSel(selNoop()))},
		{args: ". | .", want: mkAST(pipe(exprSel(selNoop()), exprSel(selNoop())))},
		{args: ". | . | .", want: mkAST(
			pipe(
				exprSel(selNoop()),
				pipe(
					exprSel(selNoop()),
					exprSel(selNoop()),
				),
			),
		)},
		{args: ". | . | . | .", want: mkAST(
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
		{args: "true", want: mkAST(exprLit(litBool(true)))},
		{args: `""`, want: mkAST(exprLit(litString("")))},
		{args: `1`, want: mkAST(exprLit(litInt(1)))},
		{args: `1.0`, want: mkAST(exprLit(litFloat(1)))},
		{args: `null`, want: mkAST(exprLit(litNull()))},
		{args: ".hello", want: mkAST(exprSel(
			selMember(exprLit(litString("hello")), nil),
		))},
		{args: ".hello.bye", want: mkAST(exprSel(
			selMember(exprLit(litString("hello")),
				selMember(exprLit(litString("bye")), nil),
			),
		))},
		{args: ".[]", want: mkAST(
			exprSel(
				selSlice(nil, nil, nil),
			),
		)},
		{args: ".[1]", want: mkAST(
			exprSel(
				selMember(exprLit(litInt(1)), nil),
			),
		)},

		{args: ".[1:2]", want: mkAST(
			exprSel(
				selSlice(
					exprLit(litInt(1)),
					exprLit(litInt(2)),
					nil,
				),
			),
		)},
		{args: ".hello[]", want: mkAST(
			exprSel(
				selMember(
					exprLit(litString("hello")),
					selSlice(nil, nil, nil),
				),
			),
		)},
		{args: ".hello[1]", want: mkAST(
			exprSel(
				selMember(
					exprLit(litString("hello")),
					selMember(exprLit(litInt(1)), nil),
				),
			),
		)},
		{args: ".hello[1:2]", want: mkAST(
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
		{args: ".[][]", want: mkAST(
			exprSel(
				selSlice(nil, nil,
					selSlice(nil, nil, nil),
				),
			),
		)},
		{args: ".[1][1]", want: mkAST(
			exprSel(
				selMember(
					exprLit(litInt(1)),
					selMember(exprLit(litInt(1)), nil),
				),
			),
		)},
		{args: ".[1:2][1:2]", want: mkAST(
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

		{args: ".hello[].bye", want: mkAST(
			exprSel(
				selMember(
					exprLit(litString("hello")),
					selSlice(nil, nil,
						selMember(exprLit(litString("bye")), nil),
					),
				),
			),
		)},
		{args: ".hello[1].bye", want: mkAST(
			exprSel(
				selMember(
					exprLit(litString("hello")),
					selMember(exprLit(litInt(1)),
						selMember(exprLit(litString("bye")), nil),
					),
				),
			),
		)},
		{args: ".hello[1:2].bye", want: mkAST(
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

		{args: `!.`, want: mkAST(
			exprUnOp(opNot(exprSel(selNoop()))),
		)},
		{args: `. && .`, want: mkAST(
			exprBinOp(opAnd(
				exprSel(selNoop()),
				exprSel(selNoop()),
			)),
		)},
		{args: `. || .`, want: mkAST(
			exprBinOp(opOr(
				exprSel(selNoop()),
				exprSel(selNoop()),
			)),
		)},
		{args: `. && . && .`, want: mkAST(
			exprBinOp(opAnd(
				exprBinOp(opAnd(
					exprSel(selNoop()),
					exprSel(selNoop()),
				)),
				exprSel(selNoop()),
			)),
		)},
		{args: `. || . || .`, want: mkAST(
			exprBinOp(opOr(
				exprBinOp(opOr(
					exprSel(selNoop()),
					exprSel(selNoop()),
				)),
				exprSel(selNoop()),
			)),
		)},
		{args: `!. || !. || !.`, want: mkAST(
			exprBinOp(opOr(
				exprBinOp(opOr(
					exprUnOp(opNot(exprSel(selNoop()))),
					exprUnOp(opNot(exprSel(selNoop()))),
				)),
				exprUnOp(opNot(exprSel(selNoop()))),
			)),
		)},
		{args: `!(. || .) || !.`, want: mkAST(
			exprBinOp(opOr(
				exprUnOp(opNot(
					exprBinOp(opOr(
						exprSel(selNoop()),
						exprSel(selNoop()),
					)),
				)),
				exprUnOp(opNot(exprSel(selNoop()))),
			)),
		)},
		{args: `!. || !(. || .)`, want: mkAST(
			exprBinOp(opOr(
				exprUnOp(opNot(exprSel(selNoop()))),
				exprUnOp(opNot(
					exprBinOp(opOr(
						exprSel(selNoop()),
						exprSel(selNoop()),
					)),
				)),
			)),
		)},

		{args: `. + .`, want: mkAST(
			exprBinOp(opAdd(
				exprSel(selNoop()),
				exprSel(selNoop()),
			)),
		)},
		{args: `. - .`, want: mkAST(
			exprBinOp(opSub(
				exprSel(selNoop()),
				exprSel(selNoop()),
			)),
		)},
		{args: `. / .`, want: mkAST(
			exprBinOp(opDiv(
				exprSel(selNoop()),
				exprSel(selNoop()),
			)),
		)},
		{args: `. * .`, want: mkAST(
			exprBinOp(opMul(
				exprSel(selNoop()),
				exprSel(selNoop()),
			)),
		)},
		{args: `. + . + .`, want: mkAST(
			exprBinOp(opAdd(
				exprBinOp(opAdd(
					exprSel(selNoop()),
					exprSel(selNoop()),
				)),
				exprSel(selNoop()),
			)),
		)},
		{args: `. - . - .`, want: mkAST(
			exprBinOp(opSub(
				exprBinOp(opSub(
					exprSel(selNoop()),
					exprSel(selNoop()),
				)),
				exprSel(selNoop()),
			)),
		)},
		{args: `. / . / .`, want: mkAST(
			exprBinOp(opDiv(
				exprBinOp(opDiv(
					exprSel(selNoop()),
					exprSel(selNoop()),
				)),
				exprSel(selNoop()),
			)),
		)},
		{args: `. * . * .`, want: mkAST(
			exprBinOp(opMul(
				exprBinOp(opMul(
					exprSel(selNoop()),
					exprSel(selNoop()),
				)),
				exprSel(selNoop()),
			)),
		)},

		{args: `1 + 2 - 3`, want: mkAST(
			exprBinOp(opAdd(
				exprLit(litInt(1)),
				exprBinOp(opSub(
					exprLit(litInt(2)),
					exprLit(litInt(3)),
				)),
			)),
		)},
		{args: `1 - 2 + 3`, want: mkAST(
			exprBinOp(opAdd(
				exprBinOp(opSub(
					exprLit(litInt(1)),
					exprLit(litInt(2)),
				)),
				exprLit(litInt(3)),
			)),
		)},
		{args: `1 * 2 / 3`, want: mkAST(
			exprBinOp(opMul(
				exprLit(litInt(1)),
				exprBinOp(opDiv(
					exprLit(litInt(2)),
					exprLit(litInt(3)),
				)),
			)),
		)},
		{args: `1 / 2 * 3`, want: mkAST(
			exprBinOp(opMul(
				exprBinOp(opDiv(
					exprLit(litInt(1)),
					exprLit(litInt(2)),
				)),
				exprLit(litInt(3)),
			)),
		)},

		{args: `1 - (2 + 3)`, want: mkAST(
			exprBinOp(opSub(
				exprLit(litInt(1)),
				exprBinOp(opAdd(
					exprLit(litInt(2)),
					exprLit(litInt(3)),
				)),
			)),
		)},
		{args: `1 / (2 * 3)`, want: mkAST(
			exprBinOp(opDiv(
				exprLit(litInt(1)),
				exprBinOp(opMul(
					exprLit(litInt(2)),
					exprLit(litInt(3)),
				)),
			)),
		)},

		{args: "select(true)", want: mkAST(
			exprFn(fn("select", exprLit(litBool(true)))),
		)},
		{args: "select(true && true)", want: mkAST(
			exprFn(fn("select",
				exprBinOp(opAnd(
					exprLit(litBool(true)),
					exprLit(litBool(true)),
				)),
			)),
		)},
		{args: "select(.)", want: mkAST(
			exprFn(fn("select", exprSel(selNoop()))),
		)},
		{args: "select(.lol)", want: mkAST(
			exprFn(fn("select", exprSel(
				selMember(exprLit(litString("lol")), nil),
			))),
		)},
		{args: "select(bool(.lol) && true)", want: mkAST(
			exprFn(fn("select", exprBinOp(
				opAnd(
					exprFn(fn("bool", exprSel(
						selMember(exprLit(litString("lol")), nil),
					))),
					exprLit(litBool(true)),
				),
			))),
		)},
		{args: "select(.lol && true)", want: mkAST(
			exprFn(fn("select", exprBinOp(
				opAnd(
					exprSel(selMember(exprLit(litString("lol")), nil)),
					exprLit(litBool(true)),
				),
			))),
		)},

		{args: `1+1>2 && 2+2 == 4 || 1+1>2 && !(2+2 == 4)`, want: mkAST(
			exprBinOp(opOr(
				// 1+1>2 && 2+2 == 4
				exprBinOp(opAnd(
					exprBinOp(opGt(
						exprBinOp(opAdd(
							exprLit(litInt(1)),
							exprLit(litInt(1)),
						)),
						exprLit(litInt(2)),
					)),
					exprBinOp(opEq(
						exprBinOp(opAdd(
							exprLit(litInt(2)),
							exprLit(litInt(2)),
						)),
						exprLit(litInt(4)),
					)),
				)),
				// 1+1>2 && !2+2 == 4
				exprBinOp(opAnd(
					exprBinOp(opGt(
						exprBinOp(opAdd(
							exprLit(litInt(1)),
							exprLit(litInt(1)),
						)),
						exprLit(litInt(2)),
					)),
					exprUnOp(opNot(
						exprBinOp(opEq(
							exprBinOp(opAdd(
								exprLit(litInt(2)),
								exprLit(litInt(2)),
							)),
							exprLit(litInt(4)),
						)),
					)),
				)),
			)),
		)},

		{args: `.lol[0:1] | select(.is_red && string(.size) == "large")`, want: mkAST(
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
					exprBinOp(opAnd(
						exprSel(selMember(exprLit(litString("is_red")), nil)),
						exprBinOp(opEq(
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
		{args: `.lol[0:1] | select(.is_red && string(.size) == "large") | select(.)`, want: mkAST(
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
						exprBinOp(opAnd(
							exprSel(selMember(exprLit(litString("is_red")), nil)),
							exprBinOp(opEq(
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

		{args: "select(.)", want: mkAST(
			exprFn(fn("select", exprSel(
				selNoop(),
			))),
		)},
		{args: "select(.size > 1)", want: mkAST(
			exprFn(fn("select", exprBinOp(
				opGt(
					exprSel(selMember(exprLit(litString("size")), nil)),
					exprLit(litInt(1)),
				),
			))),
		)},
		{args: "select(.size > 1) / 1000", want: mkAST(
			exprBinOp(opDiv(
				exprFn(fn("select", exprBinOp(
					opGt(
						exprSel(selMember(exprLit(litString("size")), nil)),
						exprLit(litInt(1)),
					),
				))),
				exprLit(litInt(1000)),
			)),
		)},
		{args: "reduce(select(.size > 1) / 1000)", want: mkAST(
			exprFn(fn(
				"reduce",
				exprBinOp(opDiv(
					exprFn(fn("select", exprBinOp(
						opGt(
							exprSel(selMember(exprLit(litString("size")), nil)),
							exprLit(litInt(1)),
						),
					))),
					exprLit(litInt(1000)),
				)),
			)),
		)},
		{args: "reduce(select(.size > 1) | . / 1000)", want: mkAST(
			exprFn(fn(
				"reduce",
				pipe(
					exprFn(fn("select", exprBinOp(
						opGt(
							exprSel(selMember(exprLit(litString("size")), nil)),
							exprLit(litInt(1)),
						),
					))),
					exprBinOp(opDiv(
						exprSel(selNoop()),
						exprLit(litInt(1000)),
					)),
				),
			)),
		)},
		{args: "select(.cond.keep) | .name", want: mkAST(
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

		{args: `-.`},
		{args: `-42`},
		{args: `-4.2`},
		{args: `-int(.)`},
		{args: `-int(42)`},
		{args: `-int("4.2")`},
		{args: `-float(.)`},
		{args: `-float(42)`},
		{args: `-float("4.2")`},
		{args: `-.`},
		{args: `-42`},
		{args: `-4.2`},
		{args: `-int(.)`},
		{args: `-int(42)`},
		{args: `-int("4.2")`},
		{args: `-float(.)`},
		{args: `-float(42)`},
		{args: `-float("4.2")`},

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
