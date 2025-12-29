/**
 * Hook for authenticated GraphQL requests
 *
 * Gets the access token from csd-core's sessionStorage and provides
 * request functions that automatically include authentication.
 *
 * Note: We use sessionStorage directly rather than csd-core's useAuth hook
 * to avoid module federation context issues. The token is the same - csd-core
 * stores it in sessionStorage as 'csd_access_token'.
 */

import { useCallback } from 'react';
import { useGraphQLUrl, useCoreGraphQLUrl } from '../../ServiceConfigContext';
import { graphqlRequest } from '../graphql';

/**
 * Get the access token from sessionStorage
 * csd-core stores the JWT token as 'csd_access_token'
 */
function getAccessToken(): string | undefined {
  try {
    return sessionStorage.getItem('csd_access_token') || undefined;
  } catch {
    return undefined;
  }
}

/**
 * Hook that provides authenticated GraphQL request functions
 *
 * Usage:
 * ```tsx
 * const { request, coreRequest } = useGraphQL();
 *
 * // For csd-pilote backend
 * const data = await request<MyDataType>(query, variables);
 *
 * // For csd-core backend
 * const coreData = await coreRequest<CoreDataType>(query, variables);
 * ```
 */
export function useGraphQL() {
  const graphqlUrl = useGraphQLUrl();
  const coreGraphqlUrl = useCoreGraphQLUrl();

  // Request function for csd-pilote backend
  // Token is read at request time to ensure we always use the current token
  const request = useCallback(
    async <T>(query: string, variables?: Record<string, unknown>): Promise<T> => {
      if (!graphqlUrl) {
        throw new Error('GraphQL URL not configured');
      }
      // Get fresh token at request time from sessionStorage
      const token = getAccessToken();
      return graphqlRequest<T>(graphqlUrl, query, variables, token);
    },
    [graphqlUrl]
  );

  // Request function for csd-core backend
  const coreRequest = useCallback(
    async <T>(query: string, variables?: Record<string, unknown>): Promise<T> => {
      if (!coreGraphqlUrl) {
        throw new Error('Core GraphQL URL not configured');
      }
      // Get fresh token at request time from sessionStorage
      const token = getAccessToken();
      return graphqlRequest<T>(coreGraphqlUrl, query, variables, token);
    },
    [coreGraphqlUrl]
  );

  return {
    request,
    coreRequest,
    graphqlUrl,
    coreGraphqlUrl,
    // Token is always available if user is logged in (csd-core handles auth)
    isAuthenticated: !!getAccessToken(),
  };
}

export default useGraphQL;
