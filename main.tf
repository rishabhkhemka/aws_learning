terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
}

provider "aws" {
  region                   = "ap-south-1"
  shared_credentials_files = ["$HOME/.aws/credentials"]
}

resource "aws_dynamodb_table" "user_contact_info_tf" {
  name           = "user_contact_info_tf"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "userID"
  stream_enabled = false

  attribute {
    name = "userID"
    type = "S"
  }

  attribute {
    name = "firstName"
    type = "S"
  }

  attribute {
    name = "lastName"
    type = "S"
  }

  global_secondary_index {
    name               = "firstNameIndex"
    hash_key           = "firstName"
    projection_type    = "ALL"
    read_capacity      = 5
    write_capacity     = 5
    non_key_attributes = []
  }

  global_secondary_index {
    name               = "lastNameIndex"
    hash_key           = "lastName"
    projection_type    = "ALL"
    read_capacity      = 5
    write_capacity     = 5
    non_key_attributes = []
  }
}

# user_tf lambda
resource "aws_lambda_function" "user_tf" {
  function_name = "user_tf"
  runtime       = "go1.x"
  handler       = "main"
  filename      = "./lambdas/user.zip"
  role          = aws_iam_role.lambda_role_tf.arn
  environment {
    variables = {
      TABLE_NAME = "user_contact_info_tf"
    }
  }
}

# users_tf lambda
resource "aws_lambda_function" "users_tf" {
  function_name = "users_tf"
  runtime       = "go1.x"
  handler       = "main"
  filename      = "./lambdas/users.zip"
  role          = aws_iam_role.lambda_role_tf.arn
  environment {
    variables = {
      TABLE_NAME = "user_contact_info_tf"
    }
  }
}

# iam role for lambdas
resource "aws_iam_role" "lambda_role_tf" {
  name = "lambda_role_tf"
  assume_role_policy = jsonencode(
    {
      "Version" : "2012-10-17",
      "Statement" : [
        {
          "Effect" : "Allow",
          "Principal" : {
            "Service" : "lambda.amazonaws.com"
          },
          "Action" : "sts:AssumeRole"
        }
      ]
    },
  )
}

# policy for the lambda role
resource "aws_iam_policy" "lambda_policy_tf" {
  name = "lambda_policy_tf"
  policy = jsonencode(
    {
      "Version" : "2012-10-17",
      "Id" : "default",
      "Statement" : [
        {
          "Effect" : "Allow",
          "Action" : "lambda:*",
          "Resource" : "*",
        },
        {
          "Effect" : "Allow",
          "Action" : "dynamodb:*",
          "Resource" : "*"
        }
      ]
    }
  )
}

# attaching lambda policy with the role
resource "aws_iam_role_policy_attachment" "role_policy_attachment_tf" {
  role       = aws_iam_role.lambda_role_tf.name
  policy_arn = aws_iam_policy.lambda_policy_tf.arn
}

resource "aws_lambda_permission" "api_gw" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.users_tf.function_name
  principal     = "apigateway.amazonaws.com"
}

# api gateway
resource "aws_api_gateway_rest_api" "user_info_api_tf" {
  name        = "user_info_api_tf"
  description = "api for user contact management"
  policy      = jsonencode(
    {
      "Version" : "2012-10-17",
      "Statement" : [
        {
          "Effect" : "Allow",
          "Principal" : {
            "Service" : "apigateway.amazonaws.com"
          },
          "Action" : "sts:AssumeRole"
        }
      ]
    }
  )
}

# resource on gateway
resource "aws_api_gateway_resource" "users" {
  rest_api_id = aws_api_gateway_rest_api.user_info_api_tf.id
  parent_id   = aws_api_gateway_rest_api.user_info_api_tf.root_resource_id
  path_part   = "users"
}

# GET /users
resource "aws_api_gateway_method" "get_users" {
  rest_api_id   = aws_api_gateway_rest_api.user_info_api_tf.id
  resource_id   = aws_api_gateway_resource.users.id
  http_method   = "GET"
  authorization = "NONE"
}

# /users GET method integration with lambda
resource "aws_api_gateway_integration" "get_users_integration" {
  rest_api_id             = aws_api_gateway_rest_api.user_info_api_tf.id
  resource_id             = aws_api_gateway_resource.users.id
  http_method             = aws_api_gateway_method.get_users.http_method
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.users_tf.invoke_arn
  timeout_milliseconds    = 29000
  integration_http_method = "POST"
}

# /users GET method integration response
resource "aws_api_gateway_integration_response" "get_users_integration_response" {
  rest_api_id = aws_api_gateway_rest_api.user_info_api_tf.id
  resource_id = aws_api_gateway_resource.users.id
  http_method = aws_api_gateway_method.get_users.http_method
  status_code = "200"
}

resource "aws_api_gateway_deployment" "deploy" {
  rest_api_id = aws_api_gateway_rest_api.user_info_api_tf.id
  stage_name  = "dev"

  depends_on = [
    aws_api_gateway_integration.get_users_integration
  ]
}