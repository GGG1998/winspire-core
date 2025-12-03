# Quick Start Guide

Get Winspire running locally in 3 steps.

## Prerequisites

- Docker & Docker Compose
- Supabase CLI (`brew install supabase/tap/supabase`)

## Steps

### 1. Start Supabase

```bash
cd ../supabase
supabase start
```

**Copy these values** from the output:
- `JWT secret`
- `anon key`

### 2. Configure Environment

```bash
cd ../local

# Create .env file
cp env.example .env

# Edit .env and paste your values
nano .env
```

### 3. Start Everything

```bash
# Option A: Using script
./start.sh

# Option B: Using make
make start

# Option C: Using docker-compose directly
docker-compose up
```

## Verify It Works

```bash
# Test API Gateway
curl http://localhost/v1/cups
curl http://localhost/v1/games

# Open Traefik dashboard
open http://localhost:8080
```

## Common Commands

```bash
# View logs
docker-compose logs -f

# Stop services
./stop.sh
# or
make stop

# Restart with rebuild
make rebuild

# Test load balancing (2 instances)
make scale
```

## Next Steps

- See [README.md](README.md) for detailed documentation
- Test SSE connections: `curl -N http://localhost/v1/stream/cup/123`
- Load test: `k6 run loadtest.js`

## Troubleshooting

**Services won't start?**
```bash
# Check Supabase is running
supabase status

# Check ports aren't in use
lsof -i :80 -i :8080 -i :6379
```

**Can't connect to database?**
```bash
# Test connection
nc -zv localhost 54322

# Or connect directly
psql postgresql://postgres:postgres@localhost:54322/postgres
```


