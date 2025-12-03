-- Migration to update existing participant display_name from UUID to a placeholder
-- This fixes records created before the DisplayName fix was implemented
-- Note: We can't access user_profiles/streamer_profiles from this database,
-- so we'll update to a placeholder. Real nicknames will be set on next participation confirmation.

-- Update display_name for records that currently have UUID values
UPDATE tournament_registrations
SET display_name = 'Player-' || SUBSTRING(user_id::text FROM 1 FOR 8),
    updated_at = NOW()
WHERE display_name ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'; -- Only update if current value is UUID

