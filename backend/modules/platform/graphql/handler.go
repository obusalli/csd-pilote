package graphql

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	csdcore "csd-pilote/backend/modules/platform/csd-core"
	"csd-pilote/backend/modules/platform/middleware"
	"csd-pilote/backend/modules/platform/ratelimit"
	"csd-pilote/backend/modules/platform/validation"
)

// Handler handles GraphQL requests
type Handler struct {
	csdCoreClient *csdcore.Client
}

// NewHandler creates a new GraphQL handler
func NewHandler(csdCoreClient *csdcore.Client) *Handler {
	return &Handler{
		csdCoreClient: csdCoreClient,
	}
}

// MaxRequestBodySize is the maximum allowed request body size (1MB)
const MaxRequestBodySize = 1 << 20 // 1 MB

// Pre-compiled regex patterns for GraphQL parsing (performance optimization)
var (
	graphqlOperationPattern = regexp.MustCompile(`^(query|mutation|subscription)\s+(\w+)`)
	graphqlTypePattern      = regexp.MustCompile(`^(query|mutation)\s*`)
	graphqlFieldPattern     = regexp.MustCompile(`\{\s*(\w+)`)
)

// ServeHTTP handles GraphQL HTTP requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(NewErrorResponse("Method not allowed"))
		return
	}

	// Limit request body size to prevent DoS
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodySize)

	var req GraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(NewErrorResponse("Invalid request body"))
		return
	}

	// Parse the operation from the query
	opType, opName := parseOperation(req.Query, req.OperationName)

	// Try to handle locally first
	if handled := h.handleLocal(w, r, opType, opName, req.Variables); handled {
		return
	}

	// Forward to csd-core if not handled locally
	h.forwardToCSDCore(w, r, req)
}

// handleLocal tries to handle the operation locally
func (h *Handler) handleLocal(w http.ResponseWriter, r *http.Request, opType, opName string, variables map[string]interface{}) bool {
	var op *Operation
	var ok bool

	switch opType {
	case "query":
		op, ok = GetQuery(opName)
	case "mutation":
		op, ok = GetMutation(opName)
	default:
		return false
	}

	if !ok {
		return false
	}

	// Check rate limit for ALL operations (queries and mutations)
	if err := ratelimit.CheckRateLimit(r, opName); err != nil {
		if rlErr, ok := err.(*ratelimit.RateLimitError); ok {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(NewErrorResponseWithCode(
				"RATE_LIMIT_EXCEEDED",
				validation.NewRateLimitError(rlErr.Operation).Message,
			))
			return true
		}
	}

	// Check permission if required
	if op.Permission != "" {
		token, hasToken := middleware.GetTokenFromContext(r.Context())
		if !hasToken {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(NewErrorResponseWithCode(
				string(validation.ErrCodeUnauthorized),
				validation.NewUnauthorizedError().Message,
			))
			return true
		}

		if h.csdCoreClient != nil {
			// Check for system.admin first (grants all permissions)
			hasAdmin, _ := h.csdCoreClient.CheckPermission(r.Context(), token, "system.admin")

			if !hasAdmin {
				// Check for specific permission
				hasPermission, err := h.csdCoreClient.CheckPermission(r.Context(), token, op.Permission)
				if err != nil || !hasPermission {
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(NewErrorResponseWithCode(
						string(validation.ErrCodeForbidden),
						validation.NewForbiddenError(op.Permission).Message,
					))
					return true
				}
			}
		}
	}

	// Execute handler
	op.Handler(r.Context(), w, variables)
	return true
}

// forwardToCSDCore forwards the request to csd-core
func (h *Handler) forwardToCSDCore(w http.ResponseWriter, r *http.Request, req GraphQLRequest) {
	if h.csdCoreClient == nil {
		json.NewEncoder(w).Encode(NewErrorResponse("Operation not found and csd-core not configured"))
		return
	}

	token, _ := middleware.GetTokenFromContext(r.Context())

	// Extract operation name if not provided in request
	operationName := req.OperationName
	if operationName == "" {
		operationName = extractOperationNameFromQuery(req.Query)
	}

	resp, err := h.csdCoreClient.ExecuteWithName(r.Context(), token, operationName, req.Query, req.Variables)
	if err != nil {
		// Sanitize error to prevent information disclosure
		json.NewEncoder(w).Encode(NewErrorResponse(validation.SafeErrorMessage(err, "csd-core proxy")))
		return
	}

	// Forward the response as-is
	result := map[string]interface{}{}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		json.NewEncoder(w).Encode(GraphQLResponse{Data: resp.Data})
		return
	}
	json.NewEncoder(w).Encode(GraphQLResponse{Data: result})
}

// extractOperationNameFromQuery extracts the operation name from a GraphQL query
func extractOperationNameFromQuery(query string) string {
	query = strings.TrimSpace(query)
	matches := graphqlOperationPattern.FindStringSubmatch(query)
	if len(matches) > 2 {
		return matches[2]
	}
	return ""
}

// parseOperation extracts operation type and name from a GraphQL query
func parseOperation(query string, operationName string) (opType string, opName string) {
	query = strings.TrimSpace(query)

	matches := graphqlTypePattern.FindStringSubmatch(query)
	if len(matches) > 1 {
		opType = matches[1]
	} else {
		opType = "query"
	}

	// Try to find the first field name
	fieldMatches := graphqlFieldPattern.FindStringSubmatch(query)
	if len(fieldMatches) > 1 {
		opName = fieldMatches[1]
	}

	if operationName != "" {
		if opName == "" {
			opName = operationName
		}
	}

	return opType, opName
}
