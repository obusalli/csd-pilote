/**
 * Routes - Module Federation exposed module
 *
 * This file is exposed to csd-core via Module Federation.
 * It contains all the routes for csd-pilote that will be mounted
 * under the routePath configured in the service (e.g., /pilote/*).
 */

import React from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { ServiceConfigProvider, ServiceConfig } from './ServiceConfigContext';

// Lazy load pages for better performance
const DashboardPage = React.lazy(() =>
  import('@pilot/dashboard/DashboardPage').then((m) => ({ default: m.DashboardPage }))
);
const ClustersPage = React.lazy(() =>
  import('@pilot/clusters/ClustersPage').then((m) => ({ default: m.ClustersPage }))
);
const ClusterDetailPage = React.lazy(() =>
  import('@pilot/clusters/ClusterDetailPage').then((m) => ({ default: m.ClusterDetailPage }))
);
const HypervisorsPage = React.lazy(() =>
  import('@pilot/hypervisors/HypervisorsPage').then((m) => ({ default: m.HypervisorsPage }))
);
const HypervisorDetailPage = React.lazy(() =>
  import('@pilot/hypervisors/HypervisorDetailPage').then((m) => ({ default: m.HypervisorDetailPage }))
);
const ContainerEnginesPage = React.lazy(() =>
  import('@pilot/containers/ContainerEnginesPage').then((m) => ({ default: m.ContainerEnginesPage }))
);
const ContainerEngineDetailPage = React.lazy(() =>
  import('@pilot/containers/ContainerEngineDetailPage').then((m) => ({ default: m.ContainerEngineDetailPage }))
);

// Security pages
const SecurityRulesPage = React.lazy(() =>
  import('@pilot/security/RulesPage').then((m) => ({ default: m.RulesPage }))
);
const SecurityProfilesPage = React.lazy(() =>
  import('@pilot/security/ProfilesPage').then((m) => ({ default: m.ProfilesPage }))
);
const SecurityTemplatesPage = React.lazy(() =>
  import('@pilot/security/TemplatesPage').then((m) => ({ default: m.TemplatesPage }))
);
const SecurityDeploymentsPage = React.lazy(() =>
  import('@pilot/security/DeploymentsPage').then((m) => ({ default: m.DeploymentsPage }))
);

// Loading fallback
const Loading: React.FC = () => (
  <div style={{
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    height: '200px',
    color: '#666',
  }}>
    Loading...
  </div>
);

// Route configuration
const ROUTES: Record<string, React.LazyExoticComponent<React.FC>> = {
  '/pilote/dashboard': DashboardPage,
  '/pilote/kubernetes/clusters': ClustersPage,
  '/pilote/libvirt/hypervisors': HypervisorsPage,
  '/pilote/containers/engines': ContainerEnginesPage,
  '/pilote/security/rules': SecurityRulesPage,
  '/pilote/security/profiles': SecurityProfilesPage,
  '/pilote/security/templates': SecurityTemplatesPage,
  '/pilote/security/deployments': SecurityDeploymentsPage,
};

// Dynamic routes (with parameters)
const DYNAMIC_ROUTES: Array<{
  pattern: RegExp;
  component: React.LazyExoticComponent<React.FC>;
}> = [
  {
    pattern: /^\/pilote\/kubernetes\/clusters\/([^/]+)$/,
    component: ClusterDetailPage,
  },
  {
    pattern: /^\/pilote\/libvirt\/hypervisors\/([^/]+)$/,
    component: HypervisorDetailPage,
  },
  {
    pattern: /^\/pilote\/containers\/engines\/([^/]+)$/,
    component: ContainerEngineDetailPage,
  },
];

interface PiloteRoutesProps {
  serviceConfig?: ServiceConfig;
}

const PiloteRoutes: React.FC<PiloteRoutesProps> = ({ serviceConfig }) => {
  const location = useLocation();
  const navigate = useNavigate();
  const path = location.pathname;

  // Find matching static route
  let Component = ROUTES[path];

  // Check dynamic routes if no static match
  if (!Component) {
    for (const route of DYNAMIC_ROUTES) {
      if (route.pattern.test(path)) {
        Component = route.component;
        break;
      }
    }
  }

  // Handle redirects
  React.useEffect(() => {
    if (path === '/pilote' || path === '/pilote/') {
      navigate('/pilote/dashboard', { replace: true });
      return;
    }
    if (!Component) {
      navigate('/pilote/dashboard', { replace: true });
    }
  }, [path, navigate, Component]);

  if (!Component) {
    return <Loading />;
  }

  return (
    <ServiceConfigProvider serviceConfig={serviceConfig}>
      <React.Suspense fallback={<Loading />}>
        <Component />
      </React.Suspense>
    </ServiceConfigProvider>
  );
};

export default PiloteRoutes;
