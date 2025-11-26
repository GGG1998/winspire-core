import { useState, useEffect } from 'react';
import { apiClient } from '../../../shared/api/client';
import type { Lobby } from '../types';

export function useLobby(tournamentId: string) {
  const [lobby, setLobby] = useState<Lobby | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchLobby = async () => {
      setIsLoading(true);
      const response = await apiClient.get<Lobby>(`/lobby/${tournamentId}`);
      if (response.data) {
        setLobby(response.data);
      } else {
        setError(response.error?.message || 'Failed to load lobby');
      }
      setIsLoading(false);
    };

    fetchLobby();
  }, [tournamentId]);

  return { lobby, isLoading, error };
}



