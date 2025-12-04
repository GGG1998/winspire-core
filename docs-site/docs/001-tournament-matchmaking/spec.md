# Feature Specification: Tournament Matchmaking System

**Feature Branch**: `001-tournament-matchmaking`  
**Created**: 2025-12-03  
**Status**: Draft  
**Input**: Build a matchmaking system for tournament competitions with bracket generation, round management, match handling, and real-time event flow. Default mode is single elimination 1v1, inspired by e-sports best practices.

## Clarifications

### Session 2025-12-03

- Q: How should byes be assigned when participant count is not a power of 2? → A: Random assignment - system randomly selects bye recipients from all participants
- Q: What should a player with a bye experience while waiting for next match? → A: Auto-advance + notification - player is notified of bye, automatically placed in next round slot, waits for opponent's match to complete
- Q: Can new players join after the tournament has started (queue behavior)? → A: Roster locked at start - no new players after tournament starts; bracket is final at generation
- Q: Should missing host controls (Publish, UpdateSettings, OpenRegistration, CloseRegistration) be in this spec? → A: No - keep matchmaking-focused; tournament lifecycle controls belong to existing tournament management system
- Q: Should referee assignment (RefereeAssigned event) be included? → A: Defer to future - skip for MVP; host handles disputes directly via existing controls (FR-025, FR-028)
- Q: Can player withdraw/forfeit after tournament starts? → A: Forfeit allowed - player can forfeit current/upcoming match (opponent gets walkover); "withdraw" command only allowed before tournament starts
- Q: What happens when player refreshes browser while in lobby? → A: Seamless return - player returns to same lobby state; ready status is preserved server-side
- Q: How to resolve disconnect during active match (CS:GO style)? → A: Single disconnect = disconnected player loses 1 point + 30s reconnect window; if no return = disqualification. Both disconnect = first to disconnect gives opponent 1 point; both get 30s to reconnect; first to return continues. Timer starts from app exit.
- Q: What happens if player shares lobby link with colleague (account sharing)? → A: User ID authentication required - lobby access requires authenticated session matching registered player's user ID; shared links denied for other users
- Q: How do players submit match scores? → A: Automatic from game API - system reads match results directly from game API; no manual submission needed (reduces disputes and fraud)
- Q: Can players dispute automatically retrieved API results? → A: No disputes allowed - API results are final; if API has technical issue, only host can override (no player dispute flow)

### Session 2025-12-04

- Q: What is the game API integration pattern for match results? → A: Polling with fraud validation - After match completes, game client sends score (with game-specific fraud rules) to Game API; Matchmaking service polls Game API to retrieve validated results
- Q: What observability/monitoring approach should be used? → A: AWS CloudWatch + minimal app metrics - Use AWS-provided infrastructure metrics (CPU, memory, health checks) plus emit critical application metrics (bracket generation time, match completion rates) to CloudWatch Logs and Metrics
- Q: What is the uptime/availability target for the matchmaking service? → A: Best effort - No formal SLA; focus on fast incident detection and recovery rather than high-availability infrastructure complexity for MVP

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Host Starts Tournament and Bracket Generates Automatically (Priority: P1)

A host who has created and published a tournament with registered participants wants to start the tournament. Upon starting, the system automatically generates the tournament bracket, creates all rounds, and assigns participants to first-round matches.

**Why this priority**: This is the core value proposition - without automatic bracket generation and match creation, there is no tournament matchmaking. Everything else depends on this flow working correctly.

**Independent Test**: Can be fully tested by creating a tournament with 8 registered participants, starting it, and verifying the bracket shows 4 first-round matches with all participants assigned.

**Acceptance Scenarios**:

1. **Given** a scheduled tournament with 8 confirmed participants, **When** the host clicks "Start Tournament", **Then** the system generates a single elimination bracket with 3 rounds (Quarter-finals, Semi-finals, Final) and 4 first-round matches with participants randomly assigned.

2. **Given** a tournament with 5 confirmed participants (non-power-of-2), **When** the host starts the tournament, **Then** the system generates a bracket where 3 participants receive first-round byes and 2 participants are matched in round 1.

3. **Given** a tournament in "registration_open" state with fewer than minimum required participants, **When** the host attempts to start, **Then** the system prevents starting and displays the minimum participant requirement.

---

### User Story 2 - Player Receives Match Assignment and Joins Lobby (Priority: P1)

A registered player whose tournament has started wants to know their opponent and prepare for their match. They receive a notification about their match assignment and can join the match lobby to see opponent details and indicate readiness.

**Why this priority**: Players must know who they're playing against and have a space to prepare - this is essential for any competitive tournament experience.

**Independent Test**: Can be tested by starting a tournament and verifying each assigned player receives match information and can access their lobby.

**Acceptance Scenarios**:

1. **Given** a started tournament with generated matches, **When** a player is assigned to a match, **Then** they receive notification with opponent name/handle and match details (round, position in bracket).

2. **Given** a player assigned to a match, **When** they join the match lobby, **Then** they see opponent information, match status, and a "Ready" button.

3. **Given** a player in the lobby, **When** they click "Ready", **Then** their ready status is visible to their opponent and the system.

4. **Given** a registered player shares their lobby link with an unauthorized user, **When** the unauthorized user attempts to access the lobby, **Then** they are denied with "Unauthorized - you are not a participant in this match" message.

---

### User Story 3 - Match Completes and Winner Advances (Priority: P1)

After a match is played, the winner needs to be recorded and automatically advance to the next round while the loser is eliminated from the tournament.

**Why this priority**: Tournament progression is fundamental - without match completion and advancement, the tournament cannot reach a conclusion.

**Independent Test**: Can be tested by completing a single match and verifying winner appears in next round while loser is marked eliminated.

**Acceptance Scenarios**:

1. **Given** a match in "started" state, **When** the match result is submitted (score or winner declaration), **Then** the match is marked "completed" with the recorded result.

2. **Given** a completed match, **When** the system processes the result, **Then** the winner is assigned to the next round's match slot and the loser's status changes to "eliminated".

3. **Given** a completed semi-final match, **When** both semi-finals are done, **Then** the final match is ready with both winners assigned.

---

### User Story 4 - Match Starts When Both Players Ready (Priority: P2)

Both players in a match have indicated readiness and the match should begin. The system tracks when both are ready and initiates the match.

**Why this priority**: Ensures fair play - both competitors should be prepared before the match begins. Not blocking for MVP but essential for good UX.

**Independent Test**: Can be tested by having two players in a lobby both click ready and verifying match transitions to "started" state.

**Acceptance Scenarios**:

1. **Given** both players in a match lobby have clicked "Ready", **When** the system detects both ready states, **Then** the match status changes to "started" and both players are notified.

2. **Given** one player is ready and waiting, **When** 2 minutes pass without opponent readying, **Then** the waiting player sees a countdown and can request walkover.

3. **Given** auto_force_ready is enabled for the tournament, **When** match start time arrives, **Then** the match starts automatically regardless of ready status.

---

### User Story 5 - Player Check-in Before Tournament Start (Priority: P2)

Registered players need to confirm their presence shortly before the tournament begins to ensure they're actually available to compete.

**Why this priority**: Reduces no-shows and allows for last-minute roster adjustments. Important for tournament quality but not blocking core flow.

**Independent Test**: Can be tested by opening check-in window and verifying only checked-in players are included in bracket generation.

**Acceptance Scenarios**:

1. **Given** a scheduled tournament with check-in enabled, **When** the check-in window opens (15 minutes before start), **Then** confirmed participants can check in.

2. **Given** a participant who has checked in, **When** the tournament starts, **Then** they are included in the bracket.

3. **Given** a confirmed participant who did not check in, **When** the tournament starts, **Then** they are excluded from the bracket and marked as "no-show".

---

### User Story 6 - Handle No-Show with Walkover (Priority: P2)

When a player fails to show up for their match within the allowed time window, their opponent should be awarded a walkover (automatic win) and advance.

**Why this priority**: Prevents tournaments from stalling due to absent players. Critical for tournament completion but dependent on match flow being established first.

**Independent Test**: Can be tested by having only one player ready in a match, waiting for timeout, and verifying walkover is granted.

**Acceptance Scenarios**:

1. **Given** a match where one player is ready and opponent hasn't joined lobby for 2 minutes, **When** timeout expires, **Then** the ready player is offered to claim walkover.

2. **Given** a walkover is granted, **When** the system processes it, **Then** the present player wins and advances, absent player is eliminated.

3. **Given** both players absent from lobby for 5 minutes, **When** timeout expires, **Then** host is notified and can manually resolve (assign winner or eliminate both).

---

### User Story 7 - Automatic Score Retrieval and Verification (Priority: P2)

After a match is played in the external game, the result needs to be automatically retrieved and recorded. The game client submits score to Game API (with fraud validation), then matchmaking service polls Game API to retrieve validated results.

**Why this priority**: Accurate score recording is essential for tournament integrity. Automatic retrieval with fraud detection reduces disputes and fraud. Depends on match start flow being complete.

**Independent Test**: Can be tested by completing a match in the game, game client sending score to Game API, and verifying matchmaking service automatically polls and retrieves validated result.

**Acceptance Scenarios**:

1. **Given** a match in "started" state, **When** the match completes in the external game, **Then** the matchmaking service begins polling Game API for results (5-second intervals).

2. **Given** the Game API has received and validated the score, **When** the matchmaking service polls for results, **Then** the validated result is returned, match is marked completed, and both players are notified.

3. **Given** the Game API is unavailable or polling times out (60 seconds), **When** the system detects the failure, **Then** the match is flagged for manual host entry and host is notified.

---

### User Story 8 - Host Override for API Issues (Priority: P3)

When the game API returns incorrect or suspicious results due to technical issues (e.g., wrong match session matched, API bug), the host needs to manually override the result to maintain tournament integrity.

**Why this priority**: Rare exception flow for handling API failures. Most matches complete correctly via API. Can be added after core flows work.

**Independent Test**: Can be tested by simulating an API error scenario and having host manually correct the result.

**Acceptance Scenarios**:

1. **Given** a completed match with API-retrieved results, **When** host detects a technical issue with the result, **Then** host can access the match and manually override the result.

2. **Given** a host is reviewing a match result, **When** they enter a corrected result, **Then** the match is updated with the new result and both players are notified.

3. **Given** a match with overridden results, **When** viewing match details, **Then** the system shows "Result manually corrected by host" indicator.

---

### User Story 9 - Tournament Completes with Winner Declaration (Priority: P2)

When the final match is completed, the tournament should automatically declare the winner and complete.

**Why this priority**: Natural conclusion to the tournament. Depends on match progression working through all rounds.

**Independent Test**: Can be tested by completing all matches through to final, verifying tournament status changes to "completed" with winner recorded.

**Acceptance Scenarios**:

1. **Given** the final match is completed, **When** the system processes the result, **Then** the winner is declared tournament champion and all participants are notified.

2. **Given** a tournament is completed, **When** viewing tournament details, **Then** full bracket results, final standings, and winner are displayed.

---

### Edge Cases

- **Odd number of participants**: System randomly assigns byes, advancing selected players without opponents to next round (equal probability for all participants). Bye recipients receive notification of auto-advancement and wait for their next opponent's match to complete.
- **Browser refresh / temporary disconnect in lobby**: Player seamlessly returns to same lobby state; ready status preserved server-side; no "rejoin" required
- **Single player disconnection during active match (CS:GO style)**: Disconnected player loses 1 round point (e.g., score was 1:0 → becomes 2:0). Match pauses, 30 second reconnect window starts from app exit. If player returns within 30s, match continues. If no return after 30s → disqualification (opponent wins by walkover).
- **Both players disconnect during active match**: First player to disconnect gives opponent 1 round point. Both players get 30 seconds to reconnect (timer starts from their respective app exit). First player to successfully reconnect continues; if opponent doesn't return within their 30s window → they are disqualified.
- **Host cancels mid-tournament**: Tournament cancelled, all active matches voided, participants notified
- **API returns incorrect result (technical issue)**: Host can manually override the API-retrieved result. System logs override with timestamp and host ID for audit. No player dispute mechanism - API is source of truth.
- **Player forfeits mid-tournament**: Player can forfeit their current/upcoming match (distinct from "withdraw" which is blocked after start); opponent receives walkover, forfeiting player is eliminated. No replacement player is assigned - bracket remains fixed.
- **All remaining players withdraw**: Tournament ends with no winner, marked as cancelled
- **Power failure / system restart**: Tournament state persisted, can resume from last known state
- **Late registration attempt**: Players attempting to register after tournament starts receive "registration closed" message; no queue or waitlist system exists
- **Account sharing / lobby link sharing**: If player shares their lobby URL with someone else, the unauthorized user is denied access. Lobby requires authenticated session matching the registered participant's user ID. System displays "Unauthorized - you are not a participant in this match" error.
- **Game API unavailable or polling timeout**: If Game API fails to return results, times out (60 seconds), or fraud validation fails, match is flagged for manual host entry. Host receives notification with match details and can manually enter results. Tournament progression continues via manual path.

## Requirements *(mandatory)*

### Functional Requirements

**Tournament Lifecycle**
- **FR-001**: System MUST support tournament states: draft, scheduled, registration_open, registration_closed, started, completed, cancelled
- **FR-002**: System MUST prevent state transitions that violate the tournament lifecycle (e.g., cannot start a completed tournament)
- **FR-003**: System MUST lock participant roster when tournament starts (no new registrations after bracket generation)
- **FR-004**: System MUST support configurable minimum and maximum participant counts
- **FR-005**: System MUST support automatic registration closure when maximum capacity is reached

**Bracket Management**
- **FR-006**: System MUST generate single elimination brackets for any participant count within tournament's configured limits (minimum 2 participants required)
- **FR-007**: System MUST calculate and randomly assign byes when participant count is not a power of 2 (any participant has equal chance of receiving a bye)
- **FR-008**: System MUST randomize participant seeding for bracket assignment (default behavior)
- **FR-009**: System MUST create all rounds and match slots upon bracket generation
- **FR-010**: System MUST track bracket position for each match (round number, match number within round)

**Match Management**
- **FR-011**: System MUST support match states: pending, ready, started, paused, completed, disputed, cancelled
- **FR-012**: System MUST track both participants for each match
- **FR-013**: System MUST record match results including winner and scores (round-by-round if applicable)
- **FR-014**: System MUST automatically advance winners to their next match slot
- **FR-015**: System MUST automatically mark losers as eliminated (in single elimination mode)
- **FR-016**: System MUST support walkover assignment for no-shows

**Disconnect Handling (CS:GO Style)**
- **FR-016a**: System MUST detect player disconnection during active match and pause the match
- **FR-016b**: System MUST award 1 round point to online opponent when player disconnects
- **FR-016c**: System MUST provide 30 second reconnect window (timer starts from app exit timestamp)
- **FR-016d**: System MUST disqualify player (walkover to opponent) if reconnect window expires without return
- **FR-016e**: System MUST track disconnect timestamps independently for each player (for dual disconnect scenarios)
- **FR-016f**: System MUST resume match when disconnected player returns within 30 second window

**Player Participation**
- **FR-017**: System MUST track participant status: registered, confirmed, checked_in, active, eliminated, withdrawn
- **FR-018**: System MUST support player check-in within configurable window before tournament start
- **FR-019**: System MUST track player ready status within match lobbies (persisted server-side, survives browser refresh)
- **FR-020**: System MUST allow players to withdraw from tournament before it starts (WithdrawFromTournament - blocked once tournament is in 'started' state)
- **FR-020a**: System MUST allow players to forfeit their current/upcoming match during tournament (ForfeitMatch - gives opponent walkover)
- **FR-021**: System MUST notify players of their match assignments
- **FR-022**: System MUST notify bye recipients of their automatic advancement and when their next match becomes ready

**Security & Access Control**
- **FR-023**: System MUST authenticate lobby access by matching authenticated user ID to registered participant ID
- **FR-024**: System MUST deny lobby access to users not registered as participants in that match
- **FR-025**: System MUST prevent session sharing/account impersonation (lobby URLs cannot be used by unauthorized users)

**Score and Results**
- **FR-026**: System MUST poll Game API for match results after match completes in external game (5-second polling interval, 60-second timeout)
- **FR-026a**: Game API MUST validate submitted scores using game-specific fraud detection rules before returning results to matchmaking service
- **FR-027**: System MUST handle Game API unavailability or polling timeout (fallback to manual host entry if API fails or 60s timeout expires)
- **FR-028**: System MUST support host override for API-retrieved results (for technical issues only; no player dispute mechanism)
- **FR-029**: System MUST validate fraud detection status from Game API response before marking match as completed
- **FR-030**: System MUST log when results are manually overridden by host (audit trail)

**Host Controls**
- **FR-031**: Host MUST be able to start tournament when minimum participants are met
- **FR-032**: Host MUST be able to manually assign walkover in no-show situations
- **FR-033**: Host MUST be able to manually enter match results if game API is unavailable
- **FR-034**: Host MUST be able to override API-retrieved results for technical issues (with audit logging)
- **FR-035**: Host MUST be able to cancel tournament (with restrictions during active matches)
- **FR-036**: Host MUST receive notifications for API failures and issues requiring intervention

**Observability & Monitoring**
- **FR-037**: System MUST emit CloudWatch metrics for critical performance indicators: bracket generation duration, match state update latency, tournament completion rate, score submission success rate
- **FR-038**: System MUST log all state transitions (tournament started, match created, match completed, errors) to CloudWatch Logs in structured format
- **FR-039**: System MUST expose health check endpoint for ECS container health monitoring
- **FR-040**: System MUST emit CloudWatch alarms when success criteria thresholds are violated (SC-001 through SC-012)

**Reliability & Availability**
- **FR-041**: System operates on best-effort availability (no formal uptime SLA for MVP)
- **FR-042**: System MUST persist tournament state to database to allow recovery after service restarts
- **FR-043**: System MUST implement graceful degradation: if real-time features fail, tournament can continue via manual host intervention
- **FR-044**: System SHOULD recover from crashes within 2 minutes via ECS container auto-restart

### Key Entities

- **Tournament**: Competition event with defined rules, participant limits, schedule, and current state. Contains metadata like name, description, game reference, and timing settings.

- **Bracket**: The elimination structure for a tournament. Defines the progression path from first round to finals. Linked to exactly one tournament.

- **Round**: A stage within the bracket (e.g., Round of 16, Quarter-finals, Semi-finals, Final). Contains multiple matches that must complete before next round begins.

- **Match**: A single competition between two participants. Tracks participants, ready states, scores, result, and current status. Belongs to one round.

- **Participant**: A player registered for a tournament. Tracks registration status, check-in status, current position in bracket, and elimination state.

- **Match Lobby**: Virtual waiting room for a specific match where participants gather, indicate readiness, and receive match start signals.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Bracket generation completes within 2 seconds for large tournaments (performance tested with 128 participants)
- **SC-002**: 95% of tournaments complete without requiring host intervention for technical issues
- **SC-003**: Players receive match assignment notifications within 5 seconds of bracket generation
- **SC-004**: Match state updates (ready, started, completed) propagate to all viewers within 3 seconds
- **SC-005**: System supports 50 concurrent tournaments with 64 participants each without degradation
- **SC-006**: No-show detection and walkover flow completes within configured timeout + 30 seconds
- **SC-007**: 99% of score submissions process correctly on first attempt (no false conflicts)
- **SC-008**: Tournament completion rate (started to finished) exceeds 90% for tournaments with minimum viable participants
- **SC-009**: Average time from "both players ready" to "match started" is under 5 seconds
- **SC-010**: Players can view their next match opponent within 10 seconds of previous match completing
- **SC-011**: 95% of match results are automatically retrieved from Game API via polling without requiring manual host entry
- **SC-012**: Game API result retrieval via polling completes within 10 seconds of match completion 95% of the time

## Assumptions

- Players have authenticated accounts in the platform before joining tournaments
- Tournament creation and basic tournament settings already exist in the system
- **Tournament lifecycle management (publish, settings update, registration open/close) exists in the tournament management system** - this matchmaking spec focuses on bracket generation, match handling, and tournament progression
- **Game integration**: Actual gameplay happens in external game system; game clients submit scores to Game API with game-specific fraud validation rules; matchmaking service polls Game API to retrieve validated results
- Game API provides endpoints to fetch match results with winner/loser, scores, and fraud validation status
- Real-time communication infrastructure is available for notifications
- A host role with appropriate permissions exists in the system

## Out of Scope

- Tournament creation (CreateTournament command)
- Tournament publishing (PublishTournament command - draft → scheduled)
- Tournament settings management (UpdateTournamentSettings command)
- Manual registration open/close (OpenRegistration, CloseRegistration commands)
- Prize distribution (PrizesDistributed - belongs to Rewards bounded context)
- Referee assignment (RefereeAssigned - deferred to future; host handles disputes for MVP)
- Player dispute mechanism (DisputeMatch - API results are source of truth; only host can override for technical issues)
