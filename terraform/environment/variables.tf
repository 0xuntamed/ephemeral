variable "environment_id" {
  description = "Stable, DNS-safe identifier allocated by the control plane."
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9-]{2,62}$", var.environment_id))
    error_message = "environment_id must be a 3-63 character lowercase DNS label."
  }
}

variable "image" {
  description = "Immutable container image URI, ideally including a digest."
  type        = string
}

variable "cpu" {
  description = "Fargate CPU units."
  type        = number
  default     = 512
}

variable "memory" {
  description = "Fargate memory in MiB."
  type        = number
  default     = 1024
}

variable "container_port" {
  description = "Port exposed by the preview workload."
  type        = number
  default     = 8080
}

variable "health_check_path" {
  description = "HTTP endpoint used by the target group."
  type        = string
  default     = "/healthz"
}

variable "host_name" {
  description = "Host matched by the shared ALB listener rule."
  type        = string
}

variable "listener_rule_priority" {
  description = "Unique listener rule priority allocated by the control plane."
  type        = number
}

variable "database_schema" {
  description = "Isolated schema created for this environment, or an empty string."
  type        = string
  default     = ""
}

variable "aws_region" {
  type    = string
  default = "us-east-1"
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
  description = "Useful for the low-cost demo VPC; use private subnets in production."
  type        = bool
  default     = true
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
  description = "Secrets Manager ARN whose JSON document contains a password key."
  type        = string
  default     = ""
}

variable "tags" {
  type    = map(string)
  default = {}
}
