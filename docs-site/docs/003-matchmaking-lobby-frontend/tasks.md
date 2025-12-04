# Tasks: Matchmaking Lobby Frontend

**Feature**: 003-matchmaking-lobby-frontend  
**Branch**: `003-matchmaking-lobby-frontend`  
**Date**: 2025-12-04

## Overview

This document breaks down the implementation of the matchmaking lobby frontend into actionable, dependency-ordered tasks organized by user story. Each user story phase is independently testable.

**Total Tasks**: 133  
**Parallel Opportunities**: 45 parallelizable tasks marked with [P]

## Task Format

```
- [ ] [TaskID] [P?] [Story?] Description with file path
```

- **TaskID**: Sequential number (T001, T002, ...)
- **[P]**: Parallelizable (different files, no dependencies)
- **[Story]**: User story label (US0, US1, ...) for story-specific tasks
- **Description**: Clear action with exact file path

## Implementation Strategy

**MVP Scope**: US0 + US1 + US2 + US3 + US4 (Pre-lobby, Match lobby entry, Ready flow, Game iframe, Bracket view)

**Incremental Delivery**:
1. Phase 1-2: Foundation (Setup + Shared infrastructure)
2. Phase 3-7: MVP (US0-US4) - Complete pre-lobby to game flow
3. Phase 8-12: Enhanced Features (US5-US9) - Matches list, real-time, disconnects
4. Phase 13: Polish & Cross-cutting

---

## Phase 1: Setup

**Goal**: Initialize project structure and install dependencies

- [x] T001 Verify TypeScript 5.9+ and React 19 are installed in frontends/winspire-app/package.json
- [x] T002 Install required dependencies: @tanstack/react-query, zod in frontends/winspire-app/ (NOTE: @tanstack/react-query pending manual install)
- [x] T003 Create feature directory structure in frontends/winspire-app/src/features/lobby/
- [x] T004 Create api subdirectory in frontends/winspire-app/src/features/lobby/api/
- [x] T005 Create components subdirectory in frontends/winspire-app/src/features/lobby/components/
- [x] T006 Create hooks subdirectory in frontends/winspire-app/src/features/lobby/hooks/
- [x] T007 Create pages subdirectory in frontends/winspire-app/src/features/lobby/pages/
- [x] T008 Create shared hooks directory in frontends/winspire-app/src/shared/hooks/

---

## Phase 2: Foundational (Blocking Prerequisites)

**Goal**: Build shared infrastructure that all user stories depend on

**Dependencies**: Phase 1 complete

### Core Types

- [x] T009 [P] Define Match, Bracket, Round types in frontends/winspire-app/src/features/lobby/types.ts
- [x] T010 [P] Define TournamentPreLobbyState and PreLobbyParticipant types in frontends/winspire-app/src/features/lobby/types.ts
- [x] T011 [P] Define PlayerInfo and LobbyState types in frontends/winspire-app/src/features/lobby/types.ts
- [x] T012 [P] Define WebSocket message types (ClientMessageType, ServerMessageType) in frontends/winspire-app/src/features/lobby/types.ts
- [x] T013 [P] Define WebSocket payload types (LobbyStatePayload, MatchAssignedPayload, etc.) in frontends/winspire-app/src/features/lobby/types.ts

### Validation Schemas

- [x] T014 [P] Create Zod schemas for Match validation in frontends/winspire-app/src/features/lobby/schemas.ts
- [x] T015 [P] Create Zod schemas for PreLobby validation in frontends/winspire-app/src/features/lobby/schemas.ts
- [x] T016 [P] Create Zod schemas for WebSocket messages in frontends/winspire-app/src/features/lobby/schemas.ts

### API Client

- [x] T017 Create base matchmaking API client with fetch wrapper in frontends/winspire-app/src/features/lobby/api/matchmakingApi.ts
- [x] T018 Add getPreLobbyState(tournamentId) endpoint in frontends/winspire-app/src/features/lobby/api/matchmakingApi.ts
- [x] T019 Add getMatch(matchId) endpoint in frontends/winspire-app/src/features/lobby/api/matchmakingApi.ts
- [x] T020 Add markReady(matchId, playerId) endpoint in frontends/winspire-app/src/features/lobby/api/matchmakingApi.ts

### WebSocket Infrastructure

- [x] T021 Implement useWebSocket hook with connection lifecycle in frontends/winspire-app/src/shared/hooks/useWebSocket.ts
- [x] T022 Add auto-reconnect with exponential backoff (1s, 2s, 4s, max 30s) in frontends/winspire-app/src/shared/hooks/useWebSocket.ts
- [x] T023 Add connection state tracking (connecting, connected, disconnected, reconnecting) in frontends/winspire-app/src/shared/hooks/useWebSocket.ts
- [x] T024 Add message queue for pending messages during reconnect in frontends/winspire-app/src/shared/hooks/useWebSocket.ts

### Constants

- [x] T025 [P] Define lobby constants (RECONNECT_TIMEOUT, GRACE_PERIOD_DURATION, etc.) in frontends/winspire-app/src/features/lobby/constants.ts

---

## Phase 3: User Story 0 - Tournament Pre-Lobby (P1)

**Goal**: Implement waiting room where players gather before tournament starts

**Dependencies**: Phase 2 complete

**Independent Test**: Navigate to `/tournaments/:tournamentId/lobby`, verify countdown, participant list, and activity feed display

### Components

- [x] T026 [P] [US0] Create TournamentPreLobbyPage with route `/tournaments/:tournamentId/lobby` in frontends/winspire-app/src/features/lobby/pages/TournamentPreLobbyPage.tsx
- [x] T027 [P] [US0] Create ParticipantList component showing avatars and display names in frontends/winspire-app/src/features/lobby/components/ParticipantList.tsx
- [x] T028 [P] [US0] Create ActivityFeed component for joins/leaves events in frontends/winspire-app/src/features/lobby/components/ActivityFeed.tsx
- [x] T029 [P] [US0] Create GracePeriodIndicator component with 30s countdown in frontends/winspire-app/src/features/lobby/components/GracePeriodIndicator.tsx

### State Management

- [x] T030 [US0] Create useTournamentPreLobby hook with WebSocket connection in frontends/winspire-app/src/features/lobby/hooks/useTournamentPreLobby.ts
- [x] T031 [US0] Add prelobby_state message handler in frontends/winspire-app/src/features/lobby/hooks/useTournamentPreLobby.ts
- [x] T032 [US0] Add participant_joined message handler in frontends/winspire-app/src/features/lobby/hooks/useTournamentPreLobby.ts
- [x] T033 [US0] Add participant_left message handler in frontends/winspire-app/src/features/lobby/hooks/useTournamentPreLobby.ts
- [x] T034 [US0] Add grace_period_started message handler in frontends/winspire-app/src/features/lobby/hooks/useTournamentPreLobby.ts
- [x] T035 [US0] Add roster_updated message handler (participant count changes during grace period) in frontends/winspire-app/src/features/lobby/hooks/useTournamentPreLobby.ts
- [x] T036 [US0] Add match_assigned message handler with 2s delay + redirect logic in frontends/winspire-app/src/features/lobby/hooks/useTournamentPreLobby.ts

### Integration

- [x] T037 [US0] Integrate TournamentPreLobbyPage with useTournamentPreLobby hook in frontends/winspire-app/src/features/lobby/pages/TournamentPreLobbyPage.tsx
- [x] T038 [US0] Add authentication check (verify user is registered participant) in frontends/winspire-app/src/features/lobby/pages/TournamentPreLobbyPage.tsx
- [x] T039 [US0] Add late arrival toast notification (after grace period) in frontends/winspire-app/src/features/lobby/pages/TournamentPreLobbyPage.tsx
- [x] T040 [US0] Add route registration for /tournaments/:tournamentId/lobby in frontends/winspire-app/src/App.tsx

---

## Phase 4: User Story 1 - Player Enters Match Lobby (P1)

**Goal**: Display match lobby with player VS opponent layout

**Dependencies**: Phase 2 complete (can run parallel with US0)

**Independent Test**: Navigate to `/lobby/:tournamentId/match/:matchId`, verify VS display with avatars

### Components

- [x] T041 [P] [US1] Create MatchLobbyPage with route `/lobby/:tournamentId/match/:matchId` in frontends/winspire-app/src/features/lobby/pages/MatchLobbyPage.tsx
- [x] T042 [P] [US1] Create PlayerVsDisplay component with left-right layout in frontends/winspire-app/src/features/lobby/components/PlayerVsDisplay.tsx
- [x] T043 [P] [US1] Add VS separator with styling in frontends/winspire-app/src/features/lobby/components/PlayerVsDisplay.tsx
- [x] T044 [P] [US1] Add placeholder state for missing opponent ("Waiting for opponent...") in frontends/winspire-app/src/features/lobby/components/PlayerVsDisplay.tsx

### State Management

- [x] T045 [US1] Create useMatchLobby hook with WebSocket connection to /v1/matches/:id/lobby in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts
- [x] T046 [US1] Add lobby_state message handler in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts
- [x] T047 [US1] Add player_joined message handler in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts
- [x] T048 [US1] Add player_left message handler in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts

### Integration

- [x] T049 [US1] Integrate MatchLobbyPage with useMatchLobby hook in frontends/winspire-app/src/features/lobby/pages/MatchLobbyPage.tsx
- [x] T050 [US1] Add authentication check (verify user is match participant) in frontends/winspire-app/src/features/lobby/pages/MatchLobbyPage.tsx
- [x] T051 [US1] Add unauthorized error page for non-participants in frontends/winspire-app/src/features/lobby/pages/MatchLobbyPage.tsx
- [x] T052 [US1] Add route registration for /lobby/:tournamentId/match/:matchId in frontends/winspire-app/src/App.tsx

---

## Phase 5: User Story 2 - Player Ready Status (P1)

**Goal**: Implement ready button and countdown when both players ready

**Dependencies**: Phase 4 (US1) complete

**Independent Test**: Click Ready button, verify status updates and countdown triggers

### Components

- [ ] T053 [P] [US2] Create ReadyButton component with click handler in frontends/winspire-app/src/features/lobby/components/ReadyButton.tsx
- [ ] T054 [P] [US2] Add ready status indicators (checkmark/clock icons) in frontends/winspire-app/src/features/lobby/components/PlayerVsDisplay.tsx
- [ ] T055 [P] [US2] Create MatchStartCountdown component (3, 2, 1) in frontends/winspire-app/src/features/lobby/components/MatchStartCountdown.tsx

### State Management

- [ ] T056 [US2] Create useReadyState hook with optimistic updates in frontends/winspire-app/src/features/lobby/hooks/useReadyState.ts
- [ ] T057 [US2] Add ready_updated WebSocket message handler in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts
- [ ] T058 [US2] Add match_starting WebSocket message handler (triggers countdown) in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts
- [ ] T059 [US2] Add server-side ready state persistence on page load in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts

### Integration

- [ ] T060 [US2] Integrate ReadyButton with useReadyState in frontends/winspire-app/src/features/lobby/pages/MatchLobbyPage.tsx
- [ ] T061 [US2] Add countdown display when match_starting received in frontends/winspire-app/src/features/lobby/pages/MatchLobbyPage.tsx

---

## Phase 6: User Story 3 - Game Iframe (P1)

**Goal**: Load game in iframe when match starts

**Dependencies**: Phase 5 (US2) complete

**Independent Test**: Start match, verify game iframe loads with correct session parameters

### Components

- [ ] T062 [P] [US3] Create GameFrame component with iframe container in frontends/winspire-app/src/features/lobby/components/GameFrame.tsx
- [ ] T063 [P] [US3] Add game loading state with spinner in frontends/winspire-app/src/features/lobby/components/GameFrame.tsx
- [ ] T064 [P] [US3] Add iframe error handling with retry button in frontends/winspire-app/src/features/lobby/components/GameFrame.tsx
- [ ] T065 [P] [US3] Create MatchResult component for post-match summary in frontends/winspire-app/src/features/lobby/components/MatchResult.tsx

### Communication

- [ ] T066 [US3] Implement postMessage listener for game completion signals in frontends/winspire-app/src/features/lobby/components/GameFrame.tsx
- [ ] T067 [US3] Add session token passing via URL parameters in frontends/winspire-app/src/features/lobby/components/GameFrame.tsx
- [ ] T068 [US3] Add 30-second timeout for game load failure in frontends/winspire-app/src/features/lobby/components/GameFrame.tsx

### State Management

- [ ] T069 [US3] Add match_started WebSocket message handler (provides gameUrl) in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts
- [ ] T070 [US3] Add match_completed WebSocket message handler in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts

### Integration

- [ ] T071 [US3] Integrate GameFrame into MatchLobbyPage (show when match started) in frontends/winspire-app/src/features/lobby/pages/MatchLobbyPage.tsx
- [ ] T072 [US3] Replace iframe with MatchResult when match completes in frontends/winspire-app/src/features/lobby/pages/MatchLobbyPage.tsx

---

## Phase 7: User Story 4 - Bracket View (P1)

**Goal**: Display interactive tournament bracket

**Dependencies**: Phase 2 complete (can run parallel with US0-US3)

**Independent Test**: Load tournament detail page, click Bracket tab, verify all rounds display

### API Integration

- [ ] T073 [P] [US4] Add getBracket(tournamentId) endpoint in frontends/winspire-app/src/features/host/api/tournamentApi.ts
- [ ] T074 [P] [US4] Update Bracket and Round types in frontends/winspire-app/src/features/host/types.ts

### Components

- [ ] T075 [P] [US4] Update BracketView component with real API integration in frontends/winspire-app/src/features/host/components/BracketView.tsx
- [ ] T076 [P] [US4] Create BracketMatch component with SVG connectors in frontends/winspire-app/src/features/host/components/BracketMatch.tsx
- [ ] T077 [P] [US4] Add winner highlighting and eliminated player dimming in frontends/winspire-app/src/features/host/components/BracketMatch.tsx
- [ ] T078 [P] [US4] Add bye indicators ("BYE") in frontends/winspire-app/src/features/host/components/BracketMatch.tsx
- [ ] T079 [P] [US4] Add click handler for navigating to match lobby in frontends/winspire-app/src/features/host/components/BracketMatch.tsx

### Layout

- [ ] T080 [US4] Implement horizontal scroll for large brackets in frontends/winspire-app/src/features/host/components/BracketView.tsx
- [ ] T081 [US4] Add CSS Grid layout for round columns in frontends/winspire-app/src/features/host/components/BracketView.tsx
- [ ] T082 [US4] Add responsive mobile layout (horizontal scroll + zoom) in frontends/winspire-app/src/features/host/components/BracketView.tsx

### Integration

- [ ] T083 [US4] Update TournamentDetailPage Bracket tab to use new BracketView in frontends/winspire-app/src/features/host/pages/TournamentDetailPage.tsx

---

## Phase 8: User Story 5 - Matches List View (P2)

**Goal**: Display matches in list format grouped by round

**Dependencies**: Phase 7 (US4) complete

**Independent Test**: Load Matches tab, verify all matches display with correct status badges

### API Integration

- [ ] T084 [P] [US5] Add getMatches(tournamentId) endpoint in frontends/winspire-app/src/features/host/api/tournamentApi.ts

### Components

- [ ] T085 [P] [US5] Update MatchesView component with real API integration in frontends/winspire-app/src/features/host/components/MatchesView.tsx
- [ ] T086 [P] [US5] Create MatchCard component showing player names, avatars, score, status in frontends/winspire-app/src/features/host/components/MatchCard.tsx
- [ ] T087 [P] [US5] Add pulsing "Live" badge for active matches in frontends/winspire-app/src/features/host/components/MatchCard.tsx
- [ ] T088 [P] [US5] Add "Join Lobby" button for user's active matches in frontends/winspire-app/src/features/host/components/MatchCard.tsx

### Layout

- [ ] T089 [US5] Add round grouping (Runda 1, Półfinały, Finał) in frontends/winspire-app/src/features/host/components/MatchesView.tsx

### Integration

- [ ] T090 [US5] Update TournamentDetailPage Matches tab to use new MatchesView in frontends/winspire-app/src/features/host/pages/TournamentDetailPage.tsx

---

## Phase 9: User Story 6 - Real-time Updates (P2)

**Goal**: WebSocket updates propagate to all views without refresh

**Dependencies**: Phase 4 (US1) and Phase 7 (US4) complete

**Independent Test**: Have player ready up in lobby, verify bracket/matches list update in real-time

### WebSocket Integration

- [ ] T091 [US6] Add WebSocket subscription for bracket updates in frontends/winspire-app/src/features/host/components/BracketView.tsx
- [ ] T092 [US6] Add WebSocket subscription for matches list updates in frontends/winspire-app/src/features/host/components/MatchesView.tsx
- [ ] T093 [US6] Add React Query cache invalidation on WebSocket messages in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts

### Connection Handling

- [ ] T094 [P] [US6] Create ConnectionIndicator component for reconnection state in frontends/winspire-app/src/shared/components/ConnectionIndicator.tsx
- [ ] T095 [US6] Add "Reconnecting..." indicator when WebSocket drops in frontends/winspire-app/src/features/lobby/pages/MatchLobbyPage.tsx
- [ ] T096 [US6] Add auto-refresh data on connection restore in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts

---

## Phase 10: User Story 7 - Bye Handling (P2)

**Goal**: Show waiting state for players with byes

**Dependencies**: Phase 7 (US4) complete

**Independent Test**: Start tournament with odd participant count, verify bye recipient sees waiting message

### Components

- [ ] T097 [P] [US7] Create ByeWaitingState component with "Auto-advanced" message in frontends/winspire-app/src/features/lobby/components/ByeWaitingState.tsx
- [ ] T098 [P] [US7] Add next match slot display ("Winner of Match X") in frontends/winspire-app/src/features/lobby/components/ByeWaitingState.tsx

### State Management

- [ ] T099 [US7] Add bye detection logic in useTournamentPreLobby (handle match_assigned with bye flag) in frontends/winspire-app/src/features/lobby/hooks/useTournamentPreLobby.ts
- [ ] T100 [US7] Add notification when next match becomes ready in frontends/winspire-app/src/features/lobby/hooks/useTournamentPreLobby.ts

### Integration

- [ ] T101 [US7] Show ByeWaitingState instead of match redirect for bye recipients in frontends/winspire-app/src/features/lobby/pages/TournamentPreLobbyPage.tsx

---

## Phase 11: User Story 8 - Disconnect Handling (P2)

**Goal**: Handle disconnect with 30s reconnect window

**Dependencies**: Phase 4 (US1) complete

**Independent Test**: Disconnect during match, verify countdown and reconnect flow

### Components

- [ ] T102 [P] [US8] Create DisconnectOverlay component with 30s countdown in frontends/winspire-app/src/features/lobby/components/DisconnectOverlay.tsx
- [ ] T103 [P] [US8] Add "Opponent Disqualified" message when timer expires in frontends/winspire-app/src/features/lobby/components/DisconnectOverlay.tsx

### State Management

- [ ] T104 [US8] Create useDisconnect hook tracking disconnect state in frontends/winspire-app/src/features/lobby/hooks/useDisconnect.ts
- [ ] T105 [US8] Add player_disconnected WebSocket message handler in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts
- [ ] T106 [US8] Add player_reconnected WebSocket message handler in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts
- [ ] T107 [US8] Calculate remaining reconnect time from disconnectedAt timestamp in frontends/winspire-app/src/features/lobby/hooks/useDisconnect.ts

### Integration

- [ ] T108 [US8] Show DisconnectOverlay when opponent disconnects in frontends/winspire-app/src/features/lobby/pages/MatchLobbyPage.tsx
- [ ] T109 [US8] Handle seamless rejoin for returning players in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts

---

## Phase 12: User Story 9 - Host Manual Results (P3)

**Goal**: Allow host to manually enter match results

**Dependencies**: Phase 8 (US5) complete

**Independent Test**: Simulate API failure, verify host can enter results

### API Integration

- [ ] T110 [P] [US9] Add submitManualResult(matchId, winnerId, scores) endpoint in frontends/winspire-app/src/features/host/api/tournamentApi.ts

### Components

- [ ] T111 [P] [US9] Create ManualResultForm component with winner selection and score inputs in frontends/winspire-app/src/features/host/components/ManualResultForm.tsx
- [ ] T112 [P] [US9] Add "API Error - Manual Entry Required" alert in frontends/winspire-app/src/features/host/components/MatchCard.tsx
- [ ] T113 [P] [US9] Add "Manually Entered" indicator badge in frontends/winspire-app/src/features/host/components/MatchCard.tsx

### Integration

- [ ] T114 [US9] Add host authentication check in frontends/winspire-app/src/features/host/pages/TournamentDetailPage.tsx
- [ ] T115 [US9] Show ManualResultForm for flagged matches in frontends/winspire-app/src/features/host/components/MatchesView.tsx

---

## Phase 13: User Story 10 - Walkover Claims (P3)

**Goal**: Allow player to claim walkover after timeout

**Dependencies**: Phase 4 (US1) complete

**Independent Test**: Wait 2 minutes without opponent, verify "Claim Walkover" button appears

### API Integration

- [ ] T116 [P] [US10] Add claimWalkover(matchId) endpoint in frontends/winspire-app/src/features/lobby/api/matchmakingApi.ts

### Components

- [ ] T117 [P] [US10] Create WalkoverButton component with 2-minute timer in frontends/winspire-app/src/features/lobby/components/WalkoverButton.tsx

### State Management

- [ ] T118 [US10] Add walkover timer logic (show button after 2 min) in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts
- [ ] T119 [US10] Handle walkover success and advancement in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts

### Integration

- [ ] T120 [US10] Show WalkoverButton when timeout expires in frontends/winspire-app/src/features/lobby/pages/MatchLobbyPage.tsx
- [ ] T120a [US10] Implement host notification for double no-show (5min timeout per FR-038) in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts

---

## Phase 14: Polish & Cross-Cutting

**Goal**: Final touches, mobile responsiveness, error handling

**Dependencies**: All user story phases complete

### Mobile Responsiveness

- [ ] T121 [P] Add mobile breakpoints and responsive styles to PlayerVsDisplay in frontends/winspire-app/src/features/lobby/components/PlayerVsDisplay.tsx
- [ ] T122 [P] Add mobile layout (stacked) for MatchLobbyPage in frontends/winspire-app/src/features/lobby/pages/MatchLobbyPage.tsx
- [ ] T123 [P] Add mobile-optimized GameFrame (viewport adaptation) in frontends/winspire-app/src/features/lobby/components/GameFrame.tsx
- [ ] T124 [P] Test bracket horizontal scroll on mobile devices in frontends/winspire-app/src/features/host/components/BracketView.tsx

### Error Handling

- [ ] T125 [P] Add error boundaries for lobby pages in frontends/winspire-app/src/features/lobby/pages/
- [ ] T126 [P] Add fallback UI for WebSocket connection failures in frontends/winspire-app/src/features/lobby/pages/MatchLobbyPage.tsx
- [ ] T127 [P] Add toast notifications for critical errors in frontends/winspire-app/src/features/lobby/pages/
- [ ] T128 [P] Add loading skeletons for slow network conditions in frontends/winspire-app/src/features/lobby/components/

### Edge Cases

- [ ] T129 [P] Handle multiple browser tabs (show "Session active in another tab") in frontends/winspire-app/src/features/lobby/hooks/useMatchLobby.ts
- [ ] T130 [P] Handle tournament cancellation during lobby in frontends/winspire-app/src/features/lobby/pages/MatchLobbyPage.tsx
- [ ] T131 [P] Handle player navigating to wrong match lobby in frontends/winspire-app/src/features/lobby/pages/MatchLobbyPage.tsx
- [ ] T132 [P] Handle insufficient participants during grace period in frontends/winspire-app/src/features/lobby/pages/TournamentPreLobbyPage.tsx

---

## Dependencies Graph

```
Phase 1 (Setup)
    ↓
Phase 2 (Foundational) ← BLOCKS ALL USER STORIES
    ↓
    ├─→ Phase 3 (US0 - Pre-Lobby) [P1]
    ├─→ Phase 4 (US1 - Match Lobby Entry) [P1] ← parallel with US0
    │       ↓
    │   Phase 5 (US2 - Ready Status) [P1]
    │       ↓
    │   Phase 6 (US3 - Game Iframe) [P1]
    │
    └─→ Phase 7 (US4 - Bracket View) [P1] ← parallel with US0-US3
            ↓
        Phase 8 (US5 - Matches List) [P2]
            ↓
        Phase 12 (US9 - Host Manual Results) [P3]

Phase 4 (US1) + Phase 7 (US4)
    ↓
Phase 9 (US6 - Real-time Updates) [P2]

Phase 7 (US4)
    ↓
Phase 10 (US7 - Bye Handling) [P2]

Phase 4 (US1)
    ↓
Phase 11 (US8 - Disconnect Handling) [P2]
    ↓
Phase 13 (US10 - Walkover Claims) [P3]

All Phases
    ↓
Phase 14 (Polish)
```

## Parallel Execution Examples

### Phase 3 (US0 - Pre-Lobby)
**Parallel Group 1** (after T025 complete):
- T026, T027, T028, T029 (all components, different files)

### Phase 4 (US1 - Match Lobby)
**Parallel Group 2** (after T040 complete):
- T041, T042, T043, T044 (all components, different files)

### Phase 7 (US4 - Bracket)
**Parallel Group 3** (after T074 complete):
- T075, T076, T077, T078, T079 (all components, different files)

### Phase 14 (Polish)
**Parallel Group 4** (all tasks can run in parallel):
- T121, T122, T123, T124, T125, T126, T127, T128, T129, T130, T131, T132

## Validation Checklist

### Format Validation ✅
- [x] All tasks have checkbox format `- [ ]`
- [x] All tasks have sequential TaskID (T001-T132, plus T120a)
- [x] All parallelizable tasks marked with [P]
- [x] All story tasks have [Story] label (US0-US10)
- [x] All tasks include file paths

### Completeness ✅
- [x] Each user story has all needed tasks
- [x] Dependencies are clearly specified
- [x] Tasks map to functional requirements in spec.md
- [x] Independent test criteria provided per story
- [x] Parallel opportunities identified (45 tasks)

### Organization ✅
- [x] Tasks organized by user story
- [x] Setup and foundational phases separate
- [x] Polish phase at end
- [x] Clear dependency graph

---

## Suggested MVP Scope

**Phase 1-7 (Tasks T001-T083)**: Complete pre-lobby → match lobby → ready → game → bracket flow

This covers:
- ✅ US0: Tournament Pre-Lobby with grace period
- ✅ US1: Match Lobby Entry
- ✅ US2: Ready Status
- ✅ US3: Game Iframe
- ✅ US4: Bracket View

**Estimated Tasks**: 83  
**Estimated Complexity**: Medium-High (many integration points)

---

**Next Steps**: 
1. Execute Phase 1-2 (Setup + Foundation) first
2. Implement MVP scope (Phase 3-7)
3. Test complete flow end-to-end
4. Add enhanced features (Phase 8-13)
5. Polish and handle edge cases (Phase 14)
