package ast

type FiltersStmt struct {
	Filters []*FilterStmt `json:"filters"`
}

type FilterStmt struct {
	Selectors []*SelectorStmt `json:"selectors"`
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
		From int `json:"from,omitempty"`
		To   int `json:"to,omitempty"`
	}
	IndexSelectorStmt struct {
		Index int `json:"index,omitempty"`
	}
)
