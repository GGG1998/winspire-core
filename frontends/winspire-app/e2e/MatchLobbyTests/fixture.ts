/**
 * Match Lobby E2E Test Fixtures
 *
 * Re-exports from round-transition.fixture.ts for backward compatibility.
 * All fixture logic is consolidated in round-transition.fixture.ts to avoid duplication.
 */

export {
  // Types
  type MatchLobbySeed,
  type MatchSeed,
  type TournamentBracketSeed,

  // Seed functions
  seedMatchLobby1v1,
  seedMatchInLoadingState,
  seedMatchInStartedState,
  seedTournamentBracket,

  // Utility functions
  attachSecondParticipant,
  updateMatchStatus,
  completeMatchViaApi,

  // Constants
  POSTGRES_DSN,
  MATCHMAKING_API_URL,
  DEFAULT_HOST_ID,
} from './round-transition.fixture';
