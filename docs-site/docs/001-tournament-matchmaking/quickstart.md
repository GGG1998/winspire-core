# Quickstart: Tournament Matchmaking

**Feature**: 001-tournament-matchmaking  
**Date**: 2025-12-04  
**Prerequisites**: Go 1.25, PostgreSQL 14+, Redis 7+, Docker (optional)

This guide walks you through setting up and running the tournament matchmaking **microservice** locally.

---

## Prerequisites

### Required Software

| Software | Version | Purpose |
|----------|---------|---------|
| **Go** | 1.25 | Backend microservice |
| **PostgreSQL** | 14+ | Database (separate matchmaking_db) |
| **Redis** | 7+ | Event pub/sub (microservice communication) |
| **Docker** (optional) | Latest | Local infrastructure |
| **sqlc** | 1.25+ | SQL code generation |
| **goose** | Latest | Database migrations |

### Install Prerequisites

```bash
# macOS (Homebrew)
brew install go postgresql redis sqlc goose

# Ubuntu/Debian
sudo apt-get update
sudo apt-get install golang postgresql redis-server

# Install sqlc and goose
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/pressly/goose/v3/cmd/goose@latest
```

---

## Quick Start (Docker Compose)

The fastest way to get started is using Docker Compose for infrastructure:

```bash
# From repository root
cd platform/local

# Copy example environment
cp env.example .env

# Start PostgreSQL + Redis + Supabase (optional)
make up

# Verify services running
docker ps
```

### Environment Variables

Create `services/matchmaking/.env.dev` with:

```bash
# Database (separate from competition_db)
DATABASE_URL="postgres://postgres:postgres@localhost:5432/matchmaking_db?sslmode=disable"

# Redis (for event-driven communication with competition service)
REDIS_URL="redis://localhost:6379"
REDIS_CHANNEL_PREFIX="winspire.events"

# Game API (external)
GAME_API_BASE_URL="https://game-api.example.com"
GAME_API_KEY="your-api-key"

# JWT (shared secret with competition service)
JWT_SECRET="dev-secret-key"  # Use secure key in production

# Server
PORT=8081  # Different port from competition service (8080)
ENV=development
```

---

## Database Setup

### 1. Create Matchmaking Database

```bash
# Create separate database for matchmaking service
psql -U postgres -c "CREATE DATABASE matchmaking_db;"
```

### 2. Run Migrations

```bash
cd services/matchmaking

# Apply all migrations (new matchmaking tables)
goose -dir migrations postgres "$DATABASE_URL" up

# Verify tables created
psql "$DATABASE_URL" -c "\dt"
```

**Expected output**:
```
 Schema |          Name                | Type  |  Owner
--------+------------------------------+-------+----------
| public | tournament_brackets          | table | postgres
| public | tournament_rounds            | table | postgres
| public | tournament_matches           | table | postgres
```

**Note**: This is a separate database from `competition_db` (which has `tournaments`, `tournament_registrations` tables). The two services communicate via Redis Pub/Sub events.

### 3. Generate SQL Code (SQLC)

```bash
cd services/matchmaking

# Generate type-safe Go code from SQL queries
sqlc generate

# Verify generated files
ls internal/store/*.go
```

---

## Build & Run

### 1. Install Dependencies

```bash
cd services/matchmaking

# Download Go modules
go mod download

# Verify go.work references matchmaking service
cd ../..
grep matchmaking go.work
```

### 2. Build Service

```bash
cd services/matchmaking

# Build binary
make build

# OR build for development with hot-reload (requires air)
go install github.com/cosmtrek/air@latest
air
```

### 3. Run Service

```bash
cd services/matchmaking

# Run with environment file
./matchmaking

# OR run directly with go run
go run cmd/matchmaking/main.go
```

**Expected output**:
```
INFO: Starting matchmaking service on :8081
INFO: Using Gin HTTP framework
INFO: Database migrations: up-to-date (matchmaking_db)
INFO: Connected to Redis at localhost:6379
INFO: Subscribed to event: TournamentStarted
INFO: Health check endpoint: GET /health
```

### 4. Verify Service Running

```bash
# Health check
curl http://localhost:8081/health

# Expected response:
{
  "status": "healthy",
  "database": "connected",
  "redis": "connected",
  "timestamp": "2025-12-04T10:00:00Z"
}
```

---

## Microservice Communication

The matchmaking service communicates with the competition service via Redis Pub/Sub:

```bash
# Terminal 1: Start competition service (publishes TournamentStarted)
cd services/competition
go run cmd/competition/main.go

# Terminal 2: Start matchmaking service (subscribes to TournamentStarted)
cd services/matchmaking
go run cmd/matchmaking/main.go

# Terminal 3: Trigger tournament start (in competition service)
curl -X POST http://localhost:8080/v1/tournaments/{tournament_id}/start \
  -H "Authorization: Bearer $TOKEN"

# Matchmaking service will receive event and generate bracket automatically
```

---

## API Examples

### 1. Get Bracket Visualization

```bash
# Get complete bracket structure (from matchmaking service)
curl http://localhost:8081/v1/brackets/{bracket_id}

# Or by tournament ID
curl http://localhost:8081/v1/tournaments/{tournament_id}/bracket

# Response (200 OK):
{
  "bracket_id": "uuid",
  "tournament_id": "uuid",
  "total_rounds": 3,
  "total_matches": 7,
  "byes_count": 1,
  "generated_at": "2025-12-04T10:00:00Z",
  "rounds": [
    {
      "round_id": "uuid",
      "round_number": 1,
      "round_name": "Quarter-finals",
      "matches_count": 4,
      "status": "in_progress",
      "matches": [
        {
          "match_id": "uuid",
          "match_number": 1,
          "participant1_id": "uuid",
          "participant2_id": "uuid",
          "status": "started",
          "participant1_ready": true,
          "participant2_ready": true,
          "score_player1": 2,
          "score_player2": 1
        }
      ]
    }
  ]
}
```

### 2. Mark Player Ready in Lobby

```bash
# Player marks ready (matchmaking service)
curl -X POST http://localhost:8081/v1/matches/{match_id}/ready \
  -H "Authorization: Bearer $TOKEN"

# Response (200 OK):
{
  "match_id": "uuid",
  "player_ready": true,
  "opponent_ready": false,
  "match_status": "ready"
}
```

### 3. Forfeit Match

```bash
# Player forfeits (opponent gets walkover)
curl -X POST http://localhost:8081/v1/matches/{match_id}/forfeit \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Internet connection issues"}'

# Response (200 OK):
{
  "match_id": "uuid",
  "status": "completed",
  "winner_id": "opponent-uuid",
  "result_source": "walkover"
}
```

### 4. Host Override Result (Manual Entry)

```bash
# Host manually enters result (API failure fallback)
curl -X PATCH http://localhost:8081/v1/matches/{match_id}/result \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "winner_id": "player-uuid",
    "score_player1": 3,
    "score_player2": 1,
    "override_reason": "Game API timeout - manually verified via screenshot"
  }'

# Response (200 OK):
{
  "match_id": "uuid",
  "status": "completed",
  "winner_id": "player-uuid",
  "score_player1": 3,
  "score_player2": 1,
  "result_source": "host_manual",
  "completed_at": "2025-12-04T10:30:00Z"
}
```

---

## Event Monitoring

### Subscribe to Events (Redis Pub/Sub)

```bash
# Monitor all matchmaking events
redis-cli
> SUBSCRIBE winspire.events.matchmaking.*

# Monitor specific events
> SUBSCRIBE winspire.events.matchmaking.bracket_generated
> SUBSCRIBE winspire.events.matchmaking.match_created
> SUBSCRIBE winspire.events.matchmaking.match_completed

# Example event published by matchmaking service:
{
  "event_id": "01JFXYZ...",
  "event_type": "MatchCreated",
  "aggregate_id": "match-uuid",
  "bounded_context": "Matchmaking",
  "timestamp": "2025-12-04T10:00:00Z",
  "payload": {
    "match_id": "uuid",
    "round_id": "uuid",
    "tournament_id": "uuid",
    "participant1_id": "uuid",
    "participant2_id": "uuid",
    "status": "pending"
  }
}
```

---

## Testing

### Unit Tests

```bash
cd services/matchmaking

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/application/...
```

### Integration Tests

```bash
# Start test database
docker run --name test-matchmaking-db -e POSTGRES_PASSWORD=test -p 5433:5432 -d postgres:14

# Create test database
psql -U postgres -h localhost -p 5433 -c "CREATE DATABASE matchmaking_test;"

# Run integration tests
TEST_DATABASE_URL="postgres://postgres:test@localhost:5433/matchmaking_test?sslmode=disable" \
  go test -tags=integration ./...
```

### API Tests (curl)

```bash
# View bracket
curl http://localhost:8081/v1/tournaments/<tournament-id>/bracket | jq

# Get match details
curl http://localhost:8081/v1/matches/<match-id> | jq
```

---

## Troubleshooting

### Issue: "Database connection failed"

**Solution**: Verify PostgreSQL running and `matchmaking_db` exists

```bash
# Check PostgreSQL status
pg_isready -h localhost -p 5432

# Verify database exists
psql -U postgres -l | grep matchmaking_db

# Create if missing
psql -U postgres -c "CREATE DATABASE matchmaking_db;"
```

### Issue: "Redis connection failed"

**Solution**: Verify Redis running

```bash
# Check Redis status
redis-cli ping
# Expected: PONG
```

### Issue: "Not receiving TournamentStarted events"

**Solution**: Verify both services are running and using same Redis instance

```bash
# Check if competition service is publishing events
redis-cli
> MONITOR

# Start tournament in competition service and watch for event publication
```

### Issue: "Bracket generation takes too long"

**Solution**: Check database indexes

```bash
psql "$DATABASE_URL" -c "\d tournament_matches"

# Verify indexes exist:
# - idx_matches_round_id
# - idx_matches_next_match_id
```

### Issue: "Game API polling not working"

**Solution**: Check game API credentials and network

```bash
# Test game API directly
curl -H "Authorization: Bearer $GAME_API_KEY" \
  $GAME_API_BASE_URL/api/health

# Check service logs
tail -f logs/matchmaking.log | grep "game_api"
```

---

## Development Workflow

### 1. Add New Migration

```bash
cd services/matchmaking

# Create new migration
goose -dir migrations create add_new_column sql

# Edit migration file
vim migrations/000004_add_new_column.sql

# Apply migration
goose -dir migrations postgres "$DATABASE_URL" up
```

### 2. Add New SQL Query (SQLC)

```bash
cd services/matchmaking

# Edit query file
vim internal/store/match_queries.sql

# Add query:
-- name: GetActiveMatches :many
SELECT * FROM tournament_matches
WHERE status IN ('pending', 'started');

# Generate Go code
sqlc generate
```

### 3. Hot Reload (Development)

```bash
cd services/matchmaking

# Install air (if not installed)
go install github.com/cosmtrek/air@latest

# Run with hot reload
air

# Edit any .go file → service auto-restarts
```

### 4. Add New Event Handler

```bash
cd services/matchmaking

# Edit event handler
vim internal/application/event_handler.go

# Subscribe to new event type
func (h *EventHandler) HandleTournamentEvent(event Event) {
    switch event.Type {
    case "TournamentStarted":
        h.bracketService.GenerateBracket(event.TournamentID)
    case "TournamentCancelled":  // NEW
        h.bracketService.CancelBracket(event.TournamentID)
    }
}
```

---

## Next Steps

1. **Implement Game API Client** (`internal/gameapi/client.go`)
   - See `research.md` for polling pattern

2. **Implement WebSocket Server** (for realtime updates)
   - See `.specify/darft/001-matchmaking.md` for flow

3. **Add Disconnect Handling** (`internal/application/disconnect_handler.go`)
   - CS:GO-style 30s reconnection window

4. **Add Event Publishing** (`internal/pubsub/publisher.go`)
   - Redis Pub/Sub integration for `MatchCreated`, `MatchCompleted`

5. **Create Admin UI** (for host overrides)
   - Build on existing frontend in `frontends/winspire-app`

---

## Resources

- **API Documentation**: [api.openapi.yaml](./contracts/api.openapi.yaml)
- **Event Schemas**: [events.yaml](./contracts/events.yaml)
- **Data Model**: [data-model.md](./data-model.md)
- **Research Decisions**: [research.md](./research.md)
- **Architecture Docs**: `docs-site/docs/platform/`

---

## Microservice Architecture

```
┌──────────────────────┐         Redis Pub/Sub          ┌──────────────────────┐
│ Competition Service  │────────────────────────────────▶│ Matchmaking Service  │
│ (port 8080)          │  TournamentStarted event       │ (port 8081)          │
│                      │                                 │                      │
│ - competition_db     │                                 │ - matchmaking_db     │
│ - Tournament mgmt    │                                 │ - Bracket generation │
└──────────────────────┘                                 │ - Match management   │
                                                          └──────────────────────┘
                                                                   │
                                                                   ▼
                                                          Publishes events:
                                                          - BracketGenerated
                                                          - MatchCreated
                                                          - MatchCompleted
```

---

## Support

- **Issues**: File on GitHub repository
- **Questions**: Ask in #engineering Slack channel
- **Architecture Decisions**: Review `docs-site/docs/reference/adr/`

Happy coding! 🚀
