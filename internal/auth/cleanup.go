package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/eddie-wainaina1/maggiesb/internal/database"
)

// StartTokenCleanupRoutine starts a background goroutine that periodically cleans up expired tokens from MongoDB
// This prevents the token blacklist collection from growing unbounded
// interval: how often to run cleanup (e.g., 1 * time.Hour)
func StartTokenCleanupRoutine(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			CleanupBlacklist()
		}
	}()
}

// CleanupBlacklist removes expired tokens from the MongoDB blacklist
func CleanupBlacklist() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tokenRepo := database.NewTokenRepository()
	if err := tokenRepo.CleanupExpiredTokens(ctx); err != nil {
		fmt.Printf("Error cleaning up expired tokens: %v\n", err)
	}
}
