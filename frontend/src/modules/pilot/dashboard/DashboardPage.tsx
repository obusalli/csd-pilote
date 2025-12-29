import React from 'react';
import {
  CSDLayoutPage,
  CSDBox,
  CSDTypography,
  CSDPaper,
  CSDGrid,
  CSDStatCard,
  CSDStatsGrid,
} from 'csd_core/UI';
import { useBreadcrumb } from 'csd_core/Providers';

export const DashboardPage: React.FC = () => {
  useBreadcrumb([
    { label: 'Pilote', path: '/pilote' },
    { label: 'Dashboard' },
  ]);

  return (
    <CSDLayoutPage title="Infrastructure Dashboard">
      <CSDStatsGrid>
        <CSDStatCard
          title="Kubernetes Clusters"
          value="0"
          icon="cloud"
          color="primary"
          linkTo="/pilote/kubernetes/clusters"
        />
        <CSDStatCard
          title="Libvirt Hypervisors"
          value="0"
          icon="dns"
          color="secondary"
          linkTo="/pilote/libvirt/hypervisors"
        />
        <CSDStatCard
          title="Container Engines"
          value="0"
          icon="view_in_ar"
          color="info"
          linkTo="/pilote/containers/engines"
        />
        <CSDStatCard
          title="Active VMs"
          value="0"
          icon="computer"
          color="success"
        />
      </CSDStatsGrid>

      <CSDBox sx={{ mt: 4 }}>
        <CSDGrid container spacing={3}>
          <CSDGrid item xs={12} md={6}>
            <CSDPaper sx={{ p: 3 }}>
              <CSDTypography variant="h6" gutterBottom>
                Recent Activity
              </CSDTypography>
              <CSDTypography color="text.secondary">
                No recent activity
              </CSDTypography>
            </CSDPaper>
          </CSDGrid>
          <CSDGrid item xs={12} md={6}>
            <CSDPaper sx={{ p: 3 }}>
              <CSDTypography variant="h6" gutterBottom>
                System Status
              </CSDTypography>
              <CSDTypography color="text.secondary">
                All systems operational
              </CSDTypography>
            </CSDPaper>
          </CSDGrid>
        </CSDGrid>
      </CSDBox>
    </CSDLayoutPage>
  );
};

export default DashboardPage;
