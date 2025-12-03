/**
 * useTournamentStatus Hook
 * Feature: 002-streamer-tournament-creation
 * 
 * Provides tournament status configuration from backend status
 */

import { TOURNAMENT_STATUS_CONFIG } from '../constants'
import type { TournamentStatus, UseTournamentStatusReturn } from '../types'

/**
 * Hook for getting tournament status configuration
 * Uses backend status directly instead of calculating
 */
export function useTournamentStatus(
  status: TournamentStatus
): UseTournamentStatusReturn {
  // Get configuration for the status
  const config = TOURNAMENT_STATUS_CONFIG[status]

  return {
    status,
    label: config.label,
    colorClasses: {
      bg: config.bgColor,
      text: config.textColor,
      border: config.borderColor
    }
  }
}

/**
 * Get row background color class for tournament status
 */
export function getStatusRowColor(status: TournamentStatus): string {
  return TOURNAMENT_STATUS_CONFIG[status].rowBgColor
}



