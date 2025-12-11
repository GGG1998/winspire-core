-- Add completed_at timestamp to tournament_brackets table
ALTER TABLE tournament_brackets 
ADD COLUMN completed_at TIMESTAMP NULL;

COMMENT ON COLUMN tournament_brackets.completed_at IS 'Timestamp when the tournament bracket was completed (final match finished)';










