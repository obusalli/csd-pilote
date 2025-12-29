import React from 'react';
import {
  CSDLayoutPage,
  CSDBox,
  CSDPaper,
  CSDChip,
  CSDButton,
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
  CSDTable,
  CSDTableHead,
  CSDTableBody,
  CSDTableRow,
  CSDTableCell,
} from 'csd_core/UI';
import { useBreadcrumb, useNavigate, useGraphQL } from 'csd_core/Providers';

const CLUSTERS_QUERY = `
  query Clusters($limit: Int, $offset: Int) {
    clusters(limit: $limit, offset: $offset) {
      id
      name
      description
      mode
      distribution
      status
      createdAt
    }
    clustersCount
  }
`;

const KUBERNETES_AGENTS_QUERY = `
  query KubernetesAgents {
    kubernetesAgents {
      id
      name
      hostname
      status
    }
  }
`;

const ALL_DISTRIBUTIONS_QUERY = `
  query AllKubernetesDistributions {
    allKubernetesDistributions {
      id
      name
      description
      deployable
    }
  }
`;

const DEPLOYABLE_DISTRIBUTIONS_QUERY = `
  query KubernetesDistributions {
    kubernetesDistributions {
      id
      name
      description
    }
  }
`;

const CREATE_CLUSTER = `
  mutation CreateCluster($input: CreateClusterInput!) {
    createCluster(input: $input) {
      id
    }
  }
`;

const DEPLOY_CLUSTER = `
  mutation DeployCluster($input: DeployClusterInput!) {
    deployCluster(input: $input) {
      id
    }
  }
`;

const statusColors: Record<string, 'success' | 'error' | 'warning' | 'default' | 'info'> = {
  CONNECTED: 'success',
  DISCONNECTED: 'error',
  ERROR: 'error',
  PENDING: 'warning',
  DEPLOYING: 'info',
};

interface Agent {
  id: string;
  name: string;
  hostname: string;
  status: string;
}

interface Distribution {
  id: string;
  name: string;
  description: string;
  deployable?: boolean;
}

interface Cluster {
  id: string;
  name: string;
  description: string;
  mode: string;
  distribution: string;
  status: string;
  createdAt: string;
}

export const ClustersPage: React.FC = () => {
  const navigate = useNavigate();
  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [mode, setMode] = React.useState<'CONNECT' | 'DEPLOY'>('CONNECT');
  const [submitting, setSubmitting] = React.useState(false);

  // Form state - CONNECT mode
  const [connectForm, setConnectForm] = React.useState({
    name: '',
    description: '',
    agentId: '',
    artifactKey: '',
    distribution: '',
  });

  // Form state - DEPLOY mode
  const [deployForm, setDeployForm] = React.useState({
    name: '',
    description: '',
    distribution: '',
    version: '',
    masterNodes: '',
    workerNodes: '',
  });

  useBreadcrumb([
    { label: 'Pilote', path: '/pilote' },
    { label: 'Kubernetes', path: '/pilote/kubernetes' },
    { label: 'Clusters' },
  ]);

  // Fetch clusters
  const clustersResult = useGraphQL(CLUSTERS_QUERY, { limit: 100, offset: 0 });
  const clustersData = clustersResult.data as { clusters: Cluster[]; clustersCount: number } | undefined;
  const clustersLoading = clustersResult.loading;

  // Fetch agents for CONNECT mode
  const connectAgentsResult = useGraphQL(KUBERNETES_AGENTS_QUERY, {});
  const connectAgentsData = connectAgentsResult.data as { kubernetesAgents: Agent[] } | undefined;

  // Fetch all distributions
  const allDistributionsResult = useGraphQL(ALL_DISTRIBUTIONS_QUERY, {});
  const allDistributionsData = allDistributionsResult.data as { allKubernetesDistributions: Distribution[] } | undefined;

  // Fetch deployable distributions
  const deployableDistributionsResult = useGraphQL(DEPLOYABLE_DISTRIBUTIONS_QUERY, {});
  const deployableDistributionsData = deployableDistributionsResult.data as { kubernetesDistributions: Distribution[] } | undefined;

  const handleOpenDialog = () => {
    setDialogOpen(true);
    setMode('CONNECT');
    setConnectForm({ name: '', description: '', agentId: '', artifactKey: '', distribution: '' });
    setDeployForm({ name: '', description: '', distribution: '', version: '', masterNodes: '', workerNodes: '' });
  };

  const handleCloseDialog = () => {
    setDialogOpen(false);
  };

  const handleModeChange = (_: unknown, newMode: 'CONNECT' | 'DEPLOY' | null) => {
    if (newMode) setMode(newMode);
  };

  const handleSubmit = async () => {
    setSubmitting(true);
    try {
      if (mode === 'CONNECT') {
        await clustersResult.refetch({
          query: CREATE_CLUSTER,
          variables: {
            input: {
              name: connectForm.name,
              description: connectForm.description,
              agentId: connectForm.agentId,
              artifactKey: connectForm.artifactKey,
              distribution: connectForm.distribution || undefined,
            },
          },
        });
      } else {
        await clustersResult.refetch({
          query: DEPLOY_CLUSTER,
          variables: {
            input: {
              name: deployForm.name,
              description: deployForm.description,
              distribution: deployForm.distribution,
              version: deployForm.version || undefined,
              masterNodes: deployForm.masterNodes.split(',').map(s => s.trim()).filter(Boolean),
              workerNodes: deployForm.workerNodes ? deployForm.workerNodes.split(',').map(s => s.trim()).filter(Boolean) : [],
            },
          },
        });
      }
      handleCloseDialog();
      clustersResult.refetch();
    } catch (error) {
      console.error('Failed:', error);
    } finally {
      setSubmitting(false);
    }
  };

  const isConnectFormValid = connectForm.name && connectForm.agentId && connectForm.artifactKey;
  const isDeployFormValid = deployForm.name && deployForm.distribution && deployForm.masterNodes;

  const connectAgents = connectAgentsData?.kubernetesAgents || [];
  const allDistributions = allDistributionsData?.allKubernetesDistributions || [];
  const deployableDistributions = deployableDistributionsData?.kubernetesDistributions || [];

  return (
    <CSDLayoutPage
      title="Kubernetes Clusters"
      actions={
        <CSDButton variant="contained" color="primary" onClick={handleOpenDialog}>
          Add Cluster
        </CSDButton>
      }
    >
      <CSDPaper>
        {clustersLoading ? (
          <CSDBox sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
            <CSDCircularProgress />
          </CSDBox>
        ) : (
          <CSDTable>
            <CSDTableHead>
              <CSDTableRow>
                <CSDTableCell>Name</CSDTableCell>
                <CSDTableCell>Mode</CSDTableCell>
                <CSDTableCell>Distribution</CSDTableCell>
                <CSDTableCell>Status</CSDTableCell>
                <CSDTableCell>Description</CSDTableCell>
              </CSDTableRow>
            </CSDTableHead>
            <CSDTableBody>
              {(clustersData?.clusters || []).map((cluster) => (
                <CSDTableRow
                  key={cluster.id}
                  hover
                  onClick={() => navigate(`/pilote/kubernetes/clusters/${cluster.id}`)}
                  sx={{ cursor: 'pointer' }}
                >
                  <CSDTableCell>{cluster.name}</CSDTableCell>
                  <CSDTableCell>
                    <CSDChip label={cluster.mode} color={cluster.mode === 'DEPLOY' ? 'secondary' : 'primary'} size="small" />
                  </CSDTableCell>
                  <CSDTableCell>{cluster.distribution || '-'}</CSDTableCell>
                  <CSDTableCell>
                    <CSDChip label={cluster.status} color={statusColors[cluster.status] || 'default'} size="small" />
                  </CSDTableCell>
                  <CSDTableCell>{cluster.description || '-'}</CSDTableCell>
                </CSDTableRow>
              ))}
            </CSDTableBody>
          </CSDTable>
        )}
      </CSDPaper>

      <CSDDialog open={dialogOpen} onClose={handleCloseDialog} maxWidth="md" fullWidth>
        <CSDDialogTitle>Add Kubernetes Cluster</CSDDialogTitle>
        <CSDDialogContent>
          <CSDBox sx={{ mb: 3, mt: 1 }}>
            <CSDToggleButtonGroup
              value={mode}
              exclusive
              onChange={handleModeChange}
              fullWidth
            >
              <CSDToggleButton value="CONNECT">
                Connect to Existing
              </CSDToggleButton>
              <CSDToggleButton value="DEPLOY">
                Deploy New
              </CSDToggleButton>
            </CSDToggleButtonGroup>
          </CSDBox>

          {mode === 'CONNECT' ? (
            <CSDBox sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              <CSDAlert severity="info">
                Connect to an existing Kubernetes cluster using a kubeconfig artifact.
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
                >
                  {connectAgents.map((agent) => (
                    <CSDMenuItem key={agent.id} value={agent.id}>
                      {agent.name} ({agent.hostname})
                    </CSDMenuItem>
                  ))}
                </CSDSelect>
              </CSDFormControl>

              <CSDTextField
                label="Kubeconfig Artifact Key"
                value={connectForm.artifactKey}
                onChange={(e) => setConnectForm({ ...connectForm, artifactKey: e.target.value })}
                required
                fullWidth
                helperText="Key of the kubeconfig artifact in csd-core"
              />

              <CSDFormControl fullWidth>
                <CSDInputLabel>Distribution (Optional)</CSDInputLabel>
                <CSDSelect
                  value={connectForm.distribution}
                  onChange={(e) => setConnectForm({ ...connectForm, distribution: e.target.value as string })}
                  label="Distribution (Optional)"
                >
                  <CSDMenuItem value="">Not specified</CSDMenuItem>
                  {allDistributions.map((dist) => (
                    <CSDMenuItem key={dist.id} value={dist.id}>
                      {dist.name}
                    </CSDMenuItem>
                  ))}
                </CSDSelect>
              </CSDFormControl>
            </CSDBox>
          ) : (
            <CSDBox sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              <CSDAlert severity="info">
                Deploy a new Kubernetes cluster. Enter agent IDs separated by commas.
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
                <CSDInputLabel>Distribution</CSDInputLabel>
                <CSDSelect
                  value={deployForm.distribution}
                  onChange={(e) => setDeployForm({ ...deployForm, distribution: e.target.value as string })}
                  label="Distribution"
                >
                  {deployableDistributions.map((dist) => (
                    <CSDMenuItem key={dist.id} value={dist.id}>
                      {dist.name} - {dist.description}
                    </CSDMenuItem>
                  ))}
                </CSDSelect>
              </CSDFormControl>

              <CSDTextField
                label="Version (Optional)"
                value={deployForm.version}
                onChange={(e) => setDeployForm({ ...deployForm, version: e.target.value })}
                fullWidth
                helperText="Leave empty for latest version"
              />

              <CSDTextField
                label="Master Node Agent IDs"
                value={deployForm.masterNodes}
                onChange={(e) => setDeployForm({ ...deployForm, masterNodes: e.target.value })}
                required
                fullWidth
                helperText="Comma-separated agent IDs for master nodes"
              />

              <CSDTextField
                label="Worker Node Agent IDs (Optional)"
                value={deployForm.workerNodes}
                onChange={(e) => setDeployForm({ ...deployForm, workerNodes: e.target.value })}
                fullWidth
                helperText="Comma-separated agent IDs for worker nodes"
              />
            </CSDBox>
          )}
        </CSDDialogContent>
        <CSDDialogActions>
          <CSDButton onClick={handleCloseDialog}>Cancel</CSDButton>
          <CSDButton
            variant="contained"
            color="primary"
            onClick={handleSubmit}
            disabled={submitting || (mode === 'CONNECT' ? !isConnectFormValid : !isDeployFormValid)}
          >
            {submitting ? 'Processing...' : mode === 'CONNECT' ? 'Connect' : 'Deploy'}
          </CSDButton>
        </CSDDialogActions>
      </CSDDialog>
    </CSDLayoutPage>
  );
};

export default ClustersPage;
