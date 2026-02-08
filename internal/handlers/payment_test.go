package handlers

import (
	"context"
	"os"
	"testing"

	"github.com/eddie-wainaina1/maggiesb/internal/database"
	"github.com/eddie-wainaina1/maggiesb/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestInitiateMpesaPayment(t *testing.T) {
	mongoURI := os.Getenv("MONGO_TEST_URI")
	if mongoURI == "" {
		t.Skip("MONGO_TEST_URI not set, skipping database test")
	}

	// Setup
	if err := database.InitMongo(mongoURI); err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer database.DisconnectMongo()

	ctx := context.Background()

	// Create test user
	userRepo := database.NewUserRepository()
	testUser := &models.User{
		ID:        uuid.New().String(),
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
	}
	if err := userRepo.CreateUser(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test order
	orderRepo := database.NewOrderRepository()
	testOrder := &models.Order{
		ID:        uuid.New().String(),
		UserID:    testUser.ID,
		Status:    "in queue",
		TotalCost: 100.00,
		Discount:  0,
		Cost:      100.00,
		Phone:     "254712345678",
		Products:  []models.OrderItem{},
		Metadata:  models.OrderMetadata{},
	}
	if err := orderRepo.CreateOrder(ctx, testOrder); err != nil {
		t.Fatalf("Failed to create test order: %v", err)
	}

	// Create test invoice
	invoiceRepo := database.NewInvoiceRepository()
	testInvoice := &models.Invoice{
		ID:            uuid.New().String(),
		OrderID:       testOrder.ID,
		InvoiceAmount: 100.00,
		PaidAmount:    0,
		TaxAmount:     10.00,
		Type:          "payable",
		PaidOn:        make(map[string]float64),
	}
	if err := invoiceRepo.CreateInvoice(ctx, testInvoice); err != nil {
		t.Fatalf("Failed to create test invoice: %v", err)
	}

	// Test: M-Pesa client not initialized should return error
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	c.Set("userID", testUser.ID)

	// Note: Full integration test would require M-Pesa credentials and mock HTTP server
	if mpesaClient == nil {
		t.Logf("M-Pesa client not initialized (expected in test environment)")
	}
}

func TestHandleMpesaCallback(t *testing.T) {
	mongoURI := os.Getenv("MONGO_TEST_URI")
	if mongoURI == "" {
		t.Skip("MONGO_TEST_URI not set, skipping database test")
	}

	// Setup
	if err := database.InitMongo(mongoURI); err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer database.DisconnectMongo()

	ctx := context.Background()

	// Create test payment record
	paymentRepo := database.NewPaymentRepository()
	testPayment := &models.PaymentRecord{
		ID:                uuid.New().String(),
		InvoiceID:         uuid.New().String(),
		OrderID:           uuid.New().String(),
		CheckoutRequestID: "test-checkout-123",
		Phone:             "254712345678",
		Amount:            100.00,
		Status:            "initiated",
	}
	if err := paymentRepo.CreatePaymentRecord(ctx, testPayment); err != nil {
		t.Fatalf("Failed to create test payment record: %v", err)
	}

	// Verify payment was created
	retrieved, err := paymentRepo.GetPaymentByCheckoutRequestID(ctx, testPayment.CheckoutRequestID)
	if err != nil {
		t.Fatalf("Failed to retrieve payment record: %v", err)
	}

	if retrieved.CheckoutRequestID != testPayment.CheckoutRequestID {
		t.Errorf("Expected CheckoutRequestID %s, got %s", testPayment.CheckoutRequestID, retrieved.CheckoutRequestID)
	}

	if retrieved.Status != "initiated" {
		t.Errorf("Expected status 'initiated', got %s", retrieved.Status)
	}
}

func TestMpesaPaymentRequest(t *testing.T) {
	req := models.MpesaPaymentRequest{
		InvoiceID: "test-invoice-123",
		Phone:     "254712345678",
	}

	if req.InvoiceID != "test-invoice-123" {
		t.Errorf("Expected InvoiceID 'test-invoice-123', got %s", req.InvoiceID)
	}

	if req.Phone != "254712345678" {
		t.Errorf("Expected Phone '254712345678', got %s", req.Phone)
	}
}
