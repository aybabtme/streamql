package parser

import (
	"errors"
	"fmt"
	"io"

	"strconv"

	"github.com/aybabtme/streamql/ast"
	"github.com/aybabtme/streamql/scanner"
	"github.com/aybabtme/streamql/token"
)

// Parser represents a parser.
type Parser struct {
	s   *scanner.Scanner
	buf struct {
		tok  token.Token
		lit  string
		used bool
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: scanner.NewScanner(r)}
}

var parseComplete = errors.New("parse completed")

func (p *Parser) Parse() (*ast.FiltersStmt, error) {
	stmt := &ast.FiltersStmt{}
	err := p.scanFiltersStmt(stmt)
	switch err {
	case parseComplete, nil:
		return stmt, nil
	case io.EOF:
		return stmt, io.ErrUnexpectedEOF
	default:
		return stmt, err
	}
}

// scanners

func (p *Parser) scanFiltersStmt(stmt *ast.FiltersStmt) error {
	if err := p.scanFilterStmt(stmt); err != nil {
		return err
	}
	if err := p.scanWhitespace(); err != nil {
		return err
	}
	if err := p.scanFilterChain(stmt); err != nil {
		return err
	}
	return nil
}

func (p *Parser) scanFilterStmt(stmt *ast.FiltersStmt) error {
	cur := &ast.FilterStmt{}
	err := p.scanSelectorsStmt(cur)
	switch err {
	case parseComplete, nil:
		if cur != nil {
			stmt.Filters = append(stmt.Filters, cur)
		}
		return err
	default:
		return err
	}
}

func (p *Parser) scanFilterChain(stmt *ast.FiltersStmt) error {
	_, _, err := p.scan()
	switch err {
	case io.EOF:
		return parseComplete
	case nil:
		p.unscan() // continue
	default:
		return err
	}

	if err := p.scanComma(); err != nil {
		return err
	}
	if err := p.scanWhitespace(); err != nil {
		return err
	}
	if err := p.scanFiltersStmt(stmt); err != nil {
		return err
	}
	return nil
}

func (p *Parser) scanSelectorsStmt(stmt *ast.FilterStmt) error {
	if err := p.scanRootSelector(stmt); err != nil {
		return err
	}
	if err := p.scanWhitespace(); err != nil {
		return err
	}
	tok, _, err := p.scan()
	if err != nil {
		return err
	}
	p.unscan()
	switch tok {
	case token.Comma:
		return nil
	case token.Pipe:
		if err := p.scanSelectorChain(stmt); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("expecting a %v or a %v, not a %v",
			token.Comma, token.Pipe, tok,
		)
	}
}

func (p *Parser) scanSelectorChain(stmt *ast.FilterStmt) error {
	if err := p.scanPipe(); err != nil {
		return err
	}
	if err := p.scanWhitespace(); err != nil {
		return err
	}

	if err := p.scanSelectorsStmt(stmt); err != nil {

		return err
	}
	return nil
}

func (p *Parser) scanRootSelector(stmt *ast.FilterStmt) error {
	if err := p.scanDot(); err != nil {
		return err
	}
	cur, err := p.scanRootObjectSelector()
	switch err {
	case nil, parseComplete:
		if cur != nil {
			stmt.Selectors = append(stmt.Selectors, cur)
		}
	default:
	}
	return err
}

func (p *Parser) scanRootObjectSelector() (*ast.SelectorStmt, error) {

	// figure out what type we're looking at
	tok, _, err := p.scan()
	switch err {
	case io.EOF:
		// there's nothing but an empty selector
		return &ast.SelectorStmt{}, parseComplete
	case nil: // continue
	default:
		return nil, err
	}
	p.unscan()
	switch tok {
	case token.InlineString: // RootObject has the leading Dot scan'ed away
		object, err := p.scanMemberSelector()
		if err != nil {
			return nil, err
		}
		stmt := &ast.SelectorStmt{Object: object}

		// if there's more, they're child selectors
		child, err := p.scanSelector()
		switch err {
		case parseComplete, nil:
			if child != nil {
				stmt.Object.Child = child
			}
			return stmt, err
		default:
			return nil, err
		}

	case token.LeftBracket:
		array, err := p.scanArraySelector()
		if err != nil {
			return nil, err
		}
		stmt := &ast.SelectorStmt{Array: array}

		child, err := p.scanSelector()
		switch err {
		case parseComplete, nil:
			if child != nil {
				stmt.Array.Child = child
			}
			return stmt, err
		default:
			return nil, err
		}

	case token.Whitespace:
		err := p.scanWhitespace()
		switch err {
		case io.EOF:
			return nil, parseComplete
		case nil: // continue
		default:
			return nil, err
		}
		return nil, nil

	case token.Comma, token.Pipe:
		// we return an empty statement because there
		// was a Dot
		return new(ast.SelectorStmt), nil

	default: // can be epsilon, aka anything
		return nil, nil
	}
}

func (p *Parser) scanSelector() (*ast.SelectorStmt, error) {
	// figure out what type we're looking at
	tok, _, err := p.scan()
	switch err {
	case io.EOF:
		return nil, parseComplete
	case nil: // continue
	default:

		return nil, err
	}
	p.unscan()
	switch tok {
	case token.Dot: // RootObject has the leading Dot scan'ed away
		object, err := p.scanObjectSelector()
		if err != nil {
			return nil, err
		}
		stmt := &ast.SelectorStmt{Object: object}

		// if there's more, they're child selectors
		child, err := p.scanSelector()
		switch err {
		case parseComplete, nil:
			if child != nil {
				stmt.Object.Child = child
			}
			return stmt, err
		default:
			return nil, err
		}

	case token.LeftBracket:
		array, err := p.scanArraySelector()
		if err != nil {
			return nil, err
		}
		stmt := &ast.SelectorStmt{Array: array}

		child, err := p.scanSelector()
		switch err {
		case parseComplete, nil:
			if child != nil {
				stmt.Array.Child = child
			}
			return stmt, err
		default:
			return nil, err
		}

	case token.Whitespace:
		err := p.scanWhitespace()
		switch err {
		case io.EOF:
			err = parseComplete
		case nil: // continue
		default:
			return nil, err
		}
		return nil, err

	case token.Comma, token.Pipe:
		return nil, nil

	default:
		return nil, nil
	}
}

func (p *Parser) scanObjectSelector() (*ast.ObjectSelectorStmt, error) {
	if err := p.scanDot(); err != nil {
		return nil, err
	}
	return p.scanMemberSelector()
}

func (p *Parser) scanMemberSelector() (*ast.ObjectSelectorStmt, error) {
	str, err := p.scanString()
	if err != nil {
		return nil, err
	}
	return &ast.ObjectSelectorStmt{Member: str}, nil
}

func (p *Parser) scanArraySelector() (*ast.ArraySelectorStmt, error) {
	if err := p.scanLeftBracket(); err != nil {
		return nil, err
	}
	return p.scanArrayOpSelector()
}

func (p *Parser) scanArrayOpSelector() (*ast.ArraySelectorStmt, error) {

	tok, _, err := p.scan()
	if err != nil {
		return nil, err
	}
	if tok == token.RightBracket {
		return &ast.ArraySelectorStmt{
			Each: &ast.EachSelectorStmt{},
		}, nil
	}
	p.unscan() // we just want to peek

	if err := p.scanWhitespace(); err != nil {
		return nil, err
	}

	first, err := p.scanInteger()
	if err != nil {
		return nil, err
	}

	if err := p.scanWhitespace(); err != nil {
		return nil, err
	}

	stmt := new(ast.ArraySelectorStmt)
	return stmt, p.scanArrayOpIndexor(stmt, first)
}

func (p *Parser) scanArrayOpIndexor(stmt *ast.ArraySelectorStmt, first int) error {
	tok, _, err := p.scan()
	if err != nil {
		return err
	}
	// we don't unscan

	switch tok {
	case token.Colon:
		// we consumed the colon, don't need to scan it
		if err := p.scanWhitespace(); err != nil {
			return err
		}
		second, err := p.scanInteger()
		if err != nil {
			return err
		}
		if err := p.scanWhitespace(); err != nil {
			return err
		}
		if err := p.scanRightBracket(); err != nil {
			return err
		}
		stmt.RangeEach = &ast.RangeEachSelectorStmt{From: first, To: second}
		return nil

	case token.RightBracket:
		stmt.Index = &ast.IndexSelectorStmt{Index: first}
		return nil
	default:
		return fmt.Errorf("expected a %v or a %v, not a %v",
			token.Colon, token.RightBracket, tok,
		)
	}
}

func (p *Parser) scanString() (string, error) {
	tok, lit, err := p.scan()
	if err != nil {
		return "", err
	}
	if tok != token.InlineString {
		return "", fmt.Errorf("expected a %v, not a %v", token.InlineString, tok)
	}
	return lit, nil
}

func (p *Parser) scanInteger() (int, error) {
	tok, lit, err := p.scan()
	if err != nil {
		return 0, err
	}
	if tok != token.Integer {
		return 0, fmt.Errorf("expected a %v, not a %v", token.Integer, tok)
	}
	return strconv.Atoi(lit)
}

func (p *Parser) scanWhitespace() error {
	for {
		tok, _, err := p.scan()
		if err != nil {
			return err
		}
		if tok != token.Whitespace {
			p.unscan() // went too far!
			return nil
		}
	}
}

func (p *Parser) scanComma() error        { return p.scanToken(token.Comma) }
func (p *Parser) scanPipe() error         { return p.scanToken(token.Pipe) }
func (p *Parser) scanDot() error          { return p.scanToken(token.Dot) }
func (p *Parser) scanLeftBracket() error  { return p.scanToken(token.LeftBracket) }
func (p *Parser) scanRightBracket() error { return p.scanToken(token.RightBracket) }

func (p *Parser) scanToken(want token.Token) error {
	got, _, err := p.scan()
	if err != nil {
		return err
	}
	if want != got {
		return fmt.Errorf("expected a %v, not a %v", want, got)
	}
	return nil
}

// helpers

func (p *Parser) scan() (token.Token, string, error) {
	if p.buf.used {
		p.buf.used = false

		return p.buf.tok, p.buf.lit, nil
	}
	tok, lit, err := p.s.Scan()
	p.buf.tok = tok
	p.buf.lit = lit
	// don't set `p.buf.used` because it indicates an unscan. this would
	// work anyways, but would force the `if p.buf.used` branch to be
	// taken every 2nd call to `scan()`

	return tok, lit, err
}

func (p *Parser) unscan() { p.buf.used = true }
