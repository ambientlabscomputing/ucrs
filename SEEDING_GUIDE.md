# UCRS Seeding System

## Overview

The UCRS seeding system automatically loads Provider and Capability definitions from YAML files on service startup. This allows you to manage registry data through version control and ensures consistent baseline data across environments.

## Features

- **Environment-Specific Seeds**: Separate seed data for `dev` and `prod` environments
- **Automatic Upserts**: Seeds automatically create or update existing entries
- **Git-Managed**: Seed files are version-controlled, allowing easy updates
- **Startup Integration**: Seeds load automatically when UCRS starts

## Directory Structure

```
ucrs/service/seeds/
├── README.md
├── dev/
│   ├── capabilities.yaml
│   └── providers.yaml
└── prod/
    ├── capabilities.yaml
    └── providers.yaml
```

## Configuration

Set the environment in `config.yaml`:

```yaml
environment: "dev"  # Options: "dev" or "prod"
```

The default is `dev` if not specified.

## Seed File Format

### Capabilities (`capabilities.yaml`)

```yaml
capabilities:
  - id: mesh.service.discovery
    version: 1.0.0
    description: Service discovery and registration in mesh network
    risk_class: low
    allowed_verbs:
      - register
      - discover
      - query
    permission_class: mesh.discovery
```

### Providers (`providers.yaml`)

```yaml
providers:
  - provider_id: underleaf.mma
    version: dev  # "dev" in dev environment, "1.0.0" in prod
    capabilities:
      - id: mesh.event.routing
        version_range: ^1.0
    artifact:
      type: binary
      uri: https://github.com/ambientlabscomputing/mycelium_mesh_agent/releases/download/{version}/mycelium-mesh-agent-{os}-{arch}
      digest: ""
      checksums:
        uri: https://github.com/ambientlabscomputing/mycelium_mesh_agent/releases/download/{version}/checksums.txt
      supported_platforms:
        - os: darwin
          arch: amd64
        - os: linux
          arch: amd64
    trust_tier: local
    sandbox_profile: standard
    runtime_requirements:
      args: []
      health_endpoint: http://localhost:10080/health
      network:
        - lan
      env:
        MMA_HTTP_PORT: "10080"
      port: 10080
    description: Mycelium Mesh Agent - Event-driven mesh networking agent
    homepage: https://github.com/ambientlabscomputing/mycelium_mesh_agent
    maintainer: Ambient Labs
```

## Behavior

### On Startup

1. UCRS checks the configured `environment` setting
2. Loads seed files from `seeds/{environment}/`
3. For each capability and provider:
   - If it exists (matching ID/version): **Updates** it with seed data
   - If it doesn't exist: **Creates** it

### Timestamps

- `created_at`: Preserved from existing record, or set to current time for new entries
- `updated_at`: Always set to current time when seed is applied

## Updating Seeds

To update registry data:

1. Edit the appropriate YAML file in `seeds/dev/` or `seeds/prod/`
2. Commit the changes to git
3. Restart UCRS

The updated data will automatically be applied to the database.

## Dev vs Prod Differences

The main difference between dev and prod seeds is the **provider version**:

- **Dev**: Uses `version: dev` for providers (pointing to development builds)
- **Prod**: Uses semantic versions like `version: 1.0.0`

Capabilities remain the same between environments.

## Adding New Seeds

1. Add the capability or provider definition to the appropriate YAML file
2. Follow the existing format and structure
3. Ensure all required fields are present
4. Restart UCRS to apply

## Troubleshooting

### Seeds Not Loading

Check the log output on startup:
```
level=INFO msg="loading registry seeds" environment=dev
level=INFO msg="upserted capability" id=mesh.service.discovery version=1.0.0
level=INFO msg="loaded capabilities" count=2
level=INFO msg="upserted provider" id=underleaf.mma version=dev
level=INFO msg="loaded providers" count=1
level=INFO msg="successfully loaded all registry seeds"
```

### Seed Directory Not Found

If you see:
```
level=WARN msg="seed directory not found, skipping seeding" path=seeds/dev
```

Ensure:
1. The `seeds/` directory exists in the working directory where UCRS runs
2. The environment subdirectory (`dev/` or `prod/`) exists
3. Run UCRS from the `ucrs/service/` directory

### YAML Parse Errors

If seed files fail to parse:
1. Validate YAML syntax using a linter
2. Ensure proper indentation (2 or 4 spaces, no tabs)
3. Check that all required fields are present
4. Verify enum values (e.g., `risk_class`, `trust_tier`) match expected values

## Implementation Details

The seeding system is implemented in the `/service/seeder/` package:

- **`seeder.go`**: Core seeding logic
- Integrated in `main.go` during startup sequence
- Uses repository layer for database operations
- Upsert logic preserves `created_at` timestamps

## Example Workflow

### Local Development

1. Use `environment: dev` in `config.yaml`
2. Provider version is `dev`
3. Points to development builds/branches
4. Iterate quickly with `underleaf.mma` provider

### Production Deployment

1. Set `environment: prod` in production config
2. Provider version is `1.0.0` (or latest stable)
3. Points to tagged releases
4. Stable, tested configurations

## Current Seed Data

### Capabilities

- `mesh.service.discovery` (v1.0.0) - Service discovery in mesh network
- `mesh.event.routing` (v1.0.0) - Event routing and mesh networking

### Providers

- `underleaf.mma` - Mycelium Mesh Agent
  - Implements both mesh capabilities
  - Binary artifact from GitHub releases
  - Supports darwin/linux/windows on amd64/arm64
