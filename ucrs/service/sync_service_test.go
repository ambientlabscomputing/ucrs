package service

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"github.com/ambientlabscomputing/ucrs/repository"
	"github.com/ambientlabscomputing/ucrs/types"
	"github.com/ambientlabscomputing/ucrs/utils"
)

type mockSnapshotRepoForSync struct {
	createFunc        func(ctx context.Context, snapshot *types.RegistrySnapshot) error
	getLatestFunc     func(ctx context.Context) (*types.RegistrySnapshot, error)
	getByVersionFunc  func(ctx context.Context, version string) (*types.RegistrySnapshot, error)
	getByETagFunc     func(ctx context.Context, etag string) (*types.RegistrySnapshot, error)
	listVersionsFunc  func(ctx context.Context, limit int) ([]string, error)
	getDeltaSinceFunc func(ctx context.Context, sinceVersion string) (*types.DeltaUpdate, error)
}

func (m *mockSnapshotRepoForSync) Create(ctx context.Context, snapshot *types.RegistrySnapshot) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, snapshot)
	}
	return nil
}

func (m *mockSnapshotRepoForSync) GetLatest(ctx context.Context) (*types.RegistrySnapshot, error) {
	if m.getLatestFunc != nil {
		return m.getLatestFunc(ctx)
	}
	return nil, nil
}

func (m *mockSnapshotRepoForSync) GetByVersion(ctx context.Context, version string) (*types.RegistrySnapshot, error) {
	if m.getByVersionFunc != nil {
		return m.getByVersionFunc(ctx, version)
	}
	return nil, nil
}

func (m *mockSnapshotRepoForSync) GetByETag(ctx context.Context, etag string) (*types.RegistrySnapshot, error) {
	if m.getByETagFunc != nil {
		return m.getByETagFunc(ctx, etag)
	}
	return nil, nil
}

func (m *mockSnapshotRepoForSync) ListVersions(ctx context.Context, limit int) ([]string, error) {
	if m.listVersionsFunc != nil {
		return m.listVersionsFunc(ctx, limit)
	}
	return []string{}, nil
}

func (m *mockSnapshotRepoForSync) GetDeltaSince(ctx context.Context, sinceVersion string) (*types.DeltaUpdate, error) {
	if m.getDeltaSinceFunc != nil {
		return m.getDeltaSinceFunc(ctx, sinceVersion)
	}
	return nil, nil
}

type mockCapabilityRepoForSync struct {
	listFunc func(ctx context.Context, limit, offset int) ([]types.Capability, int, error)
}

func (m *mockCapabilityRepoForSync) Create(ctx context.Context, capability *types.Capability) error {
	return nil
}
func (m *mockCapabilityRepoForSync) GetByID(ctx context.Context, id string) (*types.Capability, error) {
	return nil, nil
}
func (m *mockCapabilityRepoForSync) Update(ctx context.Context, capability *types.Capability) error {
	return nil
}
func (m *mockCapabilityRepoForSync) Delete(ctx context.Context, id string) error { return nil }
func (m *mockCapabilityRepoForSync) Query(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]types.Capability, int, error) {
	return []types.Capability{}, 0, nil
}
func (m *mockCapabilityRepoForSync) List(ctx context.Context, limit, offset int) ([]types.Capability, int, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, offset)
	}
	return []types.Capability{}, 0, nil
}
func (m *mockCapabilityRepoForSync) GetAllVersions(ctx context.Context, id string) ([]types.Capability, error) {
	return []types.Capability{}, nil
}

type mockProviderRepoForSync struct {
	listFunc func(ctx context.Context, limit, offset int) ([]types.Provider, int, error)
}

func (m *mockProviderRepoForSync) Create(ctx context.Context, provider *types.Provider) error {
	return nil
}
func (m *mockProviderRepoForSync) GetByID(ctx context.Context, providerID, version string) (*types.Provider, error) {
	return nil, nil
}
func (m *mockProviderRepoForSync) Update(ctx context.Context, provider *types.Provider) error {
	return nil
}
func (m *mockProviderRepoForSync) Delete(ctx context.Context, providerID, version string) error {
	return nil
}
func (m *mockProviderRepoForSync) Query(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]types.Provider, int, error) {
	return []types.Provider{}, 0, nil
}
func (m *mockProviderRepoForSync) List(ctx context.Context, limit, offset int) ([]types.Provider, int, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, offset)
	}
	return []types.Provider{}, 0, nil
}
func (m *mockProviderRepoForSync) GetByCapability(ctx context.Context, capabilityID string) ([]types.Provider, error) {
	return []types.Provider{}, nil
}
func (m *mockProviderRepoForSync) GetLatestVersion(ctx context.Context, providerID string) (*types.Provider, error) {
	return nil, nil
}

type mockRepoForSync struct {
	capabilitiesRepo *mockCapabilityRepoForSync
	providersRepo    *mockProviderRepoForSync
	snapshotsRepo    *mockSnapshotRepoForSync
}

func (m *mockRepoForSync) Capabilities() repository.CapabilityRepository {
	return m.capabilitiesRepo
}

func (m *mockRepoForSync) Providers() repository.ProviderRepository {
	return m.providersRepo
}

func (m *mockRepoForSync) Snapshots() repository.SnapshotRepository {
	return m.snapshotsRepo
}

// Test GetSnapshot
func TestSyncService_GetSnapshot(t *testing.T) {
	ctx, settings := setupTest()
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)

	t.Run("returns latest snapshot", func(t *testing.T) {
		expectedSnapshot := &types.RegistrySnapshot{
			Version:   "123456",
			Timestamp: time.Now(),
			ETag:      "\"abc123\"",
		}

		mockRepo := &mockRepoForSync{
			snapshotsRepo: &mockSnapshotRepoForSync{
				getLatestFunc: func(ctx context.Context) (*types.RegistrySnapshot, error) {
					return expectedSnapshot, nil
				},
			},
		}

		svc := NewSyncService(mockRepo, settings, privateKey, publicKey)

		snapshot, notModified, err := svc.GetSnapshot(ctx, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if notModified {
			t.Error("expected not modified to be false")
		}

		if snapshot == nil {
			t.Fatal("expected snapshot, got nil")
		}

		if snapshot.Version != expectedSnapshot.Version {
			t.Errorf("expected version %s, got %s", expectedSnapshot.Version, snapshot.Version)
		}
	})

	t.Run("returns not modified when etag matches", func(t *testing.T) {
		etag := "\"abc123\""
		expectedSnapshot := &types.RegistrySnapshot{
			Version:   "123456",
			Timestamp: time.Now(),
			ETag:      etag,
		}

		mockRepo := &mockRepoForSync{
			snapshotsRepo: &mockSnapshotRepoForSync{
				getLatestFunc: func(ctx context.Context) (*types.RegistrySnapshot, error) {
					return expectedSnapshot, nil
				},
			},
		}

		svc := NewSyncService(mockRepo, settings, privateKey, publicKey)

		snapshot, notModified, err := svc.GetSnapshot(ctx, etag)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !notModified {
			t.Error("expected not modified to be true")
		}

		if snapshot != nil {
			t.Error("expected nil snapshot when not modified")
		}
	})

	t.Run("returns nil when no snapshot exists", func(t *testing.T) {
		mockRepo := &mockRepoForSync{
			snapshotsRepo: &mockSnapshotRepoForSync{
				getLatestFunc: func(ctx context.Context) (*types.RegistrySnapshot, error) {
					return nil, nil
				},
			},
		}

		svc := NewSyncService(mockRepo, settings, privateKey, publicKey)

		snapshot, notModified, err := svc.GetSnapshot(ctx, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if notModified {
			t.Error("expected not modified to be false")
		}

		if snapshot != nil {
			t.Error("expected nil snapshot")
		}
	})
}

// Test GetDelta
func TestSyncService_GetDelta(t *testing.T) {
	ctx, settings := setupTest()
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)

	t.Run("returns delta update", func(t *testing.T) {
		expectedDelta := &types.DeltaUpdate{
			SinceVersion:        "123",
			ToVersion:           "456",
			AddedCapabilities:   []types.Capability{{ID: "new.cap", Version: "1.0.0"}},
			UpdatedCapabilities: []types.Capability{},
			RemovedCapabilities: []string{},
		}

		mockRepo := &mockRepoForSync{
			snapshotsRepo: &mockSnapshotRepoForSync{
				getDeltaSinceFunc: func(ctx context.Context, sinceVersion string) (*types.DeltaUpdate, error) {
					if sinceVersion == "123" {
						return expectedDelta, nil
					}
					return nil, nil
				},
			},
		}

		svc := NewSyncService(mockRepo, settings, privateKey, publicKey)

		delta, err := svc.GetDelta(ctx, "123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if delta == nil {
			t.Fatal("expected delta, got nil")
		}

		if delta.SinceVersion != "123" {
			t.Errorf("expected from version 123, got %s", delta.SinceVersion)
		}

		if len(delta.AddedCapabilities) != 1 {
			t.Errorf("expected 1 added capability, got %d", len(delta.AddedCapabilities))
		}
	})
}

// Test GenerateSnapshot
func TestSyncService_GenerateSnapshot(t *testing.T) {
	ctx, settings := setupTest()
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	settings.Signing = utils.SigningConfig{KeyID: "test-key-001"}

	t.Run("generates snapshot successfully", func(t *testing.T) {
		caps := []types.Capability{
			{ID: "cap1", Version: "1.0.0", Description: "First"},
			{ID: "cap2", Version: "1.0.0", Description: "Second"},
		}

		provs := []types.Provider{
			{ProviderID: "prov1", Version: "1.0.0", TrustTier: types.TrustCertified},
		}

		var createdSnapshot *types.RegistrySnapshot

		mockRepo := &mockRepoForSync{
			capabilitiesRepo: &mockCapabilityRepoForSync{
				listFunc: func(ctx context.Context, limit, offset int) ([]types.Capability, int, error) {
					return caps, len(caps), nil
				},
			},
			providersRepo: &mockProviderRepoForSync{
				listFunc: func(ctx context.Context, limit, offset int) ([]types.Provider, int, error) {
					return provs, len(provs), nil
				},
			},
			snapshotsRepo: &mockSnapshotRepoForSync{
				createFunc: func(ctx context.Context, snapshot *types.RegistrySnapshot) error {
					createdSnapshot = snapshot
					return nil
				},
			},
		}

		svc := NewSyncService(mockRepo, settings, privateKey, publicKey)

		snapshot, err := svc.GenerateSnapshot(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if snapshot == nil {
			t.Fatal("expected snapshot, got nil")
		}

		if snapshot.Version == "" {
			t.Error("expected version to be set")
		}

		if len(snapshot.Capabilities) != 2 {
			t.Errorf("expected 2 capabilities, got %d", len(snapshot.Capabilities))
		}

		if len(snapshot.Providers) != 1 {
			t.Errorf("expected 1 provider, got %d", len(snapshot.Providers))
		}

		if snapshot.ETag == "" {
			t.Error("expected ETag to be computed")
		}

		if snapshot.Manifest.Signature == "" {
			t.Error("expected snapshot to be signed")
		}

		if snapshot.Manifest.KeyID != "test-key-001" {
			t.Errorf("expected key ID test-key-001, got %s", snapshot.Manifest.KeyID)
		}

		if createdSnapshot == nil {
			t.Fatal("snapshot was not stored")
		}
	})

	t.Run("handles capability list error", func(t *testing.T) {
		expectedErr := errors.New("database error")

		mockRepo := &mockRepoForSync{
			capabilitiesRepo: &mockCapabilityRepoForSync{
				listFunc: func(ctx context.Context, limit, offset int) ([]types.Capability, int, error) {
					return nil, 0, expectedErr
				},
			},
		}

		svc := NewSyncService(mockRepo, settings, privateKey, publicKey)

		_, err := svc.GenerateSnapshot(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("handles provider list error", func(t *testing.T) {
		expectedErr := errors.New("provider database error")

		mockRepo := &mockRepoForSync{
			capabilitiesRepo: &mockCapabilityRepoForSync{
				listFunc: func(ctx context.Context, limit, offset int) ([]types.Capability, int, error) {
					return []types.Capability{}, 0, nil
				},
			},
			providersRepo: &mockProviderRepoForSync{
				listFunc: func(ctx context.Context, limit, offset int) ([]types.Provider, int, error) {
					return nil, 0, expectedErr
				},
			},
		}

		svc := NewSyncService(mockRepo, settings, privateKey, publicKey)

		_, err := svc.GenerateSnapshot(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

// Test EnsureSnapshotExists
func TestSyncService_EnsureSnapshotExists(t *testing.T) {
	ctx, settings := setupTest()
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	settings.Signing = utils.SigningConfig{KeyID: "test-key"}

	t.Run("does nothing when snapshot exists", func(t *testing.T) {
		existingSnapshot := &types.RegistrySnapshot{
			Version:   "123",
			Timestamp: time.Now(),
		}

		createCalled := false

		mockRepo := &mockRepoForSync{
			snapshotsRepo: &mockSnapshotRepoForSync{
				getLatestFunc: func(ctx context.Context) (*types.RegistrySnapshot, error) {
					return existingSnapshot, nil
				},
				createFunc: func(ctx context.Context, snapshot *types.RegistrySnapshot) error {
					createCalled = true
					return nil
				},
			},
		}

		svc := NewSyncService(mockRepo, settings, privateKey, publicKey)

		err := svc.EnsureSnapshotExists(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if createCalled {
			t.Error("snapshot should not be created when one exists")
		}
	})

	t.Run("generates snapshot when none exists", func(t *testing.T) {
		createCalled := false

		mockRepo := &mockRepoForSync{
			snapshotsRepo: &mockSnapshotRepoForSync{
				getLatestFunc: func(ctx context.Context) (*types.RegistrySnapshot, error) {
					return nil, nil
				},
				createFunc: func(ctx context.Context, snapshot *types.RegistrySnapshot) error {
					createCalled = true
					return nil
				},
			},
			capabilitiesRepo: &mockCapabilityRepoForSync{
				listFunc: func(ctx context.Context, limit, offset int) ([]types.Capability, int, error) {
					return []types.Capability{}, 0, nil
				},
			},
			providersRepo: &mockProviderRepoForSync{
				listFunc: func(ctx context.Context, limit, offset int) ([]types.Provider, int, error) {
					return []types.Provider{}, 0, nil
				},
			},
		}

		svc := NewSyncService(mockRepo, settings, privateKey, publicKey)

		err := svc.EnsureSnapshotExists(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !createCalled {
			t.Error("snapshot should be created when none exists")
		}
	})
}

// Test computeETag
func TestSyncService_ComputeETag(t *testing.T) {
	_, settings := setupTest()
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)

	svc := NewSyncService(&mockRepoForSync{}, settings, privateKey, publicKey)

	t.Run("produces consistent etag", func(t *testing.T) {
		timestamp := time.Now()
		snapshot := &types.RegistrySnapshot{
			Version:   "123",
			Timestamp: timestamp,
		}

		etag1 := svc.computeETag(snapshot)
		etag2 := svc.computeETag(snapshot)

		if etag1 != etag2 {
			t.Error("etag should be consistent for same snapshot")
		}

		if etag1 == "" {
			t.Error("etag should not be empty")
		}
	})

	t.Run("produces different etag for different versions", func(t *testing.T) {
		timestamp := time.Now()
		snapshot1 := &types.RegistrySnapshot{
			Version:   "123",
			Timestamp: timestamp,
		}
		snapshot2 := &types.RegistrySnapshot{
			Version:   "456",
			Timestamp: timestamp,
		}

		etag1 := svc.computeETag(snapshot1)
		etag2 := svc.computeETag(snapshot2)

		if etag1 == etag2 {
			t.Error("etag should be different for different versions")
		}
	})
}

// Test signature verification
func TestSyncService_SignAndVerify(t *testing.T) {
	_, settings := setupTest()
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	settings.Signing = utils.SigningConfig{KeyID: "test-key-001"}

	svc := NewSyncService(&mockRepoForSync{}, settings, privateKey, publicKey)

	t.Run("signs and verifies valid snapshot", func(t *testing.T) {
		snapshot := &types.RegistrySnapshot{
			Version:      "123",
			Timestamp:    time.Now(),
			Capabilities: []types.Capability{{ID: "cap1", Version: "1.0.0"}},
			Providers:    []types.Provider{{ProviderID: "prov1", Version: "1.0.0"}},
		}

		err := svc.signSnapshot(snapshot)
		if err != nil {
			t.Fatalf("unexpected error signing: %v", err)
		}

		if snapshot.Manifest.Signature == "" {
			t.Error("signature should be set")
		}

		if snapshot.Manifest.KeyID != "test-key-001" {
			t.Errorf("expected key ID test-key-001, got %s", snapshot.Manifest.KeyID)
		}

		if snapshot.Manifest.Algorithm != "Ed25519" {
			t.Errorf("expected algorithm Ed25519, got %s", snapshot.Manifest.Algorithm)
		}

		valid := svc.VerifySignature(snapshot)
		if !valid {
			t.Error("signature should be valid")
		}
	})

	t.Run("detects tampered snapshot", func(t *testing.T) {
		snapshot := &types.RegistrySnapshot{
			Version:      "123",
			Timestamp:    time.Now(),
			Capabilities: []types.Capability{{ID: "cap1", Version: "1.0.0"}},
			Providers:    []types.Provider{},
		}

		err := svc.signSnapshot(snapshot)
		if err != nil {
			t.Fatalf("unexpected error signing: %v", err)
		}

		// Tamper with the snapshot
		snapshot.Version = "999"

		valid := svc.VerifySignature(snapshot)
		if valid {
			t.Error("signature should be invalid after tampering")
		}
	})

	t.Run("detects tampered capability count", func(t *testing.T) {
		snapshot := &types.RegistrySnapshot{
			Version:      "123",
			Timestamp:    time.Now(),
			Capabilities: []types.Capability{{ID: "cap1", Version: "1.0.0"}},
			Providers:    []types.Provider{},
		}

		err := svc.signSnapshot(snapshot)
		if err != nil {
			t.Fatalf("unexpected error signing: %v", err)
		}

		// Add a capability after signing
		snapshot.Capabilities = append(snapshot.Capabilities, types.Capability{ID: "cap2", Version: "1.0.0"})

		valid := svc.VerifySignature(snapshot)
		if valid {
			t.Error("signature should be invalid after adding capability")
		}
	})
}
