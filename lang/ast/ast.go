package ast

type AST struct {
	Expr *Expr `json:"expr,omitempty"`
}

type Expr struct {
	// oneof
	Literal  *Literal  `json:"literal,omitempty"`
	Selector *Selector `json:"selector,omitempty"`
	Operator *Operator `json:"operator,omitempty"`
	FuncCall *FuncCall `json:"func_call,omitempty"`
	Next     *Expr     `json:"next,omitempty"`
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

type Operator struct {
	// oneof
	LogNot    *OperandLogNot    `json:"not,omitempty"`
	LogAnd    *OperandLogAnd    `json:"and,omitempty"`
	LogOr     *OperandLogOr     `json:"or,omitempty"`
	NumAdd    *OperandNumAdd    `json:"add,omitempty"`
	NumSub    *OperandNumSub    `json:"sub,omitempty"`
	NumDiv    *OperandNumDiv    `json:"div,omitempty"`
	NumMul    *OperandNumMul    `json:"mul,omitempty"`
	CmpEq     *OperandCmpEq     `json:"eq,omitempty"`
	CmpNotEq  *OperandCmpNotEq  `json:"not_eq,omitempty"`
	CmpGt     *OperandCmpGt     `json:"gt,omitempty"`
	CmpGtOrEq *OperandCmpGtOrEq `json:"gte,omitempty"`
	CmpLs     *OperandCmpLs     `json:"ls,omitempty"`
	CmpLsOrEq *OperandCmpLsOrEq `json:"lse,omitempty"`
}

type OperandLogNot struct {
	Arg *Expr `json:"arg,omitempty"`
}
type OperandLogAnd struct {
	LHS *Expr `json:"lhs,omitempty"`
	RHS *Expr `json:"rhs,omitempty"`
}
type OperandLogOr struct {
	LHS *Expr `json:"lhs,omitempty"`
	RHS *Expr `json:"rhs,omitempty"`
}
type OperandNumAdd struct {
	LHS *Expr `json:"lhs,omitempty"`
	RHS *Expr `json:"rhs,omitempty"`
}
type OperandNumSub struct {
	LHS *Expr `json:"lhs,omitempty"`
	RHS *Expr `json:"rhs,omitempty"`
}
type OperandNumDiv struct {
	LHS *Expr `json:"lhs,omitempty"`
	RHS *Expr `json:"rhs,omitempty"`
}
type OperandNumMul struct {
	LHS *Expr `json:"lhs,omitempty"`
	RHS *Expr `json:"rhs,omitempty"`
}
type OperandCmpEq struct {
	LHS *Expr `json:"lhs,omitempty"`
	RHS *Expr `json:"rhs,omitempty"`
}
type OperandCmpNotEq struct {
	LHS *Expr `json:"lhs,omitempty"`
	RHS *Expr `json:"rhs,omitempty"`
}
type OperandCmpGt struct {
	LHS *Expr `json:"lhs,omitempty"`
	RHS *Expr `json:"rhs,omitempty"`
}
type OperandCmpGtOrEq struct {
	LHS *Expr `json:"lhs,omitempty"`
	RHS *Expr `json:"rhs,omitempty"`
}
type OperandCmpLs struct {
	LHS *Expr `json:"lhs,omitempty"`
	RHS *Expr `json:"rhs,omitempty"`
}
type OperandCmpLsOrEq struct {
	LHS *Expr `json:"lhs,omitempty"`
	RHS *Expr `json:"rhs,omitempty"`
}
