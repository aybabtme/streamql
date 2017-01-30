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
	String
	Integer
	Float

	// MiscCharacters
	Comma
	Pipe
	Dot
	LeftBracket
	RightBracket
	LeftParens
	RightParens
	Quote
	Colon
	PlusSymbol
	MinusSymbol
	MultiplySymbol
	DivideSymbol
)
