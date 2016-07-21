package msg

type Message interface {
	Member(string) (Message, bool)
	Each() ([]Message, bool)
	Range(from, to int) ([]Message, bool)
	Index(int) (Message, bool)
}
