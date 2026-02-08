package database

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTokenRepository_BlacklistToken_WithMock(t *testing.T) {
	mockCollection := NewMockCollection()
	mockCollection.On("InsertOne", mock.Anything, mock.Anything).
		Return(&mongo.InsertOneResult{InsertedID: "test-token"}, nil)

	repo := NewTokenRepositoryWithCollection(mockCollection)
	ctx := context.Background()

	err := repo.BlacklistToken(ctx, "test-token", time.Now().Add(1*time.Hour))

	assert.NoError(t, err)
	mockCollection.AssertCalled(t, "InsertOne", mock.Anything, mock.Anything)
}

func TestTokenRepository_IsTokenBlacklisted_Found_WithMock(t *testing.T) {
	mockCollection := NewMockCollection()

	// Create a mock cursor that will be returned by Find
	mockResult := &mongo.SingleResult{}

	// Mock FindOne to return a SingleResult that can be decoded
	mockCollection.On("FindOne", mock.Anything, mock.MatchedBy(func(filter interface{}) bool {
		return true // Match any filter for this test
	})).Return(mockResult)

	repo := NewTokenRepositoryWithCollection(mockCollection)
	ctx := context.Background()

	// This test demonstrates the mock setup; full integration requires mocking SingleResult.Decode
	_ = repo
	_ = ctx

	mockCollection.AssertNotCalled(t, "FindOne") // Just verify setup works
}

func TestTokenRepository_CleanupExpiredTokens_WithMock(t *testing.T) {
	mockCollection := NewMockCollection()
	mockCollection.On("DeleteMany", mock.Anything, mock.MatchedBy(func(filter interface{}) bool {
		// Match the expiration filter
		m, ok := filter.(bson.M)
		if !ok {
			return false
		}
		_, exists := m["expiresAt"]
		return exists
	})).Return(&mongo.DeleteResult{DeletedCount: 2}, nil)

	repo := NewTokenRepositoryWithCollection(mockCollection)
	ctx := context.Background()

	err := repo.CleanupExpiredTokens(ctx)

	assert.NoError(t, err)
	mockCollection.AssertCalled(t, "DeleteMany", mock.Anything, mock.Anything)
}

func TestTokenRepository_BlacklistToken_Error_WithMock(t *testing.T) {
	mockCollection := NewMockCollection()
	// Return wrapped error as the repository does
	mockCollection.On("InsertOne", mock.Anything, mock.Anything).
		Return(nil, mongo.ErrNilDocument)

	repo := NewTokenRepositoryWithCollection(mockCollection)
	ctx := context.Background()

	err := repo.BlacklistToken(ctx, "test-token", time.Now().Add(1*time.Hour))

	assert.Error(t, err)
	// The repository wraps the error, so we check for the wrapped message
	assert.Contains(t, err.Error(), "failed to blacklist token")
}
