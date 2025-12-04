<!--
Sync Impact Report:
Version change: 1.2.0 → 1.3.0 (added File Naming Conventions principle)
Added sections: VI. File Naming Conventions (Go: no _repo/_handler suffixes, converters.go; React: PascalCase components, use-prefix hooks)
Removed sections: None
Templates requiring updates:
  ✅ golang_http.mdc - Should document file naming patterns
  ✅ typescript_react.mdc - Should document component/hook naming patterns
Follow-up TODOs:
  - Update existing code to follow new naming conventions
  - Add file naming examples to rule templates
Previous version history:
  - 1.1.0 → 1.2.0 (added Technology Stack principle)
-->

# winspire-core Constitution

## Core Principles

### I. Modular Monorepo
The project MUST be organized as a modular monorepo with clear separation of concerns. The repository structure MUST follow the established layout: `services/` for microservices, `libs/` for shared libraries, `frontends/` for frontend applications, and `platform/` for infrastructure code. Each component MUST be independently buildable, testable, and deployable while maintaining coordination through shared tooling and conventions.

**Rationale**: A modular monorepo enables code sharing, coordinated releases, and simplified dependency management while preserving component independence and clear boundaries.

### II. Independent Modules
Each service and library MUST be an independent module with its own dependency management. Go services MUST have their own `go.mod` file and be included in the root `go.work` file. Frontend applications MUST be organized as pnpm workspaces with independent `package.json` files. Infrastructure code MUST be isolated per environment.

**Rationale**: Independent modules enable component-level versioning, reduce coupling, and allow teams to work autonomously on different parts of the system.

### III. Shared Libraries
Common functionality MUST be extracted into shared libraries under `libs/`. Libraries MUST be organized by language (e.g., `libs/go/`, `libs/ts/`) and MUST have clear, documented purposes. Libraries MUST be self-contained, independently testable, and versioned separately from consuming services.

**Rationale**: Shared libraries reduce duplication, ensure consistency across services, and enable reuse of well-tested components.

### III.5. Bounded Contexts & Domain-Driven Design
Services MUST be organized around **Bounded Contexts** as defined by Domain-Driven Design. Each service MUST own a clearly defined domain with specific aggregates, events, and business rules. Services MUST NOT share databases or domain logic. Communication between services MUST happen via domain events (async) or well-defined REST APIs (sync). Domain events MUST be documented with their schema, owning bounded context, and consuming services.

**Rationale**: Bounded contexts ensure clear ownership, reduce coupling, and enable independent evolution of domain models. Event-driven communication decouples services temporally and allows for scalable, resilient architectures.

**Examples**:
- ✅ `competition` service owns Tournament lifecycle (TournamentManagement BC)
- ✅ `matchmaking` service owns Match creation and brackets (Matchmaking BC)
- ✅ Services communicate via `TournamentStarted`, `MatchCreated` events
- ❌ `competition` does NOT directly call `matchmaking` REST API for domain operations
- ❌ Services do NOT share tables or schemas across bounded contexts

### IV. Technology Stack
The project MUST adhere to the standardized technology stack to ensure consistency, maintainability, and team productivity.

**Backend (Go Services)**:
- **Language**: Go 1.25
- **HTTP Framework**: Gin (`github.com/gin-gonic/gin`)
- **Database Access**: SQLC (type-safe SQL code generation)
- **PostgreSQL Driver**: pgx/v5
- **Shared Libraries**: `libs/go/httpx` (middleware), `libs/go/auth` (JWT validation)

**Frontend (React Applications)**:
- **Language**: TypeScript 5.9+
- **Framework**: React 19
- **Build Tool**: Vite 7
- **Styling**: Tailwind CSS v4
- **Routing**: React Router v7
- **Forms & Validation**: React Hook Form + Zod
- **UI Components**: Headless UI (@headlessui/react)
- **Data Tables**: TanStack Table
- **State Management**: React Context + React Query (TanStack Query)

**Infrastructure**:
- **Database**: PostgreSQL (managed via migrations and SQLC)
- **Cache/PubSub**: Redis
- **Authentication**: Supabase Auth
- **Containerization**: Docker + Docker Compose

**Rationale**: A standardized stack reduces cognitive overhead, enables code reuse, ensures consistent patterns across services and frontends, and simplifies onboarding and maintenance.

### V. Independent Versioning
Each component (service, library, frontend, infrastructure) MUST be versioned independently using scoped Git tags following the format `<path>/<component>/v<MAJOR>.<MINOR>.<PATCH>`. Components MUST NOT be forced to version together unless there is a breaking change that requires coordinated release.

**Rationale**: Independent versioning allows components to evolve at their own pace, reduces deployment risk, and enables selective updates of system parts.

### VI. File Naming Conventions
All files MUST follow clear, descriptive naming patterns that reflect their purpose without redundant suffixes.

**Go Services (Backend)**:
- **Repository files**: Use the entity name (e.g., `bracket.go` for `BracketRepository`, NOT `bracket_repo.go`)
- **Handler files**: Use the entity name (e.g., `tournament.go` for tournament handlers, NOT `tournament_handler.go`)
- **Converters/Helpers**: Use `converters.go` for type conversion utilities (e.g., pgtype ↔ Go types)
- **Domain entities**: Use the entity name (e.g., `match.go` for `Match` aggregate)
- **Configuration**: Use `config.go` for configuration loading
- **Main entrypoint**: Use `main.go` in `cmd/<service>/`

**TypeScript/React (Frontend)**:
- **Components**: Use PascalCase for component files (e.g., `TournamentCard.tsx`, NOT `tournament-card.tsx`)
- **Hooks**: Use camelCase with `use` prefix (e.g., `useTournament.ts`, NOT `tournament-hook.ts`)
- **Utilities**: Use descriptive names (e.g., `api.ts`, `validation.ts`, NOT `utils.ts`)
- **Types**: Use `types.ts` for shared type definitions per module

**General Rules**:
- Avoid redundant suffixes that duplicate type information already clear from directory structure
- Use singular names for files that define a single entity or aggregate
- Use plural names for files that handle collections or utilities (e.g., `converters.go`, `events.go`)
- Be descriptive and consistent within each service/module

**Rationale**: Clear file naming reduces cognitive overhead, makes codebases easier to navigate, and prevents naming redundancy that clutters file listings. The directory structure already provides context (e.g., `repository/bracket.go` is clearly a repository).

**Examples**:
- ✅ `internal/repository/bracket.go` (BracketRepository)
- ✅ `internal/http/tournament.go` (tournament handlers)
- ✅ `internal/repository/converters.go` (type converters)
- ✅ `components/TournamentCard.tsx` (React component)
- ❌ `internal/repository/bracket_repo.go` (redundant `_repo` suffix)
- ❌ `internal/http/tournament_handler.go` (redundant `_handler` suffix)
- ❌ `components/tournament-card.tsx` (should use PascalCase)

## Development Workflow

All development MUST respect the modular monorepo structure. New features MUST be placed in the appropriate directory (`services/`, `libs/`, `frontends/`, or `platform/`) based on their purpose. Cross-component changes MUST be coordinated and documented. The root `Makefile` and workspace configuration files (`go.work`, `package.json`) MUST be updated when adding new modules.

## Governance

This constitution supersedes all other development practices and architectural decisions. Amendments to this constitution require:

1. Documentation of the proposed change and rationale
2. Review of impact on existing components and templates
3. Update of dependent templates and documentation
4. Version increment following semantic versioning:
   - **MAJOR**: Backward incompatible governance/principle removals or redefinitions
   - **MINOR**: New principle/section added or materially expanded guidance
   - **PATCH**: Clarifications, wording, typo fixes, non-semantic refinements

All pull requests and code reviews MUST verify compliance with these principles. Complexity that violates these principles MUST be justified with explicit rationale.

**Version**: 1.3.0 | **Ratified**: 2025-01-27 | **Last Amended**: 2025-12-04
