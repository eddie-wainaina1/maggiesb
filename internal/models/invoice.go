package models

import "time"

// Invoice represents a financial document for an order
type Invoice struct {
	ID            string             `json:"id" bson:"_id"`
	OrderID       string             `json:"orderId" bson:"orderId"`
	InvoiceAmount float64            `json:"invoiceAmount" bson:"invoiceAmount"` // total invoice amount
	PaidAmount    float64            `json:"paidAmount" bson:"paidAmount"`       // total amount paid so far
	TaxAmount     float64            `json:"taxAmount" bson:"taxAmount"`         // tax applied (default 0)
	Type          string             `json:"type" bson:"type"`                   // "payable" or "receivable"
	PaidOn        map[string]float64 `json:"paidOn" bson:"paidOn"`               // map of dates (YYYY-MM-DD) to amounts paid
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// Invoice types
const (
	InvoiceTypePayable   = "payable"   // customer needs to pay
	InvoiceTypeReceivable = "receivable" // customer paid/refund due
)

// CreateInvoiceRequest payload to create an invoice
type CreateInvoiceRequest struct {
	OrderID       string  `json:"orderId" binding:"required"`
	InvoiceAmount float64 `json:"invoiceAmount" binding:"required,gt=0"`
	TaxAmount     float64 `json:"taxAmount" binding:"min=0"`
	Type          string  `json:"type" binding:"required"`
}

// RecordPaymentRequest payload to record a payment on an invoice
type RecordPaymentRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
	Date   string  `json:"date" binding:"required"` // YYYY-MM-DD format
}

// ReverseInvoiceRequest is used by admin to request a manual reversal/refund
type ReverseInvoiceRequest struct {
	// If Amount is zero, reverse entire paid amount
	Amount   float64 `json:"amount" binding:"gte=0"`
	Date     string  `json:"date"`                             // optional YYYY-MM-DD
	Phone    string  `json:"phone"`                            // customer phone for M-Pesa reversal (optional)
	UseMpesa bool    `json:"useMpesa"`                         // attempt M-Pesa reversal when true
	Reason   string  `json:"reason" binding:"required"`       // reason for reversal
}

// ReversalRecord represents an audit record for a refund/reversal
type ReversalRecord struct {
	ID        string    `json:"id" bson:"_id"`
	InvoiceID string    `json:"invoiceId" bson:"invoiceId"`
	Amount    float64   `json:"amount" bson:"amount"`
	Date      string    `json:"date" bson:"date"` // YYYY-MM-DD
	Phone     string    `json:"phone" bson:"phone"`
	AdminID   string    `json:"adminId" bson:"adminId"`
	Reason    string    `json:"reason" bson:"reason"`
	ReceiptURL string   `json:"receiptUrl,omitempty" bson:"receiptUrl,omitempty"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}
