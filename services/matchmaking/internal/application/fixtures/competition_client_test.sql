-- Fixtures for CompetitionClient integration tests
-- These fixtures are used to test the GetTournamentInfo and IsUserRegistered functions

-- Create default host (same as competition service migration 000002)
INSERT INTO hosts (id, name, description)
VALUES (
    '00000000-0000-0000-0000-000000000001'::UUID,
    'Default Tournament Host',
    'Default host for development and testing purposes'
)
ON CONFLICT (id) DO NOTHING;

-- Tournament with all fields populated
INSERT INTO tournaments (
    id, host_id, name, description, status,
    scheduled_start_time_at, minimum_team_count, game_id
) VALUES (
    '11111111-1111-1111-1111-111111111111'::UUID,
    '00000000-0000-0000-0000-000000000001'::UUID,
    'Test Tournament',
    'A test tournament for integration tests',
    'registration_open',
    NOW() + INTERVAL '24 hours',
    8,
    '22222222-2222-2222-2222-222222222222'::UUID
);

-- Tournament without game ID
INSERT INTO tournaments (
    id, host_id, name, description, status,
    scheduled_start_time_at, minimum_team_count, game_id
) VALUES (
    '33333333-3333-3333-3333-333333333333'::UUID,
    '00000000-0000-0000-0000-000000000001'::UUID,
    'No Game Tournament',
    NULL,
    'draft',
    NULL,
    4,
    NULL
);

-- Tournament with min participants less than 2 (to test default behavior)
INSERT INTO tournaments (
    id, host_id, name, description, status,
    scheduled_start_time_at, minimum_team_count, game_id
) VALUES (
    '44444444-4444-4444-4444-444444444444'::UUID,
    '00000000-0000-0000-0000-000000000001'::UUID,
    'Small Tournament',
    NULL,
    'started',
    NULL,
    1, -- Less than 2, should default to 2 in client
    NULL
);

-- Tournament for registration tests
INSERT INTO tournaments (
    id, host_id, name, description, status,
    scheduled_start_time_at, minimum_team_count
) VALUES (
    '55555555-5555-5555-5555-555555555555'::UUID,
    '00000000-0000-0000-0000-000000000001'::UUID,
    'Registration Test Tournament',
    'Tournament for testing user registration checks',
    'registration_open',
    NOW() + INTERVAL '48 hours',
    8
);

-- Registration for a test user
INSERT INTO tournament_registrations (
    tournament_id, user_id, status, display_name, avatar_url
) VALUES (
    '55555555-5555-5555-5555-555555555555'::UUID,
    '66666666-6666-6666-6666-666666666666'::UUID,
    'confirmed',
    'Test Player',
    'https://example.com/avatar.png'
);
