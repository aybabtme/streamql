package gomsg

import "github.com/aybabtme/streamql/lang/spec/msg"

var _ msg.Msg = (internalMsg)(nil)

type internalMsg interface {
	msg.Msg
	isGoMsg()
}

var (
	_ internalMsg = (*concreteNull)(nil)
	_ msg.Null    = (*concreteNull)(nil)
)

type concreteNull struct{}

func (concreteNull) isGoMsg()                      {}
func (concreteNull) Type() msg.Type                { return msg.TypeNull }
func (concreteNull) Member(string) msg.Msg         { panic("undefined on Null") }
func (concreteNull) Keys() []string                { panic("undefined on Null") }
func (concreteNull) Slice(int64, int64) msg.Source { panic("undefined on Null") }
func (concreteNull) Index(int64) msg.Msg           { panic("undefined on Null") }
func (concreteNull) Len() int64                    { panic("undefined on Null") }
func (concreteNull) StringVal() string             { panic("undefined on Null") }
func (concreteNull) IntVal() int64                 { panic("undefined on Null") }
func (concreteNull) FloatVal() float64             { panic("undefined on Null") }
func (concreteNull) BoolVal() bool                 { panic("undefined on Null") }
func (concreteNull) IsNull() bool                  { return true }

var (
	_ internalMsg = (*concreteObj)(nil)
	_ msg.Object  = (*concreteObj)(nil)
)

type concreteObj struct {
	keys    []string
	members map[string]msg.Msg
}

func (concreteObj) isGoMsg()                    {}
func (co *concreteObj) Type() msg.Type          { return msg.TypeObject }
func (co *concreteObj) IsNull() bool            { return co.members == nil }
func (co *concreteObj) Keys() []string          { return co.keys }
func (co *concreteObj) Member(v string) msg.Msg { return co.members[v] }

func (concreteObj) Slice(int64, int64) msg.Source { panic("undefined on Object") }
func (concreteObj) Index(int64) msg.Msg           { panic("undefined on Object") }
func (concreteObj) Len() int64                    { panic("undefined on Object") }
func (concreteObj) StringVal() string             { panic("undefined on Object") }
func (concreteObj) IntVal() int64                 { panic("undefined on Object") }
func (concreteObj) FloatVal() float64             { panic("undefined on Object") }
func (concreteObj) BoolVal() bool                 { panic("undefined on Object") }

var (
	_ internalMsg = (*concreteArr)(nil)
	_ msg.Array   = (*concreteArr)(nil)
)

type concreteArr struct {
	elems []msg.Msg
}

func (concreteArr) isGoMsg()                  {}
func (ca *concreteArr) Type() msg.Type        { return msg.TypeArray }
func (ca concreteArr) IsNull() bool           { return ca.elems == nil }
func (ca *concreteArr) Len() int64            { return int64(len(ca.elems)) }
func (ca *concreteArr) Index(i int64) msg.Msg { return ca.elems[i] }
func (ca *concreteArr) Slice(from, to int64) msg.Source {
	i := from
	n := to
	return func() (msg.Msg, bool, error) {
		if i >= n {
			return nil, false, nil
		}
		i++
		return ca.elems[i-1], true, nil
	}
}

func (concreteArr) Member(string) msg.Msg { panic("undefined on Array") }
func (concreteArr) Keys() []string        { panic("undefined on Array") }
func (concreteArr) StringVal() string     { panic("undefined on Array") }
func (concreteArr) IntVal() int64         { panic("undefined on Array") }
func (concreteArr) FloatVal() float64     { panic("undefined on Array") }
func (concreteArr) BoolVal() bool         { panic("undefined on Array") }

var (
	_ internalMsg = concreteStr("hello")
	_ msg.String  = concreteStr("hello")
)

type concreteStr string

func (concreteStr) isGoMsg()             {}
func (concreteStr) Type() msg.Type       { return msg.TypeString }
func (concreteStr) IsNull() bool         { return false }
func (cs concreteStr) StringVal() string { return string(cs) }

func (concreteStr) Member(string) msg.Msg         { panic("undefined on String") }
func (concreteStr) Keys() []string                { panic("undefined on String") }
func (concreteStr) Slice(int64, int64) msg.Source { panic("undefined on String") }
func (concreteStr) Index(int64) msg.Msg           { panic("undefined on String") }
func (concreteStr) Len() int64                    { panic("undefined on String") }
func (concreteStr) IntVal() int64                 { panic("undefined on String") }
func (concreteStr) FloatVal() float64             { panic("undefined on String") }
func (concreteStr) BoolVal() bool                 { panic("undefined on String") }

var (
	_ internalMsg = concreteInt(1)
	_ msg.Int     = concreteInt(1)
)

type concreteInt int64

func (concreteInt) isGoMsg()         {}
func (concreteInt) Type() msg.Type   { return msg.TypeInt }
func (concreteInt) IsNull() bool     { return false }
func (ci concreteInt) IntVal() int64 { return int64(ci) }

func (concreteInt) Member(string) msg.Msg         { panic("undefined on Int") }
func (concreteInt) Keys() []string                { panic("undefined on Int") }
func (concreteInt) Slice(int64, int64) msg.Source { panic("undefined on Int") }
func (concreteInt) Index(int64) msg.Msg           { panic("undefined on Int") }
func (concreteInt) Len() int64                    { panic("undefined on Int") }
func (concreteInt) StringVal() string             { panic("undefined on Int") }
func (concreteInt) FloatVal() float64             { panic("undefined on Int") }
func (concreteInt) BoolVal() bool                 { panic("undefined on Int") }

var (
	_ internalMsg = concreteFloat(3.14159)
	_ msg.Float   = concreteFloat(3.14159)
)

type concreteFloat float64

func (concreteFloat) isGoMsg()             {}
func (concreteFloat) Type() msg.Type       { return msg.TypeFloat }
func (concreteFloat) IsNull() bool         { return false }
func (cf concreteFloat) FloatVal() float64 { return float64(cf) }

func (concreteFloat) Member(string) msg.Msg         { panic("undefined on Float") }
func (concreteFloat) Keys() []string                { panic("undefined on Float") }
func (concreteFloat) Slice(int64, int64) msg.Source { panic("undefined on Float") }
func (concreteFloat) Index(int64) msg.Msg           { panic("undefined on Float") }
func (concreteFloat) Len() int64                    { panic("undefined on Float") }
func (concreteFloat) StringVal() string             { panic("undefined on Float") }
func (concreteFloat) IntVal() int64                 { panic("undefined on Float") }
func (concreteFloat) BoolVal() bool                 { panic("undefined on Float") }

var (
	_ internalMsg = concreteBool(true)
	_ msg.Bool    = concreteBool(false)
)

type concreteBool bool

func (concreteBool) isGoMsg()         {}
func (concreteBool) Type() msg.Type   { return msg.TypeBool }
func (concreteBool) IsNull() bool     { return false }
func (cb concreteBool) BoolVal() bool { return bool(cb) }

func (concreteBool) Member(string) msg.Msg         { panic("undefined on Bool") }
func (concreteBool) Keys() []string                { panic("undefined on Bool") }
func (concreteBool) Slice(int64, int64) msg.Source { panic("undefined on Bool") }
func (concreteBool) Index(int64) msg.Msg           { panic("undefined on Bool") }
func (concreteBool) Len() int64                    { panic("undefined on Bool") }
func (concreteBool) StringVal() string             { panic("undefined on Bool") }
func (concreteBool) IntVal() int64                 { panic("undefined on Bool") }
func (concreteBool) FloatVal() float64             { panic("undefined on Bool") }
