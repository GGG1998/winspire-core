import { apiClient } from '../../../shared/api/client';
import type { ApiResponse } from '../../../shared/types';
import {
  API_ENDPOINTS
} from '../constants';
import type {
  GetPreLobbyStateResponse,
  MatchApiData,
  MarkReadyResponse,
  ClaimWalkoverResponse,
  TournamentPreLobbyState,
  Match,
  PreLobbyParticipant,
  ActivityFeedItem,
} from '../types';

// ============================================================================
// Response Transformers (snake_case API → camelCase frontend)
// ============================================================================

function transformPreLobbyParticipant(apiData: any): PreLobbyParticipant {
  return {
    id: apiData.id,
    displayName: apiData.display_name,
    avatarUrl: apiData.avatar_url,
    joinedAt: apiData.joined_at,
  };
}

function transformActivityFeedItem(apiData: any): ActivityFeedItem {
  return {
    id: apiData.id,
    type: apiData.type,
    message: apiData.message,
    timestamp: apiData.timestamp,
    participantName: apiData.participant_name,
  };
}

function transformMatch(apiData: MatchApiData): Match {
  return {
    id: apiData.id,
    roundId: '', // Not provided in MatchApiData, will be set by caller if needed
    matchNumber: apiData.match_number,
    nextMatchId: apiData.next_match_id,
    participant1Id: apiData.participant1_id,
    participant2Id: apiData.participant2_id,
    status: apiData.status as any,
    participant1Ready: apiData.participant1_ready,
    participant2Ready: apiData.participant2_ready,
    winnerId: apiData.winner_id,
    scorePlayer1: apiData.score_player1,
    scorePlayer2: apiData.score_player2,
    resultSource: null, // Not in MatchApiData
    disconnectedPlayerId: null, // Not in MatchApiData
    disconnectedAt: null, // Not in MatchApiData
    gameApiMatchId: null, // Not in MatchApiData
    createdAt: '', // Not in MatchApiData
    startedAt: apiData.started_at,
    completedAt: apiData.completed_at,
    updatedAt: '', // Not in MatchApiData
  };
}

// ============================================================================
// Tournament Pre-Lobby API
// ============================================================================

/**
 * Get tournament pre-lobby state
 * Returns current pre-lobby information including participants and grace period status
 */
export async function getPreLobbyState(
  tournamentId: string
): Promise<ApiResponse<TournamentPreLobbyState>> {
  const response = await apiClient.get<GetPreLobbyStateResponse>(
    API_ENDPOINTS.GET_PRE_LOBBY_STATE(tournamentId)
  );

  if (response.error) {
    return { error: response.error };
  }

  // Transform snake_case API response to camelCase frontend model
  const transformed: TournamentPreLobbyState = {
    tournamentId: response.data!.tournament_id,
    tournamentName: response.data!.tournament_name,
    creatorId: response.data!.creator_id || '', // Creator ID from backend
    startTime: response.data!.start_time,
    status: response.data!.status as any,
    participants: response.data!.participants.map(transformPreLobbyParticipant),
    participantCount: response.data!.participant_count,
    minimumParticipants: response.data!.minimum_participants,
    gracePeriodEndsAt: response.data!.grace_period_ends_at,
    activityFeed: response.data!.activity_feed.map(transformActivityFeedItem),
  };

  return { data: transformed };
}

// ============================================================================
// Match Lobby API
// ============================================================================

// Response type for getMatch API
export interface GetMatchApiResponse {
  match: MatchApiData;
  participant1: {
    id: string;
    display_name: string;
    avatar_url: string | null;
  };
  participant2: {
    id: string;
    display_name: string;
    avatar_url: string | null;
  } | null;
  round_number: number;
  tournament?: {
    id: string;
    name: string;
    creator_id: string;
  };
}

// Transformed response for frontend
export interface GetMatchResponse {
  match: Match;
  participant1: import('../types').PlayerInfo;
  participant2: import('../types').PlayerInfo | null;
  roundNumber: number;
  tournament?: {
    id: string;
    name: string;
    creatorId: string;
  };
}

/**
 * Get match details by match ID
 * Returns match information including participants and status
 */
export async function getMatch(matchId: string): Promise<GetMatchResponse> {
  const response = await apiClient.get<GetMatchApiResponse>(API_ENDPOINTS.GET_MATCH(matchId));

  if (response.error) {
    throw new Error(typeof response.error === 'string' ? response.error : response.error.message || 'Failed to fetch match');
  }

  const data = response.data!;

  const transformed: GetMatchResponse = {
    match: transformMatch(data.match),
    participant1: {
      id: data.participant1.id,
      displayName: data.participant1.display_name,
      avatarUrl: data.participant1.avatar_url,
    },
    participant2: data.participant2 ? {
      id: data.participant2.id,
      displayName: data.participant2.display_name,
      avatarUrl: data.participant2.avatar_url,
    } : null,
    roundNumber: data.round_number,
    tournament: data.tournament ? {
      id: data.tournament.id,
      name: data.tournament.name,
      creatorId: data.tournament.creator_id,
    } : undefined,
  };

  return transformed;
}

/**
 * Mark player as ready for match
 * Updates ready status for the current player
 */
export async function markReady(
  matchId: string,
  playerId: string,
  ready: boolean = true
): Promise<ApiResponse<MarkReadyResponse>> {
  const response = await apiClient.post<MarkReadyResponse>(API_ENDPOINTS.MARK_READY(matchId), {
    player_id: playerId,
    ready,
  });

  return response;
}

/**
 * Claim walkover win due to opponent no-show
 * Awards win to present player after timeout period
 */
export async function claimWalkover(
  matchId: string,
  playerId: string
): Promise<ApiResponse<ClaimWalkoverResponse>> {
  const response = await apiClient.post<ClaimWalkoverResponse>(API_ENDPOINTS.CLAIM_WALKOVER(matchId), {
    player_id: playerId,
  });

  return response;
}

// ============================================================================
// WebSocket URL Helpers
// ============================================================================

/**
 * Get WebSocket URL for tournament pre-lobby
 * @param tournamentId Tournament UUID
 * @param wsBaseUrl WebSocket base URL (e.g., ws://localhost:8082/api/matchmaking)
 * @returns Full WebSocket URL for pre-lobby connection
 */
export function getPreLobbyWebSocketUrl(tournamentId: string, wsBaseUrl: string): string {
  return `${wsBaseUrl}${API_ENDPOINTS.WS_PRE_LOBBY(tournamentId)}`;
}

/**
 * Get WebSocket URL for match lobby
 * @param matchId Match UUID
 * @param wsBaseUrl WebSocket base URL (e.g., ws://localhost:8082/api/matchmaking)
 * @returns Full WebSocket URL for match lobby connection
 */
export function getMatchLobbyWebSocketUrl(matchId: string, wsBaseUrl: string): string {
  return `${wsBaseUrl}${API_ENDPOINTS.WS_MATCH_LOBBY(matchId)}`;
}

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Check if user is authorized to access pre-lobby
 * (This is a client-side check - server will enforce authorization)
 */
export function canAccessPreLobby(userId: string, participants: PreLobbyParticipant[]): boolean {
  return participants.some((p) => p.id === userId);
}

/**
 * Check if user is authorized to access match lobby
 * (This is a client-side check - server will enforce authorization)
 */
export function canAccessMatchLobby(userId: string, match: Match): boolean {
  return userId === match.participant1Id || userId === match.participant2Id;
}

/**
 * Calculate remaining seconds until tournament start
 */
export function getTimeUntilStart(startTime: string): number {
  const start = new Date(startTime).getTime();
  const now = Date.now();
  const diff = start - now;
  return Math.max(0, Math.floor(diff / 1000));
}

/**
 * Calculate remaining seconds in grace period
 */
export function getGracePeriodRemaining(gracePeriodEndsAt: string | null): number {
  if (!gracePeriodEndsAt) return 0;
  const end = new Date(gracePeriodEndsAt).getTime();
  const now = Date.now();
  const diff = end - now;
  return Math.max(0, Math.floor(diff / 1000));
}

// ============================================================================
// Default Export
// ============================================================================

export const matchmakingApi = {
  getPreLobbyState,
  getMatch,
  markReady,
  claimWalkover,
  getPreLobbyWebSocketUrl,
  getMatchLobbyWebSocketUrl,
  canAccessPreLobby,
  canAccessMatchLobby,
  getTimeUntilStart,
  getGracePeriodRemaining,
};

