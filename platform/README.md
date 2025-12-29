# Platform

Infrastructure and platform services for Winspire.

## Structure

```
platform/
├── local/              # Local development environment with Traefik
│   ├── docker-compose.yaml
│   ├── README.md       # Full documentation
│   ├── QUICK_START.md  # Get started in 3 steps
│   ├── start.sh        # Quick start script
│   ├── stop.sh         # Quick stop script
│   ├── Makefile        # Common commands
│   └── env.example     # Environment variables template
│
└── supabase/           # Supabase configuration (Auth, Database, Storage)
    ├── config.toml
    ├── migrations/
    └── Makefile
```

## Quick Links

### Local Development
- **Quick Start**: [local/QUICK_START.md](local/QUICK_START.md) - Get running in 3 steps
- **Full Guide**: [local/README.md](local/README.md) - Complete documentation
- **Commands**: `cd local && make help`

### Supabase
- **Setup**: `cd supabase && supabase start`
- **Migrations**: `cd supabase && make migrate`
- **Stop**: `cd supabase && supabase stop`

## Features

### Local Development Environment

Simulates AWS ALB + ECS locally using Docker and Traefik:

✅ **Path-based routing** - Like AWS ALB  
✅ **Sticky sessions** - For SSE connections  
✅ **Load balancing** - Test with multiple instances  
✅ **Health checks** - Automatic service monitoring  
✅ **Service discovery** - Via Docker labels  
✅ **Redis caching** - JWT validation + SSE Pub/Sub  

### Traefik Dashboard

Monitor all services, routes, and health checks:
- **URL**: http://localhost:8080
- **Features**: Real-time metrics, route visualization, service status

## Usage

### Start Everything

```bash
# Option 1: Quick start script
cd local && ./start.sh

# Option 2: Makefile
cd local && make start

# Option 3: Docker Compose directly
cd local && docker-compose up
```

### Common Tasks

```bash
cd local

# View logs
make logs

# Restart services
make restart

# Test load balancing (2 instances)
make scale

# Open Traefik dashboard
make dashboard

# Connect to Redis
make redis-cli

# Run health checks
make test

# Cleanup
make clean
```

## Architecture

### Local (Development)

```
Browser → Traefik (localhost:80)
  ├─ /v1/stream/* → tournament (sticky sessions)
  ├─ /v1/cups/* → tournament
  ├─ /v1/tournaments/* → tournament
  ├─ /v1/games/* → game-management
  └─ /v1/admin/* → game-management
  ↓
Services
  ├─ Redis (cache + Pub/Sub)
  └─ Supabase (Auth + Database + Storage)
```

### Production (AWS)

```
Internet → CloudFront → ALB
  ├─ /v1/stream/* → ECS Service (tournament)
  ├─ /v1/cups/* → ECS Service (tournament)
  ├─ /v1/games/* → ECS Service (game-management)
  └─ /v1/admin/* → ECS Service (game-management)
  ↓
ECS Fargate (Auto-scaling 2-20 instances)
  ├─ Redis ElastiCache
  ├─ RDS PostgreSQL
  └─ CloudWatch
```

See [../infra/README.md](../infra/README.md) for production deployment.

## Differences: Local vs Production

| Feature | Local | Production |
|---------|-------|------------|
| Gateway | Traefik | AWS ALB |
| Compute | Docker | ECS Fargate |
| Database | Supabase local | RDS PostgreSQL |
| Cache | Docker Redis | ElastiCache |
| Scaling | Manual | Auto-scaling |
| SSL | HTTP | HTTPS (ACM) |
| Monitoring | Traefik Dashboard | CloudWatch |
| Cost | $0 | ~$750/month |

## Troubleshooting

### Supabase not running
```bash
cd supabase
supabase start
supabase status
```

### Ports already in use
```bash
# Check what's using the ports
lsof -i :80 -i :8080 -i :6379

# Stop conflicting services
# ... then restart
```

### Services won't connect to Supabase
```bash
# Test connection from host
nc -zv localhost 54322

# Check from Docker
docker exec -it tournament sh
> apk add postgresql-client
> psql postgresql://postgres:postgres@host.docker.internal:54322/postgres
```

## Next Steps

1. ✅ Set up local environment → [local/QUICK_START.md](local/QUICK_START.md)
2. 🧪 Test features → [local/README.md](local/README.md)
3. 📊 Monitor with Traefik → http://localhost:8080
4. 🚀 Deploy to AWS → [../infra/README.md](../infra/README.md)

## Support

- **Local issues**: See [local/README.md](local/README.md#troubleshooting)
- **Supabase issues**: `supabase --help` or https://supabase.com/docs
- **Infrastructure**: See [../infra/README.md](../infra/README.md)


