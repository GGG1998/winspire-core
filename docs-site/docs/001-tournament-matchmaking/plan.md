# Implementation Plan: Tournament Matchmaking System

**Branch**: `001-tournament-matchmaking` | **Date**: 2025-12-04 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `./docs-site/docs/001-tournament-matchmaking/spec.md`

## Summary

Build a tournament matchmaking system for single elimination 1v1 tournaments with automatic bracket generation, round management, match handling, and real-time event flow. The system automatically retrieves match results from game API, manages player disconnections with CS:GO-style reconnection (30s window), and supports host overrides for technical issues. Features include random bye assignment for odd participant counts, lobby-based ready states with server-side persistence, and no-queue design (roster locked at tournament start).

**Technical Approach**: Event-driven architecture using Domain-Driven Design with two bounded contexts: **TournamentManagement** (lifecycle, registration in competition service) and **Matchmaking** (brackets, rounds, matches in new matchmaking microservice). Creates new independent `services/matchmaking/` microservice that communicates with competition service via Redis Pub/Sub events.

## Technical Context

**Language/Version**: Go 1.25  
**Primary Dependencies**: 
- `github.com/gin-gonic/gin` (HTTP framework)
- `github.com/jackc/pgx/v5` (PostgreSQL driver)
- `github.com/redis/go-redis/v9` (Pub/Sub for event distribution)
- sqlc (type-safe SQL query generation)
- `libs/go/httpx` (shared HTTP middleware)
- `libs/go/auth` (shared JWT validation)

**Storage**: PostgreSQL (new independent database `matchmaking_db` with 3 tables: `tournament_brackets`, `tournament_rounds`, `tournament_matches`)  
**Testing**: Go testing stdlib + testify for assertions  
**Target Platform**: Linux server (Docker containers, AWS ECS)  
**Project Type**: Backend microservice (new independent service `services/matchmaking/`)  
**Performance Goals**: 
- Bracket generation <2s for large tournaments (tested with 128 participants)
- 50 concurrent tournaments with 64 participants each
- Match state updates propagate <3s

**Constraints**: 
- API result retrieval <10s (p95)
- 95% of matches auto-retrieved from game API
- 30s disconnect reconnection window (hard requirement)

**Scale/Scope**: 
- Configurable max participants per tournament (set by host)
- 35+ domain events
- 30+ functional requirements
- 3 new database tables
- Independent microservice with separate database

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Modular Monorepo Compliance ✅

**Service Placement**: PASS  
New independent service created at `services/matchmaking/` with own `go.mod`.

**Rationale**: Matchmaking bounded context is separate from TournamentManagement. Event-driven communication via Redis Pub/Sub enables independent deployment, scaling, and database ownership. Eventual consistency model aligns with DDD best practices.

### Bounded Contexts & DDD ✅

**Bounded Context**: PASS  
Two bounded contexts identified:
1. **TournamentManagement** BC - Owns: Tournament, TournamentParticipant aggregates
2. **Matchmaking** BC - Owns: Bracket, Round, Match, Team aggregates

**Event-Driven Communication**: PASS  
- TournamentManagement publishes: `TournamentStarted`, `ParticipantCheckedIn`, `TournamentCompleted`
- Matchmaking subscribes to: `TournamentStarted` → triggers bracket generation
- Matchmaking publishes: `BracketGenerated`, `MatchCreated`, `MatchCompleted`
- Realtime service subscribes to Matchmaking events for WebSocket notifications

**No Shared Database**: PASS  
- TournamentManagement BC: Uses `competition_db` (existing)
  - Tables: `tournaments`, `tournament_registrations`, `hosts`
- Matchmaking BC: Uses `matchmaking_db` (NEW, independent)
  - Tables: `tournament_brackets`, `tournament_rounds`, `tournament_matches`

**Communication**: Services communicate via domain events over Redis Pub/Sub. Bracket generation is triggered by `TournamentStarted` event (eventual consistency model).

### Independent Modules ✅

**Dependency Management**: PASS  
New service has its own `services/matchmaking/go.mod` and is registered in root `go.work` file.

### Shared Libraries ✅

**Reuse**: PASS  
- Uses existing `libs/go/auth/` for JWT validation
- Uses existing `libs/go/httpx/` for HTTP middleware

## Project Structure

### Documentation (this feature)

```text
docs-site/docs/001-tournament-matchmaking/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file (Phase 0-1 output)
├── research.md          # Phase 0 output (next)
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (event schemas + OpenAPI)
│   ├── events.yaml      # Domain event schemas
│   └── api.openapi.yaml # REST API spec
├── checklists/          # Quality validation
│   └── requirements.md  # Spec quality checklist (completed)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
services/matchmaking/                             # NEW SERVICE
├── cmd/matchmaking/
│   └── main.go                      # Service entrypoint, Gin router setup
├── internal/
│   ├── application/
│   │   ├── bracket_service.go       # Bracket generation logic
│   │   ├── match_service.go         # Match state management
│   │   └── event_handler.go         # Redis event subscriber (TournamentStarted)
│   ├── domain/
│   │   ├── bracket.go               # Bracket aggregate
│   │   ├── round.go                 # Round entity
│   │   ├── match.go                 # Match aggregate
│   │   └── events.go                # Domain event definitions
│   ├── http/
│   │   ├── bracket_handler.go       # GET /brackets/:id endpoints
│   │   ├── match_handler.go         # POST /matches/:id/ready, /forfeit
│   │   └── middleware.go            # Auth, logging (uses libs/go/httpx)
│   ├── repository/
│   │   ├── bracket_repo.go          # Bracket persistence
│   │   └── match_repo.go            # Match persistence
│   ├── store/
│   │   ├── schema.sql               # SQLC schema reference
│   │   ├── bracket_queries.sql      # SQLC bracket queries
│   │   └── match_queries.sql        # SQLC match queries
│   └── pubsub/
│       ├── publisher.go             # Redis event publisher
│       └── subscriber.go            # Redis event subscriber
├── migrations/
│   ├── 000001_create_brackets_table.sql         # NEW
│   ├── 000002_create_rounds_table.sql           # NEW
│   └── 000003_create_matches_table.sql          # NEW
├── go.mod                           # Independent module
├── go.sum
├── Makefile
└── Dockerfile

services/competition/                             # EXISTING SERVICE
├── (unchanged - owns TournamentManagement BC)
└── Publishes: TournamentStarted event via Redis
```

**Structure Decision**: New independent service at `services/matchmaking/` for Matchmaking bounded context. Follows Go project layout with Gin for HTTP, pgx/v5 for PostgreSQL, SQLC for queries. Communicates with `competition` service via Redis Pub/Sub (event-driven). Uses shared libraries (`libs/go/httpx`, `libs/go/auth`) for common functionality.

## Complexity Tracking

> **No violations requiring justification**

All constitution principles are fully satisfied:
- ✅ Modular Monorepo: New service in `services/matchmaking/`
- ✅ Independent Modules: Own `go.mod`, registered in `go.work`
- ✅ Shared Libraries: Uses `libs/go/httpx` and `libs/go/auth`
- ✅ Bounded Contexts & DDD: Separate databases, event-driven communication
- ✅ Technology Stack: Gin, pgx/v5, SQLC, Go 1.25

---

# Phase 0: Outline & Research

## Research Questions

### 1. Bracket Generation Algorithms

**Question**: How to generate single elimination brackets with automatic bye handling for any participant count?

**Research Needed**:
- Algorithm for calculating number of rounds (log2 ceiling)
- Bye placement strategy (top of bracket vs distributed)
- Match numbering scheme for progression tracking
- Handling power-of-2 vs non-power-of-2 participant counts

### 2. Game API Integration Patterns

**Question**: How to poll/webhook game API for match results with retry logic and fallback?

**Research Needed**:
- Polling frequency (avoid rate limits)
- Exponential backoff for failures
- Circuit breaker pattern for API outages
- Webhook validation for push-based results
- API contract assumptions (endpoints, auth, data format)

### 3. Disconnect Detection & Reconnection

**Question**: How to detect player disconnection, track 30s window, and resume match state?

**Research Needed**:
- WebSocket heartbeat mechanism
- Server-side timer implementation
- Match pause/resume state machine
- Disconnect timestamp persistence
- Handling both players disconnecting simultaneously

### 4. Redis Pub/Sub for Event Distribution

**Question**: How to reliably publish domain events to Redis channels for realtime service consumption?

**Research Needed**:
- Channel naming conventions
- Event payload serialization (JSON/Protobuf)
- At-least-once delivery guarantees
- Subscriber connection handling
- Replay/recovery mechanisms for missed events

### 5. SQLC Best Practices

**Question**: How to structure SQL queries with sqlc for complex joins (brackets + rounds + matches + participants)?

**Research Needed**:
- Query organization (one file vs per-aggregate)
- Complex JOIN patterns for bracket visualization
- Transaction handling with sqlc
- Nullable foreign keys (bye matches have one participant)
- Performance optimization (indexes, query plans)

---

# Phase 1: Design & Contracts

## Design Artifacts To Generate

1. **data-model.md** - Database schema design
   - `tournament_brackets` table
   - `tournament_rounds` table  
   - `tournament_matches` table
   - Relationships and foreign keys
   - Indexes for query performance

2. **contracts/events.yaml** - Domain Event Schemas
   - All 35+ events from DDD catalog
   - Payload structure (JSON schema)
   - Invariants as validation rules
   - Event versioning strategy

3. **contracts/api.openapi.yaml** - REST API Spec
   - POST `/tournaments/{id}/start` - Start tournament (triggers bracket generation)
   - GET `/tournaments/{id}/bracket` - Get bracket with rounds/matches
   - GET `/matches/{id}` - Get match details
   - POST `/matches/{id}/ready` - Mark player ready
   - POST `/matches/{id}/forfeit` - Forfeit match
   - PATCH `/matches/{id}/result` - Host override result
   - GET `/tournaments/{id}/lobby` - Get lobby state for player

4. **quickstart.md** - Developer Setup Guide
   - Prerequisites (Go 1.21+, PostgreSQL, Redis)
   - Database migration steps
   - Environment variables
   - Running locally with docker-compose
   - API examples (curl commands)

---

# Phase 2: Task Decomposition

**STOP**: Phase 2 (task breakdown) is handled by `/speckit.tasks` command.

This plan document ends after Phase 1 design artifacts are generated.

---

## Phase Completion Status

### ✅ Phase 0: Research (COMPLETED)

**Artifacts Generated**:
- [research.md](./research.md) - All 5 research questions answered
  - Bracket generation algorithms (balanced binary tree)
  - Game API integration (polling + circuit breaker)
  - Disconnect handling (WebSocket heartbeat + 30s timer)
  - Redis Pub/Sub (channel-per-event with JSON)
  - SQLC patterns (aggregate-per-file organization)

### ✅ Phase 1: Design & Contracts (COMPLETED)

**Artifacts Generated**:
1. ✅ [data-model.md](./data-model.md) - 3 new tables with migrations
   - `tournament_brackets` (1 per tournament)
   - `tournament_rounds` (N per bracket)
   - `tournament_matches` (N per round, with disconnect tracking)

2. ✅ [contracts/events.yaml](./contracts/events.yaml) - Domain event schemas
   - 10 events from DDD catalog (in-scope only)
   - CloudEvents-compatible JSON schemas
   - Invariants documented per event

3. ✅ [contracts/api.openapi.yaml](./contracts/api.openapi.yaml) - REST API spec
   - 6 endpoints (start tournament, get bracket, match operations)
   - OpenAPI 3.0 specification
   - Security: Bearer JWT authentication

4. ✅ [quickstart.md](./quickstart.md) - Developer setup guide
   - Prerequisites and installation
   - Docker Compose setup
   - Database migrations
   - API examples (curl commands)
   - Troubleshooting guide

5. ✅ **Agent context updated** - Cursor IDE rules file updated with:
   - Language: Go 1.21+
   - Database: PostgreSQL with 3 new tables

### ⏸️ Phase 2: Task Decomposition

**Status**: NOT STARTED (requires `/speckit.tasks` command)

---

## Implementation Summary

### Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Service Placement** | New `services/matchmaking/` | Independent bounded context, separate database |
| **HTTP Framework** | Gin | Constitution-mandated, performant, idiomatic |
| **Database Driver** | pgx/v5 | Constitution-mandated, native PostgreSQL support |
| **Bracket Algorithm** | Balanced binary tree | Simple, fair, standard in e-sports |
| **Score Retrieval** | Game API polling | Reliable, works with any game API |
| **Disconnect Handling** | WebSocket + 30s server timer | CS:GO-style fairness |
| **Event Distribution** | Redis Pub/Sub (JSON) | Low-latency, event-driven microservices |
| **Database Queries** | SQLC | Type-safe, performant, constitution-mandated |

### DDD Event Catalog Coverage

From your `.specify/darft/001-matchmaking.md` event list:

| Event | Status | Notes |
|-------|--------|-------|
| TournamentStarted | ✅ Subscribed | Triggers bracket generation |
| BracketGenerated | ✅ Published | Matchmaking BC |
| RoundCreated | ✅ Published | Matchmaking BC |
| MatchCreated | ✅ Published | Matchmaking BC |
| MatchStarted | ✅ Published | Matchmaking BC |
| MatchCompleted | ✅ Published | Matchmaking BC |
| ParticipantEliminated | ✅ Published | Matchmaking BC |
| ParticipantAdvanced | ✅ Published | Matchmaking BC |
| WalkoverGranted | ✅ Published | Matchmaking BC |
| PlayerConnectionLost | ✅ Published | Matchmaking BC |
| PlayerConnectionRestored | ✅ Published | Matchmaking BC |
| ParticipantJoinedLobby | ✅ Published | Matchmaking BC |
| TournamentCompleted | ✅ Published | TournamentManagement BC |
| TournamentCancelled | ✅ Published | TournamentManagement BC |
| MatchDisputed | ⏸️ Out of Scope | No player disputes (API = source of truth) |
| DisputeResolved | ⏸️ Out of Scope | No disputes |
| RefereeAssigned | ⏸️ Deferred | Future enhancement |
| TeamFormed | ⏸️ Out of Scope | 1v1 only (no teams in MVP) |
| PrizesDistributed | ⏸️ Out of Scope | Rewards BC |
| ScoreStreamedToCache | 🔄 Implicit | Part of realtime updates |

---

## Next Steps

Run the following command to break down into implementation tasks:

```bash
/speckit.tasks
```

This will generate `tasks.md` with detailed implementation checklist organized by component.
