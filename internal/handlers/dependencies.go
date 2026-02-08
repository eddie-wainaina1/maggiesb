package handlers

import (
	"context"

	"github.com/eddie-wainaina1/maggiesb/internal/database"
	"github.com/eddie-wainaina1/maggiesb/internal/models"
)

// Repository interfaces for dependency injection
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrderByID(ctx context.Context, orderID string) (*models.Order, error)
	GetOrdersByUser(ctx context.Context, userID string, page int, limit int) ([]*models.Order, error)
	GetAllOrders(ctx context.Context, page int, limit int) ([]*models.Order, error)
	GetOrderCountByUser(ctx context.Context, userID string) (int64, error)
	GetOrderCount(ctx context.Context) (int64, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status string) error
}

type InvoiceRepository interface {
	CreateInvoice(ctx context.Context, invoice *models.Invoice) error
	GetInvoiceByID(ctx context.Context, invoiceID string) (*models.Invoice, error)
	GetInvoiceByOrderID(ctx context.Context, orderID string) (*models.Invoice, error)
	RecordPayment(ctx context.Context, invoiceID string, amount float64, dateStr string) error
	ReverseAllPayments(ctx context.Context, invoiceID string, dateStr string) error
	ReversePaymentAmount(ctx context.Context, invoiceID string, amount float64, dateStr string) error
	GetInvoicesByType(ctx context.Context, invoiceType string, page int, limit int) ([]*models.Invoice, error)
	GetInvoiceCount(ctx context.Context) (int64, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	UserExists(ctx context.Context, email string) (bool, error)
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	FindUserByID(ctx context.Context, userID string) (*models.User, error)
	UpdateUser(ctx context.Context, userID string, user *models.User) error
	DeleteUser(ctx context.Context, userID string) error
}

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *models.Product) error
	GetProductByID(ctx context.Context, productID string) (*models.Product, error)
	GetAllProducts(ctx context.Context, page int, limit int) ([]*models.Product, error)
	SearchProducts(ctx context.Context, query string, page int, limit int) ([]*models.Product, error)
	UpdateProduct(ctx context.Context, productID string, updates map[string]interface{}) error
	DeleteProduct(ctx context.Context, productID string) error
	GetProductCount(ctx context.Context) (int64, error)
}

type PaymentRepository interface {
	CreatePaymentRecord(ctx context.Context, payment *models.PaymentRecord) error
	GetPaymentByCheckoutRequestID(ctx context.Context, checkoutRequestID string) (*models.PaymentRecord, error)
	GetPaymentByInvoiceID(ctx context.Context, invoiceID string) (*models.PaymentRecord, error)
	UpdatePaymentStatus(ctx context.Context, checkoutID string, status string, receiptNum string, transDate string) error
	ReversePaymentsByInvoiceID(ctx context.Context, invoiceID string) error
}

type ReversalRepository interface {
	CreateReversalRecord(ctx context.Context, record *models.ReversalRecord) error
}

type ReportRepository interface {
	GetSummaryReport(ctx context.Context, startDate, endDate string) (*models.SummaryReport, error)
	GetDailyBreakdown(ctx context.Context, startDate, endDate string) ([]models.DailySalesReport, error)
}

// DI variables - can be overridden in tests before handlers are called
var (
	NewOrderRepository    OrderRepository
	NewInvoiceRepository  InvoiceRepository
	NewUserRepository     UserRepository
	NewProductRepository  ProductRepository
	NewPaymentRepository  PaymentRepository
	NewReversalRepository ReversalRepository
	NewReportRepository   ReportRepository
)

// InitDependencies initializes all repositories (called from main)
func InitDependencies() {
	if NewOrderRepository == nil {
		NewOrderRepository = database.NewOrderRepository()
	}
	if NewInvoiceRepository == nil {
		NewInvoiceRepository = database.NewInvoiceRepository()
	}
	if NewUserRepository == nil {
		NewUserRepository = database.NewUserRepository()
	}
	if NewProductRepository == nil {
		NewProductRepository = database.NewProductRepository()
	}
	if NewPaymentRepository == nil {
		NewPaymentRepository = database.NewPaymentRepository()
	}
	if NewReversalRepository == nil {
		NewReversalRepository = database.NewReversalRepository()
	}
	if NewReportRepository == nil {
		NewReportRepository = database.NewReportRepository()
	}
}
