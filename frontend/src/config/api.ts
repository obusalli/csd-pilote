/**
 * Centralized API configuration for csd-pilote frontend
 *
 * Configuration is loaded from YAML files:
 * - Development: public/csd-pilote.yaml (absolute URLs)
 * - Production: public/csd-pilote.production.yaml (relative URLs for nginx)
 *
 * In federated mode, csd-core can override PILOTE.GRAPHQL via ServiceConfig.
 */

import yaml from 'js-yaml';

// Final parsed config
interface PiloteConfig {
  core_graphql_url: string;
  pilote_graphql_url: string;
}

// Raw YAML structure (matches common/frontend pattern)
interface RawConfig {
  common?: {
    graphql?: {
      core_url?: string;
      pilote_url?: string;
    };
  };
  frontend?: {
    dev?: {
      enabled?: boolean;
      port?: number;
      host?: string;
    };
  };
  // Legacy flat format support
  core_graphql_url?: string;
  pilote_graphql_url?: string;
}

interface ConfigCache {
  config: PiloteConfig;
}

let cachedConfig: ConfigCache | null = null;

/**
 * Load configuration from YAML file
 * Note: This is only used in standalone mode. In federated mode,
 * URLs are provided via serviceConfig from csd-core.
 */
async function loadConfig(): Promise<PiloteConfig> {
  if (cachedConfig) {
    return cachedConfig.config;
  }

  // Use different config file for production
  const isProduction = window.location.protocol === 'https:' ||
    (!window.location.hostname.includes('localhost') && !window.location.hostname.includes('127.0.0.1'));

  const configFile = isProduction
    ? '/csd-pilote.production.yaml'
    : '/csd-pilote.yaml';

  const response = await fetch(configFile);

  if (!response.ok) {
    throw new Error(`Failed to load configuration file ${configFile}: ${response.statusText}`);
  }

  const yamlText = await response.text();
  const rawConfig = yaml.load(yamlText) as RawConfig;

  // Parse new format (common.graphql) or legacy format
  const config: PiloteConfig = {
    core_graphql_url: rawConfig.common?.graphql?.core_url || rawConfig.core_graphql_url || '',
    pilote_graphql_url: rawConfig.common?.graphql?.pilote_url || rawConfig.pilote_graphql_url || '',
  };

  if (!config.core_graphql_url || !config.pilote_graphql_url) {
    throw new Error(`Missing required URLs in ${configFile}`);
  }

  cachedConfig = { config };
  return config;
}

/**
 * Get csd-core GraphQL URL
 */
export async function getCoreGraphQLUrl(): Promise<string> {
  const config = await loadConfig();
  return config.core_graphql_url;
}

/**
 * Get csd-pilote GraphQL URL
 */
export async function getPiloteGraphQLUrl(): Promise<string> {
  const config = await loadConfig();
  return config.pilote_graphql_url;
}

/**
 * Get full configuration
 */
export async function getConfig(): Promise<PiloteConfig> {
  return loadConfig();
}

// For synchronous access after initial load (use with caution)
export function getCachedConfig(): PiloteConfig | null {
  return cachedConfig?.config || null;
}
