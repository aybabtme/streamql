package vm

type Message interface {
	Member(string) (Message, bool)
	Each() ([]Message, bool)
	Range(from, to int) ([]Message, bool)
	Index(int) (Message, bool)
}

type Engine interface {
	Filter(in []Message) (out [][]Message)
}

// TBD

type Source interface {
	Next() (Message, bool)
}

type Sink interface {
	Next() (Message, bool)
}
