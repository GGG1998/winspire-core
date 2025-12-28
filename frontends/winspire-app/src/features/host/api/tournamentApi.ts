/**
 * Tournament API Module
 * Feature: 002-streamer-tournament-creation
 * 
 * Handles all tournament-related API operations
 */

import { apiClient } from '../../../shared/api/client'
import { supabase } from '../../../shared/api/supabase'
import type {
  Tournament,
  TournamentFormData,
  CreateTournamentInput,
  EditTournamentInput,
  TournamentApiData,
  CreateTournamentResponse,
  EditTournamentResponse,
  TournamentParticipant,
  Bracket,
  Match,
  MatchWithPlayers,
  MatchPlayer
} from '../types'

/**
 * Get current user's host ID
 * Uses the user's ID as the host ID - backend will auto-create personal host if needed
 */
async function getCurrentHostId(): Promise<string> {
  const { data: { session } } = await supabase.auth.getSession()
  if (!session?.user?.id) {
    throw new Error('User not authenticated')
  }
  // Use user ID as host ID - backend auto-creates personal host on first request
  return session.user.id
}

/**
 * Convert form data to API input format
 */
export function formToApiInput(formData: TournamentFormData): CreateTournamentInput {
  // Parse team size from teamMode (e.g., "1v1" -> 1)
  const teamSize = parseInt(formData.teamMode.split('v')[0], 10) || 1

  // Convert datetime-local string to ISO 8601
  const scheduledStartTimeAt = formData.startTime 
    ? new Date(formData.startTime).toISOString()
    : undefined

  return {
    name: formData.name.trim(),
    competitionParameters: {
      tournamentParameters: {
        teamSize
      }
    },
    scheduledStartTimeAt,
    minimumTeamCount: formData.minPlayers,
    maximumTeamCount: formData.maxPlayers,
    gameSnapshot: formData.gameSnapshot ? {
      id: formData.gameSnapshot.id,
      slug: formData.gameSnapshot.slug,
      name: formData.gameSnapshot.name,
      version: formData.gameSnapshot.version,
      logoUrl: formData.gameSnapshot.logoUrl,
      description: formData.gameSnapshot.description,
      storagePath: formData.gameSnapshot.storagePath
    } : undefined,
    // TODO clean, we use gameSnapshot instead of gameId
    game: formData.game ? {
      id: formData.game
    } : undefined
  }
}

/**
 * Convert API tournament data to UI Tournament model
 */
export function apiToUiTournament(apiData: TournamentApiData): Tournament {
  // Use scheduledStartTimeAt if available, otherwise fall back to createdAt
  const startTime = apiData.scheduledStartTimeAt 
    ? new Date(apiData.scheduledStartTimeAt) 
    : new Date(apiData.createdAt)
  
  return {
    id: apiData.id,
    name: apiData.name,
    status: apiData.status as Tournament['status'],
    startTime,
    // TODO clean 
    game: apiData.game || 'Unknown',
    gameLogoUrl: apiData.gameLogoUrl,
    gameSnapshot: apiData.gameSnapshot ? {
      id: apiData.gameSnapshot.id,
      slug: apiData.gameSnapshot.slug,
      name: apiData.gameSnapshot.name,
      version: apiData.gameSnapshot.version,
      logoUrl: apiData.gameSnapshot.logoUrl,
      description: apiData.gameSnapshot.description,
      storagePath: apiData.gameSnapshot.storagePath
    } : undefined,

    bannerUrl: apiData.bannerUrl,
    creatorId: apiData.creatorId || '',
    roomLink: apiData.roomLink || `${window.location.origin}/h/${apiData.creatorId}/tournaments/${apiData.id}`,
    isCompleted: apiData.isCompleted || false,
    createdAt: new Date(apiData.createdAt),
    updatedAt: apiData.updatedAt ? new Date(apiData.updatedAt) : new Date(apiData.createdAt),
    // Optional fields
    organizer: apiData.organizer,
    format: apiData.format,
    readyWindow: apiData.readyWindow ? {
      startsAt: new Date(apiData.readyWindow.startsAt),
      endsAt: new Date(apiData.readyWindow.endsAt)
    } : undefined,
    prize: apiData.prize,
    me: apiData.me
      ? {
          participationStatus: apiData.me.participationStatus,
          match: apiData.me.match
            ? {
                matchId: apiData.me.match.matchId,
                tournamentId: apiData.me.match.tournamentId,
                status: apiData.me.match.status,
                round: apiData.me.match.round,
                table: apiData.me.match.table,
                startedAt: apiData.me.match.startedAt ? new Date(apiData.me.match.startedAt) : undefined,
                updatedAt: apiData.me.match.updatedAt ? new Date(apiData.me.match.updatedAt) : undefined,
              }
            : undefined,
        }
      : undefined,
    lastActivity: apiData.lastActivity ? new Date(apiData.lastActivity) : undefined,
    participantCount: apiData.participantCount
  }
}

/**
 * Tournament API client
 */
export const tournamentApi = {
  /**
   * Fetch all tournaments for the current host
   */
  async getTournaments(): Promise<Tournament[]> {
    try {
      const hostId = await getCurrentHostId()
      const response = await apiClient.get<{ tournaments: TournamentApiData[] }>(
        `/v1/${hostId}/tournaments`
      )

      if (response.error) {
        throw new Error(response.error.message)
      }

      // Return empty array if no data
      if (!response.data?.tournaments) {
        return []
      }

      return response.data.tournaments.map(apiToUiTournament)
    } catch (error) {
      console.error('Failed to fetch tournaments:', error)
      throw error
    }
  },

  /**
   * Get a single tournament by ID
   */
  async getTournament(tournamentId: string): Promise<Tournament> {
    try {
      const hostId = await getCurrentHostId()
      const response = await apiClient.get<{ tournament: TournamentApiData }>(
        `/v1/${hostId}/tournaments/${tournamentId}`
      )

      if (response.error) {
        // Check if it's a 403 Forbidden error
        if (response.error.message?.includes('permission') || 
            response.error.message?.includes('Forbidden') ||
            response.error.message?.includes('FORBIDDEN')) {
          throw new Error('403: You do not have permission to view this tournament')
        }
        throw new Error(response.error.message)
      }

      if (!response.data?.tournament) {
        throw new Error('Tournament not found')
      }

      return apiToUiTournament(response.data.tournament)
    } catch (error) {
      console.error('Failed to fetch tournament:', error)
      throw error
    }
  },

  /**
   * Create a new tournament
   */
  async createTournament(input: CreateTournamentInput): Promise<Tournament> {
    try {
      const hostId = await getCurrentHostId()
      const response = await apiClient.post<CreateTournamentResponse>(
        `/v1/${hostId}/tournaments`,
        input
      )

      if (response.error) {
        throw new Error(response.error.message)
      }

      if (!response.data?.success) {
        throw new Error(response.data?.errorCode || 'Tournament creation failed')
      }

      if (!response.data.tournamentId) {
        throw new Error('No tournament ID returned')
      }

      // Fetch the created tournament to return complete data
      // For now, return a minimal tournament object since the API only returns ID
      const tournament: Tournament = {
        id: response.data.tournamentId,
        name: input.name,
        status: 'draft',
        startTime: input.scheduledStartTimeAt ? new Date(input.scheduledStartTimeAt) : new Date(),
        game: input.game?.slug || 'Unknown',
        creatorId: hostId,
        roomLink: `${window.location.origin}/tournament/${response.data.tournamentId}`,
        isCompleted: false,
        createdAt: new Date(),
        updatedAt: new Date()

        // TODO: add minimum team count, maximum team count
      }

      return tournament
    } catch (error) {
      console.error('Failed to create tournament:', error)
      throw error
    }
  },

  /**
   * Edit an existing tournament
   */
  async editTournament(tournamentId: string, input: EditTournamentInput): Promise<void> {
    try {
      const hostId = await getCurrentHostId()
      const response = await apiClient.put<EditTournamentResponse>(
        `/v1/${hostId}/tournaments/${tournamentId}`,
        input
      )

      if (response.error) {
        throw new Error(response.error.message)
      }

      if (!response.data?.success) {
        throw new Error(response.data?.errorCode || 'Tournament edit failed')
      }
    } catch (error) {
      console.error('Failed to edit tournament:', error)
      throw error
    }
  },

  /**
   * Start a tournament
   * Uses the unified edit endpoint with status transition
   */
  async startTournament(tournamentId: string): Promise<void> {
    return this.editTournament(tournamentId, { status: 'started' })
  },

  /**
   * Cancel a tournament
   * Uses the unified edit endpoint with status transition
   */
  async cancelTournament(tournamentId: string): Promise<void> {
    return this.editTournament(tournamentId, { status: 'cancelled' })
  },

  /**
   * Publish a tournament (transition from draft to scheduled)
   * Sets default values for format, ready window, and prize
   */
  async publishTournament(tournamentId: string, scheduledStartTimeAt: Date): Promise<void> {
    const now = new Date()
    
    const input: EditTournamentInput = {
      status: 'open',
      format: {
        type: 'single_elimination',
        teamSize: 1,
        maxSlots: 250
      },
      readyWindow: {
        startsAt: now.toISOString(),
        endsAt: scheduledStartTimeAt.toISOString()
      },
      prize: {
        type: 'custom',
        description: ''
      }
    }
    
    return this.editTournament(tournamentId, input)
  },

  /**
   * Open registration for a tournament (transition from scheduled to registration_open)
   * Allows players to sign up for the tournament
   */
  async openRegistration(tournamentId: string): Promise<void> {
    return this.editTournament(tournamentId, { status: 'registration_open' })
  },

  /**
   * Confirm participation in a tournament
   * User confirms they have a spot/place in the tournament
   */
  async confirmParticipation(tournamentId: string): Promise<void> {
    try {
      const hostId = await getCurrentHostId()
      const response = await apiClient.post<{ success: boolean }>(
        `/v1/${hostId}/tournaments/${tournamentId}/confirm-participation`,
        {}
      )

      if (response.error) {
        throw new Error(response.error.message)
      }

      if (!response.data?.success) {
        throw new Error('Failed to confirm participation')
      }
    } catch (error) {
      console.error('Failed to confirm participation:', error)
      throw error
    }
  },

  /**
   * Join a tournament (requires confirmed participation)
   * Returns the room link for the tournament
   */
  async joinTournament(tournamentId: string): Promise<string> {
    try {
      const hostId = await getCurrentHostId()
      const response = await apiClient.post<{ success: boolean; roomLink?: string }>(
        `/v1/${hostId}/tournaments/${tournamentId}/join`,
        {}
      )

      if (response.error) {
        throw new Error(response.error.message)
      }

      if (!response.data?.success) {
        throw new Error('Failed to join tournament')
      }

      return response.data.roomLink || `/tournament/${tournamentId}/room`
    } catch (error) {
      console.error('Failed to join tournament:', error)
      throw error
    }
  },

  /**
   * Get participants list for a tournament
   * Fetches paginated list of tournament participants
   */
  async getParticipants(tournamentId: string, limit: number = 50, offset: number = 0): Promise<TournamentParticipant[]> {
    try {
      const hostId = await getCurrentHostId()
      const response = await apiClient.get<{
        participants: Array<{
          id: string
          userId: string
          name: string
          avatarUrl?: string
          isReady: boolean
          status: string
          registeredAt: string
          checkedInAt?: string
        }>
        total: number
        limit: number
        offset: number
      }>(`/v1/${hostId}/tournaments/${tournamentId}/participants?limit=${limit}&offset=${offset}`)

      if (response.error) {
        throw new Error(response.error.message)
      }

      if (!response.data?.participants) {
        return []
      }

      // Convert API data to UI format
      return response.data.participants.map(p => ({
        id: p.id,
        name: p.name,
        avatarUrl: p.avatarUrl,
        isReady: p.isReady,
        registeredAt: new Date(p.registeredAt)
      }))
    } catch (error) {
      console.error('Failed to fetch participants:', error)
      throw error
    }
  },

  /**
   * Get tournament bracket with all rounds and matches
   */
  async getBracket(tournamentId: string): Promise<Bracket> {
    try {
      const response = await apiClient.get<{ bracket: Bracket }>(
        `:8088/v1/matchmaking/tournaments/${tournamentId}/bracket`
      )

      if (response.error) {
        throw new Error(response.error.message)
      }

      if (!response.data?.bracket) {
        throw new Error('Bracket not found')
      }

      return response.data.bracket
    } catch (error) {
      console.error('Failed to fetch bracket:', error)
      throw error
    }
  },

  /**
   * Get all matches for a tournament (grouped by round)
   * Returns matches with player details
   */
  async getMatches(tournamentId: string): Promise<MatchWithPlayers[]> {
    try {
      const response = await apiClient.get<{
        matches: Array<{
          match: Match
          participant1: MatchPlayer | null
          participant2: MatchPlayer | null
          roundNumber: number
          roundName: string
        }>
      }>(`:8088/v1/matchmaking/tournaments/${tournamentId}/matches`)

      if (response.error) {
        throw new Error(response.error.message)
      }

      if (!response.data?.matches) {
        return []
      }

      // Transform to MatchWithPlayers format
      return response.data.matches.map(item => ({
        ...item.match,
        player1: item.participant1,
        player2: item.participant2,
      }))
    } catch (error) {
      console.error('Failed to fetch matches:', error)
      throw error
    }
  },

  /**
   * Submit manual result for a match (host only)
   * Used when game API fails or is unavailable
   */
  async submitManualResult(
    matchId: string,
    winnerId: string,
    scorePlayer1: number | null = null,
    scorePlayer2: number | null = null
  ): Promise<void> {
    try {
      const response = await apiClient.post<{ success: boolean }>(
        `/v1/matchmaking/matches/${matchId}/manual-result`,
        {
          winnerId,
          scorePlayer1,
          scorePlayer2,
        }
      )

      if (response.error) {
        throw new Error(response.error.message)
      }

      if (!response.data?.success) {
        throw new Error('Failed to submit manual result')
      }
    } catch (error) {
      console.error('Failed to submit manual result:', error)
      throw error
    }
  }
}
