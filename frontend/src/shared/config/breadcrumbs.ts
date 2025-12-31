/**
 * Breadcrumb Configuration for CSD-Pilote
 *
 * Centralized breadcrumb category constants with icons matching menu structure.
 * Pages use useBreadcrumb() hook with these constants.
 *
 * @example
 * ```tsx
 * // In ClustersPage.tsx:
 * import { useBreadcrumb } from 'csd_core/Providers';
 * import { BREADCRUMBS } from '../../../../../shared/config/breadcrumbs';
 *
 * useBreadcrumb([BREADCRUMBS.KUBERNETES, { labelKey: 'breadcrumb.clusters' }]);
 * ```
 */

export interface BreadcrumbConfig {
  /** Translation key for the label */
  labelKey: string;
  /** Optional link path (null/undefined = non-clickable) */
  path?: string | null;
  /** Optional icon name (matches menu icons) */
  icon?: string;
}

/**
 * Mapping of breadcrumb labelKeys to their icons
 * Used by useBreadcrumb hook to auto-resolve icons
 */
export const BREADCRUMB_ICONS: Record<string, string> = {
  // Main categories
  'breadcrumb.pilote': 'dashboard',
  'breadcrumb.kubernetes': 'cloud',
  'breadcrumb.libvirt': 'server',
  'breadcrumb.containers': 'cube',

  // Kubernetes pages
  'breadcrumb.clusters': 'cluster',
  'breadcrumb.namespaces': 'folder',
  'breadcrumb.deployments': 'rocket',
  'breadcrumb.pods': 'pod',
  'breadcrumb.services': 'network',

  // Libvirt pages
  'breadcrumb.hypervisors': 'server',
  'breadcrumb.domains': 'desktop',
  'breadcrumb.networks': 'network',
  'breadcrumb.storage': 'database',

  // Containers pages
  'breadcrumb.container_engines': 'cube',

  // Security pages
  'breadcrumb.security': 'shield',
  'breadcrumb.rules': 'list',
  'breadcrumb.profiles': 'folder',
  'breadcrumb.templates': 'template',
};

/**
 * Reusable breadcrumb categories (first level items)
 * Icons match the sidebar menu icons for visual consistency
 */
export const BREADCRUMBS = {
  // Main categories
  PILOTE: {
    labelKey: 'breadcrumb.pilote',
    path: '/pilote',
    icon: 'dashboard',
  } as BreadcrumbConfig,

  KUBERNETES: {
    labelKey: 'breadcrumb.kubernetes',
    icon: 'cloud',
  } as BreadcrumbConfig,

  LIBVIRT: {
    labelKey: 'breadcrumb.libvirt',
    icon: 'server',
  } as BreadcrumbConfig,

  CONTAINERS: {
    labelKey: 'breadcrumb.containers',
    icon: 'cube',
  } as BreadcrumbConfig,

  SECURITY: {
    labelKey: 'breadcrumb.security',
    icon: 'shield',
  } as BreadcrumbConfig,
} as const;
