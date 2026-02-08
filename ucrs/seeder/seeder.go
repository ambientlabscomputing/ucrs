package seeder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ambientlabscomputing/underleaf/capability_registry_service/repository"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/types"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/utils"
	"gopkg.in/yaml.v3"
)

// SeedData represents the structure of seed YAML files
type SeedData struct {
	Capabilities []types.Capability `yaml:"capabilities"`
	Providers    []types.Provider   `yaml:"providers"`
}

// Seeder handles loading and upserting seed data into the registry
type Seeder struct {
	repo        repository.Repository
	environment string
	seedsPath   string
}

// NewSeeder creates a new seeder instance
func NewSeeder(repo repository.Repository, environment string) *Seeder {
	if environment == "" {
		environment = "dev"
	}
	return &Seeder{
		repo:        repo,
		environment: environment,
		seedsPath:   "seeds",
	}
}

// LoadSeeds loads and upserts all seed data for the current environment
func (s *Seeder) LoadSeeds(ctx context.Context) error {
	logger := utils.GetLogger(ctx)
	logger.Info("loading registry seeds", "environment", s.environment)

	envPath := filepath.Join(s.seedsPath, s.environment)

	// Check if environment directory exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		logger.Warn("seed directory not found, skipping seeding", "path", envPath)
		return nil
	}

	// Load capabilities
	if err := s.loadCapabilities(ctx, envPath); err != nil {
		return fmt.Errorf("failed to load capabilities: %w", err)
	}

	// Load providers
	if err := s.loadProviders(ctx, envPath); err != nil {
		return fmt.Errorf("failed to load providers: %w", err)
	}

	logger.Info("successfully loaded all registry seeds")
	return nil
}

// loadCapabilities loads capability seeds from YAML file
func (s *Seeder) loadCapabilities(ctx context.Context, envPath string) error {
	logger := utils.GetLogger(ctx)
	capabilitiesFile := filepath.Join(envPath, "capabilities.yaml")

	// Check if file exists
	if _, err := os.Stat(capabilitiesFile); os.IsNotExist(err) {
		logger.Info("no capabilities seed file found, skipping", "file", capabilitiesFile)
		return nil
	}

	// Read YAML file
	data, err := os.ReadFile(capabilitiesFile)
	if err != nil {
		return fmt.Errorf("failed to read capabilities file: %w", err)
	}

	// Parse YAML
	var seedData SeedData
	if err := yaml.Unmarshal(data, &seedData); err != nil {
		return fmt.Errorf("failed to parse capabilities YAML: %w", err)
	}

	// Upsert each capability
	for _, capability := range seedData.Capabilities {
		if err := s.upsertCapability(ctx, &capability); err != nil {
			logger.Error("failed to upsert capability", "id", capability.ID, "error", err)
			return fmt.Errorf("failed to upsert capability %s: %w", capability.ID, err)
		}
		logger.Info("upserted capability", "id", capability.ID, "version", capability.Version)
	}

	logger.Info("loaded capabilities", "count", len(seedData.Capabilities))
	return nil
}

// loadProviders loads provider seeds from YAML file
func (s *Seeder) loadProviders(ctx context.Context, envPath string) error {
	logger := utils.GetLogger(ctx)
	providersFile := filepath.Join(envPath, "providers.yaml")

	// Check if file exists
	if _, err := os.Stat(providersFile); os.IsNotExist(err) {
		logger.Info("no providers seed file found, skipping", "file", providersFile)
		return nil
	}

	// Read YAML file
	data, err := os.ReadFile(providersFile)
	if err != nil {
		return fmt.Errorf("failed to read providers file: %w", err)
	}

	// Parse YAML
	var seedData SeedData
	if err := yaml.Unmarshal(data, &seedData); err != nil {
		return fmt.Errorf("failed to parse providers YAML: %w", err)
	}

	// Upsert each provider
	for _, provider := range seedData.Providers {
		if err := s.upsertProvider(ctx, &provider); err != nil {
			logger.Error("failed to upsert provider", "id", provider.ProviderID, "version", provider.Version, "error", err)
			return fmt.Errorf("failed to upsert provider %s@%s: %w", provider.ProviderID, provider.Version, err)
		}
		logger.Info("upserted provider", "id", provider.ProviderID, "version", provider.Version)
	}

	logger.Info("loaded providers", "count", len(seedData.Providers))
	return nil
}

// upsertCapability inserts or updates a capability
func (s *Seeder) upsertCapability(ctx context.Context, capability *types.Capability) error {
	// Check if capability already exists
	existing, err := s.repo.Capabilities().GetByID(ctx, capability.ID)
	if err != nil {
		return err
	}

	now := time.Now()

	if existing == nil {
		// Create new capability
		capability.CreatedAt = now
		capability.UpdatedAt = now
		return s.repo.Capabilities().Create(ctx, capability)
	}

	// Update existing capability (preserve CreatedAt)
	capability.CreatedAt = existing.CreatedAt
	capability.UpdatedAt = now
	return s.repo.Capabilities().Update(ctx, capability)
}

// upsertProvider inserts or updates a provider
func (s *Seeder) upsertProvider(ctx context.Context, provider *types.Provider) error {
	// Check if provider already exists
	existing, err := s.repo.Providers().GetByID(ctx, provider.ProviderID, provider.Version)
	if err != nil {
		return err
	}

	now := time.Now()

	if existing == nil {
		// Create new provider
		provider.CreatedAt = now
		provider.UpdatedAt = now
		return s.repo.Providers().Create(ctx, provider)
	}

	// Update existing provider (preserve CreatedAt)
	provider.CreatedAt = existing.CreatedAt
	provider.UpdatedAt = now
	return s.repo.Providers().Update(ctx, provider)
}
