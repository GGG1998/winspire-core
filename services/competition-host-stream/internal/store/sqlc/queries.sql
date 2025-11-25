-- Cup host projections
-- name: UpsertCupHostView :exec
INSERT INTO cup_host_views (
    cup_id,
    competition_context_id,
    stage_statuses,
    attendance_counts,
    dependency_health,
    updated_at
) VALUES (
    sqlc.arg('cup_id'),
    sqlc.arg('competition_context_id'),
    sqlc.arg('stage_statuses'),
    sqlc.arg('attendance_counts'),
    sqlc.arg('dependency_health'),
    NOW()
)
ON CONFLICT (cup_id) DO UPDATE
SET stage_statuses     = EXCLUDED.stage_statuses,
    attendance_counts  = EXCLUDED.attendance_counts,
    dependency_health  = EXCLUDED.dependency_health,
    updated_at         = NOW();

-- Tournament host projections
-- name: UpsertTournamentHostView :exec
INSERT INTO tournament_host_views (
    tournament_id,
    cup_id,
    settings_hash,
    lineup_status,
    seeding_window,
    match_gate,
    updated_at
) VALUES (
    sqlc.arg('tournament_id'),
    sqlc.arg('cup_id'),
    sqlc.arg('settings_hash'),
    sqlc.arg('lineup_status'),
    sqlc.arg('seeding_window'),
    sqlc.arg('match_gate'),
    NOW()
)
ON CONFLICT (tournament_id) DO UPDATE
SET cup_id        = EXCLUDED.cup_id,
    settings_hash = EXCLUDED.settings_hash,
    lineup_status = EXCLUDED.lineup_status,
    seeding_window = EXCLUDED.seeding_window,
    match_gate    = EXCLUDED.match_gate,
    updated_at    = NOW();

-- Attendance snapshots
-- name: UpsertAttendanceSnapshot :exec
INSERT INTO attendance_snapshots (
    scope_type,
    scope_id,
    total_joined,
    total_confirmed,
    restrictions_breached,
    last_event_id,
    updated_at
) VALUES (
    sqlc.arg('scope_type'),
    sqlc.arg('scope_id'),
    sqlc.arg('total_joined'),
    sqlc.arg('total_confirmed'),
    sqlc.arg('restrictions_breached'),
    sqlc.arg('last_event_id'),
    NOW()
)
ON CONFLICT (scope_type, scope_id) DO UPDATE
SET total_joined          = EXCLUDED.total_joined,
    total_confirmed       = EXCLUDED.total_confirmed,
    restrictions_breached = EXCLUDED.restrictions_breached,
    last_event_id         = EXCLUDED.last_event_id,
    updated_at            = NOW();

-- Match lobby projections
-- name: UpsertMatchLobbyView :exec
INSERT INTO match_lobby_views (
    match_id,
    tournament_id,
    lobby_information,
    queue_state,
    updated_at
) VALUES (
    sqlc.arg('match_id'),
    sqlc.arg('tournament_id'),
    sqlc.arg('lobby_information'),
    sqlc.arg('queue_state'),
    NOW()
)
ON CONFLICT (match_id) DO UPDATE
SET tournament_id    = EXCLUDED.tournament_id,
    lobby_information = EXCLUDED.lobby_information,
    queue_state      = EXCLUDED.queue_state,
    updated_at       = NOW();

-- Host subscriptions
-- name: LeaseHostSubscription :exec
INSERT INTO host_subscriptions (
    subscription_id,
    host_id,
    scope_type,
    scope_id,
    last_delivered_event_id,
    leased_at
) VALUES (
    sqlc.arg('subscription_id'),
    sqlc.arg('host_id'),
    sqlc.arg('scope_type'),
    sqlc.arg('scope_id'),
    sqlc.arg('last_delivered_event_id'),
    NOW()
)
ON CONFLICT (host_id, scope_type, scope_id) DO UPDATE
SET subscription_id         = EXCLUDED.subscription_id,
    last_delivered_event_id = EXCLUDED.last_delivered_event_id,
    leased_at               = NOW();

-- name: ReleaseExpiredSubscriptions :exec
DELETE FROM host_subscriptions
WHERE scope_type = sqlc.arg('scope_type')
  AND scope_id = sqlc.arg('scope_id')
  AND leased_at < NOW() - sqlc.arg('ttl');

