-- +migrate Up
CREATE TABLE IF NOT EXISTS cup_host_views (
    cup_id UUID PRIMARY KEY,
    competition_context_id UUID NOT NULL,
    stage_statuses JSONB NOT NULL DEFAULT '[]'::jsonb,
    attendance_counts JSONB NOT NULL DEFAULT '{}'::jsonb,
    dependency_health JSONB NOT NULL DEFAULT '[]'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tournament_host_views (
    tournament_id UUID PRIMARY KEY,
    cup_id UUID,
    settings_hash TEXT NOT NULL,
    lineup_status JSONB NOT NULL DEFAULT '[]'::jsonb,
    seeding_window TSRANGE,
    match_gate JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS attendance_snapshots (
    scope_type TEXT NOT NULL,
    scope_id UUID NOT NULL,
    total_joined INTEGER NOT NULL DEFAULT 0,
    total_confirmed INTEGER NOT NULL DEFAULT 0,
    restrictions_breached JSONB NOT NULL DEFAULT '[]'::jsonb,
    last_event_id BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (scope_type, scope_id)
);

CREATE TABLE IF NOT EXISTS match_lobby_views (
    match_id UUID PRIMARY KEY,
    tournament_id UUID NOT NULL,
    lobby_information JSONB NOT NULL DEFAULT '{}'::jsonb,
    queue_state JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS host_subscriptions (
    subscription_id UUID PRIMARY KEY,
    host_id UUID NOT NULL,
    scope_type TEXT NOT NULL,
    scope_id UUID NOT NULL,
    last_delivered_event_id BIGINT NOT NULL DEFAULT 0,
    leased_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (host_id, scope_type, scope_id)
);

CREATE INDEX IF NOT EXISTS idx_tournament_host_views_cup_id ON tournament_host_views (cup_id);
CREATE INDEX IF NOT EXISTS idx_match_lobby_views_tournament_id ON match_lobby_views (tournament_id);
CREATE INDEX IF NOT EXISTS idx_attendance_snapshots_event ON attendance_snapshots (last_event_id DESC);

-- +migrate Down
DROP TABLE IF EXISTS host_subscriptions;
DROP TABLE IF EXISTS match_lobby_views;
DROP TABLE IF EXISTS attendance_snapshots;
DROP TABLE IF EXISTS tournament_host_views;
DROP TABLE IF EXISTS cup_host_views;




