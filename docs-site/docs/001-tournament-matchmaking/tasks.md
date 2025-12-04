# Tasks: Tournament Matchmaking System

**Input**: Design documents from `/docs-site/docs/001-tournament-matchmaking/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Not explicitly requested in spec - tasks focus on implementation only

**Organization**: Tasks grouped by user story to enable independent implementation and testing of each story

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

**Modular Monorepo (winspire-core)**:
- New Service: `services/matchmaking/`
- Shared Libraries: `libs/go/`
- All paths below use absolute service paths

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create new matchmaking microservice structure

- [X] T001 Create services/matchmaking/ directory structure (cmd/, internal/, migrations/)
- [X] T002 Initialize Go module in services/matchmaking/go.mod with Go 1.25.4
- [X] T003 [P] Add matchmaking service to root go.work file
- [X] T004 [P] Create services/matchmaking/Makefile with build, test, and sqlc generate targets
- [X] T005 [P] Create services/matchmaking/Dockerfile for containerized deployment
- [X] T006 [P] Create services/matchmaking/.env.example with required environment variables
- [X] T007 [P] Setup services/matchmaking/.gitignore for Go artifacts

**Checkpoint**: Service skeleton ready

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Database & Migrations

- [X] T008 Create services/matchmaking/migrations/000001_create_brackets_table.sql (tournament_brackets table with no FK to tournaments)
- [X] T009 Create services/matchmaking/migrations/000002_create_rounds_table.sql (tournament_rounds table)
- [X] T010 Create services/matchmaking/migrations/000003_create_matches_table.sql (tournament_matches table with disconnect tracking)
- [X] T011 Configure goose migration tool in services/matchmaking/Makefile
- [X] T012 Create services/matchmaking/internal/store/schema.sql with all table definitions for SQLC reference

### SQLC Setup

- [X] T013 Create services/matchmaking/sqlc.yaml configuration file
- [X] T014 [P] Create services/matchmaking/internal/store/bracket_queries.sql (CreateBracket, GetBracketByTournamentID, GetBracketWithRoundsAndMatches)
- [X] T015 [P] Create services/matchmaking/internal/store/round_queries.sql (CreateRound, GetRoundsByBracketID, UpdateRoundStatus)
- [X] T016 [P] Create services/matchmaking/internal/store/match_queries.sql (CreateMatch, GetMatchByID, UpdateMatchStatus, UpdateMatchResult, UpdateMatchDisconnect)
- [X] T017 Generate SQLC code: run sqlc generate to create services/matchmaking/internal/store/*.go files
- [X] T018 Create services/matchmaking/internal/store/db.go with database connection pool using pgx/v5

### Domain Models & Events

- [X] T019 [P] Create services/matchmaking/internal/domain/bracket.go (Bracket aggregate with business logic)
- [X] T020 [P] Create services/matchmaking/internal/domain/round.go (Round entity with state transitions)
- [X] T021 [P] Create services/matchmaking/internal/domain/match.go (Match aggregate with states: pending, ready, started, paused, completed)
- [X] T022 [P] Create services/matchmaking/internal/domain/events.go (DomainEvent interface and event types: BracketGenerated, MatchCreated, MatchStarted, MatchCompleted)

### Redis Pub/Sub Infrastructure

- [X] T023 Create services/matchmaking/internal/pubsub/publisher.go (EventPublisher with Redis client, Publish method using channel-per-event pattern)
- [X] T024 Create services/matchmaking/internal/pubsub/subscriber.go (EventSubscriber with Redis Pub/Sub, Subscribe method, event routing)
- [X] T025 Create services/matchmaking/internal/pubsub/channels.go (Channel naming constants: events:matchmaking:*, events:tournament_management:*)

### WebSocket Infrastructure

- [X] T026 Create services/matchmaking/internal/websocket/hub.go (WebSocket hub for managing client connections per match)
- [X] T027 Create services/matchmaking/internal/websocket/client.go (WebSocket client with heartbeat monitoring, disconnect detection)
- [X] T028 Create services/matchmaking/internal/websocket/message.go (Message types for lobby, ready status, score submission)

### HTTP Framework & Middleware

- [X] T029 Create services/matchmaking/cmd/matchmaking/main.go with Gin router setup, graceful shutdown
- [X] T030 [P] Create services/matchmaking/internal/http/middleware.go (Auth middleware using libs/go/auth, Logging middleware using libs/go/httpx)
- [X] T031 [P] Create services/matchmaking/internal/http/health_handler.go (GET /health endpoint for ECS health checks)
- [X] T032 Configure Gin router with middleware, CORS, and error handling in main.go

### Configuration & Observability

- [X] T033 Create services/matchmaking/internal/config/config.go (Environment variable loading: DATABASE_URL, REDIS_URL, JWT_SECRET, PORT)
- [X] T034 [P] Create services/matchmaking/internal/observability/metrics.go (CloudWatch metrics emitter for bracket generation time, match completion rates)
- [X] T035 [P] Create services/matchmaking/internal/observability/logger.go (Structured logging to CloudWatch Logs with context fields)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Host Starts Tournament and Bracket Generates Automatically (Priority: P1) 🎯 MVP

**Goal**: When host starts tournament in competition service, matchmaking service receives event and automatically generates bracket with rounds and matches

**Independent Test**: Start tournament with 8 participants via competition service, verify bracket appears with 4 first-round matches in matchmaking database

### Implementation for User Story 1

- [X] T036 [P] [US1] Create services/matchmaking/internal/repository/bracket_repo.go (BracketRepository interface and implementation using SQLC queries)
- [X] T037 [P] [US1] Create services/matchmaking/internal/repository/round_repo.go (RoundRepository interface and implementation)
- [X] T038 [P] [US1] Create services/matchmaking/internal/repository/match_repo.go (MatchRepository interface and implementation)
- [X] T039 [US1] Create services/matchmaking/internal/application/bracket_service.go (BracketService with GenerateBracket method)
- [X] T040 [US1] Implement bracket generation algorithm in bracket_service.go (binary tree algorithm from research.md: calculate rounds, assign byes, create match tree)
- [X] T041 [US1] Implement bye assignment logic (random selection, equal probability for all participants)
- [X] T042 [US1] Implement match numbering and progression tracking (next_match_id linking)
- [X] T043 [US1] Create services/matchmaking/internal/application/event_handler.go (HandleTournamentStarted event handler, subscribes to events:tournament_management:tournament_started)
- [X] T044 [US1] Integrate event_handler.go with bracket_service.go (call GenerateBracket when TournamentStarted received)
- [X] T045 [US1] Implement transaction wrapper for bracket generation (atomic create of bracket + rounds + matches)
- [X] T046 [US1] Publish BracketGenerated event after successful generation via events:matchmaking:bracket_generated
- [X] T047 [US1] Add error handling and rollback for bracket generation failures
- [X] T048 [US1] Add CloudWatch metric emission for bracket generation duration (SC-001: <2s target)
- [X] T049 [US1] Start event subscriber in main.go (subscribe to TournamentStarted on service startup)

**Checkpoint**: At this point, User Story 1 should be fully functional - tournament start triggers bracket generation

---

## Phase 4: User Story 2 - Player Receives Match Assignment and Joins Lobby (Priority: P1)

**Goal**: Players can view their match assignments, join match lobbies, and indicate readiness

**Independent Test**: After bracket generation, player can GET their match, see opponent details, join lobby via WebSocket, and mark ready

### Implementation for User Story 2

- [X] T050 [P] [US2] Create services/matchmaking/internal/http/bracket_handler.go (GET /v1/brackets/:id endpoint, GET /v1/tournaments/:id/bracket endpoint)
- [X] T051 [P] [US2] Create services/matchmaking/internal/http/match_handler.go (GET /v1/matches/:id endpoint, POST /v1/matches/:id/ready endpoint)
- [X] T052 [US2] Implement GetBracket handler in bracket_handler.go (fetch bracket with all rounds and matches using complex join query)
- [X] T053 [US2] Implement GetMatch handler in match_handler.go (fetch single match with participant details)
- [X] T054 [US2] Implement MarkPlayerReady handler in match_handler.go (update participant1_ready or participant2_ready, persist server-side per FR-019)
- [X] T055 [US2] Add JWT authentication to match lobby endpoints (verify user ID matches participant ID per FR-023)
- [X] T056 [US2] Implement lobby access control (deny access if authenticated user ID not in match participants per FR-024, FR-025)
- [X] T057 [US2] Create services/matchmaking/internal/http/websocket_handler.go (WebSocket upgrade endpoint: GET /v1/matches/:id/lobby)
- [X] T058 [US2] Implement WebSocket lobby connection (register client in hub, send initial lobby state)
- [X] T059 [US2] Implement WebSocket ready status broadcasting (when player marks ready, notify opponent via WebSocket)
- [X] T060 [US2] Publish ParticipantJoinedLobby event when player connects to lobby WebSocket
- [X] T061 [US2] Add error responses for unauthorized lobby access attempts (SC message from FR-024)
- [X] T062 [US2] Register bracket and match handlers in Gin router

**Checkpoint**: At this point, User Stories 1 AND 2 should both work - players can view brackets and join lobbies

---

## Phase 5: User Story 3 - Match Completes and Winner Advances (Priority: P1)

**Goal**: When match completes, winner automatically advances to next round and loser is marked eliminated

**Independent Test**: Complete a first-round match, verify winner appears in next round match, loser marked eliminated

### Implementation for User Story 3

- [X] T063 [P] [US3] Create services/matchmaking/internal/application/match_service.go (MatchService with CompleteMatch, AdvanceWinner methods)
- [X] T064 [US3] Implement CompleteMatch in match_service.go (update match with winner_id, scores, status=completed, completed_at timestamp)
- [X] T065 [US3] Implement winner advancement logic (find next_match_id, assign winner to participant slot in next match per FR-014)
- [X] T066 [US3] Implement loser elimination tracking (mark loser as eliminated per FR-015)
- [X] T067 [US3] Handle bye matches (auto-advance participant1 when participant2_id is NULL)
- [X] T068 [US3] Publish MatchCompleted event with winner_id and loser_id
- [X] T069 [US3] Publish ParticipantAdvanced event when winner assigned to next match
- [X] T070 [US3] Publish ParticipantEliminated event for losing player
- [X] T071 [US3] Add CloudWatch logging for match completion events

**Checkpoint**: Tournament progression now works - matches complete and advance winners

---

## Phase 6: User Story 4 - Match Starts When Both Players Ready (Priority: P2)

**Goal**: When both players mark ready in lobby, match automatically transitions to "started" state

**Independent Test**: Two players in lobby both click ready, verify match status changes to "started" and both receive notification

### Implementation for User Story 4

- [X] T072 [US4] Add ready state detection in match_service.go (CheckBothPlayersReady method)
- [X] T073 [US4] Implement automatic match start trigger (when both ready, update status to "started", set started_at timestamp)
- [X] T074 [US4] Publish MatchStarted event when match begins
- [X] T075 [US4] Broadcast match start notification via WebSocket to both players
- [X] T076 [US4] Implement auto-force-ready logic for tournaments with auto_force_ready enabled (if specified start time arrives, force start per user story acceptance)
- [X] T077 [US4] Add CloudWatch metric for time between both-ready and match-started (SC-009: <5s target)

**Checkpoint**: Match lifecycle complete - ready detection works

---

## Phase 7: User Story 5 - Player Check-in Before Tournament Start (Priority: P2)

**Goal**: Players can check in before tournament starts, only checked-in players included in bracket

**Independent Test**: Enable check-in, have some players check in and others not, verify only checked-in players in generated bracket

**Note**: This story spans competition and matchmaking services. Matchmaking only receives participant list from TournamentStarted event - check-in logic lives in competition service.

### Implementation for User Story 5

- [X] T078 [US5] Update event_handler.go to filter participants from TournamentStarted event (only include checked-in participants if check-in enabled)
- [X] T079 [US5] Add participant status validation in bracket generation (ensure participant list only contains eligible players)
- [X] T080 [US5] Handle edge case: minimum participant count not met after check-in filtering (reject bracket generation, publish error event)

**Checkpoint**: Check-in integration complete

---

## Phase 8: User Story 6 - Handle No-Show with Walkover (Priority: P2)

**Goal**: If player doesn't join lobby within timeout, opponent awarded walkover and advances

**Independent Test**: Start match, have only one player join lobby, wait for timeout, verify walkover granted

### Implementation for User Story 6

- [X] T081 [US6] Add no-show timeout tracking in websocket/hub.go (track lobby join timestamps per player)
- [X] T082 [US6] Implement timeout detection (if opponent doesn't join within 2 minutes per FR-032)
- [X] T083 [US6] Implement walkover grant logic in match_service.go (GrantWalkover method, mark present player as winner per FR-016)
- [X] T084 [US6] Publish WalkoverGranted event with winner and reason (no-show)
- [X] T085 [US6] Handle both players absent scenario (notify host for manual resolution per acceptance criteria)
- [X] T086 [US6] Add POST /v1/matches/:id/claim-walkover endpoint for player-initiated walkover claim
- [X] T087 [US6] Add host notification for both-absent scenario (integrate with notification service or log for host dashboard)

**Checkpoint**: No-show handling works

---

## Phase 9: User Story 7 - Automatic Score Retrieval and Verification (Priority: P2)

**Goal**: After match completes in external game, system polls Game API for validated results

**Independent Test**: Complete match in game, game client sends score to Game API, matchmaking polls and retrieves validated result

**Note**: Uses polling approach from research.md - game client submits to Game API (with fraud validation), matchmaking polls to retrieve

### Implementation for User Story 7

- [X] T088 [P] [US7] Create services/matchmaking/internal/gameapi/client.go (GameAPIClient with PollMatchResult method, circuit breaker pattern from research.md)
- [X] T089 [US7] Implement polling trigger when match status changes to 'started' (start polling after both players ready)
- [X] T090 [US7] Implement polling loop with 5-second intervals (poll Game API GET /api/matches/:id/result per research.md)
- [X] T091 [US7] Handle successful result retrieval (Game API returns validated score with fraud check results, call match_service.CompleteMatch per FR-029)
- [X] T092 [US7] Handle Game API fraud validation failure (if API indicates fraud detected, flag match for host review per FR-027)
- [X] T093 [US7] Handle Game API timeout (60s polling timeout → flag match for manual host entry per FR-027)
- [X] T094 [US7] Implement circuit breaker for Game API calls (prevent cascading failures, open after 5 consecutive failures per research.md)
- [X] T095 [US7] Add CloudWatch metric for polling success rate (SC-011: 95% target)
- [X] T096 [US7] Add CloudWatch metric for result retrieval latency (SC-012: <10s target from polling start to result)
- [X] T097 [US7] Publish ScoreRetrieved event when Game API returns validated result
- [X] T098 [US7] Add audit logging for all polling attempts (timestamp, match_id, poll_count, result, validation_status)

**Checkpoint**: Real-time score submission with fraud detection works

---

## Phase 10: User Story 8 - Host Override for API Issues (Priority: P3)

**Goal**: Host can manually enter/override match results when Game API fails or returns incorrect results

**Independent Test**: Simulate Game API failure, host manually enters result, match completes with host-entered data

### Implementation for User Story 8

- [ ] T099 [US8] Add PATCH /v1/matches/:id/result endpoint in match_handler.go (host-only, requires host role per FR-034)
- [ ] T100 [US8] Implement host authentication check (verify JWT contains host role for tournament)
- [ ] T101 [US8] Implement manual result entry (accept winner_id, scores, override_reason per FR-028)
- [ ] T102 [US8] Add result_source field update (set to 'host_manual' instead of 'game_api' or 'walkover')
- [ ] T103 [US8] Implement audit logging for manual overrides (log timestamp, host_id, original_result, new_result, reason per FR-030)
- [ ] T104 [US8] Publish MatchResultOverridden event with audit details
- [ ] T105 [US8] Add "Result manually corrected by host" indicator in GetMatch response (when result_source='host_manual')

**Checkpoint**: Host override capability complete

---

## Phase 11: User Story 9 - Tournament Completes with Winner Declaration (Priority: P2)

**Goal**: When final match completes, tournament automatically marked as completed and winner declared

**Independent Test**: Complete final match, verify TournamentCompleted event published with champion ID

### Implementation for User Story 9

- [X] T106 [US9] Add final match detection in match_service.go (check if completed match has next_match_id = NULL)
- [X] T107 [US9] Implement tournament completion logic (when final match completes, determine champion)
- [X] T108 [US9] Publish TournamentCompleted event with tournament_id and champion_id (consumed by competition service)
- [X] T109 [US9] Add bracket completion status tracking (update bracket record with completed_at timestamp)
- [X] T110 [US9] Generate final standings data (all participants with placement: 1st, 2nd, eliminated in round X)

**Checkpoint**: Tournament lifecycle complete end-to-end

---

## Phase 12: Disconnect Handling (CS:GO Style) - Cross-Cutting for Active Matches

**Goal**: Handle player disconnections during active matches with 30s reconnect window and point loss

**Independent Test**: Player disconnects during match, verify opponent gets +1 point, 30s timer starts, reconnect resumes match or timeout disqualifies

**Note**: This applies to User Stories 4, 7, 9 (any active match)

### Implementation for Disconnect Handling

- [X] T111 Implement WebSocket heartbeat monitoring in websocket/client.go (detect missed heartbeats per research.md, 10s threshold)
- [X] T112 Implement disconnect detection handler (when heartbeat timeout occurs, trigger disconnect flow per FR-016a)
- [X] T113 Implement point award on disconnect (update match score: opponent gets +1 point per FR-016b)
- [X] T114 Implement match pause on disconnect (set status='paused', store disconnected_at timestamp per FR-016a)
- [X] T115 Implement 30-second reconnection timer (use time.AfterFunc, start from app exit timestamp per FR-016c)
- [X] T116 Implement reconnection handler (if player reconnects within 30s, resume match per FR-016f)
- [X] T117 Implement disqualification on timeout (if 30s expires without reconnect, grant walkover per FR-016d)
- [X] T118 Handle both players disconnect scenario (track timestamps independently per FR-016e, first to disconnect gives point, both get 30s timers)
- [X] T119 Publish PlayerConnectionLost event when disconnect detected
- [X] T120 Publish PlayerConnectionRestored event when player reconnects successfully
- [X] T121 Add disconnect/reconnect state to match database (disconnected_player_id, disconnected_at columns already in schema)
- [X] T122 Add CloudWatch logging for disconnect events (frequency, duration, reconnect success rate)

**Checkpoint**: Disconnect handling fully implemented

---

## Phase 13: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

### Observability & Monitoring

- [ ] T123 [P] Implement CloudWatch metrics for all success criteria (SC-001 through SC-012):
  - SC-001: Emit bracket_generation_duration_seconds histogram
  - SC-002: Emit tournament_completion_rate gauge (requires aggregation)
  - SC-003: Emit match_assignment_notification_latency_seconds histogram
  - SC-004: Emit match_state_update_propagation_latency_seconds histogram
  - SC-005: Emit concurrent_tournaments_count gauge
  - SC-006: Emit walkover_flow_duration_seconds histogram
  - SC-007: Emit score_submission_success_rate gauge
  - SC-008: Emit tournament_completion_success_rate gauge
  - SC-009: Emit ready_to_started_latency_seconds histogram
  - SC-010: Emit opponent_view_latency_seconds histogram
  - SC-011: Emit automatic_score_retrieval_rate gauge
  - SC-012: Emit game_api_polling_latency_seconds histogram
  - Use CloudWatch embedded metrics format (EMF) for efficient emission
  
- [ ] T124 [P] Add structured logging to CloudWatch Logs for all state transitions:
  - Tournament events: TournamentStarted received (with participant count)
  - Bracket events: BracketGenerated (with rounds, matches, byes)
  - Match events: MatchCreated, MatchStarted, MatchCompleted, MatchCancelled
  - Player events: ParticipantJoinedLobby, ParticipantReady, ParticipantDisconnected
  - Polling events: GameAPIPollStarted, GameAPIPollSuccess, GameAPIPollTimeout
  - Error events: All errors with context (match_id, player_id, error_type)
  - Include structured fields: timestamp, level, match_id, tournament_id, player_id, event_type
  - Log format: JSON with correlation_id for request tracing
  
- [ ] T125 [P] Configure CloudWatch alarms for SLO violations (12 alarms total):
  - Alarm: BracketGenerationSlow (bracket_generation_duration_seconds >2s for >5% of requests)
  - Alarm: TournamentCompletionLow (tournament_completion_rate <90%)
  - Alarm: MatchNotificationSlow (match_assignment_notification_latency_seconds >5s for >5%)
  - Alarm: MatchStatePropagationSlow (match_state_update_propagation_latency_seconds >3s for >5%)
  - Alarm: ConcurrentTournamentLimit (concurrent_tournaments_count >50)
  - Alarm: WalkoverFlowSlow (walkover_flow_duration_seconds >32s for >5%)
  - Alarm: ScoreSubmissionLow (score_submission_success_rate <99%)
  - Alarm: ReadyToStartedSlow (ready_to_started_latency_seconds >5s for >5%)
  - Alarm: OpponentViewSlow (opponent_view_latency_seconds >10s for >5%)
  - Alarm: AutoScoreRetrievalLow (automatic_score_retrieval_rate <95%)
  - Alarm: GameAPIPollingS low (game_api_polling_latency_seconds >10s for >5%)
  - Configure SNS topic for alarm notifications to ops team
  
- [ ] T126 [P] Add request tracing headers to all API endpoints and WebSocket connections:
  - Generate correlation_id (ULID) for each inbound request
  - Propagate correlation_id to all downstream calls (Game API, Redis events)
  - Add causation_id to domain events (links event to triggering request)
  - Include trace context in all log entries
  - Add X-Correlation-ID response header for client correlation
  - Store correlation_id in match/bracket records for debugging

### Error Handling & Resilience

- [ ] T127 Implement circuit breaker for Game API calls using gobreaker (prevent cascading failures per research.md)
- [ ] T128 Add retry logic with exponential backoff for Redis Pub/Sub connection failures
- [ ] T129 Implement graceful degradation (if Redis fails, log events locally and retry publish)
- [ ] T130 Add database connection pool health monitoring

### Documentation

- [ ] T131 [P] Update services/matchmaking/README.md with service architecture, endpoints, event flows
- [ ] T132 [P] Document WebSocket message formats in docs-site/docs/001-tournament-matchmaking/websocket-protocol.md
- [ ] T133 [P] Add API examples to quickstart.md for all implemented endpoints
- [ ] T134 [P] Document deployment steps (Docker build, ECS task definition, environment variables)

### Security Hardening

- [ ] T135 Add rate limiting to API endpoints (prevent abuse)
- [ ] T136 Implement request size limits for WebSocket messages (prevent DoS)
- [ ] T137 Add HTTPS/WSS requirement validation for production deployment
- [ ] T138 Audit all SQL queries for injection vulnerabilities (SQLC provides protection, but double-check)

### Testing Support

- [ ] T139 Run quickstart.md validation (follow developer setup guide, verify all steps work)
- [ ] T140 Create services/matchmaking/scripts/seed-test-tournament.sh (seed database with sample tournament for manual testing)
- [ ] T141 Verify all CloudWatch metrics are emitting correctly in staging environment

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-13)**: All depend on Foundational phase completion
  - US1 (P1): Can start after Foundational - No dependencies on other stories
  - US2 (P1): Can start after Foundational - No dependencies on other stories (though typically implemented after US1)
  - US3 (P1): Integrates with US1 (bracket structure) but independently testable
  - US4 (P2): Integrates with US2 (lobby system)
  - US5 (P2): Integrates with US1 (bracket generation)
  - US6 (P2): Integrates with US2 (lobby system)
  - US7 (P2): Integrates with US3 (match completion)
  - US8 (P3): Extends US7 (manual override of automatic scores)
  - US9 (P2): Integrates with US3 (match completion detection)
  - Disconnect Handling: Cross-cutting, affects US4, US7, US9
- **Polish (Phase 13)**: Depends on all desired user stories being complete

### User Story Dependencies

```
US1 (Bracket Generation) ─┬─> US2 (Lobby & Ready) ──> US4 (Match Start) ─┬─> US7 (Score Submission) ──> US9 (Tournament Complete)
                          │                                               │
                          └─> US3 (Winner Advance) ────────────────────────┘
                          │
                          └─> US5 (Check-in)
                          
US2 ──> US6 (No-show/Walkover)

US7 ──> US8 (Host Override)

Disconnect Handling: Applies to US4, US7, US9 (any active match)
```

### Within Each User Story

- Repositories before services
- Services before handlers
- Handlers before WebSocket integration
- Core implementation before event publishing
- Story complete before moving to next priority

### Parallel Opportunities

**Phase 1 (Setup)**: T003, T004, T005, T006, T007 can run in parallel

**Phase 2 (Foundational)**: 
- T014, T015, T016 (SQL query files) can run in parallel
- T019, T020, T021, T022 (domain models) can run in parallel
- T030, T031, T034, T035 (middleware, health, metrics, logging) can run in parallel

**Phase 3 (US1)**:
- T036, T037, T038 (repositories) can run in parallel

**Phase 4 (US2)**:
- T050, T051 (handlers) can run in parallel

**Phase 5 (US3)**:
- T063 is single service, no parallelization

**Phase 14 (Polish)**:
- T128, T129, T130, T131 (observability) can run in parallel
- T136, T137, T138, T139 (documentation) can run in parallel

**Team Parallelization**: After Foundational (Phase 2) completes, different developers can work on:
- Developer A: US1 (Bracket Generation)
- Developer B: US2 (Lobby System)
- Developer C: Disconnect Handling infrastructure (prepares for US4, US7)

---

## Parallel Example: Foundational Phase

```bash
# Launch all SQL query files together:
Task T014: "Create services/matchmaking/internal/store/bracket_queries.sql"
Task T015: "Create services/matchmaking/internal/store/round_queries.sql"
Task T016: "Create services/matchmaking/internal/store/match_queries.sql"

# Launch all domain models together:
Task T019: "Create services/matchmaking/internal/domain/bracket.go"
Task T020: "Create services/matchmaking/internal/domain/round.go"
Task T021: "Create services/matchmaking/internal/domain/match.go"
Task T022: "Create services/matchmaking/internal/domain/events.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 + User Story 2 + User Story 3)

**Minimum Viable Product**: Basic tournament flow

1. Complete Phase 1: Setup (T001-T007)
2. Complete Phase 2: Foundational (T008-T035) - **CRITICAL**
3. Complete Phase 3: US1 - Bracket Generation (T036-T049)
4. Complete Phase 4: US2 - Lobby System (T050-T062)
5. Complete Phase 5: US3 - Winner Advancement (T063-T071)
6. **STOP and VALIDATE**: 
   - Start tournament → bracket generates
   - Players join lobbies
   - Complete match → winner advances
7. Deploy/demo if ready

**MVP Delivers**:
- ✅ Automatic bracket generation
- ✅ Player lobby access
- ✅ Tournament progression through completion
- ❌ No ready detection (host can manually start matches)
- ❌ No automatic scoring (host enters results)
- ❌ No disconnect handling

### Incremental Delivery

**Iteration 1 (MVP)**: US1 + US2 + US3 → Core bracket flow  
**Iteration 2**: Add US4 (Ready Detection) + US6 (No-show) → Better player experience  
**Iteration 3**: Add US7 (Polling Score Retrieval) + Disconnect Handling → CS:GO-style competition  
**Iteration 4**: Add US5 (Check-in) + US9 (Completion) + US8 (Host Override) + Polish → Production-ready

### Parallel Team Strategy

With 3 developers after Foundational phase complete:

**Week 1-2**:
- Dev A: US1 (Bracket Generation) → T036-T049
- Dev B: US2 (Lobby System) → T050-T062  
- Dev C: Disconnect Handling infrastructure → T110-T121

**Week 3**:
- Dev A: US3 (Winner Advancement) → T063-T071
- Dev B: US4 (Ready Detection) → T072-T077
- Dev C: US7 (Score Submission) → T088-T097

**Week 4**:
- All devs: Integration, testing, polish

---

## Task Summary

- **Total Tasks**: 141
- **Setup (Phase 1)**: 7 tasks
- **Foundational (Phase 2)**: 28 tasks ← **CRITICAL BLOCKING PHASE**
- **User Story 1 (P1)**: 14 tasks (T036-T049) 🎯 MVP
- **User Story 2 (P1)**: 13 tasks (T050-T062) 🎯 MVP
- **User Story 3 (P1)**: 9 tasks (T063-T071) 🎯 MVP
- **User Story 4 (P2)**: 6 tasks (T072-T077)
- **User Story 5 (P2)**: 3 tasks (T078-T080)
- **User Story 6 (P2)**: 7 tasks (T081-T087)
- **User Story 7 (P2)**: 11 tasks (T088-T098)
- **User Story 8 (P3)**: 7 tasks (T099-T105)
- **User Story 9 (P2)**: 5 tasks (T106-T110)
- **Disconnect Handling (Cross-cutting)**: 12 tasks (T111-T122)
- **Polish (Phase 13)**: 19 tasks (T123-T141)

**Parallel Opportunities**: 25 tasks marked [P] can run concurrently with other tasks in their phase

**MVP Scope** (T001-T071): 57 tasks = 40% of total  
**Production-Ready** (MVP + US4-US9 + Disconnect): 124 tasks = 88% of total

---

## Notes

- [P] tasks = different files, no dependencies within phase
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Focus on MVP first (US1-US3) before adding complexity
- WebSocket implementation is critical for US2, US7, and real-time updates
- Redis Pub/Sub is critical for inter-service communication (competition → matchmaking)
- SQLC code generation (T017) must run after creating all SQL query files
- All database migrations must run before any service code can access database

