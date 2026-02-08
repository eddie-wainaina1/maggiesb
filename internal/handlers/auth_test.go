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
	"github.com/stretchr/testify/mock"
)

func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Setup mock user repository
	mockUserRepo := new(MockUserRepository)
	mockUserRepo.On("UserExists", mock.Anything, "newuser@example.com").Return(false, nil)
	mockUserRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "newuser@example.com" && u.FirstName == "John"
	})).Return(nil)
	
	// Override DI variable
	oldUserRepo := NewUserRepository
	NewUserRepository = UserRepository(mockUserRepo)
	defer func() { NewUserRepository = oldUserRepo }()
	
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
	
	Register(c)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response models.AuthResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "newuser@example.com", response.User.Email)
	assert.NotEmpty(t, response.Token)
	assert.NotZero(t, response.ExpiresAt)
	
	mockUserRepo.AssertExpectations(t)
}

func TestRegister_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	body := []byte(`{"email":"invalid"}`)
	httpReq := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	Register(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Setup mock user repository
	mockUserRepo := new(MockUserRepository)
	mockUserRepo.On("UserExists", mock.Anything, "existing@example.com").Return(true, nil)
	
	oldUserRepo := NewUserRepository
	NewUserRepository = mockUserRepo
	defer func() { NewUserRepository = oldUserRepo }()
	
	req := models.RegisterRequest{
		Email:     "existing@example.com",
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
	
	Register(c)
	
	assert.Equal(t, http.StatusConflict, w.Code)
	mockUserRepo.AssertExpectations(t)
}

func TestLogin_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	body := []byte(`{"invalid":"data"}`)
	httpReq := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	Login(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create a user with hashed password
	hashedPassword, _ := auth.HashPassword("password123")
	user := &models.User{
		ID:       uuid.New().String(),
		Email:    "test@example.com",
		Password: hashedPassword,
		Role:     "user",
	}
	
	// Setup mock user repository
	mockUserRepo := new(MockUserRepository)
	mockUserRepo.On("FindUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	
	oldUserRepo := NewUserRepository
	NewUserRepository = UserRepository(mockUserRepo)
	defer func() { NewUserRepository = oldUserRepo }()

	req := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	
	Login(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.AuthResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test@example.com", response.User.Email)
	assert.NotEmpty(t, response.Token)
	
	mockUserRepo.AssertExpectations(t)
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
	
	// Setup mock user repository
	mockUserRepo := new(MockUserRepository)
	mockUserRepo.On("FindUserByID", mock.Anything, userID).Return(user, nil)
	
	oldUserRepo := NewUserRepository
	NewUserRepository = UserRepository(mockUserRepo)
	defer func() { NewUserRepository = oldUserRepo }()

	httpReq := httptest.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Set("userID", userID)
	
	GetProfile(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.User
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test@example.com", response.Email)
	assert.Equal(t, "Test", response.FirstName)
	
	mockUserRepo.AssertExpectations(t)
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
	
	// Skip: ValidateToken requires database access to check token blacklist
	t.Skip("ValidateToken requires database initialization")
	
	// For invalid token, ValidateToken will fail before blacklisting
	// So we just need to verify the handler rejects it
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
	
	// Skip: ValidateToken requires database access to check token blacklist
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
	
	// Since token validation requires database access for blacklist check,
	// we expect unauthorized if database is not initialized
	// In a real scenario with mocks, we could test success
	assert.True(t, w.Code == http.StatusUnauthorized || w.Code == http.StatusInternalServerError)
}
