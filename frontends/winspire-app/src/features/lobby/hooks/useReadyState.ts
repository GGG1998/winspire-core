import { useState, useCallback } from 'react';
import { matchmakingApi } from '../api/matchmakingApi';
import { toast } from 'react-hot-toast';

interface UseReadyStateReturn {
  isReady: boolean;
  isLoading: boolean;
  toggleReady: () => Promise<void>;
  setReady: (ready: boolean) => void;
}

/**
 * useReadyState Hook
 * 
 * Manages player ready status with optimistic updates
 * Features:
 * - Optimistic UI updates (immediate feedback)
 * - Rollback on API error
 * - Loading state during API call
 * - Toast notifications on success/error
 */
export function useReadyState(
  matchId: string,
  playerId: string,
  initialReady: boolean = false
): UseReadyStateReturn {
  const [isReady, setIsReady] = useState(initialReady);
  const [isLoading, setIsLoading] = useState(false);

  // Sync ready state with server value when it changes
  const setReady = useCallback((ready: boolean) => {
    setIsReady(ready);
  }, []);

  // Toggle ready status with optimistic update
  const toggleReady = useCallback(async () => {
    const newReadyState = !isReady;
    const previousState = isReady;

    // Optimistic update - change UI immediately
    setIsReady(newReadyState);
    setIsLoading(true);

    try {
      // Call API to persist change
      const response = await matchmakingApi.markReady(matchId, playerId, newReadyState);

      if (response.error) {
        // Rollback on error
        setIsReady(previousState);
        toast.error('Nie udało się zaktualizować statusu gotowości');
        console.error('[useReadyState] API error:', response.error);
        return;
      }

      // Success feedback
      if (newReadyState) {
        toast.success('Oznaczono jako gotowy!');
      } else {
        toast.success('Anulowano gotowość');
      }

      console.log('[useReadyState] Ready status updated:', response.data);
    } catch (err) {
      // Rollback on exception
      setIsReady(previousState);
      toast.error('Błąd podczas aktualizacji statusu');
      console.error('[useReadyState] Exception:', err);
    } finally {
      setIsLoading(false);
    }
  }, [matchId, playerId, isReady]);

  return {
    isReady,
    isLoading,
    toggleReady,
    setReady,
  };
}

