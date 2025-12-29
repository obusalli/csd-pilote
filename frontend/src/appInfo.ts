/**
 * CSD Pilote - Application Information
 *
 * This module is exposed via Module Federation as ./AppInfo
 * and is loaded by csd-core's About dialog to display service information.
 */

import type { ServiceAppInfo } from './types';

export type { ServiceAppInfo } from './types';

const appInfo: ServiceAppInfo = {
  name: 'CSD Pilote',
  version: '0.1.0',
  author: 'Koraisoft',
  description: 'Infrastructure management module for CSD Core. Provides management of Kubernetes clusters, Libvirt hypervisors, and Docker/Podman container engines.',
  copyright: `© ${new Date().getFullYear()} Koraisoft. All rights reserved.`,

  components: [
    {
      name: 'CSD Pilote Frontend',
      version: '0.1.0',
      description: 'React-based infrastructure management interface',
    },
    {
      name: 'CSD Pilote Backend',
      version: '0.1.0',
      description: 'Go GraphQL API for infrastructure management',
    },
  ],

  openSourceLibraries: [
    {
      name: 'React',
      version: '19.0.0',
      description: 'A JavaScript library for building user interfaces',
      layer: 'Frontend',
      license: 'MIT',
    },
    {
      name: 'React DOM',
      version: '19.0.0',
      description: 'React package for working with the DOM',
      layer: 'Frontend',
      license: 'MIT',
    },
    {
      name: 'React Router DOM',
      version: '6.x',
      description: 'Declarative routing for React applications',
      layer: 'Frontend',
      license: 'MIT',
    },
    {
      name: 'MUI (Material-UI)',
      version: '7.x',
      description: "React components implementing Google's Material Design",
      layer: 'Frontend',
      license: 'MIT',
    },
    {
      name: 'Apollo Client',
      version: '3.x',
      description: 'Comprehensive state management library for JavaScript with GraphQL',
      layer: 'Frontend',
      license: 'MIT',
    },
    {
      name: 'GraphQL',
      version: '16.8.1',
      description: 'GraphQL query language and runtime',
      layer: 'Shared',
      license: 'MIT',
    },
    {
      name: 'Vite',
      version: '6.0.11',
      description: 'Next generation frontend build tool',
      layer: 'Build',
      license: 'MIT',
    },
    {
      name: 'TypeScript',
      version: '5.9.3',
      description: 'Typed superset of JavaScript that compiles to plain JavaScript',
      layer: 'Build',
      license: 'Apache-2.0',
    },
  ],
};

export default appInfo;
