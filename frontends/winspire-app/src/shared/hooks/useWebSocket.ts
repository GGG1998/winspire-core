import { useState, useEffect, useRef, useCallback } from 'react';
import type { ConnectionState } from '../../features/lobby/types';

// ============================================================================
// Types
// ============================================================================

interface UseWebSocketOptions {
  url: string | null; // WebSocket URL (null = don't connect)
  getToken?: () => Promise<string | null>; // Async function to get auth token
  onMessage?: (event: MessageEvent) => void;
  onOpen?: () => void;
  onClose?: () => void;
  onError?: (error: Event) => void;
  reconnect?: boolean; // Enable auto-reconnect (default: true)
  reconnectInterval?: number; // Initial reconnect interval in ms (default: 1000)
  reconnectBackoff?: number[]; // Backoff sequence in ms (default: [1000, 2000, 4000, 8000, 16000, 30000])
  maxReconnectAttempts?: number; // Max reconnect attempts (default: Infinity)
  heartbeatInterval?: number; // Heartbeat interval in ms (default: 15000)
  messageQueueMaxSize?: number; // Max queued messages (default: 100)
}

interface UseWebSocketReturn {
  status: ConnectionState;
  isConnected: boolean;
  reconnectAttempts: number;
  error: string | null;
  lastCloseEvent: CloseEventInfo | null;
  send: (data: string | object) => boolean;
  close: () => void;
  connect: () => void;
}

interface CloseEventInfo {
  code: number;
  reason: string;
  wasClean: boolean;
}

const DEFAULT_RECONNECT_BACKOFF = [1000, 2000, 4000, 8000, 16000, 30000, 30000, 30000];
const DEFAULT_HEARTBEAT_INTERVAL = 15000;
const DEFAULT_MESSAGE_QUEUE_MAX_SIZE = 100;

// ============================================================================
// useWebSocket Hook
// ============================================================================

/**
 * Generic WebSocket hook with automatic reconnection and message queuing
 * 
 * Features:
 * - Automatic reconnection with exponential backoff
 * - Message queuing during disconnection
 * - Heartbeat/ping support
 * - Connection state tracking
 * - Error handling
 * 
 * @example
 * ```tsx
 * const { status, isConnected, send } = useWebSocket({
 *   url: 'ws://localhost:8082/lobby/123',
 *   onMessage: (event) => {
 *     const data = JSON.parse(event.data);
 *     console.log('Received:', data);
 *   },
 * });
 * 
 * // Send a message
 * send({ type: 'ready', payload: {} });
 * ```
 */
export function useWebSocket(options: UseWebSocketOptions): UseWebSocketReturn {
  const {
    url,
    getToken,
    onMessage,
    onOpen,
    onClose,
    onError,
    reconnect = true,
    reconnectBackoff = DEFAULT_RECONNECT_BACKOFF,
    maxReconnectAttempts = Infinity,
    heartbeatInterval = DEFAULT_HEARTBEAT_INTERVAL,
    messageQueueMaxSize = DEFAULT_MESSAGE_QUEUE_MAX_SIZE,
  } = options;

  // State
  const [status, setStatus] = useState<ConnectionState>('disconnected');
  const [reconnectAttempts, setReconnectAttempts] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [lastCloseEvent, setLastCloseEvent] = useState<CloseEventInfo | null>(null);

  // Refs (don't trigger re-renders)
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const heartbeatIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const messageQueueRef = useRef<(string | object)[]>([]);
  const isManualCloseRef = useRef(false);

  // ========================================================================
  // Helper Functions
  // ========================================================================

  /**
   * Get reconnect delay based on attempt count using exponential backoff
   */
  const getReconnectDelay = useCallback(
    (attempt: number): number => {
      const index = Math.min(attempt, reconnectBackoff.length - 1);
      return reconnectBackoff[index] || reconnectBackoff[reconnectBackoff.length - 1];
    },
    [reconnectBackoff]
  );

  /**
   * Clear reconnect timeout
   */
  const clearReconnectTimeout = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
  }, []);

  /**
   * Clear heartbeat interval
   */
  const clearHeartbeatInterval = useCallback(() => {
    if (heartbeatIntervalRef.current) {
      clearInterval(heartbeatIntervalRef.current);
      heartbeatIntervalRef.current = null;
    }
  }, []);

  /**
   * Clear all timers
   */
  const clearTimers = useCallback(() => {
    clearReconnectTimeout();
    clearHeartbeatInterval();
  }, [clearReconnectTimeout, clearHeartbeatInterval]);

  /**
   * Send heartbeat/ping message
   */
  const sendHeartbeat = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      try {
        wsRef.current.send(JSON.stringify({ type: 'ping', payload: {}, timestamp: new Date().toISOString() }));
      } catch (err) {
        console.error('[WebSocket] Heartbeat failed:', err);
      }
    }
  }, []);

  /**
   * Start heartbeat interval
   */
  const startHeartbeat = useCallback(() => {
    clearHeartbeatInterval();
    heartbeatIntervalRef.current = setInterval(sendHeartbeat, heartbeatInterval);
  }, [heartbeatInterval, sendHeartbeat, clearHeartbeatInterval]);

  /**
   * Flush queued messages
   */
  const flushMessageQueue = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN && messageQueueRef.current.length > 0) {
      console.log(`[WebSocket] Flushing ${messageQueueRef.current.length} queued messages`);
      
      messageQueueRef.current.forEach((msg) => {
        try {
          const data = typeof msg === 'string' ? msg : JSON.stringify(msg);
          wsRef.current?.send(data);
        } catch (err) {
          console.error('[WebSocket] Failed to send queued message:', err);
        }
      });

      messageQueueRef.current = [];
    }
  }, []);

  /**
   * Handle WebSocket open event
   */
  const handleOpen = useCallback(() => {
    console.log('[WebSocket] Connection opened');
    setStatus('connected');
    setReconnectAttempts(0);
    setError(null);
    setLastCloseEvent(null);
    clearReconnectTimeout();
    
    startHeartbeat();
    sendHeartbeat();
    flushMessageQueue();
    
    onOpen?.();
  }, [onOpen, startHeartbeat, flushMessageQueue, sendHeartbeat, clearReconnectTimeout]);

  /**
   * Handle WebSocket message event
   */
  const handleMessage = useCallback(
    (event: MessageEvent) => {
      onMessage?.(event);
    },
    [onMessage]
  );

  /**
   * Handle WebSocket error event
   */
  const handleError = useCallback(
    (event: Event) => {
      console.error('[WebSocket] Error:', event);
      setStatus('error');
      setError('WebSocket connection error');
      onError?.(event);
    },
    [onError]
  );

  /**
   * Handle WebSocket close event
   */
  const handleClose = useCallback((event: CloseEvent) => {
    console.log(
      `[WebSocket] Connection closed (code=${event.code}, reason="${event.reason}", clean=${event.wasClean})`
    );
    setLastCloseEvent({
      code: event.code,
      reason: event.reason,
      wasClean: event.wasClean,
    });
    if (!event.wasClean || event.code !== 1000) {
      setError(event.reason || `WebSocket closed with code ${event.code}`);
    }

    clearTimers();
    setStatus('disconnected');
    onClose?.();

    // Attempt reconnection if not manually closed and reconnect is enabled
    if (!isManualCloseRef.current && reconnect && reconnectAttempts < maxReconnectAttempts) {
      const delay = getReconnectDelay(reconnectAttempts);
      console.log(`[WebSocket] Reconnecting in ${delay}ms (attempt ${reconnectAttempts + 1})`);
      
      setStatus('reconnecting');
      setReconnectAttempts((prev) => prev + 1);

      reconnectTimeoutRef.current = setTimeout(() => {
        connect();
      }, delay);
    } else if (reconnectAttempts >= maxReconnectAttempts) {
      console.error('[WebSocket] Max reconnection attempts reached');
      setError('Failed to reconnect after maximum attempts');
    }
  }, [reconnect, reconnectAttempts, maxReconnectAttempts, getReconnectDelay, clearTimers, onClose]);

  /**
   * Connect to WebSocket
   */
  const connect = useCallback(async () => {
    if (!url) {
      console.warn('[WebSocket] No URL provided, skipping connection');
      return;
    }

    if (wsRef.current?.readyState === WebSocket.OPEN || wsRef.current?.readyState === WebSocket.CONNECTING) {
      console.warn('[WebSocket] Already connected or connecting');
      return;
    }

    try {
      // Build WebSocket URL with authentication token if available
      let wsUrl = url;
      if (getToken) {
        const token = await getToken();
        if (token) {
          const separator = url.includes('?') ? '&' : '?';
          wsUrl = `${url}${separator}token=${token}`;
          console.log('[WebSocket] Connecting with authentication token');
        }
      }

      console.log('[WebSocket] Connecting to:', url);
      console.log('[WebSocket-H15] FINAL WS URL:', wsUrl);
      // H15: Log WebSocket connection attempt
      fetch('http://127.0.0.1:7242/ingest/79a701eb-2b46-4fc0-9477-fca1e4d8fd71',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({location:'useWebSocket.ts:288',message:'WebSocket connecting',data:{url,wsUrl,hasToken:!!getToken},timestamp:Date.now(),sessionId:'debug-session',hypothesisId:'H15'})}).catch(()=>{});
      setStatus('connecting');
      isManualCloseRef.current = false;
      setLastCloseEvent(null);

      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = handleOpen;
      ws.onmessage = handleMessage;
      ws.onerror = handleError;
      ws.onclose = handleClose;
    } catch (err) {
      console.error('[WebSocket] Connection failed:', err);
      setStatus('error');
      setError(err instanceof Error ? err.message : 'Failed to connect');
    }
  }, [url, getToken, handleOpen, handleMessage, handleError, handleClose]);

  /**
   * Close WebSocket connection
   */
  const close = useCallback(() => {
    console.log('[WebSocket] Closing connection manually');
    isManualCloseRef.current = true;
    clearTimers();
    
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    
    setStatus('disconnected');
    setReconnectAttempts(0);
    messageQueueRef.current = [];
  }, [clearTimers]);

  /**
   * Send message through WebSocket
   * Queues message if not connected and returns false
   */
  const send = useCallback(
    (data: string | object): boolean => {
      const ws = wsRef.current;

      if (ws?.readyState === WebSocket.OPEN) {
        try {
          const message = typeof data === 'string' ? data : JSON.stringify(data);
          ws.send(message);
          return true;
        } catch (err) {
          console.error('[WebSocket] Send failed:', err);
          return false;
        }
      } else {
        // Queue message for later if not at max size
        if (messageQueueRef.current.length < messageQueueMaxSize) {
          console.log('[WebSocket] Queueing message (not connected)');
          messageQueueRef.current.push(data);
          return false;
        } else {
          console.warn('[WebSocket] Message queue full, dropping message');
          return false;
        }
      }
    },
    [messageQueueMaxSize]
  );

  // ========================================================================
  // Effects
  // ========================================================================

  /**
   * Restart heartbeat when tab becomes visible again to avoid idle disconnects
   */
  useEffect(() => {
    if (typeof document === 'undefined') {
      return;
    }

    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible' && wsRef.current?.readyState === WebSocket.OPEN) {
        sendHeartbeat();
        startHeartbeat();
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, [sendHeartbeat, startHeartbeat]);

  /**
   * Connect on mount if URL provided
   */
  useEffect(() => {
    if (url) {
      connect();
    }

    return () => {
      close();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [url]); // Only reconnect when URL changes

  // ========================================================================
  // Return API
  // ========================================================================

  return {
    status,
    isConnected: status === 'connected',
    reconnectAttempts,
    error,
    lastCloseEvent,
    send,
    close,
    connect,
  };
}

