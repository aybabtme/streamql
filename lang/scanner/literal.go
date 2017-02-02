package scanner

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
)

func ParseBoolean(lit string) (bool, error) {
	switch lit {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value %q", lit)
	}
}

func ParseInlineString(lit string) (string, error) {
	rd := strings.NewReader(lit)
	ch, _, err := rd.ReadRune()
	switch err {
	case io.EOF:
		return "", nil
	case nil: // continue
	default:
		return "", err
	}

	// read the first Letter | EscapedKeyword
	switch {
	case isLetter(ch): // can not start with a digit
	case isEscapeCharacter(ch):
		ch, _, err = rd.ReadRune()
		switch {
		case err == io.EOF:
			return "", errors.New("unfinished escape sequence")
		case err != nil:
			return "", err
		}
		if !isEscapedKeyword(ch) {
			return "", errors.New("unknown escape sequence")
		}
	default:
		return "", errors.New("invalid character in inline string")
	}
	buf := bytes.NewBuffer(nil)
	buf.WriteRune(ch)

	// then read subsequent Letter | Digit | EscapedKeyword
	for {
		ch, _, err := rd.ReadRune()
		switch err {
		case io.EOF:
			return buf.String(), nil
		case nil: // continue
		default:
			return "", err
		}
		switch {
		case isDigit(ch): // accept digits after first char
		case isLetter(ch):
		case isEscapeCharacter(ch):
			ch, _, err = rd.ReadRune() // read past the \
			switch {
			case err == io.EOF:
				return "", errors.New("unfinished escape sequence")
			case err != nil:
				return "", err
			}
			if !isEscapedKeyword(ch) {
				return "", errors.New("unknown escape sequence")
			}
		default:
			// past the end of what's a valid InlineString
			return "", errors.New("invalid character in inline string")
		}
		buf.WriteRune(ch)
	}
}

func ParseString(lit string) (string, error) {
	rd := strings.NewReader(lit)
	ch, _, err := rd.ReadRune()
	switch {
	case err == io.EOF:
		return "", errors.New("missing opening quote")
	case err != nil:
		return "", err
	case !isQuote(ch):
		return "", errors.New("doesn't start with an opening quote")
	}

	buf := bytes.NewBuffer(nil)

	for {
		ch, _, err := rd.ReadRune()
		switch {
		case err == io.EOF:
			return "", errors.New("missing closing quote")
		case err != nil:
			return "", err
		}
		switch {
		case isStringCharacter(ch):
			buf.WriteRune(ch)
		case isEscapeCharacter(ch):
			ch2, _, err := rd.ReadRune() // read past the \
			switch {
			case err == io.EOF:
				return "", errors.New("unfinished escape sequence")
			case err != nil:
				return "", err
			}
			switch {
			case isQuote(ch2), isEscapeCharacter(ch2): // continue
				buf.WriteRune(ch2)
			case isControlCode(ch2):
				switch ch2 {
				case 'n':
					buf.WriteRune('\n')
				case 't':
					buf.WriteRune('\t')
				case 'r':
					buf.WriteRune('\r')
				default:
					panic(fmt.Sprintf("bug: missing case for accepted control code %#v", ch2))
				}
			default:
				return "", fmt.Errorf("unknown escape sequence: %q", ch)
			}
		case isQuote(ch):
			// we're done reading the string
			return buf.String(), nil
		default:
			return "", fmt.Errorf("invalid character: %q", ch)
		}
	}
}

func eofAsUnexpectedEOF(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}

func ParseNumber(lit string) (float64, error) {
	// this function is implemented like horse shit
	neg := len(lit) > 0 && lit[0] == '-'
	v := 0.0

	decIdx := strings.IndexRune(lit, '.')
	expIdx := -1
	if decIdx != -1 {
		expIdx = strings.IndexRune(lit[decIdx:], 'e')
	} else {
		expIdx = strings.IndexRune(lit, 'e')
	}

	var (
		hasDecimal  = decIdx != -1
		hasExponent = expIdx != -1
		intPart     int64
		err         error
	)

	if hasDecimal {
		intPart, err = ParseInteger(lit[:decIdx])
	} else if hasExponent {
		intPart, err = ParseInteger(lit[:expIdx])
	} else {
		intPart, err = ParseInteger(lit)
		return float64(intPart), err
	}
	if err != nil {
		return 0, err
	}
	v += float64(intPart)

	var (
		decLit string
		expLit string
	)
	if hasDecimal && hasExponent {
		expIdx += decIdx
		decLit = lit[decIdx+1 : expIdx]
	} else if hasDecimal && !hasExponent {
		decLit = lit[decIdx+1:]
	}
	if hasExponent {
		expLit = lit[expIdx+1:]
	}
	if hasDecimal {
		decPart := 0.0
		for i, ch := range []byte(decLit) {

			if isZeroDigit(rune(ch)) {
				continue
			}
			if isDigit(rune(ch)) {
				decPart += float64(ch-'0') * 1 / math.Pow10(i+1)
			} else {
				return 0, fmt.Errorf("invalid character: %q", rune(ch))
			}
		}
		if neg {
			v -= decPart
		} else {
			v += decPart
		}
	}

	if hasExponent {
		expPart, err := ParseInteger(expLit)
		if err != nil {
			return 0, err
		}
		v *= math.Pow10(int(expPart))
	}
	return v, nil
}

func ParseInteger(lit string) (int64, error) {
	return parseInteger(strings.NewReader(lit))
}

func parseInteger(rd *strings.Reader) (int64, error) {
	neg := false
	v := int64(0)

	ch, _, err := rd.ReadRune()
	if err != nil {
		return 0, eofAsUnexpectedEOF(err)
	}
	if isNegativeSign(ch) {
		neg = true
	} else {
		rd.UnreadRune()
	}
	digit, err := parseDigit(rd)
	if err != nil {
		return 0, eofAsUnexpectedEOF(err)
	}
	if digit == 0 {
		_, err := parseDigit(rd)
		switch err {
		case io.EOF:
			if neg {
				v = -0
			}
			return v, nil
		case nil:
			return v, fmt.Errorf("leading 0 must have nothing but a 0")
		default:
			return v, err
		}

	}
	v += int64(digit)
	for {
		digit, err := parseDigit(rd)
		switch err {
		case io.EOF:
			if neg {
				v = -v
			}
			return v, nil
		case nil:
			v *= 10
			v += int64(digit)
		default:
			return v, err
		}
	}
}

func parseDigit(rd io.RuneReader) (uint8, error) {
	ch, _, err := rd.ReadRune()
	switch err {
	case io.EOF:
		return 0, err
	case nil:
		if !isDigit(ch) {
			return 0, fmt.Errorf("invalid digit value %q", ch)
		}
	default:
		return 0, err
	}
	switch ch {
	case '0':
		return 0, nil
	case '1':
		return 1, nil
	case '2':
		return 2, nil
	case '3':
		return 3, nil
	case '4':
		return 4, nil
	case '5':
		return 5, nil
	case '6':
		return 6, nil
	case '7':
		return 7, nil
	case '8':
		return 8, nil
	case '9':
		return 9, nil
	default:
		panic(fmt.Sprintf("bug: missing case for %q", ch))
	}
}
