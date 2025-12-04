import { useState, useEffect, useRef, useCallback } from 'react';
import { useWebSocket } from '../../../shared/hooks/useWebSocket';
import { matchmakingApi } from '../api/matchmakingApi';
import { useAuth } from '../../auth';
import type {
  Match,
  PlayerInfo,
  ConnectionState,
  ServerMessageType,
  LobbyStatePayload,
  ReadyUpdatedPayload,
  MatchStartingPayload,
  MatchStartedPayload,
  MatchCompletedPayload,
  PlayerDisconnectedPayload,
  PlayerReconnectedPayload,
} from '../types';

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
  matchStarting: boolean; // Countdown active
  countdownSeconds: number | null; // 3, 2, 1, null
  gameUrl: string | null; // Set when match starts
  disconnectedPlayerId: string | null; // ID of disconnected player
  disconnectedAt: string | null; // Timestamp of disconnect
  canClaimWalkover: boolean; // 2 minutes elapsed, opponent not present
  walkoverClaimableAt: string | null; // Timestamp when walkover becomes available
}

interface UseMatchLobbyReturn {
  matchState: MatchLobbyState | null;
  isLoading: boolean;
  error: string | null;
  connectionStatus: ConnectionState;
  claimWalkover: () => Promise<void>;
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
  const { user } = useAuth();
  const [matchState, setMatchState] = useState<MatchLobbyState | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Track if initial data is loaded
  const initialLoadComplete = useRef(false);

  // Construct WebSocket URL
  // Note: The actual WS base URL would be configured in environment
  const wsUrl = matchId ? `/v1/matches/${matchId}/lobby` : '';

  // Track previous connection status for auto-refresh on reconnect
  const previousStatusRef = useRef<ConnectionState | null>(null);

  // Initialize WebSocket connection
  const { status: connectionStatus } = useWebSocket({
    url: wsUrl,
    onMessage: handleWebSocketMessage,
    onOpen: () => {
      console.log('[useMatchLobby] WebSocket connected');
      
      // Auto-refresh data on reconnection (not on initial connection)
      if (previousStatusRef.current && previousStatusRef.current !== 'connected' && matchId) {
        console.log('[useMatchLobby] Auto-refreshing match data after reconnection');
        
        // Re-fetch match data
        matchmakingApi.getMatch(matchId).then((matchData) => {
          setMatchState({
            match: matchData.match,
            player1: matchData.participant1,
            player2: matchData.participant2,
            roundNumber: matchData.roundNumber || 1,
            tournament: matchData.tournament,
            status: matchData.match.status,
            matchStarting: false,
            countdownSeconds: null,
            gameUrl: matchData.match.status === 'started' ? matchData.match.gameApiMatchId : null,
            disconnectedPlayerId: matchData.match.disconnectedPlayerId || null,
            disconnectedAt: matchData.match.disconnectedAt || null,
            canClaimWalkover: false,
            walkoverClaimableAt: null,
          });
        }).catch((err) => {
          console.error('[useMatchLobby] Exception during match data refresh:', err);
        });
      }
      
      previousStatusRef.current = connectionStatus;
    },
    onClose: () => {
      console.log('[useMatchLobby] WebSocket disconnected');
      previousStatusRef.current = connectionStatus;
    },
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
          matchStarting: false,
          countdownSeconds: null,
          gameUrl: null,
          disconnectedPlayerId: matchData.match.disconnectedPlayerId || null,
          disconnectedAt: matchData.match.disconnectedAt || null,
          canClaimWalkover: false,
          walkoverClaimableAt: null,
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
        case 'match_completed':
          handleMatchCompleted(message.payload as MatchCompletedPayload);
          break;
        case 'player_disconnected':
          handlePlayerDisconnected(message.payload as PlayerDisconnectedPayload);
          break;
        case 'player_reconnected':
          handlePlayerReconnected(message.payload as PlayerReconnectedPayload);
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
      matchStarting: prev?.matchStarting || false,
      countdownSeconds: prev?.countdownSeconds || null,
      gameUrl: prev?.gameUrl || null,
      disconnectedPlayerId: payload.match.disconnectedPlayerId || null,
      disconnectedAt: payload.match.disconnectedAt || null,
      canClaimWalkover: prev?.canClaimWalkover || false,
      walkoverClaimableAt: prev?.walkoverClaimableAt || null,
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
        status: 'started',
        matchStarting: true,
        countdownSeconds: payload.countdownSeconds,
      };
    });
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
        matchStarting: false,
        countdownSeconds: null,
        gameUrl: payload.gameUrl,
      };
    });
  }, []);

  // Handle match_completed message (match finishes)
  const handleMatchCompleted = useCallback((payload: MatchCompletedPayload) => {
    console.log('[useMatchLobby] Match completed:', payload);
    
    setMatchState((prev) => {
      if (!prev) return null;

      const updatedMatch = {
        ...prev.match,
        status: 'completed' as const,
        winnerId: payload.winnerId,
        scorePlayer1: payload.scorePlayer1,
        scorePlayer2: payload.scorePlayer2,
        resultSource: payload.resultSource,
        completedAt: new Date().toISOString(),
      };

      return {
        ...prev,
        match: updatedMatch,
        status: 'completed',
      };
    });
  }, []);

  // Handle player_disconnected message
  const handlePlayerDisconnected = useCallback((payload: PlayerDisconnectedPayload) => {
    console.log('[useMatchLobby] Player disconnected:', payload);
    
    setMatchState((prev) => {
      if (!prev) return null;

      const updatedMatch = {
        ...prev.match,
        disconnectedPlayerId: payload.playerId,
        disconnectedAt: payload.disconnectedAt,
      };

      return {
        ...prev,
        match: updatedMatch,
        disconnectedPlayerId: payload.playerId,
        disconnectedAt: payload.disconnectedAt,
      };
    });
  }, []);

  // Handle player_reconnected message
  const handlePlayerReconnected = useCallback((payload: PlayerReconnectedPayload) => {
    console.log('[useMatchLobby] Player reconnected:', payload);
    
    setMatchState((prev) => {
      if (!prev) return null;

      const updatedMatch = {
        ...prev.match,
        disconnectedPlayerId: null,
        disconnectedAt: null,
      };

      return {
        ...prev,
        match: updatedMatch,
        disconnectedPlayerId: null,
        disconnectedAt: null,
      };
    });
  }, []);

  // Walkover timer - enable claim button after 2 minutes if opponent doesn't join
  useEffect(() => {
    if (!matchState || matchState.status !== 'ready') {
      return;
    }

    // Check if opponent is missing (only 1 player in lobby)
    const opponentMissing = !matchState.player2;
    
    if (!opponentMissing) {
      // Both players present, clear walkover state
      setMatchState((prev) => {
        if (!prev) return null;
        return {
          ...prev,
          canClaimWalkover: false,
          walkoverClaimableAt: null,
        };
      });
      return;
    }

    // Calculate when walkover becomes claimable (2 minutes from match creation)
    const matchCreatedAt = new Date(matchState.match.createdAt).getTime();
    const now = Date.now();
    const elapsed = now - matchCreatedAt;
    const TWO_MINUTES_MS = 2 * 60 * 1000;

    if (elapsed >= TWO_MINUTES_MS) {
      // Walkover already claimable
      setMatchState((prev) => {
        if (!prev) return null;
        return {
          ...prev,
          canClaimWalkover: true,
          walkoverClaimableAt: new Date(matchCreatedAt + TWO_MINUTES_MS).toISOString(),
        };
      });
    } else {
      // Set timer for when walkover becomes claimable
      const remainingMs = TWO_MINUTES_MS - elapsed;
      const timeout = setTimeout(() => {
        setMatchState((prev) => {
          if (!prev) return null;
          return {
            ...prev,
            canClaimWalkover: true,
            walkoverClaimableAt: new Date(matchCreatedAt + TWO_MINUTES_MS).toISOString(),
          };
        });
      }, remainingMs);

      return () => clearTimeout(timeout);
    }
  }, [matchState]);

  // Claim walkover function
  const claimWalkover = useCallback(async () => {
    if (!matchState || !matchId || !user) {
      throw new Error('Match state or user not available');
    }

    try {
      const response = await matchmakingApi.claimWalkover(matchId, user.id);
      
      if (response.error) {
        throw new Error(response.error.message);
      }

      console.log('[useMatchLobby] Walkover claimed successfully - awaiting server confirmation');
      
      // The match_completed event from server will update the state
      // In a real implementation, show a success toast notification here
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Nie udało się zgłosić walkoweru';
      console.error('[useMatchLobby] Walkover claim failed:', errorMessage);
      throw err;
    }
  }, [matchState, matchId, user]);

  return {
    matchState,
    isLoading,
    error,
    connectionStatus,
    claimWalkover,
  };
}

