# --- CodeBuild ---
# artifact storage
resource "aws_s3_bucket" "builds_bucket" {
  bucket_prefix = "bsl-builds-bucket"
  acl           = "private"
}

# permissions
resource "aws_iam_role" "build_role" {
  name_prefix = "bsl-builds-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "codebuild.amazonaws.com"
        }
      },
    ]
  })
}
resource "aws_iam_policy" "build_policy" {
  name_prefix = "bsl-builds-policy"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      # CodeBuild
      {
        Effect = "Allow"
        Action = [
          "codebuild:CreateReportGroup",
          "codebuild:CreateReport",
          "codebuild:UpdateReport",
          "codebuild:BatchPutTestCases",
          "codebuild:BatchPutCodeCoverages"
        ],
        Resource : "arn:aws:codebuild:${var.region}:${data.aws_caller_identity.current.account_id}:report-group/*"
      },
      # CloudWatch
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "*"
      },
      # S3
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:GetBucketAcl",
          "s3:GetBucketLocation"
        ]
        Resource = [
          aws_s3_bucket.builds_bucket.arn,
          "${aws_s3_bucket.builds_bucket.arn}/*"
        ]
      },
    ]
  })
}
resource "aws_iam_role_policy_attachment" "build_policy_attach" {
  role       = aws_iam_role.build_role.name
  policy_arn = aws_iam_policy.build_policy.arn
}

# codebuild
resource "aws_codebuild_project" "bsl_build" {
  name           = "bsl-build-project"
  description    = "bsl_build_project"
  build_timeout  = "10"
  queued_timeout = "60"
  service_role   = aws_iam_role.build_role.arn

  artifacts {
    type     = "S3"
    location = aws_s3_bucket.builds_bucket.id
  }

  environment {
    type = "LINUX_CONTAINER"

    image                       = "aws/codebuild/standard:5.0"
    compute_type                = "BUILD_GENERAL1_SMALL"
    image_pull_credentials_type = "CODEBUILD"
  }

  logs_config {
    cloudwatch_logs {
      group_name  = "bsl-logs"
      stream_name = "bsl-build-logs"
    }
  }

  source {
    type            = "GITHUB"
    location        = "https://github.com/AJ2O/bytesizelinks.git"
    git_clone_depth = 1
  }
}

# --- CodeDeploy ---
# permissions
resource "aws_iam_role" "deploy_role" {
  name_prefix = "bsl-deploy-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "codedeploy.amazonaws.com"
        }
      },
    ]
  })
}
resource "aws_iam_policy" "deploy_launch_template_policy" {
  name_prefix = "bsl-deploy-launch-template-policy"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      # Needed for CodeDeploy to work with AutoScaling Groups properly
      {
        Effect = "Allow"
        Action = [
          "iam:PassRole",
          "ec2:CreateTags",
          "ec2:RunInstances"
        ],
        Resource : "*"
      },
    ]
  })
}
resource "aws_iam_role_policy_attachment" "deploy_role_attach" {
  role       = aws_iam_role.deploy_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSCodeDeployRole"
}
resource "aws_iam_role_policy_attachment" "deploy_launch_template_policy_attach" {
  role       = aws_iam_role.deploy_role.name
  policy_arn = aws_iam_policy.deploy_launch_template_policy.arn
}
# codedeploy (web client)
resource "aws_codedeploy_app" "bsl_deploy" {
  compute_platform = "Server"
  name             = "bsl-webclient"
}
/*
resource "aws_codedeploy_deployment_group" "bsl_dg" {
  app_name = aws_codedeploy_app.bsl_deploy.name
  deployment_group_name = "bsl-deployment-group"
  service_role_arn = aws_iam_role.deploy_role.arn

  deployment_style {
    deployment_type = "BLUE_GREEN"

  }

  load_balancer_info {
    target_group_info = 
  }
}*/

# --- CodePipeline ---
