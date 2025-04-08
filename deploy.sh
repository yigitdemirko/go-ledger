#!/bin/bash

# Update system packages
sudo apt-get update
sudo apt-get upgrade -y

# Install required packages
sudo apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    nginx \
    certbot \
    python3-certbot-nginx

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add current user to docker group
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.3/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Create application directory
sudo mkdir -p /opt/go-ledger
cd /opt/go-ledger

# Copy application files (you'll need to copy your files here)
# For now, we'll create a placeholder docker-compose.yml
cat > docker-compose.yml << 'EOL'
version: '3.8'

services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=db
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_NAME=ledger_db
      - JWT_SECRET=your-secret-key
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=ledger_db
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  postgres_data:
EOL

# Configure Nginx
sudo tee /etc/nginx/sites-available/go-ledger << 'EOL'
server {
    listen 80;
    server_name _;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
EOL

# Enable Nginx site
sudo ln -s /etc/nginx/sites-available/go-ledger /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default

# Test Nginx configuration
sudo nginx -t

# Start Nginx
sudo systemctl restart nginx

# Start the application
sudo docker compose up -d

echo "Deployment complete! Please make sure to:"
echo "1. Copy your application files to /opt/go-ledger"
echo "2. Update the JWT_SECRET in docker-compose.yml"
echo "3. Configure your domain name in Nginx"
echo "4. Run 'sudo certbot --nginx' to set up HTTPS" 