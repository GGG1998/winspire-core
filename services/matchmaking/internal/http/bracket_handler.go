// Package http provides HTTP handlers for the matchmaking service
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/repository"
)

// BracketHandler handles bracket-related HTTP requests
type BracketHandler struct {
	bracketRepo repository.BracketRepository
	roundRepo   repository.RoundRepository
	matchRepo   repository.MatchRepository
}

// NewBracketHandler creates a new bracket handler
func NewBracketHandler(
	bracketRepo repository.BracketRepository,
	roundRepo repository.RoundRepository,
	matchRepo repository.MatchRepository,
) *BracketHandler {
	return &BracketHandler{
		bracketRepo: bracketRepo,
		roundRepo:   roundRepo,
		matchRepo:   matchRepo,
	}
}

// GetBracket retrieves a bracket by its ID with all rounds and matches
// GET /v1/brackets/:id
func (h *BracketHandler) GetBracket(c *gin.Context) {
	// Parse bracket ID
	bracketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bracket ID", "details": err.Error()})
		return
	}

	// First get the bracket to obtain its tournament ID
	bracketInfo, err := h.bracketRepo.GetByID(c.Request.Context(), bracketID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bracket not found", "details": err.Error()})
		return
	}

	// Fetch bracket with rounds and matches using tournament ID
	bracket, rounds, matches, err := h.bracketRepo.GetWithRoundsAndMatches(c.Request.Context(), bracketInfo.TournamentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bracket not found", "details": err.Error()})
		return
	}

	// Organize matches by round for response
	roundMap := make(map[uuid.UUID][]interface{})
	for _, match := range matches {
		matchData := map[string]interface{}{
			"id":                 match.ID,
			"match_number":       match.MatchNumber,
			"participant1_id":    match.Participant1ID,
			"participant2_id":    match.Participant2ID,
			"next_match_id":      match.NextMatchID,
			"status":             match.Status,
			"participant1_ready": match.Participant1Ready,
			"participant2_ready": match.Participant2Ready,
			"winner_id":          match.WinnerID,
			"score_player1":      match.ScorePlayer1,
			"score_player2":      match.ScorePlayer2,
			"started_at":         match.StartedAt,
			"completed_at":       match.CompletedAt,
		}
		roundMap[match.RoundID] = append(roundMap[match.RoundID], matchData)
	}

	// Build rounds response with matches
	roundsData := make([]interface{}, len(rounds))
	for i, round := range rounds {
		roundsData[i] = map[string]interface{}{
			"id":            round.ID,
			"round_number":  round.RoundNumber,
			"round_name":    round.RoundName,
			"matches_count": round.MatchesCount,
			"status":        round.Status,
			"matches":       roundMap[round.ID],
		}
	}

	// Build final response
	bracketData := map[string]interface{}{
		"id":            bracket.ID,
		"tournament_id": bracket.TournamentID,
		"total_rounds":  bracket.TotalRounds,
		"total_matches": bracket.TotalMatches,
		"byes_count":    bracket.ByesCount,
		"generated_at":  bracket.GeneratedAt,
		"rounds":        roundsData,
	}

	c.JSON(http.StatusOK, gin.H{"bracket": bracketData})
}

// GetBracketByTournament retrieves a bracket by tournament ID
// GET /v1/tournaments/:id/bracket
func (h *BracketHandler) GetBracketByTournament(c *gin.Context) {
	// Parse tournament ID
	tournamentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tournament ID", "details": err.Error()})
		return
	}

	// Get full bracket details with rounds and matches by tournament ID
	fullBracket, rounds, matches, err := h.bracketRepo.GetWithRoundsAndMatches(c.Request.Context(), tournamentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bracket not found for tournament", "details": err.Error()})
		return
	}

	// Organize matches by round
	roundMap := make(map[uuid.UUID][]interface{})
	for _, match := range matches {
		matchData := map[string]interface{}{
			"id":                 match.ID,
			"match_number":       match.MatchNumber,
			"participant1_id":    match.Participant1ID,
			"participant2_id":    match.Participant2ID,
			"next_match_id":      match.NextMatchID,
			"status":             match.Status,
			"participant1_ready": match.Participant1Ready,
			"participant2_ready": match.Participant2Ready,
			"winner_id":          match.WinnerID,
			"score_player1":      match.ScorePlayer1,
			"score_player2":      match.ScorePlayer2,
			"started_at":         match.StartedAt,
			"completed_at":       match.CompletedAt,
		}
		roundMap[match.RoundID] = append(roundMap[match.RoundID], matchData)
	}

	// Build rounds response
	roundsData := make([]interface{}, len(rounds))
	for i, round := range rounds {
		roundsData[i] = map[string]interface{}{
			"id":            round.ID,
			"round_number":  round.RoundNumber,
			"round_name":    round.RoundName,
			"matches_count": round.MatchesCount,
			"status":        round.Status,
			"matches":       roundMap[round.ID],
		}
	}

	// Build final response
	bracketData := map[string]interface{}{
		"id":            fullBracket.ID,
		"tournament_id": fullBracket.TournamentID,
		"total_rounds":  fullBracket.TotalRounds,
		"total_matches": fullBracket.TotalMatches,
		"byes_count":    fullBracket.ByesCount,
		"generated_at":  fullBracket.GeneratedAt,
		"rounds":        roundsData,
	}

	c.JSON(http.StatusOK, gin.H{"bracket": bracketData})
}
