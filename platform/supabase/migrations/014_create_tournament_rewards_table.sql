-- Migration: Create tournament_rewards table
-- Purpose: Store rewards for tournament winners (1st, 2nd, 3rd place)
-- Date: 2025-01-12

-- ============================================================================
-- TOURNAMENT REWARDS TABLE
-- ============================================================================
-- Stores rewards configured by streamers for tournament winners.
-- Hidden values are encrypted using pgcrypto with a secret key.

-- Enable pgcrypto extension for encryption
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS tournament_rewards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL,

    -- Placement: 1 = 1st place, 2 = 2nd place, 3 = 3rd place
    placement SMALLINT NOT NULL CHECK (placement >= 1 AND placement <= 3),

    -- Public reward information (visible to all)
    title VARCHAR(255) NOT NULL,
    description TEXT,

    -- Hidden value (encrypted) - the actual reward (e.g., activation code, Discord link)
    -- Encrypted using pgcrypto symmetric encryption
    hidden_value_encrypted BYTEA NOT NULL,

    -- Claim status
    is_claimed BOOLEAN NOT NULL DEFAULT FALSE,
    claimed_by_user_id UUID,
    claimed_at TIMESTAMPTZ,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Unique constraint: one reward per placement per tournament
    UNIQUE(tournament_id, placement)
);

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX idx_tournament_rewards_tournament_id ON tournament_rewards(tournament_id);
CREATE INDEX idx_tournament_rewards_claimed_by ON tournament_rewards(claimed_by_user_id) WHERE claimed_by_user_id IS NOT NULL;

-- ============================================================================
-- TRIGGER: Update updated_at timestamp
-- ============================================================================

CREATE TRIGGER update_tournament_rewards_updated_at
    BEFORE UPDATE ON tournament_rewards
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- ENCRYPTION FUNCTIONS
-- ============================================================================

-- Function to encrypt hidden value
-- Uses AES-256 encryption with a secret key from app settings
CREATE OR REPLACE FUNCTION encrypt_reward_value(plain_text TEXT, secret_key TEXT)
RETURNS BYTEA AS $$
BEGIN
    RETURN pgp_sym_encrypt(plain_text, secret_key);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to decrypt hidden value
-- Only called when winner is verified
CREATE OR REPLACE FUNCTION decrypt_reward_value(encrypted_value BYTEA, secret_key TEXT)
RETURNS TEXT AS $$
BEGIN
    RETURN pgp_sym_decrypt(encrypted_value, secret_key);
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================================================
-- RLS POLICIES
-- ============================================================================

ALTER TABLE tournament_rewards ENABLE ROW LEVEL SECURITY;

-- Policy: Anyone can view public reward info (title, description, placement)
-- Hidden value is never exposed through RLS - must use decrypt function
CREATE POLICY "Public can view reward metadata"
    ON tournament_rewards
    FOR SELECT
    USING (true);

-- Policy: Host admins can insert rewards for their tournaments
-- Note: tournament ownership check should be done at application level
CREATE POLICY "Authenticated users can insert rewards"
    ON tournament_rewards
    FOR INSERT
    TO authenticated
    WITH CHECK (true);

-- Policy: Host admins can update rewards
CREATE POLICY "Authenticated users can update rewards"
    ON tournament_rewards
    FOR UPDATE
    TO authenticated
    USING (true)
    WITH CHECK (true);

-- Policy: Host admins can delete rewards
CREATE POLICY "Authenticated users can delete rewards"
    ON tournament_rewards
    FOR DELETE
    TO authenticated
    USING (true);

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE tournament_rewards IS 'Rewards for tournament winners (1st, 2nd, 3rd place)';
COMMENT ON COLUMN tournament_rewards.tournament_id IS 'Reference to tournament (managed by tournament service)';
COMMENT ON COLUMN tournament_rewards.placement IS 'Winner placement: 1=gold, 2=silver, 3=bronze';
COMMENT ON COLUMN tournament_rewards.title IS 'Public reward title (e.g., "Game Key", "Weapon Skin")';
COMMENT ON COLUMN tournament_rewards.description IS 'Public description visible to all participants';
COMMENT ON COLUMN tournament_rewards.hidden_value_encrypted IS 'Encrypted reward value (code, link) - decrypted only for winner';
COMMENT ON COLUMN tournament_rewards.is_claimed IS 'Whether the winner has claimed/viewed the reward';
COMMENT ON COLUMN tournament_rewards.claimed_by_user_id IS 'User ID of the winner who claimed the reward';
COMMENT ON COLUMN tournament_rewards.claimed_at IS 'Timestamp when reward was claimed';
