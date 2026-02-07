package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
// DBName is the default database name used by repositories. Can be overridden at runtime.
var DBName = "maggiesb"

// InitMongo initializes MongoDB connection
func InitMongo(mongoURI string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify connection
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	MongoClient = client
	return nil
}

// SetDBName overrides the package database name used by GetDB/GetCollection
func SetDBName(name string) {
	if name == "" {
		return
	}
	DBName = name
}

// DisconnectMongo closes the MongoDB connection
func DisconnectMongo() error {
	if MongoClient == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return MongoClient.Disconnect(ctx)
}

// GetDB returns the MongoDB database
func GetDB(dbName string) *mongo.Database {
	if MongoClient == nil {
		panic("MongoDB client not initialized")
	}
	return MongoClient.Database(dbName)
}

// GetCollection returns a MongoDB collection
func GetCollection(dbName, collectionName string) *mongo.Collection {
	return GetDB(dbName).Collection(collectionName)
}
