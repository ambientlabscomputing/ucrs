package service

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"os"

	"github.com/ambientlabscomputing/underleaf/capability_registry_service/repository"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/utils"
)

// Service is the main service interface
type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Capability operations
	Capabilities() *CapabilityService
	// Provider operations
	Providers() *ProviderService
	// Sync operations
	Sync() *SyncService
}

// AppService is the main service implementation
type AppService struct {
	repo         repository.Repository
	settings     *utils.Settings
	capabilities *CapabilityService
	providers    *ProviderService
	sync         *SyncService

	// Ed25519 signing keys
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewAppService creates a new application service
func NewAppService(repo repository.Repository, settings *utils.Settings) (*AppService, error) {
	// Load Ed25519 signing keys
	privateKey, publicKey, err := loadSigningKeys(settings)
	if err != nil {
		return nil, err
	}

	svc := &AppService{
		repo:       repo,
		settings:   settings,
		privateKey: privateKey,
		publicKey:  publicKey,
	}

	// Initialize sub-services
	svc.capabilities = NewCapabilityService(repo, settings)
	svc.providers = NewProviderService(repo, settings)
	svc.sync = NewSyncService(repo, settings, privateKey, publicKey)

	return svc, nil
}

func (s *AppService) Start(ctx context.Context) error {
	logger := utils.GetLogger(ctx)
	logger.Info("Starting Capability Registry Service")

	// Pre-generate initial snapshot if none exists
	if err := s.sync.EnsureSnapshotExists(ctx); err != nil {
		logger.Error("failed to ensure initial snapshot", "error", err)
		return err
	}

	return nil
}

func (s *AppService) Stop(ctx context.Context) error {
	logger := utils.GetLogger(ctx)
	logger.Info("Stopping Capability Registry Service")
	return nil
}

func (s *AppService) Capabilities() *CapabilityService {
	return s.capabilities
}

func (s *AppService) Providers() *ProviderService {
	return s.providers
}

func (s *AppService) Sync() *SyncService {
	return s.sync
}

// loadSigningKeys loads Ed25519 signing keys from the configured paths
func loadSigningKeys(settings *utils.Settings) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	// Read private key
	privateKeyBytes, err := os.ReadFile(settings.Signing.PrivateKeyPath)
	if err != nil {
		return nil, nil, err
	}

	// Decode base64-encoded private key
	privateKey, err := base64.StdEncoding.DecodeString(string(privateKeyBytes))
	if err != nil {
		return nil, nil, err
	}

	// Read public key
	publicKeyBytes, err := os.ReadFile(settings.Signing.PublicKeyPath)
	if err != nil {
		return nil, nil, err
	}

	// Decode base64-encoded public key
	publicKey, err := base64.StdEncoding.DecodeString(string(publicKeyBytes))
	if err != nil {
		return nil, nil, err
	}

	return ed25519.PrivateKey(privateKey), ed25519.PublicKey(publicKey), nil
}
