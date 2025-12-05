# Research: Tournament Pre-Lobby Backend

**Feature**: 004-tournament-prelobby-backend  
**Date**: 2025-12-05

## 1. WebSocket Hub Extension for Tournament Rooms

### Decision
Extend the existing `Hub` struct to support both match-based and tournament-based rooms using a unified room abstraction.

### Rationale
The current Hub is organized by `matchID`. Pre-lobby needs organization by `tournamentID`. Rather than creating a separate hub, we extend the existing one with a `RoomType` discriminator to reuse connection management, heartbeat monitoring, and broadcast logic.

### Implementation Approach
```go
type RoomType string

const (
    RoomTypeMatch      RoomType = "match"
    RoomTypeTournament RoomType = "tournament"
)

type Client struct {
    RoomType  RoomType
    RoomID    uuid.UUID  // Either matchID or tournamentID
    PlayerID  uuid.UUID
    // ... existing fields
}
```

The Hub maintains separate maps:
- `matches map[uuid.UUID]map[uuid.UUID]*Client` (existing)
- `tournaments map[uuid.UUID]map[uuid.UUID]*Client` (new)

### Alternatives Considered
1. **Separate PreLobbyHub**: Rejected - duplicates connection management, heartbeat logic
2. **Generic Room abstraction**: Considered but adds complexity; discriminated union is simpler
3. **Single unified map with composite key**: Rejected - harder to iterate by room type

---

## 2. Competition Service Integration

### Decision
Matchmaking service calls competition service REST API to verify tournament registration. Cache registration status for 30 seconds to handle temporary inconsistencies.

### Rationale
Per constitution (III.5 Bounded Contexts), services communicate via events for domain operations. However, authorization checks are cross-cutting concerns that can use synchronous REST calls. The competition service already exposes `GET /v1/{hostId}/tournaments/{tournamentId}/participants/{userId}` for registration status.

### Implementation Approach
```go
type CompetitionClient interface {
    GetParticipantStatus(ctx context.Context, tournamentID, userID uuid.UUID) (string, error)
    GetTournamentInfo(ctx context.Context, tournamentID uuid.UUID) (*TournamentInfo, error)
}
```

Internal HTTP client with:
- Base URL from config: `COMPETITION_SERVICE_URL`
- 500ms timeout per request
- 30-second cache using sync.Map or Redis

### Alternatives Considered
1. **Redis cache of all registrations**: Rejected - requires competition service to publish all registration changes
2. **JWT claims with registration status**: Rejected - tokens issued before registration wouldn't have status
3. **Event-sourced registration projection**: Over-engineering for this use case

---

## 3. Grace Period Timer Implementation

### Decision
Use `time.AfterFunc` for grace period countdown with state persisted to PostgreSQL. Timer recreated on service restart by checking active grace periods.

### Rationale
Go's `time.AfterFunc` provides accurate timing with minimal overhead. Persisting grace period state (start_time, end_time, status) to database ensures recovery after service restart. The 30-second duration is short enough that minor timing variations (±1s) are acceptable.

### Implementation Approach
```go
type GracePeriodManager struct {
    activeTimers map[uuid.UUID]*time.Timer
    mu           sync.Mutex
    repo         PreLobbyRepository
    onComplete   func(tournamentID uuid.UUID)
}

func (m *GracePeriodManager) StartGracePeriod(tournamentID uuid.UUID) error {
    // 1. Persist to DB: status='grace_period', end_time=now()+30s
    // 2. Start timer: time.AfterFunc(30*time.Second, callback)
    // 3. Store timer reference for cancellation
}
```

On service startup:
```go
func (m *GracePeriodManager) RecoverActiveGracePeriods(ctx context.Context) error {
    // Query DB for tournaments with status='grace_period' and end_time > now()
    // For each, calculate remaining duration and start timer
}
```

### Alternatives Considered
1. **Cron-style scheduler**: Over-engineering for single 30s timer
2. **Redis TTL with keyspace notifications**: Adds external dependency complexity
3. **Client-side countdown only**: Rejected - server must be authoritative for bracket generation

---

## 4. Participant Snapshot Schema

### Decision
Store participant snapshots in a dedicated `prelobby_participant_snapshots` table with JSONB array for participant list. Snapshot is immutable once created.

### Rationale
JSONB array provides:
- Atomic write of entire participant list
- Efficient retrieval without joins
- Schema flexibility for future participant metadata
- Immutability guarantee (no updates after creation)

### Implementation Approach
```sql
CREATE TABLE prelobby_participant_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL UNIQUE,
    participants JSONB NOT NULL,  -- [{user_id, display_name, avatar_url}]
    participant_count INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

The snapshot is created exactly once when grace period ends:
```go
func (s *PreLobbyService) FinalizeGracePeriod(ctx context.Context, tournamentID uuid.UUID) error {
    // 1. Get current connected participants from Hub
    // 2. Create snapshot in DB
    // 3. Update pre-lobby status to 'generating_bracket'
    // 4. Trigger bracket generation with snapshot
}
```

### Alternatives Considered
1. **Normalized table with foreign keys**: Rejected - adds join overhead, doesn't provide atomicity
2. **Store in tournament_brackets table**: Rejected - bracket may not exist yet, different bounded context
3. **In-memory only**: Rejected - doesn't survive restart, no audit trail

---

## 5. Activity Feed Storage

### Decision
Store activity feed in `prelobby_activity_feed` table with per-tournament partitioning via tournament_id. Limit to 20 most recent entries per tournament in API response.

### Rationale
Persistent storage enables:
- Historical audit of pre-lobby events
- Recovery of feed after service restart
- Consistent feed across multiple server instances

### Implementation Approach
```sql
CREATE TABLE prelobby_activity_feed (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    participant_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activity_feed_tournament_time 
    ON prelobby_activity_feed(tournament_id, created_at DESC);
```

Query for recent events:
```sql
SELECT * FROM prelobby_activity_feed 
WHERE tournament_id = $1 
ORDER BY created_at DESC 
LIMIT 20;
```

### Alternatives Considered
1. **Redis list with LTRIM**: Rejected - doesn't persist across Redis restarts
2. **In-memory ring buffer**: Rejected - doesn't share across instances
3. **Event sourcing**: Over-engineering for simple activity log

---

## 6. Pre-Lobby State Machine

### Decision
Implement pre-lobby status as a finite state machine with explicit transitions.

### States
```
waiting → grace_period → generating_bracket → started
    ↓          ↓
cancelled  cancelled
```

### Transitions
| From | To | Trigger |
|------|-----|---------|
| waiting | grace_period | TournamentStarted event received |
| waiting | cancelled | Tournament cancelled or insufficient participants |
| grace_period | generating_bracket | 30s timer completes with ≥ min participants |
| grace_period | cancelled | Participant count drops to 0 or below minimum |
| generating_bracket | started | Bracket generation completes |

### Implementation
```go
type PreLobbyStatus string

const (
    PreLobbyStatusWaiting           PreLobbyStatus = "waiting"
    PreLobbyStatusGracePeriod       PreLobbyStatus = "grace_period"
    PreLobbyStatusGeneratingBracket PreLobbyStatus = "generating_bracket"
    PreLobbyStatusStarted           PreLobbyStatus = "started"
    PreLobbyStatusCancelled         PreLobbyStatus = "cancelled"
)

func (s PreLobbyStatus) CanTransitionTo(target PreLobbyStatus) bool {
    transitions := map[PreLobbyStatus][]PreLobbyStatus{
        PreLobbyStatusWaiting:           {PreLobbyStatusGracePeriod, PreLobbyStatusCancelled},
        PreLobbyStatusGracePeriod:       {PreLobbyStatusGeneratingBracket, PreLobbyStatusCancelled},
        PreLobbyStatusGeneratingBracket: {PreLobbyStatusStarted},
    }
    for _, allowed := range transitions[s] {
        if allowed == target {
            return true
        }
    }
    return false
}
```

---

## Summary

All research topics resolved. No NEEDS CLARIFICATION items remain.

| Topic | Decision | Confidence |
|-------|----------|------------|
| WebSocket Hub Extension | Discriminated union with RoomType | High |
| Competition Integration | REST API with 30s cache | High |
| Grace Period Timer | time.AfterFunc + DB persistence | High |
| Participant Snapshot | JSONB array in dedicated table | High |
| Activity Feed | PostgreSQL table with index | High |
| State Machine | Explicit FSM with transitions | High |

