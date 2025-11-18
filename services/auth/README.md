# Auth Service

Authentication service for Winspire platform using Supabase Auth.

> **Note:** The auth service relies entirely on Supabase Auth APIs and does not require a direct connection to PostgreSQL. A local database is still managed by Supabase, but this service communicates only with Supabase's HTTP interfaces.

## Quick Start

### 1. Prerequisites

- Go 1.25.4+
- Docker & Docker Compose
- Supabase running locally

### 2. Setup

```bash
# 1. Start Supabase
cd ../../supabase
docker-compose up -d

# 2. Get Supabase credentials
# Open http://localhost:8000 (Supabase Studio)
# Go to Settings → API
# Copy: URL, Anon Key, Service Role Key, JWT Secret

# 3. Create .env file
cd ../services/auth
cp .env.example .env
# Edit .env and add your Supabase credentials
# Required values:
#   SUPABASE_URL
#   SUPABASE_ANON_KEY
#   SUPABASE_SERVICE_ROLE_KEY
#   SUPABASE_JWT_SECRET
#   PORT (optional, defaults to 8080)
#   ENV  (optional, defaults to development)

# 4. Install dependencies
go mod tidy

# 5. Run the service
make run
# or
go run ./cmd/auth
```

### 3. Test the Service

```bash
# Run automated test script
./test-endpoints.sh

# Or test manually (see TESTING.md for details)
curl http://localhost:8080/health
```

## Available Endpoints

- `GET /health` - Health check
- `POST /v1/auth/register` - User registration
- `GET /v1/auth/verify` - Email verification
- `POST /v1/auth/login` - User login
- `POST /v1/auth/refresh` - Refresh access token
- `POST /v1/auth/logout` - User logout (protected)
- `POST /v1/auth/password/reset` - Request password reset
- `POST /v1/auth/password/reset/confirm` - Confirm password reset
- `GET /v1/auth/oauth/{provider}` - Initiate OAuth flow
- `GET /v1/auth/oauth/{provider}/callback` - OAuth callback

## Documentation

- [TESTING.md](./TESTING.md) - Detailed testing guide with curl examples
- [../specs/001-supabase-auth/quickstart.md](../specs/001-supabase-auth/quickstart.md) - Development guide

## Make Commands

```bash
make build          # Build the service
make test           # Run tests
make run            # Run the service
make clean          # Clean build artifacts
```

