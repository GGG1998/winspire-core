# Tasks: Supabase Authentication Integration

**Input**: Design documents from `/specs/001-supabase-auth/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Tests**: Tests are OPTIONAL per specification - not included in initial implementation. Can be added later if requested.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Modular monorepo (winspire-core)**: 
  - Services: `services/auth/cmd/`, `services/auth/internal/`
  - Libraries: `libs/go/auth/`
  - Migrations: `services/auth/migrations/`
  - Queries: `services/auth/internal/queries/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create project structure: `services/auth/` and `libs/go/auth/` directories per implementation plan
- [X] T002 [P] Initialize Go module for auth service: `cd services/auth && go mod init github.com/winspire/winspire-core/services/auth`
- [X] T003 [P] Initialize Go module for auth library: `cd libs/go/auth && go mod init github.com/winspire/winspire-core/libs/go/auth`
- [X] T004 Create root `go.work` file and add both modules: `go work init services/auth libs/go/auth`
- [X] T005 [P] Create directory structure: `services/auth/cmd/auth/`, `services/auth/internal/{handlers,services,config,queries}`, `services/auth/migrations/`
- [X] T006 [P] Create directory structure: `libs/go/auth/{jwt,middleware,types}/`
- [X] T007 [P] Create `services/auth/Makefile` with build, test, and run targets
- [X] T008 [P] Create `services/auth/.env.example` with required environment variables
- [X] T009 [P] Create `services/auth/.gitignore` to exclude binaries and .env files

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

-- Note: We rely entirely on Supabase's built-in structures:
-- - auth.users: User accounts (no custom tables needed)
-- - auth.identities: OAuth provider links (managed by Supabase)
-- - Password reset: Handled internally by Supabase
-- - Sessions: JWT-based, stateless
-- No database migrations needed at this time.
- [X] T018 [P] Create configuration package: `services/auth/internal/config/config.go` using envconfig for Supabase URL, keys, JWT secret, database URL, OAuth credentials
- [X] T019 [P] Create error handling: `services/auth/internal/errors/errors.go` with custom error types and error response formatting
- [X] T020 [P] Create logging setup: `services/auth/internal/logger/logger.go` with structured logging configuration
- [X] T021 [P] Create database connection: `services/auth/internal/db/db.go` with pgx/v5 connection pool setup
- [X] T022 [P] Create Supabase client: `services/auth/internal/supabase/client.go` with Supabase Go client initialization
- [X] T023 [P] Create JWT validator: `libs/go/auth/jwt/validator.go` with Supabase JWT validation logic (signature, expiration, issuer, audience)
- [X] T024 [P] Create JWT parser: `libs/go/auth/jwt/parser.go` with JWT token parsing and claims extraction
- [X] T025 [P] Create user context types: `libs/go/auth/types/user.go` with UserID, Email, Role, UserContext structs
- [X] T026 [P] Create JWT middleware: `libs/go/auth/middleware/auth.go` with ValidateJWTMiddleware for Gin that extracts user context
- [X] T027 Create main application: `services/auth/cmd/auth/main.go` with Gin router setup, health check endpoint, and server startup
- [X] T028 [P] Create health handler: `services/auth/internal/handlers/health.go` with health check endpoint implementation

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - User Registration and Email Verification (Priority: P1) 🎯 MVP

**Goal**: Users can create accounts with email and password, receive verification emails, and verify their email to access the platform

**Independent Test**: Create a new account with a valid email, receive a verification email, click the verification link, and successfully log in

### Implementation for User Story 1

- [X] T029 [US1] Create registration service: `services/auth/internal/services/registration.go` with RegisterUser function that calls Supabase Auth SignUp
- [X] T030 [US1] Create registration handler: `services/auth/internal/handlers/auth.go` with Register endpoint that validates request, calls registration service, and returns auth response
- [X] T031 [US1] Add registration route: Update `services/auth/cmd/auth/main.go` to register POST `/v1/auth/register` endpoint
- [X] T032 [US1] Implement email verification handler: `services/auth/internal/handlers/auth.go` with VerifyEmail endpoint that calls Supabase Auth to verify email token
- [X] T033 [US1] Add email verification route: Update `services/auth/cmd/auth/main.go` to register GET `/v1/auth/verify` endpoint
- [X] T034 [US1] Add request validation: Create validation functions in `services/auth/internal/handlers/auth.go` for RegisterRequest (email format, password strength)
- [X] T035 [US1] Add error handling: Handle Supabase errors (user exists, invalid email) and return appropriate HTTP status codes
- [X] T036 [US1] Add logging: Log registration attempts, successes, and failures in `services/auth/internal/services/registration.go`

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently - users can register and verify email

---

## Phase 4: User Story 2 - User Login with Email and Password (Priority: P1)

**Goal**: Users can log in with verified email and password, receive JWT tokens, and access protected features

**Independent Test**: Log in with valid credentials, receive a session token, and access protected resources

### Implementation for User Story 2

- [X] T037 [US2] Create login service: `services/auth/internal/services/auth.go` with LoginUser function that calls Supabase Auth SignIn
- [X] T038 [US2] Create login handler: `services/auth/internal/handlers/auth.go` with Login endpoint that validates credentials and returns JWT tokens
- [X] T039 [US2] Add login route: Update `services/auth/cmd/auth/main.go` to register POST `/v1/auth/login` endpoint
- [X] T040 [US2] Implement email verification check: In login service, verify user email is confirmed before allowing login
- [X] T041 [US2] Add token refresh handler: `services/auth/internal/handlers/auth.go` with RefreshToken endpoint that refreshes access token using refresh token
- [X] T042 [US2] Add token refresh route: Update `services/auth/cmd/auth/main.go` to register POST `/v1/auth/refresh` endpoint
- [X] T043 [US2] Add error handling: Handle invalid credentials, unverified email, and return appropriate error messages
- [X] T044 [US2] Add logging: Log login attempts, successes, and failures in `services/auth/internal/services/auth.go`

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - users can register, verify email, and log in

---

## Phase 5: User Story 3 - Password Recovery (Priority: P2)

**Goal**: Users can request password reset, receive reset email, and set a new password to regain account access

**Independent Test**: Request password reset, receive reset email, click link, set new password, and log in with new password

### Implementation for User Story 3

- [X] T045 [US3] Create password reset service: `services/auth/internal/services/password.go` with RequestPasswordReset function that calls Supabase Auth ResetPasswordForEmail
- [X] T046 [US3] Create password reset request handler: `services/auth/internal/handlers/password.go` with RequestPasswordReset endpoint
- [X] T047 [US3] Add password reset request route: Update `services/auth/cmd/auth/main.go` to register POST `/v1/auth/password/reset` endpoint
- [X] T048 [US3] Create password reset confirmation service: `services/auth/internal/services/password.go` with ConfirmPasswordReset function that calls Supabase Auth UpdateUser
- [X] T049 [US3] Create password reset confirmation handler: `services/auth/internal/handlers/password.go` with ConfirmPasswordReset endpoint
- [X] T050 [US3] Add password reset confirmation route: Update `services/auth/cmd/auth/main.go` to register POST `/v1/auth/password/reset/confirm` endpoint
- [X] T051 [US3] Add token validation: Validate reset token expiration and usage in password reset confirmation
- [X] T052 [US3] Add error handling: Handle expired tokens, invalid tokens, and return appropriate error messages
- [X] T053 [US3] Add logging: Log password reset requests and completions in `services/auth/internal/services/password.go`

**Checkpoint**: At this point, User Stories 1, 2, AND 3 should all work independently

---

## Phase 6: User Story 4 - Social Authentication (OAuth) (Priority: P2)

**Goal**: Users can register or log in using OAuth providers (Discord/Twitch for Streamers, Google/Facebook for Users)

**Independent Test**: Click social login button, complete OAuth flow, and be authenticated into the platform

### Implementation for User Story 4

- [X] T054 [US4] Configure OAuth providers in Supabase: Set up Discord, Twitch, Google, Facebook OAuth apps and add credentials to Supabase dashboard (Manual configuration in Supabase dashboard)
- [X] T055 [US4] Create OAuth service: `services/auth/internal/services/oauth.go` with InitiateOAuthFlow function that generates OAuth URL
- [X] T056 [US4] Create OAuth initiation handler: `services/auth/internal/handlers/oauth.go` with InitiateOAuth endpoint that redirects to provider
- [X] T057 [US4] Add OAuth initiation route: Update `services/auth/cmd/auth/main.go` to register GET `/v1/auth/oauth/{provider}` endpoint
- [X] T058 [US4] Create OAuth callback service: `services/auth/internal/services/oauth.go` with HandleOAuthCallback function that exchanges code for tokens and creates/links user
- [X] T059 [US4] Create OAuth callback handler: `services/auth/internal/handlers/oauth.go` with OAuthCallback endpoint that processes OAuth response
- [X] T060 [US4] Add OAuth callback route: Update `services/auth/cmd/auth/main.go` to register GET `/v1/auth/oauth/{provider}/callback` endpoint
- [X] T061 [US4] Implement OAuth provider link storage: OAuth identities are managed by Supabase in auth.identities table (no custom table needed)
- [X] T062 [US4] Add provider validation: Validate that Streamers can only use Discord/Twitch and Users can only use Google/Facebook
- [X] T063 [US4] Implement account linking: Supabase handles automatic account linking by email (via auth.identities)
- [X] T064 [US4] Add error handling: Handle OAuth provider errors, cancelled flows, and return appropriate error messages
- [X] T065 [US4] Add logging: Log OAuth initiation, callbacks, and account linking in `services/auth/internal/services/oauth.go`

**Checkpoint**: At this point, User Stories 1, 2, 3, AND 4 should all work independently

---

## Phase 7: User Story 5 - Session Management and Logout (Priority: P2)

**Goal**: Users can log out to invalidate their session and must re-authenticate for future access

**Independent Test**: Log in, access protected resources, log out, and verify protected resources are no longer accessible

### Implementation for User Story 5

- [X] T066 [US5] Create logout service: `services/auth/internal/services/auth.go` with LogoutUser function that calls Supabase Auth SignOut
- [X] T067 [US5] Create logout handler: `services/auth/internal/handlers/auth.go` with Logout endpoint that invalidates session
- [X] T068 [US5] Add logout route: Update `services/auth/cmd/auth/main.go` to register POST `/v1/auth/logout` endpoint
- [X] T069 [US5] Add JWT middleware protection: Apply JWT validation middleware to logout endpoint to ensure user is authenticated
- [X] T070 [US5] Implement session expiration handling: JWT expiration is already handled by ValidateJWTMiddleware in jwt/validator.go
- [X] T071 [US5] Add error handling: Handle invalid tokens, expired sessions, and return appropriate error messages
- [X] T072 [US5] Add logging: Log logout events and session expirations in `services/auth/internal/services/auth.go`

**Checkpoint**: At this point, User Stories 1, 2, 3, 4, AND 5 should all work independently

---

## Phase 8: User Story 6 - Multi-Factor Authentication (MFA) (Priority: P3)

**Goal**: Users can enable MFA and must provide both password and authenticator code when logging in

**Independent Test**: Enable MFA, log out, log in with password and MFA code, and access platform

### Implementation for User Story 6

- [ ] T073 [US6] Create MFA service: `services/auth/internal/services/mfa.go` with EnableMFA function that generates TOTP secret and QR code
- [ ] T074 [US6] Create MFA enable handler: `services/auth/internal/handlers/mfa.go` with EnableMFA endpoint that returns QR code for authenticator setup
- [ ] T075 [US6] Add MFA enable route: Update `services/auth/cmd/auth/main.go` to register POST `/v1/auth/mfa/enable` endpoint
- [ ] T076 [US6] Create MFA verification service: `services/auth/internal/services/mfa.go` with VerifyMFASetup function that validates TOTP code during setup
- [ ] T077 [US6] Create MFA verification handler: `services/auth/internal/handlers/mfa.go` with VerifyMFASetup endpoint
- [ ] T078 [US6] Add MFA verification route: Update `services/auth/cmd/auth/main.go` to register POST `/v1/auth/mfa/verify` endpoint
- [ ] T079 [US6] Update login service: Modify `services/auth/internal/services/auth.go` LoginUser to check MFA status and require MFA code if enabled
- [ ] T080 [US6] Create MFA challenge handler: `services/auth/internal/handlers/mfa.go` with MFAChallenge endpoint that accepts MFA code after password verification
- [ ] T081 [US6] Add MFA challenge route: Update `services/auth/cmd/auth/main.go` to register POST `/v1/auth/mfa/challenge` endpoint
- [ ] T082 [US6] Add MFA storage: Store MFA secret in user metadata or separate table
- [ ] T083 [US6] Add error handling: Handle invalid MFA codes, expired challenges, and return appropriate error messages
- [ ] T084 [US6] Add logging: Log MFA enablement, verification attempts, and challenges in `services/auth/internal/services/mfa.go`

**Checkpoint**: At this point, all user stories should be independently functional

---

-- Phase 9 (RBAC) removed - we rely on Supabase's user_metadata for role storage
-- Roles can be stored in user_metadata.role field in JWT claims
-- No custom RBAC tables needed at this time

---

## Phase 10: User Management APIs

**Goal**: Provide APIs for users to view their profile and for services to validate tokens

**Note**: These endpoints support the authentication system and other microservices

- [ ] T094 Create user service: `services/auth/internal/services/user.go` with GetCurrentUser, GetUserByID functions
- [ ] T095 Create user handler: `services/auth/internal/handlers/users.go` with GetCurrentUser, GetUserByID endpoints
- [ ] T096 Add user routes: Update `services/auth/cmd/auth/main.go` to register GET `/v1/users/me`, GET `/v1/users/{userId}` endpoints
- [ ] T097 Create token validation service: `services/auth/internal/services/validation.go` with ValidateToken function that validates JWT and returns user info
- [ ] T098 Create token validation handler: `services/auth/internal/handlers/validation.go` with ValidateToken endpoint for other services to use
- [ ] T099 Add token validation route: Update `services/auth/cmd/auth/main.go` to register POST `/v1/validate` endpoint
- [ ] T100 Add error handling: Handle invalid tokens, missing users, and return appropriate error messages
- [ ] T101 Add logging: Log token validation requests in `services/auth/internal/services/validation.go`

---

## Phase 11: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T102 [P] Update OpenAPI contract: Ensure `contracts/auth-service.yaml` matches implemented endpoints
- [ ] T103 [P] Add request rate limiting: Implement rate limiting middleware for authentication endpoints to prevent brute force attacks
- [ ] T104 [P] Add CORS configuration: Configure CORS middleware in `services/auth/cmd/auth/main.go` for frontend access
- [ ] T105 [P] Add request validation: Create validation middleware for request bodies using Gin's binding
- [ ] T106 [P] Add response formatting: Standardize error response format across all endpoints
- [ ] T107 [P] Add audit logging: Log all authentication events (registration, login, password reset, OAuth, logout) for compliance
- [ ] T108 [P] Add metrics: Instrument endpoints with metrics (request count, latency, error rate) for observability
- [ ] T109 [P] Add health check improvements: Enhance health check to verify database and Supabase connectivity
- [ ] T110 [P] Add documentation: Update quickstart.md with actual endpoint examples and integration patterns
- [ ] T111 [P] Code cleanup: Review and refactor code for consistency, remove unused code
- [ ] T112 [P] Add Kong configuration: Update `supabase/volumes/api/kong.yml` to route `/auth-service/v1/*` to auth service
- [ ] T113 [P] Performance optimization: Review and optimize database queries, add connection pooling tuning
- [ ] T114 [P] Security hardening: Review security practices, add input sanitization, validate all inputs

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-8)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **RBAC (Phase 9)**: Removed - using Supabase user_metadata for roles
- **User Management (Phase 10)**: Depends on Foundational completion, can run in parallel with user stories
- **Polish (Phase 11)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories (uses Supabase Auth directly)
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 4 (P2)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 5 (P2)**: Can start after Foundational (Phase 2) - Requires JWT middleware from Foundational
- **User Story 6 (P3)**: Can start after Foundational (Phase 2) - Requires login service from US2

### Within Each User Story

- Services before handlers
- Handlers before routes
- Core implementation before error handling and logging
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, user stories can start in parallel (if team capacity allows)
- Services and handlers within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members
- User Management (Phase 10) can run in parallel with user stories
- All Polish tasks marked [P] can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all foundational components together:
Task: "Create registration service in services/auth/internal/services/registration.go"
Task: "Create registration handler in services/auth/internal/handlers/auth.go"
Task: "Add request validation in services/auth/internal/handlers/auth.go"
```

---

## Parallel Example: Foundational Phase

```bash
# Launch all foundational components together:
Task: "Create configuration package in services/auth/internal/config/config.go"
Task: "Create error handling in services/auth/internal/errors/errors.go"
Task: "Create logging setup in services/auth/internal/logger/logger.go"
Task: "Create database connection in services/auth/internal/db/db.go"
Task: "Create Supabase client in services/auth/internal/supabase/client.go"
Task: "Create JWT validator in libs/go/auth/jwt/validator.go"
Task: "Create JWT parser in libs/go/auth/jwt/parser.go"
Task: "Create user context types in libs/go/auth/types/user.go"
Task: "Create JWT middleware in libs/go/auth/middleware/auth.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Registration)
4. Complete Phase 4: User Story 2 (Login)
5. **STOP and VALIDATE**: Test User Stories 1 & 2 independently
6. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (Registration MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo (Login MVP!)
4. Add User Story 3 → Test independently → Deploy/Demo (Password Recovery)
5. Add User Story 4 → Test independently → Deploy/Demo (OAuth)
6. Add User Story 5 → Test independently → Deploy/Demo (Logout)
7. Add User Story 6 → Test independently → Deploy/Demo (MFA)
8. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (Registration)
   - Developer B: User Story 2 (Login)
   - Developer C: User Management (Phase 10)
3. After P1 stories complete:
   - Developer A: User Story 3 (Password Recovery)
   - Developer B: User Story 4 (OAuth)
   - Developer C: User Story 5 (Logout)
4. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
- All tasks use exact file paths per modular monorepo structure
- JWT validation library (`libs/go/auth/`) is built in Foundational phase and used by all stories
- We rely entirely on Supabase's built-in structures (auth.users, auth.identities)
- Roles stored in user_metadata.role field in JWT claims
- Supabase Auth handles core authentication; our service wraps it

---

## Summary

**Total Tasks**: 96 (reduced from 114 - removed migrations and RBAC)
- Phase 1 (Setup): 9 tasks
- Phase 2 (Foundational): 11 tasks (removed 8 migration/sqlc tasks)
- Phase 3 (US1 - Registration): 8 tasks
- Phase 4 (US2 - Login): 8 tasks
- Phase 5 (US3 - Password Recovery): 9 tasks
- Phase 6 (US4 - OAuth): 12 tasks
- Phase 7 (US5 - Logout): 7 tasks
- Phase 8 (US6 - MFA): 12 tasks
- Phase 9 (RBAC): Removed - using Supabase user_metadata
- Phase 10 (User Management): 8 tasks
- Phase 11 (Polish): 13 tasks

**MVP Scope**: Phases 1, 2, 3, 4 (Setup + Foundational + Registration + Login)

**Independent Test Criteria**:
- US1: Register → Verify Email → Login
- US2: Login → Access Protected Resource
- US3: Request Reset → Receive Email → Reset Password → Login
- US4: Click OAuth → Complete Flow → Authenticated
- US5: Login → Access Resource → Logout → Access Denied
- US6: Enable MFA → Login with Password + MFA → Authenticated

**Parallel Opportunities**: 
- 9 tasks in Setup phase
- 15 tasks in Foundational phase
- User stories can run in parallel after Foundational
- 13 tasks in Polish phase

