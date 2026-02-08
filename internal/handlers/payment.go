package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/eddie-wainaina1/maggiesb/internal/database"
	"github.com/eddie-wainaina1/maggiesb/internal/models"
	mpesa "github.com/eddie-wainaina1/maggiesb/internal/payment"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

var mpesaClient *mpesa.Client

// InitMpesaClient initializes the M-Pesa client with credentials from environment
func InitMpesaClient() error {
	config := mpesa.Config{
		ConsumerKey:       os.Getenv("MPESA_CONSUMER_KEY"),
		ConsumerSecret:    os.Getenv("MPESA_CONSUMER_SECRET"),
		BusinessShortCode: os.Getenv("MPESA_BUSINESS_SHORTCODE"),
		PassKey:           os.Getenv("MPESA_PASSKEY"),
		CallbackURL:       os.Getenv("MPESA_CALLBACK_URL"),
		Environment:       os.Getenv("MPESA_ENV"),
	}

	if config.ConsumerKey == "" || config.ConsumerSecret == "" {
		return fmt.Errorf("M-Pesa credentials not configured")
	}

	if config.Environment == "" {
		config.Environment = "sandbox"
	}

	mpesaClient = mpesa.NewClient(config)
	return nil
}

// InitiateMpesaPayment initiates an M-Pesa STK Push payment for an invoice
func InitiateMpesaPayment(c *gin.Context) {
	if mpesaClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "M-Pesa client not initialized"})
		return
	}

	var req models.MpesaPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Get invoice
	invoiceRepo := database.NewInvoiceRepository()
	invoice, err := invoiceRepo.GetInvoiceByID(context.Background(), req.InvoiceID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve invoice"})
		return
	}

	// Verify user owns the order
	orderRepo := database.NewOrderRepository()
	order, err := orderRepo.GetOrderByID(context.Background(), invoice.OrderID)
	if err != nil || order.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// Initiate STK Push
	stkResp, err := mpesaClient.InitiateSTKPush(
		req.Phone,
		strconv.FormatFloat(invoice.InvoiceAmount, 'f', 2, 64),
		req.InvoiceID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create payment record
	paymentRepo := database.NewPaymentRepository()
	payment := &models.PaymentRecord{
		ID:                uuid.New().String(),
		InvoiceID:         req.InvoiceID,
		OrderID:           invoice.OrderID,
		CheckoutRequestID: stkResp.CheckoutRequestID,
		MerchantRequestID: stkResp.MerchantRequestID,
		Phone:             req.Phone,
		Amount:            invoice.InvoiceAmount,
		Status:            "initiated",
	}

	if err := paymentRepo.CreatePaymentRecord(context.Background(), payment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record payment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"checkoutRequestId": stkResp.CheckoutRequestID,
		"customerMessage":   stkResp.CustomerMessage,
		"paymentId":         payment.ID,
	})
}

// HandleMpesaCallback handles M-Pesa payment callback
func HandleMpesaCallback(c *gin.Context) {
	var callback models.MpesaCallback
	if err := c.ShouldBindJSON(&callback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid callback"})
		return
	}

	stkCallback := callback.Body.StkCallback
	paymentRepo := database.NewPaymentRepository()
	invoiceRepo := database.NewInvoiceRepository()

	// Get payment record
	payment, err := paymentRepo.GetPaymentByCheckoutRequestID(context.Background(), stkCallback.CheckoutRequestID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ResultCode": "1", "ResultDesc": "Payment not found"})
		return
	}

	// Update payment status
	status := "failed"
	receiptNum := ""
	transDate := ""

	if stkCallback.ResultCode == 0 {
		status = "completed"
		// Extract M-Pesa receipt number and transaction date from callback
		for _, item := range stkCallback.CallbackMetadata.Item {
			if item.Name == "MpesaReceiptNumber" {
				receiptNum = fmt.Sprintf("%v", item.Value)
			}
			if item.Name == "TransactionDate" {
				transDate = fmt.Sprintf("%v", item.Value)
			}
		}
	}

	// Update payment record
	if err := paymentRepo.UpdatePaymentStatus(context.Background(), stkCallback.CheckoutRequestID, status, receiptNum, transDate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ResultCode": "1", "ResultDesc": "Failed to update payment"})
		return
	}

	// If payment successful, record it in invoice
	if status == "completed" {
		invoice, err := invoiceRepo.GetInvoiceByID(context.Background(), payment.InvoiceID)
		if err == nil {
			// Record payment in invoice
			dateStr := transDate // Already formatted by M-Pesa
			if dateStr == "" {
				dateStr = "2006-01-02" // Fallback to today
			}
			invoiceRepo.RecordPayment(context.Background(), payment.InvoiceID, invoice.InvoiceAmount, dateStr)
		}
	}

	c.JSON(http.StatusOK, gin.H{"ResultCode": "0", "ResultDesc": "Callback received"})
}

// GetPaymentStatus retrieves payment status
func GetPaymentStatus(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Note: Add a GetPaymentByID method to the repository for this
	// For now, we'll skip ownership verification

	c.JSON(http.StatusOK, gin.H{"error": "not yet implemented"})
}
