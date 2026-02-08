package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
)

func TestReversalAndInvoiceReversalFlow(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping reversal tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	invRepo := NewInvoiceRepository()
	payRepo := NewPaymentRepository()
	revRepo := NewReversalRepository()
	ctx := context.Background()

	// Cleanup
	invRepo.collection.DeleteMany(ctx, map[string]interface{}{})
	payRepo.collection.DeleteMany(ctx, map[string]interface{}{})
	revRepo.collection.DeleteMany(ctx, map[string]interface{}{})

	// Create invoice with paid amount
	invoice := &models.Invoice{
		ID:            "inv-rev-test-1",
		OrderID:       "order-rev-1",
		InvoiceAmount: 200.0,
		PaidAmount:    100.0,
		TaxAmount:     0,
		Type:          models.InvoiceTypePayable,
		PaidOn:        map[string]float64{"2026-02-08": 100.0},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := invRepo.CreateInvoice(ctx, invoice); err != nil {
		t.Fatalf("CreateInvoice error: %v", err)
	}

	// Create two payment records
	p1 := &models.PaymentRecord{
		ID:                "pay-1",
		InvoiceID:         invoice.ID,
		OrderID:           invoice.OrderID,
		CheckoutRequestID: "chk-1",
		MerchantRequestID: "m-1",
		Phone:             "254700000001",
		Amount:            60.0,
		Status:            "completed",
	}
	p2 := &models.PaymentRecord{
		ID:                "pay-2",
		InvoiceID:         invoice.ID,
		OrderID:           invoice.OrderID,
		CheckoutRequestID: "chk-2",
		MerchantRequestID: "m-2",
		Phone:             "254700000002",
		Amount:            40.0,
		Status:            "completed",
	}

	if err := payRepo.CreatePaymentRecord(ctx, p1); err != nil {
		t.Fatalf("CreatePaymentRecord p1 error: %v", err)
	}
	if err := payRepo.CreatePaymentRecord(ctx, p2); err != nil {
		t.Fatalf("CreatePaymentRecord p2 error: %v", err)
	}

	// Partial reversal of 30
	if err := invRepo.ReversePaymentAmount(ctx, invoice.ID, 30.0, "2026-02-09"); err != nil {
		t.Fatalf("ReversePaymentAmount error: %v", err)
	}

	invAfter, err := invRepo.GetInvoiceByID(ctx, invoice.ID)
	if err != nil {
		t.Fatalf("GetInvoiceByID after partial error: %v", err)
	}
	if invAfter.PaidAmount != 70.0 {
		t.Fatalf("expected paid amount 70.0 after partial reversal, got %f", invAfter.PaidAmount)
	}
	if v, ok := invAfter.PaidOn["2026-02-09"]; !ok || v != -30.0 {
		t.Fatalf("expected paidOn entry -30.0, got %v (ok=%v)", v, ok)
	}

	// Full reversal
	if err := invRepo.ReverseAllPayments(ctx, invoice.ID, "2026-02-10"); err != nil {
		t.Fatalf("ReverseAllPayments error: %v", err)
	}

	invFinal, err := invRepo.GetInvoiceByID(ctx, invoice.ID)
	if err != nil {
		t.Fatalf("GetInvoiceByID after full reversal error: %v", err)
	}
	if invFinal.PaidAmount != 0 {
		t.Fatalf("expected paid amount 0 after full reversal, got %f", invFinal.PaidAmount)
	}
	if len(invFinal.PaidOn) != 0 {
		t.Fatalf("expected paidOn map empty after full reversal, got %v", invFinal.PaidOn)
	}
	if invFinal.Type != models.InvoiceTypeReceivable {
		t.Fatalf("expected invoice type receivable after reversal, got %s", invFinal.Type)
	}

	// Mark payment records reversed
	if err := payRepo.ReversePaymentsByInvoiceID(ctx, invoice.ID); err != nil {
		t.Fatalf("ReversePaymentsByInvoiceID error: %v", err)
	}

	// Verify payment records updated
	count, err := payRepo.collection.CountDocuments(ctx, map[string]interface{}{"invoiceId": invoice.ID, "status": "reversed"})
	if err != nil {
		t.Fatalf("count documents error: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected some payment records with status reversed, got count=0")
	}

	// Create reversal audit record
	rev := &models.ReversalRecord{
		ID:        "rev-1",
		InvoiceID: invoice.ID,
		Amount:    100.0,
		Date:      "2026-02-10",
		Phone:     "254700000001",
		AdminID:   "admin-1",
		Reason:    "Customer returned items",
	}
	if err := revRepo.CreateReversalRecord(ctx, rev); err != nil {
		t.Fatalf("CreateReversalRecord error: %v", err)
	}

	// Verify reversal record exists
	cnt, err := revRepo.collection.CountDocuments(ctx, map[string]interface{}{"invoiceId": invoice.ID})
	if err != nil {
		t.Fatalf("count reversal docs error: %v", err)
	}
	if cnt == 0 {
		t.Fatalf("expected reversal record inserted, got count=0")
	}

	// cleanup
	invRepo.collection.DeleteMany(ctx, map[string]interface{}{})
	payRepo.collection.DeleteMany(ctx, map[string]interface{}{})
	revRepo.collection.DeleteMany(ctx, map[string]interface{}{})
}
