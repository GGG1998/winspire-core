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
	ContextMatchmaking        = "matchmaking"
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

// Matchmaking event channels
var (
	ChannelBracketGenerated       = ChannelName(ContextMatchmaking, "BracketGenerated")
	ChannelRoundCreated           = ChannelName(ContextMatchmaking, "RoundCreated")
	ChannelMatchCreated           = ChannelName(ContextMatchmaking, "MatchCreated")
	ChannelMatchStarted           = ChannelName(ContextMatchmaking, "MatchStarted")
	ChannelMatchCompleted         = ChannelName(ContextMatchmaking, "MatchCompleted")
	ChannelParticipantAdvanced    = ChannelName(ContextMatchmaking, "ParticipantAdvanced")
	ChannelParticipantEliminated  = ChannelName(ContextMatchmaking, "ParticipantEliminated")
	ChannelWalkoverGranted        = ChannelName(ContextMatchmaking, "WalkoverGranted")
	ChannelPlayerConnectionLost   = ChannelName(ContextMatchmaking, "PlayerConnectionLost")
	ChannelPlayerConnectionRestored = ChannelName(ContextMatchmaking, "PlayerConnectionRestored")

	// Pre-lobby event channels
	ChannelPreLobbyCreated           = ChannelName(ContextMatchmaking, "PreLobbyCreated")
	ChannelPreLobbyGracePeriodStarted = ChannelName(ContextMatchmaking, "PreLobbyGracePeriodStarted")
	ChannelPreLobbyGracePeriodEnded  = ChannelName(ContextMatchmaking, "PreLobbyGracePeriodEnded")
	ChannelPreLobbyParticipantSnapshot = ChannelName(ContextMatchmaking, "PreLobbyParticipantSnapshot")
	ChannelPreLobbyCancelled         = ChannelName(ContextMatchmaking, "PreLobbyCancelled")
)

// Tournament Management event channels (consumed by matchmaking)
var (
	ChannelTournamentStarted = ChannelName(ContextTournamentManagement, "TournamentStarted")
)

// GetMatchmakingChannels returns all channels that matchmaking publishes to
func GetMatchmakingChannels() []string {
	return []string{
		ChannelBracketGenerated,
		ChannelRoundCreated,
		ChannelMatchCreated,
		ChannelMatchStarted,
		ChannelMatchCompleted,
		ChannelParticipantAdvanced,
		ChannelParticipantEliminated,
		ChannelWalkoverGranted,
		ChannelPlayerConnectionLost,
		ChannelPlayerConnectionRestored,
		// Pre-lobby channels
		ChannelPreLobbyCreated,
		ChannelPreLobbyGracePeriodStarted,
		ChannelPreLobbyGracePeriodEnded,
		ChannelPreLobbyParticipantSnapshot,
		ChannelPreLobbyCancelled,
	}
}

// GetSubscriptionChannels returns all channels that matchmaking subscribes to
func GetSubscriptionChannels() []string {
	return []string{
		ChannelTournamentStarted,
	}
}


