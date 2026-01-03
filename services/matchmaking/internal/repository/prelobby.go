package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/store/sqlc"
	pgtypeconv "github.com/winspire/winspire-core/libs/go/pgtype"
)

// UpdatePreLobbyParams contains optional fields for flexible pre-lobby updates
type UpdatePreLobbyParams struct {
	Status           *domain.PreLobbyStatus
	GracePeriodStart *time.Time
	GracePeriodEnd   *time.Time
	MinParticipants  *int
}

// PreLobbyRepository handles pre-lobby persistence
type PreLobbyRepository interface {
	// Pre-lobby CRUD
	Create(ctx context.Context, preLobby *domain.PreLobby) error
	GetByTournamentID(ctx context.Context, tournamentID uuid.UUID) (*domain.PreLobby, error)
	UpdateStatus(ctx context.Context, tournamentID uuid.UUID, status domain.PreLobbyStatus) (*domain.PreLobby, error)
	UpdateWithParams(ctx context.Context, tournamentID uuid.UUID, params UpdatePreLobbyParams) (*domain.PreLobby, error)
	StartGracePeriod(ctx context.Context, tournamentID uuid.UUID) (*domain.PreLobby, error)
	GetActiveGracePeriods(ctx context.Context) ([]*domain.PreLobby, error)
	Delete(ctx context.Context, tournamentID uuid.UUID) error

	// Participant snapshot
	CreateParticipantSnapshot(ctx context.Context, snapshot *domain.ParticipantSnapshot) error
	GetParticipantSnapshot(ctx context.Context, tournamentID uuid.UUID) (*domain.ParticipantSnapshot, error)
	ParticipantSnapshotExists(ctx context.Context, tournamentID uuid.UUID) (bool, error)

	// Activity feed
	AddActivityFeedEvent(ctx context.Context, item *domain.ActivityFeedItem) error
	GetRecentActivityFeed(ctx context.Context, tournamentID uuid.UUID) ([]*domain.ActivityFeedItem, error)
}

type preLobbyRepository struct {
	queries *sqlc.Queries
	db      sqlc.DBTX
	pool    *pgxpool.Pool
}

// NewPreLobbyRepository creates a new pre-lobby repository
func NewPreLobbyRepository(queries *sqlc.Queries, db sqlc.DBTX, pool *pgxpool.Pool) PreLobbyRepository {
	return &preLobbyRepository{
		queries: queries,
		db:      db,
		pool:    pool,
	}
}

// Create creates a new pre-lobby
func (r *preLobbyRepository) Create(ctx context.Context, preLobby *domain.PreLobby) error {
	result, err := r.queries.CreatePreLobby(ctx, sqlc.CreatePreLobbyParams{
		TournamentID:    pgtypeconv.UUIDToPgtype(preLobby.TournamentID),
		MinParticipants: int32(preLobby.MinParticipants),
	})
	if err != nil {
		return fmt.Errorf("create pre-lobby: %w", err)
	}

	// Update the domain object with generated values
	preLobby.ID = pgtypeconv.PgtypeToUUID(result.ID)
	preLobby.CreatedAt = result.CreatedAt
	preLobby.UpdatedAt = result.UpdatedAt

	return nil
}

// GetByTournamentID gets a pre-lobby by tournament ID
func (r *preLobbyRepository) GetByTournamentID(ctx context.Context, tournamentID uuid.UUID) (*domain.PreLobby, error) {
	result, err := r.queries.GetPreLobbyByTournament(ctx, pgtypeconv.UUIDToPgtype(tournamentID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get pre-lobby by tournament: %w", err)
	}

	return r.toDomain(result), nil
}

// UpdateStatus updates the pre-lobby status
func (r *preLobbyRepository) UpdateStatus(ctx context.Context, tournamentID uuid.UUID, status domain.PreLobbyStatus) (*domain.PreLobby, error) {
	result, err := r.queries.UpdatePreLobbyStatus(ctx, sqlc.UpdatePreLobbyStatusParams{
		TournamentID: pgtypeconv.UUIDToPgtype(tournamentID),
		Status:       string(status),
	})
	if err != nil {
		return nil, fmt.Errorf("update pre-lobby status: %w", err)
	}

	return r.toDomain(result), nil
}

// UpdateWithParams updates the pre-lobby with flexible optional parameters
func (r *preLobbyRepository) UpdateWithParams(ctx context.Context, tournamentID uuid.UUID, params UpdatePreLobbyParams) (*domain.PreLobby, error) {
	sqlcParams := sqlc.UpdatePreLobbyWithParamsParams{
		TournamentID: pgtypeconv.UUIDToPgtype(tournamentID),
	}

	if params.Status != nil {
		sqlcParams.Status = pgtype.Text{String: string(*params.Status), Valid: true}
	}
	if params.GracePeriodStart != nil {
		sqlcParams.GracePeriodStart = pgtype.Timestamp{Time: *params.GracePeriodStart, Valid: true}
	}
	if params.GracePeriodEnd != nil {
		sqlcParams.GracePeriodEnd = pgtype.Timestamp{Time: *params.GracePeriodEnd, Valid: true}
	}
	if params.MinParticipants != nil {
		sqlcParams.MinParticipants = pgtype.Int4{Int32: int32(*params.MinParticipants), Valid: true}
	}

	result, err := r.queries.UpdatePreLobbyWithParams(ctx, sqlcParams)
	if err != nil {
		return nil, fmt.Errorf("update pre-lobby with params: %w", err)
	}

	return r.toDomain(result), nil
}

// StartGracePeriod starts the grace period for a pre-lobby
func (r *preLobbyRepository) StartGracePeriod(ctx context.Context, tournamentID uuid.UUID) (*domain.PreLobby, error) {
	result, err := r.queries.StartGracePeriod(ctx, pgtypeconv.UUIDToPgtype(tournamentID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("pre-lobby not found or not in waiting status")
		}
		return nil, fmt.Errorf("start grace period: %w", err)
	}

	return r.toDomain(result), nil
}

// GetActiveGracePeriods gets all pre-lobbies currently in grace period
func (r *preLobbyRepository) GetActiveGracePeriods(ctx context.Context) ([]*domain.PreLobby, error) {
	results, err := r.queries.GetActiveGracePeriods(ctx)
	if err != nil {
		return nil, fmt.Errorf("get active grace periods: %w", err)
	}

	preLobbies := make([]*domain.PreLobby, len(results))
	for i, result := range results {
		preLobbies[i] = r.toDomain(result)
	}

	return preLobbies, nil
}

// Delete deletes a pre-lobby
func (r *preLobbyRepository) Delete(ctx context.Context, tournamentID uuid.UUID) error {
	err := r.queries.DeletePreLobby(ctx, pgtypeconv.UUIDToPgtype(tournamentID))
	if err != nil {
		return fmt.Errorf("delete pre-lobby: %w", err)
	}
	return nil
}

// CreateParticipantSnapshot creates an immutable participant snapshot
func (r *preLobbyRepository) CreateParticipantSnapshot(ctx context.Context, snapshot *domain.ParticipantSnapshot) error {
	// ParticipantsJSON implements driver.Valuer so pgx serializes it as JSONB automatically
	result, err := r.queries.CreateParticipantSnapshot(ctx, sqlc.CreateParticipantSnapshotParams{
		TournamentID:     pgtypeconv.UUIDToPgtype(snapshot.TournamentID),
		Participants:     domain.ParticipantsJSON(snapshot.Participants),
		ParticipantCount: int32(snapshot.ParticipantCount),
	})
	if err != nil {
		return fmt.Errorf("create participant snapshot: %w", err)
	}

	snapshot.ID = pgtypeconv.PgtypeToUUID(result.ID)
	snapshot.CreatedAt = result.CreatedAt

	return nil
}

// GetParticipantSnapshot gets the participant snapshot for a tournament
func (r *preLobbyRepository) GetParticipantSnapshot(ctx context.Context, tournamentID uuid.UUID) (*domain.ParticipantSnapshot, error) {
	result, err := r.queries.GetParticipantSnapshot(ctx, pgtypeconv.UUIDToPgtype(tournamentID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get participant snapshot: %w", err)
	}

	// ParticipantsJSON implements sql.Scanner so it's automatically deserialized
	return &domain.ParticipantSnapshot{
		ID:               pgtypeconv.PgtypeToUUID(result.ID),
		TournamentID:     pgtypeconv.PgtypeToUUID(result.TournamentID),
		Participants:     []domain.PreLobbyParticipant(result.Participants),
		ParticipantCount: int(result.ParticipantCount),
		CreatedAt:        result.CreatedAt,
	}, nil
}

// ParticipantSnapshotExists checks if a snapshot exists for a tournament
func (r *preLobbyRepository) ParticipantSnapshotExists(ctx context.Context, tournamentID uuid.UUID) (bool, error) {
	exists, err := r.queries.ParticipantSnapshotExists(ctx, pgtypeconv.UUIDToPgtype(tournamentID))
	if err != nil {
		return false, fmt.Errorf("check participant snapshot exists: %w", err)
	}
	return exists, nil
}

// AddActivityFeedEvent adds an event to the activity feed
func (r *preLobbyRepository) AddActivityFeedEvent(ctx context.Context, item *domain.ActivityFeedItem) error {
	result, err := r.queries.AddActivityFeedEvent(ctx, sqlc.AddActivityFeedEventParams{
		TournamentID:    pgtypeconv.UUIDToPgtype(item.TournamentID),
		EventType:       string(item.EventType),
		Message:         item.Message,
		ParticipantName: pgtypeconv.StringPtrToPgtype(item.ParticipantName),
	})
	if err != nil {
		return fmt.Errorf("add activity feed event: %w", err)
	}

	item.ID = pgtypeconv.PgtypeToUUID(result.ID)
	item.CreatedAt = result.CreatedAt

	return nil
}

// GetRecentActivityFeed gets the 20 most recent activity feed events
func (r *preLobbyRepository) GetRecentActivityFeed(ctx context.Context, tournamentID uuid.UUID) ([]*domain.ActivityFeedItem, error) {
	results, err := r.queries.GetRecentActivityFeed(ctx, pgtypeconv.UUIDToPgtype(tournamentID))
	if err != nil {
		return nil, fmt.Errorf("get recent activity feed: %w", err)
	}

	items := make([]*domain.ActivityFeedItem, len(results))
	for i, result := range results {
		items[i] = &domain.ActivityFeedItem{
			ID:              pgtypeconv.PgtypeToUUID(result.ID),
			TournamentID:    pgtypeconv.PgtypeToUUID(result.TournamentID),
			EventType:       domain.ActivityFeedEventType(result.EventType),
			Message:         result.Message,
			ParticipantName: pgtypeconv.PgtypeToStringPtr(result.ParticipantName),
			CreatedAt:       result.CreatedAt,
		}
	}

	return items, nil
}

// toDomain converts a SQLC model to a domain model
func (r *preLobbyRepository) toDomain(m sqlc.Prelobby) *domain.PreLobby {
	preLobby := &domain.PreLobby{
		ID:              pgtypeconv.PgtypeToUUID(m.ID),
		TournamentID:    pgtypeconv.PgtypeToUUID(m.TournamentID),
		Status:          domain.PreLobbyStatus(m.Status),
		MinParticipants: int(m.MinParticipants),
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}

	if m.GracePeriodStart.Valid {
		preLobby.GracePeriodStart = &m.GracePeriodStart.Time
	}
	if m.GracePeriodEnd.Valid {
		preLobby.GracePeriodEnd = &m.GracePeriodEnd.Time
	}

	return preLobby
}
