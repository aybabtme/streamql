package vmtest

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aybabtme/streamql/lang/ast"
	"github.com/aybabtme/streamql/lang/grammar"
	"github.com/aybabtme/streamql/lang/msg"
	"github.com/aybabtme/streamql/lang/msg/gomsg"
	"github.com/aybabtme/streamql/lang/vm"
)

// Verify that a VM emits the expected messages, given input messages and a query.
func Verify(t *testing.T, mkVM func(*ast.AST, *vm.Options) vm.VM) {
	bd := gomsg.Build()

	tests := []struct {
		name   string
		strict bool
		input  []msg.Msg
		query  []string
		want   []msg.Msg
	}{
		{"passthru of nothing", true,
			list(),
			[]string{""},
			list(),
		},
		{"passthru of nothing", true,
			list(),
			[]string{"."},
			list(),
		},
		{"passthru", true,
			list(mustBool(bd, true)),
			[]string{""},
			list(mustBool(bd, true)),
		},
		{"passthru", true,
			list(mustBool(bd, true)),
			[]string{"."},
			list(mustBool(bd, true)),
		},
		{"explode", true,
			list(
				mustArray(bd,
					mustInt(bd, 1),
					mustInt(bd, 2),
					mustInt(bd, 3),
				),
				mustArray(bd,
					mustInt(bd, 4),
					mustInt(bd, 5),
					mustInt(bd, 6),
				),
			),
			[]string{".[]"},
			list(
				mustInt(bd, 1),
				mustInt(bd, 2),
				mustInt(bd, 3),
				mustInt(bd, 4),
				mustInt(bd, 5),
				mustInt(bd, 6),
			),
		},
		{"explode", true,
			list(
				mustArray(bd,
					mustInt(bd, 1),
					mustInt(bd, 2),
					mustInt(bd, 3),
				),
				mustArray(bd,
					mustInt(bd, 4),
					mustInt(bd, 5),
					mustInt(bd, 6),
				),
			),
			[]string{".[]"},
			list(
				mustInt(bd, 1),
				mustInt(bd, 2),
				mustInt(bd, 3),
				mustInt(bd, 4),
				mustInt(bd, 5),
				mustInt(bd, 6),
			),
		},
		{"explode recursively", true,
			list(
				mustArray(bd,
					mustArray(bd,
						mustInt(bd, 1),
						mustInt(bd, 2),
						mustInt(bd, 3),
					),
					mustArray(bd,
						mustInt(bd, 4),
						mustInt(bd, 5),
						mustInt(bd, 6),
					),
				),
				mustArray(bd,
					mustArray(bd,
						mustInt(bd, 7),
						mustInt(bd, 8),
						mustInt(bd, 9),
					),
					mustArray(bd,
						mustInt(bd, 10),
						mustInt(bd, 11),
						mustInt(bd, 12),
					),
				),
			),
			[]string{".[][]"},
			list(
				mustInt(bd, 1),
				mustInt(bd, 2),
				mustInt(bd, 3),
				mustInt(bd, 4),
				mustInt(bd, 5),
				mustInt(bd, 6),
				mustInt(bd, 7),
				mustInt(bd, 8),
				mustInt(bd, 9),
				mustInt(bd, 10),
				mustInt(bd, 11),
				mustInt(bd, 12),
			),
		},
		{"explode parts skip first", true,
			list(
				mustArray(bd,
					mustInt(bd, 1),
					mustInt(bd, 2),
					mustInt(bd, 3),
				),
				mustArray(bd,
					mustInt(bd, 4),
					mustInt(bd, 5),
					mustInt(bd, 6),
				),
			),
			[]string{".[1:]"},
			list(
				mustInt(bd, 2),
				mustInt(bd, 3),
				mustInt(bd, 5),
				mustInt(bd, 6),
			),
		},
		{"explode parts skip last", true,
			list(
				mustArray(bd,
					mustInt(bd, 1),
					mustInt(bd, 2),
					mustInt(bd, 3),
				),
				mustArray(bd,
					mustInt(bd, 4),
					mustInt(bd, 5),
					mustInt(bd, 6),
				),
			),
			[]string{".[:2]"},
			list(
				mustInt(bd, 1),
				mustInt(bd, 2),
				mustInt(bd, 4),
				mustInt(bd, 5),
			),
		},
		{"explode parts skip first and last", true,
			list(
				mustArray(bd,
					mustInt(bd, 1),
					mustInt(bd, 2),
					mustInt(bd, 3),
				),
				mustArray(bd,
					mustInt(bd, 4),
					mustInt(bd, 5),
					mustInt(bd, 6),
				),
			),
			[]string{".[1:2]"},
			list(
				mustInt(bd, 2),
				mustInt(bd, 5),
			),
		},
		{"index into an object", true,
			list(
				mustObject(bd, map[string]msg.Msg{
					"hello": mustString(bd, "world"),
				}),
			),
			[]string{".hello"},
			list(
				mustString(bd, "world"),
			),
		},
		{"index into recursively into an object", true,
			list(
				mustObject(bd, map[string]msg.Msg{
					"hello": mustObject(bd, map[string]msg.Msg{
						"world": mustFloat(bd, 3.14159),
					}),
				}),
				mustObject(bd, map[string]msg.Msg{}),
				mustObject(bd, map[string]msg.Msg{
					"hello": mustObject(bd, map[string]msg.Msg{}),
				}),
			),
			[]string{".hello.world"},
			list(
				mustFloat(bd, 3.14159),
			),
		},
		{"maybe index into recursively into an object", true,
			list(
				mustObject(bd, map[string]msg.Msg{
					"hello": mustObject(bd, map[string]msg.Msg{
						"world": mustFloat(bd, 3.14159),
					}),
				}),
				mustObject(bd, map[string]msg.Msg{
					"hello": mustObject(bd, map[string]msg.Msg{
						"world": mustFloat(bd, 2*3.14159),
					}),
				}),
			),
			[]string{".hello.world | select(. > 4.0)"},
			list(
				mustFloat(bd, 2*3.14159),
			),
		},
		{"select into an object", true,
			list(
				mustObject(bd, map[string]msg.Msg{
					"keep": mustBool(bd, true),
					"name": mustString(bd, "item0"),
				}),
				mustObject(bd, map[string]msg.Msg{
					"keep": mustBool(bd, false),
					"name": mustString(bd, "item1"),
				}),
				mustObject(bd, map[string]msg.Msg{
					"keep": mustBool(bd, true),
					"name": mustString(bd, "item2"),
				}),
			),
			[]string{"select(.keep) | .name"},
			list(
				mustString(bd, "item0"),
				mustString(bd, "item2"),
			),
		},
		{"select recursively into an object", true,
			list(
				mustObject(bd, map[string]msg.Msg{
					"cond": mustObject(bd, map[string]msg.Msg{
						"keep": mustBool(bd, true),
					}),
					"name": mustString(bd, "item0"),
				}),
				mustObject(bd, map[string]msg.Msg{
					"cond": mustObject(bd, map[string]msg.Msg{
						"keep": mustBool(bd, false),
					}),
					"name": mustString(bd, "item1"),
				}),
				mustObject(bd, map[string]msg.Msg{
					"cond": mustObject(bd, map[string]msg.Msg{
						"keep": mustBool(bd, true),
					}),
					"name": mustString(bd, "item2"),
				}),
			),
			[]string{"select(.cond.keep) | .name"},
			list(
				mustString(bd, "item0"),
				mustString(bd, "item2"),
			),
		},
		{"equality", true,
			list(
				// float equality
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustFloat(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustFloat(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 3.1415), "r": mustFloat(bd, 3.1415)}),

				// int equality
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustInt(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 2), "r": mustInt(bd, 1)}),

				// int~float equality
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 2), "r": mustInt(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 2), "r": mustFloat(bd, 2)}),

				// string equality
				mustObject(bd, map[string]msg.Msg{"l": mustString(bd, "hello"), "r": mustString(bd, "hello")}),
				mustObject(bd, map[string]msg.Msg{"l": mustString(bd, "hello"), "r": mustString(bd, "world")}),

				// bool equality
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, true), "r": mustBool(bd, true)}),
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, true), "r": mustBool(bd, false)}),

				// array equality
				mustObject(bd, map[string]msg.Msg{
					"l": mustArray(bd,
						mustInt(bd, 0),
						mustInt(bd, 1),
					),
					"r": mustArray(bd,
						mustInt(bd, 0),
						mustInt(bd, 1),
					),
				}),
				mustObject(bd, map[string]msg.Msg{
					"l": mustArray(bd,
						mustInt(bd, 1),
						mustInt(bd, 0),
					),
					"r": mustArray(bd,
						mustInt(bd, 0),
						mustInt(bd, 1),
					),
				}),

				// object equality
				mustObject(bd, map[string]msg.Msg{
					"l": mustObject(bd, map[string]msg.Msg{"hello": mustString(bd, "world")}),
					"r": mustObject(bd, map[string]msg.Msg{"hello": mustString(bd, "world")}),
				}),
				mustObject(bd, map[string]msg.Msg{
					"l": mustObject(bd, map[string]msg.Msg{"hello": mustString(bd, "world")}),
					"r": mustObject(bd, map[string]msg.Msg{"bye": mustString(bd, "world")}),
				}),

				// incompatible types equality
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, true), "r": mustInt(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 2), "r": mustString(bd, "")}),
			),
			[]string{".l == .r"},
			list(
				mustBool(bd, true),
				mustBool(bd, false),
				mustBool(bd, true),

				mustBool(bd, true),
				mustBool(bd, false),

				mustBool(bd, true),
				mustBool(bd, true),

				mustBool(bd, true),
				mustBool(bd, false),

				mustBool(bd, true),
				mustBool(bd, false),

				mustBool(bd, true),
				mustBool(bd, false),

				mustBool(bd, true),
				mustBool(bd, false),

				mustBool(bd, false),
				mustBool(bd, false),
			),
		},
		{"non-equality", true,
			list(
				// float equality
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustFloat(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustFloat(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 3.1415), "r": mustFloat(bd, 3.1415)}),

				// int equality
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustInt(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 2), "r": mustInt(bd, 1)}),

				// int~float equality
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 2), "r": mustInt(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 2), "r": mustFloat(bd, 2)}),

				// string equality
				mustObject(bd, map[string]msg.Msg{"l": mustString(bd, "hello"), "r": mustString(bd, "hello")}),
				mustObject(bd, map[string]msg.Msg{"l": mustString(bd, "hello"), "r": mustString(bd, "world")}),

				// bool equality
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, true), "r": mustBool(bd, true)}),
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, true), "r": mustBool(bd, false)}),

				// array equality
				mustObject(bd, map[string]msg.Msg{
					"l": mustArray(bd,
						mustInt(bd, 0),
						mustInt(bd, 1),
					),
					"r": mustArray(bd,
						mustInt(bd, 0),
						mustInt(bd, 1),
					),
				}),
				mustObject(bd, map[string]msg.Msg{
					"l": mustArray(bd,
						mustInt(bd, 1),
						mustInt(bd, 0),
					),
					"r": mustArray(bd,
						mustInt(bd, 0),
						mustInt(bd, 1),
					),
				}),

				// object equality
				mustObject(bd, map[string]msg.Msg{
					"l": mustObject(bd, map[string]msg.Msg{"hello": mustString(bd, "world")}),
					"r": mustObject(bd, map[string]msg.Msg{"hello": mustString(bd, "world")}),
				}),
				mustObject(bd, map[string]msg.Msg{
					"l": mustObject(bd, map[string]msg.Msg{"hello": mustString(bd, "world")}),
					"r": mustObject(bd, map[string]msg.Msg{"bye": mustString(bd, "world")}),
				}),

				// incompatible types equality
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, true), "r": mustInt(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 2), "r": mustString(bd, "")}),
			),
			[]string{".l != .r"},
			list(
				mustBool(bd, false),
				mustBool(bd, true),
				mustBool(bd, false),

				mustBool(bd, false),
				mustBool(bd, true),

				mustBool(bd, false),
				mustBool(bd, false),

				mustBool(bd, false),
				mustBool(bd, true),

				mustBool(bd, false),
				mustBool(bd, true),

				mustBool(bd, false),
				mustBool(bd, true),

				mustBool(bd, false),
				mustBool(bd, true),

				mustBool(bd, true),
				mustBool(bd, true),
			),
		},
		{"greater than", true,
			list(
				// float equality
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustFloat(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 2), "r": mustFloat(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 3.1415), "r": mustFloat(bd, 3)}),

				// int equality
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustInt(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 2), "r": mustInt(bd, 1)}),

				// int~float equality
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 2), "r": mustInt(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 2), "r": mustFloat(bd, 1)}),

				// string equality
				mustObject(bd, map[string]msg.Msg{"l": mustString(bd, "hello"), "r": mustString(bd, "hello")}),
				mustObject(bd, map[string]msg.Msg{"l": mustString(bd, "hello"), "r": mustString(bd, "bye")}),
			),
			[]string{".l > .r"},
			list(
				mustBool(bd, false),
				mustBool(bd, true),
				mustBool(bd, true),

				mustBool(bd, false),
				mustBool(bd, true),

				mustBool(bd, true),
				mustBool(bd, true),

				mustBool(bd, false),
				mustBool(bd, true),
			),
		},
		{"greater than or eq", true,
			list(
				// float equality
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustFloat(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 2), "r": mustFloat(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 3.1415), "r": mustFloat(bd, 3)}),

				// int equality
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustInt(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 2), "r": mustInt(bd, 1)}),

				// int~float equality
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 2), "r": mustInt(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 2), "r": mustFloat(bd, 1)}),

				// string equality
				mustObject(bd, map[string]msg.Msg{"l": mustString(bd, "hello"), "r": mustString(bd, "hello")}),
				mustObject(bd, map[string]msg.Msg{"l": mustString(bd, "hello"), "r": mustString(bd, "bye")}),
			),
			[]string{".l >= .r"},
			list(
				mustBool(bd, true),
				mustBool(bd, true),
				mustBool(bd, true),

				mustBool(bd, true),
				mustBool(bd, true),

				mustBool(bd, true),
				mustBool(bd, true),

				mustBool(bd, true),
				mustBool(bd, true),
			),
		},
		{"less than", true,
			list(
				// float equality
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustFloat(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 2), "r": mustFloat(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustFloat(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 3.1415), "r": mustFloat(bd, 3)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 3), "r": mustFloat(bd, 3.1415)}),

				// int equality
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustInt(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 2), "r": mustInt(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustInt(bd, 2)}),

				// int~float equality
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 2), "r": mustInt(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 2), "r": mustFloat(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustInt(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustFloat(bd, 2)}),

				// string equality
				mustObject(bd, map[string]msg.Msg{"l": mustString(bd, "hello"), "r": mustString(bd, "hello")}),
				mustObject(bd, map[string]msg.Msg{"l": mustString(bd, "hello"), "r": mustString(bd, "bye")}),
				mustObject(bd, map[string]msg.Msg{"l": mustString(bd, "bye"), "r": mustString(bd, "hello")}),
			),
			[]string{".l < .r"},
			list(
				mustBool(bd, false),
				mustBool(bd, false),
				mustBool(bd, true),
				mustBool(bd, false),
				mustBool(bd, true),

				mustBool(bd, false),
				mustBool(bd, false),
				mustBool(bd, true),

				mustBool(bd, false),
				mustBool(bd, false),
				mustBool(bd, true),
				mustBool(bd, true),

				mustBool(bd, false),
				mustBool(bd, false),
				mustBool(bd, true),
			),
		},
		{"less than or eq", true,
			list(
				// float equality
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustFloat(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 2), "r": mustFloat(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 3.1415), "r": mustFloat(bd, 3)}),

				// int equality
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustInt(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 2), "r": mustInt(bd, 1)}),

				// int~float equality
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 2), "r": mustInt(bd, 1)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 2), "r": mustFloat(bd, 1)}),

				// string equality
				mustObject(bd, map[string]msg.Msg{"l": mustString(bd, "hello"), "r": mustString(bd, "hello")}),
				mustObject(bd, map[string]msg.Msg{"l": mustString(bd, "hello"), "r": mustString(bd, "bye")}),
			),
			[]string{".l <= .r"},
			list(
				mustBool(bd, true),
				mustBool(bd, false),
				mustBool(bd, false),

				mustBool(bd, true),
				mustBool(bd, false),

				mustBool(bd, false),
				mustBool(bd, false),

				mustBool(bd, true),
				mustBool(bd, false),
			),
		},
		{"NOT logic table", true,
			list(
				mustBool(bd, false),
				mustBool(bd, true),
			),
			[]string{"!."},
			list(
				mustBool(bd, true),
				mustBool(bd, false),
			),
		},
		{"AND logic table", true,
			list(
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, false), "r": mustBool(bd, false)}),
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, false), "r": mustBool(bd, true)}),
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, true), "r": mustBool(bd, false)}),
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, true), "r": mustBool(bd, true)}),
			),
			[]string{".l && .r"},
			list(
				mustBool(bd, false),
				mustBool(bd, false),
				mustBool(bd, false),
				mustBool(bd, true),
			),
		},
		{"OR logic table", true,
			list(
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, false), "r": mustBool(bd, false)}),
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, false), "r": mustBool(bd, true)}),
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, true), "r": mustBool(bd, false)}),
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, true), "r": mustBool(bd, true)}),
			),
			[]string{".l || .r"},
			list(
				mustBool(bd, false),
				mustBool(bd, true),
				mustBool(bd, true),
				mustBool(bd, true),
			),
		},
		{"XOR logic table", true,
			list(
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, false), "r": mustBool(bd, false)}),
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, false), "r": mustBool(bd, true)}),
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, true), "r": mustBool(bd, false)}),
				mustObject(bd, map[string]msg.Msg{"l": mustBool(bd, true), "r": mustBool(bd, true)}),
			),
			[]string{".l != .r"},
			list(
				mustBool(bd, false),
				mustBool(bd, true),
				mustBool(bd, true),
				mustBool(bd, false),
			),
		},
		{"addition", true,
			list(
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustFloat(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustInt(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustString(bd, "hello "), "r": mustString(bd, "world")}),

				// int promotion to float
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustInt(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustFloat(bd, 2)}),

				// int promotion to string
				mustObject(bd, map[string]msg.Msg{"l": mustString(bd, "hello "), "r": mustInt(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustString(bd, "world")}),

				// float promotion to string
				mustObject(bd, map[string]msg.Msg{"l": mustString(bd, "hello "), "r": mustFloat(bd, 2.2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1.1), "r": mustString(bd, "world")}),
			),
			[]string{".l + .r"},
			list(
				mustFloat(bd, 3),
				mustInt(bd, 3),
				mustString(bd, "hello world"),

				mustFloat(bd, 3),
				mustFloat(bd, 3),

				mustString(bd, "hello 2"),
				mustString(bd, "1world"),

				mustString(bd, "hello 2.2"),
				mustString(bd, "1.1world"),
			),
		},

		{"subtraction", true,
			list(
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustFloat(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustInt(bd, 2)}),
				// int promotion to float
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustInt(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustFloat(bd, 2)}),
			),
			[]string{".l - .r"},
			list(
				mustFloat(bd, -1),
				mustInt(bd, -1),
				mustFloat(bd, -1),
				mustFloat(bd, -1),
			),
		},
		{"single int subtraction (negation)", true,
			list(
				mustBool(bd, true),
			),
			[]string{"-1"},
			list(
				mustInt(bd, -1),
			),
		},
		{"single float subtraction (negation)", true,
			list(
				mustBool(bd, true),
			),
			[]string{"-1.2"},
			list(
				mustFloat(bd, -1.2),
			),
		},
		{"division", true,
			list(
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustFloat(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustInt(bd, 2)}),
				// int promotion to float
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustInt(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustFloat(bd, 2)}),
			),
			[]string{".l / .r"},
			list(
				mustFloat(bd, 1.0/2.0),
				mustInt(bd, 1/2),
				mustFloat(bd, 1.0/2.0),
				mustFloat(bd, 1.0/2.0),
			),
		},
		{"division by zero", false, // not strict, we want to skip the divisions by zero
			list(
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustFloat(bd, 0)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustInt(bd, 0)}),
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 1), "r": mustInt(bd, 0)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 1), "r": mustFloat(bd, 0)}),
			),
			[]string{".l / .r"},
			list(),
		},
		{"multiplication", true,
			list(
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 3.5), "r": mustFloat(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 3), "r": mustInt(bd, 2)}),
				// int promotion to float
				mustObject(bd, map[string]msg.Msg{"l": mustFloat(bd, 3.5), "r": mustInt(bd, 2)}),
				mustObject(bd, map[string]msg.Msg{"l": mustInt(bd, 3), "r": mustFloat(bd, 2.5)}),
			),
			[]string{".l * .r"},
			list(
				mustFloat(bd, 7),
				mustInt(bd, 6),
				mustFloat(bd, 7),
				mustFloat(bd, 7.5),
			),
		},

		{"priority of operations", true,
			list(mustBool(bd, true)),
			[]string{"4-1"},
			list(mustInt(bd, 3)),
		},
		{"priority of operations", true,
			list(mustBool(bd, true)),
			[]string{"4.0-(1+4)"},
			list(mustFloat(bd, -1)),
		},
		{"priority of operations", true,
			list(mustBool(bd, true)),
			[]string{"4-1+4"},
			list(mustInt(bd, 7)),
		},
		{"priority of operations", true,
			list(mustBool(bd, true)),
			[]string{"4-(1+4)"},
			list(mustInt(bd, -1)),
		},
		{"priority of operations", true,
			list(mustBool(bd, true)),
			[]string{"4.0-(1+4)"},
			list(mustFloat(bd, -1)),
		},
		{"priority of operations", true,
			list(mustBool(bd, true)),
			[]string{"4-1+4"},
			list(mustInt(bd, 7)),
		},
		{"priority of operations", true,
			list(mustBool(bd, true)),
			[]string{"4.0-1+4"},
			list(mustFloat(bd, 7)),
		},
		{"priority of operations", true,
			list(mustBool(bd, true)),
			[]string{"4/(1*4)"},
			list(mustInt(bd, 1)),
		},
		{"priority of operations", true,
			list(mustBool(bd, true)),
			[]string{"4.0/(1.0*4.0)"},
			list(mustFloat(bd, 1.0)),
		},
		{"priority of operations", true,
			list(mustBool(bd, true)),
			[]string{"4/1*4"},
			list(mustInt(bd, 16)),
		},
		{"priority of operations", true,
			list(mustBool(bd, true)),
			[]string{"4.0/1.0*4.0"},
			list(mustFloat(bd, 16.0)),
		},

		// need way to assert errors
		// {"invalid function call", false,
		// 	list(mustBool(bd, true)),
		// 	"select(.too, .many, .arg)",
		// 	list(mustFloat(bd, 16.0)),
		// },

		{"passthru if true", true,
			list(mustBool(bd, true), mustBool(bd, false), mustBool(bd, true), mustBool(bd, false)),
			[]string{
				`select(., .)`,
				`select(.)`,
				`. | select(.)`,
				`select(select(.))`,
				`select(., select(.))`,
				`select(., select(., .))`,
				`select(.) | select(.) | select(.)`,
			},
			list(mustBool(bd, true), mustBool(bd, true)),
		},

		{"regexp", true,
			list(
				mustObject(bd, map[string]msg.Msg{
					"s":       mustString(bd, "aaaaa"),
					"pattern": mustString(bd, "a"),
				}),
			),
			[]string{"select(regexp(.s, .pattern))"},
			list(
				mustObject(bd, map[string]msg.Msg{
					"s":       mustString(bd, "aaaaa"),
					"pattern": mustString(bd, "a"),
				}),
			),
		},

		{"contains", true,
			list(
				mustObject(bd, map[string]msg.Msg{
					"s":         mustString(bd, "aaaaa"),
					"substring": mustString(bd, "a"),
				}),
			),
			[]string{"select(contains(.s, .substring))"},
			list(
				mustObject(bd, map[string]msg.Msg{
					"s":         mustString(bd, "aaaaa"),
					"substring": mustString(bd, "a"),
				}),
			),
		},

		{"length", true,
			list(
				mustString(bd, "hello world"),
				mustArray(bd,
					mustInt(bd, 1),
					mustInt(bd, 2),
					mustInt(bd, 3),
				),
				mustObject(bd, map[string]msg.Msg{
					"1": mustInt(bd, 1),
					"2": mustInt(bd, 2),
				}),
			),
			[]string{
				"length(.)",
				"length",
				". | length",
			},
			list(
				mustInt(bd, 11),
				mustInt(bd, 3),
				mustInt(bd, 2),
			),
		},

		{"keys", true,
			list(
				mustArray(bd,
					mustString(bd, "lol"),
					mustString(bd, "lol"),
					mustString(bd, "lol"),
				),
				mustObject(bd, map[string]msg.Msg{
					"key1": mustInt(bd, 1),
					"key2": mustInt(bd, 2),
				}),
			),
			[]string{
				"keys(.)",
				"keys",
				". | keys",
			},
			list(
				mustArray(bd,
					mustInt(bd, 0),
					mustInt(bd, 1),
					mustInt(bd, 2),
				),
				mustArray(bd,
					mustString(bd, "key1"),
					mustString(bd, "key2"),
				),
			),
		},

		{"has", true,
			list(
				mustObject(bd, map[string]msg.Msg{
					"s": mustString(bd, "aaaaa"),
				}),
			),
			[]string{
				`has(., "s")`,
				`has("s")`,
				`. | has("s")`,
			},
			list(
				mustBool(bd, true),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, query := range tt.query {
				t.Logf("query=\t%s", query)
				tree, err := grammar.Parse(strings.NewReader(query))
				if err != nil {
					t.Fatal(err)
				}
				vm := mkVM(tree, &vm.Options{
					Strict: tt.strict,
				})

				var got []msg.Msg
				err = vm.Run(bd, ArraySource(tt.input), func(m msg.Msg) error {
					if m == nil {
						panic("derpo")
					}
					got = append(got, m)
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
				if want, got := tt.want, got; !reflect.DeepEqual(want, got) {
					t.Errorf("want=%#v", want)
					t.Errorf(" got=%#v", got)
				}
			}
		})
	}
}

func list(allOfThem ...msg.Msg) []msg.Msg { return allOfThem }

type arraySource struct {
	data []msg.Msg
}

func ArraySource(data []msg.Msg) msg.Source {
	n := len(data)
	i := 0
	return func() (msg.Msg, bool, error) {
		i++
		if i > n {
			return nil, false, nil
		}
		return data[i-1], i <= n, nil
	}
}

func mustObject(bd msg.Builder, obj map[string]msg.Msg) msg.Msg {
	return mustMsg(bd.Object(func(ob msg.ObjectBuilder) error {
		for k, v := range obj {
			err := ob.AddMember(k, func(_ msg.Builder) (msg.Msg, error) {
				return v, nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	}))
}
func mustArray(bd msg.Builder, arr ...msg.Msg) msg.Msg {
	return mustMsg(bd.Array(func(ab msg.ArrayBuilder) error {
		for _, el := range arr {
			err := ab.AddElem(func(_ msg.Builder) (msg.Msg, error) {
				return el, nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	}))
}
func mustString(bd msg.Builder, v string) msg.Msg { return mustMsg(bd.String(v)) }
func mustInt(bd msg.Builder, v int64) msg.Msg     { return mustMsg(bd.Int(v)) }
func mustFloat(bd msg.Builder, v float64) msg.Msg { return mustMsg(bd.Float(v)) }
func mustBool(bd msg.Builder, v bool) msg.Msg     { return mustMsg(bd.Bool(v)) }
func mustNull(bd msg.Builder) msg.Msg             { return mustMsg(bd.Null()) }

func mustMsg(m msg.Msg, err error) msg.Msg {
	if err != nil {
		panic(err)
	}
	return m
}
