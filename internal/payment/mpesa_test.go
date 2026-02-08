package mpesa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestEncryptSecurityCredential ensures encryptSecurityCredential produces a
// base64-encoded RSA-encrypted blob that can be decrypted with the private key.
func TestEncryptSecurityCredential(t *testing.T) {
	// generate temporary RSA key pair
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("marshal pub failed: %v", err)
	}

	// write public key PEM to temp file
	dir := t.TempDir()
	pubPath := filepath.Join(dir, "test_pub.pem")
	f, err := os.Create(pubPath)
	if err != nil {
		t.Fatalf("create pub file: %v", err)
	}
	if err := pem.Encode(f, &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}); err != nil {
		t.Fatalf("pem encode failed: %v", err)
	}
	f.Close()

	c := NewClient(Config{PublicKeyPath: pubPath})

	secret := "super-secret-password"
	encB64, err := c.encryptSecurityCredential(secret)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	// decode and decrypt using private key
	raw, err := base64.StdEncoding.DecodeString(encB64)
	if err != nil {
		t.Fatalf("base64 decode failed: %v", err)
	}

	dec, err := rsa.DecryptPKCS1v15(rand.Reader, priv, raw)
	if err != nil {
		t.Fatalf("rsa decrypt failed: %v", err)
	}

	if string(dec) != secret {
		t.Fatalf("decrypted secret mismatch: expected %q got %q", secret, string(dec))
	}
}

// TestGeneratePassword ensures generatePassword returns expected base64 encoding
func TestGeneratePassword(t *testing.T) {
	short := "12345"
	pass := "passkey"
	ts := "20260101120000"

	got := generatePassword(short, pass, ts)
	expected := base64.StdEncoding.EncodeToString([]byte(short + pass + ts))
	if got != expected {
		t.Fatalf("generatePassword mismatch: expected %s got %s", expected, got)
	}
}

// TestBase64Encode tests the base64Encode helper
func TestBase64Encode(t *testing.T) {
	data := "hello world"
	encoded := base64Encode(data)
	expected := base64.StdEncoding.EncodeToString([]byte(data))
	assert.Equal(t, expected, encoded)
}

// TestNewClient creates a client and verifies configuration
func TestNewClient_Sandbox(t *testing.T) {
	config := Config{
		ConsumerKey:    "test_key",
		ConsumerSecret: "test_secret",
		Environment:    "sandbox",
	}
	c := NewClient(config)

	assert.Equal(t, SandboxBaseURL, c.baseURL)
	assert.Equal(t, config.ConsumerKey, c.config.ConsumerKey)
}

func TestNewClient_Production(t *testing.T) {
	config := Config{
		ConsumerKey:    "test_key",
		ConsumerSecret: "test_secret",
		Environment:    "production",
	}
	c := NewClient(config)

	assert.Equal(t, ProductionBaseURL, c.baseURL)
}

func TestNewClient_DefaultSandbox(t *testing.T) {
	config := Config{
		ConsumerKey:    "test_key",
		ConsumerSecret: "test_secret",
		Environment:    "",
	}
	c := NewClient(config)

	assert.Equal(t, SandboxBaseURL, c.baseURL)
}

// TestGetAccessToken_CacheHit tests that cached tokens are reused
func TestGetAccessToken_CacheHit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test_token_123",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	config := Config{
		ConsumerKey:    "test_key",
		ConsumerSecret: "test_secret",
		Environment:    "sandbox",
	}
	c := NewClient(config)
	c.baseURL = server.URL

	// First call - should fetch from server
	token1, err := c.GetAccessToken()
	assert.NoError(t, err)
	assert.Equal(t, "test_token_123", token1)

	// Second call - should use cache
	token2, err := c.GetAccessToken()
	assert.NoError(t, err)
	assert.Equal(t, "test_token_123", token2)
}

// TestGetAccessToken_Expiry tests that expired tokens are refreshed
func TestGetAccessToken_Expiry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "token_1",
				"expires_in":   1, // Expire in 1 second
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "token_2",
				"expires_in":   3600,
			})
		}
	}))
	defer server.Close()

	config := Config{
		ConsumerKey:    "test_key",
		ConsumerSecret: "test_secret",
		Environment:    "sandbox",
	}
	c := NewClient(config)
	c.baseURL = server.URL

	// First call
	token1, _ := c.GetAccessToken()
	assert.Equal(t, "token_1", token1)

	// Wait for expiry
	time.Sleep(1100 * time.Millisecond)

	// Second call - should fetch new token
	token2, _ := c.GetAccessToken()
	assert.Equal(t, "token_2", token2)
	assert.Equal(t, 2, callCount)
}

// TestGetAccessToken_NetworkError tests error handling
func TestGetAccessToken_NetworkError(t *testing.T) {
	config := Config{
		ConsumerKey:    "test_key",
		ConsumerSecret: "test_secret",
		Environment:    "sandbox",
	}
	c := NewClient(config)
	c.baseURL = "http://invalid-url-that-doesnt-exist"

	_, err := c.GetAccessToken()
	assert.Error(t, err)
}

// TestGetAccessToken_InvalidResponse tests JSON parse errors
func TestGetAccessToken_InvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"invalid": json`))
	}))
	defer server.Close()

	config := Config{
		ConsumerKey:    "test_key",
		ConsumerSecret: "test_secret",
		Environment:    "sandbox",
	}
	c := NewClient(config)
	c.baseURL = server.URL

	_, err := c.GetAccessToken()
	assert.Error(t, err)
}

// TestInitiateSTKPush_Success tests successful STK Push initiation
func TestInitiateSTKPush_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Token endpoint
		if r.RequestURI == "/oauth/v1/generate?grant_type=client_credentials" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test_token",
				"expires_in":   3600,
			})
			return
		}

		// STK Push endpoint
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(STKPushResponse{
			MerchantRequestID: "29115-34620561-1",
			CheckoutRequestID: "ws_CO_191220191020375136",
			ResponseCode:      "0",
			ResponseDescription: "Success. Request accepted for processing",
			CustomerMessage:   "Success. Request accepted for processing",
		})
	}))
	defer server.Close()

	config := Config{
		ConsumerKey:       "test_key",
		ConsumerSecret:    "test_secret",
		BusinessShortCode: "174379",
		PassKey:           "bfb279f9aa9bdbcf158e97dd1a503b30",
		CallbackURL:       "https://example.com/callback",
		Environment:       "sandbox",
	}
	c := NewClient(config)
	c.baseURL = server.URL

	resp, err := c.InitiateSTKPush("254712345678", "100", "invoice-123")
	assert.NoError(t, err)
	assert.Equal(t, "0", resp.ResponseCode)
	assert.Equal(t, "ws_CO_191220191020375136", resp.CheckoutRequestID)
}

// TestInitiateSTKPush_NoToken tests failure when token cannot be obtained
func TestInitiateSTKPush_NoToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	config := Config{
		ConsumerKey:       "invalid_key",
		ConsumerSecret:    "invalid_secret",
		BusinessShortCode: "174379",
		PassKey:           "bfb279f9aa9bdbcf158e97dd1a503b30",
		CallbackURL:       "https://example.com/callback",
		Environment:       "sandbox",
	}
	c := NewClient(config)
	c.baseURL = server.URL

	_, err := c.InitiateSTKPush("254712345678", "100", "invoice-123")
	assert.Error(t, err)
}

// TestInitiateSTKPush_ErrorResponse tests handling of API errors
func TestInitiateSTKPush_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/oauth/v1/generate?grant_type=client_credentials" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test_token",
				"expires_in":   3600,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(STKPushResponse{
			ResponseCode:        "1",
			ResponseDescription: "Some error occurred",
		})
	}))
	defer server.Close()

	config := Config{
		ConsumerKey:       "test_key",
		ConsumerSecret:    "test_secret",
		BusinessShortCode: "174379",
		PassKey:           "bfb279f9aa9bdbcf158e97dd1a503b30",
		CallbackURL:       "https://example.com/callback",
		Environment:       "sandbox",
	}
	c := NewClient(config)
	c.baseURL = server.URL

	_, err := c.InitiateSTKPush("254712345678", "100", "invoice-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "STK push failed")
}

// TestInitiateReversal_NotConfigured tests error when reversal not configured
func TestInitiateReversal_NotConfigured(t *testing.T) {
	config := Config{
		Environment: "sandbox",
	}
	c := NewClient(config)

	err := c.InitiateReversal("254712345678", "100", "invoice-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

// TestInitiateReversal_InvalidPublicKey tests error with invalid public key
func TestInitiateReversal_InvalidPublicKey(t *testing.T) {
	dir := t.TempDir()
	pubPath := filepath.Join(dir, "invalid.pem")
	os.WriteFile(pubPath, []byte("invalid pem"), 0644)

	config := Config{
		InitiatorName:     "testuser",
		InitiatorPassword: "testpass",
		PublicKeyPath:     pubPath,
		Environment:       "sandbox",
	}
	c := NewClient(config)

	err := c.InitiateReversal("254712345678", "100", "invoice-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SecurityCredential")
}

// TestInitiateReversal_Success tests successful reversal initiation
func TestInitiateReversal_Success(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubBytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)

	dir := t.TempDir()
	pubPath := filepath.Join(dir, "test_pub.pem")
	f, _ := os.Create(pubPath)
	pem.Encode(f, &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	f.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/oauth/v1/generate?grant_type=client_credentials" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test_token",
				"expires_in":   3600,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ConversationID":      "AG_20240101_1234567890abcdef",
			"OriginatorConversationID": "12345-1234567-1",
			"ResponseCode":        "0",
			"ResponseDescription": "Accept the service request successfully.",
		})
	}))
	defer server.Close()

	config := Config{
		InitiatorName:      "testuser",
		InitiatorPassword:  "testpass",
		PublicKeyPath:      pubPath,
		BusinessShortCode:  "174379",
		ReversalResultURL:  "https://example.com/reversal/result",
		ReversalTimeoutURL: "https://example.com/reversal/timeout",
		Environment:        "sandbox",
	}
	c := NewClient(config)
	c.baseURL = server.URL

	err := c.InitiateReversal("254712345678", "100", "invoice-123")
	assert.NoError(t, err)
}

// TestInitiateReversal_ErrorResponse tests handling of reversal errors
func TestInitiateReversal_ErrorResponse(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubBytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)

	dir := t.TempDir()
	pubPath := filepath.Join(dir, "test_pub.pem")
	f, _ := os.Create(pubPath)
	pem.Encode(f, &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	f.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/oauth/v1/generate?grant_type=client_credentials" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test_token",
				"expires_in":   3600,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessage": "Invalid request",
		})
	}))
	defer server.Close()

	config := Config{
		InitiatorName:      "testuser",
		InitiatorPassword:  "testpass",
		PublicKeyPath:      pubPath,
		BusinessShortCode:  "174379",
		ReversalResultURL:  "https://example.com/reversal/result",
		ReversalTimeoutURL: "https://example.com/reversal/timeout",
		Environment:        "sandbox",
	}
	c := NewClient(config)
	c.baseURL = server.URL

	err := c.InitiateReversal("254712345678", "100", "invoice-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mpesa reversal error")
}

// TestEncryptSecurityCredential_InvalidKeyPath tests error handling
func TestEncryptSecurityCredential_InvalidKeyPath(t *testing.T) {
	config := Config{
		PublicKeyPath: "/nonexistent/path/key.pem",
	}
	c := NewClient(config)

	_, err := c.encryptSecurityCredential("secret")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read public key")
}

// TestGeneratePassword_DifferentInputs tests password generation variations
func TestGeneratePassword_DifferentInputs(t *testing.T) {
	tests := []struct {
		shortCode string
		passKey   string
		timestamp string
	}{
		{"123", "key", "20240101"},
		{"000", "", ""},
		{"999", "complex-key-123", "20250131235959"},
	}

	for _, tt := range tests {
		got := generatePassword(tt.shortCode, tt.passKey, tt.timestamp)
		expected := base64.StdEncoding.EncodeToString([]byte(tt.shortCode + tt.passKey + tt.timestamp))
		assert.Equal(t, expected, got)
	}
}
