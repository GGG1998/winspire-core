# Winspire Project Overview

## Purpose
Tournament/competition management platform with Go microservices, React frontends, and shared libraries.

## Tech Stack
- **Backend**: Go microservices (Gin framework), PostgreSQL, Redis pub/sub
- **Frontend**: React 19, Vite 7, Tailwind v4, React Query, Zod, React Hook Form
- **Infrastructure**: Docker Compose (local), Supabase, AWS ECS (production)

## Architecture
- DDD bounded contexts: tournament, matchmaking, game-management
- Each service owns its PostgreSQL schema
- Inter-service communication via Redis pub/sub (async) and REST APIs (sync)

## Key Patterns
- **SQLC**: SQL query generation from `.sql` files - never write raw SQL in Go
- **Atlas**: Database migrations in `services/<name>/migrations/`
- **Feature-based frontend**: `src/features/<feature>/` with api/, components/, hooks/, schemas.ts, types.ts

## Service Ports
- tournament: 8089
- matchmaking: 8088
- game-management: 8087
