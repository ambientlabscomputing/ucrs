RFC-UL-CR-001

Underleaf Capability Registry Cloud Service (UCRS)

Status: Draft
Version: 1.0
Audience: Platform Infrastructure, Governance, Security
Scope: Cloud control-plane service

⸻

1. Purpose

The Underleaf Capability Registry Service (UCRS) provides the global semantic authority and provider index used by Underleaf Agents to resolve, verify, and install capability providers.

UCRS is responsible for:
	•	Capability schema governance
	•	Provider metadata registry
	•	Trust tier classification
	•	Artifact location indexing
	•	Registry snapshot distribution

UCRS does not host runtime artifacts or manage deployments.

⸻

2. Design Principles

UCRS MUST:
	•	Remain product-agnostic
	•	Be deterministic and cache-friendly
	•	Support offline-first agents
	•	Maintain cryptographic integrity guarantees
	•	Avoid runtime coupling

UCRS MUST NOT:
	•	Execute workloads
	•	Host binary artifacts
	•	Perform installation
	•	Act as consumer-facing app marketplace

⸻

3. System Role

Cloud Authority Layer

Capability Definitions
Provider Index
Trust Metadata
Compatibility Rules
↓
Distributed to Underleaf Agents


⸻

4. Capability Schema Authority

4.1 Capability Namespace Format

Capabilities MUST use hierarchical DNS-style naming:

<domain>.<category>.<action>

Examples:
	•	iot.light.control
	•	iot.hue.bridge
	•	sensor.motion.read

⸻

4.2 Capability Schema Model

Each capability entry MUST include:

Field	Required
id	yes
version	yes
description	yes
risk_class	yes
allowed_verbs	yes
permission_class	yes
deprecation_policy	optional

Example:

{
  "id": "iot.light.control",
  "version": "1.0.0",
  "risk_class": "medium",
  "allowed_verbs": ["on", "off", "set_brightness"],
  "permission_class": "device.control",
  "description": "Generic lighting control capability"
}


⸻

4.3 Versioning Requirements

Capabilities MUST follow semantic versioning.

Breaking behavior changes MUST increment major version.

Backward compatibility MUST be preserved for minor versions.

⸻

4.4 Risk Classification

Capabilities MUST declare operational risk:

Level	Meaning
low	informational
medium	physical state change
high	safety/security critical


⸻

5. Provider Artifact Index

5.1 Provider Metadata Model

Each provider entry MUST contain:

{
  "provider_id": "ambient.hue-mcp",
  "version": "2.1.0",
  "capabilities": [
    {"id": "iot.hue.bridge", "version_range": "^1.0"}
  ],
  "artifact": {
    "type": "oci",
    "uri": "registry.underleaf.io/ambient/hue-mcp:2.1.0",
    "digest": "sha256:..."
  },
  "trust_tier": "official",
  "sandbox_profile": "standard",
  "runtime_requirements": {
    "network": ["lan"],
    "secrets": ["hue_api_key"]
  }
}


⸻

5.2 Artifact Hosting

Artifacts MUST NOT be hosted by UCRS.

Supported artifact types:
	•	OCI container images
	•	Signed binaries
	•	Git repositories
	•	Package registries (npm, pip)

⸻

5.3 Trust Tier Classification

Each provider MUST be assigned:
	•	official
	•	certified
	•	community
	•	experimental
	•	local

⸻

6. Registry Distribution API

6.1 Snapshot Sync

GET /v1/sync/snapshot

Returns:
	•	capability schemas
	•	provider index
	•	trust metadata
	•	signed manifest

⸻

6.2 Delta Sync

GET /v1/sync/delta?since=<timestamp>

Returns incremental updates.

⸻

6.3 Integrity Protection

All registry snapshots MUST be signed.

Agents MUST reject unsigned or tampered payloads.

⸻

7. Governance

7.1 Capability Approval

New capabilities MUST undergo:
	•	schema review
	•	naming approval
	•	compatibility analysis

⸻

7.2 Provider Certification

Certified providers MUST:
	•	pass security review
	•	conform to sandbox rules
	•	verify capability correctness

⸻

8. Non-Goals

UCRS does not:
	•	manage billing
	•	host UI marketplace
	•	handle runtime telemetry
	•	perform deployments
