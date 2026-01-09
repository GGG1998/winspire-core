# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Winspire is a tournament/competition management platform built as a monorepo with Go microservices, React frontends, and shared libraries.

## Architecture

### Bounded Contexts (Domain-Driven Design)

- **tournament**: Tournament lifecycle & participant registration (TournamentCreated, Published, Started, Completed)
- **matchmaking**: Match creation, brackets, lobbies, scoring (BracketGenerated, MatchCreated, ScoreSubmitted)
- **game-management**: Game catalog & asset management (GameBundle storage in S3)

Services communicate via Redis pub/sub (async events) and REST APIs (sync queries). Each service owns its own PostgreSQL database schema.

### Repository Structure

```
services/           # Go microservices (each has own go.mod)
├── tournament/     # Port 8089
├── matchmaking/    # Port 8088
├── game-management/# Port 8087
└── template/       # Cookiecutter template for new services

libs/go/            # Shared Go libraries
├── auth/           # JWT middleware, UserContext types
├── httpx/          # HTTP middleware (CORS, Recovery, RequestLogger)
└── pgtype/         # pgtype <-> Go type converters

frontends/
├── winspire-app/   # Main React app (React 19, Vite 7, Tailwind v4)
└── mini-admin/     # Admin dashboard (React 18, Vite 5)

platform/
├── local/          # Docker Compose development environment
├── supabase/       # Supabase configuration
└── terraform/      # Infrastructure as code
```

## Common Commands

### Local Development (from platform/local/)
```bash
make start          # Start all services via Docker Compose
make rebuild        # Rebuild and restart all services (use after Go code changes)
make logs           # View all service logs
make logs-service SERVICE=tournament  # View specific service logs
make stop           # Stop all services
make clean          # Remove all containers, volumes, images
make redis-cli      # Connect to Redis CLI
make db-connect     # Connect to Postgres
```

### Go Services (from services/<name>/)
```bash
make build          # Build binary
make run            # Build and run service
make test           # Run tests
make test-coverage  # Tests with coverage report
make sqlc           # Generate SQLC code from SQL queries
make migrate        # Apply database migrations (Atlas)
make lint           # Run golangci-lint
```

### Frontend winspire-app (from frontends/winspire-app/)
```bash
yarn dev            # Start dev server
yarn build          # TypeScript compile + Vite build
yarn lint           # ESLint
yarn test:e2e       # Playwright tests
yarn test:e2e:ui    # Playwright with UI
yarn storybook      # Run Storybook
```

### Frontend mini-admin (from frontends/mini-admin/)
```bash
yarn dev            # Start dev server
yarn build          # TypeScript compile + Vite build
```

### Root Commands
```bash
make new-service-interactive  # Create new microservice from template
make sync                     # Sync go.work with all services
make dev-mini-admin           # Run game-management + mini-admin together
```

## Go Development Guidelines

### Database: SQLC + Atlas Migrations

- **Never write raw SQL** in Go code - use sqlc-generated queries
- Migrations live in `services/<name>/migrations/` (numbered: `000001_description.sql`)
- After schema changes: `make migrate` then `make sqlc`

### SQLC Query Annotations
```sql
-- name: GetByID :one      -- Returns single row
-- name: List :many        -- Returns []Row
-- name: Create :exec      -- No return value
-- name: Update :execrows  -- Returns rows affected
```

### Type Conversions (use libs/go/pgtype)
```go
uuidToPgtype(id)         // uuid.UUID -> pgtype.UUID
pgtypeToUUID(id)         // pgtype.UUID -> uuid.UUID
stringToPgtypeText(s)    // *string -> pgtype.Text
```

### HTTP Handler Pattern (Gin framework)
- Use `c.Request.Context()` for context propagation
- Group handlers by domain in separate files
- Dependencies via struct injection, no globals
- Validation via Gin binding tags (`binding:"required,min=2,max=100"`)

## TypeScript/React Guidelines

### Frontend Structure (Feature-based)
```
src/
├── features/<feature>/
│   ├── api/          # React Query hooks, fetch calls
│   ├── components/   # Feature-specific components
│   ├── hooks/        # Feature-specific hooks
│   ├── schemas.ts    # Zod validation schemas
│   ├── types.ts      # TypeScript interfaces
│   └── constants.ts  # UI labels, validation rules
└── shared/
    ├── components/ui/  # Reusable UI components
    └── api/            # Shared API clients
```

### Key Patterns
- **Zod** for all validation schemas (export types via `z.infer<typeof schema>`)
- **React Hook Form** with `zodResolver` for forms
- **React Query** for server state management
- Import UI components from `@/shared/components/ui/`

## Important Notes

- Go services require `make rebuild` in platform/local after code changes (Docker must rebuild)
- Each service has its own database - no shared tables between services
- Inter-service communication uses events (Redis pub/sub), not direct DB queries
- JWT auth middleware from `libs/go/auth/middleware`
- Supabase auth best practices documented in `.claude/rules/supabase_auth.md`
- **Debugging state machines** (multi-layer fixes) documented in `.claude/rules/debugging_state_machines.md`
