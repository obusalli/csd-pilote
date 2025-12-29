package graphql

// GraphQLRequest represents an incoming GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   interface{}    `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string `json:"message"`
}

// NewErrorResponse creates an error response
func NewErrorResponse(message string) GraphQLResponse {
	return GraphQLResponse{
		Errors: []GraphQLError{{Message: message}},
	}
}

// NewDataResponse creates a data response
func NewDataResponse(data interface{}) GraphQLResponse {
	return GraphQLResponse{
		Data: data,
	}
}
