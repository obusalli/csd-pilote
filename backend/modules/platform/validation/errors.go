package validation

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"csd-pilote/backend/modules/platform/logger"
)

// ErrorCode represents a standardized error code
type ErrorCode string

const (
	ErrCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrCodeValidation       ErrorCode = "VALIDATION_ERROR"
	ErrCodeConflict         ErrorCode = "CONFLICT"
	ErrCodeRateLimit        ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrCodeInternalError    ErrorCode = "INTERNAL_ERROR"
	ErrCodeBadRequest       ErrorCode = "BAD_REQUEST"
	ErrCodeServiceUnavail   ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeOperationFailed  ErrorCode = "OPERATION_FAILED"
	ErrCodePermissionDenied ErrorCode = "PERMISSION_DENIED"
)

// APIError represents a safe error response
type APIError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	TraceID    string    `json:"traceId,omitempty"`
	innerError error     // Never exposed to client
}

func (e *APIError) Error() string {
	return e.Message
}

// generateTraceID generates a unique trace ID for error tracking
func generateTraceID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError() *APIError {
	return &APIError{
		Code:    ErrCodeUnauthorized,
		Message: "Authentication required",
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(permission string) *APIError {
	return &APIError{
		Code:    ErrCodeForbidden,
		Message: fmt.Sprintf("Permission denied: %s", permission),
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *APIError {
	return &APIError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string) *APIError {
	return &APIError{
		Code:    ErrCodeValidation,
		Message: message,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *APIError {
	return &APIError{
		Code:    ErrCodeConflict,
		Message: message,
	}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(operation string) *APIError {
	return &APIError{
		Code:    ErrCodeRateLimit,
		Message: fmt.Sprintf("Too many requests for %s. Please try again later.", operation),
	}
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string) *APIError {
	return &APIError{
		Code:    ErrCodeBadRequest,
		Message: message,
	}
}

// NewInternalError creates an internal error (logs details, returns safe message)
func NewInternalError(err error, context string) *APIError {
	traceID := generateTraceID()

	// Log the actual error internally
	logger.Error("[%s] Internal error in %s: %s", traceID, context, err.Error())

	return &APIError{
		Code:       ErrCodeInternalError,
		Message:    "An internal error occurred. Please try again later.",
		TraceID:    traceID,
		innerError: err,
	}
}

// NewOperationError creates an operation failed error
func NewOperationError(operation string, err error) *APIError {
	traceID := generateTraceID()

	// Log the actual error internally
	logger.Error("[%s] Operation '%s' failed: %s", traceID, operation, err.Error())

	return &APIError{
		Code:       ErrCodeOperationFailed,
		Message:    fmt.Sprintf("Operation '%s' failed. Please try again later.", operation),
		TraceID:    traceID,
		innerError: err,
	}
}

// SanitizeError converts any error to a safe API error
func SanitizeError(err error, context string) *APIError {
	if err == nil {
		return nil
	}

	// If already an APIError, return as-is
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}

	// If validation errors, return validation error
	if validationErr, ok := err.(*ValidationErrors); ok {
		return NewValidationError(validationErr.Error())
	}

	// Check for common safe errors
	errMsg := strings.ToLower(err.Error())

	// Not found errors
	if strings.Contains(errMsg, "record not found") ||
		strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "no rows") {
		return NewNotFoundError("Resource")
	}

	// Duplicate/conflict errors
	if strings.Contains(errMsg, "duplicate") ||
		strings.Contains(errMsg, "unique constraint") ||
		strings.Contains(errMsg, "already exists") {
		return NewConflictError("Resource already exists")
	}

	// Permission errors
	if strings.Contains(errMsg, "permission denied") ||
		strings.Contains(errMsg, "forbidden") {
		return &APIError{
			Code:    ErrCodeForbidden,
			Message: "Permission denied",
		}
	}

	// Connection/service errors
	if strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "unavailable") {
		traceID := generateTraceID()
		logger.Error("[%s] Service unavailable in %s: %s", traceID, context, err.Error())
		return &APIError{
			Code:    ErrCodeServiceUnavail,
			Message: "Service temporarily unavailable. Please try again later.",
			TraceID: traceID,
		}
	}

	// Default: internal error (hide details)
	return NewInternalError(err, context)
}

// SafeErrorMessage returns a safe error message string for GraphQL responses
func SafeErrorMessage(err error, context string) string {
	apiErr := SanitizeError(err, context)
	if apiErr == nil {
		return ""
	}
	if apiErr.TraceID != "" {
		return fmt.Sprintf("%s (trace: %s)", apiErr.Message, apiErr.TraceID)
	}
	return apiErr.Message
}
