# Digitalocean

**Admin**:
- gabriel.domanowski@gowinspire.com

**Envirnonemnt**:
- dev

## Dev

**Url**: https://winspire-dev-s63lr.ondigitalocean.app/auth/user/login

### Deployment Proces

1. Create PR

Create a Pull Request targeting the winspire-app/dev branch:
https://github.com/Winspire-Lab/winspire-core/compare/winspire-app/dev...<your-branch>

2. Code Review

- PR must be reviewed and approved by a team member
- All CI checks must pass

3. Merge & Auto-Deploy

- Once approved, merge the PR
- Deployment to dev environment triggers automatically after merge

That's it - the pipeline handles the rest.

---

# Mini-Admin (AWS S3 + CloudFront)

**Infrastructure**: AWS S3 + CloudFront
**Environment**: dev
**Source**: `frontends/mini-admin/`

## Infrastructure

Mini-admin is hosted on AWS using:
- **S3 Bucket**: `winspire-mini-admin-dev` - stores static files
- **CloudFront**: CDN distribution with HTTPS

Infrastructure is managed via Terraform in `platform/terraform/environments/dev/main.tf`.

## Get Deployment Info

```bash
cd platform/terraform/environments/dev
terraform output mini_admin_url              # CloudFront URL
terraform output mini_admin_bucket           # S3 bucket name
terraform output mini_admin_distribution_id  # For cache invalidation
```

## Manual Deployment

### 1. Build

```bash
cd frontends/mini-admin
VITE_API_URL=https://dev-api.gowinspire.com/v1 yarn build
```

### 2. Deploy to S3

```bash
aws s3 sync dist/ s3://winspire-mini-admin-dev --delete
```

### 3. Invalidate CloudFront Cache (after updates)

```bash
aws cloudfront create-invalidation \
  --distribution-id <DISTRIBUTION_ID> \
  --paths "/*"
```

Or using Terraform output:
```bash
aws cloudfront create-invalidation \
  --distribution-id $(terraform output -raw mini_admin_distribution_id) \
  --paths "/*"
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `VITE_API_URL` | Backend API URL | `http://localhost:8088/v1` |
| `VITE_INTERNAL_SERVICE_KEY` | Internal auth key | `IlikeCookies` |

## Quick Deploy Script

```bash
#!/bin/bash
# deploy-mini-admin.sh

set -e

cd frontends/mini-admin
VITE_API_URL=https://dev-api.gowinspire.com/v1 yarn build

aws s3 sync dist/ s3://winspire-mini-admin-dev --delete

DIST_ID=$(cd ../../platform/terraform/environments/dev && terraform output -raw mini_admin_distribution_id)
aws cloudfront create-invalidation --distribution-id $DIST_ID --paths "/*"

echo "Deployed! URL: $(cd ../../platform/terraform/environments/dev && terraform output -raw mini_admin_url)"
```