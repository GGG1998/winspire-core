# Tasks: Competition Lifecycle Host Streaming

**Input**: Design documents from `/docs-site/docs/001-competition-spec/`
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/openapi.yaml`, `quickstart.md`

**Tests**: Integration tests per user story ensure independent verification of REST + SSE flows.

**Organization**: Dependency-ordered phases (Setup → Foundational → US1 → US2 → Polish) with checklist-format tasks mapped to user stories.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish repository structure, tooling, and local dev workflow.

- [x] T001 Scaffold service directories `services/competition-host-stream/{cmd/competition-host-stream,internal/{config,http,projections,sse,store},dev,migrations}` according to `plan.md`.
- [x] T002 Initialize Go module + toolchain with Gin, pgx, Redis, SSE, testify, and testcontainers in `services/competition-host-stream/go.mod`.
- [x] T003 [P] Register the module inside root `go.work` and ensure replace directives cover `services/competition-host-stream`.
- [x] T004 [P] Create local Postgres + Redis stack in `services/competition-host-stream/dev/compose.yaml` aligned with `quickstart.md`.
- [x] T005 [P] Author `services/competition-host-stream/Makefile` with `fmt`, `lint`, `test`, `run`, and `sqlc` targets referenced by quickstart instructions.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that all user stories depend on. No story work can begin until these are complete.

- [x] T006 Implement env + config loader (Postgres, Redis, SSE limits, auth) in `services/competition-host-stream/internal/config/config.go`.
- [x] T007 Build base Gin server with logging, recovery, auth middleware, and error envelope in `services/competition-host-stream/internal/http/server.go`.
- [x] T008 Define Atlas/SQL migrations for projections (`cup_host_views`, `tournament_host_views`, `attendance_snapshots`, `match_lobby_views`, `host_subscriptions`) in `services/competition-host-stream/migrations/*.sql` per `data-model.md`.
- [ ] T009 [P] Configure `services/competition-host-stream/sqlc.yaml` and author projection queries in `services/competition-host-stream/internal/store/sqlc/queries.sql`, then run `sqlc generate`.
- [x] T010 [P] Implement Postgres + Redis clients with health checks + pooling in `services/competition-host-stream/internal/store/clients.go`.
- [x] T011 [P] Create SSE broker + subscription registry scaffolding (Redis fan-out, Last-Event-ID cursors) in `services/competition-host-stream/internal/sse/broker.go`.

---

## Phase 3: User Story 1 - Cup-to-Tournament Orchestration Clarity (Priority: P1) 🎯 MVP

**Goal**: Hosts can inspect Cup/Tournament projections, dependencies, and attendance rollups without engineering assistance.
**Independent Test**: Seed Cup + Tournament data, call `/v1/hosts/{hostId}/cups/{cupId}` and `/v1/hosts/{hostId}/tournaments/{tournamentId}`, then connect to `/v1/hosts/{hostId}/streams/cup/{cupId}` to verify dependency health + attendance deltas stream in under 2 s.

### Implementation

- [x] T012 [P] [US1] Build `cup_host_views` projector + repository in `services/competition-host-stream/internal/projections/cup.go` using Cup + Participation events.
- [x] T013 [P] [US1] Build `tournament_host_views` projector + repository in `services/competition-host-stream/internal/projections/tournament.go` including lineage to Cup.
- [x] T014 [US1] Implement attendance snapshot worker covering `OwnCupParticipation`/`OwnTournamentParticipation` in `services/competition-host-stream/internal/projections/attendance.go`.
- [x] T015 [US1] Expose GET/PATCH Cup + Tournament handlers with validation (StageStatus, DependencyHealth) in `services/competition-host-stream/internal/http/handlers/cup_tournament.go`.
- [x] T016 [US1] Extend SSE handler in `services/competition-host-stream/internal/sse/stream_handler.go` to lease `host_subscriptions` and emit Cup/Tournament frames.
- [ ] T017 [P] [US1] Write integration test `services/competition-host-stream/internal/tests/us1_cup_to_tournament_test.go` covering projections + SSE replay cursors.
- [x] T018 [US1] Document Cup/Tournament aggregates, roles, GraphQL mutations, and supplier relationships (customer-supplier map + Mermaid context) in `docs-site/docs/001-competition-spec/spec.md`.

---

## Phase 4: User Story 2 - Match & Eligibility Sync Reference (Priority: P2)

**Goal**: Operations leads can trace matchmaking + eligibility ownership via REST + SSE without reverse-engineering services.
**Independent Test**: Seed lineup + queue data, fetch `/v1/hosts/{hostId}/tournaments/{tournamentId}/matches`, and subscribe to `/v1/hosts/{hostId}/streams/tournament/{tournamentId}` to identify owning bounded context for a stuck offer using event payloads.

### Implementation

- [x] T019 [P] [US2] Implement `match_lobby_views` projector consuming matchmaking queue events in `services/competition-host-stream/internal/projections/match.go`.
- [x] T020 [P] [US2] Overlay Participation + `AllowedTournamentAction` data into tournament lineup status in `services/competition-host-stream/internal/projections/participation.go`.
- [x] T021 [US2] Serve `/v1/hosts/{hostId}/tournaments/{tournamentId}/matches` handler returning lobby + queue metadata from `internal/http/handlers/match.go`.
- [ ] T022 [US2] Route `MatchmakingQueueUpdated` + `TournamentParticipationUpdate` events through SSE broker in `services/competition-host-stream/internal/sse/event_router.go`.
- [ ] T023 [P] [US2] Add integration test `services/competition-host-stream/internal/tests/us2_match_eligibility_test.go` validating stuck-offer traceability.
- [x] T024 [US2] Expand `docs-site/docs/001-competition-spec/spec.md` with Match/Participation dependencies, AllowedAction exposure, and troubleshooting checklist.

---

## Phase 5: Polish & Cross-Cutting Concerns

- [x] T025 [P] Capture cross-domain components (`CompetitionContext`, `AllowedTournamentAction`, `LobbyInformation`, `GameSession`) and ownership table in `docs-site/docs/001-competition-spec/spec.md`.
- [x] T026 [P] Add full textual + Mermaid context map plus edge-case notes/refactoring suggestions (lobby service extraction, VO validation) to `docs-site/docs/001-competition-spec/spec.md`.
- [ ] T027 Harden App Runner env + secret handling and Redis reconnect logic in `services/competition-host-stream/internal/config/runtime.go`.
- [ ] T028 Run `docs-site/docs/001-competition-spec/quickstart.md` steps end-to-end (sqlc generate, atlas migrate, go test, SSE curl) and record verification log in `docs-site/docs/001-competition-spec/quickstart.md`.

---

## Dependencies & Execution Order

- **Phase sequencing**: Phase 1 → Phase 2 → Phase 3 (US1) → Phase 4 (US2) → Phase 5.
- **Story dependencies**: US1 depends on Foundational; US2 depends on Foundational + US1 projections/SSE; Polish depends on completion of targeted stories.
- **Task dependencies**:
  - T006 precedes T007–T011 (server relies on config + infra).
  - T012–T016 depend on sqlc + migrations (T008–T010).
  - T019–T023 depend on Cup/Tournament baseline (T012–T016).
  - Documentation tasks (T018, T024, T025, T026) require corresponding implementations to be available.

### Parallel Opportunities

- Setup: T003–T005 can run in parallel after T001–T002.
- Foundational: T009–T011 operate on separate files once migrations (T008) exist.
- US1: T012 & T013 can proceed concurrently; T017 can begin after repositories stubbed; T018 drags after endpoints defined.
- US2: T019 & T020 parallelize; T023 can start after projections compile; T024 finalizes while SSE routing (T022) stabilizes.
- Polish: T025 & T026 edit documentation in distinct sections; T027 runs concurrently with T028 (doc validation).

### Parallel Execution Example — User Story 1

```
Task T012: Implement cup_host_views projector
Task T013: Implement tournament_host_views projector
Task T017: Integration test for Cup↔Tournament SSE replay
```

### Parallel Execution Example — User Story 2

```
Task T019: Build match_lobby_views projector
Task T020: Overlay participation + allowed actions
Task T023: Integration test for stuck matchmaking offer
```

---

## Implementation Strategy

- **MVP (US1 only)**:
  1. Complete Phases 1–2 to solidify infrastructure.
  2. Deliver all US1 tasks (T012–T018) and validate via SSE + REST integration test.
  3. Demo Cup-to-Tournament orchestration clarity before proceeding.
- **Incremental Delivery**:
  1. After MVP, complete US2 tasks (T019–T024) to unlock match troubleshooting.
  2. Execute Polish tasks (T025–T028) for cross-domain documentation, edge cases, and operations hardening.
  3. Each story remains independently testable for staged deployments.

---

