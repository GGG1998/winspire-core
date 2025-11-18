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

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

