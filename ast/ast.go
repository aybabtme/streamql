package ast

type FiltersStmt struct {
	Filters []*FilterStmt
}

type FilterStmt struct {
	Selectors []*SelectorStmt
}

type SelectorStmt struct {
	// oneof
	Object *ObjectSelectorStmt
	Array  *ArraySelectorStmt
}

type ObjectSelectorStmt struct {
	// Member on which the Child will be applied.
	Member string

	Child *SelectorStmt
}

type ArraySelectorStmt struct {
	// oneof
	Each      *EachSelectorStmt
	RangeEach *RangeEachSelectorStmt
	Index     *IndexSelectorStmt

	// Member on which the Child will be applied.
	Child *SelectorStmt
}

type (
	EachSelectorStmt      struct{}
	RangeEachSelectorStmt struct{ From, To int }
	IndexSelectorStmt     struct{ Index int }
)
