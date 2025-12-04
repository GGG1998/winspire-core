-- Create tournament_rounds table
CREATE TABLE tournament_rounds (
    -- Primary key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Foreign key to bracket
    bracket_id UUID NOT NULL REFERENCES tournament_brackets(id) ON DELETE CASCADE,
    
    -- Round metadata
    round_number INTEGER NOT NULL CHECK (round_number > 0),
    round_name VARCHAR(100) NOT NULL, -- "Round of 16", "Quarter-finals", "Semi-finals", "Final"
    matches_count INTEGER NOT NULL CHECK (matches_count > 0),
    
    -- State tracking
    status VARCHAR(20) NOT NULL DEFAULT 'pending' 
        CHECK (status IN ('pending', 'in_progress', 'completed')),
    
    -- Timestamps
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    
    -- Constraints
    CONSTRAINT unique_bracket_round_number UNIQUE (bracket_id, round_number)
);

CREATE INDEX idx_rounds_bracket_id ON tournament_rounds(bracket_id);
CREATE INDEX idx_rounds_status ON tournament_rounds(status) WHERE status = 'in_progress';

COMMENT ON TABLE tournament_rounds IS 'Represents rounds within a bracket (e.g., Quarter-finals, Semi-finals, Final)';
COMMENT ON COLUMN tournament_rounds.round_name IS 'Human-readable name: "Final" (1 match), "Semi-finals" (2), "Quarter-finals" (4), "Round of X" (8+)';
COMMENT ON COLUMN tournament_rounds.status IS 'pending → in_progress → completed';
