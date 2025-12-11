package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/application"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire/winspire-core/libs/go/httpx"
)

// GetCurrentUserMatch returns the latest active match for the authenticated user
// GET /v1/matchmaking/me
func (h *MatchHandler) GetCurrentUserMatch(c *gin.Context) {
	user := httpx.MustGetUser(c)
	userID, err := uuid.Parse(string(user.ID))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID", "details": err.Error()})
		return
	}

	match, round, tournamentID, err := h.matchService.GetCurrentMatchForUser(c.Request.Context(), userID)
	if err != nil {
		// No active match found for the user
		c.JSON(http.StatusOK, gin.H{"match": nil})
		return
	}

	response := gin.H{
		"matchId":      match.ID,
		"tournamentId": tournamentID,
		"status":       match.Status,
		"round":        round.RoundNumber,
		"table":        match.MatchNumber,
		"startedAt":    match.StartedAt,
		"updatedAt":    match.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"match": response,
	})
}

// compile-time assertion that MatchService has required method
var _ interface {
	GetCurrentMatchForUser(ctx context.Context, userID uuid.UUID) (*domain.Match, *domain.Round, uuid.UUID, error)
} = (*application.MatchService)(nil)


