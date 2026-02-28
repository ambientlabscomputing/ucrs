package repository

import (
	"context"
	"time"

	"github.com/ambientlabscomputing/ucrs/sdk/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoCapabilityRepository struct {
	collection *mongo.Collection
}

func NewMongoCapabilityRepository(db *mongo.Database) *MongoCapabilityRepository {
	return &MongoCapabilityRepository{
		collection: db.Collection("capabilities"),
	}
}

func (r *MongoCapabilityRepository) Create(ctx context.Context, capability *types.Capability) error {
	capability.CreatedAt = time.Now()
	capability.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, capability)
	return err
}

func (r *MongoCapabilityRepository) GetByID(ctx context.Context, id string) (*types.Capability, error) {
	var capability types.Capability
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&capability)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &capability, err
}

func (r *MongoCapabilityRepository) Update(ctx context.Context, capability *types.Capability) error {
	capability.UpdatedAt = time.Now()
	update := bson.M{"$set": capability}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": capability.ID}, update)
	return err
}

func (r *MongoCapabilityRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *MongoCapabilityRepository) Query(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]types.Capability, int, error) {
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
		SetSort(bson.D{{Key: "_id", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var capabilities []types.Capability
	if err := cursor.All(ctx, &capabilities); err != nil {
		return nil, 0, err
	}

	return capabilities, int(total), nil
}

func (r *MongoCapabilityRepository) List(ctx context.Context, limit, offset int) ([]types.Capability, int, error) {
	return r.Query(ctx, map[string]interface{}{}, limit, offset)
}

func (r *MongoCapabilityRepository) GetAllVersions(ctx context.Context, id string) ([]types.Capability, error) {
	// For future multi-version support - currently returns single capability
	capability, err := r.GetByID(ctx, id)
	if err != nil || capability == nil {
		return []types.Capability{}, err
	}
	return []types.Capability{*capability}, nil
}
