# Development Environment Configuration
# Serverless setup with App Runner (scale-to-zero) + Upstash Redis

terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "Winspire"
      Environment = "dev"
      ManagedBy   = "Terraform"
    }
  }
}

# =============================================================================
# Variables
# =============================================================================

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "eu-central-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

# Database & Auth (Supabase)
variable "postgres_dsn" {
  description = "PostgreSQL connection string (Supabase)"
  type        = string
  sensitive   = true
}

variable "jwt_secret" {
  description = "JWT secret (Supabase)"
  type        = string
  sensitive   = true
}

variable "jwt_issuer" {
  description = "JWT issuer URL (Supabase)"
  type        = string
  default     = "https://your-project.supabase.co/auth/v1"
}

variable "jwt_audience" {
  description = "JWT audience"
  type        = string
  default     = "authenticated"
}

variable "supabase_url" {
  description = "Supabase API URL"
  type        = string
}

variable "supabase_anon_key" {
  description = "Supabase anon key"
  type        = string
  sensitive   = true
}

variable "supabase_service_key" {
  description = "Supabase service key"
  type        = string
  sensitive   = true
  default     = ""
}

# Redis (Upstash - serverless)
variable "upstash_redis_url" {
  description = "Upstash Redis URL (redis://...)"
  type        = string
  sensitive   = true
}

variable "upstash_redis_password" {
  description = "Upstash Redis password"
  type        = string
  sensitive   = true
  default     = ""
}

# Optional
variable "game_api_url" {
  description = "External game API URL"
  type        = string
  default     = ""
}

variable "game_api_key" {
  description = "External game API key"
  type        = string
  sensitive   = true
  default     = ""
}

# =============================================================================
# Data Sources
# =============================================================================

data "aws_caller_identity" "current" {}

# =============================================================================
# App Runner - Matchmaking Service (scale-to-zero)
# =============================================================================

module "matchmaking" {
  source = "../../modules/app-runner"

  environment   = var.environment
  service_name  = "matchmaking"

  container_image = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/dev-winspire-matchmaking:latest"
  container_port  = 8081

  # Resources (smallest for dev)
  cpu    = "256"   # 0.25 vCPU
  memory = "512"   # 512 MB

  # Auto-scaling configuration (App Runner pauses instances when idle)
  min_instances   = 1   # AWS minimum, but paused when no traffic
  max_instances   = 5
  max_concurrency = 100 # Requests per instance before scaling

  health_check_path = "/health"
  auto_deploy       = true

  # Custom domain
  custom_domain = "dev-api.gowinspire.com"

  environment_variables = {
    APP_ENV                           = "production"
    SERVICE_PORT                      = "8081"
    PORT                              = "8081"
    GIN_MODE                          = "release"

    # Upstash Redis (serverless)
    REDIS_URL                         = var.upstash_redis_url
    REDIS_ADDR                        = replace(replace(var.upstash_redis_url, "redis://", ""), "rediss://", "")
    REDIS_PASSWORD                    = var.upstash_redis_password
    REDIS_DB                          = "0"
    REDIS_TLS_ENABLED                 = "true"

    # Supabase
    POSTGRES_DSN                      = var.postgres_dsn
    DATABASE_URL                      = var.postgres_dsn
    HOST_JWT_SECRET                   = var.jwt_secret
    HOST_JWT_ISSUER                   = var.jwt_issuer
    HOST_JWT_AUDIENCE                 = var.jwt_audience
    SUPABASE_URL                      = var.supabase_url
    SUPABASE_ANON_KEY                 = var.supabase_anon_key
    SUPABASE_SERVICE_KEY              = var.supabase_service_key

    # Service URLs (self-referencing for monolith)
    GAME_MANAGEMENT_URL               = "https://placeholder.awsapprunner.com"
    GAME_MANAGEMENT_INTERNAL_API_KEY  = "dev-internal-api-key-12345"

    # Game API
    GAME_API_URL                      = var.game_api_url
    GAME_API_KEY                      = var.game_api_key
    GAME_API_POLL_INTERVAL            = "5s"
    GAME_API_POLL_TIMEOUT             = "60s"

    # Feature flags
    ENABLE_WEBSOCKET_LOBBIES          = "true"
    ENABLE_AUTO_SCORE_RETRIEVAL       = "true"
    ENABLE_DISCONNECT_HANDLING        = "true"
    SCHEDULER_ENABLED                 = "true"
    SCHEDULER_INTERVAL                = "*/2 * * * *"

    # Timeouts
    LOBBY_JOIN_TIMEOUT                = "2m"
    DISCONNECT_RECONNECT_WINDOW       = "30s"
    READY_CHECK_TIMEOUT               = "5m"

    # Limits
    MAX_CONCURRENT_TOURNAMENTS        = "50"
    MAX_PARTICIPANTS_PER_TOURNAMENT   = "256"
    DATABASE_MAX_CONNS                = "10"
    DATABASE_MAX_IDLE_CONNS           = "2"

    # CORS
    CORS_ALLOWED_ORIGINS              = "https://winspire-dev-s63lr.ondigitalocean.app,https://d2mjyl2e92t3yg.cloudfront.net,http://localhost:5173"

    # S3 for game bundles
    AWS_REGION                        = var.aws_region
    AWS_S3_BUCKET                     = "gowinspire-game"

    # Logging
    LOG_LEVEL                         = "info"
    LOG_FORMAT                        = "json"
  }

  # S3 permissions for game assets
  instance_policy_json = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          "arn:aws:s3:::gowinspire-game",
          "arn:aws:s3:::gowinspire-game/*"
        ]
      }
    ]
  })

  tags = { Team = "Platform" }
}

# =============================================================================
# Mini-Admin Frontend (S3 + CloudFront) - unchanged
# =============================================================================

module "mini_admin_bucket" {
  source = "../../modules/s3-frontend-hosting"

  bucket_name          = "winspire-mini-admin-${var.environment}"
  environment          = var.environment
  spa_mode             = true
  enable_public_access = false

  tags = {
    Service = "mini-admin"
    Team    = "Platform"
  }
}

module "mini_admin_cdn" {
  source = "../../modules/cloudfront"

  bucket_id                   = module.mini_admin_bucket.bucket_id
  bucket_arn                  = module.mini_admin_bucket.bucket_arn
  bucket_regional_domain_name = module.mini_admin_bucket.bucket_regional_domain_name
  environment                 = var.environment
  comment                     = "Mini-Admin Dashboard CDN"

  tags = {
    Service = "mini-admin"
    Team    = "Platform"
  }
}

# =============================================================================
# Outputs
# =============================================================================

output "matchmaking_url" {
  description = "Matchmaking API URL (App Runner)"
  value       = module.matchmaking.service_url
}

output "matchmaking_status" {
  description = "Matchmaking service status"
  value       = module.matchmaking.service_status
}

output "mini_admin_url" {
  description = "Mini-Admin Dashboard URL (CloudFront)"
  value       = module.mini_admin_cdn.distribution_url
}

output "mini_admin_bucket" {
  description = "Mini-Admin S3 bucket name"
  value       = module.mini_admin_bucket.bucket_id
}

output "mini_admin_distribution_id" {
  description = "Mini-Admin CloudFront distribution ID"
  value       = module.mini_admin_cdn.distribution_id
}

output "custom_domain_dns_records" {
  description = "DNS records to add at your registrar for dev-api.gowinspire.com"
  value       = module.matchmaking.custom_domain_dns_records
}
