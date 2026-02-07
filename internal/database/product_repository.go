package database

import (
	"context"
	"fmt"
	"time"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ProductsCollectionName = "products"
)

type ProductRepository struct {
	collection *mongo.Collection
}

// NewProductRepository creates a new product repository
func NewProductRepository() *ProductRepository {
	return &ProductRepository{
		collection: GetCollection(DBName, ProductsCollectionName),
	}
}

// CreateProduct inserts a new product into the database
func (pr *ProductRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	_, err := pr.collection.InsertOne(ctx, product)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetProductByID retrieves a product by ID
func (pr *ProductRepository) GetProductByID(ctx context.Context, productID string) (*models.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var product models.Product
	err := pr.collection.FindOne(ctx, bson.M{"_id": productID}).Decode(&product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

// GetAllProducts retrieves all products with pagination
func (pr *ProductRepository) GetAllProducts(ctx context.Context, page, limit int) ([]*models.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().SetSkip(skip).SetLimit(int64(limit))

	cursor, err := pr.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}
	defer cursor.Close(ctx)

	var products []*models.Product
	err = cursor.All(ctx, &products)
	if err != nil {
		return nil, fmt.Errorf("failed to decode products: %w", err)
	}

	return products, nil
}

// UpdateProduct updates product information
func (pr *ProductRepository) UpdateProduct(ctx context.Context, productID string, updates map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	updates["updatedAt"] = time.Now()

	result, err := pr.collection.UpdateOne(ctx, bson.M{"_id": productID}, bson.M{"$set": updates})
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// DeleteProduct deletes a product by ID
func (pr *ProductRepository) DeleteProduct(ctx context.Context, productID string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := pr.collection.DeleteOne(ctx, bson.M{"_id": productID})
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// SearchProducts searches products by name or description
func (pr *ProductRepository) SearchProducts(ctx context.Context, query string, page, limit int) ([]*models.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().SetSkip(skip).SetLimit(int64(limit))

	filter := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	cursor, err := pr.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}
	defer cursor.Close(ctx)

	var products []*models.Product
	err = cursor.All(ctx, &products)
	if err != nil {
		return nil, fmt.Errorf("failed to decode products: %w", err)
	}

	return products, nil
}

// GetProductCount returns the total count of products
func (pr *ProductRepository) GetProductCount(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	count, err := pr.collection.EstimatedDocumentCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}

	return count, nil
}
