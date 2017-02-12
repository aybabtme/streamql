package gomsg

import (
	"testing"

	"reflect"

	"github.com/aybabtme/streamql/lang/spec/msg"
	"github.com/aybabtme/streamql/lang/spec/msg/msgutil"
)

type emitter func(msg.Builder) (msg.Msg, error)

var (
	mkObject = func(v map[string]emitter) emitter {
		return func(bd msg.Builder) (msg.Msg, error) {
			return bd.Object(func(ab msg.ObjectBuilder) error {
				for key, val := range v {
					if err := ab.AddMember(key, val); err != nil {
						return err
					}
				}
				return nil
			})
		}
	}
	mkArray = func(elems ...emitter) emitter {
		return func(bd msg.Builder) (msg.Msg, error) {
			return bd.Array(func(ab msg.ArrayBuilder) error {
				for _, elem := range elems {
					if err := ab.AddElem(elem); err != nil {
						return err
					}
				}
				return nil
			})
		}
	}
	mkString = func(v string) emitter {
		return func(bd msg.Builder) (msg.Msg, error) { return bd.String(v) }
	}
	mkFloat = func(v float64) emitter {
		return func(bd msg.Builder) (msg.Msg, error) { return bd.Float(v) }
	}
	mkInt = func(v int64) emitter {
		return func(bd msg.Builder) (msg.Msg, error) { return bd.Int(v) }
	}
	mkBool = func(v bool) emitter {
		return func(bd msg.Builder) (msg.Msg, error) { return bd.Bool(v) }
	}
	mkNull = func() emitter {
		return func(bd msg.Builder) (msg.Msg, error) { return bd.Null() }
	}
)

func TestBuilder(t *testing.T) {
	tests := []struct {
		name  string
		want  interface{}
		wantT msg.Type
		toMsg func(msg.Builder) (msg.Msg, error)
	}{

		{"create String", "hello world", msg.TypeString, func(bd msg.Builder) (msg.Msg, error) { return bd.String("hello world") }},
		{"create Int", int64(42), msg.TypeInt, func(bd msg.Builder) (msg.Msg, error) { return bd.Int(42) }},
		{"create Float", 3.14159, msg.TypeFloat, func(bd msg.Builder) (msg.Msg, error) { return bd.Float(3.14159) }},
		{"create Bool", true, msg.TypeBool, func(bd msg.Builder) (msg.Msg, error) { return bd.Bool(true) }},
		{"create Bool", false, msg.TypeBool, func(bd msg.Builder) (msg.Msg, error) { return bd.Bool(false) }},
		{"create Null", nil, msg.TypeNull, func(bd msg.Builder) (msg.Msg, error) { return bd.Null() }},

		{
			"create simple Object",
			map[string]interface{}{
				"mkString": "hello world",
				"mkInt":    int64(42),
				"mkFloat":  3.14159,
				"mkBool":   true,
				"mkNull":   nil,
			},
			msg.TypeObject,
			mkObject(map[string]emitter{
				"mkString": mkString("hello world"),
				"mkInt":    mkInt(42),
				"mkFloat":  mkFloat(3.14159),
				"mkBool":   mkBool(true),
				"mkNull":   mkNull(),
			}),
		},

		{
			"create simple Array",
			[]interface{}{"hello world", int64(42), 3.14159, true, false, nil},
			msg.TypeArray,
			mkArray(
				mkString("hello world"),
				mkInt(42),
				mkFloat(3.14159),
				mkBool(true),
				mkBool(false),
				mkNull(),
			),
		},

		{
			"create Object",
			map[string]interface{}{
				"mkString": "hello world",
				"mkInt":    int64(42),
				"mkFloat":  3.14159,
				"mkBool":   true,
				"mkNull":   nil,
				"mkArray": []interface{}{"hello world", int64(42), 3.14159, true, false, nil, map[string]interface{}{
					"mkString": "hello world",
					"mkInt":    int64(42),
					"mkFloat":  3.14159,
					"mkBool":   true,
					"mkNull":   nil,
					"mkArray":  []interface{}{"hello world", int64(42), 3.14159, true, false, nil},
				}},
			},
			msg.TypeObject,
			mkObject(map[string]emitter{
				"mkString": mkString("hello world"),
				"mkInt":    mkInt(42),
				"mkFloat":  mkFloat(3.14159),
				"mkBool":   mkBool(true),
				"mkNull":   mkNull(),
				"mkArray": mkArray(
					mkString("hello world"),
					mkInt(42),
					mkFloat(3.14159),
					mkBool(true),
					mkBool(false),
					mkNull(),
					mkObject(map[string]emitter{
						"mkString": mkString("hello world"),
						"mkInt":    mkInt(42),
						"mkFloat":  mkFloat(3.14159),
						"mkBool":   mkBool(true),
						"mkNull":   mkNull(),
						"mkArray": mkArray(
							mkString("hello world"),
							mkInt(42),
							mkFloat(3.14159),
							mkBool(true),
							mkBool(false),
							mkNull(),
						),
					}),
				),
			}),
		},

		{
			"create Array",
			[]interface{}{"hello world", int64(42), 3.14159, true, false, nil, map[string]interface{}{
				"mkString": "hello world",
				"mkInt":    int64(42),
				"mkFloat":  3.14159,
				"mkBool":   true,
				"mkNull":   nil,
				"mkArray": []interface{}{"hello world", int64(42), 3.14159, true, false, nil, map[string]interface{}{
					"mkString": "hello world",
					"mkInt":    int64(42),
					"mkFloat":  3.14159,
					"mkBool":   true,
					"mkNull":   nil,
					"mkArray":  []interface{}{"hello world", int64(42), 3.14159, true, false, nil},
				}},
			}},
			msg.TypeArray,
			mkArray(
				mkString("hello world"),
				mkInt(42),
				mkFloat(3.14159),
				mkBool(true),
				mkBool(false),
				mkNull(),
				mkObject(map[string]emitter{
					"mkString": mkString("hello world"),
					"mkInt":    mkInt(42),
					"mkFloat":  mkFloat(3.14159),
					"mkBool":   mkBool(true),
					"mkNull":   mkNull(),
					"mkArray": mkArray(
						mkString("hello world"),
						mkInt(42),
						mkFloat(3.14159),
						mkBool(true),
						mkBool(false),
						mkNull(),
						mkObject(map[string]emitter{
							"mkString": mkString("hello world"),
							"mkInt":    mkInt(42),
							"mkFloat":  mkFloat(3.14159),
							"mkBool":   mkBool(true),
							"mkNull":   mkNull(),
							"mkArray": mkArray(
								mkString("hello world"),
								mkInt(42),
								mkFloat(3.14159),
								mkBool(true),
								mkBool(false),
								mkNull(),
							),
						}),
					),
				}),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bd := Build()
			v, err := tt.toMsg(bd)
			if err != nil {
				t.Fatal(err)
			}
			if want, got := tt.wantT, v.Type(); !reflect.DeepEqual(want, got) {
				t.Errorf("want=%#v", want)
				t.Errorf(" got=%#v", got)
			}

			got, err := msgutil.Reveal(v)
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
