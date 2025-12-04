import { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWebSocket } from '../../../shared/hooks/useWebSocket';
import { getPreLobbyState, getPreLobbyWebSocketUrl } from '../api/matchmakingApi';
import type {
  TournamentPreLobbyState,
  PreLobbyParticipant,
  ActivityFeedItem,
  WebSocketMessage,
  PreLobbyStatePayload,
  ParticipantJoinedPayload,
  ParticipantLeftPayload,
  GracePeriodStartedPayload,
  RosterUpdatedPayload,
  MatchAssignedPayload,
  ConnectionState,
} from '../types';
import { MATCH_ASSIGNED_NOTIFICATION_DURATION } from '../constants';

// WebSocket base URL from environment
const WS_BASE_URL = import.meta.env.VITE_MATCHMAKING_WS_URL || 'ws://localhost:8082/api/matchmaking';

interface UseTournamentPreLobbyReturn {
  preLobbyState: TournamentPreLobbyState | null;
  isLoading: boolean;
  error: string | null;
  connectionStatus: ConnectionState;
  isGracePeriodActive: boolean;
  gracePeriodRemaining: number;
  isParticipantCountUpdating: boolean;
}

/**
 * Hook for managing tournament pre-lobby state with WebSocket connection
 * 
 * Handles:
 * - Initial pre-lobby state loading
 * - Real-time participant updates
 * - Grace period tracking
 * - Match assignment with auto-redirect
 * - Activity feed updates
 * - Connection state management
 */
export function useTournamentPreLobby(tournamentId: string | null): UseTournamentPreLobbyReturn {
  const navigate = useNavigate();
  
  // State
  const [preLobbyState, setPreLobbyState] = useState<TournamentPreLobbyState | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isGracePeriodActive, setIsGracePeriodActive] = useState(false);
  const [gracePeriodRemaining, setGracePeriodRemaining] = useState(0);
  const [isParticipantCountUpdating, setIsParticipantCountUpdating] = useState(false);

  // Refs
  const gracePeriodTimerRef = useRef<NodeJS.Timeout | null>(null);
  const participantUpdateTimerRef = useRef<NodeJS.Timeout | null>(null);

  // ========================================================================
  // WebSocket Message Handlers
  // ========================================================================

  /**
   * T031: Handle prelobby_state message (initial state on connect)
   */
  const handlePreLobbyState = useCallback((payload: PreLobbyStatePayload) => {
    console.log('[PreLobby] Received initial state:', payload);
    
    setPreLobbyState({
      tournamentId: payload.tournament.id,
      tournamentName: payload.tournament.name,
      startTime: payload.tournament.startTime,
      status: 'waiting',
      participants: payload.participants,
      participantCount: payload.participants.length,
      minimumParticipants: 2, // Default, should come from backend
      gracePeriodEndsAt: payload.gracePeriodEndsAt,
      activityFeed: [],
    });

    setIsGracePeriodActive(payload.gracePeriodActive);
    setIsLoading(false);
  }, []);

  /**
   * T032: Handle participant_joined message
   */
  const handleParticipantJoined = useCallback((payload: ParticipantJoinedPayload) => {
    console.log('[PreLobby] Participant joined:', payload);
    
    setPreLobbyState((prev) => {
      if (!prev) return prev;

      // Add participant if not already in list
      const exists = prev.participants.some((p) => p.id === payload.participant.id);
      const newParticipants = exists
        ? prev.participants
        : [...prev.participants, payload.participant];

      // Add activity feed item
      const feedItem: ActivityFeedItem = {
        id: crypto.randomUUID(),
        type: 'participant_joined',
        message: `${payload.participant.displayName} dołączył do turnieju`,
        timestamp: new Date().toISOString(),
        participantName: payload.participant.displayName,
      };

      return {
        ...prev,
        participants: newParticipants,
        participantCount: payload.totalCount,
        activityFeed: [...prev.activityFeed, feedItem],
      };
    });

    // Flash participant count indicator during grace period
    if (isGracePeriodActive) {
      setIsParticipantCountUpdating(true);
      if (participantUpdateTimerRef.current) {
        clearTimeout(participantUpdateTimerRef.current);
      }
      participantUpdateTimerRef.current = setTimeout(() => {
        setIsParticipantCountUpdating(false);
      }, 1500);
    }
  }, [isGracePeriodActive]);

  /**
   * T033: Handle participant_left message
   */
  const handleParticipantLeft = useCallback((payload: ParticipantLeftPayload) => {
    console.log('[PreLobby] Participant left:', payload);
    
    setPreLobbyState((prev) => {
      if (!prev) return prev;

      // Remove participant from list
      const participant = prev.participants.find((p) => p.id === payload.participantId);
      const newParticipants = prev.participants.filter((p) => p.id !== payload.participantId);

      // Add activity feed item
      const feedItem: ActivityFeedItem = {
        id: crypto.randomUUID(),
        type: 'participant_left',
        message: `${participant?.displayName || 'Gracz'} opuścił turniej`,
        timestamp: new Date().toISOString(),
        participantName: participant?.displayName,
      };

      return {
        ...prev,
        participants: newParticipants,
        participantCount: payload.totalCount,
        activityFeed: [...prev.activityFeed, feedItem],
      };
    });

    // Flash participant count indicator during grace period
    if (isGracePeriodActive) {
      setIsParticipantCountUpdating(true);
      if (participantUpdateTimerRef.current) {
        clearTimeout(participantUpdateTimerRef.current);
      }
      participantUpdateTimerRef.current = setTimeout(() => {
        setIsParticipantCountUpdating(false);
      }, 1500);
    }
  }, [isGracePeriodActive]);

  /**
   * T034: Handle grace_period_started message
   */
  const handleGracePeriodStarted = useCallback((payload: GracePeriodStartedPayload) => {
    console.log('[PreLobby] Grace period started:', payload);
    
    setPreLobbyState((prev) => {
      if (!prev) return prev;

      // Add activity feed item
      const feedItem: ActivityFeedItem = {
        id: crypto.randomUUID(),
        type: 'grace_period_started',
        message: 'Rozpoczęto okres łaski - spóźnieni gracze mogą jeszcze dołączyć (30s)',
        timestamp: new Date().toISOString(),
      };

      return {
        ...prev,
        status: 'grace_period',
        gracePeriodEndsAt: payload.endsAt,
        activityFeed: [...prev.activityFeed, feedItem],
      };
    });

    setIsGracePeriodActive(true);
    setGracePeriodRemaining(payload.durationSeconds);

    // Start countdown
    if (gracePeriodTimerRef.current) {
      clearInterval(gracePeriodTimerRef.current);
    }
    gracePeriodTimerRef.current = setInterval(() => {
      setGracePeriodRemaining((prev) => {
        const next = prev - 1;
        if (next <= 0) {
          setIsGracePeriodActive(false);
          if (gracePeriodTimerRef.current) {
            clearInterval(gracePeriodTimerRef.current);
          }
          return 0;
        }
        return next;
      });
    }, 1000);
  }, []);

  /**
   * T035: Handle roster_updated message (participant count changes during grace period)
   */
  const handleRosterUpdated = useCallback((payload: RosterUpdatedPayload) => {
    console.log('[PreLobby] Roster updated:', payload);
    
    setPreLobbyState((prev) => {
      if (!prev) return prev;
      return {
        ...prev,
        participantCount: payload.participantCount,
      };
    });

    // Flash participant count indicator
    setIsParticipantCountUpdating(true);
    if (participantUpdateTimerRef.current) {
      clearTimeout(participantUpdateTimerRef.current);
    }
    participantUpdateTimerRef.current = setTimeout(() => {
      setIsParticipantCountUpdating(false);
    }, 1500);
  }, []);

  /**
   * T036: Handle match_assigned message with 2s delay + redirect logic
   */
  const handleMatchAssigned = useCallback(
    (payload: MatchAssignedPayload) => {
      console.log('[PreLobby] Match assigned:', payload);

      // Check if this is a bye
      if (payload.isBye) {
        setPreLobbyState((prev) => {
          if (!prev) return prev;

          const feedItem: ActivityFeedItem = {
            id: crypto.randomUUID(),
            type: 'tournament_starting',
            message: `Masz BYE - automatyczny awans do ${payload.roundName}`,
            timestamp: new Date().toISOString(),
          };

          return {
            ...prev,
            status: 'started',
            activityFeed: [...prev.activityFeed, feedItem],
          };
        });

        // TODO: Handle bye recipient waiting state (US7)
        // For now, just show notification
        return;
      }

      // Normal match assignment
      setPreLobbyState((prev) => {
        if (!prev) return prev;

        const feedItem: ActivityFeedItem = {
          id: crypto.randomUUID(),
          type: 'tournament_starting',
          message: `Przydzielono mecz: ${payload.roundName} vs ${payload.opponentName}`,
          timestamp: new Date().toISOString(),
        };

        return {
          ...prev,
          status: 'started',
          activityFeed: [...prev.activityFeed, feedItem],
        };
      });

      // Show notification for 2 seconds, then redirect
      setTimeout(() => {
        navigate(`/lobby/${tournamentId}/match/${payload.matchId}`);
      }, MATCH_ASSIGNED_NOTIFICATION_DURATION);
    },
    [navigate, tournamentId]
  );

  /**
   * Handle incoming WebSocket messages
   */
  const handleWebSocketMessage = useCallback(
    (event: MessageEvent) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        console.log('[PreLobby] WebSocket message:', message.type);

        switch (message.type) {
          case 'prelobby_state':
            handlePreLobbyState(message.payload as PreLobbyStatePayload);
            break;
          case 'participant_joined':
            handleParticipantJoined(message.payload as ParticipantJoinedPayload);
            break;
          case 'participant_left':
            handleParticipantLeft(message.payload as ParticipantLeftPayload);
            break;
          case 'grace_period_started':
            handleGracePeriodStarted(message.payload as GracePeriodStartedPayload);
            break;
          case 'roster_updated':
            handleRosterUpdated(message.payload as RosterUpdatedPayload);
            break;
          case 'match_assigned':
            handleMatchAssigned(message.payload as MatchAssignedPayload);
            break;
          case 'error':
            console.error('[PreLobby] Server error:', message.payload);
            setError(String(message.payload));
            break;
          default:
            console.warn('[PreLobby] Unknown message type:', message.type);
        }
      } catch (err) {
        console.error('[PreLobby] Failed to parse WebSocket message:', err);
      }
    },
    [
      handlePreLobbyState,
      handleParticipantJoined,
      handleParticipantLeft,
      handleGracePeriodStarted,
      handleRosterUpdated,
      handleMatchAssigned,
    ]
  );

  // ========================================================================
  // WebSocket Connection
  // ========================================================================

  const wsUrl = tournamentId ? getPreLobbyWebSocketUrl(tournamentId, WS_BASE_URL) : null;

  const { status: connectionStatus } = useWebSocket({
    url: wsUrl,
    onMessage: handleWebSocketMessage,
    onOpen: () => console.log('[PreLobby] WebSocket connected'),
    onClose: () => console.log('[PreLobby] WebSocket disconnected'),
    onError: (err) => console.error('[PreLobby] WebSocket error:', err),
    reconnect: true,
  });

  // ========================================================================
  // Initial Data Loading
  // ========================================================================

  useEffect(() => {
    if (!tournamentId) {
      setIsLoading(false);
      return;
    }

    // Load initial state from REST API
    const loadInitialState = async () => {
      setIsLoading(true);
      setError(null);

      const result = await getPreLobbyState(tournamentId);

      if (result.error) {
        setError(result.error.message);
        setIsLoading(false);
        return;
      }

      if (result.data) {
        setPreLobbyState(result.data);
        setIsGracePeriodActive(result.data.gracePeriodEndsAt !== null);
      }

      setIsLoading(false);
    };

    loadInitialState();
  }, [tournamentId]);

  // ========================================================================
  // Cleanup
  // ========================================================================

  useEffect(() => {
    return () => {
      if (gracePeriodTimerRef.current) {
        clearInterval(gracePeriodTimerRef.current);
      }
      if (participantUpdateTimerRef.current) {
        clearTimeout(participantUpdateTimerRef.current);
      }
    };
  }, []);

  // ========================================================================
  // Return API
  // ========================================================================

  return {
    preLobbyState,
    isLoading,
    error,
    connectionStatus,
    isGracePeriodActive,
    gracePeriodRemaining,
    isParticipantCountUpdating,
  };
}

