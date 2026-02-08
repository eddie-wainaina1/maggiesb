# M-Pesa Integration Guide

This guide walks through integrating Safaricom Daraja M-Pesa Express (STK Push) into the Maggiesb ecommerce application.

## Overview

M-Pesa Express (also known as "Lipa na M-Pesa Online") is an API that lets you initiate mobile money payment prompts on customer phones. When a customer initiates a payment for an order:

1. Your server calls the Safaricom Daraja API with the customer's phone number
2. Safaricom sends an STK Push prompt to the customer's M-Pesa app
3. Customer enters their PIN to confirm
4. Safaricom sends a callback to your server with the result
5. Your server records the payment and updates the invoice

## Setup Steps

### 1. Register with Safaricom Developer Portal

1. Go to [https://developer.safaricom.co.ke](https://developer.safaricom.co.ke)
2. Create a developer account
3. Create a new app with the following:
    - **App name**: "Maggiesb" (or your app name)
    - **Description**: E-commerce platform payment integration
4. Generate credentials (Consumer Key and Consumer Secret)
5. In the app settings, add your callback URL:
    - Sandbox: `https://yourdomain-staging.com/api/v1/mpesa/callback`
    - Production: `https://yourdomain.com/api/v1/mpesa/callback`

### 2. Get Your Shortcode and Passkey

- **Shortcode**: Usually provided after app creation (test: 174379)
- **Passkey**: A unique key for your shortcode (get from your account dashboard)

### 3. Configure Environment Variables

Create a `.env` file with:

```env
# M-Pesa Configuration
MPESA_CONSUMER_KEY=your_key_from_daraja
MPESA_CONSUMER_SECRET=your_secret_from_daraja
MPESA_BUSINESS_SHORTCODE=174379
MPESA_PASSKEY=your_passkey_here
MPESA_CALLBACK_URL=https://yourdomain.com/api/v1/mpesa/callback
MPESA_ENV=sandbox  # or "production" for live
```

### 4. Start Your Application

```bash
go run main.go
```

The app will:

- Load environment variables from `.env`
- Initialize the M-Pesa client (or warn if credentials are missing)
- Start the HTTP server with payment endpoints

## API Usage

### Initiate Payment

**Endpoint**: `POST /api/v1/orders/:id/pay`

**Authentication**: Required (Bearer token)

**Request Body**:

```json
{
    "invoiceId": "invoice-uuid-here",
    "phone": "254712345678"
}
```

**Response (201)**:

```json
{
    "checkoutRequestId": "ws_CO_191220191020375651",
    "customerMessage": "Please enter your M-Pesa PIN",
    "paymentId": "payment-record-uuid"
}
```

**Error Cases**:

- `400`: Invalid request (missing fields)
- `401`: Not authenticated
- `404`: Invoice not found
- `403`: Invoice belongs to another user
- `500`: M-Pesa API error or payment record creation failed

### Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│ Client (Mobile/Web)                                         │
└─────────────────────────────────────────────────────────────┘
                            │
                POST /api/v1/orders/:id/pay
                (invoiceId, phone)
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Your Server (Maggiesb)                                      │
│ - Verify invoice ownership                                  │
│ - Call M-Pesa STK Push API                                  │
│ - Create PaymentRecord (status: "initiated")                │
└─────────────────────────────────────────────────────────────┘
                            │
            Send STK Push Request (OAuth token + payload)
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Safaricom Daraja API                                        │
│ - Validate credentials                                      │
│ - Send STK Push to customer                                 │
└─────────────────────────────────────────────────────────────┘
                            │
                STK Push on Customer Phone
                            │
                            ▼
        Customer Enters PIN → Payment Processed
                            │
                            ▼
        Safaricom Sends Callback with Result
                            │
                POST /api/v1/mpesa/callback
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Your Server (Callback Handler)                              │
│ - Parse callback                                            │
│ - Update PaymentRecord (status: "completed")                │
│ - Update Invoice (add to PaidOn map)                        │
│ - Return 200 OK                                             │
└─────────────────────────────────────────────────────────────┘
```

## Code Structure

### Models (internal/models/payment.go)

```go
type MpesaPaymentRequest struct {
    InvoiceID string `json:"invoiceId" binding:"required"`
    Phone     string `json:"phone" binding:"required"`
}

type PaymentRecord struct {
    ID                string             `bson:"_id"`
    InvoiceID         string             `bson:"invoiceId"`
    OrderID           string             `bson:"orderId"`
    CheckoutRequestID string             `bson:"checkoutRequestId"`
    Phone             string             `bson:"phone"`
    Amount            float64            `bson:"amount"`
    MpesaReceiptNumber string             `bson:"mpesaReceiptNumber"`
    Status            string             `bson:"status"` // "initiated", "completed", "failed"
    CreatedAt         time.Time          `bson:"createdAt"`
    UpdatedAt         time.Time          `bson:"updatedAt"`
}

type MpesaCallback struct {
    Body struct {
        StkCallback struct {
            MerchantRequestID string `json:"MerchantRequestID"`
            CheckoutRequestID string `json:"CheckoutRequestID"`
            ResultCode        int    `json:"ResultCode"` // 0 = success
            ResultDesc        string `json:"ResultDesc"`
            CallbackMetadata  struct {
                Item []struct {
                    Name  string      `json:"Name"`
                    Value interface{} `json:"Value"`
                } `json:"Item"`
            } `json:"CallbackMetadata"`
        } `json:"stkCallback"`
    } `json:"Body"`
}
```

### M-Pesa Client (internal/payment/mpesa.go)

```go
type Client struct {
    config      Config
    accessToken string
    tokenExpiry time.Time
    tokenMutex  sync.RWMutex // Protects token access
    httpClient  *http.Client
}

func (c *Client) GetAccessToken() (string, error)
func (c *Client) InitiateSTKPush(phone, amount, invoiceID string) (*STKPushResponse, error)
```

### Payment Handlers (internal/handlers/payment.go)

```go
func InitMpesaClient() error
func InitiateMpesaPayment(c *gin.Context)
func HandleMpesaCallback(c *gin.Context)
func GetPaymentStatus(c *gin.Context)
```

### Payment Repository (internal/database/payment_repository.go)

```go
func (r *PaymentRepository) CreatePaymentRecord(ctx context.Context, payment *models.PaymentRecord) error
func (r *PaymentRepository) GetPaymentByCheckoutRequestID(ctx context.Context, id string) (*models.PaymentRecord, error)
func (r *PaymentRepository) GetPaymentByInvoiceID(ctx context.Context, invoiceID string) (*models.PaymentRecord, error)
func (r *PaymentRepository) UpdatePaymentStatus(ctx context.Context, checkoutID, status, receiptNum, transDate string) error
```

## Testing

### Unit Tests

```bash
# Test payment request handling (no DB required)
go test -v ./internal/handlers -run TestMpesaPaymentRequest

# Test callback parsing
go test -v ./internal/handlers -run TestHandleMpesaCallback
```

### Integration Testing with Safaricom Sandbox

**Test Credentials** (from `.env.example`):

- Shortcode: `174379`
- Passkey: `bfb279f9aa9bdbcf158e97dd71a467cd2e0ff47d142c1692b53a8f95b491f50a`
- Test Phone: `254708374149` (Safaricom test account)
- Test PIN: `1234`

**Manual Test**:

1. Start the server:

    ```bash
    go run main.go
    ```

2. Create a user:

    ```bash
    curl -X POST http://localhost:8080/api/v1/auth/register \
      -H "Content-Type: application/json" \
      -d '{
        "email": "testuser@example.com",
        "password": "testpass123",
        "firstName": "Test",
        "lastName": "User"
      }'
    ```

3. Login to get token:

    ```bash
    curl -X POST http://localhost:8080/api/v1/auth/login \
      -H "Content-Type: application/json" \
      -d '{
        "email": "testuser@example.com",
        "password": "testpass123"
      }'
    ```

4. Create an order and get the invoice ID

5. Initiate payment:

    ```bash
    curl -X POST http://localhost:8080/api/v1/orders/<order-id>/pay \
      -H "Authorization: Bearer <token>" \
      -H "Content-Type: application/json" \
      -d '{
        "invoiceId": "<invoice-id>",
        "phone": "254708374149"
      }'
    ```

6. Customer receives STK Push and enters PIN

7. Safaricom sends callback to `/api/v1/mpesa/callback`

8. Payment record is updated and invoice shows paid amount

## Common Issues & Solutions

### Issue: "M-Pesa client not initialized"

**Solution**: Ensure all `MPESA_*` environment variables are set in `.env`:

```bash
MPESA_CONSUMER_KEY=xxx
MPESA_CONSUMER_SECRET=xxx
MPESA_BUSINESS_SHORTCODE=174379
MPESA_PASSKEY=xxx
MPESA_CALLBACK_URL=https://yourdomain.com/api/v1/mpesa/callback
MPESA_ENV=sandbox
```

### Issue: "Invalid consumer key/secret"

**Solution**:

1. Verify you're using credentials from Daraja (not production API)
2. Ensure credentials haven't been rotated
3. Check for extra spaces or newlines in `.env`

### Issue: "ResultCode 1" in callback

**Possible Causes**:

- Payment amount is 0 or invalid
- Customer has insufficient M-Pesa balance
- Phone number format is incorrect (should include country code: 254...)
- Timeout: Customer didn't enter PIN within 5 minutes

### Issue: Callback not received

**Troubleshooting**:

1. Verify callback URL is **publicly accessible** (not localhost)
2. Check callback URL is **HTTPS** (Safaricom requires it)
3. Verify firewall/router allows incoming requests
4. Check server logs for any errors during callback processing
5. Whitelist Safaricom IP addresses if behind a firewall

### Issue: "Phone number format invalid"

**Solution**: M-Pesa phone numbers must:

- Include country code: `254` (Kenya)
- Be 12 digits total: `254712345678`
- No leading `+` or spaces

## Security Considerations

### For Sandbox:

- Keep credentials in `.env` (gitignored)
- No signature validation needed

### For Production:

1. **Use environment variables only** - never hardcode credentials
2. **Validate callback signatures** using Safaricom's public key (implement webhook verification)
3. **Use HTTPS only** for callback URLs
4. **Implement idempotency** - handle duplicate callbacks gracefully
5. **Log all callbacks** for audit trails
6. **Rate limit** payment endpoints (prevent abuse)
7. **Verify amounts** - confirm invoice amount matches payment amount
8. **Implement retry logic** for failed Daraja API calls
9. **Store raw callback** data for debugging and compliance

### Sample Callback Verification (Future Enhancement)

```go
// Verify M-Pesa callback signature
func VerifyCallbackSignature(body []byte, signature string) bool {
    // Use Safaricom's public key to verify HMAC
    h := hmac.New(sha256.New, []byte(publicKey))
    h.Write(body)
    computed := base64.StdEncoding.EncodeToString(h.Sum(nil))
    return hmac.Equal([]byte(computed), []byte(signature))
}
```

## Next Steps

1. **Testing**: Configure sandbox credentials and test the full flow
2. **Frontend Integration**: Build UI to display STK Push prompts and payment status
3. **Notifications**: Send email/SMS confirmations on successful payment
4. **Admin Dashboard**: Display payment analytics and transaction history
5. **Refunds**: Implement M-Pesa reversal/refund functionality
6. **Webhook Validation**: Add signature verification for production security
7. **Audit Logging**: Store all payment transactions for compliance

## Resources

- **Safaricom Daraja Docs**: https://developer.safaricom.co.ke/documentation
- **M-Pesa Express API**: https://developer.safaricom.co.ke/docs?python#lipa-na-m-pesa-online-api
- **OAuth 2.0 Guide**: https://developer.safaricom.co.ke/docs?python#authentication
- **Error Codes**: https://developer.safaricom.co.ke/docs?python#error-codes-and-responses

## Example Database Queries

### Find all initiated payments

```javascript
db.paymentrecords.find({ status: "initiated" });
```

### Find completed payments for a date range

```javascript
db.paymentrecords.find({
    status: "completed",
    createdAt: { $gte: ISODate("2024-01-01"), $lt: ISODate("2024-02-01") },
});
```

### Join payments with invoices

```javascript
db.paymentrecords.aggregate([
    {
        $lookup: {
            from: "invoices",
            localField: "invoiceId",
            foreignField: "_id",
            as: "invoice",
        },
    },
]);
```

## Support

For issues with:

- **Safaricom API**: Contact [Daraja Support](https://developer.safaricom.co.ke/support)
- **This Implementation**: Check the error logs and ensure `.env` is configured
- **MongoDB**: Check collection indexes and connection settings
