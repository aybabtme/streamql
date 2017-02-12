package msg

type Source func() (msg Msg, more bool, err error)
type Sink func(Msg) error

type Builder interface {
	IsOwnType(Msg) bool

	Object(func(ObjectBuilder) error) (Msg, error)
	Array(func(ArrayBuilder) error) (Msg, error)
	String(string) (Msg, error)
	Int(int64) (Msg, error)
	Float(float64) (Msg, error)
	Bool(bool) (Msg, error)
	Null() (Msg, error)
}

type ObjectBuilder interface {
	AddMember(string, func(Builder) (Msg, error)) error
}

type ArrayBuilder interface {
	AddElem(func(Builder) (Msg, error)) error
}

type Type uint16

const (
	TypeObject = iota
	TypeArray
	TypeString
	TypeInt
	TypeFloat
	TypeBool
	TypeNull
)

type Msg interface {
	Type() Type

	Object
	Array
	String
	Int
	Float
	Bool
	Null
}

type Object interface {
	Member(string) Msg
	Keys() []string
}

type Array interface {
	Slice(int64, int64) Source
	Index(int64) Msg
	Len() int64
}

type String interface {
	StringVal() string
}

type Int interface {
	IntVal() int64
}
type Float interface {
	FloatVal() float64
}
type Bool interface {
	BoolVal() bool
}
type Null interface {
	IsNull() bool
}
