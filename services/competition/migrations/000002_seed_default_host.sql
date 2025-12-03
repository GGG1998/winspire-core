-- Migration: Seed default host for development
-- Purpose: Create a default host and helper function to auto-register users as host admins
-- Date: 2024-11-28

-- ============================================================================
-- CREATE DEFAULT HOST
-- ============================================================================
-- This is useful for development/testing
-- In production, hosts should be created through proper API endpoints

INSERT INTO hosts (id, name, description)
VALUES (
    '00000000-0000-0000-0000-000000000001'::UUID,
    'Default Tournament Host',
    'Default host for development and testing purposes'
)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- HELPER FUNCTION: Auto-register user as host admin
-- ============================================================================
-- This function can be called to automatically register a user as admin of default host
-- Useful for development: SELECT ensure_user_is_host_admin('user-uuid-here');

CREATE OR REPLACE FUNCTION ensure_user_is_host_admin(
    p_user_id UUID,
    p_host_id UUID DEFAULT '00000000-0000-0000-0000-000000000001'::UUID
)
RETURNS VOID AS $$
BEGIN
    -- Insert user as admin if not already a member
    INSERT INTO host_members (host_id, user_id, role)
    VALUES (p_host_id, p_user_id, 'admin')
    ON CONFLICT (host_id, user_id) 
    DO UPDATE SET 
        role = CASE 
            WHEN host_members.role = 'member' THEN 'admin'::VARCHAR
            ELSE host_members.role
        END,
        updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION ensure_user_is_host_admin IS 'Helper function to ensure a user is registered as host admin (useful for dev/testing)';

-- ============================================================================
-- HELPER FUNCTION: Create host for user
-- ============================================================================
-- Creates a new host and makes the user an owner

CREATE OR REPLACE FUNCTION create_host_for_user(
    p_user_id UUID,
    p_host_name VARCHAR(255),
    p_description TEXT DEFAULT NULL
)
RETURNS UUID AS $$
DECLARE
    v_host_id UUID;
BEGIN
    -- Create the host
    INSERT INTO hosts (name, description)
    VALUES (p_host_name, p_description)
    RETURNING id INTO v_host_id;
    
    -- Make user the owner
    INSERT INTO host_members (host_id, user_id, role)
    VALUES (v_host_id, p_user_id, 'owner');
    
    RETURN v_host_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_host_for_user IS 'Creates a new host and makes the specified user an owner';

