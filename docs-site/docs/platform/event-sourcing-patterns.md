# Event Sourcing Patterns

Guide for implementing Event Sourcing in Winspire microservices.

## Current Architecture

The `competition-host-stream` service already uses event-driven patterns:

```go
// services/competition-host-stream/internal/sse/event_router.go
type EventRouter struct {
    broker *Broker
}

func (r *EventRouter) CupUpdated(ctx context.Context, cupID uuid.UUID, payload any) {
    r.publish(ctx, Scope{Type: "cup", ID: cupID}, "CupHostViewUpdated", payload)
}

func (r *EventRouter) MatchmakingQueueUpdated(ctx context.Context, tournamentID uuid.UUID, payload any) {
    r.publish(ctx, Scope{Type: "tournament", ID: tournamentID}, "MatchmakingQueueUpdated", payload)
}

func (r *EventRouter) TournamentParticipationUpdate(ctx context.Context, tournamentID uuid.UUID, payload any) {
    r.publish(ctx, Scope{Type: "tournament", ID: tournamentID}, "TournamentParticipationUpdate", payload)
}
```

**Current flow:**
```
Command → Service → Database (write) → Event Router → SSE Broker → Clients
                                    ↓
                              Projections (read model)
```

**This is already CQRS!**
- Commands update the database
- Events trigger projection updates
- Clients read from projections

## Event Sourcing Evolution

### Phase 1: Current State (CQRS without Event Store)

**What we have:**
- ✅ Command handlers
- ✅ Event publishing (SSE)
- ✅ Read model projections
- ❌ Events stored in database
- ❌ Event history

**Example:**

```go
// Cup creation command
func (s *Service) CreateCup(ctx context.Context, cmd CreateCupCommand) error {
    // Write to database
    cup := &Cup{ID: cmd.CupID, Name: cmd.Name}
    if err := s.repo.SaveCup(ctx, cup); err != nil {
        return err
    }
    
    // Publish event
    s.eventRouter.CupUpdated(ctx, cup.ID, cup)
    
    return nil
}
```

**Pros:**
- Simple to understand
- Easy to debug (database is source of truth)
- Low operational complexity

**Cons:**
- Can't rebuild state from events
- No audit trail
- Limited time-travel debugging

### Phase 2: Event Store Introduction

**What changes:**
- Add event store (DynamoDB or PostgreSQL)
- Store events alongside database updates
- Keep database as source of truth (for now)

**Architecture:**

```
Command → Service → [Event Store + Database] → Event Bus → Projections
                           ↓
                    Event History (audit)
```

**Implementation:**

```go
// Event store interface
type EventStore interface {
    Append(ctx context.Context, streamID string, events []Event) error
    Read(ctx context.Context, streamID string, fromVersion int64) ([]Event, error)
}

// Cup creation with event store
func (s *Service) CreateCup(ctx context.Context, cmd CreateCupCommand) error {
    // 1. Create event
    event := CupCreatedEvent{
        CupID:     cmd.CupID,
        Name:      cmd.Name,
        Timestamp: time.Now(),
        Version:   1,
    }
    
    // 2. Store event in event store
    if err := s.eventStore.Append(ctx, "cup-"+cmd.CupID.String(), []Event{event}); err != nil {
        return err
    }
    
    // 3. Update database (still source of truth)
    cup := &Cup{ID: cmd.CupID, Name: cmd.Name}
    if err := s.repo.SaveCup(ctx, cup); err != nil {
        return err
    }
    
    // 4. Publish event to SSE
    s.eventRouter.CupUpdated(ctx, cup.ID, cup)
    
    return nil
}
```

**DynamoDB Schema:**

```hcl
resource "aws_dynamodb_table" "event_store" {
  name         = "${var.environment}-winspire-events"
  billing_mode = "PAY_PER_REQUEST"
  
  hash_key  = "stream_id"    # "cup-{uuid}"
  range_key = "version"      # event version number
  
  attribute {
    name = "stream_id"
    type = "S"
  }
  
  attribute {
    name = "version"
    type = "N"
  }
  
  attribute {
    name = "timestamp"
    type = "N"
  }
  
  global_secondary_index {
    name            = "timestamp-index"
    hash_key        = "stream_id"
    range_key       = "timestamp"
    projection_type = "ALL"
  }
}
```

**Benefits:**
- ✅ Complete audit trail
- ✅ Event history for debugging
- ✅ Can replay events
- ✅ Database still source of truth (low risk)

**Cons:**
- Additional storage (DynamoDB)
- Dual-write complexity (event store + database)

**Cost:** ~$5-10/month for DynamoDB (pay-per-request)

### Phase 3: Event Sourcing (Event Store as Source of Truth)

**What changes:**
- Event store becomes primary data source
- Database becomes read model (built from events)
- Rebuild read models from event history

**Architecture:**

```
Command → Service → Event Store → Event Bus → [Projections → Database]
                         ↑
                    Source of Truth
```

**Implementation:**

```go
// Cup creation (pure event sourcing)
func (s *Service) CreateCup(ctx context.Context, cmd CreateCupCommand) error {
    streamID := "cup-" + cmd.CupID.String()
    
    // 1. Read current events (if any)
    events, err := s.eventStore.Read(ctx, streamID, 0)
    if err != nil {
        return err
    }
    
    // 2. Rebuild state from events
    cup := &Cup{}
    for _, event := range events {
        cup.Apply(event)
    }
    
    // 3. Validate command
    if cup.ID != uuid.Nil {
        return fmt.Errorf("cup already exists")
    }
    
    // 4. Create new event
    newEvent := CupCreatedEvent{
        CupID:     cmd.CupID,
        Name:      cmd.Name,
        Timestamp: time.Now(),
        Version:   int64(len(events)) + 1,
    }
    
    // 5. Append to event store (single source of truth)
    if err := s.eventStore.Append(ctx, streamID, []Event{newEvent}); err != nil {
        return err
    }
    
    // 6. Publish to event bus for projections
    s.eventBus.Publish(ctx, newEvent)
    
    return nil
}

// Projection builder (separate process)
func (p *Projector) HandleCupCreated(ctx context.Context, event CupCreatedEvent) error {
    // Update database read model
    cup := &Cup{
        ID:   event.CupID,
        Name: event.Name,
    }
    return p.repo.SaveCup(ctx, cup)
}
```

**Benefits:**
- ✅ True event sourcing
- ✅ Rebuild read models anytime
- ✅ Time-travel debugging
- ✅ Perfect audit trail
- ✅ Event replay for migrations

**Cons:**
- Complex to implement correctly
- Event versioning challenges
- Eventual consistency
- Higher operational complexity

**When to use:**
- Audit requirements (financial, regulatory)
- Need to replay events
- Complex domain logic
- Multiple read models from same events

### Phase 4: Event Bus (SNS/SQS or EventBridge)

**What changes:**
- Add AWS EventBridge or SNS/SQS
- Services communicate via events
- Decoupled architecture

**Architecture:**

```
Service A → Event Store → EventBridge → [Service B, Service C]
                              ↓
                        Dead Letter Queue
```

**Implementation:**

```go
// Publish to EventBridge
func (s *Service) PublishEvent(ctx context.Context, event Event) error {
    input := &eventbridge.PutEventsInput{
        Entries: []types.PutEventsRequestEntry{
            {
                Source:       aws.String("winspire.competition-host-stream"),
                DetailType:   aws.String(event.Type()),
                Detail:       aws.String(event.JSON()),
                EventBusName: aws.String("winspire-events"),
            },
        },
    }
    
    _, err := s.eventBridgeClient.PutEvents(ctx, input)
    return err
}

// Subscribe to events (in another service)
// EventBridge rule routes to SQS queue
func (s *GameManagementService) ProcessEvents(ctx context.Context) {
    for {
        messages, err := s.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
            QueueUrl:            aws.String(s.queueURL),
            MaxNumberOfMessages: 10,
            WaitTimeSeconds:     20,
        })
        
        for _, msg := range messages.Messages {
            event := ParseEvent(msg.Body)
            
            switch event.Type() {
            case "CupCreated":
                s.handleCupCreated(ctx, event)
            case "TournamentUpdated":
                s.handleTournamentUpdated(ctx, event)
            }
            
            // Delete message after processing
            s.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
                QueueUrl:      aws.String(s.queueURL),
                ReceiptHandle: msg.ReceiptHandle,
            })
        }
    }
}
```

**Terraform:**

```hcl
resource "aws_cloudwatch_event_bus" "winspire" {
  name = "winspire-events"
}

resource "aws_cloudwatch_event_rule" "cup_events" {
  name           = "cup-events"
  event_bus_name = aws_cloudwatch_event_bus.winspire.name
  
  event_pattern = jsonencode({
    source      = ["winspire.competition-host-stream"]
    detail-type = ["CupCreated", "CupUpdated", "CupDeleted"]
  })
}

resource "aws_cloudwatch_event_target" "cup_events_queue" {
  rule           = aws_cloudwatch_event_rule.cup_events.name
  event_bus_name = aws_cloudwatch_event_bus.winspire.name
  target_id      = "game-management-queue"
  arn            = aws_sqs_queue.game_management_events.arn
}

resource "aws_sqs_queue" "game_management_events" {
  name = "${var.environment}-game-management-events"
  
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.game_management_dlq.arn
    maxReceiveCount     = 3
  })
}

resource "aws_sqs_queue" "game_management_dlq" {
  name = "${var.environment}-game-management-events-dlq"
}
```

**Benefits:**
- ✅ Decoupled services
- ✅ Async communication
- ✅ Dead letter queues
- ✅ Event filtering
- ✅ Multiple subscribers

**Cost:**
- EventBridge: $1.00 per million events
- SQS: $0.40 per million requests (first 1M free)
- ~$5-10/month for typical usage

## Recommended Migration Path

### For Most Teams: Stay on Phase 1-2

**Phase 1 (Current) is sufficient if:**
- Audit requirements are basic
- Single read model per aggregate
- No need to replay events
- Team comfortable with current patterns

**Move to Phase 2 if:**
- Need audit trail for compliance
- Want event history for debugging
- Planning for future event replay

### For Advanced Teams: Phase 3-4

**Move to Phase 3 (Pure Event Sourcing) if:**
- Financial/regulatory audit requirements
- Multiple read models from same events
- Complex domain logic with many state transitions
- Need to replay events for migrations

**Move to Phase 4 (Event Bus) if:**
- 5+ microservices
- Services need to react to events from other services
- Building event-driven architecture
- Need async processing

## Event Versioning Strategy

As events evolve, you'll need versioning:

```go
// v1
type CupCreatedEventV1 struct {
    CupID uuid.UUID `json:"cup_id"`
    Name  string    `json:"name"`
}

// v2 (added field)
type CupCreatedEventV2 struct {
    CupID       uuid.UUID `json:"cup_id"`
    Name        string    `json:"name"`
    Description string    `json:"description"` // new field
}

// Upcaster
func UpcastCupCreated(event Event) Event {
    switch event.Version() {
    case 1:
        v1 := event.(*CupCreatedEventV1)
        return &CupCreatedEventV2{
            CupID:       v1.CupID,
            Name:        v1.Name,
            Description: "", // default value
        }
    case 2:
        return event
    default:
        panic("unknown version")
    }
}
```

## Testing Event-Driven Systems

```go
func TestCupCreation(t *testing.T) {
    // Arrange
    eventStore := NewInMemoryEventStore()
    service := NewCupService(eventStore)
    
    // Act
    cmd := CreateCupCommand{CupID: uuid.New(), Name: "Test Cup"}
    err := service.CreateCup(context.Background(), cmd)
    
    // Assert
    require.NoError(t, err)
    
    events, err := eventStore.Read(context.Background(), "cup-"+cmd.CupID.String(), 0)
    require.NoError(t, err)
    require.Len(t, events, 1)
    require.IsType(t, &CupCreatedEvent{}, events[0])
}

func TestProjectionBuilder(t *testing.T) {
    // Arrange
    repo := NewInMemoryRepository()
    projector := NewProjector(repo)
    
    // Act - replay events
    events := []Event{
        &CupCreatedEvent{CupID: uuid.New(), Name: "Cup 1"},
        &CupCreatedEvent{CupID: uuid.New(), Name: "Cup 2"},
    }
    
    for _, event := range events {
        err := projector.Handle(context.Background(), event)
        require.NoError(t, err)
    }
    
    // Assert - check read model
    cups, err := repo.ListCups(context.Background())
    require.NoError(t, err)
    require.Len(t, cups, 2)
}
```

## Summary

**Current architecture (Phase 1) is already well-designed:**
- ✅ CQRS pattern
- ✅ Event-driven SSE
- ✅ Projection-based reads

**Recommendations:**
- **Now**: Stay on Phase 1 (works well for current scale)
- **+6 months**: Add Phase 2 (event store for audit trail) if needed
- **+12 months**: Consider Phase 3-4 if you have 5+ services and complex event-driven flows

**Remember:** Event Sourcing adds complexity. Only adopt it when the benefits clearly outweigh the costs.


