# Go Ledger API

A Go-based ledger system for managing user balances and transactions.

## Prerequisites

- Go 1.21 or higher
- PostgreSQL

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=ledger_db
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=100
DB_MIN_CONNECTIONS=10
JWT_SECRET=your_jwt_secret
```

## Getting Started

1. Clone the repository
2. Set up environment variables
3. Run database migrations:
   ```bash
   make db-reset
   ```
4. Start the server:
   ```bash
   make run
   ```

## API Endpoints

### Authentication

#### Register a New User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123",
    "name": "Test User"
  }'
```

#### Register an Admin User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123",
    "name": "Admin User",
    "role": "ADMIN"
  }'
```

#### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

### User Operations

#### Get User Details
```bash
curl -X GET http://localhost:8080/api/v1/users/{user_id} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Get All Users (Admin Only)
```bash
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Change Password
```bash
curl -X POST http://localhost:8080/api/v1/users/change-password \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "password123",
    "new_password": "newpassword123"
  }'
```

### Transaction Operations

#### Transfer Credits
```bash
curl -X POST http://localhost:8080/api/v1/transfer \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "from_user_id": 1,
    "to_user_id": 2,
    "amount": 100.00
  }'
```

#### Get Transaction History
```bash
curl -X GET http://localhost:8080/api/v1/users/{user_id}/transactions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Get Historical Balance
```bash
curl -X GET "http://localhost:8080/api/v1/users/{user_id}/balance/historical?timestamp=2024-01-20T15:00:00Z" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Response Examples

### Successful Registration
```json
{
  "token": "eyJhbGc...",
  "user": {
    "id": 1,
    "name": "Test User",
    "username": "testuser",
    "role": "USER"
  }
}
```

### Successful Login
```json
{
  "token": "eyJhbGc...",
  "user": {
    "id": 1,
    "username": "testuser",
    "role": "USER"
  }
}
```

### User Details
```json
{
  "id": 1,
  "name": "Test User",
  "balance": 100.00,
  "created_at": "2024-01-20T15:00:00Z",
  "updated_at": "2024-01-20T15:00:00Z"
}
```

### Successful Transfer
```json
{
  "message": "Transfer successful",
  "from_user": {
    "id": 1,
    "balance": 0.00
  },
  "to_user": {
    "id": 2,
    "balance": 100.00
  },
  "transaction": {
    "id": 1,
    "from_user_id": 1,
    "to_user_id": 2,
    "amount": 100.00,
    "transaction_type": "TRANSFER",
    "created_at": "2024-01-20T15:00:00Z"
  }
}
```

## Error Responses

All error responses follow this format:
```json
{
  "error": "Error message description"
}
```

Common HTTP status codes:
- 200: Success
- 201: Created
- 400: Bad Request
- 401: Unauthorized
- 403: Forbidden
- 404: Not Found
- 500: Internal Server Error

## Development

### Available Make Commands
```bash
make build      # Build the application
make run        # Run the application
make db-reset   # Reset and migrate database
make fmt        # Format code
make lint       # Run linter
``` 