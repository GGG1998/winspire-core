# App Runner Module Outputs

output "service_url" {
  description = "The URL of the App Runner service (HTTPS)"
  value       = "https://${aws_apprunner_service.main.service_url}"
}

output "service_arn" {
  description = "ARN of the App Runner service"
  value       = aws_apprunner_service.main.arn
}

output "service_id" {
  description = "ID of the App Runner service"
  value       = aws_apprunner_service.main.service_id
}

output "service_status" {
  description = "Current status of the App Runner service"
  value       = aws_apprunner_service.main.status
}

output "auto_scaling_configuration_arn" {
  description = "ARN of the auto scaling configuration"
  value       = aws_apprunner_auto_scaling_configuration_version.main.arn
}

output "custom_domain_dns_records" {
  description = "DNS records to add for custom domain validation"
  value = var.custom_domain != null ? {
    for record in aws_apprunner_custom_domain_association.main[0].certificate_validation_records :
    record.name => {
      type  = record.type
      value = record.value
    }
  } : {}
}
