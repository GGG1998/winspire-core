package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/winspire-core/services/matchmaking/internal/games"
)

// GameDeps contains dependencies for game handlers.
// Note: Public game listing endpoints have been removed.
// Frontend now fetches games directly from Supabase.
type GameDeps struct {
	Repo *games.GameRepository
}

// RegisterGameRoutes registers public game routes.
// Note: Public game listing endpoints (GET /games, GET /games/:gameId, GET /g/:slug)
// have been removed. Frontend fetches games directly from Supabase.
// Only bundle streaming routes remain for serving game assets from S3.
func RegisterGameRoutes(_ *gin.RouterGroup, _ GameDeps) {
	// No public game routes - frontend uses Supabase directly
	// Bundle streaming is handled by RegisterBundleRoutes
	// Admin operations are handled by RegisterAdminRoutes
}

// gameToResponse converts a Game domain model to API response.
// Used by admin handlers.
func gameToResponse(game games.Game) GameResponse {
	resp := GameResponse{
		ID:          game.ID.String(),
		Slug:        game.Slug,
		Name:        game.Name,
		Version:     game.Version,
		StoragePath: game.StoragePath,
		IsActive:    game.IsActive,
		CreatedAt:   game.CreatedAt,
		UpdatedAt:   game.UpdatedAt,
	}

	if game.GameIntegrationID != nil {
		id := game.GameIntegrationID.String()
		resp.GameIntegrationID = &id
	}
	if game.Description != nil {
		resp.Description = game.Description
	}
	if game.LogoURL != nil {
		resp.LogoURL = game.LogoURL
	}

	return resp
}
