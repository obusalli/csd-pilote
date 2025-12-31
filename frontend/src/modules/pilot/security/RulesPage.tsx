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
  CSDBox,
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
import { useCloneEntity } from '../../../shared/hooks/useCloneEntity';
import { BREADCRUMBS } from '../../../shared/config/breadcrumbs';

// Types
interface FirewallRule {
  id: string;
  name: string;
  description: string;
  chain: string;
  protocol: string;
  sourceIp: string;
  sourcePort: string;
  destIp: string;
  destPort: string;
  action: string;
  priority: number;
  enabled: boolean;
  comment: string;
  ruleExpr: string;
  // Interface matching
  inInterface: string;
  outInterface: string;
  // Connection tracking
  ctState: string;
  // Rate limiting
  rateLimit: string;
  rateBurst: number;
  limitOver: string;
  // NAT options
  natToAddr: string;
  natToPort: string;
  // Logging options
  logPrefix: string;
  logLevel: string;
  createdAt: string;
  updatedAt: string;
}

interface RuleInput {
  name: string;
  description?: string;
  chain: string;
  protocol: string;
  sourceIp?: string;
  sourcePort?: string;
  destIp?: string;
  destPort?: string;
  action: string;
  priority?: number;
  enabled?: boolean;
  comment?: string;
  ruleExpr?: string;
  // Interface matching
  inInterface?: string;
  outInterface?: string;
  // Connection tracking
  ctState?: string;
  // Rate limiting
  rateLimit?: string;
  rateBurst?: number;
  limitOver?: string;
  // NAT options
  natToAddr?: string;
  natToPort?: string;
  // Logging options
  logPrefix?: string;
  logLevel?: string;
}

// Action colors
const actionColors: Record<string, 'success' | 'error' | 'warning' | 'default' | 'info'> = {
  accept: 'success',
  drop: 'error',
  reject: 'warning',
  log: 'info',
  masquerade: 'default',
  snat: 'default',
  dnat: 'default',
  redirect: 'default',
};

// Initial form state
const INITIAL_FORM: Partial<RuleInput> = {
  name: '',
  description: '',
  chain: 'input',
  protocol: 'tcp',
  sourceIp: '',
  sourcePort: '',
  destIp: '',
  destPort: '',
  action: 'accept',
  priority: 100,
  enabled: true,
  comment: '',
  ruleExpr: '',
  // Interface matching
  inInterface: '',
  outInterface: '',
  // Connection tracking
  ctState: '',
  // Rate limiting
  rateLimit: '',
  rateBurst: 0,
  limitOver: '',
  // NAT options
  natToAddr: '',
  natToPort: '',
  // Logging options
  logPrefix: '',
  logLevel: '',
};

export const RulesPage: React.FC = () => {
  const { request } = useGraphQL();
  const { t } = useTranslation();
  const pagination = usePagination('security-rules');

  // Breadcrumb
  useBreadcrumb([
    BREADCRUMBS.PILOTE,
    BREADCRUMBS.SECURITY,
    { labelKey: 'breadcrumb.rules', icon: 'list' },
  ]);

  // State
  const [rules, setRules] = useState<FirewallRule[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [totalCount, setTotalCount] = useState(0);

  // Filter state
  const [simpleFilter, setSimpleFilter] = useState<Record<string, unknown> | null>(null);

  // Dialog state
  const [formOpen, setFormOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [bulkDeleteOpen, setBulkDeleteOpen] = useState(false);
  const [editingRule, setEditingRule] = useState<FirewallRule | null>(null);

  // Selection state for bulk actions
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

  // Form state
  const [form, setForm] = useState<Partial<RuleInput>>(INITIAL_FORM);
  const [initialForm, setInitialForm] = useState<Partial<RuleInput>>(INITIAL_FORM);
  const [formError, setFormError] = useState<string | null>(null);

  // Stats
  const [stats, setStats] = useState({
    total: 0,
    enabled: 0,
    disabled: 0,
    accept: 0,
  });

  // Check if form has changes
  const hasChanges = useMemo(() => {
    return JSON.stringify(form) !== JSON.stringify(initialForm);
  }, [form, initialForm]);

  // Check if form is valid
  const isFormValid = useMemo(() => {
    return !!(form.name?.trim() && form.chain && form.protocol && form.action);
  }, [form]);

  // Chain options
  const chainOptions = useMemo(() => [
    { value: 'input', label: t('security.rules.chain_input') },
    { value: 'output', label: t('security.rules.chain_output') },
    { value: 'forward', label: t('security.rules.chain_forward') },
    { value: 'prerouting', label: t('security.rules.chain_prerouting') },
    { value: 'postrouting', label: t('security.rules.chain_postrouting') },
  ], [t]);

  // Protocol options
  const protocolOptions = useMemo(() => [
    { value: 'all', label: 'All' },
    { value: 'tcp', label: 'TCP' },
    { value: 'udp', label: 'UDP' },
    { value: 'icmp', label: 'ICMP' },
  ], []);

  // Action options
  const actionOptions = useMemo(() => [
    { value: 'accept', label: t('security.rules.action_accept') },
    { value: 'drop', label: t('security.rules.action_drop') },
    { value: 'reject', label: t('security.rules.action_reject') },
    { value: 'log', label: t('security.rules.action_log') },
    { value: 'masquerade', label: t('security.rules.action_masquerade') },
    { value: 'snat', label: 'SNAT' },
    { value: 'dnat', label: 'DNAT' },
    { value: 'redirect', label: 'REDIRECT' },
  ], [t]);

  // Connection tracking state options
  const ctStateOptions = useMemo(() => [
    { value: '', label: '-' },
    { value: 'new', label: 'NEW' },
    { value: 'established', label: 'ESTABLISHED' },
    { value: 'related', label: 'RELATED' },
    { value: 'invalid', label: 'INVALID' },
    { value: 'established,related', label: 'ESTABLISHED,RELATED' },
    { value: 'new,established,related', label: 'NEW,ESTABLISHED,RELATED' },
  ], []);

  // Log level options
  const logLevelOptions = useMemo(() => [
    { value: '', label: '-' },
    { value: 'emerg', label: 'Emergency' },
    { value: 'alert', label: 'Alert' },
    { value: 'crit', label: 'Critical' },
    { value: 'err', label: 'Error' },
    { value: 'warn', label: 'Warning' },
    { value: 'notice', label: 'Notice' },
    { value: 'info', label: 'Info' },
    { value: 'debug', label: 'Debug' },
  ], []);

  // Status options for filter
  const chainFilterOptions = useMemo(() => [
    { value: '', label: t('common.all') },
    ...chainOptions,
  ], [t, chainOptions]);

  const actionFilterOptions = useMemo(() => [
    { value: '', label: t('common.all') },
    ...actionOptions,
  ], [t, actionOptions]);

  // Filter fields
  const filterFields: FilterFieldDefinition[] = useMemo(() => [
    { id: 'name', label: t('security.rules.name'), type: 'string' },
    { id: 'chain', label: t('security.rules.chain'), type: 'select', options: chainFilterOptions },
    { id: 'action', label: t('security.rules.action'), type: 'select', options: actionFilterOptions },
  ], [t, chainFilterOptions, actionFilterOptions]);

  // Simple filters
  const simpleFilters: SimpleFilterField[] = useMemo(() => [
    { type: 'search', placeholder: 'security.rules.search_placeholder' },
    { type: 'select', name: 'chain', label: 'security.rules.chain', options: chainFilterOptions },
  ], [chainFilterOptions]);

  // Handle simple filter changes
  const handleSimpleFilterChange = useCallback((filters: Record<string, unknown>) => {
    const ruleFilter: Record<string, unknown> = {};
    if (filters.search && typeof filters.search === 'string' && filters.search.trim()) {
      ruleFilter.search = filters.search.trim();
    }
    if (filters.chain && filters.chain !== '') {
      ruleFilter.chain = filters.chain;
    }
    setSimpleFilter(Object.keys(ruleFilter).length > 0 ? ruleFilter : null);
  }, []);

  // Load rules
  const loadRules = useCallback(async () => {
    setLoading(true);
    try {
      const listQuery = `
        query ListSecurityRules($limit: Int, $offset: Int, $filter: FirewallRuleFilter) {
          securityRules(limit: $limit, offset: $offset, filter: $filter) {
            id name description chain protocol sourceIp sourcePort destIp destPort action priority enabled comment ruleExpr
            inInterface outInterface ctState rateLimit rateBurst limitOver natToAddr natToPort logPrefix logLevel
            createdAt updatedAt
          }
          securityRulesCount
        }
      `;

      const variables = {
        limit: pagination.rowsPerPage,
        offset: pagination.page * pagination.rowsPerPage,
        filter: simpleFilter,
      };

      const data = await request<{ securityRules: FirewallRule[]; securityRulesCount: number }>(listQuery, variables);
      const ruleList = data.securityRules || [];
      setRules(ruleList);
      setTotalCount(data.securityRulesCount || ruleList.length);

      // Calculate stats
      setStats({
        total: data.securityRulesCount || ruleList.length,
        enabled: ruleList.filter(r => r.enabled).length,
        disabled: ruleList.filter(r => !r.enabled).length,
        accept: ruleList.filter(r => r.action === 'accept').length,
      });

      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('security.rules.error_loading'));
    } finally {
      setLoading(false);
    }
  }, [request, pagination.page, pagination.rowsPerPage, simpleFilter, t]);

  useEffect(() => {
    loadRules();
  }, [loadRules]);

  // Handlers
  const handleOpen = (rule: FirewallRule | null) => {
    setFormError(null);
    if (rule) {
      setEditingRule(rule);
      const formState: Partial<RuleInput> = {
        name: rule.name || '',
        description: rule.description || '',
        chain: rule.chain || 'input',
        protocol: rule.protocol || 'tcp',
        sourceIp: rule.sourceIp || '',
        sourcePort: rule.sourcePort || '',
        destIp: rule.destIp || '',
        destPort: rule.destPort || '',
        action: rule.action || 'accept',
        priority: rule.priority || 100,
        enabled: rule.enabled,
        comment: rule.comment || '',
        ruleExpr: rule.ruleExpr || '',
        // Interface matching
        inInterface: rule.inInterface || '',
        outInterface: rule.outInterface || '',
        // Connection tracking
        ctState: rule.ctState || '',
        // Rate limiting
        rateLimit: rule.rateLimit || '',
        rateBurst: rule.rateBurst || 0,
        limitOver: rule.limitOver || '',
        // NAT options
        natToAddr: rule.natToAddr || '',
        natToPort: rule.natToPort || '',
        // Logging options
        logPrefix: rule.logPrefix || '',
        logLevel: rule.logLevel || '',
      };
      setForm(formState);
      setInitialForm(formState);
    } else {
      setEditingRule(null);
      setForm(INITIAL_FORM);
      setInitialForm(INITIAL_FORM);
    }
    setFormOpen(true);
  };

  const handleClose = () => {
    setFormOpen(false);
    setEditingRule(null);
    setFormError(null);
  };

  const handleDelete = (rule: FirewallRule) => {
    setEditingRule(rule);
    setDeleteOpen(true);
  };

  const handleFormSubmit = async () => {
    try {
      if (editingRule) {
        const mutation = `
          mutation UpdateSecurityRule($id: ID!, $input: UpdateFirewallRuleInput!) {
            updateSecurityRule(id: $id, input: $input) { id }
          }
        `;
        await request(mutation, { id: editingRule.id, input: form });
      } else {
        const mutation = `
          mutation CreateSecurityRule($input: CreateFirewallRuleInput!) {
            createSecurityRule(input: $input) { id }
          }
        `;
        await request(mutation, { input: form });
      }
      handleClose();
      await loadRules();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : t('security.rules.error_saving'));
    }
  };

  const handleConfirmDelete = async () => {
    if (!editingRule) return;
    try {
      await request(`mutation DeleteSecurityRule($id: ID!) { deleteSecurityRule(id: $id) }`, { id: editingRule.id });
      setDeleteOpen(false);
      setEditingRule(null);
      await loadRules();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('security.rules.error_deleting'));
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
        `mutation BulkDeleteSecurityRules($ids: [ID!]!) { bulkDeleteSecurityRules(ids: $ids) }`,
        { ids: Array.from(selectedIds) }
      );
      setBulkDeleteOpen(false);
      setSelectedIds(new Set());
      await loadRules();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('security.rules.error_deleting'));
    }
  };

  // Clone entity hook - opens form with cloned data
  const { cloneEntity } = useCloneEntity<FirewallRule>({
    nameField: 'name',
    additionalFields: [
      'description', 'chain', 'protocol', 'sourceIp', 'sourcePort',
      'destIp', 'destPort', 'action', 'priority', 'enabled', 'comment',
      'ruleExpr', 'inInterface', 'outInterface', 'ctState', 'rateLimit',
      'rateBurst', 'limitOver', 'natToAddr', 'natToPort', 'logPrefix', 'logLevel',
    ],
    onClone: (clonedData) => {
      setEditingRule(null); // Ensure create mode
      setForm(clonedData as Partial<RuleInput>);
      setInitialForm(clonedData as Partial<RuleInput>);
      setFormOpen(true);
    },
  });

  // Stat cards
  const statCards: StatCardData[] = [
    { title: t('clusters.total'), value: String(stats.total), icon: 'list', color: 'primary' },
    { title: t('security.rules.enabled'), value: String(stats.enabled), icon: 'checkCircle', color: 'success' },
    { title: t('security.rules.action_accept'), value: String(stats.accept), icon: 'shield', color: 'info' },
  ];

  // DataGrid columns
  const columns = [
    {
      id: 'name',
      label: 'security.rules.name',
      sortable: true,
      render: (rule: FirewallRule) => (
        <CSDTypography variant="body2" fontWeight="bold">{rule.name}</CSDTypography>
      ),
    },
    {
      id: 'chain',
      label: 'security.rules.chain',
      sortable: true,
      render: (rule: FirewallRule) => (
        <CSDChip label={rule.chain.toUpperCase()} size="small" variant="outlined" />
      ),
    },
    {
      id: 'protocol',
      label: 'security.rules.protocol',
      sortable: true,
      render: (rule: FirewallRule) => rule.protocol.toUpperCase(),
    },
    {
      id: 'destPort',
      label: 'security.rules.dest_port',
      sortable: true,
      render: (rule: FirewallRule) => rule.destPort || '-',
    },
    {
      id: 'action',
      label: 'security.rules.action',
      sortable: true,
      render: (rule: FirewallRule) => (
        <CSDChip label={rule.action.toUpperCase()} size="small" color={actionColors[rule.action] || 'default'} />
      ),
    },
    {
      id: 'enabled',
      label: 'security.rules.enabled',
      sortable: true,
      render: (rule: FirewallRule) => (
        <CSDIcon name={rule.enabled ? 'checkCircle' : 'cancel'} color={rule.enabled ? 'success' : 'disabled'} />
      ),
    },
    {
      id: 'priority',
      label: 'security.rules.priority',
      sortable: true,
    },
  ];

  // Row actions
  const actions = [
    { icon: 'edit', onClick: (rule: FirewallRule) => handleOpen(rule), tooltip: 'common.edit', color: 'primary' as const },
    { icon: 'copy', onClick: (rule: FirewallRule) => cloneEntity(rule), tooltip: 'common.clone', color: 'secondary' as const },
    { icon: 'delete', onClick: (rule: FirewallRule) => handleDelete(rule), tooltip: 'common.delete', color: 'error' as const },
  ];

  // Form dialog
  const formDialog = (
    <CSDFormDialog
      id="security-rules-form-dialog"
      open={formOpen}
      onClose={handleClose}
      title={editingRule ? 'security.rules.edit_rule' : 'security.rules.add_rule'}
      icon={editingRule ? 'edit' : 'add'}
      error={formError}
      trackChanges={!!editingRule}
      hasChanges={hasChanges}
      onResetAll={() => setForm(initialForm)}
      actions={
        <>
          <CSDActionButton variant="outlined" onClick={handleClose}>
            {t('common.cancel')}
          </CSDActionButton>
          <CSDActionButton
            variant="contained"
            onClick={handleFormSubmit}
            disabled={!isFormValid || (editingRule && !hasChanges)}
          >
            {t('common.save')}
          </CSDActionButton>
        </>
      }
    >
      <CSDStack spacing={2}>
        <CSDTextField
          id="security-rules-form-name"
          name="name"
          label={t('security.rules.name')}
          value={form.name || ''}
          onChange={(v) => setForm({ ...form, name: v })}
          required
          fullWidth
        />
        <CSDTextField
          id="security-rules-form-description"
          name="description"
          label={t('security.rules.description')}
          value={form.description || ''}
          onChange={(v) => setForm({ ...form, description: v })}
          multiline
          rows={2}
          fullWidth
        />
        <CSDStack direction="row" spacing={2}>
          <CSDSelect
            id="security-rules-form-chain"
            name="chain"
            label={t('security.rules.chain')}
            value={form.chain || 'input'}
            onChange={(v) => setForm({ ...form, chain: v })}
            options={chainOptions}
            required
            fullWidth
          />
          <CSDSelect
            id="security-rules-form-protocol"
            name="protocol"
            label={t('security.rules.protocol')}
            value={form.protocol || 'tcp'}
            onChange={(v) => setForm({ ...form, protocol: v })}
            options={protocolOptions}
            required
            fullWidth
          />
        </CSDStack>
        <CSDStack direction="row" spacing={2}>
          <CSDTextField
            id="security-rules-form-source-ip"
            name="sourceIp"
            label={t('security.rules.source_ip')}
            value={form.sourceIp || ''}
            onChange={(v) => setForm({ ...form, sourceIp: v })}
            fullWidth
          />
          <CSDTextField
            id="security-rules-form-source-port"
            name="sourcePort"
            label={t('security.rules.source_port')}
            value={form.sourcePort || ''}
            onChange={(v) => setForm({ ...form, sourcePort: v })}
            fullWidth
          />
        </CSDStack>
        <CSDStack direction="row" spacing={2}>
          <CSDTextField
            id="security-rules-form-dest-ip"
            name="destIp"
            label={t('security.rules.dest_ip')}
            value={form.destIp || ''}
            onChange={(v) => setForm({ ...form, destIp: v })}
            fullWidth
          />
          <CSDTextField
            id="security-rules-form-dest-port"
            name="destPort"
            label={t('security.rules.dest_port')}
            value={form.destPort || ''}
            onChange={(v) => setForm({ ...form, destPort: v })}
            fullWidth
          />
        </CSDStack>
        <CSDStack direction="row" spacing={2}>
          <CSDSelect
            id="security-rules-form-action"
            name="action"
            label={t('security.rules.action')}
            value={form.action || 'accept'}
            onChange={(v) => setForm({ ...form, action: v })}
            options={actionOptions}
            required
            fullWidth
          />
          <CSDTextField
            id="security-rules-form-priority"
            name="priority"
            label={t('security.rules.priority')}
            value={String(form.priority || 100)}
            onChange={(v) => setForm({ ...form, priority: parseInt(v) || 100 })}
            type="number"
            fullWidth
          />
        </CSDStack>

        {/* Interface matching */}
        <CSDTypography variant="subtitle2" color="textSecondary">
          {t('security.rules.interface_matching')}
        </CSDTypography>
        <CSDStack direction="row" spacing={2}>
          <CSDTextField
            id="security-rules-form-in-interface"
            name="inInterface"
            label={t('security.rules.in_interface')}
            value={form.inInterface || ''}
            onChange={(v) => setForm({ ...form, inInterface: v })}
            fullWidth
            helperText="e.g., eth0, lo"
          />
          <CSDTextField
            id="security-rules-form-out-interface"
            name="outInterface"
            label={t('security.rules.out_interface')}
            value={form.outInterface || ''}
            onChange={(v) => setForm({ ...form, outInterface: v })}
            fullWidth
            helperText="e.g., eth1, br0"
          />
        </CSDStack>

        {/* Connection tracking */}
        <CSDSelect
          id="security-rules-form-ct-state"
          name="ctState"
          label={t('security.rules.ct_state')}
          value={form.ctState || ''}
          onChange={(v) => setForm({ ...form, ctState: v })}
          options={ctStateOptions}
          fullWidth
        />

        {/* Rate limiting */}
        <CSDTypography variant="subtitle2" color="textSecondary">
          {t('security.rules.rate_limiting')}
        </CSDTypography>
        <CSDStack direction="row" spacing={2}>
          <CSDTextField
            id="security-rules-form-rate-limit"
            name="rateLimit"
            label={t('security.rules.rate_limit')}
            value={form.rateLimit || ''}
            onChange={(v) => setForm({ ...form, rateLimit: v })}
            fullWidth
            helperText="e.g., 10/second, 100/minute"
          />
          <CSDTextField
            id="security-rules-form-rate-burst"
            name="rateBurst"
            label={t('security.rules.rate_burst')}
            value={String(form.rateBurst || '')}
            onChange={(v) => setForm({ ...form, rateBurst: parseInt(v) || 0 })}
            type="number"
            fullWidth
          />
        </CSDStack>

        {/* NAT options - only for NAT actions */}
        {['snat', 'dnat', 'redirect'].includes(form.action || '') && (
          <>
            <CSDTypography variant="subtitle2" color="textSecondary">
              {t('security.rules.nat_options')}
            </CSDTypography>
            <CSDStack direction="row" spacing={2}>
              <CSDTextField
                id="security-rules-form-nat-to-addr"
                name="natToAddr"
                label={t('security.rules.nat_to_addr')}
                value={form.natToAddr || ''}
                onChange={(v) => setForm({ ...form, natToAddr: v })}
                fullWidth
                helperText="Target address for SNAT/DNAT"
              />
              <CSDTextField
                id="security-rules-form-nat-to-port"
                name="natToPort"
                label={t('security.rules.nat_to_port')}
                value={form.natToPort || ''}
                onChange={(v) => setForm({ ...form, natToPort: v })}
                fullWidth
                helperText="Target port for DNAT/REDIRECT"
              />
            </CSDStack>
          </>
        )}

        {/* Logging options - only for LOG action */}
        {form.action === 'log' && (
          <>
            <CSDTypography variant="subtitle2" color="textSecondary">
              {t('security.rules.log_options')}
            </CSDTypography>
            <CSDStack direction="row" spacing={2}>
              <CSDTextField
                id="security-rules-form-log-prefix"
                name="logPrefix"
                label={t('security.rules.log_prefix')}
                value={form.logPrefix || ''}
                onChange={(v) => setForm({ ...form, logPrefix: v })}
                fullWidth
                helperText="Prefix for log messages"
              />
              <CSDSelect
                id="security-rules-form-log-level"
                name="logLevel"
                label={t('security.rules.log_level')}
                value={form.logLevel || ''}
                onChange={(v) => setForm({ ...form, logLevel: v })}
                options={logLevelOptions}
                fullWidth
              />
            </CSDStack>
          </>
        )}

        <CSDTextField
          id="security-rules-form-rule-expr"
          name="ruleExpr"
          label={t('security.rules.rule_expr')}
          value={form.ruleExpr || ''}
          onChange={(v) => setForm({ ...form, ruleExpr: v })}
          fullWidth
          helperText="Optional raw nftables expression"
        />
        <CSDTextField
          id="security-rules-form-comment"
          name="comment"
          label={t('security.rules.comment')}
          value={form.comment || ''}
          onChange={(v) => setForm({ ...form, comment: v })}
          fullWidth
        />
        <CSDSwitch
          id="security-rules-form-enabled"
          name="enabled"
          label={t('security.rules.enabled')}
          checked={form.enabled ?? true}
          onChange={(checked) => setForm({ ...form, enabled: checked })}
        />
      </CSDStack>
    </CSDFormDialog>
  );

  // Delete dialog
  const deleteDialog = (
    <>
      <CSDConfirmDialog
        id="security-rules-delete-dialog"
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        onConfirm={handleConfirmDelete}
        title={t('security.rules.delete_rule')}
        message={t('security.rules.delete_confirmation', { name: editingRule?.name || '' })}
        confirmLabel={t('common.delete')}
        cancelLabel={t('common.cancel')}
        severity="error"
      />
      <CSDConfirmDialog
        id="security-rules-bulk-delete-dialog"
        open={bulkDeleteOpen}
        onClose={() => setBulkDeleteOpen(false)}
        onConfirm={handleBulkDelete}
        title={t('security.rules.bulk_delete')}
        message={t('security.rules.bulk_delete_confirmation', { count: selectedIds.size })}
        confirmLabel={t('common.delete')}
        cancelLabel={t('common.cancel')}
        severity="error"
      />
    </>
  );

  return (
    <AdvancedFilterManager
      storageKey="security-rules"
      filterFields={filterFields}
      simpleFilters={simpleFilters}
      data={rules}
      onSimpleFilterChange={handleSimpleFilterChange}
      pagination={pagination}
      autoSave={false}
      useServerFiltering={true}
      actions={
        <CSDBox sx={{ display: 'flex', gap: 1 }}>
          {selectedIds.size > 0 && (
            <CSDActionButton
              id="security-rules-bulk-delete-button"
              variant="outlined"
              color="error"
              startIcon={<CSDIcon name="delete" />}
              onClick={() => setBulkDeleteOpen(true)}
            >
              {t('common.delete_selected', { count: selectedIds.size })}
            </CSDActionButton>
          )}
          <CSDActionButton
            id="security-rules-add-button"
            variant="contained"
            startIcon={<CSDIcon name="add" />}
            onClick={() => handleOpen(null)}
          >
            {t('security.rules.add_rule')}
          </CSDActionButton>
        </CSDBox>
      }
    >
      {({ filteredData, filterBar }) => (
        <CSDCrudPage
          entityName="security-rules"
          statCards={statCards}
          filterBar={filterBar}
          loading={loading}
          error={error}
          formDialog={formDialog}
          deleteDialog={deleteDialog}
        >
          <CSDDataGrid
            id="security-rules-data-grid"
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

export default RulesPage;
