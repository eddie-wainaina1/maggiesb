package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eddie-wainaina1/maggiesb/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_MissingAuthHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_InvalidAuthHeaderFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_MissingBearerPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic sometoken")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Skip this test as ValidateToken requires database access
	t.Skip("ValidateToken requires database initialization")

	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_StoresClaimsStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Skip test that requires database access via ValidateToken
	t.Skip("ValidateToken requires database initialization")
	
	userID := uuid.New().String()
	email := "test@example.com"
	role := "admin"
	token, err := auth.GenerateToken(userID, email, role, 1*time.Hour)
	assert.NoError(t, err)

	router := gin.New()

	var storedUserID, storedEmail, storedRole string
	
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		storedUserID = c.GetString("userID")
		storedEmail = c.GetString("email")
		storedRole = c.GetString("role")
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, userID, storedUserID)
	assert.Equal(t, email, storedEmail)
	assert.Equal(t, role, storedRole)
}

func TestRequireRole_Forbidden_InvalidRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// middleware expects role to be set in context
	router.Use(func(c *gin.Context) {
		c.Set("role", "user")
		c.Next()
	})

	router.GET("/admin", RequireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireRole_Forbidden_MissingRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/admin", RequireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireRole_Allowed_ValidRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})

	router.GET("/admin", RequireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireRole_MultipleRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("role", "moderator")
		c.Next()
	})

	router.GET("/user", RequireRole("user"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/user", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}
