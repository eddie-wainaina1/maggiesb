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

func TestInvoiceRepository_GetInvoicesByType_And_Count(t *testing.T) {
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

	// Create payable invoices
	for i := 0; i < 3; i++ {
		invoice := &models.Invoice{
			ID:            "inv-payable-" + string(rune(48+i)),
			OrderID:       "order-pay-" + string(rune(48+i)),
			InvoiceAmount: 100.0,
			PaidAmount:    0,
			TaxAmount:     0,
			Type:          models.InvoiceTypePayable,
			PaidOn:        make(map[string]float64),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		repo.CreateInvoice(ctx, invoice)
	}

	// Create receivable invoices
	for i := 0; i < 2; i++ {
		invoice := &models.Invoice{
			ID:            "inv-recv-" + string(rune(48+i)),
			OrderID:       "order-recv-" + string(rune(48+i)),
			InvoiceAmount: 50.0,
			PaidAmount:    50.0,
			TaxAmount:     0,
			Type:          models.InvoiceTypeReceivable,
			PaidOn:        map[string]float64{"2026-02-08": 50.0},
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		repo.CreateInvoice(ctx, invoice)
	}

	// Test GetInvoicesByType
	payables, err := repo.GetInvoicesByType(ctx, models.InvoiceTypePayable, 1, 10)
	if err != nil {
		t.Fatalf("GetInvoicesByType payable error: %v", err)
	}
	if len(payables) < 3 {
		t.Fatalf("expected at least 3 payable invoices, got %d", len(payables))
	}

	receivables, err := repo.GetInvoicesByType(ctx, models.InvoiceTypeReceivable, 1, 10)
	if err != nil {
		t.Fatalf("GetInvoicesByType receivable error: %v", err)
	}
	if len(receivables) < 2 {
		t.Fatalf("expected at least 2 receivable invoices, got %d", len(receivables))
	}

	// Test GetInvoiceCount
	count, err := repo.GetInvoiceCount(ctx)
	if err != nil {
		t.Fatalf("GetInvoiceCount error: %v", err)
	}
	if count < 5 {
		t.Fatalf("expected at least 5 invoices, got %d", count)
	}

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{})
}
