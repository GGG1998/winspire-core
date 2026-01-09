# Debugging State Machines & Multi-Layer Changes

## Lesson Learned: Check ALL Layers

When modifying state machine flows (e.g., match status transitions), a fix in one layer is NOT enough. You must verify changes across the entire stack:

```
HTTP Handler → Service → Repository → SQL Query → Database Constraints
     ↓            ↓           ↓            ↓              ↓
   Go code     Go code     Go code    *.sql files    migrations/
```

### Real Example: Match State Flow Bug

**Problem**: Changed flow from `pending → ready → loading → started` to `pending → loading → ready → started`

**What happened**:
1. Fixed Go domain code (`match.go`) - tests passed
2. Deployed → still 500 error
3. Found SQL query had hardcoded `AND status = 'loading'`
4. Fixed SQL, regenerated SQLC → still 500 error
5. Found database CHECK constraint `check_loading_requires_both_ready` blocking new flow
6. Dropped constraint → finally worked

**Lesson**: Each layer had its own enforcement of the old flow. Fixing one layer wasn't enough.

## Checklist for State Machine Changes

- [ ] **Domain code** (`internal/domain/*.go`) - state transition logic, `validTransitions` map
- [ ] **Service layer** (`internal/application/*.go`) - business logic using states
- [ ] **SQL queries** (`internal/store/*.sql`) - WHERE clauses with status checks
- [ ] **Generated SQLC** (`internal/store/sqlc/*.go`) - run `make sqlc` after SQL changes
- [ ] **Database migrations** (`migrations/*.sql`) - CHECK constraints, triggers
- [ ] **Comments/docs** - update state flow documentation

## Finding Database Constraints

```bash
# Search for CHECK constraints in migrations
grep -r "CHECK\|CONSTRAINT" services/<name>/migrations/

# Find all constraints for a specific table
grep -rn "tournament_matches" services/matchmaking/migrations/ | grep -i constraint

# Search for status-related constraints
grep -rn "status" services/<name>/migrations/ | grep -i "check\|constraint"
```

## Debugging Production Errors

### 1. Check AWS CloudWatch Logs First

```bash
# Recent errors
aws logs filter-log-events \
  --log-group-name /ecs/dev/<service> \
  --filter-pattern "ERROR" \
  --start-time $(date -v-30M +%s)000 \
  --region eu-central-1

# Specific endpoint
aws logs filter-log-events \
  --log-group-name /ecs/dev/matchmaking \
  --filter-pattern "game-loaded" \
  --start-time $(date -v-10M +%s)000 \
  --region eu-central-1
```

### 2. Trace Full Execution Path

Before declaring a fix complete:
1. HTTP handler receives request
2. Service method is called
3. Repository/store method executes SQL
4. Database applies constraints
5. Response returned

### 3. Verify Generated Code

```bash
# After changing SQL queries
cd services/<name>
make sqlc

# Check the generated file
cat internal/store/sqlc/<query_file>.go | grep -A 20 "YourQueryName"
```

### 4. Check Migration Status

```sql
-- Check applied migrations
SELECT * FROM atlas_schema_revisions ORDER BY applied DESC LIMIT 10;
```

## Common Pitfalls

| Symptom | Likely Cause | Solution |
|---------|--------------|----------|
| Go tests pass, production 500 | SQL query has old logic | Check `*.sql` files, regenerate SQLC |
| SQL looks correct, still fails | Database constraint blocking | Search migrations for CHECK constraints |
| Fix deployed but old error | Service not restarted | Check ECS deployment, force redeploy |
| Migration fails midway | Partial migration state | Check `atlas_schema_revisions`, manual fix |

## Unit Tests vs Integration Tests

**Unit tests in Go domain layer DON'T test:**
- SQL query logic (hardcoded WHERE clauses)
- Database constraints (CHECK, FOREIGN KEY, TRIGGERS)
- Transaction behavior
- Concurrent access patterns

**For state machine changes, always:**
1. Write unit tests for domain logic
2. Test SQL queries against real database
3. Verify no blocking constraints exist

## Quick Commands Reference

```bash
# Search for state-related code
grep -rn "MatchStatus\|status" services/matchmaking/internal/

# Find all SQL files
find services/matchmaking -name "*.sql" -type f

# Check recent migrations
ls -la services/matchmaking/migrations/ | tail -10

# Run migration on dev database (Atlas)
cd services/matchmaking && make migrate

# Direct constraint drop (emergency)
psql "$DATABASE_URL" -c "ALTER TABLE tournament_matches DROP CONSTRAINT IF EXISTS constraint_name;"
```
