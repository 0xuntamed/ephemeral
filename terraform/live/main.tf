module "environment" {
  source = "../environment"

  environment_id               = var.environment_id
  image                        = var.image
  cpu                          = var.cpu
  memory                       = var.memory
  database_schema              = var.database_schema
  host_name                    = var.host_name
  listener_rule_priority       = var.listener_rule_priority
  aws_region                   = var.aws_region
  vpc_id                       = var.vpc_id
  subnet_ids                   = var.subnet_ids
  ecs_cluster_arn              = var.ecs_cluster_arn
  https_listener_arn           = var.https_listener_arn
  task_execution_role_arn      = var.task_execution_role_arn
  task_role_arn                = var.task_role_arn
  service_security_group_id    = var.service_security_group_id
  assign_public_ip             = var.assign_public_ip
  database_host                = var.database_host
  database_name                = var.database_name
  database_user                = var.database_user
  database_password_secret_arn = var.database_password_secret_arn
}

output "preview_url" {
  value = module.environment.preview_url
}

output "ecs_service_arn" {
  value = module.environment.ecs_service_arn
}

output "target_group_arn" {
  value = module.environment.target_group_arn
}

output "listener_rule_arn" {
  value = module.environment.listener_rule_arn
}
