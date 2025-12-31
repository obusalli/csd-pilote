import React, { useState, useEffect } from 'react';
import {
  CSDLayoutPage,
  CSDBox,
  CSDTypography,
  CSDPaper,
  CSDGrid,
  CSDStatCard,
  CSDStatsGrid,
  CSDCircularProgress,
} from 'csd_core/UI';
import { useBreadcrumb, useTranslation } from 'csd_core/Providers';
import { useGraphQL } from '../../../shared/hooks/useGraphQL';

// GraphQL query for dashboard stats
const DASHBOARD_STATS_QUERY = `
  query DashboardStats {
    dashboardStats {
      clustersCount
      clustersConnected
      clustersDeploying
      clustersError
      hypervisorsCount
      hypervisorsConnected
      hypervisorsDeploying
      hypervisorsError
      containerEnginesCount
      enginesConnected
      enginesError
    }
  }
`;

interface DashboardStats {
  clustersCount: number;
  clustersConnected: number;
  clustersDeploying: number;
  clustersError: number;
  hypervisorsCount: number;
  hypervisorsConnected: number;
  hypervisorsDeploying: number;
  hypervisorsError: number;
  containerEnginesCount: number;
  enginesConnected: number;
  enginesError: number;
}

interface DashboardStatsResponse {
  dashboardStats: DashboardStats;
}

export const DashboardPage: React.FC = () => {
  const { t } = useTranslation();
  const { request } = useGraphQL();
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useBreadcrumb([
    { label: 'Pilote', path: '/pilote' },
    { label: t('pilot.dashboard.title', 'Dashboard') },
  ]);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        setLoading(true);
        const response = await request<DashboardStatsResponse>(DASHBOARD_STATS_QUERY);
        setStats(response.dashboardStats);
        setError(null);
      } catch (err) {
        console.error('Failed to fetch dashboard stats:', err);
        setError(err instanceof Error ? err.message : 'Failed to fetch stats');
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
  }, [request]);

  const formatStatSubtitle = (connected: number, deploying: number, error: number): string => {
    const parts: string[] = [];
    if (connected > 0) parts.push(`${connected} ${t('pilot.dashboard.connected', 'connected')}`);
    if (deploying > 0) parts.push(`${deploying} ${t('pilot.dashboard.deploying', 'deploying')}`);
    if (error > 0) parts.push(`${error} ${t('pilot.dashboard.error', 'error')}`);
    return parts.join(', ') || t('pilot.dashboard.noResources', 'No resources');
  };

  if (loading) {
    return (
      <CSDLayoutPage title={t('pilot.dashboard.title', 'Infrastructure Dashboard')}>
        <CSDBox sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 300 }}>
          <CSDCircularProgress />
        </CSDBox>
      </CSDLayoutPage>
    );
  }

  return (
    <CSDLayoutPage title={t('pilot.dashboard.title', 'Infrastructure Dashboard')}>
      <CSDStatsGrid>
        <CSDStatCard
          title={t('pilot.dashboard.clusters', 'Kubernetes Clusters')}
          value={String(stats?.clustersCount ?? 0)}
          subtitle={stats ? formatStatSubtitle(stats.clustersConnected, stats.clustersDeploying, stats.clustersError) : undefined}
          icon="cloud"
          color="primary"
          linkTo="/pilote/kubernetes/clusters"
        />
        <CSDStatCard
          title={t('pilot.dashboard.hypervisors', 'Libvirt Hypervisors')}
          value={String(stats?.hypervisorsCount ?? 0)}
          subtitle={stats ? formatStatSubtitle(stats.hypervisorsConnected, stats.hypervisorsDeploying, stats.hypervisorsError) : undefined}
          icon="dns"
          color="secondary"
          linkTo="/pilote/libvirt/hypervisors"
        />
        <CSDStatCard
          title={t('pilot.dashboard.containerEngines', 'Container Engines')}
          value={String(stats?.containerEnginesCount ?? 0)}
          subtitle={stats ? formatStatSubtitle(stats.enginesConnected, 0, stats.enginesError) : undefined}
          icon="view_in_ar"
          color="info"
          linkTo="/pilote/containers/engines"
        />
        <CSDStatCard
          title={t('pilot.dashboard.totalResources', 'Total Resources')}
          value={String((stats?.clustersCount ?? 0) + (stats?.hypervisorsCount ?? 0) + (stats?.containerEnginesCount ?? 0))}
          icon="inventory_2"
          color="success"
        />
      </CSDStatsGrid>

      <CSDBox sx={{ mt: 4 }}>
        <CSDGrid container spacing={3}>
          <CSDGrid size={{ xs: 12, md: 6 }}>
            <CSDPaper sx={{ p: 3 }}>
              <CSDTypography variant="h6" gutterBottom>
                {t('pilot.dashboard.statusSummary', 'Status Summary')}
              </CSDTypography>
              {error ? (
                <CSDTypography color="error">{error}</CSDTypography>
              ) : stats ? (
                <CSDBox>
                  <CSDTypography variant="body2" sx={{ mb: 1 }}>
                    <strong>{t('pilot.dashboard.clusters', 'Clusters')}:</strong> {stats.clustersConnected} {t('pilot.dashboard.connected', 'connected')}, {stats.clustersError} {t('pilot.dashboard.withErrors', 'with errors')}
                  </CSDTypography>
                  <CSDTypography variant="body2" sx={{ mb: 1 }}>
                    <strong>{t('pilot.dashboard.hypervisors', 'Hypervisors')}:</strong> {stats.hypervisorsConnected} {t('pilot.dashboard.connected', 'connected')}, {stats.hypervisorsError} {t('pilot.dashboard.withErrors', 'with errors')}
                  </CSDTypography>
                  <CSDTypography variant="body2">
                    <strong>{t('pilot.dashboard.containerEngines', 'Container Engines')}:</strong> {stats.enginesConnected} {t('pilot.dashboard.connected', 'connected')}, {stats.enginesError} {t('pilot.dashboard.withErrors', 'with errors')}
                  </CSDTypography>
                </CSDBox>
              ) : (
                <CSDTypography color="text.secondary">
                  {t('pilot.dashboard.noData', 'No data available')}
                </CSDTypography>
              )}
            </CSDPaper>
          </CSDGrid>
          <CSDGrid size={{ xs: 12, md: 6 }}>
            <CSDPaper sx={{ p: 3 }}>
              <CSDTypography variant="h6" gutterBottom>
                {t('pilot.dashboard.systemStatus', 'System Status')}
              </CSDTypography>
              {stats && stats.clustersError === 0 && stats.hypervisorsError === 0 && stats.enginesError === 0 ? (
                <CSDTypography color="success.main">
                  {t('pilot.dashboard.allOperational', 'All systems operational')}
                </CSDTypography>
              ) : stats ? (
                <CSDTypography color="warning.main">
                  {t('pilot.dashboard.issuesDetected', 'Some resources have issues')}
                </CSDTypography>
              ) : (
                <CSDTypography color="text.secondary">
                  {t('pilot.dashboard.checkingStatus', 'Checking status...')}
                </CSDTypography>
              )}
            </CSDPaper>
          </CSDGrid>
        </CSDGrid>
      </CSDBox>
    </CSDLayoutPage>
  );
};

export default DashboardPage;
