package database

import (
	"context"
	"os"
	"testing"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
)

func TestReversalRepository_CreateReversalRecord(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping reversal repository tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewReversalRepository()
	ctx := context.Background()

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{})

	rev := &models.ReversalRecord{
		ID:        "rev-1",
		InvoiceID: "inv-123",
		Amount:    100.0,
		Date:      "2026-02-08",
		Phone:     "254700000001",
		AdminID:   "admin-1",
		Reason:    "Customer returned items",
	}

	if err := repo.CreateReversalRecord(ctx, rev); err != nil {
		t.Fatalf("CreateReversalRecord error: %v", err)
	}

	// verify it was inserted
	var result models.ReversalRecord
	if err := repo.collection.FindOne(ctx, map[string]interface{}{"_id": rev.ID}).Decode(&result); err != nil {
		t.Fatalf("failed to find inserted reversal: %v", err)
	}

	if result.InvoiceID != rev.InvoiceID {
		t.Fatalf("expected invoiceId %s, got %s", rev.InvoiceID, result.InvoiceID)
	}
	if result.Amount != rev.Amount {
		t.Fatalf("expected amount %.2f, got %.2f", rev.Amount, result.Amount)
	}

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{})
}
