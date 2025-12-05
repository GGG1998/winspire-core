# Tasks: Tournament Pre-Lobby Backend

**Input**: Design documents from `docs-site/docs/004-tournament-prelobby-backend/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓

**Tests**: Not explicitly requested in specification - test tasks omitted.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions (Modular Monorepo)

- Services: `services/matchmaking/cmd/`, `services/matchmaking/internal/`
- Migrations: `services/matchmaking/migrations/`
- Store: `services/matchmaking/internal/store/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Database schema, SQLC configuration, and shared types

- [x] T001 Create database migration in `services/matchmaking/migrations/000005_create_prelobby_tables.sql` with prelobbies, prelobby_participant_snapshots, and prelobby_activity_feed tables
- [x] T002 Add pre-lobby tables to schema reference in `services/matchmaking/internal/store/schema.sql`
- [x] T003 Create SQLC queries in `services/matchmaking/internal/store/prelobby_queries.sql` for all pre-lobby operations
- [x] T004 Update `services/matchmaking/sqlc.yaml` to include prelobby_queries.sql
- [x] T005 Run `make sqlc` to generate Go code in `services/matchmaking/internal/store/sqlc/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Domain entities, repository layer, and WebSocket infrastructure extensions that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T006 Create PreLobby domain aggregate with status state machine in `services/matchmaking/internal/domain/prelobby.go`
- [x] T007 [P] Add pre-lobby WebSocket message types (prelobby_state, participant_joined, participant_left, grace_period_started, roster_updated, grace_period_ended, match_assigned, tournament_cancelled) in `services/matchmaking/internal/websocket/message.go`
- [x] T008 [P] Add pre-lobby event channels (ChannelPreLobbyGracePeriodStarted, ChannelPreLobbyParticipantSnapshot) in `services/matchmaking/internal/pubsub/channels.go`
- [x] T009 [P] Add pre-lobby domain events (PreLobbyCreated, GracePeriodStarted, GracePeriodEnded, ParticipantSnapshotCreated) in `services/matchmaking/internal/domain/events.go`
- [x] T010 Create PreLobbyRepository interface and implementation in `services/matchmaking/internal/repository/prelobby.go`
- [x] T011 Extend WebSocket Hub to support tournament-based rooms with RoomType discriminator in `services/matchmaking/internal/websocket/hub.go`

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Pre-Lobby State API (Priority: P1) 🎯 MVP

**Goal**: Frontend applications can retrieve current pre-lobby state via REST API

**Independent Test**: Call GET `/v1/tournaments/:id/lobby` and verify response contains tournament metadata, participant list, and lobby status

### Implementation for User Story 1

- [x] T012 [US1] Create PreLobbyService with GetState method in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T013 [US1] Implement GET handler for `/v1/tournaments/:tournamentId/lobby` in `services/matchmaking/internal/http/prelobby.go`
- [x] T014 [US1] Add pre-lobby routes to router setup in `services/matchmaking/cmd/matchmaking/main.go`
- [x] T015 [US1] Implement PreLobbyState response DTO with tournament metadata, participants, status, and grace period info in `services/matchmaking/internal/http/prelobby.go`

**Checkpoint**: REST API returns pre-lobby state - can be tested with curl/Postman

---

## Phase 4: User Story 2 - WebSocket Connection Management (Priority: P1)

**Goal**: Real-time participant presence tracking via WebSocket connections

**Independent Test**: Connect multiple WebSocket clients and verify participant_joined/participant_left events broadcast to all

### Implementation for User Story 2

- [x] T016 [US2] Add tournament room management methods (RegisterTournamentClient, UnregisterTournamentClient, GetTournamentParticipants) to Hub in `services/matchmaking/internal/websocket/hub.go`
- [x] T017 [US2] Implement WebSocket upgrade handler for `/v1/tournaments/:tournamentId/lobby` in `services/matchmaking/internal/http/prelobby.go`
- [x] T018 [US2] Add PreLobbyService methods for participant join/leave with broadcast in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T019 [US2] Implement initial prelobby_state message on WebSocket connection in `services/matchmaking/internal/http/prelobby.go`
- [x] T020 [US2] Implement participant_joined broadcast when player connects in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T021 [US2] Implement participant_left broadcast when player disconnects with 5s detection in `services/matchmaking/internal/application/prelobby_service.go`

**Checkpoint**: WebSocket connections work - participants see each other join/leave in real-time

---

## Phase 5: User Story 3 - Grace Period Management (Priority: P1)

**Goal**: 30-second grace period after tournament start with participant snapshot for bracket generation

**Independent Test**: Trigger tournament start, connect player during 30s window, verify they're included in bracket

### Implementation for User Story 3

- [x] T022 [US3] Create GracePeriodManager with timer and recovery logic in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T023 [US3] Add TournamentStarted event handler to subscribe to Redis pub/sub in `services/matchmaking/internal/application/event_handler.go`
- [x] T024 [US3] Implement StartGracePeriod method that persists state and starts 30s timer in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T025 [US3] Implement grace_period_started broadcast to all connected participants in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T026 [US3] Implement roster_updated broadcast when participant count changes during grace period in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T027 [US3] Implement FinalizeGracePeriod that creates participant snapshot in database in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T028 [US3] Implement grace_period_ended broadcast and trigger bracket generation in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T029 [US3] Implement tournament cancellation if participant count < minimum when grace period ends in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T030 [US3] Add RecoverActiveGracePeriods on service startup in `services/matchmaking/cmd/matchmaking/main.go`

**Checkpoint**: Grace period works end-to-end - tournament start triggers 30s countdown, snapshot created, bracket generation triggered

---

## Phase 6: User Story 4 - Tournament-Matchmaking Integration (Priority: P2)

**Goal**: Verify participant eligibility by checking registration with competition service

**Independent Test**: Connect with registered user (succeeds) and unregistered user (fails)

### Implementation for User Story 4

- [x] T031 [US4] Create CompetitionClient interface for registration verification in `services/matchmaking/internal/application/competition_client.go`
- [x] T032 [US4] Implement HTTP client for competition service with 500ms timeout in `services/matchmaking/internal/application/competition_client.go`
- [x] T033 [US4] Add 30-second cache for registration status using sync.Map in `services/matchmaking/internal/application/competition_client.go`
- [x] T034 [US4] Integrate registration check into REST handler before returning state in `services/matchmaking/internal/http/prelobby.go`
- [x] T035 [US4] Integrate registration check into WebSocket handler before accepting connection in `services/matchmaking/internal/http/prelobby.go`
- [x] T036 [US4] Implement tournament status validation (reject draft, completed, cancelled) in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T037 [US4] Add graceful error handling for competition service unavailability (503 response) in `services/matchmaking/internal/http/prelobby.go`

**Checkpoint**: Only registered participants can access pre-lobby

---

## Phase 7: User Story 5 - Activity Feed Events (Priority: P3)

**Goal**: Chronological feed of recent activity (joins, leaves, announcements)

**Independent Test**: Perform join/leave actions and verify activity feed in state response shows events in order

### Implementation for User Story 5

- [x] T038 [US5] Add AddActivityFeedEvent method to PreLobbyRepository in `services/matchmaking/internal/repository/prelobby.go`
- [x] T039 [US5] Add GetRecentActivityFeed method (limit 20) to PreLobbyRepository in `services/matchmaking/internal/repository/prelobby.go`
- [x] T040 [US5] Record participant_joined event when player connects in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T041 [US5] Record participant_left event when player disconnects in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T042 [US5] Record grace_period_started event when grace period begins in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T043 [US5] Record tournament_cancelled event when start is cancelled in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T044 [US5] Include activity feed in REST API response in `services/matchmaking/internal/http/prelobby.go`
- [x] T045 [US5] Include activity feed in WebSocket prelobby_state message in `services/matchmaking/internal/http/prelobby.go`

**Checkpoint**: Activity feed shows recent events - complete pre-lobby functionality

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T046 [P] Add structured logging for all pre-lobby operations in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T047 [P] Add metrics for pre-lobby connections, grace period duration, bracket generation time in `services/matchmaking/internal/observability/metrics.go`
- [x] T048 Implement match_assigned broadcast to participants after bracket generation in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T049 Add edge case handling: duplicate tournament start commands (ignore if grace period active) in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T050 Add edge case handling: participant count drops to 0 during grace period (cancel tournament) in `services/matchmaking/internal/application/prelobby_service.go`
- [x] T051 Run database migration and verify schema in local environment
- [x] T052 Validate quickstart.md instructions work end-to-end

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - US1, US2, US3 are all P1 priority - implement in order (dependencies between them)
  - US4 (P2) can start after US2 is complete
  - US5 (P3) can start after US1 is complete
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

```
Setup (Phase 1)
    ↓
Foundational (Phase 2)
    ↓
┌───────────────────────────────────────┐
│                                       │
US1 (REST API) ──→ US2 (WebSocket) ──→ US3 (Grace Period)
       │                                       │
       │                                       ↓
       └──→ US5 (Activity Feed)         US4 (Integration)
                                               │
                                               ↓
                                        Polish (Phase 8)
```

### Within Each User Story

- Models/domain before services
- Services before handlers
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- T007, T008, T009 can run in parallel (different files)
- T046, T047 can run in parallel (different files)
- Once Foundational phase completes, US5 can be worked on in parallel with US2/US3/US4

---

## Parallel Example: Foundational Phase

```bash
# Launch these tasks together (different files, no dependencies):
Task T007: "Add pre-lobby WebSocket message types in websocket/message.go"
Task T008: "Add pre-lobby event channels in pubsub/channels.go"
Task T009: "Add pre-lobby domain events in domain/events.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T005)
2. Complete Phase 2: Foundational (T006-T011)
3. Complete Phase 3: User Story 1 (T012-T015)
4. **STOP and VALIDATE**: Test REST API returns pre-lobby state
5. Deploy/demo if ready - frontend can display basic pre-lobby info

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test REST API → Frontend can show pre-lobby (MVP!)
3. Add User Story 2 → Test WebSocket → Real-time presence working
4. Add User Story 3 → Test grace period → Full tournament start flow
5. Add User Story 4 → Test authorization → Security complete
6. Add User Story 5 → Test activity feed → Enhanced UX
7. Each story adds value without breaking previous stories

### Recommended Order

For single developer:
1. Setup (Phase 1): ~30 min
2. Foundational (Phase 2): ~2 hours
3. US1 REST API (Phase 3): ~1 hour
4. US2 WebSocket (Phase 4): ~2 hours
5. US3 Grace Period (Phase 5): ~3 hours
6. US4 Integration (Phase 6): ~1.5 hours
7. US5 Activity Feed (Phase 7): ~1 hour
8. Polish (Phase 8): ~1 hour

**Total estimated time**: ~12 hours

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- All file paths are relative to repository root
- Run `make sqlc` after modifying SQL queries
- Run `make migrate-up` after adding migrations

