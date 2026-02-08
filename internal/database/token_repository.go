package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	TokensCollectionName = "token_blacklist"
)

type BlacklistedToken struct {
	Token     string    `bson:"_id"`
	ExpiresAt time.Time `bson:"expiresAt"`
	CreatedAt time.Time `bson:"createdAt"`
}

type TokenRepository struct {
	collection Collection
}

// NewTokenRepository creates a new token repository
func NewTokenRepository() *TokenRepository {
	return &TokenRepository{
		collection: NewMongoCollection(GetCollection(DBName, TokensCollectionName)),
	}
}

// NewTokenRepositoryWithCollection creates a token repository with custom collection (for testing)
func NewTokenRepositoryWithCollection(c Collection) *TokenRepository {
	return &TokenRepository{
		collection: c,
	}
}

// BlacklistToken adds a token to the blacklist
func (tr *TokenRepository) BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	blacklistedToken := BlacklistedToken{
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	_, err := tr.collection.InsertOne(ctx, blacklistedToken)
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// IsTokenBlacklisted checks if a token is in the blacklist
func (tr *TokenRepository) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var result BlacklistedToken
	err := tr.collection.FindOne(ctx, bson.M{"_id": token}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	// Check if token is still valid (not expired)
	if result.ExpiresAt.Before(time.Now()) {
		return false, nil
	}

	return true, nil
}

// CleanupExpiredTokens removes expired tokens from the blacklist
func (tr *TokenRepository) CleanupExpiredTokens(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := tr.collection.DeleteMany(ctx, bson.M{
		"expiresAt": bson.M{
			"$lt": time.Now(),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	if result.DeletedCount > 0 {
		fmt.Printf("Cleaned up %d expired tokens from blacklist\n", result.DeletedCount)
	}

	return nil
}
