import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  CSDCrudPage,
  CSDDataGrid,
  CSDConfirmDialog,
  CSDFormDialog,
  CSDActionButton,
  CSDTextField,
  CSDSelect,
  CSDAutoComplete,
  CSDStack,
  CSDIcon,
  CSDChip,
  CSDTypography,
  CSDToggleButtonGroup,
  CSDToggleButton,
  CSDAlert,
  CSDBox,
  StatCardData,
} from 'csd_core/UI';
import {
  AdvancedFilterManager,
  usePagination,
  useTranslation,
  useBreadcrumb,
  formatDate,
  type SimpleFilterField,
  type FilterFieldDefinition,
} from 'csd_core/Providers';
import { useGraphQL } from '../../../shared/hooks/useGraphQL';
import { BREADCRUMBS } from '../../../shared/config/breadcrumbs';

// Types
interface Cluster {
  id: string;
  name: string;
  description: string;
  mode: string;
  distribution: string;
  version: string;
  status: string;
  statusMessage: string;
  createdAt: string;
  updatedAt: string;
}

interface Agent {
  id: string;
  name: string;
  hostname: string;
  status: string;
  supportedDistributions?: string[];
}

interface Distribution {
  id: string;
  name: string;
  description: string;
  deployable?: boolean;
}

interface ConnectClusterInput {
  name: string;
  description?: string;
  agentId: string;
  artifactKey: string;
  distribution?: string;
}

interface DeployClusterInput {
  name: string;
  description?: string;
  distribution: string;
  version?: string;
  masterNodes: string[];
  workerNodes?: string[];
}

// Status colors
const statusColors: Record<string, 'success' | 'error' | 'warning' | 'default' | 'info'> = {
  CONNECTED: 'success',
  DISCONNECTED: 'error',
  ERROR: 'error',
  PENDING: 'warning',
  DEPLOYING: 'info',
};

// Initial form states
const INITIAL_CONNECT_FORM: Partial<ConnectClusterInput> = {
  name: '',
  description: '',
  agentId: '',
  artifactKey: '',
  distribution: '',
};

const INITIAL_DEPLOY_FORM: Partial<DeployClusterInput> = {
  name: '',
  description: '',
  distribution: '',
  version: '',
  masterNodes: [],
  workerNodes: [],
};

export const ClustersPage: React.FC = () => {
  const { request } = useGraphQL();
  const { t } = useTranslation();
  const pagination = usePagination('clusters');

  // Breadcrumb
  useBreadcrumb([BREADCRUMBS.PILOTE, { labelKey: 'breadcrumb.clusters', icon: 'cloud' }]);

  // State
  const [clusters, setClusters] = useState<Cluster[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [totalCount, setTotalCount] = useState(0);

  // Reference data
  const [agents, setAgents] = useState<Agent[]>([]);
  const [deployAgents, setDeployAgents] = useState<Agent[]>([]);
  const [distributions, setDistributions] = useState<Distribution[]>([]);
  const [allDistributions, setAllDistributions] = useState<Distribution[]>([]);

  // Filter state
  const [simpleFilter, setSimpleFilter] = useState<Record<string, unknown> | null>(null);

  // Dialog state
  const [formOpen, setFormOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [bulkDeleteOpen, setBulkDeleteOpen] = useState(false);
  const [editingCluster, setEditingCluster] = useState<Cluster | null>(null);

  // Selection state for bulk actions
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

  // Form mode
  const [formMode, setFormMode] = useState<'CONNECT' | 'DEPLOY'>('CONNECT');

  // Form state
  const [connectForm, setConnectForm] = useState<Partial<ConnectClusterInput>>(INITIAL_CONNECT_FORM);
  const [deployForm, setDeployForm] = useState<Partial<DeployClusterInput>>(INITIAL_DEPLOY_FORM);
  const [initialConnectForm, setInitialConnectForm] = useState<Partial<ConnectClusterInput>>(INITIAL_CONNECT_FORM);
  const [initialDeployForm, setInitialDeployForm] = useState<Partial<DeployClusterInput>>(INITIAL_DEPLOY_FORM);
  const [formError, setFormError] = useState<string | null>(null);

  // Stats
  const [stats, setStats] = useState({
    total: 0,
    connected: 0,
    deploying: 0,
    error: 0,
  });

  // Check if form has changes
  const hasChanges = useMemo(() => {
    if (formMode === 'CONNECT') {
      return (
        connectForm.name !== initialConnectForm.name ||
        connectForm.description !== initialConnectForm.description ||
        connectForm.agentId !== initialConnectForm.agentId ||
        connectForm.artifactKey !== initialConnectForm.artifactKey ||
        connectForm.distribution !== initialConnectForm.distribution
      );
    }
    return (
      deployForm.name !== initialDeployForm.name ||
      deployForm.description !== initialDeployForm.description ||
      deployForm.distribution !== initialDeployForm.distribution ||
      deployForm.version !== initialDeployForm.version ||
      JSON.stringify(deployForm.masterNodes) !== JSON.stringify(initialDeployForm.masterNodes) ||
      JSON.stringify(deployForm.workerNodes) !== JSON.stringify(initialDeployForm.workerNodes)
    );
  }, [formMode, connectForm, deployForm, initialConnectForm, initialDeployForm]);

  // Check if form is valid
  const isFormValid = useMemo(() => {
    if (formMode === 'CONNECT') {
      return !!(connectForm.name?.trim() && connectForm.agentId && connectForm.artifactKey);
    }
    return !!(deployForm.name?.trim() && deployForm.distribution && deployForm.masterNodes?.length);
  }, [formMode, connectForm, deployForm]);

  // Agent options for selects
  const agentOptions = useMemo(() =>
    agents.map(a => ({ value: a.id, label: `${a.name} (${a.hostname})` })),
    [agents]
  );

  const deployAgentOptions = useMemo(() =>
    deployAgents.map(a => ({ value: a.id, label: `${a.name} (${a.hostname})` })),
    [deployAgents]
  );

  // Distribution options
  const distributionOptions = useMemo(() =>
    distributions.map(d => ({ value: d.id, label: `${d.name} - ${d.description}` })),
    [distributions]
  );

  const allDistributionOptions = useMemo(() =>
    [{ value: '', label: t('clusters.not_specified') }, ...allDistributions.map(d => ({ value: d.id, label: d.name }))],
    [allDistributions, t]
  );

  // Status options for filter
  const statusOptions = useMemo(() => [
    { value: '', label: t('common.all') },
    { value: 'CONNECTED', label: t('clusters.status_connected') },
    { value: 'DISCONNECTED', label: t('clusters.status_disconnected') },
    { value: 'DEPLOYING', label: t('clusters.status_deploying') },
    { value: 'PENDING', label: t('clusters.status_pending') },
    { value: 'ERROR', label: t('clusters.status_error') },
  ], [t]);

  // Mode options for filter
  const modeOptions = useMemo(() => [
    { value: '', label: t('common.all') },
    { value: 'CONNECT', label: t('clusters.mode_connect') },
    { value: 'DEPLOY', label: t('clusters.mode_deploy') },
  ], [t]);

  // Filter fields
  const filterFields: FilterFieldDefinition[] = useMemo(() => [
    { id: 'name', label: t('clusters.name'), type: 'string' },
    { id: 'status', label: t('clusters.status'), type: 'select', options: statusOptions },
    { id: 'mode', label: t('clusters.mode'), type: 'select', options: modeOptions },
  ], [t, statusOptions, modeOptions]);

  // Simple filters
  const simpleFilters: SimpleFilterField[] = useMemo(() => [
    { type: 'search', placeholder: 'clusters.search_placeholder' },
    { type: 'select', name: 'status', label: 'clusters.status', options: statusOptions },
  ], [statusOptions]);

  // Handle simple filter changes
  const handleSimpleFilterChange = useCallback((filters: Record<string, unknown>) => {
    const clusterFilter: Record<string, unknown> = {};
    if (filters.search && typeof filters.search === 'string' && filters.search.trim()) {
      clusterFilter.search = filters.search.trim();
    }
    if (filters.status && filters.status !== '') {
      clusterFilter.status = filters.status;
    }
    setSimpleFilter(Object.keys(clusterFilter).length > 0 ? clusterFilter : null);
  }, []);

  // Load reference data
  const loadReferenceData = useCallback(async () => {
    try {
      const [agentsRes, distribRes, allDistribRes] = await Promise.all([
        request<{ kubernetesAgents: Agent[] }>(`query { kubernetesAgents { id name hostname status } }`),
        request<{ kubernetesDistributions: Distribution[] }>(`query { kubernetesDistributions { id name description } }`),
        request<{ allKubernetesDistributions: Distribution[] }>(`query { allKubernetesDistributions { id name description deployable } }`),
      ]);
      setAgents(agentsRes.kubernetesAgents || []);
      setDistributions(distribRes.kubernetesDistributions || []);
      setAllDistributions(allDistribRes.allKubernetesDistributions || []);
    } catch {
      // Ignore reference data errors
    }
  }, [request]);

  // Load deploy agents when distribution changes
  useEffect(() => {
    if (formMode === 'DEPLOY' && deployForm.distribution) {
      const loadDeployAgents = async () => {
        try {
          const res = await request<{ kubernetesDeployAgents: Agent[] }>(
            `query($distribution: String) { kubernetesDeployAgents(distribution: $distribution) { id name hostname status supportedDistributions } }`,
            { distribution: deployForm.distribution?.toLowerCase() }
          );
          setDeployAgents(res.kubernetesDeployAgents || []);
        } catch {
          setDeployAgents([]);
        }
      };
      loadDeployAgents();
    }
  }, [formMode, deployForm.distribution, request]);

  // Load clusters
  const loadClusters = useCallback(async () => {
    setLoading(true);
    try {
      const listQuery = `
        query ListClusters($limit: Int, $offset: Int, $filter: ClusterFilter) {
          clusters(limit: $limit, offset: $offset, filter: $filter) {
            id name description mode distribution version status statusMessage createdAt updatedAt
          }
          clustersCount
        }
      `;

      const variables = {
        limit: pagination.rowsPerPage,
        offset: pagination.page * pagination.rowsPerPage,
        filter: simpleFilter,
      };

      const data = await request<{ clusters: Cluster[]; clustersCount: number }>(listQuery, variables);
      const clusterList = data.clusters || [];
      setClusters(clusterList);
      setTotalCount(data.clustersCount || clusterList.length);

      // Calculate stats
      setStats({
        total: data.clustersCount || clusterList.length,
        connected: clusterList.filter(c => c.status === 'CONNECTED').length,
        deploying: clusterList.filter(c => c.status === 'DEPLOYING').length,
        error: clusterList.filter(c => c.status === 'ERROR' || c.status === 'DISCONNECTED').length,
      });

      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('clusters.error_loading'));
    } finally {
      setLoading(false);
    }
  }, [request, pagination.page, pagination.rowsPerPage, simpleFilter, t]);

  useEffect(() => {
    loadClusters();
  }, [loadClusters]);

  useEffect(() => {
    loadReferenceData();
  }, [loadReferenceData]);

  // Handlers
  const handleOpen = (cluster: Cluster | null) => {
    setFormError(null);
    if (cluster) {
      setEditingCluster(cluster);
      setFormMode(cluster.mode as 'CONNECT' | 'DEPLOY');
      if (cluster.mode === 'CONNECT') {
        const formState = {
          name: cluster.name || '',
          description: cluster.description || '',
          agentId: '',
          artifactKey: '',
          distribution: cluster.distribution || '',
        };
        setConnectForm(formState);
        setInitialConnectForm(formState);
      }
    } else {
      setEditingCluster(null);
      setFormMode('CONNECT');
      setConnectForm(INITIAL_CONNECT_FORM);
      setInitialConnectForm(INITIAL_CONNECT_FORM);
      setDeployForm(INITIAL_DEPLOY_FORM);
      setInitialDeployForm(INITIAL_DEPLOY_FORM);
    }
    setFormOpen(true);
  };

  const handleClose = () => {
    setFormOpen(false);
    setEditingCluster(null);
    setFormError(null);
  };

  const handleDelete = (cluster: Cluster) => {
    setEditingCluster(cluster);
    setDeleteOpen(true);
  };

  const handleFormSubmit = async () => {
    try {
      if (formMode === 'CONNECT') {
        const mutation = `
          mutation CreateCluster($input: CreateClusterInput!) {
            createCluster(input: $input) { id }
          }
        `;
        await request(mutation, {
          input: {
            name: connectForm.name,
            description: connectForm.description || undefined,
            agentId: connectForm.agentId,
            artifactKey: connectForm.artifactKey,
            distribution: connectForm.distribution || undefined,
          },
        });
      } else {
        const mutation = `
          mutation DeployCluster($input: DeployClusterInput!) {
            deployCluster(input: $input) { id }
          }
        `;
        await request(mutation, {
          input: {
            name: deployForm.name,
            description: deployForm.description || undefined,
            distribution: deployForm.distribution,
            version: deployForm.version || undefined,
            masterNodes: deployForm.masterNodes,
            workerNodes: deployForm.workerNodes || [],
          },
        });
      }
      handleClose();
      await loadClusters();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : t('clusters.error_saving'));
    }
  };

  const handleConfirmDelete = async () => {
    if (!editingCluster) return;
    try {
      await request(`mutation DeleteCluster($id: ID!) { deleteCluster(id: $id) }`, { id: editingCluster.id });
      setDeleteOpen(false);
      setEditingCluster(null);
      await loadClusters();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('clusters.error_deleting'));
    }
  };

  const handleModeChange = (_: unknown, newMode: 'CONNECT' | 'DEPLOY' | null) => {
    if (newMode && !editingCluster) {
      setFormMode(newMode);
    }
  };

  // Selection handlers
  const handleSelectionChange = (ids: Set<string>) => {
    setSelectedIds(ids);
  };

  const handleBulkDelete = async () => {
    if (selectedIds.size === 0) return;
    try {
      await request(
        `mutation BulkDeleteClusters($ids: [ID!]!) { bulkDeleteClusters(ids: $ids) }`,
        { ids: Array.from(selectedIds) }
      );
      setBulkDeleteOpen(false);
      setSelectedIds(new Set());
      await loadClusters();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('clusters.error_bulk_delete'));
    }
  };

  // Stat cards
  const statCards: StatCardData[] = [
    { title: t('clusters.total'), value: String(stats.total), icon: 'cloud', color: 'primary' },
    { title: t('clusters.connected'), value: String(stats.connected), icon: 'checkCircle', color: 'success' },
    { title: t('clusters.deploying'), value: String(stats.deploying), icon: 'sync', color: 'info' },
    { title: t('clusters.errors'), value: String(stats.error), icon: 'error', color: 'error' },
  ];

  // DataGrid columns
  const columns = [
    {
      id: 'name',
      label: 'clusters.name',
      sortable: true,
      render: (cluster: Cluster) => (
        <CSDTypography variant="body2" fontWeight="bold">{cluster.name}</CSDTypography>
      ),
    },
    {
      id: 'mode',
      label: 'clusters.mode',
      sortable: true,
      render: (cluster: Cluster) => (
        <CSDChip label={cluster.mode} size="small" color={cluster.mode === 'DEPLOY' ? 'secondary' : 'primary'} />
      ),
    },
    {
      id: 'distribution',
      label: 'clusters.distribution',
      sortable: true,
      render: (cluster: Cluster) => cluster.distribution || '-',
    },
    {
      id: 'status',
      label: 'clusters.status',
      sortable: true,
      render: (cluster: Cluster) => (
        <CSDChip label={cluster.status} size="small" color={statusColors[cluster.status] || 'default'} />
      ),
    },
    {
      id: 'createdAt',
      label: 'clusters.created',
      sortable: true,
      render: (cluster: Cluster) => formatDate(cluster.createdAt, '-'),
    },
  ];

  // Row actions
  const actions = [
    { icon: 'visibility', onClick: (cluster: Cluster) => window.location.href = `/pilote/kubernetes/clusters/${cluster.id}`, tooltip: 'common.view', color: 'primary' as const },
    { icon: 'delete', onClick: (cluster: Cluster) => handleDelete(cluster), tooltip: 'common.delete', color: 'error' as const },
  ];

  // Form dialog
  const formDialog = (
    <CSDFormDialog
      id="clusters-form-dialog"
      open={formOpen}
      onClose={handleClose}
      title={editingCluster ? 'clusters.edit_cluster' : 'clusters.add_cluster'}
      icon={editingCluster ? 'edit' : 'add'}
      error={formError}
      trackChanges={!!editingCluster}
      hasChanges={hasChanges}
      onResetAll={() => {
        if (formMode === 'CONNECT') setConnectForm(initialConnectForm);
        else setDeployForm(initialDeployForm);
      }}
      actions={
        <>
          <CSDActionButton variant="outlined" onClick={handleClose}>
            {t('common.cancel')}
          </CSDActionButton>
          <CSDActionButton
            variant="contained"
            onClick={handleFormSubmit}
            disabled={!isFormValid || (editingCluster && !hasChanges)}
          >
            {formMode === 'CONNECT' ? t('clusters.connect') : t('clusters.deploy')}
          </CSDActionButton>
        </>
      }
    >
      <CSDStack spacing={2}>
        {!editingCluster && (
          <CSDToggleButtonGroup value={formMode} exclusive onChange={handleModeChange} fullWidth>
            <CSDToggleButton value="CONNECT">{t('clusters.mode_connect')}</CSDToggleButton>
            <CSDToggleButton value="DEPLOY">{t('clusters.mode_deploy')}</CSDToggleButton>
          </CSDToggleButtonGroup>
        )}

        {formMode === 'CONNECT' ? (
          <>
            <CSDAlert severity="info">{t('clusters.connect_description')}</CSDAlert>
            <CSDTextField
              id="clusters-form-name"
              name="name"
              label={t('clusters.name')}
              value={connectForm.name || ''}
              onChange={(v) => setConnectForm({ ...connectForm, name: v })}
              required
              fullWidth
            />
            <CSDTextField
              id="clusters-form-description"
              name="description"
              label={t('clusters.description')}
              value={connectForm.description || ''}
              onChange={(v) => setConnectForm({ ...connectForm, description: v })}
              multiline
              rows={2}
              fullWidth
            />
            <CSDSelect
              id="clusters-form-agent"
              name="agentId"
              label={t('clusters.agent')}
              value={connectForm.agentId || ''}
              onChange={(v) => setConnectForm({ ...connectForm, agentId: v })}
              options={agentOptions}
              required
              fullWidth
            />
            <CSDTextField
              id="clusters-form-artifact"
              name="artifactKey"
              label={t('clusters.kubeconfig_artifact')}
              value={connectForm.artifactKey || ''}
              onChange={(v) => setConnectForm({ ...connectForm, artifactKey: v })}
              required
              fullWidth
              helperText={t('clusters.kubeconfig_artifact_help')}
            />
            <CSDSelect
              id="clusters-form-distribution"
              name="distribution"
              label={t('clusters.distribution_optional')}
              value={connectForm.distribution || ''}
              onChange={(v) => setConnectForm({ ...connectForm, distribution: v })}
              options={allDistributionOptions}
              fullWidth
            />
          </>
        ) : (
          <>
            <CSDAlert severity="info">{t('clusters.deploy_description')}</CSDAlert>
            <CSDTextField
              id="clusters-form-name"
              name="name"
              label={t('clusters.name')}
              value={deployForm.name || ''}
              onChange={(v) => setDeployForm({ ...deployForm, name: v })}
              required
              fullWidth
            />
            <CSDTextField
              id="clusters-form-description"
              name="description"
              label={t('clusters.description')}
              value={deployForm.description || ''}
              onChange={(v) => setDeployForm({ ...deployForm, description: v })}
              multiline
              rows={2}
              fullWidth
            />
            <CSDSelect
              id="clusters-form-distribution"
              name="distribution"
              label={t('clusters.distribution')}
              value={deployForm.distribution || ''}
              onChange={(v) => setDeployForm({ ...deployForm, distribution: v, masterNodes: [], workerNodes: [] })}
              options={distributionOptions}
              required
              fullWidth
            />
            <CSDTextField
              id="clusters-form-version"
              name="version"
              label={t('clusters.version')}
              value={deployForm.version || ''}
              onChange={(v) => setDeployForm({ ...deployForm, version: v })}
              fullWidth
              helperText={t('clusters.version_help')}
            />
            <CSDAutoComplete
              id="clusters-form-masters"
              name="masterNodes"
              label={t('clusters.master_nodes')}
              value={deployForm.masterNodes || []}
              onChange={(values: string[]) => setDeployForm({ ...deployForm, masterNodes: values })}
              options={deployAgentOptions}
              multiple
              required
              fullWidth
              helperText={t('clusters.master_nodes_help')}
            />
            <CSDAutoComplete
              id="clusters-form-workers"
              name="workerNodes"
              label={t('clusters.worker_nodes')}
              value={deployForm.workerNodes || []}
              onChange={(values: string[]) => setDeployForm({ ...deployForm, workerNodes: values })}
              options={deployAgentOptions.filter(o => !deployForm.masterNodes?.includes(o.value))}
              multiple
              fullWidth
              helperText={t('clusters.worker_nodes_help')}
            />
          </>
        )}
      </CSDStack>
    </CSDFormDialog>
  );

  // Delete dialog
  const deleteDialog = (
    <>
      <CSDConfirmDialog
        id="clusters-delete-dialog"
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        onConfirm={handleConfirmDelete}
        title={t('clusters.delete_cluster')}
        message={t('clusters.delete_confirmation', { name: editingCluster?.name || '' })}
        confirmLabel={t('common.delete')}
        cancelLabel={t('common.cancel')}
        severity="error"
      />
      <CSDConfirmDialog
        id="clusters-bulk-delete-dialog"
        open={bulkDeleteOpen}
        onClose={() => setBulkDeleteOpen(false)}
        onConfirm={handleBulkDelete}
        title={t('clusters.bulk_delete')}
        message={t('clusters.bulk_delete_confirmation', { count: selectedIds.size })}
        confirmLabel={t('common.delete')}
        cancelLabel={t('common.cancel')}
        severity="error"
      />
    </>
  );

  return (
    <AdvancedFilterManager
      storageKey="clusters"
      filterFields={filterFields}
      simpleFilters={simpleFilters}
      data={clusters}
      onSimpleFilterChange={handleSimpleFilterChange}
      pagination={pagination}
      autoSave={false}
      useServerFiltering={true}
      actions={
        <CSDBox sx={{ display: 'flex', gap: 1 }}>
          {selectedIds.size > 0 && (
            <CSDActionButton
              id="clusters-bulk-delete-button"
              variant="outlined"
              color="error"
              startIcon={<CSDIcon name="delete" />}
              onClick={() => setBulkDeleteOpen(true)}
            >
              {t('common.delete_selected', { count: selectedIds.size })}
            </CSDActionButton>
          )}
          <CSDActionButton
            id="clusters-add-button"
            variant="contained"
            startIcon={<CSDIcon name="add" />}
            onClick={() => handleOpen(null)}
          >
            {t('clusters.add_cluster')}
          </CSDActionButton>
        </CSDBox>
      }
    >
      {({ filteredData, filterBar }) => (
        <CSDCrudPage
          entityName="clusters"
          statCards={statCards}
          filterBar={filterBar}
          loading={loading}
          error={error}
          formDialog={formDialog}
          deleteDialog={deleteDialog}
        >
          <CSDDataGrid
            id="clusters-data-grid"
            data={filteredData}
            columns={columns}
            actions={actions}
            loading={loading}
            pagination={pagination}
            totalCount={totalCount}
            keyField="id"
            hoverable
            selectable
            selectedIds={selectedIds}
            onSelectionChange={handleSelectionChange}
          />
        </CSDCrudPage>
      )}
    </AdvancedFilterManager>
  );
};

export default ClustersPage;
