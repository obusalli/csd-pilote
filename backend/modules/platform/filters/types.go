package filters

// FilterOperator defines the comparison operation
type FilterOperator string

const (
	OpEquals           FilterOperator = "eq"
	OpNotEquals        FilterOperator = "neq"
	OpGreaterThan      FilterOperator = "gt"
	OpGreaterThanEqual FilterOperator = "gte"
	OpLessThan         FilterOperator = "lt"
	OpLessThanEqual    FilterOperator = "lte"
	OpContains         FilterOperator = "contains"
	OpStartsWith       FilterOperator = "startsWith"
	OpEndsWith         FilterOperator = "endsWith"
	OpIn               FilterOperator = "in"
	OpNotIn            FilterOperator = "notIn"
	OpIsNull           FilterOperator = "isNull"
	OpIsNotNull        FilterOperator = "isNotNull"
	OpBetween          FilterOperator = "between"
)

// LogicalOperator defines how conditions are combined
type LogicalOperator string

const (
	LogicalAnd LogicalOperator = "AND"
	LogicalOr  LogicalOperator = "OR"
)

// FilterCondition represents a single filter condition
type FilterCondition struct {
	Field    string         `json:"field"`
	Operator FilterOperator `json:"operator"`
	Value    interface{}    `json:"value"`
	Values   []interface{}  `json:"values,omitempty"` // For "in", "notIn", "between"
}

// FilterGroup represents a group of conditions with a logical operator
type FilterGroup struct {
	Logic      LogicalOperator   `json:"logic"`
	Conditions []FilterCondition `json:"conditions,omitempty"`
	Groups     []FilterGroup     `json:"groups,omitempty"` // Nested groups
}

// AdvancedFilter represents the complete advanced filter structure
type AdvancedFilter struct {
	Logic      LogicalOperator   `json:"logic"`
	Conditions []FilterCondition `json:"conditions,omitempty"`
	Groups     []FilterGroup     `json:"groups,omitempty"`
}

// SortDirection defines the sort order
type SortDirection string

const (
	SortAsc  SortDirection = "ASC"
	SortDesc SortDirection = "DESC"
)

// SortField represents a field to sort by
type SortField struct {
	Field     string        `json:"field"`
	Direction SortDirection `json:"direction"`
}

// QueryOptions represents pagination and sorting options
type QueryOptions struct {
	Limit   int         `json:"limit,omitempty"`
	Offset  int         `json:"offset,omitempty"`
	OrderBy []SortField `json:"orderBy,omitempty"`
}
