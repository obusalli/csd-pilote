package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware recovers from panics and returns a 500 error
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID for tracing
				reqID, _ := GetRequestIDFromContext(r.Context())

				// Log the panic with stack trace
				log.Printf("[PANIC] request_id=%s path=%s error=%v\n%s",
					reqID, r.URL.Path, err, debug.Stack())

				// Return 500 error
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"errors": []map[string]string{
						{"message": "Internal server error"},
					},
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
