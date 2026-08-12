variable "environment_id" {
  type = string
}

variable "image" {
  type = string
}

variable "cpu" {
  type = number
}

variable "memory" {
  type = number
}

variable "database_schema" {
  type = string
}

variable "host_name" {
  type = string
}

variable "listener_rule_priority" {
  type = number
}

variable "aws_region" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "subnet_ids" {
  type = list(string)
}

variable "ecs_cluster_arn" {
  type = string
}

variable "https_listener_arn" {
  type = string
}

variable "task_execution_role_arn" {
  type = string
}

variable "task_role_arn" {
  type = string
}

variable "service_security_group_id" {
  type = string
}

variable "assign_public_ip" {
  type    = bool
  default = true
}

variable "database_host" {
  type    = string
  default = ""
}

variable "database_name" {
  type    = string
  default = "preview"
}

variable "database_user" {
  type    = string
  default = "platform"
}

variable "database_password_secret_arn" {
  type    = string
  default = ""
}
