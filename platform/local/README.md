# Local Development with Traefik

Guide for running Winspire locally with Traefik as API Gateway (simulates AWS ALB).

## Architecture

```
Browser
  ↓
Traefik (localhost:80) - API Gateway
  ├─ /v1/stream/* → tournament:8089 (sticky sessions)
  ├─ /v1/cups/* → tournament:8089
  ├─ /v1/tournaments/* → tournament:8089
  ├─ /v1/games/* → game-management:8085
  └─ /v1/admin/* → game-management:8085
  ↓
Services (Docker Containers)
  ├─ tournament (1-2 instances)
  ├─ game-management
  └─ Redis (JWT cache + SSE Pub/Sub)
  ↓
Supabase Local (host.docker.internal:54321)
  └─ PostgreSQL (port 54322)
```

**Traefik Dashboard:** http://localhost:8080

## Prerequisites

1. **Docker & Docker Compose** installed
2. **Supabase CLI** installed and running locally
3. **Go 1.25+** (for local development without Docker)

## Quick Start

### 1. Start Supabase

```bash
# From project root
cd platform/supabase
supabase start

# Note the credentials displayed (especially JWT secret)
```

### 2. Configure Environment

```bash
# From platform/local directory
cd platform/local

# Copy example environment file
cp .env.example .env

# Edit with your Supabase credentials
nano .env
```

Get your Supabase credentials from the `supabase start` output:
- Copy `JWT secret` → `HOST_JWT_SECRET`
- Copy `anon key` → `SUPABASE_ANON_KEY`
- Update `HOST_JWT_ISSUER` with your project URL

### 3. Start Services

```bash
# Make sure you're in platform/local/
cd platform/local

# Start all services
docker-compose up

# Or in detached mode
docker-compose up -d

# View logs
docker-compose logs -f
```

### 4. Verify Services

**Health Checks:**
```bash
# Tournament Service
curl http://localhost/v1/cups

# Game Management
curl http://localhost/v1/games

# Traefik Dashboard
open http://localhost:8080
```

## Quick Commands

```bash
# All commands assume you're in platform/local/

# Start everything
docker-compose up -d

# Stop everything
docker-compose down

# Rebuild and restart
docker-compose up --build

# View logs
docker-compose logs -f tournament

# Scale up (test load balancing)
docker-compose --profile scale up -d

# Stop scaled services
docker-compose --profile scale down
```

## Testing Features

### 1. Path-Based Routing (like ALB)

```bash
# These all route through Traefik
curl http://localhost/v1/cups
curl http://localhost/v1/tournaments
curl http://localhost/v1/games
curl http://localhost/v1/admin/games
```

### 2. Sticky Sessions for SSE

```bash
# Connect to SSE stream (gets sticky cookie)
curl -N -v http://localhost/v1/stream/cup/123456

# Check for Set-Cookie: winspire_sse=...
# Traefik will route subsequent requests to same instance
```

### 3. Load Balancing (Multiple Instances)

```bash
# Start multiple instances
docker-compose --profile scale up -d

# This starts:
# - tournament (instance 1)
# - tournament-2 (instance 2)

# Test load balancing
for i in {1..10}; do
  curl -s http://localhost/v1/cups | jq
done

# Watch Traefik dashboard to see requests distributed
open http://localhost:8080
```

### 4. JWT Authentication

```bash
# Get JWT from Supabase (using your frontend or Supabase client)

# Test authenticated endpoint
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost/v1/cups
```

### 5. Redis Pub/Sub for SSE

```bash
# Terminal 1: Subscribe to SSE stream
curl -N http://localhost/v1/stream/cup/abc123

# Terminal 2: Publish event via Redis
docker exec -it winspire-redis redis-cli
> PUBLISH sse:events '{"scope_key":"cup:abc123","event_type":"CupUpdated","data":"test","timestamp":1234567890}'

# Terminal 1 should receive the event!
```

### 6. Health Checks

```bash
# Traefik automatically monitors service health

# Stop a service
docker-compose stop tournament

# Check Traefik dashboard - service shows as unhealthy
open http://localhost:8080

# Requests fail gracefully
curl http://localhost/v1/cups
# Returns 503 Service Unavailable

# Restart service
docker-compose start tournament
# Traefik automatically detects it's healthy
```

## Development Workflows

### Workflow A: All Services in Docker

Best for testing the complete stack:

```bash
cd platform/local
docker-compose up --build
```

### Workflow B: Services Running Locally (Hot Reload)

Best for active development:

```bash
# 1. Start infrastructure only
cd platform/local
docker-compose up traefik redis

# 2. Start Supabase
cd ../supabase
supabase start

# 3. Run services locally
cd ../../services/tournament
export $(cat ../../platform/local/.env | xargs)
export SERVICE_PORT=8086
go run cmd/tournament/main.go

# In another terminal:
cd services/game-management
export $(cat ../../platform/local/.env | xargs)
export SERVICE_PORT=8085
go run cmd/game-management/main.go
```

Note: When running locally, services will be accessible directly on their ports (8086, 8085) bypassing Traefik.

### Workflow C: Hybrid (Some Docker, Some Local)

```bash
# Start infrastructure + one service
cd platform/local
docker-compose up traefik redis game-management

# Run tournament service locally for debugging
cd ../../services/tournament
export $(cat ../../platform/local/.env | xargs)
go run cmd/tournament/main.go
```

## Debugging

### View Logs

```bash
cd platform/local

# All services
docker-compose logs -f

# Specific service
docker-compose logs -f tournament

# Traefik access logs
docker-compose logs traefik | grep "request"
```

### Traefik Dashboard

Open http://localhost:8080 and check:
- **HTTP Routes**: See all configured routes
- **Services**: Check health status
- **Middlewares**: View applied middleware

### Redis Monitoring

```bash
# Monitor Redis commands
docker exec -it winspire-redis redis-cli monitor

# Check pub/sub channels
docker exec -it winspire-redis redis-cli
> PUBSUB CHANNELS
> PUBSUB NUMSUB sse:events

# Check cached JWTs
> KEYS jwt:*
> GET jwt:v1:abc123...
```

### Database Debugging

```bash
# Connect to Supabase database
supabase db connect

# Or via psql
psql postgresql://postgres:postgres@localhost:54322/postgres

# Check tables
\dt

# View data
SELECT * FROM cups LIMIT 5;
SELECT * FROM host_subscriptions;
```

## Performance Testing

### Load Testing with k6

```bash
# Install k6
brew install k6  # macOS

# Create test script
cat > loadtest.js <<EOF
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 100 },
    { duration: '1m', target: 100 },
    { duration: '30s', target: 0 },
  ],
};

export default function() {
  let res = http.get('http://localhost/v1/cups');
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
}
EOF

# Run from platform/local/
k6 run loadtest.js
```

### SSE Connection Test

```bash
# Test many concurrent SSE connections
for i in {1..100}; do
  curl -N -s http://localhost/v1/stream/cup/test$i > /dev/null &
done

# Monitor in Traefik dashboard
open http://localhost:8080

# Check Redis Pub/Sub
docker exec -it winspire-redis redis-cli
> PUBSUB NUMSUB sse:events
```

## Cleanup

```bash
cd platform/local

# Stop all services
docker-compose down

# Stop including scaled services
docker-compose --profile scale down

# Stop and remove volumes
docker-compose down -v

# Remove images
docker-compose down --rmi all

# Stop Supabase
cd ../supabase
supabase stop
```

## Troubleshooting

### Service Won't Start

```bash
# Check service logs
cd platform/local
docker-compose logs tournament

# Common issues:
# - Missing .env file → copy from .env.example
# - Supabase not running → supabase start
# - Port already in use (80, 8080, 6379)
```

### Traefik Not Routing

```bash
# Check Traefik dashboard
open http://localhost:8080

# Verify labels on service
docker inspect tournament | grep traefik

# Check Traefik logs
docker-compose logs traefik
```

### Can't Connect to Supabase

```bash
# Check if Supabase is running
supabase status

# Verify port 54322 is accessible
nc -zv localhost 54322

# Check from Docker container
docker exec -it tournament sh
> apk add postgresql-client
> psql postgresql://postgres:postgres@host.docker.internal:54322/postgres -c "SELECT version();"
```

### Redis Connection Failed

```bash
# Check Redis is running
docker-compose ps redis

# Test connection
docker exec -it winspire-redis redis-cli ping

# Check from service
docker exec -it tournament sh
> apk add redis
> redis-cli -h redis ping
```

## Architecture Comparison

### Local (Traefik) vs Production (AWS ALB)

| Feature | Local (Traefik) | Production (ALB) |
|---------|----------------|------------------|
| Path routing | ✅ | ✅ |
| Sticky sessions | ✅ | ✅ |
| Health checks | ✅ | ✅ |
| Load balancing | ✅ | ✅ |
| SSL/TLS | ⚠️ Manual | ✅ Automatic |
| Auto-scaling | ⚠️ Manual | ✅ Automatic |
| CloudWatch | ❌ Disabled | ✅ Enabled |
| Multi-AZ | ❌ N/A | ✅ Yes |

### What Works Locally

✅ **Same as AWS:**
- Path-based routing
- Sticky sessions for SSE
- Health checks
- Load balancing
- Service discovery
- Redis caching
- JWT validation

⚠️ **Different:**
- HTTP instead of HTTPS
- Manual scaling instead of auto-scaling
- Supabase local instead of RDS
- No CloudWatch metrics

## Next Steps

1. ✅ Get services running locally
2. ✅ Test all routes through Traefik
3. ✅ Verify sticky sessions work
4. ✅ Test Redis Pub/Sub for SSE
5. ✅ Load test with multiple instances
6. 🚀 Deploy to AWS with Terraform (see `/infra/`)

---

**Pro Tip:** Use Traefik dashboard (http://localhost:8080) to visually see all routes, services, and their health status. It's like AWS Console for your local environment!


