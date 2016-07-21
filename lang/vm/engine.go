package vm

import "github.com/aybabtme/streamql/lang/vm/msg"

type Engine interface {
	Filter(src Source, sink Sink)
}

type Source func() (msg.Message, bool)

type Sink func(msg.Message) bool
