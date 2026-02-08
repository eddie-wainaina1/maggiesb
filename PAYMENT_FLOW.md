# M-Pesa Payment Flow - Complete Reference

This document provides a step-by-step walkthrough of the M-Pesa payment integration with actual code examples.

## 1. User Initiates Payment

### Client Request

```bash
curl -X POST http://localhost:8080/api/v1/orders/order-123/pay \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiI..." \
  -H "Content-Type: application/json" \
  -d '{
    "invoiceId": "inv-456",
    "phone": "254712345678"
  }'
```

### Request Validation in Handler

```go
// internal/handlers/payment.go - InitiateMpesaPayment()
var req models.MpesaPaymentRequest
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}

// Verify user is authenticated
userID, exists := c.Get("userID")
if !exists {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
    return
}

// Get invoice from database
invoiceRepo := database.NewInvoiceRepository()
invoice, err := invoiceRepo.GetInvoiceByID(context.Background(), req.InvoiceID)
if err != nil {
    if err == mongo.ErrNoDocuments {
        c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
        return
    }
}

// Verify user owns the order
orderRepo := database.NewOrderRepository()
order, err := orderRepo.GetOrderByID(context.Background(), invoice.OrderID)
if order.UserID != userID.(string) {
    c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
    return
}
```

## 2. Server Initiates STK Push

### Call M-Pesa API

```go
// internal/handlers/payment.go - InitiateMpesaPayment()
stkResp, err := mpesaClient.InitiateSTKPush(
    req.Phone,
    strconv.FormatFloat(invoice.InvoiceAmount, 'f', 2, 64),
    req.InvoiceID,
)
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
}
```

### M-Pesa Client Implementation

```go
// internal/payment/mpesa.go - InitiateSTKPush()
func (c *Client) InitiateSTKPush(phone, amount, invoiceID string) (*STKPushResponse, error) {
    // Get OAuth token
    token, err := c.GetAccessToken()
    if err != nil {
        return nil, err
    }

    // Generate timestamp and password
    timestamp := time.Now().Format("20060102150405")
    password := c.generatePassword(timestamp)

    // Build request payload
    payload := map[string]interface{}{
        "BusinessShortCode": c.config.BusinessShortCode,
        "Password":          password,
        "Timestamp":         timestamp,
        "TransactionType":   "CustomerPayBillOnline",
        "Amount":            amount,
        "PartyA":            phone,
        "PartyB":            c.config.BusinessShortCode,
        "PhoneNumber":       phone,
        "CallBackURL":       c.config.CallbackURL,
        "AccountReference":  invoiceID,
        "TransactionDesc":   "Payment for invoice " + invoiceID,
    }

    // Call Safaricom API
    url := c.baseURL + "/mpesa/stkpush/v1/processrequest"
    // ... HTTP request with OAuth token ...
    // Parse response and return CheckoutRequestID
}
```

### Create Payment Record

```go
// internal/handlers/payment.go - InitiateMpesaPayment()
paymentRepo := database.NewPaymentRepository()
payment := &models.PaymentRecord{
    ID:                uuid.New().String(),
    InvoiceID:         req.InvoiceID,
    OrderID:           invoice.OrderID,
    CheckoutRequestID: stkResp.CheckoutRequestID,
    MerchantRequestID: stkResp.MerchantRequestID,
    Phone:             req.Phone,
    Amount:            invoice.InvoiceAmount,
    Status:            "initiated",
}

if err := paymentRepo.CreatePaymentRecord(context.Background(), payment); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record payment"})
    return
}
```

### MongoDB Payment Record Created

```json
{
    "_id": "uuid-123",
    "invoiceId": "inv-456",
    "orderId": "order-123",
    "checkoutRequestId": "ws_CO_191220191020375651",
    "merchantRequestId": "29115-1234567-1",
    "phone": "254712345678",
    "amount": 100.0,
    "status": "initiated",
    "createdAt": "2024-01-15T10:30:00Z",
    "updatedAt": "2024-01-15T10:30:00Z"
}
```

### Response to Client

```json
{
    "checkoutRequestId": "ws_CO_191220191020375651",
    "customerMessage": "Please enter your M-Pesa PIN",
    "paymentId": "uuid-123"
}
```

## 3. Customer Receives STK Push

Safaricom sends an STK prompt to the customer's phone:

```
LIPA NA MPESA
Karibu! Ingiza PIN yako kwenye Lipa na M-Pesa
Amount: Ksh 100.00
Business: Your Business Name
```

Customer enters their M-Pesa PIN → Payment is processed by Safaricom

## 4. Safaricom Sends Callback

### Callback URL

```
POST https://yourdomain.com/api/v1/mpesa/callback
```

### Callback Payload (Success Case)

```json
{
    "Body": {
        "stkCallback": {
            "MerchantRequestID": "29115-1234567-1",
            "CheckoutRequestID": "ws_CO_191220191020375651",
            "ResultCode": 0,
            "ResultDesc": "The service request has been processed successfully.",
            "CallbackMetadata": {
                "Item": [
                    {
                        "Name": "Amount",
                        "Value": 100
                    },
                    {
                        "Name": "MpesaReceiptNumber",
                        "Value": "NHY4GT5HJI"
                    },
                    {
                        "Name": "TransactionDate",
                        "Value": "20240115103015"
                    },
                    {
                        "Name": "PhoneNumber",
                        "Value": 254712345678
                    }
                ]
            }
        }
    }
}
```

### Callback Payload (Failure Case)

```json
{
    "Body": {
        "stkCallback": {
            "MerchantRequestID": "29115-1234567-1",
            "CheckoutRequestID": "ws_CO_191220191020375651",
            "ResultCode": 1032,
            "ResultDesc": "Request cancelled by user",
            "CallbackMetadata": {
                "Item": []
            }
        }
    }
}
```

## 5. Server Processes Callback

### Handler Implementation

```go
// internal/handlers/payment.go - HandleMpesaCallback()
var callback models.MpesaCallback
if err := c.ShouldBindJSON(&callback); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid callback"})
    return
}

stkCallback := callback.Body.StkCallback
paymentRepo := database.NewPaymentRepository()
invoiceRepo := database.NewInvoiceRepository()

// Find existing payment
payment, err := paymentRepo.GetPaymentByCheckoutRequestID(
    context.Background(),
    stkCallback.CheckoutRequestID,
)
if err != nil {
    c.JSON(http.StatusOK, gin.H{"ResultCode": "1", "ResultDesc": "Payment not found"})
    return
}

// Determine status based on ResultCode
status := "failed"
receiptNum := ""
transDate := ""

if stkCallback.ResultCode == 0 {
    status = "completed"
    // Extract metadata
    for _, item := range stkCallback.CallbackMetadata.Item {
        if item.Name == "MpesaReceiptNumber" {
            receiptNum = fmt.Sprintf("%v", item.Value)
        }
        if item.Name == "TransactionDate" {
            transDate = fmt.Sprintf("%v", item.Value)
        }
    }
}

// Update payment record
if err := paymentRepo.UpdatePaymentStatus(
    context.Background(),
    stkCallback.CheckoutRequestID,
    status,
    receiptNum,
    transDate,
); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
        "ResultCode": "1",
        "ResultDesc": "Failed to update payment",
    })
    return
}
```

### Update Invoice on Success

```go
// internal/handlers/payment.go - HandleMpesaCallback()
if status == "completed" {
    invoice, err := invoiceRepo.GetInvoiceByID(context.Background(), payment.InvoiceID)
    if err == nil {
        // Record payment in invoice
        invoiceRepo.RecordPayment(
            context.Background(),
            payment.InvoiceID,
            invoice.InvoiceAmount,
            transDate,
        )
    }
}
```

### MongoDB Updates

**Payment Record Update**:

```json
{
    "_id": "uuid-123",
    "invoiceId": "inv-456",
    "orderId": "order-123",
    "checkoutRequestId": "ws_CO_191220191020375651",
    "phone": "254712345678",
    "amount": 100.0,
    "mpesaReceiptNumber": "NHY4GT5HJI",
    "transactionDate": "20240115103015",
    "status": "completed",
    "createdAt": "2024-01-15T10:30:00Z",
    "updatedAt": "2024-01-15T10:35:00Z"
}
```

**Invoice Update**:

```json
{
    "_id": "inv-456",
    "orderId": "order-123",
    "invoiceAmount": 100.0,
    "paidAmount": 100.0,
    "taxAmount": 10.0,
    "type": "payable",
    "paidOn": {
        "20240115103015": 100.0
    },
    "updatedAt": "2024-01-15T10:35:00Z"
}
```

### Response to Safaricom

```json
{
    "ResultCode": "0",
    "ResultDesc": "Callback received"
}
```

## 6. Complete Data Flow Diagram

```
Customer's Phone          Your Server              Safaricom API
─────────────────        ───────────────          ──────────────

                          User initiates payment
                          POST /api/v1/orders/:id/pay
                          {invoiceId, phone}
                          ├─ Validate invoice
                          ├─ Call Daraja API
                          │  └─ Get OAuth token
                          │  └─ STK Push request
                          ├─ Create PaymentRecord (status: initiated)
                          └─ Return CheckoutRequestID

                          STK Push Call ──────────→
                                                    ├─ Validate
                                                    ├─ Check balance
                                                    └─ Send prompt

Receives STK Prompt ←─────────────────────────────
│
├─ User sees payment prompt
│  Amount: 100 KSH
│  Business: Your Shop
│
├─ Enters M-Pesa PIN
│  ✓ Accepted
│  ✓ Processing...
│
└─ Payment Processed ─────────────────────────────→
                                                    ├─ Validate PIN
                                                    ├─ Deduct balance
                                                    ├─ Send callback

                          ←───────────────────────
                          Callback received
                          POST /api/v1/mpesa/callback
                          {ResultCode: 0, Receipt: NHY4GT5HJI}
                          ├─ Find PaymentRecord
                          ├─ Update status: completed
                          ├─ Update Invoice
                          │  └─ Add to PaidOn map
                          └─ Return 200 OK
                                                    └─ Callback received
```

## 7. Error Scenarios

### Scenario: Invalid Invoice

```
POST /api/v1/orders/order-123/pay
{invoiceId: "nonexistent"}

Response (404):
{
  "error": "invoice not found"
}
```

### Scenario: User Not Authenticated

```
POST /api/v1/orders/order-123/pay
(No Authorization header)

Response (401):
{
  "error": "user not authenticated"
}
```

### Scenario: User Doesn't Own Invoice

```
POST /api/v1/orders/order-123/pay
(Invoice belongs to different user)

Response (403):
{
  "error": "forbidden"
}
```

### Scenario: M-Pesa API Error

```go
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": "M-Pesa API error: invalid consumer key"
    })
    return
}
```

### Scenario: Customer Cancels Payment

```json
{
  "Body": {
    "stkCallback": {
      "ResultCode": 1032,
      "ResultDesc": "Request cancelled by user"
    }
  }
}

PaymentRecord status updated to "failed"
Invoice remains unpaid
```

## 8. State Machine

```
Payment States:

initiated
    │
    ├─→ completed (ResultCode == 0)
    │       └─→ Invoice updated with PaidAmount
    │       └─→ Invoice PaidOn map updated
    │
    └─→ failed (ResultCode != 0)
            └─→ Invoice remains unpaid
            └─→ User can retry payment
```

## 9. Database Queries

### Find payment for invoice

```javascript
db.paymentrecords.findOne({ invoiceId: "inv-456" });
```

### Find all pending payments

```javascript
db.paymentrecords.find({ status: "initiated" }).limit(10);
```

### Find payments by phone

```javascript
db.paymentrecords.find({ phone: "254712345678" });
```

### Calculate daily payment volume

```javascript
db.paymentrecords.aggregate([
    {
        $match: {
            status: "completed",
            createdAt: { $gte: ISODate("2024-01-15") },
        },
    },
    { $group: { _id: null, total: { $sum: "$amount" }, count: { $sum: 1 } } },
]);
```

## 10. Testing the Flow

### Manual Test (Step by Step)

```bash
# 1. Start server
go run main.go

# 2. Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -d '{"email":"test@example.com","password":"test123","firstName":"Test","lastName":"User"}'
# Response: { "token": "eyJhbGciOiJIUzI1NiI...", "user": {...} }

# 3. Create product (as admin)
TOKEN="your_admin_token"
curl -X POST http://localhost:8080/api/v1/admin/products \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Product 1","price":100,"description":"Test"}'
# Response: { "id": "prod-123" }

# 4. Create order
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"products":[{"productId":"prod-123","quantity":1}]}'
# Response: { "id": "order-123", "invoiceId": "inv-456" }

# 5. Initiate payment
curl -X POST http://localhost:8080/api/v1/orders/order-123/pay \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"invoiceId":"inv-456","phone":"254708374149"}'
# Response: { "checkoutRequestId": "ws_CO_xxx" }

# 6. Check payment in MongoDB
mongosh
use maggiesb
db.paymentrecords.findOne({})

# 7. Simulate callback (for testing)
curl -X POST http://localhost:8080/api/v1/mpesa/callback \
  -H "Content-Type: application/json" \
  -d '{
    "Body": {
      "stkCallback": {
        "CheckoutRequestID": "ws_CO_xxx",
        "ResultCode": 0,
        "CallbackMetadata": {"Item": [
          {"Name": "MpesaReceiptNumber", "Value": "TEST123"},
          {"Name": "TransactionDate", "Value": "20240115103015"}
        ]}
      }
    }
  }'

# 8. Check invoice was updated
db.invoices.findOne({_id: "inv-456"})
# Should show paidAmount: 100, PaidOn: {"20240115103015": 100}
```

## 11. Monitoring Checklist

During payment processing, monitor:

- [ ] PaymentRecord created with "initiated" status
- [ ] M-Pesa API response contains CheckoutRequestID
- [ ] STK Push delivered to customer phone
- [ ] Callback received within 5 minutes
- [ ] PaymentRecord updated with receipt number
- [ ] Invoice PaidAmount increased
- [ ] Invoice PaidOn map contains transaction entry
- [ ] No database errors in logs

---

**See Also**:

- [MPESA_INTEGRATION.md](./MPESA_INTEGRATION.md) - Setup and configuration
- [IMPLEMENTATION_SUMMARY.md](./IMPLEMENTATION_SUMMARY.md) - Architecture overview
- [README.md](./README.md) - API documentation
