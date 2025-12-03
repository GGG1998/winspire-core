/**
 * useTournaments Hook
 * Feature: 002-streamer-tournament-creation
 * 
 * Wrapper hook for accessing tournaments context
 * Note: This hook now delegates to TournamentsContext to prevent multiple API calls
 */

import { useTournamentsContext } from '../contexts/TournamentsContext'
import type { UseTournamentsReturn } from '../types'

/**
 * Hook for managing tournaments data and operations
 * 
 * This is a convenience wrapper around useTournamentsContext
 * that maintains backward compatibility with existing components.
 * All state is now managed centrally by TournamentsProvider.
 */
export function useTournaments(): UseTournamentsReturn {
  return useTournamentsContext()
}



