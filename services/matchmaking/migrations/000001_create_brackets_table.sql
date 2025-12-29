-- Create tournament_brackets table
CREATE TABLE tournament_brackets (
    -- Primary key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Tournament reference (UUID only, no FK - tournaments table in different DB)
    tournament_id UUID NOT NULL UNIQUE,
    
    -- Bracket metadata
    total_rounds INTEGER NOT NULL CHECK (total_rounds > 0),
    total_matches INTEGER NOT NULL CHECK (total_matches > 0),
    byes_count INTEGER NOT NULL DEFAULT 0 CHECK (byes_count >= 0),
    
    -- Timestamps
    generated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_brackets_tournament_id ON tournament_brackets(tournament_id);

COMMENT ON TABLE tournament_brackets IS 'Stores bracket structure for each tournament. One bracket per tournament.';
COMMENT ON COLUMN tournament_brackets.tournament_id IS 'References tournaments.id in tournament_db (no FK due to separate database)';
COMMENT ON COLUMN tournament_brackets.total_rounds IS 'Number of rounds calculated as CEIL(LOG2(participant_count))';
COMMENT ON COLUMN tournament_brackets.byes_count IS 'Number of byes assigned: (2^total_rounds) - participant_count';
