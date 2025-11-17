<!--
Sync Impact Report:
Version change: N/A → 1.0.0 (initial constitution)
Added sections: Core Principles (Modular Monorepo, Independent Modules, Shared Libraries, Independent Versioning)
Removed sections: None (initial version)
Templates requiring updates:
  ✅ plan-template.md - Updated Project Structure section to reference monorepo layout
  ✅ spec-template.md - No changes needed (already structure-agnostic)
  ✅ tasks-template.md - Updated Path Conventions to include monorepo structure
Follow-up TODOs: None
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

### IV. Independent Versioning
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

**Version**: 1.0.0 | **Ratified**: 2025-01-27 | **Last Amended**: 2025-01-27
