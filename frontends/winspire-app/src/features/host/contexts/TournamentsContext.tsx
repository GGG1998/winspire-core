/**
 * TournamentsContext
 * Feature: 002-streamer-tournament-creation
 * 
 * Shared context for tournament data to prevent multiple API calls
 * Provides centralized state management for tournaments across components
 */

import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react'
import { tournamentApi } from '../api/tournamentApi'
import { VALIDATION_MESSAGES } from '../constants'
import type { 
  Tournament, 
  CreateTournamentInput, 
  TournamentError,
  UseTournamentsReturn 
} from '../types'

interface TournamentsContextValue extends UseTournamentsReturn {}

const TournamentsContext = createContext<TournamentsContextValue | null>(null)

/**
 * Provider component that manages tournament state
 */
export function TournamentsProvider({ children }: { children: ReactNode }) {
  const [tournaments, setTournaments] = useState<Tournament[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<TournamentError | null>(null)

  /**
   * Fetch all tournaments
   */
  const refreshTournaments = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)
      const data = await tournamentApi.getTournaments()
      // Sort by creation date (newest first)
      const sorted = data.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime())
      setTournaments(sorted)
    } catch (err) {
      const tournamentError = mapToTournamentError(err)
      setError(tournamentError)
    } finally {
      setIsLoading(false)
    }
  }, [])

  /**
   * Create a new tournament
   */
  const createTournament = useCallback(async (data: CreateTournamentInput): Promise<Tournament> => {
    try {
      const tournament = await tournamentApi.createTournament(data)
      // Add to beginning of list (newest first)
      setTournaments(prev => [tournament, ...prev])
      return tournament
    } catch (err) {
      const tournamentError = mapToTournamentError(err)
      setError(tournamentError)
      throw tournamentError
    }
  }, [])

  /**
   * Clear error state
   */
  const clearError = useCallback(() => {
    setError(null)
  }, [])

  // Fetch tournaments on mount - only once for the entire context
  useEffect(() => {
    refreshTournaments()
  }, [refreshTournaments])

  const value: TournamentsContextValue = {
    tournaments,
    isLoading,
    error,
    createTournament,
    refreshTournaments,
    clearError
  }

  return (
    <TournamentsContext.Provider value={value}>
      {children}
    </TournamentsContext.Provider>
  )
}

/**
 * Hook to access tournaments context
 * Must be used within TournamentsProvider
 */
export function useTournamentsContext(): TournamentsContextValue {
  const context = useContext(TournamentsContext)
  if (!context) {
    throw new Error('useTournamentsContext must be used within TournamentsProvider')
  }
  return context
}

/**
 * Map unknown error to TournamentError
 */
function mapToTournamentError(err: unknown): TournamentError {
  // Network error
  if (err instanceof TypeError && err.message.includes('fetch')) {
    return {
      type: 'network',
      message: VALIDATION_MESSAGES.general.network,
      details: err
    }
  }

  // API error with response
  if (err && typeof err === 'object' && 'status' in err) {
    const apiError = err as { status: number; message?: string; validationErrors?: Array<{ field: string; message: string }> }
    
    if (apiError.status === 401) {
      return {
        type: 'authorization',
        message: VALIDATION_MESSAGES.general.authorization,
        details: err
      }
    }
    
    if (apiError.status === 400) {
      return {
        type: 'validation',
        message: VALIDATION_MESSAGES.general.validation,
        details: err,
        fieldErrors: apiError.validationErrors?.reduce((acc, e) => {
          acc[e.field] = e.message
          return acc
        }, {} as Record<string, string>)
      }
    }
    
    if (apiError.status >= 500) {
      return {
        type: 'server',
        message: VALIDATION_MESSAGES.general.server,
        details: err
      }
    }
  }

  // Unknown error
  return {
    type: 'unknown',
    message: VALIDATION_MESSAGES.general.unknown,
    details: err
  }
}

