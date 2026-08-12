terraform {
  required_version = ">= 1.10.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }

  backend "s3" {
    key                  = "ephemeral/platform.tfstate"
    workspace_key_prefix = "environments"
    use_lockfile         = true
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project   = "ephemeral-platform"
      ManagedBy = "terraform"
    }
  }
}
