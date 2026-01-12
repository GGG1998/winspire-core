# AWS App Runner Service Module
# Serverless container hosting with scale-to-zero support

terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# IAM Role for App Runner to access ECR
resource "aws_iam_role" "apprunner_ecr_access" {
  name = "${var.environment}-${var.service_name}-apprunner-ecr"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "build.apprunner.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "apprunner_ecr_access" {
  role       = aws_iam_role.apprunner_ecr_access.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSAppRunnerServicePolicyForECRAccess"
}

# IAM Role for App Runner instance (runtime permissions)
resource "aws_iam_role" "apprunner_instance" {
  name = "${var.environment}-${var.service_name}-apprunner-instance"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "tasks.apprunner.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

# Attach custom policy if provided (e.g., S3 access)
resource "aws_iam_role_policy" "apprunner_instance_custom" {
  count = var.instance_policy_json != null ? 1 : 0

  name   = "${var.environment}-${var.service_name}-custom"
  role   = aws_iam_role.apprunner_instance.id
  policy = var.instance_policy_json
}

# App Runner Auto Scaling Configuration
resource "aws_apprunner_auto_scaling_configuration_version" "main" {
  auto_scaling_configuration_name = "${var.environment}-${var.service_name}"

  min_size = var.min_instances
  max_size = var.max_instances

  # Scale based on concurrent requests
  max_concurrency = var.max_concurrency

  tags = var.tags
}

# App Runner Service
resource "aws_apprunner_service" "main" {
  service_name = "${var.environment}-${var.service_name}"

  source_configuration {
    authentication_configuration {
      access_role_arn = aws_iam_role.apprunner_ecr_access.arn
    }

    image_repository {
      image_identifier      = var.container_image
      image_repository_type = "ECR"

      image_configuration {
        port = tostring(var.container_port)

        runtime_environment_variables = var.environment_variables
      }
    }

    auto_deployments_enabled = var.auto_deploy
  }

  instance_configuration {
    cpu               = var.cpu
    memory            = var.memory
    instance_role_arn = aws_iam_role.apprunner_instance.arn
  }

  auto_scaling_configuration_arn = aws_apprunner_auto_scaling_configuration_version.main.arn

  health_check_configuration {
    protocol            = "HTTP"
    path                = var.health_check_path
    interval            = 10
    timeout             = 5
    healthy_threshold   = 1
    unhealthy_threshold = 5
  }

  network_configuration {
    egress_configuration {
      egress_type = "DEFAULT"
    }
  }

  tags = merge(
    var.tags,
    {
      Name        = "${var.environment}-${var.service_name}"
      Environment = var.environment
      Service     = var.service_name
    }
  )
}

# Custom Domain (optional)
resource "aws_apprunner_custom_domain_association" "main" {
  count = var.custom_domain != null ? 1 : 0

  domain_name = var.custom_domain
  service_arn = aws_apprunner_service.main.arn

  # Enable www subdomain as well
  enable_www_subdomain = false
}

# CloudWatch Log Group (App Runner creates its own, but we can reference it)
data "aws_region" "current" {}
