package token

//go:generate stringer -type=Token
type Token int

const (
	// Special tokens
	Illegal Token = iota
	EOS
	Whitespace

	// Literals
	InlineString
	Integer

	// MiscCharacters
	Comma
	Pipe
	Dot
	LeftBracket
	RightBracket
	Colon
)
