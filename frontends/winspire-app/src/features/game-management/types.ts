/**
 * Game Management Types
 *
 * Shared types for game-related data across the application
 */

/**
 * Full Game interface
 * Represents a complete game entity with all metadata
 */
export interface Game {
  id: string
  slug: string
  name: string
  version: string
  storagePath: string
  logoUrl?: string
  description?: string
  gameIntegrationId?: string
  isActive?: boolean
  createdAt?: string
  updatedAt?: string
}

/**
 * GameSnapshot
 * Lightweight snapshot of game data stored with brackets/matches
 * Contains only the essential fields needed for gameplay
 */
export type GameSnapshot = Pick<
  Game,
  'id' | 'slug' | 'name' | 'version' | 'logoUrl' | 'description' | 'storagePath'
>
