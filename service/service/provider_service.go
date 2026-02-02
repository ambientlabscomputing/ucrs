package service

import (
	"context"

	"github.com/ambientlabscomputing/underleaf/capability_registry_service/repository"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/types"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/utils"
)

type ProviderService struct {
	repo     repository.Repository
	settings *utils.Settings
}

func NewProviderService(repo repository.Repository, settings *utils.Settings) *ProviderService {
	return &ProviderService{
		repo:     repo,
		settings: settings,
	}
}

func (s *ProviderService) Create(ctx context.Context, req *types.CreateProviderRequest) (*types.Provider, error) {
	logger := utils.GetLogger(ctx)

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

	logger.Info("provider updated", "provider_id", provider.ProviderID)
	return provider, nil
}

func (s *ProviderService) Delete(ctx context.Context, providerID, version string) error {
	logger := utils.GetLogger(ctx)

	if err := s.repo.Providers().Delete(ctx, providerID, version); err != nil {
		logger.Error("failed to delete provider", "error", err, "provider_id", providerID, "version", version)
		return err
	}

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
