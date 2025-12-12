# S3 Static Website Hosting module for Mini Admin
# Provides an S3 bucket with optional versioning and static website configuration

terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# S3 Bucket
resource "aws_s3_bucket" "main" {
  bucket = var.bucket_name

  tags = merge(
    var.tags,
    {
      Name        = var.bucket_name
      Environment = var.environment
      Purpose     = "Static file hosting"
    }
  )
}

# Bucket Versioning
resource "aws_s3_bucket_versioning" "main" {
  count  = var.enable_versioning ? 1 : 0
  bucket = aws_s3_bucket.main.id

  versioning_configuration {
    status = "Enabled"
  }
}

# Bucket Public Access Block Configuration
resource "aws_s3_bucket_public_access_block" "main" {
  bucket = aws_s3_bucket.main.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

# Bucket Policy for Public Read Access
resource "aws_s3_bucket_policy" "main" {
  bucket = aws_s3_bucket.main.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "PublicReadGetObject"
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:GetObject"
        Resource  = "${aws_s3_bucket.main.arn}/*"
      }
    ]
  })

  depends_on = [aws_s3_bucket_public_access_block.main]
}

# Static Website Configuration
resource "aws_s3_bucket_website_configuration" "main" {
  count  = var.enable_static_website ? 1 : 0
  bucket = aws_s3_bucket.main.id

  index_document {
    suffix = var.index_document
  }

  error_document {
    key = var.error_document
  }
}

# CORS Configuration
resource "aws_s3_bucket_cors_configuration" "main" {
  bucket = aws_s3_bucket.main.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "HEAD"]
    allowed_origins = var.cors_allowed_origins
    expose_headers  = ["ETag"]
    max_age_seconds = 3600
  }

  dynamic "cors_rule" {
    for_each = var.enable_uploads ? [1] : []
    content {
      allowed_headers = ["*"]
      allowed_methods = ["PUT", "POST", "DELETE"]
      allowed_origins = var.cors_allowed_origins
      expose_headers  = ["ETag"]
      max_age_seconds = 3600
    }
  }
}

# Server-Side Encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "main" {
  bucket = aws_s3_bucket.main.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Lifecycle Rules (optional cleanup)
resource "aws_s3_bucket_lifecycle_configuration" "main" {
  count  = var.enable_lifecycle_rules ? 1 : 0
  bucket = aws_s3_bucket.main.id

  rule {
    id     = "delete-old-versions"
    status = "Enabled"

    noncurrent_version_expiration {
      noncurrent_days = var.noncurrent_version_expiration_days
    }
  }

  rule {
    id     = "delete-incomplete-uploads"
    status = "Enabled"

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}

# CloudWatch Metrics (optional)
resource "aws_s3_bucket_metric" "main" {
  count  = var.enable_metrics ? 1 : 0
  bucket = aws_s3_bucket.main.id
  name   = "EntireBucket"
}










