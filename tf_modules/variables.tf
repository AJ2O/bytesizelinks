# --- General ---
variable "region" {
  description = "AWS region to deploy this project to."
  type        = string
  default     = "us-east-1"
}

variable "env_type" {
  description = "The type of environment this deployment is. Either one of dev/prod"
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "test", "prod"], var.env_type)
    error_message = "The env_type value must be one of \"dev\", \"test\", or \"prod\"."
  }
}

variable "domain_name" {
  description = "The domain name given to the website."
  type        = string
  default     = "bytesize.link"
}