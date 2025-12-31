import React, { useState, useEffect, useCallback, useMemo } from 'react';
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
  CSDConfirmDialog,
  CSDFormDialog,
  CSDBox,
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

// Types
interface FirewallRule {
  id: string;
  name: string;
  description: string;
  chain: string;
  protocol: string;
  sourceIp: string;
  destIp: string;
  destPort: string;
  action: string;
  priority: number;
  enabled: boolean;
}

interface FirewallProfile {
  id: string;
  name: string;
  description: string;
  isDefault: boolean;
  enabled: boolean;
  inputPolicy: string;
  outputPolicy: string;
  forwardPolicy: string;
  enableNat: boolean;
  enableConntrack: boolean;
  allowLoopback: boolean;
  allowEstablished: boolean;
  allowIcmpPing: boolean;
  enableIpv6: boolean;
  rules: FirewallRule[];
  createdAt: string;
  updatedAt: string;
}

// GraphQL queries
const PROFILE_QUERY = `
  query SecurityProfile($id: ID!) {
    securityProfile(id: $id) {
      id name description isDefault enabled
      inputPolicy outputPolicy forwardPolicy
      enableNat enableConntrack allowLoopback allowEstablished allowIcmpPing enableIpv6
      rules { id name description chain protocol sourceIp destIp destPort action priority enabled }
      createdAt updatedAt
    }
  }
`;

const ALL_RULES_QUERY = `
  query AllSecurityRules {
    securityRules(limit: 1000) {
      id name description chain protocol destPort action priority enabled
    }
  }
`;

const ADD_RULES_MUTATION = `
  mutation AddRulesToProfile($profileId: ID!, $ruleIds: [ID!]!) {
    addRulesToProfile(profileId: $profileId, ruleIds: $ruleIds) { id }
  }
`;

const REMOVE_RULES_MUTATION = `
  mutation RemoveRulesFromProfile($profileId: ID!, $ruleIds: [ID!]!) {
    removeRulesFromProfile(profileId: $profileId, ruleIds: $ruleIds) { id }
  }
`;

// Action colors
const actionColors: Record<string, 'success' | 'error' | 'warning' | 'default' | 'info'> = {
  ACCEPT: 'success',
  DROP: 'error',
  REJECT: 'warning',
  LOG: 'info',
  MASQUERADE: 'default',
  SNAT: 'default',
  DNAT: 'default',
};

export const ProfileDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { request } = useGraphQL();
  const { showSuccess, showError } = useSnackbar();
  const { t } = useTranslation();

  // State
  const [tab, setTab] = useState(0);
  const [profile, setProfile] = useState<FirewallProfile | null>(null);
  const [allRules, setAllRules] = useState<FirewallRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [mutationLoading, setMutationLoading] = useState(false);

  // Dialog state
  const [addRulesOpen, setAddRulesOpen] = useState(false);
  const [removeConfirmOpen, setRemoveConfirmOpen] = useState(false);
  const [selectedRuleIds, setSelectedRuleIds] = useState<Set<string>>(new Set());
  const [rulesToRemove, setRulesToRemove] = useState<string[]>([]);

  // Breadcrumb
  useBreadcrumb([
    BREADCRUMBS.PILOTE,
    BREADCRUMBS.SECURITY,
    { labelKey: 'breadcrumb.profiles', path: '/pilote/security/profiles' },
    { labelKey: 'common.view' },
  ]);

  // Load profile
  const loadProfile = useCallback(async () => {
    setLoading(true);
    try {
      const data = await request<{ securityProfile: FirewallProfile }>(PROFILE_QUERY, { id });
      setProfile(data.securityProfile);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('security.profiles.error_loading'));
    } finally {
      setLoading(false);
    }
  }, [id, request, t]);

  // Load all rules for add dialog
  const loadAllRules = useCallback(async () => {
    try {
      const data = await request<{ securityRules: FirewallRule[] }>(ALL_RULES_QUERY);
      setAllRules(data.securityRules || []);
    } catch {
      // Silently handle
    }
  }, [request]);

  useEffect(() => {
    loadProfile();
    loadAllRules();
  }, [loadProfile, loadAllRules]);

  // Available rules (not in profile)
  const availableRules = useMemo(() => {
    if (!profile) return allRules;
    const profileRuleIds = new Set(profile.rules.map(r => r.id));
    return allRules.filter(r => !profileRuleIds.has(r.id));
  }, [allRules, profile]);

  // Handle add rules
  const handleAddRules = async () => {
    if (selectedRuleIds.size === 0) return;
    setMutationLoading(true);
    try {
      await request(ADD_RULES_MUTATION, {
        profileId: id,
        ruleIds: Array.from(selectedRuleIds),
      });
      showSuccess(t('security.profiles.rules_added'));
      setAddRulesOpen(false);
      setSelectedRuleIds(new Set());
      await loadProfile();
    } catch (err) {
      showError(err instanceof Error ? err.message : t('security.profiles.error_saving'));
    } finally {
      setMutationLoading(false);
    }
  };

  // Handle remove rules
  const handleRemoveRules = async () => {
    if (rulesToRemove.length === 0) return;
    setMutationLoading(true);
    try {
      await request(REMOVE_RULES_MUTATION, {
        profileId: id,
        ruleIds: rulesToRemove,
      });
      showSuccess(t('security.profiles.rules_removed'));
      setRemoveConfirmOpen(false);
      setRulesToRemove([]);
      await loadProfile();
    } catch (err) {
      showError(err instanceof Error ? err.message : t('security.profiles.error_saving'));
    } finally {
      setMutationLoading(false);
    }
  };

  const openRemoveConfirm = (ruleIds: string[]) => {
    setRulesToRemove(ruleIds);
    setRemoveConfirmOpen(true);
  };

  // Loading state
  if (loading) {
    return (
      <CSDDetailPage title={t('common.loading')} loading>
        <CSDStack spacing={2} alignItems="center" justifyContent="center" sx={{ minHeight: 200 }}>
          <CSDCircularProgress />
        </CSDStack>
      </CSDDetailPage>
    );
  }

  // Error state
  if (error || !profile) {
    return (
      <CSDDetailPage title={t('security.profiles.not_found')}>
        <CSDAlert severity="error">{error || t('security.profiles.not_found')}</CSDAlert>
      </CSDDetailPage>
    );
  }

  // Profile rules columns
  const ruleColumns = [
    {
      id: 'name',
      label: 'security.rules.name',
      render: (rule: FirewallRule) => (
        <CSDTypography variant="body2" fontWeight="bold">{rule.name}</CSDTypography>
      ),
    },
    {
      id: 'chain',
      label: 'security.rules.chain',
      render: (rule: FirewallRule) => (
        <CSDChip label={rule.chain} size="small" variant="outlined" />
      ),
    },
    {
      id: 'protocol',
      label: 'security.rules.protocol',
      render: (rule: FirewallRule) => rule.protocol?.toUpperCase() || '-',
    },
    {
      id: 'destPort',
      label: 'security.rules.dest_port',
      render: (rule: FirewallRule) => rule.destPort || '-',
    },
    {
      id: 'action',
      label: 'security.rules.action',
      render: (rule: FirewallRule) => (
        <CSDChip
          label={rule.action}
          size="small"
          color={actionColors[rule.action?.toUpperCase()] || 'default'}
        />
      ),
    },
    {
      id: 'priority',
      label: 'security.rules.priority',
      render: (rule: FirewallRule) => rule.priority,
    },
    {
      id: 'enabled',
      label: 'security.rules.enabled',
      render: (rule: FirewallRule) => (
        <CSDIcon
          name={rule.enabled ? 'checkCircle' : 'cancel'}
          color={rule.enabled ? 'success' : 'disabled'}
        />
      ),
    },
  ];

  // Available rules columns (for add dialog)
  const availableRuleColumns = [
    {
      id: 'name',
      label: 'security.rules.name',
      render: (rule: FirewallRule) => rule.name,
    },
    {
      id: 'chain',
      label: 'security.rules.chain',
      render: (rule: FirewallRule) => rule.chain,
    },
    {
      id: 'action',
      label: 'security.rules.action',
      render: (rule: FirewallRule) => (
        <CSDChip
          label={rule.action}
          size="small"
          color={actionColors[rule.action?.toUpperCase()] || 'default'}
        />
      ),
    },
    {
      id: 'destPort',
      label: 'security.rules.dest_port',
      render: (rule: FirewallRule) => rule.destPort || '-',
    },
  ];

  // Row actions for profile rules
  const ruleActions = [
    {
      icon: 'delete',
      onClick: (rule: FirewallRule) => openRemoveConfirm([rule.id]),
      tooltip: 'security.profiles.remove_rules',
      color: 'error' as const,
    },
  ];

  // Render profile rules tab
  const renderRulesTab = () => {
    if (!profile.rules || profile.rules.length === 0) {
      return (
        <CSDStack spacing={2} alignItems="center" sx={{ py: 4 }}>
          <CSDTypography color="text.secondary">{t('security.profiles.no_rules')}</CSDTypography>
          <CSDActionButton
            variant="contained"
            startIcon={<CSDIcon name="add" />}
            onClick={() => setAddRulesOpen(true)}
          >
            {t('security.profiles.add_rules')}
          </CSDActionButton>
        </CSDStack>
      );
    }

    return (
      <CSDDataGrid
        id="profile-rules-grid"
        data={profile.rules}
        columns={ruleColumns}
        actions={ruleActions}
        keyField="id"
        hoverable
      />
    );
  };

  // Render settings tab
  const renderSettingsTab = () => (
    <CSDStack spacing={3}>
      <CSDTypography variant="subtitle1" fontWeight="bold">
        {t('security.profiles.default_policies')}
      </CSDTypography>
      <CSDInfoGrid>
        <CSDInfoItem
          label={t('security.profiles.input_policy')}
          value={<CSDChip label={profile.inputPolicy.toUpperCase()} size="small" color={profile.inputPolicy === 'accept' ? 'success' : 'error'} />}
        />
        <CSDInfoItem
          label={t('security.profiles.output_policy')}
          value={<CSDChip label={profile.outputPolicy.toUpperCase()} size="small" color={profile.outputPolicy === 'accept' ? 'success' : 'error'} />}
        />
        <CSDInfoItem
          label={t('security.profiles.forward_policy')}
          value={<CSDChip label={profile.forwardPolicy.toUpperCase()} size="small" color={profile.forwardPolicy === 'accept' ? 'success' : 'error'} />}
        />
      </CSDInfoGrid>

      <CSDTypography variant="subtitle1" fontWeight="bold">
        {t('security.profiles.features')}
      </CSDTypography>
      <CSDInfoGrid>
        <CSDInfoItem
          label={t('security.profiles.allow_loopback')}
          value={<CSDIcon name={profile.allowLoopback ? 'checkCircle' : 'cancel'} color={profile.allowLoopback ? 'success' : 'disabled'} />}
        />
        <CSDInfoItem
          label={t('security.profiles.allow_established')}
          value={<CSDIcon name={profile.allowEstablished ? 'checkCircle' : 'cancel'} color={profile.allowEstablished ? 'success' : 'disabled'} />}
        />
        <CSDInfoItem
          label={t('security.profiles.allow_icmp_ping')}
          value={<CSDIcon name={profile.allowIcmpPing ? 'checkCircle' : 'cancel'} color={profile.allowIcmpPing ? 'success' : 'disabled'} />}
        />
        <CSDInfoItem
          label={t('security.profiles.enable_conntrack')}
          value={<CSDIcon name={profile.enableConntrack ? 'checkCircle' : 'cancel'} color={profile.enableConntrack ? 'success' : 'disabled'} />}
        />
        <CSDInfoItem
          label={t('security.profiles.enable_nat')}
          value={<CSDIcon name={profile.enableNat ? 'checkCircle' : 'cancel'} color={profile.enableNat ? 'success' : 'disabled'} />}
        />
        <CSDInfoItem
          label={t('security.profiles.enable_ipv6')}
          value={<CSDIcon name={profile.enableIpv6 ? 'checkCircle' : 'cancel'} color={profile.enableIpv6 ? 'success' : 'disabled'} />}
        />
      </CSDInfoGrid>
    </CSDStack>
  );

  // Render tab content
  const renderTabContent = () => {
    switch (tab) {
      case 0:
        return renderRulesTab();
      case 1:
        return renderSettingsTab();
      default:
        return null;
    }
  };

  return (
    <CSDDetailPage
      title={profile.name}
      actions={
        <CSDBox sx={{ display: 'flex', gap: 1 }}>
          <CSDActionButton
            variant="outlined"
            startIcon={<CSDIcon name="add" />}
            onClick={() => setAddRulesOpen(true)}
          >
            {t('security.profiles.add_rules')}
          </CSDActionButton>
          <CSDActionButton
            variant="outlined"
            startIcon={<CSDIcon name="sync" />}
            onClick={loadProfile}
          >
            {t('common.refresh')}
          </CSDActionButton>
        </CSDBox>
      }
    >
      <CSDStack spacing={3}>
        {/* Profile Info */}
        <CSDPaper>
          <CSDInfoGrid>
            <CSDInfoItem label={t('security.profiles.name')} value={profile.name} />
            <CSDInfoItem
              label={t('security.profiles.is_default')}
              value={profile.isDefault ? <CSDChip label="Default" color="warning" size="small" /> : '-'}
            />
            <CSDInfoItem
              label={t('security.profiles.enabled')}
              value={<CSDIcon name={profile.enabled ? 'checkCircle' : 'cancel'} color={profile.enabled ? 'success' : 'disabled'} />}
            />
            <CSDInfoItem
              label={t('security.profiles.rules')}
              value={<CSDChip label={t('security.profiles.rules_count', { count: profile.rules?.length || 0 })} size="small" variant="outlined" />}
            />
            {profile.description && (
              <CSDInfoItem label={t('security.profiles.description')} value={profile.description} fullWidth />
            )}
            <CSDInfoItem label={t('common.created')} value={formatDate(profile.createdAt, '-')} />
            <CSDInfoItem label={t('common.updated')} value={formatDate(profile.updatedAt, '-')} />
          </CSDInfoGrid>
        </CSDPaper>

        {/* Tabs */}
        <CSDPaper>
          <CSDTabs value={tab} onChange={(_: unknown, v: number) => setTab(v)}>
            <CSDTab label={`${t('security.profiles.rules')} (${profile.rules?.length || 0})`} />
            <CSDTab label={t('security.profiles.features')} />
          </CSDTabs>
          <CSDStack spacing={2} sx={{ p: 2 }}>
            {renderTabContent()}
          </CSDStack>
        </CSDPaper>
      </CSDStack>

      {/* Add Rules Dialog */}
      <CSDFormDialog
        id="profile-add-rules-dialog"
        open={addRulesOpen}
        onClose={() => {
          setAddRulesOpen(false);
          setSelectedRuleIds(new Set());
        }}
        title="security.profiles.add_rules"
        icon="add"
        maxWidth="md"
        actions={
          <>
            <CSDActionButton variant="outlined" onClick={() => setAddRulesOpen(false)}>
              {t('common.cancel')}
            </CSDActionButton>
            <CSDActionButton
              variant="contained"
              onClick={handleAddRules}
              disabled={selectedRuleIds.size === 0 || mutationLoading}
            >
              {t('security.profiles.add_rules')} ({selectedRuleIds.size})
            </CSDActionButton>
          </>
        }
      >
        {availableRules.length === 0 ? (
          <CSDStack spacing={2} alignItems="center" sx={{ py: 4 }}>
            <CSDTypography color="text.secondary">{t('security.profiles.no_available_rules')}</CSDTypography>
          </CSDStack>
        ) : (
          <CSDDataGrid
            id="available-rules-grid"
            data={availableRules}
            columns={availableRuleColumns}
            keyField="id"
            hoverable
            selectable
            selectedIds={selectedRuleIds}
            onSelectionChange={setSelectedRuleIds}
          />
        )}
      </CSDFormDialog>

      {/* Remove Rules Confirm Dialog */}
      <CSDConfirmDialog
        id="profile-remove-rules-dialog"
        open={removeConfirmOpen}
        onClose={() => setRemoveConfirmOpen(false)}
        onConfirm={handleRemoveRules}
        title={t('security.profiles.remove_rules')}
        message={t('security.profiles.remove_rules_confirmation', { count: rulesToRemove.length })}
        confirmLabel={t('common.delete')}
        cancelLabel={t('common.cancel')}
        severity="warning"
      />
    </CSDDetailPage>
  );
};

export default ProfileDetailPage;
