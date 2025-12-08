/**
 * Game Management API Module
 * 
 * Handles all game-related API operations with the game-management service
 */

import { apiClient } from './client'

/**
 * Game represents a game in the system
 * Matches GameResponse from the backend
 */
export interface Game {
  id: string
  gameIntegrationId?: string
  slug: string
  name: string
  description?: string
  logoUrl?: string
  version: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

/**
 * Response for listing games
 */
export interface GamesListResponse {
  games: Game[]
  total: number
}

/**
 * Game Management API client
 */
export const gameManagementApi = {
  /**
   * Fetch all active games
   * GET /v1/games
   */
  async getGames(): Promise<Game[]> {
    try {
      const response = await apiClient.get<GamesListResponse>('/v1/games')

      if (response.error) {
        throw new Error(response.error.message)
      }

      // Return empty array if no data
      if (!response.data?.games) {
        return []
      }

      return response.data.games
    } catch (error) {
      console.error('Failed to fetch games:', error)
      throw error
    }
  },

  /**
   * Get a single game by ID
   * GET /v1/games/:gameId
   */
  async getGameById(gameId: string): Promise<Game> {
    try {
      const response = await apiClient.get<Game>(`/v1/games/${gameId}`)

      if (response.error) {
        throw new Error(response.error.message)
      }

      if (!response.data) {
        throw new Error('Game not found')
      }

      return response.data
    } catch (error) {
      console.error('Failed to fetch game by ID:', error)
      throw error
    }
  },

  /**
   * Get a single game by slug
   * GET /v1/g/:slug
   */
  async getGameBySlug(slug: string): Promise<Game> {
    try {
      const response = await apiClient.get<Game>(`/v1/g/${slug}`)

      if (response.error) {
        throw new Error(response.error.message)
      }

      if (!response.data) {
        throw new Error('Game not found')
      }

      return response.data
    } catch (error) {
      console.error('Failed to fetch game by slug:', error)
      throw error
    }
  }
}




