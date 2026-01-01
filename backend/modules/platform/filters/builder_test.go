package filters

import (
	"testing"
)

func TestIsValidColumnName(t *testing.T) {
	tests := []struct {
		name     string
		column   string
		expected bool
	}{
		{"simple name", "name", true},
		{"with underscore", "first_name", true},
		{"with number", "column1", true},
		{"table.column", "users.name", true},
		{"starts with underscore", "_private", true},
		{"empty string", "", false},
		{"starts with number", "1column", false},
		{"contains hyphen", "first-name", false},
		{"contains space", "first name", false},
		{"too long", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false}, // 65 chars
		{"max length", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true},   // 64 chars
		{"sql injection attempt", "name; DROP TABLE", false},
		{"special characters", "name$", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidColumnName(tt.column)
			if result != tt.expected {
				t.Errorf("isValidColumnName(%q) = %v, want %v", tt.column, result, tt.expected)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"camelCase", "camel_case"},
		{"PascalCase", "pascal_case"},
		{"already_snake", "already_snake"},
		{"lowercase", "lowercase"},
		{"ID", "i_d"},
		{"userID", "user_i_d"},
		{"HTTPStatus", "h_t_t_p_status"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestQueryBuilder_GetColumnName(t *testing.T) {
	t.Run("with field mapping", func(t *testing.T) {
		qb := NewQueryBuilder(nil).
			WithFieldMapping("createdAt", "created_at")

		result := qb.getColumnName("createdAt")
		if result != "created_at" {
			t.Errorf("getColumnName(createdAt) = %q, want 'created_at'", result)
		}
	})

	t.Run("without mapping converts to snake case", func(t *testing.T) {
		qb := NewQueryBuilder(nil)

		result := qb.getColumnName("firstName")
		if result != "first_name" {
			t.Errorf("getColumnName(firstName) = %q, want 'first_name'", result)
		}
	})

	t.Run("strict mode rejects unmapped fields", func(t *testing.T) {
		qb := NewQueryBuilder(nil).
			WithStrictMode().
			WithFieldMapping("name", "name")

		// Mapped field works
		if result := qb.getColumnName("name"); result != "name" {
			t.Errorf("getColumnName(name) = %q, want 'name'", result)
		}

		// Unmapped field rejected
		if result := qb.getColumnName("unmapped"); result != "" {
			t.Errorf("getColumnName(unmapped) in strict mode = %q, want ''", result)
		}
	})

	t.Run("rejects invalid column names", func(t *testing.T) {
		qb := NewQueryBuilder(nil)

		// Even with mapping, invalid column name is rejected
		qb.WithFieldMapping("bad", "bad; DROP TABLE")
		if result := qb.getColumnName("bad"); result != "" {
			t.Errorf("getColumnName should reject invalid mapped column, got %q", result)
		}
	})
}

func TestQueryBuilder_WithFieldMappings(t *testing.T) {
	qb := NewQueryBuilder(nil).
		WithFieldMappings(map[string]string{
			"createdAt": "created_at",
			"updatedAt": "updated_at",
		})

	if qb.getColumnName("createdAt") != "created_at" {
		t.Error("createdAt should map to created_at")
	}
	if qb.getColumnName("updatedAt") != "updated_at" {
		t.Error("updatedAt should map to updated_at")
	}
}

func TestQueryBuilder_BuildCondition(t *testing.T) {
	qb := NewQueryBuilder(nil)

	tests := []struct {
		name          string
		condition     FilterCondition
		expectedSQL   string
		expectedEmpty bool
	}{
		{
			name: "equals",
			condition: FilterCondition{
				Field:    "name",
				Operator: OpEquals,
				Value:    "test",
			},
			expectedSQL: "name = ?",
		},
		{
			name: "not equals",
			condition: FilterCondition{
				Field:    "status",
				Operator: OpNotEquals,
				Value:    "active",
			},
			expectedSQL: "status != ?",
		},
		{
			name: "greater than",
			condition: FilterCondition{
				Field:    "age",
				Operator: OpGreaterThan,
				Value:    18,
			},
			expectedSQL: "age > ?",
		},
		{
			name: "contains",
			condition: FilterCondition{
				Field:    "description",
				Operator: OpContains,
				Value:    "test",
			},
			expectedSQL: "description ILIKE ?",
		},
		{
			name: "starts with",
			condition: FilterCondition{
				Field:    "name",
				Operator: OpStartsWith,
				Value:    "prefix",
			},
			expectedSQL: "name ILIKE ?",
		},
		{
			name: "is null",
			condition: FilterCondition{
				Field:    "deleted_at",
				Operator: OpIsNull,
			},
			expectedSQL: "deleted_at IS NULL",
		},
		{
			name: "is not null",
			condition: FilterCondition{
				Field:    "created_at",
				Operator: OpIsNotNull,
			},
			expectedSQL: "created_at IS NOT NULL",
		},
		{
			name: "invalid field rejected",
			condition: FilterCondition{
				Field:    "1invalid",
				Operator: OpEquals,
				Value:    "test",
			},
			expectedEmpty: true,
		},
		{
			name: "unknown operator",
			condition: FilterCondition{
				Field:    "name",
				Operator: "invalid_op",
				Value:    "test",
			},
			expectedEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, _ := qb.buildCondition(tt.condition)
			if tt.expectedEmpty {
				if sql != "" {
					t.Errorf("Expected empty SQL, got %q", sql)
				}
			} else {
				if sql != tt.expectedSQL {
					t.Errorf("buildCondition() = %q, want %q", sql, tt.expectedSQL)
				}
			}
		})
	}
}

func TestQueryBuilder_BuildCondition_In(t *testing.T) {
	qb := NewQueryBuilder(nil)

	t.Run("in with values array", func(t *testing.T) {
		condition := FilterCondition{
			Field:    "status",
			Operator: OpIn,
			Values:   []interface{}{"active", "pending"},
		}
		sql, args := qb.buildCondition(condition)
		if sql != "status IN (?, ?)" {
			t.Errorf("Expected 'status IN (?, ?)', got %q", sql)
		}
		if len(args) != 2 {
			t.Errorf("Expected 2 args, got %d", len(args))
		}
	})

	t.Run("in with value as array", func(t *testing.T) {
		condition := FilterCondition{
			Field:    "status",
			Operator: OpIn,
			Value:    []interface{}{"active", "pending"},
		}
		sql, args := qb.buildCondition(condition)
		if sql != "status IN (?, ?)" {
			t.Errorf("Expected 'status IN (?, ?)', got %q", sql)
		}
		if len(args) != 2 {
			t.Errorf("Expected 2 args, got %d", len(args))
		}
	})

	t.Run("in with empty values", func(t *testing.T) {
		condition := FilterCondition{
			Field:    "status",
			Operator: OpIn,
			Values:   []interface{}{},
		}
		sql, _ := qb.buildCondition(condition)
		if sql != "" {
			t.Errorf("Expected empty SQL for empty values, got %q", sql)
		}
	})
}

func TestQueryBuilder_BuildCondition_Between(t *testing.T) {
	qb := NewQueryBuilder(nil)

	t.Run("between with two values", func(t *testing.T) {
		condition := FilterCondition{
			Field:    "age",
			Operator: OpBetween,
			Values:   []interface{}{18, 65},
		}
		sql, args := qb.buildCondition(condition)
		if sql != "age BETWEEN ? AND ?" {
			t.Errorf("Expected 'age BETWEEN ? AND ?', got %q", sql)
		}
		if len(args) != 2 {
			t.Errorf("Expected 2 args, got %d", len(args))
		}
	})

	t.Run("between with insufficient values", func(t *testing.T) {
		condition := FilterCondition{
			Field:    "age",
			Operator: OpBetween,
			Values:   []interface{}{18},
		}
		sql, _ := qb.buildCondition(condition)
		if sql != "" {
			t.Errorf("Expected empty SQL for insufficient values, got %q", sql)
		}
	})
}

func TestFilterOperators(t *testing.T) {
	operators := []FilterOperator{
		OpEquals,
		OpNotEquals,
		OpGreaterThan,
		OpGreaterThanEqual,
		OpLessThan,
		OpLessThanEqual,
		OpContains,
		OpStartsWith,
		OpEndsWith,
		OpIn,
		OpNotIn,
		OpIsNull,
		OpIsNotNull,
		OpBetween,
	}

	for _, op := range operators {
		if op == "" {
			t.Error("Operator should not be empty")
		}
	}
}

func TestLogicalOperators(t *testing.T) {
	if LogicalAnd == "" {
		t.Error("LogicalAnd should not be empty")
	}
	if LogicalOr == "" {
		t.Error("LogicalOr should not be empty")
	}
}

func TestSortDirection(t *testing.T) {
	if SortAsc == "" {
		t.Error("SortAsc should not be empty")
	}
	if SortDesc == "" {
		t.Error("SortDesc should not be empty")
	}
}
