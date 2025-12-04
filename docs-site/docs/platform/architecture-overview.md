# Architecture Overview

The Winspire platform is built using **Domain-Driven Design (DDD)** and **Event-Driven Architecture (EDA)**. Services are organized around **Bounded Contexts** that communicate via domain events.

## System Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                         Clients                               │
│                   (Web, Mobile, Streaming)                    │
└───────────────────────────┬──────────────────────────────────┘
                            │ REST/SSE/WebSocket
┌───────────────────────────▼──────────────────────────────────┐
│                      API Gateway                              │
└───────────┬──────────────────────────────────┬────────────────┘
            │                                  │
            ▼                                  ▼
┌─────────────────────┐              ┌─────────────────────┐
│ TournamentManagement│◄────Events───┤   Matchmaking       │
│   (competition)     │              │   (matchmaking)     │
└──────────┬──────────┘              └──────────┬──────────┘
           │                                    │
           │              Events                │
           └────────────────┬───────────────────┘
                            ▼
                  ┌──────────────────────┐
                  │ Realtime/Projections │
                  │ (competition-host-   │
                  │      stream)         │
                  └──────────────────────┘
                            │ SSE/WS
                            ▼
                      [Live Updates]
```

## Bounded Contexts

See [Bounded Contexts](./bounded-contexts.md) for detailed mapping.

| Bounded Context | Service | Responsibility |
|----------------|---------|----------------|
| **TournamentManagement** | `competition` | Tournament lifecycle & registration |
| **Matchmaking** | `matchmaking` | Brackets, matches, scoring |
| **Realtime/Projections** | `competition-host-stream` | SSE streaming & read models |
| **GameLibrary** | `game-management` | Game catalog & assets |

## Architectural Principles

1. **Domain-Driven Design** - Business logic organized by bounded contexts
2. **Event-Driven Architecture** - Domains communicate via events
3. **CQRS** - Separate read/write models where beneficial
4. **API-First** - Well-defined contracts (OpenAPI, AsyncAPI)
5. **Cloud-Native** - Designed for cloud deployment
6. **Observability** - Built-in monitoring and tracing

## Technology Stack

- **API Gateway**: Kong / AWS API Gateway
- **Event Bus**: PostgreSQL NOTIFY/LISTEN (Phase 1) → SNS/SQS (Phase 2)
- **Services**: Go microservices (Gin/Echo framework)
- **Data Stores**: PostgreSQL per bounded context
- **Realtime**: SSE (Server-Sent Events) / WebSocket
- **Monitoring**: OpenTelemetry, CloudWatch, Prometheus

## Communication Patterns

### Synchronous
- REST APIs (OpenAPI specs) - for commands from clients
- Used sparingly for service-to-service (prefer events)

### Asynchronous
- **Event-driven** (primary pattern) - services react to domain events
- **Message queues** (future) - SQS for reliable event delivery

See [Integration Guidelines](./integration-guidelines.md) for details.

---

**Related Documentation**:
- [Bounded Contexts](./bounded-contexts.md)
- [Event Sourcing Patterns](./event-sourcing-patterns.md)
- [Service Mesh Migration](./service-mesh-migration.md)
