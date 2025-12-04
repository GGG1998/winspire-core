import { useState, useEffect, useCallback } from 'react';

interface DisconnectState {
  isDisconnected: boolean;
  disconnectedPlayerId: string | null;
  disconnectedAt: string | null;
  remainingSeconds: number;
}

interface UseDisconnectReturn extends DisconnectState {
  setDisconnected: (playerId: string, disconnectedAt: string) => void;
  setReconnected: () => void;
  clearDisconnect: () => void;
}

const RECONNECT_WINDOW_SECONDS = 30;

/**
 * useDisconnect Hook
 * 
 * Manages disconnect state and countdown timer
 * Features:
 * - Tracks disconnected player
 * - Calculates remaining reconnect time from timestamp
 * - Auto-updates countdown every second
 * - Handles reconnection events
 */
export function useDisconnect(): UseDisconnectReturn {
  const [state, setState] = useState<DisconnectState>({
    isDisconnected: false,
    disconnectedPlayerId: null,
    disconnectedAt: null,
    remainingSeconds: RECONNECT_WINDOW_SECONDS,
  });

  // Set player as disconnected
  const setDisconnected = useCallback((playerId: string, disconnectedAt: string) => {
    console.log('[useDisconnect] Player disconnected:', playerId, disconnectedAt);
    
    // Calculate initial remaining time
    const disconnectTime = new Date(disconnectedAt).getTime();
    const now = Date.now();
    const elapsedSeconds = Math.floor((now - disconnectTime) / 1000);
    const remaining = Math.max(0, RECONNECT_WINDOW_SECONDS - elapsedSeconds);

    setState({
      isDisconnected: true,
      disconnectedPlayerId: playerId,
      disconnectedAt,
      remainingSeconds: remaining,
    });
  }, []);

  // Set player as reconnected
  const setReconnected = useCallback(() => {
    console.log('[useDisconnect] Player reconnected');
    
    setState({
      isDisconnected: false,
      disconnectedPlayerId: null,
      disconnectedAt: null,
      remainingSeconds: RECONNECT_WINDOW_SECONDS,
    });
  }, []);

  // Clear disconnect state
  const clearDisconnect = useCallback(() => {
    setState({
      isDisconnected: false,
      disconnectedPlayerId: null,
      disconnectedAt: null,
      remainingSeconds: RECONNECT_WINDOW_SECONDS,
    });
  }, []);

  // Update countdown timer
  useEffect(() => {
    if (!state.isDisconnected || !state.disconnectedAt) {
      return;
    }

    // Recalculate remaining time every second
    const interval = setInterval(() => {
      const disconnectTime = new Date(state.disconnectedAt!).getTime();
      const now = Date.now();
      const elapsedSeconds = Math.floor((now - disconnectTime) / 1000);
      const remaining = Math.max(0, RECONNECT_WINDOW_SECONDS - elapsedSeconds);

      setState((prev) => ({
        ...prev,
        remainingSeconds: remaining,
      }));

      // Auto-clear when time expires
      if (remaining === 0) {
        console.log('[useDisconnect] Reconnect window expired');
        clearInterval(interval);
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [state.isDisconnected, state.disconnectedAt]);

  return {
    ...state,
    setDisconnected,
    setReconnected,
    clearDisconnect,
  };
}

