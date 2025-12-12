# Round Management & Tournament Flow Plan

**Status:** DRAFT (Updated)
**Priority:** CRITICAL
**Author:** Claude Code
**Created:** 2025-01-XX
**Last Updated:** 2025-01-XX - Grace period trigger changed to "all winners joined pre-lobby"

## Overview

This document outlines the implementation plan for a comprehensive Round/Tournament Manager that orchestrates the complete tournament flow from first round to final, ensuring users properly return to pre-lobby between rounds.

## Current State Analysis

### What Exists

**Backend (`services/matchmaking`):**
- `MatchService` - handles match completion, winner advancement (`CompleteMatch`, `AdvanceWinner`)
- `BracketService` - generates brackets from participants
- `PreLobbyService` - manages pre-lobby before first round
- Events: `MatchCompleted`, `ParticipantAdvanced`, `ParticipantEliminated`, `TournamentCompleted`
- WebSocket Hub - broadcasts to match/tournament rooms

**Frontend (`frontends/winspire-app`):**
- `useMatchLobby` hook - handles match state, ready states, game loading
- `useTournamentPreLobby` hook - handles pre-lobby before bracket generation
- `MatchResult` component - displays win/loss result (but NO navigation logic after)
- WebSocket message types defined in `types.ts`

### What's Missing

1. **No post-match navigation events** - After `match_completed`, user is stuck in Match Lobby
2. **No RoundManager** - No orchestration tracking winners joining pre-lobby for next round
3. **No winner-tracking state** - System doesn't know when all winners have joined pre-lobby
4. ~~**No participant validation**~~ - **EXISTS** in `CanAcceptParticipants()` (blocks eliminated players)
5. **Insufficient test coverage** for critical round transition flow

### What Already Exists (Reuse!)

1. `PreLobbyService.StartGracePeriod()` - Can reuse for inter-round grace
2. `PreLobbyService.CanAcceptParticipants()` - Already validates eliminated players
3. `PreLobbyService.BroadcastMatchAssigned()` - Can reuse for next round match assignment
4. `MatchService.GrantWalkover()` - Can reuse for absent players
5. `MatchService.HandleBothAbsent()` - Can reuse for both-absent scenario

---

## Requirements (from User)

| # | Requirement | Priority |
|---|-------------|----------|
| R1 | After match finish, winner receives event to return to pre-lobby | MUST |
| R2 | After match finish, loser receives elimination message | MUST |
| R3 | Tournament champion receives congratulation message | MUST |
| R4 | User MUST be in pre-lobby to start any round | MUST |
| R5 | Eliminated user cannot rejoin tournament | MUST |
| R6 | Backend needs RoundManager to control tournament flow | MUST |
| R7 | Frontend MUST load current state from up-to-date events | MUST |
| R8 | Tests required before implementation (unit, integration, e2e) | MUST |

---

## Architecture

### New Components

```
services/matchmaking/internal/
├── application/
│   ├── round_manager.go          # NEW: Orchestrates round progression
│   └── match_service.go          # MODIFY: Call RoundManager after completion
├── domain/
│   └── events.go                 # MODIFY: Add new events
└── websocket/
    └── hub.go                    # MODIFY: Add tournament broadcast methods
```

### Event Flow Diagram

```
                    ┌─────────────────┐
                    │  Match Started  │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │ Match Completed │   ◄── Runs in MATCH LOBBY (game)
                    └────────┬────────┘       NOT in pre-lobby!
                             │
               ┌─────────────┴─────────────┐
               │                           │
      ┌────────▼────────┐        ┌────────▼────────┐
      │  Winner Event   │        │  Loser Event    │
      │ (return_to_     │        │ (eliminated)    │
      │  prelobby)      │        └────────┬────────┘
      └────────┬────────┘                 │
               │                   ┌──────▼──────┐
               │                   │ Redirect to │
               │                   │ Home Page   │
               │                   └─────────────┘
      ┌────────▼────────┐
      │ Winner navigates│
      │ to Pre-Lobby    │
      │ (from modal)    │
      └────────┬────────┘
               │
      ┌────────▼───────────────────────────┐
      │ Winner joins Pre-Lobby WebSocket   │
      │                                    │
      │ RoundManager.OnWinnerJoined()      │
      │ tracks expected vs connected       │
      └────────┬───────────────────────────┘
               │
      ┌────────▼────────────────────────────┐
      │ ALL winners from round N            │
      │ joined pre-lobby?                   │──No──► Wait for more winners
      └────────┬────────────────────────────┘
               │ Yes
      ┌────────▼────────┐
      │ Is Final Round? │──Yes──► Tournament Complete
      └────────┬────────┘         (champion announced)
               │ No
      ┌────────▼────────────────────────────┐
      │ START GRACE PERIOD (30s)            │
      │ for late joiners                    │
      │                                     │
      │ All winners already in pre-lobby!   │
      │ Grace is for stragglers/reconnects  │
      │                                     │
      │ Broadcast: "grace_period_started"   │
      │ (reuse existing event)              │
      └────────┬────────────────────────────┘
               │
      ┌────────▼────────┐
      │ Grace Period    │
      │ Timer Running   │
      │ (countdown)     │
      └────────┬────────┘
               │
      ┌────────▼────────────────────────────┐
      │ GRACE PERIOD ENDS                   │
      │                                     │
      │ At this point all winners SHOULD    │
      │ be present (they triggered grace)   │
      │                                     │
      │ ✓ Present → match_assigned          │
      │ ✗ Disconnected → WALKOVER           │
      └────────┬────────────────────────────┘
               │
      ┌────────▼────────┐
      │ Players join    │
      │ Match Lobby     │
      └─────────────────┘
```

### Critical Insight: Where Code Runs

| Event | Location | Notes |
|-------|----------|-------|
| `MatchCompleted` | Match Lobby (game WebSocket) | Players still in game |
| `return_to_prelobby` | Match Lobby → Client | Tells winner to navigate |
| Winner navigates | Client-side | User clicks "Go to Pre-Lobby" |
| `participant_joined` | Pre-Lobby WebSocket | Winner connects to pre-lobby |
| Grace period starts | Pre-Lobby | Only when ALL winners connected |

### Key Policy Decisions

| Policy | Decision | Rationale |
|--------|----------|-----------|
| **Grace Period Trigger** | When ALL winners join pre-lobby | Grace period does NOT auto-start after round completion. Instead, RoundManager tracks winners connecting to pre-lobby. Only when ALL expected winners have connected → start 30s grace period for stragglers/reconnects. |
| Pre-Lobby Requirement | MANDATORY | Player MUST be connected to WebSocket to receive match assignment |
| Absence Handling | Walkover | If disconnected when grace ends → opponent wins by walkover |
| Grace Duration | 30 seconds | Short - all winners should already be connected when grace starts |
| Match Completion Context | Runs in Match Lobby | `CompleteMatch()` runs in game WebSocket context, NOT pre-lobby. This is why we can't auto-start grace period there. |

---

## Implementation Plan

### Phase 0: Test Infrastructure (Priority: HIGHEST)

**Goal:** Establish test foundation before any implementation.

#### 0.1 Unit Test Setup
```
services/matchmaking/internal/application/
├── round_manager_test.go         # Test round progression logic
└── match_service_test.go         # Extend with post-match event tests
```

**Test Cases:**
- [ ] `TestRoundManager_OnMatchCompleted_TracksWinner` - Winner added to expected list
- [ ] `TestRoundManager_OnMatchCompleted_RoundNotComplete` - Does not trigger when matches pending
- [ ] `TestRoundManager_OnMatchCompleted_FinalRound` - Skips final round (handled elsewhere)
- [ ] `TestRoundManager_OnParticipantJoined_NotExpectedWinner` - Ignores non-winners
- [ ] `TestRoundManager_OnParticipantJoined_TracksJoinedWinner` - Records winner joining
- [ ] `TestRoundManager_OnParticipantJoined_AllWinnersJoined` - Triggers grace period
- [ ] `TestRoundManager_GracePeriodPreventsDoubleStart` - GraceStarted flag works
- [ ] `TestRoundManager_AssignMatches_BothPresent` - Match assigned to both players
- [ ] `TestRoundManager_AssignMatches_OneAbsent` - Walkover granted
- [ ] `TestRoundManager_AssignMatches_BothAbsent` - Host notified
- [ ] `TestMatchService_PostMatchEvents_Winner` - Winner receives `return_to_prelobby`
- [ ] `TestMatchService_PostMatchEvents_Loser` - Loser receives `player_eliminated_notify`
- [ ] `TestMatchService_PostMatchEvents_Champion` - Champion receives `tournament_champion`

#### 0.2 Integration Test Setup
```
services/matchmaking/internal/
└── integration_test/
    ├── tournament_flow_test.go   # Full round progression
    └── prelobby_validation_test.go # Participant state validation
```

**Test Cases:**
- [ ] `TestTournamentFlow_TwoRounds` - 4 players, 2 rounds with winner tracking
- [ ] `TestTournamentFlow_EightPlayers` - Full bracket progression
- [ ] `TestTournamentFlow_GracePeriodOnAllWinnersJoined` - Grace starts when all winners in pre-lobby
- [ ] `TestTournamentFlow_WalkoverOnDisconnect` - Player disconnects during grace → walkover
- [ ] `TestPreLobby_EliminatedPlayerRejected` - Loser can't rejoin pre-lobby
- [ ] `TestRoundTransition_WinnerToPreLobby` - Winner navigates from match to pre-lobby

#### 0.3 E2E Test Setup
```
frontends/winspire-app/e2e/
└── TournamentFlowTests/
    ├── round-progression.spec.ts
    └── post-match-navigation.spec.ts
```

**Test Cases:**
- [ ] `test_winner_redirects_to_prelobby`
- [ ] `test_loser_redirects_to_home`
- [ ] `test_champion_sees_congratulation`

---

### Phase 1: Backend - New Events (Priority: HIGH)

**Files to modify:** `services/matchmaking/internal/domain/events.go`

#### 1.1 New Event Types

```go
// ============================================================================
// Post-Match Navigation Events (NEW)
// ============================================================================

// ReturnToPreLobbyPayload - sent to winner after match
type ReturnToPreLobbyPayload struct {
    TournamentID  uuid.UUID `json:"tournament_id"`
    MatchID       uuid.UUID `json:"match_id"`
    NextRoundNum  int       `json:"next_round_number"`
    Message       string    `json:"message"`
    PreLobbyURL   string    `json:"prelobby_url"`
}

// PlayerEliminatedNotificationPayload - sent to loser
type PlayerEliminatedNotificationPayload struct {
    TournamentID  uuid.UUID `json:"tournament_id"`
    MatchID       uuid.UUID `json:"match_id"`
    FinalPosition int       `json:"final_position"` // e.g., 5-8th place
    Message       string    `json:"message"`
}

// TournamentChampionPayload - sent to tournament winner
type TournamentChampionPayload struct {
    TournamentID uuid.UUID `json:"tournament_id"`
    ChampionID   uuid.UUID `json:"champion_id"`
    PrizeSummary string    `json:"prize_summary,omitempty"`
    Message      string    `json:"message"`
}

```

#### 1.2 Reuse Existing Events (extend payloads)

**Modify existing `PreLobbyGracePeriodStartedPayload`** - add `round_number`:

```go
// MODIFY: Add RoundNumber to existing payload
type PreLobbyGracePeriodStartedPayload struct {
    TournamentID     uuid.UUID `json:"tournament_id"`
    GracePeriodStart time.Time `json:"grace_period_start"`
    GracePeriodEnd   time.Time `json:"grace_period_end"`
    ParticipantCount int       `json:"participant_count"`
    RoundNumber      int       `json:"round_number"` // NEW: 1 = initial, 2+ = inter-round
}
```

**WebSocket messages to add (for individual player notifications):**

```go
const (
    // Post-match navigation (sent to individual player via match WebSocket)
    MsgTypeReturnToPreLobby       = "return_to_prelobby"
    MsgTypePlayerEliminatedNotify = "player_eliminated_notify"
    MsgTypeTournamentChampion     = "tournament_champion"

    // Existing - reuse for inter-round grace:
    // "grace_period_started" - already exists
    // "grace_period_ended" - already exists
)
```

---

### Phase 2: Backend - RoundManager (Priority: HIGH)

**Architecture: HYBRID** - MatchService notifies RoundManager on match completion, PreLobbyService notifies on winner join.

**Key insight:** Grace period cannot auto-start in `CompleteMatch()` because that runs in Match Lobby context (players still in game). Grace period starts only when ALL winners have connected to Pre-Lobby WebSocket.

#### 2.0 Integration Points

**2.0.1 MODIFY MatchService.CompleteMatch** - track round completion, send post-match events:

```go
// In match_service.go - at END of CompleteMatch():
func (s *MatchService) CompleteMatch(ctx context.Context, ...) error {
    // ... existing code (AdvanceWinner, ParticipantEliminated) ...

    // NEW: Send post-match navigation events to winner/loser
    s.sendPostMatchEvents(ctx, matchID, winnerID, loserID)

    // NEW: Notify round manager that match completed (for tracking)
    if s.roundManager != nil {
        s.roundManager.OnMatchCompleted(ctx, matchID, winnerID)
    }

    return nil
}

// Send WebSocket messages to players in Match Lobby
func (s *MatchService) sendPostMatchEvents(ctx context.Context, matchID, winnerID, loserID uuid.UUID) {
    match, _ := s.matchRepo.GetByID(ctx, matchID)
    round, _ := s.roundRepo.GetByID(ctx, match.RoundID)
    bracket, _ := s.bracketRepo.GetByID(ctx, round.BracketID)

    isFinalRound := match.NextMatchID == nil

    // Send to winner
    if isFinalRound {
        s.hub.SendToParticipant(bracket.TournamentID, winnerID, &websocket.Message{
            Type:    MsgTypeTournamentChampion,
            Payload: TournamentChampionPayload{...},
        })
    } else {
        s.hub.SendToParticipant(bracket.TournamentID, winnerID, &websocket.Message{
            Type:    MsgTypeReturnToPreLobby,
            Payload: ReturnToPreLobbyPayload{
                TournamentID: bracket.TournamentID,
                NextRoundNum: round.RoundNumber + 1,
                Message:      "Congratulations! Proceed to pre-lobby for next round.",
                PreLobbyURL:  fmt.Sprintf("/tournament/%s/prelobby", bracket.TournamentID),
            },
        })
    }

    // Send to loser
    s.hub.SendToParticipant(bracket.TournamentID, loserID, &websocket.Message{
        Type:    MsgTypePlayerEliminatedNotify,
        Payload: PlayerEliminatedNotificationPayload{...},
    })
}
```

**2.0.2 MODIFY PreLobbyService.JoinPreLobby** - notify RoundManager when winner joins:

```go
// In prelobby_service.go - when participant joins pre-lobby WebSocket:
func (s *PreLobbyService) JoinPreLobby(ctx context.Context, tournamentID, participantID uuid.UUID, conn *websocket.Conn) error {
    // ... existing validation (CanAcceptParticipants) ...

    // Register connection
    s.sessionStore.Add(tournamentID, participantID, conn)

    // NEW: Notify round manager that a participant (potentially winner) joined
    if s.roundManager != nil {
        s.roundManager.OnParticipantJoinedPreLobby(ctx, tournamentID, participantID)
    }

    return nil
}
```

**New file:** `services/matchmaking/internal/application/round_manager.go`

#### 2.1 Interface & State Definition

```go
// RoundManager orchestrates tournament progression between rounds
// Tracks: which round is "pending" (waiting for winners to join pre-lobby)
// Triggers: grace period when ALL expected winners have connected
type RoundManager struct {
    matchRepo       repository.MatchRepository
    roundRepo       repository.RoundRepository
    bracketRepo     repository.BracketRepository
    preLobbyService *PreLobbyService  // REUSE for grace period
    hub             *websocket.Hub
    logger          *observability.Logger

    // In-memory state: pending round transitions
    // Key: tournamentID, Value: RoundTransitionState
    pendingTransitions sync.Map
}

// RoundTransitionState tracks winners joining pre-lobby after round completes
type RoundTransitionState struct {
    mu              sync.Mutex
    TournamentID    uuid.UUID
    CompletedRound  int                // Round N that just finished
    NextRound       int                // Round N+1 to start
    ExpectedWinners map[uuid.UUID]bool // Winners who should join pre-lobby
    JoinedWinners   map[uuid.UUID]bool // Winners who have connected
    GraceStarted    bool               // Prevent double-start
}

// Key Methods:
// - OnMatchCompleted(ctx, matchID, winnerID)           -- track winner, check if round complete
// - OnParticipantJoinedPreLobby(ctx, tournamentID, participantID)  -- check if all winners joined
// - checkAndStartGracePeriod(...)                      -- start grace when all present
// - assignMatchesForRound(...)                         -- callback when grace ends
```

#### 2.2 Core Logic

```go
// OnMatchCompleted - called by MatchService after each match completion
// Tracks winner and checks if round is complete
func (rm *RoundManager) OnMatchCompleted(ctx context.Context, matchID, winnerID uuid.UUID) error {
    match, err := rm.matchRepo.GetByID(ctx, matchID)
    if err != nil {
        return fmt.Errorf("get match: %w", err)
    }

    // Skip final round - tournament completion handled elsewhere
    if match.NextMatchID == nil {
        rm.logger.Info("Final match completed, tournament ends", nil)
        return nil
    }

    round, _ := rm.roundRepo.GetByID(ctx, match.RoundID)
    bracket, _ := rm.bracketRepo.GetByID(ctx, round.BracketID)

    // Get or create transition state
    state := rm.getOrCreateTransitionState(bracket.TournamentID, round.RoundNumber)

    state.mu.Lock()
    defer state.mu.Unlock()

    // Add this winner to expected list
    state.ExpectedWinners[winnerID] = true

    // Check if ALL matches in round are done
    allMatches, _ := rm.matchRepo.GetByRoundID(ctx, round.ID)
    completedCount := 0
    for _, m := range allMatches {
        if m.Status == domain.MatchStatusCompleted {
            completedCount++
        }
    }

    if completedCount < len(allMatches) {
        rm.logger.Info("Round not complete yet", map[string]interface{}{
            "completed": completedCount,
            "total":     len(allMatches),
        })
        return nil
    }

    rm.logger.Info("Round complete! Waiting for winners to join pre-lobby", map[string]interface{}{
        "tournament_id":    bracket.TournamentID.String(),
        "completed_round":  round.RoundNumber,
        "expected_winners": len(state.ExpectedWinners),
    })

    return nil
}

// OnParticipantJoinedPreLobby - called by PreLobbyService when participant connects
// Checks if this was an expected winner and if all winners are now present
func (rm *RoundManager) OnParticipantJoinedPreLobby(ctx context.Context, tournamentID, participantID uuid.UUID) {
    val, exists := rm.pendingTransitions.Load(tournamentID)
    if !exists {
        // No pending transition for this tournament
        return
    }

    state := val.(*RoundTransitionState)
    state.mu.Lock()
    defer state.mu.Unlock()

    // Check if this is an expected winner
    if !state.ExpectedWinners[participantID] {
        return // Not a winner we're waiting for
    }

    // Mark winner as joined
    state.JoinedWinners[participantID] = true

    rm.logger.Info("Winner joined pre-lobby", map[string]interface{}{
        "tournament_id": tournamentID.String(),
        "participant":   participantID.String(),
        "joined":        len(state.JoinedWinners),
        "expected":      len(state.ExpectedWinners),
    })

    // Check if ALL winners have joined
    if len(state.JoinedWinners) >= len(state.ExpectedWinners) {
        rm.startGracePeriodForNextRound(ctx, state)
    }
}

// startGracePeriodForNextRound - ALL winners present, start 30s grace for reconnects
func (rm *RoundManager) startGracePeriodForNextRound(ctx context.Context, state *RoundTransitionState) {
    if state.GraceStarted {
        return // Prevent double-start
    }
    state.GraceStarted = true

    rm.logger.Info("All winners joined! Starting grace period", map[string]interface{}{
        "tournament_id": state.TournamentID.String(),
        "next_round":    state.NextRound,
    })

    // Callback when grace period ends
    onComplete := func(tID uuid.UUID, presentParticipantIDs []uuid.UUID) {
        rm.assignMatchesForRound(context.Background(), tID, state.NextRound, presentParticipantIDs)
        rm.pendingTransitions.Delete(tID) // Cleanup
    }

    // REUSE existing PreLobbyService.StartGracePeriod
    rm.preLobbyService.StartGracePeriod(ctx, state.TournamentID, onComplete)
}

func (rm *RoundManager) CheckRoundComplete(ctx context.Context, roundID uuid.UUID) (bool, error) {
    matches, err := rm.matchRepo.GetByRoundID(ctx, roundID)
    if err != nil {
        return false, err
    }
    for _, m := range matches {
        if m.Status != domain.MatchStatusCompleted {
            return false, nil
        }
    }
    return true, nil
}

// assignMatchesForRound - callback when grace period ends
// Assigns matches to players who are present in pre-lobby, grants walkovers to disconnected
func (rm *RoundManager) assignMatchesForRound(ctx context.Context, tournamentID uuid.UUID, roundNum int, presentIDs []uuid.UUID) {
    matches, _ := rm.matchRepo.GetByTournamentAndRound(ctx, tournamentID, roundNum)

    presentMap := make(map[uuid.UUID]bool)
    for _, id := range presentIDs {
        presentMap[id] = true
    }

    for _, match := range matches {
        p1Present := presentMap[match.Participant1ID]
        p2Present := match.Participant2ID != nil && presentMap[*match.Participant2ID]

        switch {
        case p1Present && p2Present:
            // Both present → REUSE PreLobbyService.BroadcastMatchAssigned
            p1 := rm.preLobbyService.GetParticipantDetails(tournamentID, match.Participant1ID)
            p2 := rm.preLobbyService.GetParticipantDetails(tournamentID, *match.Participant2ID)
            rm.preLobbyService.BroadcastMatchAssigned(tournamentID, match.Participant1ID, match.ID, roundNum, match.MatchNumber, &p2)
            rm.preLobbyService.BroadcastMatchAssigned(tournamentID, *match.Participant2ID, match.ID, roundNum, match.MatchNumber, &p1)

        case p1Present && !p2Present:
            // Grant walkover to P1
            rm.handleWalkover(ctx, match.ID, match.Participant1ID, *match.Participant2ID)

        case !p1Present && p2Present:
            // Grant walkover to P2
            rm.handleWalkover(ctx, match.ID, *match.Participant2ID, match.Participant1ID)

        default:
            // Both disconnected - escalate to host
            rm.handleBothAbsent(ctx, match.ID, tournamentID)
        }
    }
}

// Participant validation already exists in PreLobbyService.CanAcceptParticipants()
// See prelobby_service.go:315-323 - checks HasParticipantLostInTournament
```

#### 2.3 Helper Methods

```go
func (rm *RoundManager) getOrCreateTransitionState(tournamentID uuid.UUID, completedRound int) *RoundTransitionState {
    val, loaded := rm.pendingTransitions.LoadOrStore(tournamentID, &RoundTransitionState{
        TournamentID:    tournamentID,
        CompletedRound:  completedRound,
        NextRound:       completedRound + 1,
        ExpectedWinners: make(map[uuid.UUID]bool),
        JoinedWinners:   make(map[uuid.UUID]bool),
        GraceStarted:    false,
    })

    if loaded {
        return val.(*RoundTransitionState)
    }
    return val.(*RoundTransitionState)
}

func (rm *RoundManager) handleWalkover(ctx context.Context, matchID, winnerID, absentID uuid.UUID) {
    // Reuse existing MatchService.GrantWalkover (line 680 in match_service.go)
    // This already handles: marking match complete, advancing winner, eliminating absent player
    rm.logger.Info("Granting walkover", map[string]interface{}{
        "match_id": matchID.String(),
        "winner":   winnerID.String(),
        "absent":   absentID.String(),
    })
    // Call through injected match service reference or handle directly
}

func (rm *RoundManager) handleBothAbsent(ctx context.Context, matchID, tournamentID uuid.UUID) {
    // Reuse existing MatchService.HandleBothAbsent (line 766)
    // Notifies host to decide winner
    rm.logger.Warn("Both players absent", map[string]interface{}{
        "match_id":      matchID.String(),
        "tournament_id": tournamentID.String(),
    })
}
```

---

### Phase 3: Backend - PreLobby Validation (Priority: LOW - ALREADY EXISTS!)

> **NOTE:** Validation already implemented in `PreLobbyService.CanAcceptParticipants()` (line 298-326).
> It checks `matchRepo.HasParticipantLostInTournament()` and returns error if eliminated.

**No new code needed** - just verify existing validation is called on WebSocket connect.

#### 3.1 Existing Validation (prelobby_service.go:315-323)

```go
// ALREADY EXISTS - checks if participant has lost
if participantID != nil && (preLobby.Status == domain.PreLobbyStatusGeneratingBracket ||
                           preLobby.Status == domain.PreLobbyStatusStarted) {
    hasLost, err := s.matchRepo.HasParticipantLostInTournament(ctx, tournamentID, *participantID)
    if hasLost {
        return false, "You have been eliminated from this tournament and cannot rejoin", nil
    }
}
```

#### 3.2 Verify WebSocket Handler Uses Validation

```go
// In websocket_handler.go - ensure this is called on pre-lobby connect:
canAccept, reason, err := h.preLobbyService.CanAcceptParticipants(ctx, tournamentID, &userID)
if !canAccept {
    return sendError(conn, "ELIMINATED", reason)
}
```

---

### Phase 4: Frontend - New Event Handlers (Priority: HIGH)

**Modify:** `frontends/winspire-app/src/features/lobby/types.ts`

#### 4.1 Add New Message Types

```typescript
// Add to ServerMessageType
export type ServerMessageType =
  // ... existing types ...
  | 'return_to_prelobby'       // Winner should go to pre-lobby
  | 'player_eliminated_notify' // Loser eliminated from tournament
  | 'tournament_champion'      // User won the tournament
  | 'next_round_starting';     // Next round grace period begins

// New payload types
export interface ReturnToPreLobbyPayload {
  tournament_id: string;
  match_id: string;
  next_round_number: number;
  message: string;
  prelobby_url: string;
}

export interface PlayerEliminatedNotifyPayload {
  tournament_id: string;
  match_id: string;
  final_position: number;
  message: string;
}

export interface TournamentChampionPayload {
  tournament_id: string;
  champion_id: string;
  prize_summary?: string;
  message: string;
}
```

#### 4.2 Update useMatchLobby Hook

```typescript
// In useMatchLobby.ts - add handlers

const handleReturnToPreLobby = useCallback((payload: ReturnToPreLobbyPayload) => {
  setMatchState((prev) => {
    if (!prev) return null;
    return {
      ...prev,
      postMatchAction: {
        type: 'return_to_prelobby',
        message: payload.message,
        nextRoundNumber: payload.next_round_number,
        prelobbyUrl: payload.prelobby_url,
      }
    };
  });
}, []);

const handlePlayerEliminatedNotify = useCallback((payload: PlayerEliminatedNotifyPayload) => {
  setMatchState((prev) => {
    if (!prev) return null;
    return {
      ...prev,
      postMatchAction: {
        type: 'eliminated',
        message: payload.message,
        finalPosition: payload.final_position,
      }
    };
  });
}, []);

const handleTournamentChampion = useCallback((payload: TournamentChampionPayload) => {
  setMatchState((prev) => {
    if (!prev) return null;
    return {
      ...prev,
      postMatchAction: {
        type: 'champion',
        message: payload.message,
        prizeSummary: payload.prize_summary,
      }
    };
  });
}, []);
```

---

### Phase 5: Frontend - Post-Match Modal (Priority: MEDIUM)

**New component:** `frontends/winspire-app/src/features/lobby/components/PostMatchModal.tsx`

```typescript
interface PostMatchModalProps {
  action: PostMatchAction;
  winner: PlayerInfo | null;
  loser: PlayerInfo | null;
  currentUserId: string;
  onContinue: () => void;
}

export function PostMatchModal({ action, winner, loser, currentUserId, onContinue }: PostMatchModalProps) {
  const navigate = useNavigate();

  const handleContinue = () => {
    switch (action.type) {
      case 'return_to_prelobby':
        navigate(action.prelobbyUrl);
        break;
      case 'eliminated':
        navigate('/');
        break;
      case 'champion':
        navigate('/');
        break;
    }
    onContinue();
  };

  return (
    <Dialog open={true} onClose={() => {}}>
      <DialogPanel>
        {action.type === 'return_to_prelobby' && (
          <>
            <div className="text-8xl text-center">🎉</div>
            <DialogTitle>Wygrałeś mecz!</DialogTitle>
            <p>{action.message}</p>
            <p>Przejdź do poczekalni na rundę {action.nextRoundNumber}</p>
            <Button onClick={handleContinue}>Przejdź do poczekalni</Button>
          </>
        )}

        {action.type === 'eliminated' && (
          <>
            <div className="text-8xl text-center">😢</div>
            <DialogTitle>Koniec gry</DialogTitle>
            <p>{action.message}</p>
            <p>Zajmiesz miejsce: {action.finalPosition}</p>
            <Button onClick={handleContinue}>Wróć do strony głównej</Button>
          </>
        )}

        {action.type === 'champion' && (
          <>
            <div className="text-8xl text-center">🏆</div>
            <DialogTitle>Gratulacje! Jesteś mistrzem!</DialogTitle>
            <p>{action.message}</p>
            {action.prizeSummary && <p>Nagroda: {action.prizeSummary}</p>}
            <Button onClick={handleContinue}>Zakończ</Button>
          </>
        )}
      </DialogPanel>
    </Dialog>
  );
}
```

---

### Phase 6: Database Schema (Priority: MEDIUM)

**New migration:** `services/matchmaking/migrations/000XXX_add_participant_status.sql`

```sql
-- Add participant status tracking for tournament flow
ALTER TABLE prelobby_participants
ADD COLUMN IF NOT EXISTS tournament_status VARCHAR(20) DEFAULT 'active';

-- Possible values: 'active', 'eliminated', 'champion', 'withdrawn'

CREATE INDEX IF NOT EXISTS idx_prelobby_participants_status
ON prelobby_participants(tournament_id, user_id, tournament_status);

-- Track which round each participant is waiting for
ALTER TABLE prelobby_participants
ADD COLUMN IF NOT EXISTS waiting_for_round INT DEFAULT 1;
```

---

## Task Priority Matrix

| Task | Phase | Priority | Effort | Dependencies | Status |
|------|-------|----------|--------|--------------|--------|
| Unit test setup | 0.1 | HIGHEST | 3h | None | TODO |
| Integration test setup | 0.2 | HIGHEST | 4h | 0.1 | TODO |
| Post-match WebSocket events | 1.1 | HIGH | 1h | None | TODO |
| Modify MatchService.CompleteMatch | 2.0.1 | HIGH | 1h | 1.1 | TODO |
| Modify PreLobbyService.JoinPreLobby | 2.0.2 | HIGH | 0.5h | None | TODO |
| RoundManager (state tracking) | 2.1 | HIGH | 2h | 2.0.x | TODO |
| RoundManager (grace + assign) | 2.2-2.3 | HIGH | 2h | 2.1 | TODO |
| PreLobby validation | 3.x | LOW | 0h | - | **EXISTS** |
| Frontend types | 4.1 | HIGH | 1h | 1.1 | TODO |
| useMatchLobby handlers | 4.2 | HIGH | 2h | 4.1 | TODO |
| PostMatchModal component | 5 | MEDIUM | 2h | 4.x | TODO |
| Database migration | 6 | LOW | 0h | - | **NOT NEEDED** |
| E2E tests | 0.3 | HIGH | 4h | All above | TODO |

**Estimated Total:** ~18h

**Key Changes from Previous Design:**
- RoundManager now tracks winner joins (not just round completion)
- Grace period triggered by "all winners joined" (not "round complete")
- Two integration points: MatchService + PreLobbyService

---

## Implementation Order

```
1. [TESTS FIRST] Phase 0 - Write failing tests
   └── 0.1 Unit tests for RoundManager
       ├── TestRoundManager_OnMatchCompleted_TracksWinner
       ├── TestRoundManager_OnParticipantJoined_AllWinnersJoined
       └── TestRoundManager_AssignMatches_*
   └── 0.2 Unit tests for MatchService post-match events
   └── 0.3 Integration tests for round transitions

2. [BACKEND] Phase 1 - Post-match WebSocket message types
   └── 1.1 Add to domain/events.go:
       ├── MsgTypeReturnToPreLobby
       ├── MsgTypePlayerEliminatedNotify
       └── MsgTypeTournamentChampion
   └── 1.2 Add payload structs

3. [BACKEND] Phase 2 - RoundManager + Integration Points
   └── 2.0.1 Modify MatchService.CompleteMatch:
       ├── Call sendPostMatchEvents() for winner/loser
       └── Call roundManager.OnMatchCompleted(matchID, winnerID)
   └── 2.0.2 Modify PreLobbyService.JoinPreLobby:
       └── Call roundManager.OnParticipantJoinedPreLobby()
   └── 2.1 Create RoundManager struct + RoundTransitionState
   └── 2.2 Implement OnMatchCompleted (track expected winners)
   └── 2.3 Implement OnParticipantJoinedPreLobby (trigger grace when all join)
   └── 2.4 Implement assignMatchesForRound (callback after grace)

4. [BACKEND] Phase 3 - Validation (SKIP - already exists)
   └── Verify CanAcceptParticipants blocks eliminated players

5. [FRONTEND] Phase 4 - Event Handling
   └── 4.1 Add TypeScript payload types to types.ts
   └── 4.2 Add handlers in useMatchLobby:
       ├── handleReturnToPreLobby
       ├── handlePlayerEliminatedNotify
       └── handleTournamentChampion
   └── 4.3 Add postMatchAction state

6. [FRONTEND] Phase 5 - UI
   └── 5.1 PostMatchModal component with 3 variants:
       ├── Winner: "Proceed to Pre-Lobby" button
       ├── Loser: "Return to Home" button
       └── Champion: Celebration + "Finish" button

7. [TESTS] Phase 0.3 - E2E Tests
   └── test_winner_navigates_to_prelobby
   └── test_loser_redirects_home
   └── test_full_tournament_two_rounds
```

---

## Acceptance Criteria

**Post-Match Events (in Match Lobby):**
- [ ] After match, winner receives `return_to_prelobby` WebSocket message
- [ ] After match, loser receives `player_eliminated_notify` message
- [ ] Tournament champion receives `tournament_champion` message

**Frontend Navigation:**
- [ ] Winner is shown modal with "Przejdź do poczekalni" button
- [ ] Clicking button navigates to `/tournament/{id}/prelobby`
- [ ] Loser is shown modal with "Wróć do strony głównej" button
- [ ] Clicking button navigates to `/` (home)

**Grace Period Trigger (KEY CHANGE):**
- [ ] Grace period does NOT start when round completes
- [ ] Grace period starts ONLY when ALL winners from round N join pre-lobby WebSocket
- [ ] RoundManager tracks expected vs joined winners
- [ ] When last expected winner joins → `StartGracePeriod()` is called
- [ ] Grace period reuses existing `PreLobbyService.StartGracePeriod()`

**Match Assignment (after grace ends):**
- [ ] Both present → match assigned to both players via `BroadcastMatchAssigned()`
- [ ] One present, one disconnected → walkover to present player
- [ ] Both disconnected → host notified to decide

**Validation:**
- [ ] Eliminated players rejected by existing `CanAcceptParticipants()` validation
- [ ] Losers cannot reconnect to pre-lobby after elimination

**Testing:**
- [ ] Unit tests pass for RoundManager (all 10+ test cases)
- [ ] Integration tests pass for round flow (grace trigger on all winners joined)
- [ ] E2E tests verify full tournament with round transitions

---

## Open Questions (ANSWERED)

1. ~~**Grace period for inter-round pre-lobby?**~~ → **YES, 30s** (reuse StartGracePeriod)
2. ~~**Must player be in pre-lobby?**~~ → **YES, mandatory** (walkover if absent)
3. ~~**When should grace period start?**~~ → **When ALL winners join pre-lobby** (not when round completes!)
   - Critical insight: `CompleteMatch()` runs in Match Lobby (game) context
   - Players are NOT in pre-lobby when match ends
   - Grace period triggers only when all expected winners have connected to pre-lobby WebSocket
4. ~~**Architecture for RoundManager?**~~ → **HYBRID** (MatchService + PreLobbyService notify RoundManager)
5. **Spectator mode?** - Should eliminated players be able to watch remaining matches? (OUT OF SCOPE)
6. **Bracket view update** - Should bracket show live progress? (OUT OF SCOPE)

---

## Related Documents

- `docs-site/docs/001-tournament-matchmaking/research.md`
- `docs-site/docs/platform/bounded-contexts.md`
- `services/matchmaking/README.md`
