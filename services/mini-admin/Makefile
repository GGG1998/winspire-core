.PHONY: help sqlc migrate-up migrate-down migrate-status build run clean, init
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

sqlc: ## Generate Go code from SQL queries
	sqlc generate

migrate-up: ## Run database migrations up
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "Error: DATABASE_URL is not set"; \
		exit 1; \
	fi
	atlas migrate apply --dir "file://migrations" --url "$(DATABASE_URL)"

migrate-down: ## Rollback last migration
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "Error: DATABASE_URL is not set"; \
		exit 1; \
	fi
	atlas migrate down --dir "file://migrations" --url "$(DATABASE_URL)"

migrate-status: ## Show migration status
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "Error: DATABASE_URL is not set"; \
		exit 1; \
	fi
	atlas migrate status --dir "file://migrations" --url "$(DATABASE_URL)"

build: ## Build the service binary
	go build -o bin/mini-admin ./cmd/mini-admin

run: ## Run the service locally
	go run ./cmd/mini-admin/main.go

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf internal/repository/sqlstore/*.go

test: ## Run tests
	go test -v ./...

lint: ## Run linter
	golangci-lint run

docker-build: ## Build Docker image
	docker build -t mini-admin:latest -f cmd/mini-admin/Dockerfile .

init:
	go mod download
	make migrate-up
	make sqlc
	make run

.DEFAULT_GOAL := help

