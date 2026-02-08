package handlers

import (
	"context"
	"net/http"
	"strconv"
	"fmt"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateOrder creates a new order for the authenticated user
func CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	productRepo := NewProductRepository
	orderRepo := NewOrderRepository

	// Build order items and calculate cost/discounts
	var items []models.OrderItem
	var cost float64
	var discountsTotal float64
	metadata := models.OrderMetadata{}
	if req.Metadata != nil {
		metadata = *req.Metadata
	}

	for _, p := range req.Products {
		prod, err := productRepo.GetProductByID(context.Background(), p.ProductID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("product not found: %s", p.ProductID)})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve product"})
			return
		}

		qty := p.Quantity
		if qty < 1 { qty = 1 }

		itemCost := prod.Price * float64(qty)
		itemDiscount := 0.0

		// apply product's inherent discount (percentage) if set
		if prod.Discount > 0 {
			itemDiscount += (prod.Discount / 100.0) * itemCost
		}

		// apply metadata overrides for per-product absolute discounts
		if metadata.Discounts != nil {
			if d, ok := metadata.Discounts[p.ProductID]; ok {
				itemDiscount += d
			}
		}

		cost += itemCost
		discountsTotal += itemDiscount

		items = append(items, models.OrderItem{
			ProductID: prod.ID,
			Quantity:  qty,
			Price:     prod.Price,
			Discount:  itemDiscount,
		})
	}

	totalCost := cost - discountsTotal
	if totalCost < 0 { totalCost = 0 }

	order := &models.Order{
		ID:        uuid.New().String(),
		Products:  items,
		Cost:      cost,
		Discount:  discountsTotal,
		TotalCost: totalCost,
		UserID:    userID.(string),
		Phone:     req.Phone,
		Metadata:  metadata,
	}

	if err := orderRepo.CreateOrder(context.Background(), order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	// Create corresponding invoice
	invoiceRepo := NewInvoiceRepository
	invoice := &models.Invoice{
		ID:            uuid.New().String(),
		OrderID:       order.ID,
		InvoiceAmount: order.TotalCost,
		PaidAmount:    0,
		TaxAmount:     0,
		Type:          models.InvoiceTypePayable,
		PaidOn:        make(map[string]float64),
	}

	if err := invoiceRepo.CreateInvoice(context.Background(), invoice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create invoice"})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// GetOrder returns a single order for the authenticated user
func GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	orderRepo := NewOrderRepository
	order, err := orderRepo.GetOrderByID(context.Background(), orderID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve order"})
		return
	}

	// ensure the user owns the order or has admin role
	if order.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// ListOrders lists orders for the authenticated user with pagination
func ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	orderRepo := NewOrderRepository
	orders, err := orderRepo.GetOrdersByUser(context.Background(), userID.(string), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve orders"})
		return
	}

	count, err := orderRepo.GetOrderCountByUser(context.Background(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": orders, "page": page, "limit": limit, "total": count})
}
