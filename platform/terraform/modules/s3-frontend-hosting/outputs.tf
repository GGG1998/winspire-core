output "bucket_id" {
  description = "The ID of the S3 bucket"
  value       = aws_s3_bucket.frontend.id
}

output "bucket_arn" {
  description = "The ARN of the S3 bucket"
  value       = aws_s3_bucket.frontend.arn
}

output "bucket_domain_name" {
  description = "The bucket domain name"
  value       = aws_s3_bucket.frontend.bucket_domain_name
}

output "bucket_regional_domain_name" {
  description = "The bucket regional domain name"
  value       = aws_s3_bucket.frontend.bucket_regional_domain_name
}

output "website_endpoint" {
  description = "The website endpoint URL"
  value       = aws_s3_bucket_website_configuration.frontend.website_endpoint
}

output "website_domain" {
  description = "The domain of the website endpoint"
  value       = aws_s3_bucket_website_configuration.frontend.website_domain
}

output "website_url" {
  description = "The full HTTP URL of the website"
  value       = "http://${aws_s3_bucket_website_configuration.frontend.website_endpoint}"
}

