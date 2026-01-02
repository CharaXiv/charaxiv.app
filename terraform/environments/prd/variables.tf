# Cloudflare
variable "cloudflare_api_token" {
  description = "Cloudflare API token with R2 admin permissions"
  type        = string
  sensitive   = true
}

variable "cloudflare_account_id" {
  description = "Cloudflare account ID"
  type        = string
}

# GCP
variable "gcp_project_id" {
  description = "GCP project ID"
  type        = string
}

variable "gcp_region" {
  description = "GCP region for Cloud Run"
  type        = string
  default     = "asia-northeast1" # Tokyo
}

# App
variable "app_image" {
  description = "Container image to deploy"
  type        = string
  default     = "" # Empty uses placeholder
}

# R2 credentials (for Cloud Run secrets)
variable "r2_access_key_id" {
  description = "R2 access key ID (will be stored in Secret Manager)"
  type        = string
  sensitive   = true
}

variable "r2_secret_access_key" {
  description = "R2 secret access key (will be stored in Secret Manager)"
  type        = string
  sensitive   = true
}
