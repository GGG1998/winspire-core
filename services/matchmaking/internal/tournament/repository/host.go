package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgtypeconv "github.com/winspire/winspire-core/libs/go/pgtype"

	"github.com/winspire-core/services/matchmaking/internal/store/sqlc"
	"github.com/winspire-core/services/matchmaking/internal/tournament/domain"
)

// HostRepository wraps sqlc-generated queries for hosts.
type HostRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// Errors
var (
	ErrHostNotFound = errors.New("host not found")
)

// NewHostRepository creates a new HostRepository.
func NewHostRepository(pool *pgxpool.Pool) *HostRepository {
	return &HostRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// GetByID retrieves a host by ID.
func (r *HostRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Host, error) {
	row, err := r.queries.GetTournamentHostByID(ctx, pgtypeconv.UUIDToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHostNotFound
		}
		return nil, fmt.Errorf("get host: %w", err)
	}

	return sqlcHostToDomain(row), nil
}

// IsUserAdmin checks if a user is an admin of the host.
func (r *HostRepository) IsUserAdmin(ctx context.Context, hostID, userID uuid.UUID) (bool, error) {
	params := sqlc.IsTournamentUserHostAdminParams{
		HostID: pgtypeconv.UUIDToPgtype(hostID),
		UserID: pgtypeconv.UUIDToPgtype(userID),
	}

	isAdmin, err := r.queries.IsTournamentUserHostAdmin(ctx, params)
	if err != nil {
		return false, fmt.Errorf("check host admin: %w", err)
	}
	return isAdmin, nil
}

// GetUserRole retrieves the user's role for a host.
func (r *HostRepository) GetUserRole(ctx context.Context, hostID, userID uuid.UUID) (string, error) {
	params := sqlc.GetTournamentUserHostRoleParams{
		HostID: pgtypeconv.UUIDToPgtype(hostID),
		UserID: pgtypeconv.UUIDToPgtype(userID),
	}

	role, err := r.queries.GetTournamentUserHostRole(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("user is not a member of this host")
		}
		return "", fmt.Errorf("get user host role: %w", err)
	}
	return role, nil
}

// GetUserHosts retrieves all hosts the user is a member of.
func (r *HostRepository) GetUserHosts(ctx context.Context, userID uuid.UUID) ([]*domain.Host, error) {
	rows, err := r.queries.GetTournamentUserHosts(ctx, pgtypeconv.UUIDToPgtype(userID))
	if err != nil {
		return nil, fmt.Errorf("get user hosts: %w", err)
	}

	hosts := make([]*domain.Host, len(rows))
	for i, row := range rows {
		hosts[i] = sqlcHostToDomain(row)
	}

	return hosts, nil
}

// GetOrCreateHostForUser gets an existing host for the user or creates a new one.
// If hostID matches userID, it will auto-create a personal host for that user.
func (r *HostRepository) GetOrCreateHostForUser(ctx context.Context, hostID, userID uuid.UUID) (*domain.Host, error) {
	// First check if host exists
	host, err := r.GetByID(ctx, hostID)
	if err == nil {
		// Host exists, verify user has access
		isAdmin, err := r.IsUserAdmin(ctx, hostID, userID)
		if err != nil {
			return nil, fmt.Errorf("check admin status: %w", err)
		}
		if !isAdmin {
			return nil, fmt.Errorf("user is not admin of this host")
		}
		return host, nil
	}

	// Host doesn't exist - check if we should auto-create
	// Only auto-create if hostID equals userID (personal host)
	if hostID != userID {
		return nil, fmt.Errorf("host does not exist and cannot be auto-created")
	}

	// Auto-create personal host for user
	return r.createPersonalHost(ctx, userID)
}

// createPersonalHost creates a personal host for a user and makes them owner.
// The host ID is set to the userID so that hostID == userID for personal hosts.
func (r *HostRepository) createPersonalHost(ctx context.Context, userID uuid.UUID) (*domain.Host, error) {
	params := sqlc.CreateTournamentHostWithIDAndOwnerParams{
		ID:          pgtypeconv.UUIDToPgtype(userID), // Use userID as host ID
		Name:        "Personal Host",
		Description: pgtypeconv.StringToPgtype("Auto-created personal host for tournament management"),
		LogoUrl:     pgtypeconv.StringPtrToPgtype(nil),
		WebsiteUrl:  pgtypeconv.StringPtrToPgtype(nil),
		UserID:      pgtypeconv.UUIDToPgtype(userID),
	}

	row, err := r.queries.CreateTournamentHostWithIDAndOwner(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create host with owner: %w", err)
	}

	return &domain.Host{
		ID:          pgtypeconv.PgtypeToUUID(row.ID),
		Name:        row.Name,
		Description: pgtypeconv.PgtypeToStringPtr(row.Description),
		LogoURL:     pgtypeconv.PgtypeToStringPtr(row.LogoUrl),
		WebsiteURL:  pgtypeconv.PgtypeToStringPtr(row.WebsiteUrl),
	}, nil
}

// Create creates a new host.
func (r *HostRepository) Create(ctx context.Context, name, description string) (*domain.Host, error) {
	params := sqlc.CreateTournamentHostParams{
		Name:        name,
		Description: pgtypeconv.StringToPgtype(description),
		LogoUrl:     pgtypeconv.StringPtrToPgtype(nil),
		WebsiteUrl:  pgtypeconv.StringPtrToPgtype(nil),
	}

	row, err := r.queries.CreateTournamentHost(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create host: %w", err)
	}

	return sqlcHostToDomain(row), nil
}

// sqlcHostToDomain converts SQLC host to domain host.
func sqlcHostToDomain(h sqlc.TournamentHost) *domain.Host {
	return &domain.Host{
		ID:          pgtypeconv.PgtypeToUUID(h.ID),
		Name:        h.Name,
		Description: pgtypeconv.PgtypeToStringPtr(h.Description),
		LogoURL:     pgtypeconv.PgtypeToStringPtr(h.LogoUrl),
		WebsiteURL:  pgtypeconv.PgtypeToStringPtr(h.WebsiteUrl),
	}
}
