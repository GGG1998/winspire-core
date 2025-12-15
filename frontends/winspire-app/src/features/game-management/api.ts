/**
 * Game Management API Client
 * 
 * Client for interacting with the game-management service
 */

import { apiClient } from '@/lib/api-client'
import type { Game } from './types'

/**
 * API response wrapper
 */
interface ApiResponse<T> {
  data?: T
  error?: {
    message: string
    code?: string
  }
}

/**
 * Game Management API
 */
export const gameManagementApi = {
  /**
   * Get game by ID
   * Fetches complete game information from the game-management service
   */
  async getGameById(gameId: string): Promise<Game> {
    try {
      const response = await apiClient.get<Game>(
        `/v1/games/${gameId}`
      )

      if (response.error) {
        throw new Error(response.error.message)
      }

      if (!response.data) {
        throw new Error('No game data returned')
      }

      return response.data
    } catch (error) {
      console.error('Failed to fetch game:', error)
      throw error
    }
  }
}









