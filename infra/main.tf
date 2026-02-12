terraform {
  required_version = ">= 1.5"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
  
  default_tags {
    tags = var.default_tags
  }
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Local values
locals {
  account_id = data.aws_caller_identity.current.account_id
  region     = data.aws_region.current.name
  function_name = "serverless-github-app-${local.account_id}-${local.region}"
}

module "lambda_function" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "~> 7.0"

  function_name = local.function_name
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = ["arm64"]
  
  source_path = [
    {
      path = "${path.module}/../app"
      commands = [
        "GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap .",
        ":zip"
      ]
    }
  ]

  timeout     = 30
  memory_size = 256

  environment_variables = {
    ENVIRONMENT                     = var.environment
    SSM_GITHUB_APP_ID               = var.github_app_id_ssm_path
    SSM_GITHUB_APP_PRIVATE_KEY      = var.github_app_private_key_ssm_path
    SSM_GITHUB_APP_WEBHOOK_SECRET   = var.github_app_webhook_secret_ssm_path
  }

  create_lambda_function_url = true
  authorization_type         = "NONE"

  # Attach additional IAM policies
  attach_policy_statements = true
  policy_statements = {
    ssm_read = {
      effect = "Allow"
      actions = [
        "ssm:GetParameter",
        "ssm:GetParameters"
      ]
      resources = [
        "arn:aws:ssm:${local.region}:${local.account_id}:parameter${var.github_app_id_ssm_path}",
        "arn:aws:ssm:${local.region}:${local.account_id}:parameter${var.github_app_private_key_ssm_path}",
        "arn:aws:ssm:${local.region}:${local.account_id}:parameter${var.github_app_webhook_secret_ssm_path}"
      ]
    }
  }
}
