package spec

type Source func() (msg Msg, more bool, err error)
type Sink func(Msg) error

type Builder interface {
	Object(func(ObjectBuilder)) Msg
	Array(func(ArrayBuilder)) Msg
	String(string) Msg
	Int(int64) Msg
	Float(float64) Msg
	Bool(bool) Msg
	Null() Msg
}

type ObjectBuilder interface {
	AddMember(string) Builder
}

type ArrayBuilder interface {
	AddElem() Builder
}

type MsgT uint16

const (
	MsgTStart = iota
	MsgTObject
	MsgTArray
	MsgTString
	MsgTInt
	MsgTFloat
	MsgTBool
	MsgTNull
	MsgTEnd
)

type Msg interface {
	Type() MsgT

	MsgObject
	MsgArray
	MsgString
	MsgInt
	MsgFloat
	MsgBool
	MsgNull
}

type MsgObject interface {
	Member(string) Msg
	Keys() []string
}

type MsgArray interface {
	Slice(int64, int64) Source
	Index(int64) Msg
	Len() int64
}

type MsgString interface {
	String() string
}

type MsgInt interface {
	Int() int64
}
type MsgFloat interface {
	Float() float64
}
type MsgBool interface {
	Bool() bool
}
type MsgNull interface {
	IsNull() bool
}
