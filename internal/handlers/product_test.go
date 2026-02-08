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

func TestCreateProduct_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := models.CreateProductRequest{
		Name:        "Test Product",
		Description: "A test product",
		Price:       100.0,
		Discount:    10.0,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/products", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	// Setup mock
	mockProductRepo := new(MockProductRepository)
	mockProductRepo.On("CreateProduct", mock.Anything, mock.Anything).Return(nil)

	oldProductRepo := NewProductRepository
	NewProductRepository = ProductRepository(mockProductRepo)
	defer func() { NewProductRepository = oldProductRepo }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	CreateProduct(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Product
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Test Product", response.Name)
	assert.Equal(t, 100.0, response.Price)
	assert.Equal(t, 10.0, response.Discount)

	mockProductRepo.AssertExpectations(t)
}

func TestCreateProduct_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"invalid":"data"}`)
	httpReq := httptest.NewRequest("POST", "/products", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	CreateProduct(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetProduct_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	productID := uuid.New().String()
	product := &models.Product{
		ID:          productID,
		Name:        "Test Product",
		Description: "A test product",
		Price:       100.0,
		Discount:    10.0,
	}

	// Setup mock
	mockProductRepo := new(MockProductRepository)
	mockProductRepo.On("GetProductByID", mock.Anything, productID).Return(product, nil)

	oldProductRepo := NewProductRepository
	NewProductRepository = ProductRepository(mockProductRepo)
	defer func() { NewProductRepository = oldProductRepo }()

	httpReq := httptest.NewRequest("GET", "/products/"+productID, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: productID}}

	GetProduct(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Product
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Test Product", response.Name)
	assert.Equal(t, productID, response.ID)

	mockProductRepo.AssertExpectations(t)
}

func TestListProducts_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	products := []*models.Product{
		{ID: uuid.New().String(), Name: "Product 1", Price: 100.0},
		{ID: uuid.New().String(), Name: "Product 2", Price: 200.0},
	}

	// Setup mock
	mockProductRepo := new(MockProductRepository)
	mockProductRepo.On("GetAllProducts", mock.Anything, 1, 10).Return(products, nil)
	mockProductRepo.On("GetProductCount", mock.Anything).Return(int64(2), nil)

	oldProductRepo := NewProductRepository
	NewProductRepository = ProductRepository(mockProductRepo)
	defer func() { NewProductRepository = oldProductRepo }()

	httpReq := httptest.NewRequest("GET", "/products?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	ListProducts(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(2), response["total"])

	mockProductRepo.AssertExpectations(t)
}

func TestSearchProducts_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	products := []*models.Product{
		{ID: uuid.New().String(), Name: "Test Product", Price: 100.0},
	}

	// Setup mock
	mockProductRepo := new(MockProductRepository)
	mockProductRepo.On("SearchProducts", mock.Anything, "test", 1, 10).Return(products, nil)

	oldProductRepo := NewProductRepository
	NewProductRepository = ProductRepository(mockProductRepo)
	defer func() { NewProductRepository = oldProductRepo }()

	httpReq := httptest.NewRequest("GET", "/products/search?q=test&page=1&limit=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	SearchProducts(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["data"])

	mockProductRepo.AssertExpectations(t)
}

func TestSearchProducts_MissingQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	httpReq := httptest.NewRequest("GET", "/products/search?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	SearchProducts(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateProduct_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	productID := uuid.New().String()
	req := models.UpdateProductRequest{
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       150.0,
		Discount:    15.0,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/products/"+productID, bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	// Setup mock
	mockProductRepo := new(MockProductRepository)
	mockProductRepo.On("UpdateProduct", mock.Anything, productID, mock.Anything).Return(nil)
	mockProductRepo.On("GetProductByID", mock.Anything, productID).Return(&models.Product{
		ID:          productID,
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       150.0,
		Discount:    15.0,
	}, nil)

	oldProductRepo := NewProductRepository
	NewProductRepository = ProductRepository(mockProductRepo)
	defer func() { NewProductRepository = oldProductRepo }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: productID}}

	UpdateProduct(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockProductRepo.AssertExpectations(t)
}

func TestUpdateProduct_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	productID := uuid.New().String()
	body := []byte(`{"invalid":"data"}`)
	httpReq := httptest.NewRequest("PUT", "/products/"+productID, bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: productID}}

	UpdateProduct(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteProduct_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	productID := uuid.New().String()

	// Setup mock
	mockProductRepo := new(MockProductRepository)
	mockProductRepo.On("DeleteProduct", mock.Anything, productID).Return(nil)

	oldProductRepo := NewProductRepository
	NewProductRepository = ProductRepository(mockProductRepo)
	defer func() { NewProductRepository = oldProductRepo }()

	httpReq := httptest.NewRequest("DELETE", "/products/"+productID, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Params = gin.Params{gin.Param{Key: "id", Value: productID}}

	DeleteProduct(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockProductRepo.AssertExpectations(t)
}
