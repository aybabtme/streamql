package spec

const (
	tokDot          = "."
	tokLeftBracket  = "["
	tokRightBracket = "]"
	tokLeftParens   = "("
	tokRightParens  = ")"
	tokColon        = ":"
	tokPipe         = "|"

	tokLogNot = "!"
	tokLogAnd = "&&"
	tokLogOr  = "||"

	tokNumAdd = "+"
	tokNumSub = "-"
	tokNumMul = "*"
	tokNumDiv = "/"

	tokCmpEq     = "=="
	tokCmpNotEq  = "!="
	tokCmpGt     = ">"
	tokCmpGtOrEq = ">="
	tokCmpLs     = "<"
	tokCmpLsOrEq = "<="

	tokWS         = "`ws`"
	tokNull       = "`null`"
	tokBool       = "`bool`"
	tokIdentifier = "`id`"
	tokString     = "`string`"
	tokInt        = "`int`"
	tokFloat      = "`float`"
)

type tok struct {
	id  string
	lit string
}

func (t *tok) String() string { return t.id }
