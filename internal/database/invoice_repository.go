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
	InvoicesCollectionName = "invoices"
)

type InvoiceRepository struct {
	collection *mongo.Collection
}

// NewInvoiceRepository creates a new invoice repository
func NewInvoiceRepository() *InvoiceRepository {
	return &InvoiceRepository{collection: GetCollection(DBName, InvoicesCollectionName)}
}

// CreateInvoice inserts a new invoice
func (ir *InvoiceRepository) CreateInvoice(ctx context.Context, invoice *models.Invoice) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	invoice.CreatedAt = time.Now()
	invoice.UpdatedAt = time.Now()
	if invoice.PaidOn == nil {
		invoice.PaidOn = make(map[string]float64)
	}

	_, err := ir.collection.InsertOne(ctx, invoice)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}
	return nil
}

// GetInvoiceByID retrieves an invoice by ID
func (ir *InvoiceRepository) GetInvoiceByID(ctx context.Context, invoiceID string) (*models.Invoice, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var invoice models.Invoice
	err := ir.collection.FindOne(ctx, bson.M{"_id": invoiceID}).Decode(&invoice)
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// GetInvoiceByOrderID retrieves an invoice by order ID
func (ir *InvoiceRepository) GetInvoiceByOrderID(ctx context.Context, orderID string) (*models.Invoice, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var invoice models.Invoice
	err := ir.collection.FindOne(ctx, bson.M{"orderId": orderID}).Decode(&invoice)
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// RecordPayment records a payment on an invoice
func (ir *InvoiceRepository) RecordPayment(ctx context.Context, invoiceID string, amount float64, dateStr string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Parse date
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format, use YYYY-MM-DD: %w", err)
	}

	// Update PaidOn map and PaidAmount
	result, err := ir.collection.UpdateOne(
		ctx,
		bson.M{"_id": invoiceID},
		bson.M{
			"$inc": bson.M{
				"paidAmount":               amount,
				fmt.Sprintf("paidOn.%s", dateStr): amount,
			},
			"$set": bson.M{
				"updatedAt": time.Now(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to record payment: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

// ReverseAllPayments clears paid amounts on an invoice (used for cancellations/returns)
func (ir *InvoiceRepository) ReverseAllPayments(ctx context.Context, invoiceID string, reversalDate string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// For safety, ensure invoice exists
	var inv models.Invoice
	if err := ir.collection.FindOne(ctx, bson.M{"_id": invoiceID}).Decode(&inv); err != nil {
		return err
	}

	// Set PaidAmount to 0, clear PaidOn map, mark invoice as receivable (refund due)
	update := bson.M{
		"$set": bson.M{
			"paidAmount": 0,
			"paidOn":     map[string]float64{},
			"type":       models.InvoiceTypeReceivable,
			"updatedAt":  time.Now(),
		},
	}

	result, err := ir.collection.UpdateOne(ctx, bson.M{"_id": invoiceID}, update)
	if err != nil {
		return fmt.Errorf("failed to reverse payments: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

// ReversePaymentAmount deducts a specific amount from the invoice's paid total
// and records a negative entry in the PaidOn map for the provided date.
func (ir *InvoiceRepository) ReversePaymentAmount(ctx context.Context, invoiceID string, amount float64, dateStr string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Ensure invoice exists and get current paid amount
	var inv models.Invoice
	if err := ir.collection.FindOne(ctx, bson.M{"_id": invoiceID}).Decode(&inv); err != nil {
		return err
	}

	if amount <= 0 {
		return fmt.Errorf("amount to reverse must be > 0")
	}

	if amount > inv.PaidAmount {
		amount = inv.PaidAmount
	}

	update := bson.M{
		"$inc": bson.M{
			"paidAmount": -amount,
		},
		"$set": bson.M{
			fmt.Sprintf("paidOn.%s", dateStr): -amount,
			"updatedAt": time.Now(),
		},
	}

	// If after reversal paidAmount becomes zero, mark invoice receivable
	if inv.PaidAmount-amount <= 0 {
		update["$set"].(bson.M)["type"] = models.InvoiceTypeReceivable
	}

	result, err := ir.collection.UpdateOne(ctx, bson.M{"_id": invoiceID}, update)
	if err != nil {
		return fmt.Errorf("failed to reverse payment amount: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("invoice not found")
	}
	return nil
}

// GetInvoicesByType retrieves invoices by type with pagination
func (ir *InvoiceRepository) GetInvoicesByType(ctx context.Context, invoiceType string, page, limit int) ([]*models.Invoice, error) {
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

	cursor, err := ir.collection.Find(ctx, bson.M{"type": invoiceType}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoices: %w", err)
	}
	defer cursor.Close(ctx)

	var invoices []*models.Invoice
	if err := cursor.All(ctx, &invoices); err != nil {
		return nil, fmt.Errorf("failed to decode invoices: %w", err)
	}
	return invoices, nil
}

// GetInvoiceCount returns the total count of invoices
func (ir *InvoiceRepository) GetInvoiceCount(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	count, err := ir.collection.EstimatedDocumentCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count invoices: %w", err)
	}
	return count, nil
}
