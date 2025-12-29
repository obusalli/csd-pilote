import React from 'react';
import {
  CSDLayoutPage,
  CSDBox,
  CSDTypography,
  CSDPaper,
  CSDChip,
  CSDButton,
  CSDDataGrid,
  CSDCircularProgress,
  CSDDialog,
  CSDDialogTitle,
  CSDDialogContent,
  CSDDialogActions,
  CSDTextField,
  CSDSelect,
  CSDMenuItem,
  CSDFormControl,
  CSDInputLabel,
  CSDToggleButtonGroup,
  CSDToggleButton,
  CSDAlert,
} from 'csd_core/UI';
import { useBreadcrumb, useNavigate, useGraphQL, useSnackbar } from 'csd_core/Providers';

const HYPERVISORS_QUERY = `
  query Hypervisors($limit: Int, $offset: Int, $filter: HypervisorFilter) {
    hypervisors(limit: $limit, offset: $offset, filter: $filter) {
      id
      name
      description
      mode
      driver
      uri
      status
      createdAt
    }
    hypervisorsCount
  }
`;

const LIBVIRT_AGENTS_QUERY = `
  query LibvirtAgents {
    libvirtAgents {
      id
      name
      hostname
      status
    }
  }
`;

const LIBVIRT_DEPLOY_AGENTS_QUERY = `
  query LibvirtDeployAgents($driver: String) {
    libvirtDeployAgents(driver: $driver) {
      id
      name
      hostname
      status
      supportedDrivers
    }
  }
`;

const LIBVIRT_DRIVERS_QUERY = `
  query LibvirtDrivers {
    libvirtDrivers {
      id
      name
      description
    }
  }
`;

const CREATE_HYPERVISOR = `
  mutation CreateHypervisor($input: CreateHypervisorInput!) {
    createHypervisor(input: $input) {
      id
    }
  }
`;

const DEPLOY_HYPERVISOR = `
  mutation DeployHypervisor($input: DeployHypervisorInput!) {
    deployHypervisor(input: $input) {
      id
    }
  }
`;

const DELETE_HYPERVISOR = `
  mutation DeleteHypervisor($id: ID!) {
    deleteHypervisor(id: $id)
  }
`;

const statusColors: Record<string, 'success' | 'error' | 'warning' | 'default' | 'info'> = {
  CONNECTED: 'success',
  DISCONNECTED: 'error',
  ERROR: 'error',
  PENDING: 'warning',
  DEPLOYING: 'info',
};

const modeColors: Record<string, 'primary' | 'secondary'> = {
  CONNECT: 'primary',
  DEPLOY: 'secondary',
};

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

interface Hypervisor {
  id: string;
  name: string;
  description: string;
  mode: string;
  driver: string;
  uri: string;
  status: string;
  createdAt: string;
}

export const HypervisorsPage: React.FC = () => {
  const navigate = useNavigate();
  const { showSuccess, showError } = useSnackbar();
  const { execute, loading: mutationLoading } = useGraphQL();

  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [mode, setMode] = React.useState<'CONNECT' | 'DEPLOY'>('CONNECT');

  // Form state - CONNECT mode
  const [connectForm, setConnectForm] = React.useState({
    name: '',
    description: '',
    agentId: '',
    uri: '',
    artifactKey: '',
  });

  // Form state - DEPLOY mode
  const [deployForm, setDeployForm] = React.useState({
    name: '',
    description: '',
    agentId: '',
    driver: '',
  });

  useBreadcrumb([
    { label: 'Pilote', path: '/pilote' },
    { label: 'Libvirt', path: '/pilote/libvirt' },
    { label: 'Hypervisors' },
  ]);

  // Fetch hypervisors
  const { data: hypervisorsData, loading: hypervisorsLoading, refetch } = useGraphQL<{
    hypervisors: Hypervisor[];
    hypervisorsCount: number;
  }>(HYPERVISORS_QUERY, { limit: 100, offset: 0 });

  // Fetch agents for CONNECT mode
  const { data: connectAgentsData, loading: connectAgentsLoading } = useGraphQL<{
    libvirtAgents: Agent[];
  }>(LIBVIRT_AGENTS_QUERY, {}, { skip: !dialogOpen || mode !== 'CONNECT' });

  // Fetch agents for DEPLOY mode (filtered by driver)
  const { data: deployAgentsData, loading: deployAgentsLoading } = useGraphQL<{
    libvirtDeployAgents: Agent[];
  }>(LIBVIRT_DEPLOY_AGENTS_QUERY, { driver: deployForm.driver?.toLowerCase() }, {
    skip: !dialogOpen || mode !== 'DEPLOY',
  });

  // Fetch drivers
  const { data: driversData } = useGraphQL<{
    libvirtDrivers: Driver[];
  }>(LIBVIRT_DRIVERS_QUERY, {}, { skip: !dialogOpen || mode !== 'DEPLOY' });

  const columns = [
    { field: 'name', headerName: 'Name', flex: 1 },
    {
      field: 'mode',
      headerName: 'Mode',
      width: 100,
      renderCell: (params: { value: string }) => (
        <CSDChip
          label={params.value}
          color={modeColors[params.value] || 'default'}
          size="small"
          variant="outlined"
        />
      ),
    },
    { field: 'driver', headerName: 'Driver', width: 100 },
    { field: 'uri', headerName: 'URI', flex: 1 },
    {
      field: 'status',
      headerName: 'Status',
      width: 130,
      renderCell: (params: { value: string }) => (
        <CSDChip
          label={params.value}
          color={statusColors[params.value] || 'default'}
          size="small"
        />
      ),
    },
    { field: 'description', headerName: 'Description', flex: 1 },
  ];

  const handleOpenDialog = () => {
    setDialogOpen(true);
    setMode('CONNECT');
    setConnectForm({ name: '', description: '', agentId: '', uri: '', artifactKey: '' });
    setDeployForm({ name: '', description: '', agentId: '', driver: '' });
  };

  const handleCloseDialog = () => {
    setDialogOpen(false);
  };

  const handleModeChange = (_: unknown, newMode: 'CONNECT' | 'DEPLOY' | null) => {
    if (newMode) setMode(newMode);
  };

  const handleSubmit = async () => {
    try {
      if (mode === 'CONNECT') {
        await execute(CREATE_HYPERVISOR, {
          input: {
            name: connectForm.name,
            description: connectForm.description,
            agentId: connectForm.agentId,
            uri: connectForm.uri,
            artifactKey: connectForm.artifactKey || undefined,
          },
        });
        showSuccess('Hypervisor created successfully');
      } else {
        await execute(DEPLOY_HYPERVISOR, {
          input: {
            name: deployForm.name,
            description: deployForm.description,
            agentId: deployForm.agentId,
            driver: deployForm.driver,
          },
        });
        showSuccess('Hypervisor deployment started');
      }
      handleCloseDialog();
      refetch();
    } catch (error) {
      showError(`Failed to ${mode === 'CONNECT' ? 'create' : 'deploy'} hypervisor: ${error}`);
    }
  };

  const isConnectFormValid = connectForm.name && connectForm.agentId && connectForm.uri;
  const isDeployFormValid = deployForm.name && deployForm.agentId && deployForm.driver;

  const connectAgents = connectAgentsData?.libvirtAgents || [];
  const deployAgents = deployAgentsData?.libvirtDeployAgents || [];
  const drivers = driversData?.libvirtDrivers || [];

  return (
    <CSDLayoutPage
      title="Libvirt Hypervisors"
      actions={
        <CSDButton variant="contained" color="primary" onClick={handleOpenDialog}>
          Add Hypervisor
        </CSDButton>
      }
    >
      <CSDPaper>
        {hypervisorsLoading ? (
          <CSDBox sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
            <CSDCircularProgress />
          </CSDBox>
        ) : (
          <CSDDataGrid
            rows={hypervisorsData?.hypervisors || []}
            columns={columns}
            autoHeight
            disableRowSelectionOnClick
            onRowClick={(params: { row: { id: string } }) => navigate(`/pilote/libvirt/hypervisors/${params.row.id}`)}
            sx={{ cursor: 'pointer' }}
          />
        )}
      </CSDPaper>

      <CSDDialog open={dialogOpen} onClose={handleCloseDialog} maxWidth="md" fullWidth>
        <CSDDialogTitle>Add Libvirt Hypervisor</CSDDialogTitle>
        <CSDDialogContent>
          <CSDBox sx={{ mb: 3, mt: 1 }}>
            <CSDToggleButtonGroup
              value={mode}
              exclusive
              onChange={handleModeChange}
              fullWidth
            >
              <CSDToggleButton value="CONNECT">
                Connect to Existing Hypervisor
              </CSDToggleButton>
              <CSDToggleButton value="DEPLOY">
                Deploy Libvirt on Agent
              </CSDToggleButton>
            </CSDToggleButtonGroup>
          </CSDBox>

          {mode === 'CONNECT' ? (
            <CSDBox sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              <CSDAlert severity="info">
                Connect to an existing libvirt hypervisor using a connection URI.
              </CSDAlert>

              <CSDTextField
                label="Name"
                value={connectForm.name}
                onChange={(e) => setConnectForm({ ...connectForm, name: e.target.value })}
                required
                fullWidth
              />

              <CSDTextField
                label="Description"
                value={connectForm.description}
                onChange={(e) => setConnectForm({ ...connectForm, description: e.target.value })}
                multiline
                rows={2}
                fullWidth
              />

              <CSDFormControl fullWidth required>
                <CSDInputLabel>Agent</CSDInputLabel>
                <CSDSelect
                  value={connectForm.agentId}
                  onChange={(e) => setConnectForm({ ...connectForm, agentId: e.target.value as string })}
                  label="Agent"
                  disabled={connectAgentsLoading}
                >
                  {connectAgents.map((agent) => (
                    <CSDMenuItem key={agent.id} value={agent.id}>
                      {agent.name} ({agent.hostname})
                    </CSDMenuItem>
                  ))}
                </CSDSelect>
              </CSDFormControl>

              <CSDTextField
                label="Libvirt URI"
                value={connectForm.uri}
                onChange={(e) => setConnectForm({ ...connectForm, uri: e.target.value })}
                required
                fullWidth
                helperText="e.g., qemu:///system, qemu+ssh://user@host/system"
              />

              <CSDTextField
                label="SSH Key Artifact (Optional)"
                value={connectForm.artifactKey}
                onChange={(e) => setConnectForm({ ...connectForm, artifactKey: e.target.value })}
                fullWidth
                helperText="Key of the SSH key artifact for remote connections"
              />
            </CSDBox>
          ) : (
            <CSDBox sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              <CSDAlert severity="info">
                Deploy libvirt and the selected driver on an agent. The hypervisor will be configured automatically.
              </CSDAlert>

              <CSDTextField
                label="Name"
                value={deployForm.name}
                onChange={(e) => setDeployForm({ ...deployForm, name: e.target.value })}
                required
                fullWidth
              />

              <CSDTextField
                label="Description"
                value={deployForm.description}
                onChange={(e) => setDeployForm({ ...deployForm, description: e.target.value })}
                multiline
                rows={2}
                fullWidth
              />

              <CSDFormControl fullWidth required>
                <CSDInputLabel>Driver</CSDInputLabel>
                <CSDSelect
                  value={deployForm.driver}
                  onChange={(e) => setDeployForm({
                    ...deployForm,
                    driver: e.target.value as string,
                    agentId: '',
                  })}
                  label="Driver"
                >
                  {drivers.map((driver) => (
                    <CSDMenuItem key={driver.id} value={driver.id}>
                      {driver.name} - {driver.description}
                    </CSDMenuItem>
                  ))}
                </CSDSelect>
              </CSDFormControl>

              <CSDFormControl fullWidth required>
                <CSDInputLabel>Agent</CSDInputLabel>
                <CSDSelect
                  value={deployForm.agentId}
                  onChange={(e) => setDeployForm({ ...deployForm, agentId: e.target.value as string })}
                  label="Agent"
                  disabled={deployAgentsLoading || !deployForm.driver}
                >
                  {deployAgents.length === 0 && deployForm.driver ? (
                    <CSDMenuItem disabled>
                      No agents support {deployForm.driver} deployment
                    </CSDMenuItem>
                  ) : (
                    deployAgents.map((agent) => (
                      <CSDMenuItem key={agent.id} value={agent.id}>
                        {agent.name} ({agent.hostname})
                      </CSDMenuItem>
                    ))
                  )}
                </CSDSelect>
              </CSDFormControl>

              {deployForm.driver && deployAgents.length === 0 && !deployAgentsLoading && (
                <CSDAlert severity="warning">
                  No agents support {deployForm.driver} deployment. Select a different driver or configure agent capabilities.
                </CSDAlert>
              )}
            </CSDBox>
          )}
        </CSDDialogContent>
        <CSDDialogActions>
          <CSDButton onClick={handleCloseDialog}>Cancel</CSDButton>
          <CSDButton
            variant="contained"
            color="primary"
            onClick={handleSubmit}
            disabled={mutationLoading || (mode === 'CONNECT' ? !isConnectFormValid : !isDeployFormValid)}
          >
            {mutationLoading ? 'Processing...' : mode === 'CONNECT' ? 'Connect' : 'Deploy'}
          </CSDButton>
        </CSDDialogActions>
      </CSDDialog>
    </CSDLayoutPage>
  );
};

export default HypervisorsPage;
