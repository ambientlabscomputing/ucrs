package repository

import (
	"context"

	"github.com/ambientlabscomputing/ucrs/sdk/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// Repository is the main repository interface aggregating all sub-repositories
type Repository interface {
	Capabilities() CapabilityRepository
	Providers() ProviderRepository
	Snapshots() SnapshotRepository
}

// CapabilityRepository handles capability schema storage and retrieval
type CapabilityRepository interface {
	Create(ctx context.Context, capability *types.Capability) error
	GetByID(ctx context.Context, id string) (*types.Capability, error)
	Update(ctx context.Context, capability *types.Capability) error
	Delete(ctx context.Context, id string) error
	Query(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]types.Capability, int, error)
	List(ctx context.Context, limit, offset int) ([]types.Capability, int, error)
	GetAllVersions(ctx context.Context, id string) ([]types.Capability, error)
}

// ProviderRepository handles provider metadata storage and retrieval
type ProviderRepository interface {
	Create(ctx context.Context, provider *types.Provider) error
	GetByID(ctx context.Context, providerID, version string) (*types.Provider, error)
	Update(ctx context.Context, provider *types.Provider) error
	Delete(ctx context.Context, providerID, version string) error
	Query(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]types.Provider, int, error)
	List(ctx context.Context, limit, offset int) ([]types.Provider, int, error)
	GetByCapability(ctx context.Context, capabilityID string) ([]types.Provider, error)
	GetLatestVersion(ctx context.Context, providerID string) (*types.Provider, error)
}

// SnapshotRepository handles registry snapshot storage and retrieval
type SnapshotRepository interface {
	Create(ctx context.Context, snapshot *types.RegistrySnapshot) error
	GetLatest(ctx context.Context) (*types.RegistrySnapshot, error)
	GetByVersion(ctx context.Context, version string) (*types.RegistrySnapshot, error)
	GetByETag(ctx context.Context, etag string) (*types.RegistrySnapshot, error)
	ListVersions(ctx context.Context, limit int) ([]string, error)
	GetDeltaSince(ctx context.Context, sinceVersion string) (*types.DeltaUpdate, error)
}

// MongoRepository is the MongoDB implementation of the Repository interface
type MongoRepository struct {
	db           *mongo.Database
	capabilities CapabilityRepository
	providers    ProviderRepository
	snapshots    SnapshotRepository
}

// NewMongoRepository creates a new MongoDB repository
func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{
		db:           db,
		capabilities: NewMongoCapabilityRepository(db),
		providers:    NewMongoProviderRepository(db),
		snapshots:    NewMongoSnapshotRepository(db),
	}
}

func (r *MongoRepository) Capabilities() CapabilityRepository {
	return r.capabilities
}

func (r *MongoRepository) Providers() ProviderRepository {
	return r.providers
}

func (r *MongoRepository) Snapshots() SnapshotRepository {
	return r.snapshots
}
