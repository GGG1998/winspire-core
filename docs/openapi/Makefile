.PHONY: all lint bundle bundle-all codegen-go codegen-ts serve serve-tournament serve-matchmaking serve-game-management build-docs clean install help

# Default target
all: lint bundle

# ============================================================================
# Help
# ============================================================================

help:
	@echo "OpenAPI Specification Management"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  install              Install required tools (redocly, oapi-codegen, openapi-generator)"
	@echo "  lint                 Lint all OpenAPI specifications"
	@echo "  bundle               Bundle combined specification"
	@echo "  bundle-all           Bundle all service specifications"
	@echo "  codegen-go           Generate Go types from OpenAPI specs"
	@echo "  codegen-ts           Generate TypeScript client from OpenAPI specs"
	@echo "  serve                Start Redoc documentation server (combined)"
	@echo "  serve-tournament     Start Redoc server for tournament service"
	@echo "  serve-matchmaking    Start Redoc server for matchmaking service"
	@echo "  serve-game-management Start Redoc server for game-management service"
	@echo "  build-docs           Build static HTML documentation"
	@echo "  clean                Remove generated files"
	@echo ""

# ============================================================================
# Tool Installation
# ============================================================================

install:
	@echo "Installing OpenAPI tools..."
	@command -v npm >/dev/null 2>&1 || { echo "npm is required but not installed. Aborting."; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "go is required but not installed. Aborting."; exit 1; }
	npm install -g @redocly/cli
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	npm install -g @openapitools/openapi-generator-cli
	@echo "Tools installed successfully!"
	@echo ""
	@echo "Verify installation:"
	@redocly --version
	@oapi-codegen --version
	@openapi-generator-cli version

# ============================================================================
# Validation & Linting
# ============================================================================

lint:
	@echo "Linting OpenAPI specifications..."
	redocly lint services/tournament/openapi.yaml
	redocly lint services/matchmaking/openapi.yaml
	redocly lint services/game-management/openapi.yaml
	redocly lint aggregated/combined-source.yaml
	@echo "All specifications are valid!"

lint-tournament:
	redocly lint services/tournament/openapi.yaml

lint-matchmaking:
	redocly lint services/matchmaking/openapi.yaml

lint-game-management:
	redocly lint services/game-management/openapi.yaml

# ============================================================================
# Bundling
# ============================================================================

bundle: bundle-combined
	@echo "Combined spec bundled: aggregated/openapi.yaml"

bundle-combined:
	@echo "Bundling combined specification..."
	redocly bundle aggregated/combined-source.yaml \
		--output aggregated/openapi.yaml \
		--dereferenced

bundle-tournament:
	@echo "Bundling tournament service spec..."
	@mkdir -p aggregated
	redocly bundle services/tournament/openapi.yaml \
		--output aggregated/tournament-openapi.yaml \
		--dereferenced

bundle-matchmaking:
	@echo "Bundling matchmaking service spec..."
	@mkdir -p aggregated
	redocly bundle services/matchmaking/openapi.yaml \
		--output aggregated/matchmaking-openapi.yaml \
		--dereferenced

bundle-game-management:
	@echo "Bundling game-management service spec..."
	@mkdir -p aggregated
	redocly bundle services/game-management/openapi.yaml \
		--output aggregated/game-management-openapi.yaml \
		--dereferenced

bundle-all: bundle-tournament bundle-matchmaking bundle-game-management bundle-combined
	@echo "All specifications bundled!"

# ============================================================================
# Go Code Generation (using oapi-codegen)
# ============================================================================

GO_OUT_BASE := ../../services

codegen-go: codegen-go-tournament codegen-go-matchmaking codegen-go-game-management
	@echo "Go code generation complete!"

codegen-go-tournament: bundle-tournament
	@echo "Generating Go types for tournament service..."
	@mkdir -p $(GO_OUT_BASE)/tournament/internal/openapi
	oapi-codegen \
		-generate types \
		-package openapi \
		-o $(GO_OUT_BASE)/tournament/internal/openapi/types.gen.go \
		aggregated/tournament-openapi.yaml

codegen-go-matchmaking: bundle-matchmaking
	@echo "Generating Go types for matchmaking service..."
	@mkdir -p $(GO_OUT_BASE)/matchmaking/internal/openapi
	oapi-codegen \
		-generate types \
		-package openapi \
		-o $(GO_OUT_BASE)/matchmaking/internal/openapi/types.gen.go \
		aggregated/matchmaking-openapi.yaml

codegen-go-game-management: bundle-game-management
	@echo "Generating Go types for game-management service..."
	@mkdir -p $(GO_OUT_BASE)/game-management/internal/openapi
	oapi-codegen \
		-generate types \
		-package openapi \
		-o $(GO_OUT_BASE)/game-management/internal/openapi/types.gen.go \
		aggregated/game-management-openapi.yaml

# ============================================================================
# TypeScript Code Generation (using openapi-generator)
# ============================================================================

TS_OUT_BASE := ../../frontends/winspire-app/src/shared/api/generated

codegen-ts: bundle-combined
	@echo "Generating TypeScript client for frontend..."
	@mkdir -p $(TS_OUT_BASE)
	openapi-generator-cli generate \
		-i aggregated/openapi.yaml \
		-g typescript-fetch \
		-o $(TS_OUT_BASE) \
		--additional-properties=supportsES6=true,typescriptThreePlus=true,npmName=@winspire/api-client
	@echo "TypeScript client generated in $(TS_OUT_BASE)"

# Alternative: Generate with @hey-api/openapi-ts (better React Query integration)
codegen-ts-hey-api: bundle-combined
	@echo "Generating TypeScript client with hey-api..."
	npx @hey-api/openapi-ts \
		-i aggregated/openapi.yaml \
		-o $(TS_OUT_BASE) \
		-c @hey-api/client-fetch
	@echo "TypeScript client generated in $(TS_OUT_BASE)"

# ============================================================================
# Documentation
# ============================================================================

serve:
	@echo "Starting Redoc documentation server..."
	@echo "Opening http://localhost:8080"
	redocly preview-docs aggregated/combined-source.yaml --port 8080

serve-tournament:
	@echo "Serving tournament API docs on http://localhost:8081"
	redocly preview-docs services/tournament/openapi.yaml --port 8081

serve-matchmaking:
	@echo "Serving matchmaking API docs on http://localhost:8082"
	redocly preview-docs services/matchmaking/openapi.yaml --port 8082

serve-game-management:
	@echo "Serving game-management API docs on http://localhost:8083"
	redocly preview-docs services/game-management/openapi.yaml --port 8083

build-docs: bundle-combined
	@echo "Building static documentation..."
	@mkdir -p dist
	redocly build-docs aggregated/openapi.yaml --output dist/index.html
	@echo "Static docs built: dist/index.html"

# ============================================================================
# Utilities
# ============================================================================

clean:
	@echo "Cleaning generated files..."
	rm -f aggregated/openapi.yaml
	rm -f aggregated/tournament-openapi.yaml
	rm -f aggregated/matchmaking-openapi.yaml
	rm -f aggregated/game-management-openapi.yaml
	rm -rf dist/
	@echo "Clean complete!"

# Stats about the API
stats: bundle-combined
	@echo "API Statistics:"
	redocly stats aggregated/openapi.yaml
