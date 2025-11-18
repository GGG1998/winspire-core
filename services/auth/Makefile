.PHONY: build test run migrate-up migrate-down sqlc-generate clean

# Build the auth service
build:
	go build -o bin/auth ./cmd/auth

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Run the service
run:
	go run ./cmd/auth

# Run database migrations up
migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

# Run database migrations down
migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down

# Generate sqlc models
sqlc-generate:
	sqlc generate

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

