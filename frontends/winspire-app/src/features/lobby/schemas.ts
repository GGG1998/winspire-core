import { z } from 'zod';

// ============================================================================
// Pre-Lobby Validation Schemas
// ============================================================================

// Pre-lobby status validation
export const preLobbyStatusSchema = z.enum(['waiting', 'grace_period', 'generating_bracket', 'started']);

// Pre-lobby participant validation
export const preLobbyParticipantSchema = z.object({
  id: z.string().uuid(),
  displayName: z.string().min(1).max(100),
  avatarUrl: z.string().url().nullable(),
  joinedAt: z.string().datetime(),
});

// Activity feed item validation
export const activityFeedItemSchema = z.object({
  id: z.string().uuid(),
  type: z.enum(['participant_joined', 'participant_left', 'tournament_starting', 'grace_period_started']),
  message: z.string(),
  timestamp: z.string().datetime(),
  participantName: z.string().optional(),
});

// Tournament pre-lobby state validation
export const tournamentPreLobbyStateSchema = z.object({
  tournamentId: z.string().uuid(),
  tournamentName: z.string().min(1),
  startTime: z.string().datetime(),
  status: preLobbyStatusSchema,
  participants: z.array(preLobbyParticipantSchema),
  participantCount: z.number().int().nonnegative(),
  minimumParticipants: z.number().int().positive(),
  gracePeriodEndsAt: z.string().datetime().nullable(),
  activityFeed: z.array(activityFeedItemSchema),
});

// Grace period state validation
export const gracePeriodStateSchema = z.object({
  isActive: z.boolean(),
  endsAt: z.string().datetime().nullable(),
  remainingSeconds: z.number().int().nonnegative(),
  participantCountUpdating: z.boolean(),
});

// ============================================================================
// Match Validation Schemas
// ============================================================================

// Match status validation
export const matchStatusSchema = z.enum(['pending', 'ready', 'started', 'paused', 'completed', 'disputed', 'cancelled']);

// Result source validation
export const resultSourceSchema = z.enum(['game_api', 'manual_host', 'walkover']);

// Match from API validation
export const matchSchema = z.object({
  id: z.string().uuid(),
  roundId: z.string().uuid(),
  matchNumber: z.number().int().positive(),
  nextMatchId: z.string().uuid().nullable(),
  participant1Id: z.string().uuid(),
  participant2Id: z.string().uuid().nullable(),
  status: matchStatusSchema,
  participant1Ready: z.boolean(),
  participant2Ready: z.boolean(),
  winnerId: z.string().uuid().nullable(),
  scorePlayer1: z.number().int().min(0).nullable(),
  scorePlayer2: z.number().int().min(0).nullable(),
  resultSource: resultSourceSchema.nullable(),
  disconnectedPlayerId: z.string().uuid().nullable(),
  disconnectedAt: z.string().datetime().nullable(),
  gameApiMatchId: z.string().nullable(),
  createdAt: z.string().datetime(),
  startedAt: z.string().datetime().nullable(),
  completedAt: z.string().datetime().nullable(),
  updatedAt: z.string().datetime(),
});

// ============================================================================
// Player Validation Schemas
// ============================================================================

// Player info validation
export const playerInfoSchema = z.object({
  id: z.string().uuid(),
  displayName: z.string().min(1).max(100),
  avatarUrl: z.string().url().nullable(),
});

// Participant status validation
export const participantStatusSchema = z.enum([
  'registered',
  'confirmed',
  'checked_in',
  'active',
  'eliminated',
  'withdrawn',
]);

// Tournament participant validation
export const participantSchema = z.object({
  id: z.string().uuid(),
  userId: z.string().uuid(),
  tournamentId: z.string().uuid(),
  status: participantStatusSchema,
  hasBye: z.boolean(),
  currentMatchId: z.string().uuid().nullable(),
  player: playerInfoSchema,
});

// ============================================================================
// Bracket Validation Schemas
// ============================================================================

// Round status validation
export const roundStatusSchema = z.enum(['pending', 'in_progress', 'completed']);

// Round validation
export const roundSchema = z.object({
  id: z.string().uuid(),
  roundNumber: z.number().int().positive(),
  roundName: z.string(),
  matchesCount: z.number().int().nonnegative(),
  status: roundStatusSchema,
  matches: z.array(matchSchema),
});

// Bracket validation
export const bracketSchema = z.object({
  id: z.string().uuid(),
  tournamentId: z.string().uuid(),
  totalRounds: z.number().int().positive(),
  totalMatches: z.number().int().positive(),
  byesCount: z.number().int().nonnegative(),
  generatedAt: z.string().datetime(),
  rounds: z.array(roundSchema),
});

// ============================================================================
// Lobby State Validation Schemas
// ============================================================================

// Connection state validation
export const connectionStateSchema = z.enum(['connecting', 'connected', 'disconnected', 'reconnecting', 'error']);

// Lobby state validation
export const lobbyStateSchema = z.object({
  match: matchSchema,
  currentUser: playerInfoSchema,
  opponent: playerInfoSchema.nullable(),
  currentUserReady: z.boolean(),
  opponentReady: z.boolean(),
  opponentConnected: z.boolean(),
  matchStarting: z.boolean(),
  countdownSeconds: z.number().int().nonnegative().nullable(),
  disconnectCountdown: z.number().int().nonnegative().nullable(),
  gameUrl: z.string().url().nullable(),
});

// WebSocket state validation
export const webSocketStateSchema = z.object({
  status: connectionStateSchema,
  lastConnected: z.date().nullable(),
  reconnectAttempts: z.number().int().nonnegative(),
  error: z.string().nullable(),
});

// ============================================================================
// WebSocket Message Validation Schemas
// ============================================================================

// Client message type validation
export const clientMessageTypeSchema = z.enum(['ping', 'join_prelobby', 'ready', 'unready']);

// Server message type validation
export const serverMessageTypeSchema = z.enum([
  'pong',
  'prelobby_state',
  'participant_joined',
  'participant_left',
  'grace_period_started',
  'roster_updated',
  'match_assigned',
  'lobby_state',
  'player_joined',
  'player_left',
  'ready_updated',
  'match_starting',
  'match_started',
  'player_disconnected',
  'player_reconnected',
  'match_completed',
  'match_cancelled',
  'error',
]);

// Base WebSocket message validation
export const wsMessageSchema = z.object({
  type: z.union([clientMessageTypeSchema, serverMessageTypeSchema]),
  payload: z.unknown(),
  timestamp: z.string().datetime(),
  correlationId: z.string().optional(),
});

// ============================================================================
// WebSocket Payload Validation Schemas
// ============================================================================

export const preLobbyStatePayloadSchema = z.object({
  tournament: z.object({
    id: z.string().uuid(),
    name: z.string(),
    startTime: z.string().datetime(),
  }),
  participants: z.array(preLobbyParticipantSchema),
  gracePeriodActive: z.boolean(),
  gracePeriodEndsAt: z.string().datetime().nullable(),
});

export const participantJoinedPayloadSchema = z.object({
  participant: preLobbyParticipantSchema,
  totalCount: z.number().int().positive(),
});

export const participantLeftPayloadSchema = z.object({
  participantId: z.string().uuid(),
  totalCount: z.number().int().nonnegative(),
});

export const gracePeriodStartedPayloadSchema = z.object({
  endsAt: z.string().datetime(),
  durationSeconds: z.number().int().positive(),
});

export const rosterUpdatedPayloadSchema = z.object({
  participantCount: z.number().int().nonnegative(),
});

export const matchAssignedPayloadSchema = z.object({
  matchId: z.string().uuid(),
  opponentName: z.string(),
  roundName: z.string(),
  isBye: z.boolean().optional(),
});

export const lobbyStatePayloadSchema = z.object({
  match: matchSchema,
  participant1: playerInfoSchema,
  participant2: playerInfoSchema.nullable(),
});

export const readyUpdatedPayloadSchema = z.object({
  playerId: z.string().uuid(),
  ready: z.boolean(),
});

export const matchStartingPayloadSchema = z.object({
  countdownSeconds: z.number().int().positive(),
});

export const matchStartedPayloadSchema = z.object({
  gameUrl: z.string().url(),
  gameSessionId: z.string(),
});

export const playerDisconnectedPayloadSchema = z.object({
  playerId: z.string().uuid(),
  disconnectedAt: z.string().datetime(),
  reconnectDeadline: z.string().datetime(),
});

export const matchCompletedPayloadSchema = z.object({
  winnerId: z.string().uuid(),
  scorePlayer1: z.number().int().nonnegative(),
  scorePlayer2: z.number().int().nonnegative(),
  resultSource: resultSourceSchema,
});

// ============================================================================
// API Response Validation Schemas
// ============================================================================

export const apiErrorSchema = z.object({
  code: z.string(),
  message: z.string(),
  details: z.unknown().optional(),
});

export const apiResponseSchema = <T extends z.ZodTypeAny>(dataSchema: T) =>
  z.object({
    data: dataSchema,
    error: apiErrorSchema.optional(),
  });

// Get Bracket Response validation
export const getBracketResponseSchema = z.object({
  id: z.string().uuid(),
  tournament_id: z.string().uuid(),
  total_rounds: z.number().int().positive(),
  total_matches: z.number().int().positive(),
  byes_count: z.number().int().nonnegative(),
  generated_at: z.string().datetime(),
  rounds: z.array(
    z.object({
      id: z.string().uuid(),
      round_number: z.number().int().positive(),
      round_name: z.string(),
      matches_count: z.number().int().nonnegative(),
      status: z.string(),
      matches: z.array(
        z.object({
          id: z.string().uuid(),
          match_number: z.number().int().positive(),
          participant1_id: z.string().uuid(),
          participant2_id: z.string().uuid().nullable(),
          next_match_id: z.string().uuid().nullable(),
          status: z.string(),
          participant1_ready: z.boolean(),
          participant2_ready: z.boolean(),
          winner_id: z.string().uuid().nullable(),
          score_player1: z.number().int().nullable(),
          score_player2: z.number().int().nullable(),
          started_at: z.string().datetime().nullable(),
          completed_at: z.string().datetime().nullable(),
        })
      ),
    })
  ),
});

// Mark Ready Response validation
export const markReadyResponseSchema = z.object({
  message: z.string(),
  match_id: z.string().uuid(),
  player_id: z.string().uuid(),
  ready: z.boolean(),
});

// Claim Walkover Response validation
export const claimWalkoverResponseSchema = z.object({
  message: z.string(),
  match_id: z.string().uuid(),
  winner_id: z.string().uuid(),
  no_show_player: z.string().uuid(),
});

// Get Pre-Lobby State Response validation
export const getPreLobbyStateResponseSchema = z.object({
  tournament_id: z.string().uuid(),
  tournament_name: z.string(),
  start_time: z.string().datetime(),
  status: z.string(),
  participants: z.array(
    z.object({
      id: z.string().uuid(),
      display_name: z.string(),
      avatar_url: z.string().url().nullable(),
      joined_at: z.string().datetime(),
    })
  ),
  participant_count: z.number().int().nonnegative(),
  minimum_participants: z.number().int().positive(),
  grace_period_ends_at: z.string().datetime().nullable(),
  activity_feed: z.array(
    z.object({
      id: z.string().uuid(),
      type: z.string(),
      message: z.string(),
      timestamp: z.string().datetime(),
      participant_name: z.string().optional(),
    })
  ),
});

