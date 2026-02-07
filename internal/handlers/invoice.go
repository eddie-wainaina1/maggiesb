package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/eddie-wainaina1/maggiesb/internal/database"
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

	orderRepo := database.NewOrderRepository()
	invoiceRepo := database.NewInvoiceRepository()

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

	invoiceRepo := database.NewInvoiceRepository()
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
	orderRepo := database.NewOrderRepository()
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

	invoiceRepo := database.NewInvoiceRepository()

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

	invoiceRepo := database.NewInvoiceRepository()
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
