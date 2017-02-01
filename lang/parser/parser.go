package parser

import (
	"errors"
	"fmt"
	"io"
	"log"

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

func newUnknownKeywordError(got string, want ...string) error {
	return &UnknownKeywordError{Expected: want, Actual: got}
}

type UnknownKeywordError struct {
	Expected []string
	Actual   string
}

func (e *UnknownKeywordError) Error() string {
	var expect string
	for i, exp := range e.Expected {
		switch i {
		case 0:
			expect = fmt.Sprintf("%q", exp)
		case len(e.Expected) - 1:
			expect += " or " + fmt.Sprintf("%q", exp)
		default:
			expect += ", " + fmt.Sprintf("%q", exp)
		}
	}
	return fmt.Sprintf("expected %s, got %q", expect, e.Actual)
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
	log.Printf("Parse")
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
	log.Printf("scanFiltersStmt")
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
	log.Printf("scanFilterStmt")
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
	log.Printf("scanFilterChain")
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
	log.Printf("scanFuncsStmt")
	if err := p.scanFuncStmt(stmt); err != nil {

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
		if err := p.scanFuncChainStmt(stmt); err != nil {
			return err
		}
		return nil
	default:
		return newSyntaxError(tok, token.Comma, token.Pipe)
	}
}

func (p *Parser) scanFuncStmt(stmt *ast.FilterStmt) error {
	log.Printf("scanFuncStmt")
	tok, _, err := p.scan()
	if err != nil {
		return err
	}
	p.unscan()

	switch tok {
	case token.Dot:
		sel, err := p.scanSelector()
		switch err {
		case parseComplete, nil:
			stmt.Funcs = append(stmt.Funcs, &ast.FuncStmt{
				Selector: sel,
			})
			return err
		default:
			return err
		}

	case token.LeftParens, token.InlineString, token.Float, token.Integer:
		emitFunc, err := p.scanEmitFunc()
		switch err {
		case parseComplete, nil:
			stmt.Funcs = append(stmt.Funcs, &ast.FuncStmt{
				EmitFunc: emitFunc,
			})
			return err
		default:
			return err
		}

	default:
		return newSyntaxError(tok, token.Dot, token.LeftParens, token.InlineString, token.Float, token.Integer)
	}
}

func (p *Parser) scanFuncChainStmt(stmt *ast.FilterStmt) error {
	log.Printf("scanFuncChainStmt")
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
	log.Printf("scanEmitFunc")
	tok, lit, err := p.scan()
	if err != nil {
		return nil, err
	}
	p.unscan()
	switch tok {
	case token.InlineString:
		// is it a boolean literal?
		if lit == trueKeyword || lit == falseKeyword {
			emitBool, err := p.scanEmitBooleanFunc()
			return &ast.EmitFuncStmt{EmitBooleanFunc: emitBool}, err
		}
		// or a function with a type?
		switch lit {

		case containsKeyword, regexpKeyword, notKeyword:
			emitBool, err := p.scanEmitBooleanFunc()
			return &ast.EmitFuncStmt{EmitBooleanFunc: emitBool}, err

		case substringKeyword:
			emitStr, err := p.scanEmitStringFunc()
			return &ast.EmitFuncStmt{EmitStringFunc: emitStr}, err

		case selectKeyword:
			emitAny, err := p.scanEmitAnyFunc()
			return &ast.EmitFuncStmt{EmitAnyFunc: emitAny}, err

		case atofKeyword, lengthKeyword:
			emitNum, err := p.scanEmitNumberFunc()
			return &ast.EmitFuncStmt{EmitNumberFunc: emitNum}, err

		default:
			return nil, newUnknownKeywordError(lit,
				containsKeyword, regexpKeyword, notKeyword, substringKeyword, selectKeyword, atofKeyword, lengthKeyword,
			)
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
	log.Printf("scanSelector")
	if err := p.scanDot(); err != nil {
		return nil, err
	}
	return p.scanRootObjectSelector()
}

// selectors

func (p *Parser) scanRootObjectSelector() (*ast.SelectorStmt, error) {
	log.Printf("scanRootObjectSelector")

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
	log.Printf("scanSubSelector")
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
	log.Printf("scanObjectSelector")
	if err := p.scanDot(); err != nil {
		return nil, err
	}
	return p.scanMemberSelector()
}

func (p *Parser) scanMemberSelector() (*ast.ObjectSelectorStmt, error) {
	log.Printf("scanMemberSelector")
	str, err := p.scanInlineString()
	if err != nil {
		return nil, err
	}
	return &ast.ObjectSelectorStmt{Member: str}, nil
}

func (p *Parser) scanArraySelector() (*ast.ArraySelectorStmt, error) {
	log.Printf("scanArraySelector")
	if err := p.scanLeftBracket(); err != nil {
		return nil, err
	}
	return p.scanArrayOpSelector()
}

func (p *Parser) scanArrayOpSelector() (*ast.ArraySelectorStmt, error) {
	log.Printf("scanArrayOpSelector")

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
	log.Printf("scanArrayOpIndexor")
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

const (
	trueKeyword  = `true`
	falseKeyword = `false`

	notKeyword = `not`
	andKeyword = `and`
	orKeyword  = `or`
	xorKeyword = `xor`

	substringKeyword = `substring`
	containsKeyword  = `contains`
	regexpKeyword    = `regexp`

	selectKeyword = `select`
	atofKeyword   = `atof`
	lengthKeyword = `length`
)

func (p *Parser) scanEmitStringFunc() (*ast.EmitStringFunc, error) {
	log.Printf("scanEmitStringFunc")
	tok, lit, err := p.scan()
	if err != nil {
		return nil, err
	}
	p.unscan()
	if tok != token.InlineString {
		return nil, newSyntaxError(tok, token.InlineString)
	}
	switch lit {
	case substringKeyword:
		return p.scanBuiltInStrFunc()
	default:
		return nil, newUnknownKeywordError(lit, substringKeyword)
	}
}

func (p *Parser) scanEmitAnyFunc() (*ast.EmitAnyFunc, error) {
	log.Printf("scanEmitAnyFunc")
	tok, lit, err := p.scan()
	if err != nil {
		return nil, err
	}
	p.unscan()
	if tok != token.InlineString {
		return nil, newSyntaxError(tok, token.InlineString)
	}
	switch lit {
	case selectKeyword:
		return p.scanEmitAnyFunc()
	default:
		return nil, newUnknownKeywordError(lit, selectKeyword)
	}
}

func (p *Parser) scanEmitNumberFunc() (*ast.EmitNumberFunc, error) {
	log.Printf("scanEmitNumberFunc")
	tok, lit, err := p.scan()
	if err != nil {
		return nil, err
	}
	p.unscan()

	numFunc := new(ast.EmitNumberFunc)

	switch tok {
	case token.InlineString: // continue
		switch lit {
		case atofKeyword:
		default:
			return nil, newUnknownKeywordError(lit, atofKeyword)
		}

	case token.LeftParens, token.Integer, token.Float: // continue

	default:
		return nil, newSyntaxError(tok, token.InlineString, token.LeftParens, token.Integer, token.Float)
	}
	return numFunc, p.scanAlgNumberOps(numFunc)
}

func (p *Parser) scanEmitBooleanFunc() (*ast.EmitBooleanFunc, error) {
	log.Printf("scanEmitBooleanFunc")
	tok, lit, err := p.scan()
	if err != nil {
		return nil, err
	}
	p.unscan()

	boolFunc := new(ast.EmitBooleanFunc)

	switch tok {
	case token.InlineString: // continue
		switch lit {
		case trueKeyword, falseKeyword:
		case notKeyword:
		case regexpKeyword, containsKeyword:
		default:
			return nil, newUnknownKeywordError(lit, trueKeyword, falseKeyword, notKeyword, regexpKeyword, containsKeyword, atofKeyword)
		}

	default:
		return nil, newSyntaxError(tok, token.InlineString)
	}

	return boolFunc, p.scanAlgBoolOps(boolFunc)
}

func (p *Parser) scanBuiltInStrFunc() (*ast.EmitStringFunc, error) {
	log.Printf("scanBuiltInStrFunc")
	tok, lit, err := p.scan()
	if err != nil {
		return nil, err
	}
	p.unscan()

	strFunc := new(ast.EmitStringFunc)

	var rule func(*ast.EmitStringFunc) error
	switch tok {
	case token.InlineString: // continue
		switch lit {
		case substringKeyword:
			rule = p.scanFuncStringSubStr

		default:
			return nil, newUnknownKeywordError(lit, substringKeyword)
		}

	default:
		return nil, newSyntaxError(tok, token.InlineString)
	}

	return strFunc, rule(strFunc)
}
func (p *Parser) scanBuiltInAnyFunc() (*ast.EmitAnyFunc, error) {
	log.Printf("scanBuiltInAnyFunc")
	tok, lit, err := p.scan()
	if err != nil {
		return nil, err
	}
	p.unscan()

	anyFunc := new(ast.EmitAnyFunc)

	var rule func(*ast.EmitAnyFunc) error
	switch tok {
	case token.InlineString: // continue
		switch lit {
		case selectKeyword:
			rule = p.scanFuncAnySelect

		default:
			return nil, newUnknownKeywordError(lit, selectKeyword)
		}

	default:
		return nil, newSyntaxError(tok, token.InlineString)
	}

	return anyFunc, rule(anyFunc)
}
func (p *Parser) scanBuiltInIntFunc() (*ast.EmitIntFunc, error) {
	log.Printf("scanBuiltInIntFunc")
	tok, lit, err := p.scan()
	if err != nil {
		return nil, err
	}
	p.unscan()

	intFunc := new(ast.EmitIntFunc)

	var rule func(*ast.EmitIntFunc) error
	switch tok {
	case token.InlineString: // continue
		switch lit {
		case lengthKeyword:
			rule = p.scanFuncStringLength

		default:
			return nil, newUnknownKeywordError(lit, lengthKeyword)
		}

	default:
		return nil, newSyntaxError(tok, token.InlineString)
	}

	return intFunc, rule(intFunc)
}

func (p *Parser) scanBuiltInFloatFunc() (*ast.EmitFloatFunc, error) {
	log.Printf("scanBuiltInFloatFunc")
	tok, lit, err := p.scan()
	if err != nil {
		return nil, err
	}
	p.unscan()

	floatFunc := new(ast.EmitFloatFunc)

	var rule func(*ast.EmitFloatFunc) error
	switch tok {
	case token.InlineString: // continue
		switch lit {
		case atofKeyword:
			rule = p.scanFuncStringAtof

		default:
			return nil, newUnknownKeywordError(lit, atofKeyword)
		}

	default:
		return nil, newSyntaxError(tok, token.InlineString)
	}

	return floatFunc, rule(floatFunc)
}

func (p *Parser) scanBuiltInBoolFunc() (*ast.EmitBooleanFunc, error) {
	log.Printf("scanBuiltInBoolFunc")
	tok, lit, err := p.scan()
	if err != nil {
		return nil, err
	}
	p.unscan()

	boolFunc := new(ast.EmitBooleanFunc)

	var rule func(*ast.EmitBooleanFunc) error
	switch tok {
	case token.InlineString: // continue
		switch lit {
		case regexpKeyword:
			rule = p.scanFuncStringRegexp
		case containsKeyword:
			rule = p.scanFuncStringContains

		default:
			return nil, newUnknownKeywordError(lit, regexpKeyword, containsKeyword)
		}

	default:
		return nil, newSyntaxError(tok, token.InlineString)
	}

	return boolFunc, rule(boolFunc)
}

func (p *Parser) scanAlgBoolOps(boolFunc *ast.EmitBooleanFunc) error {
	log.Printf("scanAlgBoolOps")
	tok, lit, err := p.scan()
	if err != nil {
		return err
	}
	p.unscan()

	switch tok {
	case token.InlineString: // continue

		switch lit {

		case notKeyword:
			return p.scanAlgBoolOpsUnary(boolFunc)

		case regexpKeyword, containsKeyword, trueKeyword, falseKeyword:
			lhsArg, err := p.scanBooleanArg()
			if err != nil {
				return err
			}
			return p.scanAlgBoolOpsTwoAry(boolFunc, lhsArg)

		default:
			return newUnknownKeywordError(lit, notKeyword, regexpKeyword, containsKeyword, trueKeyword, falseKeyword)
		}

	default:
		return newSyntaxError(tok, token.InlineString)
	}
}

func (p *Parser) scanAlgBoolOpsUnary(boolFunc *ast.EmitBooleanFunc) error {
	log.Printf("scanAlgBoolOpsUnary")
	tok, lit, err := p.scan()
	if err != nil {
		return err
	}
	p.unscan()

	boolFunc.Algebra = new(ast.AlgebraBooleanOps)
	switch tok {
	case token.InlineString: // continue

		switch lit {
		case notKeyword:
			return p.scanFuncBooleanNot(boolFunc.Algebra)
		default:
			return newUnknownKeywordError(lit, notKeyword)
		}

	default:
		return newSyntaxError(tok, token.InlineString)
	}
}
func (p *Parser) scanAlgBoolOpsTwoAry(boolFunc *ast.EmitBooleanFunc, lhs *ast.BooleanArg) error {
	log.Printf("scanAlgBoolOpsTwoAry")

	tok, lit, err := p.scan()
	switch err {
	case io.EOF:
		if lhs.Boolean != nil {
			boolFunc.Literal = lhs.Boolean
		} else if lhs.EmitBooleanFunc != nil {
			boolFunc.Literal = lhs.EmitBooleanFunc.Literal
			boolFunc.StringContains = lhs.EmitBooleanFunc.StringContains
			boolFunc.StringRegexp = lhs.EmitBooleanFunc.StringRegexp
		}
		return parseComplete
	case nil: // continue
	default:
		return err
	}
	p.unscan()

	boolFunc.Algebra = new(ast.AlgebraBooleanOps)
	switch tok {

	case token.InlineString: // continue

		switch lit {
		case xorKeyword:
			return p.scanFuncBooleanXOR(boolFunc.Algebra, lhs)
		case andKeyword:
			return p.scanFuncBooleanAnd(boolFunc.Algebra, lhs)
		case orKeyword:
			return p.scanFuncBooleanOr(boolFunc.Algebra, lhs)

		default:
			return newUnknownKeywordError(lit, xorKeyword, andKeyword, orKeyword)
		}

	default:
		return nil // could be anything, aka epsilon
	}
}

func (p *Parser) scanAlgNumberOps(numFunc *ast.EmitNumberFunc) error {
	log.Printf("scanAlgNumberOps")
	tok, lit, err := p.scan()
	if err != nil {
		return err
	}
	p.unscan()

	switch tok {
	case token.InlineString: // continue

		switch lit {

		case lengthKeyword, atofKeyword:
			lhsArg, err := p.scanNumberArg()
			if err != nil {
				return err
			}
			err = p.scanAlgNumberOpsTwoAry(numFunc, lhsArg)
			return err

		default:
			return newUnknownKeywordError(lit, lengthKeyword, atofKeyword)
		}

	case token.LeftParens:
		if err := p.scanLeftParens(); err != nil {
			return err
		}
		if err := p.scanAlgNumberOps(numFunc); err != nil {
			return err
		}
		if err := p.scanRightParens(); err != nil {
			return err
		}
		return nil

	case token.Integer, token.Float:
		lhsArg, err := p.scanNumberArg()
		if err != nil {
			return err
		}
		return p.scanAlgNumberOpsTwoAry(numFunc, lhsArg)

	default:
		return newSyntaxError(tok, token.InlineString)
	}
}

func (p *Parser) scanAlgNumberOpsTwoAry(numFunc *ast.EmitNumberFunc, lhs *ast.NumberArg) error {
	log.Printf("scanAlgNumberOpsTwoAry")
	tok, _, err := p.scan()
	switch err {
	case io.EOF:
		numFunc.Algebra = lhs.EmitNumberFunc.Algebra
		numFunc.Float = lhs.EmitNumberFunc.Float
		numFunc.Int = lhs.EmitNumberFunc.Int
		return parseComplete
	case nil: // continue
	default:
		return err
	}
	p.unscan()

	numFunc.Algebra = new(ast.AlgebraNumberOps)
	switch tok {
	case token.PlusSymbol:
		return p.scanFuncNumberAdd(numFunc.Algebra, lhs)
	case token.MinusSymbol:
		return p.scanFuncNumberSubtract(numFunc.Algebra, lhs)
	case token.MultiplySymbol:
		return p.scanFuncNumberMultiply(numFunc.Algebra, lhs)
	case token.DivideSymbol:
		return p.scanFuncNumberDivide(numFunc.Algebra, lhs)

	default:
		return nil // could be anything, aka epsilon
	}
}

// string functions

func (p *Parser) scanFuncStringContains(boolFunc *ast.EmitBooleanFunc) (err error) {
	log.Printf("scanFuncStringContains")
	fn := new(ast.FuncStringContains)
	err = p.scanFunc(
		containsKeyword,
		func() error {
			fn.SubString, err = p.scanStringArg()
			return err
		},
		func() error {
			fn.Selector, err = p.scanSelector()
			return err
		},
	)
	boolFunc.StringContains = fn
	return nil
}

func (p *Parser) scanFuncStringRegexp(boolFunc *ast.EmitBooleanFunc) (err error) {
	log.Printf("scanFuncStringRegexp")
	fn := new(ast.FuncStringRegexp)
	err = p.scanFunc(
		regexpKeyword,
		func() error {
			fn.Expression, err = p.scanStringArg()
			return err
		},
		func() error {
			fn.Selector, err = p.scanSelector()
			return err
		},
	)
	boolFunc.StringRegexp = fn
	return err
}
func (p *Parser) scanFuncStringSubStr(strFunc *ast.EmitStringFunc) (err error) {
	log.Printf("scanFuncStringSubStr")
	fn := new(ast.FuncStringSubStr)
	err = p.scanFunc(
		substringKeyword,
		func() error {
			fn.String, err = p.scanStringArg()
			return err
		},
		func() error {
			fn.From, err = p.scanIntegerArg()
			return err
		},
		func() error {
			fn.To, err = p.scanIntegerArg()
			return err
		},
	)
	strFunc.StringSubStr = fn
	return nil
}

func (p *Parser) scanFuncStringLength(intFunc *ast.EmitIntFunc) (err error) {
	log.Printf("scanFuncStringLength")
	fn := new(ast.FuncStringLength)
	err = p.scanFunc(
		lengthKeyword,
		func() error {
			fn.String, err = p.scanStringArg()
			return err
		},
	)
	intFunc.StringLength = fn
	return err
}

func (p *Parser) scanFuncStringAtof(floatFunc *ast.EmitFloatFunc) (err error) {
	log.Printf("scanFuncStringAtof")
	fn := new(ast.FuncStringAtof)
	err = p.scanFunc(
		atofKeyword,
		func() error {
			fn.String, err = p.scanStringArg()
			return err
		},
	)
	floatFunc.StringAtof = fn
	return err
}

// any functions

func (p *Parser) scanFuncAnySelect(anyFunc *ast.EmitAnyFunc) (err error) {
	log.Printf("scanFuncAnySelect")
	fn := new(ast.FuncAnySelect)
	err = p.scanFunc(
		selectKeyword,
		func() error {
			fn.Condition, err = p.scanBooleanArg()
			return err
		},
	)
	anyFunc.AnySelect = fn
	return err
}

// boolean functions

func (p *Parser) scanFuncBooleanOr(alg *ast.AlgebraBooleanOps, lhs *ast.BooleanArg) (err error) {
	log.Printf("scanFuncBooleanOr")
	fn := &ast.FuncBooleanOr{LHS: lhs}
	if err = p.scanInlineStringKeyword(orKeyword); err != nil {
		return err
	}
	if fn.RHS, err = p.scanBooleanArg(); err != nil {
		return err
	}
	alg.Or = fn
	return err
}

func (p *Parser) scanFuncBooleanAnd(alg *ast.AlgebraBooleanOps, lhs *ast.BooleanArg) (err error) {
	log.Printf("scanFuncBooleanAnd")
	fn := &ast.FuncBooleanAnd{LHS: lhs}
	if err = p.scanInlineStringKeyword(andKeyword); err != nil {
		return err
	}
	if fn.RHS, err = p.scanBooleanArg(); err != nil {
		return err
	}
	alg.And = fn
	return err
}
func (p *Parser) scanFuncBooleanXOR(alg *ast.AlgebraBooleanOps, lhs *ast.BooleanArg) (err error) {
	log.Printf("scanFuncBooleanXOR")
	fn := &ast.FuncBooleanXOR{LHS: lhs}
	if err = p.scanInlineStringKeyword(xorKeyword); err != nil {
		return err
	}
	if fn.RHS, err = p.scanBooleanArg(); err != nil {
		return err
	}
	alg.XOR = fn
	return err
}
func (p *Parser) scanFuncBooleanNot(alg *ast.AlgebraBooleanOps) (err error) {
	log.Printf("scanFuncBooleanNot")
	fn := new(ast.FuncBooleanNot)
	if err = p.scanInlineStringKeyword(notKeyword); err != nil {
		return err
	}
	if fn.Boolean, err = p.scanBooleanArg(); err != nil {
		return err
	}
	alg.Not = fn
	return err
}

// number functions

func (p *Parser) scanFuncNumberAdd(alg *ast.AlgebraNumberOps, lhs *ast.NumberArg) (err error) {
	log.Printf("scanFuncNumberAdd")
	fn := &ast.FuncNumberAdd{LHS: lhs}
	if err = p.scanPlusSymbol(); err != nil {
		return err
	}
	if fn.RHS, err = p.scanNumberArg(); err != nil {
		return err
	}
	alg.Add = fn
	return err
}
func (p *Parser) scanFuncNumberSubtract(alg *ast.AlgebraNumberOps, lhs *ast.NumberArg) (err error) {
	log.Printf("scanFuncNumberSubtract")
	fn := &ast.FuncNumberSubtract{LHS: lhs}
	if err = p.scanMinusSymbol(); err != nil {
		return err
	}
	if fn.RHS, err = p.scanNumberArg(); err != nil {
		return err
	}
	alg.Subtract = fn
	return err
}
func (p *Parser) scanFuncNumberMultiply(alg *ast.AlgebraNumberOps, lhs *ast.NumberArg) (err error) {
	log.Printf("scanFuncNumberMultiply")
	fn := &ast.FuncNumberMultiply{LHS: lhs}
	if err = p.scanMultiplySymbol(); err != nil {
		return err
	}
	if fn.RHS, err = p.scanNumberArg(); err != nil {
		return err
	}
	alg.Multiply = fn
	return err
}
func (p *Parser) scanFuncNumberDivide(alg *ast.AlgebraNumberOps, lhs *ast.NumberArg) (err error) {
	log.Printf("scanFuncNumberDivide")
	fn := &ast.FuncNumberDivide{LHS: lhs}
	if err = p.scanDivideSymbol(); err != nil {
		return err
	}
	if fn.RHS, err = p.scanNumberArg(); err != nil {
		return err
	}
	alg.Divide = fn
	return err
}

// args

func (p *Parser) scanStringArg() (*ast.StringArg, error) {
	log.Printf("scanStringArg")
	tok, lit, err := p.scan()
	if err != nil {
		return nil, err
	}
	p.unscan()

	switch tok {
	case token.String:
		v, err := p.scanString()
		return &ast.StringArg{String: &v}, err

	case token.InlineString:
		switch lit {
		case substringKeyword:
			fn, err := p.scanBuiltInStrFunc()
			return &ast.StringArg{EmitStringFunc: fn}, err
		default:
			return nil, newUnknownKeywordError(lit, substringKeyword)
		}
	default:
		return nil, newSyntaxError(tok, token.String, token.InlineString)
	}
}
func (p *Parser) scanBooleanArg() (*ast.BooleanArg, error) {
	log.Printf("scanBooleanArg")
	tok, lit, err := p.scan()
	if err != nil {
		return nil, err
	}
	p.unscan()

	switch tok {
	case token.InlineString:
		switch lit {
		case regexpKeyword, containsKeyword:
			fn, err := p.scanBuiltInBoolFunc()
			return &ast.BooleanArg{EmitBooleanFunc: fn}, err
		case trueKeyword, falseKeyword:
			v, err := p.scanBoolean()
			return &ast.BooleanArg{Boolean: &v}, err

		default:
			return nil, newUnknownKeywordError(lit, regexpKeyword, containsKeyword, trueKeyword, falseKeyword)
		}
	default:
		return nil, newSyntaxError(tok, token.InlineString)
	}
}
func (p *Parser) scanNumberArg() (*ast.NumberArg, error) {
	log.Printf("scanNumberArg")
	tok, lit, err := p.scan()
	if err != nil {
		return nil, err
	}
	p.unscan()

	switch tok {
	case token.Integer, token.Float:
		v, err := p.scanNumber()
		return &ast.NumberArg{Number: &v}, err

	case token.InlineString:
		switch lit {
		case atofKeyword:
			fn, err := p.scanBuiltInFloatFunc()
			return &ast.NumberArg{EmitNumberFunc: &ast.EmitNumberFunc{Float: fn}}, err

		case lengthKeyword:
			fn, err := p.scanBuiltInIntFunc()
			return &ast.NumberArg{EmitNumberFunc: &ast.EmitNumberFunc{Int: fn}}, err

		default:
			return nil, newUnknownKeywordError(lit, atofKeyword, lengthKeyword)
		}
	default:
		return nil, newSyntaxError(tok, token.Quote, token.InlineString)
	}
}
func (p *Parser) scanIntegerArg() (*ast.IntegerArg, error) {
	log.Printf("scanIntegerArg")
	tok, lit, err := p.scan()
	if err != nil {
		return nil, err
	}
	p.unscan()

	switch tok {
	case token.Integer:
		v, err := p.scanInteger()
		return &ast.IntegerArg{Integer: &v}, err

	case token.InlineString:
		switch lit {
		case lengthKeyword:
			fn, err := p.scanBuiltInIntFunc()
			return &ast.IntegerArg{EmitIntFunc: fn}, err

		default:
			return nil, newUnknownKeywordError(lit, lengthKeyword)
		}
	default:
		return nil, newSyntaxError(tok, token.Quote, token.InlineString)
	}
}

// literals and stuff

func (p *Parser) scanInlineString() (string, error) {
	log.Printf("scanInlineString")
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
	log.Printf("scanString")
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
	log.Printf("scanBoolean")
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
	log.Printf("scanNumber")
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
	log.Printf("scanInteger")
	tok, lit, err := p.scan()
	if err != nil {
		return 0, err
	}
	if tok != token.Integer {
		return 0, newSyntaxError(tok, token.Integer)
	}
	return scanner.ParseInteger(lit)
}

func (p *Parser) scanFunc(keyword string, scanArgs ...func() error) error {
	log.Printf("scanFunc")
	if err := p.scanInlineStringKeyword(keyword); err != nil {
		return err
	}
	if err := p.scanLeftParens(); err != nil {
		return err
	}
	for i, scanArg := range scanArgs {
		if i != 0 {
			if err := p.scanComma(); err != nil {
				return err
			}
		}
		if err := scanArg(); err != nil {
			return err
		}
	}
	return p.scanRightParens()
}

func (p *Parser) scanInlineStringKeyword(kw string) error {
	log.Printf("scanInlineStringKeyword")
	tok, lit, err := p.scan()
	if err != nil {
		return err
	}
	if tok != token.InlineString {
		return newSyntaxError(tok, token.InlineString)
	}
	if lit != kw {
		return newUnknownKeywordError(lit, kw)
	}
	return nil
}

func (p *Parser) scanWhitespace() error {
	log.Printf("scanWhitespace")
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

func (p *Parser) scanComma() error {
	log.Printf("scanComma")
	return p.scanToken(token.Comma)
}
func (p *Parser) scanPipe() error {
	log.Printf("scanPipe")
	return p.scanToken(token.Pipe)
}
func (p *Parser) scanDot() error {
	log.Printf("scanDot")
	return p.scanToken(token.Dot)
}
func (p *Parser) scanLeftBracket() error {
	log.Printf("scanLeftBracket")
	return p.scanToken(token.LeftBracket)
}
func (p *Parser) scanRightBracket() error {
	log.Printf("scanRightBracket")
	return p.scanToken(token.RightBracket)
}
func (p *Parser) scanLeftParens() error {
	log.Printf("scanLeftParens")
	return p.scanToken(token.LeftParens)
}
func (p *Parser) scanRightParens() error {
	log.Printf("scanRightParens")
	return p.scanToken(token.RightParens)
}
func (p *Parser) scanPlusSymbol() error {
	log.Printf("scanPlusSymbol")
	return p.scanToken(token.PlusSymbol)
}
func (p *Parser) scanMinusSymbol() error {
	log.Printf("scanMinusSymbol")
	return p.scanToken(token.MinusSymbol)
}
func (p *Parser) scanMultiplySymbol() error {
	log.Printf("scanMultiplySymbol")
	return p.scanToken(token.MultiplySymbol)
}
func (p *Parser) scanDivideSymbol() error {
	log.Printf("scanDivideSymbol")
	return p.scanToken(token.DivideSymbol)
}

func (p *Parser) scanToken(want token.Token) error {
	log.Printf("scanToken")
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
	log.Printf("scan")
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

func (p *Parser) unscan() {
	log.Printf("unscan")
	p.buf.used = true
}
