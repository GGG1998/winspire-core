variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "service_name" {
  description = "Name of the microservice"
  type        = string
}

variable "ecs_cluster_id" {
  description = "ID of the ECS cluster"
  type        = string
}

variable "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "private_subnet_ids" {
  description = "List of subnet IDs for ECS tasks"
  type        = list(string)
}

variable "assign_public_ip" {
  description = "Assign public IP to ECS tasks (required for public subnets without NAT)"
  type        = bool
  default     = false
}

variable "alb_security_group_id" {
  description = "Security group ID of the ALB"
  type        = string
}

variable "container_image" {
  description = "Docker image for the service"
  type        = string
}

variable "container_port" {
  description = "Port the container listens on"
  type        = number
  default     = 8080
}

variable "task_cpu" {
  description = "CPU units for the task (1024 = 1 vCPU)"
  type        = number
  default     = 256
}

variable "task_memory" {
  description = "Memory for the task in MB"
  type        = number
  default     = 512
}

variable "desired_count" {
  description = "Desired number of tasks"
  type        = number
  default     = 2
}

variable "enable_auto_scaling" {
  description = "Enable auto scaling for the service"
  type        = bool
  default     = true
}

variable "min_capacity" {
  description = "Minimum number of tasks"
  type        = number
  default     = 2
}

variable "max_capacity" {
  description = "Maximum number of tasks"
  type        = number
  default     = 20
}

variable "cpu_target_value" {
  description = "Target CPU utilization percentage for auto scaling"
  type        = number
  default     = 70
}

variable "memory_target_value" {
  description = "Target memory utilization percentage for auto scaling"
  type        = number
  default     = 80
}

variable "enable_sticky_sessions" {
  description = "Enable sticky sessions on the target group (for SSE)"
  type        = bool
  default     = false
}

variable "sticky_session_duration" {
  description = "Duration of sticky sessions in seconds"
  type        = number
  default     = 3600
}

variable "environment_variables" {
  description = "Map of environment variables to set in the container"
  type        = map(string)
  default     = {}
}

variable "secrets" {
  description = "Map of secret names to ARNs (from Secrets Manager or Parameter Store)"
  type        = map(string)
  default     = {}
}

variable "log_retention_days" {
  description = "CloudWatch log retention in days"
  type        = number
  default     = 30
}

variable "enable_execute_command" {
  description = "Enable ECS Exec for debugging"
  type        = bool
  default     = false
}

variable "task_policy_json" {
  description = "Custom IAM policy JSON for the task role"
  type        = string
  default     = null
}

variable "health_check_path" {
  description = "Health check path for the target group"
  type        = string
  default     = "/healthz"
}

variable "tags" {
  description = "Additional tags for resources"
  type        = map(string)
  default     = {}
}

