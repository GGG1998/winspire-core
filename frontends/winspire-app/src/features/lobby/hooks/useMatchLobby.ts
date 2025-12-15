import { useState, useEffect, useRef, useCallback } from 'react';
import { useWebSocket } from '../../../shared/hooks/useWebSocket';
import { matchmakingApi, getMatchLobbyWebSocketUrl } from '../api/matchmakingApi';
import { WS_BASE_URL } from '../../../shared/config/websocket';
import { supabase } from '../../../shared/api/supabase';
import { useAuth } from '../../auth';
import type { GameSnapshot } from '../../game-management';
import type {
  Match,
  PlayerInfo,
  ConnectionState,
  ServerMessageType,
  LobbyStatePayload,
  ReadyUpdatedPayload,
  MatchReadyToLoadPayload,
  GameLoadedPayload,
  MatchStartingPayload,
  MatchStartedPayload,
  MatchCompletedPayload,
  PlayerDisconnectedPayload,
  PlayerReconnectedPayload,
  PostMatchOutcome,
  ReturnToPreLobbyPayload,
  PlayerEliminatedNotifyPayload,
  TournamentChampionPayload,
  MatchTournamentInfo,
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
  gameSnapshot?: GameSnapshot | null;
  tournamentInfo?: MatchTournamentInfo | null; // Tournament info (from match response, no separate API call)
  status: Match['status'];
  matchStarting: boolean; // Countdown active
  countdownSeconds: number | null; // 3, 2, 1, null
  disconnectedPlayerId: string | null; // ID of disconnected player
  disconnectedAt: string | null; // Timestamp of disconnect
  canClaimWalkover: boolean; // 2 minutes elapsed, opponent not present
  walkoverClaimableAt: string | null; // Timestamp when walkover becomes available
  postMatchOutcome: PostMatchOutcome | null; // Post-match navigation (winner/loser/champion)
}

interface UseMatchLobbyReturn {
  matchState: MatchLobbyState | null;
  isLoading: boolean;
  error: string | null;
  connectionStatus: ConnectionState;
  serverRestarting: boolean;
  claimWalkover: () => Promise<void>;
  clearPostMatchOutcome: () => void;
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
  const [serverRestarting, setServerRestarting] = useState(false);
  
  // Track if initial data is loaded
  const initialLoadComplete = useRef(false);

  // Construct WebSocket URL using centralized config
  const wsUrl = matchId ? getMatchLobbyWebSocketUrl(matchId, WS_BASE_URL) : null;

  // Track previous connection status for auto-refresh on reconnect
  const previousStatusRef = useRef<ConnectionState | null>(null);

  // Get authentication token for WebSocket connection
  const getToken = useCallback(async () => {
    const { data: { session } } = await supabase.auth.getSession();
    return session?.access_token || null;
  }, []);

  // Our custom hook handles token auth internally via getToken callback

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
        
        // Initialize match state from REST response using functional updater
        // to preserve any WebSocket updates that arrived while REST call was in-flight
        setMatchState((prev) => {
          const mergedP1 = prev?.match.participant1GameLoaded ?? matchData.match.participant1GameLoaded;
          const mergedP2 = prev?.match.participant2GameLoaded ?? matchData.match.participant2GameLoaded;
          
          const newState = {
            match: {
              ...matchData.match,
              // Preserve game loaded status from WebSocket if already updated
              participant1GameLoaded: mergedP1,
              participant2GameLoaded: mergedP2,
            },
            player1: matchData.participant1,
            player2: matchData.participant2,
            roundNumber: matchData.roundNumber || 1,
            gameSnapshot: matchData.gameSnapshot,
            tournamentInfo: matchData.tournamentInfo || null,
            status: matchData.match.status,
            matchStarting: prev?.matchStarting || false,
            countdownSeconds: prev?.countdownSeconds || null,
            disconnectedPlayerId: matchData.match.disconnectedPlayerId || null,
            disconnectedAt: matchData.match.disconnectedAt || null,
            canClaimWalkover: false,
            walkoverClaimableAt: null,
            postMatchOutcome: prev?.postMatchOutcome || null,
          };
          
          return newState;
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

  // Handle lobby_state message (full state update)
  const handleLobbyState = useCallback((payload: LobbyStatePayload) => {
    console.log('[useMatchLobby] Lobby state update:', payload);
    
    setMatchState((prev) => {
      // Backend sends snake_case, convert to camelCase
      const payloadP1 = (payload.match as any).participant1_game_loaded ?? payload.match.participant1GameLoaded;
      const payloadP2 = (payload.match as any).participant2_game_loaded ?? payload.match.participant2GameLoaded;
      
      const mergedP1 = payloadP1 ?? prev?.match.participant1GameLoaded ?? false;
      const mergedP2 = payloadP2 ?? prev?.match.participant2GameLoaded ?? false;
      
      const newState = {
        match: {
          ...payload.match,
          // Preserve game loaded fields from previous state if not in payload
          participant1GameLoaded: mergedP1,
          participant2GameLoaded: mergedP2,
        },
        // Preserve player info from REST API if not in WebSocket payload
        player1: payload.participant1 ?? prev?.player1 ?? null,
        player2: payload.participant2 ?? prev?.player2 ?? null,
        roundNumber: prev?.roundNumber || 1,
        gameSnapshot: prev?.gameSnapshot,
        status: payload.match.status,
        matchStarting: prev?.matchStarting || false,
        countdownSeconds: prev?.countdownSeconds || null,
        disconnectedPlayerId: payload.match.disconnectedPlayerId || null,
        disconnectedAt: payload.match.disconnectedAt || null,
        canClaimWalkover: prev?.canClaimWalkover || false,
        walkoverClaimableAt: prev?.walkoverClaimableAt || null,
        postMatchOutcome: prev?.postMatchOutcome || null,
      };
      
      return newState;
    });
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
      
      const isP1 = payload.playerId === updatedMatch.participant1Id;
      const isP2 = payload.playerId === updatedMatch.participant2Id;
      
      if (isP1) {
        updatedMatch.participant1Ready = payload.ready;
      } else if (isP2) {
        updatedMatch.participant2Ready = payload.ready;
      }

      return {
        ...prev,
        match: updatedMatch,
      };
    });
  }, []);

  // Handle match_ready_to_load message (both players ready)
  const handleMatchReadyToLoad = useCallback((payload: MatchReadyToLoadPayload) => {
    console.log('[useMatchLobby] Match ready to load game:', payload.gameUrl);
    
    setMatchState((prev) => {
      if (!prev) return null;

      const updatedMatch = {
        ...prev.match,
        status: 'loading' as const,
        gameUrl: payload.gameUrl,
        gameApiMatchId: payload.gameSessionId,
      };

      return {
        ...prev,
        match: updatedMatch,
        status: 'loading',
      };
    });
  }, []);

  // Handle game_loaded message (player loaded game)
  const handleGameLoaded = useCallback((payload: GameLoadedPayload) => {
    console.log('[useMatchLobby] Game loaded:', payload);
    
    setMatchState((prev) => {
      if (!prev) return null;

      const updatedMatch = {
        ...prev.match,
        participant1GameLoaded: payload.participant1GameLoaded,
        participant2GameLoaded: payload.participant2GameLoaded,
      };

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

  // Handle return_to_prelobby message (winner should return to pre-lobby for next round)
  const handleReturnToPreLobby = useCallback((payload: ReturnToPreLobbyPayload) => {
    console.log('[useMatchLobby] Return to pre-lobby:', payload);

    setMatchState((prev) => {
      if (!prev) return null;

      return {
        ...prev,
        postMatchOutcome: {
          type: 'winner',
          payload,
        },
      };
    });
  }, []);

  // Handle player_eliminated_notify message (loser notification)
  const handlePlayerEliminatedNotify = useCallback((payload: PlayerEliminatedNotifyPayload) => {
    console.log('[useMatchLobby] Player eliminated:', payload);

    setMatchState((prev) => {
      if (!prev) return null;

      return {
        ...prev,
        postMatchOutcome: {
          type: 'eliminated',
          payload,
        },
      };
    });
  }, []);

  // Handle tournament_champion message (tournament winner)
  const handleTournamentChampion = useCallback((payload: TournamentChampionPayload) => {
    console.log('[useMatchLobby] Tournament champion:', payload);

    setMatchState((prev) => {
      if (!prev) return null;

      return {
        ...prev,
        postMatchOutcome: {
          type: 'champion',
          payload,
        },
      };
    });
  }, []);

  // WebSocket message handler (memoized to prevent stale closures)
  const handleWebSocketMessage = useCallback((event: MessageEvent) => {
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
        case 'match_ready_to_load':
          handleMatchReadyToLoad(message.payload as MatchReadyToLoadPayload);
          break;
        case 'game_loaded':
          handleGameLoaded(message.payload as GameLoadedPayload);
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
        case 'return_to_prelobby':
          handleReturnToPreLobby(message.payload as ReturnToPreLobbyPayload);
          break;
        case 'player_eliminated_notify':
          handlePlayerEliminatedNotify(message.payload as PlayerEliminatedNotifyPayload);
          break;
        case 'tournament_champion':
          handleTournamentChampion(message.payload as TournamentChampionPayload);
          break;
        case 'server_restarting':
          console.log('[useMatchLobby] Server is restarting, preparing for reconnect...');
          setServerRestarting(true);
          setError(null); // Clear any existing errors
          break;
        default:
          console.warn('[useMatchLobby] Unknown message type:', messageType);
      }
    } catch (err) {
      console.error('[useMatchLobby] Failed to parse WebSocket message:', err);
    }
  }, [
    matchId,
    handleLobbyState,
    handlePlayerJoined,
    handlePlayerLeft,
    handleReadyUpdated,
    handleMatchReadyToLoad,
    handleGameLoaded,
    handleMatchStarting,
    handleMatchStarted,
    handleMatchCompleted,
    handlePlayerDisconnected,
    handlePlayerReconnected,
    handleReturnToPreLobby,
    handlePlayerEliminatedNotify,
    handleTournamentChampion,
  ]);

  // Initialize WebSocket connection using our custom hook
  const { 
    status: wsStatus,
    isConnected: _isConnected, // Keep for future use
  } = useWebSocket({
    url: wsUrl, // Our hook handles token internally via getToken
    getToken: getToken,
    onMessage: handleWebSocketMessage,
    onOpen: () => {
      console.log('[useMatchLobby] WebSocket connection opened');

      // Clear server restarting flag on successful connection
      setServerRestarting(false);
      setError(null);

      // Auto-refresh data on reconnection (not on initial connection)
      if (previousStatusRef.current && previousStatusRef.current !== 'connected' && matchId) {
        console.log('[useMatchLobby] Auto-refreshing match data after reconnection');

        // Re-fetch match data
        matchmakingApi.getMatch(matchId)
          .then((matchData) => {
            // Use functional updater to preserve WebSocket-updated values
            setMatchState((prev) => {
              const mergedP1 = prev?.match.participant1GameLoaded ?? matchData.match.participant1GameLoaded;
              const mergedP2 = prev?.match.participant2GameLoaded ?? matchData.match.participant2GameLoaded;

              return {
                match: {
                  ...matchData.match,
                  participant1GameLoaded: mergedP1,
                  participant2GameLoaded: mergedP2,
                },
                player1: matchData.participant1,
                player2: matchData.participant2,
                roundNumber: matchData.roundNumber || 1,
                gameSnapshot: matchData.gameSnapshot,
                tournamentInfo: matchData.tournamentInfo || prev?.tournamentInfo || null,
                status: matchData.match.status,
                matchStarting: false,
                countdownSeconds: null,
                disconnectedPlayerId: matchData.match.disconnectedPlayerId || null,
                disconnectedAt: matchData.match.disconnectedAt || null,
                canClaimWalkover: false,
                walkoverClaimableAt: null,
                postMatchOutcome: prev?.postMatchOutcome || null,
              };
            });
          })
          .catch((err) => {
            console.error('[useMatchLobby] Exception during match data refresh:', err);
          });
      }

      previousStatusRef.current = 'connected';
    },
    onClose: () => {
      console.log('[useMatchLobby] WebSocket connection closed');
      previousStatusRef.current = 'disconnected';
    },
    onError: (error) => {
      console.error('[useMatchLobby] WebSocket error:', error);
      setError('Błąd połączenia WebSocket');
    },
    reconnect: true,
    maxReconnectAttempts: Infinity,
    reconnectBackoff: [1000, 2000, 4000, 8000, 16000, 30000],
    heartbeatInterval: 15000,
  });

  // Messages are handled via onMessage callback in useWebSocket options
  // wsStatus already has correct ConnectionState type from our custom hook

  // Walkover timer - enable claim button after 2 minutes if opponent doesn't join
  useEffect(() => {
    if (!matchState || matchState.status !== 'ready') {
      return;
    }

    // Check if opponent is missing (only 1 player in lobby)
    const opponentMissing = !matchState.player2;
    
    if (!opponentMissing) {
      // Both players present, clear walkover state
      // Only update if walkover flags are currently set (prevent infinite loop)
      if (matchState.canClaimWalkover || matchState.walkoverClaimableAt) {
        setMatchState((prev) => {
          if (!prev) return null;
          return {
            ...prev,
            canClaimWalkover: false,
            walkoverClaimableAt: null,
          };
        });
      }
      return;
    }

    // Calculate when walkover becomes claimable (2 minutes from match creation)
    const matchCreatedAt = new Date(matchState.match.createdAt).getTime();
    const now = Date.now();
    const elapsed = now - matchCreatedAt;
    const TWO_MINUTES_MS = 2 * 60 * 1000;

    if (elapsed >= TWO_MINUTES_MS) {
      // Walkover already claimable
      // Only update if walkover is not already set (prevent infinite loop)
      if (!matchState.canClaimWalkover) {
        setMatchState((prev) => {
          if (!prev) return null;
          return {
            ...prev,
            canClaimWalkover: true,
            walkoverClaimableAt: new Date(matchCreatedAt + TWO_MINUTES_MS).toISOString(),
          };
        });
      }
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

  // Clear post-match outcome (e.g., when modal is closed)
  const clearPostMatchOutcome = useCallback(() => {
    setMatchState((prev) => {
      if (!prev) return null;
      return {
        ...prev,
        postMatchOutcome: null,
      };
    });
  }, []);

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
    connectionStatus: wsStatus,
    serverRestarting,
    claimWalkover,
    clearPostMatchOutcome,
  };
}

