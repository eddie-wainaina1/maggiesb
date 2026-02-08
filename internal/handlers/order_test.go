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

func TestCreateOrder_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	body := []byte(`{"invalid":"data"}`)
	httpReq := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	CreateOrder(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateOrder_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	userID := uuid.New().String()
	productID := uuid.New().String()
	
	product := &models.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: 100,
	}
	
	// Setup mocks
	mockProductRepo := new(MockProductRepository)
	mockProductRepo.On("GetProductByID", mock.Anything, productID).Return(product, nil)
	
	mockOrderRepo := new(MockOrderRepository)
	mockOrderRepo.On("CreateOrder", mock.Anything, mock.MatchedBy(func(o *models.Order) bool {
		return o.UserID == userID && len(o.Products) > 0
	})).Return(nil)
	
	mockInvoiceRepo := new(MockInvoiceRepository)
	mockInvoiceRepo.On("CreateInvoice", mock.Anything, mock.Anything).Return(nil)
	
	oldProductRepo := NewProductRepository
	oldOrderRepo := NewOrderRepository
	oldInvoiceRepo := NewInvoiceRepository
	NewProductRepository = ProductRepository(mockProductRepo)
	NewOrderRepository = OrderRepository(mockOrderRepo)
	NewInvoiceRepository = InvoiceRepository(mockInvoiceRepo)
	defer func() {
		NewProductRepository = oldProductRepo
		NewOrderRepository = oldOrderRepo
		NewInvoiceRepository = oldInvoiceRepo
	}()
	
	req := models.CreateOrderRequest{
		Phone: "254712345678",
		Products: []struct {
			ProductID string `json:"productId" binding:"required"`
			Quantity  int    `json:"quantity" binding:"required,gt=0"`
		}{
			{ProductID: productID, Quantity: 2},
		},
	}
	
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Set("userID", userID)
	
	CreateOrder(c)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	mockProductRepo.AssertExpectations(t)
	mockOrderRepo.AssertExpectations(t)
	mockInvoiceRepo.AssertExpectations(t)
}

func TestCreateOrder_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	body := []byte(`{"phone":"254712345678","products":[]}`)
	httpReq := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	CreateOrder(c)
	
	// Should fail on JSON binding validation before authentication check
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetOrder_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	httpReq := httptest.NewRequest("GET", "/orders/test-id", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: "test-id"}}
	
	GetOrder(c)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetOrder_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	userID := uuid.New().String()
	orderID := uuid.New().String()
	
	order := &models.Order{
		ID:     orderID,
		UserID: userID,
	}
	
	// Setup mock
	mockOrderRepo := new(MockOrderRepository)
	mockOrderRepo.On("GetOrderByID", mock.Anything, orderID).Return(order, nil)
	
	oldOrderRepo := NewOrderRepository
	NewOrderRepository = OrderRepository(mockOrderRepo)
	defer func() { NewOrderRepository = oldOrderRepo }()

	httpReq := httptest.NewRequest("GET", "/orders/test-id", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: orderID}}
	c.Set("userID", userID)
	
	GetOrder(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockOrderRepo.AssertExpectations(t)
}

func TestListOrders_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	httpReq := httptest.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	ListOrders(c)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListOrders_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	userID := uuid.New().String()
	
	orders := []*models.Order{
		{ID: uuid.New().String(), UserID: userID},
		{ID: uuid.New().String(), UserID: userID},
	}
	
	// Setup mock
	mockOrderRepo := new(MockOrderRepository)
	mockOrderRepo.On("GetOrdersByUser", mock.Anything, userID, 1, 10).Return(orders, nil)
	mockOrderRepo.On("GetOrderCountByUser", mock.Anything, userID).Return(int64(2), nil)
	
	oldOrderRepo := NewOrderRepository
	NewOrderRepository = OrderRepository(mockOrderRepo)
	defer func() { NewOrderRepository = oldOrderRepo }()

	httpReq := httptest.NewRequest("GET", "/orders?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Set("userID", userID)
	
	ListOrders(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(2), response["total"])
	
	mockOrderRepo.AssertExpectations(t)
}
