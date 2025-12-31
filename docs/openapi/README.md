# OpenAPI Specifications

This directory contains OpenAPI 3.1 specifications for all Winspire platform APIs.

## Directory Structure

```
docs/openapi/
├── shared/                      # Reusable components across services
│   ├── components/
│   │   ├── parameters.yaml      # Common path/query parameters
│   │   ├── responses.yaml       # Standard error responses
│   │   └── schemas/             # Shared schema definitions
│   └── security/
│       └── schemes.yaml         # Authentication schemes
│
├── services/                    # Per-service specifications
│   ├── tournament/              # Tournament Service (port 8089)
│   ├── matchmaking/             # Matchmaking Service (port 8088)
│   └── game-management/         # Game Management Service (port 8087)
│
├── aggregated/                  # Generated bundled specs
│   ├── combined-source.yaml     # Source for combined spec
│   └── openapi.yaml             # Generated (do not edit)
│
├── .redocly.yaml                # Redocly configuration
└── Makefile                     # Build targets
```

## Quick Start

### Install Tools

```bash
make install
```

This installs:
- `@redocly/cli` - Linting and bundling
- `oapi-codegen` - Go code generation
- `@openapitools/openapi-generator-cli` - TypeScript generation

### Common Commands

```bash
# Lint all specifications
make lint

# Bundle all specs into single files
make bundle-all

# Generate Go types for all services
make codegen-go

# Generate TypeScript client for frontend
make codegen-ts

# Start documentation server
make serve
```

### View Documentation

```bash
# Combined API docs (all services)
make serve                    # http://localhost:8080

# Individual service docs
make serve-tournament         # http://localhost:8081
make serve-matchmaking        # http://localhost:8082
make serve-game-management    # http://localhost:8083
```

## Adding New Endpoints

1. **Add path definition** in `services/<service>/paths/<feature>.yaml`
2. **Add schemas** in `services/<service>/schemas/` if needed
3. **Reference in openapi.yaml** under `paths:`
4. **Run lint** to validate: `make lint`
5. **Generate code**: `make codegen-go codegen-ts`

### Cross-Reference Pattern

Reference shared components:
```yaml
$ref: "../../../shared/components/responses.yaml#/responses/BadRequest"
```

Reference service-specific schemas:
```yaml
$ref: "../schemas/Tournament.yaml#/Tournament"
```

## Code Generation

### Go Types

Generated files are placed in `services/<service>/internal/openapi/types.gen.go`.

```bash
make codegen-go-tournament
make codegen-go-matchmaking
make codegen-go-game-management
```

### TypeScript Client

Generated client is placed in `frontends/winspire-app/src/shared/api/generated/`.

```bash
make codegen-ts
```

## Root Makefile Targets

From the repository root:

```bash
make openapi-install   # Install tools
make openapi-lint      # Lint specs
make openapi-bundle    # Bundle all specs
make openapi-codegen   # Generate Go + TypeScript
make openapi-serve     # Start docs server
make openapi-all       # Lint + Bundle + Codegen
```
