# Feature Specification: Matchmaking Lobby Frontend

**Feature Branch**: `003-matchmaking-lobby-frontend`  
**Created**: 2025-12-04  
**Status**: Draft  
**Input**: Frontend for tournament matchmaking system: lobby entry, player avatars display, bracket view, matches tab, and game iframe integration based on spec.md from 001-tournament-matchmaking

## Clarifications

### Session 2025-12-04

- Q: What layout style should the match lobby use? → A: Full-screen immersive layout with player avatars at top center (you vs opponent), game iframe centered below
- Q: Should bracket visualization be interactive? → A: Yes - clickable matches that can navigate to match lobby when user is a participant
- Q: How should real-time updates be delivered to frontend? → A: WebSocket/SSE for match status changes, ready states, and tournament progression
- Q: What happens when a player isn't assigned to any match yet (has a bye)? → A: Show "Waiting for next round" state with countdown/info about their auto-advancement
- Q: When players click "Join" button (5 minutes before tournament), where should they land in the frontend? → A: Tournament pre-lobby page - shows countdown to start, participant list, chat (waiting room before matches created)
- Q: When tournament starts and bracket is generated, how should players in pre-lobby be notified of their match assignment? → A: WebSocket event `match_assigned` with match ID, auto-redirect after 2s delay
- Q: The pre-lobby includes a "chat/activity feed" - what functionality should this provide? → A: Activity feed only - shows participant joins/leaves, tournament updates, no player messaging
- Q: The tournament pre-lobby at `/tournaments/:tournamentId/lobby` - who can access this page? → A: Registered participants only (requires authentication and tournament registration)
- Q: Player is in pre-lobby waiting room, then closes browser tab or refreshes page. What should happen? → A: Seamless return to pre-lobby, preserve state via server
- Q: Host clicks "Start Tournament" to begin bracket generation. During the generation process, a player joins the pre-lobby late. What happens? → A: Grace period exists (30 seconds) - late arrivals during bracket generation can still be included
- Q: During the 30-second grace period, the frontend pre-lobby shows "Finalizing bracket..." What should happen if roster changes (new player joins)? → A: Show real-time participant count updates ("8 players... 9 players...") during grace period
- Q: What if during the 30-second grace period, participants leave the pre-lobby (close browser, disconnect)? → A: Roster remains fluid during grace period - departures reduce participant count, only present players at grace period end are included
- Q: Grace period ends, bracket is finalized. How are late-arriving players (after grace period) informed they cannot join? → A: Toast notification + redirect to detail page

## User Scenarios & Testing *(mandatory)*

### User Story 0 - Player Joins Tournament Pre-Lobby (Priority: P1)

A registered player clicks "Join" button (appears 5 minutes before tournament start) and enters a waiting room where they can see other participants and wait for the tournament to begin.

**Why this priority**: This is the entry point for all players - without pre-lobby, players have no clear waiting state before matches are created.

**Independent Test**: Can be tested by clicking "Join" button before tournament starts, verifying pre-lobby displays with countdown and participant list.

**Acceptance Scenarios**:

1. **Given** a scheduled tournament 5 minutes before start time, **When** a registered player clicks "Join" button, **Then** they navigate to `/tournaments/:tournamentId/lobby` and see tournament pre-lobby page.

2. **Given** a player in the tournament pre-lobby, **When** viewing the page, **Then** they see countdown timer to tournament start, list of joined participants with avatars, and an activity feed showing participant joins/leaves.

3. **Given** a player in the pre-lobby, **When** the tournament starts (host clicks "Start Tournament"), **Then** they receive notification "Tournament Starting - Generating Bracket..." and transition to their match assignment.

4. **Given** a registered player joins pre-lobby during the 30-second grace period after tournament start, **When** bracket generation is still in progress, **Then** they are included in the bracket and receive match assignment notification.

4a. **Given** tournament start has been triggered and grace period is active, **When** viewing the pre-lobby, **Then** players see "Finalizing bracket..." message with real-time participant count (e.g., "8 players... 9 players..." as new players join).

5. **Given** the tournament starts and bracket is generated, **When** player receives match assignment, **Then** they are automatically redirected to `/lobby/:tournamentId/match/:matchId` with their match.

6. **Given** a player who received a bye, **When** bracket generation completes, **Then** they see "You have a bye - Waiting for Round 2" message instead of match redirect.

7. **Given** a registered player attempts to join pre-lobby after grace period has ended, **When** they navigate to `/tournaments/:tournamentId/lobby`, **Then** they see toast notification "Tournament in progress - registration closed" and are redirected to tournament detail page.

8. **Given** a non-registered user attempts to access pre-lobby, **When** they navigate to `/tournaments/:tournamentId/lobby`, **Then** they see "Unauthorized - you must be registered for this tournament" error message.

---

### User Story 1 - Player Enters Match Lobby (Priority: P1)

A player who has been assigned to a match in a tournament wants to join their match lobby. They navigate to the lobby page where they can see their opponent, indicate readiness, and prepare for the match.

**Why this priority**: This is the core value proposition - the match lobby is where all player interaction happens. Without a functional lobby, no matches can be played.

**Independent Test**: Can be tested by assigning a player to a match, having them navigate to the lobby URL, and verifying they see opponent info and ready controls.

**Acceptance Scenarios**:

1. **Given** a player assigned to a match in a started tournament, **When** they navigate to `/lobby/:tournamentId/match/:matchId`, **Then** they see the lobby page with their avatar/nick on the left and opponent on the right at top center.

2. **Given** a player in the match lobby, **When** they view the header section, **Then** they see their own avatar and display name, a "VS" separator, and their opponent's avatar and display name.

3. **Given** a player in the match lobby, **When** opponent has not joined yet, **Then** opponent side shows a placeholder avatar with "Waiting for opponent..." text.

4. **Given** an unauthorized user attempts to access a match lobby URL, **When** they navigate to the lobby, **Then** they see an "Unauthorized - you are not a participant in this match" error page.

---

### User Story 2 - Player Indicates Ready Status (Priority: P1)

Both players in a match lobby need to indicate they are ready before the match can start. Players see a clear "Ready" button and can track both their own and opponent's ready status.

**Why this priority**: Ready flow is essential for fair match starts - both competitors must be prepared before gameplay begins.

**Independent Test**: Can be tested by having a player click Ready and verifying status updates for both players.

**Acceptance Scenarios**:

1. **Given** a player in the match lobby who has not indicated ready, **When** they click the "Ready" button, **Then** their ready status changes to "Ready" and is visible to opponent.

2. **Given** a player who has clicked Ready, **When** they are waiting for opponent, **Then** they see their own "Ready" indicator (green checkmark) and opponent's "Not Ready" indicator.

3. **Given** both players have clicked Ready, **When** the system detects both ready states, **Then** both players see a "Match Starting..." countdown (3, 2, 1) and the match transitions to "started" state.

4. **Given** a player refreshes their browser while in the lobby, **When** the page reloads, **Then** their ready status is preserved and they return to the same lobby state seamlessly.

---

### User Story 3 - Game Loads in Lobby Iframe (Priority: P1)

When a match starts, the game loads in an iframe within the lobby so players can compete without leaving the tournament context.

**Why this priority**: The game iframe is the actual gameplay surface - without it, matches cannot be played in the tournament flow.

**Independent Test**: Can be tested by starting a match and verifying the game iframe loads with correct session parameters.

**Acceptance Scenarios**:

1. **Given** a match in "started" state, **When** the lobby page renders, **Then** a centered game iframe appears with the game loaded using match-specific session parameters.

2. **Given** the game iframe is displayed, **When** the player interacts with it, **Then** game controls and gameplay function correctly within the iframe context.

3. **Given** a match completes in the external game, **When** results are retrieved from Game API, **Then** the iframe is replaced with a match result summary showing winner, scores, and "Return to Tournament" button.

4. **Given** the game fails to load in the iframe, **When** the timeout expires (30 seconds), **Then** an error message appears with "Retry" and "Contact Host" options.

---

### User Story 4 - Tournament Bracket View (Priority: P1)

A participant or spectator wants to see the tournament bracket to understand match progression, current standings, and upcoming matches.

**Why this priority**: Bracket visualization is fundamental for tournament UX - players need to see where they are and who they might face next.

**Independent Test**: Can be tested by loading a started tournament and verifying bracket displays all rounds with correct match assignments.

**Acceptance Scenarios**:

1. **Given** a started tournament with generated bracket, **When** user navigates to the "Bracket" tab on tournament detail page, **Then** they see a visual bracket tree showing all rounds (e.g., Quarter-finals → Semi-finals → Final).

2. **Given** a bracket view, **When** viewing a completed match, **Then** winner is highlighted and loser is shown as eliminated (crossed out or dimmed).

3. **Given** a bracket view with upcoming matches, **When** viewing a pending match slot, **Then** it shows "TBD" or the assigned players depending on bracket progression.

4. **Given** a participant viewing the bracket, **When** they click on their own match, **Then** they are navigated to their match lobby (if match is ready) or see match details modal.

5. **Given** a bracket with bye assignments, **When** viewing round 1, **Then** bye recipients are shown as auto-advanced to round 2 with "BYE" indicator.

---

### User Story 5 - Matches List View (Priority: P2)

A user wants to see all matches in the tournament in a list format, grouped by round, with status indicators and ability to navigate to match details.

**Why this priority**: List view provides an alternative to bracket for detailed match information. Important for tournament management but not blocking core flow.

**Independent Test**: Can be tested by loading tournament matches tab and verifying all matches display with correct status and player info.

**Acceptance Scenarios**:

1. **Given** a started tournament, **When** user navigates to the "Matches" tab, **Then** they see all matches grouped by round (Runda 1, Półfinały, Finał).

2. **Given** a match in the list, **When** viewing its card, **Then** they see: match number, both player names/avatars, current score (if any), and status badge (Oczekuje/Na żywo/Zakończony).

3. **Given** a match in progress, **When** viewing the matches list, **Then** the live match has a pulsing "Na żywo" (Live) badge and is prominently displayed.

4. **Given** the user is a participant in a match, **When** viewing the matches list, **Then** their match is highlighted and shows "Join Lobby" button.

---

### User Story 6 - Real-time Status Updates (Priority: P2)

Tournament participants and spectators see live updates as matches progress, players ready up, and results are recorded - without manually refreshing.

**Why this priority**: Real-time updates create an engaging tournament experience. Critical for competitive feel but can fall back to manual refresh.

**Independent Test**: Can be tested by having two players in a match while a third watches the bracket, verifying all see updates in real-time.

**Acceptance Scenarios**:

1. **Given** a user viewing the bracket or matches tab, **When** a match status changes (started, completed), **Then** the UI updates within 3 seconds without page refresh.

2. **Given** a player in a match lobby, **When** opponent clicks Ready, **Then** the player sees opponent's ready status update immediately (within 1 second).

3. **Given** a spectator watching a match, **When** the match completes, **Then** they see the result and winner advancing in the bracket.

4. **Given** connection to real-time service is lost, **When** the disconnect occurs, **Then** user sees a "Reconnecting..." indicator and data refreshes automatically when connection restores.

---

### User Story 7 - Player with Bye Waits for Next Round (Priority: P2)

A player who received a bye in round 1 needs to understand their status and wait for their next opponent to be determined.

**Why this priority**: Bye recipients need clear communication about their auto-advancement. Important for UX but affects fewer players.

**Independent Test**: Can be tested by starting a tournament with 5 participants and verifying bye recipients see correct waiting state.

**Acceptance Scenarios**:

1. **Given** a player who received a bye, **When** they view their tournament status, **Then** they see "You have a bye this round - auto-advancing to next round" message.

2. **Given** a bye recipient viewing the bracket, **When** their next round match slot is visible, **Then** they see themselves assigned with opponent showing "Winner of Match X".

3. **Given** a bye recipient waiting, **When** their potential opponent's match completes, **Then** they receive notification that their next match is ready and can join lobby.

---

### User Story 8 - Disconnect Handling in Lobby (Priority: P2)

A player's connection drops during the match. The system handles reconnection gracefully and applies disconnect penalties according to game rules.

**Why this priority**: Network issues are inevitable in online gaming. Graceful handling prevents frustration and ensures fair outcomes.

**Independent Test**: Can be tested by simulating disconnect during match and verifying reconnect window and penalty application.

**Acceptance Scenarios**:

1. **Given** a player in an active match, **When** they disconnect (close browser/lose connection), **Then** opponent sees "Opponent disconnected - 30s reconnect window" countdown.

2. **Given** a disconnected player, **When** they return within 30 seconds, **Then** they rejoin the match lobby seamlessly and gameplay continues.

3. **Given** the 30-second window expires without reconnection, **When** the timer reaches zero, **Then** present player sees "Opponent disqualified" message and receives walkover win.

4. **Given** a player who was disconnected, **When** they return to the match (within window), **Then** they see current match state including any penalty points awarded during disconnect.

---

### User Story 9 - Host Manually Enters Results (Priority: P3)

When automatic score retrieval fails, the host can manually enter match results to keep the tournament progressing.

**Why this priority**: Fallback flow for API failures. Rare but necessary for tournament completion.

**Independent Test**: Can be tested by simulating API failure and having host enter results through the UI.

**Acceptance Scenarios**:

1. **Given** a match flagged for manual entry, **When** host views the match in their dashboard, **Then** they see "API Error - Manual Entry Required" alert with input fields.

2. **Given** a host entering manual results, **When** they submit winner and score, **Then** match is marked complete and winner advances (with "Manually Entered" indicator).

3. **Given** a manually resolved match, **When** viewing match details, **Then** system shows "Result manually corrected by host" indicator with timestamp.

---

### User Story 10 - Walkover Request for No-Show (Priority: P3)

When opponent doesn't show up within the timeout window, the present player can request a walkover win.

**Why this priority**: No-show handling prevents tournaments from stalling. Important but relies on core flow being established.

**Independent Test**: Can be tested by having only one player join lobby, waiting for timeout, and requesting walkover.

**Acceptance Scenarios**:

1. **Given** a player in lobby waiting for opponent, **When** 2 minutes pass without opponent joining, **Then** player sees "Opponent not present - Claim Walkover" button.

2. **Given** a player clicks "Claim Walkover", **When** the system processes the request, **Then** the player wins by walkover and advances to next round.

3. **Given** both players absent for 5 minutes, **When** timeout expires, **Then** host receives notification to manually resolve the match.

---

### Edge Cases

- **Player refreshes browser while in pre-lobby**: Seamlessly returns to pre-lobby, state preserved server-side (consistent with match lobby FR-008)
- **Player joins pre-lobby during 30s grace period**: Included in bracket generation, receives match assignment like other participants
- **Player leaves pre-lobby during grace period (disconnect/close)**: Removed from roster, participant count decreases, not included in bracket
- **Player joins pre-lobby after grace period expired**: Toast notification "Tournament in progress - registration closed" shown, automatically redirected to tournament detail page to view bracket as spectator
- **Participant count drops below minimum during grace period**: Tournament start cancelled, players shown "Insufficient participants" message and returned to tournament detail page
- **Tournament cancelled while player in pre-lobby**: Show "Tournament Cancelled" notification and redirect to tournament detail page
- **Player navigates away during ready countdown**: Match start is cancelled, ready states reset, both players must re-ready
- **Multiple browser tabs open to same lobby**: Only most recent tab is active; others show "Session active in another tab" message
- **Tournament cancelled while player is in lobby**: Player sees "Tournament Cancelled" modal with explanation and link to tournament list
- **Game iframe crashes**: Show error state with "Reload Game" button; if persistent, fallback to manual host entry
- **Bracket display for large tournaments (64+ participants)**: Horizontal scroll with zoom controls; show condensed view option
- **Slow network conditions**: Show loading skeleton states; degrade gracefully with stale data indicators
- **Player attempts to join wrong match lobby**: Show "You are not a participant in this match" with link to their actual match
- **Mobile device accessing lobby**: Responsive layout with stacked player cards; game iframe adapts to viewport

## Requirements *(mandatory)*

### Functional Requirements

**Tournament Pre-Lobby (Waiting Room)**
- **FR-000**: System MUST render tournament pre-lobby page at route `/tournaments/:tournamentId/lobby`
- **FR-000a**: System MUST authenticate pre-lobby access by verifying user is registered participant in tournament
- **FR-000b**: System MUST deny pre-lobby access to non-participants with clear error message
- **FR-000c**: System MUST display countdown timer to tournament start time
- **FR-000d**: System MUST show list of joined participants with avatars and display names
- **FR-000e**: System MUST display activity feed showing participant joins/leaves and tournament status updates (no player messaging)
- **FR-000f**: System MUST listen for tournament start event via WebSocket
- **FR-000g**: System MUST listen for `match_assigned` WebSocket event containing match ID
- **FR-000h**: System MUST display "Match Assigned" notification with opponent name for 2 seconds
- **FR-000i**: System MUST automatically redirect player to their match lobby after 2-second notification delay
- **FR-000j**: System MUST show "You have a bye" notification for bye recipients instead of match redirect
- **FR-000k**: System MUST allow seamless return to pre-lobby on browser refresh or reconnection (server-side state preservation)
- **FR-000l**: System MUST support 30-second grace period after tournament start - players joining pre-lobby during this window are included in bracket generation
- **FR-000m**: System MUST track roster changes during grace period - both arrivals and departures affect final bracket participant list
- **FR-000n**: System MUST display "Finalizing bracket..." indicator with real-time participant count updates during grace period (e.g., "Finalizing bracket... 8 players")
- **FR-000o**: System MUST finalize bracket with only participants present in pre-lobby at end of grace period (disconnected players excluded)
- **FR-000p**: System MUST show toast notification "Tournament in progress - registration closed" to late arrivals after grace period
- **FR-000q**: System MUST automatically redirect late arrivals to tournament detail page (where they can view bracket as spectator)

**Match Lobby Entry & Display**
- **FR-001**: System MUST render match lobby page at route `/lobby/:tournamentId/match/:matchId`
- **FR-002**: System MUST display player avatars and display names in top-center "VS" layout (current user on left, opponent on right)
- **FR-003**: System MUST show placeholder state when opponent has not yet joined lobby
- **FR-004**: System MUST authenticate lobby access by matching current user ID to match participant ID
- **FR-005**: System MUST deny access with clear error message for non-participants

**Ready State Management**
- **FR-006**: System MUST display prominent "Ready" button for players to indicate readiness
- **FR-007**: System MUST show real-time ready status indicators for both players (checkmark/clock icons)
- **FR-008**: System MUST persist ready status server-side (survives browser refresh)
- **FR-009**: System MUST show "Match Starting..." countdown (3, 2, 1) when both players are ready
- **FR-010**: System MUST transition to game view when countdown completes

**Game Iframe Integration**
- **FR-011**: System MUST render centered game iframe when match is in "started" state
- **FR-012**: System MUST pass match session parameters to game iframe via URL or postMessage
- **FR-013**: System MUST handle game loading errors with retry functionality
- **FR-014**: System MUST show loading state while game initializes
- **FR-015**: System MUST replace iframe with result summary when match completes

**Bracket Visualization**
- **FR-016**: System MUST display interactive bracket tree for single elimination tournaments
- **FR-017**: System MUST show all rounds with correct labels (Runda 1, Ćwierćfinały, Półfinały, Finał)
- **FR-018**: System MUST highlight winners and indicate eliminated players
- **FR-019**: System MUST show bye assignments with "BYE" indicator
- **FR-020**: System MUST allow clicking on matches to view details or join lobby
- **FR-021**: System MUST support horizontal scrolling for large brackets

**Matches List View**
- **FR-022**: System MUST display matches grouped by round
- **FR-023**: System MUST show match cards with: match number, player names, avatars, score, status badge
- **FR-024**: System MUST highlight live matches with pulsing indicator
- **FR-025**: System MUST show "Join Lobby" button for user's active matches

**Real-time Updates**
- **FR-026**: System MUST subscribe to WebSocket/SSE for match status changes
- **FR-027**: System MUST update UI within 3 seconds of server-side state change
- **FR-028**: System MUST show reconnection indicator when connection drops
- **FR-029**: System MUST auto-reconnect and refresh data on connection restore

**Disconnect Handling**
- **FR-030**: System MUST display disconnect countdown (30 seconds) when opponent disconnects
- **FR-031**: System MUST show "Opponent Disqualified" when timer expires
- **FR-032**: System MUST allow seamless rejoin for players returning within window

**Bye Handling**
- **FR-033**: System MUST show "Auto-advanced" message for bye recipients
- **FR-034**: System MUST display bye recipient's next match slot with "Winner of Match X" opponent
- **FR-035**: System MUST notify bye recipient when their next match becomes ready

**Walkover & No-Show**
- **FR-036**: System MUST show "Claim Walkover" button after 2-minute opponent no-show
- **FR-037**: System MUST process walkover and advance present player
- **FR-038**: System MUST notify host when both players are absent after 5 minutes

**Host Controls (Frontend)**
- **FR-039**: System MUST show "Manual Entry Required" alert for flagged matches
- **FR-040**: System MUST provide result input form for hosts
- **FR-041**: System MUST display "Manually Entered" indicator on host-resolved matches

### Key Entities

- **TournamentPreLobby**: Waiting room state before tournament starts, including participant list, countdown timer, and activity feed (joins/leaves/updates - no player messaging)

- **MatchLobby**: Virtual waiting room state including participant IDs, ready states, connection status, and match reference

- **Bracket**: Visual representation of tournament progression with rounds, matches, and their relationships

- **MatchCard**: UI component showing match details including players, scores, status, and available actions

- **PlayerDisplay**: Reusable component for showing player avatar, name, ready status, and connection state

- **GameFrame**: Iframe container with game session management, error handling, and communication layer

## Success Criteria *(mandatory)*

### Measurable Outcomes

**Measurement Context**: All timing criteria measured under standard conditions: 50ms network latency, warm browser cache, desktop browser (Chrome/Firefox/Safari latest).

- **SC-001**: Players can navigate to match lobby and see opponent information within 2 seconds of page load
- **SC-002**: Ready status updates are visible to opponent within 1 second of clicking
- **SC-003**: 95% of players successfully complete the ready → match start flow without errors
- **SC-004**: Bracket visualization renders correctly for tournaments up to 128 participants
- **SC-005**: Real-time updates propagate to all viewers within 3 seconds of state change
- **SC-006**: 90% of players can return from disconnect within the 30-second window without issues
- **SC-007**: Game iframe loads and becomes interactive within 5 seconds of match start
- **SC-008**: Mobile users can access all lobby functionality without horizontal scrolling on player info sections
- **SC-009**: Walkover claim flow completes within 10 seconds of button click
- **SC-010**: Host can manually enter results and see bracket update within 5 seconds

## Assumptions

- Backend matchmaking service (001-tournament-matchmaking) is implemented and provides REST API + WebSocket endpoints
- User authentication system exists and provides current user context with profile data (avatar, display name)
- Game integration provides embeddable iframe URL with session token parameter
- Tournament detail page (`TournamentDetailPage.tsx`) exists and can be extended with new tab functionality
- Design system components (Button, Badge, Avatar, etc.) are available in shared components library
- WebSocket infrastructure is available for real-time communication

## Out of Scope

- Tournament creation and registration flows (handled by existing tournament management)
- Game client implementation (external system, only iframe integration)
- Host dashboard for tournament management (separate feature)
- Spectator chat during matches (future feature)
- Video/audio communication between players (future feature)
- Double elimination bracket visualization (single elimination first)
- Tournament standings/leaderboards beyond bracket display
