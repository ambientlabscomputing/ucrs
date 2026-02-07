package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ambientlabscomputing/underleaf/capability_registry_service/repository"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/types"
)

type mockProviderRepoForService struct {
	createFunc           func(ctx context.Context, provider *types.Provider) error
	getByIDFunc          func(ctx context.Context, providerID, version string) (*types.Provider, error)
	updateFunc           func(ctx context.Context, provider *types.Provider) error
	deleteFunc           func(ctx context.Context, providerID, version string) error
	queryFunc            func(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]types.Provider, int, error)
	listFunc             func(ctx context.Context, limit, offset int) ([]types.Provider, int, error)
	getByCapabilityFunc  func(ctx context.Context, capabilityID string) ([]types.Provider, error)
	getLatestVersionFunc func(ctx context.Context, providerID string) (*types.Provider, error)
}

func (m *mockProviderRepoForService) Create(ctx context.Context, provider *types.Provider) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, provider)
	}
	return nil
}

func (m *mockProviderRepoForService) GetByID(ctx context.Context, providerID, version string) (*types.Provider, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, providerID, version)
	}
	return nil, nil
}

func (m *mockProviderRepoForService) Update(ctx context.Context, provider *types.Provider) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, provider)
	}
	return nil
}

func (m *mockProviderRepoForService) Delete(ctx context.Context, providerID, version string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, providerID, version)
	}
	return nil
}

func (m *mockProviderRepoForService) Query(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]types.Provider, int, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, filters, limit, offset)
	}
	return []types.Provider{}, 0, nil
}

func (m *mockProviderRepoForService) List(ctx context.Context, limit, offset int) ([]types.Provider, int, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, offset)
	}
	return []types.Provider{}, 0, nil
}

func (m *mockProviderRepoForService) GetByCapability(ctx context.Context, capabilityID string) ([]types.Provider, error) {
	if m.getByCapabilityFunc != nil {
		return m.getByCapabilityFunc(ctx, capabilityID)
	}
	return []types.Provider{}, nil
}

func (m *mockProviderRepoForService) GetLatestVersion(ctx context.Context, providerID string) (*types.Provider, error) {
	if m.getLatestVersionFunc != nil {
		return m.getLatestVersionFunc(ctx, providerID)
	}
	return nil, nil
}

type mockRepoForProvider struct {
	capabilitiesRepo *mockCapabilityRepo
	providersRepo    *mockProviderRepoForService
	snapshotsRepo    *mockSnapshotRepo
}

func (m *mockRepoForProvider) Capabilities() repository.CapabilityRepository {
	return m.capabilitiesRepo
}

func (m *mockRepoForProvider) Providers() repository.ProviderRepository {
	return m.providersRepo
}

func (m *mockRepoForProvider) Snapshots() repository.SnapshotRepository {
	return m.snapshotsRepo
}

// Test Create provider
func TestProviderService_Create(t *testing.T) {
	ctx, settings := setupTest()

	t.Run("successful creation", func(t *testing.T) {
		var createdProvider *types.Provider
		mockRepo := &mockRepoForProvider{
			providersRepo: &mockProviderRepoForService{
				createFunc: func(ctx context.Context, provider *types.Provider) error {
					createdProvider = provider
					return nil
				},
			},
		}

		svc := NewProviderService(mockRepo, settings, nil)

		req := &types.CreateProviderRequest{
			ProviderID:     "test.provider",
			Version:        "1.0.0",
			TrustTier:      types.TrustCertified,
			Artifact:       types.Artifact{Type: types.ArtifactOCI, URI: "docker.io/test/provider:1.0.0"},
			Capabilities:   []types.CapabilityRef{{ID: "cap1"}, {ID: "cap2"}},
			SandboxProfile: "default",
		}

		result, err := svc.Create(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected result, got nil")
		}

		if result.ProviderID != req.ProviderID {
			t.Errorf("expected ID %s, got %s", req.ProviderID, result.ProviderID)
		}

		if createdProvider == nil {
			t.Fatal("create was not called")
		}

		if createdProvider.TrustTier != req.TrustTier {
			t.Errorf("expected trust tier %v, got %v", req.TrustTier, createdProvider.TrustTier)
		}

		if len(createdProvider.Capabilities) != 2 {
			t.Errorf("expected 2 capabilities, got %d", len(createdProvider.Capabilities))
		}
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo := &mockRepoForProvider{
			providersRepo: &mockProviderRepoForService{
				createFunc: func(ctx context.Context, provider *types.Provider) error {
					return expectedErr
				},
			},
		}

		svc := NewProviderService(mockRepo, settings, nil)

		req := &types.CreateProviderRequest{
			ProviderID:     "test.provider",
			Version:        "1.0.0",
			TrustTier:      types.TrustCertified,
			Artifact:       types.Artifact{Type: types.ArtifactOCI, URI: "docker.io/test/provider:1.0.0"},
			Capabilities:   []types.CapabilityRef{},
			SandboxProfile: "default",
		}

		_, err := svc.Create(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

// Test Get provider
func TestProviderService_Get(t *testing.T) {
	ctx, settings := setupTest()

	t.Run("provider found", func(t *testing.T) {
		expected := &types.Provider{
			ProviderID: "test.provider",
			Version:    "1.0.0",
			TrustTier:  types.TrustOfficial,
			Artifact:   types.Artifact{Type: types.ArtifactBinary},
		}

		mockRepo := &mockRepoForProvider{
			providersRepo: &mockProviderRepoForService{
				getByIDFunc: func(ctx context.Context, providerID, version string) (*types.Provider, error) {
					if providerID == "test.provider" && version == "1.0.0" {
						return expected, nil
					}
					return nil, nil
				},
			},
		}

		svc := NewProviderService(mockRepo, settings, nil)

		result, err := svc.Get(ctx, "test.provider", "1.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected result, got nil")
		}

		if result.ProviderID != expected.ProviderID {
			t.Errorf("expected ID %s, got %s", expected.ProviderID, result.ProviderID)
		}

		if result.TrustTier != expected.TrustTier {
			t.Errorf("expected trust tier %v, got %v", expected.TrustTier, result.TrustTier)
		}
	})

	t.Run("provider not found", func(t *testing.T) {
		mockRepo := &mockRepoForProvider{
			providersRepo: &mockProviderRepoForService{
				getByIDFunc: func(ctx context.Context, providerID, version string) (*types.Provider, error) {
					return nil, nil
				},
			},
		}

		svc := NewProviderService(mockRepo, settings, nil)

		result, err := svc.Get(ctx, "nonexistent", "1.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != nil {
			t.Error("expected nil result for not found")
		}
	})
}

// Test Update provider
func TestProviderService_Update(t *testing.T) {
	ctx, settings := setupTest()

	t.Run("successful update", func(t *testing.T) {
		existing := &types.Provider{
			ProviderID: "test.provider",
			Version:    "1.0.0",
			TrustTier:  types.TrustCommunity,
			Artifact:   types.Artifact{Type: types.ArtifactOCI},
		}

		var updatedProvider *types.Provider
		mockRepo := &mockRepoForProvider{
			providersRepo: &mockProviderRepoForService{
				getByIDFunc: func(ctx context.Context, providerID, version string) (*types.Provider, error) {
					if providerID == "test.provider" {
						return existing, nil
					}
					return nil, nil
				},
				updateFunc: func(ctx context.Context, provider *types.Provider) error {
					updatedProvider = provider
					return nil
				},
			},
		}

		svc := NewProviderService(mockRepo, settings, nil)

		newTier := types.TrustCertified
		req := &types.UpdateProviderRequest{
			TrustTier: &newTier,
		}

		result, err := svc.Update(ctx, "test.provider", "1.0.0", req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected result, got nil")
		}

		if result.TrustTier != newTier {
			t.Errorf("expected trust tier %v, got %v", newTier, result.TrustTier)
		}

		if updatedProvider == nil {
			t.Fatal("update was not called")
		}
	})

	t.Run("provider not found", func(t *testing.T) {
		mockRepo := &mockRepoForProvider{
			providersRepo: &mockProviderRepoForService{
				getByIDFunc: func(ctx context.Context, providerID, version string) (*types.Provider, error) {
					return nil, nil
				},
			},
		}

		svc := NewProviderService(mockRepo, settings, nil)

		newTier := types.TrustCertified
		req := &types.UpdateProviderRequest{
			TrustTier: &newTier,
		}

		result, err := svc.Update(ctx, "nonexistent", "1.0.0", req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != nil {
			t.Error("expected nil result for not found")
		}
	})
}

// Test Delete provider
func TestProviderService_Delete(t *testing.T) {
	ctx, settings := setupTest()

	t.Run("successful deletion", func(t *testing.T) {
		deletedID := ""
		deletedVersion := ""
		mockRepo := &mockRepoForProvider{
			providersRepo: &mockProviderRepoForService{
				deleteFunc: func(ctx context.Context, providerID, version string) error {
					deletedID = providerID
					deletedVersion = version
					return nil
				},
			},
		}

		svc := NewProviderService(mockRepo, settings, nil)

		err := svc.Delete(ctx, "test.provider", "1.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if deletedID != "test.provider" {
			t.Errorf("expected deleted ID 'test.provider', got '%s'", deletedID)
		}

		if deletedVersion != "1.0.0" {
			t.Errorf("expected deleted version '1.0.0', got '%s'", deletedVersion)
		}
	})
}

// Test List providers
func TestProviderService_List(t *testing.T) {
	ctx, settings := setupTest()

	t.Run("list with results", func(t *testing.T) {
		providers := []types.Provider{
			{ProviderID: "prov1", Version: "1.0.0", TrustTier: types.TrustOfficial},
			{ProviderID: "prov2", Version: "1.0.0", TrustTier: types.TrustCertified},
			{ProviderID: "prov3", Version: "2.0.0", TrustTier: types.TrustCommunity},
		}

		mockRepo := &mockRepoForProvider{
			providersRepo: &mockProviderRepoForService{
				listFunc: func(ctx context.Context, limit, offset int) ([]types.Provider, int, error) {
					return providers, 3, nil
				},
			},
		}

		svc := NewProviderService(mockRepo, settings, nil)

		results, total, err := svc.List(ctx, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}

		if total != 3 {
			t.Errorf("expected total 3, got %d", total)
		}
	})
}

// Test Query providers
func TestProviderService_Query(t *testing.T) {
	ctx, settings := setupTest()

	t.Run("query with filters", func(t *testing.T) {
		var capturedFilters map[string]interface{}

		mockRepo := &mockRepoForProvider{
			providersRepo: &mockProviderRepoForService{
				queryFunc: func(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]types.Provider, int, error) {
					capturedFilters = filters
					return []types.Provider{
						{ProviderID: "filtered", Version: "1.0.0", TrustTier: types.TrustOfficial},
					}, 1, nil
				},
			},
		}

		svc := NewProviderService(mockRepo, settings, nil)

		tier := types.TrustOfficial
		req := &types.QueryProvidersRequest{
			TrustTier: &tier,
			Limit:     10,
			Offset:    0,
		}

		results, total, err := svc.Query(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d", len(results))
		}

		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}

		if capturedFilters["trust_tier"] != types.TrustOfficial {
			t.Error("filter not passed correctly")
		}
	})
}
