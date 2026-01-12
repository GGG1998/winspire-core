# App Runner Module Variables

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "service_name" {
  description = "Name of the service"
  type        = string
}

variable "container_image" {
  description = "ECR image URI (e.g., 123456789.dkr.ecr.region.amazonaws.com/repo:tag)"
  type        = string
}

variable "container_port" {
  description = "Port the container listens on"
  type        = number
  default     = 8081
}

variable "cpu" {
  description = "CPU units (256, 512, 1024, 2048, 4096)"
  type        = string
  default     = "256"
}

variable "memory" {
  description = "Memory in MB (512, 1024, 2048, 3072, 4096, 6144, 8192, 10240, 12288)"
  type        = string
  default     = "512"
}

variable "min_instances" {
  description = "Minimum number of instances (0 for scale-to-zero)"
  type        = number
  default     = 0
}

variable "max_instances" {
  description = "Maximum number of instances"
  type        = number
  default     = 5
}

variable "max_concurrency" {
  description = "Maximum concurrent requests per instance before scaling"
  type        = number
  default     = 100
}

variable "health_check_path" {
  description = "Health check endpoint path"
  type        = string
  default     = "/health"
}

variable "environment_variables" {
  description = "Environment variables for the container"
  type        = map(string)
  default     = {}
}

variable "auto_deploy" {
  description = "Enable automatic deployments when new image is pushed"
  type        = bool
  default     = true
}

variable "instance_policy_json" {
  description = "Custom IAM policy JSON for instance role (e.g., S3 access)"
  type        = string
  default     = null
}

variable "custom_domain" {
  description = "Custom domain name (e.g., dev-api.gowinspire.com). Set to null to skip."
  type        = string
  default     = null
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}
