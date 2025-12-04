// ============================================================================
// Timing Constants
// ============================================================================

// WebSocket reconnection timeout (in milliseconds)
export const RECONNECT_TIMEOUT = 30000; // 30 seconds

// Grace period duration for late arrivals (in seconds)
export const GRACE_PERIOD_DURATION = 30; // 30 seconds

// Match start countdown duration (in seconds)
export const MATCH_START_COUNTDOWN = 3; // 3, 2, 1

// Disconnect reconnect window (in seconds)
export const DISCONNECT_RECONNECT_WINDOW = 30; // 30 seconds

// Walkover claim timeout (in milliseconds)
export const WALKOVER_CLAIM_TIMEOUT = 120000; // 2 minutes

// Double no-show timeout (in milliseconds)
export const DOUBLE_NO_SHOW_TIMEOUT = 300000; // 5 minutes

// Game iframe load timeout (in milliseconds)
export const GAME_IFRAME_LOAD_TIMEOUT = 30000; // 30 seconds

// Match assigned notification display duration (in milliseconds)
export const MATCH_ASSIGNED_NOTIFICATION_DURATION = 2000; // 2 seconds

// ============================================================================
// WebSocket Configuration
// ============================================================================

// WebSocket reconnection backoff sequence (in milliseconds)
export const WS_RECONNECT_BACKOFF = [1000, 2000, 4000, 8000, 16000, 30000]; // 1s, 2s, 4s, 8s, 16s, max 30s

// WebSocket heartbeat interval (in milliseconds)
export const WS_HEARTBEAT_INTERVAL = 30000; // 30 seconds

// WebSocket message queue max size
export const WS_MESSAGE_QUEUE_MAX_SIZE = 100;

// ============================================================================
// UI Constants
// ============================================================================

// Activity feed max items to display
export const ACTIVITY_FEED_MAX_ITEMS = 50;

// Participant list items per page
export const PARTICIPANTS_PER_PAGE = 20;

// Bracket horizontal scroll speed
export const BRACKET_SCROLL_SPEED = 10;

// Bracket zoom levels
export const BRACKET_ZOOM_LEVELS = [0.5, 0.75, 1.0, 1.25, 1.5];
export const BRACKET_DEFAULT_ZOOM = 1.0;

// ============================================================================
// Status Labels
// ============================================================================

export const MATCH_STATUS_LABELS: Record<string, string> = {
  pending: 'Oczekuje',
  ready: 'Gotowy',
  started: 'Na żywo',
  paused: 'Wstrzymany',
  completed: 'Zakończony',
  disputed: 'Sporny',
  cancelled: 'Anulowany',
};

export const ROUND_NAME_TRANSLATIONS: Record<string, string> = {
  'Round 1': 'Runda 1',
  'Round of 16': '1/8 finału',
  'Quarter-finals': 'Ćwierćfinały',
  'Quarterfinals': 'Ćwierćfinały',
  'Semi-finals': 'Półfinały',
  'Semifinals': 'Półfinały',
  Final: 'Finał',
};

export const PRE_LOBBY_STATUS_LABELS: Record<string, string> = {
  waiting: 'Oczekiwanie',
  grace_period: 'Okres Łaski',
  generating_bracket: 'Generowanie Drabinki',
  started: 'Rozpoczęty',
};

// ============================================================================
// Error Messages
// ============================================================================

export const ERROR_MESSAGES = {
  UNAUTHORIZED: 'Unauthorized - you are not a participant in this match',
  NOT_REGISTERED: 'Unauthorized - you must be registered for this tournament',
  NOT_MATCH_PARTICIPANT: 'Nie masz uprawnień do tego meczu - nie jesteś jego uczestnikiem',
  TOURNAMENT_IN_PROGRESS: 'Tournament in progress - registration closed',
  CONNECTION_LOST: 'Connection lost - attempting to reconnect...',
  GAME_LOAD_FAILED: 'Failed to load game - please retry',
  WALKOVER_FAILED: 'Failed to claim walkover',
  READY_FAILED: 'Failed to update ready status',
  MATCH_NOT_FOUND: 'Match not found',
  TOURNAMENT_NOT_FOUND: 'Tournament not found',
  BRACKET_NOT_FOUND: 'Bracket not generated yet',
} as const;

// ============================================================================
// API Endpoints (relative to matchmaking service base)
// ============================================================================

export const API_ENDPOINTS = {
  // Tournament Pre-Lobby
  GET_PRE_LOBBY_STATE: (tournamentId: string) => `/v1/tournaments/${tournamentId}/lobby`,
  WS_PRE_LOBBY: (tournamentId: string) => `/v1/tournaments/${tournamentId}/lobby`,

  // Match Lobby
  GET_MATCH: (matchId: string) => `/v1/matches/${matchId}`,
  MARK_READY: (matchId: string) => `/v1/matches/${matchId}/ready`,
  CLAIM_WALKOVER: (matchId: string) => `/v1/matches/${matchId}/walkover`,
  WS_MATCH_LOBBY: (matchId: string) => `/v1/matches/${matchId}/lobby`,

  // Bracket
  GET_BRACKET: (tournamentId: string) => `/v1/tournaments/${tournamentId}/bracket`,

  // Matches List
  GET_MATCHES: (tournamentId: string) => `/v1/tournaments/${tournamentId}/matches`,
} as const;

// ============================================================================
// Game Iframe Configuration
// ============================================================================

export const GAME_IFRAME_CONFIG = {
  LOAD_TIMEOUT_MS: 30000, // 30 seconds
  ASPECT_RATIO: '16:9',
  ALLOW_PERMISSIONS: 'fullscreen; gamepad; microphone',
  SANDBOX_FLAGS: 'allow-scripts allow-same-origin allow-forms allow-popups',
} as const;

// ============================================================================
// Local Storage Keys
// ============================================================================

export const STORAGE_KEYS = {
  LAST_TOURNAMENT_ID: 'winspire:lastTournamentId',
  LAST_MATCH_ID: 'winspire:lastMatchId',
  BRACKET_ZOOM_LEVEL: 'winspire:bracketZoomLevel',
} as const;

