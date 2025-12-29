# GitHub Actions Workflows

## Required Configuration

### Secrets (Settings → Secrets and variables → Actions → Secrets)

| Secret | Description | Where to find |
|--------|-------------|---------------|
| `AWS_ACCESS_KEY_ID` | AWS IAM access key | AWS Console → IAM → Users → Security credentials |
| `AWS_SECRET_ACCESS_KEY` | AWS IAM secret key | AWS Console → IAM → Users → Security credentials |
| `POSTGRES_DSN` | PostgreSQL connection string | Supabase Dashboard → Settings → Database → Connection string → URI |
| `JWT_SECRET` | JWT secret for auth | Supabase Dashboard → Settings → API → JWT Secret |
| `SUPABASE_ANON_KEY` | Supabase anon/public key | Supabase Dashboard → Settings → API → anon/public |
| `SUPABASE_SERVICE_KEY` | Supabase service_role key | Supabase Dashboard → Settings → API → service_role |
| `REDIS_PASSWORD` | Redis password (empty if none) | Your Redis configuration |
| `GAME_API_KEY` | External game API key (optional) | Your game API provider |

### Variables (Settings → Secrets and variables → Actions → Variables)

| Variable | Description | Example |
|----------|-------------|---------|
| `JWT_ISSUER` | JWT issuer URL | `https://your-project.supabase.co/auth/v1` |
| `JWT_AUDIENCE` | JWT audience | `authenticated` |
| `SUPABASE_URL` | Supabase API URL | `https://your-project.supabase.co` |
| `GAME_API_URL` | External game API URL (optional) | `https://api.game.com` |

## Workflows

### `build-and-push.yml`
Builds Docker images and pushes to ECR.
- **Trigger**: Manual (`workflow_dispatch`) or push to `main` (services changed)
- **Options**: Choose services, environment, deploy flag

### `migrate.yml`
Runs database migrations with Atlas.
- **Trigger**: Manual only
- **Options**: Choose service and action (apply/status/down)

### `claude-code-review.yml`
AI code review on pull requests.
- **Trigger**: Pull request opened/synchronized

## How Secrets Flow

```
GitHub Secrets → Terraform variables → ECS Task Definition (env vars)
```

Secrets are passed to Terraform via CLI or tfvars during `terraform apply`:
```bash
terraform apply \
  -var="postgres_dsn=$POSTGRES_DSN" \
  -var="jwt_secret=$JWT_SECRET" \
  ...
```

## First-time Setup

1. Add all secrets to GitHub (Settings → Secrets)
2. Add all variables to GitHub (Settings → Variables)
3. Run Terraform to create infrastructure
4. Run `build-and-push.yml` to build and deploy services
