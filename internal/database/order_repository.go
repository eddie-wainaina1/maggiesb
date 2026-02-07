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
	OrdersCollectionName = "orders"
)

type OrderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{collection: GetCollection(DBName, OrdersCollectionName)}
}

func (or *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	_, err := or.collection.InsertOne(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

func (or *OrderRepository) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var order models.Order
	err := or.collection.FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (or *OrderRepository) GetOrdersByUser(ctx context.Context, userID string, page, limit int) ([]*models.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if page < 1 { page = 1 }
	if limit < 1 { limit = 10 }

	skip := int64((page-1)*limit)
	opts := options.Find().SetSkip(skip).SetLimit(int64(limit)).SetSort(bson.D{{Key:"createdAt", Value:-1}})

	cursor, err := or.collection.Find(ctx, bson.M{"user": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []*models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("failed to decode orders: %w", err)
	}
	return orders, nil
}

func (or *OrderRepository) GetAllOrders(ctx context.Context, page, limit int) ([]*models.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if page < 1 { page = 1 }
	if limit < 1 { limit = 10 }

	skip := int64((page-1)*limit)
	opts := options.Find().SetSkip(skip).SetLimit(int64(limit)).SetSort(bson.D{{Key:"createdAt", Value:-1}})

	cursor, err := or.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []*models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("failed to decode orders: %w", err)
	}
	return orders, nil
}

func (or *OrderRepository) GetOrderCountByUser(ctx context.Context, userID string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	count, err := or.collection.CountDocuments(ctx, bson.M{"user": userID})
	if err != nil {
		return 0, fmt.Errorf("failed to count orders: %w", err)
	}
	return count, nil
}

func (or *OrderRepository) GetOrderCount(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	count, err := or.collection.EstimatedDocumentCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count orders: %w", err)
	}
	return count, nil
}

// UpdateOrderStatus updates only the status of an order
func (or *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := or.collection.UpdateOne(ctx, bson.M{"_id": orderID}, bson.M{"$set": bson.M{"status": status, "updatedAt": time.Now()}})
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}
