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

	case isNegativeSign(ch) || isNonZeroDigit(ch) || isZeroDigit(ch):
		rule = s.scanNumber

	case isLetter(ch) || isEscapeCharacter(ch):
		rule = s.scanInlineString

	case isQuote(ch):
		rule = s.scanString

	case isWhitespace(ch):
		rule = s.scanWhitespace

	case isComma(ch),
		isPipe(ch),
		isDot(ch),
		isRightBracket(ch),
		isLeftBracket(ch),
		isRightParens(ch),
		isLeftParens(ch),
		isQuote(ch),
		isColon(ch):

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

func (s *Scanner) scanInlineString() (token.Token, string, error) {
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
		buf.WriteRune(ch)
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
			buf.WriteRune(ch)
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

func (s *Scanner) scanString() (token.Token, string, error) {
	buf := bytes.NewBuffer(nil)
	ch, err := s.read()
	if err != nil {
		return s.eosIfEOF(token.Illegal, buf.String(), err)
	}
	if !isQuote(ch) {
		return s.eosIfEOF(token.Illegal, buf.String(), err)
	}
	buf.WriteRune(ch)

	// read subsequent Letter | Digit | Escaped quotes & whitespace
	for {
		ch, err := s.read()
		switch err {
		case io.EOF:
			// incomplete string, missing closing quote
			return s.eosIfEOF(token.Illegal, buf.String(), nil)
		case nil: // continue
		default:
			return s.eosIfEOF(token.Illegal, buf.String(), err)
		}
		switch {
		case isStringCharacter(ch):
		case isEscapeCharacter(ch):
			buf.WriteRune(ch)
			ch, err = s.read() // read past the \
			if err != nil {
				return s.eosIfEOF(token.Illegal, buf.String(), err)
			}
			if !isQuote(ch) && !isEscapeCharacter(ch) && !isControlCode(ch) {
				return s.eosIfEOF(token.Illegal, buf.String(), err)
			}
		case isQuote(ch):
			// we're done reading the string
			buf.WriteRune(ch)
			return token.String, buf.String(), nil
		default:
			return s.eosIfEOF(token.Illegal, buf.String(), err)
		}
		buf.WriteRune(ch)
	}

}

func (s *Scanner) scanNumber() (token.Token, string, error) {
	buf := bytes.NewBuffer(nil)

	tok, lit, err := s.scanInteger()
	switch err {
	case io.EOF:
		return tok, lit, nil
	case nil: // continue
	default:
		return s.eosIfEOF(token.Illegal, lit, err)
	}
	buf.WriteString(lit)

	// check if there's a decimal or scientific part
	ch, err := s.read()
	switch err {
	case io.EOF:
		s.unread()
		return tok, lit, err
	case nil: // continue
	default:
		return s.eosIfEOF(token.Illegal, buf.String(), err)
	}

	switch {
	case isDot(ch):
		buf.WriteRune(ch)
		// read the decimal part
		err = s.scanDigits(buf)
		switch err {
		case io.EOF:
			return token.Float, buf.String(), nil
		case nil: // continue
		default:
			return s.eosIfEOF(token.Illegal, buf.String(), err)
		}

	case isScientificExponent(ch): // continue
		s.unread() // we'll scan it back

	default:
		// there was no decimal part
		s.unread()
		return token.Integer, buf.String(), nil
	}

	// if we're still scanning a number, it should be
	// a scientific exponent part
	ch, err = s.read()
	switch err {
	case io.EOF:
		s.unread() // no exponent part
		return token.Float, buf.String(), nil
	case nil: // continue
	default:
		return s.eosIfEOF(token.Illegal, buf.String(), err)
	}

	if !isScientificExponent(ch) {
		s.unread() // no exponent part
		return token.Float, buf.String(), nil
	}

	// we have an exponent part
	buf.WriteRune(ch)
	tok, lit, err = s.scanInteger() // don't care for token
	if err != nil && err != io.EOF {
		return s.eosIfEOF(token.Illegal, buf.String(), err)
	}
	buf.WriteString(lit)
	if tok != token.Integer {
		return s.eosIfEOF(token.Illegal, buf.String(), err)
	}
	return token.Float, buf.String(), nil
}

func (s *Scanner) scanInteger() (token.Token, string, error) {
	buf := bytes.NewBuffer(nil)
	ch, err := s.read()
	switch err {
	case io.EOF:
		return s.eosIfEOF(token.Illegal, buf.String(), err)
	case nil: // continue
	default:
		return s.eosIfEOF(token.Illegal, buf.String(), err)
	}

	if isNegativeSign(ch) {
		buf.WriteRune(ch)
		ch, err = s.read() // consume it
		switch err {
		case io.EOF:
			// expected something
			return s.eosIfEOF(token.Illegal, buf.String(), err)
		case nil: // continue
		default:
			return s.eosIfEOF(token.Illegal, buf.String(), err)
		}
	}
	buf.WriteRune(ch)

	if isZeroDigit(ch) {
		// make sure next char is not a digit
		ch2, err := s.read()
		if err != nil && err != io.EOF {
			return s.eosIfEOF(token.Illegal, buf.String(), err)
		}
		if isNonZeroDigit(ch2) {
			buf.WriteRune(ch2) // add it to the error
			tok, lit, err := s.eosIfEOF(token.Illegal, buf.String(), err)
			return tok, lit, err
		}
		s.unread() // return what we read
		return token.Integer, buf.String(), err
	}

	if !isDigit(ch) {
		return s.eosIfEOF(token.Illegal, buf.String(), err)
	}

	// read the rest of the digits
	err = s.scanDigits(buf)
	switch err {
	case io.EOF, nil: // continue
		return token.Integer, buf.String(), err
	default:
		return s.eosIfEOF(token.Illegal, buf.String(), err)
	}
}

func (s *Scanner) scanDigits(buf *bytes.Buffer) error {
	for {
		ch, err := s.read()
		if err != nil {
			return err
		}
		if !isDigit(ch) {
			// past the end of what are digits
			s.unread()
			return nil
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
	case isComma(ch):
		return token.Comma, string(ch), nil
	case isPipe(ch):
		return token.Pipe, string(ch), nil
	case isDot(ch):
		return token.Dot, string(ch), nil
	case isRightBracket(ch):
		return token.RightBracket, string(ch), nil
	case isLeftBracket(ch):
		return token.LeftBracket, string(ch), nil
	case isRightParens(ch):
		return token.RightParens, string(ch), nil
	case isLeftParens(ch):
		return token.LeftParens, string(ch), nil
	case isQuote(ch):
		return token.Quote, string(ch), nil
	case isColon(ch):
		return token.Colon, string(ch), nil
	default:
		return s.eosIfEOF(token.Illegal, "", err)
	}
}

// lexicon's first set

func isStringCharacter(ch rune) bool    { return isSpace(ch) || isLetter(ch) || isDigit(ch) }
func isWhitespace(ch rune) bool         { return isSpace(ch) || isControlChar(ch) }
func isSpace(ch rune) bool              { return ch == ' ' }
func isNegativeSign(ch rune) bool       { return ch == '-' }
func isZeroDigit(ch rune) bool          { return ch == '0' }
func isNonZeroDigit(ch rune) bool       { return ch > '0' && ch <= '9' }
func isScientificExponent(ch rune) bool { return ch == 'e' }
func isComma(ch rune) bool              { return (ch == ',') }
func isPipe(ch rune) bool               { return (ch == '|') }
func isDot(ch rune) bool                { return (ch == '.') }
func isLeftBracket(ch rune) bool        { return (ch == '[') }
func isRightBracket(ch rune) bool       { return (ch == ']') }
func isLeftParens(ch rune) bool         { return (ch == '(') }
func isRightParens(ch rune) bool        { return (ch == ')') }
func isQuote(ch rune) bool              { return (ch == '"') }
func isColon(ch rune) bool              { return (ch == ':') }
func isEscapeCharacter(ch rune) bool    { return (ch == '\\') }
func isLetter(ch rune) bool             { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }
func isControlChar(ch rune) bool        { return ch == '\t' || ch == '\n' || ch == '\r' }
func isControlCode(ch rune) bool        { return ch == 't' || ch == 'n' || ch == 'r' }

func isDigit(ch rune) bool { return isZeroDigit(ch) || isNonZeroDigit(ch) }

func isEscapedKeyword(ch rune) bool {
	return isComma(ch) ||
		isPipe(ch) ||
		isDot(ch) ||
		isLeftBracket(ch) ||
		isRightBracket(ch) ||
		isLeftParens(ch) ||
		isRightParens(ch) ||
		isQuote(ch) ||
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
