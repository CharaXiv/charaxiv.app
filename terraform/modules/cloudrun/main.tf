terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

# Enable required APIs
resource "google_project_service" "run" {
  service            = "run.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "artifactregistry" {
  service            = "artifactregistry.googleapis.com"
  disable_on_destroy = false
}

# Artifact Registry for container images
resource "google_artifact_registry_repository" "app" {
  location      = var.region
  repository_id = "${var.project_name}-${var.environment}"
  format        = "DOCKER"

  depends_on = [google_project_service.artifactregistry]
}

# Service account for Cloud Run
resource "google_service_account" "cloudrun" {
  account_id   = "${var.project_name}-${var.environment}-run"
  display_name = "Cloud Run service account for ${var.project_name} ${var.environment}"
}

# Grant Cloud Run service account access to GCS bucket
resource "google_storage_bucket_iam_member" "cloudrun_gcs" {
  bucket = var.gcs_bucket_name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.cloudrun.email}"
}

# Cloud Run service
resource "google_cloud_run_v2_service" "app" {
  name     = "${var.project_name}-${var.environment}"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.cloudrun.email

    scaling {
      min_instance_count = var.min_instances
      max_instance_count = var.max_instances
    }

    containers {
      image = var.image != "" ? var.image : "us-docker.pkg.dev/cloudrun/container/hello"

      ports {
        container_port = 8000
      }

      resources {
        limits = {
          cpu    = var.cpu
          memory = var.memory
        }
        cpu_idle          = true  # Scale to zero
        startup_cpu_boost = true
      }

      env {
        name  = "ENV"
        value = var.environment
      }

      env {
        name  = "GCS_BUCKET"
        value = var.gcs_bucket_name
      }

      startup_probe {
        http_get {
          path = "/health"
        }
        initial_delay_seconds = 0
        period_seconds        = 10
        failure_threshold     = 3
      }

      liveness_probe {
        http_get {
          path = "/health"
        }
        period_seconds = 30
      }
    }
  }

  depends_on = [
    google_project_service.run,
  ]
}

# Allow unauthenticated access
resource "google_cloud_run_v2_service_iam_member" "public" {
  location = google_cloud_run_v2_service.app.location
  name     = google_cloud_run_v2_service.app.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
