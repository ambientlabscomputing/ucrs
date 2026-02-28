package repository

import (
	"context"
	"time"

	"github.com/ambientlabscomputing/ucrs/sdk/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoSnapshotRepository struct {
	collection *mongo.Collection
}

func NewMongoSnapshotRepository(db *mongo.Database) *MongoSnapshotRepository {
	return &MongoSnapshotRepository{
		collection: db.Collection("snapshots"),
	}
}

func (r *MongoSnapshotRepository) Create(ctx context.Context, snapshot *types.RegistrySnapshot) error {
	snapshot.Timestamp = time.Now()
	_, err := r.collection.InsertOne(ctx, snapshot)
	return err
}

func (r *MongoSnapshotRepository) GetLatest(ctx context.Context) (*types.RegistrySnapshot, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	var snapshot types.RegistrySnapshot

	err := r.collection.FindOne(ctx, bson.M{}, opts).Decode(&snapshot)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &snapshot, err
}

func (r *MongoSnapshotRepository) GetByVersion(ctx context.Context, version string) (*types.RegistrySnapshot, error) {
	var snapshot types.RegistrySnapshot
	err := r.collection.FindOne(ctx, bson.M{"version": version}).Decode(&snapshot)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &snapshot, err
}

func (r *MongoSnapshotRepository) GetByETag(ctx context.Context, etag string) (*types.RegistrySnapshot, error) {
	var snapshot types.RegistrySnapshot
	err := r.collection.FindOne(ctx, bson.M{"etag": etag}).Decode(&snapshot)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &snapshot, err
}

func (r *MongoSnapshotRepository) ListVersions(ctx context.Context, limit int) ([]string, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit)).
		SetProjection(bson.M{"version": 1, "_id": 0})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Version string `bson:"version"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	versions := make([]string, len(results))
	for i, r := range results {
		versions[i] = r.Version
	}

	return versions, nil
}

// GetDeltaSince computes the delta between a source version and the latest snapshot
// This implementation computes deltas dynamically by comparing snapshots
// For production, consider storing pre-computed deltas or using change streams
func (r *MongoSnapshotRepository) GetDeltaSince(ctx context.Context, sinceVersion string) (*types.DeltaUpdate, error) {
	// Get the source snapshot
	sourceSnapshot, err := r.GetByVersion(ctx, sinceVersion)
	if err != nil || sourceSnapshot == nil {
		return nil, err
	}

	// Get the latest snapshot
	latestSnapshot, err := r.GetLatest(ctx)
	if err != nil || latestSnapshot == nil {
		return nil, err
	}

	// If versions match, return empty delta
	if sourceSnapshot.Version == latestSnapshot.Version {
		return &types.DeltaUpdate{
			SinceVersion: sinceVersion,
			ToVersion:    latestSnapshot.Version,
			Timestamp:    time.Now(),
			Manifest:     latestSnapshot.Manifest,
		}, nil
	}

	// Compute delta (simplified implementation)
	delta := &types.DeltaUpdate{
		SinceVersion: sinceVersion,
		ToVersion:    latestSnapshot.Version,
		Timestamp:    time.Now(),
		Manifest:     latestSnapshot.Manifest,
	}

	// Build maps for comparison
	oldCapabilities := make(map[string]types.Capability)
	for _, cap := range sourceSnapshot.Capabilities {
		oldCapabilities[cap.ID] = cap
	}

	newCapabilities := make(map[string]types.Capability)
	for _, cap := range latestSnapshot.Capabilities {
		newCapabilities[cap.ID] = cap
	}

	// Identify added and updated capabilities
	for id, newCap := range newCapabilities {
		if oldCap, exists := oldCapabilities[id]; exists {
			if oldCap.Version != newCap.Version || oldCap.UpdatedAt != newCap.UpdatedAt {
				delta.UpdatedCapabilities = append(delta.UpdatedCapabilities, newCap)
			}
		} else {
			delta.AddedCapabilities = append(delta.AddedCapabilities, newCap)
		}
	}

	// Identify removed capabilities
	for id := range oldCapabilities {
		if _, exists := newCapabilities[id]; !exists {
			delta.RemovedCapabilities = append(delta.RemovedCapabilities, id)
		}
	}

	// Similar logic for providers
	oldProviders := make(map[string]types.Provider)
	for _, prov := range sourceSnapshot.Providers {
		key := prov.ProviderID + ":" + prov.Version
		oldProviders[key] = prov
	}

	newProviders := make(map[string]types.Provider)
	for _, prov := range latestSnapshot.Providers {
		key := prov.ProviderID + ":" + prov.Version
		newProviders[key] = prov
	}

	for key, newProv := range newProviders {
		if oldProv, exists := oldProviders[key]; exists {
			if oldProv.UpdatedAt != newProv.UpdatedAt {
				delta.UpdatedProviders = append(delta.UpdatedProviders, newProv)
			}
		} else {
			delta.AddedProviders = append(delta.AddedProviders, newProv)
		}
	}

	for key := range oldProviders {
		if _, exists := newProviders[key]; !exists {
			delta.RemovedProviders = append(delta.RemovedProviders, key)
		}
	}

	return delta, nil
}
