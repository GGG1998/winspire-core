import { useState, useCallback } from 'react';
import type { ApiResponse } from '../types';

export function useApi<T>() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const execute = useCallback(
    async (apiCall: () => Promise<ApiResponse<T>>): Promise<T | null> => {
      setLoading(true);
      setError(null);

      try {
        const response = await apiCall();
        if (response.error) {
          setError(response.error.message);
          return null;
        }
        return response.data || null;
      } catch (err) {
        const message = err instanceof Error ? err.message : 'An error occurred';
        setError(message);
        return null;
      } finally {
        setLoading(false);
      }
    },
    []
  );

  return { execute, loading, error };
}

