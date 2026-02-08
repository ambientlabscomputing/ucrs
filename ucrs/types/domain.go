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
	ID                string    `json:"id" bson:"_id" yaml:"id"`               // e.g., "iot.light.control"
	Version           string    `json:"version" bson:"version" yaml:"version"` // Semantic version
	Description       string    `json:"description" bson:"description" yaml:"description"`
	RiskClass         RiskClass `json:"risk_class" bson:"risk_class" yaml:"risk_class"`
	AllowedVerbs      []string  `json:"allowed_verbs" bson:"allowed_verbs" yaml:"allowed_verbs"`          // e.g., ["on", "off", "set_brightness"]
	PermissionClass   string    `json:"permission_class" bson:"permission_class" yaml:"permission_class"` // e.g., "device.control"
	DeprecationPolicy string    `json:"deprecation_policy,omitempty" bson:"deprecation_policy,omitempty" yaml:"deprecation_policy,omitempty"`
	CreatedAt         time.Time `json:"created_at" bson:"created_at" yaml:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" bson:"updated_at" yaml:"updated_at"`
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

// PlatformSpec represents a specific OS/architecture combination
type PlatformSpec struct {
	OS   string `json:"os" bson:"os" yaml:"os"`       // e.g., "darwin", "linux", "windows"
	Arch string `json:"arch" bson:"arch" yaml:"arch"` // e.g., "amd64", "arm64"
}

// Artifact represents the location and integrity data for a provider artifact
type Artifact struct {
	Type               ArtifactType       `json:"type" bson:"type" yaml:"type"`
	URI                string             `json:"uri" bson:"uri" yaml:"uri"`                                                                               // e.g., "registry.underleaf.io/ambient/hue-mcp:2.1.0" or template with {os}/{arch}/{version}
	Digest             string             `json:"digest,omitempty" bson:"digest,omitempty" yaml:"digest,omitempty"`                                        // e.g., "sha256:..." (for single-platform artifacts)
	Checksums          *ChecksumsArtifact `json:"checksums,omitempty" bson:"checksums,omitempty" yaml:"checksums,omitempty"`                               // For multi-platform binaries
	SupportedPlatforms []PlatformSpec     `json:"supported_platforms,omitempty" bson:"supported_platforms,omitempty" yaml:"supported_platforms,omitempty"` // Platforms this binary supports (empty = platform-agnostic)
}

// ChecksumsArtifact points to a checksums file for multi-platform binaries
type ChecksumsArtifact struct {
	URI string `json:"uri" bson:"uri" yaml:"uri"` // e.g., "https://github.com/.../checksums.txt"
}

// CapabilityRef represents a reference to a capability with version constraints
type CapabilityRef struct {
	ID           string `json:"id" bson:"id" yaml:"id"`                                  // e.g., "iot.hue.bridge"
	VersionRange string `json:"version_range" bson:"version_range" yaml:"version_range"` // e.g., "^1.0"
}

// RuntimeRequirements specifies runtime dependencies for a provider
type RuntimeRequirements struct {
	Network        []string          `json:"network,omitempty" bson:"network,omitempty" yaml:"network,omitempty"`                         // e.g., ["lan", "wan"]
	Secrets        []string          `json:"secrets,omitempty" bson:"secrets,omitempty" yaml:"secrets,omitempty"`                         // e.g., ["hue_api_key"]
	Env            map[string]string `json:"env,omitempty" bson:"env,omitempty" yaml:"env,omitempty"`                                     // Environment variables
	Args           []string          `json:"args,omitempty" bson:"args,omitempty" yaml:"args,omitempty"`                                  // Command-line arguments for binary providers
	Port           int               `json:"port,omitempty" bson:"port,omitempty" yaml:"port,omitempty"`                                  // Port the provider listens on
	HealthEndpoint string            `json:"health_endpoint,omitempty" bson:"health_endpoint,omitempty" yaml:"health_endpoint,omitempty"` // HTTP health check endpoint (e.g., "http://localhost:8080/health")
}

// Provider represents a capability provider in the registry
type Provider struct {
	ProviderID          string              `json:"provider_id" bson:"_id" yaml:"provider_id"`            // e.g., "ambient.hue-mcp"
	Version             string              `json:"version" bson:"version" yaml:"version"`                // Semantic version
	Capabilities        []CapabilityRef     `json:"capabilities" bson:"capabilities" yaml:"capabilities"` // Capabilities this provider implements
	Artifact            Artifact            `json:"artifact" bson:"artifact" yaml:"artifact"`
	TrustTier           TrustTier           `json:"trust_tier" bson:"trust_tier" yaml:"trust_tier"`
	SandboxProfile      string              `json:"sandbox_profile" bson:"sandbox_profile" yaml:"sandbox_profile"` // e.g., "standard", "isolated"
	RuntimeRequirements RuntimeRequirements `json:"runtime_requirements" bson:"runtime_requirements" yaml:"runtime_requirements"`
	Description         string              `json:"description,omitempty" bson:"description,omitempty" yaml:"description,omitempty"`
	Maintainer          string              `json:"maintainer,omitempty" bson:"maintainer,omitempty" yaml:"maintainer,omitempty"`
	Homepage            string              `json:"homepage,omitempty" bson:"homepage,omitempty" yaml:"homepage,omitempty"`
	CreatedAt           time.Time           `json:"created_at" bson:"created_at" yaml:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at" bson:"updated_at" yaml:"updated_at"`
}

// SignedManifest represents cryptographic integrity proof for registry snapshots
type SignedManifest struct {
	KeyID     string    `json:"key_id" bson:"key_id" yaml:"key_id"`          // Signing key identifier
	Signature string    `json:"signature" bson:"signature" yaml:"signature"` // Ed25519 signature (base64)
	Algorithm string    `json:"algorithm" bson:"algorithm" yaml:"algorithm"` // "Ed25519"
	SignedAt  time.Time `json:"signed_at" bson:"signed_at" yaml:"signed_at"`
}

// RegistrySnapshot represents a complete snapshot of the capability registry
type RegistrySnapshot struct {
	Version      string         `json:"version" bson:"version" yaml:"version"` // Snapshot version
	Timestamp    time.Time      `json:"timestamp" bson:"timestamp" yaml:"timestamp"`
	Capabilities []Capability   `json:"capabilities" bson:"capabilities" yaml:"capabilities"`
	Providers    []Provider     `json:"providers" bson:"providers" yaml:"providers"`
	Manifest     SignedManifest `json:"manifest" bson:"manifest" yaml:"manifest"`
	ETag         string         `json:"etag" bson:"etag" yaml:"etag"` // For caching
}

// DeltaUpdate represents incremental changes since a previous snapshot
type DeltaUpdate struct {
	SinceVersion        string         `json:"since_version" bson:"since_version" yaml:"since_version"`
	ToVersion           string         `json:"to_version" bson:"to_version" yaml:"to_version"`
	Timestamp           time.Time      `json:"timestamp" bson:"timestamp" yaml:"timestamp"`
	AddedCapabilities   []Capability   `json:"added_capabilities" bson:"added_capabilities" yaml:"added_capabilities"`
	UpdatedCapabilities []Capability   `json:"updated_capabilities" bson:"updated_capabilities" yaml:"updated_capabilities"`
	RemovedCapabilities []string       `json:"removed_capabilities" bson:"removed_capabilities" yaml:"removed_capabilities"` // IDs
	AddedProviders      []Provider     `json:"added_providers" bson:"added_providers" yaml:"added_providers"`
	UpdatedProviders    []Provider     `json:"updated_providers" bson:"updated_providers" yaml:"updated_providers"`
	RemovedProviders    []string       `json:"removed_providers" bson:"removed_providers" yaml:"removed_providers"` // Provider IDs
	Manifest            SignedManifest `json:"manifest" bson:"manifest" yaml:"manifest"`
}
