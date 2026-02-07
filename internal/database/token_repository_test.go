package database

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestTokenRepository_BlacklistAndCleanup(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping token repository tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewTokenRepository()
	ctx := context.Background()

	// ensure collection clean
	repo.collection.DeleteMany(ctx, map[string]interface{}{})

	token := "test-token-123"
	expires := time.Now().Add(1 * time.Hour)
	if err := repo.BlacklistToken(ctx, token, expires); err != nil {
		t.Fatalf("BlacklistToken error: %v", err)
	}

	isBlacklisted, err := repo.IsTokenBlacklisted(ctx, token)
	if err != nil {
		t.Fatalf("IsTokenBlacklisted error: %v", err)
	}
	if !isBlacklisted {
		t.Fatalf("expected token to be blacklisted")
	}

	// cleanup expired tokens and test deletion
	// insert expired token
	expiredToken := "expired-token"
	repo.collection.InsertOne(ctx, map[string]interface{}{"_id": expiredToken, "expiresAt": time.Now().Add(-1 * time.Hour), "createdAt": time.Now()})
	if err := repo.CleanupExpiredTokens(ctx); err != nil {
		t.Fatalf("CleanupExpiredTokens error: %v", err)
	}

	isBlacklisted, err = repo.IsTokenBlacklisted(ctx, expiredToken)
	if err != nil {
		t.Fatalf("IsTokenBlacklisted error: %v", err)
	}
	if isBlacklisted {
		t.Fatalf("expected expired token to be cleaned up")
	}
}
