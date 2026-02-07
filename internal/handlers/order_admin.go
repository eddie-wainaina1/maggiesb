package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/eddie-wainaina1/maggiesb/internal/database"
	"github.com/eddie-wainaina1/maggiesb/internal/models"
	"github.com/gin-gonic/gin"
)

// AdminListOrders lists all orders (admin)
func AdminListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	orderRepo := database.NewOrderRepository()
	orders, err := orderRepo.GetAllOrders(context.Background(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve orders"})
		return
	}

	count, err := orderRepo.GetOrderCount(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": orders, "page": page, "limit": limit, "total": count})
}

// AdminUpdateOrderStatus updates an order's status (admin)
func AdminUpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	var payload struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status
	s := payload.Status
	switch s {
	case models.OrderStatusInQueue, models.OrderStatusProcessing, models.OrderStatusShipped, models.OrderStatusAwaitingPickup, models.OrderStatusComplete, models.OrderStatusCancelled, models.OrderStatusReturned:
		// ok
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
		return
	}

	orderRepo := database.NewOrderRepository()
	if err := orderRepo.UpdateOrderStatus(context.Background(), orderID, s); err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order status"})
		return
	}

	// Return updated order
	order, err := orderRepo.GetOrderByID(context.Background(), orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve updated order"})
		return
	}

	c.JSON(http.StatusOK, order)
}
