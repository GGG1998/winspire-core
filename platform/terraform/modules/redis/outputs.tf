output "redis_endpoint" {
  description = "Primary endpoint of the Redis cluster"
  value       = aws_elasticache_replication_group.main.primary_endpoint_address
}

output "redis_port" {
  description = "Port of the Redis cluster"
  value       = aws_elasticache_replication_group.main.port
}

output "redis_connection_string" {
  description = "Redis connection string (rediss:// with TLS)"
  value       = var.transit_encryption_enabled ? "rediss://${aws_elasticache_replication_group.main.primary_endpoint_address}:${aws_elasticache_replication_group.main.port}" : "redis://${aws_elasticache_replication_group.main.primary_endpoint_address}:${aws_elasticache_replication_group.main.port}"
}

output "redis_security_group_id" {
  description = "Security group ID of the Redis cluster"
  value       = aws_security_group.redis.id
}

output "replication_group_id" {
  description = "ID of the replication group"
  value       = aws_elasticache_replication_group.main.id
}

output "replication_group_arn" {
  description = "ARN of the replication group"
  value       = aws_elasticache_replication_group.main.arn
}

