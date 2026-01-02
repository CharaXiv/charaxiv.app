variable "project_name" {
  description = "Project name prefix for resources"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, stg, prd)"
  type        = string
}

variable "region" {
  description = "GCP region for Cloud Run"
  type        = string
}

variable "image" {
  description = "Container image to deploy (leave empty for placeholder)"
  type        = string
  default     = ""
}

variable "min_instances" {
  description = "Minimum number of instances (0 for scale to zero)"
  type        = number
  default     = 0
}

variable "max_instances" {
  description = "Maximum number of instances"
  type        = number
  default     = 10
}

variable "cpu" {
  description = "CPU limit"
  type        = string
  default     = "1"
}

variable "memory" {
  description = "Memory limit"
  type        = string
  default     = "512Mi"
}

# R2 configuration
variable "r2_account_id" {
  description = "Cloudflare account ID for R2"
  type        = string
}

variable "r2_bucket_name" {
  description = "R2 bucket name"
  type        = string
}

variable "r2_access_key_id" {
  description = "R2 access key ID"
  type        = string
  sensitive   = true
}

variable "r2_secret_access_key" {
  description = "R2 secret access key"
  type        = string
  sensitive   = true
}
