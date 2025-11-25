# Phase 1 — Data Model

The Host Streaming service maintains slim relational projections sourced from Cup, Tournament, Participation, and Match domains. Every table is append-only or idempotently upserted via sqlc so application pods remain stateless; streaming cursors are calculated per request.

## Entities & Fields

### `cup_host_views`
| Field | Type | Notes |
|-------|------|-------|
| `cup_id` | UUID (PK, FK → cups.id) | Inherited from Cup Management |
| `competition_context_id` | UUID | Shared lineage identifier |
| `stage_statuses` | JSONB | Array of `{stageId,status,startAt,endAt}` |
| `attendance_counts` | JSONB | `{total, confirmed, waitlisted}` derived from `OwnCupParticipation` |
| `dependency_health` | JSONB | Supplier readiness flags (`tournamentLifecycle`, `matchLifecycle`, etc.) |
| `updated_at` | TIMESTAMPTZ | For conditional GET/SSE replay |

Validation: stage metadata must mirror blueprint definitions; dependency health enumerations align with FR-004 relationship types.

### `tournament_host_views`
| Field | Type | Notes |
|-------|------|-------|
| `tournament_id` | UUID (PK) |
| `cup_id` | UUID (nullable, when standalone tournaments exist) |
| `settings_hash` | TEXT | Hash of `TournamentSettingsGroup` to detect drift |
| `lineup_status` | JSONB | Per-lineup `{lineupId,confirmed,allowedActions}` |
| `seeding_window` | TSRANGE | Derived from `TournamentSchedule` |
| `match_gate` | JSONB | `{readyForMatch,blockedReason}` |
| `updated_at` | TIMESTAMPTZ |

Validation: `lineup_status.allowedActions` must be subset of Participation API enumerations; `match_gate.readyForMatch` flips only when Participation + Match both report readiness.

### `attendance_snapshots`
| Field | Type | Notes |
|-------|------|-------|
| `scope_type` | ENUM(`cup`,`tournament`) |
| `scope_id` | UUID |
| `total_joined` | INT |
| `total_confirmed` | INT |
| `restrictions_breached` | JSONB | List of failing `CompetitionRestriction` ids |
| `last_event_id` | BIGINT | Monotonic cursor for SSE replay |
| `updated_at` | TIMESTAMPTZ |

Validation: `total_confirmed <= total_joined`; restrictions array cannot exceed 50 items to keep payloads light for SSE.

### `match_lobby_views`
| Field | Type | Notes |
|-------|------|-------|
| `match_id` | UUID |
| `tournament_id` | UUID |
| `lobby_information` | JSONB | Mirrors `LobbyInformation` VO (`maximumLobbyMinutes`, `gameSessionTag`, etc.) |
| `queue_state` | JSONB | `{offerId,status,queueType}` from Matchmaking queue |
| `updated_at` | TIMESTAMPTZ |

Validation: `lobby_information.maximumLobbyMinutes` enforced ≥ `maximumGoToGameMinutes`; `queue_state.status` limited to OPEN/ACCEPTED/PLAYING/COMPLETE.

### `host_subscriptions`
| Field | Type | Notes |
|-------|------|-------|
| `subscription_id` | UUID (PK) |
| `host_id` | UUID | Auth subject |
| `scope_type` | ENUM(`cup`,`tournament`,`match`) |
| `scope_id` | UUID |
| `last_delivered_event_id` | BIGINT |
| `leased_at` | TIMESTAMPTZ | TTL-limited lease for SSE pooling |

Validation: TTL ≤ 90 s to enforce stateless reconnects; unique constraint on (`host_id`,`scope_type`,`scope_id`) keeps a single cursor per host view.

## Relationships

- `cup_host_views.cup_id` → `tournament_host_views.cup_id` (1:N) ensures Cup dashboards enumerate derived tournaments.
- `tournament_host_views.tournament_id` → `match_lobby_views.tournament_id` (1:N) to drill from tournament sheet to match console.
- `attendance_snapshots.scope_*` joined to either cup or tournament view depending on `scope_type`.
- `host_subscriptions` reference any of the above scopes, enabling SSE fan-out keyed by (`scope_type`,`scope_id`).

## State Transitions

1. **Cup blueprint change**: Cup Management mutation emits event → projection worker upserts `cup_host_views` + resets dependent tournament hashes.
2. **Tournament lineup change**: Participation event updates `tournament_host_views.lineup_status` and `attendance_snapshots`.
3. **Match queue change**: Matchmaking event alters `match_lobby_views.queue_state`; if tournament gate flips, SSE notifies host subscribers.
4. **SSE delivery**: When host connects, `host_subscriptions` is leased; events with `event_id > last_delivered_event_id` stream to client. Lease expiry or disconnect simply clears row—no sticky state in pods.




