package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
)

func TestOrderRepository_CreateAndGet(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping order repository tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewOrderRepository()
	ctx := context.Background()

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{})

	order := &models.Order{
		ID: "test-order-1",
		Products: []models.OrderItem{{ProductID: "p1", Quantity: 1, Price: 5.0, Discount: 0}},
		Cost: 5.0,
		Discount: 0,
		TotalCost: 5.0,
		UserID: "u1",
		Phone: "123456",
		Metadata: models.OrderMetadata{Discounts: map[string]float64{}},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.CreateOrder(ctx, order); err != nil {
		t.Fatalf("CreateOrder error: %v", err)
	}

	f, err := repo.GetOrderByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetOrderByID error: %v", err)
	}
	if f.ID != order.ID {
		t.Fatalf("expected id %s, got %s", order.ID, f.ID)
	}

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{})
}

func TestOrderRepository_GetOrdersByUser(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping order repository tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewOrderRepository()
	ctx := context.Background()
	repo.collection.DeleteMany(ctx, map[string]interface{}{})

	userID := "user-1"
	for i := 0; i < 15; i++ {
		order := &models.Order{
			ID:        "order-" + string(rune(48+i)),
			Products:  []models.OrderItem{},
			Cost:      10.0,
			Discount:  0,
			TotalCost: 10.0,
			UserID:    userID,
			Phone:     "123456",
		}
		repo.CreateOrder(ctx, order)
	}

	orders, err := repo.GetOrdersByUser(ctx, userID, 1, 10)
	if err != nil {
		t.Fatalf("GetOrdersByUser error: %v", err)
	}
	if len(orders) != 10 {
		t.Fatalf("expected 10 orders, got %d", len(orders))
	}

	orders, err = repo.GetOrdersByUser(ctx, userID, 2, 10)
	if err != nil {
		t.Fatalf("GetOrdersByUser page 2 error: %v", err)
	}
	if len(orders) != 5 {
		t.Fatalf("expected 5 orders on page 2, got %d", len(orders))
	}

	repo.collection.DeleteMany(ctx, map[string]interface{}{})
}

func TestOrderRepository_UpdateOrderStatus(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping order repository tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewOrderRepository()
	ctx := context.Background()
	repo.collection.DeleteMany(ctx, map[string]interface{}{})

	order := &models.Order{
		ID:        "update-order-1",
		Products:  []models.OrderItem{},
		Cost:      10.0,
		Discount:  0,
		TotalCost: 10.0,
		UserID:    "user-1",
		Phone:     "123456",
		Status:    models.OrderStatusInQueue,
	}
	repo.CreateOrder(ctx, order)

	if err := repo.UpdateOrderStatus(ctx, order.ID, models.OrderStatusProcessing); err != nil {
		t.Fatalf("UpdateOrderStatus error: %v", err)
	}

	updated, err := repo.GetOrderByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetOrderByID error: %v", err)
	}
	if updated.Status != models.OrderStatusProcessing {
		t.Fatalf("expected status %s, got %s", models.OrderStatusProcessing, updated.Status)
	}

	// test not found
	if err := repo.UpdateOrderStatus(ctx, "nonexistent", models.OrderStatusShipped); err == nil {
		t.Fatalf("expected error for nonexistent order")
	}

	repo.collection.DeleteMany(ctx, map[string]interface{}{})
}
