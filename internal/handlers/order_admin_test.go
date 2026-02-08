package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdminListOrders_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orders := []*models.Order{
		{ID: uuid.New().String(), UserID: uuid.New().String()},
		{ID: uuid.New().String(), UserID: uuid.New().String()},
	}

	// Setup mock
	mockOrderRepo := new(MockOrderRepository)
	mockOrderRepo.On("GetAllOrders", mock.Anything, 1, 10).Return(orders, nil)
	mockOrderRepo.On("GetOrderCount", mock.Anything).Return(int64(2), nil)

	oldOrderRepo := NewOrderRepository
	NewOrderRepository = OrderRepository(mockOrderRepo)
	defer func() { NewOrderRepository = oldOrderRepo }()

	httpReq := httptest.NewRequest("GET", "/admin/orders?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	AdminListOrders(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(2), response["total"])
	assert.NotNil(t, response["data"])

	mockOrderRepo.AssertExpectations(t)
}

func TestAdminListOrders_Failed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mock to return error
	mockOrderRepo := new(MockOrderRepository)
	mockOrderRepo.On("GetAllOrders", mock.Anything, 1, 10).Return(nil, assert.AnError)

	oldOrderRepo := NewOrderRepository
	NewOrderRepository = OrderRepository(mockOrderRepo)
	defer func() { NewOrderRepository = oldOrderRepo }()

	httpReq := httptest.NewRequest("GET", "/admin/orders?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	AdminListOrders(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockOrderRepo.AssertExpectations(t)
}

func TestAdminUpdateOrderStatus_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderID := uuid.New().String()
	req := struct {
		Status string `json:"status"`
	}{
		Status: models.OrderStatusComplete,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/admin/orders/"+orderID+"/status", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	// Setup mock
	mockOrderRepo := new(MockOrderRepository)
	mockOrderRepo.On("UpdateOrderStatus", mock.Anything, orderID, models.OrderStatusComplete).Return(nil)
	mockOrderRepo.On("GetOrderByID", mock.Anything, orderID).Return(&models.Order{
		ID:        orderID,
		UserID:    uuid.New().String(),
		Status:    models.OrderStatusComplete,
		TotalCost: 100.0,
	}, nil)

	oldOrderRepo := NewOrderRepository
	NewOrderRepository = OrderRepository(mockOrderRepo)
	defer func() { NewOrderRepository = oldOrderRepo }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: orderID}}

	AdminUpdateOrderStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Order
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, models.OrderStatusComplete, response.Status)

	mockOrderRepo.AssertExpectations(t)
}

func TestAdminUpdateOrderStatus_InvalidStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderID := uuid.New().String()
	req := struct {
		Status string `json:"status"`
	}{
		Status: "invalid_status",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/admin/orders/"+orderID+"/status", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: orderID}}

	AdminUpdateOrderStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminUpdateOrderStatus_MissingStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderID := uuid.New().String()
	body := []byte(`{}`)
	httpReq := httptest.NewRequest("PUT", "/admin/orders/"+orderID+"/status", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: orderID}}

	AdminUpdateOrderStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminUpdateOrderStatus_OrderNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderID := uuid.New().String()
	req := struct {
		Status string `json:"status"`
	}{
		Status: models.OrderStatusComplete,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/admin/orders/"+orderID+"/status", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	// Setup mock to return "order not found" error
	mockOrderRepo := new(MockOrderRepository)
	mockOrderRepo.On("UpdateOrderStatus", mock.Anything, orderID, models.OrderStatusComplete).Return(assert.AnError)

	oldOrderRepo := NewOrderRepository
	NewOrderRepository = OrderRepository(mockOrderRepo)
	defer func() { NewOrderRepository = oldOrderRepo }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: orderID}}

	AdminUpdateOrderStatus(c)

	// Should return 500 (internal server error) since our mock doesn't match "order not found" message
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockOrderRepo.AssertExpectations(t)
}

func TestAdminUpdateOrderStatus_CancelledWithReversal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderID := uuid.New().String()
	invoiceID := uuid.New().String()
	req := struct {
		Status string `json:"status"`
	}{
		Status: models.OrderStatusCancelled,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/admin/orders/"+orderID+"/status", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	// Setup mocks
	mockOrderRepo := new(MockOrderRepository)
	mockOrderRepo.On("UpdateOrderStatus", mock.Anything, orderID, models.OrderStatusCancelled).Return(nil)
	mockOrderRepo.On("GetOrderByID", mock.Anything, orderID).Return(&models.Order{
		ID:        orderID,
		UserID:    uuid.New().String(),
		Status:    models.OrderStatusCancelled,
		TotalCost: 100.0,
	}, nil)

	mockInvoiceRepo := new(MockInvoiceRepository)
	mockInvoiceRepo.On("GetInvoiceByOrderID", mock.Anything, orderID).Return(&models.Invoice{
		ID:         invoiceID,
		OrderID:    orderID,
		Type:       models.InvoiceTypePayable,
		PaidAmount: 50.0,
	}, nil)
	mockInvoiceRepo.On("ReverseAllPayments", mock.Anything, invoiceID, mock.Anything).Return(nil)

	mockPaymentRepo := new(MockPaymentRepository)
	mockPaymentRepo.On("ReversePaymentsByInvoiceID", mock.Anything, invoiceID).Return(nil)

	oldOrderRepo := NewOrderRepository
	oldInvoiceRepo := NewInvoiceRepository
	oldPaymentRepo := NewPaymentRepository
	NewOrderRepository = OrderRepository(mockOrderRepo)
	NewInvoiceRepository = InvoiceRepository(mockInvoiceRepo)
	NewPaymentRepository = PaymentRepository(mockPaymentRepo)
	defer func() {
		NewOrderRepository = oldOrderRepo
		NewInvoiceRepository = oldInvoiceRepo
		NewPaymentRepository = oldPaymentRepo
	}()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: orderID}}

	AdminUpdateOrderStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockOrderRepo.AssertExpectations(t)
	mockInvoiceRepo.AssertExpectations(t)
	mockPaymentRepo.AssertExpectations(t)
}
