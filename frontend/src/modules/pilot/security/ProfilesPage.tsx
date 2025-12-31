import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
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
  CSDAlert,
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
interface FirewallProfile {
  id: string;
  name: string;
  description: string;
  isDefault: boolean;
  enabled: boolean;
  // Default policies
  inputPolicy: string;
  outputPolicy: string;
  forwardPolicy: string;
  // Features
  enableNat: boolean;
  enableConntrack: boolean;
  allowLoopback: boolean;
  allowEstablished: boolean;
  allowIcmpPing: boolean;
  enableIpv6: boolean;
  rules: { id: string; name: string }[];
  createdAt: string;
  updatedAt: string;
}

interface ProfileInput {
  name: string;
  description?: string;
  isDefault?: boolean;
  enabled?: boolean;
  // Default policies
  inputPolicy?: string;
  outputPolicy?: string;
  forwardPolicy?: string;
  // Features
  enableNat?: boolean;
  enableConntrack?: boolean;
  allowLoopback?: boolean;
  allowEstablished?: boolean;
  allowIcmpPing?: boolean;
  enableIpv6?: boolean;
}

// Initial form state
const INITIAL_FORM: Partial<ProfileInput> = {
  name: '',
  description: '',
  isDefault: false,
  enabled: true,
  // Default policies
  inputPolicy: 'drop',
  outputPolicy: 'accept',
  forwardPolicy: 'drop',
  // Features
  enableNat: false,
  enableConntrack: true,
  allowLoopback: true,
  allowEstablished: true,
  allowIcmpPing: true,
  enableIpv6: false,
};

export const ProfilesPage: React.FC = () => {
  const { request } = useGraphQL();
  const { t } = useTranslation();
  const pagination = usePagination('security-profiles');

  // Breadcrumb
  useBreadcrumb([
    BREADCRUMBS.PILOTE,
    BREADCRUMBS.SECURITY,
    { labelKey: 'breadcrumb.profiles', icon: 'folder' },
  ]);

  // State
  const [profiles, setProfiles] = useState<FirewallProfile[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [totalCount, setTotalCount] = useState(0);

  // Filter state
  const [simpleFilter, setSimpleFilter] = useState<Record<string, unknown> | null>(null);

  // Dialog state
  const [formOpen, setFormOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [editingProfile, setEditingProfile] = useState<FirewallProfile | null>(null);

  // Form state
  const [form, setForm] = useState<Partial<ProfileInput>>(INITIAL_FORM);
  const [initialForm, setInitialForm] = useState<Partial<ProfileInput>>(INITIAL_FORM);
  const [formError, setFormError] = useState<string | null>(null);

  // Stats
  const [stats, setStats] = useState({
    total: 0,
    enabled: 0,
    default: 0,
  });

  // Import dialog state
  const [importOpen, setImportOpen] = useState(false);
  const [importData, setImportData] = useState<string>('');
  const [importError, setImportError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Check if form has changes
  const hasChanges = useMemo(() => {
    return JSON.stringify(form) !== JSON.stringify(initialForm);
  }, [form, initialForm]);

  // Check if form is valid
  const isFormValid = useMemo(() => {
    return !!form.name?.trim();
  }, [form]);

  // Filter fields
  const filterFields: FilterFieldDefinition[] = useMemo(() => [
    { id: 'name', label: t('security.profiles.name'), type: 'string' },
  ], [t]);

  // Simple filters
  const simpleFilters: SimpleFilterField[] = useMemo(() => [
    { type: 'search', placeholder: 'security.profiles.search_placeholder' },
  ], []);

  // Policy options
  const policyOptions = useMemo(() => [
    { value: 'accept', label: t('security.profiles.policy_accept') },
    { value: 'drop', label: t('security.profiles.policy_drop') },
  ], [t]);

  // Handle simple filter changes
  const handleSimpleFilterChange = useCallback((filters: Record<string, unknown>) => {
    const profileFilter: Record<string, unknown> = {};
    if (filters.search && typeof filters.search === 'string' && filters.search.trim()) {
      profileFilter.search = filters.search.trim();
    }
    setSimpleFilter(Object.keys(profileFilter).length > 0 ? profileFilter : null);
  }, []);

  // Load profiles
  const loadProfiles = useCallback(async () => {
    setLoading(true);
    try {
      const listQuery = `
        query ListSecurityProfiles($limit: Int, $offset: Int, $filter: FirewallProfileFilter) {
          securityProfiles(limit: $limit, offset: $offset, filter: $filter) {
            id name description isDefault enabled
            inputPolicy outputPolicy forwardPolicy
            enableNat enableConntrack allowLoopback allowEstablished allowIcmpPing enableIpv6
            rules { id name } createdAt updatedAt
          }
          securityProfilesCount
        }
      `;

      const variables = {
        limit: pagination.rowsPerPage,
        offset: pagination.page * pagination.rowsPerPage,
        filter: simpleFilter,
      };

      const data = await request<{ securityProfiles: FirewallProfile[]; securityProfilesCount: number }>(listQuery, variables);
      const profileList = data.securityProfiles || [];
      setProfiles(profileList);
      setTotalCount(data.securityProfilesCount || profileList.length);

      // Calculate stats
      setStats({
        total: data.securityProfilesCount || profileList.length,
        enabled: profileList.filter(p => p.enabled).length,
        default: profileList.filter(p => p.isDefault).length,
      });

      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('security.profiles.error_loading'));
    } finally {
      setLoading(false);
    }
  }, [request, pagination.page, pagination.rowsPerPage, simpleFilter, t]);

  useEffect(() => {
    loadProfiles();
  }, [loadProfiles]);

  // Handlers
  const handleOpen = (profile: FirewallProfile | null) => {
    setFormError(null);
    if (profile) {
      setEditingProfile(profile);
      const formState: Partial<ProfileInput> = {
        name: profile.name || '',
        description: profile.description || '',
        isDefault: profile.isDefault,
        enabled: profile.enabled,
        // Default policies
        inputPolicy: profile.inputPolicy || 'drop',
        outputPolicy: profile.outputPolicy || 'accept',
        forwardPolicy: profile.forwardPolicy || 'drop',
        // Features
        enableNat: profile.enableNat ?? false,
        enableConntrack: profile.enableConntrack ?? true,
        allowLoopback: profile.allowLoopback ?? true,
        allowEstablished: profile.allowEstablished ?? true,
        allowIcmpPing: profile.allowIcmpPing ?? true,
        enableIpv6: profile.enableIpv6 ?? false,
      };
      setForm(formState);
      setInitialForm(formState);
    } else {
      setEditingProfile(null);
      setForm(INITIAL_FORM);
      setInitialForm(INITIAL_FORM);
    }
    setFormOpen(true);
  };

  const handleClose = () => {
    setFormOpen(false);
    setEditingProfile(null);
    setFormError(null);
  };

  const handleDelete = (profile: FirewallProfile) => {
    setEditingProfile(profile);
    setDeleteOpen(true);
  };

  const handleFormSubmit = async () => {
    try {
      if (editingProfile) {
        const mutation = `
          mutation UpdateSecurityProfile($id: ID!, $input: UpdateFirewallProfileInput!) {
            updateSecurityProfile(id: $id, input: $input) { id }
          }
        `;
        await request(mutation, { id: editingProfile.id, input: form });
      } else {
        const mutation = `
          mutation CreateSecurityProfile($input: CreateFirewallProfileInput!) {
            createSecurityProfile(input: $input) { id }
          }
        `;
        await request(mutation, { input: form });
      }
      handleClose();
      await loadProfiles();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : t('security.profiles.error_saving'));
    }
  };

  const handleConfirmDelete = async () => {
    if (!editingProfile) return;
    try {
      await request(`mutation DeleteSecurityProfile($id: ID!) { deleteSecurityProfile(id: $id) }`, { id: editingProfile.id });
      setDeleteOpen(false);
      setEditingProfile(null);
      await loadProfiles();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('security.profiles.error_deleting'));
    }
  };

  // Export profile
  const handleExport = async (profile: FirewallProfile) => {
    try {
      const data = await request<{ exportSecurityProfile: string }>(
        `query ExportSecurityProfile($id: ID!) { exportSecurityProfile(id: $id) }`,
        { id: profile.id }
      );
      // Download as JSON file
      const blob = new Blob([data.exportSecurityProfile], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `firewall-profile-${profile.name.replace(/\s+/g, '-').toLowerCase()}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('security.profiles.error_exporting'));
    }
  };

  // Import profile
  const handleImport = async () => {
    if (!importData.trim()) return;
    try {
      const parsed = JSON.parse(importData);
      await request(
        `mutation ImportSecurityProfile($input: ProfileImportInput!) { importSecurityProfile(input: $input) { id } }`,
        { input: parsed }
      );
      setImportOpen(false);
      setImportData('');
      setImportError(null);
      await loadProfiles();
    } catch (err) {
      if (err instanceof SyntaxError) {
        setImportError(t('security.profiles.invalid_json'));
      } else {
        setImportError(err instanceof Error ? err.message : t('security.profiles.error_importing'));
      }
    }
  };

  // Handle file upload for import
  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (e) => {
      const content = e.target?.result as string;
      setImportData(content);
      setImportError(null);
    };
    reader.onerror = () => {
      setImportError(t('security.profiles.file_read_error'));
    };
    reader.readAsText(file);
  };

  // Clone entity hook - opens form with cloned data
  const { cloneEntity } = useCloneEntity<FirewallProfile>({
    nameField: 'name',
    additionalFields: [
      'description', 'enabled', 'inputPolicy', 'outputPolicy', 'forwardPolicy',
      'enableNat', 'enableConntrack', 'allowLoopback', 'allowEstablished',
      'allowIcmpPing', 'enableIpv6',
    ],
    onClone: (clonedData) => {
      setEditingProfile(null); // Ensure create mode
      setForm(clonedData as Partial<ProfileInput>);
      setInitialForm(clonedData as Partial<ProfileInput>);
      setFormOpen(true);
    },
  });

  // Stat cards
  const statCards: StatCardData[] = [
    { title: t('clusters.total'), value: String(stats.total), icon: 'folder', color: 'primary' },
    { title: t('security.profiles.enabled'), value: String(stats.enabled), icon: 'checkCircle', color: 'success' },
    { title: t('security.profiles.is_default'), value: String(stats.default), icon: 'star', color: 'warning' },
  ];

  // DataGrid columns
  const columns = [
    {
      id: 'name',
      label: 'security.profiles.name',
      sortable: true,
      render: (profile: FirewallProfile) => (
        <CSDStack direction="row" spacing={1} alignItems="center">
          <CSDTypography variant="body2" fontWeight="bold">{profile.name}</CSDTypography>
          {profile.isDefault && <CSDChip label="Default" size="small" color="warning" />}
        </CSDStack>
      ),
    },
    {
      id: 'description',
      label: 'security.profiles.description',
      sortable: true,
      render: (profile: FirewallProfile) => profile.description || '-',
    },
    {
      id: 'rules',
      label: 'security.profiles.rules',
      render: (profile: FirewallProfile) => (
        <CSDChip label={t('security.profiles.rules_count', { count: profile.rules?.length || 0 })} size="small" variant="outlined" />
      ),
    },
    {
      id: 'enabled',
      label: 'security.profiles.enabled',
      sortable: true,
      render: (profile: FirewallProfile) => (
        <CSDIcon name={profile.enabled ? 'checkCircle' : 'cancel'} color={profile.enabled ? 'success' : 'disabled'} />
      ),
    },
    {
      id: 'createdAt',
      label: 'clusters.created',
      sortable: true,
      render: (profile: FirewallProfile) => formatDate(profile.createdAt, '-'),
    },
  ];

  // Row actions
  const actions = [
    { icon: 'visibility', onClick: (profile: FirewallProfile) => window.location.href = `/pilote/security/profiles/${profile.id}`, tooltip: 'common.view', color: 'primary' as const },
    { icon: 'download', onClick: (profile: FirewallProfile) => handleExport(profile), tooltip: 'security.profiles.export', color: 'info' as const },
    { icon: 'copy', onClick: (profile: FirewallProfile) => cloneEntity(profile), tooltip: 'common.clone', color: 'secondary' as const },
    { icon: 'edit', onClick: (profile: FirewallProfile) => handleOpen(profile), tooltip: 'common.edit', color: 'primary' as const },
    { icon: 'delete', onClick: (profile: FirewallProfile) => handleDelete(profile), tooltip: 'common.delete', color: 'error' as const },
  ];

  // Form dialog
  const formDialog = (
    <CSDFormDialog
      id="security-profiles-form-dialog"
      open={formOpen}
      onClose={handleClose}
      title={editingProfile ? 'security.profiles.edit_profile' : 'security.profiles.add_profile'}
      icon={editingProfile ? 'edit' : 'add'}
      error={formError}
      trackChanges={!!editingProfile}
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
            disabled={!isFormValid || (editingProfile && !hasChanges)}
          >
            {t('common.save')}
          </CSDActionButton>
        </>
      }
    >
      <CSDStack spacing={2}>
        <CSDTextField
          id="security-profiles-form-name"
          name="name"
          label={t('security.profiles.name')}
          value={form.name || ''}
          onChange={(v) => setForm({ ...form, name: v })}
          required
          fullWidth
        />
        <CSDTextField
          id="security-profiles-form-description"
          name="description"
          label={t('security.profiles.description')}
          value={form.description || ''}
          onChange={(v) => setForm({ ...form, description: v })}
          multiline
          rows={3}
          fullWidth
        />
        <CSDSwitch
          id="security-profiles-form-default"
          name="isDefault"
          label={t('security.profiles.is_default')}
          checked={form.isDefault ?? false}
          onChange={(checked) => setForm({ ...form, isDefault: checked })}
        />
        <CSDSwitch
          id="security-profiles-form-enabled"
          name="enabled"
          label={t('security.profiles.enabled')}
          checked={form.enabled ?? true}
          onChange={(checked) => setForm({ ...form, enabled: checked })}
        />

        {/* Default Policies */}
        <CSDTypography variant="subtitle2" color="textSecondary">
          {t('security.profiles.default_policies')}
        </CSDTypography>
        <CSDStack direction="row" spacing={2}>
          <CSDSelect
            id="security-profiles-form-input-policy"
            name="inputPolicy"
            label={t('security.profiles.input_policy')}
            value={form.inputPolicy || 'drop'}
            onChange={(v) => setForm({ ...form, inputPolicy: v })}
            options={policyOptions}
            fullWidth
          />
          <CSDSelect
            id="security-profiles-form-output-policy"
            name="outputPolicy"
            label={t('security.profiles.output_policy')}
            value={form.outputPolicy || 'accept'}
            onChange={(v) => setForm({ ...form, outputPolicy: v })}
            options={policyOptions}
            fullWidth
          />
          <CSDSelect
            id="security-profiles-form-forward-policy"
            name="forwardPolicy"
            label={t('security.profiles.forward_policy')}
            value={form.forwardPolicy || 'drop'}
            onChange={(v) => setForm({ ...form, forwardPolicy: v })}
            options={policyOptions}
            fullWidth
          />
        </CSDStack>

        {/* Features */}
        <CSDTypography variant="subtitle2" color="textSecondary">
          {t('security.profiles.features')}
        </CSDTypography>
        <CSDStack direction="row" spacing={2} flexWrap="wrap">
          <CSDSwitch
            id="security-profiles-form-allow-loopback"
            name="allowLoopback"
            label={t('security.profiles.allow_loopback')}
            checked={form.allowLoopback ?? true}
            onChange={(checked) => setForm({ ...form, allowLoopback: checked })}
          />
          <CSDSwitch
            id="security-profiles-form-allow-established"
            name="allowEstablished"
            label={t('security.profiles.allow_established')}
            checked={form.allowEstablished ?? true}
            onChange={(checked) => setForm({ ...form, allowEstablished: checked })}
          />
          <CSDSwitch
            id="security-profiles-form-allow-icmp-ping"
            name="allowIcmpPing"
            label={t('security.profiles.allow_icmp_ping')}
            checked={form.allowIcmpPing ?? true}
            onChange={(checked) => setForm({ ...form, allowIcmpPing: checked })}
          />
        </CSDStack>
        <CSDStack direction="row" spacing={2}>
          <CSDSwitch
            id="security-profiles-form-enable-conntrack"
            name="enableConntrack"
            label={t('security.profiles.enable_conntrack')}
            checked={form.enableConntrack ?? true}
            onChange={(checked) => setForm({ ...form, enableConntrack: checked })}
          />
          <CSDSwitch
            id="security-profiles-form-enable-nat"
            name="enableNat"
            label={t('security.profiles.enable_nat')}
            checked={form.enableNat ?? false}
            onChange={(checked) => setForm({ ...form, enableNat: checked })}
          />
          <CSDSwitch
            id="security-profiles-form-enable-ipv6"
            name="enableIpv6"
            label={t('security.profiles.enable_ipv6')}
            checked={form.enableIpv6 ?? false}
            onChange={(checked) => setForm({ ...form, enableIpv6: checked })}
          />
        </CSDStack>
      </CSDStack>
    </CSDFormDialog>
  );

  // Delete dialog
  const deleteDialog = (
    <CSDConfirmDialog
      id="security-profiles-delete-dialog"
      open={deleteOpen}
      onClose={() => setDeleteOpen(false)}
      onConfirm={handleConfirmDelete}
      title={t('security.profiles.delete_profile')}
      message={t('security.profiles.delete_confirmation', { name: editingProfile?.name || '' })}
      confirmLabel={t('common.delete')}
      cancelLabel={t('common.cancel')}
      severity="error"
    />
  );

  // Import dialog
  const importDialog = (
    <CSDFormDialog
      id="security-profiles-import-dialog"
      open={importOpen}
      onClose={() => {
        setImportOpen(false);
        setImportData('');
        setImportError(null);
      }}
      title="security.profiles.import_profile"
      icon="upload"
      error={importError}
      actions={
        <>
          <CSDActionButton variant="outlined" onClick={() => setImportOpen(false)}>
            {t('common.cancel')}
          </CSDActionButton>
          <CSDActionButton
            variant="contained"
            onClick={handleImport}
            disabled={!importData.trim()}
          >
            {t('security.profiles.import')}
          </CSDActionButton>
        </>
      }
    >
      <CSDStack spacing={2}>
        <CSDAlert severity="info">
          {t('security.profiles.import_info')}
        </CSDAlert>
        <input
          ref={fileInputRef}
          type="file"
          accept=".json"
          onChange={handleFileUpload}
          style={{ display: 'none' }}
        />
        <CSDActionButton
          variant="outlined"
          startIcon={<CSDIcon name="uploadFile" />}
          onClick={() => fileInputRef.current?.click()}
          fullWidth
        >
          {t('security.profiles.select_file')}
        </CSDActionButton>
        <CSDTextField
          id="security-profiles-import-data"
          name="importData"
          label={t('security.profiles.json_data')}
          value={importData}
          onChange={setImportData}
          multiline
          rows={10}
          fullWidth
          placeholder='{"name": "My Profile", "rules": [...]}'
        />
      </CSDStack>
    </CSDFormDialog>
  );

  return (
    <AdvancedFilterManager
      storageKey="security-profiles"
      filterFields={filterFields}
      simpleFilters={simpleFilters}
      data={profiles}
      onSimpleFilterChange={handleSimpleFilterChange}
      pagination={pagination}
      autoSave={false}
      useServerFiltering={true}
      actions={
        <CSDBox sx={{ display: 'flex', gap: 1 }}>
          <CSDActionButton
            id="security-profiles-import-button"
            variant="outlined"
            startIcon={<CSDIcon name="upload" />}
            onClick={() => setImportOpen(true)}
          >
            {t('security.profiles.import')}
          </CSDActionButton>
          <CSDActionButton
            id="security-profiles-add-button"
            variant="contained"
            startIcon={<CSDIcon name="add" />}
            onClick={() => handleOpen(null)}
          >
            {t('security.profiles.add_profile')}
          </CSDActionButton>
        </CSDBox>
      }
    >
      {({ filteredData, filterBar }) => (
        <CSDCrudPage
          entityName="security-profiles"
          statCards={statCards}
          filterBar={filterBar}
          loading={loading}
          error={error}
          formDialog={<>{formDialog}{importDialog}</>}
          deleteDialog={deleteDialog}
        >
          <CSDDataGrid
            id="security-profiles-data-grid"
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

export default ProfilesPage;
