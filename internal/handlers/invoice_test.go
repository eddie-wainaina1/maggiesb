package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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

// NOTE: To fully test GetInvoiceByOrder with mocks, the handlers would need to be refactored
// to accept repositories as function parameters or use interface-based DI instead of concrete types.
// For now, unit tests focus on validation logic that doesn't require database calls.

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
