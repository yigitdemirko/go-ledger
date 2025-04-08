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

## CI/CD

This project uses GitHub Actions for continuous integration and deployment.

### CI Pipeline
- Runs on every push and pull request to main
- Tests the application
- Builds and pushes Docker image to Docker Hub

### CD Pipeline
- Runs on every push to main
- Deploys the application to EC2

### Setup
1. Add these secrets to your GitHub repository:
   - `DOCKERHUB_USERNAME`: Your Docker Hub username
   - `DOCKERHUB_TOKEN`: Your Docker Hub access token
   - `EC2_HOST`: Your EC2 instance public IP
   - `EC2_SSH_KEY`: Your EC2 SSH private key

2. Make sure your EC2 instance has:
   - Docker and Docker Compose installed
   - The application directory set up
   - Proper security group settings 