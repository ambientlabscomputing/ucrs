# Underleaf Platform Integration Guide

**Version:** 1.0-draft
**Audience:** Application developers integrating with Underleaf
**Last Updated:** 2026-02-05

---

## Table of Contents

1. [Introduction](#introduction)
2. [Architecture Overview](#architecture-overview)
3. [Core Concepts](#core-concepts)
4. [Getting Started](#getting-started)
5. [Requesting Capabilities](#requesting-capabilities)
6. [Providing Capabilities](#providing-capabilities)
7. [Accessing Capabilities via the Mesh](#accessing-capabilities-via-the-mesh)
8. [Policy and Consent](#policy-and-consent)
9. [Telemetry and Observability](#telemetry-and-observability)
10. [Advanced Topics](#advanced-topics)
11. [Troubleshooting](#troubleshooting)
12. [Reference](#reference)

---

## Introduction

Underleaf is a distributed edge computing platform that enables secure, policy-governed capability-based communication between services across nodes. This guide walks you through integrating your applications with the Underleaf platform.

### What You'll Learn

- How to request and access capabilities from other services
- How to provide capabilities to the mesh
- How to work with the Mycelium Mesh Agent (MMA) for runtime connectivity
- How policy and consent enforcement works
- Best practices for edge-native development

### Prerequisites

- Underleaf Agent (UA) running on your node
- Basic understanding of service-oriented architecture
- Familiarity with REST APIs or gRPC

---

## Architecture Overview

Underleaf consists of three main components:

### 1. Underleaf Agent (UA)

**Authority:** System of record for cluster state

The UA manages:
- Cluster membership and health
- Node and service identity issuance
- Service lifecycle (install, start, stop, update)
- Policy document and consent state storage
- Capability Registry (UCRS) snapshot verification and caching

**Key Point:** UA is the source of truth for "what exists" and "what's allowed to run"

### 2. Mycelium Mesh Agent (MMA)

**Authority:** Runtime data plane and capability binding

The MMA provides:
- Runtime capability binding decisions
- Service-to-service mTLS connectivity
- Locality-aware routing (node → LAN → WAN)
- Policy evaluation and enforcement
- Telemetry collection from mesh activity

**Key Point:** MMA enforces "who can talk to whom" and "how" at runtime

### 3. Underleaf Capability Registry Service (UCRS)

**Authority:** Global capability schema definitions

The UCRS defines:
- Capability schemas and versioning
- Provider metadata and artifact locations
- Trust tier classifications (official, certified, community, experimental, local)
- Risk class assignments (low, medium, high)

**Key Point:** UCRS defines "what capabilities mean" globally

### Data Flow

```
┌─────────────────────────────────────────────────────┐
│                  UCRS (Cloud)                       │
│  • Capability schemas                               │
│  • Provider index                                   │
│  • Trust metadata                                   │
└────────────────┬────────────────────────────────────┘
                 │ (verified snapshots)
                 ▼
┌─────────────────────────────────────────────────────┐
│         Underleaf Agent (UA) - Node Level           │
│  • Cluster membership                               │
│  • Identity issuance                                │
│  • Service lifecycle                                │
│  • Policy storage                                   │
│  • Local capability cache                           │
└────────────────┬────────────────────────────────────┘
                 │ (event stream)
                 ▼
┌─────────────────────────────────────────────────────┐
│      Mycelium Mesh Agent (MMA) - Node Level         │
│  • Runtime binding engine                           │
│  • Service discovery                                │
│  • mTLS routing                                     │
│  • Policy evaluation                                │
└────────────────┬────────────────────────────────────┘
                 │ (mTLS connections)
                 ▼
          Your Services
```

---

## Core Concepts

### Capabilities

A **capability** is a semantic interface definition registered in UCRS. Capabilities use hierarchical DNS-style naming:

```
<domain>.<category>.<action>

Examples:
  iot.light.control
  iot.hue.bridge
  sensor.motion.read
```

Each capability has:
- **ID**: Unique identifier (e.g., `iot.light.control`)
- **Version**: Semantic version (e.g., `1.2.0`)
- **Risk Class**: `low`, `medium`, or `high`
- **Allowed Verbs**: Permitted operations
- **Schema**: JSON schema defining the interface
- **Permission Class**: Authorization category

### Providers

A **provider** is a running service that advertises a capability. Providers are:
- Validated against the capability cache
- Assigned a trust tier (official, certified, community, experimental, local)
- Associated with service identity and endpoints

### Bindings

A **binding** is a runtime grant connecting a client service to a provider. Bindings:
- Are issued by the MMA after policy evaluation
- Have time-to-live (TTL) and can expire
- Enforce constraints (locality, rate limits, trust tier)
- Can be revoked

### Locality

Underleaf prefers local-first routing:
1. **Same Node**: Services on the same machine
2. **Same LAN**: Services on the local network
3. **Same Site**: Services in the same datacenter/site
4. **WAN**: Remote services (if policy allows)

### Trust Tiers

Providers are classified by trust level:
- **Official**: Vetted by Underleaf/platform vendor
- **Certified**: Passed security review
- **Community**: Community-contributed
- **Experimental**: Development/testing
- **Local**: Locally developed, not published

---

## Getting Started

### Installation

1. **Install Underleaf Agent** on your node:

```bash
# See underleaf_client/README.md for installation instructions
curl -L -o ufctl https://github.com/ambientlabscomputing/underleaf_client/releases/download/${VERSION}/ufctl-linux-amd64
chmod +x ufctl
sudo mv ufctl /usr/local/bin/
```

2. **Authenticate** with the Underleaf control plane:

```bash
ufctl auth login
```

3. **Register** your node as a server:

```bash
ufctl register
```

4. **Start the agent** (includes MMA):

```bash
ufctl agent start -d
```

### Verify Installation

Check that both UA and MMA are running:

```bash
# Check UA status
ufctl agent status

# Check MMA mesh health
curl http://localhost:8080/api/v1/introspect/mesh

# Check MMA server health
curl http://localhost:8080/health
```

### MMA Configuration

The MMA is configured via the UA control channel. Default configuration includes:

**Control Channel:**
- Transport: `grpc_uds` (Unix Domain Socket)
- Address: `/tmp/ua_mma.sock`
- Fallback: TCP on `localhost:8000`

**Binding Engine:**
- Default TTL: 1 hour (configurable)
- Max TTL: 24 hours
- Default locality: `lan_preferred`
- Rate limiting: Configurable per-capability

**Telemetry:**
- Buffer size: 1MB
- Flush interval: 30 seconds
- Log level: info

**Runtime:**
- Data plane mode: `ambient`
- Connection pool: 100 connections
- Connection timeout: 5s
- QUIC enabled: true (when implemented)

Configuration can be updated via:
```bash
# Via UA deployment
# Config is pushed from UA to MMA via POST /api/v1/config/apply

# Hot reload (SIGHUP signal)
kill -HUP $(pgrep -f mycelium_mesh_agent)
```

---

## Requesting Capabilities

### Overview

When your service needs to use a capability provided by another service, you request a **binding** from the MMA.

### Binding Request Flow

```
1. Your Service
   ↓ (POST /v1/bindings/request)
2. MMA Binding Engine
   ↓ (validate capability exists)
3. MMA checks capability cache
   ↓ (find providers)
4. MMA Discovery Registry
   ↓ (sort by locality)
5. MMA Policy Evaluator
   ↓ (evaluate each candidate)
6. Policy + Consent Check
   ↓ (if allowed)
7. Issue Binding Grant
   ↓ (return provider info)
8. Your Service
```

### Making a Binding Request

**Endpoint:** `POST http://localhost:8080/api/v1/bindings/request`

The MMA HTTP API server runs on port 8080 by default. The control channel between UA and MMA uses Unix Domain Socket at `/tmp/ua_mma.sock` by default, or TCP at `localhost:8000` for cross-platform compatibility.

**Request Body:**

```json
{
  "request_id": "unique-request-id",
  "client": {
    "service_id": "your-service-id",
    "identity": "spiffe://cluster.local/your-service"
  },
  "capability": {
    "id": "iot.light.control",
    "version_constraint": "^1.0"
  },
  "constraints": {
    "locality": "lan_preferred",
    "max_latency_ms": 100,
    "min_trust_tier": "certified",
    "ttl_ms": 300000,
    "rate_limit": {
      "rps": 10.0,
      "burst": 20
    }
  },
  "context": {
    "user_present": true,
    "purpose": "Turn on living room lights"
  }
}
```

**Field Descriptions:**

- `request_id`: Unique idempotency key
- `client.service_id`: Your service's identifier (assigned by UA)
- `client.identity`: Your service's SPIFFE identity (issued by UA)
- `capability.id`: The capability you want to use
- `capability.version_constraint`: Semver range (optional)
- `constraints.locality`: One of:
  - `node_only`: Only same-node providers
  - `lan_preferred`: Prefer LAN, allow WAN
  - `lan_only`: Only LAN providers
  - `wan_allowed`: Any location
- `constraints.max_latency_ms`: Maximum acceptable latency (optional)
- `constraints.min_trust_tier`: Minimum provider trust tier
- `constraints.ttl_ms`: Binding lifetime in milliseconds
- `constraints.rate_limit`: Rate limiting (optional)
- `context.user_present`: Whether user is actively present (affects consent)
- `context.purpose`: Human-readable purpose (for audit)

**Response (Success):**

```json
{
  "decision": "allow",
  "reason_code": "",
  "binding_id": "binding-abc-123",
  "provider": {
    "service_id": "hue-bridge-01",
    "identity": "spiffe://cluster.local/hue-bridge",
    "endpoint": {
      "proto": "https",
      "host": "192.168.1.100",
      "port": 8443,
      "scope": "lan"
    }
  },
  "grant": {
    "token_ref": "binding-abc-123",
    "expires_at": "2026-02-05T10:15:00Z",
    "created_at": "2026-02-05T10:10:00Z"
  },
  "enforced_constraints": {
    "locality": "lan_preferred",
    "ttl_ms": 300000
  }
}
```

**Response (Denied):**

```json
{
  "decision": "deny",
  "reason_code": "POLICY_DENIED",
  "binding_id": "request-xyz-789"
}
```

**Reason Codes:**

- `CAPABILITY_UNKNOWN`: Capability not in local cache
- `NO_PROVIDER_AVAILABLE`: No providers running
- `POLICY_DENIED`: Policy evaluation rejected
- `CONSENT_REQUIRED`: User consent needed
- `TRUST_TIER_INSUFFICIENT`: No providers meet trust tier requirement
- `PROVENANCE_MISMATCH`: Provider digest doesn't match
- `LOCALITY_VIOLATION`: No providers match locality constraint
- `RATE_LIMITED`: Rate limit exceeded
- `INTERNAL_ERROR`: MMA internal error

### Using the Binding

Once you receive an "allow" decision:

1. **Extract the provider endpoint**
2. **Use the binding ID** as authorization token
3. **Establish mTLS connection** to the provider
4. **Track binding expiration** (`expires_at`)

**Example (pseudo-code):**

```go
// Make binding request
resp := makeBindingRequest(req)

if resp.Decision == "deny" {
    log.Error("Binding denied", "reason", resp.ReasonCode)
    return err
}

// Connect to provider
endpoint := resp.Provider.Endpoint
url := fmt.Sprintf("%s://%s:%d", endpoint.Proto, endpoint.Host, endpoint.Port)

// Use binding ID for auth (implementation varies by provider)
client := newMTLSClient(url, bindingID: resp.BindingID)

// Make capability call
result := client.Call("turn_on", params)
```

> **⚠️ Note:** The exact mechanism for presenting the binding ID to the provider is still being finalized. Check with your provider's documentation.

### Checking Binding Status

**Endpoint:** `GET http://localhost:8080/api/v1/bindings/{binding_id}`

**Response:**

```json
{
  "binding_id": "binding-abc-123",
  "state": "active",
  "client_service_id": "your-service-id",
  "provider_service_id": "hue-bridge-01",
  "capability_id": "iot.light.control",
  "created_at": "2026-02-05T10:10:00Z",
  "expires_at": "2026-02-05T10:15:00Z",
  "enforced_constraints": {
    "locality": "lan_preferred",
    "ttl_ms": 300000
  },
  "last_error": ""
}
```

**States:**
- `active`: Binding is valid
- `expired`: Binding has expired
- `revoked`: Binding was revoked
- `failed`: Binding failed to establish

---

## Providing Capabilities

### Overview

To provide a capability to the mesh, your service must:
1. Be registered with the UA
2. Advertise the capability in its service manifest
3. Implement the capability interface
4. Accept binding-authorized requests

### Registering as a Provider

#### Step 1: Service Registration

Your service must be deployed and started by the UA. The UA issues a service identity and notifies the MMA via the event stream.

The UA includes a full deployment system with:
- **Compiler**: Processes deployment manifests
- **Reconciler**: Computes required changes
- **Runner**: Executes deployment actions
- **Event Bus Integration**: Publishes deployment events

Services are deployed via the UA deployment API and can be native processes or Docker containers. Use `ufctl deploy apply` to deploy services.

#### Step 2: Advertise Capabilities

When UA starts your service, it emits a `service.started` event to MMA:

```json
{
  "event_type": "service.started",
  "payload": {
    "service_id": "hue-bridge-01",
    "service_identity": "spiffe://cluster.local/hue-bridge",
    "node_id": "node-123",
    "endpoints": [
      {
        "proto": "https",
        "host": "192.168.1.100",
        "port": 8443,
        "scope": "lan"
      }
    ],
    "capabilities_provided": [
      {
        "capability_id": "iot.light.control",
        "version": "1.2.0"
      },
      {
        "capability_id": "iot.hue.bridge",
        "version": "2.0.0"
      }
    ],
    "labels": {
      "app_id": "hue-bridge",
      "env": "production"
    }
  }
}
```

The MMA validates that each advertised capability exists in the local capability cache (synced from UCRS).

#### Step 3: Implement Capability Interface

Your service must implement the capability schema. The schema is defined in UCRS and cached locally by UA.

**Example Capability Schema (simplified):**

```json
{
  "id": "iot.light.control",
  "version": "1.2.0",
  "risk_class": "medium",
  "allowed_verbs": ["turn_on", "turn_off", "set_brightness", "get_state"],
  "description": "Generic lighting control capability",
  "schema": {
    "turn_on": {
      "params": {},
      "returns": {"success": "boolean"}
    },
    "turn_off": {
      "params": {},
      "returns": {"success": "boolean"}
    },
    "set_brightness": {
      "params": {"level": "integer (0-100)"},
      "returns": {"success": "boolean"}
    },
    "get_state": {
      "params": {},
      "returns": {"on": "boolean", "brightness": "integer"}
    }
  }
}
```

> **⚠️ Note:** The exact RPC/API protocol for capability invocations (REST, gRPC, custom protocol) is being standardized. Providers currently implement HTTP/gRPC endpoints.

#### Step 4: Accept Binding-Authorized Requests

Your service should:
1. Accept mTLS connections from clients
2. Validate client identity against binding grants
3. Enforce rate limits and constraints
4. Return results per capability schema

**Pseudo-code Example:**

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    // Extract binding ID from request (e.g., header, JWT, etc.)
    bindingID := r.Header.Get("X-Binding-ID")

    // Validate binding with MMA (or cache locally)
    binding := validateBinding(bindingID)
    if binding.State != "active" {
        http.Error(w, "Invalid binding", http.StatusUnauthorized)
        return
    }

    // Enforce rate limits
    if !checkRateLimit(binding.ClientServiceID, binding.EnforcedConstraints.RateLimit) {
        http.Error(w, "Rate limited", http.StatusTooManyRequests)
        return
    }

    // Execute capability
    result := executeCapability(r)
    json.NewEncoder(w).Encode(result)
}
```

> **⚠️ Note:** The binding validation mechanism (direct MMA query vs. cached validation) is being refined for production use.

### Provider Trust Tier

Your provider's trust tier is determined by UCRS based on:
- **Official**: Published by Underleaf/platform vendor
- **Certified**: Passed security audit and review
- **Community**: Community-contributed, verified
- **Experimental**: Development/testing phase
- **Local**: Locally developed, not submitted to UCRS

Trust tier affects whether clients can bind to your provider based on their `min_trust_tier` constraint.

---

## Accessing Capabilities via the Mesh

### mTLS Connectivity

All mesh communication uses mutual TLS (mTLS) for authentication and encryption.

#### Identity Management

- UA issues service identities (SPIFFE-like URIs)
- Certificates are rotated automatically
- MMA handles certificate presentation and validation

#### Connection Establishment

```
Client Service                    MMA                    Provider Service
      |                            |                             |
      |--- Binding Request ------->|                             |
      |                            |--- Policy Check             |
      |                            |--- Find Provider            |
      |<-- Binding Granted --------|                             |
      |                            |                             |
      |--- mTLS Handshake -------------------------------->      |
      |                  (MMA may intercept/proxy)               |
      |                            |                             |
      |--- Capability Call ----------------------------------->  |
      |<-- Result ---------------------------------------------|  |
```

### Routing Modes

The MMA supports three deployment modes configured via `mesh_mode`:

#### Ambient Mode (Default)

In ambient mode, MMA runs at the node level:
- Shared node-level routing
- Minimal per-service footprint
- Configuration: `mesh_mode: "ambient"`

This is the default and recommended mode.

> **⚠️ Note:** Transparent traffic interception via eBPF/tproxy is planned but not yet implemented. Currently services make direct HTTP requests to MMA API.

#### SDK Mode

Services use the MMA API directly:
- Make explicit binding requests to `http://localhost:8080/api/v1/bindings/request`
- Use returned endpoint information
- Establish mTLS connections to providers
- Handle connection lifecycle

This mode is fully operational and recommended for current development.

#### Sidecar Mode (Planned)

In sidecar mode, MMA would run as a proxy per service:
- Intercepts all outbound traffic
- Performs routing decisions
- Handles mTLS automatically

> **⚠️ Note:** Sidecar mode configuration exists but transparent proxying implementation is not yet complete.

### Locality-Aware Routing

MMA automatically routes to the nearest provider:

1. **Same Node**: Direct localhost communication
2. **Same LAN**: Local network routing
3. **WAN**: Remote routing (if allowed by policy)

The `locality` constraint in your binding request controls this behavior.

---

## Policy and Consent

### Policy Evaluation

When you request a binding, MMA evaluates policy using:

**Inputs:**
- Capability risk class (from cache)
- Provider trust tier (from discovery)
- Policy rules (from UA)
- Consent state (from UA)
- Binding constraints (from request)

**Default Policy:**

If no explicit policy is defined:
- **Low risk** capabilities: Allow (configurable)
- **Medium risk** capabilities: Allow if provider trust tier ≥ `certified`
- **High risk** capabilities: Deny unless explicitly allowed

### Consent Requirements

High-risk capabilities may require user consent:

1. **Consent Required**: MMA returns `CONSENT_REQUIRED` reason code
2. **UA Consent Flow**: Your application should trigger UA consent flow
3. **Consent Granted**: UA updates consent state via `consent_state.updated` event
4. **Retry Binding**: Request binding again

The MMA policy evaluator includes full consent state management:
- Consent states are stored per subject-capability pair
- The `UpdateConsent()` API allows UA to manage consent
- High-risk capabilities automatically check for valid consent

> **⚠️ Note:** End-user consent prompting UI/UX (how users grant consent via UI) is under development. The backend consent state management is production-ready.

### Custom Policies

Policies are stored in UA and evaluated by MMA. The policy evaluator is fully implemented with:
- **Trust tier enforcement**: Validates provider trust level meets requirements
- **Risk-based defaults**: Low/medium/high risk class handling per RFC
- **Consent checking**: Validates user consent for high-risk capabilities
- **Rate limiting**: Token bucket algorithm for per-client/capability rate limits
- **Provenance validation**: Ensures high-risk capabilities come from trusted sources

Policy documents define:
- Condition expressions
- Allow/deny decisions
- Additional constraints
- Priority/ordering

**Example Policy Rule:**

```json
{
  "id": "high-risk-require-certification",
  "condition": "capability.risk_class == 'high'",
  "decision": "deny",
  "constraints": {
    "min_trust_tier": "certified"
  },
  "priority": 100
}
```

**Policy Evaluation Flow:**

The MMA policy evaluator performs these checks in order:
1. Validate capability exists in cache
2. Check locality constraints
3. Check trust tier (default minimum: `certified`)
4. Check provenance for high-risk capabilities
5. Evaluate custom policy rules
6. Check consent for high-risk capabilities
7. Apply rate limiting

> **⚠️ Note:** Policy authoring UI/API and policy management tools are being developed. The policy evaluation engine is production-ready.

### Policy Explainability

Every deny decision includes:
- **Reason Code**: Machine-readable code
- **Policy Trace Reference**: Pointer to policy rule (optional)
- **Capability ID**: Which capability was denied

This enables debugging and audit trails.

---

## Telemetry and Observability

### MMA Telemetry

MMA collects telemetry on:
- Binding decisions (granted, denied, revoked)
- Connection metrics (latency, errors)
- Policy decisions
- Mesh health events

### Accessing Telemetry

UA can request telemetry from MMA:

**Endpoint:** `POST http://localhost:8080/api/v1/telemetry/flush`

**Request:**

```json
{
  "types": ["audit", "metrics"],
  "since": "2026-02-05T10:00:00Z",
  "max_bytes": 1048576
}
```

**Response:**

```json
{
  "delivered": {
    "audit_count": 150,
    "metric_points": 1200
  },
  "remaining_queued": 50,
  "delivery_ref": "telemetry-batch-001",
  "events": [
    {
      "event_type": "mesh.binding.granted",
      "timestamp": "2026-02-05T10:05:00Z",
      "binding_id": "binding-abc-123",
      "client_service_id": "app-001",
      "provider_service_id": "sensor-002",
      "capability_id": "sensor.motion.read"
    },
    {
      "event_type": "mesh.binding.denied",
      "timestamp": "2026-02-05T10:06:00Z",
      "request_id": "req-456",
      "client_service_id": "app-003",
      "capability_id": "iot.thermostat.control",
      "reason_code": "POLICY_DENIED"
    }
  ]
}
```

The MMA telemetry system is fully implemented with:
- **Buffer**: Thread-safe circular buffer for telemetry events (configurable size, default 1MB)
- **Collector**: Real-time metric collection (binding counts, error rates)
- **Flusher**: On-demand and periodic flushing to UA
- **Event Types**: Binding granted/denied/revoked, policy decisions, health changes

Telemetry configuration is managed via the MMA config:
- `buffer_size_bytes`: Max buffer size (default: 1MB)
- `flush_interval_ms`: Auto-flush interval (default: 30s)
- `include_traces`: Whether to include detailed policy traces

### Mesh Health

Check mesh health:

**Endpoint:** `GET http://localhost:8080/api/v1/introspect/mesh`

**Response:**

```json
{
  "mesh_state": "ready",
  "known_members": 5,
  "known_services": 23,
  "capability_cache_version": "2026-02-05-001",
  "policy_version": "v1.2.3",
  "data_plane": {
    "mode": "ambient",
    "status": "active"
  },
  "buffers": {
    "telemetry_bytes": 102400,
    "audit_events_queued": 15
  }
}
```

**Mesh States:**
- `ready`: Normal operation
- `degraded`: Partial functionality (e.g., missing policy)
- `offline`: Cannot operate (e.g., no capability cache)

---

## Advanced Topics

### Service Introspection

Query detailed service information:

**Endpoint:** `POST http://localhost:8080/api/v1/introspect/service`

**Request:**

```json
{
  "service_id": "hue-bridge-01",
  "include_routes": true,
  "include_policies": true
}
```

**Response:**

```json
{
  "service": {
    "service_id": "hue-bridge-01",
    "identity": "spiffe://cluster.local/hue-bridge",
    "node_id": "node-123",
    "endpoints": [
      {
        "proto": "https",
        "host": "192.168.1.100",
        "port": 8443,
        "scope": "lan"
      }
    ],
    "labels": {
      "app_id": "hue-bridge",
      "env": "production"
    }
  },
  "provides": [
    {
      "capability_id": "iot.light.control",
      "version": "1.2.0"
    }
  ],
  "consumes": [
    {
      "binding_id": "binding-xyz-789",
      "capability_id": "sensor.motion.read",
      "state": "active"
    }
  ],
  "routes": [
    {
      "peer_service_id": "app-001",
      "locality": "same_lan",
      "latency_ms": 5
    }
  ]
}
```

### Binding Revocation

Revoke an active binding:

> **⚠️ Note:** Binding revocation API is defined but the mechanism for external revocation triggers (e.g., admin console, policy change) is TBD.

### Offline Operation

MMA continues operating when:
- UA temporarily unavailable
- WAN disconnected
- UCRS unreachable

Using last-known-good:
- Policy snapshot
- Capability cache
- Membership view

**Behavior:**
- Existing bindings continue until expiration
- New bindings use cached policy
- Mesh state may transition to `degraded` if critical data is stale

### Host Adapter for Native Devices

The MMA includes a Host Adapter system for bridging native/embedded devices into the mesh:

**Use Cases:**
- IoT devices (sensors, actuators)
- Embedded systems
- Legacy hardware
- Native applications without mesh awareness

**Device Provider API:**

```go
// Register a device as a service
adapter.RegisterService(&types.ServiceEntry{
    ServiceID: "my-device-01",
    ServiceIdentity: "spiffe://cluster.local/my-device",
    NodeID: "node-123",
    Endpoints: []types.Endpoint{{
        Proto: "http",
        Host: "192.168.1.50",
        Port: 8080,
        Scope: "lan",
    }},
    Labels: map[string]string{
        "device_type": "sensor",
    },
})

// Advertise capabilities
adapter.AdvertiseCapability("my-device-01", &types.Capability{
    ID: "sensor.temperature.read",
    Version: "1.0.0",
    RiskClass: types.RiskClassLow,
})
```

The Host Adapter:
- Validates capabilities against cache
- Enforces policy at adapter boundary
- Handles binding requests for device services
- Manages device lifecycle (register/unregister)

Host adapters are managed by the Adapter Manager which supports multiple adapter types per node.

### Multi-Node Scenarios

**Same LAN:**
- mDNS may assist discovery
- Direct IP routing preferred
- Low latency

**WAN:**
- Requires `wan_allowed` locality
- QUIC support configurable via `runtime_config.enable_quic` (default: `true`)
- Higher latency

The runtime config includes QUIC support configuration. WAN routing is implemented with configurable connection timeouts and pooling:
- `connection_timeout_ms`: 5000ms default
- `connection_idle_timeout_ms`: 30000ms default
- `connection_pool_size`: 100 connections default

> **⚠️ Note:** QUIC protocol implementation for WAN connections is in progress. TCP/HTTP connections are currently used.

---

## Troubleshooting

### Binding Denied: CAPABILITY_UNKNOWN

**Cause:** Capability not in local cache

**Solutions:**
1. Verify capability exists in UCRS
2. Check UA capability cache sync: `ufctl config` (verify cache version)
3. Wait for UA to sync latest UCRS snapshot
4. Check MMA logs for cache update events

> **⚠️ Note:** UA capability cache sync mechanics and CLI commands are being finalized.

### Binding Denied: NO_PROVIDER_AVAILABLE

**Cause:** No running services provide the capability

**Solutions:**
1. Verify provider service is running: `ufctl servers list`
2. Check provider advertises capability in service manifest
3. Verify provider's capability version matches your constraint
4. Check provider is healthy and reachable

### Binding Denied: POLICY_DENIED

**Cause:** Policy evaluation rejected the binding

**Solutions:**
1. Check capability risk class (low/medium/high)
2. Verify provider trust tier meets your `min_trust_tier`
3. Review policy rules (if custom policy defined)
4. Check MMA logs for detailed policy evaluation:
   ```bash
   tail -f /path/to/mma.log | grep "policy"
   ```
5. For high-risk capabilities, verify consent has been granted

The policy evaluator logs detailed information about:
- Which check failed (locality, trust tier, provenance, consent, rate limit)
- The provider being evaluated
- The reason code returned

**Example log entry:**
```
INFO Policy denied binding reason=trust_tier_insufficient provider=sensor-01 client=app-03 capability=iot.thermostat.control
```

> **⚠️ Note:** Structured policy trace inspection UI is under development. Current debugging relies on logs.

### Binding Denied: LOCALITY_VIOLATION

**Cause:** No providers match your locality constraint

**Solutions:**
1. Relax locality constraint (e.g., `lan_only` → `lan_preferred`)
2. Verify provider is on expected network scope
3. Check node/LAN discovery is functioning
4. Review provider endpoint scopes

### Mesh State: Degraded

**Causes:**
- Missing or stale policy
- Missing or stale capability cache
- UA unavailable

**Solutions:**
1. Check UA status: `ufctl agent status`
2. Verify MMA can connect to UA event stream
3. Check MMA logs for error messages
4. Restart MMA if necessary: `ufctl agent restart`

### Connection Failed After Binding Granted

**Causes:**
- Provider service stopped
- Network connectivity issue
- Binding expired
- mTLS certificate invalid

**Solutions:**
1. Check binding status: `GET /v1/bindings/{binding_id}`
2. Verify provider is running
3. Check network connectivity to provider endpoint
4. Verify identity certificates are valid (check MMA logs)
5. Request new binding if expired

---

## Reference

### MMA HTTP API

All MMA APIs are served on `http://localhost:8080` by default.

### Binding Request API

- **Endpoint:** `POST /api/v1/bindings/request`
- **Auth:** Node-local only
- **Response:** Binding decision with provider info or deny reason

### Binding Status API

- **Endpoint:** `GET /api/v1/bindings/{binding_id}`
- **Auth:** Node-local only
- **Response:** Current binding state and metadata

### Telemetry API

- **Endpoint:** `POST /api/v1/telemetry/flush`
- **Auth:** UA-only (authenticated via control channel)
- **Response:** Telemetry events and summary

### Mesh Introspection API

- **Endpoint:** `GET /api/v1/introspect/mesh`
- **Auth:** Node-local only
- **Response:** Mesh health and statistics

### Service Introspection API

- **Endpoint:** `POST /api/v1/introspect/service`
- **Auth:** Node-local only
- **Response:** Service details, capabilities, and routes

### Configuration API

- **Endpoint:** `POST /api/v1/config/apply`
- **Auth:** UA-only (authenticated via control channel)
- **Request:** MMA configuration object
- **Response:** Applied status and warnings

### Health Check API

- **Endpoint:** `GET /health`
- **Auth:** None required
- **Response:** Server health status

### Metrics API

- **Endpoint:** `GET /metrics`
- **Auth:** Node-local only
- **Response:** Current telemetry metrics

### Locality Preferences

- `node_only`: Same node only
- `lan_preferred`: Prefer LAN, allow WAN
- `lan_only`: LAN only
- `wan_allowed`: Any location

### Trust Tiers

- `official`: Official/vendor-provided
- `certified`: Security-reviewed
- `community`: Community-verified
- `experimental`: Development/testing
- `local`: Local development

### Risk Classes

- `low`: Informational, read-only
- `medium`: State-changing, physical control
- `high`: Safety/security critical

### Binding States

- `active`: Valid and usable
- `expired`: TTL exceeded
- `revoked`: Explicitly revoked
- `failed`: Failed to establish

### Mesh States

- `ready`: Normal operation
- `degraded`: Partial functionality
- `offline`: Cannot operate

### Reason Codes

- `CAPABILITY_UNKNOWN`
- `NO_PROVIDER_AVAILABLE`
- `POLICY_DENIED`
- `CONSENT_REQUIRED`
- `TRUST_TIER_INSUFFICIENT`
- `PROVENANCE_MISMATCH`
- `LOCALITY_VIOLATION`
- `RATE_LIMITED`
- `INTERNAL_ERROR`

### Event Types (UA → MMA)

**Membership:**
- `cluster.snapshot`
- `member.joined`
- `member.updated`
- `member.left`
- `member.health`

**Identity:**
- `identity.trust_roots.updated`
- `identity.service.issued`
- `identity.service.revoked`

**Services:**
- `service.started`
- `service.updated`
- `service.stopped`

**Capabilities:**
- `capability_cache.snapshot.updated`
- `capability_cache.delta.updated`

**Policy:**
- `mesh_policy.updated`
- `consent_state.updated`

### Event Types (MMA → UA)

- `mesh.binding.granted`
- `mesh.binding.denied`
- `mesh.binding.revoked`
- `mesh.policy.decision`
- `mesh.telemetry.summary`
- `mesh.health.changed`

---

## Additional Resources

- **UCRS Architecture:** See `capability_registry_service/guiding_docs/UCRS_Arch.md`
- **MMA RFC:** See `mycelium_mesh_agent/agent_docs/MYCELLIUM_MESH_AGENT_RFC.md`
- **Interface Contract:** See `mycelium_mesh_agent/agent_docs/INTERFACE_CONTRACT.md`
- **Underleaf Client:** See `underleaf_client/README.md`

---

## Notes on Production Readiness

This guide reflects the current state of Underleaf components as of 2026-02-05.

### Production-Ready Components

The following components are fully implemented and operational:

1. **MMA HTTP API Server**: All v1 API endpoints operational on port 8080
2. **Binding Engine**: Full RFC-compliant capability binding with policy evaluation
3. **Policy Evaluator**: Trust tier checking, risk-based defaults, consent management, rate limiting
4. **Telemetry System**: Buffer, collector, and flusher with configurable behavior
5. **Service Discovery**: UA event stream consumer, registry with locality-aware routing
6. **Configuration Management**: Full config store with validation and hot-reload (SIGHUP)
7. **Host Adapter**: Device provider adapter for bridging native/embedded services
8. **UA Deployment System**: Compiler, reconciler, runner for service deployments
9. **Control Channel**: UA↔MMA communication via Unix Domain Socket or TCP

### Under Active Development

The following areas are being refined for production use:

1. **MMA Client SDK**: HTTP client library for capability requests
2. **Binding ID Validation**: Standardized mechanism for providers to validate binding IDs from clients
3. **Transparent Proxying**: eBPF/tproxy implementation for sidecar and ambient modes
4. **Consent UI/UX**: End-user consent prompting interface (backend consent management is complete)
5. **Policy Authoring Tools**: UI/API for creating and managing custom policies (evaluation engine is complete)
6. **QUIC Protocol**: QUIC transport implementation for WAN connections (configuration exists)
7. **MMA Management CLI**: Command-line tools for MMA operations

### Recommended Current Usage

For current development and early production:
- Use **SDK mode** with direct HTTP API calls to MMA
- Deploy services via **UA deployment system** (`ufctl deploy apply`)
- Configure policies via **UA policy storage** and event stream
- Monitor via **MMA telemetry APIs** and metrics endpoint
- Use **localhost/LAN** routing (WAN capabilities exist but are less tested)

Check individual component documentation and CHANGELOG files for latest status.

---

**Questions or Issues?**

- File issues at: `https://github.com/ambientlabscomputing/underleaf/issues`
- Documentation feedback: Contact the Underleaf platform team

---

*Copyright © 2026 Ambient Labs. All rights reserved.*