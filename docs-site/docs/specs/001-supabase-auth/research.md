# Research: Supabase Authentication Integration

**Date**: 2025-01-27  
**Feature**: Supabase Authentication Integration  
**Phase**: 0 - Research & Technology Decisions

## Research Questions

This document consolidates research findings for all "NEEDS CLARIFICATION" items identified in the implementation plan.

---

## 1. HTTP Framework Selection

**Question**: Which HTTP framework should we use for the auth service: Gin, Echo, or Chi?

**Research Findings**:

### Gin
- **Pros**: 
  - Most popular Go web framework (60k+ stars)
  - Excellent performance (fastest among the three)
  - Rich middleware ecosystem
  - Good documentation and community support
  - Built-in JSON binding and validation
- **Cons**: 
  - More opinionated structure
  - Slightly larger binary size
- **Performance**: ~40k req/s (benchmarks)

### Echo
- **Pros**: 
  - Lightweight and fast
  - Excellent documentation
  - Built-in middleware
  - Good for REST APIs
- **Cons**: 
  - Smaller community than Gin
  - Less third-party middleware
- **Performance**: ~35k req/s (benchmarks)

### Chi
- **Pros**: 
  - Minimal and composable
  - Standard library compatible
  - Very lightweight
  - Good for microservices
- **Cons**: 
  - Less built-in features (need to add more dependencies)
  - Smaller ecosystem
- **Performance**: ~30k req/s (benchmarks)

**Decision**: **Gin**  
**Rationale**: 
- Best performance for high-throughput auth service (1,000 concurrent requests requirement)
- Largest ecosystem and community support (important for team onboarding)
- Built-in features reduce boilerplate (JSON binding, validation)
- Proven in production at scale
- Good balance of features and performance

**Alternatives Considered**: Echo (good but smaller ecosystem), Chi (too minimal for our needs)

---

## 2. Configuration Management

**Question**: Should we use Viper or envconfig for configuration management?

**Research Findings**:

### Viper
- **Pros**: 
  - Supports multiple config sources (env vars, files, flags, etc.)
  - Automatic type conversion
  - Watch for config file changes
  - Very flexible
- **Cons**: 
  - More complex API
  - Heavier dependency
  - Can be overkill for simple env-based config

### envconfig
- **Pros**: 
  - Simple and lightweight
  - Type-safe struct-based configuration
  - Good for 12-factor app patterns (env vars)
  - Minimal dependencies
- **Cons**: 
  - Only supports environment variables
  - Less flexible for complex config scenarios

**Decision**: **envconfig**  
**Rationale**: 
- Microservices should follow 12-factor app principles (config via env vars)
- Simpler API reduces cognitive load
- Type-safe struct configuration prevents runtime errors
- Lighter dependency footprint
- Can add Viper later if needed for file-based config

**Alternatives Considered**: Viper (too complex for current needs, can add later if needed)

---

## 3. Contract Testing Approach

**Question**: What approach should we use for contract testing?

**Research Findings**:

### Options:
1. **Pact** - Consumer-driven contract testing
2. **OpenAPI contract testing** - Validate against OpenAPI spec
3. **Manual contract validation** - Test against documented API contracts
4. **Postman/Newman** - Collection-based contract tests

**Decision**: **OpenAPI contract testing with manual validation initially**  
**Rationale**: 
- OpenAPI specs will be generated for API contracts
- Can use tools like `openapi-validator` or `spectral` to validate requests/responses
- Simpler to start with, can add Pact later if needed for consumer-driven contracts
- Aligns with API-first design approach

**Implementation**: 
- Generate OpenAPI specs in `contracts/` directory
- Use Go libraries like `go-openapi` for runtime validation
- Add contract tests that validate request/response against OpenAPI schema

**Alternatives Considered**: Pact (good for complex microservices, but adds complexity we don't need yet)

---

## 4. JWT Validation Library Structure

**Question**: Should JWT validation be in a separate library (`libs/go/auth/`) initially or in the service?

**Research Findings**:

### Separate Library (libs/go/auth/)
- **Pros**: 
  - Reusable across services immediately
  - Clear separation of concerns
  - Can be versioned independently
  - Other services can import without depending on auth service
- **Cons**: 
  - More upfront complexity
  - Need to define stable API early
  - Two modules to maintain from start

### In-Service Initially
- **Pros**: 
  - Simpler initial setup
  - Can refactor to library after patterns emerge
  - Less upfront design needed
- **Cons**: 
  - Other services would need to depend on auth service code
  - Harder to extract later (coupling)
  - Violates DRY if multiple services need JWT validation

**Decision**: **Separate library (`libs/go/auth/`) from the start**  
**Rationale**: 
- Multiple services will need JWT validation (per microservices architecture)
- Better to establish the pattern early than refactor later
- Independent versioning allows library to evolve separately
- Clear API boundaries force better design
- Constitution requires shared libraries for common functionality

**Alternatives Considered**: In-service initially (simpler but creates technical debt)

---

## 5. Supabase Auth Integration Patterns in Go

**Question**: What are best practices for integrating Supabase Auth in Go microservices?

**Research Findings**:

### Key Patterns:

1. **JWT Validation**:
   - Supabase issues JWTs signed with a secret key
   - Services validate JWTs using the JWT secret (from Supabase project settings)
   - Use `github.com/golang-jwt/jwt/v5` for JWT parsing and validation
   - Validate: signature, expiration, issuer, audience

2. **User Metadata**:
   - User data stored in Supabase `auth.users` table
   - Custom user metadata in `auth.users.user_metadata` JSONB field
   - Roles can be stored in `user_metadata.role` or separate `public.user_roles` table

3. **Service Architecture**:
   - Auth service acts as a proxy/wrapper around Supabase Auth API
   - Other services validate JWTs independently (no need to call auth service for every request)
   - Auth service provides endpoints for: registration, login, password reset, OAuth flows
   - Auth service provides user info endpoints for other services

4. **Database Schema**:
   - Use Supabase PostgreSQL for auth-related tables
   - Create `public.user_roles` table for RBAC (separate from Supabase auth.users)
   - Use sqlc to generate type-safe queries for role/permission lookups

**Decision**: **Hybrid approach**  
**Rationale**: 
- Use Supabase Auth for identity management (registration, login, OAuth)
- Build custom RBAC layer in our PostgreSQL schema (roles, permissions)
- Services validate JWTs independently using shared library
- Auth service provides user management and role assignment APIs

**Implementation Details**:
- JWT validation in `libs/go/auth/` using Supabase JWT secret
- Role/permission queries in `services/auth/` using sqlc
- Auth service exposes REST API for user operations
- Other services use `libs/go/auth/` middleware for JWT validation

---

## 6. sqlc Setup and Best Practices

**Question**: How should we structure SQL queries and sqlc configuration for the auth service?

**Research Findings**:

### sqlc Best Practices:

1. **Project Structure**:
   ```
   services/auth/
   ├── migrations/          # SQL migrations
   ├── internal/
   │   └── queries/        # SQL query files (.sql)
   │   └── models/         # Generated Go models (sqlc output)
   └── sqlc.yaml           # sqlc configuration
   ```

2. **Query Organization**:
   - One `.sql` file per domain concept (e.g., `users.sql`, `roles.sql`, `permissions.sql`)
   - Use named parameters (`:param_name`) for type safety
   - Use `-- name: FunctionName` comments to name generated functions

3. **Configuration**:
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
   ```

4. **Migration Management**:
   - Use `golang-migrate` or `atlas` for migration management
   - Store migrations in `services/auth/migrations/`
   - Version migrations with timestamps

**Decision**: **Use sqlc with pgx/v5 driver**  
**Rationale**: 
- Type-safe SQL queries prevent runtime errors
- Generated code is fast and efficient
- pgx/v5 is the recommended PostgreSQL driver for Go
- Clear separation: migrations define schema, queries define operations

**Implementation**:
- SQL migrations in `services/auth/migrations/`
- Query files in `services/auth/internal/queries/`
- Generated models in `services/auth/internal/models/`
- Use `golang-migrate` for migration management

---

## 7. Role-Based Access Control (RBAC) Patterns

**Question**: How should we implement RBAC in a microservices architecture with Supabase?

**Research Findings**:

### RBAC Architecture Patterns:

1. **Database-Driven RBAC**:
   - Store roles and permissions in PostgreSQL tables
   - `user_roles` table: user_id → role_id mapping
   - `roles` table: role definitions (Streamer, User, Admin, etc.)
   - `permissions` table: permission definitions
   - `role_permissions` table: role → permission mapping

2. **JWT Claims for Roles**:
   - Include role in JWT `user_metadata` or custom claims
   - Services can read role from JWT without DB lookup
   - Trade-off: JWT size vs. DB query performance

3. **Hybrid Approach**:
   - Store roles in database (source of truth)
   - Include role in JWT for fast access
   - Refresh JWT when roles change
   - Services can validate against DB for critical operations

**Decision**: **Database-driven RBAC with JWT role claims**  
**Rationale**: 
- Database is source of truth for roles/permissions
- Include role in JWT for performance (avoid DB lookup on every request)
- Auth service manages role assignments and JWT refresh
- Services validate roles from JWT, with optional DB validation for sensitive operations

**Schema Design**:
```sql
-- Roles table
CREATE TABLE roles (
  id UUID PRIMARY KEY,
  name VARCHAR(50) UNIQUE NOT NULL,  -- 'streamer', 'user', 'admin'
  description TEXT
);

-- User roles (many-to-many)
CREATE TABLE user_roles (
  user_id UUID REFERENCES auth.users(id),
  role_id UUID REFERENCES roles(id),
  assigned_at TIMESTAMP DEFAULT NOW(),
  PRIMARY KEY (user_id, role_id)
);

-- Permissions table
CREATE TABLE permissions (
  id UUID PRIMARY KEY,
  name VARCHAR(100) UNIQUE NOT NULL,  -- 'tournament:create', 'stream:manage'
  resource VARCHAR(50) NOT NULL
);

-- Role permissions (many-to-many)
CREATE TABLE role_permissions (
  role_id UUID REFERENCES roles(id),
  permission_id UUID REFERENCES permissions(id),
  PRIMARY KEY (role_id, permission_id)
);
```

**Implementation**:
- Auth service provides APIs for role assignment
- When roles change, issue new JWT with updated claims
- `libs/go/auth/` provides middleware to extract and validate roles from JWT
- Services use middleware to check role-based access

---

## 8. JWT Validation Patterns for Supabase JWTs

**Question**: How should services validate Supabase-issued JWTs?

**Research Findings**:

### Validation Requirements:

1. **JWT Structure**:
   - Supabase JWTs use HS256 (HMAC) or RS256 (RSA) signing
   - JWT secret available in Supabase project settings
   - Claims include: `sub` (user ID), `email`, `role`, `user_metadata`

2. **Validation Steps**:
   - Verify signature using JWT secret
   - Check expiration (`exp` claim)
   - Validate issuer (`iss` claim) - should be Supabase project URL
   - Validate audience (`aud` claim) - should match expected audience
   - Extract user ID and role from claims

3. **Middleware Pattern**:
   - HTTP middleware that validates JWT from `Authorization: Bearer <token>` header
   - Extract user context (ID, email, role) and attach to request context
   - Return 401 if invalid, 403 if role insufficient

**Decision**: **Middleware-based JWT validation in shared library**  
**Rationale**: 
- Consistent validation logic across all services
- Reusable middleware reduces code duplication
- Type-safe user context extraction
- Easy to test and maintain

**Implementation**:
- `libs/go/auth/middleware.go` - JWT validation middleware
- `libs/go/auth/jwt.go` - JWT parsing and validation logic
- `libs/go/auth/types.go` - User context types (UserID, Role, etc.)
- Services use middleware: `router.Use(auth.ValidateJWTMiddleware(config))`

---

## Summary of Decisions

| Question | Decision | Rationale |
|----------|----------|-----------|
| HTTP Framework | **Gin** | Best performance, largest ecosystem, proven at scale |
| Config Management | **envconfig** | Simple, type-safe, follows 12-factor principles |
| Contract Testing | **OpenAPI validation** | Aligns with API-first design, simpler to start |
| JWT Library Structure | **Separate library** | Reusable, independent versioning, better architecture |
| Supabase Integration | **Hybrid (Auth + RBAC)** | Use Supabase for identity, custom RBAC layer |
| sqlc Setup | **pgx/v5 + migrations** | Type-safe queries, standard Go patterns |
| RBAC Pattern | **DB-driven + JWT claims** | Database as source of truth, JWT for performance |
| JWT Validation | **Middleware in library** | Reusable, consistent, type-safe |

---

## Next Steps

All "NEEDS CLARIFICATION" items have been resolved. Proceed to Phase 1:
- Generate data model documentation
- Create API contracts (OpenAPI specs)
- Create quickstart guide
- Update agent context

