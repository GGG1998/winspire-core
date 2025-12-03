variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID where the ALB will be created"
  type        = string
}

variable "public_subnet_ids" {
  description = "List of public subnet IDs for the ALB"
  type        = list(string)
}

variable "certificate_arn" {
  description = "ARN of ACM certificate for HTTPS (optional)"
  type        = string
  default     = null
}

variable "enable_deletion_protection" {
  description = "Enable deletion protection for the ALB"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Additional tags for resources"
  type        = map(string)
  default     = {}
}

