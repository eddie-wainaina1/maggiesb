package handlers

import (
	"context"
	"net/http"
	"strconv"
	"fmt"
	"time"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetInvoiceByOrder retrieves an invoice for the authenticated user's order
func GetInvoiceByOrder(c *gin.Context) {
	orderID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	orderRepo := NewOrderRepository
	invoiceRepo := NewInvoiceRepository

	// Verify the order belongs to the user
	order, err := orderRepo.GetOrderByID(context.Background(), orderID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve order"})
		return
	}

	if order.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// Get invoice for the order
	invoice, err := invoiceRepo.GetInvoiceByOrderID(context.Background(), orderID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve invoice"})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// GetInvoice retrieves a single invoice by ID (user-facing, checks ownership)
func GetInvoice(c *gin.Context) {
	invoiceID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	invoiceRepo := NewInvoiceRepository
	invoice, err := invoiceRepo.GetInvoiceByID(context.Background(), invoiceID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve invoice"})
		return
	}

	// Verify user owns the order associated with invoice
	orderRepo := NewOrderRepository
	order, err := orderRepo.GetOrderByID(context.Background(), invoice.OrderID)
	if err != nil || order.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// AdminListInvoices lists all invoices with optional type filter (admin)
func AdminListInvoices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	invoiceType := c.Query("type")

	invoiceRepo := NewInvoiceRepository

	var invoices []*models.Invoice
	var count int64
	var err error

	if invoiceType != "" {
		invoices, err = invoiceRepo.GetInvoicesByType(context.Background(), invoiceType, page, limit)
	} else {
		// For simplicity, use a generic query if no type filter
		// You could add a GetAllInvoices method to the repository
		invoices, err = invoiceRepo.GetInvoicesByType(context.Background(), "", page, limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve invoices"})
		return
	}

	count, err = invoiceRepo.GetInvoiceCount(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count invoices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": invoices, "page": page, "limit": limit, "total": count})
}

// AdminRecordPayment records a payment on an invoice (admin)
func AdminRecordPayment(c *gin.Context) {
	invoiceID := c.Param("id")

	var req models.RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoiceRepo := NewInvoiceRepository
	if err := invoiceRepo.RecordPayment(context.Background(), invoiceID, req.Amount, req.Date); err != nil {
		if err.Error() == "invoice not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return updated invoice
	invoice, err := invoiceRepo.GetInvoiceByID(context.Background(), invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve updated invoice"})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// AdminReverseInvoice performs a manual reversal/refund for an invoice.
func AdminReverseInvoice(c *gin.Context) {
	invoiceID := c.Param("id")

	var req models.ReverseInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoiceRepo := NewInvoiceRepository
	invoice, err := invoiceRepo.GetInvoiceByID(context.Background(), invoiceID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve invoice"})
		return
	}

	if invoice.Type != models.InvoiceTypePayable {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only payable invoices can be reversed via this endpoint"})
		return
	}

	// Determine amount to reverse
	amt := req.Amount
	if amt <= 0 {
		amt = invoice.PaidAmount
	}

	if amt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no paid amount available to reverse"})
		return
	}

	// Optional MPesa reversal
	if req.UseMpesa {
		if mpesaClient == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "M-Pesa client not initialized or not configured"})
			return
		}

		if err := mpesaClient.InitiateReversal(req.Phone, fmt.Sprintf("%.2f", amt), invoice.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("mpesa reversal failed: %v", err)})
			return
		}
	}

	// Apply DB reversals
	if amt >= invoice.PaidAmount {
		// full reversal
		if err := invoiceRepo.ReverseAllPayments(context.Background(), invoiceID, time.Now().Format("2006-01-02")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reverse invoice payments"})
			return
		}

		paymentRepo := NewPaymentRepository
		if err := paymentRepo.ReversePaymentsByInvoiceID(context.Background(), invoiceID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark payment records reversed"})
			return
		}
	} else {
		// partial reversal
		dateStr := req.Date
		if dateStr == "" {
			dateStr = time.Now().Format("2006-01-02")
		}
		if err := invoiceRepo.ReversePaymentAmount(context.Background(), invoiceID, amt, dateStr); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to apply partial reversal"})
			return
		}
	}

	// Record reversal audit
	adminID, _ := c.Get("userID")
	revRepo := NewReversalRepository
	rev := &models.ReversalRecord{
		ID:        fmt.Sprintf("rev_%s_%d", invoiceID, time.Now().Unix()),
		InvoiceID: invoiceID,
		Amount:    amt,
		Date:      req.Date,
		Phone:     req.Phone,
		AdminID:   adminID.(string),
		Reason:    req.Reason,
	}
	_ = revRepo.CreateReversalRecord(context.Background(), rev)

	updated, err := invoiceRepo.GetInvoiceByID(context.Background(), invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve updated invoice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invoice": updated, "reversal": rev})
}
