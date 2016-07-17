package ast

type FiltersStmt struct {
	Filters []*FilterStmt
}

type FilterStmt struct {
	Selectors []*SelectorStmt
}

type SelectorStmt struct {
	// optional
	Member *MemberStmt

	// oneof
	Each      *EachStmt
	RangeEach *RangeEachStmt
	Index     *IndexStmt
}

type MemberStmt struct {
	FieldName string
}

type EachStmt struct{}

type RangeEachStmt struct {
	From int
	To   int
}

type IndexStmt struct {
	Index int
}
