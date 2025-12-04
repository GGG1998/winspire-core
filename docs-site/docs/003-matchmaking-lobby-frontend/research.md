# Research: Matchmaking Lobby Frontend

**Feature**: 003-matchmaking-lobby-frontend  
**Date**: 2025-12-04

## Research Tasks Completed

### 1. Tournament Pre-Lobby Flow

**Context**: Players need a waiting room before matches are created, with grace period for late arrivals.

**Decision**: Two-stage lobby system
- Stage 1: Tournament Pre-Lobby (`/tournaments/:tournamentId/lobby`) - waiting room before bracket generation
- Stage 2: Match Lobby (`/lobby/:tournamentId/match/:matchId`) - per-match room after assignment
- 30-second grace period after tournament start for late arrivals

**Rationale**:
- Backend API exposes `/tournaments/{tournament_id}/lobby` endpoint
- Prevents premature bracket generation before all players ready
- Grace period provides fairness for slight delays (network, tab switching)
- Matches backend behavior from 001-tournament-matchmaking spec

**Alternatives Considered**:
- Direct to match: Confusing for players before matches exist
- No grace period: Too strict, excludes players with minor delays
- Longer grace period (60s+): Delays tournament start too much

### 2. Backend API Integration Pattern

**Context**: Need to understand how frontend will communicate with matchmaking service.

**Decision**: Hybrid REST + WebSocket approach
- REST API for initial data fetch and mutations (ready status, walkover claims)
- WebSocket for real-time state updates in match lobby

**Rationale**: 
- Backend already exposes REST endpoints (`GET /v1/matches/:id`, `POST /v1/matches/:id/ready`)
- WebSocket endpoint exists (`GET /v1/matches/:id/lobby`) for real-time lobby updates
- This pattern is already established in the codebase

**Alternatives Considered**:
- Server-Sent Events (SSE): Simpler but uni-directional, WebSocket already implemented
- Polling: Higher latency and server load, rejected for lobby use case

### 2. Bracket Visualization Approach

**Context**: Need interactive bracket tree for single elimination tournaments up to 128 participants.

**Decision**: Custom SVG-based bracket component with CSS Grid layout
- Horizontal layout: Rounds flow left to right
- Connector lines: SVG paths between match cards
- Interactions: Click to navigate/view details, hover states

**Rationale**:
- Existing libraries (react-tournament-bracket, react-brackets) don't match design aesthetics
- Custom implementation provides full control over styling and animations
- SVG paths allow smooth connector animations

**Alternatives Considered**:
- react-brackets: Limited customization, doesn't support our design
- Canvas-based: Harder to make accessible and interactive
- Pure CSS: Connector lines become complex for large brackets

### 3. Game Iframe Communication

**Context**: Game loads in iframe, need to communicate match session and receive completion signals.

**Decision**: URL parameters + postMessage API
- Session token passed via iframe src URL parameter
- Match completion detected via postMessage from game
- Fallback: Backend polling for match result if postMessage fails

**Rationale**:
- postMessage is the secure cross-origin communication standard
- URL parameters provide initial session without same-origin requirement
- Backend polling ensures reliability if iframe communication fails

**Alternatives Considered**:
- Only URL parameters: No way to receive game completion signal
- Only polling: Adds latency to result detection
- Shared cookies: Security concerns with third-party game

### 4. WebSocket State Management

**Context**: Need to manage WebSocket connection lifecycle, reconnection, and state synchronization.

**Decision**: Custom hook with React Query integration
- `useWebSocket` hook manages connection lifecycle
- Connection state tracked in React state
- Message handlers update React Query cache for consistency
- Auto-reconnect with exponential backoff (1s, 2s, 4s, max 30s)

**Rationale**:
- React Query handles caching and cache invalidation
- Custom hook provides reusable pattern
- Exponential backoff prevents server overload during outages

**Alternatives Considered**:
- socket.io-client: Overkill for our needs, adds bundle size
- Native WebSocket only: Need reconnection logic anyway
- Zustand for state: Over-engineering for this scope

### 5. Mobile Responsiveness Strategy

**Context**: Lobby must work on mobile devices with smaller screens.

**Decision**: Responsive breakpoints with layout restructuring
- Desktop: VS display horizontal, game iframe full width
- Tablet: Same as desktop, smaller fonts
- Mobile: VS display stacked vertically, game iframe adapts to viewport
- Bracket: Horizontal scroll with pinch-to-zoom on mobile

**Rationale**:
- Game iframe must remain playable, not shrunk too small
- VS display information is critical, must be visible
- Bracket naturally scrolls horizontally even on desktop

**Alternatives Considered**:
- Separate mobile app: Out of scope, not requested
- Hide bracket on mobile: Users need bracket visibility
- Force landscape: Poor UX, many users won't rotate

### 6. Ready State Persistence

**Context**: Ready status must persist across browser refresh per spec (FR-008).

**Decision**: Server-side persistence with optimistic UI updates
- POST `/v1/matches/:id/ready` persists state server-side
- UI shows optimistic "Ready" immediately on click
- WebSocket confirms actual state to both players
- On page load, fetch current match state including ready status

**Rationale**:
- Backend already implements ready state persistence
- Optimistic update provides snappy UX
- WebSocket sync handles race conditions

**Alternatives Considered**:
- localStorage: Wouldn't work across devices/browsers
- Session-only: Spec explicitly requires persistence across refresh

### 7. Disconnect Detection Strategy

**Context**: Need to detect when opponent disconnects and show countdown.

**Decision**: WebSocket close event + heartbeat + backend notification
- WebSocket close triggers disconnect detection
- Backend sends `player_disconnected` event via WebSocket to opponent
- Frontend shows 30s countdown based on `disconnected_at` timestamp
- Reconnection detected via `player_reconnected` event

**Rationale**:
- Backend already tracks disconnect timestamps (`disconnected_at`, `disconnected_player_id`)
- Server-authoritative timestamps ensure fair countdown
- WebSocket events provide real-time notification

**Alternatives Considered**:
- Client-side heartbeat only: Can't trust client timing
- Polling for disconnect: Too slow for 30s window

## Technology Decisions Summary

| Area | Technology | Version |
|------|------------|---------|
| Pre-Lobby Flow | Two-stage lobby system + 30s grace period | N/A |
| State Management | React Query + WebSocket hook | TanStack Query v5 |
| Real-time | Native WebSocket with custom reconnect | N/A |
| Bracket Rendering | Custom SVG + CSS Grid | N/A |
| Iframe Communication | postMessage API | Browser standard |
| API Client | Fetch with typed responses | N/A |
| Validation | Zod schemas | Zod 3.x |

## Open Questions (Resolved)

1. ~~What is the WebSocket message format?~~ → JSON with `type` and `payload` fields
2. ~~How does backend signal match start?~~ → `match_started` WebSocket event
3. ~~What game URL format is used?~~ → Configurable via tournament settings, includes session token

