/**
 * Tournament Feature Types
 * Feature: 002-streamer-tournament-creation
 */

import type { GameSnapshot } from '../../features/game-management'

// ============================================================================
// Core Entities
// ============================================================================

/**
 * Participant/Player in a tournament
 */
export interface TournamentParticipant {
  /** Unique participant identifier */
  id: string
  /** Display name */
  name: string
  /** Avatar URL (optional) */
  avatarUrl?: string
  /** Whether player has confirmed ready status */
  isReady: boolean
  /** Registration timestamp */
  registeredAt: Date
}

/**
 * Tournament format configuration
 */
export interface TournamentFormat {
  /** Format type (Single Elimination, Double Elimination, Round Robin, etc.) */
  type: 'single_elimination' | 'double_elimination' | 'round_robin' | 'swiss'
  /** Team size (1v1, 2v2, 3v3, etc.) */
  teamSize: number
  /** Maximum number of slots/participants */
  maxSlots: number
  /** Best of N for matches */
  bestOf?: number
}

/**
 * Ready window configuration
 */
export interface ReadyWindow {
  /** Window start time */
  startsAt: Date
  /** Window end time */
  endsAt: Date
}

/**
 * Prize configuration
 */
export interface TournamentPrize {
  /** Prize type */
  type: 'custom' | 'cash' | 'points'
  /** Prize description */
  description: string
  /** Prize value (for cash/points) */
  value?: number
  /** Currency (for cash prizes) */
  currency?: string
}

/**
 * Organization/Host info
 */
export interface TournamentOrganizer {
  /** Organizer ID */
  id: string
  /** Organizer name */
  name: string
  /** Logo URL */
  logoUrl?: string
}

/**
 * User-specific tournament information
 */
export interface TournamentMeInfo {
  /** Current user's participation status */
  participationStatus?: 'registered' | 'confirmed' | 'checked_in' | null
}

/**
 * Tournament entity representing a gaming competition
 */
export interface Tournament {
  /** Unique tournament identifier (UUID) */
  id: string
  
  /** Tournament display name (3-100 characters) */
  name: string
  
  /** Tournament status from backend */
  status: TournamentStatus
  
  /** Scheduled start time */
  startTime: Date
  
  /** Game identifier (currently always "Packman") */
  // TODO clean, we use gameSnapshot instead of gameId
  game: string
  
  /** Game logo/icon URL */
  gameLogoUrl?: string

  /** Game snapshot data from game-management service */
  gameSnapshot?: GameSnapshot
  
  /** Tournament banner/cover image URL */
  bannerUrl?: string
  
  /** Streamer ID who created this tournament */
  creatorId: string
  
  /** Organizer information */
  organizer?: TournamentOrganizer
  
  /** Unique room link for participants to join */
  roomLink: string
  
  /** Whether tournament has been completed */
  isCompleted: boolean
  
  /** Tournament format configuration */
  format?: TournamentFormat
  
  /** Ready window for check-in */
  readyWindow?: ReadyWindow
  
  /** Prize configuration */
  prize?: TournamentPrize
  
  /** Last player activity timestamp (for inactivity detection) */
  lastActivity?: Date
  
  /** Tournament creation timestamp */
  createdAt: Date
  
  /** Last update timestamp */
  updatedAt: Date
  
  /** User-specific tournament information */
  me?: TournamentMeInfo
  
  /** Number of confirmed participants */
  participantCount?: number
}

/**
 * Tournament status from backend
 * Maps to the 7 status values in the database
 */
export type TournamentStatus = 'draft' | 'scheduled' | 'registration_open' | 'registration_closed' | 'started' | 'completed' | 'cancelled'

// ============================================================================
// Input Types
// ============================================================================

/**
 * Tournament parameters for creating a tournament
 */
export interface TournamentParametersInput {
  /** Template ID for tournament structure */
  templateId?: string
  
  /** Required number of players per lineup */
  teamSize?: number
}

/**
 * Competition parameters
 */
export interface CompetitionParametersInput {
  /** Tournament parameters */
  tournamentParameters?: TournamentParametersInput
}

/**
 * Game identifier input
 */
export interface GameIdentifierInput {
  /** Game UUID */
  id?: string
  
  /** Human-readable slug for the game */
  slug?: string
}

/**
 * Space identifier input
 */
export interface SpaceIdentifierInput {
  /** Space UUID */
  id?: string
  
  /** Human-readable slug for the space */
  slug?: string
}

/**
 * Data required to create a new tournament (API format)
 */
export interface CreateTournamentInput {
  /** Tournament name (3-100 characters) */
  name: string
  
  /** Description (max 4000 characters) */
  description?: string
  
  /** Caller-provided idempotency key */
  externalId?: string
  
  /** Competition parameters */
  competitionParameters: CompetitionParametersInput
  
  /** Registration window open time (ISO 8601) */
  registrationWindowOpenAt?: string
  
  /** Scheduled start time (ISO 8601) */
  scheduledStartTimeAt?: string
  
  /** Auto force ready (default: true) */
  autoForceReady?: boolean
  
  /** Minimum team count (default: 8) */
  minimumTeamCount?: number

  /** Maximum team count (default: 250) */
  maximumTeamCount?: number

  /** Team size (default: 1) */
  teamSize?: number

  /** Game identifier */
  game?: GameIdentifierInput
  
  /** Game snapshot data from game-management service */
  gameSnapshot?: GameSnapshot
  
  /** Space identifier */
  space?: SpaceIdentifierInput
}

/**
 * Data required to edit a tournament
 * Supports partial updates including status transitions
 */
export interface EditTournamentInput {
  /** Tournament name (3-100 characters) */
  name?: string
  
  /** Description (max 4000 characters) */
  description?: string
  
  /** Tournament status (for state transitions: open, started, completed, cancelled) */
  status?: 'draft' | 'open' | 'started' | 'completed' | 'cancelled'
  
  /** Scheduled start time (ISO 8601) */
  scheduledStartTimeAt?: string
  
  /** Minimum team count (minimum: 1) */
  minimumTeamCount?: number
  
  /** Maximum team count */
  maximumTeamCount?: number
  
  /** Tournament format configuration */
  format?: {
    type: 'single_elimination' | 'double_elimination' | 'round_robin' | 'swiss'
    teamSize: number
    maxSlots: number
    bestOf?: number
  }
  
  /** Ready window for check-in */
  readyWindow?: {
    startsAt: string
    endsAt: string
  }
  
  /** Prize configuration */
  prize?: {
    type: 'custom' | 'cash' | 'points'
    description: string
    value?: number
    currency?: string
  }
}

/**
 * Join tournament request
 */
export interface JoinTournamentInput {
  /** Social team ID */
  socialTeamId?: string
  
  /** Team name */
  teamName?: string
  
  /** Array of player UUIDs */
  players: string[]
}

/**
 * Form data structure for controlled components
 */
export interface TournamentFormData {
  /** Tournament name from text input */
  name: string
  
  /** Start time from datetime-local input (YYYY-MM-DDTHH:mm format) */
  startTime: string
  
  /** Game selection (pre-filled, disabled) */
  // TODO clean, we use gameSnapshot instead of gameId
  game: string

  /** Game snapshot data from game-management service */
  gameSnapshot: GameSnapshot
  
  /** Team mode (e.g., '1v1', '2v2', etc.) */
  teamMode: string
  
  /** Minimum number of players */
  minPlayers: number
  /** Maximum number of players */
  maxPlayers: number
}

// ============================================================================
// API Response Types
// ============================================================================

/**
 * Standard success response
 */
export interface SuccessResponse {
  /** Operation success status */
  success: boolean
  
  /** Error code if operation failed */
  errorCode?: string
}

/**
 * Create tournament response
 */
export interface CreateTournamentResponse extends SuccessResponse {
  /** Newly created tournament ID (UUID string) */
  tournamentId?: string
}

/**
 * API response for tournament list
 */
export interface TournamentListResponse {
  tournaments: TournamentApiData[]
  total: number
}

/**
 * API response for single tournament
 */
export interface TournamentCreateResponse {
  tournament: TournamentApiData
  message: string
}

/**
 * Join tournament response
 */
export interface JoinTournamentResponse extends SuccessResponse {}

/**
 * Leave tournament response
 */
export interface LeaveTournamentResponse extends SuccessResponse {}

/**
 * Confirm participation response
 */
export interface ConfirmParticipationResponse extends SuccessResponse {}

/**
 * Edit tournament response
 */
export interface EditTournamentResponse extends SuccessResponse {}

/**
 * Start tournament response
 */
export interface StartTournamentResponse extends SuccessResponse {}

/**
 * Cancel tournament response
 */
export interface CancelTournamentResponse extends SuccessResponse {}

/**
 * Tournament data as received from API (dates as strings)
 * This matches the TournamentListItem response from the backend
 */
export interface TournamentApiData {
  /** Tournament ID (UUID) */
  id: string
  /** Tournament name */
  name: string
  /** Tournament status (draft, scheduled, registration_open, registration_closed, starting, started, completed, cancelled) */
  status: string
  /** Scheduled start time (ISO 8601, optional) */
  scheduledStartTimeAt?: string
  /** Creation timestamp (ISO 8601) */
  createdAt: string
  /** Game slug */
  // TODO clean, we use gameSnapshot instead of gameId
  game?: string
  /** Game logo URL */
  gameLogoUrl?: string
  /** Banner URL */
  /** Game snapshot data from game-management service */
  gameSnapshot?: GameSnapshot
  bannerUrl?: string
  /** Creator ID */
  creatorId?: string
  /** Room link */
  roomLink?: string
  /** Is completed */
  isCompleted?: boolean
  /** Last activity timestamp */
  lastActivity?: string
  /** Updated at timestamp */
  updatedAt?: string
  /** Organizer info */
  organizer?: {
    id: string
    name: string
    logoUrl?: string
  }
  /** Format info */
  format?: {
    type: 'single_elimination' | 'double_elimination' | 'round_robin' | 'swiss'
    teamSize: number
    maxSlots: number
    bestOf?: number
  }
  /** Ready window */
  readyWindow?: {
    startsAt: string
    endsAt: string
  }
  /** Prize info */
  prize?: {
    type: 'custom' | 'cash' | 'points'
    description: string
    value?: number
    currency?: string
  }
  /** User-specific tournament information */
  me?: {
    participationStatus?: 'registered' | 'confirmed' | 'checked_in' | null
  }
  /** Current user's participation status (deprecated - use me.participationStatus) */
  userParticipationStatus?: 'registered' | 'confirmed' | 'checked_in' | null
  /** Number of confirmed participants */
  participantCount?: number
}

/**
 * Tournament detail tab type
 */
export type TournamentTab = 'overview' | 'bracket' | 'matches' | 'players' | 'results'

// ============================================================================
// Error Types
// ============================================================================

/**
 * Form validation errors
 */
export interface FormErrors {
  name?: string
  startTime?: string
  game?: string
  _general?: string
}

/**
 * Tournament operation error
 */
export interface TournamentError {
  type: 'network' | 'validation' | 'authorization' | 'server' | 'unknown'
  message: string
  details?: unknown
  fieldErrors?: Record<string, string>
}

// ============================================================================
// Component Props Types
// ============================================================================

/**
 * Props for TournamentTable component
 */
export interface TournamentTableProps {
  tournaments: Tournament[]
  isLoading: boolean
  error?: TournamentError | null
  onRefresh?: () => void
  onTournamentClick?: (tournament: Tournament) => void
}

/**
 * Props for CreateTournamentModal component
 */
export interface CreateTournamentModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: (tournament: Tournament) => void
}

/**
 * Props for TournamentSuccessModal component
 */
export interface TournamentSuccessModalProps {
  isOpen: boolean
  tournament: Tournament
  onClose: () => void
  onGoToRoom: (roomLink: string) => void
}

/**
 * Props for StatusBadge component
 */
export interface StatusBadgeProps {
  status: TournamentStatus
  size?: 'sm' | 'md' | 'lg'
}

// ============================================================================
// Hook Return Types
// ============================================================================

/**
 * Return type for useTournaments hook
 */
export interface UseTournamentsReturn {
  tournaments: Tournament[]
  isLoading: boolean
  error: TournamentError | null
  createTournament: (data: CreateTournamentInput) => Promise<Tournament>
  refreshTournaments: () => Promise<void>
  clearError: () => void
}

/**
 * Return type for useTournamentStatus hook
 */
export interface UseTournamentStatusReturn {
  status: TournamentStatus
  label: string
  colorClasses: {
    bg: string
    text: string
    border: string
  }
}

/**
 * Return type for useClipboard hook
 */
export interface UseClipboardReturn {
  copy: (text: string) => Promise<boolean>
  isCopied: boolean
  error: Error | null
  reset: () => void
}

// ============================================================================
// Bracket & Match Types (for tournament detail viewing)
// ============================================================================

/**
 * Match status
 */
export type MatchStatus = 
  | 'ready'       // Both players present, waiting for ready
  | 'started'     // Game in progress
  | 'paused'      // Game paused
  | 'completed'   // Match finished
  | 'disputed'    // Result under review
  | 'cancelled';  // Match cancelled

/**
 * Result source
 */
export type ResultSource = 
  | 'game_api'      // Result from game
  | 'manual_host'   // Host manually set winner
  | 'walkover';     // One player didn't show

/**
 * Match entity
 */
export interface Match {
  id: string;
  roundId: string;
  matchNumber: number;
  nextMatchId: string | null;
  participant1Id: string;
  participant2Id: string | null;
  status: MatchStatus;
  participant1Ready: boolean;
  participant2Ready: boolean;
  winnerId: string | null;
  scorePlayer1: number | null;
  scorePlayer2: number | null;
  resultSource: ResultSource | null;
  disconnectedPlayerId: string | null;
  disconnectedAt: string | null;
  gameApiMatchId: string | null;
  createdAt: string;
  startedAt: string | null;
  completedAt: string | null;
  updatedAt: string;
}

/**
 * Round status
 */
export type RoundStatus = 
  | 'pending'     // Not yet started
  | 'in_progress' // Some matches active
  | 'completed';  // All matches done

/**
 * Round in bracket
 */
export interface Round {
  id: string;
  roundNumber: number;
  roundName: string;
  matchesCount: number;
  status: RoundStatus;
  matches: Match[];
}

/**
 * Bracket structure
 */
export interface Bracket {
  id: string;
  tournamentId: string;
  totalRounds: number;
  totalMatches: number;
  byesCount: number;
  generatedAt: string;
  rounds: Round[];
}

/**
 * Player display info (for matches/bracket display)
 */
export interface MatchPlayer {
  id: string;
  displayName: string;
  avatarUrl: string | null;
}

/**
 * Match with player details (for display)
 */
export interface MatchWithPlayers extends Match {
  player1: MatchPlayer | null;
  player2: MatchPlayer | null;
}


