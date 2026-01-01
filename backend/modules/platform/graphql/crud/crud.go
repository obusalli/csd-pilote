// Package crud provides generic CRUD handler helpers for GraphQL operations
package crud

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"csd-pilote/backend/modules/platform/graphql"
	"csd-pilote/backend/modules/platform/middleware"
)

// HandlerContext contains common context values extracted from a request
type HandlerContext struct {
	Ctx      context.Context
	TenantID uuid.UUID
	UserID   uuid.UUID
	Token    string
}

// ExtractTenantContext extracts tenant context for read operations
// Returns nil and writes error if extraction fails
func ExtractTenantContext(ctx context.Context, w http.ResponseWriter) *HandlerContext {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return nil
	}

	return &HandlerContext{
		Ctx:      ctx,
		TenantID: tenantID,
	}
}

// ExtractFullContext extracts full context for write operations (includes user)
// Returns nil and writes error if extraction fails
func ExtractFullContext(ctx context.Context, w http.ResponseWriter) *HandlerContext {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return nil
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return nil
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	return &HandlerContext{
		Ctx:      ctx,
		TenantID: tenantID,
		UserID:   user.UserID,
		Token:    token,
	}
}

// ParseID extracts and validates a UUID from variables
// Returns zero UUID and writes error if parsing fails
func ParseID(w http.ResponseWriter, variables map[string]interface{}, key string) (uuid.UUID, bool) {
	id, err := graphql.ParseUUID(variables, key)
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return uuid.Nil, false
	}
	return id, true
}

// ParseIDs extracts and validates multiple UUIDs from variables
// Returns nil and writes error if parsing fails
func ParseIDs(w http.ResponseWriter, variables map[string]interface{}, key string) ([]uuid.UUID, bool) {
	idsRaw, ok := variables[key].([]interface{})
	if !ok || len(idsRaw) == 0 {
		graphql.WriteValidationError(w, key+" is required and must be a non-empty array")
		return nil, false
	}

	ids := make([]uuid.UUID, 0, len(idsRaw))
	for _, idRaw := range idsRaw {
		idStr, ok := idRaw.(string)
		if !ok {
			graphql.WriteValidationError(w, "invalid ID format in "+key)
			return nil, false
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			graphql.WriteValidationError(w, "invalid UUID: "+idStr)
			return nil, false
		}
		ids = append(ids, id)
	}

	return ids, true
}

// GetResult is a generic result for single-item get operations
type GetResult[T any] struct {
	Item T
	Key  string // The key name for the response (e.g., "cluster", "hypervisor")
}

// WriteGetResult writes a successful get response
func WriteGetResult[T any](w http.ResponseWriter, key string, item T) {
	graphql.WriteSuccess(w, map[string]interface{}{
		key: item,
	})
}

// ListResult is a generic result for list operations
type ListResult[T any] struct {
	Items    []T
	Count    int64
	ItemsKey string // The key name for items (e.g., "clusters")
	CountKey string // The key name for count (e.g., "clustersCount")
}

// WriteListResult writes a successful list response
func WriteListResult[T any](w http.ResponseWriter, itemsKey, countKey string, items []T, count int64) {
	graphql.WriteSuccess(w, map[string]interface{}{
		itemsKey: items,
		countKey: count,
	})
}

// WriteDeleteResult writes a successful delete response
func WriteDeleteResult(w http.ResponseWriter, success bool) {
	graphql.WriteSuccess(w, map[string]interface{}{
		"success": success,
	})
}

// WriteBulkDeleteResult writes a successful bulk delete response
func WriteBulkDeleteResult(w http.ResponseWriter, deleted int64) {
	graphql.WriteSuccess(w, map[string]interface{}{
		"deleted": deleted,
	})
}

// WriteCreateResult writes a successful create response
func WriteCreateResult[T any](w http.ResponseWriter, key string, item T) {
	graphql.WriteSuccess(w, map[string]interface{}{
		key: item,
	})
}

// WriteUpdateResult writes a successful update response
func WriteUpdateResult[T any](w http.ResponseWriter, key string, item T) {
	graphql.WriteSuccess(w, map[string]interface{}{
		key: item,
	})
}

// HandleError writes an error response with context
func HandleError(w http.ResponseWriter, err error, operation string) {
	graphql.WriteError(w, err, operation)
}

// HandleValidationError writes a validation error response
func HandleValidationError(w http.ResponseWriter, message string) {
	graphql.WriteValidationError(w, message)
}
