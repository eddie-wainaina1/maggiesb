package handlers

import (
	"context"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockOrderRepository mocks the order repository
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderRepository) GetOrdersByUserID(ctx context.Context, userID string, page int, limit int) ([]*models.Order, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockOrderRepository) GetOrdersByUser(ctx context.Context, userID string, page int, limit int) ([]*models.Order, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockOrderRepository) GetAllOrders(ctx context.Context, page int, limit int) ([]*models.Order, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockOrderRepository) GetOrderCountByUser(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) GetOrderCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	args := m.Called(ctx, orderID, status)
	return args.Error(0)
}

// MockInvoiceRepository mocks the invoice repository
type MockInvoiceRepository struct {
	mock.Mock
}

func (m *MockInvoiceRepository) CreateInvoice(ctx context.Context, invoice *models.Invoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *MockInvoiceRepository) GetInvoiceByID(ctx context.Context, invoiceID string) (*models.Invoice, error) {
	args := m.Called(ctx, invoiceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) GetInvoiceByOrderID(ctx context.Context, orderID string) (*models.Invoice, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) RecordPayment(ctx context.Context, invoiceID string, amount float64, dateStr string) error {
	args := m.Called(ctx, invoiceID, amount, dateStr)
	return args.Error(0)
}

func (m *MockInvoiceRepository) ReverseAllPayments(ctx context.Context, invoiceID string, dateStr string) error {
	args := m.Called(ctx, invoiceID, dateStr)
	return args.Error(0)
}

func (m *MockInvoiceRepository) ReversePaymentAmount(ctx context.Context, invoiceID string, amount float64, dateStr string) error {
	args := m.Called(ctx, invoiceID, amount, dateStr)
	return args.Error(0)
}

func (m *MockInvoiceRepository) GetInvoicesByType(ctx context.Context, invoiceType string, page int, limit int) ([]*models.Invoice, error) {
	args := m.Called(ctx, invoiceType, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) GetInvoiceCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return int64(args.Int(0)), args.Error(1)
}

// MockUserRepository mocks the user repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindUserByID(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, userID string, user *models.User) error {
	args := m.Called(ctx, userID, user)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) UserExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// MockProductRepository mocks the product repository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) GetProductByID(ctx context.Context, productID string) (*models.Product, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) GetAllProducts(ctx context.Context, page int, limit int) ([]*models.Product, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) SearchProducts(ctx context.Context, query string, page int, limit int) ([]*models.Product, error) {
	args := m.Called(ctx, query, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) UpdateProduct(ctx context.Context, productID string, updates map[string]interface{}) error {
	args := m.Called(ctx, productID, updates)
	return args.Error(0)
}

func (m *MockProductRepository) DeleteProduct(ctx context.Context, productID string) error {
	args := m.Called(ctx, productID)
	return args.Error(0)
}

func (m *MockProductRepository) GetProductCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// MockPaymentRepository mocks the payment repository
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) CreatePaymentRecord(ctx context.Context, payment *models.PaymentRecord) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetPaymentByCheckoutRequestID(ctx context.Context, checkoutID string) (*models.PaymentRecord, error) {
	args := m.Called(ctx, checkoutID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaymentRecord), args.Error(1)
}

func (m *MockPaymentRepository) GetPaymentByInvoiceID(ctx context.Context, invoiceID string) (*models.PaymentRecord, error) {
	args := m.Called(ctx, invoiceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaymentRecord), args.Error(1)
}

func (m *MockPaymentRepository) UpdatePaymentStatus(ctx context.Context, checkoutID string, status string, receiptNum string, transDate string) error {
	args := m.Called(ctx, checkoutID, status, receiptNum, transDate)
	return args.Error(0)
}

func (m *MockPaymentRepository) ReversePaymentsByInvoiceID(ctx context.Context, invoiceID string) error {
	args := m.Called(ctx, invoiceID)
	return args.Error(0)
}

// MockReversalRepository mocks the reversal repository
type MockReversalRepository struct {
	mock.Mock
}

func (m *MockReversalRepository) CreateReversalRecord(ctx context.Context, reversal *models.ReversalRecord) error {
	args := m.Called(ctx, reversal)
	return args.Error(0)
}
// MockReportRepository mocks the report repository
type MockReportRepository struct {
	mock.Mock
}

func (m *MockReportRepository) GetSummaryReport(ctx context.Context, startDate, endDate string) (*models.SummaryReport, error) {
	args := m.Called(ctx, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SummaryReport), args.Error(1)
}

func (m *MockReportRepository) GetDailyBreakdown(ctx context.Context, startDate, endDate string) ([]models.DailySalesReport, error) {
	args := m.Called(ctx, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DailySalesReport), args.Error(1)
}