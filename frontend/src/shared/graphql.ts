/**
 * Shared GraphQL request utility
 *
 * Authentication is handled via JWT token from csd-core's auth context.
 * Use the useGraphQL hook in components to get an authenticated request function.
 */

export interface GraphQLResponse<T> {
  data?: T;
  errors?: Array<{ message: string }>;
}

/**
 * Extract operation name from a GraphQL query string
 * Matches patterns like: query MyQuery, mutation MyMutation, subscription MySub
 */
function extractOperationName(query: string): string | undefined {
  // Match operation name regardless of leading whitespace
  const match = query.match(/(query|mutation|subscription)\s+(\w+)/);
  return match ? match[2] : undefined;
}

/**
 * Make a GraphQL request with authentication
 *
 * @param endpoint - GraphQL endpoint URL
 * @param query - GraphQL query string
 * @param variables - Optional query variables
 * @param token - Auth token (required for authenticated requests)
 */
export async function graphqlRequest<T>(
  endpoint: string,
  query: string,
  variables?: Record<string, unknown>,
  token?: string
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  // Add Authorization header if token is provided
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  // Extract operation name from query for GraphQL spec compliance
  const operationName = extractOperationName(query) || 'Anonymous';

  // Build URL with query parameters (like csd-core/csd-stocks)
  // v= is a cache buster, service= identifies the caller, operationName= helps with debugging/logging
  const separator = endpoint.includes('?') ? '&' : '?';
  const cacheBuster = Date.now();
  const url = `${endpoint}${separator}v=${cacheBuster}&service=csd-pilote&operationName=${encodeURIComponent(operationName)}`;

  const response = await fetch(url, {
    method: 'POST',
    headers,
    body: JSON.stringify({ query, variables, operationName }),
  });

  const result: GraphQLResponse<T> = await response.json();

  if (result.errors && result.errors.length > 0) {
    throw new Error(result.errors[0].message);
  }

  if (!result.data) {
    throw new Error('No data returned from GraphQL query');
  }

  return result.data;
}

/**
 * Create an authenticated GraphQL request function
 * Use this to create a request function bound to a specific token
 */
export function createAuthenticatedRequest(token: string | undefined) {
  return async function<T>(
    endpoint: string,
    query: string,
    variables?: Record<string, unknown>
  ): Promise<T> {
    return graphqlRequest<T>(endpoint, query, variables, token);
  };
}
