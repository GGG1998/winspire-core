---
targets:
  - '*'
root: false
globs:
  - services/*/internal/http/**/*.go
  - services/*/cmd/**/*.go
cursor:
  alwaysApply: false
  globs:
    - services/*/internal/http/**/*.go
    - services/*/cmd/**/*.go
---
# Golang REST HTTP API Guidelines

## Project Structure

Każdy mikroserwis HTTP powinien mieć strukturę:

```
services/<service-name>/
├── cmd/<service-name>/
│   ├── main.go              # Entry point
│   └── Dockerfile
├── internal/
│   ├── config/
│   │   └── config.go        # Konfiguracja z ENV
│   ├── http/
│   │   ├── server.go        # Router i middleware
│   │   └── handlers/
│   │       ├── types.go     # Request/Response DTOs
│   │       ├── helpers.go   # Pomocnicze funkcje (parseUUID, marshal)
│   │       └── <domain>.go  # Handlery pogrupowane domenowo
│   ├── repository/
│   │   └── <entity>.go      # Warstwa dostępu do danych (używa sqlc)
│   └── store/
│       └── sqlc/
│           ├── queries.sql  # SQL queries z adnotacjami sqlc
│           ├── db.go        # Wygenerowany przez sqlc
│           ├── models.go    # Wygenerowany przez sqlc
│           └── queries.sql.go # Wygenerowany przez sqlc
├── migrations/              # Database migrations (Atlas)
│   ├── 000001_initial.sql
│   └── 000002_add_table.sql
├── atlas.hcl                # Atlas configuration
├── go.mod
├── sqlc.yaml               # Konfiguracja sqlc
└── Makefile
```

## Database Migrations - Atlas

**KRYTYCZNE**: Używamy [Atlas](https://atlasgo.io/) do zarządzania migracjami bazy danych. Każdy serwis ma własne migracje w katalogu `migrations/`.

### Uruchamianie migracji

```bash
# Przejdź do katalogu serwisu
cd services/<service-name>

# Uruchom migracje
export DATABASE_URL="postgresql://user:pass@host:port/db"
make migrate-up

# Sprawdź status migracji
make migrate-status
```

### Konwencja plików migracji

- Numerowane sekwencyjnie: `000001_`, `000002_`, etc.
- Opisowa nazwa: `create_users_table.sql`, `add_tournaments_index.sql`
- Format: `{number}_{description}.sql`

### Przykład migracji

```sql
-- 000001_create_users_table.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

### Workflow z migracjami

1. **Utwórz nową migrację** w `services/<service>/migrations/000XXX_description.sql`
2. **Uruchom migracje**: `cd services/<service> && make migrate-up`
3. **Zweryfikuj schemat**: Sprawdź czy tabele zostały utworzone
4. **Zaktualizuj SQLC queries**: Jeśli dodałeś nowe tabele, dodaj queries w `internal/store/sqlc/queries.sql`
5. **Wygeneruj kod**: `make sqlc`

**WAŻNE**: Przed pierwszym uruchomieniem nowego serwisu, zawsze uruchom migracje!

## SQLC - Type-Safe SQL Code Generation

**KRYTYCZNE**: Używamy [sqlc](https://docs.sqlc.dev/) do generowania type-safe Go kodu z SQL queries. 
**NIE PISZ** surowych SQL queries w kodzie Go!

### Workflow SQLC

1. **Zdefiniuj schema** w `migrations/*.sql`
2. **Napisz queries** w `internal/store/sqlc/queries.sql` z adnotacjami sqlc
3. **Wygeneruj kod**: `make sqlc` lub `sqlc generate`
4. **Używaj wygenerowanego kodu** w repository

### Konfiguracja (sqlc.yaml)

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/store/sqlc/queries.sql"
    schema: "migrations/"
    gen:
      go:
        package: "sqlc"
        out: "internal/store/sqlc"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_empty_slices: true
```

### Queries File Structure (internal/store/sqlc/queries.sql)

```sql
-- name: GetUserByID :one
-- Retrieves a user by ID
SELECT id, name, email, created_at
FROM users
WHERE id = $1;

-- name: ListUsers :many
-- Lists all users with pagination
SELECT id, name, email, created_at
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateUser :one
-- Creates a new user
INSERT INTO users (name, email)
VALUES ($1, $2)
RETURNING id, name, email, created_at;

-- name: UpdateUser :exec
-- Updates user information
UPDATE users
SET name = $2, email = $3
WHERE id = $1;

-- name: DeleteUser :exec
-- Deletes a user
DELETE FROM users WHERE id = $1;
```

### Query Annotations

| Annotation | Zwraca | Użycie |
|------------|--------|---------|
| `:one` | Single row | SELECT pojedynczy rekord |
| `:many` | []Row | SELECT wiele rekordów |
| `:exec` | error only | INSERT/UPDATE/DELETE bez RETURNING |
| `:execrows` | rows affected + error | INSERT/UPDATE/DELETE z licznikiem |
| `:execresult` | sql.Result | Raw result |

### Nullable Fields

```sql
-- name: GetUser :one
SELECT 
    id, 
    name, 
    bio,           -- nullable text
    avatar_url     -- nullable text
FROM users
WHERE id = $1;
```

Generuje:
```go
type User struct {
    ID        pgtype.UUID        `json:"id"`
    Name      string             `json:"name"`
    Bio       pgtype.Text        `json:"bio"`        // nullable
    AvatarUrl pgtype.Text        `json:"avatar_url"` // nullable
}
```

### Optional Parameters (sqlc.narg)

```sql
-- name: UpdateUser :one
UPDATE users
SET 
    name = COALESCE(sqlc.narg('name'), name),
    bio = COALESCE(sqlc.narg('bio'), bio)
WHERE id = $1
RETURNING *;
```

Generuje:
```go
type UpdateUserParams struct {
    ID   pgtype.UUID `json:"id"`
    Name pgtype.Text `json:"name"` // optional
    Bio  pgtype.Text `json:"bio"`  // optional
}
```

### Generating Code

```bash
# Generate code from SQL
make sqlc

# Or directly
sqlc generate

# Verify queries (requires sqlc cloud)
sqlc verify
```

Wygenerowane pliki (NIE EDYTUJ):
- `internal/store/sqlc/db.go` - Queries struct i interface
- `internal/store/sqlc/models.go` - Type definitions
- `internal/store/sqlc/queries.sql.go` - Query implementations

## HTTP Framework - Gin

Używamy frameworka **Gin** (`github.com/gin-gonic/gin`) oraz współdzielonej biblioteki `libs/go/httpx`.

### Router Setup (server.go)

```go
package httpx

import (
    "context"
    "log/slog"

    "github.com/gin-gonic/gin"
    authmw "github.com/winspire/winspire-core/libs/go/auth/middleware"
    sharedhttp "github.com/winspire/winspire-core/libs/go/httpx"

    "github.com/winspire/<service>/internal/config"
    "github.com/winspire/<service>/internal/http/handlers"
)

// HealthFunc defines the signature for dependency health checks.
type HealthFunc func(ctx context.Context) error

// ServerDeps aggregates the dependencies required to construct the HTTP router.
type ServerDeps struct {
    Config      config.Config
    Logger      *slog.Logger
    HealthCheck HealthFunc
    // Dodaj repozytoria i inne zależności
}

// NewRouter builds a fully-configured Gin engine with middleware, auth, and handlers.
func NewRouter(deps ServerDeps) *gin.Engine {
    gin.SetMode(gin.ReleaseMode)
    router := gin.New()

    // Use shared httpx middleware
    httpCfg := sharedhttp.DefaultConfig()
    httpCfg.ServiceName = "<service-name>"

    router.Use(
        sharedhttp.Recovery(deps.Logger),
        sharedhttp.CORS(httpCfg),
        sharedhttp.SecurityHeaders(httpCfg),
        sharedhttp.RequestLogger(deps.Logger),
        sharedhttp.ErrorResponder(),
    )

    // Health check endpoint (no auth required)
    router.GET("/healthz", sharedhttp.HealthCheck(deps.HealthCheck))

    // API routes with auth
    api := router.Group("/v1")
    if deps.Config.HasAuth() {
        api.Use(authmw.ValidateJWTMiddleware(authmw.Config{
            JWTSecret: deps.Config.HostJWTSecret,
            Issuer:    deps.Config.HostJWTIssuer,
            Audience:  deps.Config.HostJWTAudience,
        }))
    }

    // Register handlers
    handlers.RegisterXxxRoutes(api, handlers.XxxDeps{...})

    // Admin routes (require admin role)
    adminGroup := api.Group("/admin")
    adminGroup.Use(sharedhttp.RequireAdminRole())
    handlers.RegisterAdminRoutes(adminGroup, handlers.AdminDeps{...})

    return router
}
```

## Handlers

### Struktura Handlera

```go
package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// XxxDeps contains dependencies for xxx handlers.
type XxxDeps struct {
    Repo *repository.XxxRepository
    // inne zależności
}

// RegisterXxxRoutes registers routes for xxx domain.
func RegisterXxxRoutes(group *gin.RouterGroup, deps XxxDeps) {
    // GET - List
    group.GET("/items", func(c *gin.Context) {
        items, err := deps.Repo.List(c.Request.Context())
        if err != nil {
            c.JSON(http.StatusInternalServerError, ErrorResponse{
                Error:   "failed to list items",
                Details: err.Error(),
            })
            return
        }
        c.JSON(http.StatusOK, ItemsListResponse{Items: items, Total: len(items)})
    })

    // GET - Single item
    group.GET("/items/:itemId", func(c *gin.Context) {
        itemID, err := uuid.Parse(c.Param("itemId"))
        if err != nil {
            c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid item ID"})
            return
        }

        item, err := deps.Repo.GetByID(c.Request.Context(), itemID)
        if err != nil {
            c.JSON(http.StatusNotFound, ErrorResponse{Error: "item not found"})
            return
        }

        c.JSON(http.StatusOK, itemToResponse(*item))
    })

    // POST - Create
    group.POST("/items", func(c *gin.Context) {
        var req CreateItemRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, ErrorResponse{
                Error:   "invalid request body",
                Details: err.Error(),
            })
            return
        }

        item, err := deps.Repo.Create(c.Request.Context(), req.ToInput())
        if err != nil {
            c.JSON(http.StatusInternalServerError, ErrorResponse{
                Error:   "failed to create item",
                Details: err.Error(),
            })
            return
        }

        c.JSON(http.StatusCreated, itemToResponse(*item))
    })

    // PATCH/PUT - Update
    group.PUT("/items/:itemId", func(c *gin.Context) {
        itemID, err := uuid.Parse(c.Param("itemId"))
        if err != nil {
            c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid item ID"})
            return
        }

        var req UpdateItemRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, ErrorResponse{
                Error:   "invalid request body",
                Details: err.Error(),
            })
            return
        }

        if err := deps.Repo.Update(c.Request.Context(), itemID, req.ToInput()); err != nil {
            c.JSON(http.StatusInternalServerError, ErrorResponse{
                Error:   "failed to update item",
                Details: err.Error(),
            })
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "item updated successfully"})
    })

    // DELETE
    group.DELETE("/items/:itemId", func(c *gin.Context) {
        itemID, err := uuid.Parse(c.Param("itemId"))
        if err != nil {
            c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid item ID"})
            return
        }

        if err := deps.Repo.Delete(c.Request.Context(), itemID); err != nil {
            c.JSON(http.StatusInternalServerError, ErrorResponse{
                Error:   "failed to delete item",
                Details: err.Error(),
            })
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "item deleted successfully"})
    })
}
```

### Helper Functions (helpers.go)

```go
package handlers

import (
    "encoding/json"
    "fmt"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// parseUUIDParam extracts and validates a UUID path parameter.
func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
    val := c.Param(name)
    id, err := uuid.Parse(val)
    if err != nil {
        c.Error(fmt.Errorf("invalid %s: %w", name, err))
        return uuid.UUID{}, false
    }
    return id, true
}

// marshalOrError marshals value to JSON, using fallback on nil.
func marshalOrError(c *gin.Context, v any, fallback []byte) (json.RawMessage, bool) {
    if v == nil {
        return json.RawMessage(fallback), true
    }
    buf, err := json.Marshal(v)
    if err != nil {
        c.Error(err)
        return nil, false
    }
    return json.RawMessage(buf), true
}
```

## Request/Response Types (types.go)

```go
package handlers

import (
    "time"

    "github.com/google/uuid"
)

// === Response Types ===

// ItemResponse represents an item in API responses.
type ItemResponse struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description *string   `json:"description,omitempty"`
    IsActive    bool      `json:"isActive"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}

// ItemsListResponse represents paginated list response.
type ItemsListResponse struct {
    Items []ItemResponse `json:"items"`
    Total int            `json:"total"`
}

// === Request Types ===

// CreateItemRequest represents the request body for creating an item.
type CreateItemRequest struct {
    Name        string  `json:"name" binding:"required,min=2,max=100"`
    Description *string `json:"description"`
}

// UpdateItemRequest represents the request body for updating an item.
type UpdateItemRequest struct {
    Name        *string `json:"name" binding:"omitempty,min=2,max=100"`
    Description *string `json:"description"`
    IsActive    *bool   `json:"isActive"`
}

// === Error Response ===

// ErrorResponse represents a standardized error response.
type ErrorResponse struct {
    Error   string `json:"error"`
    Details string `json:"details,omitempty"`
}

// === Nested Types ===

// StageStatus example of nested type with validation.
type StageStatus struct {
    StageID uuid.UUID `json:"stageId" binding:"required"`
    Status  string    `json:"status" binding:"required,oneof=PENDING ACTIVE COMPLETED"`
}
```

### Gin Binding Tags

| Tag | Opis |
|-----|------|
| `binding:"required"` | Pole wymagane |
| `binding:"omitempty"` | Waliduj tylko jeśli podane |
| `binding:"min=2,max=100"` | Długość stringa |
| `binding:"oneof=A B C"` | Dozwolone wartości |
| `binding:"dive"` | Waliduj elementy slice'a |
| `binding:"email"` | Format email |
| `binding:"url"` | Format URL |

## Configuration (config.go)

```go
package config

import (
    "fmt"
    "os"
    "strconv"
    "time"
)

// Config captures all runtime tunables provided via environment variables.
type Config struct {
    AppEnv        string
    ServicePort   int
    PostgresDSN   string
    RedisAddr     string
    ReadTimeout   time.Duration
    WriteTimeout  time.Duration
    ShutdownGrace time.Duration

    // JWT configuration
    HostJWTSecret   string
    HostJWTIssuer   string
    HostJWTAudience string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (Config, error) {
    cfg := Config{
        AppEnv:          valueOrDefault("APP_ENV", "development"),
        ServicePort:     intFromEnv("SERVICE_PORT", 8080),
        PostgresDSN:     valueOrDefault("POSTGRES_DSN", ""),
        RedisAddr:       valueOrDefault("REDIS_ADDR", "localhost:6379"),
        ReadTimeout:     durationFromEnv("HTTP_READ_TIMEOUT", 15*time.Second),
        WriteTimeout:    durationFromEnv("HTTP_WRITE_TIMEOUT", 15*time.Second),
        ShutdownGrace:   durationFromEnv("SHUTDOWN_GRACE", 10*time.Second),
        HostJWTSecret:   valueOrDefault("HOST_JWT_SECRET", ""),
        HostJWTIssuer:   valueOrDefault("HOST_JWT_ISSUER", ""),
        HostJWTAudience: valueOrDefault("HOST_JWT_AUDIENCE", ""),
    }

    if cfg.PostgresDSN == "" {
        return Config{}, fmt.Errorf("POSTGRES_DSN must be provided")
    }

    return cfg, nil
}

// HasAuth indicates whether JWT middleware should be enabled.
func (c Config) HasAuth() bool {
    return c.HostJWTSecret != "" && c.HostJWTIssuer != "" && c.HostJWTAudience != ""
}

func valueOrDefault(key, def string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return def
}

func intFromEnv(key string, def int) int {
    if val := os.Getenv(key); val != "" {
        if out, err := strconv.Atoi(val); err == nil {
            return out
        }
    }
    return def
}

func durationFromEnv(key string, def time.Duration) time.Duration {
    if val := os.Getenv(key); val != "" {
        if out, err := time.ParseDuration(val); err == nil {
            return out
        }
    }
    return def
}
```

## Main Entry Point (main.go)

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "github.com/winspire/<service>/internal/config"
    httpx "github.com/winspire/<service>/internal/http"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        panic(err)
    }

    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))

    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    // Initialize dependencies (DB, Redis, etc.)
    // ...

    router := httpx.NewRouter(httpx.ServerDeps{
        Config:    cfg,
        Logger:    logger,
        HealthCheck: func(ctx context.Context) error {
            // Check dependencies health
            return nil
        },
    })

    srv := &http.Server{
        Addr:              fmt.Sprintf(":%d", cfg.ServicePort),
        Handler:           router,
        ReadHeaderTimeout: cfg.ReadTimeout,
    }

    go func() {
        logger.Info("server starting", "addr", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Error("server error", "error", err)
            os.Exit(1)
        }
    }()

    <-ctx.Done()
    shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownGrace)
    defer cancel()
    _ = srv.Shutdown(shutdownCtx)
    logger.Info("server stopped")
}
```

## Repository Pattern with SQLC

**ZAWSZE** używaj wygenerowanego kodu sqlc w repository. **NIE PISZ** surowych SQL queries.

### Repository Structure

```go
package repository

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"
    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/winspire/<service>/internal/store/sqlc"
)

// Repository wraps sqlc-generated queries with business logic.
type Repository struct {
    pool    *pgxpool.Pool
    queries *sqlc.Queries
}

// NewRepository creates a new Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
    return &Repository{
        pool:    pool,
        queries: sqlc.New(pool),
    }
}

// Entity is the business domain model (can differ from sqlc.Entity).
type Entity struct {
    ID          uuid.UUID
    Name        string
    Description *string
    IsActive    bool
    CreatedAt   string
    UpdatedAt   string
}

// GetByID retrieves an entity by ID using sqlc-generated query.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Entity, error) {
    sqlcEntity, err := r.queries.GetEntityByID(ctx, uuidToPgtype(id))
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, fmt.Errorf("entity not found")
        }
        return nil, fmt.Errorf("get entity: %w", err)
    }

    return sqlcEntityToEntity(sqlcEntity), nil
}

// Create creates a new entity using sqlc-generated query.
func (r *Repository) Create(ctx context.Context, name string, desc *string) (*Entity, error) {
    params := sqlc.CreateEntityParams{
        Name:        name,
        Description: stringToPgtypeText(desc),
    }

    sqlcEntity, err := r.queries.CreateEntity(ctx, params)
    if err != nil {
        return nil, fmt.Errorf("create entity: %w", err)
    }

    return sqlcEntityToEntity(sqlcEntity), nil
}

// Update with transaction example.
func (r *Repository) UpdateWithRelations(ctx context.Context, id uuid.UUID, data UpdateData) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback(ctx)

    // Use transaction-aware queries
    qtx := r.queries.WithTx(tx)

    // Update entity
    err = qtx.UpdateEntity(ctx, sqlc.UpdateEntityParams{
        ID:   uuidToPgtype(id),
        Name: data.Name,
    })
    if err != nil {
        return fmt.Errorf("update entity: %w", err)
    }

    // Update related data
    // ...

    return tx.Commit(ctx)
}

// ============================================================================
// Type Converters (pgtype <-> standard Go types)
// ============================================================================

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
    return pgtype.UUID{Bytes: id, Valid: true}
}

func pgtypeToUUID(id pgtype.UUID) uuid.UUID {
    if !id.Valid {
        return uuid.UUID{}
    }
    return id.Bytes
}

func stringToPgtypeText(s *string) pgtype.Text {
    if s == nil {
        return pgtype.Text{Valid: false}
    }
    return pgtype.Text{String: *s, Valid: true}
}

func pgtypeTextToString(t pgtype.Text) *string {
    if !t.Valid {
        return nil
    }
    return &t.String
}

func sqlcEntityToEntity(e sqlc.Entity) *Entity {
    return &Entity{
        ID:          pgtypeToUUID(e.ID),
        Name:        e.Name,
        Description: pgtypeTextToString(e.Description),
        IsActive:    e.IsActive,
        CreatedAt:   e.CreatedAt.Time.Format(time.RFC3339),
        UpdatedAt:   e.UpdatedAt.Time.Format(time.RFC3339),
    }
}
```

## Kluczowe Zasady

1. **Zawsze używaj `c.Request.Context()`** - przekazuj context do repozytorium/serwisów
2. **Jeden plik handlers = jeden domain** - np. `game.go`, `tournament.go`
3. **Dependencies przez struktury** - nie używaj globalnych zmiennych
4. **Walidacja przez binding tags** - nie pisz ręcznej walidacji
5. **Standardowe kody HTTP**:
   - `200 OK` - sukces GET/PUT
   - `201 Created` - sukces POST
   - `202 Accepted` - async operacje
   - `400 Bad Request` - błąd walidacji
   - `401 Unauthorized` - brak/zły token
   - `403 Forbidden` - brak uprawnień
   - `404 Not Found` - zasób nie istnieje
   - `500 Internal Server Error` - błąd serwera
6. **Graceful shutdown** - zawsze obsługuj sygnały SIGINT/SIGTERM
7. **Health check** - endpoint `/healthz` bez auth

## Shared Libraries

- `libs/go/httpx` - middleware (Recovery, CORS, SecurityHeaders, RequestLogger, ErrorResponder, HealthCheck)
- `libs/go/auth/middleware` - JWT validation middleware
- `libs/go/auth/types` - UserContext, Role types

## Routing Conventions

- Wersjonowanie: `/v1/...`
- REST naming: `/v1/items`, `/v1/items/:itemId`
- Nested resources: `/v1/hosts/:hostId/tournaments/:tournamentId`
- Admin routes: `/v1/admin/...`

## Local Development - KRYTYCZNE

**ZAWSZE** po wprowadzeniu zmian w kodzie Go (handlers, middleware, config, repository, etc.) musisz przebudować i zrestartować serwisy:

```bash
cd platform/local

# Opcja 1: Pełny rebuild (zalecane przy większych zmianach)
make rebuild

# Opcja 2: Restart (szybsze, ale nie rebuiluje obrazów)
make restart
```

### Kiedy używać rebuild vs restart?

| Operacja | Użyj | Powód |
|----------|------|-------|
| `make rebuild` | Zmiany w kodzie Go | Musi przebudować Docker image z nowym kodem |
| `make rebuild` | Zmiany w Dockerfile | Wymaga rebuildu obrazu |
| `make rebuild` | Dodanie nowych dependencji (go.mod) | Musi zainstalować nowe pakiety |
| `make rebuild` | Zmiany w sqlc queries | Musi wygenerować nowy kod i przebudować |
| `make restart` | Zmiany w ENV variables | Wystarczy restart z nowym .env |
| `make restart` | Zmiany w config.toml (Supabase) | Restart wystarczy |

### Przydatne komendy Makefile

```bash
# Zobacz wszystkie dostępne komendy
make help

# Logi ze wszystkich serwisów
make logs

# Logi z konkretnego serwisu
make logs-service SERVICE=tournament

# Health check wszystkich serwisów
make test

# Połącz się z Redis CLI
make redis-cli

# Połącz się z Postgres
make db-connect

# Czyste środowisko (usuwa wszystko)
make clean
```

### Workflow przy rozwoju backendu

1. **Wprowadź zmiany w kodzie Go**
2. **Rebuild i restart**: `cd platform/local && make rebuild`
3. **Sprawdź logi**: `make logs-service SERVICE=<nazwa-serwisu>`
4. **Testuj endpoint**: `curl http://localhost/v1/...`
5. **W razie problemów**: Sprawdź logi lub zrób `make clean && make start` dla świeżego startu

### Znane problemy

- **Port zajęty**: Uruchom `make stop` przed `make start`
- **Stare obrazy**: Użyj `make clean` żeby wyczyścić wszystko
- **Zmiany nie widoczne**: Zawsze rób `make rebuild` po zmianach w Go
- **Database schema zmiany**: Wymaga `make clean` i ponownego `make start` dla świeżej bazy danych
