package parser

import (
	"errors"
	"fmt"
	"io"

	"github.com/aybabtme/streamql/lang/ast"
	"github.com/aybabtme/streamql/lang/scanner"
	"github.com/aybabtme/streamql/lang/token"
)

var _ error = (*SyntaxError)(nil)

func newSyntaxError(got token.Token, want ...token.Token) error {
	return &SyntaxError{Expected: want, Actual: got}
}

type SyntaxError struct {
	Expected []token.Token
	Actual   token.Token
}

func (e *SyntaxError) Error() string {
	var expect string
	for i, exp := range e.Expected {
		switch i {
		case 0:
			expect = exp.String()
		case len(e.Expected) - 1:
			expect += " or " + exp.String()
		default:
			expect += ", " + exp.String()
		}
	}
	return fmt.Sprintf("expected %s, got %s", expect, e.Actual.String())
}

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
	err := p.scanFuncsStmt(cur)
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

func (p *Parser) scanFuncsStmt(stmt *ast.FilterStmt) error {
	if err := p.scanFuncStmt(stmt); err != nil {
		return err
	}
	if err := p.scanWhitespace(); err != nil {
		return err
	}
	if err := p.scanFuncChainStmt(stmt); err != nil {
		return err
	}
	return nil
}

func (p *Parser) scanFuncStmt(stmt *ast.FilterStmt) error {
	tok, _, err := p.scan()
	if err != nil {
		return err
	}
	p.unscan()

	switch tok {
	case token.Dot:
		sel, err := p.scanSelector()
		if err != nil {
			return err
		}
		stmt.Funcs = append(stmt.Funcs, &ast.FuncStmt{
			Selector: sel,
		})

	case token.LeftParens, token.InlineString, token.Float, token.Integer:
		emitFunc, err := p.scanEmitFunc()
		if err != nil {
			return err
		}
		stmt.Funcs = append(stmt.Funcs, &ast.FuncStmt{
			EmitFunc: emitFunc,
		})

	default:
		return newSyntaxError(tok, token.Dot, token.LeftParens, token.InlineString, token.Float, token.Integer)
	}
	return nil
}

func (p *Parser) scanFuncChainStmt(stmt *ast.FilterStmt) error {
	_, _, err := p.scan()
	switch err {
	case io.EOF:
		return parseComplete
	case nil:
		p.unscan() // continue
	default:
		return err
	}
	if err := p.scanPipe(); err != nil {
		return err
	}
	if err := p.scanWhitespace(); err != nil {
		return err
	}

	if err := p.scanFuncsStmt(stmt); err != nil {
		return err
	}
	return nil
}

// func

func (p *Parser) scanEmitFunc() (*ast.EmitFuncStmt, error) {
	tok, lit, err := p.scan()
	if err != nil {
		return nil, err
	}
	p.unscan()

	switch tok {
	case token.InlineString:
		// is it a boolean literal?
		if lit == "true" || lit == "false" {
			emitBool, err := p.scanEmitBooleanFunc()
			if err != nil {
				return nil, err
			}
			return &ast.EmitFuncStmt{EmitBooleanFunc: emitBool}, nil
		}
		// or a function with a type?
		switch lit {

		case "contains", "regexp", "not":
			emitBool, err := p.scanEmitBooleanFunc()
			if err != nil {
				return nil, err
			}
			return &ast.EmitFuncStmt{EmitBooleanFunc: emitBool}, nil

		case "substring":
			emitStr, err := p.scanEmitStringFunc()
			if err != nil {
				return nil, err
			}
			return &ast.EmitFuncStmt{EmitStringFunc: emitStr}, nil

		case "select":
			emitAny, err := p.scanEmitAnyFunc()
			if err != nil {
				return nil, err
			}
			return &ast.EmitFuncStmt{EmitAnyFunc: emitAny}, nil

		case "atof", "length":
			emitNum, err := p.scanEmitNumberFunc()
			if err != nil {
				return nil, err
			}
			return &ast.EmitFuncStmt{EmitNumberFunc: emitNum}, nil

		default:
			// FIXME: return a proper error type
			return nil, fmt.Errorf("unknown keyword %q", lit)
		}
	case token.LeftParens, // FIXME: can also be for boolean algebra
		token.Float, token.Integer:
		emitNum, err := p.scanEmitNumberFunc()
		if err != nil {
			return nil, err
		}
		return &ast.EmitFuncStmt{EmitNumberFunc: emitNum}, nil
	default:
		return nil, newSyntaxError(tok, token.LeftParens, token.InlineString, token.Float, token.Integer)
	}
}

func (p *Parser) scanSelector() (*ast.SelectorStmt, error) {
	if err := p.scanDot(); err != nil {
		return nil, err
	}
	return p.scanRootObjectSelector()
}

// selectors

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
		child, err := p.scanSubSelector()
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

		child, err := p.scanSubSelector()
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

func (p *Parser) scanSubSelector() (*ast.SelectorStmt, error) {
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
		child, err := p.scanSubSelector()
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

		child, err := p.scanSubSelector()
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
	str, err := p.scanInlineString()
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

	first, err := p.scanIntegerArg()
	if err != nil {
		return nil, err
	}

	if err := p.scanWhitespace(); err != nil {
		return nil, err
	}

	stmt := new(ast.ArraySelectorStmt)
	return stmt, p.scanArrayOpIndexor(stmt, first)
}

func (p *Parser) scanArrayOpIndexor(stmt *ast.ArraySelectorStmt, first *ast.IntegerArg) error {
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
		second, err := p.scanIntegerArg()
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
		return newSyntaxError(tok, token.Colon, token.RightBracket)
	}
}

// functions

func (p *Parser) scanEmitStringFunc() (*ast.EmitStringFunc, error)   { panic("not impl") }
func (p *Parser) scanEmitAnyFunc() (*ast.EmitAnyFunc, error)         { panic("not impl") }
func (p *Parser) scanEmitNumberFunc() (*ast.EmitNumberFunc, error)   { panic("not impl") }
func (p *Parser) scanEmitBooleanFunc() (*ast.EmitBooleanFunc, error) { panic("not impl") }

func (p *Parser) scanBuiltInStrFunc() (*ast.EmitStringFunc, error)   { panic("not impl") }
func (p *Parser) scanBuiltInAnyFunc() (*ast.EmitAnyFunc, error)      { panic("not impl") }
func (p *Parser) scanBuiltInIntFunc() (*ast.EmitIntFunc, error)      { panic("not impl") }
func (p *Parser) scanBuiltInFloatFunc() (*ast.EmitFloatFunc, error)  { panic("not impl") }
func (p *Parser) scanBuiltInBoolFunc() (*ast.EmitBooleanFunc, error) { panic("not impl") }

func (p *Parser) scanAlgBoolOps() (*ast.AlgebraBooleanOps, error)   { panic("not impl") }
func (p *Parser) scanAlgBoolOpsUnary(*ast.AlgebraBooleanOps) error  { panic("not impl") }
func (p *Parser) scanAlgBoolOpsTwoAry(*ast.AlgebraBooleanOps) error { panic("not impl") }

func (p *Parser) scanAlgNumberOps() (*ast.AlgebraNumberOps, error)   { panic("not impl") }
func (p *Parser) scanAlgNumberOpsTwoAry(*ast.AlgebraNumberOps) error { panic("not impl") }

// string functions

func (p *Parser) scanFuncStringContains() (*ast.FuncStringContains, error) { panic("not impl") }
func (p *Parser) scanFuncStringRegexp() (*ast.FuncStringRegexp, error)     { panic("not impl") }
func (p *Parser) scanFuncStringSubStr() (*ast.FuncStringSubStr, error)     { panic("not impl") }
func (p *Parser) scanFuncStringLength() (*ast.FuncStringLength, error)     { panic("not impl") }
func (p *Parser) scanFuncStringAtof() (*ast.FuncStringAtof, error)         { panic("not impl") }

// any functions

func (p *Parser) scanFuncAnySelect() (*ast.FuncAnySelect, error) { panic("not impl") }

// boolean functions

func (p *Parser) scanFuncBooleanOr() (*ast.FuncBooleanOr, error)   { panic("not impl") }
func (p *Parser) scanFuncBooleanAnd() (*ast.FuncBooleanAnd, error) { panic("not impl") }
func (p *Parser) scanFuncBooleanXOR() (*ast.FuncBooleanXOR, error) { panic("not impl") }
func (p *Parser) scanFuncBooleanNot() (*ast.FuncBooleanNot, error) { panic("not impl") }

// number functions

func (p *Parser) scanFuncNumberAdd() (*ast.FuncNumberAdd, error)           { panic("not impl") }
func (p *Parser) scanFuncNumberSubtract() (*ast.FuncNumberSubtract, error) { panic("not impl") }
func (p *Parser) scanFuncNumberMultiply() (*ast.FuncNumberMultiply, error) { panic("not impl") }
func (p *Parser) scanFuncNumberDivide() (*ast.FuncNumberDivide, error)     { panic("not impl") }

// args

func (p *Parser) scanStringArg() (*ast.StringArg, error)   { panic("not implemented") }
func (p *Parser) scanBooleanArg() (*ast.BooleanArg, error) { panic("not implemented") }
func (p *Parser) scanNumberArg() (*ast.NumberArg, error)   { panic("not implemented") }
func (p *Parser) scanIntegerArg() (*ast.IntegerArg, error) { panic("not implemented") }

// literals and stuff

func (p *Parser) scanInlineString() (string, error) {
	tok, lit, err := p.scan()
	if err != nil {
		return "", err
	}
	if tok != token.InlineString {
		return "", newSyntaxError(tok, token.InlineString)
	}

	return scanner.ParseInlineString(lit)
}

func (p *Parser) scanString() (string, error) {
	tok, lit, err := p.scan()
	if err != nil {
		return "", err
	}
	if tok != token.String {
		return "", newSyntaxError(tok, token.String)
	}

	return scanner.ParseString(lit)
}

func (p *Parser) scanBoolean() (bool, error) {
	tok, lit, err := p.scan()
	if err != nil {
		return false, err
	}
	if tok != token.InlineString {
		return false, newSyntaxError(tok, token.InlineString)
	}

	return scanner.ParseBoolean(lit)
}

func (p *Parser) scanNumber() (float64, error) {
	tok, lit, err := p.scan()
	if err != nil {
		return 0, err
	}
	if tok != token.Float && tok != token.Integer {
		return 0, newSyntaxError(tok, token.Float, token.Integer)
	}
	return scanner.ParseNumber(lit)
}

func (p *Parser) scanInteger() (int, error) {
	tok, lit, err := p.scan()
	if err != nil {
		return 0, err
	}
	if tok != token.Integer {
		return 0, newSyntaxError(tok, token.Integer)
	}
	return scanner.ParseInteger(lit)
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
		return newSyntaxError(got, want)
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
