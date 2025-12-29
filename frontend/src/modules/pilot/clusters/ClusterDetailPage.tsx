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
  CSDIconButton,
  CSDTooltip,
} from 'csd_core/UI';
import { useBreadcrumb, useParams, useGraphQL, useSnackbar } from 'csd_core/Providers';

const CLUSTER_QUERY = `
  query Cluster($id: ID!) {
    cluster(id: $id) {
      id
      name
      description
      status
      apiServerUrl
      artifactKey
      lastCheckedAt
      createdAt
      updatedAt
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

const statusColors: Record<string, 'success' | 'error' | 'warning' | 'default'> = {
  CONNECTED: 'success',
  DISCONNECTED: 'error',
  ERROR: 'error',
  PENDING: 'warning',
};

interface Cluster {
  id: string;
  name: string;
  description: string;
  status: string;
  apiServerUrl: string;
  artifactKey: string;
  lastCheckedAt: string;
  createdAt: string;
  updatedAt: string;
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
  const { data: namespacesData, loading: namespacesLoading } = useGraphQL(NAMESPACES_QUERY, { clusterId: id }, { skip: tab !== 0 || clusterData?.cluster?.status !== 'CONNECTED' });
  const { data: deploymentsData, loading: deploymentsLoading } = useGraphQL(DEPLOYMENTS_QUERY, { clusterId: id }, { skip: tab !== 1 || clusterData?.cluster?.status !== 'CONNECTED' });
  const { data: podsData, loading: podsLoading } = useGraphQL(PODS_QUERY, { clusterId: id }, { skip: tab !== 2 || clusterData?.cluster?.status !== 'CONNECTED' });
  const { data: servicesData, loading: servicesLoading } = useGraphQL(SERVICES_QUERY, { clusterId: id }, { skip: tab !== 3 || clusterData?.cluster?.status !== 'CONNECTED' });

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
  const isConnected = cluster.status === 'CONNECTED';

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

  const renderTabContent = () => {
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

    switch (tab) {
      case 0:
        return namespacesLoading ? (
          <CSDBox sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
            <CSDCircularProgress />
          </CSDBox>
        ) : (
          <CSDDataGrid
            rows={namespacesData?.namespaces || []}
            columns={namespaceColumns}
            getRowId={(row: { name: string }) => row.name}
            autoHeight
            disableRowSelectionOnClick
          />
        );
      case 1:
        return deploymentsLoading ? (
          <CSDBox sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
            <CSDCircularProgress />
          </CSDBox>
        ) : (
          <CSDDataGrid
            rows={deploymentsData?.deployments || []}
            columns={deploymentColumns}
            getRowId={(row: { namespace: string; name: string }) => `${row.namespace}/${row.name}`}
            autoHeight
            disableRowSelectionOnClick
          />
        );
      case 2:
        return podsLoading ? (
          <CSDBox sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
            <CSDCircularProgress />
          </CSDBox>
        ) : (
          <CSDDataGrid
            rows={podsData?.pods || []}
            columns={podColumns}
            getRowId={(row: { namespace: string; name: string }) => `${row.namespace}/${row.name}`}
            autoHeight
            disableRowSelectionOnClick
          />
        );
      case 3:
        return servicesLoading ? (
          <CSDBox sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
            <CSDCircularProgress />
          </CSDBox>
        ) : (
          <CSDDataGrid
            rows={servicesData?.k8sServices || []}
            columns={serviceColumns}
            getRowId={(row: { namespace: string; name: string }) => `${row.namespace}/${row.name}`}
            autoHeight
            disableRowSelectionOnClick
          />
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
            <CSDGrid item xs={12} md={6}>
              <CSDTypography variant="subtitle2" color="text.secondary">
                Status
              </CSDTypography>
              <CSDChip
                label={cluster.status}
                color={statusColors[cluster.status] || 'default'}
                size="small"
              />
            </CSDGrid>
            <CSDGrid item xs={12} md={6}>
              <CSDTypography variant="subtitle2" color="text.secondary">
                API Server
              </CSDTypography>
              <CSDTypography>{cluster.apiServerUrl || '-'}</CSDTypography>
            </CSDGrid>
            <CSDGrid item xs={12} md={6}>
              <CSDTypography variant="subtitle2" color="text.secondary">
                Artifact Key
              </CSDTypography>
              <CSDTypography>{cluster.artifactKey}</CSDTypography>
            </CSDGrid>
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
