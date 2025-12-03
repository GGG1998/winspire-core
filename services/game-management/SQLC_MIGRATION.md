# SQLC Migration - Game Management Service

## Overview

This service has been migrated from **raw SQL queries** to [sqlc](https://docs.sqlc.dev/) for type-safe database operations.

## What Changed

### ❌ Before (Raw SQL)

```go
func (r *GameRepository) GetByID(ctx context.Context, id uuid.UUID) (*Game, error) {
    query := `
        SELECT id, game_integration_id, slug, name, description, logo_url, 
               storage_path, version, is_active, created_at, updated_at
        FROM games WHERE id = $1
    `
    
    var game Game
    var gameIntegrationID pgtype.UUID
    var description, logoURL pgtype.Text
    
    err := r.pool.QueryRow(ctx, query, id).Scan(
        &game.ID, &gameIntegrationID, &game.Slug, &game.Name,
        &description, &logoURL, &game.StoragePath, &game.Version,
        &game.IsActive, &game.CreatedAt, &game.UpdatedAt,
    )
    // ... manual error handling and type conversion
}
```

### ✅ After (SQLC)

**1. Define Query in SQL** (`internal/repository/queries.sql`):
```sql
-- name: GetGameByID :one
SELECT id, game_integration_id, slug, name, description, logo_url, 
       storage_path, version, is_active, created_at, updated_at
FROM games
WHERE id = $1;
```

**2. Generate Code**:
```bash
sqlc generate
```

**3. Use Generated Code**:
```go
func (r *GameRepository) GetByID(ctx context.Context, id uuid.UUID) (*Game, error) {
    sqlcGame, err := r.queries.GetGameByID(ctx, uuidToPgtype(id))
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, fmt.Errorf("game not found")
        }
        return nil, fmt.Errorf("get game by id: %w", err)
    }
    
    return sqlcGameToGame(sqlcGame), nil
}
```

## Files Structure

```
services/game-management/
├── internal/
│   └── repository/
│       ├── queries.sql              # SQL queries with sqlc annotations
│       ├── game.go                  # Business logic using sqlc
│       └── sqlstore/                # GENERATED - DO NOT EDIT
│           ├── db.go                # Queries struct
│           ├── models.go            # Type definitions
│           ├── querier.go           # Interface
│           └── queries.sql.go       # Query implementations
├── migrations/
│   └── 000001_init.sql             # Database schema
└── sqlc.yaml                        # sqlc configuration
```

## Available Queries

All queries are defined in `internal/repository/queries.sql`:

### Read Operations
- `GetGameByID` - Get game by UUID
- `GetGameBySlug` - Get active game by slug
- `GetGameByIntegrationID` - Get game by integration ID
- `ListGames` - List all active games
- `ListAllGames` - List all games (admin)

### Write Operations
- `CreateGame` - Create new game
- `UpdateGame` - Update game (optional parameters)
- `DeleteGame` - Soft delete (sets is_active = false)
- `HardDeleteGame` - Permanent delete

### Utility Queries
- `GameExists` - Check if game exists
- `GameSlugExists` - Check if slug is taken
- `CountGames` - Count active games
- `CountAllGames` - Count all games

## Type Conversions

The repository provides helper functions for pgtype conversions:

```go
// UUID conversions
uuidToPgtype(uuid.UUID) -> pgtype.UUID
pgtypeToUUID(pgtype.UUID) -> uuid.UUID
uuidToPgtypeNullable(*uuid.UUID) -> pgtype.UUID
pgtypeToUUIDNullable(pgtype.UUID) -> *uuid.UUID

// String conversions
stringToPgtypeText(*string) -> pgtype.Text
pgtypeTextToString(pgtype.Text) -> *string
optionalStringToPgtypeText(*string) -> pgtype.Text

// Bool conversions
optionalBoolToPgtype(*bool) -> pgtype.Bool

// Model conversion
sqlcGameToGame(sqlstore.Game) -> *repository.Game
```

## Usage Examples

### Creating a Game

```go
game, err := repo.Create(ctx, repository.CreateGameInput{
    Slug:        "packman",
    Name:        "Packman",
    Description: &desc,
    LogoURL:     &logoURL,
    StoragePath: "/games/packman",
    Version:     "1.0.0",
})
```

### Updating a Game (Optional Fields)

```go
err := repo.Update(ctx, repository.UpdateGameInput{
    ID:          gameID,
    Name:        &newName,      // Update name
    Description: &newDesc,      // Update description
    // Other fields nil - won't be updated
})
```

### Querying Games

```go
// Get by ID
game, err := repo.GetByID(ctx, gameID)

// Get by slug
game, err := repo.GetBySlug(ctx, "packman")

// List active games
games, err := repo.List(ctx)

// List all (admin)
allGames, err := repo.ListAll(ctx)
```

## Benefits of SQLC

### ✅ Type Safety
- Compile-time checks for SQL queries
- No runtime SQL parsing errors
- Autocomplete in IDE

### ✅ Performance
- No reflection overhead
- Direct scanning to structs
- Efficient parameterized queries

### ✅ Maintainability
- SQL queries in one place (`queries.sql`)
- Generated code is consistent
- Easy to add new queries

### ✅ Testability
- Mock-friendly interface (`querier.go`)
- Clear separation of concerns
- Easy to test business logic

## Development Workflow

### Adding a New Query

1. **Edit `queries.sql`**:
```sql
-- name: GetGamesByVersion :many
SELECT * FROM games 
WHERE version = $1 AND is_active = true;
```

2. **Generate code**:
```bash
make sqlc
# or
sqlc generate
```

3. **Use in repository**:
```go
func (r *GameRepository) GetByVersion(ctx context.Context, version string) ([]Game, error) {
    sqlcGames, err := r.queries.GetGamesByVersion(ctx, version)
    if err != nil {
        return nil, err
    }
    
    games := make([]Game, len(sqlcGames))
    for i, sg := range sqlcGames {
        games[i] = *sqlcGameToGame(sg)
    }
    return games, nil
}
```

### Modifying Schema

1. Create new migration in `migrations/`
2. Update `queries.sql` if needed
3. Run: `sqlc generate`
4. Update repository if needed

## Configuration

`sqlc.yaml`:
```yaml
version: "2"
sql:
  - engine: "postgresql"
    schema:
      - "migrations"
    queries:
      - "internal/repository/queries.sql"
    gen:
      go:
        package: "sqlstore"
        out: "internal/repository/sqlstore"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_interface: true
```

## Migration Notes

### Breaking Changes
- `Game.CreatedAt` and `Game.UpdatedAt` changed from `time.Time` to `string` (ISO 8601)
- This affects `GameResponse` in API handlers
- Frontend receives ISO 8601 strings directly

### Non-Breaking Changes
- All public repository methods maintain same signature
- Internal implementation uses sqlc
- Performance improved (no manual scanning)

## Testing

```bash
# Build to verify compilation
make build

# Run tests
make test

# Generate sqlc code
make sqlc
```

## Resources

- [SQLC Documentation](https://docs.sqlc.dev/)
- [Project SQLC Guide](../../services/competition/SQLC_GUIDE.md)
- [Golang HTTP Rules](.cursor/rules/golang_http.mdc)

