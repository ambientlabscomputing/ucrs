package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/ambientlabscomputing/ucrs/repository"
	"github.com/ambientlabscomputing/ucrs/sdk/types"
	"github.com/ambientlabscomputing/ucrs/utils"
)

type ProviderService struct {
	repo     repository.Repository
	settings *utils.Settings
	sync     *SyncService
}

func NewProviderService(repo repository.Repository, settings *utils.Settings, sync *SyncService) *ProviderService {
	return &ProviderService{
		repo:     repo,
		settings: settings,
		sync:     sync,
	}
}

func (s *ProviderService) refreshSnapshot(ctx context.Context) {
	if s.sync == nil {
		return
	}

	logger := utils.GetLogger(ctx)
	if _, err := s.sync.GenerateSnapshot(ctx); err != nil {
		logger.Error("failed to regenerate snapshot", "error", err)
	}
}

// validateArtifact validates artifact configuration for binary providers
func (s *ProviderService) validateArtifact(ctx context.Context, artifact *types.Artifact) error {
	logger := utils.GetLogger(ctx)

	// Only binary artifacts need platform validation
	if artifact.Type != types.ArtifactBinary {
		return nil
	}

	hasOSPlaceholder := strings.Contains(artifact.URI, "{os}")
	hasArchPlaceholder := strings.Contains(artifact.URI, "{arch}")
	hasTemplates := hasOSPlaceholder || hasArchPlaceholder

	// If URI contains platform templates, require supported_platforms
	if hasTemplates && len(artifact.SupportedPlatforms) == 0 {
		return fmt.Errorf("binary artifact with {os}/{arch} templates must declare supported_platforms")
	}

	// If no templates but supported_platforms declared, warn (unusual but valid for single-platform binaries)
	if !hasTemplates && len(artifact.SupportedPlatforms) > 0 {
		logger.Warn("artifact has supported_platforms but URI has no {os}/{arch} templates - this may be intentional for single-platform binaries")
	}

	// Validate supported_platforms entries
	validOSes := map[string]bool{"darwin": true, "linux": true, "windows": true}
	validArches := map[string]bool{"amd64": true, "arm64": true, "386": true, "arm": true}

	for _, platform := range artifact.SupportedPlatforms {
		if !validOSes[platform.OS] {
			return fmt.Errorf("unsupported OS in supported_platforms: %s (valid: darwin, linux, windows)", platform.OS)
		}
		if !validArches[platform.Arch] {
			return fmt.Errorf("unsupported architecture in supported_platforms: %s (valid: amd64, arm64, 386, arm)", platform.Arch)
		}
	}

	// If multi-platform (>1 entry), require checksums for verification
	if len(artifact.SupportedPlatforms) > 1 {
		if artifact.Checksums == nil || artifact.Checksums.URI == "" {
			return fmt.Errorf("multi-platform binary artifacts must provide a checksums.uri for verification")
		}
	}

	return nil
}

func (s *ProviderService) Create(ctx context.Context, req *types.CreateProviderRequest) (*types.Provider, error) {
	logger := utils.GetLogger(ctx)

	// Validate artifact configuration
	if err := s.validateArtifact(ctx, &req.Artifact); err != nil {
		logger.Error("artifact validation failed", "error", err, "provider_id", req.ProviderID)
		return nil, fmt.Errorf("invalid artifact: %w", err)
	}

	provider := &types.Provider{
		ProviderID:          req.ProviderID,
		Version:             req.Version,
		Capabilities:        req.Capabilities,
		Artifact:            req.Artifact,
		TrustTier:           req.TrustTier,
		SandboxProfile:      req.SandboxProfile,
		RuntimeRequirements: req.RuntimeRequirements,
		Description:         req.Description,
		Maintainer:          req.Maintainer,
		Homepage:            req.Homepage,
	}

	if err := s.repo.Providers().Create(ctx, provider); err != nil {
		logger.Error("failed to create provider", "error", err, "provider_id", req.ProviderID)
		return nil, err
	}

	s.refreshSnapshot(ctx)

	logger.Info("provider created", "provider_id", provider.ProviderID, "version", provider.Version)
	return provider, nil
}

func (s *ProviderService) Get(ctx context.Context, providerID, version string) (*types.Provider, error) {
	logger := utils.GetLogger(ctx)

	provider, err := s.repo.Providers().GetByID(ctx, providerID, version)
	if err != nil {
		logger.Error("failed to get provider", "error", err, "provider_id", providerID, "version", version)
		return nil, err
	}

	if provider == nil {
		logger.Debug("provider not found", "provider_id", providerID, "version", version)
		return nil, nil
	}

	return provider, nil
}

func (s *ProviderService) Update(ctx context.Context, providerID, version string, req *types.UpdateProviderRequest) (*types.Provider, error) {
	logger := utils.GetLogger(ctx)

	provider, err := s.repo.Providers().GetByID(ctx, providerID, version)
	if err != nil {
		logger.Error("failed to get provider for update", "error", err, "provider_id", providerID)
		return nil, err
	}

	if provider == nil {
		logger.Debug("provider not found for update", "provider_id", providerID)
		return nil, nil
	}

	// Apply updates
	if req.Artifact != nil {
		// Validate artifact before applying
		if err := s.validateArtifact(ctx, req.Artifact); err != nil {
			logger.Error("artifact validation failed", "error", err, "provider_id", providerID)
			return nil, fmt.Errorf("invalid artifact: %w", err)
		}
		provider.Artifact = *req.Artifact
	}
	if req.TrustTier != nil {
		provider.TrustTier = *req.TrustTier
	}
	if req.SandboxProfile != nil {
		provider.SandboxProfile = *req.SandboxProfile
	}
	if req.RuntimeRequirements != nil {
		provider.RuntimeRequirements = *req.RuntimeRequirements
	}
	if req.Description != nil {
		provider.Description = *req.Description
	}
	if req.Maintainer != nil {
		provider.Maintainer = *req.Maintainer
	}
	if req.Homepage != nil {
		provider.Homepage = *req.Homepage
	}

	if err := s.repo.Providers().Update(ctx, provider); err != nil {
		logger.Error("failed to update provider", "error", err, "provider_id", providerID)
		return nil, err
	}

	s.refreshSnapshot(ctx)

	logger.Info("provider updated", "provider_id", provider.ProviderID)
	return provider, nil
}

func (s *ProviderService) Delete(ctx context.Context, providerID, version string) error {
	logger := utils.GetLogger(ctx)

	if err := s.repo.Providers().Delete(ctx, providerID, version); err != nil {
		logger.Error("failed to delete provider", "error", err, "provider_id", providerID, "version", version)
		return err
	}

	s.refreshSnapshot(ctx)

	logger.Info("provider deleted", "provider_id", providerID, "version", version)
	return nil
}

func (s *ProviderService) Query(ctx context.Context, req *types.QueryProvidersRequest) ([]types.Provider, int, error) {
	logger := utils.GetLogger(ctx)

	filters := make(map[string]interface{})
	if req.CapabilityID != nil {
		filters["capabilities.id"] = *req.CapabilityID
	}
	if req.TrustTier != nil {
		filters["trust_tier"] = *req.TrustTier
	}
	if req.ArtifactType != nil {
		filters["artifact.type"] = *req.ArtifactType
	}

	providers, total, err := s.repo.Providers().Query(ctx, filters, req.Limit, req.Offset)
	if err != nil {
		logger.Error("failed to query providers", "error", err)
		return nil, 0, err
	}

	logger.Debug("providers queried", "count", len(providers), "total", total)
	return providers, total, nil
}

func (s *ProviderService) List(ctx context.Context, limit, offset int) ([]types.Provider, int, error) {
	logger := utils.GetLogger(ctx)

	providers, total, err := s.repo.Providers().List(ctx, limit, offset)
	if err != nil {
		logger.Error("failed to list providers", "error", err)
		return nil, 0, err
	}

	logger.Debug("providers listed", "count", len(providers), "total", total)
	return providers, total, nil
}
