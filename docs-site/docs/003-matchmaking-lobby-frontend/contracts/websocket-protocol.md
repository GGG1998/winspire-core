# WebSocket Protocol: Tournament & Match Lobbies

**Feature**: 003-matchmaking-lobby-frontend  
**Version**: 1.1.0

## Overview

This protocol covers two WebSocket endpoints:
1. **Tournament Pre-Lobby**: `/v1/tournaments/{tournamentId}/lobby` - Waiting room before bracket generation
2. **Match Lobby**: `/v1/matches/{matchId}/lobby` - Per-match lobby after assignment

## Tournament Pre-Lobby Connection

### Endpoint
```
wss://api.winspire.com/api/matchmaking/v1/tournaments/{tournamentId}/lobby
```

### Authentication
- Bearer token required in initial HTTP upgrade request
- Token must be JWT from Supabase Auth
- User must be a registered participant in the tournament

### Pre-Lobby Messages

#### Server → Client: prelobby_state
Initial state sent immediately after connection.

```json
{
  "type": "prelobby_state",
  "payload": {
    "tournament": {
      "id": "tournament-uuid",
      "name": "Championship Finals",
      "startTime": "2025-12-04T20:00:00Z"
    },
    "participants": [
      {
        "id": "user-uuid-1",
        "displayName": "Player1",
        "avatarUrl": "https://example.com/avatar1.png",
        "joinedAt": "2025-12-04T19:55:00Z"
      }
    ],
    "gracePeriodActive": false,
    "gracePeriodEndsAt": null
  },
  "timestamp": "2025-12-04T19:55:00Z"
}
```

#### Server → Client: participant_joined
New participant joined the pre-lobby.

```json
{
  "type": "participant_joined",
  "payload": {
    "participant": {
      "id": "user-uuid-2",
      "displayName": "Player2",
      "avatarUrl": "https://example.com/avatar2.png",
      "joinedAt": "2025-12-04T19:56:00Z"
    },
    "totalCount": 8
  },
  "timestamp": "2025-12-04T19:56:00Z"
}
```

#### Server → Client: participant_left
Participant left the pre-lobby.

```json
{
  "type": "participant_left",
  "payload": {
    "participantId": "user-uuid-2",
    "totalCount": 7
  },
  "timestamp": "2025-12-04T19:57:00Z"
}
```

#### Server → Client: grace_period_started
Tournament start triggered, 30-second grace period begins.

```json
{
  "type": "grace_period_started",
  "payload": {
    "endsAt": "2025-12-04T20:00:30Z",
    "durationSeconds": 30
  },
  "timestamp": "2025-12-04T20:00:00Z"
}
```

#### Server → Client: roster_updated
Participant count changed during grace period.

```json
{
  "type": "roster_updated",
  "payload": {
    "participantCount": 9
  },
  "timestamp": "2025-12-04T20:00:10Z"
}
```

#### Server → Client: match_assigned
Player assigned to a match, includes 2-second redirect notification.

```json
{
  "type": "match_assigned",
  "payload": {
    "matchId": "match-uuid",
    "opponentName": "Player5",
    "roundName": "Quarter-finals"
  },
  "timestamp": "2025-12-04T20:00:30Z"
}
```

Frontend should show 2-second notification then redirect to `/lobby/:tournamentId/match/:matchId`.

---

## Match Lobby Connection

### Endpoint
```
wss://api.winspire.app/api/matchmaking/v1/matches/{matchId}/lobby
```

### Authentication
- Bearer token required in initial HTTP upgrade request
- Token must be JWT from Supabase Auth
- User must be a participant in the match

### Connection Example
```typescript
const ws = new WebSocket(
  `wss://api.winspire.app/api/matchmaking/v1/matches/${matchId}/lobby`,
  {
    headers: {
      Authorization: `Bearer ${accessToken}`
    }
  }
);
```

## Message Format

All messages are JSON with this structure:

```typescript
interface WebSocketMessage {
  type: string;           // Message type
  payload: unknown;       // Type-specific payload
  timestamp: string;      // ISO 8601 timestamp
  correlationId?: string; // Optional request-response correlation
}
```

## Client → Server Messages

### ping
Heartbeat to keep connection alive. Send every 30 seconds.

```json
{
  "type": "ping",
  "payload": {},
  "timestamp": "2025-12-04T10:00:00Z"
}
```

### ready
Mark player as ready for the match.

```json
{
  "type": "ready",
  "payload": {},
  "timestamp": "2025-12-04T10:00:00Z"
}
```

### unready
Cancel ready status (only allowed before countdown starts).

```json
{
  "type": "unready",
  "payload": {},
  "timestamp": "2025-12-04T10:00:00Z"
}
```

## Server → Client Messages

### pong
Response to ping heartbeat.

```json
{
  "type": "pong",
  "payload": {},
  "timestamp": "2025-12-04T10:00:00Z"
}
```

### lobby_state
Initial state sent immediately after connection.

```json
{
  "type": "lobby_state",
  "payload": {
    "match": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "round_id": "660e8400-e29b-41d4-a716-446655440001",
      "match_number": 1,
      "status": "ready",
      "participant1_id": "user-uuid-1",
      "participant2_id": "user-uuid-2",
      "participant1_ready": false,
      "participant2_ready": false,
      "winner_id": null,
      "score_player1": null,
      "score_player2": null
    },
    "participant1": {
      "id": "user-uuid-1",
      "displayName": "Player1",
      "avatarUrl": "https://example.com/avatar1.png"
    },
    "participant2": {
      "id": "user-uuid-2",
      "displayName": "Player2",
      "avatarUrl": "https://example.com/avatar2.png"
    }
  },
  "timestamp": "2025-12-04T10:00:00Z"
}
```

### player_joined
Opponent has joined the lobby.

```json
{
  "type": "player_joined",
  "payload": {
    "playerId": "user-uuid-2",
    "displayName": "Player2",
    "avatarUrl": "https://example.com/avatar2.png"
  },
  "timestamp": "2025-12-04T10:00:05Z"
}
```

### player_left
Opponent has left the lobby (closed connection).

```json
{
  "type": "player_left",
  "payload": {
    "playerId": "user-uuid-2"
  },
  "timestamp": "2025-12-04T10:00:10Z"
}
```

### ready_updated
A player's ready status has changed.

```json
{
  "type": "ready_updated",
  "payload": {
    "playerId": "user-uuid-1",
    "ready": true
  },
  "timestamp": "2025-12-04T10:00:15Z"
}
```

### match_starting
Both players are ready, countdown is starting.

```json
{
  "type": "match_starting",
  "payload": {
    "countdownSeconds": 3
  },
  "timestamp": "2025-12-04T10:00:20Z"
}
```

Server sends this message multiple times as countdown progresses:
- `countdownSeconds: 3`
- `countdownSeconds: 2`
- `countdownSeconds: 1`
- Then `match_started`

### match_started
Match has begun, game iframe should load.

```json
{
  "type": "match_started",
  "payload": {
    "gameUrl": "https://game.winspire.app/play?session=abc123&token=xyz789",
    "gameSessionId": "abc123"
  },
  "timestamp": "2025-12-04T10:00:23Z"
}
```

### player_disconnected
Opponent has disconnected during active match.

```json
{
  "type": "player_disconnected",
  "payload": {
    "playerId": "user-uuid-2",
    "disconnectedAt": "2025-12-04T10:05:00Z",
    "reconnectDeadline": "2025-12-04T10:05:30Z"
  },
  "timestamp": "2025-12-04T10:05:00Z"
}
```

Frontend should show 30-second countdown based on `reconnectDeadline`.

### player_reconnected
Disconnected player has returned within the window.

```json
{
  "type": "player_reconnected",
  "payload": {
    "playerId": "user-uuid-2"
  },
  "timestamp": "2025-12-04T10:05:15Z"
}
```

### match_completed
Match has finished with a result.

```json
{
  "type": "match_completed",
  "payload": {
    "winnerId": "user-uuid-1",
    "scorePlayer1": 3,
    "scorePlayer2": 1,
    "resultSource": "game_api"
  },
  "timestamp": "2025-12-04T10:15:00Z"
}
```

`resultSource` values:
- `game_api`: Automatic from game API
- `manual_host`: Host manually entered
- `walkover`: Opponent no-show/disconnect

### match_cancelled
Match has been cancelled (tournament cancelled or other reason).

```json
{
  "type": "match_cancelled",
  "payload": {
    "reason": "tournament_cancelled"
  },
  "timestamp": "2025-12-04T10:00:00Z"
}
```

### error
Error message from server.

```json
{
  "type": "error",
  "payload": {
    "code": "INVALID_STATE",
    "message": "Cannot mark ready, match has already started"
  },
  "timestamp": "2025-12-04T10:00:00Z"
}
```

Error codes:
- `INVALID_STATE`: Action not allowed in current state
- `NOT_PARTICIPANT`: User is not a participant
- `INTERNAL_ERROR`: Server error

## Connection Lifecycle

```
1. Client connects with JWT
2. Server validates JWT and participant status
3. Server sends `lobby_state` with current match state
4. Client sends `ping` every 30 seconds
5. Server responds with `pong`
6. When player clicks Ready: Client sends `ready`
7. Server broadcasts `ready_updated` to both clients
8. When both ready: Server sends `match_starting` countdown
9. Server sends `match_started` with game URL
10. Game plays (iframe)
11. Server sends `match_completed` with result
12. Client can close connection
```

## Reconnection Handling

If WebSocket connection drops:
1. Client should attempt reconnect with exponential backoff
2. Backoff: 1s → 2s → 4s → 8s → 16s → 30s (max)
3. On reconnect, server sends fresh `lobby_state`
4. Ready status is preserved server-side
5. If in game, `gameUrl` is included in lobby state

```typescript
let reconnectDelay = 1000;
const maxDelay = 30000;

function reconnect() {
  setTimeout(() => {
    ws = new WebSocket(url);
    ws.onclose = () => {
      reconnectDelay = Math.min(reconnectDelay * 2, maxDelay);
      reconnect();
    };
    ws.onopen = () => {
      reconnectDelay = 1000; // Reset on successful connect
    };
  }, reconnectDelay);
}
```

