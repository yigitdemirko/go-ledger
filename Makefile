# Go parameters
BINARY_NAME=ledger
GO=go
GOBUILD=$(GO) build
GORUN=$(GO) run
GOCLEAN=$(GO) clean
GOTEST=$(GO) test
GOGET=$(GO) get
GOMOD=$(GO) mod
MAIN_FILE=main.go

# Build flags
LDFLAGS=-ldflags "-w -s"

.PHONY: all build clean run test deps tidy fmt lint help generate-secret test-api

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

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Database commands
.PHONY: db-create db-migrate db-reset

db-create: ## Create database
	psql -U postgres -c "DROP DATABASE IF EXISTS ledger_db;"
	psql -U postgres -c "CREATE DATABASE ledger_db;"

db-migrate: ## Run database migrations (placeholder)
	@echo "Running database migrations..."
	$(GORUN) $(MAIN_FILE)

db-reset: db-create db-migrate ## Reset database and run migrations

# Default target
.DEFAULT_GOAL := help 