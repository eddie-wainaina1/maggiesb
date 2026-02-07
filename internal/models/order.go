package models

import "time"

// OrderItem represents a product entry in an order
type OrderItem struct {
	ProductID string  `json:"productId" bson:"productId"`
	Quantity  int     `json:"quantity" bson:"quantity"`
	Price     float64 `json:"price" bson:"price"`    // snapshot price at purchase time
	Discount  float64 `json:"discount" bson:"discount"` // snapshot discount amount applied to this item (absolute)
}

// OrderMetadata holds additional order metadata
type OrderMetadata struct {
	Discounts      map[string]float64 `json:"discounts" bson:"discounts"`           // per-product absolute discount amounts
	Notes          string             `json:"notes" bson:"notes"`
	LocationDetails string            `json:"locationDetails" bson:"locationDetails"`
}

// Order represents a customer's order
type Order struct {
	ID         string         `json:"id" bson:"_id"`
	Products   []OrderItem    `json:"products" bson:"products"`
	Cost       float64        `json:"cost" bson:"cost"`                   // total before discounts
	Discount   float64        `json:"discount" bson:"discount"`           // total discount applied at creation (absolute)
	TotalCost  float64        `json:"totalCost" bson:"totalCost"`         // final cost after discounts
	Status     string         `json:"status" bson:"status"`
	UserID     string         `json:"user" bson:"user"`
	Phone      string         `json:"phone" bson:"phone"`
	Metadata   OrderMetadata  `json:"metadata" bson:"metadata"`
	CreatedAt  time.Time      `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt" bson:"updatedAt"`
}

// Order status values
const (
	OrderStatusInQueue      = "in queue"
	OrderStatusProcessing   = "processing"
	OrderStatusShipped      = "shipped"
	OrderStatusAwaitingPickup = "awaiting pick-up"
	OrderStatusComplete     = "order complete"
	OrderStatusCancelled    = "cancelled"
	OrderStatusReturned     = "returned"
)

// CreateOrderRequest is the payload to create an order
type CreateOrderRequest struct {
	Products []struct {
		ProductID string `json:"productId" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required,gt=0"`
	} `json:"products" binding:"required,min=1"`
	Phone    string            `json:"phone" binding:"required"`
	Metadata *OrderMetadata   `json:"metadata"`
}

// GetOrdersQuery used for listing orders
type GetOrdersQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}
