# S3 Static Website Hosting Module

This Terraform module creates an S3 bucket configured for static website hosting with optional versioning and public read access.

## Features

- S3 bucket with public read access
- Optional versioning support
- Static website hosting configuration
- CORS configuration for cross-origin access
- Server-side encryption (AES256)
- Lifecycle rules for cleanup
- CloudWatch metrics (optional)

## Usage

```hcl
module "games_bucket" {
  source = "../../modules/s3-static"

  bucket_name           = "mini-admin-games-${var.environment}"
  environment           = var.environment
  enable_versioning     = true
  enable_static_website = true
  cors_allowed_origins  = ["*"]
  enable_uploads        = true

  tags = {
    Project = "mini-admin"
    Service = "game-hosting"
  }
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| bucket_name | Name of the S3 bucket | `string` | n/a | yes |
| environment | Environment name | `string` | n/a | yes |
| enable_versioning | Enable versioning | `bool` | `false` | no |
| enable_static_website | Enable static website hosting | `bool` | `true` | no |
| index_document | Index document | `string` | `"index.html"` | no |
| error_document | Error document | `string` | `"error.html"` | no |
| cors_allowed_origins | CORS allowed origins | `list(string)` | `["*"]` | no |
| enable_uploads | Enable upload CORS rules | `bool` | `true` | no |
| enable_lifecycle_rules | Enable lifecycle rules | `bool` | `true` | no |
| noncurrent_version_expiration_days | Days before version expiration | `number` | `90` | no |
| enable_metrics | Enable CloudWatch metrics | `bool` | `false` | no |
| tags | Additional tags | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| bucket_id | The ID of the S3 bucket |
| bucket_arn | The ARN of the S3 bucket |
| bucket_domain_name | The bucket domain name |
| bucket_regional_domain_name | The bucket regional domain name |
| website_endpoint | The website endpoint |
| website_domain | The domain of the website endpoint |

## Notes

- **Versioning**: Enabling versioning will incur additional S3 storage costs
- **Public Access**: This module configures the bucket for public read access - suitable for static game files
- **CORS**: By default allows all origins; restrict in production environments
- **Lifecycle**: Old versions expire after 90 days to manage storage costs












