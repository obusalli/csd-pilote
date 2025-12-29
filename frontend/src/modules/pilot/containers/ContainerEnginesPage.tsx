import React from 'react';
import {
  CSDCrudPage,
  CSDChip,
} from 'csd_core/UI';
import { useBreadcrumb, useNavigate } from 'csd_core/Providers';

const ENGINES_QUERY = `
  query ContainerEngines($limit: Int, $offset: Int, $filter: ContainerEngineFilter) {
    containerEngines(limit: $limit, offset: $offset, filter: $filter) {
      id
      name
      description
      engineType
      host
      status
      engineVersion
      containerCount
      imageCount
      createdAt
    }
    containerEnginesCount
  }
`;

const CREATE_ENGINE = `
  mutation CreateContainerEngine($input: CreateContainerEngineInput!) {
    createContainerEngine(input: $input) {
      id
    }
  }
`;

const UPDATE_ENGINE = `
  mutation UpdateContainerEngine($id: ID!, $input: UpdateContainerEngineInput!) {
    updateContainerEngine(id: $id, input: $input) {
      id
    }
  }
`;

const DELETE_ENGINE = `
  mutation DeleteContainerEngine($id: ID!) {
    deleteContainerEngine(id: $id)
  }
`;

const statusColors: Record<string, 'success' | 'error' | 'warning' | 'default'> = {
  CONNECTED: 'success',
  DISCONNECTED: 'error',
  ERROR: 'error',
  PENDING: 'warning',
};

const engineTypeColors: Record<string, 'primary' | 'secondary'> = {
  DOCKER: 'primary',
  PODMAN: 'secondary',
};

export const ContainerEnginesPage: React.FC = () => {
  const navigate = useNavigate();

  useBreadcrumb([
    { label: 'Pilote', path: '/pilote' },
    { label: 'Containers', path: '/pilote/containers' },
    { label: 'Engines' },
  ]);

  const columns = [
    { field: 'name', headerName: 'Name', flex: 1 },
    {
      field: 'engineType',
      headerName: 'Type',
      width: 100,
      renderCell: (params: { row: unknown; value: unknown }) => (
        <CSDChip
          label={params.value as string}
          color={engineTypeColors[params.value as string] || 'default'}
          size="small"
        />
      ),
    },
    { field: 'host', headerName: 'Host', flex: 1 },
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
    { field: 'engineVersion', headerName: 'Version', width: 120 },
    { field: 'containerCount', headerName: 'Containers', width: 100 },
    { field: 'imageCount', headerName: 'Images', width: 100 },
  ];

  const formFields = [
    { name: 'name', label: 'Name', type: 'text', required: true },
    { name: 'description', label: 'Description', type: 'text', multiline: true, rows: 2 },
    {
      name: 'engineType',
      label: 'Engine Type',
      type: 'select',
      required: true,
      options: [
        { value: 'DOCKER', label: 'Docker' },
        { value: 'PODMAN', label: 'Podman' },
      ],
    },
    { name: 'host', label: 'Host', type: 'text', required: true, helperText: 'e.g., unix:///var/run/docker.sock or tcp://host:2375' },
    { name: 'artifactKey', label: 'TLS Certs Artifact', type: 'text', helperText: 'Optional: Key of the TLS certificates artifact' },
  ];

  return (
    <CSDCrudPage
      title="Container Engines"
      icon="view_in_ar"
      entityName="container engine"
      entityNamePlural="container engines"
      columns={columns}
      formFields={formFields}
      queries={{
        list: ENGINES_QUERY,
        create: CREATE_ENGINE,
        update: UPDATE_ENGINE,
        delete: DELETE_ENGINE,
      }}
      dataKeys={{
        list: 'containerEngines',
        count: 'containerEnginesCount',
        create: 'createContainerEngine',
        update: 'updateContainerEngine',
        delete: 'deleteContainerEngine',
      }}
      onRowClick={(row: { id: string }) => navigate(`/pilote/containers/engines/${row.id}`)}
      permissions={{
        create: 'csd-pilote.containers.create',
        update: 'csd-pilote.containers.update',
        delete: 'csd-pilote.containers.delete',
      }}
    />
  );
};

export default ContainerEnginesPage;
