# Data Model: Tournament Matchmaking

**Feature**: 001-tournament-matchmaking  
**Phase**: 1 (Design)  
**Date**: 2025-12-04

## Overview

Three new tables in the independent `matchmaking_db` PostgreSQL database:
1. **tournament_brackets** - Bracket structure for each tournament
2. **tournament_rounds** - Rounds within a bracket (Quarter-finals, Semi-finals, Final)
3. **tournament_matches** - Individual 1v1 matches with participants and results

**Database Separation**: This is a separate database from `competition_db`. Tournament IDs are stored as UUIDs for reference but there are NO foreign key constraints across databases. Data consistency is maintained via event-driven architecture (Redis Pub/Sub).

## Entity Relationships

```
┌─────────────────────────┐
│ competition_db          │
│ - tournaments (existing)│ ---(event: TournamentStarted)--→
│ - users (existing)      │
└─────────────────────────┘

┌──────────────────────────────────────┐
│ matchmaking_db                       │
│                                      │
│ tournament_brackets (NEW)            │
│     ↓ 1:N                            │
│ tournament_rounds (NEW)              │
│     ↓ 1:N                            │
│ tournament_matches (NEW)             │
│   (stores tournament_id, user IDs    │
│    as UUIDs, no FK constraints)      │
└──────────────────────────────────────┘
```

**Cross-Database References**: `tournament_id` and `participant_id` fields store UUIDs but do NOT have foreign key constraints. Services maintain referential integrity via domain events.

---

## Table: `tournament_brackets`

**Purpose**: Stores the bracket structure for each tournament. One bracket per tournament.

### Schema

```sql
CREATE TABLE tournament_brackets (
    -- Primary key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Tournament reference (UUID only, no FK - tournaments table in different DB)
    tournament_id UUID NOT NULL UNIQUE,
    
    -- Bracket metadata
    total_rounds INTEGER NOT NULL CHECK (total_rounds > 0),
    total_matches INTEGER NOT NULL CHECK (total_matches > 0),
    byes_count INTEGER NOT NULL DEFAULT 0 CHECK (byes_count >= 0),
    
    -- Timestamps
    generated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_brackets_tournament_id ON tournament_brackets(tournament_id);

-- Note: No foreign key constraint on tournament_id because tournaments 
-- table exists in competition_db (separate database)
```

### Fields

| Field | Type | Nullable | Description |
|-------|------|----------|-------------|
| `id` | UUID | No | Primary key |
| `tournament_id` | UUID | No | Foreign key to `tournaments` (UNIQUE - one bracket per tournament) |
| `total_rounds` | INTEGER | No | Number of rounds (e.g., 3 for 8 players = QF, SF, F) |
| `total_matches` | INTEGER | No | Total matches across all rounds |
| `byes_count` | INTEGER | No | Number of byes assigned (for odd participant counts) |
| `generated_at` | TIMESTAMP | No | When bracket was generated |

### Validation Rules

- **FR-006**: `total_rounds` calculated as `CEIL(LOG2(participant_count))`
- **FR-007**: `byes_count = (2^total_rounds) - participant_count`
- **Event-Driven Integrity**: Bracket creation is triggered by `TournamentStarted` event from competition service. Application layer validates tournament existence via event payload.

### Example Data

```sql
-- Tournament with 5 participants (non-power-of-2)
INSERT INTO tournament_brackets VALUES (
    'bracket-uuid-1',
    'tournament-uuid-1',
    3,  -- total_rounds (Round 1, Semi-finals, Final)
    4,  -- total_matches (1 in R1, 2 in SF, 1 in F)
    3,  -- byes_count (8 slots - 5 participants)
    '2025-12-04 10:00:00'
);
```

---

## Table: `tournament_rounds`

**Purpose**: Represents rounds within a bracket (e.g., Quarter-finals, Semi-finals, Final).

### Schema

```sql
CREATE TABLE tournament_rounds (
    -- Primary key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Foreign key to bracket
    bracket_id UUID NOT NULL REFERENCES tournament_brackets(id) ON DELETE CASCADE,
    
    -- Round metadata
    round_number INTEGER NOT NULL CHECK (round_number > 0),
    round_name VARCHAR(100) NOT NULL, -- "Round of 16", "Quarter-finals", "Semi-finals", "Final"
    matches_count INTEGER NOT NULL CHECK (matches_count > 0),
    
    -- State tracking
    status VARCHAR(20) NOT NULL DEFAULT 'pending' 
        CHECK (status IN ('pending', 'in_progress', 'completed')),
    
    -- Timestamps
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    
    -- Constraints
    CONSTRAINT fk_round_bracket FOREIGN KEY (bracket_id) 
        REFERENCES tournament_brackets(id) ON DELETE CASCADE,
    CONSTRAINT unique_bracket_round_number UNIQUE (bracket_id, round_number)
);

CREATE INDEX idx_rounds_bracket_id ON tournament_rounds(bracket_id);
CREATE INDEX idx_rounds_status ON tournament_rounds(status) WHERE status = 'in_progress';
```

### Fields

| Field | Type | Nullable | Description |
|-------|------|----------|-------------|
| `id` | UUID | No | Primary key |
| `bracket_id` | UUID | No | Foreign key to `tournament_brackets` |
| `round_number` | INTEGER | No | Sequential round number (1, 2, 3...) |
| `round_name` | VARCHAR(100) | No | Human-readable name (e.g., "Semi-finals") |
| `matches_count` | INTEGER | No | Number of matches in this round |
| `status` | VARCHAR(20) | No | `pending`, `in_progress`, `completed` |
| `started_at` | TIMESTAMP | Yes | When first match of round started |
| `completed_at` | TIMESTAMP | Yes | When all matches completed |

### Validation Rules

- **FR-009**: `round_number` must be unique within bracket
- **FR-225**: Previous round must be `completed` before next round can start (for sequential rounds)
- **Round Names**: Auto-generated based on `matches_count`:
  - 1 match → "Final"
  - 2 matches → "Semi-finals"
  - 4 matches → "Quarter-finals"
  - 8+ matches → "Round of X"

### Example Data

```sql
-- 8-player tournament rounds
INSERT INTO tournament_rounds VALUES
    ('round-1', 'bracket-uuid-1', 1, 'Quarter-finals', 4, 'completed', '2025-12-04 10:00', '2025-12-04 10:30'),
    ('round-2', 'bracket-uuid-1', 2, 'Semi-finals', 2, 'in_progress', '2025-12-04 10:31', NULL),
    ('round-3', 'bracket-uuid-1', 3, 'Final', 1, 'pending', NULL, NULL);
```

---

## Table: `tournament_matches`

**Purpose**: Individual 1v1 matches between participants. Core entity for match state management.

### Schema

```sql
CREATE TABLE tournament_matches (
    -- Primary key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Foreign key to round (same database)
    round_id UUID NOT NULL REFERENCES tournament_rounds(id) ON DELETE CASCADE,
    
    -- Match metadata
    match_number INTEGER NOT NULL CHECK (match_number > 0),
    next_match_id UUID REFERENCES tournament_matches(id), -- NULL for final
    
    -- Participants (UUIDs only, no FK - users table in different DB)
    -- participant2_id NULL = bye for participant1
    participant1_id UUID NOT NULL,
    participant2_id UUID, -- Nullable for byes
    
    -- Match state
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN (
        'pending',     -- Created, waiting for players
        'ready',       -- Both players ready
        'started',     -- Match in progress
        'paused',      -- Disconnection pause (CS:GO style)
        'completed',   -- Match finished
        'cancelled'    -- Host cancelled
    )),
    
    -- Ready state (per player)
    participant1_ready BOOLEAN NOT NULL DEFAULT false,
    participant2_ready BOOLEAN NOT NULL DEFAULT false,
    
    -- Results
    winner_id UUID, -- Must be participant1_id or participant2_id (checked in app layer)
    score_player1 INTEGER CHECK (score_player1 >= 0),
    score_player2 INTEGER CHECK (score_player2 >= 0),
    result_source VARCHAR(20) CHECK (result_source IN ('game_api', 'host_manual', 'walkover')),
    
    -- Disconnect tracking (CS:GO style)
    disconnected_player_id UUID, -- Must be participant1_id or participant2_id
    disconnected_at TIMESTAMP,
    
    -- Game API integration
    game_api_match_id VARCHAR(255),  -- External game match ID
    game_api_poll_attempts INTEGER DEFAULT 0,
    game_api_last_poll TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT fk_match_round FOREIGN KEY (round_id) 
        REFERENCES tournament_rounds(id) ON DELETE CASCADE,
    CONSTRAINT fk_match_next FOREIGN KEY (next_match_id) 
        REFERENCES tournament_matches(id) ON DELETE SET NULL,
    CONSTRAINT check_completed_has_winner CHECK (
        (status = 'completed' AND winner_id IS NOT NULL) OR status != 'completed'
    ),
    CONSTRAINT unique_round_match_number UNIQUE (round_id, match_number)
);

-- Note: No foreign keys on participant1_id, participant2_id, winner_id, 
-- or disconnected_player_id because users table exists in competition_db 
-- (separate database). Application layer validates user existence.

-- Indexes
CREATE INDEX idx_matches_round_id ON tournament_matches(round_id);
CREATE INDEX idx_matches_next_match_id ON tournament_matches(next_match_id) 
    WHERE next_match_id IS NOT NULL;
CREATE INDEX idx_matches_participants ON tournament_matches(participant1_id, participant2_id)
    WHERE status IN ('pending', 'ready', 'started');
CREATE INDEX idx_matches_game_api_pending ON tournament_matches(game_api_match_id, game_api_last_poll)
    WHERE status = 'started' AND result_source IS NULL;
CREATE INDEX idx_matches_status ON tournament_matches(status);
```

### Fields

| Field | Type | Nullable | Description |
|-------|------|----------|-------------|
| `id` | UUID | No | Primary key |
| `round_id` | UUID | No | Foreign key to `tournament_rounds` |
| `match_number` | INTEGER | No | Sequential number within round (1, 2, 3...) |
| `next_match_id` | UUID | Yes | Winner advances to this match (NULL for final) |
| `participant1_id` | UUID | No | First player (always present) |
| `participant2_id` | UUID | Yes | Second player (NULL = bye for participant1) |
| `status` | VARCHAR(20) | No | Match state (pending/ready/started/paused/completed/cancelled) |
| `participant1_ready` | BOOLEAN | No | Player 1 marked ready in lobby |
| `participant2_ready` | BOOLEAN | No | Player 2 marked ready in lobby |
| `winner_id` | UUID | Yes | Winner (set when completed) |
| `score_player1` | INTEGER | Yes | Score for player 1 (round-based) |
| `score_player2` | INTEGER | Yes | Score for player 2 (round-based) |
| `result_source` | VARCHAR(20) | Yes | How result was determined (game_api/host_manual/walkover) |
| `disconnected_player_id` | UUID | Yes | Player who disconnected (CS:GO style tracking) |
| `disconnected_at` | TIMESTAMP | Yes | When disconnect occurred (for 30s window) |
| `game_api_match_id` | VARCHAR(255) | Yes | External game system match ID |
| `game_api_poll_attempts` | INTEGER | No | Number of API polling attempts |
| `game_api_last_poll` | TIMESTAMP | Yes | Last time API was polled |
| `created_at` | TIMESTAMP | No | Match created timestamp |
| `started_at` | TIMESTAMP | Yes | Match started timestamp |
| `completed_at` | TIMESTAMP | Yes | Match completed timestamp |
| `updated_at` | TIMESTAMP | No | Last update timestamp |

### State Machine

```
pending → ready → started → completed
             ↓        ↓
          cancelled  paused → started (reconnect within 30s)
                              ↓
                           completed (walkover if no reconnect)
```

### Validation Rules

- **FR-011**: Status must follow valid state transitions
- **FR-012**: Both participant IDs must exist (or participant2 NULL for bye)
- **FR-013**: Completed matches must have `winner_id`, `score_player1`, `score_player2`
- **FR-014**: Winner must be one of the participants
- **FR-016b**: Disconnect awards 1 point to online player
- **FR-016c**: 30s reconnect window enforced in application layer

### Example Data

```sql
-- Normal match (both players)
INSERT INTO tournament_matches VALUES (
    'match-1',
    'round-1',
    1,
    'match-5', -- Winner goes to semi-final match 5
    'player-a-uuid',
    'player-b-uuid',
    'completed',
    true, true,  -- Both were ready
    'player-a-uuid',  -- Winner
    3, 1,  -- Score 3-1
    'game_api',
    NULL, NULL,  -- No disconnects
    'external-game-match-123',
    1,  -- Polled once
    '2025-12-04 10:15:00',
    '2025-12-04 10:00:00',
    '2025-12-04 10:05:00',
    '2025-12-04 10:20:00',
    '2025-12-04 10:20:00'
);

-- Bye match (only participant1)
INSERT INTO tournament_matches VALUES (
    'match-2',
    'round-1',
    2,
    'match-5', -- Advances to same semi-final
    'player-c-uuid',
    NULL,  -- Bye (no opponent)
    'completed',
    true, false,  -- Only player1 ready
    'player-c-uuid',  -- Auto-winner
    0, 0,  -- No score for bye
    'walkover',
    NULL, NULL,
    NULL,
    0, NULL,
    '2025-12-04 10:00:00',
    '2025-12-04 10:00:00',
    '2025-12-04 10:00:00',
    '2025-12-04 10:00:00'
);
```

---

## Indexes Strategy

### Performance-Critical Queries

1. **Get bracket for tournament** (bracket visualization)
   - Primary: `idx_brackets_tournament_id` on `tournament_brackets(tournament_id)`
   - Supporting: `idx_rounds_bracket_id`, `idx_matches_round_id`

2. **Find active match for player** (lobby view)
   - Composite: `idx_matches_participants` on `(participant1_id, participant2_id)` WHERE status IN ('pending', 'ready', 'started')

3. **Poll game API for results** (background job)
   - Composite: `idx_matches_game_api_pending` on `(game_api_match_id, game_api_last_poll)` WHERE status = 'started'

4. **Match progression** (winner advances)
   - Foreign key: `idx_matches_next_match_id` on `next_match_id`

### Query Plan Estimates

```sql
-- Bracket visualization (hot path)
EXPLAIN ANALYZE
SELECT b.*, r.*, m.*
FROM tournament_brackets b
JOIN tournament_rounds r ON r.bracket_id = b.id
JOIN tournament_matches m ON m.round_id = r.id
WHERE b.tournament_id = 'tournament-uuid';

-- Performance benchmark example: <50ms for large tournament with 128 players (7 rounds, 127 matches)
```

---

## Migrations

### Migration Files

```
services/competition/migrations/
├── 000008_create_brackets_table.sql
├── 000009_create_rounds_table.sql
└── 000010_create_matches_table.sql
```

### Migration: 000008_create_brackets_table.sql

```sql
-- +goose Up
CREATE TABLE tournament_brackets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL UNIQUE REFERENCES tournaments(id) ON DELETE CASCADE,
    total_rounds INTEGER NOT NULL CHECK (total_rounds > 0),
    total_matches INTEGER NOT NULL CHECK (total_matches > 0),
    byes_count INTEGER NOT NULL DEFAULT 0 CHECK (byes_count >= 0),
    generated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_brackets_tournament_id ON tournament_brackets(tournament_id);

-- +goose Down
DROP TABLE IF EXISTS tournament_brackets CASCADE;
```

### Migration: 000009_create_rounds_table.sql

```sql
-- +goose Up
CREATE TABLE tournament_rounds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bracket_id UUID NOT NULL REFERENCES tournament_brackets(id) ON DELETE CASCADE,
    round_number INTEGER NOT NULL CHECK (round_number > 0),
    round_name VARCHAR(100) NOT NULL,
    matches_count INTEGER NOT NULL CHECK (matches_count > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed')),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    CONSTRAINT unique_bracket_round_number UNIQUE (bracket_id, round_number)
);

CREATE INDEX idx_rounds_bracket_id ON tournament_rounds(bracket_id);
CREATE INDEX idx_rounds_status ON tournament_rounds(status) WHERE status = 'in_progress';

-- +goose Down
DROP TABLE IF EXISTS tournament_rounds CASCADE;
```

### Migration: 000010_create_matches_table.sql

```sql
-- +goose Up
CREATE TABLE tournament_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    round_id UUID NOT NULL REFERENCES tournament_rounds(id) ON DELETE CASCADE,
    match_number INTEGER NOT NULL CHECK (match_number > 0),
    next_match_id UUID REFERENCES tournament_matches(id),
    participant1_id UUID NOT NULL REFERENCES users(id),
    participant2_id UUID REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'ready', 'started', 'paused', 'completed', 'cancelled')),
    participant1_ready BOOLEAN NOT NULL DEFAULT false,
    participant2_ready BOOLEAN NOT NULL DEFAULT false,
    winner_id UUID REFERENCES users(id),
    score_player1 INTEGER CHECK (score_player1 >= 0),
    score_player2 INTEGER CHECK (score_player2 >= 0),
    result_source VARCHAR(20) CHECK (result_source IN ('game_api', 'host_manual', 'walkover')),
    disconnected_player_id UUID REFERENCES users(id),
    disconnected_at TIMESTAMP,
    game_api_match_id VARCHAR(255),
    game_api_poll_attempts INTEGER DEFAULT 0,
    game_api_last_poll TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT check_winner_is_participant CHECK (winner_id IS NULL OR winner_id IN (participant1_id, participant2_id)),
    CONSTRAINT check_completed_has_winner CHECK ((status = 'completed' AND winner_id IS NOT NULL) OR status != 'completed'),
    CONSTRAINT unique_round_match_number UNIQUE (round_id, match_number)
);

CREATE INDEX idx_matches_round_id ON tournament_matches(round_id);
CREATE INDEX idx_matches_next_match_id ON tournament_matches(next_match_id) WHERE next_match_id IS NOT NULL;
CREATE INDEX idx_matches_participants ON tournament_matches(participant1_id, participant2_id) WHERE status IN ('pending', 'ready', 'started');
CREATE INDEX idx_matches_game_api_pending ON tournament_matches(game_api_match_id, game_api_last_poll) WHERE status = 'started' AND result_source IS NULL;
CREATE INDEX idx_matches_status ON tournament_matches(status);

-- +goose Down
DROP TABLE IF EXISTS tournament_matches CASCADE;
```

---

## Summary

### Tables Added

| Table | Purpose | Rows (est. for 64-player tournament) |
|-------|---------|--------------------------------------|
| `tournament_brackets` | Bracket structure | 1 per tournament |
| `tournament_rounds` | Rounds (QF, SF, F) | 6 rounds |
| `tournament_matches` | Individual matches | 63 matches |

### Foreign Key Relationships

```
tournaments (1) ←→ (1) tournament_brackets
tournament_brackets (1) ←→ (N) tournament_rounds
tournament_rounds (1) ←→ (N) tournament_matches
tournament_matches (N) ←→ (1) tournament_matches (self-ref for progression)
tournament_matches (N) ←→ (1) users (participants, winner)
```

### Storage Estimates

- **Per Tournament** (64 players): ~10 KB (bracket + rounds + matches)
- **1000 Tournaments**: ~10 MB
- **With indexes**: ~20 MB total

All data model decisions finalized. Ready for contracts generation.

