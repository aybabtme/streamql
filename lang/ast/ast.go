package ast

type FiltersStmt struct {
	Filters []*FilterStmt `json:"filters"`
}

type FilterStmt struct {
	Funcs []*FuncStmt `json:"funcs"`
}

type FuncStmt struct {
	// oneof
	EmitFunc *EmitFuncStmt `json:"emit_func,omitempty"`
	Selector *SelectorStmt `json:"selector,omitempty"`
}

type EmitFuncStmt struct {
	// oneof
	EmitBooleanFunc *EmitBooleanFunc `json:"emit_boolean_func,omitempty"`
	EmitStringFunc  *EmitStringFunc  `json:"emit_string_func,omitempty"`
	EmitNumberFunc  *EmitNumberFunc  `json:"emit_number_func,omitempty"`
	EmitAnyFunc     *EmitAnyFunc     `json:"emit_any_func,omitempty"`
}

type SelectorStmt struct {
	// oneof
	Object *ObjectSelectorStmt `json:"object,omitempty"`
	Array  *ArraySelectorStmt  `json:"array,omitempty"`
}

type ObjectSelectorStmt struct {
	// Member on which the Child will be applied.
	Member string        `json:"member,omitempty"`
	Child  *SelectorStmt `json:"child,omitempty"`
}

type ArraySelectorStmt struct {
	// oneof
	Each      *EachSelectorStmt      `json:"each,omitempty"`
	RangeEach *RangeEachSelectorStmt `json:"range_each,omitempty"`
	Index     *IndexSelectorStmt     `json:"index,omitempty"`

	// Member on which the Child will be applied.
	Child *SelectorStmt `json:"child,omitempty"`
}

type (
	EachSelectorStmt      struct{}
	RangeEachSelectorStmt struct {
		From *IntegerArg `json:"from,omitempty"`
		To   *IntegerArg `json:"to,omitempty"`
	}
	IndexSelectorStmt struct {
		Index *IntegerArg `json:"index,omitempty"`
	}
)

type EmitBooleanFunc struct {
	// oneof
	Literal        *BooleanArg         `json:"literal,omitempty"`
	StringContains *FuncStringContains `json:"string_contains,omitempty"`
	StringRegexp   *FuncStringRegexp   `json:"string_regexp,omitempty"`
	Algebra        *AlgebraBooleanOps  `json:"algebra,omitempty"`
}
type EmitStringFunc struct {
	// oneof
	Literal      *StringArg        `json:"literal,omitempty"`
	StringSubStr *FuncStringSubStr `json:"string_substr,omitempty"`
}
type EmitNumberFunc struct {
	// oneof
	Int     *EmitIntFunc      `json:"int,omitempty"`
	Float   *EmitFloatFunc    `json:"float,omitempty"`
	Algebra *AlgebraNumberOps `json:"algebra,omitempty"`
}
type EmitAnyFunc struct {
	// oneof
	AnySelect *FuncAnySelect `json:"select,omitempty"`
}

type EmitIntFunc struct {
	// oneof
	Literal      *IntegerArg       `json:"literal,omitempty"`
	StringLength *FuncStringLength `json:"string_length,omitempty"`
}
type EmitFloatFunc struct {
	// oneof
	Literal    *NumberArg      `json:"literal,omitempty"`
	StringAtof *FuncStringAtof `json:"string_atof,omitempty"`
}

type AlgebraBooleanOps struct {
	// oneof
	Or  *FuncBooleanOr  `json:"or,omitempty"`
	And *FuncBooleanAnd `json:"and,omitempty"`
	Not *FuncBooleanNot `json:"not,omitempty"`
	XOR *FuncBooleanXOR `json:"xor,omitempty"`
}

type AlgebraNumberOps struct {
	// oneof
	Add      *FuncNumberAdd      `json:"add,omitempty"`
	Subtract *FuncNumberSubtract `json:"subtract,omitempty"`
	Multiply *FuncNumberMultiply `json:"multiply,omitempty"`
	Divide   *FuncNumberDivide   `json:"divide,omitempty"`
}

type StringArg struct {
	// oneof
	String         *string         `json:"string,omitempty"`
	EmitStringFunc *EmitStringFunc `json:"emit_string_func,omitempty"`
	Selector       *SelectorStmt   `json:"selector,omitempty"`
}

type BooleanArg struct {
	// oneof
	Boolean         *bool            `json:"boolean,omitempty"`
	EmitBooleanFunc *EmitBooleanFunc `json:"emit_boolean_func,omitempty"`
	Selector        *SelectorStmt    `json:"selector,omitempty"`
}

type NumberArg struct {
	// oneof
	Number         *float64        `json:"number,omitempty"`
	EmitNumberFunc *EmitNumberFunc `json:"emit_number_func,omitempty"`
	Selector       *SelectorStmt   `json:"selector,omitempty"`
}

type IntegerArg struct {
	// oneof
	Integer     *int64        `json:"integer,omitempty"`
	EmitIntFunc *EmitIntFunc  `json:"emit_int_func,omitempty"`
	Selector    *SelectorStmt `json:"selector,omitempty"`
}

type FuncStringContains struct {
	SubString *StringArg `json:"substring"`
	Target    *StringArg `json:"target"`
}
type FuncStringRegexp struct {
	Expression *StringArg `json:"expression"`
	Target     *StringArg `json:"target"`
}
type FuncStringSubStr struct {
	String *StringArg  `json:"string"`
	From   *IntegerArg `json:"from"`
	To     *IntegerArg `json:"to"`
}
type FuncStringLength struct {
	String *StringArg `json:"string"`
}
type FuncStringAtof struct {
	String *StringArg `json:"string"`
}
type FuncAnySelect struct {
	Condition *BooleanArg `json:"condition"`
}

type FuncBooleanOr struct {
	LHS *BooleanArg `json:"lhs"`
	RHS *BooleanArg `json:"rhs"`
}
type FuncBooleanAnd struct {
	LHS *BooleanArg `json:"lhs"`
	RHS *BooleanArg `json:"rhs"`
}
type FuncBooleanNot struct {
	Boolean *BooleanArg `json:"boolean"`
}
type FuncBooleanXOR struct {
	LHS *BooleanArg `json:"lhs"`
	RHS *BooleanArg `json:"rhs"`
}

type FuncNumberAdd struct {
	LHS *NumberArg `json:"lhs"`
	RHS *NumberArg `json:"rhs"`
}
type FuncNumberSubtract struct {
	LHS *NumberArg `json:"lhs"`
	RHS *NumberArg `json:"rhs"`
}
type FuncNumberMultiply struct {
	LHS *NumberArg `json:"lhs"`
	RHS *NumberArg `json:"rhs"`
}
type FuncNumberDivide struct {
	LHS *NumberArg `json:"lhs"`
	RHS *NumberArg `json:"rhs"`
}
