package filters

import (
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// QueryBuilder builds GORM queries from advanced filters
type QueryBuilder struct {
	db            *gorm.DB
	fieldMappings map[string]string // Maps JSON field names to DB column names
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(db *gorm.DB) *QueryBuilder {
	return &QueryBuilder{
		db:            db,
		fieldMappings: make(map[string]string),
	}
}

// WithFieldMapping adds a field mapping from JSON name to DB column
func (qb *QueryBuilder) WithFieldMapping(jsonField, dbColumn string) *QueryBuilder {
	qb.fieldMappings[jsonField] = dbColumn
	return qb
}

// WithFieldMappings adds multiple field mappings
func (qb *QueryBuilder) WithFieldMappings(mappings map[string]string) *QueryBuilder {
	for k, v := range mappings {
		qb.fieldMappings[k] = v
	}
	return qb
}

// getColumnName returns the DB column name for a JSON field
func (qb *QueryBuilder) getColumnName(field string) string {
	if col, ok := qb.fieldMappings[field]; ok {
		return col
	}
	// Convert camelCase to snake_case by default
	return toSnakeCase(field)
}

// toSnakeCase converts camelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// ApplyFilter applies an advanced filter to a GORM query
func (qb *QueryBuilder) ApplyFilter(query *gorm.DB, filter *AdvancedFilter) *gorm.DB {
	if filter == nil {
		return query
	}

	whereClause, args := qb.buildGroup(FilterGroup{
		Logic:      filter.Logic,
		Conditions: filter.Conditions,
		Groups:     filter.Groups,
	})

	if whereClause != "" {
		query = query.Where(whereClause, args...)
	}

	return query
}

// ApplyFilterJSON applies a filter from JSON input
func (qb *QueryBuilder) ApplyFilterJSON(query *gorm.DB, filterJSON interface{}) (*gorm.DB, error) {
	if filterJSON == nil {
		return query, nil
	}

	// Handle if it's already a map
	filterMap, ok := filterJSON.(map[string]interface{})
	if !ok {
		return query, fmt.Errorf("invalid filter format")
	}

	// Convert map to AdvancedFilter struct
	jsonBytes, err := json.Marshal(filterMap)
	if err != nil {
		return query, fmt.Errorf("failed to marshal filter: %w", err)
	}

	var filter AdvancedFilter
	if err := json.Unmarshal(jsonBytes, &filter); err != nil {
		return query, fmt.Errorf("failed to parse filter: %w", err)
	}

	return qb.ApplyFilter(query, &filter), nil
}

// buildGroup builds a WHERE clause from a filter group
func (qb *QueryBuilder) buildGroup(group FilterGroup) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	// Build conditions
	for _, cond := range group.Conditions {
		clause, condArgs := qb.buildCondition(cond)
		if clause != "" {
			clauses = append(clauses, clause)
			args = append(args, condArgs...)
		}
	}

	// Build nested groups
	for _, nestedGroup := range group.Groups {
		nestedClause, nestedArgs := qb.buildGroup(nestedGroup)
		if nestedClause != "" {
			clauses = append(clauses, "("+nestedClause+")")
			args = append(args, nestedArgs...)
		}
	}

	if len(clauses) == 0 {
		return "", nil
	}

	logic := " AND "
	if group.Logic == LogicalOr {
		logic = " OR "
	}

	return strings.Join(clauses, logic), args
}

// buildCondition builds a single WHERE condition
func (qb *QueryBuilder) buildCondition(cond FilterCondition) (string, []interface{}) {
	column := qb.getColumnName(cond.Field)

	switch cond.Operator {
	case OpEquals:
		return fmt.Sprintf("%s = ?", column), []interface{}{cond.Value}

	case OpNotEquals:
		return fmt.Sprintf("%s != ?", column), []interface{}{cond.Value}

	case OpGreaterThan:
		return fmt.Sprintf("%s > ?", column), []interface{}{cond.Value}

	case OpGreaterThanEqual:
		return fmt.Sprintf("%s >= ?", column), []interface{}{cond.Value}

	case OpLessThan:
		return fmt.Sprintf("%s < ?", column), []interface{}{cond.Value}

	case OpLessThanEqual:
		return fmt.Sprintf("%s <= ?", column), []interface{}{cond.Value}

	case OpContains:
		return fmt.Sprintf("%s ILIKE ?", column), []interface{}{"%" + fmt.Sprint(cond.Value) + "%"}

	case OpStartsWith:
		return fmt.Sprintf("%s ILIKE ?", column), []interface{}{fmt.Sprint(cond.Value) + "%"}

	case OpEndsWith:
		return fmt.Sprintf("%s ILIKE ?", column), []interface{}{"%" + fmt.Sprint(cond.Value)}

	case OpIn:
		values := cond.Values
		if values == nil && cond.Value != nil {
			if arr, ok := cond.Value.([]interface{}); ok {
				values = arr
			}
		}
		if len(values) == 0 {
			return "", nil
		}
		placeholders := make([]string, len(values))
		for i := range values {
			placeholders[i] = "?"
		}
		return fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", ")), values

	case OpNotIn:
		values := cond.Values
		if values == nil && cond.Value != nil {
			if arr, ok := cond.Value.([]interface{}); ok {
				values = arr
			}
		}
		if len(values) == 0 {
			return "", nil
		}
		placeholders := make([]string, len(values))
		for i := range values {
			placeholders[i] = "?"
		}
		return fmt.Sprintf("%s NOT IN (%s)", column, strings.Join(placeholders, ", ")), values

	case OpIsNull:
		return fmt.Sprintf("%s IS NULL", column), nil

	case OpIsNotNull:
		return fmt.Sprintf("%s IS NOT NULL", column), nil

	case OpBetween:
		values := cond.Values
		if len(values) >= 2 {
			return fmt.Sprintf("%s BETWEEN ? AND ?", column), []interface{}{values[0], values[1]}
		}
		return "", nil

	default:
		return "", nil
	}
}

// ApplySort applies sorting to a query
func (qb *QueryBuilder) ApplySort(query *gorm.DB, sortFields []SortField) *gorm.DB {
	for _, sf := range sortFields {
		column := qb.getColumnName(sf.Field)
		dir := "ASC"
		if sf.Direction == SortDesc {
			dir = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", column, dir))
	}
	return query
}

// ApplyPagination applies limit and offset to a query
func (qb *QueryBuilder) ApplyPagination(query *gorm.DB, limit, offset int) *gorm.DB {
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	return query
}

// ApplyOptions applies full query options (filter, sort, pagination)
func (qb *QueryBuilder) ApplyOptions(query *gorm.DB, filter *AdvancedFilter, options *QueryOptions) *gorm.DB {
	query = qb.ApplyFilter(query, filter)

	if options != nil {
		if len(options.OrderBy) > 0 {
			query = qb.ApplySort(query, options.OrderBy)
		}
		query = qb.ApplyPagination(query, options.Limit, options.Offset)
	}

	return query
}
