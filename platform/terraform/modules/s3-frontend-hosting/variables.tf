variable "bucket_name" {
  description = "Name of the S3 bucket for frontend hosting"
  type        = string
}

variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
}

variable "index_document" {
  description = "Index document for static website"
  type        = string
  default     = "index.html"
}

variable "error_document" {
  description = "Error document for static website"
  type        = string
  default     = "index.html"
}

variable "spa_mode" {
  description = "Enable SPA mode (redirect 404s to index.html for client-side routing)"
  type        = bool
  default     = true
}

variable "cors_allowed_origins" {
  description = "List of allowed origins for CORS"
  type        = list(string)
  default     = ["*"]
}

variable "enable_lifecycle_rules" {
  description = "Enable lifecycle rules for cleanup"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Additional tags for resources"
  type        = map(string)
  default     = {}
}




