import React from 'react';
import {
  CSDLayoutPage,
  CSDBox,
  CSDTypography,
  CSDPaper,
  CSDGrid,
  CSDChip,
  CSDTabs,
  CSDTab,
  CSDButton,
  CSDDataGrid,
  CSDCircularProgress,
  CSDAlert,
  CSDDivider,
} from 'csd_core/UI';
import { useBreadcrumb, useParams, useGraphQL, useSnackbar } from 'csd_core/Providers';

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
  const [tab, setTab] = React.useState(0);
  const { showSuccess, showError } = useSnackbar();
  const { execute, loading: mutationLoading } = useGraphQL();

  useBreadcrumb([
    { label: 'Pilote', path: '/pilote' },
    { label: 'Kubernetes', path: '/pilote/kubernetes' },
    { label: 'Clusters', path: '/pilote/kubernetes/clusters' },
    { label: 'Details' },
  ]);

  // Fetch cluster details
  const { data: clusterData, loading: clusterLoading, error: clusterError, refetch: refetchCluster } = useGraphQL<{ cluster: Cluster }>(CLUSTER_QUERY, { id });

  // Fetch resources based on tab
  const isConnected = clusterData?.cluster?.status === 'CONNECTED';
  const { data: namespacesData, loading: namespacesLoading } = useGraphQL(NAMESPACES_QUERY, { clusterId: id }, { skip: tab !== 1 || !isConnected });
  const { data: deploymentsData, loading: deploymentsLoading } = useGraphQL(DEPLOYMENTS_QUERY, { clusterId: id }, { skip: tab !== 2 || !isConnected });
  const { data: podsData, loading: podsLoading } = useGraphQL(PODS_QUERY, { clusterId: id }, { skip: tab !== 3 || !isConnected });
  const { data: servicesData, loading: servicesLoading } = useGraphQL(SERVICES_QUERY, { clusterId: id }, { skip: tab !== 4 || !isConnected });

  const handleTestConnection = async () => {
    try {
      await execute(TEST_CONNECTION, { id });
      showSuccess('Connection test successful');
      refetchCluster();
    } catch (error) {
      showError(`Connection test failed: ${error}`);
    }
  };

  if (clusterLoading) {
    return (
      <CSDLayoutPage title="Loading...">
        <CSDBox sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CSDCircularProgress />
        </CSDBox>
      </CSDLayoutPage>
    );
  }

  if (clusterError || !clusterData?.cluster) {
    return (
      <CSDLayoutPage title="Cluster Not Found">
        <CSDAlert severity="error">
          {clusterError?.message || 'Cluster not found'}
        </CSDAlert>
      </CSDLayoutPage>
    );
  }

  const cluster = clusterData.cluster;
  const isDeployMode = cluster.mode === 'DEPLOY';

  const nodeColumns = [
    { field: 'role', headerName: 'Role', width: 100 },
    { field: 'hostname', headerName: 'Hostname', flex: 1 },
    { field: 'ip', headerName: 'IP', width: 140 },
    {
      field: 'status',
      headerName: 'Status',
      width: 120,
      renderCell: (params: { value: string }) => (
        <CSDChip
          label={params.value}
          color={nodeStatusColors[params.value] || 'default'}
          size="small"
        />
      ),
    },
    { field: 'message', headerName: 'Message', flex: 1 },
  ];

  const namespaceColumns = [
    { field: 'name', headerName: 'Name', flex: 1 },
    { field: 'status', headerName: 'Status', width: 120 },
  ];

  const deploymentColumns = [
    { field: 'namespace', headerName: 'Namespace', width: 150 },
    { field: 'name', headerName: 'Name', flex: 1 },
    {
      field: 'replicas',
      headerName: 'Replicas',
      width: 120,
      renderCell: (params: { row: { readyReplicas: number; replicas: number } }) => (
        <CSDTypography>
          {params.row.readyReplicas}/{params.row.replicas}
        </CSDTypography>
      ),
    },
    {
      field: 'images',
      headerName: 'Image',
      flex: 1,
      renderCell: (params: { value: string[] }) => params.value?.[0] || '-',
    },
  ];

  const podColumns = [
    { field: 'namespace', headerName: 'Namespace', width: 150 },
    { field: 'name', headerName: 'Name', flex: 1 },
    {
      field: 'status',
      headerName: 'Status',
      width: 120,
      renderCell: (params: { value: string }) => {
        const color = params.value === 'Running' ? 'success' : params.value === 'Pending' ? 'warning' : 'error';
        return <CSDChip label={params.value} color={color} size="small" />;
      },
    },
    { field: 'ready', headerName: 'Ready', width: 80 },
    { field: 'restarts', headerName: 'Restarts', width: 80 },
    { field: 'age', headerName: 'Age', width: 80 },
    { field: 'node', headerName: 'Node', width: 150 },
  ];

  const serviceColumns = [
    { field: 'namespace', headerName: 'Namespace', width: 150 },
    { field: 'name', headerName: 'Name', flex: 1 },
    { field: 'type', headerName: 'Type', width: 120 },
    { field: 'clusterIP', headerName: 'Cluster IP', width: 150 },
    {
      field: 'ports',
      headerName: 'Ports',
      flex: 1,
      renderCell: (params: { value: Array<{ port: number; targetPort: string; protocol: string }> }) =>
        params.value?.map((p) => `${p.port}/${p.protocol}`).join(', ') || '-',
    },
  ];

  const renderNodesTab = () => {
    if (!isDeployMode) {
      return (
        <CSDBox sx={{ p: 3, textAlign: 'center' }}>
          <CSDTypography color="text.secondary">
            Node information is only available for deployed clusters.
          </CSDTypography>
        </CSDBox>
      );
    }

    if (!cluster.nodes || cluster.nodes.length === 0) {
      return (
        <CSDBox sx={{ p: 3, textAlign: 'center' }}>
          <CSDTypography color="text.secondary">No nodes found.</CSDTypography>
        </CSDBox>
      );
    }

    return (
      <CSDDataGrid
        rows={cluster.nodes}
        columns={nodeColumns}
        autoHeight
        disableRowSelectionOnClick
      />
    );
  };

  const renderResourcesTab = (
    loading: boolean,
    data: unknown[] | undefined,
    columns: { field: string; headerName: string; flex?: number; width?: number; renderCell?: (params: unknown) => React.ReactNode }[],
    getRowId: (row: Record<string, unknown>) => string
  ) => {
    if (!isConnected) {
      return (
        <CSDBox sx={{ p: 3, textAlign: 'center' }}>
          <CSDTypography color="text.secondary" sx={{ mb: 2 }}>
            Connect to the cluster to view resources
          </CSDTypography>
          <CSDButton
            variant="contained"
            color="primary"
            onClick={handleTestConnection}
            disabled={mutationLoading}
          >
            {mutationLoading ? 'Testing...' : 'Test Connection'}
          </CSDButton>
        </CSDBox>
      );
    }

    if (loading) {
      return (
        <CSDBox sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CSDCircularProgress />
        </CSDBox>
      );
    }

    return (
      <CSDDataGrid
        rows={data || []}
        columns={columns}
        getRowId={getRowId}
        autoHeight
        disableRowSelectionOnClick
      />
    );
  };

  const renderTabContent = () => {
    switch (tab) {
      case 0:
        return renderNodesTab();
      case 1:
        return renderResourcesTab(
          namespacesLoading,
          namespacesData?.namespaces,
          namespaceColumns,
          (row) => row.name as string
        );
      case 2:
        return renderResourcesTab(
          deploymentsLoading,
          deploymentsData?.deployments,
          deploymentColumns,
          (row) => `${row.namespace}/${row.name}`
        );
      case 3:
        return renderResourcesTab(
          podsLoading,
          podsData?.pods,
          podColumns,
          (row) => `${row.namespace}/${row.name}`
        );
      case 4:
        return renderResourcesTab(
          servicesLoading,
          servicesData?.k8sServices,
          serviceColumns,
          (row) => `${row.namespace}/${row.name}`
        );
      default:
        return null;
    }
  };

  return (
    <CSDLayoutPage
      title={cluster.name}
      actions={
        <CSDBox sx={{ display: 'flex', gap: 1 }}>
          <CSDButton
            variant="outlined"
            color="primary"
            onClick={handleTestConnection}
            disabled={mutationLoading}
          >
            {mutationLoading ? 'Testing...' : 'Test Connection'}
          </CSDButton>
        </CSDBox>
      }
    >
      <CSDPaper sx={{ mb: 3 }}>
        <CSDBox sx={{ p: 3 }}>
          <CSDGrid container spacing={2}>
            <CSDGrid item xs={12} md={6}>
              <CSDTypography variant="subtitle2" color="text.secondary">
                Name
              </CSDTypography>
              <CSDTypography>{cluster.name}</CSDTypography>
            </CSDGrid>
            <CSDGrid item xs={12} md={3}>
              <CSDTypography variant="subtitle2" color="text.secondary">
                Mode
              </CSDTypography>
              <CSDChip
                label={cluster.mode}
                color={modeColors[cluster.mode] || 'default'}
                size="small"
                variant="outlined"
              />
            </CSDGrid>
            <CSDGrid item xs={12} md={3}>
              <CSDTypography variant="subtitle2" color="text.secondary">
                Status
              </CSDTypography>
              <CSDChip
                label={cluster.status}
                color={statusColors[cluster.status] || 'default'}
                size="small"
              />
            </CSDGrid>

            {cluster.distribution && (
              <CSDGrid item xs={12} md={6}>
                <CSDTypography variant="subtitle2" color="text.secondary">
                  Distribution
                </CSDTypography>
                <CSDTypography>{cluster.distribution}</CSDTypography>
              </CSDGrid>
            )}

            {cluster.version && (
              <CSDGrid item xs={12} md={6}>
                <CSDTypography variant="subtitle2" color="text.secondary">
                  Version
                </CSDTypography>
                <CSDTypography>{cluster.version}</CSDTypography>
              </CSDGrid>
            )}

            {cluster.apiServerUrl && (
              <CSDGrid item xs={12} md={6}>
                <CSDTypography variant="subtitle2" color="text.secondary">
                  API Server
                </CSDTypography>
                <CSDTypography>{cluster.apiServerUrl}</CSDTypography>
              </CSDGrid>
            )}

            {cluster.artifactKey && (
              <CSDGrid item xs={12} md={6}>
                <CSDTypography variant="subtitle2" color="text.secondary">
                  Kubeconfig Artifact
                </CSDTypography>
                <CSDTypography>{cluster.artifactKey}</CSDTypography>
              </CSDGrid>
            )}

            {cluster.statusMessage && (
              <CSDGrid item xs={12}>
                <CSDTypography variant="subtitle2" color="text.secondary">
                  Status Message
                </CSDTypography>
                <CSDTypography color={cluster.status === 'ERROR' ? 'error' : 'textPrimary'}>
                  {cluster.statusMessage}
                </CSDTypography>
              </CSDGrid>
            )}

            {cluster.description && (
              <CSDGrid item xs={12}>
                <CSDTypography variant="subtitle2" color="text.secondary">
                  Description
                </CSDTypography>
                <CSDTypography>{cluster.description}</CSDTypography>
              </CSDGrid>
            )}
          </CSDGrid>
        </CSDBox>
      </CSDPaper>

      <CSDPaper>
        <CSDTabs value={tab} onChange={(_: unknown, v: number) => setTab(v)}>
          <CSDTab
            label={`Nodes${isDeployMode && cluster.nodes?.length ? ` (${cluster.nodes.length})` : ''}`}
            disabled={!isDeployMode}
          />
          <CSDTab label={`Namespaces${namespacesData?.namespacesCount ? ` (${namespacesData.namespacesCount})` : ''}`} />
          <CSDTab label={`Deployments${deploymentsData?.deploymentsCount ? ` (${deploymentsData.deploymentsCount})` : ''}`} />
          <CSDTab label={`Pods${podsData?.podsCount ? ` (${podsData.podsCount})` : ''}`} />
          <CSDTab label={`Services${servicesData?.k8sServicesCount ? ` (${servicesData.k8sServicesCount})` : ''}`} />
        </CSDTabs>

        <CSDBox sx={{ p: 2 }}>
          {renderTabContent()}
        </CSDBox>
      </CSDPaper>
    </CSDLayoutPage>
  );
};

export default ClusterDetailPage;
