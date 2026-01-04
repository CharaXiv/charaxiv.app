terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }

  backend "gcs" {}
}

provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
}

module "gcs" {
  source = "../../modules/gcs"

  project_id    = var.gcp_project_id
  project_name  = "charaxiv"
  environment   = "prd"
  location      = "ASIA-NORTHEAST1" # Tokyo
  cors_origins  = ["https://charaxiv.app", "https://www.charaxiv.app"]
  create_sa_key = false # Cloud Run uses workload identity
}

module "cloudrun" {
  source = "../../modules/cloudrun"

  project_id   = var.gcp_project_id
  project_name = "charaxiv"
  environment  = "prd"
  region       = var.gcp_region
  image        = var.app_image

  # Production settings
  min_instances = 0  # Scale to zero to save costs
  max_instances = 10
  cpu           = "1"
  memory        = "512Mi"

  # GCS bucket
  gcs_bucket_name = module.gcs.bucket_name
}

# Outputs
output "bucket_name" {
  value = module.gcs.bucket_name
}

output "bucket_url" {
  value = module.gcs.bucket_url
}

output "service_account_email" {
  value = module.gcs.service_account_email
}

output "cloudrun_url" {
  value = module.cloudrun.service_url
}

output "artifact_registry" {
  value = module.cloudrun.artifact_registry
}
