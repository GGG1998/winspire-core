# Service Mesh Migration Guide

This document outlines the migration path from the current ALB-based architecture to AWS App Mesh for service-to-service communication.

## Current Architecture (Phase 1)

```
┌─────────┐
│ Client  │
└────┬────┘
     │
┌────▼────────────────┐
│ Application Load    │
│ Balancer (ALB)      │
└──┬──────────────┬───┘
   │              │
┌──▼──────────┐  ┌▼──────────────┐
│ competition- │  │ game-         │
│ host-stream  │  │ management    │
└──────┬───────┘  └───────┬───────┘
       │                  │
       └────────┬─────────┘
                │
        ┌───────▼────────┐
        │ PostgreSQL     │
        │ Redis          │
        └────────────────┘
```

**Characteristics:**
- ✅ Simple architecture
- ✅ Shared middleware library (`libs/go/httpx`)
- ✅ JWT validation with Redis caching
- ✅ Horizontal scaling with Redis Pub/Sub for SSE
- ❌ No service-to-service communication yet
- ❌ No advanced traffic management

## Target Architecture (Phase 4)

```
┌─────────┐
│ Client  │
└────┬────┘
     │
┌────▼────────────────┐
│ Application Load    │
│ Balancer (ALB)      │
└──┬──────────────┬───┘
   │              │
┌──▼──────────┐  ┌▼──────────────┐
│ Envoy Proxy │  │ Envoy Proxy   │
│     ↓       │  │      ↓        │
│ competition-│  │ game-         │
│ host-stream │◄─┤ management    │
└──────┬───────┘  └───────┬───────┘
       │                  │
       └────────┬─────────┘
                │
        ┌───────▼────────┐
        │ PostgreSQL     │
        │ Redis          │
        │                │
        │ ┌────────────┐ │
        │ │ App Mesh   │ │
        │ │ Control    │ │
        │ │ Plane      │ │
        │ └────────────┘ │
        └────────────────┘
```

**Added Capabilities:**
- ✅ Service-to-service communication via Envoy
- ✅ Circuit breakers & retries
- ✅ Distributed tracing (X-Ray)
- ✅ mTLS between services
- ✅ Traffic splitting (canary deployments)
- ✅ **Zero code changes required!**

## Migration Phases

### Phase 1: Current State (✅ Implemented)

**Infrastructure:**
- ALB with path-based routing
- ECS Fargate services
- Redis for JWT caching and SSE broadcasting
- Shared `httpx` middleware library

**Benefits:**
- Simple to understand and debug
- Low operational overhead
- Fast to deploy

**When to stay here:**
- Single-service or limited service-to-service communication
- Team < 10 developers
- Traffic < 10,000 concurrent users

### Phase 2: App Mesh Introduction

**What changes:**
- Add App Mesh virtual nodes and services
- Deploy Envoy sidecar proxies with ECS tasks
- Configure service discovery

**What stays the same:**
- Application code (no changes!)
- ALB for external traffic
- Redis and database setup
- Deployment process

**Terraform changes:**

```hcl
# Add App Mesh resources
resource "aws_appmesh_mesh" "winspire" {
  name = "winspire-mesh"
}

resource "aws_appmesh_virtual_node" "competition_host_stream" {
  name      = "competition-host-stream"
  mesh_name = aws_appmesh_mesh.winspire.name

  spec {
    listener {
      port_mapping {
        port     = 8086
        protocol = "http"
      }
    }

    service_discovery {
      aws_cloud_map {
        service_name   = aws_service_discovery_service.competition_host_stream.name
        namespace_name = aws_service_discovery_private_dns_namespace.winspire.name
      }
    }
  }
}

# Update ECS task definition to include Envoy sidecar
resource "aws_ecs_task_definition" "competition_host_stream" {
  # ... existing configuration ...

  proxy_configuration {
    type           = "APPMESH"
    container_name = "envoy"
    properties = {
      AppPorts         = "8086"
      EgressIgnoredIPs = "169.254.170.2,169.254.169.254"
      IgnoredUID       = "1337"
      ProxyEgressPort  = 15001
      ProxyIngressPort = 15000
    }
  }
}
```

**Duration:** 1-2 weeks (including testing)

**Rollback:** Remove App Mesh resources, services continue working

### Phase 3: Service-to-Service Communication

**What changes:**
- Enable internal service calls via App Mesh
- Example: `game-management` calls `competition-host-stream` for tournament data

**Application code changes:**

```go
// Before (no service-to-service communication)
// All data fetched directly from database

// After (with App Mesh)
import "net/http"

func (s *Service) GetTournamentInfo(ctx context.Context, tournamentID string) (*Tournament, error) {
    // Call competition-host-stream via service mesh
    // App Mesh handles service discovery, load balancing, retries
    url := "http://competition-host-stream.winspire.local:8086/v1/tournaments/" + tournamentID
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    // Envoy sidecar intercepts this call
    // - Adds tracing headers
    // - Applies retry policy
    // - Enforces circuit breaker
    // - Routes to healthy instance
    resp, err := http.DefaultClient.Do(req)
    // ... handle response ...
}
```

**App Mesh configuration:**

```hcl
resource "aws_appmesh_virtual_router" "competition_host_stream" {
  name      = "competition-host-stream-router"
  mesh_name = aws_appmesh_mesh.winspire.name

  spec {
    listener {
      port_mapping {
        port     = 8086
        protocol = "http"
      }
    }
  }
}

resource "aws_appmesh_route" "competition_host_stream_default" {
  name                = "default"
  mesh_name           = aws_appmesh_mesh.winspire.name
  virtual_router_name = aws_appmesh_virtual_router.competition_host_stream.name

  spec {
    http_route {
      match {
        prefix = "/"
      }

      retry_policy {
        http_retry_events = [
          "server-error",
          "gateway-error",
        ]
        max_retries = 3
        per_retry_timeout {
          unit  = "s"
          value = 5
        }
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.competition_host_stream.name
          weight       = 100
        }
      }
    }
  }
}
```

**Duration:** 2-4 weeks (design, implement, test)

**Rollback:** Revert to database queries

### Phase 4: Advanced Features

**Traffic Splitting (Canary Deployments):**

```hcl
resource "aws_appmesh_route" "canary_deployment" {
  # ... configuration ...

  spec {
    http_route {
      match {
        prefix = "/"
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.competition_host_stream_v1.name
          weight       = 90  # 90% to old version
        }
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.competition_host_stream_v2.name
          weight       = 10  # 10% to new version
        }
      }
    }
  }
}
```

**Circuit Breaker:**

```hcl
resource "aws_appmesh_virtual_node" "game_management" {
  # ... configuration ...

  spec {
    listener {
      outlier_detection {
        max_ejection_percent = 50
        interval {
          unit  = "s"
          value = 10
        }
        base_ejection_duration {
          unit  = "s"
          value = 30
        }
      }
    }
  }
}
```

**mTLS Encryption:**

```hcl
resource "aws_appmesh_virtual_node" "competition_host_stream" {
  # ... configuration ...

  spec {
    listener {
      tls {
        mode = "STRICT"
        certificate {
          acm {
            certificate_arn = aws_acm_certificate.service.arn
          }
        }
      }
    }

    backend_defaults {
      client_policy {
        tls {
          enforce = true
          validation {
            trust {
              acm {
                certificate_authority_arns = [
                  aws_acm_pca_certificate_authority.mesh.arn
                ]
              }
            }
          }
        }
      }
    }
  }
}
```

**Duration:** Ongoing (enable features as needed)

## Decision Matrix: Stay on Phase 1 vs Move to App Mesh

### Stay on Phase 1 (ALB + Shared Library) if:
- ✅ 1-3 microservices
- ✅ Limited service-to-service communication
- ✅ Team comfortable with current architecture
- ✅ <10,000 concurrent users
- ✅ Monolithic-style deployments are acceptable

### Move to App Mesh if:
- ✅ 4+ microservices
- ✅ Frequent service-to-service calls
- ✅ Need for advanced traffic management (canary, A/B)
- ✅ mTLS requirement
- ✅ Distributed tracing needed
- ✅ >10,000 concurrent users
- ✅ Complex deployment strategies

## Cost Considerations

**Phase 1 (Current):**
- ALB: ~$20/month
- ECS: ~$150/month (20 tasks @ 512 CPU / 1024 MB)
- Redis: ~$150/month
- **Total: ~$320/month**

**Phase 4 (With App Mesh):**
- ALB: ~$20/month
- ECS: ~$180/month (20 tasks + Envoy sidecars)
- Redis: ~$150/month
- App Mesh: ~$0 (no additional cost, charges only for Envoy resources)
- X-Ray: ~$5/month (tracing)
- **Total: ~$355/month (+11%)**

The incremental cost is minimal (~$35/month for advanced capabilities).

## Testing App Mesh Locally

Use AWS Copilot or Docker Compose with Envoy:

```yaml
# docker-compose.app-mesh.yml
version: '3.8'

services:
  competition-host-stream:
    build: ./services/competition-host-stream
    environment:
      - SERVICE_PORT=8086
    depends_on:
      - envoy-competition-host-stream

  envoy-competition-host-stream:
    image: public.ecr.aws/appmesh/aws-appmesh-envoy:v1.29.7.0-prod
    environment:
      - APPMESH_RESOURCE_ARN=mesh/winspire/virtualNode/competition-host-stream
    volumes:
      - ./envoy-config.yaml:/etc/envoy/envoy.yaml
```

## Summary

The migration to App Mesh is **incremental** and **non-disruptive**:
- ✅ No code changes to existing services
- ✅ Gradual feature adoption
- ✅ Easy rollback at any phase
- ✅ Minimal cost increase
- ✅ Maintains current architecture benefits

**Recommendation:** Stay on Phase 1 until you have 3+ services with significant inter-service communication.


