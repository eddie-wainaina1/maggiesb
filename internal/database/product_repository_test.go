package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
)

func TestProductRepository_CreateAndGet(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping product repository tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewProductRepository()
	ctx := context.Background()

	// cleanup test product
	testName := "test-product"
	repo.collection.DeleteMany(ctx, map[string]interface{}{"name": testName})

	p := &models.Product{
		ID: "test-prod-1",
		Name: testName,
		Description: "desc",
		Price: 10.0,
		Discount: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.CreateProduct(ctx, p); err != nil {
		t.Fatalf("CreateProduct error: %v", err)
	}

	f, err := repo.GetProductByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetProductByID error: %v", err)
	}
	if f.Name != p.Name {
		t.Fatalf("expected name %s, got %s", p.Name, f.Name)
	}

	// cleanup
	repo.collection.DeleteMany(ctx, map[string]interface{}{"name": testName})
}
