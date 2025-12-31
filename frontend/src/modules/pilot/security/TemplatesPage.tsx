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
interface FirewallTemplate {
  id: string;
  name: string;
  description: string;
  category: string;
  isBuiltIn: boolean;
  rulesJson: string;
  createdAt: string;
  updatedAt: string;
}

interface TemplateInput {
  name: string;
  description?: string;
  category: string;
  rulesJson?: string;
}

// Category colors
const categoryColors: Record<string, 'primary' | 'secondary' | 'success' | 'warning' | 'info' | 'error' | 'default'> = {
  server: 'primary',
  security: 'error',
  network: 'info',
  custom: 'default',
};

// Initial form state
const INITIAL_FORM: Partial<TemplateInput> = {
  name: '',
  description: '',
  category: 'custom',
  rulesJson: '[]',
};

export const TemplatesPage: React.FC = () => {
  const { request } = useGraphQL();
  const { t } = useTranslation();
  const pagination = usePagination('security-templates');

  // Breadcrumb
  useBreadcrumb([
    BREADCRUMBS.PILOTE,
    BREADCRUMBS.SECURITY,
    { labelKey: 'breadcrumb.templates', icon: 'template' },
  ]);

  // State
  const [templates, setTemplates] = useState<FirewallTemplate[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [totalCount, setTotalCount] = useState(0);

  // Filter state
  const [simpleFilter, setSimpleFilter] = useState<Record<string, unknown> | null>(null);

  // Dialog state
  const [formOpen, setFormOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [editingTemplate, setEditingTemplate] = useState<FirewallTemplate | null>(null);

  // Form state
  const [form, setForm] = useState<Partial<TemplateInput>>(INITIAL_FORM);
  const [initialForm, setInitialForm] = useState<Partial<TemplateInput>>(INITIAL_FORM);
  const [formError, setFormError] = useState<string | null>(null);

  // Stats
  const [stats, setStats] = useState({
    total: 0,
    builtIn: 0,
    custom: 0,
  });

  // Check if form has changes
  const hasChanges = useMemo(() => {
    return JSON.stringify(form) !== JSON.stringify(initialForm);
  }, [form, initialForm]);

  // Check if form is valid
  const isFormValid = useMemo(() => {
    return !!(form.name?.trim() && form.category);
  }, [form]);

  // Category options
  const categoryOptions = useMemo(() => [
    { value: 'server', label: t('security.templates.category_server') },
    { value: 'security', label: t('security.templates.category_security') },
    { value: 'network', label: t('security.templates.category_network') },
    { value: 'custom', label: t('security.templates.category_custom') },
  ], [t]);

  // Filter options
  const categoryFilterOptions = useMemo(() => [
    { value: '', label: t('common.all') },
    ...categoryOptions,
  ], [t, categoryOptions]);

  // Filter fields
  const filterFields: FilterFieldDefinition[] = useMemo(() => [
    { id: 'name', label: t('security.templates.name'), type: 'string' },
    { id: 'category', label: t('security.templates.category'), type: 'select', options: categoryFilterOptions },
  ], [t, categoryFilterOptions]);

  // Simple filters
  const simpleFilters: SimpleFilterField[] = useMemo(() => [
    { type: 'search', placeholder: 'security.templates.search_placeholder' },
    { type: 'select', name: 'category', label: 'security.templates.category', options: categoryFilterOptions },
  ], [categoryFilterOptions]);

  // Handle simple filter changes
  const handleSimpleFilterChange = useCallback((filters: Record<string, unknown>) => {
    const templateFilter: Record<string, unknown> = {};
    if (filters.search && typeof filters.search === 'string' && filters.search.trim()) {
      templateFilter.search = filters.search.trim();
    }
    if (filters.category && filters.category !== '') {
      templateFilter.category = filters.category;
    }
    setSimpleFilter(Object.keys(templateFilter).length > 0 ? templateFilter : null);
  }, []);

  // Load templates
  const loadTemplates = useCallback(async () => {
    setLoading(true);
    try {
      const listQuery = `
        query ListSecurityTemplates($limit: Int, $offset: Int, $filter: FirewallTemplateFilter) {
          securityTemplates(limit: $limit, offset: $offset, filter: $filter) {
            id name description category isBuiltIn rulesJson createdAt updatedAt
          }
          securityTemplatesCount
        }
      `;

      const variables = {
        limit: pagination.rowsPerPage,
        offset: pagination.page * pagination.rowsPerPage,
        filter: simpleFilter,
      };

      const data = await request<{ securityTemplates: FirewallTemplate[]; securityTemplatesCount: number }>(listQuery, variables);
      const templateList = data.securityTemplates || [];
      setTemplates(templateList);
      setTotalCount(data.securityTemplatesCount || templateList.length);

      // Calculate stats
      setStats({
        total: data.securityTemplatesCount || templateList.length,
        builtIn: templateList.filter(t => t.isBuiltIn).length,
        custom: templateList.filter(t => !t.isBuiltIn).length,
      });

      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('security.templates.error_loading'));
    } finally {
      setLoading(false);
    }
  }, [request, pagination.page, pagination.rowsPerPage, simpleFilter, t]);

  useEffect(() => {
    loadTemplates();
  }, [loadTemplates]);

  // Count rules in JSON
  const countRules = (rulesJson: string): number => {
    try {
      const rules = JSON.parse(rulesJson || '[]');
      return Array.isArray(rules) ? rules.length : 0;
    } catch {
      return 0;
    }
  };

  // Handlers
  const handleOpen = (template: FirewallTemplate | null) => {
    setFormError(null);
    if (template) {
      setEditingTemplate(template);
      const formState: Partial<TemplateInput> = {
        name: template.name || '',
        description: template.description || '',
        category: template.category || 'custom',
        rulesJson: template.rulesJson || '[]',
      };
      setForm(formState);
      setInitialForm(formState);
    } else {
      setEditingTemplate(null);
      setForm(INITIAL_FORM);
      setInitialForm(INITIAL_FORM);
    }
    setFormOpen(true);
  };

  const handleClose = () => {
    setFormOpen(false);
    setEditingTemplate(null);
    setFormError(null);
  };

  const handleDelete = (template: FirewallTemplate) => {
    if (template.isBuiltIn) {
      setError('Cannot delete built-in templates');
      return;
    }
    setEditingTemplate(template);
    setDeleteOpen(true);
  };

  const handleFormSubmit = async () => {
    try {
      if (editingTemplate) {
        const mutation = `
          mutation UpdateSecurityTemplate($id: ID!, $input: UpdateFirewallTemplateInput!) {
            updateSecurityTemplate(id: $id, input: $input) { id }
          }
        `;
        await request(mutation, { id: editingTemplate.id, input: form });
      } else {
        const mutation = `
          mutation CreateSecurityTemplate($input: CreateFirewallTemplateInput!) {
            createSecurityTemplate(input: $input) { id }
          }
        `;
        await request(mutation, { input: form });
      }
      handleClose();
      await loadTemplates();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : t('security.templates.error_saving'));
    }
  };

  const handleConfirmDelete = async () => {
    if (!editingTemplate) return;
    try {
      await request(`mutation DeleteSecurityTemplate($id: ID!) { deleteSecurityTemplate(id: $id) }`, { id: editingTemplate.id });
      setDeleteOpen(false);
      setEditingTemplate(null);
      await loadTemplates();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('security.templates.error_deleting'));
    }
  };

  // Stat cards
  const statCards: StatCardData[] = [
    { title: t('clusters.total'), value: String(stats.total), icon: 'template', color: 'primary' },
    { title: t('security.templates.is_built_in'), value: String(stats.builtIn), icon: 'lock', color: 'warning' },
    { title: t('security.templates.category_custom'), value: String(stats.custom), icon: 'edit', color: 'info' },
  ];

  // DataGrid columns
  const columns = [
    {
      id: 'name',
      label: 'security.templates.name',
      sortable: true,
      render: (template: FirewallTemplate) => (
        <CSDStack direction="row" spacing={1} alignItems="center">
          <CSDTypography variant="body2" fontWeight="bold">{template.name}</CSDTypography>
          {template.isBuiltIn && <CSDChip label={t('security.templates.is_built_in')} size="small" color="warning" variant="outlined" />}
        </CSDStack>
      ),
    },
    {
      id: 'description',
      label: 'security.templates.description',
      sortable: true,
      render: (template: FirewallTemplate) => (
        <CSDTypography variant="body2" sx={{ maxWidth: 300, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {template.description || '-'}
        </CSDTypography>
      ),
    },
    {
      id: 'category',
      label: 'security.templates.category',
      sortable: true,
      render: (template: FirewallTemplate) => (
        <CSDChip
          label={t(`security.templates.category_${template.category}`)}
          size="small"
          color={categoryColors[template.category] || 'default'}
        />
      ),
    },
    {
      id: 'rules',
      label: 'security.profiles.rules',
      render: (template: FirewallTemplate) => (
        <CSDChip label={t('security.templates.rules_count', { count: countRules(template.rulesJson) })} size="small" variant="outlined" />
      ),
    },
    {
      id: 'createdAt',
      label: 'clusters.created',
      sortable: true,
      render: (template: FirewallTemplate) => formatDate(template.createdAt, '-'),
    },
  ];

  // Row actions
  const actions = [
    { icon: 'edit', onClick: (template: FirewallTemplate) => handleOpen(template), tooltip: 'common.edit', color: 'primary' as const, disabled: (template: FirewallTemplate) => template.isBuiltIn },
    { icon: 'delete', onClick: (template: FirewallTemplate) => handleDelete(template), tooltip: 'common.delete', color: 'error' as const, disabled: (template: FirewallTemplate) => template.isBuiltIn },
  ];

  // Form dialog
  const formDialog = (
    <CSDFormDialog
      id="security-templates-form-dialog"
      open={formOpen}
      onClose={handleClose}
      title={editingTemplate ? 'security.templates.edit_template' : 'security.templates.add_template'}
      icon={editingTemplate ? 'edit' : 'add'}
      error={formError}
      trackChanges={!!editingTemplate}
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
            disabled={!isFormValid || (editingTemplate && !hasChanges) || editingTemplate?.isBuiltIn}
          >
            {t('common.save')}
          </CSDActionButton>
        </>
      }
    >
      <CSDStack spacing={2}>
        <CSDTextField
          id="security-templates-form-name"
          name="name"
          label={t('security.templates.name')}
          value={form.name || ''}
          onChange={(v) => setForm({ ...form, name: v })}
          required
          fullWidth
          disabled={editingTemplate?.isBuiltIn}
        />
        <CSDTextField
          id="security-templates-form-description"
          name="description"
          label={t('security.templates.description')}
          value={form.description || ''}
          onChange={(v) => setForm({ ...form, description: v })}
          multiline
          rows={3}
          fullWidth
          disabled={editingTemplate?.isBuiltIn}
        />
        <CSDSelect
          id="security-templates-form-category"
          name="category"
          label={t('security.templates.category')}
          value={form.category || 'custom'}
          onChange={(v) => setForm({ ...form, category: v })}
          options={categoryOptions}
          required
          fullWidth
          disabled={editingTemplate?.isBuiltIn}
        />
        <CSDTextField
          id="security-templates-form-rules"
          name="rulesJson"
          label="Rules JSON"
          value={form.rulesJson || '[]'}
          onChange={(v) => setForm({ ...form, rulesJson: v })}
          multiline
          rows={10}
          fullWidth
          disabled={editingTemplate?.isBuiltIn}
          helperText="JSON array of rule definitions"
        />
      </CSDStack>
    </CSDFormDialog>
  );

  // Delete dialog
  const deleteDialog = (
    <CSDConfirmDialog
      id="security-templates-delete-dialog"
      open={deleteOpen}
      onClose={() => setDeleteOpen(false)}
      onConfirm={handleConfirmDelete}
      title={t('security.templates.delete_template')}
      message={t('security.templates.delete_confirmation', { name: editingTemplate?.name || '' })}
      confirmLabel={t('common.delete')}
      cancelLabel={t('common.cancel')}
      severity="error"
    />
  );

  return (
    <AdvancedFilterManager
      storageKey="security-templates"
      filterFields={filterFields}
      simpleFilters={simpleFilters}
      data={templates}
      onSimpleFilterChange={handleSimpleFilterChange}
      pagination={pagination}
      autoSave={false}
      useServerFiltering={true}
      actions={
        <CSDBox sx={{ display: 'flex', gap: 1 }}>
          <CSDActionButton
            id="security-templates-add-button"
            variant="contained"
            startIcon={<CSDIcon name="add" />}
            onClick={() => handleOpen(null)}
          >
            {t('security.templates.add_template')}
          </CSDActionButton>
        </CSDBox>
      }
    >
      {({ filteredData, filterBar }) => (
        <CSDCrudPage
          entityName="security-templates"
          statCards={statCards}
          filterBar={filterBar}
          loading={loading}
          error={error}
          formDialog={formDialog}
          deleteDialog={deleteDialog}
        >
          <CSDDataGrid
            id="security-templates-data-grid"
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

export default TemplatesPage;
