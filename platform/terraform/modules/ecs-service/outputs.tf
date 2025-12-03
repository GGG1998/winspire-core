output "service_id" {
  description = "ID of the ECS service"
  value       = aws_ecs_service.main.id
}

output "service_name" {
  description = "Name of the ECS service"
  value       = aws_ecs_service.main.name
}

output "task_definition_arn" {
  description = "ARN of the task definition"
  value       = aws_ecs_task_definition.main.arn
}

output "target_group_arn" {
  description = "ARN of the target group"
  value       = aws_lb_target_group.main.arn
}

output "security_group_id" {
  description = "Security group ID of the ECS service"
  value       = aws_security_group.ecs_service.id
}

output "log_group_name" {
  description = "Name of the CloudWatch log group"
  value       = aws_cloudwatch_log_group.main.name
}

output "task_role_arn" {
  description = "ARN of the task IAM role"
  value       = aws_iam_role.ecs_task.arn
}

output "execution_role_arn" {
  description = "ARN of the execution IAM role"
  value       = aws_iam_role.ecs_execution.arn
}

