# Winspire Infrastructure

Terraform modules for deploying Winspire microservices on AWS ECS Fargate with Application Load Balancer.

## Architecture

```
Internet
  ↓
CloudFront (optional)
  ↓
Application Load Balancer (ALB)
  ├─→ /v1/stream/* → competition-host-stream (sticky sessions)
  ├─→ /v1/cups/* → competition-host-stream
  ├─→ /v1/tournaments/* → competition-host-stream
  ├─→ /v1/games/* → game-management
  └─→ /v1/admin/* → game-management (admin only)
  ↓
ECS Fargate Services
  ├─→ competition-host-stream (2-20 instances, auto-scaling)
  ├─→ game-management (2-10 instances, auto-scaling)
  └─→ [future services]
  ↓
Redis ElastiCache (JWT cache + SSE Pub/Sub)
RDS PostgreSQL (via RDS Proxy)
```

## Modules

### ALB (`modules/alb/`)
- Application Load Balancer with path-based routing
- SSL/TLS termination
- Security groups
- Health checks

### ECS Service (`modules/ecs-service/`)
- Reusable ECS service module
- Fargate launch type
- Auto-scaling (CPU, memory, custom metrics)
- CloudWatch logs
- IAM roles
- Sticky sessions support for SSE

### Redis (`modules/redis/`)
- ElastiCache Redis cluster
- Multi-AZ with automatic failover
- Encryption at rest and in transit
- CloudWatch alarms
- Parameter groups

## Quick Start

### Prerequisites

1. **AWS Account** with appropriate permissions
2. **Terraform** >= 1.5
3. **VPC** with public and private subnets
4. **Container images** pushed to ECR

### Deploy Development Environment

```bash
cd infra/environments/dev

# Initialize Terraform
terraform init

# Configure variables
cat > terraform.tfvars <<EOF
aws_region         = "eu-central-1"
vpc_id             = "vpc-xxxxx"
public_subnet_ids  = ["subnet-xxxxx", "subnet-yyyyy"]
private_subnet_ids = ["subnet-aaaaa", "subnet-bbbbb"]
EOF

# Plan
terraform plan

# Apply
terraform apply
```

### Access the Services

```bash
# Get ALB DNS name
terraform output alb_dns_name

# Test health checks
curl http://ALB_DNS_NAME/healthz

# Test specific service
curl http://ALB_DNS_NAME/v1/games
```

## Scaling Configuration

### Auto-Scaling Triggers

**competition-host-stream:**
- CPU > 70% → scale out
- Memory > 80% → scale out
- SSE connections > 500/instance → scale out (custom metric)
- Min: 2 instances
- Max: 20 instances

**game-management:**
- CPU > 70% → scale out
- Memory > 80% → scale out
- Min: 2 instances
- Max: 10 instances

### Manual Scaling

```bash
# Update desired count
terraform apply -var="competition_host_stream_desired_count=5"
```

## Monitoring

### CloudWatch Logs

```bash
# competition-host-stream logs
aws logs tail /ecs/dev/competition-host-stream --follow

# game-management logs
aws logs tail /ecs/dev/game-management --follow
```

### CloudWatch Metrics

Key metrics to monitor:
- `ECSServiceAverageCPUUtilization`
- `ECSServiceAverageMemoryUtilization`
- `TargetResponseTime`
- `HealthyHostCount`
- `UnHealthyHostCount`
- Custom: `SSEActiveConnections`

### CloudWatch Logs Insights

```sql
# Find errors
fields @timestamp, method, path, status, error
| filter status >= 400
| sort @timestamp desc
| limit 100

# Slow requests
fields @timestamp, method, path, latency_ms
| filter latency_ms > 1000
| sort latency_ms desc
| limit 50

# SSE metrics
fields @timestamp, sse.active_connections, sse.messages_published
| stats avg(sse.active_connections) by bin(5m)
```

## Deployment

### CI/CD Pipeline

1. **Build**: Docker images → ECR
2. **Test**: Run integration tests
3. **Deploy**: Update ECS task definition
4. **Verify**: Health checks pass
5. **Rollback**: Automatic on failure

### Blue/Green Deployment

ECS supports blue/green deployments via CodeDeploy:

```bash
# Deploy new version
aws ecs update-service \
  --cluster dev-winspire-cluster \
  --service competition-host-stream \
  --force-new-deployment

# Monitor deployment
aws ecs describe-services \
  --cluster dev-winspire-cluster \
  --services competition-host-stream
```

## Cost Optimization

### Development
- **ECS**: 2 tasks @ 256 CPU / 512 MB = ~$15/month
- **Redis**: cache.t4g.micro = ~$12/month
- **ALB**: ~$20/month
- **RDS**: db.t4g.micro = ~$15/month
- **Total**: ~$62/month

### Production (10,000 players peak)
- **ECS**: 20 tasks @ 512 CPU / 1024 MB = ~$150/month
- **Redis**: cache.r7g.large (2 nodes) = ~$150/month
- **ALB**: ~$50/month
- **RDS**: db.r7g.xlarge + RDS Proxy = ~$300/month
- **Data transfer**: ~$100/month
- **Total**: ~$750/month

## Security

### Network Security
- ALB in public subnets
- ECS tasks in private subnets
- Redis in private subnets
- RDS in private subnets
- Security groups: least privilege

### Secrets Management
- JWT secrets → AWS Secrets Manager
- Database credentials → AWS Secrets Manager
- Environment variables → ECS task definition
- IAM roles: task-specific permissions

### Compliance
- Encryption at rest (EBS, S3, RDS, Redis)
- Encryption in transit (HTTPS, TLS)
- CloudTrail logging
- VPC Flow Logs
- GuardDuty enabled

## Migration to Service Mesh

### Current: ALB + Shared Library
```
Client → ALB → ECS Service → Database
```

### Future: ALB + App Mesh
```
Client → ALB → ECS Service (+ Envoy sidecar) ⟷ ECS Service (+ Envoy sidecar)
                     ↓
              App Mesh Control Plane
```

### Migration Steps

1. **Phase 1**: Deploy current architecture (ALB + httpx)
2. **Phase 2**: Add App Mesh resources (no code changes)
3. **Phase 3**: Enable service-to-service communication via mesh
4. **Phase 4**: Advanced features (circuit breakers, retries, observability)

**Zero code changes required!**

## Troubleshooting

### Service not starting

```bash
# Check task logs
aws logs tail /ecs/dev/service-name --follow

# Check task definition
aws ecs describe-task-definition --task-definition dev-service-name

# Check service events
aws ecs describe-services --cluster dev-winspire-cluster --services service-name
```

### High latency

```bash
# Check CloudWatch metrics
# Check ECS CPU/Memory
# Check Redis performance
# Check RDS performance
# Check ALB target response time
```

### SSE connections failing

```bash
# Verify sticky sessions enabled
# Check Redis Pub/Sub
# Check ECS task count
# Check ALB health checks
```

## Support

For issues or questions:
- Check CloudWatch logs first
- Review ECS service events
- Verify security group rules
- Check IAM permissions

