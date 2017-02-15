package spec

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aybabtme/streamql/lang/spec/msg"
	"github.com/aybabtme/streamql/lang/spec/msg/gomsg"
)

func TestVM(t *testing.T) {
	bd := gomsg.Build()

	tests := []struct {
		name  string
		input []msg.Msg
		query string
		want  []msg.Msg
	}{
		{"passthru of nothing",
			list(),
			"",
			list(),
		},
		{"passthru of nothing",
			list(),
			".",
			list(),
		},
		{"passthru",
			list(mustBool(bd, true)),
			"",
			list(mustBool(bd, true)),
		},
		{"passthru",
			list(mustBool(bd, true)),
			".",
			list(mustBool(bd, true)),
		},
		{"passthru if true",
			list(mustBool(bd, true), mustBool(bd, false), mustBool(bd, true), mustBool(bd, false)),
			"select(.)",
			list(mustBool(bd, true), mustBool(bd, true)),
		},
		{"explode",
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
			".[]",
			list(
				mustInt(bd, 1),
				mustInt(bd, 2),
				mustInt(bd, 3),
				mustInt(bd, 4),
				mustInt(bd, 5),
				mustInt(bd, 6),
			),
		},
		{"explode",
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
			".[]",
			list(
				mustInt(bd, 1),
				mustInt(bd, 2),
				mustInt(bd, 3),
				mustInt(bd, 4),
				mustInt(bd, 5),
				mustInt(bd, 6),
			),
		},
		{"explode recursively",
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
			".[][]",
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
		{"explode parts skip first",
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
			".[1:]",
			list(
				mustInt(bd, 2),
				mustInt(bd, 3),
				mustInt(bd, 5),
				mustInt(bd, 6),
			),
		},
		{"explode parts skip last",
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
			".[:2]",
			list(
				mustInt(bd, 1),
				mustInt(bd, 2),
				mustInt(bd, 4),
				mustInt(bd, 5),
			),
		},
		{"explode parts skip first and last",
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
			".[1:2]",
			list(
				mustInt(bd, 2),
				mustInt(bd, 5),
			),
		},
		{"index into an object",
			list(
				mustObject(bd, map[string]msg.Msg{
					"hello": mustString(bd, "world"),
				}),
			),
			".hello",
			list(
				mustString(bd, "world"),
			),
		},
		{"index into recursively into an object",
			list(
				mustObject(bd, map[string]msg.Msg{
					"hello": mustObject(bd, map[string]msg.Msg{
						"world": mustFloat(bd, 3.14159),
					}),
				}),
			),
			".hello.world",
			list(
				mustFloat(bd, 3.14159),
			),
		},
		{"maybe index into recursively into an object",
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
			".hello.world | select(. > 4.0)",
			list(
				mustFloat(bd, 2*3.14159),
			),
		},
		{"equality",
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
			".l == .r",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(strings.NewReader(tt.query))
			if err != nil {
				t.Fatal(err)
			}
			vm := &ASTInterpreter{tree: ast}

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
