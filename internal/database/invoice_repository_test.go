package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
)

func TestInvoiceRepository_CreateAndGet(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping invoice repository tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewInvoiceRepository()
	ctx := context.Background()

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{})

	invoice := &models.Invoice{
		ID:            "inv-test-1",
		OrderID:       "order-123",
		InvoiceAmount: 100.0,
		PaidAmount:    0,
		TaxAmount:     10.0,
		Type:          models.InvoiceTypePayable,
		PaidOn:        make(map[string]float64),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := repo.CreateInvoice(ctx, invoice); err != nil {
		t.Fatalf("CreateInvoice error: %v", err)
	}

	f, err := repo.GetInvoiceByID(ctx, invoice.ID)
	if err != nil {
		t.Fatalf("GetInvoiceByID error: %v", err)
	}
	if f.ID != invoice.ID {
		t.Fatalf("expected id %s, got %s", invoice.ID, f.ID)
	}

	f2, err := repo.GetInvoiceByOrderID(ctx, invoice.OrderID)
	if err != nil {
		t.Fatalf("GetInvoiceByOrderID error: %v", err)
	}
	if f2.OrderID != invoice.OrderID {
		t.Fatalf("expected order id %s, got %s", invoice.OrderID, f2.OrderID)
	}
}

func TestInvoiceRepository_RecordPayment(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping invoice repository tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewInvoiceRepository()
	ctx := context.Background()

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{})

	invoice := &models.Invoice{
		ID:            "inv-test-2",
		OrderID:       "order-456",
		InvoiceAmount: 100.0,
		PaidAmount:    0,
		TaxAmount:     0,
		Type:          models.InvoiceTypePayable,
		PaidOn:        make(map[string]float64),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := repo.CreateInvoice(ctx, invoice); err != nil {
		t.Fatalf("CreateInvoice error: %v", err)
	}

	// Record payment
	if err := repo.RecordPayment(ctx, invoice.ID, 50.0, "2026-02-08"); err != nil {
		t.Fatalf("RecordPayment error: %v", err)
	}

	f, err := repo.GetInvoiceByID(ctx, invoice.ID)
	if err != nil {
		t.Fatalf("GetInvoiceByID error: %v", err)
	}
	if f.PaidAmount != 50.0 {
		t.Fatalf("expected paid amount 50.0, got %f", f.PaidAmount)
	}
	if v, ok := f.PaidOn["2026-02-08"]; !ok || v != 50.0 {
		t.Fatalf("expected payment entry not found in paidOn map")
	}

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{})
}
