output "lambda_function_name" {
  description = "Name of the deployed Lambda function"
  value       = module.lambda_function.lambda_function_name
}

output "lambda_function_arn" {
  description = "ARN of the deployed Lambda function"
  value       = module.lambda_function.lambda_function_arn
}

output "lambda_public_url" {
  description = "Public URL to access the Lambda function"
  value       = module.lambda_function.lambda_function_url
}

output "lambda_role_arn" {
  description = "IAM role ARN for the Lambda function"
  value       = module.lambda_function.lambda_role_arn
}
