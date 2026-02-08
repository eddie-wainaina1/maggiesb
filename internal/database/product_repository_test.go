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

func TestProductRepository_GetAll_Search_Update_Delete(t *testing.T) {
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
	repo.collection.DeleteMany(ctx, map[string]interface{}{})

	// Create test products
	p1 := &models.Product{
		ID:          "test-prod-all-1",
		Name:        "laptop",
		Description: "computer device",
		Price:       500.0,
		Discount:    5,
	}
	p2 := &models.Product{
		ID:          "test-prod-all-2",
		Name:        "mouse",
		Description: "pointing device",
		Price:       25.0,
		Discount:    0,
	}
	repo.CreateProduct(ctx, p1)
	repo.CreateProduct(ctx, p2)

	// Test GetAllProducts
	products, err := repo.GetAllProducts(ctx, 1, 10)
	if err != nil {
		t.Fatalf("GetAllProducts error: %v", err)
	}
	if len(products) < 2 {
		t.Fatalf("expected at least 2 products, got %d", len(products))
	}

	// Test GetProductCount
	count, err := repo.GetProductCount(ctx)
	if err != nil {
		t.Fatalf("GetProductCount error: %v", err)
	}
	if count < 2 {
		t.Fatalf("expected count >= 2, got %d", count)
	}

	// Test SearchProducts
	results, err := repo.SearchProducts(ctx, "laptop", 1, 10)
	if err != nil {
		t.Fatalf("SearchProducts error: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected search results for 'laptop'")
	}

	results, err = repo.SearchProducts(ctx, "device", 1, 10)
	if err != nil {
		t.Fatalf("SearchProducts desc error: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected search results for 'device' in descriptions")
	}

	// Test UpdateProduct
	updates := map[string]interface{}{
		"price":    450.0,
		"discount": 10,
	}
	if err := repo.UpdateProduct(ctx, p1.ID, updates); err != nil {
		t.Fatalf("UpdateProduct error: %v", err)
	}

	updated, err := repo.GetProductByID(ctx, p1.ID)
	if err != nil {
		t.Fatalf("GetProductByID after update error: %v", err)
	}
	if updated.Price != 450.0 {
		t.Fatalf("expected price 450.0, got %.2f", updated.Price)
	}

	// Test update nonexistent
	if err := repo.UpdateProduct(ctx, "nonexistent", updates); err == nil {
		t.Fatalf("expected error for nonexistent product")
	}

	// Test DeleteProduct
	if err := repo.DeleteProduct(ctx, p2.ID); err != nil {
		t.Fatalf("DeleteProduct error: %v", err)
	}

	_, err = repo.GetProductByID(ctx, p2.ID)
	if err == nil {
		t.Fatalf("expected error for deleted product")
	}

	// Test delete nonexistent
	if err := repo.DeleteProduct(ctx, "nonexistent"); err == nil {
		t.Fatalf("expected error for deleting nonexistent product")
	}

	repo.collection.DeleteMany(ctx, map[string]interface{}{})
}
