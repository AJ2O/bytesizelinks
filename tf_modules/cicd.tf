# --- CodeBuild ---
# artifact storage
resource "aws_s3_bucket" "builds_bucket" {
  bucket_prefix = "bsl-builds-bucket"
  acl           = "private"
  force_destroy = true
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
          "s3:GetBucketAcl",
          "s3:GetBucketLocation",
          "s3:GetObject",
          "s3:GetObjectVersion",
          "s3:PutObject"
        ]
        Resource = [
          aws_s3_bucket.builds_bucket.arn,
          "${aws_s3_bucket.builds_bucket.arn}/*",
          aws_s3_bucket.pipeline_bucket.arn,
          "${aws_s3_bucket.pipeline_bucket.arn}/*"
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
    location        = format("https://github.com/%s.git", var.github_repo)
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
resource "aws_codedeploy_deployment_group" "bsl_dg" {
  app_name               = aws_codedeploy_app.bsl_deploy.name
  deployment_group_name  = "bsl-deployment-group"
  deployment_config_name = "CodeDeployDefault.AllAtOnce"
  service_role_arn       = aws_iam_role.deploy_role.arn

  auto_rollback_configuration {
    enabled = true
    events  = ["DEPLOYMENT_FAILURE"]
  }

  autoscaling_groups = [aws_autoscaling_group.bsl_asg.name]

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout = "CONTINUE_DEPLOYMENT"
    }

    green_fleet_provisioning_option {
      action = "COPY_AUTO_SCALING_GROUP"
    }

    terminate_blue_instances_on_deployment_success {
      action                           = "TERMINATE"
      termination_wait_time_in_minutes = 60
    }
  }

  deployment_style {
    deployment_type   = "BLUE_GREEN"
    deployment_option = "WITH_TRAFFIC_CONTROL"
  }

  load_balancer_info {
    target_group_info {
      name = aws_lb_target_group.bsl_tg.name
    }
  }

  lifecycle {
    ignore_changes = [autoscaling_groups]
  }
}

# --- CodePipeline ---
# artifact storage
resource "aws_s3_bucket" "pipeline_bucket" {
  bucket_prefix = "bsl-pipeline-bucket"
  acl           = "private"
  force_destroy = true
}
# codestar connection to GitHub -> must be confirmed in console
resource "aws_codestarconnections_connection" "github" {
  name          = "github-connection"
  provider_type = "GitHub"
}

# permissions
resource "aws_iam_role" "pipeline_role" {
  name_prefix = "bsl-pipeline-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "codepipeline.amazonaws.com"
        }
      },
    ]
  })
}
resource "aws_iam_policy" "pipeline_policy" {
  name_prefix = "bsl-pipeline-policy"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      # CodeBuild
      {
        Effect = "Allow"
        Action = [
          "codebuild:BatchGetBuilds",
          "codebuild:StartBuild"
        ]
        Resource = "*"
      },

      # CodeDeploy
      {
        Effect = "Allow"
        Action = [
          "codedeploy:CreateDeployment",
          "codedeploy:GetApplication",
          "codedeploy:GetApplicationRevision",
          "codedeploy:GetDeployment",
          "codedeploy:GetDeploymentConfig",
          "codedeploy:RegisterApplicationRevision"
        ]
        Resource = "*"
      },

      # CodeStar
      {
        Effect = "Allow"
        Action = [
          "codestar-connections:UseConnection"
        ]
        Resource = aws_codestarconnections_connection.github.arn
      },

      # S3
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:GetObjectVersion",
          "s3:GetBucketVersioning",
          "s3:PutObjectAcl",
          "s3:PutObject"
        ]
        Resource = [
          aws_s3_bucket.pipeline_bucket.arn,
          "${aws_s3_bucket.pipeline_bucket.arn}/*"
        ]
      },
    ]
  })
}
resource "aws_iam_role_policy_attachment" "pipeline_policy_attach" {
  role       = aws_iam_role.pipeline_role.name
  policy_arn = aws_iam_policy.pipeline_policy.arn
}

# main pipeline
resource "aws_codepipeline" "web_client" {
  name     = "bsl-webclient-pipeline"
  role_arn = aws_iam_role.pipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.pipeline_bucket.id
    type     = "S3"
  }

  # scm - https://docs.aws.amazon.com/codepipeline/latest/userguide/action-reference-CodestarConnectionSource.html
  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["SourceArtifact"]

      configuration = {
        ConnectionArn        = aws_codestarconnections_connection.github.arn
        FullRepositoryId     = var.github_repo
        BranchName           = "main"
        DetectChanges        = true
        OutputArtifactFormat = "CODE_ZIP"
      }
    }
  }

  # build - https://docs.aws.amazon.com/codepipeline/latest/userguide/action-reference-CodeBuild.html
  stage {
    name = "Build"

    action {
      name             = "Build"
      category         = "Build"
      owner            = "AWS"
      provider         = "CodeBuild"
      version          = "1"
      input_artifacts  = ["SourceArtifact"]
      output_artifacts = ["BuildArtifact"]

      configuration = {
        ProjectName = aws_codebuild_project.bsl_build.name
      }
    }
  }

  # deploy - https://docs.aws.amazon.com/codepipeline/latest/userguide/action-reference-CodeDeploy.html
  stage {
    name = "Deploy"

    action {
      name            = "Deploy"
      category        = "Deploy"
      owner           = "AWS"
      provider        = "CodeDeploy"
      version         = "1"
      input_artifacts = ["BuildArtifact"]

      configuration = {
        ApplicationName     = aws_codedeploy_app.bsl_deploy.name
        DeploymentGroupName = aws_codedeploy_deployment_group.bsl_dg.deployment_group_name
      }
    }
  }
}