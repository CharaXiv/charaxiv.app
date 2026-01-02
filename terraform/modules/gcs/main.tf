terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

resource "google_storage_bucket" "assets" {
  name     = "${var.project_id}-${var.environment}-assets"
  location = var.location

  # Prevent accidental deletion
  force_destroy = false

  # Uniform bucket-level access (recommended)
  uniform_bucket_level_access = true

  # CORS for direct browser uploads if needed
  cors {
    origin          = var.cors_origins
    method          = ["GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["Content-Type", "Content-Length", "Content-MD5"]
    max_age_seconds = 3600
  }

  # Lifecycle rules for cost optimization
  lifecycle_rule {
    condition {
      days_since_noncurrent_time = 365
    }
    action {
      type = "Delete"
    }
  }
}

# Service account for the application
resource "google_service_account" "app" {
  account_id   = "${var.project_name}-${var.environment}-storage"
  display_name = "Storage access for ${var.project_name} ${var.environment}"
}

# Grant storage access to the service account
resource "google_storage_bucket_iam_member" "app_access" {
  bucket = google_storage_bucket.assets.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.app.email}"
}

# Create a key for the service account (for dev environment)
resource "google_service_account_key" "app" {
  count              = var.create_sa_key ? 1 : 0
  service_account_id = google_service_account.app.name
}
