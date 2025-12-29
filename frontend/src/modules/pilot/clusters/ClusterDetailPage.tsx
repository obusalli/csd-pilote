import React from 'react';
import {
  CSDLayoutPage,
  CSDBox,
  CSDTypography,
  CSDPaper,
  CSDGrid,
  CSDChip,
  CSDTabs,
  CSDTab,
  CSDButton,
} from 'csd_core/UI';
import { useBreadcrumb, useParams } from 'csd_core/Providers';

export const ClusterDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [tab, setTab] = React.useState(0);

  useBreadcrumb([
    { label: 'Pilote', path: '/pilote' },
    { label: 'Kubernetes', path: '/pilote/kubernetes' },
    { label: 'Clusters', path: '/pilote/kubernetes/clusters' },
    { label: 'Details' },
  ]);

  // TODO: Fetch cluster details using GraphQL

  return (
    <CSDLayoutPage
      title="Cluster Details"
      actions={
        <CSDButton variant="outlined" color="primary">
          Test Connection
        </CSDButton>
      }
    >
      <CSDPaper sx={{ mb: 3 }}>
        <CSDBox sx={{ p: 3 }}>
          <CSDGrid container spacing={2}>
            <CSDGrid item xs={12} md={6}>
              <CSDTypography variant="subtitle2" color="text.secondary">
                Cluster ID
              </CSDTypography>
              <CSDTypography>{id}</CSDTypography>
            </CSDGrid>
            <CSDGrid item xs={12} md={6}>
              <CSDTypography variant="subtitle2" color="text.secondary">
                Status
              </CSDTypography>
              <CSDChip label="PENDING" color="warning" size="small" />
            </CSDGrid>
          </CSDGrid>
        </CSDBox>
      </CSDPaper>

      <CSDPaper>
        <CSDTabs value={tab} onChange={(_: unknown, v: number) => setTab(v)}>
          <CSDTab label="Namespaces" />
          <CSDTab label="Deployments" />
          <CSDTab label="Pods" />
          <CSDTab label="Services" />
        </CSDTabs>

        <CSDBox sx={{ p: 3 }}>
          {tab === 0 && (
            <CSDTypography color="text.secondary">
              Connect to the cluster to view namespaces
            </CSDTypography>
          )}
          {tab === 1 && (
            <CSDTypography color="text.secondary">
              Connect to the cluster to view deployments
            </CSDTypography>
          )}
          {tab === 2 && (
            <CSDTypography color="text.secondary">
              Connect to the cluster to view pods
            </CSDTypography>
          )}
          {tab === 3 && (
            <CSDTypography color="text.secondary">
              Connect to the cluster to view services
            </CSDTypography>
          )}
        </CSDBox>
      </CSDPaper>
    </CSDLayoutPage>
  );
};

export default ClusterDetailPage;
