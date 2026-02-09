# Registry Seeds

This directory contains seed data for the UCRS (Underleaf Capability Registry Service).

## Structure

- `dev/` - Development environment seeds (uses version "dev" for providers)
- `prod/` - Production environment seeds (uses semantic versions like "1.0.0")

## Files

Each environment directory contains:
- `capabilities.yaml` - Capability definitions
- `providers.yaml` - Provider metadata and configurations

## Usage

Seeds are automatically loaded when UCRS starts. The environment is determined by the `ENVIRONMENT` configuration variable (defaults to "dev").

Seeds will **override** existing entries with matching IDs, allowing you to manage the registry through version control.

## Adding New Seeds

1. Edit the appropriate YAML file in `dev/` or `prod/`
2. Maintain parity between dev and prod (except for version numbers)
3. Restart UCRS to apply changes
