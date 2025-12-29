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

interface Domain {
  uuid: string;
  name: string;
  state: string;
  maxMemory: number;
  memory: number;
  vcpus: number;
  autostart: boolean;
}

interface Network {
  uuid: string;
  name: string;
  bridge: string;
  active: boolean;
  autostart: boolean;
}

interface StoragePool {
  uuid: string;
  name: string;
  state: string;
  capacity: number;
  allocation: number;
  available: number;
  active: boolean;
  volumesCount: number;
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
  const { request } = useGraphQL();
  const { showSuccess, showError } = useSnackbar();
  const { t } = useTranslation();

  const [tab, setTab] = useState(0);
  const [hypervisor, setHypervisor] = useState<Hypervisor | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [mutationLoading, setMutationLoading] = useState(false);

  // Resources data
  const [domainsData, setDomainsData] = useState<{ domains: Domain[]; domainsCount: number } | null>(null);
  const [networksData, setNetworksData] = useState<{ libvirtNetworks: Network[]; libvirtNetworksCount: number } | null>(null);
  const [storageData, setStorageData] = useState<{ storagePools: StoragePool[]; storagePoolsCount: number } | null>(null);
  const [resourcesLoading, setResourcesLoading] = useState(false);

  useBreadcrumb([
    BREADCRUMBS.PILOTE,
    BREADCRUMBS.LIBVIRT,
    { labelKey: 'breadcrumb.hypervisors', path: '/pilote/libvirt/hypervisors' },
    { labelKey: 'common.view' },
  ]);

  // Load hypervisor
  const loadHypervisor = useCallback(async () => {
    setLoading(true);
    try {
      const data = await request<{ hypervisor: Hypervisor }>(HYPERVISOR_QUERY, { id });
      setHypervisor(data.hypervisor);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('hypervisors.error_loading'));
    } finally {
      setLoading(false);
    }
  }, [id, request, t]);

  // Load resources based on tab
  const loadResources = useCallback(async () => {
    if (!hypervisor || hypervisor.status !== 'CONNECTED') return;

    setResourcesLoading(true);
    try {
      switch (tab) {
        case 0: {
          const data = await request<{ domains: Domain[]; domainsCount: number }>(DOMAINS_QUERY, { hypervisorId: id });
          setDomainsData(data);
          break;
        }
        case 1: {
          const data = await request<{ libvirtNetworks: Network[]; libvirtNetworksCount: number }>(NETWORKS_QUERY, { hypervisorId: id });
          setNetworksData(data);
          break;
        }
        case 2: {
          const data = await request<{ storagePools: StoragePool[]; storagePoolsCount: number }>(STORAGE_POOLS_QUERY, { hypervisorId: id });
          setStorageData(data);
          break;
        }
      }
    } catch {
      // Silently handle resource loading errors
    } finally {
      setResourcesLoading(false);
    }
  }, [hypervisor, tab, id, request]);

  useEffect(() => {
    loadHypervisor();
  }, [loadHypervisor]);

  useEffect(() => {
    loadResources();
  }, [tab, loadResources]);

  const handleTestConnection = async () => {
    setMutationLoading(true);
    try {
      await request(TEST_CONNECTION, { id });
      showSuccess(t('hypervisors.connection_test_success'));
      await loadHypervisor();
    } catch (err) {
      showError(t('hypervisors.connection_test_failed', { error: err instanceof Error ? err.message : String(err) }));
    } finally {
      setMutationLoading(false);
    }
  };

  const handleStartDomain = async (uuid: string) => {
    setMutationLoading(true);
    try {
      await request(START_DOMAIN, { hypervisorId: id, uuid });
      showSuccess(t('hypervisors.vm_started'));
      await loadResources();
    } catch (err) {
      showError(t('hypervisors.vm_start_failed', { error: err instanceof Error ? err.message : String(err) }));
    } finally {
      setMutationLoading(false);
    }
  };

  const handleShutdownDomain = async (uuid: string) => {
    setMutationLoading(true);
    try {
      await request(SHUTDOWN_DOMAIN, { hypervisorId: id, uuid });
      showSuccess(t('hypervisors.vm_shutdown_initiated'));
      await loadResources();
    } catch (err) {
      showError(t('hypervisors.vm_shutdown_failed', { error: err instanceof Error ? err.message : String(err) }));
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

  if (error || !hypervisor) {
    return (
      <CSDDetailPage title={t('hypervisors.not_found')}>
        <CSDAlert severity="error">{error || t('hypervisors.not_found')}</CSDAlert>
      </CSDDetailPage>
    );
  }

  const isConnected = hypervisor.status === 'CONNECTED';

  const domainColumns = [
    { id: 'name', label: 'common.name', render: (row: Domain) => row.name },
    {
      id: 'state',
      label: 'hypervisors.state',
      render: (row: Domain) => (
        <CSDChip label={row.state} color={statusColors[row.state] || 'default'} size="small" />
      ),
    },
    { id: 'vcpus', label: 'hypervisors.vcpus', render: (row: Domain) => row.vcpus },
    {
      id: 'memory',
      label: 'hypervisors.memory',
      render: (row: Domain) => formatBytes(row.memory * 1024),
    },
    {
      id: 'actions',
      label: 'common.actions',
      render: (row: Domain) => (
        <CSDStack direction="row" spacing={1}>
          {row.state === 'SHUTOFF' ? (
            <CSDActionButton size="small" color="success" onClick={() => handleStartDomain(row.uuid)} disabled={mutationLoading}>
              {t('hypervisors.start')}
            </CSDActionButton>
          ) : row.state === 'RUNNING' ? (
            <CSDActionButton size="small" color="warning" onClick={() => handleShutdownDomain(row.uuid)} disabled={mutationLoading}>
              {t('hypervisors.shutdown')}
            </CSDActionButton>
          ) : null}
        </CSDStack>
      ),
    },
  ];

  const networkColumns = [
    { id: 'name', label: 'common.name', render: (row: Network) => row.name },
    { id: 'bridge', label: 'hypervisors.bridge', render: (row: Network) => row.bridge },
    {
      id: 'active',
      label: 'hypervisors.active',
      render: (row: Network) => (
        <CSDChip label={row.active ? t('common.yes') : t('common.no')} color={row.active ? 'success' : 'default'} size="small" />
      ),
    },
    {
      id: 'autostart',
      label: 'hypervisors.autostart',
      render: (row: Network) => (
        <CSDChip label={row.autostart ? t('common.yes') : t('common.no')} color={row.autostart ? 'success' : 'default'} size="small" />
      ),
    },
  ];

  const storageColumns = [
    { id: 'name', label: 'common.name', render: (row: StoragePool) => row.name },
    { id: 'state', label: 'hypervisors.state', render: (row: StoragePool) => row.state },
    { id: 'capacity', label: 'hypervisors.capacity', render: (row: StoragePool) => formatBytes(row.capacity) },
    { id: 'allocation', label: 'hypervisors.used', render: (row: StoragePool) => formatBytes(row.allocation) },
    { id: 'available', label: 'hypervisors.available', render: (row: StoragePool) => formatBytes(row.available) },
    { id: 'volumesCount', label: 'hypervisors.volumes', render: (row: StoragePool) => row.volumesCount },
  ];

  const renderTabContent = () => {
    if (!isConnected) {
      return (
        <CSDStack spacing={2} alignItems="center" sx={{ py: 4 }}>
          <CSDTypography color="text.secondary">{t('hypervisors.connect_to_view')}</CSDTypography>
          <CSDActionButton variant="contained" onClick={handleTestConnection} disabled={mutationLoading}>
            {mutationLoading ? t('common.testing') : t('hypervisors.test_connection')}
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

    switch (tab) {
      case 0:
        return (
          <CSDDataGrid
            id="hypervisor-domains-grid"
            data={domainsData?.domains || []}
            columns={domainColumns}
            keyField="uuid"
          />
        );
      case 1:
        return (
          <CSDDataGrid
            id="hypervisor-networks-grid"
            data={networksData?.libvirtNetworks || []}
            columns={networkColumns}
            keyField="uuid"
          />
        );
      case 2:
        return (
          <CSDDataGrid
            id="hypervisor-storage-grid"
            data={storageData?.storagePools || []}
            columns={storageColumns}
            keyField="uuid"
          />
        );
      default:
        return null;
    }
  };

  return (
    <CSDDetailPage
      title={hypervisor.name}
      actions={
        <CSDActionButton
          variant="outlined"
          startIcon={<CSDIcon name="sync" />}
          onClick={handleTestConnection}
          disabled={mutationLoading}
        >
          {mutationLoading ? t('common.testing') : t('hypervisors.test_connection')}
        </CSDActionButton>
      }
    >
      <CSDStack spacing={3}>
        <CSDPaper>
          <CSDInfoGrid>
            <CSDInfoItem label={t('hypervisors.name')} value={hypervisor.name} />
            <CSDInfoItem
              label={t('hypervisors.mode')}
              value={<CSDChip label={hypervisor.mode} color={modeColors[hypervisor.mode] || 'default'} size="small" variant="outlined" />}
            />
            <CSDInfoItem
              label={t('hypervisors.status')}
              value={<CSDChip label={hypervisor.status} color={statusColors[hypervisor.status] || 'default'} size="small" />}
            />
            {hypervisor.driver && <CSDInfoItem label={t('hypervisors.driver')} value={hypervisor.driver} />}
            <CSDInfoItem label={t('hypervisors.uri')} value={hypervisor.uri} />
            <CSDInfoItem label={t('hypervisors.version')} value={hypervisor.version || '-'} />
            <CSDInfoItem label={t('hypervisors.hostname')} value={hypervisor.hostname || '-'} />
            {hypervisor.statusMessage && (
              <CSDInfoItem
                label={t('hypervisors.status_message')}
                value={<CSDTypography color={hypervisor.status === 'ERROR' ? 'error' : 'textPrimary'}>{hypervisor.statusMessage}</CSDTypography>}
                fullWidth
              />
            )}
            {hypervisor.description && <CSDInfoItem label={t('hypervisors.description')} value={hypervisor.description} fullWidth />}
            <CSDInfoItem label={t('common.created')} value={formatDate(hypervisor.createdAt, '-')} />
            <CSDInfoItem label={t('common.updated')} value={formatDate(hypervisor.updatedAt, '-')} />
          </CSDInfoGrid>
        </CSDPaper>

        <CSDPaper>
          <CSDTabs value={tab} onChange={(_: unknown, v: number) => setTab(v)}>
            <CSDTab label={`${t('hypervisors.virtual_machines')}${domainsData?.domainsCount ? ` (${domainsData.domainsCount})` : ''}`} />
            <CSDTab label={`${t('hypervisors.networks')}${networksData?.libvirtNetworksCount ? ` (${networksData.libvirtNetworksCount})` : ''}`} />
            <CSDTab label={`${t('hypervisors.storage_pools')}${storageData?.storagePoolsCount ? ` (${storageData.storagePoolsCount})` : ''}`} />
          </CSDTabs>
          <CSDStack spacing={2} sx={{ p: 2 }}>
            {renderTabContent()}
          </CSDStack>
        </CSDPaper>
      </CSDStack>
    </CSDDetailPage>
  );
};

export default HypervisorDetailPage;
