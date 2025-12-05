# Quickstart: Tournament Pre-Lobby Backend

**Feature**: 004-tournament-prelobby-backend  
**Date**: 2025-12-05

## Prerequisites

- Go 1.25+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (for local development)

## Local Development Setup

### 1. Start Infrastructure

```bash
# From repository root
docker-compose up -d postgres redis
```

### 2. Run Migrations

```bash
cd services/matchmaking
make migrate-up
```

### 3. Generate SQLC Code

```bash
cd services/matchmaking
make sqlc
```

### 4. Start Matchmaking Service

```bash
cd services/matchmaking
make run
```

The service starts on `http://localhost:8082` by default.

## Testing the API

### REST Endpoint

```bash
# Get pre-lobby state (requires JWT)
curl -X GET http://localhost:8082/api/v1/tournaments/{tournamentId}/lobby \
  -H "Authorization: Bearer <jwt_token>"
```

### WebSocket Connection

```javascript
// JavaScript example
const ws = new WebSocket(
  'ws://localhost:8082/api/v1/tournaments/{tournamentId}/lobby?token=<jwt_token>'
);

ws.onopen = () => {
  console.log('Connected to pre-lobby');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message.type, message.payload);
};

// Send heartbeat every 30s
setInterval(() => {
  ws.send(JSON.stringify({
    type: 'heartbeat',
    timestamp: new Date().toISOString(),
    payload: { player_id: '<user_id>' }
  }));
}, 30000);
```

## Configuration

Environment variables for matchmaking service:

```bash
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/matchmaking?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# Competition Service (for registration verification)
COMPETITION_SERVICE_URL=http://localhost:8081

# Server
PORT=8082
```

## Key Files

| File | Purpose |
|------|---------|
| `internal/domain/prelobby.go` | PreLobby aggregate and business rules |
| `internal/application/prelobby_service.go` | Pre-lobby business logic |
| `internal/http/prelobby.go` | REST and WebSocket handlers |
| `internal/repository/prelobby.go` | Database operations |
| `internal/websocket/hub.go` | WebSocket connection management |
| `migrations/000005_create_prelobby_tables.sql` | Database schema |

## Testing

### Unit Tests

```bash
cd services/matchmaking
go test ./internal/... -v
```

### Integration Tests

```bash
cd services/matchmaking
go test ./... -tags=integration -v
```

## Common Issues

### "Not registered for tournament"

The user must be registered in the competition service before accessing the pre-lobby. Verify registration:

```bash
curl http://localhost:8081/api/v1/{hostId}/tournaments/{tournamentId}/participants/{userId}
```

### WebSocket disconnects immediately

- Check JWT token is valid and not expired
- Verify tournament is in valid state (scheduled, registration_open, registration_closed)
- Ensure tournament start time is within 15 minutes

### Grace period not starting

The grace period starts when competition service publishes `TournamentStarted` event via Redis pub/sub. Verify:

1. Competition service is running
2. Redis pub/sub is working
3. Matchmaking service is subscribed to `events:tournament_management:tournament_started` channel

## Architecture Notes

```
┌─────────────────┐     Redis Pub/Sub      ┌─────────────────┐
│   Competition   │ ────────────────────▶  │   Matchmaking   │
│    Service      │   TournamentStarted    │    Service      │
└─────────────────┘                        └─────────────────┘
        │                                          │
        │ REST API                                 │ WebSocket
        │ (registration check)                     │ (real-time)
        ▼                                          ▼
┌─────────────────┐                        ┌─────────────────┐
│   PostgreSQL    │                        │    Frontend     │
│  (competition)  │                        │   (pre-lobby)   │
└─────────────────┘                        └─────────────────┘
```

## Next Steps

After implementing the backend:

1. Update frontend to use new pre-lobby endpoints
2. Add monitoring/alerting for grace period failures
3. Load test with 50 concurrent participants

