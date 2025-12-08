/**
 * Mock WebSocket for Storybook
 * 
 * Provides a fake WebSocket that can be controlled per story
 */

export interface MockWebSocketMessage {
  type: string;
  payload: any;
  timestamp: string;
}

export class MockWebSocketServer {
  private static instance: MockWebSocketServer;
  private connections: Map<string, MockWebSocket> = new Map();
  private messageQueue: MockWebSocketMessage[] = [];
  private autoConnect: boolean = true;

  static getInstance(): MockWebSocketServer {
    if (!MockWebSocketServer.instance) {
      MockWebSocketServer.instance = new MockWebSocketServer();
    }
    return MockWebSocketServer.instance;
  }

  registerConnection(url: string, ws: MockWebSocket) {
    this.connections.set(url, ws);
  }

  unregisterConnection(url: string) {
    this.connections.delete(url);
  }

  sendMessage(url: string, message: MockWebSocketMessage) {
    const connection = this.connections.get(url);
    if (connection && connection.readyState === WebSocket.OPEN) {
      connection.triggerMessage(message);
    }
  }

  broadcastMessage(message: MockWebSocketMessage) {
    this.connections.forEach((connection) => {
      if (connection.readyState === WebSocket.OPEN) {
        connection.triggerMessage(message);
      }
    });
  }

  queueMessage(message: MockWebSocketMessage) {
    this.messageQueue.push(message);
  }

  flushQueue(url?: string) {
    const messages = [...this.messageQueue];
    this.messageQueue = [];
    
    if (url) {
      messages.forEach((msg) => this.sendMessage(url, msg));
    } else {
      messages.forEach((msg) => this.broadcastMessage(msg));
    }
  }

  setAutoConnect(auto: boolean) {
    this.autoConnect = auto;
  }

  getAutoConnect(): boolean {
    return this.autoConnect;
  }

  reset() {
    this.connections.clear();
    this.messageQueue = [];
    this.autoConnect = true;
  }
}

export class MockWebSocket extends EventTarget implements WebSocket {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;

  readonly CONNECTING = 0;
  readonly OPEN = 1;
  readonly CLOSING = 2;
  readonly CLOSED = 3;

  url: string;
  readyState: number = WebSocket.CONNECTING;
  protocol: string = '';
  extensions: string = '';
  bufferedAmount: number = 0;
  binaryType: BinaryType = 'blob';

  onopen: ((this: WebSocket, ev: Event) => any) | null = null;
  onclose: ((this: WebSocket, ev: CloseEvent) => any) | null = null;
  onmessage: ((this: WebSocket, ev: MessageEvent) => any) | null = null;
  onerror: ((this: WebSocket, ev: Event) => any) | null = null;

  private server = MockWebSocketServer.getInstance();

  constructor(url: string | URL, protocols?: string | string[]) {
    super();
    this.url = url.toString();
    this.protocol = Array.isArray(protocols) ? protocols[0] || '' : protocols || '';

    // Register with server
    this.server.registerConnection(this.url, this);

    // Auto-connect if enabled
    if (this.server.getAutoConnect()) {
      setTimeout(() => {
        this.readyState = WebSocket.OPEN;
        const event = new Event('open');
        if (this.onopen) this.onopen.call(this, event);
        this.dispatchEvent(event);

        // Flush any queued messages
        this.server.flushQueue(this.url);
      }, 100);
    }
  }

  send(data: string | ArrayBufferLike | Blob | ArrayBufferView): void {
    if (this.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket is not open');
    }

    // Parse and log the message
    try {
      const message = typeof data === 'string' ? JSON.parse(data) : data;
      console.log('[MockWebSocket] Sent:', message);
    } catch (e) {
      console.log('[MockWebSocket] Sent (binary):', data);
    }
  }

  close(code?: number, reason?: string): void {
    if (this.readyState === WebSocket.CLOSED || this.readyState === WebSocket.CLOSING) {
      return;
    }

    this.readyState = WebSocket.CLOSING;

    setTimeout(() => {
      this.readyState = WebSocket.CLOSED;
      const event = new CloseEvent('close', {
        code: code || 1000,
        reason: reason || '',
        wasClean: true,
      });
      if (this.onclose) this.onclose.call(this, event);
      this.dispatchEvent(event);

      // Unregister from server
      this.server.unregisterConnection(this.url);
    }, 0);
  }

  // Helper method for triggering messages from server
  triggerMessage(message: MockWebSocketMessage): void {
    if (this.readyState !== WebSocket.OPEN) {
      return;
    }

    const data = JSON.stringify(message);
    const event = new MessageEvent('message', { data });
    
    if (this.onmessage) this.onmessage.call(this, event);
    this.dispatchEvent(event);
  }

  // Helper method for triggering errors
  triggerError(_error?: string): void {
    const event = new Event('error');
    if (this.onerror) this.onerror.call(this, event);
    this.dispatchEvent(event);
  }
}

// Install mock WebSocket globally for Storybook
export function installMockWebSocket() {
  if (typeof window !== 'undefined') {
    // Store original WebSocket
    if (!(window as any).__originalWebSocket) {
      (window as any).__originalWebSocket = window.WebSocket;
    }

    // Replace with mock
    (window as any).WebSocket = MockWebSocket;
    
    console.log('[MockWebSocket] Installed');
  }
}

// Restore original WebSocket
export function uninstallMockWebSocket() {
  if (typeof window !== 'undefined' && (window as any).__originalWebSocket) {
    (window as any).WebSocket = (window as any).__originalWebSocket;
    console.log('[MockWebSocket] Uninstalled');
  }
}

// Get the mock server instance for controlling behavior
export function getMockWebSocketServer(): MockWebSocketServer {
  return MockWebSocketServer.getInstance();
}

