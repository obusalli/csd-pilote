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
} from 'csd_core/UI';
import { useBreadcrumb, useParams, useGraphQL, useSnackbar } from 'csd_core/Providers';

const HYPERVISOR_QUERY = `
  query Hypervisor($id: ID!) {
    hypervisor(id: $id) {
      id
      name
      description
      mode
      driver
      uri
      status
      statusMessage
      version
      hostname
      artifactKey
      createdAt
      updatedAt
    }
  }
`;

const TEST_CONNECTION = `
  mutation TestHypervisorConnection($id: ID!) {
    testHypervisorConnection(id: $id)
  }
`;

const DOMAINS_QUERY = `
  query Domains($hypervisorId: ID!) {
    domains(hypervisorId: $hypervisorId) {
      uuid
      name
      state
      maxMemory
      memory
      vcpus
      autostart
    }
    domainsCount
  }
`;

const NETWORKS_QUERY = `
  query LibvirtNetworks($hypervisorId: ID!) {
    libvirtNetworks(hypervisorId: $hypervisorId) {
      uuid
      name
      bridge
      active
      autostart
    }
    libvirtNetworksCount
  }
`;

const STORAGE_POOLS_QUERY = `
  query StoragePools($hypervisorId: ID!) {
    storagePools(hypervisorId: $hypervisorId) {
      uuid
      name
      state
      capacity
      allocation
      available
      active
      volumesCount
    }
    storagePoolsCount
  }
`;

const START_DOMAIN = `
  mutation StartDomain($hypervisorId: ID!, $uuid: String!) {
    startDomain(hypervisorId: $hypervisorId, uuid: $uuid) { uuid state }
  }
`;

const SHUTDOWN_DOMAIN = `
  mutation ShutdownDomain($hypervisorId: ID!, $uuid: String!) {
    shutdownDomain(hypervisorId: $hypervisorId, uuid: $uuid) { uuid state }
  }
`;

const statusColors: Record<string, 'success' | 'error' | 'warning' | 'default' | 'info'> = {
  CONNECTED: 'success',
  DISCONNECTED: 'error',
  ERROR: 'error',
  PENDING: 'warning',
  DEPLOYING: 'info',
  RUNNING: 'success',
  SHUTOFF: 'default',
  PAUSED: 'warning',
};

const modeColors: Record<string, 'primary' | 'secondary'> = {
  CONNECT: 'primary',
  DEPLOY: 'secondary',
};

interface Hypervisor {
  id: string;
  name: string;
  description: string;
  mode: string;
  driver: string;
  uri: string;
  status: string;
  statusMessage: string;
  version: string;
  hostname: string;
  artifactKey: string;
  createdAt: string;
  updatedAt: string;
}

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

export const HypervisorDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [tab, setTab] = React.useState(0);
  const { showSuccess, showError } = useSnackbar();
  const { execute, loading: mutationLoading } = useGraphQL();

  useBreadcrumb([
    { label: 'Pilote', path: '/pilote' },
    { label: 'Libvirt', path: '/pilote/libvirt' },
    { label: 'Hypervisors', path: '/pilote/libvirt/hypervisors' },
    { label: 'Details' },
  ]);

  const { data: hvData, loading: hvLoading, error: hvError, refetch: refetchHv } = useGraphQL<{ hypervisor: Hypervisor }>(HYPERVISOR_QUERY, { id });

  const { data: domainsData, loading: domainsLoading, refetch: refetchDomains } = useGraphQL(DOMAINS_QUERY, { hypervisorId: id }, { skip: tab !== 0 || hvData?.hypervisor?.status !== 'CONNECTED' });
  const { data: networksData, loading: networksLoading } = useGraphQL(NETWORKS_QUERY, { hypervisorId: id }, { skip: tab !== 1 || hvData?.hypervisor?.status !== 'CONNECTED' });
  const { data: storageData, loading: storageLoading } = useGraphQL(STORAGE_POOLS_QUERY, { hypervisorId: id }, { skip: tab !== 2 || hvData?.hypervisor?.status !== 'CONNECTED' });

  const handleTestConnection = async () => {
    try {
      await execute(TEST_CONNECTION, { id });
      showSuccess('Connection test successful');
      refetchHv();
    } catch (error) {
      showError(`Connection test failed: ${error}`);
    }
  };

  const handleStartDomain = async (uuid: string) => {
    try {
      await execute(START_DOMAIN, { hypervisorId: id, uuid });
      showSuccess('VM started');
      refetchDomains();
    } catch (error) {
      showError(`Failed to start VM: ${error}`);
    }
  };

  const handleShutdownDomain = async (uuid: string) => {
    try {
      await execute(SHUTDOWN_DOMAIN, { hypervisorId: id, uuid });
      showSuccess('VM shutdown initiated');
      refetchDomains();
    } catch (error) {
      showError(`Failed to shutdown VM: ${error}`);
    }
  };

  if (hvLoading) {
    return (
      <CSDLayoutPage title="Loading...">
        <CSDBox sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CSDCircularProgress />
        </CSDBox>
      </CSDLayoutPage>
    );
  }

  if (hvError || !hvData?.hypervisor) {
    return (
      <CSDLayoutPage title="Hypervisor Not Found">
        <CSDAlert severity="error">
          {hvError?.message || 'Hypervisor not found'}
        </CSDAlert>
      </CSDLayoutPage>
    );
  }

  const hypervisor = hvData.hypervisor;
  const isConnected = hypervisor.status === 'CONNECTED';

  const domainColumns = [
    { field: 'name', headerName: 'Name', flex: 1 },
    {
      field: 'state',
      headerName: 'State',
      width: 120,
      renderCell: (params: { value: string }) => (
        <CSDChip label={params.value} color={statusColors[params.value] || 'default'} size="small" />
      ),
    },
    { field: 'vcpus', headerName: 'vCPUs', width: 80 },
    {
      field: 'memory',
      headerName: 'Memory',
      width: 120,
      renderCell: (params: { value: number }) => formatBytes(params.value * 1024),
    },
    {
      field: 'actions',
      headerName: 'Actions',
      width: 150,
      renderCell: (params: { row: { uuid: string; state: string } }) => (
        <CSDBox sx={{ display: 'flex', gap: 1 }}>
          {params.row.state === 'SHUTOFF' ? (
            <CSDButton size="small" color="success" onClick={() => handleStartDomain(params.row.uuid)}>
              Start
            </CSDButton>
          ) : params.row.state === 'RUNNING' ? (
            <CSDButton size="small" color="warning" onClick={() => handleShutdownDomain(params.row.uuid)}>
              Shutdown
            </CSDButton>
          ) : null}
        </CSDBox>
      ),
    },
  ];

  const networkColumns = [
    { field: 'name', headerName: 'Name', flex: 1 },
    { field: 'bridge', headerName: 'Bridge', width: 150 },
    {
      field: 'active',
      headerName: 'Active',
      width: 100,
      renderCell: (params: { value: boolean }) => (
        <CSDChip label={params.value ? 'Yes' : 'No'} color={params.value ? 'success' : 'default'} size="small" />
      ),
    },
    {
      field: 'autostart',
      headerName: 'Autostart',
      width: 100,
      renderCell: (params: { value: boolean }) => (
        <CSDChip label={params.value ? 'Yes' : 'No'} color={params.value ? 'success' : 'default'} size="small" />
      ),
    },
  ];

  const storageColumns = [
    { field: 'name', headerName: 'Name', flex: 1 },
    { field: 'state', headerName: 'State', width: 100 },
    {
      field: 'capacity',
      headerName: 'Capacity',
      width: 120,
      renderCell: (params: { value: number }) => formatBytes(params.value),
    },
    {
      field: 'allocation',
      headerName: 'Used',
      width: 120,
      renderCell: (params: { value: number }) => formatBytes(params.value),
    },
    {
      field: 'available',
      headerName: 'Available',
      width: 120,
      renderCell: (params: { value: number }) => formatBytes(params.value),
    },
    { field: 'volumesCount', headerName: 'Volumes', width: 80 },
  ];

  const renderTabContent = () => {
    if (!isConnected) {
      return (
        <CSDBox sx={{ p: 3, textAlign: 'center' }}>
          <CSDTypography color="text.secondary" sx={{ mb: 2 }}>
            Connect to the hypervisor to view resources
          </CSDTypography>
          <CSDButton variant="contained" color="primary" onClick={handleTestConnection} disabled={mutationLoading}>
            {mutationLoading ? 'Testing...' : 'Test Connection'}
          </CSDButton>
        </CSDBox>
      );
    }

    switch (tab) {
      case 0:
        return domainsLoading ? (
          <CSDBox sx={{ display: 'flex', justifyContent: 'center', p: 4 }}><CSDCircularProgress /></CSDBox>
        ) : (
          <CSDDataGrid
            rows={domainsData?.domains || []}
            columns={domainColumns}
            getRowId={(row: { uuid: string }) => row.uuid}
            autoHeight
            disableRowSelectionOnClick
          />
        );
      case 1:
        return networksLoading ? (
          <CSDBox sx={{ display: 'flex', justifyContent: 'center', p: 4 }}><CSDCircularProgress /></CSDBox>
        ) : (
          <CSDDataGrid
            rows={networksData?.libvirtNetworks || []}
            columns={networkColumns}
            getRowId={(row: { uuid: string }) => row.uuid}
            autoHeight
            disableRowSelectionOnClick
          />
        );
      case 2:
        return storageLoading ? (
          <CSDBox sx={{ display: 'flex', justifyContent: 'center', p: 4 }}><CSDCircularProgress /></CSDBox>
        ) : (
          <CSDDataGrid
            rows={storageData?.storagePools || []}
            columns={storageColumns}
            getRowId={(row: { uuid: string }) => row.uuid}
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
      title={hypervisor.name}
      actions={
        <CSDButton variant="outlined" color="primary" onClick={handleTestConnection} disabled={mutationLoading}>
          {mutationLoading ? 'Testing...' : 'Test Connection'}
        </CSDButton>
      }
    >
      <CSDPaper sx={{ mb: 3 }}>
        <CSDBox sx={{ p: 3 }}>
          <CSDGrid container spacing={2}>
            <CSDGrid item xs={12} md={6}>
              <CSDTypography variant="subtitle2" color="text.secondary">Name</CSDTypography>
              <CSDTypography>{hypervisor.name}</CSDTypography>
            </CSDGrid>
            <CSDGrid item xs={12} md={3}>
              <CSDTypography variant="subtitle2" color="text.secondary">Mode</CSDTypography>
              <CSDChip label={hypervisor.mode} color={modeColors[hypervisor.mode] || 'default'} size="small" variant="outlined" />
            </CSDGrid>
            <CSDGrid item xs={12} md={3}>
              <CSDTypography variant="subtitle2" color="text.secondary">Status</CSDTypography>
              <CSDChip label={hypervisor.status} color={statusColors[hypervisor.status] || 'default'} size="small" />
            </CSDGrid>
            {hypervisor.driver && (
              <CSDGrid item xs={12} md={6}>
                <CSDTypography variant="subtitle2" color="text.secondary">Driver</CSDTypography>
                <CSDTypography>{hypervisor.driver}</CSDTypography>
              </CSDGrid>
            )}
            <CSDGrid item xs={12} md={6}>
              <CSDTypography variant="subtitle2" color="text.secondary">URI</CSDTypography>
              <CSDTypography>{hypervisor.uri}</CSDTypography>
            </CSDGrid>
            <CSDGrid item xs={12} md={6}>
              <CSDTypography variant="subtitle2" color="text.secondary">Version</CSDTypography>
              <CSDTypography>{hypervisor.version || '-'}</CSDTypography>
            </CSDGrid>
            <CSDGrid item xs={12} md={6}>
              <CSDTypography variant="subtitle2" color="text.secondary">Hostname</CSDTypography>
              <CSDTypography>{hypervisor.hostname || '-'}</CSDTypography>
            </CSDGrid>
            {hypervisor.statusMessage && (
              <CSDGrid item xs={12}>
                <CSDTypography variant="subtitle2" color="text.secondary">Status Message</CSDTypography>
                <CSDTypography color={hypervisor.status === 'ERROR' ? 'error' : 'textPrimary'}>
                  {hypervisor.statusMessage}
                </CSDTypography>
              </CSDGrid>
            )}
            {hypervisor.description && (
              <CSDGrid item xs={12}>
                <CSDTypography variant="subtitle2" color="text.secondary">Description</CSDTypography>
                <CSDTypography>{hypervisor.description}</CSDTypography>
              </CSDGrid>
            )}
          </CSDGrid>
        </CSDBox>
      </CSDPaper>

      <CSDPaper>
        <CSDTabs value={tab} onChange={(_: unknown, v: number) => setTab(v)}>
          <CSDTab label={`Virtual Machines${domainsData?.domainsCount ? ` (${domainsData.domainsCount})` : ''}`} />
          <CSDTab label={`Networks${networksData?.libvirtNetworksCount ? ` (${networksData.libvirtNetworksCount})` : ''}`} />
          <CSDTab label={`Storage Pools${storageData?.storagePoolsCount ? ` (${storageData.storagePoolsCount})` : ''}`} />
        </CSDTabs>
        <CSDBox sx={{ p: 2 }}>{renderTabContent()}</CSDBox>
      </CSDPaper>
    </CSDLayoutPage>
  );
};

export default HypervisorDetailPage;
