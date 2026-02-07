# Maggiesb - Ecommerce App

A Go-based ecommerce application with JWT authentication.

## Features

- User registration and login with JWT authentication
- Password hashing with bcrypt
- Role-based access control
- Protected API endpoints

## Project Structure

```
├── internal/
│   ├── auth/          # JWT and password utilities
│   ├── handlers/      # HTTP request handlers
│   ├── middleware/    # Authentication middleware
│   └── models/        # Data models and request/response types
├── main.go            # Application entry point
├── go.mod             # Go module dependencies
└── README.md          # This file
```

## API Endpoints

### Public Endpoints

#### Register User

```
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

```
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

#### Health Check

```
GET /health

Response (200):
{
  "status": "ok"
}
```

### Protected Endpoints

All protected endpoints require the `Authorization` header:

```
Authorization: Bearer <your_jwt_token>
```

#### Get User Profile

```
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

```
POST /api/v1/logout

Response (200):
{
  "message": "logged out successfully"
}
```

## Getting Started

### Prerequisites

- Go 1.25.6 or higher

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

3. Run the application:

```bash
go run main.go
```

The server will start on `http://localhost:8080`

## Testing with cURL

### Register a new user

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "firstName": "Test",
    "lastName": "User"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### Get profile (replace TOKEN with actual JWT)

```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer TOKEN"
```

## Security Notes

⚠️ **Important**: The JWT secret key is hardcoded for development. Before deploying to production:

1. Move the secret key to an environment variable
2. Use a strong, randomly generated secret key
3. Store the secret securely (e.g., in a secrets manager)
4. Replace the in-memory user storage with a database
5. Enable HTTPS
6. Add rate limiting
7. Implement token refresh mechanism

## Environment Variables

Consider adding the following environment variables:

```
JWT_SECRET_KEY=your-secret-key-here
PORT=8080
```

## Future Enhancements

- [ ] Database integration (MongoDB/PostgreSQL)
- [ ] Email verification
- [ ] Password reset functionality
- [ ] Refresh tokens
- [ ] Rate limiting
- [ ] Product management
- [ ] Shopping cart
- [ ] Order management
- [ ] Payment integration
