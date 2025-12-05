# Implementation Plan: Tournament Pre-Lobby Backend

**Branch**: `004-tournament-prelobby-backend` | **Date**: 2025-12-05 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `./spec.md`

## Summary

Add tournament pre-lobby backend functionality to the existing **matchmaking service**. This includes:
- REST API endpoint `GET /v1/tournaments/:id/lobby` for retrieving pre-lobby state
- WebSocket endpoint `WS /v1/tournaments/:id/lobby` for real-time participant presence
- Grace period management (30-second window after tournament start)
- Participant snapshot persistence for bracket generation
- Integration with competition service for registration verification via Redis pub/sub

The implementation extends existing matchmaking infrastructure (WebSocket Hub, pub/sub subscriber, domain events) rather than creating new services.

## Technical Context

**Language/Version**: Go 1.25  
**Primary Dependencies**: Gin (HTTP framework), gorilla/websocket (WebSocket), pgx/v5 (PostgreSQL), go-redis/v9  
**Storage**: PostgreSQL (grace period state, participant snapshots, activity feed), In-memory (WebSocket connections)  
**Testing**: Go testing package with table-driven tests  
**Target Platform**: Linux server (Docker container)  
**Project Type**: Modular monorepo - extending `services/matchmaking/`  
**Performance Goals**: 200ms p95 REST API, 1s WebSocket event propagation, 50 concurrent participants per lobby  
**Constraints**: 5s disconnect detection, 30s grace period accuracy ±1s  
**Scale/Scope**: 50 participants per pre-lobby, multiple concurrent tournaments

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Modular Monorepo Compliance**: ✅ PASS
- Feature extends existing service: `services/matchmaking/`
- No new services created
- Follows established patterns in matchmaking service

**Independent Modules**: ✅ PASS
- Changes contained within matchmaking service's `go.mod`
- No new workspace entries required

**Bounded Contexts & Domain-Driven Design**: ✅ PASS
- `matchmaking` service owns pre-lobby real-time mechanics (Matchmaking BC)
- `competition` service owns tournament registration (TournamentManagement BC)
- Communication via Redis pub/sub events (`TournamentStarted`)
- No direct REST calls between services for domain operations

**Technology Stack**: ✅ PASS
- Go 1.25, Gin, SQLC, pgx/v5 (per constitution)
- Redis for pub/sub (existing infrastructure)
- PostgreSQL for persistence (existing database)

**File Naming Conventions**: ✅ PASS
- Repository files: `prelobby.go` (not `prelobby_repo.go`)
- Handler files: `prelobby.go` (not `prelobby_handler.go`)
- Domain entities: `prelobby.go` for PreLobby aggregate

## Project Structure

### Documentation (this feature)

```text
docs-site/docs/004-tournament-prelobby-backend/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 research output
├── data-model.md        # Phase 1 data model
├── quickstart.md        # Phase 1 quickstart guide
├── contracts/           # Phase 1 API contracts
│   └── prelobby-api.yaml
└── tasks.md             # Phase 2 tasks (created by /speckit.tasks)
```

### Source Code (repository root)

```text
services/matchmaking/
├── cmd/matchmaking/
│   └── main.go                    # Add pre-lobby routes and subscriber
├── internal/
│   ├── application/
│   │   ├── prelobby_service.go    # NEW: Pre-lobby business logic
│   │   └── event_handler.go       # MODIFY: Add TournamentStarted handler
│   ├── domain/
│   │   ├── prelobby.go            # NEW: PreLobby aggregate
│   │   └── events.go              # MODIFY: Add pre-lobby events
│   ├── http/
│   │   └── prelobby.go            # NEW: REST + WebSocket handlers
│   ├── pubsub/
│   │   └── channels.go            # MODIFY: Add pre-lobby channels
│   ├── repository/
│   │   └── prelobby.go            # NEW: PreLobby repository
│   ├── store/
│   │   ├── prelobby_queries.sql   # NEW: SQLC queries
│   │   └── schema.sql             # MODIFY: Add pre-lobby tables
│   └── websocket/
│       ├── hub.go                 # MODIFY: Support tournament-based rooms
│       └── message.go             # MODIFY: Add pre-lobby message types
├── migrations/
│   └── 000005_create_prelobby_tables.sql  # NEW: Database migration
└── sqlc.yaml                      # MODIFY: Add prelobby queries
```

**Structure Decision**: Extends existing matchmaking service following established patterns. New files follow constitution naming conventions (no `_handler`, `_repo` suffixes). Pre-lobby is treated as a new aggregate within the Matchmaking bounded context.

## Complexity Tracking

> No Constitution Check violations. No complexity justifications needed.

---

## Phase 0: Research

### Research Topics

1. **WebSocket Hub Extension**: How to support tournament-based rooms alongside existing match-based rooms
2. **Competition Service Integration**: REST API contract for verifying tournament registration
3. **Grace Period Timer**: Best practices for accurate server-side countdown timers in Go
4. **Participant Snapshot**: Optimal PostgreSQL schema for storing participant lists

### Research Findings

See [research.md](./research.md) for detailed findings.

---

## Phase 1: Design

### Data Model

See [data-model.md](./data-model.md) for complete entity definitions.

### API Contracts

See [contracts/prelobby-api.yaml](./contracts/prelobby-api.yaml) for OpenAPI specification.

### Quickstart Guide

See [quickstart.md](./quickstart.md) for development setup instructions.
