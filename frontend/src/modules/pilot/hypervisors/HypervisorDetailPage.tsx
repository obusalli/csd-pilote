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

export const HypervisorDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [tab, setTab] = React.useState(0);

  useBreadcrumb([
    { label: 'Pilote', path: '/pilote' },
    { label: 'Libvirt', path: '/pilote/libvirt' },
    { label: 'Hypervisors', path: '/pilote/libvirt/hypervisors' },
    { label: 'Details' },
  ]);

  return (
    <CSDLayoutPage
      title="Hypervisor Details"
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
                Hypervisor ID
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
          <CSDTab label="Virtual Machines" />
          <CSDTab label="Networks" />
          <CSDTab label="Storage Pools" />
          <CSDTab label="Volumes" />
        </CSDTabs>

        <CSDBox sx={{ p: 3 }}>
          {tab === 0 && (
            <CSDTypography color="text.secondary">
              Connect to the hypervisor to view VMs
            </CSDTypography>
          )}
          {tab === 1 && (
            <CSDTypography color="text.secondary">
              Connect to the hypervisor to view networks
            </CSDTypography>
          )}
          {tab === 2 && (
            <CSDTypography color="text.secondary">
              Connect to the hypervisor to view storage pools
            </CSDTypography>
          )}
          {tab === 3 && (
            <CSDTypography color="text.secondary">
              Connect to the hypervisor to view volumes
            </CSDTypography>
          )}
        </CSDBox>
      </CSDPaper>
    </CSDLayoutPage>
  );
};

export default HypervisorDetailPage;
