.PHONY: help run build test fmt lint clean dev docker-build docker-run

# Variables
BINARY_NAME=space-backend
MAIN_PATH=./cmd/server/main.go

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the application
	@echo "Starting server..."
	go run $(MAIN_PATH)

build: ## Build the application
	@echo "Building..."
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: bin/$(BINARY_NAME)"

dev: ## Run in development mode with hot reload (requires air)
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Install it with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to regular run..."; \
		make run; \
	fi

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

lint: ## Run linter
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it from https://golangci-lint.run/usage/install/"; \
	fi

clean: ## Clean build files
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out
	@echo "Clean complete"

install-deps: ## Install Go dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(BINARY_NAME):latest

migrate: ## Run database migrations
	@echo "Running migrations..."
	go run $(MAIN_PATH) --migrate-only

.DEFAULT_GOAL := help
