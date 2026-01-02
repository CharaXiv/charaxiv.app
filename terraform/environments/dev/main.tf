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
  environment   = "dev"
  location      = "US-WEST1" # Close to exe.dev
  cors_origins  = ["https://charaxiv.exe.xyz:8080", "http://localhost:8000"]
  create_sa_key = true # Create key for local dev
}

output "bucket_name" {
  value = module.gcs.bucket_name
}

output "bucket_url" {
  value = module.gcs.bucket_url
}

output "service_account_email" {
  value = module.gcs.service_account_email
}

output "service_account_key" {
  value     = module.gcs.service_account_key
  sensitive = true
}
