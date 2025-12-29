package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgtypeconv "github.com/winspire/winspire-core/libs/go/pgtype"

	"github.com/winspire/tournament/internal/store/sqlc"
)

// HostRepository handles database operations for host authorization.
// It wraps the sqlc-generated Queries for type safety and business logic.
type HostRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewHostRepository creates a new HostRepository.
func NewHostRepository(pool *pgxpool.Pool) *HostRepository {
	return &HostRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// IsUserHostAdmin checks if a user is an admin/owner of the specified host.
// Returns true if the user has admin rights for the host.
func (r *HostRepository) IsUserHostAdmin(ctx context.Context, hostID, userID uuid.UUID) (bool, error) {
	isAdmin, err := r.queries.IsUserHostAdmin(ctx, sqlc.IsUserHostAdminParams{
		HostID: pgtypeconv.UUIDToPgtype(hostID),
		UserID: pgtypeconv.UUIDToPgtype(userID),
	})
	if err != nil {
		return false, fmt.Errorf("check host admin: %w", err)
	}
	return isAdmin, nil
}

// GetHostByID retrieves host information by ID.
// Returns error if host doesn't exist.
func (r *HostRepository) GetHostByID(ctx context.Context, hostID uuid.UUID) (*Host, error) {
	host, err := r.queries.GetHostByID(ctx, pgtypeconv.UUIDToPgtype(hostID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("host not found")
		}
		return nil, fmt.Errorf("get host by id: %w", err)
	}

	return sqlcHostToHost(host), nil
}

// GetOrCreateHostForUser gets an existing host for the user or creates a new one.
// If hostID matches userID, it will auto-create a personal host for that user.
// This enables seamless onboarding without requiring manual host creation.
func (r *HostRepository) GetOrCreateHostForUser(ctx context.Context, hostID, userID uuid.UUID) (*Host, error) {
	// First check if host exists
	host, err := r.GetHostByID(ctx, hostID)
	if err == nil {
		// Host exists, verify user has access
		isAdmin, err := r.IsUserHostAdmin(ctx, hostID, userID)
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
// This is called automatically when a user makes their first tournament-related request.
// The host ID is set to the user's ID for a 1:1 relationship.
func (r *HostRepository) createPersonalHost(ctx context.Context, userID uuid.UUID) (*Host, error) {
	// Use transaction for atomicity
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Create host with user's ID as host ID using CTE that creates both host and member
	sqlcHost, err := qtx.CreateHostWithOwner(ctx, sqlc.CreateHostWithOwnerParams{
		ID:          pgtypeconv.UUIDToPgtype(userID),
		Name:        "Personal Host",
		Description: pgtypeconv.StringToPgtype("Auto-created personal host for tournament management"),
		LogoUrl:     pgtypeconv.StringPtrToPgtype(nil), // NULL
		WebsiteUrl:  pgtypeconv.StringPtrToPgtype(nil), // NULL
		UserID:      pgtypeconv.UUIDToPgtype(userID),
	})
	if err != nil {
		return nil, fmt.Errorf("create host with owner: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return sqlcCreateHostWithOwnerRowToHost(sqlcHost), nil
}

// GetUserHosts returns all hosts where the user is a member.
func (r *HostRepository) GetUserHosts(ctx context.Context, userID uuid.UUID) ([]Host, error) {
	sqlcHosts, err := r.queries.GetUserHosts(ctx, pgtypeconv.UUIDToPgtype(userID))
	if err != nil {
		return nil, fmt.Errorf("query user hosts: %w", err)
	}

	hosts := make([]Host, 0, len(sqlcHosts))
	for _, h := range sqlcHosts {
		hosts = append(hosts, *sqlcHostToHost(h))
	}

	return hosts, nil
}

// Host represents a host entity for business logic.
type Host struct {
	ID          uuid.UUID
	Name        string
	Description *string
	LogoURL     *string
	WebsiteURL  *string
	CreatedAt   string
	UpdatedAt   string
}

// ============================================================================
// Type Converters - using shared converters from converters.go
// ============================================================================

// sqlcHostToHost converts sqlc.Host to repository.Host
func sqlcHostToHost(h sqlc.Host) *Host {
	return &Host{
		ID:          pgtypeconv.PgtypeToUUID(h.ID),
		Name:        h.Name,
		Description: pgtypeconv.PgtypeToStringPtr(h.Description),
		LogoURL:     pgtypeconv.PgtypeToStringPtr(h.LogoUrl),
		WebsiteURL:  pgtypeconv.PgtypeToStringPtr(h.WebsiteUrl),
		CreatedAt:   pgtypeconv.PgtypeTimestamptzToString(h.CreatedAt),
		UpdatedAt:   pgtypeconv.PgtypeTimestamptzToString(h.UpdatedAt),
	}
}

// sqlcCreateHostWithOwnerRowToHost converts sqlc.CreateHostWithOwnerRow to repository.Host
func sqlcCreateHostWithOwnerRowToHost(h sqlc.CreateHostWithOwnerRow) *Host {
	return &Host{
		ID:          pgtypeconv.PgtypeToUUID(h.ID),
		Name:        h.Name,
		Description: pgtypeconv.PgtypeToStringPtr(h.Description),
		LogoURL:     pgtypeconv.PgtypeToStringPtr(h.LogoUrl),
		WebsiteURL:  pgtypeconv.PgtypeToStringPtr(h.WebsiteUrl),
		CreatedAt:   pgtypeconv.PgtypeTimestamptzToString(h.CreatedAt),
		UpdatedAt:   pgtypeconv.PgtypeTimestamptzToString(h.UpdatedAt),
	}
}
