# Implementation Plan: Matchmaking Lobby Frontend

**Branch**: `003-matchmaking-lobby-frontend` | **Date**: 2025-12-04 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `docs-site/docs/003-matchmaking-lobby-frontend/spec.md`

## Summary

Build the frontend for the tournament matchmaking system including: tournament pre-lobby with grace period, match lobby with player VS display and game iframe, interactive bracket visualization, matches list view with real-time updates, and ready/disconnect handling. Integrates with existing matchmaking service REST API and WebSocket endpoints.

## Technical Context

**Language/Version**: TypeScript 5.9+  
**Primary Dependencies**: React 19, Vite 7, Tailwind CSS v4, React Router v7, React Query (TanStack Query), Zod  
**Storage**: N/A (frontend consumes backend API)  
**Testing**: Playwright (E2E), Vitest (unit)  
**Target Platform**: Web browser (Chrome, Firefox, Safari, Edge), Mobile responsive  
**Project Type**: Single Page Application (SPA) - frontend  
**Performance Goals**: 
- Page load < 2s
- Real-time updates visible within 1s (WebSocket)
- Game iframe interactive within 5s of match start
- Grace period roster updates < 500ms
**Constraints**: 
- Must integrate with existing matchmaking service API
- Must handle WebSocket reconnection gracefully
- Must support responsive design for mobile
- Grace period timing must be precise (30s window)
**Scale/Scope**: Support tournaments up to 128 participants, 50 concurrent tournaments

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Modular Monorepo Compliance ✅

| Principle | Status | Details |
|-----------|--------|---------|
| Feature placement | ✅ | Frontend app in `frontends/winspire-app/` |
| Feature-based structure | ✅ | New components in `features/lobby/` and extending `features/host/` |
| Independent module | ✅ | Uses existing `package.json` in winspire-app |
| Shared libraries | ✅ | Reuses `shared/components/ui/` for UI elements |

### Technology Stack Compliance ✅

| Component | Required | Planned | Status |
|-----------|----------|---------|--------|
| Language | TypeScript 5.9+ | TypeScript 5.9+ | ✅ |
| Framework | React 19 | React 19 | ✅ |
| Build Tool | Vite 7 | Vite 7 | ✅ |
| Styling | Tailwind CSS v4 | Tailwind CSS v4 | ✅ |
| Routing | React Router v7 | React Router v7 | ✅ |
| Forms | React Hook Form + Zod | Zod for validation | ✅ |
| State | React Context + React Query | React Query + Context | ✅ |
| UI Components | Headless UI | Headless UI | ✅ |

### File Naming Conventions ✅

- Components: PascalCase (e.g., `TournamentPreLobby.tsx`, `BracketTree.tsx`)
- Hooks: camelCase with `use` prefix (e.g., `useTournamentPreLobby.ts`, `useBracket.ts`)
- Types: `types.ts` per feature module
- API: `matchmakingApi.ts` in `features/lobby/api/`

### Bounded Contexts ✅

Frontend consumes matchmaking API without crossing bounded context boundaries:
- ✅ Lobby feature calls matchmaking service REST/WebSocket endpoints
- ✅ Tournament detail page consumes competition service for base tournament data
- ✅ No direct database access - all via API

## Project Structure

### Documentation (this feature)

```text
docs-site/docs/003-matchmaking-lobby-frontend/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (frontend API client contracts)
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
frontends/winspire-app/src/
├── features/
│   ├── lobby/                      # EXTEND existing
│   │   ├── api/
│   │   │   └── matchmakingApi.ts   # NEW: Matchmaking service API client
│   │   ├── components/
│   │   │   ├── TournamentPreLobby.tsx        # NEW: Pre-lobby waiting room
│   │   │   ├── GracePeriodIndicator.tsx     # NEW: Grace period countdown
│   │   │   ├── ParticipantList.tsx          # NEW: Live participant list
│   │   │   ├── ActivityFeed.tsx             # NEW: Joins/leaves feed
│   │   │   ├── LobbyView.tsx                # UPDATE: Full match lobby layout
│   │   │   ├── PlayerVsDisplay.tsx          # NEW: VS header with avatars
│   │   │   ├── ReadyButton.tsx              # NEW: Ready state button
│   │   │   ├── GameFrame.tsx                # NEW: Game iframe container
│   │   │   ├── DisconnectOverlay.tsx        # NEW: Reconnect countdown
│   │   │   └── MatchResult.tsx              # NEW: Post-match result display
│   │   ├── hooks/
│   │   │   ├── useTournamentPreLobby.ts     # NEW: Pre-lobby state management
│   │   │   ├── useMatchLobby.ts             # UPDATE: WebSocket state management
│   │   │   ├── useReadyState.ts             # NEW: Ready state logic
│   │   │   └── useDisconnect.ts             # NEW: Disconnect handling
│   │   ├── pages/
│   │   │   ├── TournamentPreLobbyPage.tsx   # NEW: Pre-lobby page
│   │   │   └── MatchLobbyPage.tsx           # UPDATE: Full lobby page
│   │   ├── schemas.ts              # NEW: Zod validation schemas
│   │   ├── types.ts                # UPDATE: Match/lobby types
│   │   └── constants.ts            # NEW: Lobby-specific constants
│   │
│   └── host/                       # EXTEND existing
│       ├── components/
│       │   ├── BracketView.tsx     # UPDATE: Interactive bracket tree
│       │   ├── BracketMatch.tsx    # NEW: Single match in bracket
│       │   ├── MatchesView.tsx     # UPDATE: Real data integration
│       │   └── MatchCard.tsx       # NEW: Match card component
│       ├── api/
│       │   └── tournamentApi.ts    # UPDATE: Add bracket endpoints
│       └── types.ts                # UPDATE: Bracket/match types
│
└── shared/
    └── hooks/
        └── useWebSocket.ts         # NEW: Generic WebSocket hook
```

**Structure Decision**: Extends existing `features/lobby/` and `features/host/` directories per constitution's feature-based structure. Adds tournament pre-lobby as entry point before match lobby.

## Complexity Tracking

> No constitution violations - complexity is within acceptable bounds.

| Decision | Rationale |
|----------|-----------|
| Pre-lobby with grace period | Essential for fair late-arrival handling, prevents no-shows |
| WebSocket in shared hook | Reusable real-time infrastructure for future features |
| Separate GameFrame component | Isolates iframe complexity and security handling |
| Bracket visualization custom-built | Existing libs (react-brackets) don't match design requirements |
