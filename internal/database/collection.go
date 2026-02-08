package database

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection defines the interface for MongoDB collection operations
type Collection interface {
	InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error)
	FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error)
	DeleteMany(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error)
	CountDocuments(ctx context.Context, filter interface{}) (int64, error)
	EstimatedDocumentCount(ctx context.Context) (int64, error)
	Aggregate(ctx context.Context, pipeline interface{}) (*mongo.Cursor, error)
}

// MongoCollection wraps mongo.Collection to implement Collection interface
type MongoCollection struct {
	collection *mongo.Collection
}

// NewMongoCollection creates a wrapper around mongo.Collection
func NewMongoCollection(c *mongo.Collection) Collection {
	return &MongoCollection{collection: c}
}

func (mc *MongoCollection) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	return mc.collection.InsertOne(ctx, document)
}

func (mc *MongoCollection) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	return mc.collection.FindOne(ctx, filter)
}

func (mc *MongoCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	return mc.collection.Find(ctx, filter, opts...)
}

func (mc *MongoCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return mc.collection.UpdateOne(ctx, filter, update, opts...)
}

func (mc *MongoCollection) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return mc.collection.UpdateMany(ctx, filter, update, opts...)
}

func (mc *MongoCollection) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	return mc.collection.DeleteOne(ctx, filter)
}

func (mc *MongoCollection) DeleteMany(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	return mc.collection.DeleteMany(ctx, filter)
}

func (mc *MongoCollection) CountDocuments(ctx context.Context, filter interface{}) (int64, error) {
	return mc.collection.CountDocuments(ctx, filter)
}

func (mc *MongoCollection) EstimatedDocumentCount(ctx context.Context) (int64, error) {
	return mc.collection.EstimatedDocumentCount(ctx)
}
func (mc *MongoCollection) Aggregate(ctx context.Context, pipeline interface{}) (*mongo.Cursor, error) {
	return mc.collection.Aggregate(ctx, pipeline)
}