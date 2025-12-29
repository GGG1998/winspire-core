import { randomUUID } from 'crypto';
import { type SeededAccount } from '../fixtures/auth.fixture';
import { Pool } from 'pg';
import type { APIRequestContext } from '@playwright/test';

// ============================================================================
// Configuration
// ============================================================================

const POSTGRES_DSN = process.env.POSTGRES_DSN ?? 'postgresql://postgres:postgres@localhost:54322/postgres?sslmode=disable';
const MATCHMAKING_API_URL = process.env.MATCHMAKING_API_URL ?? 'http://localhost:8088';

// Default host ID (must exist in hosts table)
const DEFAULT_HOST_ID = '00000000-0000-0000-0000-000000000001';

type Id = string;

// Shared pool instance for reuse within a single fixture call
let _pool: Pool | null = null;

function getPool(): Pool {
  if (!_pool) {
    _pool = new Pool({ connectionString: POSTGRES_DSN });
  }
  return _pool;
}

async function closePool(): Promise<void> {
  if (_pool) {
    await _pool.end();
    _pool = null;
  }
}

// ============================================================================
// Tournament Service Helpers (tournament + registrations)
// ============================================================================

/**
 * Ensures the default E2E host exists. Safe to call multiple times.
 */
async function ensureDefaultHost(pool: Pool): Promise<void> {
  await pool.query(
    `INSERT INTO hosts (id, name, description)
     VALUES ($1, $2, $3)
     ON CONFLICT (id) DO NOTHING`,
    [DEFAULT_HOST_ID, 'E2E Test Host', 'Default host for E2E testing']
  );
}

interface CreateTournamentOptions {
  tournamentId: Id;
  name?: string;
  status?: 'draft' | 'registration_open' | 'registration_closed' | 'started' | 'completed' | 'cancelled';
  minPlayers?: number;
  maxPlayers?: number;
}

/**
 * Creates a tournament in the tournament service database.
 */
async function createTournament(pool: Pool, options: CreateTournamentOptions): Promise<void> {
  await ensureDefaultHost(pool);

  await pool.query(
    `INSERT INTO tournaments (
      id, host_id, name, description, status,
      scheduled_start_time_at, actual_start_time_at,
      minimum_team_count, maximum_team_count, team_size
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    ON CONFLICT (id) DO NOTHING`,
    [
      options.tournamentId,
      DEFAULT_HOST_ID,
      options.name ?? 'E2E Test Tournament',
      'Tournament created for E2E testing',
      options.status ?? 'started',
      new Date().toISOString(),
      new Date().toISOString(),
      options.minPlayers ?? 2,
      options.maxPlayers ?? 16,
      1, // team_size (solo)
    ]
  );
}

interface RegisterPlayerOptions {
  tournamentId: Id;
  player: SeededAccount;
  status?: 'pending' | 'confirmed' | 'checked_in' | 'withdrawn' | 'disqualified';
}

/**
 * Registers a player for a tournament in the tournament service database.
 */
async function registerPlayer(pool: Pool, options: RegisterPlayerOptions): Promise<void> {
  const displayName = options.player.session.user?.user_metadata?.nickname
    || `Player ${options.player.id.slice(0, 8)}`;

  await pool.query(
    `INSERT INTO tournament_registrations (
      tournament_id, user_id, status, display_name, avatar_url, checked_in_at, is_ready
    ) VALUES ($1, $2, $3, $4, $5, $6, $7)
    ON CONFLICT (tournament_id, user_id) DO NOTHING`,
    [
      options.tournamentId,
      options.player.id,
      options.status ?? 'checked_in',
      displayName,
      null,
      new Date().toISOString(),
      true,
    ]
  );
}

/**
 * Registers multiple players for a tournament.
 */
async function registerPlayers(
  pool: Pool,
  tournamentId: Id,
  players: SeededAccount[],
  status?: 'pending' | 'confirmed' | 'checked_in' | 'withdrawn' | 'disqualified'
): Promise<void> {
  for (const player of players) {
    await registerPlayer(pool, { tournamentId, player, status });
  }
}

// ============================================================================
// Pre-lobby Helpers
// ============================================================================

interface CreatePrelobbyOptions {
  tournamentId: Id;
  status?: 'waiting' | 'grace_period' | 'generating_bracket' | 'started' | 'cancelled';
  minParticipants?: number;
}

/**
 * Creates a pre-lobby record in the matchmaking service database.
 * This is required for the pre-lobby page to load correctly.
 */
async function createPrelobby(pool: Pool, options: CreatePrelobbyOptions): Promise<Id> {
  const { rows } = await pool.query<{ id: Id }>(
    `INSERT INTO prelobbies (
      id, tournament_id, status, min_participants
    ) VALUES ($1, $2, $3, $4)
    ON CONFLICT (tournament_id) DO UPDATE SET
      status = EXCLUDED.status,
      updated_at = NOW()
    RETURNING id`,
    [
      randomUUID(),
      options.tournamentId,
      options.status ?? 'started',
      options.minParticipants ?? 2,
    ]
  );
  return rows[0].id;
}

// ============================================================================
// Matchmaking Service Helpers (brackets, rounds, matches)
// ============================================================================

interface CreateBracketOptions {
  tournamentId: Id;
  totalRounds?: number;
  totalMatches?: number;
  byesCount?: number;
  gameSlug?: string;
}

/**
 * Creates a bracket in the matchmaking service database.
 */
async function createBracket(pool: Pool, options: CreateBracketOptions): Promise<Id> {
  const gameSnapshot = {
    id: randomUUID(),
    slug: options.gameSlug ?? 'test-game-e2e',
    name: 'Test Game',
    version: '1.0.0',
    storagePath: '/games/test',
  };

  const { rows } = await pool.query<{ id: Id }>(
    `INSERT INTO tournament_brackets (
      id, tournament_id, game_snapshot, total_rounds, total_matches, byes_count
    ) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
    [
      randomUUID(),
      options.tournamentId,
      JSON.stringify(gameSnapshot),
      options.totalRounds ?? 1,
      options.totalMatches ?? 1,
      options.byesCount ?? 0,
    ]
  );
  return rows[0].id;
}

interface CreateRoundOptions {
  bracketId: Id;
  roundNumber: number;
  roundName: string;
  matchesCount: number;
  status?: 'pending' | 'in_progress' | 'completed';
}

/**
 * Creates a round in the matchmaking service database.
 */
async function createRound(pool: Pool, options: CreateRoundOptions): Promise<Id> {
  const { rows } = await pool.query<{ id: Id }>(
    `INSERT INTO tournament_rounds (
      id, bracket_id, round_number, round_name, matches_count, status
    ) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
    [
      randomUUID(),
      options.bracketId,
      options.roundNumber,
      options.roundName,
      options.matchesCount,
      options.status ?? 'pending',
    ]
  );
  return rows[0].id;
}

type MatchStatus = 'pending' | 'ready' | 'loading' | 'started' | 'paused' | 'completed' | 'disputed' | 'cancelled';

interface CreateMatchOptions {
  roundId: Id;
  matchNumber: number;
  participant1Id: Id | null;
  participant2Id: Id | null;
  nextMatchId?: Id | null;
  status?: MatchStatus;
  participant1Ready?: boolean;
  participant2Ready?: boolean;
  participant1GameLoaded?: boolean;
  participant2GameLoaded?: boolean;
  startedAt?: string | null;
}

/**
 * Creates a match in the matchmaking service database.
 */
async function createMatch(pool: Pool, options: CreateMatchOptions): Promise<Id> {
  const { rows } = await pool.query<{ id: Id }>(
    `INSERT INTO tournament_matches (
      id, round_id, match_number, participant1_id, participant2_id,
      next_match_id, status,
      participant1_ready, participant2_ready,
      participant1_game_loaded, participant2_game_loaded,
      started_at
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id`,
    [
      randomUUID(),
      options.roundId,
      options.matchNumber,
      options.participant1Id,
      options.participant2Id,
      options.nextMatchId ?? null,
      options.status ?? 'pending',
      options.participant1Ready ?? false,
      options.participant2Ready ?? false,
      options.participant1GameLoaded ?? false,
      options.participant2GameLoaded ?? false,
      options.startedAt ?? null,
    ]
  );
  return rows[0].id;
}

// ============================================================================
// Exported Types
// ============================================================================

export interface MatchSeed {
  matchId: Id;
  participant1Id: Id;
  participant2Id: Id;
}

export interface TournamentBracketSeed {
  tournamentId: Id;
  bracketId: Id;
  round1Id: Id;
  round2Id: Id;
  finalMatchId: Id;
  round1Matches: MatchSeed[];
  prelobbyUrl: string;
}

/**
 * Seed result for a 6-player bracket with 2 byes.
 * Structure:
 *   Round 1: 2 matches (4 players play, 2 have byes)
 *   Round 2: 2 matches (R1 winners vs bye players)
 *   Round 3: 1 final match
 */
export interface SixPlayerBracketSeed {
  tournamentId: Id;
  bracketId: Id;
  rounds: {
    round1Id: Id;
    round2Id: Id;
    round3Id: Id;  // Final
  };
  matches: {
    // Round 1
    match1: MatchSeed;  // player1 vs player2
    match2: MatchSeed;  // player5 vs player6
    // Round 2
    match3: MatchSeed;  // R1M1 winner vs player3 (bye)
    match4: MatchSeed;  // R1M2 winner vs player4 (bye)
    // Round 3
    finalMatch: MatchSeed;
  };
  byePlayers: [Id, Id];  // player3, player4
  prelobbyUrl: string;
}

export interface MatchLobbySeed {
  tournamentId: Id;
  bracketId: Id;
  roundId: Id;
  matchId: Id;
  lobbyUrl: string;
}

// ============================================================================
// High-Level Seed Functions
// ============================================================================

interface SeedBracketOptions {
  players: SeededAccount[]; // Expects 4 players for a simple bracket
}

/**
 * Seeds a 4-player single elimination bracket:
 *
 * Round 1:
 *   Match 1: player[0] vs player[1] -> winner goes to final
 *   Match 2: player[2] vs player[3] -> winner goes to final
 *
 * Round 2 (Final):
 *   Match 3: placeholders (actual winners set by backend)
 *
 * Seeds both tournament service data (tournament + registrations)
 * and matchmaking service data (bracket, rounds, matches).
 */
export async function seedTournamentBracket(options: SeedBracketOptions): Promise<TournamentBracketSeed> {
  const pool = getPool();
  const { players } = options;

  if (players.length !== 4) {
    throw new Error('seedTournamentBracket requires exactly 4 players');
  }

  const tournamentId = randomUUID();

  // === Tournament Service Data ===
  await createTournament(pool, { tournamentId, name: 'E2E 4-Player Bracket', minPlayers: 4 });
  await registerPlayers(pool, tournamentId, players);

  // === Pre-lobby Data ===
  await createPrelobby(pool, { tournamentId, status: 'started', minParticipants: 4 });

  // === Matchmaking Service Data ===
  const bracketId = await createBracket(pool, { tournamentId, totalRounds: 2, totalMatches: 3 });

  const round1Id = await createRound(pool, {
    bracketId,
    roundNumber: 1,
    roundName: 'Round 1',
    matchesCount: 2,
    status: 'in_progress',
  });

  const round2Id = await createRound(pool, {
    bracketId,
    roundNumber: 2,
    roundName: 'Final',
    matchesCount: 1,
    status: 'pending',
  });

  // Create final match first (for FK reference)
  const finalMatchId = await createMatch(pool, {
    roundId: round2Id,
    matchNumber: 1,
    participant1Id: players[0].id, // Placeholder
    participant2Id: players[2].id, // Placeholder
    nextMatchId: null,
    status: 'pending',
  });

  // Create round 1 matches
  const match1Id = await createMatch(pool, {
    roundId: round1Id,
    matchNumber: 1,
    participant1Id: players[0].id,
    participant2Id: players[1].id,
    nextMatchId: finalMatchId,
    status: 'pending',
  });

  const match2Id = await createMatch(pool, {
    roundId: round1Id,
    matchNumber: 2,
    participant1Id: players[2].id,
    participant2Id: players[3].id,
    nextMatchId: finalMatchId,
    status: 'pending',
  });

  // Ensure all writes are committed before closing the pool
  await closePool();

  // Small delay to ensure database has processed all writes
  await new Promise(resolve => setTimeout(resolve, 100));

  console.log(`[seedTournamentBracket] Created tournament ${tournamentId} with matches:`, {
    match1Id,
    match2Id,
    finalMatchId,
  });

  return {
    tournamentId,
    bracketId,
    round1Id,
    round2Id,
    finalMatchId,
    round1Matches: [
      { matchId: match1Id, participant1Id: players[0].id, participant2Id: players[1].id },
      { matchId: match2Id, participant1Id: players[2].id, participant2Id: players[3].id },
    ],
    prelobbyUrl: `/tournaments/${tournamentId}/prelobby`,
  };
}

interface SeedSixPlayerBracketOptions {
  players: SeededAccount[]; // Expects exactly 6 players
}

/**
 * Seeds a 6-player single elimination bracket with 2 byes:
 *
 * Round 1 (2 matches):
 *   Match 1: player[0] vs player[1] -> winner goes to Round 2 Match 3
 *   Match 2: player[4] vs player[5] -> winner goes to Round 2 Match 4
 *
 * Round 2 (2 matches):
 *   Match 3: player[2] (bye) vs R1M1 winner -> winner goes to Final
 *   Match 4: player[3] (bye) vs R1M2 winner -> winner goes to Final
 *
 * Round 3 (Final):
 *   Match 5: R2M3 winner vs R2M4 winner
 *
 * Seeds both tournament service data (tournament + registrations)
 * and matchmaking service data (bracket, rounds, matches).
 */
export async function seedSixPlayerBracket(options: SeedSixPlayerBracketOptions): Promise<SixPlayerBracketSeed> {
  const pool = getPool();
  const { players } = options;

  if (players.length !== 6) {
    throw new Error('seedSixPlayerBracket requires exactly 6 players');
  }

  const tournamentId = randomUUID();

  // === Tournament Service Data ===
  await createTournament(pool, { tournamentId, name: 'E2E 6-Player Bracket', minPlayers: 6 });
  await registerPlayers(pool, tournamentId, players);

  // === Pre-lobby Data ===
  await createPrelobby(pool, { tournamentId, status: 'started', minParticipants: 6 });

  // === Matchmaking Service Data ===
  const bracketId = await createBracket(pool, { tournamentId, totalRounds: 3, totalMatches: 5, byesCount: 2 });

  // Create rounds
  const round1Id = await createRound(pool, {
    bracketId,
    roundNumber: 1,
    roundName: 'Round 1',
    matchesCount: 2,
    status: 'in_progress',
  });

  const round2Id = await createRound(pool, {
    bracketId,
    roundNumber: 2,
    roundName: 'Semifinals',
    matchesCount: 2,
    status: 'pending',
  });

  const round3Id = await createRound(pool, {
    bracketId,
    roundNumber: 3,
    roundName: 'Final',
    matchesCount: 1,
    status: 'pending',
  });

  // Create matches in reverse order for FK references
  // Round 3 (Final) - created first
  // Note: participant1_id has NOT NULL constraint, so we pre-assign expected R2M3 winner
  // participant2_id is left null for R2M4 winner to be assigned
  const finalMatchId = await createMatch(pool, {
    roundId: round3Id,
    matchNumber: 1,
    participant1Id: players[0].id, // Pre-assign expected R2M3 winner (player1)
    participant2Id: null,          // TBD: R2M4 winner will be assigned here
    nextMatchId: null,
    status: 'pending',
  });

  // Round 2 matches - bye players pre-assigned
  const match3Id = await createMatch(pool, {
    roundId: round2Id,
    matchNumber: 1,
    participant1Id: players[2].id, // Player 3 (bye)
    participant2Id: null,          // TBD: R1M1 winner
    nextMatchId: finalMatchId,
    status: 'pending',
  });

  const match4Id = await createMatch(pool, {
    roundId: round2Id,
    matchNumber: 2,
    participant1Id: players[3].id, // Player 4 (bye)
    participant2Id: null,          // TBD: R1M2 winner
    nextMatchId: finalMatchId,
    status: 'pending',
  });

  // Round 1 matches - active players
  const match1Id = await createMatch(pool, {
    roundId: round1Id,
    matchNumber: 1,
    participant1Id: players[0].id, // Player 1
    participant2Id: players[1].id, // Player 2
    nextMatchId: match3Id,         // Winner goes to R2M3
    status: 'pending',
  });

  const match2Id = await createMatch(pool, {
    roundId: round1Id,
    matchNumber: 2,
    participant1Id: players[4].id, // Player 5
    participant2Id: players[5].id, // Player 6
    nextMatchId: match4Id,         // Winner goes to R2M4
    status: 'pending',
  });

  // Ensure all writes are committed before closing the pool
  await closePool();

  // Small delay to ensure database has processed all writes
  await new Promise(resolve => setTimeout(resolve, 100));

  console.log(`[seedSixPlayerBracket] Created tournament ${tournamentId} with matches:`, {
    round1: { match1Id, match2Id },
    round2: { match3Id, match4Id },
    final: finalMatchId,
    byePlayers: [players[2].id, players[3].id],
  });

  return {
    tournamentId,
    bracketId,
    rounds: {
      round1Id,
      round2Id,
      round3Id,
    },
    matches: {
      match1: { matchId: match1Id, participant1Id: players[0].id, participant2Id: players[1].id },
      match2: { matchId: match2Id, participant1Id: players[4].id, participant2Id: players[5].id },
      match3: { matchId: match3Id, participant1Id: players[2].id, participant2Id: '' }, // bye player, R1 winner TBD
      match4: { matchId: match4Id, participant1Id: players[3].id, participant2Id: '' }, // bye player, R1 winner TBD
      finalMatch: { matchId: finalMatchId, participant1Id: players[0].id, participant2Id: '' }, // p1 pre-assigned, p2 TBD
    },
    byePlayers: [players[2].id, players[3].id],
    prelobbyUrl: `/tournaments/${tournamentId}/prelobby`,
  };
}

/**
 * Seeds a match already in 'started' state (for testing completion flow).
 */
export async function seedMatchInStartedState(
  players: [SeededAccount, SeededAccount]
): Promise<MatchLobbySeed> {
  const pool = getPool();
  const tournamentId = randomUUID();

  // === Tournament Service Data ===
  await createTournament(pool, { tournamentId, name: 'E2E Started Match', minPlayers: 2 });
  await registerPlayers(pool, tournamentId, players);

  // === Matchmaking Service Data ===
  const bracketId = await createBracket(pool, { tournamentId, totalRounds: 1, totalMatches: 1 });

  const roundId = await createRound(pool, {
    bracketId,
    roundNumber: 1,
    roundName: 'Final',
    matchesCount: 1,
    status: 'in_progress',
  });

  const matchId = await createMatch(pool, {
    roundId,
    matchNumber: 1,
    participant1Id: players[0].id,
    participant2Id: players[1].id,
    status: 'started',
    participant1Ready: true,
    participant2Ready: true,
    participant1GameLoaded: true,
    participant2GameLoaded: true,
    startedAt: new Date().toISOString(),
  });

  await closePool();

  return {
    tournamentId,
    bracketId,
    roundId,
    matchId,
    lobbyUrl: `/lobby/${tournamentId}/match/${matchId}`,
  };
}

/**
 * Seeds a 1v1 match lobby (pending state, waiting for players).
 */
export async function seedMatchLobby1v1(options: {
  participant1: SeededAccount;
  participant2?: SeededAccount;
  attachSecond?: boolean;
}): Promise<MatchLobbySeed> {
  const pool = getPool();
  const tournamentId = randomUUID();

  const players = options.attachSecond && options.participant2
    ? [options.participant1, options.participant2]
    : [options.participant1];

  // === Tournament Service Data ===
  await createTournament(pool, { tournamentId, name: 'E2E 1v1 Match', minPlayers: 2 });
  await registerPlayers(pool, tournamentId, players);

  // === Matchmaking Service Data ===
  const bracketId = await createBracket(pool, { tournamentId });

  const roundId = await createRound(pool, {
    bracketId,
    roundNumber: 1,
    roundName: 'Round 1',
    matchesCount: 1,
    status: 'pending',
  });

  const participant2Id = options.attachSecond && options.participant2
    ? options.participant2.id
    : null;

  const matchId = await createMatch(pool, {
    roundId,
    matchNumber: 1,
    participant1Id: options.participant1.id,
    participant2Id,
    status: 'pending',
  });

  await closePool();

  return {
    tournamentId,
    bracketId,
    roundId,
    matchId,
    lobbyUrl: `/lobby/${tournamentId}/match/${matchId}`,
  };
}

/**
 * Seeds a match in 'loading' state (both players ready, waiting for game load).
 */
export async function seedMatchInLoadingState(
  participant1: SeededAccount,
  participant2: SeededAccount,
  options?: { gameSlug?: string }
): Promise<MatchLobbySeed> {
  const pool = getPool();
  const tournamentId = randomUUID();

  // === Tournament Service Data ===
  await createTournament(pool, { tournamentId, name: 'E2E Loading Match', minPlayers: 2 });
  await registerPlayers(pool, tournamentId, [participant1, participant2]);

  // === Matchmaking Service Data ===
  const bracketId = await createBracket(pool, {
    tournamentId,
    gameSlug: options?.gameSlug,
  });

  const roundId = await createRound(pool, {
    bracketId,
    roundNumber: 1,
    roundName: 'Round 1',
    matchesCount: 1,
    status: 'in_progress',
  });

  const matchId = await createMatch(pool, {
    roundId,
    matchNumber: 1,
    participant1Id: participant1.id,
    participant2Id: participant2.id,
    status: 'loading',
    participant1Ready: true,
    participant2Ready: true,
    participant1GameLoaded: false,
    participant2GameLoaded: false,
  });

  await closePool();

  return {
    tournamentId,
    bracketId,
    roundId,
    matchId,
    lobbyUrl: `/lobby/${tournamentId}/match/${matchId}`,
  };
}

// ============================================================================
// API Helpers
// ============================================================================

const TOURNAMENT_API_URL = process.env.TOURNAMENT_API_URL ?? 'http://localhost:8089';

/**
 * Verifies that a tournament exists in the tournament service via internal API.
 * Useful for debugging 404 errors in E2E tests.
 */
export async function verifyTournamentExists(
  request: APIRequestContext,
  tournamentId: Id
): Promise<{ exists: boolean; data?: Record<string, unknown>; error?: string }> {
  const url = `${TOURNAMENT_API_URL}/internal/tournaments/${tournamentId}`;
  console.log(`[verifyTournamentExists] Checking: ${url}`);

  try {
    const response = await request.get(url);
    const status = response.status();
    console.log(`[verifyTournamentExists] Response status: ${status}`);

    if (response.ok()) {
      const data = await response.json();
      console.log(`[verifyTournamentExists] Tournament found:`, data);
      return { exists: true, data };
    }

    const body = await response.text();
    console.log(`[verifyTournamentExists] Response body:`, body);

    if (status === 404) {
      return { exists: false, error: `Tournament ${tournamentId} not found (404). Body: ${body}` };
    }

    return { exists: false, error: `Unexpected status ${status}: ${body}` };
  } catch (err) {
    console.log(`[verifyTournamentExists] Request failed:`, err);
    return { exists: false, error: `Request failed: ${err}` };
  }
}

/**
 * Verifies that a match exists in the matchmaking service via API.
 */
export async function verifyMatchExists(
  request: APIRequestContext,
  matchId: Id,
  token: string
): Promise<{ exists: boolean; data?: Record<string, unknown>; error?: string }> {
  const url = `${MATCHMAKING_API_URL}/v1/matchmaking/matches/${matchId}`;
  console.log(`[verifyMatchExists] Checking: ${url}`);

  try {
    const response = await request.get(url, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });

    const status = response.status();
    console.log(`[verifyMatchExists] Response status: ${status}`);

    if (response.ok()) {
      const data = await response.json();
      console.log(`[verifyMatchExists] Match found:`, data);
      return { exists: true, data };
    }

    const body = await response.text();
    console.log(`[verifyMatchExists] Response body:`, body);

    if (status === 404) {
      return { exists: false, error: `Match ${matchId} not found (404). Body: ${body}` };
    }

    return { exists: false, error: `Unexpected status ${status}: ${body}` };
  } catch (err) {
    console.log(`[verifyMatchExists] Request failed:`, err);
    return { exists: false, error: `Request failed: ${err}` };
  }
}

interface CompleteMatchOptions {
  matchId: Id;
  winnerId: Id;
  loserId: Id;
  scoreWinner: number;
  scoreLoser: number;
  token: string;
}

/**
 * Completes a match via the matchmaking API.
 * This triggers the RoundManager to send post-match WebSocket messages.
 */
export async function completeMatchViaApi(
  request: APIRequestContext,
  options: CompleteMatchOptions
): Promise<void> {
  const response = await request.post(`${MATCHMAKING_API_URL}/v1/matchmaking/matches/${options.matchId}/complete`, {
    headers: {
      'Authorization': `Bearer ${options.token}`,
      'Content-Type': 'application/json',
    },
    data: {
      winner_id: options.winnerId,
      loser_id: options.loserId,
      score_winner: options.scoreWinner,
      score_loser: options.scoreLoser,
      result_source: 'manual_host',
    },
    timeout: 60000, // 60 seconds - match completion can take time due to round management
  });

  if (!response.ok()) {
    const body = await response.text();
    throw new Error(`Failed to complete match: ${response.status()} - ${body}`);
  }
}

/**
 * Directly update match status in DB (for testing edge cases).
 */
export async function updateMatchStatus(
  matchId: Id,
  status: string,
  winnerId?: Id,
  scorePlayer1?: number,
  scorePlayer2?: number
): Promise<void> {
  const pool = getPool();

  await pool.query(
    `UPDATE tournament_matches
     SET status = $1, winner_id = $2, score_player1 = $3, score_player2 = $4, completed_at = $5
     WHERE id = $6`,
    [status, winnerId ?? null, scorePlayer1 ?? null, scorePlayer2 ?? null, new Date().toISOString(), matchId]
  );

  await closePool();
}

/**
 * Attach second participant to an existing match.
 */
export async function attachSecondParticipant(matchId: Id, participant2Id: Id): Promise<void> {
  const pool = getPool();
  await pool.query(
    `UPDATE tournament_matches SET participant2_id = $1 WHERE id = $2`,
    [participant2Id, matchId]
  );
  await closePool();
}

// ============================================================================
// Exports
// ============================================================================

export { POSTGRES_DSN, MATCHMAKING_API_URL, DEFAULT_HOST_ID };
