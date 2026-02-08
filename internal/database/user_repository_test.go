package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
)

func TestUserRepository_CRUD(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping user repository tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewUserRepository()
	ctx := context.Background()

	// cleanup test email before/after
	testEmail := "test-user@example.com"
	repo.collection.DeleteMany(ctx, map[string]interface{}{"email": testEmail})

	u := &models.User{
		ID:        "test-user-1",
		Email:     testEmail,
		Password:  "hashed",
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.CreateUser(ctx, u); err != nil {
		t.Fatalf("CreateUser error: %v", err)
	}

	found, err := repo.FindUserByEmail(ctx, testEmail)
	if err != nil {
		t.Fatalf("FindUserByEmail error: %v", err)
	}
	if found.Email != testEmail {
		t.Fatalf("expected email %s, got %s", testEmail, found.Email)
	}

	exists, err := repo.UserExists(ctx, testEmail)
	if err != nil {
		t.Fatalf("UserExists error: %v", err)
	}
	if !exists {
		t.Fatalf("expected user to exist")
	}

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{"email": testEmail})
}

func TestUserRepository_FindByID_Update_Delete(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping user repository tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewUserRepository()
	ctx := context.Background()

	testEmail := "test-full@example.com"
	repo.collection.DeleteMany(ctx, map[string]interface{}{"email": testEmail})

	u := &models.User{
		ID:        "test-user-full",
		Email:     testEmail,
		Password:  "hashed",
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.CreateUser(ctx, u); err != nil {
		t.Fatalf("CreateUser error: %v", err)
	}

	// Test FindUserByID
	foundByID, err := repo.FindUserByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("FindUserByID error: %v", err)
	}
	if foundByID.ID != u.ID {
		t.Fatalf("expected id %s, got %s", u.ID, foundByID.ID)
	}

	// Test UpdateUser
	u.FirstName = "Updated"
	if err := repo.UpdateUser(ctx, u.ID, u); err != nil {
		t.Fatalf("UpdateUser error: %v", err)
	}

	updated, err := repo.FindUserByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("FindUserByID after update error: %v", err)
	}
	if updated.FirstName != "Updated" {
		t.Fatalf("expected FirstName 'Updated', got %s", updated.FirstName)
	}

	// Test update nonexistent
	if err := repo.UpdateUser(ctx, "nonexistent", u); err == nil {
		t.Fatalf("expected error for nonexistent user")
	}

	// Test DeleteUser
	if err := repo.DeleteUser(ctx, u.ID); err != nil {
		t.Fatalf("DeleteUser error: %v", err)
	}

	_, err = repo.FindUserByID(ctx, u.ID)
	if err == nil {
		t.Fatalf("expected error for deleted user")
	}

	// Test delete nonexistent
	if err := repo.DeleteUser(ctx, "nonexistent"); err == nil {
		t.Fatalf("expected error for deleting nonexistent user")
	}

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{"email": testEmail})
}
