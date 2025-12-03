package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	val := c.Param(name)
	id, err := uuid.Parse(val)
	if err != nil {
		c.Error(fmt.Errorf("invalid %s: %w", name, err))
		return uuid.UUID{}, false
	}
	return id, true
}

func defaultJSON(raw []byte, fallback []byte) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(fallback)
	}
	return json.RawMessage(raw)
}

func marshalOrError(c *gin.Context, v any, fallback []byte) (json.RawMessage, bool) {
	if v == nil {
		return json.RawMessage(fallback), true
	}
	buf, err := json.Marshal(v)
	if err != nil {
		c.Error(err)
		return nil, false
	}
	if len(buf) == 0 && len(fallback) > 0 {
		buf = fallback
	}
	return json.RawMessage(buf), true
}

func encodeSeedingWindow(w *SeedingWindow) string {
	if w == nil {
		return ""
	}
	start := w.OpensAt.UTC().Format(time.RFC3339Nano)
	end := w.ClosesAt.UTC().Format(time.RFC3339Nano)
	return fmt.Sprintf("[%s,%s)", start, end)
}

func normalizeAllowedActions(actions []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, action := range actions {
		upper := strings.ToUpper(action)
		if _, ok := seen[upper]; ok {
			continue
		}
		seen[upper] = struct{}{}
		out = append(out, upper)
	}
	return out
}





