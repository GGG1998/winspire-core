/**
 * MockTournamentsProvider for Storybook
 * Provides a mock context without making real API calls
 */

import { createContext, useContext, useState, useCallback } from 'react'
import type { ReactNode } from 'react'
import type { 
  Tournament, 
  CreateTournamentInput, 
  TournamentError,
  UseTournamentsReturn 
} from '../../types'

interface TournamentsContextValue extends UseTournamentsReturn {}

const TournamentsContext = createContext<TournamentsContextValue | null>(null)

/**
 * Mock provider for Storybook - doesn't make API calls
 */
export function MockTournamentsProvider({ children }: { children: ReactNode }) {
  const [tournaments, setTournaments] = useState<Tournament[]>([])
  const [isLoading] = useState(false)
  const [error, setError] = useState<TournamentError | null>(null)

  const createTournament = useCallback(async (data: CreateTournamentInput): Promise<Tournament> => {
    console.log('Mock createTournament called with:', data)
    
    const mockTournament: Tournament = {
      id: `mock-${Date.now()}`,
      name: data.name,
      status: 'draft',
      startTime: data.scheduledStartTimeAt ? new Date(data.scheduledStartTimeAt) : new Date(),
      game: data.game?.slug || 'Unknown',
      creatorId: 'mock-host-id',
      roomLink: `/tournament/mock-${Date.now()}`,
      isCompleted: false,
      createdAt: new Date(),
      updatedAt: new Date()
    }
    
    setTournaments(prev => [mockTournament, ...prev])
    return mockTournament
  }, [])

  const refreshTournaments = useCallback(async () => {
    console.log('Mock refreshTournaments called')
    // Do nothing - no API call
  }, [])

  const clearError = useCallback(() => {
    setError(null)
  }, [])

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
 * Hook to access mock tournaments context
 */
export function useTournamentsContext(): TournamentsContextValue {
  const context = useContext(TournamentsContext)
  if (!context) {
    throw new Error('useTournamentsContext must be used within MockTournamentsProvider')
  }
  return context
}


