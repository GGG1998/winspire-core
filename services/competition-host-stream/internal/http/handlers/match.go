package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/winspire/competition-host-stream/internal/projections"
)

// MatchDeps dependencies for match routes.
type MatchDeps struct {
	Reader *projections.Reader
}

// RegisterMatchRoutes wires tournament match list endpoint.
func RegisterMatchRoutes(group *gin.RouterGroup, deps MatchDeps) {
	group.GET("/hosts/:hostId/tournaments/:tournamentId/matches", func(c *gin.Context) {
		if _, ok := parseUUIDParam(c, "hostId"); !ok {
			return
		}
		tournamentID, ok := parseUUIDParam(c, "tournamentId")
		if !ok {
			return
		}
		result, err := deps.Reader.ListMatches(c.Request.Context(), tournamentID)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, result)
	})
}
