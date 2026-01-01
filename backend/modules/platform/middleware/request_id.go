package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestIDKey is the context key for request ID
type requestIDKey struct{}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID already exists in header
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}

		// Set request ID in response header
		w.Header().Set("X-Request-ID", reqID)

		// Add request ID to context
		ctx := context.WithValue(r.Context(), requestIDKey{}, reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestIDFromContext returns the request ID from context
func GetRequestIDFromContext(ctx context.Context) (string, bool) {
	reqID, ok := ctx.Value(requestIDKey{}).(string)
	return reqID, ok
}
