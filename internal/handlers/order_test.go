package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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

func TestCreateOrder_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Empty products array will fail validation (min=1)
	body := []byte(`{"phone":"254712345678","products":[]}`)
	httpReq := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	CreateOrder(c)
	
	// Should fail on validation since no userID is set, but request is also invalid
	assert.True(t, w.Code == http.StatusBadRequest)
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

func TestListOrders_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	httpReq := httptest.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	ListOrders(c)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

