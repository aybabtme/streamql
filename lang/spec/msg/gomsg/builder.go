package gomsg

import (
	"fmt"

	"github.com/aybabtme/streamql/lang/spec/msg"
)

// Build returns a msg.Builder that can create msg.Msg of the
// gomsg kind.
func Build() msg.Builder { return new(baseBuilder) }

var (
	_ msg.Builder       = (*baseBuilder)(nil)
	_ msg.ObjectBuilder = (*objBuilder)(nil)
	_ msg.ArrayBuilder  = (*arrBuilder)(nil)
)

type baseBuilder struct {
	obj *objBuilder
	arr *arrBuilder
}

func (sb *baseBuilder) String(v string) (msg.Msg, error) { return concreteStr(v), nil }
func (sb *baseBuilder) Int(v int64) (msg.Msg, error)     { return concreteInt(v), nil }
func (sb *baseBuilder) Float(v float64) (msg.Msg, error) { return concreteFloat(v), nil }
func (sb *baseBuilder) Bool(v bool) (msg.Msg, error)     { return concreteBool(v), nil }
func (sb *baseBuilder) Null() (msg.Msg, error)           { return concreteNull{}, nil }

func (sb *baseBuilder) Convert(from msg.Msg) (msg.Msg, error) {
	switch from.(type) {
	case concreteNull, *concreteNull,
		*concreteObj,
		*concreteArr,
		concreteStr, *concreteStr,
		concreteInt, *concreteInt,
		concreteFloat, *concreteFloat,
		concreteBool, *concreteBool:
		return from, nil
	}
	switch from.Type() {
	case msg.TypeObject:
		return sb.Object(func(ob msg.ObjectBuilder) error {
			fromOb := from.(msg.Object)
			for _, k := range fromOb.Keys() {
				v := fromOb.Member(k)
				return ob.AddMember(
					k,
					func(bb msg.Builder) (msg.Msg, error) {
						return bb.Convert(v)
					},
				)
			}
			return nil
		})
	case msg.TypeArray:
		return sb.Array(func(ab msg.ArrayBuilder) error {
			fromAr := from.(msg.Array)
			iter := fromAr.Slice(0, fromAr.Len())
			for {
				m, more, err := iter()
				if err != nil {
					return err
				}
				if !more {
					return nil
				}
				err = ab.AddElem(func(bb msg.Builder) (msg.Msg, error) {
					return bb.Convert(m)
				})
				if err != nil {
					return err
				}
			}
		})

	case msg.TypeString:
		return sb.String(from.StringVal())
	case msg.TypeInt:
		return sb.Int(from.IntVal())
	case msg.TypeFloat:
		return sb.Float(from.FloatVal())
	case msg.TypeBool:
		return sb.Bool(from.BoolVal())
	case msg.TypeNull:
		return sb.Null()

	default:
		return nil, fmt.Errorf("invalid message type: %v", from.Type())
	}
}

func (sb *baseBuilder) Object(fn func(msg.ObjectBuilder) error) (msg.Msg, error) {
	obj := new(objBuilder)
	if err := fn(obj); err != nil {
		return nil, err
	}
	return &concreteObj{keys: obj.keys, members: obj.members}, nil
}
func (sb *baseBuilder) Array(fn func(msg.ArrayBuilder) error) (msg.Msg, error) {
	arr := new(arrBuilder)
	if err := fn(arr); err != nil {
		return nil, err
	}
	return &concreteArr{elems: arr.elems}, nil
}

type objBuilder struct {
	keys    []string
	members map[string]msg.Msg
}

func (ob *objBuilder) AddMember(key string, msgFn func(msg.Builder) (msg.Msg, error)) error {
	v, err := msgFn(new(baseBuilder))
	if err != nil {
		return err
	}
	if ob.members == nil {
		ob.members = map[string]msg.Msg{
			key: v,
		}
		ob.keys = append(ob.keys, key)
	} else {
		_, ok := ob.members[key]
		if !ok {
			ob.keys = append(ob.keys, key)
		}
		ob.members[key] = v
	}
	return nil
}

type arrBuilder struct {
	elems []msg.Msg
}

func (ab *arrBuilder) AddElem(msgFn func(msg.Builder) (msg.Msg, error)) error {
	v, err := msgFn(new(baseBuilder))
	if err != nil {
		return err
	}
	ab.elems = append(ab.elems, v)
	return nil
}
