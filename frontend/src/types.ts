/**
 * Types for CSD Pilote
 */

export interface ServiceConfig {
  id: string;
  slug: string;
  name: string;
  graphqlUrl: string;
  coreGraphqlUrl?: string;
}

export interface ServiceAppInfo {
  name: string;
  version: string;
  author: string;
  description: string;
  copyright: string;
  components?: Array<{
    name: string;
    version: string;
    description: string;
  }>;
  openSourceLibraries?: Array<{
    name: string;
    version: string;
    description: string;
    layer: string;
    license: string;
  }>;
}
