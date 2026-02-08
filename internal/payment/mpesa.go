package mpesa

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	SandboxBaseURL    = "https://sandbox.safaricom.co.ke"
	ProductionBaseURL = "https://api.safaricom.co.ke"
)

// Config holds M-Pesa API credentials
type Config struct {
	ConsumerKey       string
	ConsumerSecret    string
	BusinessShortCode string
	PassKey           string
	CallbackURL       string
	Environment       string // "sandbox" or "production"
}

// Client handles M-Pesa API interactions
type Client struct {
	config       Config
	baseURL      string
	accessToken  string
	tokenExpiry  time.Time
	tokenMutex   sync.RWMutex
	httpClient   *http.Client
}

// NewClient creates a new M-Pesa client
func NewClient(config Config) *Client {
	baseURL := SandboxBaseURL
	if config.Environment == "production" {
		baseURL = ProductionBaseURL
	}

	return &Client{
		config:     config,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetAccessToken retrieves a valid OAuth access token
func (c *Client) GetAccessToken() (string, error) {
	c.tokenMutex.RLock()
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		defer c.tokenMutex.RUnlock()
		return c.accessToken, nil
	}
	c.tokenMutex.RUnlock()

	url := fmt.Sprintf("%s/oauth/v1/generate?grant_type=client_credentials", c.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.config.ConsumerKey, c.config.ConsumerSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	c.tokenMutex.Lock()
	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	c.tokenMutex.Unlock()

	return tokenResp.AccessToken, nil
}

// STKPushRequest payload for initiating STK Push
type STKPushRequest struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	Password          string `json:"Password"`
	Timestamp         string `json:"Timestamp"`
	TransactionType   string `json:"TransactionType"`
	Amount            string `json:"Amount"`
	PartyA            string `json:"PartyA"`
	PartyB            string `json:"PartyB"`
	PhoneNumber       string `json:"PhoneNumber"`
	CallBackURL       string `json:"CallBackURL"`
	AccountReference  string `json:"AccountReference"`
	TransactionDesc   string `json:"TransactionDesc"`
}

// STKPushResponse from Safaricom
type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
}

// InitiateSTKPush initiates an M-Pesa STK Push payment
func (c *Client) InitiateSTKPush(phone, amount, invoiceID string) (*STKPushResponse, error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")
	password := generatePassword(c.config.BusinessShortCode, c.config.PassKey, timestamp)

	payload := STKPushRequest{
		BusinessShortCode: c.config.BusinessShortCode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   "CustomerPayBillOnline",
		Amount:            amount,
		PartyA:            phone,
		PartyB:            c.config.BusinessShortCode,
		PhoneNumber:       phone,
		CallBackURL:       c.config.CallbackURL,
		AccountReference:  invoiceID,
		TransactionDesc:   "Order Payment",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/mpesa/stkpush/v1/processrequest", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send STK push request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var stkResp STKPushResponse
	if err := json.Unmarshal(body, &stkResp); err != nil {
		return nil, fmt.Errorf("failed to parse STK push response: %w", err)
	}

	if stkResp.ResponseCode != "0" {
		return nil, fmt.Errorf("STK push failed: %s - %s", stkResp.ResponseCode, stkResp.ResponseDescription)
	}

	return &stkResp, nil
}

// generatePassword generates M-Pesa password
func generatePassword(shortCode, passKey, timestamp string) string {
	data := shortCode + passKey + timestamp
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// base64Encode encodes a string to base64
func base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}
