package service

import (
	"context"

	"github.com/ambientlabscomputing/underleaf/capability_registry_service/repository"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/types"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/utils"
)

type CapabilityService struct {
	repo     repository.Repository
	settings *utils.Settings
	sync     *SyncService
}

func NewCapabilityService(repo repository.Repository, settings *utils.Settings, sync *SyncService) *CapabilityService {
	return &CapabilityService{
		repo:     repo,
		settings: settings,
		sync:     sync,
	}
}

func (s *CapabilityService) refreshSnapshot(ctx context.Context) {
	if s.sync == nil {
		return
	}

	logger := utils.GetLogger(ctx)
	if _, err := s.sync.GenerateSnapshot(ctx); err != nil {
		logger.Error("failed to regenerate snapshot", "error", err)
	}
}

func (s *CapabilityService) Create(ctx context.Context, req *types.CreateCapabilityRequest) (*types.Capability, error) {
	logger := utils.GetLogger(ctx)

	capability := &types.Capability{
		ID:                req.ID,
		Version:           req.Version,
		Description:       req.Description,
		RiskClass:         req.RiskClass,
		AllowedVerbs:      req.AllowedVerbs,
		PermissionClass:   req.PermissionClass,
		DeprecationPolicy: req.DeprecationPolicy,
	}

	if err := s.repo.Capabilities().Create(ctx, capability); err != nil {
		logger.Error("failed to create capability", "error", err, "id", req.ID)
		return nil, err
	}

	s.refreshSnapshot(ctx)

	logger.Info("capability created", "id", capability.ID, "version", capability.Version)
	return capability, nil
}

func (s *CapabilityService) Get(ctx context.Context, id string) (*types.Capability, error) {
	logger := utils.GetLogger(ctx)

	capability, err := s.repo.Capabilities().GetByID(ctx, id)
	if err != nil {
		logger.Error("failed to get capability", "error", err, "id", id)
		return nil, err
	}

	if capability == nil {
		logger.Debug("capability not found", "id", id)
		return nil, nil
	}

	return capability, nil
}

func (s *CapabilityService) Update(ctx context.Context, id string, req *types.UpdateCapabilityRequest) (*types.Capability, error) {
	logger := utils.GetLogger(ctx)

	capability, err := s.repo.Capabilities().GetByID(ctx, id)
	if err != nil {
		logger.Error("failed to get capability for update", "error", err, "id", id)
		return nil, err
	}

	if capability == nil {
		logger.Debug("capability not found for update", "id", id)
		return nil, nil
	}

	// Apply updates
	if req.Description != nil {
		capability.Description = *req.Description
	}
	if req.RiskClass != nil {
		capability.RiskClass = *req.RiskClass
	}
	if req.AllowedVerbs != nil {
		capability.AllowedVerbs = req.AllowedVerbs
	}
	if req.PermissionClass != nil {
		capability.PermissionClass = *req.PermissionClass
	}
	if req.DeprecationPolicy != nil {
		capability.DeprecationPolicy = *req.DeprecationPolicy
	}

	if err := s.repo.Capabilities().Update(ctx, capability); err != nil {
		logger.Error("failed to update capability", "error", err, "id", id)
		return nil, err
	}

	s.refreshSnapshot(ctx)

	logger.Info("capability updated", "id", capability.ID)
	return capability, nil
}

func (s *CapabilityService) Delete(ctx context.Context, id string) error {
	logger := utils.GetLogger(ctx)

	if err := s.repo.Capabilities().Delete(ctx, id); err != nil {
		logger.Error("failed to delete capability", "error", err, "id", id)
		return err
	}

	s.refreshSnapshot(ctx)

	logger.Info("capability deleted", "id", id)
	return nil
}

func (s *CapabilityService) Query(ctx context.Context, req *types.QueryCapabilitiesRequest) ([]types.Capability, int, error) {
	logger := utils.GetLogger(ctx)

	filters := make(map[string]interface{})
	if req.RiskClass != nil {
		filters["risk_class"] = *req.RiskClass
	}
	if req.PermissionClass != nil {
		filters["permission_class"] = *req.PermissionClass
	}
	if req.Verb != nil {
		filters["allowed_verbs"] = *req.Verb
	}

	capabilities, total, err := s.repo.Capabilities().Query(ctx, filters, req.Limit, req.Offset)
	if err != nil {
		logger.Error("failed to query capabilities", "error", err)
		return nil, 0, err
	}

	logger.Debug("capabilities queried", "count", len(capabilities), "total", total)
	return capabilities, total, nil
}

func (s *CapabilityService) List(ctx context.Context, limit, offset int) ([]types.Capability, int, error) {
	logger := utils.GetLogger(ctx)

	capabilities, total, err := s.repo.Capabilities().List(ctx, limit, offset)
	if err != nil {
		logger.Error("failed to list capabilities", "error", err)
		return nil, 0, err
	}

	logger.Debug("capabilities listed", "count", len(capabilities), "total", total)
	return capabilities, total, nil
}
