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

## API Endpoints

### 1. Health Check
```bash
curl http://localhost:8080/health
```
Response:
```json
{
  "status": "ok",
  "time": "2024-04-08T13:42:55+03:00"
}
```

### 2. Authentication

#### Register User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123",
    "name": "Test User"
  }'
```

#### Register Admin
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

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "name": "Test User",
    "username": "testuser",
    "role": "USER"
  }
}
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

### 3. User Management

#### Get User Details
```bash
curl -X GET http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Response:
```json
{
  "id": 1,
  "name": "Test User",
  "balance": 1000.00,
  "created_at": "2024-04-08T13:46:36.747086Z",
  "updated_at": "2024-04-08T13:46:42.630252Z"
}
```

#### List All Users (Admin Only)
```bash
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

Response:
```json
[
  {
    "id": 1,
    "name": "Test User",
    "balance": 1000.00,
    "created_at": "2024-04-08T13:46:36.747086Z",
    "updated_at": "2024-04-08T13:46:42.630252Z"
  },
  {
    "id": 2,
    "name": "Admin User",
    "balance": 0.00,
    "created_at": "2024-04-08T13:44:28.286444Z",
    "updated_at": "2024-04-08T13:44:28.286444Z"
  }
]
```

#### Initialize User Balance (Admin Only)
```bash
curl -X POST http://localhost:8080/api/v1/users/1/initialize-balance \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1000.00
  }'
```

Response:
```json
{
  "message": "Balance initialized successfully"
}
```

### 4. Transactions

#### Transfer Money
```bash
curl -X POST http://localhost:8080/api/v1/transfer \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "from_user_id": 1,
    "to_user_id": 2,
    "amount": 200.00
  }'
```

Response:
```json
{
  "message": "Transfer successful",
  "from_user": {
    "id": 1,
    "balance": 800.00
  },
  "to_user": {
    "id": 2,
    "balance": 700.00
  },
  "transaction": {
    "id": 3,
    "from_user_id": 1,
    "to_user_id": 2,
    "amount": 200.00,
    "transaction_type": "TRANSFER",
    "created_at": "2024-04-08T13:47:45.724064Z"
  }
}
```

#### View Transaction History
```bash
curl -X GET http://localhost:8080/api/v1/users/1/transactions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Response:
```json
[
  {
    "id": 3,
    "from_user_id": 1,
    "to_user_id": 2,
    "amount": 200.00,
    "transaction_type": "TRANSFER",
    "created_at": "2024-04-08T13:47:45.724064Z"
  },
  {
    "id": 1,
    "from_user_id": null,
    "to_user_id": 1,
    "amount": 1000.00,
    "transaction_type": "DEPOSIT",
    "created_at": "2024-04-08T13:46:42.630252Z"
  }
]
```

#### Get Historical Balance
```bash
curl -X GET "http://localhost:8080/api/v1/users/1/balance/historical?timestamp=2024-04-08T13:47:00Z" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Response:
```json
{
  "balance": 1000.00,
  "timestamp": "2024-04-08T13:47:00Z"
}
```

### 5. Account Management

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

Response:
```json
{
  "message": "Password updated successfully"
}
```

## Error Responses

### Insufficient Balance
```json
{
  "error": "Source user not found or insufficient balance"
}
```

### Unauthorized Access
```json
{
  "error": "Insufficient permissions"
}
```

### Invalid Request
```json
{
  "error": "Invalid request body"
}
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

## Deployment

### AWS EC2 Setup

1. Create an AWS account and launch an EC2 instance:
   - Go to AWS Console > EC2 > Launch Instance
   - Choose Ubuntu Server 22.04 LTS
   - Select t2.micro (free tier eligible)
   - Configure security group to allow:
     - SSH (port 22)
     - HTTP (port 80)
     - HTTPS (port 443)
   - Launch instance and download key pair

2. Connect to your EC2 instance:
   ```bash
   chmod 400 your-key-pair.pem
   ssh -i your-key-pair.pem ubuntu@your-ec2-public-ip
   ```

3. Deploy the application:
   ```bash
   # Copy your application files to the server
   scp -i your-key-pair.pem -r . ubuntu@your-ec2-public-ip:/opt/go-ledger/

   # Make the deployment script executable and run it
   ssh -i your-key-pair.pem ubuntu@your-ec2-public-ip
   cd /opt/go-ledger
   chmod +x deploy.sh
   ./deploy.sh
   ```

4. Set up HTTPS:
   ```bash
   sudo certbot --nginx
   ```

### Security Considerations

- Update the JWT_SECRET in docker-compose.yml
- Use strong passwords for the database
- Regularly update system packages
- Monitor application logs
- Set up automated backups 