import { useState, useEffect } from 'react';
import { apiClient } from '../../../shared/api/client';
import type { ChatMessage } from '../types';

export function useLobbyChat(tournamentId: string) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    const fetchMessages = async () => {
      const response = await apiClient.get<ChatMessage[]>(`/lobby/${tournamentId}/chat`);
      if (response.data) {
        setMessages(response.data);
      }
    };

    fetchMessages();
  }, [tournamentId]);

  const sendMessage = async (message: string) => {
    setIsLoading(true);
    const response = await apiClient.post<ChatMessage>(`/lobby/${tournamentId}/chat`, {
      message,
    });
    if (response.data) {
      setMessages((prev) => [...prev, response.data!]);
    }
    setIsLoading(false);
  };

  return { messages, sendMessage, isLoading };
}






