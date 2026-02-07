package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateIndexes creates necessary indexes for the database collections
func CreateIndexes() error {
	// Create TTL index on token_blacklist collection to auto-delete expired tokens
	tokenCollection := GetCollection(DBName, TokensCollectionName)

	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0), // Delete when expiresAt time is reached
	}

	_, err := tokenCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return fmt.Errorf("failed to create TTL index on token_blacklist: %w", err)
	}

	// Create index on users collection for email lookups
	userCollection := GetCollection(DBName, UsersCollectionName)

	userIndexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = userCollection.Indexes().CreateOne(context.Background(), userIndexModel)
	if err != nil {
		return fmt.Errorf("failed to create unique index on users email: %w", err)
	}

	// Create index on products collection for name searches
	productCollection := GetCollection(DBName, ProductsCollectionName)

	productIndexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "name", Value: "text"}, {Key: "description", Value: "text"}},
	}

	_, err = productCollection.Indexes().CreateOne(context.Background(), productIndexModel)
	if err != nil {
		return fmt.Errorf("failed to create text index on products: %w", err)
	}

	return nil
}
