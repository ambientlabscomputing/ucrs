package types

// CreateCapabilityRequest represents a request to create a new capability
type CreateCapabilityRequest struct {
	ID                string    `json:"id" binding:"required"`
	Version           string    `json:"version" binding:"required"`
	Description       string    `json:"description" binding:"required"`
	RiskClass         RiskClass `json:"risk_class" binding:"required"`
	AllowedVerbs      []string  `json:"allowed_verbs" binding:"required"`
	PermissionClass   string    `json:"permission_class" binding:"required"`
	DeprecationPolicy string    `json:"deprecation_policy,omitempty"`
}

// UpdateCapabilityRequest represents a request to update an existing capability
type UpdateCapabilityRequest struct {
	Description       *string    `json:"description,omitempty"`
	RiskClass         *RiskClass `json:"risk_class,omitempty"`
	AllowedVerbs      []string   `json:"allowed_verbs,omitempty"`
	PermissionClass   *string    `json:"permission_class,omitempty"`
	DeprecationPolicy *string    `json:"deprecation_policy,omitempty"`
}

// CreateProviderRequest represents a request to create a new provider
type CreateProviderRequest struct {
	ProviderID          string              `json:"provider_id" binding:"required"`
	Version             string              `json:"version" binding:"required"`
	Capabilities        []CapabilityRef     `json:"capabilities" binding:"required"`
	Artifact            Artifact            `json:"artifact" binding:"required"`
	TrustTier           TrustTier           `json:"trust_tier" binding:"required"`
	SandboxProfile      string              `json:"sandbox_profile" binding:"required"`
	RuntimeRequirements RuntimeRequirements `json:"runtime_requirements"`
	Description         string              `json:"description,omitempty"`
	Maintainer          string              `json:"maintainer,omitempty"`
	Homepage            string              `json:"homepage,omitempty"`
}

// UpdateProviderRequest represents a request to update an existing provider
type UpdateProviderRequest struct {
	Artifact            *Artifact            `json:"artifact,omitempty"`
	TrustTier           *TrustTier           `json:"trust_tier,omitempty"`
	SandboxProfile      *string              `json:"sandbox_profile,omitempty"`
	RuntimeRequirements *RuntimeRequirements `json:"runtime_requirements,omitempty"`
	Description         *string              `json:"description,omitempty"`
	Maintainer          *string              `json:"maintainer,omitempty"`
	Homepage            *string              `json:"homepage,omitempty"`
}

// QueryCapabilitiesRequest represents query parameters for capabilities
type QueryCapabilitiesRequest struct {
	RiskClass       *RiskClass `form:"risk_class"`
	PermissionClass *string    `form:"permission_class"`
	Verb            *string    `form:"verb"`
	Limit           int        `form:"limit" binding:"min=1,max=1000"`
	Offset          int        `form:"offset" binding:"min=0"`
}

// QueryProvidersRequest represents query parameters for providers
type QueryProvidersRequest struct {
	CapabilityID *string       `form:"capability_id"`
	TrustTier    *TrustTier    `form:"trust_tier"`
	ArtifactType *ArtifactType `form:"artifact_type"`
	Limit        int           `form:"limit" binding:"min=1,max=1000"`
	Offset       int           `form:"offset" binding:"min=0"`
}

// SyncSnapshotResponse wraps the registry snapshot for API responses
type SyncSnapshotResponse struct {
	Snapshot RegistrySnapshot `json:"snapshot"`
}

// SyncDeltaRequest represents query parameters for delta sync
type SyncDeltaRequest struct {
	Since string `form:"since" binding:"required"` // Timestamp or version
}

// SyncDeltaResponse wraps the delta update for API responses
type SyncDeltaResponse struct {
	Delta DeltaUpdate `json:"delta"`
}

// ==================== Source Resolution ====================

// ResolveSourceRequest is the query for GET /sources/resolve
type ResolveSourceRequest struct {
	Source string `form:"source" binding:"required"` // e.g. "gh:owner/repo" or "gh:owner/repo@ref"
	Ref    string `form:"ref"`                       // override ref (optional if encoded in Source)
	Token  string `form:"token"`                     // GitHub PAT for private repos (optional)
}

// ResolveSourceResponse is returned by GET /sources/resolve
type ResolveSourceResponse struct {
	Resolved ResolvedSource `json:"resolved"`
}

// ListResponse is a generic response wrapper for list endpoints
type ListResponse struct {
	Items      interface{} `json:"items"`
	TotalCount int         `json:"total_count"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
}
