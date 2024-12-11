provider "aws" {
  region = "us-east-1"
}

resource "aws_ecr_repository" "http_crud_tutorial" {
  name                 = "http-crud-tutorial-repository"
  force_delete         = true
  image_tag_mutability = "IMMUTABLE"
  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_iam_role" "lambda_execution" {
  name = "lambda-execution-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_dynamodb_table" "http_crud_tutorial_items" {
  name         = "http-crud-tutorial-items"
  billing_mode = "PAY_PER_REQUEST"
  attribute {
    name = "id"
    type = "S"
  }
  hash_key = "id"
  tags = {
    Environment = "Tutorial"
    Application = "HTTP-CRUD"
  }
}

resource "aws_iam_policy" "dynamodb_access" {
  name        = "dynamodb-access-policy"
  description = "Policy to allow Lambda to interact with DynamoDB"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:DeleteItem",
          "dynamodb:Scan",
          "dynamodb:Query",
          "dynamodb:UpdateItem"
        ]
        Resource = aws_dynamodb_table.http_crud_tutorial_items.arn
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_dynamodb_policy" {
  role       = aws_iam_role.lambda_execution.name
  policy_arn = aws_iam_policy.dynamodb_access.arn
}

resource "aws_iam_role_policy_attachment" "lambda_execution_policy" {
  role       = aws_iam_role.lambda_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_lambda_permission" "invoking_permission" {
  statement_id  = "AllowAPIGatewayInvokeLambda"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.http_crud_tutorial.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.http_crud_tutorial_api.execution_arn}/*"
}

resource "aws_lambda_function" "http_crud_tutorial" {
  function_name = "http-crud-tutorial-function"
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.http_crud_tutorial.repository_url}:v1"
  role          = aws_iam_role.lambda_execution.arn
  architectures = ["arm64"]
  memory_size   = 128
  timeout       = 10
}

resource "aws_apigatewayv2_api" "http_crud_tutorial_api" {
  name          = "http-crud-tutorial-api"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_stage" "default_stage" {
  api_id      = aws_apigatewayv2_api.http_crud_tutorial_api.id
  name        = "$default"
  auto_deploy = true
}

resource "aws_apigatewayv2_integration" "lambda_integration" {
  api_id                 = aws_apigatewayv2_api.http_crud_tutorial_api.id
  integration_type       = "AWS_PROXY"
  integration_method     = "POST"
  integration_uri        = aws_lambda_function.http_crud_tutorial.invoke_arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "get_item_by_id" {
  api_id    = aws_apigatewayv2_api.http_crud_tutorial_api.id
  route_key = "GET /items/{id}"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

resource "aws_apigatewayv2_route" "get_items" {
  api_id    = aws_apigatewayv2_api.http_crud_tutorial_api.id
  route_key = "GET /items"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

resource "aws_apigatewayv2_route" "put_items" {
  api_id    = aws_apigatewayv2_api.http_crud_tutorial_api.id
  route_key = "PUT /items"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

resource "aws_apigatewayv2_route" "delete_item_by_id" {
  api_id    = aws_apigatewayv2_api.http_crud_tutorial_api.id
  route_key = "DELETE /items/{id}"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

output "ecr_repository_url" {
  value       = aws_ecr_repository.http_crud_tutorial.repository_url
  description = "The value of ECR repo URL"
}

output "api_endpoint" {
  value       = aws_apigatewayv2_api.http_crud_tutorial_api.api_endpoint
  description = "The HTTP API endpoint URL"
}