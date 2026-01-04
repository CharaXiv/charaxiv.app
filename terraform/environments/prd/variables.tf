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
