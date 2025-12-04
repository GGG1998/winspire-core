.PHONY: help build test run clean sqlc migrate-up migrate-down migrate-status docker-build install-tools

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the matchmaking service binary"
	@echo "  test         - Run all tests"
	@echo "  run          - Run the service locally"
	@echo "  clean        - Remove build artifacts"
	@echo "  sqlc         - Generate SQLC code from SQL queries"
	@echo "  migrate-up   - Apply database migrations (Atlas)"
	@echo "  migrate-down - Rollback last migration (Atlas)"
	@echo "  migrate-status - Show migration status (Atlas)"
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

# Run migrations with Atlas
migrate-up:
	@echo "Applying migrations with Atlas..."
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "ERROR: DATABASE_URL environment variable not set"; \
		exit 1; \
	fi
	atlas migrate apply \
		--dir "file://migrations" \
		--url "$(DATABASE_URL)"

# Rollback migrations with Atlas
migrate-down:
	@echo "Rolling back migration with Atlas..."
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "ERROR: DATABASE_URL environment variable not set"; \
		exit 1; \
	fi
	@echo "Atlas doesn't support automatic rollback. Please manually apply down migration or use atlas schema apply."
	@exit 1

# Check migration status
migrate-status:
	@echo "Checking migration status..."
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "ERROR: DATABASE_URL environment variable not set"; \
		exit 1; \
	fi
	atlas migrate status \
		--dir "file://migrations" \
		--url "$(DATABASE_URL)"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t winspire-matchmaking:latest -f cmd/matchmaking/Dockerfile .

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo "Installing Atlas CLI..."
	curl -sSf https://atlasgo.sh | sh

