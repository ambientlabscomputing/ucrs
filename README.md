# Underleaf Capability Registry Service (UCRS)

A cloud-native microservice providing the global semantic authority and provider index for Underleaf Agents. UCRS manages capability schemas, provider metadata, trust tier classification, and distributes signed registry snapshots.

## Features

- **Capability Schema Governance**: Manage semantic capability definitions with versioning and risk classification
- **Provider Registry**: Index capability providers with artifact locations and trust tiers
- **Signed Snapshots**: Ed25519-signed registry snapshots for cryptographic integrity
- **Delta Sync**: Efficient incremental updates for agents
- **ETag Caching**: HTTP caching support for bandwidth optimization
- **Trust Tiers**: Classification system (official, certified, community, experimental, local)
- **Multi-Artifact Support**: OCI containers, binaries, Git repos, npm, PyPI packages
- **Automatic Seeding**: Git-managed seed data for capabilities and providers (dev/prod environments)

## Seeding System

UCRS automatically loads baseline capability and provider definitions from YAML files on startup. This allows you to:

- Manage registry data through version control
- Maintain separate seed data for dev and prod environments
- Automatically update existing entries or create new ones
- Ensure consistent baseline data across deployments

**Included Seeds:**
- **Capabilities**: `mesh.service.discovery`, `mesh.event.routing`
- **Providers**: `underleaf.mma` (Mycelium Mesh Agent)

See [SEEDING_GUIDE.md](SEEDING_GUIDE.md) for detailed documentation.

## Architecture

```
┌─────────────────┐
│ Underleaf Agent │
└────────┬────────┘
         │ Sync/Query
         ▼
┌─────────────────┐      ┌──────────┐
│      UCRS       │ ───► │ MongoDB  │
│  (Registry API) │      └──────────┘
└─────────────────┘
         │
         ▼
   Ed25519 Signing
```

## Prerequisites

- **Go** 1.24+
- **MongoDB** 7.0+
- **Auth0 Account** (for JWT authentication)
- **Docker** (optional, for containerized deployment)

## Quick Start

### 1. Clone and Setup

```bash
cd capability_registry_service
cp config.yaml.example config.yaml
```

### 2. Generate Ed25519 Signing Keys

```bash
make generate-keys
```

This creates:
- `keys/registry_signing_key.pem` (private key)
- `keys/registry_signing_key.pub` (public key)

### 3. Configure

Edit `config.yaml` with your Auth0 and MongoDB credentials.

### 4. Run Locally

```bash
# Start MongoDB
docker-compose up -d mongo

# Run service
make run
```

### 5. Or Run with Docker

```bash
docker-compose up -d
```

## API Documentation

Once running, visit:
```
http://localhost:8083/swagger/index.html
```

### Generate Documentation

```bash
make docs
```

## API Endpoints

### Sync Endpoints (Read)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/registry/sync/snapshot` | Get complete registry snapshot (with ETag) |
| `GET` | `/api/v1/registry/sync/delta?since=<version>` | Get incremental changes |

### Capability Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/api/v1/registry/capabilities` | List capabilities | `read:registry` |
| `GET` | `/api/v1/registry/capabilities/:id` | Get capability by ID | `read:registry` |
| `POST` | `/api/v1/registry/capabilities` | Create capability | `admin:registry` |
| `PATCH` | `/api/v1/registry/capabilities/:id` | Update capability | `admin:registry` |
| `DELETE` | `/api/v1/registry/capabilities/:id` | Delete capability | `admin:registry` |

### Provider Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/api/v1/registry/providers` | List providers | `read:registry` |
| `GET` | `/api/v1/registry/providers/:id/:version` | Get provider | `read:registry` |
| `POST` | `/api/v1/registry/providers` | Register provider | `write:registry` |
| `PATCH` | `/api/v1/registry/providers/:id/:version` | Update provider | `write:registry` |
| `DELETE` | `/api/v1/registry/providers/:id/:version` | Delete provider | `admin:registry` |

## Example Usage

### Create a Capability

```bash
curl -X POST http://localhost:8083/api/v1/registry/capabilities \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "iot.light.control",
    "version": "1.0.0",
    "description": "Generic lighting control capability",
    "risk_class": "medium",
    "allowed_verbs": ["on", "off", "set_brightness"],
    "permission_class": "device.control"
  }'
```

### Register a Provider

```bash
curl -X POST http://localhost:8083/api/v1/registry/providers \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "ambient.hue-mcp",
    "version": "2.1.0",
    "capabilities": [
      {"id": "iot.light.control", "version_range": "^1.0"}
    ],
    "artifact": {
      "type": "oci",
      "uri": "registry.underleaf.io/ambient/hue-mcp:2.1.0",
      "digest": "sha256:abc123..."
    },
    "trust_tier": "official",
    "sandbox_profile": "standard"
  }'
```

## Project Structure

```
capability_registry_service/
├── service/                    # Main application code
│   ├── main.go                 # Application entry point
│   ├── errors/                 # Structured error types
│   ├── repository/             # Data access layer (MongoDB)
│   ├── router/                 # HTTP handlers + middleware
│   ├── service/                # Business logic layer
│   ├── types/                  # Domain models + API types
│   ├── utils/                  # Auth, logging, settings
│   └── scripts/                # Utility scripts
├── config.yaml.example         # Configuration template
├── Makefile                    # Build commands
├── Dockerfile                  # Container image definition
└── docker-compose.yml          # Local development stack
```

## Security

All endpoints require JWT bearer tokens with appropriate scopes. Registry snapshots are signed with Ed25519 for cryptographic integrity verification.

## License

Copyright © 2026 Ambient Labs Computing. All rights reserved.
