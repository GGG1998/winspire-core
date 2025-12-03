# Po DEMO, chcielibyśmy osiągnąć coś bardziej poprawnego 
```
services/competition/
├── internal/
│   ├── domain/              # ← NOWA WARSTWA
│   │   ├── tournament/
│   │   │   ├── aggregate.go      # Aggregate root + encje
│   │   │   ├── service.go        # Domain services
│   │   │   ├── repository.go     # Interface repozytori (port)
│   │   │   ├── events.go         # Domain events
│   │   │   └── errors.go         # Domain-specific errors
│   │   └── registration/
│   │       ├── aggregate.go
│   │       └── service.go
│   │
│   ├── application/         # Application services (use cases)
│   │   └── tournament/
│   │       └── handlers.go       # Use case handlers (DTOs)
│   │
│   ├── repository/           # Implementacje repozytoriów
│   │   ├── tournament.go
│   │   ├── registration.go
│   │   └── store/
│   │       └── sqlc/
│   │
│   └── http/                # HTTP presentation layer
│       ├── handlers/
│       └── server.go
```

## Event Sourcing example
```go
package tournament

import (
    "time"
    "github.com/google/uuid"
)

// BaseEvent zawiera wspólne pola dla wszystkich eventów
type BaseEvent struct {
    EventID     uuid.UUID `json:"eventId"`
    AggregateID uuid.UUID `json:"aggregateId"`
    OccurredAt  time.Time `json:"occurredAt"`
    Version     int       `json:"version"`
}

// ============================================================================
// Tournament Lifecycle Events
// ============================================================================

// TournamentCreated - emitowany gdy turniej jest utworzony
type TournamentCreated struct {
    BaseEvent
    HostID           uuid.UUID  `json:"hostId"`
    Name             string     `json:"name"`
    MinimumTeamCount int32      `json:"minimumTeamCount"`
    MaximumTeamCount *int32     `json:"maximumTeamCount,omitempty"`
    TeamSize         int32      `json:"teamSize"`
    ScheduledStartAt *time.Time `json:"scheduledStartAt,omitempty"`
}

func (e TournamentCreated) EventName() string {
    return "tournament.created"
}

// TournamentStarted - emitowany gdy turniej został wystartowany
type TournamentStarted struct {
    BaseEvent
    ActualStartTime   time.Time `json:"actualStartTime"`
    ParticipantCount  int       `json:"participantCount"`
}

func (e TournamentStarted) EventName() string {
    return "tournament.started"
}

// TournamentCancelled - emitowany gdy turniej został anulowany
type TournamentCancelled struct {
    BaseEvent
    Reason    string    `json:"reason"`
    CancelledBy uuid.UUID `json:"cancelledBy"`
}

func (e TournamentCancelled) EventName() string {
    return "tournament.cancelled"
}

// TournamentCompleted - emitowany gdy turniej się zakończył
type TournamentCompleted struct {
    BaseEvent
    CompletedAt time.Time   `json:"completedAt"`
    WinnerID    *uuid.UUID  `json:"winnerId,omitempty"`
}

func (e TournamentCompleted) EventName() string {
    return "tournament.completed"
}

// ============================================================================
// Registration Events
// ============================================================================

// ParticipantRegistered - emitowany gdy gracz dołącza do turnieju
type ParticipantRegistered struct {
    BaseEvent
    TournamentID uuid.UUID   `json:"tournamentId"`
    UserID       uuid.UUID   `json:"userId"`
    TeamID       *uuid.UUID  `json:"teamId,omitempty"`
    TeamName     *string     `json:"teamName,omitempty"`
    RegisteredAt time.Time   `json:"registeredAt"`
}

func (e ParticipantRegistered) EventName() string {
    return "tournament.participant.registered"
}

// ParticipantRemoved - emitowany gdy gracz opuszcza turniej
type ParticipantRemoved struct {
    BaseEvent
    TournamentID uuid.UUID `json:"tournamentId"`
    UserID       uuid.UUID `json:"userId"`
    Reason       string    `json:"reason"`
}

func (e ParticipantRemoved) EventName() string {
    return "tournament.participant.removed"
}

// ============================================================================
// Configuration Change Events
// ============================================================================

// TournamentConfigUpdated - emitowany gdy zmienia się konfiguracja turnieju
type TournamentConfigUpdated struct {
    BaseEvent
    Changes map[string]interface{} `json:"changes"` // co się zmieniło
}

func (e TournamentConfigUpdated) EventName() string {
    return "tournament.config.updated"
}

// ============================================================================
// Event Interface & Collection
// ============================================================================

// DomainEvent to interfejs dla wszystkich domain events
type DomainEvent interface {
    EventName() string
    GetAggregateID() uuid.UUID
    GetOccurredAt() time.Time
}

// GetAggregateID implementacja dla BaseEvent
func (b BaseEvent) GetAggregateID() uuid.UUID {
    return b.AggregateID
}

func (b BaseEvent) GetOccurredAt() time.Time {
    return b.OccurredAt
}

// EventCollection przechowuje eventy przed ich persystencją
type EventCollection struct {
    events []DomainEvent
}

func NewEventCollection() *EventCollection {
    return &EventCollection{
        events: make([]DomainEvent, 0),
    }
}

func (ec *EventCollection) Add(event DomainEvent) {
    ec.events = append(ec.events, event)
}

func (ec *EventCollection) GetAll() []DomainEvent {
    return ec.events
}

func (ec *EventCollection) Clear() {
    ec.events = make([]DomainEvent, 0)
}
```

```go
// domain/tournament/aggregate.go
type Tournament struct {
    id               uuid.UUID
    hostID           uuid.UUID
    name             string
    status           Status
    // ...
    
    // Niewyeksportowane eventy - zgromadzone przed persystencją
    events *EventCollection
}

// Start rozpoczyna turniej i emituje event
func (t *Tournament) Start() error {
    if t.status != StatusRegistrationClosed {
        return ErrCannotStart
    }
    
    // Zmień stan
    t.status = StatusStarted
    t.actualStartTimeAt = time.Now()
    
    // Dodaj event
    t.events.Add(TournamentStarted{
        BaseEvent: BaseEvent{
            EventID:     uuid.New(),
            AggregateID: t.id,
            OccurredAt:  time.Now(),
            Version:     t.version,
        },
        ActualStartTime:  *t.actualStartTimeAt,
        ParticipantCount: len(t.participants),
    })
    
    return nil
}

// GetEvents zwraca zgromadzone eventy (potem wyczyszczone po zapisie)
func (t *Tournament) GetEvents() []DomainEvent {
    return t.events.GetAll()
}

func (t *Tournament) ClearEvents() {
    t.events.Clear()
}
```

