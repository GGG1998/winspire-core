package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	httpxtypes "github.com/winspire/winspire-core/libs/go/auth/types"
)

type mockMatchService struct {
	match        *domain.Match
	round        *domain.Round
	tournamentID uuid.UUID
	err          error
}

func (m *mockMatchService) GetCurrentMatchForUser(_ context.Context, _ uuid.UUID) (*domain.Match, *domain.Round, uuid.UUID, error) {
	return m.match, m.round, m.tournamentID, m.err
}

func TestGetCurrentUserMatch_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	roundID := uuid.New()
	tournamentID := uuid.New()
	now := time.Now()

	mockService := &mockMatchService{
		match: &domain.Match{
			ID:          uuid.New(),
			RoundID:     roundID,
			MatchNumber: 2,
			Status:      domain.MatchStatusStarted,
			StartedAt:   &now,
			UpdatedAt:   now,
		},
		round: &domain.Round{
			ID:          roundID,
			RoundNumber: 3,
		},
		tournamentID: tournamentID,
	}

	handler := &MatchHandler{
		matchService: mockService,
	}

	router := gin.New()
	router.GET("/v1/matchmaking/me", func(c *gin.Context) {
		c.Set("user", &httpxtypes.UserContext{ID: httpxtypes.UserID(userID.String())})
		handler.GetCurrentUserMatch(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/matchmaking/me", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), `"matchId"`)
	assert.Contains(t, resp.Body.String(), `"tournamentId"`)
	assert.Contains(t, resp.Body.String(), `"round":3`)
	assert.Contains(t, resp.Body.String(), `"table":2`)
}

func TestGetCurrentUserMatch_NoMatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	mockService := &mockMatchService{
		err: assert.AnError,
	}

	handler := &MatchHandler{
		matchService: mockService,
	}

	router := gin.New()
	router.GET("/v1/matchmaking/me", func(c *gin.Context) {
		c.Set("user", &httpxtypes.UserContext{ID: httpxtypes.UserID(userID.String())})
		handler.GetCurrentUserMatch(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/matchmaking/me", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), `"match":null`)
}


