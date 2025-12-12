variable "bucket_name" {
  description = "Name of the S3 bucket"
  type        = string
}

variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
}

variable "enable_versioning" {
  description = "Enable versioning on the S3 bucket"
  type        = bool
  default     = false
}

variable "enable_static_website" {
  description = "Enable static website hosting"
  type        = bool
  default     = true
}

variable "index_document" {
  description = "Index document for static website"
  type        = string
  default     = "index.html"
}

variable "error_document" {
  description = "Error document for static website"
  type        = string
  default     = "error.html"
}

variable "cors_allowed_origins" {
  description = "List of allowed origins for CORS"
  type        = list(string)
  default     = ["*"]
}

variable "enable_uploads" {
  description = "Enable CORS rules for file uploads (PUT, POST, DELETE)"
  type        = bool
  default     = true
}

variable "enable_lifecycle_rules" {
  description = "Enable lifecycle rules for cleanup"
  type        = bool
  default     = true
}

variable "noncurrent_version_expiration_days" {
  description = "Number of days before noncurrent versions expire"
  type        = number
  default     = 90
}

variable "enable_metrics" {
  description = "Enable CloudWatch metrics"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Additional tags for resources"
  type        = map(string)
  default     = {}
}










