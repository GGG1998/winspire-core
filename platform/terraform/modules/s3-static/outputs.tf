output "bucket_id" {
  description = "The ID of the S3 bucket"
  value       = aws_s3_bucket.main.id
}

output "bucket_arn" {
  description = "The ARN of the S3 bucket"
  value       = aws_s3_bucket.main.arn
}

output "bucket_domain_name" {
  description = "The bucket domain name"
  value       = aws_s3_bucket.main.bucket_domain_name
}

output "bucket_regional_domain_name" {
  description = "The bucket regional domain name"
  value       = aws_s3_bucket.main.bucket_regional_domain_name
}

output "website_endpoint" {
  description = "The website endpoint, if the bucket is configured as a website"
  value       = var.enable_static_website ? aws_s3_bucket_website_configuration.main[0].website_endpoint : null
}

output "website_domain" {
  description = "The domain of the website endpoint"
  value       = var.enable_static_website ? aws_s3_bucket_website_configuration.main[0].website_domain : null
}










