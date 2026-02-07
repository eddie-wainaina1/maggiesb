package auth

import (
	"os"
	"testing"
	"time"

	"github.com/eddie-wainaina1/maggiesb/internal/database"
)

func TestHashAndVerifyPassword(t *testing.T) {
	pw := "mysecretpassword"
	h, err := HashPassword(pw)
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}

	if err := VerifyPassword(h, pw); err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
}

func TestGenerateAndValidateToken(t *testing.T) {
	// ValidateToken depends on token blacklist in DB; skip if no test URI provided
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping token validation test")
	}

	if err := database.InitMongo(uri); err != nil {
		t.Fatalf("failed to init mongo: %v", err)
	}
	defer database.DisconnectMongo()

	SetSecretKey("test-secret-key")
	tkn, err := GenerateToken("userid123", "user@example.com", "user", 1*time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	claims, err := ValidateToken(tkn)
	if err != nil {
		t.Fatalf("ValidateToken error: %v", err)
	}

	if claims.UserID != "userid123" {
		t.Fatalf("unexpected user id: %s", claims.UserID)
	}
}
