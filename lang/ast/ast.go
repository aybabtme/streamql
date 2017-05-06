package ast

type AST struct {
	Expr *Expr `json:"expr,omitempty"`
}

type Expr struct {
	// oneof
	Literal        *Literal        `json:"literal,omitempty"`
	Selector       *Selector       `json:"selector,omitempty"`
	UnaryOperator  *UnaryOperator  `json:"unary_operator,omitempty"`
	BinaryOperator *BinaryOperator `json:"binary_operator,omitempty"`
	FuncCall       *FuncCall       `json:"func_call,omitempty"`
	Next           *Expr           `json:"next,omitempty"`
}

type Literal struct {
	// oneof
	Bool   *bool     `json:"bool,omitempty"`
	String *string   `json:"string,omitempty"`
	Int    *int64    `json:"int64,omitempty"`
	Float  *float64  `json:"float64,omitempty"`
	Null   *struct{} `json:"null,omitempty"`
}

type Selector struct {
	// oneof
	Noop   *NoopSelector   `json:"noop,omitempty"`
	Member *MemberSelector `json:"member,omitempty"`
	Slice  *SliceSelector  `json:"slice,omitempty"`
}

type NoopSelector struct{}
type MemberSelector struct {
	Index *Expr     `json:"index,omitempty"`
	Child *Selector `json:"child,omitempty"`
}
type SliceSelector struct {
	From  *Expr     `json:"from,omitempty"`
	To    *Expr     `json:"to,omitempty"`
	Child *Selector `json:"child,omitempty"`
}

type FuncCall struct {
	Name string  `json:"name,omitempty"`
	Args []*Expr `json:"args,omitempty"`
}

type UnaryOperator struct {
	Arg *Expr `json:"arg,omitempty"`
	// oneof
	LogNot *OpLogNot `json:"not,omitempty"`
}

type BinaryOperator struct {
	LHS *Expr `json:"lhs,omitempty"`
	RHS *Expr `json:"rhs,omitempty"`
	// oneof
	LogAnd    *OpLogAnd    `json:"and,omitempty"`
	LogOr     *OpLogOr     `json:"or,omitempty"`
	NumAdd    *OpNumAdd    `json:"add,omitempty"`
	NumSub    *OpNumSub    `json:"sub,omitempty"`
	NumDiv    *OpNumDiv    `json:"div,omitempty"`
	NumMul    *OpNumMul    `json:"mul,omitempty"`
	CmpEq     *OpCmpEq     `json:"eq,omitempty"`
	CmpNotEq  *OpCmpNotEq  `json:"not_eq,omitempty"`
	CmpGt     *OpCmpGt     `json:"gt,omitempty"`
	CmpGtOrEq *OpCmpGtOrEq `json:"gte,omitempty"`
	CmpLs     *OpCmpLs     `json:"ls,omitempty"`
	CmpLsOrEq *OpCmpLsOrEq `json:"lse,omitempty"`
}

type OpLogNot struct{}
type OpLogAnd struct{}
type OpLogOr struct{}
type OpNumAdd struct{}
type OpNumSub struct{}
type OpNumDiv struct{}
type OpNumMul struct{}
type OpCmpEq struct{}
type OpCmpNotEq struct{}
type OpCmpGt struct{}
type OpCmpGtOrEq struct{}
type OpCmpLs struct{}
type OpCmpLsOrEq struct{}
