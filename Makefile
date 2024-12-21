# Go parameters
BINARY_NAME=ledger
GO=go
GOBUILD=$(GO) build
GORUN=$(GO) run
GOCLEAN=$(GO) clean
GOGET=$(GO) get
GOMOD=$(GO) mod
MAIN_FILE=main.go

# Build flags
LDFLAGS=-ldflags "-w -s"

.PHONY: all build clean run deps tidy fmt lint help generate-secret docker-* db-*

all: clean build

build: ## Build the application
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_FILE)

clean: ## Clean build files
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

run: ## Run the application
	$(GORUN) $(MAIN_FILE)

deps: ## Download dependencies
	$(GOGET) -v ./...

tidy: ## Tidy up module dependencies
	$(GOMOD) tidy

fmt: ## Format code
	$(GO) fmt ./...

lint: ## Run linter
	golangci-lint run

generate-secret: ## Generate a secure JWT secret
	@openssl rand -base64 32

# Docker commands
docker-build: ## Build Docker image
	docker compose build

docker-up: ## Start Docker containers
	docker compose up

docker-up-d: ## Start Docker containers in background
	docker compose up -d

docker-down: ## Stop Docker containers
	docker compose down

docker-logs: ## View Docker logs
	docker compose logs -f

docker-ps: ## List running containers
	docker compose ps

docker-clean: ## Remove all Docker containers and images
	docker compose down --rmi all --volumes --remove-orphans

# Database commands
db-create: ## Create database
	psql -U postgres -c "DROP DATABASE IF EXISTS ledger_db;"
	psql -U postgres -c "CREATE DATABASE ledger_db;"

db-migrate: ## Run database migrations (placeholder)
	@echo "Running database migrations..."
	$(GORUN) $(MAIN_FILE)

db-reset: db-create db-migrate ## Reset database and run migrations

db-psql: ## Connect to PostgreSQL database
	psql -U postgres -d ledger_db

# Development commands
dev: docker-up ## Start development environment

install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Default target
.DEFAULT_GOAL := help 