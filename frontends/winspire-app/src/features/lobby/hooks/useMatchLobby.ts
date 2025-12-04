import { useState, useEffect, useRef, useCallback } from 'react';
import { useWebSocket } from '../../../shared/hooks/useWebSocket';
import type {
  Match,
  PlayerInfo,
  ConnectionState,
  ServerMessageType,
  LobbyStatePayload,
  ReadyUpdatedPayload,
  MatchStartingPayload,
  MatchStartedPayload,
} from '../types';
import { matchmakingApi } from '../api/matchmakingApi';

// Player joined/left payloads (for match lobby)
interface PlayerJoinedPayload {
  playerId: string;
  player: PlayerInfo;
}

interface PlayerLeftPayload {
  playerId: string;
}

// Match lobby state (includes match + players)
export interface MatchLobbyState {
  match: Match;
  player1: PlayerInfo | null;
  player2: PlayerInfo | null;
  roundNumber: number;
  tournament?: {
    id: string;
    name: string;
    creatorId: string;
  };
  status: Match['status'];
}

interface UseMatchLobbyReturn {
  matchState: MatchLobbyState | null;
  isLoading: boolean;
  error: string | null;
  connectionStatus: ConnectionState;
}

/**
 * useMatchLobby Hook
 * 
 * Manages match lobby state and WebSocket connection
 * Features:
 * - WebSocket connection to /v1/matches/:id/lobby
 * - Real-time lobby state updates
 * - Player join/leave notifications
 * - Ready state tracking
 * - Match starting countdown
 * - Auto-reconnection on disconnect
 */
export function useMatchLobby(matchId: string | null): UseMatchLobbyReturn {
  const [matchState, setMatchState] = useState<MatchLobbyState | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Track if initial data is loaded
  const initialLoadComplete = useRef(false);

  // Construct WebSocket URL
  // Note: The actual WS base URL would be configured in environment
  const wsUrl = matchId ? `/v1/matches/${matchId}/lobby` : '';

  // Initialize WebSocket connection
  const { status: connectionStatus } = useWebSocket({
    url: wsUrl,
    onMessage: handleWebSocketMessage,
    onError: (error) => {
      console.error('[useMatchLobby] WebSocket error:', error);
      setError('Błąd połączenia WebSocket');
    },
  });

  // Load initial match data via REST API
  useEffect(() => {
    if (!matchId) {
      setIsLoading(false);
      return;
    }

    const loadMatchData = async () => {
      try {
        setIsLoading(true);
        setError(null);

        const matchData = await matchmakingApi.getMatch(matchId);
        
        // Initialize match state from REST response
        setMatchState({
          match: matchData.match,
          player1: matchData.participant1,
          player2: matchData.participant2,
          roundNumber: matchData.roundNumber || 1,
          tournament: matchData.tournament,
          status: matchData.match.status,
        });

        initialLoadComplete.current = true;
      } catch (err) {
        console.error('[useMatchLobby] Failed to load match:', err);
        setError(err instanceof Error ? err.message : 'Nie udało się załadować meczu');
      } finally {
        setIsLoading(false);
      }
    };

    loadMatchData();
  }, [matchId]);

  // WebSocket message handler
  function handleWebSocketMessage(event: MessageEvent) {
    try {
      const message = JSON.parse(event.data);
      const messageType = message.type as ServerMessageType;

      console.log('[useMatchLobby] Received WebSocket message:', messageType, message);

      switch (messageType) {
        case 'lobby_state':
          handleLobbyState(message.payload as LobbyStatePayload);
          break;
        case 'player_joined':
          handlePlayerJoined(message.payload as PlayerJoinedPayload);
          break;
        case 'player_left':
          handlePlayerLeft(message.payload as PlayerLeftPayload);
          break;
        case 'ready_updated':
          handleReadyUpdated(message.payload as ReadyUpdatedPayload);
          break;
        case 'match_starting':
          handleMatchStarting(message.payload as MatchStartingPayload);
          break;
        case 'match_started':
          handleMatchStarted(message.payload as MatchStartedPayload);
          break;
        default:
          console.warn('[useMatchLobby] Unknown message type:', messageType);
      }
    } catch (err) {
      console.error('[useMatchLobby] Failed to parse WebSocket message:', err);
    }
  }

  // Handle lobby_state message (full state update)
  const handleLobbyState = useCallback((payload: LobbyStatePayload) => {
    console.log('[useMatchLobby] Lobby state update:', payload);
    
    setMatchState((prev) => ({
      match: payload.match,
      player1: payload.participant1,
      player2: payload.participant2,
      roundNumber: prev?.roundNumber || 1,
      tournament: prev?.tournament,
      status: payload.match.status,
    }));
  }, []);

  // Handle player_joined message
  const handlePlayerJoined = useCallback((payload: PlayerJoinedPayload) => {
    console.log('[useMatchLobby] Player joined:', payload);
    
    setMatchState((prev) => {
      if (!prev) return null;

      // Determine which player slot this is
      if (payload.playerId === prev.match.participant1Id) {
        return {
          ...prev,
          player1: payload.player,
        };
      } else if (payload.playerId === prev.match.participant2Id) {
        return {
          ...prev,
          player2: payload.player,
        };
      }

      return prev;
    });
  }, []);

  // Handle player_left message
  const handlePlayerLeft = useCallback((payload: PlayerLeftPayload) => {
    console.log('[useMatchLobby] Player left:', payload);
    
    // Note: In match lobby, we don't remove players on disconnect
    // We just track connection status (handled by opponent_disconnected message)
    // This message is mainly for informational purposes
  }, []);

  // Handle ready_updated message
  const handleReadyUpdated = useCallback((payload: ReadyUpdatedPayload) => {
    console.log('[useMatchLobby] Ready updated:', payload);
    
    setMatchState((prev) => {
      if (!prev) return null;

      // Update ready status for the appropriate player
      const updatedMatch = { ...prev.match };
      
      if (payload.playerId === updatedMatch.participant1Id) {
        updatedMatch.participant1Ready = payload.ready;
      } else if (payload.playerId === updatedMatch.participant2Id) {
        updatedMatch.participant2Ready = payload.ready;
      }

      return {
        ...prev,
        match: updatedMatch,
      };
    });
  }, []);

  // Handle match_starting message (countdown begins)
  const handleMatchStarting = useCallback((payload: MatchStartingPayload) => {
    console.log('[useMatchLobby] Match starting countdown:', payload.countdownSeconds);
    
    setMatchState((prev) => {
      if (!prev) return null;

      return {
        ...prev,
        status: 'started', // Transition to started during countdown
      };
    });

    // TODO: Will be handled by MatchStartCountdown component in Phase 5
  }, []);

  // Handle match_started message (game begins)
  const handleMatchStarted = useCallback((payload: MatchStartedPayload) => {
    console.log('[useMatchLobby] Match started:', payload);
    
    setMatchState((prev) => {
      if (!prev) return null;

      const updatedMatch = {
        ...prev.match,
        status: 'started' as const,
        gameApiMatchId: payload.gameSessionId,
        startedAt: new Date().toISOString(),
      };

      return {
        ...prev,
        match: updatedMatch,
        status: 'started',
      };
    });

    // TODO: Will trigger GameFrame component in Phase 6
  }, []);

  return {
    matchState,
    isLoading,
    error,
    connectionStatus,
  };
}

