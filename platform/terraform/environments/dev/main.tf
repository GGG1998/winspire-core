# Development Environment Configuration
# This is an example showing how to use the Winspire infrastructure modules

terraform {
  required_version = ">= 1.5"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }

  # Backend configuration (uncomment and configure for your setup)
  # backend "s3" {
  #   bucket         = "winspire-terraform-state"
  #   key            = "dev/terraform.tfstate"
  #   region         = "us-east-1"
  #   encrypt        = true
  #   dynamodb_table = "winspire-terraform-locks"
  # }
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

# Variables
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "eu-central-1"
}

variable "postgres_dsn" {
  description = "PostgreSQL connection string (Supabase) - set via GitHub Secrets"
  type        = string
  sensitive   = true
  default     = "placeholder-will-be-set-by-github-ci"
}

variable "jwt_secret" {
  description = "JWT secret for authentication (Supabase) - set via GitHub Secrets"
  type        = string
  sensitive   = true
  default     = "placeholder-will-be-set-by-github-ci"
}

variable "jwt_issuer" {
  description = "JWT issuer URL (Supabase)"
  type        = string
  default     = "https://your-project.supabase.co/auth/v1"
}

variable "jwt_audience" {
  description = "JWT audience (Supabase)"
  type        = string
  default     = "authenticated"
}

variable "supabase_url" {
  description = "Supabase API URL"
  type        = string
  default     = "http://localhost:54321"
}

variable "supabase_anon_key" {
  description = "Supabase anon/public key"
  type        = string
  sensitive   = true
  default     = ""
}

variable "supabase_service_key" {
  description = "Supabase service_role/secret key (bypasses RLS)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "redis_password" {
  description = "Redis password (empty for no auth)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "game_api_url" {
  description = "External game API URL (optional)"
  type        = string
  default     = ""
}

variable "game_api_key" {
  description = "External game API key (optional)"
  type        = string
  sensitive   = true
  default     = ""
}

# VPC - creates all networking
module "vpc" {
  source = "../../modules/vpc"

  environment          = "dev"
  vpc_cidr             = "10.0.0.0/16"
  az_count             = 2
  enable_nat_gateway   = false  # Not needed - ECS in public subnets
  enable_vpc_endpoints = false  # Not needed without NAT

  tags = { Team = "Platform" }
}

# Self-signed TLS certificate for HTTPS (development only)
resource "tls_private_key" "alb" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "alb" {
  private_key_pem = tls_private_key.alb.private_key_pem

  subject {
    common_name  = "dev-winspire-alb.local"
    organization = "Winspire Dev"
  }

  dns_names = [
    "dev-winspire-alb.local",
    "*.eu-central-1.elb.amazonaws.com",
    "localhost"
  ]

  validity_period_hours = 8760 # 1 year

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "alb_self_signed" {
  private_key      = tls_private_key.alb.private_key_pem
  certificate_body = tls_self_signed_cert.alb.cert_pem

  tags = {
    Name        = "dev-winspire-alb-self-signed"
    Environment = "dev"
  }

  lifecycle {
    create_before_destroy = true
  }
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "dev-winspire-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = {
    Name = "dev-winspire-cluster"
  }
}

# Application Load Balancer
module "alb" {
  source = "../../modules/alb"

  environment                = "dev"
  vpc_id                     = module.vpc.vpc_id
  public_subnet_ids          = module.vpc.public_subnet_ids
  enable_deletion_protection = false
  certificate_arn            = aws_acm_certificate.alb_self_signed.arn

  tags = { Team = "Platform" }
}

# Redis for JWT caching and SSE broadcasting
module "redis" {
  source = "../../modules/redis"

  environment        = "dev"
  vpc_id             = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids

  # Allow ECS services to access Redis
  allowed_security_groups = [
    module.tournament.security_group_id,
    module.matchmaking.security_group_id,
    module.game_management.security_group_id,
  ]

  # Dev settings - smallest instance
  node_type            = "cache.t4g.micro"
  num_cache_nodes      = 1
  enable_cloudwatch_alarms = false

  # Disable TLS for dev - Go services don't have TLS configured
  transit_encryption_enabled = false
  at_rest_encryption_enabled = false

  tags = { Team = "Platform" }
}

# Tournament Service
module "tournament" {
  source = "../../modules/ecs-service"

  environment            = "dev"
  service_name           = "tournament"
  ecs_cluster_id         = aws_ecs_cluster.main.id
  ecs_cluster_name       = aws_ecs_cluster.main.name
  vpc_id                 = module.vpc.vpc_id
  private_subnet_ids     = module.vpc.public_subnet_ids  # Public for internet access
  assign_public_ip       = true
  alb_security_group_id  = module.alb.alb_security_group_id

  container_image = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/dev-winspire-tournament:latest"
  container_port  = 8089

  # Minimal Fargate size (0.25 vCPU, 512MB) - ~$9/month per task
  task_cpu    = 256
  task_memory = 512

  # Scale from 1 to 10 (min 1 required for ECS API compatibility)
  desired_count = 1
  min_capacity  = 1
  max_capacity  = 10

  enable_sticky_sessions = true
  sticky_session_duration = 3600

  environment_variables = {
    APP_ENV               = "production"
    SERVICE_PORT          = "8089"
    REDIS_ADDR            = "${module.redis.redis_endpoint}:6379"
    MATCHMAKING_BASE_URL  = "http://matchmaking.internal:8081"
    GAME_MANAGEMENT_URL   = "http://game-management.internal:8087"
    HOST_JWT_ISSUER       = var.jwt_issuer
    HOST_JWT_AUDIENCE     = var.jwt_audience
    SCHEDULER_ENABLED     = "true"
    SCHEDULER_INTERVAL    = "*/2 * * * *"
    LOG_LEVEL             = "info"
    HTTP_READ_TIMEOUT     = "15s"
    HTTP_WRITE_TIMEOUT    = "15s"
    SHUTDOWN_GRACE        = "10s"
    POSTGRES_DSN          = var.postgres_dsn
    HOST_JWT_SECRET       = var.jwt_secret
  }

  enable_execute_command = true
  tags = { Team = "Platform" }
}

# Matchmaking Service
module "matchmaking" {
  source = "../../modules/ecs-service"

  environment            = "dev"
  service_name           = "matchmaking"
  ecs_cluster_id         = aws_ecs_cluster.main.id
  ecs_cluster_name       = aws_ecs_cluster.main.name
  vpc_id                 = module.vpc.vpc_id
  private_subnet_ids     = module.vpc.public_subnet_ids
  assign_public_ip       = true
  alb_security_group_id  = module.alb.alb_security_group_id

  container_image = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/dev-winspire-matchmaking:latest"
  container_port  = 8081

  task_cpu    = 256
  task_memory = 512

  desired_count = 1
  min_capacity  = 1
  max_capacity  = 10

  environment_variables = {
    APP_ENV                           = "production"
    SERVICE_PORT                      = "8081"
    PORT                              = "8081"
    GIN_MODE                          = "release"
    REDIS_ADDR                        = "${module.redis.redis_endpoint}:6379"
    REDIS_URL                         = "redis://${module.redis.redis_endpoint}:6379/0"
    REDIS_DB                          = "0"
    TOURNAMENT_SERVICE_URL            = "http://tournament.internal:8089"
    GAME_MANAGEMENT_URL               = "http://game-management.internal:8087"
    HOST_JWT_ISSUER                   = var.jwt_issuer
    HOST_JWT_AUDIENCE                 = var.jwt_audience
    SUPABASE_URL                      = var.supabase_url
    GAME_API_URL                      = var.game_api_url
    GAME_API_POLL_INTERVAL            = "5s"
    GAME_API_POLL_TIMEOUT             = "60s"
    GAME_API_CIRCUIT_BREAKER_THRESHOLD = "5"
    GAME_API_CIRCUIT_BREAKER_TIMEOUT  = "30s"
    LOG_LEVEL                         = "info"
    LOG_FORMAT                        = "json"
    ENABLE_WEBSOCKET_LOBBIES          = "true"
    ENABLE_AUTO_SCORE_RETRIEVAL       = "true"
    ENABLE_DISCONNECT_HANDLING        = "true"
    LOBBY_JOIN_TIMEOUT                = "2m"
    DISCONNECT_RECONNECT_WINDOW       = "30s"
    READY_CHECK_TIMEOUT               = "5m"
    MAX_CONCURRENT_TOURNAMENTS        = "50"
    MAX_PARTICIPANTS_PER_TOURNAMENT   = "256"
    DATABASE_MAX_CONNS                = "25"
    DATABASE_MAX_IDLE_CONNS           = "5"
    POSTGRES_DSN                      = var.postgres_dsn
    DATABASE_URL                      = var.postgres_dsn
    HOST_JWT_SECRET                   = var.jwt_secret
    REDIS_PASSWORD                    = var.redis_password
    SUPABASE_ANON_KEY                 = var.supabase_anon_key
    SUPABASE_SERVICE_KEY              = var.supabase_service_key
    GAME_API_KEY                      = var.game_api_key
  }

  enable_execute_command = true
  tags = { Team = "Platform" }
}

# Game Management Service
module "game_management" {
  source = "../../modules/ecs-service"

  environment            = "dev"
  service_name           = "game-management"
  ecs_cluster_id         = aws_ecs_cluster.main.id
  ecs_cluster_name       = aws_ecs_cluster.main.name
  vpc_id                 = module.vpc.vpc_id
  private_subnet_ids     = module.vpc.public_subnet_ids
  assign_public_ip       = true
  alb_security_group_id  = module.alb.alb_security_group_id

  container_image = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/dev-winspire-game-management:latest"
  container_port  = 8087

  task_cpu    = 256
  task_memory = 512

  desired_count = 1
  min_capacity  = 1
  max_capacity  = 10

  environment_variables = {
    APP_ENV                          = "production"
    SERVICE_PORT                     = "8087"
    REDIS_ADDR                       = "${module.redis.redis_endpoint}:6379"
    HOST_JWT_ISSUER                  = var.jwt_issuer
    HOST_JWT_AUDIENCE                = "authenticated"
    POSTGRES_DSN                     = var.postgres_dsn
    HOST_JWT_SECRET                  = var.jwt_secret
    GAME_MANAGEMENT_INTERNAL_API_KEY = "dev-internal-api-key-12345"
  }

  enable_execute_command = true
  tags = { Team = "Platform" }
}

# Data source for AWS account ID
data "aws_caller_identity" "current" {}

# ALB Listener Rules for path-based routing
# Note: ALB limits 5 values per rule, so we split into multiple rules

resource "aws_lb_listener_rule" "tournament_tournaments" {
  listener_arn = module.alb.http_listener_arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = module.tournament.target_group_arn
  }

  condition {
    path_pattern {
      values = ["/v1/*/tournaments", "/v1/*/tournaments/*"]
    }
  }
}

resource "aws_lb_listener_rule" "tournament_hosts" {
  listener_arn = module.alb.http_listener_arn
  priority     = 101

  action {
    type             = "forward"
    target_group_arn = module.tournament.target_group_arn
  }

  condition {
    path_pattern {
      values = ["/v1/hosts", "/v1/hosts/*", "/v1/*/registrations", "/v1/*/registrations/*"]
    }
  }
}

resource "aws_lb_listener_rule" "matchmaking_brackets" {
  listener_arn = module.alb.http_listener_arn
  priority     = 150

  action {
    type             = "forward"
    target_group_arn = module.matchmaking.target_group_arn
  }

  condition {
    path_pattern {
      values = ["/v1/brackets", "/v1/brackets/*"]
    }
  }
}

resource "aws_lb_listener_rule" "matchmaking_matches" {
  listener_arn = module.alb.http_listener_arn
  priority     = 151

  action {
    type             = "forward"
    target_group_arn = module.matchmaking.target_group_arn
  }

  condition {
    path_pattern {
      values = ["/v1/matches", "/v1/matches/*", "/v1/lobbies", "/v1/lobbies/*"]
    }
  }
}

resource "aws_lb_listener_rule" "game_management" {
  listener_arn = module.alb.http_listener_arn
  priority     = 200

  action {
    type             = "forward"
    target_group_arn = module.game_management.target_group_arn
  }

  condition {
    path_pattern {
      values = ["/v1/games", "/v1/games/*", "/v1/bundles", "/v1/bundles/*"]
    }
  }
}

# HTTPS Listener Rules (duplicates of HTTP rules for HTTPS listener)
resource "aws_lb_listener_rule" "tournament_tournaments_https" {
  listener_arn = module.alb.https_listener_arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = module.tournament.target_group_arn
  }

  condition {
    path_pattern {
      values = ["/v1/*/tournaments", "/v1/*/tournaments/*"]
    }
  }
}

resource "aws_lb_listener_rule" "tournament_hosts_https" {
  listener_arn = module.alb.https_listener_arn
  priority     = 101

  action {
    type             = "forward"
    target_group_arn = module.tournament.target_group_arn
  }

  condition {
    path_pattern {
      values = ["/v1/hosts", "/v1/hosts/*", "/v1/*/registrations", "/v1/*/registrations/*"]
    }
  }
}

resource "aws_lb_listener_rule" "matchmaking_brackets_https" {
  listener_arn = module.alb.https_listener_arn
  priority     = 150

  action {
    type             = "forward"
    target_group_arn = module.matchmaking.target_group_arn
  }

  condition {
    path_pattern {
      values = ["/v1/brackets", "/v1/brackets/*"]
    }
  }
}

resource "aws_lb_listener_rule" "matchmaking_matches_https" {
  listener_arn = module.alb.https_listener_arn
  priority     = 151

  action {
    type             = "forward"
    target_group_arn = module.matchmaking.target_group_arn
  }

  condition {
    path_pattern {
      values = ["/v1/matches", "/v1/matches/*", "/v1/lobbies", "/v1/lobbies/*"]
    }
  }
}

resource "aws_lb_listener_rule" "game_management_https" {
  listener_arn = module.alb.https_listener_arn
  priority     = 200

  action {
    type             = "forward"
    target_group_arn = module.game_management.target_group_arn
  }

  condition {
    path_pattern {
      values = ["/v1/games", "/v1/games/*", "/v1/bundles", "/v1/bundles/*"]
    }
  }
}

# Outputs
output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = module.alb.alb_dns_name
}

output "redis_endpoint" {
  description = "Redis endpoint"
  value       = module.redis.redis_endpoint
  sensitive   = true
}

output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = aws_ecs_cluster.main.name
}

output "service_log_groups" {
  description = "CloudWatch log groups for services"
  value = {
    tournament      = module.tournament.log_group_name
    matchmaking     = module.matchmaking.log_group_name
    game_management = module.game_management.log_group_name
  }
}

