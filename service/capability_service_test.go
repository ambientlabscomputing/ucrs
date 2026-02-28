package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ambientlabscomputing/ucrs/repository"
	"github.com/ambientlabscomputing/ucrs/sdk/types"
	"github.com/ambientlabscomputing/ucrs/utils"
)

// Mock implementations for testing
type mockCapabilityRepo struct {
	createFunc         func(ctx context.Context, capability *types.Capability) error
	getByIDFunc        func(ctx context.Context, id string) (*types.Capability, error)
	updateFunc         func(ctx context.Context, capability *types.Capability) error
	deleteFunc         func(ctx context.Context, id string) error
	queryFunc          func(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]types.Capability, int, error)
	listFunc           func(ctx context.Context, limit, offset int) ([]types.Capability, int, error)
	getAllVersionsFunc func(ctx context.Context, id string) ([]types.Capability, error)
}

func (m *mockCapabilityRepo) Create(ctx context.Context, capability *types.Capability) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, capability)
	}
	return nil
}

func (m *mockCapabilityRepo) GetByID(ctx context.Context, id string) (*types.Capability, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockCapabilityRepo) Update(ctx context.Context, capability *types.Capability) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, capability)
	}
	return nil
}

func (m *mockCapabilityRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockCapabilityRepo) Query(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]types.Capability, int, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, filters, limit, offset)
	}
	return []types.Capability{}, 0, nil
}

func (m *mockCapabilityRepo) List(ctx context.Context, limit, offset int) ([]types.Capability, int, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, offset)
	}
	return []types.Capability{}, 0, nil
}

func (m *mockCapabilityRepo) GetAllVersions(ctx context.Context, id string) ([]types.Capability, error) {
	if m.getAllVersionsFunc != nil {
		return m.getAllVersionsFunc(ctx, id)
	}
	return []types.Capability{}, nil
}

type mockProviderRepo struct{}

func (m *mockProviderRepo) Create(ctx context.Context, provider *types.Provider) error { return nil }
func (m *mockProviderRepo) GetByID(ctx context.Context, providerID, version string) (*types.Provider, error) {
	return nil, nil
}
func (m *mockProviderRepo) Update(ctx context.Context, provider *types.Provider) error { return nil }
func (m *mockProviderRepo) Delete(ctx context.Context, providerID, version string) error {
	return nil
}
func (m *mockProviderRepo) Query(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]types.Provider, int, error) {
	return []types.Provider{}, 0, nil
}
func (m *mockProviderRepo) List(ctx context.Context, limit, offset int) ([]types.Provider, int, error) {
	return []types.Provider{}, 0, nil
}
func (m *mockProviderRepo) GetByCapability(ctx context.Context, capabilityID string) ([]types.Provider, error) {
	return []types.Provider{}, nil
}
func (m *mockProviderRepo) GetLatestVersion(ctx context.Context, providerID string) (*types.Provider, error) {
	return nil, nil
}

type mockSnapshotRepo struct{}

func (m *mockSnapshotRepo) Create(ctx context.Context, snapshot *types.RegistrySnapshot) error {
	return nil
}
func (m *mockSnapshotRepo) GetLatest(ctx context.Context) (*types.RegistrySnapshot, error) {
	return nil, nil
}
func (m *mockSnapshotRepo) GetByVersion(ctx context.Context, version string) (*types.RegistrySnapshot, error) {
	return nil, nil
}
func (m *mockSnapshotRepo) GetByETag(ctx context.Context, etag string) (*types.RegistrySnapshot, error) {
	return nil, nil
}
func (m *mockSnapshotRepo) ListVersions(ctx context.Context, limit int) ([]string, error) {
	return []string{}, nil
}
func (m *mockSnapshotRepo) GetDeltaSince(ctx context.Context, sinceVersion string) (*types.DeltaUpdate, error) {
	return nil, nil
}

type mockRepository struct {
	capabilitiesRepo *mockCapabilityRepo
	providersRepo    *mockProviderRepo
	snapshotsRepo    *mockSnapshotRepo
}

func (m *mockRepository) Capabilities() repository.CapabilityRepository {
	return m.capabilitiesRepo
}

func (m *mockRepository) Providers() repository.ProviderRepository {
	return m.providersRepo
}

func (m *mockRepository) Snapshots() repository.SnapshotRepository {
	return m.snapshotsRepo
}

func setupTest() (context.Context, *utils.Settings) {
	ctx := context.Background()
	settings := &utils.Settings{
		LogLevel:    "error",
		LogToStderr: true,
		LogToFile:   false,
	}
	utils.InitLogger(settings)
	return ctx, settings
}

// Test Create capability
func TestCapabilityService_Create(t *testing.T) {
	ctx, settings := setupTest()

	t.Run("successful creation", func(t *testing.T) {
		var createdCap *types.Capability
		mockRepo := &mockRepository{
			capabilitiesRepo: &mockCapabilityRepo{
				createFunc: func(ctx context.Context, capability *types.Capability) error {
					createdCap = capability
					return nil
				},
			},
		}

		svc := NewCapabilityService(mockRepo, settings, nil)

		req := &types.CreateCapabilityRequest{
			ID:              "test.capability",
			Version:         "1.0.0",
			Description:     "Test capability",
			RiskClass:       types.RiskLow,
			AllowedVerbs:    []string{"read", "write"},
			PermissionClass: "test.permission",
		}

		result, err := svc.Create(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected result, got nil")
		}

		if result.ID != req.ID {
			t.Errorf("expected ID %s, got %s", req.ID, result.ID)
		}

		if createdCap == nil {
			t.Fatal("create was not called")
		}

		if createdCap.Version != req.Version {
			t.Errorf("expected version %s, got %s", req.Version, createdCap.Version)
		}

		if createdCap.RiskClass != req.RiskClass {
			t.Errorf("expected risk class %v, got %v", req.RiskClass, createdCap.RiskClass)
		}

		if len(createdCap.AllowedVerbs) != 2 {
			t.Errorf("expected 2 allowed verbs, got %d", len(createdCap.AllowedVerbs))
		}
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo := &mockRepository{
			capabilitiesRepo: &mockCapabilityRepo{
				createFunc: func(ctx context.Context, capability *types.Capability) error {
					return expectedErr
				},
			},
		}

		svc := NewCapabilityService(mockRepo, settings, nil)

		req := &types.CreateCapabilityRequest{
			ID:              "test.capability",
			Version:         "1.0.0",
			Description:     "Test",
			RiskClass:       types.RiskLow,
			AllowedVerbs:    []string{"read"},
			PermissionClass: "test.permission",
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

// Test Get capability
func TestCapabilityService_Get(t *testing.T) {
	ctx, settings := setupTest()

	t.Run("capability found", func(t *testing.T) {
		expected := &types.Capability{
			ID:              "test.capability",
			Version:         "1.0.0",
			Description:     "Test capability",
			RiskClass:       types.RiskMedium,
			AllowedVerbs:    []string{"read"},
			PermissionClass: "test.permission",
		}

		mockRepo := &mockRepository{
			capabilitiesRepo: &mockCapabilityRepo{
				getByIDFunc: func(ctx context.Context, id string) (*types.Capability, error) {
					if id == "test.capability" {
						return expected, nil
					}
					return nil, nil
				},
			},
		}

		svc := NewCapabilityService(mockRepo, settings, nil)

		result, err := svc.Get(ctx, "test.capability")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected result, got nil")
		}

		if result.ID != expected.ID {
			t.Errorf("expected ID %s, got %s", expected.ID, result.ID)
		}

		if result.RiskClass != expected.RiskClass {
			t.Errorf("expected risk class %v, got %v", expected.RiskClass, result.RiskClass)
		}
	})

	t.Run("capability not found", func(t *testing.T) {
		mockRepo := &mockRepository{
			capabilitiesRepo: &mockCapabilityRepo{
				getByIDFunc: func(ctx context.Context, id string) (*types.Capability, error) {
					return nil, nil
				},
			},
		}

		svc := NewCapabilityService(mockRepo, settings, nil)

		result, err := svc.Get(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != nil {
			t.Error("expected nil result for not found")
		}
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo := &mockRepository{
			capabilitiesRepo: &mockCapabilityRepo{
				getByIDFunc: func(ctx context.Context, id string) (*types.Capability, error) {
					return nil, expectedErr
				},
			},
		}

		svc := NewCapabilityService(mockRepo, settings, nil)

		_, err := svc.Get(ctx, "test.capability")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

// Test Update capability
func TestCapabilityService_Update(t *testing.T) {
	ctx, settings := setupTest()

	t.Run("successful update", func(t *testing.T) {
		existing := &types.Capability{
			ID:              "test.capability",
			Version:         "1.0.0",
			Description:     "Old description",
			RiskClass:       types.RiskLow,
			AllowedVerbs:    []string{"read"},
			PermissionClass: "test.permission",
		}

		var updatedCap *types.Capability
		mockRepo := &mockRepository{
			capabilitiesRepo: &mockCapabilityRepo{
				getByIDFunc: func(ctx context.Context, id string) (*types.Capability, error) {
					if id == "test.capability" {
						return existing, nil
					}
					return nil, nil
				},
				updateFunc: func(ctx context.Context, capability *types.Capability) error {
					updatedCap = capability
					return nil
				},
			},
		}

		svc := NewCapabilityService(mockRepo, settings, nil)

		newDesc := "New description"
		newRisk := types.RiskHigh
		req := &types.UpdateCapabilityRequest{
			Description: &newDesc,
			RiskClass:   &newRisk,
		}

		result, err := svc.Update(ctx, "test.capability", req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected result, got nil")
		}

		if result.Description != newDesc {
			t.Errorf("expected description '%s', got '%s'", newDesc, result.Description)
		}

		if result.RiskClass != newRisk {
			t.Errorf("expected risk class %v, got %v", newRisk, result.RiskClass)
		}

		if updatedCap == nil {
			t.Fatal("update was not called")
		}
	})

	t.Run("capability not found", func(t *testing.T) {
		mockRepo := &mockRepository{
			capabilitiesRepo: &mockCapabilityRepo{
				getByIDFunc: func(ctx context.Context, id string) (*types.Capability, error) {
					return nil, nil
				},
			},
		}

		svc := NewCapabilityService(mockRepo, settings, nil)

		newDesc := "New description"
		req := &types.UpdateCapabilityRequest{
			Description: &newDesc,
		}

		result, err := svc.Update(ctx, "nonexistent", req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != nil {
			t.Error("expected nil result for not found")
		}
	})

	t.Run("update only specific fields", func(t *testing.T) {
		existing := &types.Capability{
			ID:              "test.capability",
			Version:         "1.0.0",
			Description:     "Original",
			RiskClass:       types.RiskLow,
			AllowedVerbs:    []string{"read"},
			PermissionClass: "test.permission",
		}

		mockRepo := &mockRepository{
			capabilitiesRepo: &mockCapabilityRepo{
				getByIDFunc: func(ctx context.Context, id string) (*types.Capability, error) {
					return existing, nil
				},
				updateFunc: func(ctx context.Context, capability *types.Capability) error {
					return nil
				},
			},
		}

		svc := NewCapabilityService(mockRepo, settings, nil)

		newVerbs := []string{"read", "write", "delete"}
		req := &types.UpdateCapabilityRequest{
			AllowedVerbs: newVerbs,
		}

		result, err := svc.Update(ctx, "test.capability", req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Description should remain unchanged
		if result.Description != "Original" {
			t.Errorf("description should be unchanged, got %s", result.Description)
		}

		// RiskClass should remain unchanged
		if result.RiskClass != types.RiskLow {
			t.Errorf("risk class should be unchanged, got %v", result.RiskClass)
		}

		// AllowedVerbs should be updated
		if len(result.AllowedVerbs) != 3 {
			t.Errorf("expected 3 allowed verbs, got %d", len(result.AllowedVerbs))
		}
	})
}

// Test Delete capability
func TestCapabilityService_Delete(t *testing.T) {
	ctx, settings := setupTest()

	t.Run("successful deletion", func(t *testing.T) {
		deletedID := ""
		mockRepo := &mockRepository{
			capabilitiesRepo: &mockCapabilityRepo{
				deleteFunc: func(ctx context.Context, id string) error {
					deletedID = id
					return nil
				},
			},
		}

		svc := NewCapabilityService(mockRepo, settings, nil)

		err := svc.Delete(ctx, "test.capability")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if deletedID != "test.capability" {
			t.Errorf("expected deleted ID 'test.capability', got '%s'", deletedID)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo := &mockRepository{
			capabilitiesRepo: &mockCapabilityRepo{
				deleteFunc: func(ctx context.Context, id string) error {
					return expectedErr
				},
			},
		}

		svc := NewCapabilityService(mockRepo, settings, nil)

		err := svc.Delete(ctx, "test.capability")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

// Test List capabilities
func TestCapabilityService_List(t *testing.T) {
	ctx, settings := setupTest()

	t.Run("list with results", func(t *testing.T) {
		caps := []types.Capability{
			{ID: "cap1", Version: "1.0.0", Description: "First", RiskClass: types.RiskLow},
			{ID: "cap2", Version: "1.0.0", Description: "Second", RiskClass: types.RiskMedium},
			{ID: "cap3", Version: "2.0.0", Description: "Third", RiskClass: types.RiskHigh},
		}

		mockRepo := &mockRepository{
			capabilitiesRepo: &mockCapabilityRepo{
				listFunc: func(ctx context.Context, limit, offset int) ([]types.Capability, int, error) {
					return caps, 3, nil
				},
			},
		}

		svc := NewCapabilityService(mockRepo, settings, nil)

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

		if results[0].ID != "cap1" {
			t.Errorf("expected first ID 'cap1', got '%s'", results[0].ID)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		var capturedLimit, capturedOffset int

		mockRepo := &mockRepository{
			capabilitiesRepo: &mockCapabilityRepo{
				listFunc: func(ctx context.Context, limit, offset int) ([]types.Capability, int, error) {
					capturedLimit = limit
					capturedOffset = offset
					return []types.Capability{}, 0, nil
				},
			},
		}

		svc := NewCapabilityService(mockRepo, settings, nil)

		_, _, err := svc.List(ctx, 25, 50)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if capturedLimit != 25 {
			t.Errorf("expected limit 25, got %d", capturedLimit)
		}

		if capturedOffset != 50 {
			t.Errorf("expected offset 50, got %d", capturedOffset)
		}
	})
}

// Test Query capabilities
func TestCapabilityService_Query(t *testing.T) {
	ctx, settings := setupTest()

	t.Run("query with filters", func(t *testing.T) {
		var capturedFilters map[string]interface{}

		mockRepo := &mockRepository{
			capabilitiesRepo: &mockCapabilityRepo{
				queryFunc: func(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]types.Capability, int, error) {
					capturedFilters = filters
					return []types.Capability{
						{ID: "filtered", Version: "1.0.0", RiskClass: types.RiskHigh},
					}, 1, nil
				},
			},
		}

		svc := NewCapabilityService(mockRepo, settings, nil)

		riskClass := types.RiskHigh
		req := &types.QueryCapabilitiesRequest{
			RiskClass: &riskClass,
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

		if capturedFilters["risk_class"] != types.RiskHigh {
			t.Error("filter not passed correctly")
		}
	})
}
