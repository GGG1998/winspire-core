<!--
Sync Impact Report:
Version change: 1.1.0 → 1.2.0 (added Technology Stack principle)
Added sections: IV. Technology Stack (Backend: Go 1.25 + Gin, Frontend: React 19 + TypeScript)
Removed sections: None
Templates requiring updates:
  ✅ golang_http.mdc - Already documents Gin, SQLC, pgx/v5
  ✅ typescript_react.mdc - Already documents React 19, Vite, Tailwind CSS v4, Zod
Follow-up TODOs: None - existing rules already enforce these standards
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

**Version**: 1.2.0 | **Ratified**: 2025-01-27 | **Last Amended**: 2025-12-04
