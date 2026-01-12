# Suggested Commands for Development

## Local Development (from platform/local/)
```bash
make start          # Start all services via Docker Compose
make rebuild        # Rebuild and restart all services (after Go changes)
make logs           # View all service logs
make logs-service SERVICE=tournament  # View specific service logs
make stop           # Stop all services
make redis-cli      # Connect to Redis CLI
make db-connect     # Connect to Postgres
```

## Go Services (from services/<name>/)
```bash
make build          # Build binary
make run            # Build and run service
make test           # Run tests
make sqlc           # Generate SQLC code from SQL queries
make migrate        # Apply database migrations (Atlas)
make lint           # Run golangci-lint
```

## Frontend winspire-app (from frontends/winspire-app/)
```bash
yarn dev            # Start dev server
yarn build          # TypeScript compile + Vite build
yarn lint           # ESLint
```

## After Code Changes
1. Go changes: `cd platform/local && make rebuild`
2. SQL changes: `cd services/<name> && make sqlc`
3. Migration changes: `cd services/<name> && make migrate`
