# Implementation Plan: Supabase Authentication Integration

**Branch**: `001-supabase-auth` | **Date**: 2025-11-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-supabase-auth/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a scalable authentication microservice in Go that integrates with Supabase Auth to handle user registration, login, OAuth flows, and role-based access control. The service will use sqlc for type-safe database queries, implement JWT validation middleware, and provide a clean API for other microservices to verify user authentication and authorization. The architecture follows microservices principles with clear boundaries, independent deployment, and horizontal scalability.

## Technical Context

**Language/Version**: Go 1.25.4 (latest stable)  
**Primary Dependencies**: 
- Supabase Go Client SDK (for auth operations)
- sqlc (for type-safe SQL code generation)
- PostgreSQL driver (pgx/v5)
- JWT validation library (github.com/golang-jwt/jwt/v5)
- HTTP framework: **Gin** (selected - see research.md)
- Configuration management: **envconfig** (selected - see research.md)

**Storage**: PostgreSQL (via Supabase, with local schema for roles/permissions)  
**Testing**: 
- Unit tests: Go standard `testing` package
- Integration tests: testcontainers or local Supabase instance
- Contract tests: OpenAPI validation (see research.md)

**Target Platform**: Linux server (containerized, deployable to AWS ECS/Fargate or Kubernetes)  
**Project Type**: microservice (backend API service)  
**Performance Goals**: 
- Handle 1,000 concurrent authentication requests (per spec SC-004)
- JWT validation < 10ms p95 latency
- Support 10,000 to 500,000 users (per spec SC-006)

**Constraints**: 
- Must integrate with existing Supabase Auth (no custom auth implementation)
- Must support horizontal scaling (stateless service design)
- Must validate JWTs issued by Supabase
- Must provide role-based access control (RBAC) layer
- SQL queries must be compatible with sqlc code generation

**Scale/Scope**: 
- Single microservice: `services/auth/`
- Shared library for JWT validation: `libs/go/auth/` (separate library from start - see research.md)
- Database schema for roles/permissions in Supabase PostgreSQL
- API contracts for other services to consume

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Modular Monorepo Compliance**: ✅ PASS
- Auth service → `services/auth/` (microservice, independent Go module)
- JWT validation library → `libs/go/auth/` (shared library for other services)
- Both components independently versioned and deployable

**Independent Modules Compliance**: ✅ PASS
- `services/auth/` will have its own `go.mod` file
- `libs/go/auth/` will have its own `go.mod` file
- Both modules will be included in root `go.work` file
- Each can be built, tested, and deployed independently

**Shared Libraries Compliance**: ✅ PASS
- JWT validation logic extracted to `libs/go/auth/` for reuse across services
- Library organized by language (`libs/go/`)
- Library will be self-contained and independently testable

**Independent Versioning Compliance**: ✅ PASS
- Service versioned as `services/auth/v1.0.0`
- Library versioned as `libs/go/auth/v1.0.0`
- No forced coordination unless breaking changes require it

**GATE STATUS**: ✅ ALL CHECKS PASSED - Proceed to Phase 0

### Post-Phase 1 Re-evaluation

**Modular Monorepo Compliance**: ✅ PASS
- Design confirms service in `services/auth/` and library in `libs/go/auth/`
- Clear separation of concerns maintained
- Both components independently deployable

**Independent Modules Compliance**: ✅ PASS
- Each component has independent `go.mod` (as designed)
- Both included in `go.work` workspace
- No coupling between modules

**Shared Libraries Compliance**: ✅ PASS
- JWT validation library properly scoped for reuse
- Library API designed for consumption by other services
- Clear boundaries established

**Independent Versioning Compliance**: ✅ PASS
- Versioning strategy confirmed: independent tags per component
- No forced coordination required

**GATE STATUS**: ✅ ALL CHECKS PASSED - Design complete, ready for implementation

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
services/[service-name]/
├── cmd/[service-name]/
├── internal/
├── pkg/
├── go.mod
└── Makefile

libs/[language]/[library-name]/
├── [library structure]
└── [go.mod or package.json]

frontends/[app-name]/
├── src/
└── package.json

platform/[component]/
└── [infrastructure code]
```

**Structure Decision**: 
- **Auth Service**: `services/auth/` - Main authentication microservice handling user registration, login, OAuth flows, and providing JWT validation endpoints for other services
- **Auth Library**: `libs/go/auth/` - Shared library containing JWT validation utilities, middleware, and common auth types that other microservices can import
- **Database Schema**: SQL migrations in `services/auth/migrations/` for role/permission tables (stored in Supabase PostgreSQL)
- **SQL Queries**: `services/auth/internal/queries/` - SQL files for sqlc code generation
- **API Contracts**: `specs/001-supabase-auth/contracts/` - OpenAPI specifications for service APIs

This structure follows the modular monorepo principles:
- Service is independently deployable and versioned
- Shared library enables code reuse without coupling
- Clear separation between service-specific and shared code
- Database schema managed within service but stored in shared Supabase instance

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | N/A | No violations - design complies with constitution |

---

## Phase Completion Status

### Phase 0: Research & Technology Decisions ✅ COMPLETE

**Output**: [research.md](./research.md)

**Resolved Clarifications**:
- ✅ HTTP framework: **Gin** selected
- ✅ Configuration management: **envconfig** selected
- ✅ Contract testing: **OpenAPI validation** approach
- ✅ JWT library structure: **Separate library** from start
- ✅ Supabase integration patterns: **Hybrid approach** documented
- ✅ sqlc setup: **pgx/v5 + migrations** pattern
- ✅ RBAC pattern: **DB-driven + JWT claims** approach
- ✅ JWT validation: **Middleware in library** pattern

### Phase 1: Design & Contracts ✅ COMPLETE

**Outputs**:
- ✅ [data-model.md](./data-model.md) - Complete data model with entities, relationships, validation rules
- ✅ [contracts/auth-service.yaml](./contracts/auth-service.yaml) - OpenAPI 3.1 specification
- ✅ [quickstart.md](./quickstart.md) - Development setup guide
- ✅ Agent context updated (Go 1.21+, PostgreSQL)

**Design Artifacts**:
- Database schema defined (roles, permissions, user_roles, role_permissions, oauth_provider_links)
- API contracts defined (authentication, OAuth, password management, user management, RBAC)
- Project structure documented
- Development workflow established

### Phase 2: Task Breakdown

**Status**: ⏳ PENDING (to be created by `/speckit.tasks` command)

**Next Steps**:
1. Run `/speckit.tasks` to generate task breakdown
2. Begin implementation following tasks
3. Follow quickstart guide for local development setup

---

## Implementation Readiness

✅ **All prerequisites met**:
- Technology decisions made and documented
- Data model designed
- API contracts defined
- Development environment setup documented
- Constitution compliance verified

**Ready for**: Task breakdown and implementation
