# CSD Pilote - Infrastructure Management Application

## Overview

CSD Pilote is a business application built on top of csd-core for managing infrastructure:
- **Kubernetes clusters** - Manage K8s clusters, namespaces, deployments, pods, services
- **Libvirt hypervisors** - Manage KVM/QEMU VMs, networks, storage pools
- **Container engines** - Manage Docker/Podman containers, images, networks, volumes

## Architecture

```
csd-pilote/
├── backend/           # Go GraphQL API (port 9092)
├── frontend/          # React Module Federation remote (port 4042)
├── cli/               # csd-pilotectl CLI tool
└── docs/              # Documentation
```

### Integration with csd-core

- **Authentication**: Uses JWT tokens from csd-core (shared secret)
- **Credentials**: Stores kubeconfig, SSH keys via csd-core artifacts
- **Operations**: Executes via csd-core playbooks (kubernetes, libvirt, docker types)
- **Permissions**: Registered in csd-core permission system
- **Frontend**: Module Federation remote loaded by csd-core host

## Backend

### Tech Stack
- Go 1.25
- GraphQL (custom implementation)
- PostgreSQL with GORM
- JWT authentication

### Modules

```
backend/modules/
├── platform/              # Infrastructure (config, database, graphql, middleware)
└── pilot/                 # Business logic
    ├── clusters/          # Kubernetes cluster configurations
    ├── hypervisors/       # Libvirt hypervisor configurations
    └── containers/        # Docker/Podman engine configurations
```

### Database Schema: `csd_pilote`

**Tables:**
- `clusters` - Kubernetes cluster configs (name, artifactKey for kubeconfig, status)
- `hypervisors` - Libvirt hypervisor configs (name, URI, artifactKey for SSH key, status)
- `container_engines` - Docker/Podman configs (name, host, engineType, status)

### Running Backend

```bash
cd backend
go run csd-pilotd.go
# or
go build -o csd-pilotd && ./csd-pilotd
```

## Frontend

### Tech Stack
- React 19
- TypeScript 5.9
- Vite 6 with Module Federation
- MUI 7 (from csd-core)
- Apollo Client (from csd-core)

### Structure

```
frontend/src/
├── Routes.tsx             # Exposed module for csd-core
├── appInfo.ts             # Exposed module for About dialog
├── modules/pilot/
│   ├── dashboard/         # Dashboard page
│   ├── clusters/          # Kubernetes pages
│   ├── hypervisors/       # Libvirt pages
│   └── containers/        # Container pages
└── translations/          # i18n
```

### Running Frontend

```bash
cd frontend
npm install
npm start
```

Access via csd-core at http://localhost:4040/pilote

## CLI

### Commands

```bash
# Database management
csd-pilotectl database init      # Initialize schema
csd-pilotectl database update    # Run migrations
csd-pilotectl database status    # Show status
csd-pilotectl database reset     # Reset database

# Seed data
csd-pilotectl seed core          # Seed permissions, menus to csd-core
csd-pilotectl seed app           # Seed application data
csd-pilotectl seed all           # Seed everything
```

### Building CLI

```bash
cd cli
go build -o csd-pilotectl
```

## Configuration

### Backend (backend/csd-pilote.yaml)

```yaml
common:
  database:
    url: "postgres://postgres@localhost:5432/csd_pilote?sslmode=prefer"
  csd-core:
    url: "http://127.0.0.1:9090"

backend:
  server:
    port: "9092"
  jwt:
    secret: "same-as-csd-core"
```

### Frontend (frontend/csd-pilote.yaml)

```yaml
frontend:
  url: "http://127.0.0.1:4042"
  route-path: "/pilote"
  dev:
    port: 4042
```

## Playbook Integration

Operations are executed via csd-core playbooks:

### Kubernetes Operations
- `pod_list`, `pod_get`, `pod_delete`, `pod_logs`, `pod_exec`
- `deployment_list`, `deployment_create`, `deployment_scale`
- `service_list`, `service_create`, `service_delete`
- `namespace_list`, `namespace_create`

### Libvirt Operations
- `vm_list`, `vm_start`, `vm_stop`, `vm_define`, `vm_undefine`
- `network_list`, `network_define`, `network_start`
- `pool_list`, `pool_define`, `volume_list`, `volume_create`

### Docker Operations
- `container_list`, `container_start`, `container_stop`
- `image_list`, `image_pull`, `image_remove`
- `network_list`, `volume_list`

## Permissions

Registered in csd-core:
- `csd-pilote.clusters.*` - Kubernetes cluster management
- `csd-pilote.namespaces.*`, `csd-pilote.deployments.*`, etc.
- `csd-pilote.hypervisors.*` - Libvirt management
- `csd-pilote.domains.*` - VM management
- `csd-pilote.containers.*` - Container engine management

## WebSocket/Console Support

For interactive consoles (kubectl exec, VM console, docker exec), use csd-core's terminal WebSocket infrastructure.

## Development Setup

1. Start csd-core (backend + frontend)
2. Create database: `createdb csd_pilote`
3. Initialize schema: `csd-pilotectl database init`
4. Seed permissions: `csd-pilotectl seed core`
5. Start backend: `cd backend && go run csd-pilotd.go`
6. Start frontend: `cd frontend && npm start`
7. Access at http://localhost:4040/pilote
