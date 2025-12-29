import React from 'react';
import {
  CSDCrudPage,
  CSDChip,
} from 'csd_core/UI';
import { useBreadcrumb, useNavigate } from 'csd_core/Providers';

const CLUSTERS_QUERY = `
  query Clusters($limit: Int, $offset: Int, $filter: ClusterFilter) {
    clusters(limit: $limit, offset: $offset, filter: $filter) {
      id
      name
      description
      status
      apiServerUrl
      createdAt
      updatedAt
    }
    clustersCount
  }
`;

const CREATE_CLUSTER = `
  mutation CreateCluster($input: CreateClusterInput!) {
    createCluster(input: $input) {
      id
    }
  }
`;

const UPDATE_CLUSTER = `
  mutation UpdateCluster($id: ID!, $input: UpdateClusterInput!) {
    updateCluster(id: $id, input: $input) {
      id
    }
  }
`;

const DELETE_CLUSTER = `
  mutation DeleteCluster($id: ID!) {
    deleteCluster(id: $id)
  }
`;

const statusColors: Record<string, 'success' | 'error' | 'warning' | 'default'> = {
  CONNECTED: 'success',
  DISCONNECTED: 'error',
  ERROR: 'error',
  PENDING: 'warning',
};

export const ClustersPage: React.FC = () => {
  const navigate = useNavigate();

  useBreadcrumb([
    { label: 'Pilote', path: '/pilote' },
    { label: 'Kubernetes', path: '/pilote/kubernetes' },
    { label: 'Clusters' },
  ]);

  const columns = [
    { field: 'name', headerName: 'Name', flex: 1 },
    { field: 'description', headerName: 'Description', flex: 1 },
    {
      field: 'status',
      headerName: 'Status',
      width: 120,
      renderCell: (params: { row: unknown; value: unknown }) => (
        <CSDChip
          label={params.value as string}
          color={statusColors[params.value as string] || 'default'}
          size="small"
        />
      ),
    },
    { field: 'apiServerUrl', headerName: 'API Server', flex: 1 },
  ];

  const formFields = [
    { name: 'name', label: 'Name', type: 'text', required: true },
    { name: 'description', label: 'Description', type: 'text', multiline: true, rows: 2 },
    { name: 'artifactKey', label: 'Kubeconfig Artifact Key', type: 'text', required: true, helperText: 'Key of the kubeconfig artifact in csd-core' },
  ];

  return (
    <CSDCrudPage
      title="Kubernetes Clusters"
      icon="cloud"
      entityName="cluster"
      entityNamePlural="clusters"
      columns={columns}
      formFields={formFields}
      queries={{
        list: CLUSTERS_QUERY,
        create: CREATE_CLUSTER,
        update: UPDATE_CLUSTER,
        delete: DELETE_CLUSTER,
      }}
      dataKeys={{
        list: 'clusters',
        count: 'clustersCount',
        create: 'createCluster',
        update: 'updateCluster',
        delete: 'deleteCluster',
      }}
      onRowClick={(row: { id: string }) => navigate(`/pilote/kubernetes/clusters/${row.id}`)}
      permissions={{
        create: 'csd-pilote.clusters.create',
        update: 'csd-pilote.clusters.update',
        delete: 'csd-pilote.clusters.delete',
      }}
    />
  );
};

export default ClustersPage;
