import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  CSDCrudPage,
  CSDDataGrid,
  CSDConfirmDialog,
  CSDFormDialog,
  CSDActionButton,
  CSDTextField,
  CSDSelect,
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
  createdAt: string;
  updatedAt: string;
}

interface Agent {
  id: string;
  name: string;
  hostname: string;
  status: string;
  supportedDrivers?: string[];
}

interface Driver {
  id: string;
  name: string;
  description: string;
}

interface ConnectHypervisorInput {
  name: string;
  description?: string;
  agentId: string;
  uri: string;
  artifactKey?: string;
}

interface DeployHypervisorInput {
  name: string;
  description?: string;
  agentId: string;
  driver: string;
}

// Status colors
const statusColors: Record<string, 'success' | 'error' | 'warning' | 'default' | 'info'> = {
  CONNECTED: 'success',
  DISCONNECTED: 'error',
  ERROR: 'error',
  PENDING: 'warning',
  DEPLOYING: 'info',
};

// Mode colors
const modeColors: Record<string, 'primary' | 'secondary'> = {
  CONNECT: 'primary',
  DEPLOY: 'secondary',
};

// Initial form states
const INITIAL_CONNECT_FORM: Partial<ConnectHypervisorInput> = {
  name: '',
  description: '',
  agentId: '',
  uri: '',
  artifactKey: '',
};

const INITIAL_DEPLOY_FORM: Partial<DeployHypervisorInput> = {
  name: '',
  description: '',
  agentId: '',
  driver: '',
};

export const HypervisorsPage: React.FC = () => {
  const { request } = useGraphQL();
  const { t } = useTranslation();
  const pagination = usePagination('hypervisors');

  // Breadcrumb
  useBreadcrumb([BREADCRUMBS.PILOTE, BREADCRUMBS.LIBVIRT, { labelKey: 'breadcrumb.hypervisors', icon: 'server' }]);

  // State
  const [hypervisors, setHypervisors] = useState<Hypervisor[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [totalCount, setTotalCount] = useState(0);

  // Reference data
  const [connectAgents, setConnectAgents] = useState<Agent[]>([]);
  const [deployAgents, setDeployAgents] = useState<Agent[]>([]);
  const [drivers, setDrivers] = useState<Driver[]>([]);

  // Filter state
  const [simpleFilter, setSimpleFilter] = useState<Record<string, unknown> | null>(null);

  // Dialog state
  const [formOpen, setFormOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [bulkDeleteOpen, setBulkDeleteOpen] = useState(false);
  const [editingHypervisor, setEditingHypervisor] = useState<Hypervisor | null>(null);

  // Selection state for bulk actions
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

  // Form mode
  const [formMode, setFormMode] = useState<'CONNECT' | 'DEPLOY'>('CONNECT');

  // Form state
  const [connectForm, setConnectForm] = useState<Partial<ConnectHypervisorInput>>(INITIAL_CONNECT_FORM);
  const [deployForm, setDeployForm] = useState<Partial<DeployHypervisorInput>>(INITIAL_DEPLOY_FORM);
  const [initialConnectForm, setInitialConnectForm] = useState<Partial<ConnectHypervisorInput>>(INITIAL_CONNECT_FORM);
  const [initialDeployForm, setInitialDeployForm] = useState<Partial<DeployHypervisorInput>>(INITIAL_DEPLOY_FORM);
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
        connectForm.uri !== initialConnectForm.uri ||
        connectForm.artifactKey !== initialConnectForm.artifactKey
      );
    }
    return (
      deployForm.name !== initialDeployForm.name ||
      deployForm.description !== initialDeployForm.description ||
      deployForm.agentId !== initialDeployForm.agentId ||
      deployForm.driver !== initialDeployForm.driver
    );
  }, [formMode, connectForm, deployForm, initialConnectForm, initialDeployForm]);

  // Check if form is valid
  const isFormValid = useMemo(() => {
    if (formMode === 'CONNECT') {
      return !!(connectForm.name?.trim() && connectForm.agentId && connectForm.uri?.trim());
    }
    return !!(deployForm.name?.trim() && deployForm.agentId && deployForm.driver);
  }, [formMode, connectForm, deployForm]);

  // Agent options for selects
  const connectAgentOptions = useMemo(() =>
    connectAgents.map(a => ({ value: a.id, label: `${a.name} (${a.hostname})` })),
    [connectAgents]
  );

  const deployAgentOptions = useMemo(() =>
    deployAgents.map(a => ({ value: a.id, label: `${a.name} (${a.hostname})` })),
    [deployAgents]
  );

  // Driver options
  const driverOptions = useMemo(() =>
    drivers.map(d => ({ value: d.id, label: `${d.name} - ${d.description}` })),
    [drivers]
  );

  // Status options for filter
  const statusOptions = useMemo(() => [
    { value: '', label: t('common.all') },
    { value: 'CONNECTED', label: t('hypervisors.status_connected') },
    { value: 'DISCONNECTED', label: t('hypervisors.status_disconnected') },
    { value: 'DEPLOYING', label: t('hypervisors.status_deploying') },
    { value: 'PENDING', label: t('hypervisors.status_pending') },
    { value: 'ERROR', label: t('hypervisors.status_error') },
  ], [t]);

  // Mode options for filter
  const modeOptions = useMemo(() => [
    { value: '', label: t('common.all') },
    { value: 'CONNECT', label: t('hypervisors.mode_connect') },
    { value: 'DEPLOY', label: t('hypervisors.mode_deploy') },
  ], [t]);

  // Filter fields
  const filterFields: FilterFieldDefinition[] = useMemo(() => [
    { id: 'name', label: t('hypervisors.name'), type: 'string' },
    { id: 'status', label: t('hypervisors.status'), type: 'select', options: statusOptions },
    { id: 'mode', label: t('hypervisors.mode'), type: 'select', options: modeOptions },
  ], [t, statusOptions, modeOptions]);

  // Simple filters
  const simpleFilters: SimpleFilterField[] = useMemo(() => [
    { type: 'search', placeholder: 'hypervisors.search_placeholder' },
    { type: 'select', name: 'status', label: 'hypervisors.status', options: statusOptions },
  ], [statusOptions]);

  // Handle simple filter changes
  const handleSimpleFilterChange = useCallback((filters: Record<string, unknown>) => {
    const hypervisorFilter: Record<string, unknown> = {};
    if (filters.search && typeof filters.search === 'string' && filters.search.trim()) {
      hypervisorFilter.search = filters.search.trim();
    }
    if (filters.status && filters.status !== '') {
      hypervisorFilter.status = filters.status;
    }
    setSimpleFilter(Object.keys(hypervisorFilter).length > 0 ? hypervisorFilter : null);
  }, []);

  // Load reference data
  const loadReferenceData = useCallback(async () => {
    try {
      const [agentsRes, driversRes] = await Promise.all([
        request<{ libvirtAgents: Agent[] }>(`query LibvirtAgents { libvirtAgents { id name hostname status } }`),
        request<{ libvirtDrivers: Driver[] }>(`query LibvirtDrivers { libvirtDrivers { id name description } }`),
      ]);
      setConnectAgents(agentsRes.libvirtAgents || []);
      setDrivers(driversRes.libvirtDrivers || []);
    } catch {
      // Ignore reference data errors
    }
  }, [request]);

  // Load deploy agents when driver changes
  useEffect(() => {
    if (formMode === 'DEPLOY' && deployForm.driver) {
      const loadDeployAgents = async () => {
        try {
          const res = await request<{ libvirtDeployAgents: Agent[] }>(
            `query LibvirtDeployAgents($driver: String) { libvirtDeployAgents(driver: $driver) { id name hostname status supportedDrivers } }`,
            { driver: deployForm.driver?.toLowerCase() }
          );
          setDeployAgents(res.libvirtDeployAgents || []);
        } catch {
          setDeployAgents([]);
        }
      };
      loadDeployAgents();
    }
  }, [formMode, deployForm.driver, request]);

  // Load hypervisors
  const loadHypervisors = useCallback(async () => {
    setLoading(true);
    try {
      const listQuery = `
        query ListHypervisors($limit: Int, $offset: Int, $filter: HypervisorFilter) {
          hypervisors(limit: $limit, offset: $offset, filter: $filter) {
            id name description mode driver uri status statusMessage version hostname createdAt updatedAt
          }
          hypervisorsCount
        }
      `;

      const variables = {
        limit: pagination.rowsPerPage,
        offset: pagination.page * pagination.rowsPerPage,
        filter: simpleFilter,
      };

      const data = await request<{ hypervisors: Hypervisor[]; hypervisorsCount: number }>(listQuery, variables);
      const hypervisorList = data.hypervisors || [];
      setHypervisors(hypervisorList);
      setTotalCount(data.hypervisorsCount || hypervisorList.length);

      // Calculate stats
      setStats({
        total: data.hypervisorsCount || hypervisorList.length,
        connected: hypervisorList.filter(h => h.status === 'CONNECTED').length,
        deploying: hypervisorList.filter(h => h.status === 'DEPLOYING').length,
        error: hypervisorList.filter(h => h.status === 'ERROR' || h.status === 'DISCONNECTED').length,
      });

      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('hypervisors.error_loading'));
    } finally {
      setLoading(false);
    }
  }, [request, pagination.page, pagination.rowsPerPage, simpleFilter, t]);

  useEffect(() => {
    loadHypervisors();
  }, [loadHypervisors]);

  useEffect(() => {
    loadReferenceData();
  }, [loadReferenceData]);

  // Handlers
  const handleOpen = (hypervisor: Hypervisor | null) => {
    setFormError(null);
    if (hypervisor) {
      setEditingHypervisor(hypervisor);
      setFormMode(hypervisor.mode as 'CONNECT' | 'DEPLOY');
      if (hypervisor.mode === 'CONNECT') {
        const formState = {
          name: hypervisor.name || '',
          description: hypervisor.description || '',
          agentId: '',
          uri: hypervisor.uri || '',
          artifactKey: '',
        };
        setConnectForm(formState);
        setInitialConnectForm(formState);
      }
    } else {
      setEditingHypervisor(null);
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
    setEditingHypervisor(null);
    setFormError(null);
  };

  const handleDelete = (hypervisor: Hypervisor) => {
    setEditingHypervisor(hypervisor);
    setDeleteOpen(true);
  };

  const handleFormSubmit = async () => {
    try {
      if (formMode === 'CONNECT') {
        const mutation = `
          mutation CreateHypervisor($input: CreateHypervisorInput!) {
            createHypervisor(input: $input) { id }
          }
        `;
        await request(mutation, {
          input: {
            name: connectForm.name,
            description: connectForm.description || undefined,
            agentId: connectForm.agentId,
            uri: connectForm.uri,
            artifactKey: connectForm.artifactKey || undefined,
          },
        });
      } else {
        const mutation = `
          mutation DeployHypervisor($input: DeployHypervisorInput!) {
            deployHypervisor(input: $input) { id }
          }
        `;
        await request(mutation, {
          input: {
            name: deployForm.name,
            description: deployForm.description || undefined,
            agentId: deployForm.agentId,
            driver: deployForm.driver,
          },
        });
      }
      handleClose();
      await loadHypervisors();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : t('hypervisors.error_saving'));
    }
  };

  const handleConfirmDelete = async () => {
    if (!editingHypervisor) return;
    try {
      await request(`mutation DeleteHypervisor($id: ID!) { deleteHypervisor(id: $id) }`, { id: editingHypervisor.id });
      setDeleteOpen(false);
      setEditingHypervisor(null);
      await loadHypervisors();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('hypervisors.error_deleting'));
    }
  };

  const handleModeChange = (_: unknown, newMode: 'CONNECT' | 'DEPLOY' | null) => {
    if (newMode && !editingHypervisor) {
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
        `mutation BulkDeleteHypervisors($ids: [ID!]!) { bulkDeleteHypervisors(ids: $ids) }`,
        { ids: Array.from(selectedIds) }
      );
      setBulkDeleteOpen(false);
      setSelectedIds(new Set());
      await loadHypervisors();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('hypervisors.error_bulk_delete'));
    }
  };

  // Stat cards
  const statCards: StatCardData[] = [
    { title: t('hypervisors.total'), value: String(stats.total), icon: 'server', color: 'primary' },
    { title: t('hypervisors.connected'), value: String(stats.connected), icon: 'checkCircle', color: 'success' },
    { title: t('hypervisors.deploying'), value: String(stats.deploying), icon: 'sync', color: 'info' },
    { title: t('hypervisors.errors'), value: String(stats.error), icon: 'error', color: 'error' },
  ];

  // DataGrid columns
  const columns = [
    {
      id: 'name',
      label: 'hypervisors.name',
      sortable: true,
      render: (hypervisor: Hypervisor) => (
        <CSDTypography variant="body2" fontWeight="bold">{hypervisor.name}</CSDTypography>
      ),
    },
    {
      id: 'mode',
      label: 'hypervisors.mode',
      sortable: true,
      render: (hypervisor: Hypervisor) => (
        <CSDChip label={hypervisor.mode} size="small" color={modeColors[hypervisor.mode] || 'default'} variant="outlined" />
      ),
    },
    {
      id: 'driver',
      label: 'hypervisors.driver',
      sortable: true,
      render: (hypervisor: Hypervisor) => hypervisor.driver || '-',
    },
    {
      id: 'uri',
      label: 'hypervisors.uri',
      sortable: false,
      render: (hypervisor: Hypervisor) => hypervisor.uri || '-',
    },
    {
      id: 'status',
      label: 'hypervisors.status',
      sortable: true,
      render: (hypervisor: Hypervisor) => (
        <CSDChip label={hypervisor.status} size="small" color={statusColors[hypervisor.status] || 'default'} />
      ),
    },
    {
      id: 'createdAt',
      label: 'hypervisors.created',
      sortable: true,
      render: (hypervisor: Hypervisor) => formatDate(hypervisor.createdAt, '-'),
    },
  ];

  // Row actions
  const actions = [
    { icon: 'visibility', onClick: (hypervisor: Hypervisor) => window.location.href = `/pilote/libvirt/hypervisors/${hypervisor.id}`, tooltip: 'common.view', color: 'primary' as const },
    { icon: 'delete', onClick: (hypervisor: Hypervisor) => handleDelete(hypervisor), tooltip: 'common.delete', color: 'error' as const },
  ];

  // Form dialog
  const formDialog = (
    <CSDFormDialog
      id="hypervisors-form-dialog"
      open={formOpen}
      onClose={handleClose}
      title={editingHypervisor ? 'hypervisors.edit_hypervisor' : 'hypervisors.add_hypervisor'}
      icon={editingHypervisor ? 'edit' : 'add'}
      error={formError}
      trackChanges={!!editingHypervisor}
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
            disabled={!isFormValid || (editingHypervisor && !hasChanges)}
          >
            {formMode === 'CONNECT' ? t('hypervisors.connect') : t('hypervisors.deploy')}
          </CSDActionButton>
        </>
      }
    >
      <CSDStack spacing={2}>
        {!editingHypervisor && (
          <CSDToggleButtonGroup value={formMode} exclusive onChange={handleModeChange} fullWidth>
            <CSDToggleButton value="CONNECT">{t('hypervisors.mode_connect')}</CSDToggleButton>
            <CSDToggleButton value="DEPLOY">{t('hypervisors.mode_deploy')}</CSDToggleButton>
          </CSDToggleButtonGroup>
        )}

        {formMode === 'CONNECT' ? (
          <>
            <CSDAlert severity="info">{t('hypervisors.connect_description')}</CSDAlert>
            <CSDTextField
              id="hypervisors-form-name"
              name="name"
              label={t('hypervisors.name')}
              value={connectForm.name || ''}
              onChange={(v) => setConnectForm({ ...connectForm, name: v })}
              required
              fullWidth
            />
            <CSDTextField
              id="hypervisors-form-description"
              name="description"
              label={t('hypervisors.description')}
              value={connectForm.description || ''}
              onChange={(v) => setConnectForm({ ...connectForm, description: v })}
              multiline
              rows={2}
              fullWidth
            />
            <CSDSelect
              id="hypervisors-form-agent"
              name="agentId"
              label={t('hypervisors.agent')}
              value={connectForm.agentId || ''}
              onChange={(v) => setConnectForm({ ...connectForm, agentId: v })}
              options={connectAgentOptions}
              required
              fullWidth
            />
            <CSDTextField
              id="hypervisors-form-uri"
              name="uri"
              label={t('hypervisors.libvirt_uri')}
              value={connectForm.uri || ''}
              onChange={(v) => setConnectForm({ ...connectForm, uri: v })}
              required
              fullWidth
              helperText={t('hypervisors.uri_help')}
            />
            <CSDTextField
              id="hypervisors-form-artifact"
              name="artifactKey"
              label={t('hypervisors.ssh_key_artifact')}
              value={connectForm.artifactKey || ''}
              onChange={(v) => setConnectForm({ ...connectForm, artifactKey: v })}
              fullWidth
              helperText={t('hypervisors.ssh_key_artifact_help')}
            />
          </>
        ) : (
          <>
            <CSDAlert severity="info">{t('hypervisors.deploy_description')}</CSDAlert>
            <CSDTextField
              id="hypervisors-form-name"
              name="name"
              label={t('hypervisors.name')}
              value={deployForm.name || ''}
              onChange={(v) => setDeployForm({ ...deployForm, name: v })}
              required
              fullWidth
            />
            <CSDTextField
              id="hypervisors-form-description"
              name="description"
              label={t('hypervisors.description')}
              value={deployForm.description || ''}
              onChange={(v) => setDeployForm({ ...deployForm, description: v })}
              multiline
              rows={2}
              fullWidth
            />
            <CSDSelect
              id="hypervisors-form-driver"
              name="driver"
              label={t('hypervisors.driver')}
              value={deployForm.driver || ''}
              onChange={(v) => setDeployForm({ ...deployForm, driver: v, agentId: '' })}
              options={driverOptions}
              required
              fullWidth
            />
            <CSDSelect
              id="hypervisors-form-agent"
              name="agentId"
              label={t('hypervisors.agent')}
              value={deployForm.agentId || ''}
              onChange={(v) => setDeployForm({ ...deployForm, agentId: v })}
              options={deployAgentOptions}
              required
              fullWidth
              disabled={!deployForm.driver}
            />
            {deployForm.driver && deployAgents.length === 0 && (
              <CSDAlert severity="warning">
                {t('hypervisors.no_agents_for_driver', { driver: deployForm.driver })}
              </CSDAlert>
            )}
          </>
        )}
      </CSDStack>
    </CSDFormDialog>
  );

  // Delete dialog
  const deleteDialog = (
    <>
      <CSDConfirmDialog
        id="hypervisors-delete-dialog"
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        onConfirm={handleConfirmDelete}
        title={t('hypervisors.delete_hypervisor')}
        message={t('hypervisors.delete_confirmation', { name: editingHypervisor?.name || '' })}
        confirmLabel={t('common.delete')}
        cancelLabel={t('common.cancel')}
        severity="error"
      />
      <CSDConfirmDialog
        id="hypervisors-bulk-delete-dialog"
        open={bulkDeleteOpen}
        onClose={() => setBulkDeleteOpen(false)}
        onConfirm={handleBulkDelete}
        title={t('hypervisors.bulk_delete')}
        message={t('hypervisors.bulk_delete_confirmation', { count: selectedIds.size })}
        confirmLabel={t('common.delete')}
        cancelLabel={t('common.cancel')}
        severity="error"
      />
    </>
  );

  return (
    <AdvancedFilterManager
      storageKey="hypervisors"
      filterFields={filterFields}
      simpleFilters={simpleFilters}
      data={hypervisors}
      onSimpleFilterChange={handleSimpleFilterChange}
      pagination={pagination}
      autoSave={false}
      useServerFiltering={true}
      actions={
        <CSDBox sx={{ display: 'flex', gap: 1 }}>
          {selectedIds.size > 0 && (
            <CSDActionButton
              id="hypervisors-bulk-delete-button"
              variant="outlined"
              color="error"
              startIcon={<CSDIcon name="delete" />}
              onClick={() => setBulkDeleteOpen(true)}
            >
              {t('common.delete_selected', { count: selectedIds.size })}
            </CSDActionButton>
          )}
          <CSDActionButton
            id="hypervisors-add-button"
            variant="contained"
            startIcon={<CSDIcon name="add" />}
            onClick={() => handleOpen(null)}
          >
            {t('hypervisors.add_hypervisor')}
          </CSDActionButton>
        </CSDBox>
      }
    >
      {({ filteredData, filterBar }) => (
        <CSDCrudPage
          entityName="hypervisors"
          statCards={statCards}
          filterBar={filterBar}
          loading={loading}
          error={error}
          formDialog={formDialog}
          deleteDialog={deleteDialog}
        >
          <CSDDataGrid
            id="hypervisors-data-grid"
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

export default HypervisorsPage;
