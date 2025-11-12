.PHONY: help build run dev test clean install deps

# Default target
help:
	@echo "Accounting Web - Makefile Commands"
	@echo "==================================="
	@echo "make build      - Build the application"
	@echo "make run        - Run the application (production binaries)"
	@echo "make dev        - Run in development mode with live reload"
	@echo "make test       - Run tests"
	@echo "make clean      - Clean build artifacts"
	@echo "make deps       - Download and tidy dependencies"
	@echo "make install    - Install for production (requires sudo)"
	@echo "make db-setup   - Setup database"
	@echo "make db-reset   - Reset database (WARNING: deletes all data)"

# Build the application
build:
	@echo "Building applications..."
	@go build -o accounting-web ./cmd/web
	@go build -o accounting-worker ./cmd/worker
	@echo "Build completed!"

# Run the application
run: build
	@echo "Starting applications..."
	@./accounting-web &
	@./accounting-worker &

# Development mode
dev:
	@echo "Starting in development mode..."
	@bash scripts/start.sh

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f accounting-web accounting-worker
	@echo "Clean completed!"

# Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated!"

# Install for production
install:
	@echo "Installing for production..."
	@sudo bash scripts/install.sh

# Setup database
db-setup:
	@echo "Setting up database..."
	@mysql -u root -p < database/schema.sql
	@echo "Database setup completed!"

# Reset database (WARNING)
db-reset:
	@echo "WARNING: This will delete all data!"
	@read -p "Are you sure? (yes/no): " confirm && [ "$$confirm" = "yes" ] || exit 1
	@mysql -u root -p -e "DROP DATABASE IF EXISTS accounting_db;"
	@mysql -u root -p < database/schema.sql
	@echo "Database reset completed!"

# Create storage directories
storage:
	@echo "Creating storage directories..."
	@mkdir -p storage/uploads
	@mkdir -p storage/exports
	@mkdir -p storage/logs
	@chmod -R 777 storage
	@echo "Storage directories created!"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted!"

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run
	@echo "Linting completed!"
