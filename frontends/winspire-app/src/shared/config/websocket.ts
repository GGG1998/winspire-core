/**
 * WebSocket Configuration
 * 
 * Centralized WebSocket URL configuration for the application.
 * All WebSocket connections should use these base URLs.
 */

/**
 * Matchmaking WebSocket Base URL
 * 
 * Used for tournament pre-lobby and match lobby WebSocket connections.
 * 
 * Format: ws://host (without path - endpoints include full paths)
 * Example: ws://localhost (local development with Traefik on port 80)
 * 
 * Note: Endpoints already include full paths like /v1/matchmaking/tournaments/...
 * This base URL is concatenated with those paths.
 * 
 * Set via VITE_MATCHMAKING_WS_URL environment variable.
 * Defaults to local development URL through Traefik if not set.
 */
export const WS_BASE_URL = import.meta.env.VITE_MATCHMAKING_WS_URL || 'ws://localhost';

