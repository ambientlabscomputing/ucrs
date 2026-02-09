package repository

import (
	"context"
	"time"

	"github.com/ambientlabscomputing/ucrs/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoProviderRepository struct {
	collection *mongo.Collection
}

func NewMongoProviderRepository(db *mongo.Database) *MongoProviderRepository {
	return &MongoProviderRepository{
		collection: db.Collection("providers"),
	}
}

func (r *MongoProviderRepository) Create(ctx context.Context, provider *types.Provider) error {
	provider.CreatedAt = time.Now()
	provider.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, provider)
	return err
}

func (r *MongoProviderRepository) GetByID(ctx context.Context, providerID, version string) (*types.Provider, error) {
	var provider types.Provider
	filter := bson.M{
		"_id":     providerID,
		"version": version,
	}
	err := r.collection.FindOne(ctx, filter).Decode(&provider)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &provider, err
}

func (r *MongoProviderRepository) Update(ctx context.Context, provider *types.Provider) error {
	provider.UpdatedAt = time.Now()
	filter := bson.M{
		"_id":     provider.ProviderID,
		"version": provider.Version,
	}
	update := bson.M{"$set": provider}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *MongoProviderRepository) Delete(ctx context.Context, providerID, version string) error {
	filter := bson.M{
		"_id":     providerID,
		"version": version,
	}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

func (r *MongoProviderRepository) Query(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]types.Provider, int, error) {
	// Build MongoDB filter from input filters
	filter := bson.M{}
	for key, value := range filters {
		filter[key] = value
	}

	// Count total documents matching the filter
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find with pagination
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "_id", Value: 1}, {Key: "version", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var providers []types.Provider
	if err := cursor.All(ctx, &providers); err != nil {
		return nil, 0, err
	}

	return providers, int(total), nil
}

func (r *MongoProviderRepository) List(ctx context.Context, limit, offset int) ([]types.Provider, int, error) {
	return r.Query(ctx, map[string]interface{}{}, limit, offset)
}

func (r *MongoProviderRepository) GetByCapability(ctx context.Context, capabilityID string) ([]types.Provider, error) {
	filter := bson.M{
		"capabilities.id": capabilityID,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var providers []types.Provider
	if err := cursor.All(ctx, &providers); err != nil {
		return nil, err
	}

	return providers, nil
}

func (r *MongoProviderRepository) GetLatestVersion(ctx context.Context, providerID string) (*types.Provider, error) {
	filter := bson.M{"_id": providerID}
	opts := options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}})

	var provider types.Provider
	err := r.collection.FindOne(ctx, filter, opts).Decode(&provider)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &provider, err
}
