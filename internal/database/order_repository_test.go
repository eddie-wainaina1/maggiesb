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
