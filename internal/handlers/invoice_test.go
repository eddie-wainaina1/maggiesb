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

func TestGetInvoiceByOrder_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	httpReq := httptest.NewRequest("GET", "/orders/test-id/invoice", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: "test-id"}}
	
	GetInvoiceByOrder(c)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetInvoiceByOrder_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	userID := uuid.New().String()
	orderID := uuid.New().String()
	invoiceID := uuid.New().String()
	
	order := &models.Order{
		ID:     orderID,
		UserID: userID,
	}
	invoice := &models.Invoice{
		ID:      invoiceID,
		OrderID: orderID,
	}
	
	// Setup mocks
	mockOrderRepo := new(MockOrderRepository)
	mockOrderRepo.On("GetOrderByID", mock.Anything, orderID).Return(order, nil)
	
	mockInvoiceRepo := new(MockInvoiceRepository)
	mockInvoiceRepo.On("GetInvoiceByOrderID", mock.Anything, orderID).Return(invoice, nil)
	
	oldOrderRepo := NewOrderRepository
	oldInvoiceRepo := NewInvoiceRepository
	NewOrderRepository = OrderRepository(mockOrderRepo)
	NewInvoiceRepository = InvoiceRepository(mockInvoiceRepo)
	defer func() {
		NewOrderRepository = oldOrderRepo
		NewInvoiceRepository = oldInvoiceRepo
	}()
	
	httpReq := httptest.NewRequest("GET", "/orders/test-id/invoice", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: orderID}}
	c.Set("userID", userID)
	
	GetInvoiceByOrder(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.Invoice
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, invoiceID, response.ID)
	
	mockOrderRepo.AssertExpectations(t)
	mockInvoiceRepo.AssertExpectations(t)
}

func TestGetInvoice_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	httpReq := httptest.NewRequest("GET", "/invoices/test-id", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: "test-id"}}
	
	GetInvoice(c)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetInvoice_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	userID := uuid.New().String()
	invoiceID := uuid.New().String()
	orderID := uuid.New().String()
	
	invoice := &models.Invoice{
		ID:      invoiceID,
		OrderID: orderID,
	}
	order := &models.Order{
		ID:     orderID,
		UserID: userID,
	}
	
	mockInvoiceRepo := new(MockInvoiceRepository)
	mockInvoiceRepo.On("GetInvoiceByID", mock.Anything, invoiceID).Return(invoice, nil)
	
	mockOrderRepo := new(MockOrderRepository)
	mockOrderRepo.On("GetOrderByID", mock.Anything, orderID).Return(order, nil)
	
	oldInvoiceRepo := NewInvoiceRepository
	oldOrderRepo := NewOrderRepository
	NewInvoiceRepository = InvoiceRepository(mockInvoiceRepo)
	NewOrderRepository = OrderRepository(mockOrderRepo)
	defer func() {
		NewInvoiceRepository = oldInvoiceRepo
		NewOrderRepository = oldOrderRepo
	}()
	
	httpReq := httptest.NewRequest("GET", "/invoices/test-id", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: invoiceID}}
	c.Set("userID", userID)
	
	GetInvoice(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockInvoiceRepo.AssertExpectations(t)
	mockOrderRepo.AssertExpectations(t)
}

func TestAdminRecordPayment_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	body := []byte(`{"invalid":"data"}`)
	httpReq := httptest.NewRequest("POST", "/admin/invoices/test-id/payment", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: "test-id"}}
	
	AdminRecordPayment(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminRecordPayment_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	invoiceID := uuid.New().String()
	
	updatedInvoice := &models.Invoice{
		ID:            invoiceID,
		InvoiceAmount: 1000,
		PaidAmount:    500,
	}
	
	// Setup mock
	mockInvoiceRepo := new(MockInvoiceRepository)
	mockInvoiceRepo.On("RecordPayment", mock.Anything, invoiceID, 500.0, "2024-01-15").Return(nil)
	mockInvoiceRepo.On("GetInvoiceByID", mock.Anything, invoiceID).Return(updatedInvoice, nil)
	
	oldInvoiceRepo := NewInvoiceRepository
	NewInvoiceRepository = InvoiceRepository(mockInvoiceRepo)
	defer func() { NewInvoiceRepository = oldInvoiceRepo }()
	
	req := models.RecordPaymentRequest{
		Amount: 500,
		Date:   "2024-01-15",
	}
	
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/admin/invoices/test-id/payment", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: invoiceID}}
	
	AdminRecordPayment(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockInvoiceRepo.AssertExpectations(t)
}

func TestAdminRecordPayment_MissingAmount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	req := models.RecordPaymentRequest{
		Date: "2024-01-15",
		// Missing Amount
	}
	
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/admin/invoices/test-id/payment", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: "test-id"}}
	
	AdminRecordPayment(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminReverseInvoice_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	body := []byte(`{"invalid":"data"}`)
	httpReq := httptest.NewRequest("PUT", "/admin/invoices/test-id/reverse", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: "test-id"}}
	
	AdminReverseInvoice(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminReverseInvoice_MissingReason(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	req := models.ReverseInvoiceRequest{
		Phone:    "254712345678",
		UseMpesa: false,
		Amount:   0,
		// Missing Reason
	}
	
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/admin/invoices/test-id/reverse", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: "test-id"}}
	
	AdminReverseInvoice(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
