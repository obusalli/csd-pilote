import React, { useState, useEffect, useCallback } from 'react';
import {
  CSDDetailPage,
  CSDStack,
  CSDTypography,
  CSDPaper,
  CSDChip,
  CSDTabs,
  CSDTab,
  CSDActionButton,
  CSDDataGrid,
  CSDCircularProgress,
  CSDAlert,
  CSDIcon,
  CSDInfoGrid,
  CSDInfoItem,
} from 'csd_core/UI';
import {
  useBreadcrumb,
  useParams,
  useSnackbar,
  useTranslation,
  formatDate,
} from 'csd_core/Providers';
import { useGraphQL } from '../../../shared/hooks/useGraphQL';
import { BREADCRUMBS } from '../../../shared/config/breadcrumbs';

const CLUSTER_QUERY = `
  query Cluster($id: ID!) {
    cluster(id: $id) {
      id
      name
      description
      mode
      distribution
      version
      status
      statusMessage
      apiServerUrl
      artifactKey
      lastCheckedAt
      createdAt
      updatedAt
      nodes {
        id
        agentId
        role
        hostname
        ip
        status
        message
      }
    }
  }
`;

const TEST_CONNECTION = `
  mutation TestClusterConnection($id: ID!) {
    testClusterConnection(id: $id)
  }
`;

const NAMESPACES_QUERY = `
  query Namespaces($clusterId: ID!) {
    namespaces(clusterId: $clusterId) {
      name
      status
      createdAt
    }
    namespacesCount
  }
`;

const DEPLOYMENTS_QUERY = `
  query Deployments($clusterId: ID!, $namespace: String) {
    deployments(clusterId: $clusterId, namespace: $namespace) {
      namespace
      name
      replicas
      readyReplicas
      availableReplicas
      images
      createdAt
    }
    deploymentsCount
  }
`;

const PODS_QUERY = `
  query Pods($clusterId: ID!, $namespace: String) {
    pods(clusterId: $clusterId, namespace: $namespace) {
      namespace
      name
      phase
      status
      ready
      restarts
      age
      ip
      node
    }
    podsCount
  }
`;

const SERVICES_QUERY = `
  query K8sServices($clusterId: ID!, $namespace: String) {
    k8sServices(clusterId: $clusterId, namespace: $namespace) {
      namespace
      name
      type
      clusterIP
      externalIP
      ports {
        port
        targetPort
        protocol
      }
      createdAt
    }
    k8sServicesCount
  }
`;

const statusColors: Record<string, 'success' | 'error' | 'warning' | 'default' | 'info'> = {
  CONNECTED: 'success',
  DISCONNECTED: 'error',
  ERROR: 'error',
  PENDING: 'warning',
  DEPLOYING: 'info',
};

const nodeStatusColors: Record<string, 'success' | 'error' | 'warning' | 'default' | 'info'> = {
  READY: 'success',
  ERROR: 'error',
  PENDING: 'warning',
  DEPLOYING: 'info',
};

const modeColors: Record<string, 'primary' | 'secondary'> = {
  CONNECT: 'primary',
  DEPLOY: 'secondary',
};

interface ClusterNode {
  id: string;
  agentId: string;
  role: string;
  hostname: string;
  ip: string;
  status: string;
  message: string;
}

interface Cluster {
  id: string;
  name: string;
  description: string;
  mode: string;
  distribution: string;
  version: string;
  status: string;
  statusMessage: string;
  apiServerUrl: string;
  artifactKey: string;
  lastCheckedAt: string;
  createdAt: string;
  updatedAt: string;
  nodes: ClusterNode[];
}

export const ClusterDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { request } = useGraphQL();
  const { showSuccess, showError } = useSnackbar();
  const { t } = useTranslation();

  const [tab, setTab] = useState(0);
  const [cluster, setCluster] = useState<Cluster | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [mutationLoading, setMutationLoading] = useState(false);

  // Resources data
  const [namespacesData, setNamespacesData] = useState<{ namespaces: unknown[]; namespacesCount: number } | null>(null);
  const [deploymentsData, setDeploymentsData] = useState<{ deployments: unknown[]; deploymentsCount: number } | null>(null);
  const [podsData, setPodsData] = useState<{ pods: unknown[]; podsCount: number } | null>(null);
  const [servicesData, setServicesData] = useState<{ k8sServices: unknown[]; k8sServicesCount: number } | null>(null);
  const [resourcesLoading, setResourcesLoading] = useState(false);

  useBreadcrumb([
    BREADCRUMBS.PILOTE,
    BREADCRUMBS.KUBERNETES,
    { labelKey: 'breadcrumb.clusters', path: '/pilote/kubernetes/clusters' },
    { labelKey: 'common.view' },
  ]);

  // Load cluster
  const loadCluster = useCallback(async () => {
    setLoading(true);
    try {
      const data = await request<{ cluster: Cluster }>(CLUSTER_QUERY, { id });
      setCluster(data.cluster);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('clusters.error_loading'));
    } finally {
      setLoading(false);
    }
  }, [id, request, t]);

  // Load resources based on tab
  const loadResources = useCallback(async () => {
    if (!cluster || cluster.status !== 'CONNECTED') return;

    setResourcesLoading(true);
    try {
      switch (tab) {
        case 1: {
          const data = await request<{ namespaces: unknown[]; namespacesCount: number }>(NAMESPACES_QUERY, { clusterId: id });
          setNamespacesData(data);
          break;
        }
        case 2: {
          const data = await request<{ deployments: unknown[]; deploymentsCount: number }>(DEPLOYMENTS_QUERY, { clusterId: id });
          setDeploymentsData(data);
          break;
        }
        case 3: {
          const data = await request<{ pods: unknown[]; podsCount: number }>(PODS_QUERY, { clusterId: id });
          setPodsData(data);
          break;
        }
        case 4: {
          const data = await request<{ k8sServices: unknown[]; k8sServicesCount: number }>(SERVICES_QUERY, { clusterId: id });
          setServicesData(data);
          break;
        }
      }
    } catch {
      // Silently handle resource loading errors
    } finally {
      setResourcesLoading(false);
    }
  }, [cluster, tab, id, request]);

  useEffect(() => {
    loadCluster();
  }, [loadCluster]);

  useEffect(() => {
    if (tab > 0) {
      loadResources();
    }
  }, [tab, loadResources]);

  const handleTestConnection = async () => {
    setMutationLoading(true);
    try {
      await request(TEST_CONNECTION, { id });
      showSuccess(t('clusters.connection_test_success'));
      await loadCluster();
    } catch (err) {
      showError(t('clusters.connection_test_failed', { error: err instanceof Error ? err.message : String(err) }));
    } finally {
      setMutationLoading(false);
    }
  };

  if (loading) {
    return (
      <CSDDetailPage title={t('common.loading')} loading>
        <CSDStack spacing={2} alignItems="center" justifyContent="center" sx={{ minHeight: 200 }}>
          <CSDCircularProgress />
        </CSDStack>
      </CSDDetailPage>
    );
  }

  if (error || !cluster) {
    return (
      <CSDDetailPage title={t('clusters.not_found')}>
        <CSDAlert severity="error">{error || t('clusters.not_found')}</CSDAlert>
      </CSDDetailPage>
    );
  }

  const isConnected = cluster.status === 'CONNECTED';
  const isDeployMode = cluster.mode === 'DEPLOY';

  const nodeColumns = [
    { id: 'role', label: 'clusters.node_role', render: (row: ClusterNode) => row.role },
    { id: 'hostname', label: 'clusters.hostname', render: (row: ClusterNode) => row.hostname },
    { id: 'ip', label: 'clusters.ip', render: (row: ClusterNode) => row.ip },
    {
      id: 'status',
      label: 'clusters.status',
      render: (row: ClusterNode) => (
        <CSDChip label={row.status} color={nodeStatusColors[row.status] || 'default'} size="small" />
      ),
    },
    { id: 'message', label: 'clusters.message', render: (row: ClusterNode) => row.message || '-' },
  ];

  const namespaceColumns = [
    { id: 'name', label: 'common.name', render: (row: Record<string, unknown>) => row.name as string },
    { id: 'status', label: 'clusters.status', render: (row: Record<string, unknown>) => row.status as string },
  ];

  const deploymentColumns = [
    { id: 'namespace', label: 'clusters.namespace', render: (row: Record<string, unknown>) => row.namespace as string },
    { id: 'name', label: 'common.name', render: (row: Record<string, unknown>) => row.name as string },
    {
      id: 'replicas',
      label: 'clusters.replicas',
      render: (row: Record<string, unknown>) => `${row.readyReplicas}/${row.replicas}`,
    },
    {
      id: 'image',
      label: 'clusters.image',
      render: (row: Record<string, unknown>) => ((row.images as string[]) || [])[0] || '-',
    },
  ];

  const podColumns = [
    { id: 'namespace', label: 'clusters.namespace', render: (row: Record<string, unknown>) => row.namespace as string },
    { id: 'name', label: 'common.name', render: (row: Record<string, unknown>) => row.name as string },
    {
      id: 'status',
      label: 'clusters.status',
      render: (row: Record<string, unknown>) => {
        const status = row.status as string;
        const color = status === 'Running' ? 'success' : status === 'Pending' ? 'warning' : 'error';
        return <CSDChip label={status} color={color} size="small" />;
      },
    },
    { id: 'ready', label: 'clusters.ready', render: (row: Record<string, unknown>) => row.ready as string },
    { id: 'restarts', label: 'clusters.restarts', render: (row: Record<string, unknown>) => row.restarts as number },
    { id: 'age', label: 'clusters.age', render: (row: Record<string, unknown>) => row.age as string },
    { id: 'node', label: 'clusters.node', render: (row: Record<string, unknown>) => row.node as string },
  ];

  const serviceColumns = [
    { id: 'namespace', label: 'clusters.namespace', render: (row: Record<string, unknown>) => row.namespace as string },
    { id: 'name', label: 'common.name', render: (row: Record<string, unknown>) => row.name as string },
    { id: 'type', label: 'clusters.type', render: (row: Record<string, unknown>) => row.type as string },
    { id: 'clusterIP', label: 'clusters.cluster_ip', render: (row: Record<string, unknown>) => row.clusterIP as string },
    {
      id: 'ports',
      label: 'clusters.ports',
      render: (row: Record<string, unknown>) =>
        ((row.ports as Array<{ port: number; protocol: string }>) || []).map((p) => `${p.port}/${p.protocol}`).join(', ') || '-',
    },
  ];

  const renderNodesTab = () => {
    if (!isDeployMode) {
      return (
        <CSDStack spacing={2} alignItems="center" sx={{ py: 4 }}>
          <CSDTypography color="text.secondary">{t('clusters.nodes_deploy_only')}</CSDTypography>
        </CSDStack>
      );
    }

    if (!cluster.nodes || cluster.nodes.length === 0) {
      return (
        <CSDStack spacing={2} alignItems="center" sx={{ py: 4 }}>
          <CSDTypography color="text.secondary">{t('clusters.no_nodes')}</CSDTypography>
        </CSDStack>
      );
    }

    return (
      <CSDDataGrid
        id="cluster-nodes-grid"
        data={cluster.nodes}
        columns={nodeColumns}
        keyField="id"
      />
    );
  };

  const renderResourcesTab = (
    data: unknown[] | undefined,
    columns: { id: string; label: string; render: (row: Record<string, unknown>) => React.ReactNode }[],
    getKey: (row: Record<string, unknown>) => string
  ) => {
    if (!isConnected) {
      return (
        <CSDStack spacing={2} alignItems="center" sx={{ py: 4 }}>
          <CSDTypography color="text.secondary">{t('clusters.connect_to_view')}</CSDTypography>
          <CSDActionButton variant="contained" onClick={handleTestConnection} disabled={mutationLoading}>
            {mutationLoading ? t('common.testing') : t('clusters.test_connection')}
          </CSDActionButton>
        </CSDStack>
      );
    }

    if (resourcesLoading) {
      return (
        <CSDStack spacing={2} alignItems="center" sx={{ py: 4 }}>
          <CSDCircularProgress />
        </CSDStack>
      );
    }

    return (
      <CSDDataGrid
        id={`cluster-resources-grid-${tab}`}
        data={(data || []) as Record<string, unknown>[]}
        columns={columns}
        keyField="name"
        getRowKey={getKey}
      />
    );
  };

  const renderTabContent = () => {
    switch (tab) {
      case 0:
        return renderNodesTab();
      case 1:
        return renderResourcesTab(namespacesData?.namespaces as Record<string, unknown>[], namespaceColumns, (row) => row.name as string);
      case 2:
        return renderResourcesTab(deploymentsData?.deployments as Record<string, unknown>[], deploymentColumns, (row) => `${row.namespace}/${row.name}`);
      case 3:
        return renderResourcesTab(podsData?.pods as Record<string, unknown>[], podColumns, (row) => `${row.namespace}/${row.name}`);
      case 4:
        return renderResourcesTab(servicesData?.k8sServices as Record<string, unknown>[], serviceColumns, (row) => `${row.namespace}/${row.name}`);
      default:
        return null;
    }
  };

  return (
    <CSDDetailPage
      title={cluster.name}
      actions={
        <CSDActionButton
          variant="outlined"
          startIcon={<CSDIcon name="sync" />}
          onClick={handleTestConnection}
          disabled={mutationLoading}
        >
          {mutationLoading ? t('common.testing') : t('clusters.test_connection')}
        </CSDActionButton>
      }
    >
      <CSDStack spacing={3}>
        <CSDPaper>
          <CSDInfoGrid>
            <CSDInfoItem label={t('clusters.name')} value={cluster.name} />
            <CSDInfoItem
              label={t('clusters.mode')}
              value={<CSDChip label={cluster.mode} color={modeColors[cluster.mode] || 'default'} size="small" variant="outlined" />}
            />
            <CSDInfoItem
              label={t('clusters.status')}
              value={<CSDChip label={cluster.status} color={statusColors[cluster.status] || 'default'} size="small" />}
            />
            {cluster.distribution && <CSDInfoItem label={t('clusters.distribution')} value={cluster.distribution} />}
            {cluster.version && <CSDInfoItem label={t('clusters.version')} value={cluster.version} />}
            {cluster.apiServerUrl && <CSDInfoItem label={t('clusters.api_server')} value={cluster.apiServerUrl} />}
            {cluster.artifactKey && <CSDInfoItem label={t('clusters.kubeconfig_artifact')} value={cluster.artifactKey} />}
            {cluster.statusMessage && (
              <CSDInfoItem
                label={t('clusters.status_message')}
                value={<CSDTypography color={cluster.status === 'ERROR' ? 'error' : 'textPrimary'}>{cluster.statusMessage}</CSDTypography>}
                fullWidth
              />
            )}
            {cluster.description && <CSDInfoItem label={t('clusters.description')} value={cluster.description} fullWidth />}
            <CSDInfoItem label={t('common.created')} value={formatDate(cluster.createdAt, '-')} />
            <CSDInfoItem label={t('common.updated')} value={formatDate(cluster.updatedAt, '-')} />
          </CSDInfoGrid>
        </CSDPaper>

        <CSDPaper>
          <CSDTabs value={tab} onChange={(_: unknown, v: number) => setTab(v)}>
            <CSDTab label={`${t('clusters.nodes')}${isDeployMode && cluster.nodes?.length ? ` (${cluster.nodes.length})` : ''}`} disabled={!isDeployMode} />
            <CSDTab label={`${t('clusters.namespaces')}${namespacesData?.namespacesCount ? ` (${namespacesData.namespacesCount})` : ''}`} />
            <CSDTab label={`${t('clusters.deployments')}${deploymentsData?.deploymentsCount ? ` (${deploymentsData.deploymentsCount})` : ''}`} />
            <CSDTab label={`${t('clusters.pods')}${podsData?.podsCount ? ` (${podsData.podsCount})` : ''}`} />
            <CSDTab label={`${t('clusters.services')}${servicesData?.k8sServicesCount ? ` (${servicesData.k8sServicesCount})` : ''}`} />
          </CSDTabs>
          <CSDStack spacing={2} sx={{ p: 2 }}>
            {renderTabContent()}
          </CSDStack>
        </CSDPaper>
      </CSDStack>
    </CSDDetailPage>
  );
};

export default ClusterDetailPage;
