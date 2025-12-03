-- Migration: Create hosts and host_members tables
-- Purpose: Enable host authorization and tournament organization
-- Date: 2024-11-28

-- ============================================================================
-- HOSTS TABLE
-- ============================================================================
-- Represents organizations or individuals that can host tournaments

CREATE TABLE IF NOT EXISTS hosts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    logo_url VARCHAR(512),
    website_url VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for name lookups
CREATE INDEX idx_hosts_name ON hosts(name);

-- ============================================================================
-- HOST MEMBERS TABLE
-- ============================================================================
-- Manages user permissions for hosts (authorization)
-- Roles: 'owner', 'admin', 'member'

CREATE TABLE IF NOT EXISTS host_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id UUID NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(host_id, user_id)
);

-- Indexes for performance
CREATE INDEX idx_host_members_host_id ON host_members(host_id);
CREATE INDEX idx_host_members_user_id ON host_members(user_id);
CREATE INDEX idx_host_members_role ON host_members(role);

-- Composite index for authorization checks
CREATE INDEX idx_host_members_host_user ON host_members(host_id, user_id);

-- ============================================================================
-- TRIGGER: Update updated_at timestamp
-- ============================================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for hosts table
CREATE TRIGGER update_hosts_updated_at
    BEFORE UPDATE ON hosts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger for host_members table
CREATE TRIGGER update_host_members_updated_at
    BEFORE UPDATE ON host_members
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE hosts IS 'Organizations or individuals that can host tournaments';
COMMENT ON COLUMN hosts.id IS 'Unique identifier for the host';
COMMENT ON COLUMN hosts.name IS 'Display name of the host organization';
COMMENT ON COLUMN hosts.description IS 'Optional description of the host';
COMMENT ON COLUMN hosts.logo_url IS 'URL to host logo/avatar';
COMMENT ON COLUMN hosts.website_url IS 'Host website URL';

COMMENT ON TABLE host_members IS 'User permissions and roles for hosts';
COMMENT ON COLUMN host_members.host_id IS 'Reference to the host';
COMMENT ON COLUMN host_members.user_id IS 'Reference to user (from auth system)';
COMMENT ON COLUMN host_members.role IS 'User role: owner, admin, or member';

