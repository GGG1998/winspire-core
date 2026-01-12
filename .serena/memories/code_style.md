# Code Style and Conventions

## Go
- Gin framework for HTTP handlers
- Dependencies via struct injection
- Validation via Gin binding tags (`binding:"required,min=2,max=100"`)
- Use `c.Request.Context()` for context propagation
- Type conversions: `uuidToPgtype()`, `pgtypeToUUID()`, `stringToPgtypeText()`

## TypeScript/React
- Zod for validation schemas, export types via `z.infer<typeof schema>`
- React Hook Form with `zodResolver`
- React Query for server state
- Import UI from `@/shared/components/ui/`
- Feature structure: api/, components/, hooks/, schemas.ts, types.ts, constants.ts

## SQLC Query Annotations
```sql
-- name: GetByID :one      -- Returns single row
-- name: List :many        -- Returns []Row
-- name: Create :exec      -- No return value
-- name: Update :execrows  -- Returns rows affected
```
