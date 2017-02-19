package msgutil

import (
	"fmt"

	"github.com/aybabtme/streamql/lang/msg"
)

// FromGo turns a `v` into a msg.Msg, if `v` is a basic type or a
// slice/map of the same.
func FromGo(dst msg.Builder, v interface{}) (msg.Msg, error) {
	switch vt := v.(type) {
	case int:
		return dst.Int(int64(vt))
	case int8:
		return dst.Int(int64(vt))
	case int16:
		return dst.Int(int64(vt))
	case int32:
		return dst.Int(int64(vt))
	case int64:
		return dst.Int(vt)
	case uint:
		return dst.Int(int64(vt))
	case uint8:
		return dst.Int(int64(vt))
	case uint16:
		return dst.Int(int64(vt))
	case uint32:
		return dst.Int(int64(vt))
	case uint64:
		return dst.Int(int64(vt))

	case float32:
		return dst.Float(float64(vt))
	case float64:
		return dst.Float(vt)

	case bool:
		return dst.Bool(vt)

	case string:
		return dst.String(vt)

	case []interface{}:
		return fromSlice(dst, vt)

	case map[string]interface{}:
		return fromMap(dst, vt)
	}

	return nil, fmt.Errorf("unsupported type for conversion from Go to Msg: %T", v)
}

func fromSlice(dst msg.Builder, elems []interface{}) (msg.Msg, error) {
	return dst.Array(func(ab msg.ArrayBuilder) error {
		for _, v := range elems {
			if err := ab.AddElem(func(b msg.Builder) (msg.Msg, error) {
				return FromGo(b, v)
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func fromMap(dst msg.Builder, kv map[string]interface{}) (msg.Msg, error) {
	return dst.Object(func(ob msg.ObjectBuilder) error {
		for k, v := range kv {
			if err := ob.AddMember(k, func(b msg.Builder) (msg.Msg, error) {
				return FromGo(b, v)
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

// Convert takes a Msg and converts it to the type produced
// by the given Builder. If the Msg is already of a type
// built by the Builder, the message is returned unchanged.
func Convert(dst msg.Builder, in msg.Msg) (msg.Msg, error) {
	if dst.IsOwnType(in) {
		return in, nil
	}
	if in.IsNull() {
		return dst.Null()
	}
	var out msg.Msg
	return out, ActionOnConcreteType(in,
		IfObject(func(val msg.Object) (err error) {
			out, err = convertObject(dst, val)
			return err
		}),
		IfArray(func(val msg.Array) (err error) {
			out, err = convertArray(dst, val)
			return err
		}),
		IfString(func(val msg.String) (err error) {
			out, err = dst.String(val.StringVal())
			return err
		}),
		IfInt(func(val msg.Int) (err error) {
			out, err = dst.Int(val.IntVal())
			return err
		}),
		IfFloat(func(val msg.Float) (err error) {
			out, err = dst.Float(val.FloatVal())
			return err
		}),
		IfBool(func(val msg.Bool) (err error) {
			out, err = dst.Bool(val.BoolVal())
			return err
		}),
	)
}

func convertObject(dst msg.Builder, in msg.Object) (msg.Msg, error) {
	return dst.Object(func(ob msg.ObjectBuilder) error {
		for _, k := range in.Keys() {
			v, ok := in.Member(k)
			if !ok {
				panic("invalid object, .Keys() returned a key that has no Member(key)")
			}
			return ob.AddMember(
				k,
				func(bb msg.Builder) (msg.Msg, error) {
					return Convert(bb, v)
				},
			)
		}
		return nil
	})
}

func convertArray(dst msg.Builder, in msg.Array) (msg.Msg, error) {
	return dst.Array(func(ab msg.ArrayBuilder) error {
		iter := in.Slice(0, in.Len())
		for {
			m, more, err := iter()
			if err != nil {
				return err
			}
			if !more {
				return nil
			}
			err = ab.AddElem(func(bb msg.Builder) (msg.Msg, error) {
				return Convert(bb, m)
			})
			if err != nil {
				return err
			}
		}
	})
}

// Reveal transforms a Msg into a regular Go type. The transformation
// goes like:
//
//     Object -> map[string]interface{}
//     Array -> []interface{}
//     String -> string
//     Int -> int64
//     Float -> float64
//     Bool -> bool
//
// If `IsNull` is true, the returned value is `nil`.
func Reveal(in msg.Msg) (interface{}, error) {
	if in.IsNull() {
		return nil, nil
	}
	var out interface{}
	return out, ActionOnConcreteType(in,
		IfObject(func(val msg.Object) error {
			keys := val.Keys()
			outv := make(map[string]interface{}, len(keys))
			var err error
			for _, k := range keys {
				v, ok := val.Member(k)
				if !ok {
					panic("invalid object, .Keys() returned a key that has no Member(key)")
				}
				outv[k], err = Reveal(v)
				if err != nil {
					return err
				}
			}
			out = outv
			return nil
		}),
		IfArray(func(val msg.Array) error {
			n := val.Len()
			outv := make([]interface{}, 0, int(n))
			iter := val.Slice(0, n)
			for {
				el, more, err := iter()
				if err != nil {
					return err
				}
				if !more {
					out = outv
					return nil
				}
				v, err := Reveal(el)
				if err != nil {
					return err
				}
				outv = append(outv, v)
			}
		}),
		IfString(func(val msg.String) error {
			out = val.StringVal()
			return nil
		}),
		IfInt(func(val msg.Int) error {
			out = val.IntVal()
			return nil
		}),
		IfFloat(func(val msg.Float) error {
			out = val.FloatVal()
			return nil
		}),
		IfBool(func(val msg.Bool) error {
			out = val.BoolVal()
			return nil
		}),
		IfNull(func(val msg.Null) error {
			out = nil
			return nil
		}),
	)
}

// ConcreteType turns a Msg into the specific interface
// of the type.
//
// For instance, a Msg that has `Type() == TypeObject` will
// return a `msg.Object`. You can then `switch .(type)` on
// the returned `interface{}` to get a more specific instance
// to work with.
func ConcreteType(in msg.Msg) interface{} {
	switch in.Type() {
	case msg.TypeObject:
		return scopedObject{in}
	case msg.TypeArray:
		return scopedArray{in}
	case msg.TypeString:
		return scopedString{in}
	case msg.TypeInt:
		return scopedInt{in}
	case msg.TypeFloat:
		return scopedFloat{in}
	case msg.TypeBool:
		return scopedBool{in}
	case msg.TypeNull:
		return scopedNull{in}
	default:
		panic(fmt.Sprintf("bug: unhandled type: %v", in.Type()))
	}
}

// ActionOnConcreteType performs an action with the realized
// type of Msg. If you wish to perform an action for a specific
// type, provide a non-nil callback for it. `nil` callbacks
// are ignored.
func ActionOnConcreteType(in msg.Msg, actions ...actionOnType) error {
	cb := &actionOnTypeOpts{
		ifObject: func(msg.Object) error { return nil },
		ifArray:  func(msg.Array) error { return nil },
		ifString: func(msg.String) error { return nil },
		ifInt:    func(msg.Int) error { return nil },
		ifFloat:  func(msg.Float) error { return nil },
		ifBool:   func(msg.Bool) error { return nil },
		ifNull:   func(msg.Null) error { return nil },
	}
	for _, o := range actions {
		o(cb)
	}

	v := ConcreteType(in)

	switch vt := v.(type) {
	case msg.Object:
		return cb.ifObject(vt)
	case msg.Array:
		return cb.ifArray(vt)
	case msg.String:
		return cb.ifString(vt)
	case msg.Int:
		return cb.ifInt(vt)
	case msg.Float:
		return cb.ifFloat(vt)
	case msg.Bool:
		return cb.ifBool(vt)
	case msg.Null:
		return cb.ifNull(vt)
	default:
		panic(fmt.Sprintf("bug: unhandled type: %#v", vt))
	}
}

type actionOnTypeOpts struct {
	ifObject func(msg.Object) error
	ifArray  func(msg.Array) error
	ifString func(msg.String) error
	ifInt    func(msg.Int) error
	ifFloat  func(msg.Float) error
	ifBool   func(msg.Bool) error
	ifNull   func(msg.Null) error
}

type actionOnType func(*actionOnTypeOpts)

// IfObject performs `action` if the Msg is of TypeObject.
func IfObject(action func(msg.Object) error) actionOnType {
	return func(opts *actionOnTypeOpts) { opts.ifObject = action }
}

// IfArray performs `action` if the Msg is of TypeArray.
func IfArray(action func(msg.Array) error) actionOnType {
	return func(opts *actionOnTypeOpts) { opts.ifArray = action }
}

// IfString performs `action` if the Msg is of TypeString.
func IfString(action func(msg.String) error) actionOnType {
	return func(opts *actionOnTypeOpts) { opts.ifString = action }
}

// IfInt performs `action` if the Msg is of TypeInt.
func IfInt(action func(msg.Int) error) actionOnType {
	return func(opts *actionOnTypeOpts) { opts.ifInt = action }
}

// IfFloat performs `action` if the Msg is of TypeFloat.
func IfFloat(action func(msg.Float) error) actionOnType {
	return func(opts *actionOnTypeOpts) { opts.ifFloat = action }
}

// IfBool performs `action` if the Msg is of TypeBool.
func IfBool(action func(msg.Bool) error) actionOnType {
	return func(opts *actionOnTypeOpts) { opts.ifBool = action }
}

// IfNull performs `action` if the Msg is of TypeNull.
func IfNull(action func(msg.Null) error) actionOnType {
	return func(opts *actionOnTypeOpts) { opts.ifNull = action }
}

type (
	scopedObject struct{ under msg.Object }
	scopedArray  struct{ under msg.Array }
	scopedString struct{ under msg.String }
	scopedInt    struct{ under msg.Int }
	scopedFloat  struct{ under msg.Float }
	scopedBool   struct{ under msg.Msg }
	scopedNull   struct{ under msg.Msg }
)

func (sc scopedObject) Member(m string) (msg.Msg, bool) { return sc.under.Member(m) }
func (sc scopedObject) Keys() []string                  { return sc.under.Keys() }
func (sc scopedArray) Slice(from, to int64) msg.Source  { return sc.under.Slice(from, to) }
func (sc scopedArray) Index(i int64) msg.Msg            { return sc.under.Index(i) }
func (sc scopedArray) Len() int64                       { return sc.under.Len() }
func (sc scopedString) StringVal() string               { return sc.under.StringVal() }
func (sc scopedInt) IntVal() int64                      { return sc.under.IntVal() }
func (sc scopedFloat) FloatVal() float64                { return sc.under.FloatVal() }
func (sc scopedBool) BoolVal() bool                     { return sc.under.BoolVal() }
func (sc scopedNull) IsNull() bool                      { return sc.under.IsNull() }
