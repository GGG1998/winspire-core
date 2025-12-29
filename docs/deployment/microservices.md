# Microservices Deployment

## Deployment Flow

### Step 1: Database Migrations (if needed)

If you have schema changes, run migrations **before** deploying code.

**GitHub Actions → `Database Migrations`**

| Input | Options |
|-------|---------|
| Service | `all`, `tournament`, `matchmaking`, `game-management` |
| Action | `apply`, `status`, `down` |
| Environment | `dev`, `staging`, `prod` |
| Dry Run | `true`/`false` (preview SQL without executing) |

### Step 2: Build & Push to ECR

**GitHub Actions → `Build and Push to ECR`**

| Input | Options |
|-------|---------|
| Services | Comma-separated list or `all` |
| Environment | `dev`, `staging`, `prod` |
| Deploy | `true` to auto-deploy to ECS after build |

> **Note:** This action auto-triggers on push to `main` when `services/**` or `libs/go/**` change.

### Step 3: Verify Deployment

Check logs to confirm the service started correctly:

```bash
# Option 1: AWS CLI
aws logs tail /ecs/dev-winspire-tournament --follow

# Option 2: CloudWatch Console
# Navigate to CloudWatch → Log groups → /ecs/<env>-winspire-<service>

# Option 3: Claude Code
/aws-debug logs
```

## Quick Reference

| Scenario | Action |
|----------|--------|
| Code-only change | Step 2 only |
| Schema + code change | Step 1 → Step 2 |
| Rollback migration | Step 1 with `down` action |
