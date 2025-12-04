.PHONY: help build test run clean sqlc migrate-up migrate-down docker-build

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the matchmaking service binary"
	@echo "  test         - Run all tests"
	@echo "  run          - Run the service locally"
	@echo "  clean        - Remove build artifacts"
	@echo "  sqlc         - Generate SQLC code from SQL queries"
	@echo "  migrate-up   - Run database migrations"
	@echo "  migrate-down - Rollback last migration"
	@echo "  docker-build - Build Docker image"

# Build the service
build:
	@echo "Building matchmaking service..."
	go build -o bin/matchmaking ./cmd/matchmaking

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

# Run the service
run:
	@echo "Running matchmaking service..."
	go run ./cmd/matchmaking/main.go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out

# Generate SQLC code
sqlc:
	@echo "Generating SQLC code..."
	sqlc generate

# Run migrations
migrate-up:
	@echo "Running migrations..."
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "ERROR: DATABASE_URL environment variable not set"; \
		exit 1; \
	fi
	goose -dir migrations postgres "$(DATABASE_URL)" up

# Rollback migrations
migrate-down:
	@echo "Rolling back migration..."
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "ERROR: DATABASE_URL environment variable not set"; \
		exit 1; \
	fi
	goose -dir migrations postgres "$(DATABASE_URL)" down

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t winspire-matchmaking:latest -f cmd/matchmaking/Dockerfile .

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest

