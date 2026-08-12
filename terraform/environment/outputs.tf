output "preview_url" {
  value = "https://${var.host_name}"
}

output "ecs_service_arn" {
  value = aws_ecs_service.preview.id
}

output "target_group_arn" {
  value = aws_lb_target_group.preview.arn
}

output "listener_rule_arn" {
  value = aws_lb_listener_rule.preview.arn
}

output "task_definition_arn" {
  value = aws_ecs_task_definition.preview.arn
}
