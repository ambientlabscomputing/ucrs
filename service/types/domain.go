package types

import "time"

// RiskClass represents the operational risk level of a capability
type RiskClass string

const (
	RiskLow    RiskClass = "low"    // informational
	RiskMedium RiskClass = "medium" // physical state change
	RiskHigh   RiskClass = "high"   // safety/security critical
)

// Capability represents a semantic capability definition in the registry
type Capability struct {
	ID                string    `json:"id" bson:"_id"`          // e.g., "iot.light.control"
	Version           string    `json:"version" bson:"version"` // Semantic version
	Description       string    `json:"description" bson:"description"`
	RiskClass         RiskClass `json:"risk_class" bson:"risk_class"`
	AllowedVerbs      []string  `json:"allowed_verbs" bson:"allowed_verbs"`       // e.g., ["on", "off", "set_brightness"]
	PermissionClass   string    `json:"permission_class" bson:"permission_class"` // e.g., "device.control"
	DeprecationPolicy string    `json:"deprecation_policy,omitempty" bson:"deprecation_policy,omitempty"`
	CreatedAt         time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" bson:"updated_at"`
}

// TrustTier represents the trust classification of a provider
type TrustTier string

const (
	TrustOfficial     TrustTier = "official"     // First-party providers
	TrustCertified    TrustTier = "certified"    // Vetted third-party providers
	TrustCommunity    TrustTier = "community"    // Community-contributed providers
	TrustExperimental TrustTier = "experimental" // Experimental/beta providers
	TrustLocal        TrustTier = "local"        // Local development providers
)

// ArtifactType represents the type of provider artifact
type ArtifactType string

const (
	ArtifactOCI    ArtifactType = "oci"    // OCI container image
	ArtifactBinary ArtifactType = "binary" // Signed binary
	ArtifactGit    ArtifactType = "git"    // Git repository
	ArtifactNPM    ArtifactType = "npm"    // NPM package
	ArtifactPyPI   ArtifactType = "pypi"   // Python package
)

// Artifact represents the location and integrity data for a provider artifact
type Artifact struct {
	Type   ArtifactType `json:"type" bson:"type"`
	URI    string       `json:"uri" bson:"uri"`       // e.g., "registry.underleaf.io/ambient/hue-mcp:2.1.0"
	Digest string       `json:"digest" bson:"digest"` // e.g., "sha256:..."
}

// CapabilityRef represents a reference to a capability with version constraints
type CapabilityRef struct {
	ID           string `json:"id" bson:"id"`                       // e.g., "iot.hue.bridge"
	VersionRange string `json:"version_range" bson:"version_range"` // e.g., "^1.0"
}

// RuntimeRequirements specifies runtime dependencies for a provider
type RuntimeRequirements struct {
	Network []string          `json:"network,omitempty" bson:"network,omitempty"` // e.g., ["lan", "wan"]
	Secrets []string          `json:"secrets,omitempty" bson:"secrets,omitempty"` // e.g., ["hue_api_key"]
	Env     map[string]string `json:"env,omitempty" bson:"env,omitempty"`         // Environment variables
}

// Provider represents a capability provider in the registry
type Provider struct {
	ProviderID          string              `json:"provider_id" bson:"_id"`           // e.g., "ambient.hue-mcp"
	Version             string              `json:"version" bson:"version"`           // Semantic version
	Capabilities        []CapabilityRef     `json:"capabilities" bson:"capabilities"` // Capabilities this provider implements
	Artifact            Artifact            `json:"artifact" bson:"artifact"`
	TrustTier           TrustTier           `json:"trust_tier" bson:"trust_tier"`
	SandboxProfile      string              `json:"sandbox_profile" bson:"sandbox_profile"` // e.g., "standard", "isolated"
	RuntimeRequirements RuntimeRequirements `json:"runtime_requirements" bson:"runtime_requirements"`
	Description         string              `json:"description,omitempty" bson:"description,omitempty"`
	Maintainer          string              `json:"maintainer,omitempty" bson:"maintainer,omitempty"`
	Homepage            string              `json:"homepage,omitempty" bson:"homepage,omitempty"`
	CreatedAt           time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at" bson:"updated_at"`
}

// SignedManifest represents cryptographic integrity proof for registry snapshots
type SignedManifest struct {
	KeyID     string    `json:"key_id" bson:"key_id"`       // Signing key identifier
	Signature string    `json:"signature" bson:"signature"` // Ed25519 signature (base64)
	Algorithm string    `json:"algorithm" bson:"algorithm"` // "Ed25519"
	SignedAt  time.Time `json:"signed_at" bson:"signed_at"`
}

// RegistrySnapshot represents a complete snapshot of the capability registry
type RegistrySnapshot struct {
	Version      string         `json:"version" bson:"version"` // Snapshot version
	Timestamp    time.Time      `json:"timestamp" bson:"timestamp"`
	Capabilities []Capability   `json:"capabilities" bson:"capabilities"`
	Providers    []Provider     `json:"providers" bson:"providers"`
	Manifest     SignedManifest `json:"manifest" bson:"manifest"`
	ETag         string         `json:"etag" bson:"etag"` // For caching
}

// DeltaUpdate represents incremental changes since a previous snapshot
type DeltaUpdate struct {
	SinceVersion        string         `json:"since_version" bson:"since_version"`
	ToVersion           string         `json:"to_version" bson:"to_version"`
	Timestamp           time.Time      `json:"timestamp" bson:"timestamp"`
	AddedCapabilities   []Capability   `json:"added_capabilities" bson:"added_capabilities"`
	UpdatedCapabilities []Capability   `json:"updated_capabilities" bson:"updated_capabilities"`
	RemovedCapabilities []string       `json:"removed_capabilities" bson:"removed_capabilities"` // IDs
	AddedProviders      []Provider     `json:"added_providers" bson:"added_providers"`
	UpdatedProviders    []Provider     `json:"updated_providers" bson:"updated_providers"`
	RemovedProviders    []string       `json:"removed_providers" bson:"removed_providers"` // Provider IDs
	Manifest            SignedManifest `json:"manifest" bson:"manifest"`
}
