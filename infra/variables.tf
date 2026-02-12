variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "default_tags" {
  description = "Default tags to apply to all resources"
  type        = map(string)
  default = {
    Environment = "dev"
    Owner       = "salsiy"
  }
}

variable "github_app_id_ssm_path" {
  description = "SSM Parameter Store path for GitHub App ID"
  type        = string
  default     = "/dev/github-app/app-id"
}

variable "github_app_private_key_ssm_path" {
  description = "SSM Parameter Store path for GitHub App private key"
  type        = string
  default     = "/dev/github-app-private-key"
}

variable "github_app_webhook_secret_ssm_path" {
  description = "SSM Parameter Store path for GitHub App webhook secret"
  type        = string
  default     = "/dev/github-app-webhook-secret"
}
