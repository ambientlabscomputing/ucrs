#!/bin/bash
# Register MMA (Mycelium Mesh Agent) as a test provider in UCRS

set -e

# First, let's create the required capabilities
echo "Creating capabilities..."

# Event routing capability
curl -X POST http://localhost:8083/api/v1/registry/capabilities \
  -H "Content-Type: application/json" \
  -d '{
    "id": "mesh.event.routing",
    "version": "1.0.0",
    "description": "Event routing and mesh networking capability",
    "risk_class": "medium",
    "allowed_verbs": ["subscribe", "publish", "route"],
    "permission_class": "mesh.network"
  }' || echo "Capability may already exist"

# Service discovery capability  
curl -X POST http://localhost:8083/api/v1/registry/capabilities \
  -H "Content-Type: application/json" \
  -d '{
    "id": "mesh.service.discovery",
    "version": "1.0.0", 
    "description": "Service discovery and registration in mesh network",
    "risk_class": "low",
    "allowed_verbs": ["register", "discover", "query"],
    "permission_class": "mesh.discovery"
  }' || echo "Capability may already exist"

echo ""
echo "Registering MMA provider..."

# Register MMA provider
curl -X POST http://localhost:8083/api/v1/registry/providers \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "underleaf.mma",
    "version": "dev",
    "description": "Mycelium Mesh Agent - Event-driven mesh networking agent for service discovery and routing",
    "capabilities": [
      {"id": "mesh.event.routing", "version_range": "^1.0"},
      {"id": "mesh.service.discovery", "version_range": "^1.0"}
    ],
    "artifact": {
      "type": "binary",
      "uri": "https://github.com/ambientlabscomputing/mycelium_mesh_agent/releases/download/{version}/mycelium-mesh-agent-{os}-{arch}",
      "checksums": {
        "uri": "https://github.com/ambientlabscomputing/mycelium_mesh_agent/releases/download/{version}/checksums.txt"
      },
      "supported_platforms": [
        {"os": "darwin", "arch": "amd64"},
        {"os": "darwin", "arch": "arm64"},
        {"os": "linux", "arch": "amd64"},
        {"os": "linux", "arch": "arm64"},
        {"os": "windows", "arch": "amd64"}
      ]
    },
    "trust_tier": "local",
    "sandbox_profile": "standard",
    "runtime_requirements": {
      "health_endpoint": "http://localhost:8080/health"
    },
    "maintainer": "Ambient Labs",
    "homepage": "https://github.com/ambientlabscomputing/mycelium_mesh_agent"
  }'

echo ""
echo "✓ MMA provider registered successfully"

# Verify
echo ""
echo "Verifying registration..."
curl -s http://localhost:8083/api/v1/registry/sync/snapshot | jq '.providers | length'
echo "providers in registry"
