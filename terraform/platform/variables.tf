variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "project_name" {
  type    = string
  default = "ephemeral-platform"
}

variable "vpc_cidr" {
  type    = string
  default = "10.42.0.0/16"
}

variable "route53_zone_id" {
  description = "Existing public Route53 hosted zone that owns the preview domain."
  type        = string
}

variable "preview_domain" {
  description = "Base domain, for example preview.example.com."
  type        = string
}

variable "certificate_arn" {
  description = "ACM certificate covering *.preview.example.com."
  type        = string
}

variable "database_name" {
  type    = string
  default = "preview"
}

variable "database_username" {
  type    = string
  default = "platform"
}

variable "database_instance_class" {
  type    = string
  default = "db.t4g.micro"
}
