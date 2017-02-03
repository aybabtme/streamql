package parser

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/aybabtme/streamql/lang/ast"
	"github.com/aybabtme/streamql/lang/token"
)

func TestPositiveParse(t *testing.T) {

	tests := []struct {
		input string
		want  *ast.FiltersStmt
	}{
		{
			input: ".",
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{}},
					}},
				},
			},
		},
		{
			input: ".,.",
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{}},
					}},
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{}},
					}},
				},
			},
		},
		{
			input: ".|.",
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{}},
						{Selector: &ast.SelectorStmt{}},
					}},
				},
			},
		},
		{
			input: ".hello",
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{
							Object: &ast.ObjectSelectorStmt{
								Member: "hello",
							},
						}},
					}},
				},
			},
		},
		{
			input: `.[].id`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{
						Funcs: []*ast.FuncStmt{
							{
								Selector: &ast.SelectorStmt{
									Array: &ast.ArraySelectorStmt{
										Each: &ast.EachSelectorStmt{},
										Child: &ast.SelectorStmt{
											Object: &ast.ObjectSelectorStmt{
												Member: "id",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			input: ".hello | .bye",
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{
							Object: &ast.ObjectSelectorStmt{
								Member: "hello",
							},
						}},
						{Selector: &ast.SelectorStmt{
							Object: &ast.ObjectSelectorStmt{
								Member: "bye",
							},
						}},
					}},
				},
			},
		},
		{
			input: ".hello , .bye",
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{
							Object: &ast.ObjectSelectorStmt{
								Member: "hello",
							},
						}},
					}},
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{
							Object: &ast.ObjectSelectorStmt{
								Member: "bye",
							},
						}},
					}},
				},
			},
		},
		{
			input: ".hello    ",
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{
							Object: &ast.ObjectSelectorStmt{
								Member: "hello",
							},
						}},
					}},
				},
			},
		},
		{
			input: `.hello\ \ \ \ `,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{
							Object: &ast.ObjectSelectorStmt{
								Member: "hello    ",
							},
						}},
					}},
				},
			},
		},
		{
			input: ".[]",
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{
							Array: &ast.ArraySelectorStmt{
								Each: &ast.EachSelectorStmt{},
							},
						}},
					}},
				},
			},
		},
		{
			input: ".hello[]",
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{
							Object: &ast.ObjectSelectorStmt{
								Member: "hello",
								Child: &ast.SelectorStmt{
									Array: &ast.ArraySelectorStmt{
										Each: &ast.EachSelectorStmt{},
									},
								},
							},
						}},
					}},
				},
			},
		},
		{
			input: ".hello[1]",
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{
							Object: &ast.ObjectSelectorStmt{
								Member: "hello",
								Child: &ast.SelectorStmt{
									Array: &ast.ArraySelectorStmt{
										Index: &ast.IndexSelectorStmt{
											Index: &ast.IntegerArg{Integer: intp(1)},
										},
									},
								},
							},
						}},
					}},
				},
			},
		},
		{
			input: ".hello[1:2]",
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{
							Object: &ast.ObjectSelectorStmt{
								Member: "hello",
								Child: &ast.SelectorStmt{
									Array: &ast.ArraySelectorStmt{
										RangeEach: &ast.RangeEachSelectorStmt{
											From: &ast.IntegerArg{Integer: intp(1)},
											To:   &ast.IntegerArg{Integer: intp(2)},
										},
									},
								},
							},
						}},
					}},
				},
			},
		},
		{
			input: ".hello[1:2][42][].world",
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{
							Object: &ast.ObjectSelectorStmt{
								Member: "hello",
								Child: &ast.SelectorStmt{
									Array: &ast.ArraySelectorStmt{
										RangeEach: &ast.RangeEachSelectorStmt{
											From: &ast.IntegerArg{Integer: intp(1)},
											To:   &ast.IntegerArg{Integer: intp(2)},
										},
										Child: &ast.SelectorStmt{
											Array: &ast.ArraySelectorStmt{
												Index: &ast.IndexSelectorStmt{
													Index: &ast.IntegerArg{Integer: intp(42)},
												},
												Child: &ast.SelectorStmt{
													Array: &ast.ArraySelectorStmt{
														Each: &ast.EachSelectorStmt{},
														Child: &ast.SelectorStmt{
															Object: &ast.ObjectSelectorStmt{
																Member: "world",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						}},
					}},
				},
			},
		},

		{
			input: `.[].id1 | .[].id2, .[].id3`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{
							Array: &ast.ArraySelectorStmt{
								Each: &ast.EachSelectorStmt{},
								Child: &ast.SelectorStmt{
									Object: &ast.ObjectSelectorStmt{
										Member: "id1",
									},
								},
							},
						}},
						{Selector: &ast.SelectorStmt{
							Array: &ast.ArraySelectorStmt{
								Each: &ast.EachSelectorStmt{},
								Child: &ast.SelectorStmt{
									Object: &ast.ObjectSelectorStmt{
										Member: "id2",
									},
								},
							},
						}},
					}},
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{
							Array: &ast.ArraySelectorStmt{
								Each: &ast.EachSelectorStmt{},
								Child: &ast.SelectorStmt{
									Object: &ast.ObjectSelectorStmt{
										Member: "id3",
									},
								},
							},
						}},
					}},
				},
			},
		},
	}

	for n, tt := range tests {
		t.Run("", func(t *testing.T) {
			t.Logf("test #%d, input %q", n, tt.input)

			got, err := NewParser(strings.NewReader(tt.input)).Parse()
			if err != nil {
				t.Errorf("%+v", err)
				return
			}
			if !reflect.DeepEqual(tt.want, got) {

				t.Errorf("want=%#v", tt.want)
				t.Errorf(" got=%#v", got)

				wantData, _ := json.MarshalIndent(tt.want, "", "  ")
				gotData, _ := json.MarshalIndent(got, "", "  ")

				t.Errorf("want=%#v", tt.want)
				t.Errorf(" got=%#v", got)

				t.Errorf("want=%s", string(wantData))
				t.Errorf(" got=%s", string(gotData))

			}
		})
	}
}

func TestPositiveParseOperations(t *testing.T) {

	tests := []struct {
		input string
		want  *ast.FiltersStmt
	}{
		{
			input: `true`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{EmitFunc: &ast.EmitFuncStmt{
							EmitBooleanFunc: &ast.EmitBooleanFunc{
								Literal: &ast.BooleanArg{Boolean: boolp(true)},
							},
						}},
					}},
				},
			},
		},
		{
			input: `42`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{EmitFunc: &ast.EmitFuncStmt{
							EmitNumberFunc: &ast.EmitNumberFunc{
								Float: &ast.EmitFloatFunc{
									Literal: &ast.NumberArg{Number: float64p(42)},
								},
							},
						}},
					}},
				},
			},
		},

		{
			input: `1+1`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{EmitFunc: &ast.EmitFuncStmt{
							EmitNumberFunc: &ast.EmitNumberFunc{
								Algebra: &ast.AlgebraNumberOps{
									Add: &ast.FuncNumberAdd{
										LHS: &ast.NumberArg{Number: float64p(1)},
										RHS: &ast.NumberArg{Number: float64p(1)},
									},
								},
							},
						}},
					}},
				},
			},
		},

		{
			input: `1 + 1`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{EmitFunc: &ast.EmitFuncStmt{
							EmitNumberFunc: &ast.EmitNumberFunc{
								Algebra: &ast.AlgebraNumberOps{
									Add: &ast.FuncNumberAdd{
										LHS: &ast.NumberArg{Number: float64p(1)},
										RHS: &ast.NumberArg{Number: float64p(1)},
									},
								},
							},
						}},
					}},
				},
			},
		},

		{
			input: `atof("not a number lol")`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{EmitFunc: &ast.EmitFuncStmt{
							EmitNumberFunc: &ast.EmitNumberFunc{
								Float: &ast.EmitFloatFunc{
									Literal: &ast.NumberArg{
										EmitNumberFunc: &ast.EmitNumberFunc{
											Float: &ast.EmitFloatFunc{
												StringAtof: &ast.FuncStringAtof{
													String: &ast.StringArg{String: stringp("not a number lol")},
												},
											},
										},
									},
								},
							},
						}},
					}},
				},
			},
		},

		{
			input: `regexp("not a number lol", .)`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{EmitFunc: &ast.EmitFuncStmt{
							EmitBooleanFunc: &ast.EmitBooleanFunc{
								Literal: &ast.BooleanArg{
									EmitBooleanFunc: &ast.EmitBooleanFunc{
										StringRegexp: &ast.FuncStringRegexp{
											Expression: &ast.StringArg{String: stringp("not a number lol")},
											Target:     &ast.StringArg{Selector: &ast.SelectorStmt{}},
										},
									},
								},
							},
						}},
					}},
				},
			},
		},
		{
			input: `regexp("not a number lol", "lol")`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{EmitFunc: &ast.EmitFuncStmt{
							EmitBooleanFunc: &ast.EmitBooleanFunc{
								Literal: &ast.BooleanArg{
									EmitBooleanFunc: &ast.EmitBooleanFunc{
										StringRegexp: &ast.FuncStringRegexp{
											Expression: &ast.StringArg{String: stringp("not a number lol")},
											Target:     &ast.StringArg{String: stringp("lol")},
										},
									},
								},
							},
						}},
					}},
				},
			},
		},

		{
			input: `length("lol")`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{EmitFunc: &ast.EmitFuncStmt{
							EmitNumberFunc: &ast.EmitNumberFunc{
								Float: &ast.EmitFloatFunc{
									Literal: &ast.NumberArg{
										EmitNumberFunc: &ast.EmitNumberFunc{
											Int: &ast.EmitIntFunc{
												StringLength: &ast.FuncStringLength{
													String: &ast.StringArg{String: stringp("lol")},
												},
											},
										},
									},
								},
							},
						}},
					}},
				},
			},
		},

		{
			input: `true and true`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{EmitFunc: &ast.EmitFuncStmt{
							EmitBooleanFunc: &ast.EmitBooleanFunc{
								Algebra: &ast.AlgebraBooleanOps{
									And: &ast.FuncBooleanAnd{
										LHS: &ast.BooleanArg{Boolean: boolp(true)},
										RHS: &ast.BooleanArg{Boolean: boolp(true)},
									},
								},
							},
						}},
					}},
				},
			},
		},

		{
			input: `true xor true`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{EmitFunc: &ast.EmitFuncStmt{
							EmitBooleanFunc: &ast.EmitBooleanFunc{
								Algebra: &ast.AlgebraBooleanOps{
									XOR: &ast.FuncBooleanXOR{
										LHS: &ast.BooleanArg{Boolean: boolp(true)},
										RHS: &ast.BooleanArg{Boolean: boolp(true)},
									},
								},
							},
						}},
					}},
				},
			},
		},

		{
			input: `false or true`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{EmitFunc: &ast.EmitFuncStmt{
							EmitBooleanFunc: &ast.EmitBooleanFunc{
								Algebra: &ast.AlgebraBooleanOps{
									Or: &ast.FuncBooleanOr{
										LHS: &ast.BooleanArg{Boolean: boolp(false)},
										RHS: &ast.BooleanArg{Boolean: boolp(true)},
									},
								},
							},
						}},
					}},
				},
			},
		},

		{
			input: `select(true)`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{EmitFunc: &ast.EmitFuncStmt{
							EmitAnyFunc: &ast.EmitAnyFunc{
								AnySelect: &ast.FuncAnySelect{
									Condition: &ast.BooleanArg{Boolean: boolp(true)},
								},
							},
						}},
					}},
				},
			},
		},

		{
			input: `substring("hello", 1, 2)`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{EmitFunc: &ast.EmitFuncStmt{
							EmitStringFunc: &ast.EmitStringFunc{
								StringSubStr: &ast.FuncStringSubStr{
									String: &ast.StringArg{String: stringp("hello")},
									From:   &ast.IntegerArg{Integer: intp(1)},
									To:     &ast.IntegerArg{Integer: intp(2)},
								},
							},
						}},
					}},
				},
			},
		},

		{
			input: `substring("hello", 0, length("hello"))`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{EmitFunc: &ast.EmitFuncStmt{
							EmitStringFunc: &ast.EmitStringFunc{
								StringSubStr: &ast.FuncStringSubStr{
									String: &ast.StringArg{String: stringp("hello")},
									From:   &ast.IntegerArg{Integer: intp(0)},
									To: &ast.IntegerArg{EmitIntFunc: &ast.EmitIntFunc{
										StringLength: &ast.FuncStringLength{
											String: &ast.StringArg{String: stringp("hello")},
										},
									}},
								},
							},
						}},
					}},
				},
			},
		},

		{
			input: `.body | select(contains(.level, "warning"))`,
			want: &ast.FiltersStmt{
				Filters: []*ast.FilterStmt{
					{Funcs: []*ast.FuncStmt{
						{Selector: &ast.SelectorStmt{
							Object: &ast.ObjectSelectorStmt{
								Member: "body",
							},
						}},
						{EmitFunc: &ast.EmitFuncStmt{
							EmitAnyFunc: &ast.EmitAnyFunc{
								AnySelect: &ast.FuncAnySelect{
									Condition: &ast.BooleanArg{
										EmitBooleanFunc: &ast.EmitBooleanFunc{
											StringContains: &ast.FuncStringContains{
												SubString: &ast.StringArg{
													Selector: &ast.SelectorStmt{
														Object: &ast.ObjectSelectorStmt{
															Member: "level",
														},
													},
												},
												Target: &ast.StringArg{
													String: stringp("warning"),
												},
											},
										},
									},
								},
							},
						}},
					}},
				},
			},
		},

		// {  // integer algebra not supported in current parser ;(
		// 	input: `substring("hello", length("hello")-1, length("hello"))`,
		// 	want: &ast.FiltersStmt{
		// 		Filters: []*ast.FilterStmt{
		// 			{Funcs: []*ast.FuncStmt{
		// 				{EmitFunc: &ast.EmitFuncStmt{
		// 					EmitStringFunc: &ast.EmitStringFunc{
		// 						StringSubStr: &ast.FuncStringSubStr{
		// 							String: &ast.StringArg{String: stringp("hello")},
		// 							From:   &ast.IntegerArg{Integer: intp(1)},
		// 							To:     &ast.IntegerArg{Integer: intp(2)},
		// 						},
		// 					},
		// 				}},
		// 			}},
		// 		},
		// 	},
		// },
	}

	for n, tt := range tests {
		t.Run("", func(t *testing.T) {
			t.Logf("test #%d, input %q", n, tt.input)

			got, err := NewParser(strings.NewReader(tt.input)).Parse()
			if err != nil {
				t.Errorf("%+v", err)
				return
			}
			if !reflect.DeepEqual(tt.want, got) {

				t.Errorf("want=%#v", tt.want)
				t.Errorf(" got=%#v", got)

				wantData, _ := json.MarshalIndent(tt.want, "", "  ")
				gotData, _ := json.MarshalIndent(got, "", "  ")

				t.Errorf("want=%#v", tt.want)
				t.Errorf(" got=%#v", got)

				t.Errorf("want=%s", string(wantData))
				t.Errorf(" got=%s", string(gotData))

			}
		})
	}
}

func TestNegativeParse(t *testing.T) {

	tests := []struct {
		input string
		want  error
	}{
		{
			input: "",
			want:  io.ErrUnexpectedEOF,
		},
		{
			input: ".]",
			want: &SyntaxError{
				Expected: []token.Token{token.Comma, token.Pipe},
				Actual:   token.RightBracket,
			},
		},
		{
			input: "hello",
			want:  newUnknownKeywordError("hello", containsKeyword, regexpKeyword, notKeyword, substringKeyword, selectKeyword, atofKeyword, lengthKeyword),
		},
		{
			input: ".[1:2 | ]",
			want: &SyntaxError{
				Expected: []token.Token{token.RightBracket},
				Actual:   token.Pipe,
			},
		},
		{
			input: ". hello",
			want: &SyntaxError{
				Expected: []token.Token{token.Comma, token.Pipe},
				Actual:   token.InlineString,
			},
		},
		{
			input: ".|.|.|.|",
			want:  io.ErrUnexpectedEOF,
		},
		{
			input: ".,.,.,.,",
			want:  io.ErrUnexpectedEOF,
		},
		{
			input: ",",
			want: &SyntaxError{
				Expected: []token.Token{token.Dot, token.LeftParens, token.InlineString, token.Float, token.Integer},
				Actual:   token.Comma,
			},
		},
	}

	for n, tt := range tests {
		t.Run("", func(t *testing.T) {
			t.Logf("test #%d, input %q", n, tt.input)

			tree, got := NewParser(strings.NewReader(tt.input)).Parse()
			if got == nil {
				treeData, _ := json.MarshalIndent(tree, "", "  ")
				t.Errorf("tree=%s", string(treeData))
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("want=%v", tt.want)
				t.Errorf(" got=%v", got)
				// t.Errorf("want=%#v", tt.want)
				// t.Errorf(" got=%#v", got)
			}
		})
	}
}

func intp(v int64) *int64         { return &v }
func stringp(v string) *string    { return &v }
func boolp(v bool) *bool          { return &v }
func float64p(v float64) *float64 { return &v }
