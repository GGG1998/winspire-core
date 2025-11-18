# Quickstart: Auth Service Development

**Date**: 2025-01-18  
**Feature**: Supabase Authentication Integration

This guide provides a quick start for developing the authentication service.

---

## Prerequisites

- Go 1.25.4 or later
- Docker and Docker Compose (for local Supabase)
- PostgreSQL client (optional, for direct DB access)
- Make (for running build commands)

---

## Local Development Setup

### 1. Start Supabase Locally

```bash
# Navigate to supabase directory
cd supabase

# Start Supabase services
docker-compose up -d

# Wait for services to be ready (check logs)
docker-compose logs -f
```

Supabase will be available at:
- Kong API Gateway: 
  - External (host): `http://localhost:8000` (routes all services)
  - Internal (Docker): `http://kong:8000` (port 8000)
- Auth API: 
  - External: `http://localhost:8000/auth/v1/` (via Kong)
  - Internal: `http://auth:9999/` (GoTrue on port 9999)
- REST API: 
  - External: `http://localhost:8000/rest/v1/` (via Kong)
  - Internal: `http://rest:3000/` (PostgREST)
- Studio: `http://localhost:8000`
- Database: `postgresql://postgres:postgres@localhost:5432/postgres`

**Note**: 
- Kong runs on port **8000** internally (`kong:8000` in Docker network)
- Kong is exposed externally on port **8000** (`localhost:8000` from host)
- Supabase's GoTrue auth service runs on port **9999** internally (`auth:9999` in Docker network)
- Auth endpoints via Kong: `/auth/v1/*` → `http://auth:9999/` (Kong routes internally)

### 2. Get Supabase Credentials

1. Open Supabase Studio: http://localhost:54323
2. Go to Settings → API
3. Copy:
   - Project URL: 
     - External: `http://localhost:8000` (Kong gateway, port 8000)
     - Internal: `http://kong:8000` (Kong gateway, port 8000 in Docker network)
   - Anon Key: (public anon key, used as `apikey` header)
   - Service Role Key: (keep secret, for admin operations)
   - JWT Secret: (for JWT validation)

**Note**: 
- Kong runs on port **8000** internally, exposed on **8000** externally
- When making API requests through Kong, include the `apikey` header with the Anon Key or Service Role Key

### 3. Create Project Structure

```bash
# From repository root
mkdir -p services/auth/{cmd/auth,internal/{handlers,services,models,queries},migrations,pkg}
mkdir -p libs/go/auth/{jwt,middleware,types}

# Initialize Go modules
cd services/auth
go mod init github.com/winspire/winspire-core/services/auth

cd ../../libs/go/auth
go mod init github.com/winspire/winspire-core/libs/go/auth

# Add to root go.work (if it exists, or create it)
cd ../../..
go work init services/auth libs/go/auth
```

### 4. Configure Kong for Custom Auth Service (Optional)

If you're building a custom auth service alongside Supabase's GoTrue, add it to `supabase/volumes/api/kong.yml`:

```yaml
services:
  ## Custom Auth Service routes
  - name: auth-service-v1
    _comment: 'Custom Auth Service: /auth-service/v1/* -> http://auth-service:8080/v1/*'
    url: http://auth-service:8080/
    routes:
      - name: auth-service-v1-all
        strip_path: true
        paths:
          - /auth-service/v1/
    plugins:
      - name: cors
      - name: key-auth
        config:
          hide_credentials: false
      - name: acl
        config:
          hide_groups_header: true
          allow:
            - admin
            - anon
```

**Note**: 
- Kong runs on port **8000** internally (`kong:8000` in Docker network), exposed on **8000** externally
- Supabase's built-in GoTrue auth service runs on port **9999** internally (`auth:9999` in Docker network)
- It's already configured in Kong: `/auth/v1/*` → `http://auth:9999/`
- External access (from host): `http://localhost:8000/auth/v1/*` (through Kong on port 8000)
- Internal access (from Docker): `http://kong:8000/auth/v1/*` (through Kong on port 8000) or `http://auth:9999/` (direct to GoTrue)
- Your custom auth service can run on a different port (e.g., 8080) and be routed through Kong
- Restart Kong after updating the configuration: `docker-compose restart kong` in the `supabase/` directory
- The custom service will be accessible at `http://localhost:8000/auth-service/v1/*` (external) or `http://kong:8000/auth-service/v1/*` (internal via Kong)

### 5. Configure Environment Variables

Create `services/auth/.env`:

```bash
# Supabase Configuration
SUPABASE_URL=http://localhost:8000
SUPABASE_ANON_KEY=your_anon_key_here
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key_here
SUPABASE_JWT_SECRET=your_jwt_secret_here

# Database Configuration
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/postgres

# Server Configuration
PORT=8080
ENV=development

# Note: 
# - Kong gateway: port 8000 internally (kong:8000), port 8000 externally (localhost:8000)
# - Supabase's GoTrue auth service: port 9999 internally (auth:9999 in Docker network)
# - External access: localhost:8000/auth/v1/* (through Kong)
# - Internal access: kong:8000/auth/v1/* (through Kong) or auth:9999/ (direct)
# - Your custom auth service can use port 8080 or any other available port

# OAuth Providers (get from provider dashboards)
DISCORD_CLIENT_ID=your_discord_client_id
DISCORD_CLIENT_SECRET=your_discord_client_secret
TWITCH_CLIENT_ID=your_twitch_client_id
TWITCH_CLIENT_SECRET=your_twitch_client_secret
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
FACEBOOK_CLIENT_ID=your_facebook_client_id
FACEBOOK_CLIENT_SECRET=your_facebook_client_secret
```

### 6. Install Dependencies

```bash
cd services/auth

# Install Go dependencies
go get github.com/gin-gonic/gin
go get github.com/kelseyhightower/envconfig
go get github.com/supabase-community/supabase-go
go get github.com/golang-jwt/jwt/v5
go get github.com/jackc/pgx/v5
go get github.com/golang-migrate/migrate/v4

# Install sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### 7. Run Database Migrations

```bash
cd services/auth

# Create initial migration
migrate create -ext sql -dir migrations -seq initial_schema

# Edit migrations/000001_initial_schema.up.sql with schema from data-model.md

# Run migrations
migrate -path migrations -database "$DATABASE_URL" up
```

### 8. Configure sqlc

Create `services/auth/sqlc.yaml`:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/queries"
    schema: "migrations"
    gen:
      go:
        package: "models"
        out: "internal/models"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: false
        emit_exact_table_names: false
```

### 9. Create SQL Queries

Create `services/auth/internal/queries/users.sql`:

```sql
-- name: GetUserRoles :many
SELECT r.id, r.name, r.description
FROM roles r
JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1;

-- name: GetUserPermissions :many
SELECT DISTINCT p.id, p.name, p.resource, p.action
FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
JOIN user_roles ur ON rp.role_id = ur.role_id
WHERE ur.user_id = $1;

-- name: CheckUserPermission :one
SELECT EXISTS(
  SELECT 1
  FROM permissions p
  JOIN role_permissions rp ON p.id = rp.permission_id
  JOIN user_roles ur ON rp.role_id = ur.role_id
  WHERE ur.user_id = $1 AND p.name = $2
) as has_permission;
```

Generate Go code:

```bash
sqlc generate
```

### 10. Build and Run

```bash
cd services/auth

# Build
go build -o bin/auth ./cmd/auth

# Run
./bin/auth

# Or run directly
go run ./cmd/auth
```

---

## Project Structure

```
services/auth/
├── cmd/auth/
│   └── main.go              # Application entry point
├── internal/
│   ├── handlers/            # HTTP handlers (Gin)
│   │   ├── auth.go
│   │   ├── oauth.go
│   │   ├── users.go
│   │   └── roles.go
│   ├── services/            # Business logic
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   └── role_service.go
│   ├── models/              # sqlc-generated models
│   ├── queries/             # SQL query files
│   │   ├── users.sql
│   │   ├── roles.sql
│   │   └── permissions.sql
│   └── config/              # Configuration
│       └── config.go
├── migrations/              # Database migrations
│   ├── 000001_initial_schema.up.sql
│   └── 000001_initial_schema.down.sql
├── pkg/                     # Public packages (if any)
├── go.mod
├── go.sum
├── sqlc.yaml
├── Makefile
└── .env                     # Environment variables (gitignored)

libs/go/auth/
├── jwt/
│   ├── validator.go         # JWT validation logic
│   └── parser.go
├── middleware/
│   └── auth.go              # JWT validation middleware
├── types/
│   └── user.go              # User context types
├── go.mod
└── go.sum
```

---

## Development Workflow

### 1. Adding a New Endpoint

1. Define endpoint in `contracts/auth-service.yaml` (OpenAPI spec)
2. Create handler in `internal/handlers/`
3. Implement business logic in `internal/services/`
4. Add SQL queries if needed in `internal/queries/`
5. Run `sqlc generate` to update models
6. Write tests
7. Update API documentation

### 2. Database Changes

1. Create migration: `migrate create -ext sql -dir migrations -seq migration_name`
2. Write `up` and `down` SQL in migration files
3. Update queries in `internal/queries/` if needed
4. Run `sqlc generate` to update models
5. Test migration: `migrate -path migrations -database "$DATABASE_URL" up/down`

### 3. Testing

```bash
# Run unit tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests (requires Supabase running)
go test -tags=integration ./...
```

### 4. Code Generation

```bash
# Generate sqlc models
sqlc generate

# Format code
go fmt ./...

# Run linter (if configured)
golangci-lint run
```

---

## API Testing

### Using curl with Kong Gateway

**From host machine**: All requests go through Kong at `http://localhost:8000` (external port).  
**From Docker containers**: Use `http://kong:8000` (internal port 8000).  
Include the `apikey` header for authentication.

```bash
# Set your Supabase anon key
export SUPABASE_ANON_KEY="your_anon_key_here"

# Register user (via Kong gateway - external access from host)
curl -X POST http://localhost:8000/auth/v1/signup \
  -H "Content-Type: application/json" \
  -H "apikey: $SUPABASE_ANON_KEY" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'

# Login (via Kong gateway)
curl -X POST http://localhost:8000/auth/v1/token?grant_type=password \
  -H "Content-Type: application/json" \
  -H "apikey: $SUPABASE_ANON_KEY" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'

# Get current user (with JWT token)
curl -X GET http://localhost:8000/auth/v1/user \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "apikey: $SUPABASE_ANON_KEY"

# Note: Custom auth service endpoints (when added to Kong)
# Will be accessible at: http://localhost:8000/auth-service/v1/*
```

**Important**: 
- **Kong gateway**: Runs on port **8000** internally (`kong:8000`), exposed on **8000** externally
- **Supabase GoTrue auth**: Runs on port **9999** internally (`auth:9999` in Docker network)
- **External access** (from host machine): `http://localhost:8000/auth/v1/*` (through Kong on port 8000)
- **Internal access** (from Docker containers): 
  - Via Kong: `http://kong:8000/auth/v1/*` (Kong routes to `http://auth:9999/`)
  - Direct: `http://auth:9999/` (bypass Kong, direct to GoTrue)
- Kong routes `/auth/v1/*` → `http://auth:9999/` internally
- Our custom auth service endpoints will need to be added to Kong configuration if building separately
- All requests through Kong must include the `apikey` header with Supabase anon key or service role key

### Using OpenAPI/Swagger

1. Install Swagger UI or use online editor
2. Import `contracts/auth-service.yaml`
3. Test endpoints interactively

---

## Integration with Other Services

### Using the Auth Library

Other services can import the auth library to validate JWTs:

```go
import (
    "github.com/winspire/winspire-core/libs/go/auth/middleware"
    "github.com/winspire/winspire-core/libs/go/auth/types"
)

// In your service
router := gin.Default()

// Add JWT validation middleware
router.Use(middleware.ValidateJWTMiddleware(authConfig))

// Access user context in handlers
func handler(c *gin.Context) {
    user := types.UserFromContext(c.Request.Context())
    // user.ID, user.Email, user.Roles available
}
```

### Calling Auth Service

**From outside Docker (host machine)**:
```go
// Validate token via Supabase GoTrue (through Kong gateway - external port)
resp, err := http.Post(
    "http://localhost:8000/auth/v1/user",  // Kong external port 8000
    "application/json",
    bytes.NewBuffer([]byte(`{"token": "..."}`)),
)
req.Header.Set("apikey", supabaseAnonKey)
req.Header.Set("Authorization", "Bearer YOUR_ACCESS_TOKEN")
```

**From inside Docker network**:
```go
// Option 1: Via Kong (internal port 8000)
resp, err := http.Post(
    "http://kong:8000/auth/v1/user",  // Kong internal port 8000
    "application/json",
    bytes.NewBuffer([]byte(`{"token": "..."}`)),
)
req.Header.Set("apikey", supabaseAnonKey)

// Option 2: Direct to GoTrue (bypass Kong)
resp, err := http.Post(
    "http://auth:9999/user",  // Direct to GoTrue on port 9999
    "application/json",
    bytes.NewBuffer([]byte(`{"token": "..."}`)),
)
```

**Note**: 
- **Kong**: Port **8000** internally (`kong:8000`), **8000** externally (`localhost:8000`)
- **GoTrue**: Port **9999** internally (`auth:9999` in Docker network)
- External access: `localhost:8000/auth/v1/*` (through Kong)
- Internal access: `kong:8000/auth/v1/*` (through Kong) or `auth:9999/` (direct)
- The custom auth service needs to be added to Kong configuration (`supabase/volumes/api/kong.yml`) to be accessible through the gateway

---

## Troubleshooting

### Supabase Connection Issues

- Check Supabase is running: `docker-compose ps` in `supabase/`
- Verify credentials in `.env`
- Check Supabase logs: `docker-compose logs -f`

### Database Migration Issues

- Ensure database is accessible: `psql "$DATABASE_URL"`
- Check migration status: `migrate -path migrations -database "$DATABASE_URL" version`
- Rollback if needed: `migrate -path migrations -database "$DATABASE_URL" down 1`

### JWT Validation Issues

- Verify JWT secret matches Supabase project settings
- Check token expiration
- Validate token format (should be Supabase JWT)

### sqlc Generation Issues

- Ensure SQL queries follow sqlc syntax
- Check `sqlc.yaml` configuration
- Verify database schema matches queries

---

## Next Steps

1. Review [data-model.md](./data-model.md) for database schema
2. Review [contracts/auth-service.yaml](./contracts/auth-service.yaml) for API specification
3. Review [research.md](./research.md) for technology decisions
4. Start implementing handlers and services
5. Write tests as you develop

---

## Resources

- [Supabase Go Client](https://github.com/supabase-community/supabase-go)
- [sqlc Documentation](https://docs.sqlc.dev/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [JWT Go Library](https://github.com/golang-jwt/jwt)
- [golang-migrate](https://github.com/golang-migrate/migrate)

