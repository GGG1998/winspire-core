# Feature Specification: Tournament Pre-Lobby Backend

**Feature Branch**: `004-tournament-prelobby-backend`  
**Created**: 2025-12-05  
**Status**: Draft  
**Input**: Add tournament pre-lobby backend endpoints to matchmaking service - REST API and WebSocket for waiting room before tournament starts, with grace period and participant tracking

## Clarifications

### Session 2025-12-05

- Q: Should pre-lobby participant connections and grace period state be stored persistently, or is ephemeral (in-memory only) state acceptable? → A: Hybrid - WebSocket connections in-memory, grace period state in database
- Q: How should the final participant list be provided to bracket generation - as a snapshot persisted when grace period ends, or queried from currently connected WebSockets? → A: Persist participant list snapshot to database when grace period ends
- Q: How does the matchmaking service receive notification that the host triggered tournament start? → A: Redis pub/sub - competition service publishes, matchmaking subscribes

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Pre-Lobby State API (Priority: P1)

Frontend applications need to retrieve the current state of a tournament's pre-lobby waiting room, including which participants are currently present and waiting for the tournament to begin.

**Why this priority**: This is the foundation for pre-lobby functionality - without this REST endpoint, frontends cannot display any pre-lobby information to users.

**Independent Test**: Can be tested by calling GET `/v1/tournaments/:id/lobby` and verifying the response contains tournament metadata, participant list, and lobby status.

**Acceptance Scenarios**:

1. **Given** a tournament scheduled to start in 3 minutes, **When** a registered participant requests the pre-lobby state, **Then** the system returns tournament details, countdown time, empty participant list (if no one connected yet), and status "waiting".

2. **Given** a tournament pre-lobby with 5 connected participants, **When** any user requests the pre-lobby state, **Then** the system returns the list of all 5 currently connected participants with their display names and join timestamps.

3. **Given** a tournament that hasn't reached its join window (more than 15 minutes before start), **When** a user requests pre-lobby state, **Then** the system returns an error indicating the lobby is not yet open.

4. **Given** a tournament in grace period with 8 participants, **When** pre-lobby state is requested, **Then** the system returns status "grace_period", grace period end time, and current participant count.

5. **Given** a tournament where bracket generation has completed, **When** pre-lobby state is requested, **Then** the system returns status "started" indicating the waiting room phase is over.

---

### User Story 2 - WebSocket Connection Management (Priority: P1)

Players joining the pre-lobby need real-time presence tracking so all participants can see who else is waiting for the tournament to begin.

**Why this priority**: Real-time participant list is core to the waiting room UX - players need to see others joining to know the tournament will have enough participants.

**Independent Test**: Can be tested by connecting multiple WebSocket clients to `/v1/tournaments/:id/lobby` and verifying each connection receives participant_joined events for other connections.

**Acceptance Scenarios**:

1. **Given** a tournament pre-lobby with no connected participants, **When** a registered player establishes a WebSocket connection, **Then** the system sends the initial prelobby_state message with empty participant list.

2. **Given** Player A is connected to pre-lobby, **When** Player B connects, **Then** Player A receives a participant_joined event containing Player B's information within 1 second.

3. **Given** Player A and Player B are both connected, **When** Player B disconnects (closes browser), **Then** Player A receives a participant_left event with Player B's ID within 3 seconds.

4. **Given** a player without tournament registration, **When** they attempt to connect to pre-lobby WebSocket, **Then** the connection is refused with authentication error.

5. **Given** 10 players connected to pre-lobby, **When** a new player connects, **Then** they receive the complete current state with all 10 existing participants in the initial message.

---

### User Story 3 - Grace Period Management (Priority: P1)

When a tournament host triggers tournament start, the system must enforce a 30-second grace period during which late-arriving players can still join and be included in bracket generation.

**Why this priority**: Grace period prevents players who are slightly late from being excluded - critical for fair tournament start and reducing no-shows.

**Independent Test**: Can be tested by triggering tournament start, connecting a player during the 30-second window, and verifying they're included when bracket generation completes.

**Acceptance Scenarios**:

1. **Given** tournament start is triggered with 8 players in pre-lobby, **When** grace period begins, **Then** all connected players receive grace_period_started event with 30-second countdown.

2. **Given** grace period is active with 8 players, **When** a 9th registered player connects, **Then** all players receive roster_updated event showing participant count increased to 9.

3. **Given** grace period is active with 10 players, **When** one player disconnects, **Then** all remaining players receive roster_updated event showing count decreased to 9.

4. **Given** grace period ends with 12 players present, **When** the 30 seconds elapse, **Then** bracket generation begins with exactly those 12 players (no more late joins accepted).

5. **Given** a player attempts to connect to pre-lobby, **When** grace period has already ended and bracket generation started, **Then** their connection is refused with "tournament in progress" error.

---

### User Story 4 - Tournament-Matchmaking Integration (Priority: P2)

The pre-lobby system needs to verify participant eligibility by checking with the competition service whether users are registered for the tournament.

**Why this priority**: Authorization is essential for security but can be implemented after core pre-lobby mechanics are working.

**Independent Test**: Can be tested by attempting to connect with a registered user (succeeds) and an unregistered user (fails).

**Acceptance Scenarios**:

1. **Given** a user registered for Tournament X, **When** they request pre-lobby access for Tournament X, **Then** the system verifies their registration with competition service and grants access.

2. **Given** a user NOT registered for Tournament X, **When** they attempt to connect to Tournament X pre-lobby, **Then** the system denies access with "not registered" error.

3. **Given** the competition service is temporarily unavailable, **When** a user attempts to connect, **Then** the system returns a service unavailable error and suggests retrying.

4. **Given** a tournament with status "draft" (not published), **When** any user requests pre-lobby access, **Then** the system denies access indicating tournament hasn't started registration.

---

### User Story 5 - Activity Feed Events (Priority: P3)

Pre-lobby participants want to see a chronological feed of recent activity (joins, leaves, announcements) to understand what's happening in the waiting room.

**Why this priority**: Activity feed enhances UX but is not critical for core functionality - participants can function without historical context.

**Independent Test**: Can be tested by performing several join/leave actions and verifying the activity feed in the state response reflects these events in order.

**Acceptance Scenarios**:

1. **Given** empty pre-lobby, **When** 3 players join in sequence, **Then** subsequent state requests include activity feed with 3 "player joined" entries in chronological order.

2. **Given** pre-lobby with existing activity feed, **When** a player leaves, **Then** a "player left" entry is appended to the activity feed.

3. **Given** grace period starts, **When** state is requested, **Then** activity feed includes a "Grace period started - late arrivals accepted for 30s" entry.

4. **Given** activity feed with 50+ entries, **When** state is requested, **Then** only the most recent 20 entries are returned to limit response size.

---

### Edge Cases

- **Tournament start triggered with 0 participants in pre-lobby**: System cancels tournament start, sends "insufficient participants" event, sets tournament status back to "scheduled"
- **All players disconnect during grace period**: If count drops to 0, cancel tournament start and revert to scheduled status
- **Player attempts to join multiple pre-lobbies simultaneously**: Each WebSocket connection is independent; player can be in multiple waiting rooms (allowed)
- **Grace period active, player connects then immediately disconnects**: Roster count increments then decrements; player is NOT included in final bracket
- **Competition service returns stale registration data**: System caches registration status for 30 seconds to handle temporary inconsistency
- **WebSocket connection timeout during tournament start**: System treats as participant leaving; if during grace period, they're excluded from bracket
- **Host triggers start while grace period already active**: Ignore duplicate start command, continue with current grace period
- **Player refreshes browser during grace period**: New WebSocket connection is established, treated as new participant (seamless from user perspective)
- **Bracket generation fails mid-process**: Grace period is extended by 30 seconds, participants notified of delay
- **Minimum participant count not met when grace period ends**: Tournament start is cancelled, all players notified and returned to waiting status

## Requirements *(mandatory)*

### Functional Requirements

**Pre-Lobby State Management**
- **FR-001**: System MUST provide GET `/v1/tournaments/:id/lobby` endpoint returning current pre-lobby state including tournament metadata, participant list, status, and grace period information
- **FR-002**: System MUST track which participants are currently connected to pre-lobby in real-time using ephemeral (in-memory) storage
- **FR-003**: System MUST include participant count in pre-lobby state response
- **FR-004**: System MUST return tournament start time, minimum participant count, and current status in state response
- **FR-005**: System MUST maintain activity feed of recent pre-lobby events (joins, leaves, system messages) limited to most recent 20 entries in persistent storage

**WebSocket Real-Time Updates**
- **FR-006**: System MUST provide WebSocket endpoint at `/v1/tournaments/:id/lobby` for real-time pre-lobby updates
- **FR-007**: System MUST send `prelobby_state` message immediately upon WebSocket connection containing complete current state
- **FR-008**: System MUST broadcast `participant_joined` event to all connected participants when a new player connects
- **FR-009**: System MUST broadcast `participant_left` event to all connected participants when a player disconnects
- **FR-010**: System MUST detect WebSocket disconnection within 5 seconds and remove participant from active roster
- **FR-011**: System MUST allow same user to reconnect to pre-lobby if they disconnect and return

**Grace Period Logic**
- **FR-012**: System MUST subscribe to Redis pub/sub channel for tournament start events and initiate 30-second grace period when `tournament_start` event is received, persisting grace period state (start time, end time, status) to database
- **FR-013**: System MUST broadcast `grace_period_started` event to all connected participants with end timestamp when grace period begins
- **FR-014**: System MUST accept new participant connections during grace period and include them in roster updates
- **FR-015**: System MUST broadcast `roster_updated` event when participant count changes during grace period
- **FR-016**: System MUST persist complete participant roster snapshot (user IDs, display names) to database when grace period ends - this becomes the immutable source of truth for bracket generation
- **FR-017**: System MUST trigger bracket generation using the persisted participant snapshot (not live WebSocket connections) after grace period completes
- **FR-018**: System MUST reject pre-lobby connections attempted after grace period has ended

**Authorization & Security**
- **FR-019**: System MUST verify tournament registration status with competition service before allowing pre-lobby access
- **FR-020**: System MUST deny pre-lobby access to users not registered for the tournament
- **FR-021**: System MUST validate JWT authentication token for all pre-lobby requests
- **FR-022**: System MUST verify tournament exists and is in appropriate status ("scheduled", "registration_open", "registration_closed") before allowing pre-lobby access
- **FR-023**: System MUST reject pre-lobby access for tournaments in "draft", "completed", or "cancelled" status

**Tournament State Transitions**
- **FR-024**: System MUST support tournament status: "waiting" (pre-start), "grace_period" (30s window), "generating_bracket" (processing), "started" (matches assigned)
- **FR-025**: System MUST cancel tournament start if participant count drops to 0 during grace period
- **FR-026**: System MUST cancel tournament start if participant count is below minimum when grace period ends
- **FR-027**: System MUST broadcast `match_assigned` event to each participant when their bracket match is determined

**Error Handling**
- **FR-028**: System MUST return appropriate HTTP error codes: 404 (tournament not found), 403 (not registered), 409 (tournament not in valid state for pre-lobby)
- **FR-029**: System MUST handle competition service unavailability gracefully with 503 Service Unavailable and retry suggestion
- **FR-030**: System MUST log all pre-lobby connection attempts, disconnections, and errors for debugging

### Key Entities

- **PreLobbyState**: Represents current state of tournament waiting room including tournament metadata (ID, name, start time, minimum participants), list of connected participants, status (waiting/grace_period/started), grace period end time (if active), and activity feed

- **PreLobbyParticipant**: Represents a connected participant including user ID, display name, avatar URL, and join timestamp

- **GracePeriod**: Represents the 30-second window after tournament start including start time, end time, initial participant count, and whether it's currently active

- **ParticipantSnapshot**: Persisted list of participants present when grace period ended, including user ID, display name, and avatar URL for each participant - used as immutable input to bracket generation

- **ActivityFeedItem**: Represents a single event in the pre-lobby history including event type (participant_joined/participant_left/system_message), message text, timestamp, and optional participant name

## Success Criteria *(mandatory)*

### Measurable Outcomes

**Measurement Context**: All timing criteria measured under standard conditions: 50ms network latency, warm server cache, PostgreSQL database with proper indexes, Redis for session storage.

- **SC-001**: Pre-lobby state API responds within 200ms for 95th percentile of requests
- **SC-002**: WebSocket `participant_joined` events propagate to all connected clients within 1 second
- **SC-003**: System supports 50 concurrent participants in a single pre-lobby without performance degradation
- **SC-004**: WebSocket disconnection is detected and roster updated within 5 seconds
- **SC-005**: Grace period countdown accuracy is within ±1 second of specified 30-second duration
- **SC-006**: 99% of eligible participants successfully connect to pre-lobby on first attempt
- **SC-007**: Zero participants are incorrectly included or excluded from bracket after grace period ends
- **SC-008**: System handles tournament start cancellation (due to insufficient participants) within 2 seconds
- **SC-009**: Authorization check against competition service completes within 500ms for 95th percentile
- **SC-010**: Activity feed maintains correct chronological order for 100% of events

## Assumptions

- Competition service provides REST API endpoint to verify tournament registration status
- Tournament start event is published via event bus (Redis pub/sub) by competition service when host triggers start
- Bracket generation is handled by existing matchmaking service logic after grace period completes
- JWT authentication middleware is already implemented and available
- PostgreSQL database is used for persistent storage of grace period state, activity feed, and tournament status
- In-memory storage is used for tracking active WebSocket connections (ephemeral, rebuilt on service restart)
- Redis pub/sub is used for cross-service events: competition service publishes `tournament_start` event when host triggers start, matchmaking service subscribes and initiates grace period
- WebSocket infrastructure (Hub pattern) is already implemented in matchmaking service
- Tournament participants have display names and optional avatar URLs available from user service
- Service restart during grace period is acceptable edge case - participants must reconnect (WebSocket connections are inherently ephemeral)

## Out of Scope

- Bracket generation algorithm (already implemented in matchmaking service)
- Tournament creation and registration management (handled by competition service)
- Player chat or messaging in pre-lobby (spec only requires activity feed, not player-to-player communication)
- Spectator access to pre-lobby (only registered participants)
- Multiple lobby rooms per tournament (one waiting room per tournament)
- Pre-lobby analytics or metrics dashboard
- Tournament pause/resume functionality
- Host controls within pre-lobby UI (host manages from tournament detail page)
- Email or push notifications when grace period starts
- Pre-lobby customization (themes, banners) - future feature
