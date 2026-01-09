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
  // Get configuration for the status with fallback to draft
  const config = TOURNAMENT_STATUS_CONFIG[status] ?? TOURNAMENT_STATUS_CONFIG.draft

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
  const config = TOURNAMENT_STATUS_CONFIG[status] ?? TOURNAMENT_STATUS_CONFIG.draft
  return config.rowBgColor
}



