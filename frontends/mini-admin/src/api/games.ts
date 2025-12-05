import { apiClient } from './client'

export interface Game {
  id: string
  slug: string
  name: string
  description?: string
  logoUrl?: string
  s3Path: string
  version: string
  versioningEnabled: boolean
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface CreateGameRequest {
  slug: string
  name: string
  description?: string
  logoUrl?: string
  version: string
  versioningEnabled: boolean
}

export interface UpdateGameRequest {
  slug?: string
  name?: string
  description?: string
  logoUrl?: string
  version?: string
  versioningEnabled?: boolean
  isActive?: boolean
}

export const gamesApi = {
  // Get all games
  getAllGames: async (): Promise<Game[]> => {
    const response = await apiClient.get<{ games: Game[]; total: number }>('/games')
    return response.data.games
  },

  // Get a single game
  getGame: async (id: string): Promise<Game> => {
    const response = await apiClient.get<Game>(`/games/${id}`)
    return response.data
  },

  // Create a new game
  createGame: async (data: CreateGameRequest): Promise<Game> => {
    const response = await apiClient.post<Game>('/games', data)
    return response.data
  },

  // Update a game
  updateGame: async (id: string, data: UpdateGameRequest): Promise<Game> => {
    const response = await apiClient.put<Game>(`/games/${id}`, data)
    return response.data
  },

  // Delete a game
  deleteGame: async (id: string): Promise<void> => {
    await apiClient.delete(`/games/${id}`)
  },

  // Upload files for a game
  uploadFiles: async (gameId: string, files: File[]): Promise<{ uploadedCount: number; paths: string[] }> => {
    const formData = new FormData()
    files.forEach((file) => {
      formData.append('files', file)
    })

    const response = await apiClient.post(`/games/${gameId}/files`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data
  },

  // Get public URL for a game
  getGameUrl: async (gameId: string): Promise<{ publicUrl: string; s3Path: string }> => {
    const response = await apiClient.get(`/games/${gameId}/url`)
    return response.data
  },
}

