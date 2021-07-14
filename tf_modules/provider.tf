terraform {
  backend "s3" {
    bucket  = "bsl-terraform-remote-backend"
    encrypt = true
    key     = "terraform.tfstate"
    region  = "us-east-1"
  }
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

provider "aws" {
  region = var.region
  default_tags {
    tags = {
      Terraform   = "true"
      Project     = "ByteSize.Links"
      Environment = var.env_type
    }
  }
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}