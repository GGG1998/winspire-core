package projections

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// AllowedActionOverlay describes participant-specific actions from Participation API.
type AllowedActionOverlay struct {
	LineupID string   `json:"lineupId"`
	Actions  []string `json:"allowedActions"`
}

// AllowedActionOverride toggles a single action for a lineup entry.
type AllowedActionOverride struct {
	LineupID uuid.UUID
	Action   string
	Enabled  bool
}

// MergeAllowedActions overlays allowed actions JSON onto lineup status payload.
func MergeAllowedActions(lineupStatus json.RawMessage, overlays []AllowedActionOverlay) (json.RawMessage, error) {
	if len(lineupStatus) == 0 {
		return json.Marshal(overlays)
	}

	var current []map[string]any
	if err := json.Unmarshal(lineupStatus, &current); err != nil {
		return nil, fmt.Errorf("decode lineupStatus: %w", err)
	}

	for _, overlay := range overlays {
		updated := false
		for _, item := range current {
			if item["lineupId"] == overlay.LineupID {
				item["allowedActions"] = overlay.Actions
				updated = true
				break
			}
		}
		if !updated {
			current = append(current, map[string]any{
				"lineupId":       overlay.LineupID,
				"allowedActions": overlay.Actions,
			})
		}
	}

	return json.Marshal(current)
}

// ApplyAllowedActionOverrides mutates lineup JSON using enable/disable overrides.
func ApplyAllowedActionOverrides(lineup json.RawMessage, overrides []AllowedActionOverride) (json.RawMessage, error) {
	if len(overrides) == 0 {
		return lineup, nil
	}

	var entries []map[string]any
	if len(lineup) == 0 {
		entries = []map[string]any{}
	} else if err := json.Unmarshal(lineup, &entries); err != nil {
		return nil, fmt.Errorf("decode lineupStatus: %w", err)
	}

	var updated []map[string]any
	if entries == nil {
		entries = []map[string]any{}
	}
	for _, entry := range entries {
		updated = append(updated, entry)
	}

	getActions := func(entry map[string]any) []string {
		raw, ok := entry["allowedActions"].([]any)
		if !ok {
			return []string{}
		}
		var actions []string
		for _, item := range raw {
			if str, ok := item.(string); ok {
				actions = append(actions, strings.ToUpper(str))
			}
		}
		return actions
	}

	for _, override := range overrides {
		var target map[string]any
		for _, entry := range updated {
			if entry["lineupId"] == override.LineupID.String() {
				target = entry
				break
			}
		}
		if target == nil {
			target = map[string]any{
				"lineupId":       override.LineupID.String(),
				"allowedActions": []string{},
			}
			updated = append(updated, target)
		}

		current := getActions(target)
		target["allowedActions"] = toggleAction(current, override.Action, override.Enabled)
	}

	return json.Marshal(updated)
}

func toggleAction(actions []string, action string, enabled bool) []string {
	action = strings.ToUpper(action)
	exists := false
	next := make([]string, 0, len(actions))
	for _, current := range actions {
		cur := strings.ToUpper(current)
		if cur == action {
			exists = true
			if !enabled {
				continue
			}
		}
		next = append(next, cur)
	}
	if enabled && !exists {
		next = append(next, action)
	}
	return next
}
