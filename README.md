# Maggiesb - Ecommerce API

A production-ready Go-based ecommerce API with JWT authentication, M-Pesa payment integration, order management, and admin capabilities.

## Features

- **User Management**: Registration, login, and profile management with JWT authentication
- **Product Management**: Browse, search, and manage products with pricing and discounts
- **Order Management**: Create and track orders with itemization and status tracking
- **Invoice System**: Generate and manage invoices for orders
- **M-Pesa Integration**: Process payments via M-Pesa with callback handling
- **Role-Based Access Control**: Support for admin and user roles with protected endpoints
- **MongoDB Database**: Persistent storage with indexed collections

## Project Structure

```
├── internal/
│   ├── auth/           # JWT token management and cleanup routines
│   ├── database/       # MongoDB repositories for all entities
│   ├── handlers/       # HTTP request handlers for all endpoints
│   ├── middleware/     # Authentication and authorization middleware
│   ├── models/         # Data models (User, Product, Order, Invoice, Payment)
│   └── payment/        # M-Pesa payment integration
├── main.go             # Application entry point with router setup
├── go.mod              # Go module dependencies
└── README.md           # This file
```

## API Endpoints

### Authentication Endpoints (Public)

#### Register User

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123",
  "firstName": "John",
  "lastName": "Doe"
}

Response (201):
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid-string",
    "email": "user@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "role": "user",
    "createdAt": "2024-02-01T10:00:00Z",
    "updatedAt": "2024-02-01T10:00:00Z"
  },
  "expiresAt": 1706865600
}
```

#### Login User

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123"
}

Response (200):
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": { ... },
  "expiresAt": 1706865600
}
```

### Product Endpoints (Public)

#### List Products

```http
GET /api/v1/products

Response (200):
[
  {
    "id": "product-uuid",
    "name": "Product Name",
    "description": "Product description",
    "price": 999.99,
    "discount": 10.0,
    "createdAt": "2024-02-01T10:00:00Z",
    "updatedAt": "2024-02-01T10:00:00Z"
  }
]
```

#### Search Products

```http
GET /api/v1/products/search?q=query

Response (200):
[...]
```

#### Get Single Product

```http
GET /api/v1/products/:id

Response (200):
{
  "id": "product-uuid",
  "name": "Product Name",
  "description": "Product description",
  "price": 999.99,
  "discount": 10.0,
  "createdAt": "2024-02-01T10:00:00Z",
  "updatedAt": "2024-02-01T10:00:00Z"
}
```

### Protected User Endpoints

All protected endpoints require the `Authorization` header:

```
Authorization: Bearer <your_jwt_token>
```

#### Get User Profile

```http
GET /api/v1/profile

Response (200):
{
  "id": "uuid-string",
  "email": "user@example.com",
  "firstName": "John",
  "lastName": "Doe",
  "role": "user",
  "createdAt": "2024-02-01T10:00:00Z",
  "updatedAt": "2024-02-01T10:00:00Z"
}
```

#### Logout

```http
POST /api/v1/logout

Response (200):
{
  "message": "logged out successfully"
}
```

### Order Endpoints (Protected)

#### Create Order

```http
POST /api/v1/orders
Authorization: Bearer <token>
Content-Type: application/json

{
  "products": [
    {
      "productId": "product-uuid",
      "quantity": 2
    }
  ],
  "phone": "254712345678",
  "metadata": {
    "notes": "Please handle with care",
    "locationDetails": "Nairobi, Kenya"
  }
}

Response (201):
{
  "id": "order-uuid",
  "products": [...],
  "cost": 1999.98,
  "discount": 100.0,
  "totalCost": 1899.98,
  "status": "in queue",
  "user": "user-uuid",
  "phone": "254712345678",
  "metadata": {...},
  "createdAt": "2024-02-01T10:00:00Z",
  "updatedAt": "2024-02-01T10:00:00Z"
}
```

#### List User Orders

```http
GET /api/v1/orders
Authorization: Bearer <token>

Response (200):
[...]
```

#### Get Order Details

```http
GET /api/v1/orders/:id
Authorization: Bearer <token>

Response (200):
{...}
```

### Invoice Endpoints (Protected)

#### Get Invoice

```http
GET /api/v1/invoices/:id
Authorization: Bearer <token>

Response (200):
{
  "id": "invoice-uuid",
  "orderId": "order-uuid",
  "userId": "user-uuid",
  "status": "issued",
  "totalAmount": 1899.98,
  "paidAmount": 0.0,
  "dueDate": "2024-02-15T00:00:00Z",
  "createdAt": "2024-02-01T10:00:00Z"
}
```

#### Get Invoice by Order

```http
GET /api/v1/orders/:id/invoice
Authorization: Bearer <token>

Response (200):
{...}
```

### Payment Endpoints (Protected)

#### Initiate M-Pesa Payment

```http
POST /api/v1/orders/:id/pay
Authorization: Bearer <token>
Content-Type: application/json

{
  "invoiceId": "invoice-uuid",
  "phone": "254712345678"
}

Response (200):
{
  "checkoutRequestId": "ws_CO_DMZ_...",
  "responseCode": "0",
  "responseMessage": "Success. Request accepted for processing",
  "customerMessage": "Success. Request accepted for processing"
}
```

#### Get Payment Status

```http
GET /api/v1/payments/:id/status
Authorization: Bearer <token>

Response (200):
{
  "id": "payment-uuid",
  "invoiceId": "invoice-uuid",
  "status": "completed",
  "amount": 1899.98,
  "mpesaReceiptNumber": "NHY4GT5HJI",
  "transactionDate": "2024-02-01T10:30:00Z"
}
```

#### M-Pesa Callback (Public)

```http
POST /api/v1/mpesa/callback
Content-Type: application/json

{
  "Body": {
    "stkCallback": {
      "MerchantRequestID": "...",
      "CheckoutRequestID": "...",
      "ResultCode": 0,
      "ResultDesc": "The service request has been processed successfully.",
      "CallbackMetadata": {
        "Item": [
          {"Name": "Amount", "Value": 1899.98},
          {"Name": "MpesaReceiptNumber", "Value": "NHY4GT5HJI"},
          {"Name": "TransactionDate", "Value": "20240201103000"}
        ]
      }
    }
  }
}
```

### Admin Endpoints (Protected + Admin Role)

#### Create Product

```http
POST /api/v1/admin/products
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "Product Name",
  "description": "Product description",
  "price": 999.99,
  "discount": 10.0
}

Response (201):
{...}
```

#### Update Product

```http
PUT /api/v1/admin/products/:id
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "Updated Name",
  "price": 1099.99
}

Response (200):
{...}
```

#### Delete Product

```http
DELETE /api/v1/admin/products/:id
Authorization: Bearer <admin_token>

Response (204):
```

#### List All Orders (Admin)

```http
GET /api/v1/admin/orders
Authorization: Bearer <admin_token>

Response (200):
[...]
```

#### Update Order Status (Admin)

```http
PUT /api/v1/admin/orders/:id/status
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "status": "shipped"
}

Response (200):
{...}
```

#### List All Invoices (Admin)

```http
GET /api/v1/admin/invoices
Authorization: Bearer <admin_token>

Response (200):
[...]
```

#### Record Payment (Admin)

```http
PUT /api/v1/admin/invoices/:id/payment
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "paidAmount": 1899.98
}

Response (200):
{...}
```

### Health Check (Public)

```http
GET /health

Response (200):
{
  "status": "ok"
}
```

## Getting Started

### Prerequisites

- Go 1.25.6 or higher
- MongoDB 4.0+
- M-Pesa/Safaricom Daraja account (optional, for payment processing)

### Installation

1. Clone the repository:

```bash
git clone <repository-url>
cd maggiesb
```

2. Download dependencies:

```bash
go mod download
```

3. Set up environment variables:

```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Ensure MongoDB is running:

```bash
# Start MongoDB locally (if using Docker)
docker run -d -p 27017:27017 --name mongodb mongo:latest
```

5. Run the application:

```bash
go run main.go
```

The server will start on `http://localhost:8080` (or the port specified in `PORT` env var)

## Configuration

### Required Environment Variables

```env
# MongoDB
MONGODB_URI=mongodb://localhost:27017
DB_NAME=maggiesb

# Server
PORT=8080

# M-Pesa (Optional - app will warn if not configured)
MPESA_CONSUMER_KEY=your_key
MPESA_CONSUMER_SECRET=your_secret
MPESA_BUSINESS_SHORTCODE=174379
MPESA_PASSKEY=your_passkey
MPESA_CALLBACK_URL=https://yourdomain.com/api/v1/mpesa/callback
MPESA_ENV=sandbox
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test file
go test -v ./internal/auth -run TestJWT
```

### Building

```bash
# Build binary
go build -o bin/maggiesb main.go

# Run the built binary
./bin/maggiesb
```

- [ ] Refund/reversal handling
- [x] Refund/reversal handling (admin manual reversal endpoint: `PUT /api/v1/admin/invoices/:id/reverse`)
- [ ] Payment webhook retry logic
- [ ] Admin payment analytics dashboard
- [ ] Email notifications for payments
- [x] Refund/reversal handling implemented (DB + admin endpoint)
- [ ] Payment webhook retry logic
- [ ] Admin payment analytics dashboard
- [ ] Email notifications for payments

Admin reversal notes:

- Endpoint: `PUT /api/v1/admin/invoices/:id/reverse` (admin only)
- Payload: `ReverseInvoiceRequest` — fields: `amount`, `date` (YYYY-MM-DD), `phone`, `useMpesa`, `reason`.
- To enable automatic M-Pesa reversal, set these env vars: `MPESA_INITIATOR_NAME`, `MPESA_INITIATOR_PASSWORD`, `MPESA_PUBLIC_KEY_PATH`, and `MPESA_CALLBACK_URL`/`MPESA_REVERSAL_RESULT_URL`/`MPESA_REVERSAL_TIMEOUT_URL` as needed.
