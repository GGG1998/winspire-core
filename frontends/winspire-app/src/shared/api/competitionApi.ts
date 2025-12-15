/**
 * Competition API Module
 * 
 * Handles tournament-related API operations with the competition service
 */

import { apiClient } from './client';

/**
 * TournamentInfo represents minimal tournament data
 * Used for display and navigation in match lobbies
 */
export interface TournamentInfo {
  id: string;
  name: string;
  hostId: string;
  status: string;
  scheduledStartTimeAt?: string;
  minParticipants?: number;
}

/**
 * Competition API client
 */
export const competitionApi = {
  /**
   * Get tournament information by ID
   * Uses the internal endpoint that doesn't require host authentication
   */
  async getTournamentInfo(tournamentId: string): Promise<TournamentInfo> {
    const response = await apiClient.get<TournamentInfo>(
      `/internal/tournaments/${tournamentId}`
    );
    
    if (response.error) {
      throw new Error(response.error.message || 'Failed to fetch tournament');
    }
    
    if (!response.data) {
      throw new Error('Tournament not found');
    }
    
    return response.data;
  }
};









