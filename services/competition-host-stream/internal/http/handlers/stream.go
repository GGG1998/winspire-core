package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	ssebroker "github.com/winspire/competition-host-stream/internal/sse"
)

// StreamDeps contains dependencies for SSE streaming.
type StreamDeps struct {
	Broker   *ssebroker.Broker
	Registry *ssebroker.Registry
}

// RegisterStreamRoute attaches the SSE handler to the API group.
func RegisterStreamRoute(group *gin.RouterGroup, deps StreamDeps) {
	group.GET("/hosts/:hostId/streams/:scopeType/:scopeId", func(c *gin.Context) {
		hostID, ok := parseUUIDParam(c, "hostId")
		if !ok {
			return
		}
		scopeType := c.Param("scopeType")
		if scopeType != "cup" && scopeType != "tournament" && scopeType != "match" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":     "invalid scopeType",
				"scopeType": scopeType,
			})
			return
		}
		scopeID, ok := parseUUIDParam(c, "scopeId")
		if !ok {
			return
		}

		lastEventID := parseLastEventID(c)

		if err := deps.Registry.ReleaseExpired(c.Request.Context()); err != nil {
			c.Error(err)
			return
		}
		if _, err := deps.Registry.Lease(c.Request.Context(), hostID, scopeType, scopeID, lastEventID); err != nil {
			c.Error(err)
			return
		}

		stream := ssebroker.Scope{Type: scopeType, ID: scopeID}
		deps.Broker.Server().CreateStream(stream.Key())
		q := c.Request.URL.Query()
		q.Set("stream", stream.Key())
		if lastEventID > 0 {
			q.Set("lastEventId", strconv.FormatInt(lastEventID, 10))
		}
		c.Request.URL.RawQuery = q.Encode()
		deps.Broker.Server().ServeHTTP(c.Writer, c.Request)
	})
}

func parseLastEventID(c *gin.Context) int64 {
	if raw := c.GetHeader("Last-Event-ID"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
			return parsed
		}
	}
	if raw := c.Query("lastEventId"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
			return parsed
		}
	}
	return 0
}
