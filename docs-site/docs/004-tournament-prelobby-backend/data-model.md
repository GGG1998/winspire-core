# Data Model: Tournament Pre-Lobby Backend

**Feature**: 004-tournament-prelobby-backend  
**Date**: 2025-12-05

## Overview

This document defines the data model for the tournament pre-lobby feature. The model follows Domain-Driven Design principles with the PreLobby as the primary aggregate.

## Entities

### PreLobby (Aggregate Root)

Represents the waiting room state for a tournament before it starts.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK, auto-generated | Unique identifier |
| tournament_id | UUID | NOT NULL, UNIQUE | Reference to tournament in competition service |
| status | VARCHAR(30) | NOT NULL | Current state: waiting, grace_period, generating_bracket, started, cancelled |
| grace_period_start | TIMESTAMP | NULL | When grace period began |
| grace_period_end | TIMESTAMP | NULL | When grace period ends (start + 30s) |
| min_participants | INTEGER | NOT NULL, DEFAULT 2 | Minimum participants required to start |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | When pre-lobby was created |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | Last modification time |

**Status Values**:
- `waiting`: Pre-lobby is open, accepting connections, waiting for tournament start
- `grace_period`: 30-second window after tournament start trigger
- `generating_bracket`: Bracket generation in progress
- `started`: Tournament has started, pre-lobby closed
- `cancelled`: Tournament start was cancelled (insufficient participants)

**Business Rules**:
- Only one pre-lobby per tournament
- Status transitions follow defined state machine
- Grace period is exactly 30 seconds

---

### PreLobbyParticipantSnapshot

Immutable record of participants present when grace period ended.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK, auto-generated | Unique identifier |
| tournament_id | UUID | NOT NULL, UNIQUE, FK→prelobbies | Tournament reference |
| participants | JSONB | NOT NULL | Array of participant objects |
| participant_count | INTEGER | NOT NULL | Total count for quick access |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | Snapshot creation time |

**Participants JSONB Structure**:
```json
[
  {
    "user_id": "uuid",
    "display_name": "string",
    "avatar_url": "string|null"
  }
]
```

**Business Rules**:
- Created exactly once when grace period ends
- Immutable after creation (no updates)
- Used as input for bracket generation

---

### PreLobbyActivityFeed

Chronological log of events in the pre-lobby.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK, auto-generated | Unique identifier |
| tournament_id | UUID | NOT NULL, FK→prelobbies | Tournament reference |
| event_type | VARCHAR(50) | NOT NULL | Type of event |
| message | TEXT | NOT NULL | Human-readable message |
| participant_name | VARCHAR(255) | NULL | Name of participant (if applicable) |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | Event timestamp |

**Event Types**:
- `participant_joined`: Player connected to pre-lobby
- `participant_left`: Player disconnected from pre-lobby
- `grace_period_started`: Grace period began
- `grace_period_ended`: Grace period completed
- `tournament_cancelled`: Tournament start cancelled
- `system_message`: Administrative message

**Business Rules**:
- Events are append-only
- API returns only 20 most recent events per tournament
- Ordered by created_at DESC

---

## In-Memory Structures

### Connected Participants (Hub)

Tracked in WebSocket Hub, not persisted to database.

```go
type PreLobbyConnection struct {
    TournamentID uuid.UUID
    UserID       uuid.UUID
    DisplayName  string
    AvatarURL    *string
    JoinedAt     time.Time
    Client       *websocket.Client
}
```

**Business Rules**:
- Ephemeral - lost on service restart
- Used for real-time presence tracking
- Snapshot to database when grace period ends

---

## Relationships

```
┌─────────────────────┐
│     PreLobby        │
│  (Aggregate Root)   │
├─────────────────────┤
│ id                  │
│ tournament_id (UK)  │
│ status              │
│ grace_period_start  │
│ grace_period_end    │
│ min_participants    │
└─────────┬───────────┘
          │
          │ 1:1
          ▼
┌─────────────────────────────┐
│ PreLobbyParticipantSnapshot │
├─────────────────────────────┤
│ id                          │
│ tournament_id (FK, UK)      │
│ participants (JSONB)        │
│ participant_count           │
└─────────────────────────────┘

┌─────────────────────┐
│     PreLobby        │
└─────────┬───────────┘
          │
          │ 1:N
          ▼
┌─────────────────────────┐
│ PreLobbyActivityFeed    │
├─────────────────────────┤
│ id                      │
│ tournament_id (FK)      │
│ event_type              │
│ message                 │
│ participant_name        │
│ created_at              │
└─────────────────────────┘
```

---

## Database Migration

```sql
-- Migration: 000005_create_prelobby_tables.sql

-- ============================================================================
-- PRE-LOBBY TABLE
-- ============================================================================
CREATE TABLE prelobbies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL UNIQUE,
    status VARCHAR(30) NOT NULL DEFAULT 'waiting' CHECK (
        status IN ('waiting', 'grace_period', 'generating_bracket', 'started', 'cancelled')
    ),
    grace_period_start TIMESTAMP,
    grace_period_end TIMESTAMP,
    min_participants INTEGER NOT NULL DEFAULT 2 CHECK (min_participants >= 2),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for status queries (find active pre-lobbies)
CREATE INDEX idx_prelobbies_status ON prelobbies(status) WHERE status IN ('waiting', 'grace_period');

-- ============================================================================
-- PARTICIPANT SNAPSHOT TABLE
-- ============================================================================
CREATE TABLE prelobby_participant_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL UNIQUE REFERENCES prelobbies(tournament_id) ON DELETE CASCADE,
    participants JSONB NOT NULL,
    participant_count INTEGER NOT NULL CHECK (participant_count >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- ACTIVITY FEED TABLE
-- ============================================================================
CREATE TABLE prelobby_activity_feed (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES prelobbies(tournament_id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    participant_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for efficient retrieval of recent events
CREATE INDEX idx_activity_feed_tournament_time 
    ON prelobby_activity_feed(tournament_id, created_at DESC);
```

---

## SQLC Queries

```sql
-- name: CreatePreLobby :one
INSERT INTO prelobbies (tournament_id, min_participants)
VALUES ($1, $2)
RETURNING *;

-- name: GetPreLobbyByTournament :one
SELECT * FROM prelobbies WHERE tournament_id = $1;

-- name: UpdatePreLobbyStatus :one
UPDATE prelobbies 
SET status = $2, updated_at = NOW()
WHERE tournament_id = $1
RETURNING *;

-- name: StartGracePeriod :one
UPDATE prelobbies 
SET status = 'grace_period', 
    grace_period_start = NOW(),
    grace_period_end = NOW() + INTERVAL '30 seconds',
    updated_at = NOW()
WHERE tournament_id = $1 AND status = 'waiting'
RETURNING *;

-- name: GetActiveGracePeriods :many
SELECT * FROM prelobbies 
WHERE status = 'grace_period' AND grace_period_end > NOW();

-- name: CreateParticipantSnapshot :one
INSERT INTO prelobby_participant_snapshots (tournament_id, participants, participant_count)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetParticipantSnapshot :one
SELECT * FROM prelobby_participant_snapshots WHERE tournament_id = $1;

-- name: AddActivityFeedEvent :one
INSERT INTO prelobby_activity_feed (tournament_id, event_type, message, participant_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRecentActivityFeed :many
SELECT * FROM prelobby_activity_feed 
WHERE tournament_id = $1 
ORDER BY created_at DESC 
LIMIT 20;
```

---

## Validation Rules

### PreLobby
- `tournament_id`: Must be valid UUID, must exist in competition service
- `status`: Must be one of defined values
- `min_participants`: Must be ≥ 2
- `grace_period_start/end`: Both must be set when status is 'grace_period'

### ParticipantSnapshot
- `participants`: Must be valid JSON array
- `participant_count`: Must match array length
- Cannot be updated after creation

### ActivityFeed
- `event_type`: Must be one of defined values
- `message`: Cannot be empty
- `participant_name`: Required for participant_joined/participant_left events

