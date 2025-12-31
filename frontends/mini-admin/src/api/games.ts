import { apiClient } from './client'

export interface Game {
  id: string
  gameIntegrationId?: string
  slug: string
  name: string
  description?: string
  logoUrl?: string
  storagePath: string
  version: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface CreateGameRequest {
  gameIntegrationId?: string
  slug: string
  name: string
  storagePath: string
  description?: string
  logoUrl?: string
  version: string
}

export interface UpdateGameRequest {
  gameIntegrationId?: string
  slug?: string
  name?: string
  description?: string
  logoUrl?: string
  version?: string
  isActive?: boolean
}

export interface UploadFilesResponse {
  uploadedCount: number
  uploadedPaths: string[]
  storageBasePath: string
  message?: string
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
    const response = await apiClient.post<Game>('/admin/games', data)
    return response.data
  },

  // Update a game
  updateGame: async (id: string, data: UpdateGameRequest): Promise<Game> => {
    const response = await apiClient.put<Game>(`/admin/games/${id}`, data)
    return response.data
  },

  // Delete a game
  deleteGame: async (id: string): Promise<void> => {
    await apiClient.delete(`/admin/games/${id}`)
  },

  // Upload files for a game
  uploadFiles: async (gameId: string, files: File[]): Promise<UploadFilesResponse> => {
    const formData = new FormData()
    files.forEach((file) => {
      formData.append('files', file)
    })

    const response = await apiClient.post(`/admin/games/${gameId}/files`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data
  },

  // Upload files for a new game before creation (batch - deprecated)
  uploadNewGameFiles: async (slug: string, version: string, files: File[]): Promise<UploadFilesResponse> => {
    const formData = new FormData()
    formData.append('slug', slug)
    formData.append('version', version)
    files.forEach((file) => {
      formData.append('files', file)
    })

    const response = await apiClient.post('/admin/games/uploads', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data
  },

  // Upload a single file for a new game (async approach)
  uploadSingleFile: async (
    slug: string,
    version: string,
    file: File,
    onProgress?: (progress: number) => void
  ): Promise<{ path: string }> => {
    const formData = new FormData()
    formData.append('slug', slug)
    formData.append('version', version)
    // Send the relative path separately since browsers strip directory paths from filenames
    formData.append('filePath', file.name)
    formData.append('files', file)

    const response = await apiClient.post<UploadFilesResponse>('/admin/games/uploads', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (onProgress && e.total) {
          onProgress(Math.round((e.loaded * 100) / e.total))
        }
      },
    })
    return { path: response.data.uploadedPaths[0] }
  },

  // Get public URL for a game (constructed from game data)
  getGameUrl: async (gameId: string): Promise<{ publicUrl: string; storagePath: string }> => {
    const game = await gamesApi.getGame(gameId)
    const baseUrl = (import.meta as ImportMeta).env.VITE_API_URL || 'http://localhost/v1'
    const publicUrl = `${baseUrl}/g/${game.slug}/bundle/index.html`
    return { publicUrl, storagePath: game.storagePath }
  },
}
