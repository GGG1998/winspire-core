# Redis (ElastiCache) module for Winspire
# Provides Redis cluster for JWT caching and SSE broadcasting

terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# ElastiCache Subnet Group
resource "aws_elasticache_subnet_group" "main" {
  name       = "${var.environment}-winspire-redis"
  subnet_ids = var.private_subnet_ids

  tags = merge(
    var.tags,
    {
      Name        = "${var.environment}-winspire-redis-subnet-group"
      Environment = var.environment
    }
  )
}

# Security Group for Redis
resource "aws_security_group" "redis" {
  name        = "${var.environment}-winspire-redis-sg"
  description = "Security group for Winspire Redis cluster"
  vpc_id      = var.vpc_id

  ingress {
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = var.allowed_security_groups
    description     = "Allow Redis traffic from allowed security groups"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound traffic"
  }

  tags = merge(
    var.tags,
    {
      Name        = "${var.environment}-winspire-redis-sg"
      Environment = var.environment
    }
  )
}

# ElastiCache Replication Group (Redis Cluster)
resource "aws_elasticache_replication_group" "main" {
  replication_group_id = "${var.environment}-winspire-redis"
  description          = "Redis cluster for Winspire ${var.environment}"

  engine               = "redis"
  engine_version       = var.redis_version
  node_type            = var.node_type
  num_cache_clusters   = var.num_cache_nodes
  port                 = 6379
  parameter_group_name = aws_elasticache_parameter_group.main.name
  subnet_group_name    = aws_elasticache_subnet_group.main.name
  security_group_ids   = [aws_security_group.redis.id]

  # Automatic failover for multi-node clusters
  automatic_failover_enabled = var.num_cache_nodes > 1

  # Maintenance and backup
  maintenance_window       = var.maintenance_window
  snapshot_window          = var.snapshot_window
  snapshot_retention_limit = var.snapshot_retention_limit

  # Encryption
  at_rest_encryption_enabled = var.at_rest_encryption_enabled
  transit_encryption_enabled = var.transit_encryption_enabled

  # Notifications
  notification_topic_arn = var.notification_topic_arn

  # Auto minor version upgrade
  auto_minor_version_upgrade = true

  tags = merge(
    var.tags,
    {
      Name        = "${var.environment}-winspire-redis"
      Environment = var.environment
    }
  )
}

# Parameter Group for Redis configuration
resource "aws_elasticache_parameter_group" "main" {
  name   = "${var.environment}-winspire-redis-params"
  family = var.parameter_group_family

  # Optimize for cache workload
  parameter {
    name  = "maxmemory-policy"
    value = "allkeys-lru"
  }

  parameter {
    name  = "timeout"
    value = "300"
  }

  parameter {
    name  = "tcp-keepalive"
    value = "300"
  }

  tags = merge(
    var.tags,
    {
      Name        = "${var.environment}-winspire-redis-params"
      Environment = var.environment
    }
  )
}

# CloudWatch Alarms for Redis
resource "aws_cloudwatch_metric_alarm" "cpu" {
  count = var.enable_cloudwatch_alarms ? 1 : 0

  alarm_name          = "${var.environment}-winspire-redis-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/ElastiCache"
  period              = "300"
  statistic           = "Average"
  threshold           = "75"
  alarm_description   = "Redis CPU utilization"
  alarm_actions       = var.alarm_actions

  dimensions = {
    ReplicationGroupId = aws_elasticache_replication_group.main.id
  }

  tags = merge(
    var.tags,
    {
      Name        = "${var.environment}-winspire-redis-cpu-alarm"
      Environment = var.environment
    }
  )
}

resource "aws_cloudwatch_metric_alarm" "memory" {
  count = var.enable_cloudwatch_alarms ? 1 : 0

  alarm_name          = "${var.environment}-winspire-redis-memory"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "DatabaseMemoryUsagePercentage"
  namespace           = "AWS/ElastiCache"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "Redis memory utilization"
  alarm_actions       = var.alarm_actions

  dimensions = {
    ReplicationGroupId = aws_elasticache_replication_group.main.id
  }

  tags = merge(
    var.tags,
    {
      Name        = "${var.environment}-winspire-redis-memory-alarm"
      Environment = var.environment
    }
  )
}

