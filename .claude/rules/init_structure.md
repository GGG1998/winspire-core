---
paths: '**/*'
---
# 🧭 Winspire Platform — Repository Organization & Rules

This document defines the structure, conventions, and core technologies used across the **Winspire Platform monorepo**.  
It serves as a shared baseline for development, CI/CD automation, and release/versioning consistency.

It is only example for future work

```bash
repo/
├─ services/              # Microservices(independent module)
│  ├─ auth/
│  │  ├─ bin
│  │  ├─ tests
│  │  ├─ cmd/auth/
│  │  ├─ internal/
│  │  ├─ go.mod
│  │  ├─ Dockerfile
│  │  ├─ docker-compose.dev.yml
│  │  └─ Makefile
│  ├─ streamer/
│  └─ tournament/
│
├─ libs/                  # shared code
│  ├─ go/
│  │  ├─ common/          # go.mod, logger, config, errors
│  │  ├─ observability/   # otel, prometheus, tracing
│  │  └─ clients/         # clients do usług
│  └─ ts/
│     ├─ something1/         
│     └─ something2/        
│
├─ frontends/
│  ├─ dashboard/          # pnpm workspace
│  └─ streamer-panel/
│
├─ platform/
│  ├─ terraform/
│  │  ├─ modules/         # VPC, ECR, ECS/K8s, RDS, Redis
│  │  └─ envs/
│  │     ├─ dev/
│  │     ├─ staging/
│  │     └─ prod/
│  ├─ k8s/
│  │  ├─ base/
│  │  └─ overlays/
│  └─ ci/
│     └─ reusable/
│
├─ go.work                # connect all Go modules
├─ package.json           # pnpm workspaces
├─ Makefile               # tool helper to join tasks (build/test/deploy)
└─ .github/workflows/     # CI/CD pipeline
```

**Principles**
- Each service or library is an **independent Go module** (`go.mod`).
- Frontend apps and SDKs are **npm workspaces** (via `pnpm`).
- Infrastructure code is **isolated per environment**.
- Everything is **built and versioned independently** but stored together for coordination.

---

# 🏛️ Bounded Contexts & Service Responsibilities

Each service maps to a **Bounded Context** in Domain-Driven Design. Services communicate via **events** (async) or **REST APIs** (sync), but maintain clear ownership of their domain logic and data.

## 🎯 Current Service Map

```bash
services/
│
├── tournament/                     # BC: TournamentManagement
│   ├── Domain: Tournament lifecycle & participant registration
│   ├── Responsibilities:
│   │   ✅ Tournament creation (TournamentCreated)
│   │   ✅ Publication (TournamentPublished)
│   │   ✅ Registration management (ParticipantRegistered, ParticipantConfirmed)
│   │   ✅ Tournament status transitions (TournamentStarted, TournamentCompleted, TournamentCancelled)
│   │   ❌ NO match mechanics, brackets, or scoring
│   ├── Aggregates: Tournament, TournamentParticipant, Host
│   └── Database: PostgreSQL (tournaments, participants, hosts)
│
├── matchmaking/                    # BC: Matchmaking
│   ├── Domain: Match creation, brackets, and game flow
│   ├── Responsibilities:
│   │   ✅ Bracket generation (BracketGenerated)
│   │   ✅ Round creation (RoundCreated)
│   │   ✅ Match creation & assignment (MatchCreated)
│   │   ✅ Lobby management (ParticipantJoinedLobby)
│   │   ✅ Score submission (ScoreSubmitted, MatchCompleted)
│   │   ✅ Bracket progression (ParticipantAdvanced, ParticipantEliminated)
│   │   ❌ NO tournament registration or lifecycle
│   ├── Aggregates: Bracket, Round, Match, MatchParticipant, Lobby
│   └── Database: PostgreSQL (brackets, rounds, matches)
│
│
└── game-management/                # BC: GameLibrary
    ├── Domain: Game catalog & asset management
    ├── Responsibilities:
    │   ✅ Game metadata (name, description, rules)
    │   ✅ Version management
    │   ✅ Asset bundling & distribution
    │   ❌ NO tournament or match logic
    ├── Aggregates: Game, GameVersion, GameBundle
    └── Database: PostgreSQL + S3 (game assets)
```

## 📡 Inter-Service Communication

### Event-Driven (Async)

**Event Transport Options**:
- **Phase 1 (current)**: Redis NOTIFY/LISTEN
- **Phase 2 (future)**: AWS SNS/SQS or EventBridge or Redis Pub/Sub
- **Phase 3 (advanced)**: Event Store (DynamoDB/EventStoreDB) for now Postgres

### REST APIs (Sync)
Only for:
- Client → Service commands (POST /tournaments, POST /matches/ready)
- Service → Service queries (rare, prefer events)

## 🚫 Anti-Patterns to Avoid

- ❌ **Shared database** between services (violates BC boundaries) for now we have but create new table
- ❌ **Direct service-to-service calls** for domain logic (use events)
- ❌ **Domain logic in realtime stream** (only projections)
- ❌ **Matchmaking logic in tournament** (wrong BC)

## ✅ Adding a New Service

1. Create directory: `services/<service-name>/`
2. Initialize Go module: `cd services/<service-name> && go mod init`
3. Add to workspace: Update root `go.work`
4. Define Bounded Context in docs: `docs-site/docs/platform/bounded-contexts.md`
5. Document events: `docs-site/docs/platform/events/<service-name>.md`
6. Add service tag: `services/<service-name>/v0.1.0`

---

# 🏷️ Tagging & Versioning Rules

Each component (service, library, frontend, infra) has its own version, managed via scoped Git tags.

## 🔖 Tag Naming Convention

Tags use the following format:

```bash
<path>/<component>/v<MAJOR>.<MINOR>.<PATCH>
```

Examples:

```bash
services/auth/v1.3.0
services/chat/v0.9.2
libs/ts/ui-kit/v2.0.0
frontends/dashboard/v1.5.1
platform/terraform/v0.8.0
```
