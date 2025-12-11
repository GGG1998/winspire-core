import { randomUUID } from 'crypto';
import { type SeededAccount } from '../fixtures/auth.fixture';
import { Pool } from 'pg';

const POSTGRES_DSN = process.env.POSTGRES_DSN ?? 'postgresql://postgres:postgres@localhost:54322/postgres?sslmode=disable';

function getPool() {
  return new Pool({ connectionString: POSTGRES_DSN });
}

type Id = string;

export interface MatchLobbySeed {
  tournamentId: Id;
  bracketId: Id;
  roundId: Id;
  matchId: Id;
  lobbyUrl: string;
}

async function insertBracket(pool: Pool, tournamentId: Id) {
  const { rows } = await pool.query<{ id: Id }>(
    `INSERT INTO tournament_brackets (
        id, tournament_id, game_snapshot, total_rounds, total_matches, byes_count
     ) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
    [randomUUID(), tournamentId, null, 1, 1, 0]
  );
  return rows[0].id;
}

async function insertRound(pool: Pool, bracketId: Id) {
  const { rows } = await pool.query<{ id: Id }>(
    `INSERT INTO tournament_rounds (
        id, bracket_id, round_number, round_name, matches_count, status
     ) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
    [randomUUID(), bracketId, 1, 'Round 1', 1, 'pending']
  );
  return rows[0].id;
}

async function insertMatch(
  pool: Pool,
  roundId: Id,
  participant1Id: Id,
  participant2Id?: Id
) {
  const { rows } = await pool.query<{ id: Id }>(
    `INSERT INTO tournament_matches (
        id, round_id, match_number, participant1_id, participant2_id, status,
        participant1_ready, participant2_ready, participant1_game_loaded, participant2_game_loaded
     ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
    [
      randomUUID(),
      roundId,
      1,
      participant1Id,
      participant2Id ?? null,
      'pending',
      false,
      false,
      false,
      false,
    ]
  );
  return rows[0].id;
}

interface SeedOptions {
  participant1: SeededAccount;
  participant2?: SeededAccount;
  attachSecond?: boolean;
}

export async function seedMatchLobby1v1(options: SeedOptions): Promise<MatchLobbySeed> {
  const pool = getPool();

  const participant2Id = options.attachSecond && options.participant2 
    ? options.participant2.id 
    : undefined;

  const tournamentId = randomUUID();
  const bracketId = await insertBracket(pool, tournamentId);
  const roundId = await insertRound(pool, bracketId);
  const matchId = await insertMatch(pool, roundId, options.participant1.id, participant2Id);
  await pool.end();

  return {
    tournamentId,
    bracketId,
    roundId,
    matchId,
    lobbyUrl: `/lobby/${tournamentId}/match/${matchId}`,
  };
}

export async function attachSecondParticipant(matchId: Id, participant2Id: Id) {
  const pool = getPool();
  await pool.query(
    `UPDATE tournament_matches SET participant2_id = $1 WHERE id = $2`,
    [participant2Id, matchId]
  );
  await pool.end();
}

// Seed match in loading state (both players ready, transitioning to game load)
export async function seedMatchInLoadingState(
  participant1: SeededAccount,
  participant2: SeededAccount,
  options?: { gameSlug?: string }
): Promise<MatchLobbySeed> {
  const pool = getPool();

  const tournamentId = randomUUID();
  
  // Create bracket with game snapshot
  const gameSnapshot = {
    id: randomUUID(),
    slug: options?.gameSlug || 'test-game-e2e',
    name: 'Test Game',
    version: '1.0.0',
    storagePath: '/games/test',
  };

  const { rows: bracketRows } = await pool.query<{ id: Id }>(
    `INSERT INTO tournament_brackets (
        id, tournament_id, game_snapshot, total_rounds, total_matches, byes_count
     ) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
    [randomUUID(), tournamentId, JSON.stringify(gameSnapshot), 1, 1, 0]
  );
  const bracketId = bracketRows[0].id;

  // Create round
  const { rows: roundRows } = await pool.query<{ id: Id }>(
    `INSERT INTO tournament_rounds (
        id, bracket_id, round_number, round_name, matches_count, status
     ) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
    [randomUUID(), bracketId, 1, 'Round 1', 1, 'in_progress']
  );
  const roundId = roundRows[0].id;

  // Create match in LOADING state with both players ready
  const { rows: matchRows } = await pool.query<{ id: Id }>(
    `INSERT INTO tournament_matches (
        id, round_id, match_number, participant1_id, participant2_id, status,
        participant1_ready, participant2_ready, 
        participant1_game_loaded, participant2_game_loaded
     ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
    [
      randomUUID(),
      roundId,
      1,
      participant1.id,
      participant2.id,
      'loading',  // Status: loading (both players clicked ready)
      true,       // participant1_ready
      true,       // participant2_ready
      false,      // participant1_game_loaded
      false,      // participant2_game_loaded
    ]
  );
  const matchId = matchRows[0].id;

  await pool.end();

  return {
    tournamentId,
    bracketId,
    roundId,
    matchId,
    lobbyUrl: `/lobby/${tournamentId}/match/${matchId}`,
  };
}

export { POSTGRES_DSN };
