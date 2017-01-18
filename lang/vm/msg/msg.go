package msg

type Message interface {
	Member(string) (Message, bool)
	Each() ([]Message, bool)
	Range(from, to int) ([]Message, bool)
	Index(int) (Message, bool)
}

// Basic types

func String(v string) Message {
	return stringMsg(v)
}

type stringMsg string

func (stringMsg) Member(string) (Message, bool)        { return nil, false }
func (stringMsg) Each() ([]Message, bool)              { return nil, false }
func (stringMsg) Range(from, to int) ([]Message, bool) { return nil, false }
func (stringMsg) Index(int) (Message, bool)            { return nil, false }

func Bool(v bool) Message {
	return boolMsg(v)
}

type boolMsg bool

func (boolMsg) Member(string) (Message, bool)        { return nil, false }
func (boolMsg) Each() ([]Message, bool)              { return nil, false }
func (boolMsg) Range(from, to int) ([]Message, bool) { return nil, false }
func (boolMsg) Index(int) (Message, bool)            { return nil, false }

func Number(v float64) Message {
	return float64Msg(v)
}

type float64Msg float64

func (float64Msg) Member(string) (Message, bool)        { return nil, false }
func (float64Msg) Each() ([]Message, bool)              { return nil, false }
func (float64Msg) Range(from, to int) ([]Message, bool) { return nil, false }
func (float64Msg) Index(int) (Message, bool)            { return nil, false }
