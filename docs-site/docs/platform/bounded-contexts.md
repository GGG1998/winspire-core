# Bounded Contexts

This document defines the **Bounded Contexts** in the Winspire platform and maps them to services, aggregates, and domain events.

## Overview

The Winspire platform is organized using **Domain-Driven Design (DDD)** principles. Each **Bounded Context** represents a subdomain with its own ubiquitous language, aggregates, and business rules.

```
┌─────────────────────┐    Events     ┌─────────────────┐
│ TournamentManagement│ ────────────> │  Matchmaking    │
│  (competition)      │               │  (matchmaking)  │
└─────────────────────┘               └────────┬────────┘
         │                                     │
         │ Events                              │ Events
         └─────────────────┬───────────────────┘
                           ▼
                 ┌──────────────────────┐
                 │ Realtime/Projections │
                 │ (competition-host-   │
                 │      stream)         │
                 └──────────────────────┘
                           │ SSE/WS
                           ▼
                      [Clients]
```

---

## 1. TournamentManagement

**Service**: `services/competition/`  
**Purpose**: Manage tournament lifecycle and participant registration

### Aggregates
- **Tournament** (root)
  - States: draft, scheduled, registration_open, registration_closed, started, completed, cancelled
- **TournamentParticipant**
  - States: pending, registered, confirmed, checked_in
- **Host**

### Key Events
- `TournamentCreated`
- `TournamentPublished`
- `RegistrationOpened`
- `RegistrationClosed`
- `ParticipantRegistered`
- `ParticipantConfirmed`
- `ParticipantCheckedIn`
- `TournamentStarted` ← **triggers Matchmaking**
- `TournamentCompleted`
- `TournamentCancelled`

### Business Rules (Invariants)
- Cannot start tournament with less than `minimum_team_count`
- Cannot register if `participant_count >= maximum_team_count`
- Can only confirm participation if tournament is in `registration_open` or `scheduled`
- Check-in window is 15 minutes before `scheduled_start_time`

### Database Schema
- `tournaments`
- `tournament_registrations`
- `hosts`

---

## 2. Matchmaking

**Service**: `services/matchmaking/` (to be created)  
**Purpose**: Generate brackets, create matches, and manage game flow

### Aggregates
- **Bracket** (root)
  - Contains rounds and match assignments
- **Round**
  - Sequential stages in a tournament (e.g., Round 1, Semifinals, Finals)
- **Match**
  - States: pending, ready, started, completed, disputed
- **MatchParticipant**
  - Player state within a match
- **Lobby**
  - Pre-match waiting room

### Key Events
- `BracketGenerated` ← triggered by `TournamentStarted`
- `RoundCreated`
- `MatchCreated`
- `ParticipantJoinedLobby`
- `ParticipantMarkedReady`
- `MatchStarted`
- `ScoreSubmitted`
- `MatchCompleted`
- `MatchDisputed`
- `DisputeResolved`
- `ParticipantAdvanced`
- `ParticipantEliminated`
- `WalkoverGranted`

### Business Rules (Invariants)
- Bracket must be generated before any matches can start
- Round N cannot start until Round N-1 is completed
- Match requires both participants to be ready (or `auto_force_ready=true`)
- Score submission requires match to be in `started` state
- Dispute can only be raised within 5 minutes of match completion

### Database Schema
- `brackets`
- `rounds`
- `matches`
- `match_participants`
- `lobbies`

---

## 3. Realtime/Projections

**Service**: `services/competition-host-stream/`  
**Purpose**: Stream real-time updates to clients (SSE/WebSocket)

### Responsibilities
- Listen to events from `competition` and `matchmaking`
- Build read-optimized projections (denormalized views)
- Publish events to connected clients via SSE

### Key Projections
- **TournamentHostView** - for hosts (participant list, status)
- **TournamentLobbyView** - for participants (match assignments, opponents)
- **BracketView** - visual bracket state
- **MatchView** - live match details

### Technology
- **SSE Broker** (in-memory or Redis-backed)
- **Event Router** - routes domain events to SSE channels
- **PostgreSQL** (read-only views)

### ⚠️ Important
This service does **NOT** contain domain logic. It only:
- Listens to events
- Transforms data for frontend
- Publishes via SSE/WebSocket

---

## 4. GameLibrary

**Service**: `services/game-management/`  
**Purpose**: Catalog games and manage game assets

### Aggregates
- **Game** (root)
- **GameVersion**
- **GameBundle**

### Key Events
- `GameCreated`
- `GameVersionPublished`
- `GameBundleUploaded`

### Database Schema
- `games`
- `game_versions`
- **S3** for game assets

---

## Communication Patterns

### Event-Driven (Async)
Primary communication pattern between bounded contexts.

**Example Flow**:
1. Host clicks "Start Tournament" → `competition` service
2. `competition` publishes `TournamentStarted` event
3. `matchmaking` listens to `TournamentStarted`
4. `matchmaking` generates bracket → publishes `BracketGenerated`, `RoundCreated`, `MatchCreated`
5. `competition-host-stream` listens to all events → streams to clients

**Transport Options**:
- **Current**: PostgreSQL NOTIFY/LISTEN
- **Future**: AWS SNS/SQS, EventBridge, or dedicated Event Store

### REST APIs (Sync)
Used only for:
- Client commands (POST /tournaments, POST /matches/ready)
- Rare service-to-service queries (prefer events)

---

## Anti-Patterns

❌ **Shared Database**: Each bounded context must own its data  
❌ **Direct Service Calls**: Use events for domain operations  
❌ **Domain Logic in Stream Service**: Only projections and routing  
❌ **Mixing Contexts**: Tournament registration logic stays in `competition`, not `matchmaking`

---

## Adding New Bounded Contexts

1. **Identify the subdomain** - what business capability does it represent?
2. **Define aggregates** - what are the domain entities and their lifecycles?
3. **Document events** - what events does it publish and consume?
4. **Create service** - use `services/template/` cookiecutter
5. **Update this document** - add the new bounded context here
6. **Document event schemas** - create `docs-site/docs/platform/events/<context>.md`

---

**Last Updated**: 2025-01-27  
**Related**: [Architecture Overview](./architecture-overview.md), [Event Sourcing Patterns](./event-sourcing-patterns.md)


