# ECR Repository module for Winspire microservices
# Creates ECR repositories for each service with lifecycle policies and scanning

terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "services" {
  description = "List of service names to create repositories for"
  type        = list(string)
  default     = ["competition", "matchmaking", "game-management"]
}

variable "image_retention_count" {
  description = "Number of images to retain per repository"
  type        = number
  default     = 30
}

variable "tags" {
  description = "Additional tags to apply to resources"
  type        = map(string)
  default     = {}
}

# Create ECR repositories for each service
resource "aws_ecr_repository" "services" {
  for_each = toset(var.services)

  name                 = "${var.environment}-winspire-${each.key}"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }

  tags = merge(
    var.tags,
    {
      Name        = "${var.environment}-winspire-${each.key}"
      Environment = var.environment
      Service     = each.key
      ManagedBy   = "terraform"
    }
  )
}

# Lifecycle policy to manage image retention
resource "aws_ecr_lifecycle_policy" "services" {
  for_each   = aws_ecr_repository.services
  repository = each.value.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last ${var.image_retention_count} images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = var.image_retention_count
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# Repository policy for ECS task execution role access
resource "aws_ecr_repository_policy" "services" {
  for_each   = aws_ecr_repository.services
  repository = each.value.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowECSPull"
        Effect    = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
        Action = [
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:BatchCheckLayerAvailability"
        ]
      }
    ]
  })
}

# Outputs
output "repository_urls" {
  description = "Map of service names to ECR repository URLs"
  value = {
    for name, repo in aws_ecr_repository.services : name => repo.repository_url
  }
}

output "repository_arns" {
  description = "Map of service names to ECR repository ARNs"
  value = {
    for name, repo in aws_ecr_repository.services : name => repo.arn
  }
}

output "registry_id" {
  description = "The registry ID where the repositories are created"
  value       = length(aws_ecr_repository.services) > 0 ? values(aws_ecr_repository.services)[0].registry_id : ""
}
