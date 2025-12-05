# Quickstart: Matchmaking Lobby Frontend

**Feature**: 003-matchmaking-lobby-frontend  
**Date**: 2025-12-04

## Prerequisites

1. Node.js 20+ and pnpm installed
2. Backend matchmaking service running (see `services/matchmaking/README.md`)
3. Local development environment set up (see `platform/local/README.md`)

## Setup

```bash
# Navigate to frontend app
cd frontends/winspire-app

# Install dependencies (if not already done)
pnpm install

# Start development server
pnpm dev
```

Frontend runs at `http://localhost:5173`

## Environment Variables

Create `.env.local` in `frontends/winspire-app/`:

```env
# API Gateway (routes to matchmaking service)
VITE_API_BASE_URL=http://localhost:8080/api

# Matchmaking service direct (for WebSocket)
VITE_MATCHMAKING_WS_URL=ws://localhost:8082/api/matchmaking

# Supabase Auth
VITE_SUPABASE_URL=http://localhost:54321
VITE_SUPABASE_ANON_KEY=your-anon-key
```

## Key Routes

| Route | Component | Description |
|-------|-----------|-------------|
| `/tournaments/:tournamentId/lobby` | `TournamentPreLobbyPage` | Pre-lobby waiting room (before matches created) |
| `/lobby/:tournamentId/match/:matchId` | `MatchLobbyPage` | Match lobby with VS display |
| `/h/:streamerId/tournaments/:tournamentId` | `TournamentDetailPage` | Tournament with bracket/matches tabs |

## Feature Structure

```
src/features/
├── lobby/
│   ├── api/
│   │   └── matchmakingApi.ts     # REST + WebSocket client
│   ├── components/
│   │   ├── PlayerVsDisplay.tsx   # VS header
│   │   ├── ReadyButton.tsx       # Ready state button
│   │   ├── GameFrame.tsx         # Game iframe
│   │   └── MatchResult.tsx       # Result display
│   ├── hooks/
│   │   ├── useMatchLobby.ts      # Lobby state management
│   │   └── useWebSocket.ts       # WebSocket connection
│   ├── pages/
│   │   └── MatchLobbyPage.tsx    # Main lobby page
│   └── types.ts                  # Type definitions
│
└── host/
    └── components/
        ├── BracketView.tsx       # Interactive bracket
        └── MatchesView.tsx       # Matches list
```

## Quick Development Tasks

### 1. Test Tournament Pre-Lobby

Navigate to pre-lobby (requires tournament with registration open):

```
http://localhost:5173/tournaments/{tournamentId}/lobby
```

**What to test:**
- Participant list loads with avatars
- Countdown to tournament start displays
- Activity feed shows join/leave events
- Grace period (30s) after start
- Late arrivals update participant count
- Auto-redirect to match lobby after match assigned

### 2. Run the Match Lobby Page

Navigate to a match lobby (requires running tournament with generated bracket):

```
http://localhost:5173/lobby/{tournamentId}/match/{matchId}
```

### 3. Test WebSocket Connection

Open browser DevTools → Network → WS to see WebSocket messages.

### 4. View Bracket

Navigate to tournament detail page:

```
http://localhost:5173/h/{streamerId}/tournaments/{tournamentId}
```

Click "Drabinka" (Bracket) tab.

## API Integration Examples

### Fetch Pre-Lobby State

```typescript
// src/features/lobby/api/matchmakingApi.ts
export async function getPreLobbyState(tournamentId: string): Promise<PreLobbyState> {
  const response = await fetch(
    `${API_BASE}/matchmaking/v1/tournaments/${tournamentId}/lobby`,
    {
      headers: {
        Authorization: `Bearer ${getAccessToken()}`,
      },
    }
  );
  if (!response.ok) throw new Error('Failed to fetch pre-lobby state');
  return transformPreLobbyResponse(await response.json());
}
```

### Pre-Lobby WebSocket Hook

```typescript
// src/features/lobby/hooks/useTournamentPreLobby.ts
export function useTournamentPreLobby(tournamentId: string) {
  const [preLobbyState, setPreLobbyState] = useState<PreLobbyState | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const ws = new WebSocket(
      `${WS_BASE}/v1/tournaments/${tournamentId}/lobby`
    );

    ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      
      switch (message.type) {
        case 'prelobby_state':
          setPreLobbyState(message.payload);
          break;
        case 'participant_joined':
          // Update participant list
          break;
        case 'grace_period_started':
          // Start 30s countdown
          break;
        case 'match_assigned':
          // Show notification, redirect after 2s
          const { matchId, opponentName } = message.payload;
          showNotification(`Matched with ${opponentName}!`);
          setTimeout(() => {
            navigate(`/lobby/${tournamentId}/match/${matchId}`);
          }, 2000);
          break;
      }
    };

    wsRef.current = ws;
    return () => ws.close();
  }, [tournamentId]);

  return { preLobbyState };
}
```

### Fetch Bracket

```typescript
// src/features/host/api/tournamentApi.ts
export async function getBracket(tournamentId: string): Promise<Bracket> {
  const response = await fetch(
    `${API_BASE}/matchmaking/v1/tournaments/${tournamentId}/bracket`,
    {
      headers: {
        Authorization: `Bearer ${getAccessToken()}`,
      },
    }
  );
  if (!response.ok) throw new Error('Failed to fetch bracket');
  return transformBracketResponse(await response.json());
}
```

### Match Lobby WebSocket Hook

```typescript
// src/features/lobby/hooks/useMatchLobby.ts
export function useMatchLobby(matchId: string) {
  const [lobbyState, setLobbyState] = useState<LobbyState | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const ws = new WebSocket(
      `${WS_BASE}/v1/matches/${matchId}/lobby`
    );

    ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      handleMessage(message, setLobbyState);
    };

    wsRef.current = ws;
    return () => ws.close();
  }, [matchId]);

  const markReady = () => {
    wsRef.current?.send(JSON.stringify({ type: 'ready', payload: {} }));
  };

  return { lobbyState, markReady };
}
```

### Ready Button Component

```tsx
// src/features/lobby/components/ReadyButton.tsx
export function ReadyButton({ isReady, onReady }: ReadyButtonProps) {
  return (
    <Button
      onClick={onReady}
      disabled={isReady}
      color={isReady ? 'emerald' : 'cyan'}
    >
      {isReady ? (
        <>
          <CheckIcon className="w-5 h-5" />
          Gotowy
        </>
      ) : (
        'Gotowy!'
      )}
    </Button>
  );
}
```

## Testing

### E2E Tests (Playwright)

```bash
# Run E2E tests
pnpm test:e2e

# Run specific test file
pnpm test:e2e lobby.spec.ts
```

### Unit Tests (Vitest)

```bash
# Run unit tests
pnpm test

# Watch mode
pnpm test:watch
```

## Common Issues

### WebSocket Connection Fails

1. Check matchmaking service is running on port 8082
2. Verify JWT token is valid
3. Check CORS configuration allows WebSocket upgrade

### Bracket Not Loading

1. Verify tournament is in "started" state
2. Check bracket was generated (`GET /v1/tournaments/{id}/bracket`)
3. Ensure user has permission to view tournament

### Game Iframe Blocked

1. Check CSP headers allow iframe src
2. Verify game URL is correctly formatted
3. Test game URL directly in browser

## Next Steps

1. Read [spec.md](./spec.md) for full requirements
2. Review [data-model.md](./data-model.md) for type definitions
3. Check [contracts/matchmaking-api.yaml](./contracts/matchmaking-api.yaml) for API details
4. See [contracts/websocket-protocol.md](./contracts/websocket-protocol.md) for WebSocket messages

