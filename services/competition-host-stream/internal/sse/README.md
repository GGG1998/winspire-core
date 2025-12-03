# SSE Broker Package

Server-Sent Events (SSE) broker with support for horizontal scaling via Redis Pub/Sub.

## Architecture

### In-Memory Broker (Development)
- **Use Case**: Single instance, development, low traffic
- **Implementation**: `Broker` - uses `r3labs/sse` in-memory
- **Pros**: Simple, no external dependencies
- **Cons**: Cannot scale horizontally

### Redis-Backed Broker (Production)
- **Use Case**: Multiple instances, production, 100-10,000 players
- **Implementation**: `RedisBroker` - uses Redis Pub/Sub + `r3labs/sse`
- **Pros**: Horizontal scaling, HA, handles instance failures
- **Cons**: Requires Redis

## How It Works

### Redis Pub/Sub Pattern

```
Client → ALB (sticky sessions) → Instance A → Local SSE Server
                                  ↓ Publish
                                Redis Pub/Sub (channel: sse:events)
                                  ↓ Subscribe
                      Instance A, B, C → Local SSE Servers
                                  ↓
                             SSE Connections
```

1. **Publishing**: When a service publishes an event, it goes to Redis
2. **Broadcasting**: All instances subscribed to Redis receive the event
3. **Local Delivery**: Each instance pushes to its local SSE connections
4. **Sticky Sessions**: ALB ensures clients always hit the same instance

## Usage

### Development (In-Memory)

```go
broker := ssebroker.NewBroker(100) // pool size
defer broker.Close()

// Publish event
scope := ssebroker.Scope{Type: "cup", ID: cupID}
broker.Publish(ctx, scope, "CupUpdated", jsonData)
```

### Production (Redis-Backed)

```go
// Create Redis client
redisClient := redis.NewClient(&redis.Options{
    Addr: "redis:6379",
})

// Create Redis-backed broker
broker, err := ssebroker.NewRedisBroker(100, redisClient, logger)
if err != nil {
    log.Fatal(err)
}
defer broker.Close()

// Publish event (same interface!)
scope := ssebroker.Scope{Type: "cup", ID: cupID}
broker.Publish(ctx, scope, "CupUpdated", jsonData)
```

### Configuration

Set via environment variable:

```bash
# Development
USE_REDIS_BROKER=false

# Production
USE_REDIS_BROKER=true
REDIS_URL=redis://redis.aws.com:6379
```

## Metrics

The `MetricsCollector` tracks:
- **Active Connections**: Current SSE connections on this instance
- **Total Connections**: Lifetime total connections
- **Messages Sent**: Events pushed to clients
- **Avg Connection Duration**: How long clients stay connected

### CloudWatch Integration

```go
collector := ssebroker.NewMetricsCollector()

// In SSE handler
collector.ConnectionOpened()
defer func() {
    collector.ConnectionClosed(time.Since(startTime))
}()

// Publish metrics every 60 seconds
ticker := time.NewTicker(60 * time.Second)
go func() {
    for range ticker.C {
        collector.PublishMetricsToCloudWatch(ctx, "Winspire", "competition-host-stream")
    }
}()
```

## Scaling Considerations

### Connection Limits

- **Per Instance**: ~500-1000 connections (CPU/memory dependent)
- **Target**: 10,000 players = ~20 instances @ 500 connections each
- **Auto-Scaling**: Based on `ActiveConnections` metric

### ALB Configuration

```hcl
# Sticky sessions required
stickiness {
  type            = "lb_cookie"
  cookie_duration = 3600  # 1 hour
  enabled         = true
}

# Health check
health_check {
  path                = "/healthz"
  interval            = 30
  timeout             = 5
  healthy_threshold   = 2
  unhealthy_threshold = 2
}
```

### Registry (PostgreSQL)

The `Registry` tracks subscriptions in PostgreSQL:
- **Purpose**: Stateless - survive instance restarts
- **Schema**: `host_subscriptions` table
- **TTL**: Leases expire after inactivity
- **Cleanup**: Periodic job removes expired leases

## Testing

### Unit Tests

```bash
go test ./internal/sse -v
```

### Load Testing

```bash
# Start service with Redis
docker-compose up -d redis
USE_REDIS_BROKER=true go run cmd/competition-host-stream/main.go

# Run load test
go test ./internal/benchmarks -bench=BenchmarkSSE -benchtime=10s
```

### Multi-Instance Test

```bash
# Terminal 1
SERVICE_PORT=8086 USE_REDIS_BROKER=true go run cmd/competition-host-stream/main.go

# Terminal 2
SERVICE_PORT=8087 USE_REDIS_BROKER=true go run cmd/competition-host-stream/main.go

# Connect client to instance 1
curl -N http://localhost:8086/v1/stream/cup/123

# Publish event via instance 2
curl -X POST http://localhost:8087/v1/cups/123/update

# Client on instance 1 should receive the event!
```

## Migration Path

1. **Phase 1**: Deploy with `USE_REDIS_BROKER=false` (no changes)
2. **Phase 2**: Add Redis to infrastructure
3. **Phase 3**: Deploy with `USE_REDIS_BROKER=true`
4. **Phase 4**: Scale to multiple instances
5. **Phase 5**: Enable auto-scaling based on metrics

Zero code changes required in handlers or projections!


