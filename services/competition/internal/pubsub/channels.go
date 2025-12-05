// Package pubsub handles Redis Pub/Sub for domain events
package pubsub

import (
	"fmt"
	"strings"
)

// Channel name constants for event distribution
const (
	// Prefix for all event channels
	ChannelPrefix = "events"

	// Bounded contexts
	ContextTournamentManagement = "tournament_management"
)

// ChannelName generates a Redis channel name for an event
// Format: events:{bounded_context}:{event_name}
func ChannelName(boundedContext, eventName string) string {
	eventNameLower := strings.ToLower(
		// Convert CamelCase to snake_case
		toSnakeCase(eventName),
	)
	return fmt.Sprintf("%s:%s:%s", ChannelPrefix, boundedContext, eventNameLower)
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// Tournament Management event channels
var (
	// ChannelTournamentStartRequested represents a command/intent to start a tournament
	// This triggers the tournament start saga (grace period → bracket generation → confirmation)
	ChannelTournamentStartRequested = ChannelName(ContextTournamentManagement, "TournamentStartRequested")
	
	// ChannelTournamentStarted represents the fact that a tournament has successfully started
	// This is published after the saga completes (after grace period and bracket generation)
	ChannelTournamentStarted = ChannelName(ContextTournamentManagement, "TournamentStarted")
	
	// ChannelTournamentStartFailed represents a failed tournament start attempt
	// This triggers compensating actions (status rollback)
	ChannelTournamentStartFailed = ChannelName(ContextTournamentManagement, "TournamentStartFailed")
)

