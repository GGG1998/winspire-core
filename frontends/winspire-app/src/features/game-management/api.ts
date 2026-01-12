/**
 * Game Management API Client
 *
 * Fetches games directly from Supabase
 */

import { supabase } from '../../shared/api/supabase'
import type { Game } from './types'

/**
 * Database row type from Supabase (snake_case)
 */
interface GameRow {
  id: string
  game_integration_id: string | null
  slug: string
  name: string
  description: string | null
  logo_url: string | null
  storage_path: string
  version: string
  is_active: boolean
  created_at: string
  updated_at: string
}

/**
 * Maps database row (snake_case) to Game type (camelCase)
 */
function mapRowToGame(row: GameRow): Game {
  return {
    id: row.id,
    gameIntegrationId: row.game_integration_id ?? undefined,
    slug: row.slug,
    name: row.name,
    description: row.description ?? undefined,
    logoUrl: row.logo_url ?? undefined,
    storagePath: row.storage_path,
    version: row.version,
    isActive: row.is_active,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  }
}

/**
 * Game Management API - Supabase Direct Access
 */
export const gameManagementApi = {
  /**
   * Fetch all active games
   * Queries Supabase games table directly
   */
  async getGames(): Promise<Game[]> {
    const { data, error } = await supabase
      .from('games')
      .select('*')
      .eq('is_active', true)
      .order('name')

    if (error) {
      console.error('Failed to fetch games from Supabase:', error)
      throw new Error(`Failed to fetch games: ${error.message}`)
    }

    if (!data) {
      return []
    }

    return data.map(mapRowToGame)
  },

  /**
   * Get game by ID
   */
  async getGameById(gameId: string): Promise<Game> {
    const { data, error } = await supabase
      .from('games')
      .select('*')
      .eq('id', gameId)
      .single()

    if (error) {
      console.error('Failed to fetch game by ID:', error)
      throw new Error(`Failed to fetch game: ${error.message}`)
    }

    if (!data) {
      throw new Error('Game not found')
    }

    return mapRowToGame(data)
  },

  /**
   * Get game by slug
   */
  async getGameBySlug(slug: string): Promise<Game> {
    const { data, error } = await supabase
      .from('games')
      .select('*')
      .eq('slug', slug)
      .eq('is_active', true)
      .single()

    if (error) {
      console.error('Failed to fetch game by slug:', error)
      throw new Error(`Failed to fetch game: ${error.message}`)
    }

    if (!data) {
      throw new Error('Game not found')
    }

    return mapRowToGame(data)
  },
}
