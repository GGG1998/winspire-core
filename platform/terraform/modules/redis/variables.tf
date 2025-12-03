variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs for Redis"
  type        = list(string)
}

variable "allowed_security_groups" {
  description = "List of security group IDs allowed to access Redis"
  type        = list(string)
}

variable "redis_version" {
  description = "Redis version"
  type        = string
  default     = "7.1"
}

variable "node_type" {
  description = "ElastiCache node type"
  type        = string
  default     = "cache.t4g.micro"
}

variable "num_cache_nodes" {
  description = "Number of cache nodes (1 = no failover, 2+ = automatic failover)"
  type        = number
  default     = 2
}

variable "parameter_group_family" {
  description = "Redis parameter group family"
  type        = string
  default     = "redis7"
}

variable "maintenance_window" {
  description = "Maintenance window (e.g., sun:05:00-sun:06:00)"
  type        = string
  default     = "sun:05:00-sun:06:00"
}

variable "snapshot_window" {
  description = "Snapshot window (e.g., 03:00-04:00)"
  type        = string
  default     = "03:00-04:00"
}

variable "snapshot_retention_limit" {
  description = "Number of days to retain snapshots"
  type        = number
  default     = 5
}

variable "at_rest_encryption_enabled" {
  description = "Enable encryption at rest"
  type        = bool
  default     = true
}

variable "transit_encryption_enabled" {
  description = "Enable encryption in transit"
  type        = bool
  default     = true
}

variable "auth_token_enabled" {
  description = "Enable Redis AUTH token"
  type        = bool
  default     = true
}

variable "notification_topic_arn" {
  description = "SNS topic ARN for notifications"
  type        = string
  default     = null
}

variable "enable_cloudwatch_alarms" {
  description = "Enable CloudWatch alarms"
  type        = bool
  default     = true
}

variable "alarm_actions" {
  description = "List of ARNs to notify when alarm triggers"
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Additional tags for resources"
  type        = map(string)
  default     = {}
}

