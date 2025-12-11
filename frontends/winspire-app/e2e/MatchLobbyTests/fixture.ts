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

export { POSTGRES_DSN };
