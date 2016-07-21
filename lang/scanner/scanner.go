package scanner

import (
	"bufio"
	"bytes"
	"io"

	"github.com/aybabtme/streamql/lang/token"
)

type Scanner struct {
	r *bufio.Reader
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) Scan() (tok token.Token, lit string, err error) {
	ch, err := s.read()
	if err != nil {
		return s.eosIfEOF(token.Illegal, "", err)
	}
	var rule func() (token.Token, string, error)
	switch {

	case isNonZeroDigit(ch) || isZeroDigit(ch):
		rule = s.scanInteger

	case isLetter(ch) || isEscapeCharacter(ch):
		rule = s.scanFieldIdentifier

	case isWhitespace(ch):
		rule = s.scanWhitespace

	case isColon(ch),
		isRightBracket(ch),
		isLeftBracket(ch),
		isDot(ch),
		isPipe(ch),
		isComma(ch):

		rule = s.scanKeyword

	default:
		return token.Illegal, string(ch), err
	}
	s.unread()
	tok, lit, err = rule()
	return s.eosIfEOF(tok, lit, err)
}

// rules

func (s *Scanner) scanWhitespace() (token.Token, string, error) {
	buf := bytes.NewBuffer(nil)
	for {
		ch, err := s.read()
		if err != nil && err != io.EOF {
			return s.eosIfEOF(token.Illegal, buf.String(), err)
		}
		if !isWhitespace(ch) {
			s.unread()
			return token.Whitespace, buf.String(), nil
		}
		buf.WriteRune(ch)
	}
}

func (s *Scanner) scanFieldIdentifier() (token.Token, string, error) {
	// inline string
	buf := bytes.NewBuffer(nil)

	ch, err := s.read()
	if err != nil {
		return s.eosIfEOF(token.Illegal, buf.String(), err)
	}

	// read the first Letter | EscapedKeyword
	switch {
	case isLetter(ch):
	case isEscapeCharacter(ch):
		ch, err = s.read()
		if err != nil {
			return s.eosIfEOF(token.Illegal, buf.String(), err)
		}
		if !isEscapedKeyword(ch) {
			return s.eosIfEOF(token.Illegal, buf.String(), err)
		}
	default:
		return s.eosIfEOF(token.Illegal, buf.String(), err)
	}
	buf.WriteRune(ch)

	// then read subsequent Letter | Digit | EscapedKeyword
	for {
		ch, err := s.read()
		switch err {
		case io.EOF:
			return token.InlineString, buf.String(), nil
		case nil: // continue
		default:
			return s.eosIfEOF(token.Illegal, buf.String(), err)
		}
		switch {
		case isDigit(ch): // accept digits
		case isLetter(ch):
		case isEscapeCharacter(ch):
			ch, err = s.read() // read past the \
			if err != nil {
				return s.eosIfEOF(token.Illegal, buf.String(), err)
			}
			if !isEscapedKeyword(ch) {
				return s.eosIfEOF(token.Illegal, buf.String(), err)
			}
		default:
			// past the end of what's a valid InlineString
			s.unread()
			return token.InlineString, buf.String(), nil
		}
		buf.WriteRune(ch)
	}
}

func (s *Scanner) scanInteger() (token.Token, string, error) {
	buf := bytes.NewBuffer(nil)

	ch, err := s.read()
	if err != nil {
		return s.eosIfEOF(token.Illegal, buf.String(), err)
	}

	buf.WriteRune(ch)

	// if first digit is zero, it should be only digit
	if isZeroDigit(ch) {
		ch2, err := s.read()
		switch err {
		case io.EOF:
			return token.Integer, buf.String(), nil
		case nil: // continue
		default:
			return s.eosIfEOF(token.Illegal, buf.String(), err)
		}
		s.unread()
		if isDigit(ch2) { //
			return s.eosIfEOF(token.Illegal, buf.String(), err)
		}
		return token.Integer, buf.String(), nil
	}

	// then read subsequent Letter | Digit | EscapedKeyword
	for {
		ch, err := s.read()
		switch err {
		case io.EOF:
			return token.Integer, buf.String(), nil
		case nil: // continue
		default:
			return s.eosIfEOF(token.Illegal, buf.String(), err)
		}
		if !isDigit(ch) {
			// past the end of what's a valid Integer
			s.unread()
			return token.Integer, buf.String(), nil
		}
		buf.WriteRune(ch)
	}
}

func (s *Scanner) scanKeyword() (token.Token, string, error) {
	ch, err := s.read()
	if err != nil {
		return s.eosIfEOF(token.Illegal, "", err)
	}
	switch {
	case isColon(ch):
		return token.Colon, string(ch), nil
	case isRightBracket(ch):
		return token.RightBracket, string(ch), nil
	case isLeftBracket(ch):
		return token.LeftBracket, string(ch), nil
	case isDot(ch):
		return token.Dot, string(ch), nil
	case isPipe(ch):
		return token.Pipe, string(ch), nil
	case isComma(ch):
		return token.Comma, string(ch), nil
	default:
		return s.eosIfEOF(token.Illegal, "", err)
	}
}

// lexicon's first set

func isWhitespace(ch rune) bool      { return ch == ' ' || ch == '\t' || ch == '\n' }
func isZeroDigit(ch rune) bool       { return ch == '0' }
func isNonZeroDigit(ch rune) bool    { return ch > '0' && ch <= '9' }
func isComma(ch rune) bool           { return (ch == ',') }
func isPipe(ch rune) bool            { return (ch == '|') }
func isDot(ch rune) bool             { return (ch == '.') }
func isLeftBracket(ch rune) bool     { return (ch == '[') }
func isRightBracket(ch rune) bool    { return (ch == ']') }
func isColon(ch rune) bool           { return (ch == ':') }
func isEscapeCharacter(ch rune) bool { return (ch == '\\') }
func isLetter(ch rune) bool          { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }

func isDigit(ch rune) bool { return isZeroDigit(ch) || isNonZeroDigit(ch) }

func isEscapedKeyword(ch rune) bool {
	return isComma(ch) ||
		isPipe(ch) ||
		isDot(ch) ||
		isLeftBracket(ch) ||
		isRightBracket(ch) ||
		isColon(ch) ||
		isWhitespace(ch) || // can insert whitespace in strings
		isEscapeCharacter(ch) // need to be able to \ the \ symbol (\\)
}

// helpers

func (s *Scanner) read() (rune, error) {
	ch, _, err := s.r.ReadRune()
	return ch, err
}

func (s *Scanner) unread() { _ = s.r.UnreadRune() }

func (s *Scanner) eosIfEOF(tok token.Token, lit string, err error) (token.Token, string, error) {
	if err == io.EOF {
		tok = token.EOS
		lit = ""
	}
	if err != nil {
		tok = token.Illegal
	}
	return tok, lit, err
}
