/**
 * API Configuration
 * 
 * Centralized API URL configuration for the application.
 */

/**
 * Game Management Service URL
 * 
 * Used for accessing game files and static assets.
 * Points to the game-management service storage endpoint.
 * 
 * Format: http://host or https://host (without trailing slash)
 * Example: http://localhost (local development with Traefik on port 80)
 * 
 * Game bundle URL pattern: ${GAME_MANAGEMENT_URL}/v1/g/${slug}/bundle/
 * Where:
 * - First %s: GAME_MANAGEMENT_URL
 * - Second %s: game slug name
 * 
 * Set via VITE_GAME_MANAGEMENT_URL environment variable.
 * Defaults to local development URL through Traefik if not set.
 */
export const GAME_MANAGEMENT_URL = import.meta.env.VITE_GAME_MANAGEMENT_URL || 'http://localhost:8087';

