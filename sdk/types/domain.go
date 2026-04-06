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

// LaunchMode controls how the agent starts a provider after installation.
type LaunchMode string

const (
	// LaunchModeDaemon starts the provider as a long-running supervised process (default).
	LaunchModeDaemon LaunchMode = "daemon"
	// LaunchModeOnDemand installs the binary to disk but does NOT auto-start it.
	// The invoker (e.g. an IDE over stdio) starts the process on demand.
	LaunchModeOnDemand LaunchMode = "on-demand"
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
	LaunchMode     LaunchMode        `json:"launch_mode,omitempty" bson:"launch_mode,omitempty" yaml:"launch_mode,omitempty"`             // Controls auto-start behaviour; empty or "daemon" = supervised process (default)
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

// ==================== App Manifest (GitHub source deploy) ====================

// AppManifest is the schema for .underleaf/deploy.yaml placed in a GitHub repo.
// It is target-agnostic: no org_id, no server targeting. Those are injected by server_api.
type AppManifest struct {
	Version  string            `json:"version" yaml:"version"`               // "1"
	Name     string            `json:"name" yaml:"name"`                     // human-readable app name
	Slug     string            `json:"slug,omitempty" yaml:"slug,omitempty"` // unique slug; auto-derived from name if empty
	Services []ManifestService `json:"services" yaml:"services"`
	Networks []ManifestNetwork `json:"networks,omitempty" yaml:"networks,omitempty"`
	Volumes  []ManifestVolume  `json:"volumes,omitempty" yaml:"volumes,omitempty"`

	// Targeting allows deploy.yaml to declare placement preferences such as
	// replicas count.  Merged with (and overridden by) CLI-level targeting.
	Targeting *ManifestTargeting `json:"targeting,omitempty" yaml:"targeting,omitempty"`
}

// ManifestTargeting is the subset of targeting options that can be set in a
// deploy.yaml manifest.  It intentionally exposes only high-level placement
// knobs; full server/cluster targeting requires CLI flags.
type ManifestTargeting struct {
	Replicas *int `json:"replicas,omitempty" yaml:"replicas,omitempty"`
}

// ManifestService describes a single container service in an AppManifest.
// Exactly one of Image or Build must be set.
type ManifestService struct {
	Name        string            `json:"name" yaml:"name"`
	Image       string            `json:"image,omitempty" yaml:"image,omitempty"`
	Build       *ManifestBuild    `json:"build,omitempty" yaml:"build,omitempty"`
	Ports       []string          `json:"ports,omitempty" yaml:"ports,omitempty"`
	Environment map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
	Networks    []string          `json:"networks,omitempty" yaml:"networks,omitempty"`
	Volumes     []string          `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	Expose      *ManifestExpose   `json:"expose,omitempty" yaml:"expose,omitempty"`
}

// ManifestBuild describes how to build a Docker image from source.
type ManifestBuild struct {
	Context    string            `json:"context,omitempty" yaml:"context,omitempty"`       // relative to repo root (default ".")
	Dockerfile string            `json:"dockerfile,omitempty" yaml:"dockerfile,omitempty"` // default "Dockerfile"
	Args       map[string]string `json:"args,omitempty" yaml:"args,omitempty"`
}

// ManifestExpose declares that a service should be publicly exposed.
type ManifestExpose struct {
	Port     int    `json:"port" yaml:"port"`
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"` // auto-assigned if empty
}

// ManifestNetwork defines a Docker network in an AppManifest.
type ManifestNetwork struct {
	Name   string `json:"name" yaml:"name"`
	Driver string `json:"driver,omitempty" yaml:"driver,omitempty"` // default "bridge"
}

// ManifestVolume defines a Docker volume in an AppManifest.
type ManifestVolume struct {
	Name string `json:"name" yaml:"name"`
}

// ResolvedSource is the result of resolving a gh: source reference.
type ResolvedSource struct {
	Type       string      `json:"type"` // "github"
	Owner      string      `json:"owner"`
	Repo       string      `json:"repo"`
	Ref        string      `json:"ref"`         // resolved ref (may differ from request if default branch was used)
	ArchiveURL string      `json:"archive_url"` // tarball download URL (authenticated if private)
	Manifest   AppManifest `json:"manifest"`
	RepoMeta   RepoMeta    `json:"repo_meta"`
}

// RepoMeta carries public metadata about the resolved GitHub repo.
type RepoMeta struct {
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	Description   string `json:"description,omitempty"`
	Stars         int    `json:"stars,omitempty"`
}
