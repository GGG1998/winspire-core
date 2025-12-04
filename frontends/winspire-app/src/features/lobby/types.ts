// ============================================================================
// Tournament Pre-Lobby Types
// ============================================================================

// Pre-lobby state before bracket generation
export interface TournamentPreLobbyState {
  tournamentId: string;
  tournamentName: string;
  creatorId: string; // Tournament creator/host ID (streamer ID)
  startTime: string; // ISO timestamp
  status: 'waiting' | 'grace_period' | 'generating_bracket' | 'started';
  participants: PreLobbyParticipant[];
  participantCount: number;
  minimumParticipants: number;
  gracePeriodEndsAt: string | null; // ISO timestamp, set when grace period active
  activityFeed: ActivityFeedItem[];
}

// Participant in pre-lobby
export interface PreLobbyParticipant {
  id: string; // User UUID
  displayName: string;
  avatarUrl: string | null;
  joinedAt: string; // ISO timestamp
}

// Activity feed item
export interface ActivityFeedItem {
  id: string;
  type: 'participant_joined' | 'participant_left' | 'tournament_starting' | 'grace_period_started';
  message: string; // Localized message
  timestamp: string; // ISO timestamp
  participantName?: string;
}

// Grace period state
export interface GracePeriodState {
  isActive: boolean;
  endsAt: string | null; // ISO timestamp
  remainingSeconds: number;
  participantCountUpdating: boolean;
}

// ============================================================================
// Core Match Types
// ============================================================================

// Match status as returned from backend
export type MatchStatus =
  | 'pending' // Waiting for participants
  | 'ready' // Both participants assigned, waiting for ready
  | 'started' // Match in progress
  | 'paused' // Match paused (disconnect)
  | 'completed' // Match finished with result
  | 'disputed' // Result under review
  | 'cancelled'; // Match cancelled

// Result source indicator
export type ResultSource =
  | 'game_api' // Automatic from game API
  | 'manual_host' // Host manual entry
  | 'walkover'; // No-show walkover

// Match entity from API
export interface Match {
  id: string; // UUID
  roundId: string; // UUID
  matchNumber: number; // Position in round (1-based)
  nextMatchId: string | null; // Winner advances to this match
  participant1Id: string; // UUID of player 1
  participant2Id: string | null; // UUID of player 2 (null = bye/TBD)
  status: MatchStatus;
  participant1Ready: boolean;
  participant2Ready: boolean;
  winnerId: string | null; // UUID of winner
  scorePlayer1: number | null; // Player 1 score
  scorePlayer2: number | null; // Player 2 score
  resultSource: ResultSource | null;
  disconnectedPlayerId: string | null;
  disconnectedAt: string | null; // ISO timestamp
  gameApiMatchId: string | null; // External game session ID
  createdAt: string; // ISO timestamp
  startedAt: string | null; // ISO timestamp
  completedAt: string | null; // ISO timestamp
  updatedAt: string; // ISO timestamp
}

// ============================================================================
// Bracket Types
// ============================================================================

// Round status
export type RoundStatus =
  | 'pending' // Not yet started
  | 'in_progress' // Some matches active
  | 'completed'; // All matches done

// Round in bracket
export interface Round {
  id: string; // UUID
  roundNumber: number; // 1-based (1 = first round)
  roundName: string; // "Round 1", "Semifinals", "Final"
  matchesCount: number;
  status: RoundStatus;
  matches: Match[];
}

// Bracket structure
export interface Bracket {
  id: string; // UUID
  tournamentId: string; // UUID
  totalRounds: number;
  totalMatches: number;
  byesCount: number;
  generatedAt: string; // ISO timestamp
  rounds: Round[];
}

// ============================================================================
// Player Types
// ============================================================================

// Player display info (for lobby/bracket)
export interface PlayerInfo {
  id: string; // User UUID
  displayName: string;
  avatarUrl: string | null;
}

// Participant status in tournament
export type ParticipantStatus =
  | 'registered'
  | 'confirmed'
  | 'checked_in'
  | 'active'
  | 'eliminated'
  | 'withdrawn';

// Tournament participant
export interface Participant {
  id: string; // Participant UUID
  userId: string; // User UUID
  tournamentId: string;
  status: ParticipantStatus;
  hasBye: boolean;
  currentMatchId: string | null;
  player: PlayerInfo; // Embedded player info
}

// ============================================================================
// Lobby Types
// ============================================================================

// Lobby state (local + server combined)
export interface LobbyState {
  match: Match;
  currentUser: PlayerInfo;
  opponent: PlayerInfo | null; // null if opponent hasn't joined
  currentUserReady: boolean;
  opponentReady: boolean;
  opponentConnected: boolean;
  matchStarting: boolean; // Countdown active
  countdownSeconds: number | null;
  disconnectCountdown: number | null; // 30s reconnect window
  gameUrl: string | null; // Set when match starts
}

// Connection state
export type ConnectionState = 'connecting' | 'connected' | 'disconnected' | 'reconnecting' | 'error';

// WebSocket connection info
export interface WebSocketState {
  status: ConnectionState;
  lastConnected: Date | null;
  reconnectAttempts: number;
  error: string | null;
}

// ============================================================================
// WebSocket Message Types
// ============================================================================

// Outgoing message types (client → server)
export type ClientMessageType =
  | 'ping' // Heartbeat
  | 'join_prelobby' // Join tournament pre-lobby
  | 'ready' // Player ready
  | 'unready'; // Cancel ready (if allowed)

// Incoming message types (server → client)
export type ServerMessageType =
  | 'pong' // Heartbeat response
  | 'prelobby_state' // Initial pre-lobby state on connect
  | 'participant_joined' // Participant joined pre-lobby
  | 'participant_left' // Participant left pre-lobby
  | 'grace_period_started' // Grace period active, roster fluid
  | 'roster_updated' // Participant count changed during grace period
  | 'match_assigned' // Player assigned to match (with match ID)
  | 'lobby_state' // Initial match lobby state on connect
  | 'player_joined' // Opponent joined lobby
  | 'player_left' // Opponent left lobby
  | 'ready_updated' // Ready state changed
  | 'match_starting' // Both ready, countdown starting
  | 'match_started' // Match is now active
  | 'player_disconnected' // Opponent disconnected
  | 'player_reconnected' // Opponent reconnected
  | 'match_completed' // Match has result
  | 'match_cancelled' // Match cancelled
  | 'error'; // Error message

// Base WebSocket message structure
export interface WebSocketMessage<T = unknown> {
  type: ServerMessageType | ClientMessageType;
  payload: T;
  timestamp: string; // ISO timestamp
  correlationId?: string; // For request-response matching
}

// ============================================================================
// WebSocket Payload Types
// ============================================================================

export interface PreLobbyStatePayload {
  tournament: {
    id: string;
    name: string;
    startTime: string;
  };
  participants: PreLobbyParticipant[];
  gracePeriodActive: boolean;
  gracePeriodEndsAt: string | null;
}

export interface ParticipantJoinedPayload {
  participant: PreLobbyParticipant;
  totalCount: number;
}

export interface ParticipantLeftPayload {
  participantId: string;
  totalCount: number;
}

export interface GracePeriodStartedPayload {
  endsAt: string;
  durationSeconds: number;
}

export interface RosterUpdatedPayload {
  participantCount: number;
}

export interface MatchAssignedPayload {
  matchId: string;
  opponentName: string;
  roundName: string;
  isBye?: boolean;
}

export interface LobbyStatePayload {
  match: Match;
  participant1: PlayerInfo;
  participant2: PlayerInfo | null;
}

export interface ReadyUpdatedPayload {
  playerId: string;
  ready: boolean;
}

export interface MatchStartingPayload {
  countdownSeconds: number; // 3, 2, 1...
}

export interface MatchStartedPayload {
  gameUrl: string;
  gameSessionId: string;
}

export interface PlayerDisconnectedPayload {
  playerId: string;
  disconnectedAt: string;
  reconnectDeadline: string; // 30 seconds from disconnect
}

export interface PlayerReconnectedPayload {
  playerId: string;
  reconnectedAt: string;
}

export interface MatchCompletedPayload {
  winnerId: string;
  scorePlayer1: number;
  scorePlayer2: number;
  resultSource: ResultSource;
}

// ============================================================================
// API Response Types
// ============================================================================

// Generic API response wrapper
export interface ApiResponse<T> {
  data: T;
  error?: ApiError;
}

export interface ApiError {
  code: string;
  message: string;
  details?: unknown;
}

// Bracket API response
export interface GetBracketResponse {
  id: string;
  tournament_id: string;
  total_rounds: number;
  total_matches: number;
  byes_count: number;
  generated_at: string;
  rounds: RoundApiData[];
}

export interface RoundApiData {
  id: string;
  round_number: number;
  round_name: string;
  matches_count: number;
  status: string;
  matches: MatchApiData[];
}

export interface MatchApiData {
  id: string;
  match_number: number;
  participant1_id: string;
  participant2_id: string | null;
  next_match_id: string | null;
  status: string;
  participant1_ready: boolean;
  participant2_ready: boolean;
  winner_id: string | null;
  score_player1: number | null;
  score_player2: number | null;
  started_at: string | null;
  completed_at: string | null;
}

// Ready API response
export interface MarkReadyResponse {
  message: string;
  match_id: string;
  player_id: string;
  ready: boolean;
}

// Walkover API response
export interface ClaimWalkoverResponse {
  message: string;
  match_id: string;
  winner_id: string;
  no_show_player: string;
}

// Pre-lobby API response
export interface GetPreLobbyStateResponse {
  tournament_id: string;
  tournament_name: string;
  creator_id: string;
  start_time: string;
  status: string;
  participants: PreLobbyParticipantApiData[];
  participant_count: number;
  minimum_participants: number;
  grace_period_ends_at: string | null;
  activity_feed: ActivityFeedItemApiData[];
}

export interface PreLobbyParticipantApiData {
  id: string;
  display_name: string;
  avatar_url: string | null;
  joined_at: string;
}

export interface ActivityFeedItemApiData {
  id: string;
  type: string;
  message: string;
  timestamp: string;
  participant_name?: string;
}

// ============================================================================
// Legacy Types (for backward compatibility)
// ============================================================================

// Old Lobby interface - keeping for backward compatibility
export interface Lobby {
  tournamentId: string;
  name: string;
  status: 'OPEN' | 'ACCEPTED' | 'PLAYING' | 'COMPLETE';
}

// Old ChatMessage - keeping for backward compatibility
export interface ChatMessage {
  id: string;
  userId: string;
  userName: string;
  message: string;
  timestamp: string;
}
