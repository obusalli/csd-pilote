/**
 * ServiceConfigContext
 *
 * Context to provide API configuration to all components in csd-pilote.
 *
 * Configuration sources (in priority order):
 * 1. ServiceConfig from csd-core (federated mode) - for pilote URL
 * 2. YAML config file (public/csd-pilote.yaml or csd-pilote.production.yaml)
 */

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import type { ServiceConfig } from './types';
import { getConfig, getCachedConfig } from './config/api';

// API URLs available to components
interface ApiUrls {
  coreGraphQL: string;
  piloteGraphQL: string;
}

interface ServiceConfigContextType {
  serviceConfig: ServiceConfig | null;
  apiUrls: ApiUrls | null;
  loading: boolean;
  error: string | null;
}

const ServiceConfigContext = createContext<ServiceConfigContextType>({
  serviceConfig: null,
  apiUrls: null,
  loading: true,
  error: null,
});

interface ServiceConfigProviderProps {
  serviceConfig?: ServiceConfig;
  children: ReactNode;
}

export const ServiceConfigProvider: React.FC<ServiceConfigProviderProps> = ({
  serviceConfig,
  children,
}) => {
  const [apiUrls, setApiUrls] = useState<ApiUrls | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadApiConfig = async () => {
      try {
        // In federated mode, csd-core provides both URLs via serviceConfig
        if (serviceConfig?.graphqlUrl && serviceConfig?.coreGraphqlUrl) {
          console.log('[ServiceConfig] Federated mode - pilote:', serviceConfig.graphqlUrl, 'core:', serviceConfig.coreGraphqlUrl);
          setApiUrls({
            coreGraphQL: serviceConfig.coreGraphqlUrl,
            piloteGraphQL: serviceConfig.graphqlUrl,
          });
        } else {
          // Standalone mode - load from YAML
          console.log('[ServiceConfig] Standalone mode - loading YAML config...');
          const config = await getConfig();
          setApiUrls({
            coreGraphQL: config.core_graphql_url,
            piloteGraphQL: config.pilote_graphql_url,
          });
        }
        setError(null);
      } catch (err) {
        console.error('[ServiceConfig] Failed to load config:', err);
        setError(err instanceof Error ? err.message : 'Failed to load configuration');
      } finally {
        setLoading(false);
      }
    };

    loadApiConfig();
  }, [serviceConfig?.graphqlUrl, serviceConfig?.coreGraphqlUrl]);

  return (
    <ServiceConfigContext.Provider value={{ serviceConfig: serviceConfig || null, apiUrls, loading, error }}>
      {children}
    </ServiceConfigContext.Provider>
  );
};

/**
 * Hook to access full service configuration
 */
export const useServiceConfig = (): ServiceConfigContextType => {
  return useContext(ServiceConfigContext);
};

/**
 * Hook to get csd-pilote GraphQL URL
 */
export const useGraphQLUrl = (): string => {
  const { apiUrls } = useServiceConfig();
  if (!apiUrls) {
    // Fallback to cached config if available
    const cached = getCachedConfig();
    return cached?.pilote_graphql_url || '';
  }
  return apiUrls.piloteGraphQL;
};

/**
 * Hook to get csd-core GraphQL URL
 */
export const useCoreGraphQLUrl = (): string => {
  const { apiUrls } = useServiceConfig();
  if (!apiUrls) {
    const cached = getCachedConfig();
    return cached?.core_graphql_url || '';
  }
  return apiUrls.coreGraphQL;
};

export { ServiceConfig };
