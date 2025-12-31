import React, { useState, useEffect, useCallback, useMemo } from 'react';
import {
  CSDCrudPage,
  CSDDataGrid,
  CSDConfirmDialog,
  CSDFormDialog,
  CSDActionButton,
  CSDSelect,
  CSDStack,
  CSDIcon,
  CSDChip,
  CSDTypography,
  CSDBox,
  CSDAlert,
  CSDSwitch,
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
interface FirewallDeployment {
  id: string;
  profileId: string;
  profile: { id: string; name: string } | null;
  agentId: string;
  agentName: string;
  action: string;
  status: string;
  statusMessage: string;
  output: string;
  startedAt: string;
  completedAt: string;
  createdAt: string;
}

interface FirewallProfile {
  id: string;
  name: string;
}

interface Agent {
  id: string;
  name: string;
  hostname: string;
}

// Status colors
const statusColors: Record<string, 'success' | 'error' | 'warning' | 'default' | 'info'> = {
  pending: 'warning',
  deploying: 'info',
  applied: 'success',
  error: 'error',
  rolled_back: 'default',
};

// Action colors
const actionColors: Record<string, 'primary' | 'secondary' | 'success' | 'warning' | 'error' | 'info'> = {
  apply: 'primary',
  rollback: 'warning',
  audit: 'info',
  flush: 'error',
};

export const DeploymentsPage: React.FC = () => {
  const { request } = useGraphQL();
  const { t } = useTranslation();
  const pagination = usePagination('security-deployments');

  // Breadcrumb
  useBreadcrumb([
    BREADCRUMBS.PILOTE,
    BREADCRUMBS.SECURITY,
    { labelKey: 'breadcrumb.deployments', icon: 'rocket' },
  ]);

  // State
  const [deployments, setDeployments] = useState<FirewallDeployment[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [totalCount, setTotalCount] = useState(0);

  // Reference data
  const [profiles, setProfiles] = useState<FirewallProfile[]>([]);
  const [agents, setAgents] = useState<Agent[]>([]);

  // Filter state
  const [simpleFilter, setSimpleFilter] = useState<Record<string, unknown> | null>(null);

  // Dialog state
  const [deployOpen, setDeployOpen] = useState(false);
  const [rollbackOpen, setRollbackOpen] = useState(false);
  const [flushOpen, setFlushOpen] = useState(false);
  const [selectedDeployment, setSelectedDeployment] = useState<FirewallDeployment | null>(null);

  // Deploy form state
  const [selectedProfileId, setSelectedProfileId] = useState<string>('');
  const [selectedAgentId, setSelectedAgentId] = useState<string>('');
  const [dryRun, setDryRun] = useState<boolean>(false);
  const [deployError, setDeployError] = useState<string | null>(null);

  // Stats
  const [stats, setStats] = useState({
    total: 0,
    applied: 0,
    pending: 0,
    error: 0,
  });

  // Profile options
  const profileOptions = useMemo(() =>
    profiles.map(p => ({ value: p.id, label: p.name })),
    [profiles]
  );

  // Agent options
  const agentOptions = useMemo(() =>
    agents.map(a => ({ value: a.id, label: `${a.name} (${a.hostname})` })),
    [agents]
  );

  // Status filter options
  const statusFilterOptions = useMemo(() => [
    { value: '', label: t('common.all') },
    { value: 'pending', label: t('security.deployments.status_pending') },
    { value: 'deploying', label: t('security.deployments.status_deploying') },
    { value: 'applied', label: t('security.deployments.status_applied') },
    { value: 'error', label: t('security.deployments.status_error') },
    { value: 'rolled_back', label: t('security.deployments.status_rolled_back') },
  ], [t]);

  // Action filter options
  const actionFilterOptions = useMemo(() => [
    { value: '', label: t('common.all') },
    { value: 'apply', label: t('security.deployments.action_apply') },
    { value: 'rollback', label: t('security.deployments.action_rollback') },
    { value: 'audit', label: t('security.deployments.action_audit') },
    { value: 'flush', label: t('security.deployments.action_flush') },
  ], [t]);

  // Filter fields
  const filterFields: FilterFieldDefinition[] = useMemo(() => [
    { id: 'status', label: t('security.deployments.status'), type: 'select', options: statusFilterOptions },
    { id: 'action', label: t('security.deployments.action'), type: 'select', options: actionFilterOptions },
  ], [t, statusFilterOptions, actionFilterOptions]);

  // Simple filters
  const simpleFilters: SimpleFilterField[] = useMemo(() => [
    { type: 'search', placeholder: 'security.deployments.search_placeholder' },
    { type: 'select', name: 'status', label: 'security.deployments.status', options: statusFilterOptions },
  ], [statusFilterOptions]);

  // Handle simple filter changes
  const handleSimpleFilterChange = useCallback((filters: Record<string, unknown>) => {
    const deployFilter: Record<string, unknown> = {};
    if (filters.search && typeof filters.search === 'string' && filters.search.trim()) {
      deployFilter.search = filters.search.trim();
    }
    if (filters.status && filters.status !== '') {
      deployFilter.status = filters.status;
    }
    setSimpleFilter(Object.keys(deployFilter).length > 0 ? deployFilter : null);
  }, []);

  // Load reference data
  const loadReferenceData = useCallback(async () => {
    try {
      const [profilesRes, agentsRes] = await Promise.all([
        request<{ securityProfiles: FirewallProfile[] }>(`query { securityProfiles(limit: 100) { id name } }`),
        request<{ agents: Agent[] }>(`query { agents { id name hostname } }`),
      ]);
      setProfiles(profilesRes.securityProfiles || []);
      setAgents(agentsRes.agents || []);
    } catch {
      // Ignore reference data errors
    }
  }, [request]);

  // Load deployments
  const loadDeployments = useCallback(async () => {
    setLoading(true);
    try {
      const listQuery = `
        query ListSecurityDeployments($limit: Int, $offset: Int, $filter: FirewallDeploymentFilter) {
          securityDeployments(limit: $limit, offset: $offset, filter: $filter) {
            id profileId profile { id name } agentId agentName action status statusMessage output startedAt completedAt createdAt
          }
          securityDeploymentsCount
        }
      `;

      const variables = {
        limit: pagination.rowsPerPage,
        offset: pagination.page * pagination.rowsPerPage,
        filter: simpleFilter,
      };

      const data = await request<{ securityDeployments: FirewallDeployment[]; securityDeploymentsCount: number }>(listQuery, variables);
      const deployList = data.securityDeployments || [];
      setDeployments(deployList);
      setTotalCount(data.securityDeploymentsCount || deployList.length);

      // Calculate stats
      setStats({
        total: data.securityDeploymentsCount || deployList.length,
        applied: deployList.filter(d => d.status === 'applied').length,
        pending: deployList.filter(d => d.status === 'pending' || d.status === 'deploying').length,
        error: deployList.filter(d => d.status === 'error').length,
      });

      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('security.deployments.error_loading'));
    } finally {
      setLoading(false);
    }
  }, [request, pagination.page, pagination.rowsPerPage, simpleFilter, t]);

  useEffect(() => {
    loadDeployments();
  }, [loadDeployments]);

  useEffect(() => {
    loadReferenceData();
  }, [loadReferenceData]);

  // Handlers
  const handleOpenDeploy = () => {
    setDeployError(null);
    setSelectedProfileId('');
    setSelectedAgentId('');
    setDryRun(false);
    setDeployOpen(true);
  };

  const handleDeploy = async () => {
    if (!selectedProfileId || !selectedAgentId) return;
    try {
      const result = await request<{ deploySecurityProfile: FirewallDeployment }>(
        `mutation DeploySecurityProfile($profileId: ID!, $agentId: ID!, $dryRun: Boolean) { deploySecurityProfile(profileId: $profileId, agentId: $agentId, dryRun: $dryRun) { id status statusMessage } }`,
        { profileId: selectedProfileId, agentId: selectedAgentId, dryRun }
      );
      if (dryRun && result.deploySecurityProfile) {
        // Show dry-run result
        setDeployError(null);
        setError(`${t('security.deployments.dry_run_success')}: ${result.deploySecurityProfile.statusMessage || t('security.deployments.config_valid')}`);
      }
      setDeployOpen(false);
      await loadDeployments();
    } catch (err) {
      setDeployError(err instanceof Error ? err.message : 'Deployment failed');
    }
  };

  const handleRollback = async () => {
    if (!selectedDeployment) return;
    try {
      await request(
        `mutation RollbackSecurityDeployment($id: ID!) { rollbackSecurityDeployment(id: $id) { id } }`,
        { id: selectedDeployment.id }
      );
      setRollbackOpen(false);
      setSelectedDeployment(null);
      await loadDeployments();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Rollback failed');
    }
  };

  const handleFlush = async () => {
    if (!selectedAgentId) return;
    try {
      await request(
        `mutation FlushSecurityRules($agentId: ID!) { flushSecurityRules(agentId: $agentId) { id } }`,
        { agentId: selectedAgentId }
      );
      setFlushOpen(false);
      await loadDeployments();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Flush failed');
    }
  };

  // Stat cards
  const statCards: StatCardData[] = [
    { title: t('clusters.total'), value: String(stats.total), icon: 'rocket', color: 'primary' },
    { title: t('security.deployments.status_applied'), value: String(stats.applied), icon: 'checkCircle', color: 'success' },
    { title: t('security.deployments.status_pending'), value: String(stats.pending), icon: 'schedule', color: 'warning' },
    { title: t('security.deployments.status_error'), value: String(stats.error), icon: 'error', color: 'error' },
  ];

  // DataGrid columns
  const columns = [
    {
      id: 'profile',
      label: 'security.deployments.profile',
      sortable: true,
      render: (deployment: FirewallDeployment) => (
        <CSDTypography variant="body2" fontWeight="bold">
          {deployment.profile?.name || 'N/A'}
        </CSDTypography>
      ),
    },
    {
      id: 'agentName',
      label: 'security.deployments.agent',
      sortable: true,
    },
    {
      id: 'action',
      label: 'security.deployments.action',
      sortable: true,
      render: (deployment: FirewallDeployment) => (
        <CSDChip
          label={t(`security.deployments.action_${deployment.action}`)}
          size="small"
          color={actionColors[deployment.action] || 'default'}
        />
      ),
    },
    {
      id: 'status',
      label: 'security.deployments.status',
      sortable: true,
      render: (deployment: FirewallDeployment) => (
        <CSDChip
          label={t(`security.deployments.status_${deployment.status}`)}
          size="small"
          color={statusColors[deployment.status] || 'default'}
        />
      ),
    },
    {
      id: 'statusMessage',
      label: 'security.deployments.status_message',
      render: (deployment: FirewallDeployment) => (
        <CSDTypography variant="body2" sx={{ maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {deployment.statusMessage || '-'}
        </CSDTypography>
      ),
    },
    {
      id: 'createdAt',
      label: 'clusters.created',
      sortable: true,
      render: (deployment: FirewallDeployment) => formatDate(deployment.createdAt, '-'),
    },
  ];

  // Row actions
  const actions = [
    {
      icon: 'undo',
      onClick: (deployment: FirewallDeployment) => {
        setSelectedDeployment(deployment);
        setRollbackOpen(true);
      },
      tooltip: 'security.deployments.rollback',
      color: 'warning' as const,
      disabled: (deployment: FirewallDeployment) => deployment.status !== 'applied',
    },
  ];

  // Deploy dialog
  const deployDialog = (
    <CSDFormDialog
      id="security-deploy-dialog"
      open={deployOpen}
      onClose={() => setDeployOpen(false)}
      title="security.deployments.deploy"
      icon="rocket"
      error={deployError}
      actions={
        <>
          <CSDActionButton variant="outlined" onClick={() => setDeployOpen(false)}>
            {t('common.cancel')}
          </CSDActionButton>
          <CSDActionButton
            variant="contained"
            onClick={handleDeploy}
            disabled={!selectedProfileId || !selectedAgentId}
          >
            {t('security.deployments.deploy')}
          </CSDActionButton>
        </>
      }
    >
      <CSDStack spacing={2}>
        <CSDAlert severity="info">
          {t('security.deployments.deploy_confirmation', { profile: 'selected', agent: 'selected' })}
        </CSDAlert>
        <CSDSelect
          id="deploy-profile"
          name="profileId"
          label={t('security.deployments.profile')}
          value={selectedProfileId}
          onChange={setSelectedProfileId}
          options={profileOptions}
          required
          fullWidth
        />
        <CSDSelect
          id="deploy-agent"
          name="agentId"
          label={t('security.deployments.agent')}
          value={selectedAgentId}
          onChange={setSelectedAgentId}
          options={agentOptions}
          required
          fullWidth
        />
        <CSDSwitch
          id="deploy-dry-run"
          name="dryRun"
          label={t('security.deployments.dry_run')}
          checked={dryRun}
          onChange={setDryRun}
        />
        {dryRun && (
          <CSDAlert severity="warning">
            {t('security.deployments.dry_run_warning')}
          </CSDAlert>
        )}
      </CSDStack>
    </CSDFormDialog>
  );

  // Rollback dialog
  const rollbackDialog = (
    <CSDConfirmDialog
      id="security-rollback-dialog"
      open={rollbackOpen}
      onClose={() => {
        setRollbackOpen(false);
        setSelectedDeployment(null);
      }}
      onConfirm={handleRollback}
      title={t('security.deployments.rollback')}
      message={t('security.deployments.rollback_confirmation', { agent: selectedDeployment?.agentName || '' })}
      confirmLabel={t('security.deployments.rollback')}
      cancelLabel={t('common.cancel')}
      severity="warning"
    />
  );

  // Flush dialog
  const flushDialog = (
    <CSDFormDialog
      id="security-flush-dialog"
      open={flushOpen}
      onClose={() => setFlushOpen(false)}
      title="security.deployments.flush"
      icon="warning"
      actions={
        <>
          <CSDActionButton variant="outlined" onClick={() => setFlushOpen(false)}>
            {t('common.cancel')}
          </CSDActionButton>
          <CSDActionButton
            variant="contained"
            color="error"
            onClick={handleFlush}
            disabled={!selectedAgentId}
          >
            {t('security.deployments.flush')}
          </CSDActionButton>
        </>
      }
    >
      <CSDStack spacing={2}>
        <CSDAlert severity="error">
          {t('security.deployments.flush_confirmation', { agent: 'selected' })}
        </CSDAlert>
        <CSDSelect
          id="flush-agent"
          name="agentId"
          label={t('security.deployments.agent')}
          value={selectedAgentId}
          onChange={setSelectedAgentId}
          options={agentOptions}
          required
          fullWidth
        />
      </CSDStack>
    </CSDFormDialog>
  );

  return (
    <AdvancedFilterManager
      storageKey="security-deployments"
      filterFields={filterFields}
      simpleFilters={simpleFilters}
      data={deployments}
      onSimpleFilterChange={handleSimpleFilterChange}
      pagination={pagination}
      autoSave={false}
      useServerFiltering={true}
      actions={
        <CSDBox sx={{ display: 'flex', gap: 1 }}>
          <CSDActionButton
            id="security-flush-button"
            variant="outlined"
            color="error"
            startIcon={<CSDIcon name="deleteForever" />}
            onClick={() => {
              setSelectedAgentId('');
              setFlushOpen(true);
            }}
          >
            {t('security.deployments.flush')}
          </CSDActionButton>
          <CSDActionButton
            id="security-deploy-button"
            variant="contained"
            startIcon={<CSDIcon name="rocket" />}
            onClick={handleOpenDeploy}
          >
            {t('security.deployments.deploy')}
          </CSDActionButton>
        </CSDBox>
      }
    >
      {({ filteredData, filterBar }) => (
        <CSDCrudPage
          entityName="security-deployments"
          statCards={statCards}
          filterBar={filterBar}
          loading={loading}
          error={error}
          formDialog={<>{deployDialog}{flushDialog}</>}
          deleteDialog={rollbackDialog}
        >
          <CSDDataGrid
            id="security-deployments-data-grid"
            data={filteredData}
            columns={columns}
            actions={actions}
            loading={loading}
            pagination={pagination}
            totalCount={totalCount}
            keyField="id"
            hoverable
          />
        </CSDCrudPage>
      )}
    </AdvancedFilterManager>
  );
};

export default DeploymentsPage;
