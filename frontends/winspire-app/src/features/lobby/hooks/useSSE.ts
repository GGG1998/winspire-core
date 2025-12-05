import { useState, useEffect, useRef } from 'react';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

export function useSSE<T>(endpoint: string) {
  const [data, setData] = useState<T | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const eventSourceRef = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!endpoint) {
      setIsLoading(false);
      return;
    }

    const token = localStorage.getItem('auth_token');
    const url = `${API_BASE_URL}${endpoint}${token ? `?token=${token}` : ''}`;
    
    // Note: EventSource doesn't support custom headers, so we pass token as query param
    // For production, consider using a different SSE implementation that supports headers
    const eventSource = new EventSource(url);
    eventSourceRef.current = eventSource;

    eventSource.onmessage = (event) => {
      try {
        const parsed = JSON.parse(event.data) as T;
        setData(parsed);
        setIsLoading(false);
      } catch (err) {
        setError('Failed to parse SSE data');
        setIsLoading(false);
      }
    };

    eventSource.onerror = () => {
      setError('SSE connection error');
      setIsLoading(false);
      eventSource.close();
    };

    return () => {
      eventSource.close();
    };
  }, [endpoint]);

  return { data, isLoading, error };
}

