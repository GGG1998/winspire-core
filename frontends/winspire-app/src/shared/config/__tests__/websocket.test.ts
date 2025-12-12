/**
 * WebSocket Configuration Tests
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';

describe('WebSocket Configuration', () => {
  beforeEach(() => {
    // Clear any previous mock
    vi.unstubAllEnvs();
  });

  it('should use environment variable when set', () => {
    // Mock the environment variable
    vi.stubEnv('VITE_MATCHMAKING_WS_URL', 'ws://test.example.com:9000/api/matchmaking');
    
    // Re-import to get fresh value with mocked env
    return import('../websocket').then((module) => {
      expect(module.WS_BASE_URL).toBe('ws://test.example.com:9000/api/matchmaking');
    });
  });

  it('should use default fallback when environment variable is not set', () => {
    // Ensure env var is not set
    vi.stubEnv('VITE_MATCHMAKING_WS_URL', '');
    
    return import('../websocket').then((module) => {
      expect(module.WS_BASE_URL).toBe('ws://localhost:8082/api/matchmaking');
    });
  });

  it('should export WS_BASE_URL as a string', () => {
    return import('../websocket').then((module) => {
      expect(typeof module.WS_BASE_URL).toBe('string');
      expect(module.WS_BASE_URL).toBeTruthy();
    });
  });

  it('should have valid WebSocket URL format', () => {
    return import('../websocket').then((module) => {
      expect(module.WS_BASE_URL).toMatch(/^wss?:\/\/.+/);
    });
  });
});











