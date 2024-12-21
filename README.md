# Go Ledger API

A Go-based ledger system for managing user balances and transactions.

## Prerequisites

- Docker Desktop
- PostgreSQL (only if running locally without Docker)
- Go 1.21 or higher (only if running locally without Docker)

## Quick Start with Docker

1. Clone the repository:
```bash
git clone https://github.com/yigitdemirko/go-ledger.git
cd go-ledger
```

2. Start the application:
```bash
docker compose up --build
```

The API will be available at `http://localhost:8080`

## Running Locally (Without Docker)

1. Create a `.env` file:
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

2. Create database:
```bash
make db-create
```

3. Run the application:
```bash
make run
```

## Testing the API

### 1. Health Check
```bash
curl http://localhost:8080/health
```

### 2. Create Users

Regular User:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123",
    "name": "Test User"
  }'
```

Admin User:
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

### 3. Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

Save the token from the response for subsequent requests.

### 4. View User Details
```bash
# Replace YOUR_JWT_TOKEN with the token from login/register
curl -X GET http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 5. List All Users (Admin Only)
```bash
# Replace ADMIN_JWT_TOKEN with the admin's token
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

### 6. Transfer Money
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

### 7. View Transaction History
```bash
curl -X GET http://localhost:8080/api/v1/users/1/transactions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Development Commands

```bash
# Start the application with Docker
docker compose up --build    # Start with build
docker compose up -d        # Start in background
docker compose down        # Stop containers

# Local development commands
make build                # Build the application
make run                 # Run locally
make db-reset           # Reset database
make fmt               # Format code
make lint             # Run linter
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

## Error Handling

All errors follow this format:
```json
{
  "error": "Error message description"
}
```

Common status codes:
- 200: Success
- 201: Created
- 400: Bad Request
- 401: Unauthorized
- 403: Forbidden
- 404: Not Found
- 500: Internal Server Error

## Stopping the Application

If running with Docker:
```bash
docker compose down   # Stop and remove containers
```

If running locally:
Press `Ctrl+C` to stop the server 