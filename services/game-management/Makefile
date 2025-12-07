.PHONY: build run dev test test-coverage lint fmt sqlc migrate migrate-new deps-up deps-down docker-build docker-run clean deps mocks

# Service configuration
BINARY := game-management
SERVICE_DIR := $(shell pwd)
SERVICE_PORT := 8087

# Build the binary
build:
	go build -o bin/$(BINARY) ./cmd/$(BINARY)

# Run the service locally
run: build
	./bin/$(BINARY)

# Run with hot reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Generate SQLC code
sqlc:
	sqlc generate

# Run database migrations (requires atlas)
migrate:
	atlas migrate apply \
		--dir file://$(SERVICE_DIR)/migrations \
		--url "$${POSTGRES_DSN:-postgres://postgres:postgres@localhost:54322/postgres?sslmode=disable}"

# Create a new migration
migrate-new:
	@read -p "Migration name: " name; \
	atlas migrate new $$name --dir file://$(SERVICE_DIR)/migrations

# Start development dependencies
deps-up:
	docker compose -f docker-compose.dev.yaml up -d

# Stop development dependencies
deps-down:
	docker compose -f docker-compose.dev.yaml down

# Build Docker image
docker-build:
	docker build -t $(BINARY):latest -f cmd/$(BINARY)/Dockerfile .

# Run Docker container
docker-run:
	docker run -p $(SERVICE_PORT):$(SERVICE_PORT) --env-file .env $(BINARY):latest

# Download dependencies
deps:
	go mod download
	go mod tidy

# Generate mocks (requires mockgen)
mocks:
	go generate ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
