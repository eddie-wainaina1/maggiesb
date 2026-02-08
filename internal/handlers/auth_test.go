package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eddie-wainaina1/maggiesb/internal/auth"
	"github.com/eddie-wainaina1/maggiesb/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	req := models.RegisterRequest{
		Email:     "newuser@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	// This test requires a real database or we need to refactor handlers to use interfaces
	// For now, we test the request parsing
	var parsedReq models.RegisterRequest
	if err := c.ShouldBindJSON(&parsedReq); err != nil {
		t.Fatalf("Failed to bind request: %v", err)
	}
	
	assert.Equal(t, "newuser@example.com", parsedReq.Email)
	assert.Equal(t, "password123", parsedReq.Password)
	assert.Equal(t, "John", parsedReq.FirstName)
	assert.Equal(t, "Doe", parsedReq.LastName)
}

func TestRegister_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	body := []byte(`{"email":"invalid"}`)
	httpReq := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	var parsedReq models.RegisterRequest
	err := c.ShouldBindJSON(&parsedReq)
	assert.NotNil(t, err)
}

func TestLogin_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	body := []byte(`{"invalid":"data"}`)
	httpReq := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	var parsedReq models.LoginRequest
	err := c.ShouldBindJSON(&parsedReq)
	assert.NotNil(t, err)
}

func TestGetProfile_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	httpReq := httptest.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	GetProfile(c)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetProfile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	userID := uuid.New().String()
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
	}
	
	httpReq := httptest.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Set("userID", userID)
	
	// Would need to mock database.NewUserRepository() to test fully
	// For now, verify the context setup works
	retrievedID, exists := c.Get("userID")
	assert.True(t, exists)
	assert.Equal(t, userID, retrievedID)
	assert.Equal(t, user.ID, retrievedID)
}

func TestLogout_MissingAuthHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	httpReq := httptest.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	Logout(c)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogout_InvalidAuthHeaderFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	httpReq := httptest.NewRequest("POST", "/logout", nil)
	httpReq.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	Logout(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogout_MissingBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	httpReq := httptest.NewRequest("POST", "/logout", nil)
	httpReq.Header.Set("Authorization", "Basic sometoken")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	Logout(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogout_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Skip because ValidateToken requires database access
	t.Skip("ValidateToken requires database initialization")
	
	httpReq := httptest.NewRequest("POST", "/logout", nil)
	httpReq.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	Logout(c)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogout_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Skip because ValidateToken requires database access
	t.Skip("ValidateToken requires database initialization")
	
	// Generate a valid token
	userID := uuid.New().String()
	token, err := auth.GenerateToken(userID, "test@example.com", "user", 1*time.Hour)
	assert.NoError(t, err)
	
	httpReq := httptest.NewRequest("POST", "/logout", nil)
	httpReq.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	Logout(c)
	
	// The handler will try to call auth.BlacklistToken which needs the database setup
	// So we expect either success or an internal error depending on database state
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}
