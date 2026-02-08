package database

import (
	"context"
	"os"
	"testing"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
)

func TestPaymentRepository_CreateAndGet(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping payment repository tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewPaymentRepository()
	ctx := context.Background()

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{})

	payment := &models.PaymentRecord{
		ID:                "pay-test-1",
		InvoiceID:         "inv-test-1",
		OrderID:           "order-test-1",
		CheckoutRequestID: "chk-1",
		MerchantRequestID: "m-1",
		Phone:             "254700000001",
		Amount:            100.0,
		Status:            "initiated",
	}

	if err := repo.CreatePaymentRecord(ctx, payment); err != nil {
		t.Fatalf("CreatePaymentRecord error: %v", err)
	}

	found, err := repo.GetPaymentByCheckoutRequestID(ctx, payment.CheckoutRequestID)
	if err != nil {
		t.Fatalf("GetPaymentByCheckoutRequestID error: %v", err)
	}
	if found.ID != payment.ID {
		t.Fatalf("expected id %s, got %s", payment.ID, found.ID)
	}

	foundByInv, err := repo.GetPaymentByInvoiceID(ctx, payment.InvoiceID)
	if err != nil {
		t.Fatalf("GetPaymentByInvoiceID error: %v", err)
	}
	if foundByInv.InvoiceID != payment.InvoiceID {
		t.Fatalf("expected invoiceId %s, got %s", payment.InvoiceID, foundByInv.InvoiceID)
	}

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{})
}

func TestPaymentRepository_UpdatePaymentStatus(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping payment repository tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewPaymentRepository()
	ctx := context.Background()

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{})

	payment := &models.PaymentRecord{
		ID:                "pay-update-1",
		InvoiceID:         "inv-upd-1",
		OrderID:           "order-upd-1",
		CheckoutRequestID: "chk-upd-1",
		MerchantRequestID: "m-upd-1",
		Phone:             "254700000002",
		Amount:            50.0,
		Status:            "initiated",
	}

	if err := repo.CreatePaymentRecord(ctx, payment); err != nil {
		t.Fatalf("CreatePaymentRecord error: %v", err)
	}

	if err := repo.UpdatePaymentStatus(ctx, payment.CheckoutRequestID, "completed", "mpesa-receipt-123", "20260208"); err != nil {
		t.Fatalf("UpdatePaymentStatus error: %v", err)
	}

	updated, err := repo.GetPaymentByCheckoutRequestID(ctx, payment.CheckoutRequestID)
	if err != nil {
		t.Fatalf("GetPaymentByCheckoutRequestID after update error: %v", err)
	}
	if updated.Status != "completed" {
		t.Fatalf("expected status 'completed', got %s", updated.Status)
	}
	if updated.MpesaReceiptNumber != "mpesa-receipt-123" {
		t.Fatalf("expected receipt 'mpesa-receipt-123', got %s", updated.MpesaReceiptNumber)
	}

	// Test update nonexistent
	if err := repo.UpdatePaymentStatus(ctx, "nonexistent-chk", "completed", "", ""); err == nil {
		t.Fatalf("expected error for nonexistent payment")
	}

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{})
}

func TestPaymentRepository_ReversePaymentsByInvoiceID(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping payment repository tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewPaymentRepository()
	ctx := context.Background()

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{})

	invoiceID := "inv-rev-1"

	// Create multiple payments for same invoice
	p1 := &models.PaymentRecord{
		ID:                "pay-rev-1",
		InvoiceID:         invoiceID,
		OrderID:           "order-rev-1",
		CheckoutRequestID: "chk-rev-1",
		MerchantRequestID: "m-rev-1",
		Phone:             "254700000003",
		Amount:            30.0,
		Status:            "completed",
	}
	p2 := &models.PaymentRecord{
		ID:                "pay-rev-2",
		InvoiceID:         invoiceID,
		OrderID:           "order-rev-1",
		CheckoutRequestID: "chk-rev-2",
		MerchantRequestID: "m-rev-2",
		Phone:             "254700000004",
		Amount:            20.0,
		Status:            "completed",
	}

	repo.CreatePaymentRecord(ctx, p1)
	repo.CreatePaymentRecord(ctx, p2)

	// Reverse all payments for invoice
	if err := repo.ReversePaymentsByInvoiceID(ctx, invoiceID); err != nil {
		t.Fatalf("ReversePaymentsByInvoiceID error: %v", err)
	}

	// Verify both updated to reversed
	cnt, err := repo.collection.CountDocuments(ctx, map[string]interface{}{"invoiceId": invoiceID, "status": "reversed"})
	if err != nil {
		t.Fatalf("count documents error: %v", err)
	}
	if cnt != 2 {
		t.Fatalf("expected 2 reversed payments, got %d", cnt)
	}

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{})
}
