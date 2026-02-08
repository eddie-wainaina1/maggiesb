package database

import (
	"context"
	"fmt"
	"time"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	PaymentRecordsCollectionName = "payment_records"
)

type PaymentRepository struct {
	collection *mongo.Collection
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository() *PaymentRepository {
	return &PaymentRepository{collection: GetCollection(DBName, PaymentRecordsCollectionName)}
}

// CreatePaymentRecord inserts a new payment record
func (pr *PaymentRepository) CreatePaymentRecord(ctx context.Context, record *models.PaymentRecord) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	record.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	record.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")

	_, err := pr.collection.InsertOne(ctx, record)
	if err != nil {
		return fmt.Errorf("failed to create payment record: %w", err)
	}
	return nil
}

// GetPaymentByCheckoutRequestID retrieves a payment by checkout request ID
func (pr *PaymentRepository) GetPaymentByCheckoutRequestID(ctx context.Context, checkoutID string) (*models.PaymentRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var payment models.PaymentRecord
	err := pr.collection.FindOne(ctx, bson.M{"checkoutRequestId": checkoutID}).Decode(&payment)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// GetPaymentByInvoiceID retrieves a payment by invoice ID
func (pr *PaymentRepository) GetPaymentByInvoiceID(ctx context.Context, invoiceID string) (*models.PaymentRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var payment models.PaymentRecord
	err := pr.collection.FindOne(ctx, bson.M{"invoiceId": invoiceID}).Decode(&payment)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// UpdatePaymentStatus updates payment record status and transaction details
func (pr *PaymentRepository) UpdatePaymentStatus(ctx context.Context, checkoutID string, status, receiptNum, transDate string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	update := bson.M{
		"status":                status,
		"mpesaReceiptNumber":    receiptNum,
		"transactionDate":       transDate,
		"updatedAt":             time.Now().Format("2006-01-02 15:04:05"),
	}

	result, err := pr.collection.UpdateOne(ctx, bson.M{"checkoutRequestId": checkoutID}, bson.M{"$set": update})
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("payment record not found")
	}

	return nil
}

// ReversePaymentsByInvoiceID marks all payment records for an invoice as reversed
func (pr *PaymentRepository) ReversePaymentsByInvoiceID(ctx context.Context, invoiceID string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"status":    "reversed",
			"updatedAt": time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	_, err := pr.collection.UpdateMany(ctx, bson.M{"invoiceId": invoiceID}, update)
	if err != nil {
		return fmt.Errorf("failed to mark payments reversed: %w", err)
	}
	return nil
}
