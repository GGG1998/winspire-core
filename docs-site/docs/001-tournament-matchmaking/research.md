# Research: Tournament Matchmaking Implementation

**Feature**: 001-tournament-matchmaking  
**Phase**: 0 (Research & Technology Decisions)  
**Date**: 2025-12-04

## 1. Bracket Generation Algorithms

### Decision: Balanced Binary Tree with Top-Seeded Byes

**Rationale**:
- Single elimination tournament is naturally a binary tree structure
- Each round halves the number of participants
- Byes should be placed in early rounds to maintain bracket balance

**Algorithm**:

```go
// Calculate number of rounds needed
rounds := int(math.Ceil(math.Log2(float64(participantCount))))

// Calculate number of byes needed
totalSlots := int(math.Pow(2, float64(rounds)))
byesNeeded := totalSlots - participantCount

// Assign byes to top seeds (first byesNeeded positions)
// Remaining participants play in Round 1
```

**Match Numbering Scheme**:
- Round 1 (Quarter-finals): Matches 1-4
- Round 2 (Semi-finals): Matches 5-6  
- Round 3 (Final): Match 7
- Formula: `matchNumber = previousRoundMatchCount + positionInRound`

**Progression Tracking**:
- Each match has `next_match_id` foreign key
- Winner automatically assigned to `participant_1` or `participant_2` slot of next match
- Formula: `nextMatchId = currentMatchId / 2` (in binary tree structure)

**Alternatives Considered**:
- ❌ Swiss System - Too complex for MVP, requires multiple rounds of pairings
- ❌ Double Elimination - Requires losers bracket, significantly increases match count
- ❌ Round Robin - Not elimination-based, all players play all others

**References**:
- Challonge bracket algorithm: https://challonge.com/tournament_generator
- Binary tree traversal for match progression

---

## 2. Game API Integration Patterns

### Decision: Polling with Exponential Backoff + Circuit Breaker

**Rationale**:
- Push (webhooks) requires game API to implement callbacks (not guaranteed)
- Polling is more reliable and works with any game API
- Circuit breaker prevents cascading failures when API is down

**Implementation Pattern**:

```go
type GameAPIClient struct {
    httpClient    *http.Client
    circuitBreaker *gobreaker.CircuitBreaker
    pollInterval   time.Duration // 5 seconds default
}

// Poll for result after match starts
func (c *GameAPIClient) PollMatchResult(matchID string, timeout time.Duration) (*MatchResult, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    ticker := time.NewTicker(c.pollInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            result, err := c.circuitBreaker.Execute(func() (interface{}, error) {
                return c.fetchResult(matchID)
            })
            if err == nil && result != nil {
                return result.(*MatchResult), nil
            }
        case <-ctx.Done():
            return nil, fmt.Errorf("polling timeout after %v", timeout)
        }
    }
}
```

**Retry Strategy**:
- Initial poll: 5s after match starts
- Retry interval: 5s (constant, not exponential - match results should be available quickly)
- Total timeout: 60s before marking as "needs manual entry"
- Circuit breaker: Open after 5 consecutive failures, half-open after 30s

**Fallback Path**:
- After timeout → Match flagged for manual host entry
- Host receives notification with match details
- Host enters result via admin panel
- System logs "API_TIMEOUT" event for monitoring

**API Contract Assumptions**:

```yaml
# Expected Game API endpoint
GET /api/matches/{match_id}/result
Authorization: Bearer {api_key}

Response:
{
  "match_id": "uuid",
  "status": "completed",
  "winner_id": "player_uuid",
  "scores": {
    "player1_id": { "score": 3, "kills": 15 },
    "player2_id": { "score": 1, "kills": 8 }
  },
  "completed_at": "2025-12-04T10:30:00Z"
}
```

**Alternatives Considered**:
- ❌ Webhooks only - Requires game API support, no fallback
- ❌ Long-polling (60s timeout) - Ties up connections, doesn't work with load balancers
- ❌ Server-Sent Events - Game API unlikely to support

**Libraries**:
- `github.com/sony/gobreaker` - Circuit breaker implementation
- Standard `net/http` with custom timeout per request

---

## 3. Disconnect Detection & Reconnection

### Decision: WebSocket Heartbeat + Server-Side Timer

**Rationale**:
- WebSocket heartbeat detects disconnection within seconds
- Server-side timer ensures consistent 30s window (not client-dependent)
- Match state persisted in PostgreSQL for recovery

**Implementation Pattern**:

```go
type MatchConnection struct {
    PlayerID      string
    MatchID       string
    Conn          *websocket.Conn
    LastHeartbeat time.Time
    DisconnectedAt *time.Time
}

// Heartbeat loop (client sends every 5s)
func (m *MatchConnection) MonitorHeartbeat() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        if time.Since(m.LastHeartbeat) > 10*time.Second {
            // Missed 2 heartbeats → Consider disconnected
            m.HandleDisconnect()
            return
        }
    }
}

// Disconnect handling
func (m *MatchConnection) HandleDisconnect() {
    now := time.Now()
    m.DisconnectedAt = &now
    
    // Award 1 point to opponent (CS:GO style)
    UpdateMatchScore(m.MatchID, OpponentOf(m.PlayerID), +1)
    
    // Pause match
    SetMatchState(m.MatchID, "paused")
    
    // Start 30s reconnection timer
    time.AfterFunc(30*time.Second, func() {
        if m.DisconnectedAt != nil {
            // Player did not reconnect → Disqualify
            GrantWalkover(m.MatchID, OpponentOf(m.PlayerID))
        }
    })
}

// Reconnection handling
func (m *MatchConnection) HandleReconnect() {
    if m.DisconnectedAt != nil {
        elapsed := time.Since(*m.DisconnectedAt)
        if elapsed <= 30*time.Second {
            // Reconnected within window → Resume match
            m.DisconnectedAt = nil
            SetMatchState(m.MatchID, "started")
            NotifyOpponent(m.MatchID, "Player reconnected")
        } else {
            // Too late → Already disqualified
            return errors.New("reconnection window expired")
        }
    }
}
```

**Match State Machine**:

```
pending → ready → started → paused → started → completed
                     ↓                   ↓
                   paused           disqualified (if no reconnect)
```

**Persistence**:

```sql
ALTER TABLE tournament_matches ADD COLUMN disconnected_at TIMESTAMP;
ALTER TABLE tournament_matches ADD COLUMN disconnected_player_id UUID;
```

**Handling Simultaneous Disconnects**:
- Track `disconnected_at` per player independently
- First to disconnect: opponent gets +1 point
- Both get 30s window from their respective disconnect time
- First to reconnect continues; other player disqualified if window expires

**Alternatives Considered**:
- ❌ TCP keepalive only - Too slow (minutes), not second-precision
- ❌ Client-side timer - Exploitable, not authoritative
- ❌ Fixed 30s from match start - Unfair, doesn't account for when disconnect happened

**Libraries**:
- `github.com/gorilla/websocket` - WebSocket implementation
- Standard `time.AfterFunc` for reconnection timer

---

## 4. Redis Pub/Sub for Event Distribution

### Decision: Channel-Per-Event with JSON Serialization

**Rationale**:
- Predictable channel names enable selective subscription
- JSON is human-readable, debuggable, and language-agnostic
- At-least-once delivery acceptable (subscribers can deduplicate)

**Channel Naming Convention**:

```
Format: events:{bounded_context}:{event_name}

Examples:
- events:tournament_management:tournament_started
- events:matchmaking:bracket_generated
- events:matchmaking:match_created
- events:matchmaking:match_completed
```

**Event Payload Structure**:

```json
{
  "event_id": "01JFXYZ...",  // ULID for ordering
  "event_type": "MatchCreated",
  "aggregate_id": "match-uuid",
  "aggregate_type": "Match",
  "bounded_context": "Matchmaking",
  "timestamp": "2025-12-04T10:30:00Z",
  "payload": {
    "match_id": "uuid",
    "tournament_id": "uuid",
    "round_number": 1,
    "participant1_id": "uuid",
    "participant2_id": "uuid",
    "status": "pending"
  },
  "metadata": {
    "correlation_id": "req-xyz",
    "causation_id": "tournament_started-abc",
    "user_id": "host-uuid"
  }
}
```

**Publisher Implementation**:

```go
type EventPublisher struct {
    redisClient *redis.Client
}

func (p *EventPublisher) Publish(ctx context.Context, event DomainEvent) error {
    channel := fmt.Sprintf("events:%s:%s", 
        event.BoundedContext, 
        strings.ToLower(event.EventType))
    
    payload, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("marshal event: %w", err)
    }
    
    // Publish to Redis channel
    if err := p.redisClient.Publish(ctx, channel, payload).Err(); err != nil {
        return fmt.Errorf("publish to Redis: %w", err)
    }
    
    // Also store in events table for replay
    if err := p.storeEvent(ctx, event); err != nil {
        log.Printf("WARN: event published but not stored: %v", err)
    }
    
    return nil
}
```

**Subscriber Pattern (Realtime Service)**:

```go
func (s *RealtimeService) SubscribeToEvents(ctx context.Context) error {
    pubsub := s.redisClient.Subscribe(ctx, 
        "events:matchmaking:match_created",
        "events:matchmaking:match_started",
        "events:matchmaking:match_completed")
    
    defer pubsub.Close()
    
    for {
        select {
        case msg := <-pubsub.Channel():
            var event DomainEvent
            if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
                log.Printf("ERROR: unmarshal event: %v", err)
                continue
            }
            s.handleEvent(event)
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}
```

**Reliability Guarantees**:
- **At-least-once delivery**: Redis Pub/Sub is fire-and-forget
- **Event store**: All events persisted in `events` table for replay
- **Idempotency**: Subscribers must handle duplicate events (use `event_id` for deduplication)
- **Ordering**: Events ordered by `timestamp` + `event_id` (ULID includes timestamp)

**Recovery Mechanism**:
If realtime service crashes/restarts:
1. Query `events` table for missed events since last processed event
2. Replay missed events in order
3. Resume Redis subscription from current time

**Alternatives Considered**:
- ❌ Single channel for all events - Forces all subscribers to filter, wastes bandwidth
- ❌ Protobuf serialization - Overkill for MVP, harder to debug
- ❌ RabbitMQ/Kafka - More infrastructure, not needed for MVP scale

**Libraries**:
- `github.com/redis/go-redis/v9` - Redis client with Pub/Sub support
- `github.com/oklog/ulid` - ULID generation for event IDs

---

## 5. SQLC Best Practices

### Decision: One SQL File Per Aggregate + Complex Queries in Views

**Rationale**:
- Logical organization mirrors domain model
- Views pre-compute complex joins for bracket visualization
- SQLC generates type-safe Go code from SQL
- Transactions handled at application layer (not in SQL)

**File Organization**:

```
services/competition/internal/store/
├── queries.sql           # Existing (tournaments, participants)
├── bracket_queries.sql   # NEW - Bracket aggregate
├── round_queries.sql     # NEW - Round queries
└── match_queries.sql     # NEW - Match aggregate
```

**Example: bracket_queries.sql**

```sql
-- name: CreateBracket :one
INSERT INTO tournament_brackets (
    tournament_id, total_rounds, total_matches, generated_at
) VALUES (
    $1, $2, $3, NOW()
) RETURNING *;

-- name: GetBracketByTournamentID :one
SELECT * FROM tournament_brackets
WHERE tournament_id = $1;

-- name: GetBracketWithRoundsAndMatches :many
-- Complex join query for bracket visualization
SELECT 
    b.id as bracket_id,
    b.tournament_id,
    r.id as round_id,
    r.round_number,
    r.round_name,
    m.id as match_id,
    m.match_number,
    m.participant1_id,
    m.participant2_id,
    m.status,
    m.winner_id,
    m.score_player1,
    m.score_player2,
    p1.display_name as participant1_name,
    p2.display_name as participant2_name
FROM tournament_brackets b
JOIN tournament_rounds r ON r.bracket_id = b.id
JOIN tournament_matches m ON m.round_id = r.id
LEFT JOIN tournament_registrations p1 ON p1.user_id = m.participant1_id AND p1.tournament_id = b.tournament_id
LEFT JOIN tournament_registrations p2 ON p2.user_id = m.participant2_id AND p2.tournament_id = b.tournament_id
WHERE b.tournament_id = $1
ORDER BY r.round_number ASC, m.match_number ASC;
```

**Example: match_queries.sql**

```sql
-- name: CreateMatch :one
INSERT INTO tournament_matches (
    round_id, match_number, participant1_id, participant2_id,
    status, next_match_id
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: UpdateMatchStatus :exec
UPDATE tournament_matches 
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateMatchResult :exec
UPDATE tournament_matches
SET 
    winner_id = $2,
    score_player1 = $3,
    score_player2 = $4,
    status = 'completed',
    completed_at = NOW()
WHERE id = $1;

-- name: GetMatchByID :one
SELECT * FROM tournament_matches
WHERE id = $1;

-- name: GetMatchesForRound :many
SELECT m.*, 
    p1.display_name as participant1_name,
    p2.display_name as participant2_name
FROM tournament_matches m
LEFT JOIN tournament_registrations p1 ON p1.user_id = m.participant1_id
LEFT JOIN tournament_registrations p2 ON p2.user_id = m.participant2_id
WHERE m.round_id = $1
ORDER BY m.match_number ASC;
```

**Transaction Handling**:

```go
// Application layer handles transactions
func (s *BracketService) StartTournament(ctx context.Context, tournamentID string) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    qtx := s.queries.WithTx(tx)
    
    // 1. Update tournament status
    if err := qtx.UpdateTournamentStatus(ctx, tournamentID, "started"); err != nil {
        return err
    }
    
    // 2. Generate bracket
    bracket, err := qtx.CreateBracket(ctx, ...)
    if err != nil {
        return err
    }
    
    // 3. Create rounds
    for _, round := range rounds {
        if _, err := qtx.CreateRound(ctx, round); err != nil {
            return err
        }
    }
    
    // 4. Create matches
    for _, match := range matches {
        if _, err := qtx.CreateMatch(ctx, match); err != nil {
            return err
        }
    }
    
    return tx.Commit()
}
```

**Nullable Foreign Keys for Byes**:

```sql
-- participant2_id is nullable for byes
CREATE TABLE tournament_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    participant1_id UUID NOT NULL REFERENCES users(id),
    participant2_id UUID REFERENCES users(id),  -- NULL for byes
    ...
);

-- SQLC handles NULLs with sql.NullString
-- Generated Go type:
type Match struct {
    ID             string
    Participant1ID string
    Participant2ID sql.NullString  // NULL if bye
    ...
}
```

**Indexes for Performance**:

```sql
-- Bracket lookup by tournament
CREATE INDEX idx_brackets_tournament_id ON tournament_brackets(tournament_id);

-- Rounds by bracket
CREATE INDEX idx_rounds_bracket_id ON tournament_rounds(bracket_id);

-- Matches by round
CREATE INDEX idx_matches_round_id ON tournament_matches(round_id);

-- Match progression (find next match)
CREATE INDEX idx_matches_next_match_id ON tournament_matches(next_match_id);

-- Active matches for player (lobby view)
CREATE INDEX idx_matches_participants ON tournament_matches(participant1_id, participant2_id)
WHERE status IN ('pending', 'ready', 'started');
```

**Alternatives Considered**:
- ❌ ORM (GORM) - Less type-safe, performance overhead, complex queries hard to optimize
- ❌ Raw SQL strings - No type safety, manual scanning, error-prone
- ❌ All queries in one file - Hard to navigate, merge conflicts

**References**:
- SQLC documentation: https://docs.sqlc.dev/
- Go database/sql patterns: https://go.dev/doc/database/

---

## Summary of Decisions

| Research Area | Decision | Key Benefit |
|---------------|----------|-------------|
| Bracket Generation | Balanced binary tree with top-seeded byes | Simple algorithm, fair distribution |
| Game API Integration | Polling + circuit breaker | Reliable, works with any API |
| Disconnect Handling | WebSocket heartbeat + server-side timer | CS:GO-style reconnection, fair enforcement |
| Event Distribution | Redis Pub/Sub with JSON | Low-latency, simple, scalable |
| Database Queries | SQLC with aggregate-per-file | Type-safe, maintainable, performant |

All research questions resolved. Ready for Phase 1 design artifacts.













