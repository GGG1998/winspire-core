# Development Environment Configuration
# This is an example showing how to use the Winspire infrastructure modules

terraform {
  required_version = ">= 1.5"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
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
  default     = "us-east-1"
}

variable "vpc_id" {
  description = "VPC ID (must be created separately)"
  type        = string
}

variable "public_subnet_ids" {
  description = "List of public subnet IDs"
  type        = list(string)
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs"
  type        = list(string)
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

  environment              = "dev"
  vpc_id                   = var.vpc_id
  public_subnet_ids        = var.public_subnet_ids
  enable_deletion_protection = false

  tags = {
    Team = "Platform"
  }
}

# Redis for JWT caching and SSE broadcasting
module "redis" {
  source = "../../modules/redis"

  environment        = "dev"
  vpc_id             = var.vpc_id
  private_subnet_ids = var.private_subnet_ids
  
  # Allow ECS services to access Redis
  allowed_security_groups = [
    module.competition_host_stream.security_group_id,
    module.game_management.security_group_id,
  ]

  # Dev settings
  node_type            = "cache.t4g.micro"
  num_cache_nodes      = 1  # Single node for dev
  enable_cloudwatch_alarms = false
  
  tags = {
    Team = "Platform"
  }
}

# Competition Host Stream Service
module "competition_host_stream" {
  source = "../../modules/ecs-service"

  environment            = "dev"
  service_name           = "competition-host-stream"
  ecs_cluster_id         = aws_ecs_cluster.main.id
  ecs_cluster_name       = aws_ecs_cluster.main.name
  vpc_id                 = var.vpc_id
  private_subnet_ids     = var.private_subnet_ids
  alb_security_group_id  = module.alb.alb_security_group_id

  container_image = "competition-host-stream:latest"  # Replace with your ECR image
  container_port  = 8086
  
  task_cpu    = 512
  task_memory = 1024

  desired_count = 2
  min_capacity  = 2
  max_capacity  = 10

  # Enable sticky sessions for SSE connections
  enable_sticky_sessions   = true
  sticky_session_duration  = 3600

  environment_variables = {
    APP_ENV      = "development"
    SERVICE_PORT = "8086"
    REDIS_URL    = module.redis.redis_connection_string
    USE_REDIS_BROKER = "true"
  }

  # Secrets from AWS Secrets Manager/Parameter Store
  secrets = {
    POSTGRES_DSN    = "arn:aws:secretsmanager:${var.aws_region}:ACCOUNT_ID:secret:dev/postgres-dsn"
    HOST_JWT_SECRET = "arn:aws:secretsmanager:${var.aws_region}:ACCOUNT_ID:secret:dev/jwt-secret"
  }

  enable_execute_command = true

  tags = {
    Team = "Platform"
  }
}

# Game Management Service
module "game_management" {
  source = "../../modules/ecs-service"

  environment            = "dev"
  service_name           = "game-management"
  ecs_cluster_id         = aws_ecs_cluster.main.id
  ecs_cluster_name       = aws_ecs_cluster.main.name
  vpc_id                 = var.vpc_id
  private_subnet_ids     = var.private_subnet_ids
  alb_security_group_id  = module.alb.alb_security_group_id

  container_image = "game-management:latest"  # Replace with your ECR image
  container_port  = 8085
  
  task_cpu    = 256
  task_memory = 512

  desired_count = 2
  min_capacity  = 2
  max_capacity  = 10

  environment_variables = {
    APP_ENV      = "development"
    SERVICE_PORT = "8085"
    REDIS_URL    = module.redis.redis_connection_string
  }

  secrets = {
    POSTGRES_DSN    = "arn:aws:secretsmanager:${var.aws_region}:ACCOUNT_ID:secret:dev/postgres-dsn"
    HOST_JWT_SECRET = "arn:aws:secretsmanager:${var.aws_region}:ACCOUNT_ID:secret:dev/jwt-secret"
  }

  enable_execute_command = true

  tags = {
    Team = "Platform"
  }
}

# ALB Listener Rules for path-based routing
resource "aws_lb_listener_rule" "competition_host_stream" {
  listener_arn = module.alb.http_listener_arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = module.competition_host_stream.target_group_arn
  }

  condition {
    path_pattern {
      values = ["/v1/stream/*", "/v1/cups/*", "/v1/tournaments/*", "/v1/matches/*"]
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
      values = ["/v1/games/*", "/v1/bundles/*", "/v1/admin/*"]
    }
  }
}

# Outputs
output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = module.alb.alb_dns_name
}

output "redis_endpoint" {
  description = "Redis endpoint"
  value       = module.redis.redis_endpoint
  sensitive   = true
}

output "competition_host_stream_log_group" {
  description = "CloudWatch log group for competition-host-stream"
  value       = module.competition_host_stream.log_group_name
}

output "game_management_log_group" {
  description = "CloudWatch log group for game-management"
  value       = module.game_management.log_group_name
}

