terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }

  backend "gcs" {}
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
}

module "r2" {
  source = "../../modules/r2"

  cloudflare_account_id = var.cloudflare_account_id
  project_name          = "charaxiv"
  environment           = "prd"
  r2_location           = "APAC" # Close to Japan users
}

module "cloudrun" {
  source = "../../modules/cloudrun"

  project_name = "charaxiv"
  environment  = "prd"
  region       = var.gcp_region
  image        = var.app_image

  # Production settings
  min_instances = 0  # Scale to zero to save costs
  max_instances = 10
  cpu           = "1"
  memory        = "512Mi"

  # R2 credentials
  r2_account_id         = var.cloudflare_account_id
  r2_bucket_name        = module.r2.bucket_name
  r2_access_key_id      = var.r2_access_key_id
  r2_secret_access_key  = var.r2_secret_access_key
}

# Outputs
output "r2_bucket_name" {
  value = module.r2.bucket_name
}

output "r2_endpoint" {
  value = module.r2.endpoint
}

output "cloudrun_url" {
  value = module.cloudrun.service_url
}

output "artifact_registry" {
  value = module.cloudrun.artifact_registry
}
