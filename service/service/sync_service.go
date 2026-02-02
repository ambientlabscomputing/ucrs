package service

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/ambientlabscomputing/underleaf/capability_registry_service/repository"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/types"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/utils"
)

type SyncService struct {
	repo       repository.Repository
	settings   *utils.Settings
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func NewSyncService(repo repository.Repository, settings *utils.Settings, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) *SyncService {
	return &SyncService{
		repo:       repo,
		settings:   settings,
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (s *SyncService) GetSnapshot(ctx context.Context, ifNoneMatch string) (*types.RegistrySnapshot, bool, error) {
	logger := utils.GetLogger(ctx)

	snapshot, err := s.repo.Snapshots().GetLatest(ctx)
	if err != nil {
		logger.Error("failed to get latest snapshot", "error", err)
		return nil, false, err
	}

	if snapshot == nil {
		logger.Debug("no snapshot found")
		return nil, false, nil
	}

	// Check ETag for caching
	if ifNoneMatch != "" && ifNoneMatch == snapshot.ETag {
		logger.Debug("snapshot not modified", "etag", snapshot.ETag)
		return nil, true, nil
	}

	return snapshot, false, nil
}

func (s *SyncService) GetDelta(ctx context.Context, sinceVersion string) (*types.DeltaUpdate, error) {
	logger := utils.GetLogger(ctx)

	delta, err := s.repo.Snapshots().GetDeltaSince(ctx, sinceVersion)
	if err != nil {
		logger.Error("failed to get delta", "error", err, "since_version", sinceVersion)
		return nil, err
	}

	return delta, nil
}

func (s *SyncService) GenerateSnapshot(ctx context.Context) (*types.RegistrySnapshot, error) {
	logger := utils.GetLogger(ctx)

	// Get all capabilities
	capabilities, _, err := s.repo.Capabilities().List(ctx, 10000, 0)
	if err != nil {
		logger.Error("failed to list capabilities for snapshot", "error", err)
		return nil, err
	}

	// Get all providers
	providers, _, err := s.repo.Providers().List(ctx, 10000, 0)
	if err != nil {
		logger.Error("failed to list providers for snapshot", "error", err)
		return nil, err
	}

	// Generate version (timestamp-based)
	version := fmt.Sprintf("%d", time.Now().Unix())

	snapshot := &types.RegistrySnapshot{
		Version:      version,
		Timestamp:    time.Now(),
		Capabilities: capabilities,
		Providers:    providers,
	}

	// Compute ETag
	snapshot.ETag = s.computeETag(snapshot)

	// Sign the snapshot
	if err := s.signSnapshot(snapshot); err != nil {
		logger.Error("failed to sign snapshot", "error", err)
		return nil, err
	}

	// Store the snapshot
	if err := s.repo.Snapshots().Create(ctx, snapshot); err != nil {
		logger.Error("failed to store snapshot", "error", err)
		return nil, err
	}

	logger.Info("snapshot generated", "version", version, "capabilities", len(capabilities), "providers", len(providers))
	return snapshot, nil
}

func (s *SyncService) EnsureSnapshotExists(ctx context.Context) error {
	logger := utils.GetLogger(ctx)

	snapshot, err := s.repo.Snapshots().GetLatest(ctx)
	if err != nil {
		return err
	}

	if snapshot == nil {
		logger.Info("no snapshot found, generating initial snapshot")
		_, err := s.GenerateSnapshot(ctx)
		return err
	}

	logger.Debug("snapshot already exists", "version", snapshot.Version)
	return nil
}

func (s *SyncService) computeETag(snapshot *types.RegistrySnapshot) string {
	// Compute SHA256 hash of the snapshot version and timestamp
	hash := sha256.New()
	hash.Write([]byte(snapshot.Version))
	hash.Write([]byte(snapshot.Timestamp.String()))
	return fmt.Sprintf("\"%x\"", hash.Sum(nil))
}

func (s *SyncService) signSnapshot(snapshot *types.RegistrySnapshot) error {
	// Create a signature payload from the snapshot
	payload := fmt.Sprintf("%s:%s:%d:%d",
		snapshot.Version,
		snapshot.Timestamp.Format(time.RFC3339),
		len(snapshot.Capabilities),
		len(snapshot.Providers))

	// Sign the payload with Ed25519
	signature := ed25519.Sign(s.privateKey, []byte(payload))

	// Store the signature in the manifest
	snapshot.Manifest = types.SignedManifest{
		KeyID:     s.settings.Signing.KeyID,
		Signature: base64.StdEncoding.EncodeToString(signature),
		Algorithm: "Ed25519",
		SignedAt:  time.Now(),
	}

	return nil
}

func (s *SyncService) VerifySignature(snapshot *types.RegistrySnapshot) bool {
	// Reconstruct the payload
	payload := fmt.Sprintf("%s:%s:%d:%d",
		snapshot.Version,
		snapshot.Timestamp.Format(time.RFC3339),
		len(snapshot.Capabilities),
		len(snapshot.Providers))

	// Decode the signature
	signature, err := base64.StdEncoding.DecodeString(snapshot.Manifest.Signature)
	if err != nil {
		return false
	}

	// Verify with Ed25519
	return ed25519.Verify(s.publicKey, []byte(payload), signature)
}
