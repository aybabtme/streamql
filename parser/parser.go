package parser

import (
	"io"

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

func (p *Parser) Parse() (*ast.FiltersStmt, error) {
	stmt := &ast.FiltersStmt{}
	tok, lit, err := p.scan()
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
	// don't set `used` because it indicates an unscan. this would
	// work anyways, but would force the `if p.buf.used` branch to be
	// taken every 2nd call to `scan()`

	return tok, lit, err
}

func (p *Parser) unscan() { p.buf.used = true }
