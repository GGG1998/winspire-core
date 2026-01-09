-- Fix match flow: change from ready→loading to loading→ready
-- New flow: pending → loading → ready → started
--
-- The old constraint `check_loading_requires_both_ready` enforced that both players
-- must be ready BEFORE the match can enter 'loading' state. This is backwards.
-- In the correct flow, players load the game first, then mark ready.

-- Drop the old constraint that required both players to be ready before loading
ALTER TABLE tournament_matches DROP CONSTRAINT IF EXISTS check_loading_requires_both_ready;

-- Update comment to reflect new flow
COMMENT ON COLUMN tournament_matches.status IS 'State machine: pending → loading → ready → started → [paused →] completed';
