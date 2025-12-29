import React from 'react';
import {
  CSDCrudPage,
  CSDChip,
} from 'csd_core/UI';
import { useBreadcrumb, useNavigate } from 'csd_core/Providers';

const HYPERVISORS_QUERY = `
  query Hypervisors($limit: Int, $offset: Int, $filter: HypervisorFilter) {
    hypervisors(limit: $limit, offset: $offset, filter: $filter) {
      id
      name
      description
      uri
      status
      hostname
      hypervisorType
      totalMemoryMb
      totalCpus
      createdAt
    }
    hypervisorsCount
  }
`;

const CREATE_HYPERVISOR = `
  mutation CreateHypervisor($input: CreateHypervisorInput!) {
    createHypervisor(input: $input) {
      id
    }
  }
`;

const UPDATE_HYPERVISOR = `
  mutation UpdateHypervisor($id: ID!, $input: UpdateHypervisorInput!) {
    updateHypervisor(id: $id, input: $input) {
      id
    }
  }
`;

const DELETE_HYPERVISOR = `
  mutation DeleteHypervisor($id: ID!) {
    deleteHypervisor(id: $id)
  }
`;

const statusColors: Record<string, 'success' | 'error' | 'warning' | 'default'> = {
  CONNECTED: 'success',
  DISCONNECTED: 'error',
  ERROR: 'error',
  PENDING: 'warning',
};

export const HypervisorsPage: React.FC = () => {
  const navigate = useNavigate();

  useBreadcrumb([
    { label: 'Pilote', path: '/pilote' },
    { label: 'Libvirt', path: '/pilote/libvirt' },
    { label: 'Hypervisors' },
  ]);

  const columns = [
    { field: 'name', headerName: 'Name', flex: 1 },
    { field: 'hostname', headerName: 'Hostname', flex: 1 },
    { field: 'uri', headerName: 'URI', flex: 1 },
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
    { field: 'hypervisorType', headerName: 'Type', width: 100 },
    { field: 'totalCpus', headerName: 'CPUs', width: 80 },
    {
      field: 'totalMemoryMb',
      headerName: 'Memory',
      width: 100,
      valueFormatter: (params: { value: unknown }) =>
        params.value ? `${Math.round((params.value as number) / 1024)} GB` : '-',
    },
  ];

  const formFields = [
    { name: 'name', label: 'Name', type: 'text', required: true },
    { name: 'description', label: 'Description', type: 'text', multiline: true, rows: 2 },
    { name: 'uri', label: 'Libvirt URI', type: 'text', required: true, helperText: 'e.g., qemu+ssh://user@host/system' },
    { name: 'artifactKey', label: 'SSH Key Artifact', type: 'text', helperText: 'Optional: Key of the SSH key artifact' },
  ];

  return (
    <CSDCrudPage
      title="Libvirt Hypervisors"
      icon="dns"
      entityName="hypervisor"
      entityNamePlural="hypervisors"
      columns={columns}
      formFields={formFields}
      queries={{
        list: HYPERVISORS_QUERY,
        create: CREATE_HYPERVISOR,
        update: UPDATE_HYPERVISOR,
        delete: DELETE_HYPERVISOR,
      }}
      dataKeys={{
        list: 'hypervisors',
        count: 'hypervisorsCount',
        create: 'createHypervisor',
        update: 'updateHypervisor',
        delete: 'deleteHypervisor',
      }}
      onRowClick={(row: { id: string }) => navigate(`/pilote/libvirt/hypervisors/${row.id}`)}
      permissions={{
        create: 'csd-pilote.hypervisors.create',
        update: 'csd-pilote.hypervisors.update',
        delete: 'csd-pilote.hypervisors.delete',
      }}
    />
  );
};

export default HypervisorsPage;
