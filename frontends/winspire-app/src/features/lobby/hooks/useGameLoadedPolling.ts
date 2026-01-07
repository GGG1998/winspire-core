import { useState, useEffect, useRef, useCallback } from 'react';
import { matchmakingApi } from '../api/matchmakingApi';
import { GAME_LOADED_POLL_INTERVAL } from '../constants';

interface UseGameLoadedPollingOptions {
  matchId: string | null;
  enabled: boolean;
  onGameLoaded?: () => void;
}

interface UseGameLoadedPollingReturn {
  isGameLoaded: boolean;
  isPolling: boolean;
  participant1GameLoaded: boolean;
  participant2GameLoaded: boolean;
}

/**
 * useGameLoadedPolling Hook
 *
 * Polls the server every 5 seconds to check if the current player's game
 * has loaded. This provides a server-authoritative source for game load
 * status, complementing the postMessage approach from the game iframe.
 *
 * Usage:
 * - Enable polling when match status is 'pending' or 'loading'
 * - Disable polling once game is confirmed loaded
 * - Use the callback to trigger ready overlay display
 */
export function useGameLoadedPolling({
  matchId,
  enabled,
  onGameLoaded,
}: UseGameLoadedPollingOptions): UseGameLoadedPollingReturn {
  const [isGameLoaded, setIsGameLoaded] = useState(false);
  const [isPolling, setIsPolling] = useState(false);
  const [participant1GameLoaded, setParticipant1GameLoaded] = useState(false);
  const [participant2GameLoaded, setParticipant2GameLoaded] = useState(false);

  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const onGameLoadedRef = useRef(onGameLoaded);

  // Keep callback ref up to date
  useEffect(() => {
    onGameLoadedRef.current = onGameLoaded;
  }, [onGameLoaded]);

  // Poll function
  const pollStatus = useCallback(async () => {
    if (!matchId) return;

    try {
      const response = await matchmakingApi.getGameLoadedStatus(matchId);

      if (response.error) {
        console.error('[useGameLoadedPolling] API error:', response.error);
        return;
      }

      const data = response.data!;
      setParticipant1GameLoaded(data.participant1_game_loaded);
      setParticipant2GameLoaded(data.participant2_game_loaded);

      if (data.game_loaded && !isGameLoaded) {
        console.log('[useGameLoadedPolling] Server confirms game loaded');
        setIsGameLoaded(true);
        onGameLoadedRef.current?.();
      }
    } catch (err) {
      console.error('[useGameLoadedPolling] Exception:', err);
    }
  }, [matchId, isGameLoaded]);

  // Start/stop polling based on enabled flag
  useEffect(() => {
    if (!enabled || !matchId || isGameLoaded) {
      // Stop polling
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      setIsPolling(false);
      return;
    }

    // Start polling
    setIsPolling(true);

    // Immediate first poll
    pollStatus();

    // Set up interval
    intervalRef.current = setInterval(pollStatus, GAME_LOADED_POLL_INTERVAL);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [enabled, matchId, isGameLoaded, pollStatus]);

  // Reset state when matchId changes
  useEffect(() => {
    setIsGameLoaded(false);
    setParticipant1GameLoaded(false);
    setParticipant2GameLoaded(false);
  }, [matchId]);

  return {
    isGameLoaded,
    isPolling,
    participant1GameLoaded,
    participant2GameLoaded,
  };
}
