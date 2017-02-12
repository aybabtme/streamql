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
func (concreteNull) Type() msg.Type                { panic("wrong type") }
func (concreteNull) Member(string) msg.Msg         { panic("wrong type") }
func (concreteNull) Keys() []string                { panic("wrong type") }
func (concreteNull) Slice(int64, int64) msg.Source { panic("wrong type") }
func (concreteNull) Index(int64) msg.Msg           { panic("wrong type") }
func (concreteNull) Len() int64                    { panic("wrong type") }
func (concreteNull) StringVal() string             { panic("wrong type") }
func (concreteNull) IntVal() int64                 { panic("wrong type") }
func (concreteNull) FloatVal() float64             { panic("wrong type") }
func (concreteNull) BoolVal() bool                 { panic("wrong type") }
func (concreteNull) IsNull() bool                  { return true }

var (
	_ internalMsg = (*concreteObj)(nil)
	_ msg.Object  = (*concreteObj)(nil)
)

type concreteObj struct {
	keys    []string
	members map[string]msg.Msg
}

func (co *concreteObj) isGoMsg()                {}
func (co *concreteObj) Type() msg.Type          { return msg.TypeObject }
func (co *concreteObj) IsNull() bool            { return co.members == nil }
func (co *concreteObj) Keys() []string          { return co.keys }
func (co *concreteObj) Member(v string) msg.Msg { return co.members[v] }

func (concreteObj) Slice(int64, int64) msg.Source { panic("wrong type") }
func (concreteObj) Index(int64) msg.Msg           { panic("wrong type") }
func (concreteObj) Len() int64                    { panic("wrong type") }
func (concreteObj) StringVal() string             { panic("wrong type") }
func (concreteObj) IntVal() int64                 { panic("wrong type") }
func (concreteObj) FloatVal() float64             { panic("wrong type") }
func (concreteObj) BoolVal() bool                 { panic("wrong type") }

var (
	_ internalMsg = (*concreteArr)(nil)
	_ msg.Array   = (*concreteArr)(nil)
)

type concreteArr struct {
	elems []msg.Msg
}

func (ca *concreteArr) isGoMsg()              {}
func (ca *concreteArr) Type() msg.Type        { return msg.TypeArray }
func (ca concreteArr) IsNull() bool           { return ca.elems == nil }
func (ca *concreteArr) Len() int64            { return int64(len(ca.elems)) }
func (ca *concreteArr) Index(i int64) msg.Msg { return ca.elems[i] }
func (ca *concreteArr) Slice(from, to int64) msg.Source {
	i := from - 1
	return func() (msg.Msg, bool, error) { i++; return ca.elems[i], i+1 < to, nil }
}

func (concreteArr) Member(string) msg.Msg { panic("wrong type") }
func (concreteArr) Keys() []string        { panic("wrong type") }
func (concreteArr) StringVal() string     { panic("wrong type") }
func (concreteArr) IntVal() int64         { panic("wrong type") }
func (concreteArr) FloatVal() float64     { panic("wrong type") }
func (concreteArr) BoolVal() bool         { panic("wrong type") }

var (
	_ internalMsg = concreteStr("hello")
	_ msg.String  = concreteStr("hello")
)

type concreteStr string

func (concreteStr) isGoMsg()             {}
func (concreteStr) Type() msg.Type       { return msg.TypeString }
func (concreteStr) IsNull() bool         { return false }
func (cs concreteStr) StringVal() string { return string(cs) }

func (concreteStr) Member(string) msg.Msg         { panic("wrong type") }
func (concreteStr) Keys() []string                { panic("wrong type") }
func (concreteStr) Slice(int64, int64) msg.Source { panic("wrong type") }
func (concreteStr) Index(int64) msg.Msg           { panic("wrong type") }
func (concreteStr) Len() int64                    { panic("wrong type") }
func (concreteStr) IntVal() int64                 { panic("wrong type") }
func (concreteStr) FloatVal() float64             { panic("wrong type") }
func (concreteStr) BoolVal() bool                 { panic("wrong type") }

var (
	_ internalMsg = concreteInt(1)
	_ msg.Int     = concreteInt(1)
)

type concreteInt int64

func (concreteInt) isGoMsg()         {}
func (concreteInt) Type() msg.Type   { return msg.TypeInt }
func (concreteInt) IsNull() bool     { return false }
func (ci concreteInt) IntVal() int64 { return int64(ci) }

func (concreteInt) Member(string) msg.Msg         { panic("wrong type") }
func (concreteInt) Keys() []string                { panic("wrong type") }
func (concreteInt) Slice(int64, int64) msg.Source { panic("wrong type") }
func (concreteInt) Index(int64) msg.Msg           { panic("wrong type") }
func (concreteInt) Len() int64                    { panic("wrong type") }
func (concreteInt) StringVal() string             { panic("wrong type") }
func (concreteInt) FloatVal() float64             { panic("wrong type") }
func (concreteInt) BoolVal() bool                 { panic("wrong type") }

var (
	_ internalMsg = concreteFloat(3.14159)
	_ msg.Float   = concreteFloat(3.14159)
)

type concreteFloat float64

func (concreteFloat) isGoMsg()             {}
func (concreteFloat) Type() msg.Type       { return msg.TypeFloat }
func (concreteFloat) IsNull() bool         { return false }
func (cf concreteFloat) FloatVal() float64 { return float64(cf) }

func (concreteFloat) Member(string) msg.Msg         { panic("wrong type") }
func (concreteFloat) Keys() []string                { panic("wrong type") }
func (concreteFloat) Slice(int64, int64) msg.Source { panic("wrong type") }
func (concreteFloat) Index(int64) msg.Msg           { panic("wrong type") }
func (concreteFloat) Len() int64                    { panic("wrong type") }
func (concreteFloat) StringVal() string             { panic("wrong type") }
func (concreteFloat) IntVal() int64                 { panic("wrong type") }
func (concreteFloat) BoolVal() bool                 { panic("wrong type") }

var (
	_ internalMsg = concreteBool(true)
	_ msg.Bool    = concreteBool(false)
)

type concreteBool bool

func (concreteBool) isGoMsg()         {}
func (concreteBool) Type() msg.Type   { return msg.TypeBool }
func (concreteBool) IsNull() bool     { return false }
func (cb concreteBool) BoolVal() bool { return bool(cb) }

func (concreteBool) Member(string) msg.Msg         { panic("wrong type") }
func (concreteBool) Keys() []string                { panic("wrong type") }
func (concreteBool) Slice(int64, int64) msg.Source { panic("wrong type") }
func (concreteBool) Index(int64) msg.Msg           { panic("wrong type") }
func (concreteBool) Len() int64                    { panic("wrong type") }
func (concreteBool) StringVal() string             { panic("wrong type") }
func (concreteBool) IntVal() int64                 { panic("wrong type") }
func (concreteBool) FloatVal() float64             { panic("wrong type") }
