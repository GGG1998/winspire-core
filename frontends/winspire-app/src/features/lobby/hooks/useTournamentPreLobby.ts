import { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useWebSocket } from '../../../shared/hooks/useWebSocket';
import { getPreLobbyState, getPreLobbyWebSocketUrl } from '../api/matchmakingApi';
import { WS_BASE_URL } from '../../../shared/config/websocket';
import { supabase } from '../../../shared/api/supabase';
import type {
  TournamentPreLobbyState,
  PreLobbyParticipant,
  ActivityFeedItem,
  WebSocketMessage,
  PreLobbyStatePayload,
  ParticipantJoinedPayload,
  ParticipantLeftPayload,
  GracePeriodStartedPayload,
  GracePeriodEndedPayload,
  RosterUpdatedPayload,
  MatchAssignedPayload,
  ConnectionState,
  TournamentCancelledPayload,
} from '../types';
import { MATCH_ASSIGNED_NOTIFICATION_DURATION } from '../constants';

interface UseTournamentPreLobbyReturn {
  preLobbyState: TournamentPreLobbyState | null;
  isLoading: boolean;
  error: string | null;
  connectionStatus: ConnectionState;
  isGracePeriodActive: boolean;
  gracePeriodRemaining: number;
  isParticipantCountUpdating: boolean;
  hasBye: boolean;
  byeInfo: {
    roundName: string;
    nextMatchSlot: string;
  } | null;
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
  const [hasBye, setHasBye] = useState(false);
  const [byeInfo, setByeInfo] = useState<{ roundName: string; nextMatchSlot: string } | null>(null);

  // Refs
  const gracePeriodTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const participantUpdateTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // ========================================================================
  // WebSocket Message Handlers
  // ========================================================================

  /**
   * T031: Handle prelobby_state message (initial state on connect)
   */
  const handlePreLobbyState = useCallback((payload: PreLobbyStatePayload) => {
    // Transform participants from backend snake_case to frontend camelCase
    const participants: PreLobbyParticipant[] = payload.participants.map((p) => ({
      id: p.user_id,
      displayName: p.display_name,
      avatarUrl: p.avatar_url || null,
      joinedAt: p.joined_at,
    }));

    // Transform activity feed from backend format
    const activityFeed: ActivityFeedItem[] = (payload.activity_feed || []).map((item) => ({
      id: item.id,
      type: item.type as ActivityFeedItem['type'],
      message: item.message,
      timestamp: item.timestamp,
      participantName: item.participant_name,
    }));

    setPreLobbyState({
      tournamentId: payload.tournament_id,
      tournamentName: payload.tournament_name,
      creatorId: '', // Will be set from REST API
      startTime: payload.start_time,
      status: payload.status as TournamentPreLobbyState['status'],
      participants,
      participantCount: payload.participant_count,
      minimumParticipants: payload.min_participants,
      gracePeriodEndsAt: payload.grace_period?.ends_at || null,
      activityFeed,
    });

    setIsGracePeriodActive(payload.grace_period?.is_active || false);
    setIsLoading(false);
  }, []);

  /**
   * T032: Handle participant_joined message
   */
  const handleParticipantJoined = useCallback((payload: ParticipantJoinedPayload) => {
    setPreLobbyState((prev) => {
      if (!prev) return prev;

      // Transform backend payload to frontend participant format
      const participant: PreLobbyParticipant = {
        id: payload.user_id,
        displayName: payload.display_name,
        avatarUrl: payload.avatar_url || null,
        joinedAt: payload.joined_at,
      };

      // Add participant if not already in list
      const exists = prev.participants.some((p) => p.id === participant.id);
      const newParticipants = exists
        ? prev.participants
        : [...prev.participants, participant];

      // Add activity feed item
      const feedItem: ActivityFeedItem = {
        id: crypto.randomUUID(),
        type: 'participant_joined',
        message: `${participant.displayName} dołączył do turnieju`,
        timestamp: new Date().toISOString(),
        participantName: participant.displayName,
      };

      return {
        ...prev,
        participants: newParticipants,
        participantCount: newParticipants.length,
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
    setPreLobbyState((prev) => {
      if (!prev) return prev;

      // Remove participant from list
      const newParticipants = prev.participants.filter((p) => p.id !== payload.user_id);

      // Add activity feed item
      const feedItem: ActivityFeedItem = {
        id: crypto.randomUUID(),
        type: 'participant_left',
        message: `${payload.display_name} opuścił turniej`,
        timestamp: new Date().toISOString(),
        participantName: payload.display_name,
      };

      return {
        ...prev,
        participants: newParticipants,
        participantCount: newParticipants.length,
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
    const endTime = new Date(payload.end_time);
    const now = Date.now();
    const durationSeconds = Math.max(0, Math.round((endTime.getTime() - now) / 1000));

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
        gracePeriodEndsAt: payload.end_time,
        participantCount: payload.participant_count,
        activityFeed: [...prev.activityFeed, feedItem],
      };
    });

    setIsGracePeriodActive(true);
    setGracePeriodRemaining(durationSeconds);

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
   * Handle grace_period_ended message
   */
  const handleGracePeriodEnded = useCallback((payload: GracePeriodEndedPayload) => {
    // Stop countdown
    if (gracePeriodTimerRef.current) {
      clearInterval(gracePeriodTimerRef.current);
    }
    setIsGracePeriodActive(false);
    setGracePeriodRemaining(0);

    setPreLobbyState((prev) => {
      if (!prev) return prev;

      const status = payload.status === 'cancelled' ? 'cancelled' : 'generating_bracket';
      const message =
        status === 'cancelled'
          ? 'Turniej został anulowany po zakończeniu okresu łaski'
          : 'Zakończono okres łaski - trwa generowanie drabinki';

      const feedItem: ActivityFeedItem = {
        id: crypto.randomUUID(),
        type: 'tournament_starting',
        message,
        timestamp: new Date().toISOString(),
      };

      return {
        ...prev,
        status,
        participantCount: payload.final_participant_count,
        activityFeed: [...prev.activityFeed, feedItem],
      };
    });
  }, []);

  /**
   * T035: Handle roster_updated message (participant count changes during grace period)
   */
  const handleRosterUpdated = useCallback((payload: RosterUpdatedPayload) => {
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
  // Normalize potential snake_case or camelCase payloads (and legacy isBye)
  const normalizeMatchAssigned = (payload: MatchAssignedPayload) => {
    const matchId = payload.match_id ?? payload.matchId ?? '';
    const roundNumber = payload.round_number ?? payload.roundNumber ?? 0;
    const matchNumber = payload.match_number ?? payload.matchNumber ?? 0;

    const opponentName =
      payload.opponent?.display_name ??
      payload.opponentName ??
      'Przeciwnik';

    const opponentUserId =
      payload.opponent?.user_id ??
      (payload.isBye ? '00000000-0000-0000-0000-000000000000' : undefined);

    const isBye =
      payload.isBye === true ||
      !payload.opponent ||
      opponentUserId === '00000000-0000-0000-0000-000000000000';

    return {
      matchId,
      roundNumber,
      matchNumber,
      opponentName,
      isBye,
    };
  };

  const handleMatchAssigned = useCallback(
    (rawPayload: MatchAssignedPayload) => {
      const payload = normalizeMatchAssigned(rawPayload);

      const { matchId, roundNumber, matchNumber, opponentName, isBye } = payload;
      const roundName = `Runda ${roundNumber}`;

      // Check if this is a bye
      if (isBye) {
        setPreLobbyState((prev) => {
          if (!prev) return prev;

          const feedItem: ActivityFeedItem = {
            id: crypto.randomUUID(),
            type: 'tournament_starting',
            message: `Masz BYE - automatyczny awans do ${roundName}`,
            timestamp: new Date().toISOString(),
          };

          return {
            ...prev,
            status: 'started',
            activityFeed: [...prev.activityFeed, feedItem],
          };
        });

        // Set bye state
        setHasBye(true);
        setByeInfo({
          roundName,
          nextMatchSlot: 'Zwycięzca meczu poprzedniej rundy',
        });

        // Don't redirect - show ByeWaitingState instead
        return;
      }

      // Normal match assignment
      setPreLobbyState((prev) => {
        if (!prev) return prev;

        const feedItem: ActivityFeedItem = {
          id: crypto.randomUUID(),
          type: 'tournament_starting',
          message: `Przydzielono mecz: ${roundName} vs ${opponentName} (mecz ${matchNumber})`,
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
        navigate(`/lobby/${tournamentId}/match/${matchId}`);
      }, MATCH_ASSIGNED_NOTIFICATION_DURATION);
    },
    [navigate, tournamentId]
  );

  /**
   * Handle tournament_cancelled message
   */
  const handleTournamentCancelled = useCallback((payload: TournamentCancelledPayload) => {
    // Stop countdown if any
    if (gracePeriodTimerRef.current) {
      clearInterval(gracePeriodTimerRef.current);
    }
    setIsGracePeriodActive(false);
    setGracePeriodRemaining(0);

    setPreLobbyState((prev) => {
      if (!prev) return prev;

      const feedItem: ActivityFeedItem = {
        id: crypto.randomUUID(),
        type: 'tournament_starting',
        message: `Turniej anulowany: ${payload.reason}`,
        timestamp: new Date().toISOString(),
      };

      return {
        ...prev,
        status: 'cancelled',
        activityFeed: [...prev.activityFeed, feedItem],
      };
    });

    setError(payload.reason);
  }, []);

  /**
   * Handle incoming WebSocket messages
   */
  const handleWebSocketMessage = useCallback(
    (event: MessageEvent) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        console.log('[PreLobby] WebSocket message received:', message.type, message.payload);

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
          case 'grace_period_ended':
            handleGracePeriodEnded(message.payload as GracePeriodEndedPayload);
            break;
          case 'roster_updated':
            handleRosterUpdated(message.payload as RosterUpdatedPayload);
            break;
          case 'match_assigned':
            handleMatchAssigned(message.payload as MatchAssignedPayload);
            break;
          case 'tournament_cancelled':
            handleTournamentCancelled(message.payload as TournamentCancelledPayload);
            break;
          case 'error':
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
      handleGracePeriodEnded,
      handleRosterUpdated,
      handleMatchAssigned,
      handleTournamentCancelled,
    ]
  );

  // ========================================================================
  // WebSocket Connection
  // ========================================================================

  const wsUrl = tournamentId ? getPreLobbyWebSocketUrl(tournamentId, WS_BASE_URL) : null;

  // Get authentication token for WebSocket connection
  const getToken = useCallback(async () => {
    const { data: { session } } = await supabase.auth.getSession();
    return session?.access_token || null;
  }, []);

  const {
    status: connectionStatus,
    error: wsError,
    lastCloseEvent,
  } = useWebSocket({
    url: wsUrl,
    getToken,
    onMessage: handleWebSocketMessage,
    reconnect: true,
    heartbeatInterval: 15000,
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
  // Connection Error Handling
  // ========================================================================

  useEffect(() => {
    if (wsError) {
      setError(wsError);
    }
  }, [wsError]);

  useEffect(() => {
    if (!lastCloseEvent) return;
    const { code, reason, wasClean } = lastCloseEvent;
    if (code === 1000 && wasClean) return;

    const details = reason?.trim() || 'brak powodu (1005 – no status)';
    setError(`WebSocket zamknięty (kod ${code}${wasClean ? ', clean' : ''}): ${details}`);
  }, [lastCloseEvent]);

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
    hasBye,
    byeInfo,
  };
}

