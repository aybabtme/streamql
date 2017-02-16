//line parser.y:2
package spec

import __yyfmt__ "fmt"

//line parser.y:2
import (
	"io"
	// "log"
)

var implicitSliceIdx = struct{}{}

//line parser.y:55
type yySymType struct {
	yys  int
	node interface{}

	curID int
	cur   tok
	err   error
}

const Dot = 57346
const LeftBracket = 57347
const RightBracket = 57348
const LeftParens = 57349
const RightParens = 57350
const Colon = 57351
const Pipe = 57352
const Comma = 57353
const Null = 57354
const Bool = 57355
const Identifier = 57356
const String = 57357
const Int = 57358
const Float = 57359
const LogOr = 57360
const LogAnd = 57361
const LogNot = 57362
const CmpEq = 57363
const CmpNotEq = 57364
const CmpGt = 57365
const CmpGtOrEq = 57366
const CmpLs = 57367
const CmpLsOrEq = 57368
const NumAdd = 57369
const NumSub = 57370
const NumMul = 57371
const NumDiv = 57372

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"Dot",
	"LeftBracket",
	"RightBracket",
	"LeftParens",
	"RightParens",
	"Colon",
	"Pipe",
	"Comma",
	"Null",
	"Bool",
	"Identifier",
	"String",
	"Int",
	"Float",
	"LogOr",
	"LogAnd",
	"LogNot",
	"CmpEq",
	"CmpNotEq",
	"CmpGt",
	"CmpGtOrEq",
	"CmpLs",
	"CmpLsOrEq",
	"NumAdd",
	"NumSub",
	"NumMul",
	"NumDiv",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line parser.y:120

func cast(y yyLexer) *AST { return y.(*Lexer).parseResult.(*AST) }

func Parse(r io.Reader) (ast *AST, err error) {
	ast = new(AST)
	lex := NewLexerWithInit(r, func(l *Lexer) { l.parseResult = ast })
	// defer func() {
	//     r := recover()
	//     if r != nil {
	//         err = fmt.Errorf("%v", r)
	//     }
	// }()
	yyParse(lex)
	return ast, err
}

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
	-1, 44,
	21, 0,
	22, 0,
	-2, 35,
	-1, 45,
	21, 0,
	22, 0,
	-2, 36,
	-1, 46,
	23, 0,
	24, 0,
	25, 0,
	26, 0,
	-2, 37,
	-1, 47,
	23, 0,
	24, 0,
	25, 0,
	26, 0,
	-2, 38,
	-1, 48,
	23, 0,
	24, 0,
	25, 0,
	26, 0,
	-2, 39,
	-1, 49,
	23, 0,
	24, 0,
	25, 0,
	26, 0,
	-2, 40,
}

const yyNprod = 45
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 371

var yyAct = [...]int{

	50, 57, 22, 58, 2, 18, 59, 24, 25, 26,
	27, 28, 29, 20, 21, 23, 22, 32, 33, 35,
	67, 37, 38, 39, 40, 41, 42, 43, 44, 45,
	46, 47, 48, 49, 12, 54, 60, 15, 36, 62,
	23, 22, 11, 7, 16, 8, 9, 10, 5, 56,
	13, 21, 23, 22, 63, 6, 61, 4, 14, 66,
	69, 70, 31, 3, 34, 74, 73, 1, 0, 75,
	78, 30, 79, 51, 52, 0, 80, 84, 85, 0,
	0, 0, 87, 88, 89, 71, 0, 90, 72, 17,
	20, 21, 23, 22, 0, 0, 0, 19, 18, 0,
	24, 25, 26, 27, 28, 29, 20, 21, 23, 22,
	64, 0, 0, 65, 17, 0, 0, 0, 0, 0,
	0, 0, 19, 18, 0, 24, 25, 26, 27, 28,
	29, 20, 21, 23, 22, 86, 0, 0, 0, 17,
	0, 0, 0, 0, 0, 0, 0, 19, 18, 0,
	24, 25, 26, 27, 28, 29, 20, 21, 23, 22,
	83, 0, 0, 0, 17, 0, 0, 0, 0, 0,
	0, 0, 19, 18, 0, 24, 25, 26, 27, 28,
	29, 20, 21, 23, 22, 82, 0, 0, 0, 17,
	0, 0, 0, 0, 0, 0, 0, 19, 18, 0,
	24, 25, 26, 27, 28, 29, 20, 21, 23, 22,
	77, 0, 0, 0, 17, 0, 0, 0, 0, 0,
	0, 0, 19, 18, 0, 24, 25, 26, 27, 28,
	29, 20, 21, 23, 22, 17, 68, 0, 0, 0,
	0, 0, 0, 19, 18, 0, 24, 25, 26, 27,
	28, 29, 20, 21, 23, 22, 17, 0, 0, 0,
	0, 0, 0, 0, 19, 18, 0, 24, 25, 26,
	27, 28, 29, 20, 21, 23, 22, 24, 25, 26,
	27, 28, 29, 20, 21, 23, 22, 12, 0, 53,
	15, 0, 55, 0, 0, 11, 7, 16, 8, 9,
	10, 0, 0, 13, 12, 0, 81, 15, 0, 0,
	0, 14, 11, 7, 16, 8, 9, 10, 0, 0,
	13, 12, 0, 76, 15, 0, 0, 0, 14, 11,
	7, 16, 8, 9, 10, 0, 0, 13, 12, 0,
	0, 15, 0, 0, 0, 14, 11, 7, 16, 8,
	9, 10, 0, 0, 13, 0, 0, 0, 0, 0,
	0, 0, 14, 26, 27, 28, 29, 20, 21, 23,
	22,
}
var yyPact = [...]int{

	334, -1000, 246, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 57, 334, 334, 334, 31, 334, 334, 334,
	334, 334, 334, 334, 334, 334, 334, 334, 334, 334,
	69, 283, 256, 11, 41, 246, 334, 246, 256, -14,
	23, 11, -1000, -28, 340, 340, 63, 63, 63, 63,
	-1000, -8, 30, 69, 104, 334, -1000, 12, 225, 69,
	69, 79, 334, -1000, 69, 317, 204, -1000, 334, -1000,
	-1000, 69, 300, 179, -1000, 154, 69, 69, -1000, -1000,
	129, 69, 69, 69, -1000, -1000, 69, -1000, -1000, -1000,
	-1000,
}
var yyPgo = [...]int{

	0, 67, 3, 63, 57, 48, 55, 0, 1,
}
var yyR1 = [...]int{

	0, 1, 1, 2, 2, 2, 2, 2, 3, 3,
	3, 3, 3, 4, 4, 4, 4, 4, 4, 4,
	7, 7, 7, 7, 7, 7, 7, 5, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
	5, 5, 6, 8, 8,
}
var yyR2 = [...]int{

	0, 1, 0, 1, 1, 1, 1, 3, 1, 1,
	1, 1, 1, 1, 3, 4, 5, 7, 6, 6,
	3, 3, 4, 6, 5, 5, 0, 2, 3, 3,
	2, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 4, 1, 3,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, 13, 15, 16,
	17, 12, 4, 20, 28, 7, 14, 10, 19, 18,
	27, 28, 30, 29, 21, 22, 23, 24, 25, 26,
	14, 5, -2, -2, -5, -2, 7, -2, -2, -2,
	-2, -2, -2, -2, -2, -2, -2, -2, -2, -2,
	-7, 4, 5, 6, -2, 9, 8, -8, -2, 14,
	6, -2, 9, -7, 6, 9, -2, 8, 11, -7,
	-7, 6, 9, -2, -7, -2, 6, 6, -8, -7,
	-2, 6, 6, 6, -7, -7, 6, -7, -7, -7,
	-7,
}
var yyDef = [...]int{

	2, -2, 1, 3, 4, 5, 6, 8, 9, 10,
	11, 12, 13, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	26, 0, 27, 30, 5, 0, 0, 7, 28, 29,
	31, 32, 33, 34, -2, -2, -2, -2, -2, -2,
	14, 0, 0, 26, 0, 0, 41, 0, 43, 26,
	26, 0, 0, 15, 26, 0, 0, 42, 0, 20,
	21, 26, 0, 0, 16, 0, 26, 26, 44, 22,
	0, 26, 26, 26, 19, 18, 26, 25, 24, 17,
	23,
}
var yyTok1 = [...]int{

	1,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30,
}
var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:64
		{
			cast(yylex).Expr = expr(yyDollar[1])
		}
	case 3:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:67
		{
			yyVAL = literal(yyDollar[1])
		}
	case 4:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:68
		{
			yyVAL = selector(yyDollar[1])
		}
	case 5:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:69
		{
			yyVAL = operator(yyDollar[1])
		}
	case 6:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:70
		{
			yyVAL = funcCall(yyDollar[1])
		}
	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:71
		{
			yyVAL = pipe(yyDollar[1], yyDollar[3])
		}
	case 8:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:74
		{
			yyVAL = emitBool(yyDollar[1])
		}
	case 9:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:75
		{
			yyVAL = emitString(yyDollar[1])
		}
	case 10:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:76
		{
			yyVAL = emitInt(yyDollar[1])
		}
	case 11:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:77
		{
			yyVAL = emitFloat(yyDollar[1])
		}
	case 12:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:78
		{
			yyVAL = emitNull(yyDollar[1])
		}
	case 13:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:81
		{
			yyVAL = emitNopSelector()
		}
	case 14:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:82
		{
			yyVAL = emitMemberSelector(yyDollar[2], yyDollar[3])
		}
	case 15:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line parser.y:83
		{
			yyVAL = emitSliceSelectorEach(yyDollar[4])
		}
	case 16:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line parser.y:84
		{
			yyVAL = emitMemberSelector(yyDollar[3], yyDollar[5])
		}
	case 17:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line parser.y:85
		{
			yyVAL = emitSliceSelector(yyDollar[3], yyDollar[5], yyDollar[7])
		}
	case 18:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line parser.y:86
		{
			yyVAL = emitSliceSelector(yySymType{node: implicitSliceIdx}, yyDollar[4], yyDollar[6])
		}
	case 19:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line parser.y:87
		{
			yyVAL = emitSliceSelector(yyDollar[3], yySymType{node: implicitSliceIdx}, yyDollar[6])
		}
	case 20:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:89
		{
			yyVAL = emitMemberSelector(yyDollar[2], yyDollar[3])
		}
	case 21:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:90
		{
			yyVAL = emitSliceSelectorEach(yyDollar[3])
		}
	case 22:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line parser.y:91
		{
			yyVAL = emitMemberSelector(yyDollar[2], yyDollar[4])
		}
	case 23:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line parser.y:92
		{
			yyVAL = emitSliceSelector(yyDollar[2], yyDollar[4], yyDollar[6])
		}
	case 24:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line parser.y:93
		{
			yyVAL = emitSliceSelector(yySymType{node: implicitSliceIdx}, yyDollar[3], yyDollar[4])
		}
	case 25:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line parser.y:94
		{
			yyVAL = emitSliceSelector(yyDollar[2], yySymType{node: implicitSliceIdx}, yyDollar[5])
		}
	case 26:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line parser.y:95
		{
			yyVAL = yySymType{}
		}
	case 27:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser.y:97
		{
			yyVAL = emitOpNot(yyDollar[2])
		}
	case 28:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:98
		{
			yyVAL = emitOpAnd(yyDollar[1], yyDollar[3])
		}
	case 29:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:99
		{
			yyVAL = emitOpOr(yyDollar[1], yyDollar[3])
		}
	case 30:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser.y:100
		{
			yyVAL = emitOpNeg(yyDollar[2])
		}
	case 31:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:101
		{
			yyVAL = emitOpAdd(yyDollar[1], yyDollar[3])
		}
	case 32:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:102
		{
			yyVAL = emitOpSub(yyDollar[1], yyDollar[3])
		}
	case 33:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:103
		{
			yyVAL = emitOpDiv(yyDollar[1], yyDollar[3])
		}
	case 34:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:104
		{
			yyVAL = emitOpMul(yyDollar[1], yyDollar[3])
		}
	case 35:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:105
		{
			yyVAL = emitOpEq(yyDollar[1], yyDollar[3])
		}
	case 36:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:106
		{
			yyVAL = emitOpNotEq(yyDollar[1], yyDollar[3])
		}
	case 37:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:107
		{
			yyVAL = emitOpGt(yyDollar[1], yyDollar[3])
		}
	case 38:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:108
		{
			yyVAL = emitOpGtOrEq(yyDollar[1], yyDollar[3])
		}
	case 39:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:109
		{
			yyVAL = emitOpLs(yyDollar[1], yyDollar[3])
		}
	case 40:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:110
		{
			yyVAL = emitOpLsOrEq(yyDollar[1], yyDollar[3])
		}
	case 41:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:111
		{
			yyVAL = yyDollar[2]
		}
	case 42:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line parser.y:114
		{
			yyVAL = emitFuncCall(yyDollar[1], yyDollar[3])
		}
	case 43:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser.y:116
		{
			yyVAL = emitArg(yyDollar[1])
		}
	case 44:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser.y:117
		{
			yyVAL = emitArgs(yyDollar[1], yyDollar[3])
		}
	}
	goto yystack /* stack new state and value */
}
