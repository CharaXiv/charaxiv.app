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

      # R2 configuration
      env {
        name  = "R2_ACCOUNT_ID"
        value = var.r2_account_id
      }

      env {
        name  = "R2_BUCKET"
        value = var.r2_bucket_name
      }

      env {
        name = "R2_ACCESS_KEY_ID"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.r2_access_key.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "R2_SECRET_ACCESS_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.r2_secret_key.secret_id
            version = "latest"
          }
        }
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
    google_secret_manager_secret_version.r2_access_key,
    google_secret_manager_secret_version.r2_secret_key,
  ]
}

# Allow unauthenticated access
resource "google_cloud_run_v2_service_iam_member" "public" {
  location = google_cloud_run_v2_service.app.location
  name     = google_cloud_run_v2_service.app.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# Secret Manager for R2 credentials
resource "google_project_service" "secretmanager" {
  service            = "secretmanager.googleapis.com"
  disable_on_destroy = false
}

resource "google_secret_manager_secret" "r2_access_key" {
  secret_id = "${var.project_name}-${var.environment}-r2-access-key"

  replication {
    auto {}
  }

  depends_on = [google_project_service.secretmanager]
}

resource "google_secret_manager_secret_version" "r2_access_key" {
  secret      = google_secret_manager_secret.r2_access_key.id
  secret_data = var.r2_access_key_id
}

resource "google_secret_manager_secret" "r2_secret_key" {
  secret_id = "${var.project_name}-${var.environment}-r2-secret-key"

  replication {
    auto {}
  }

  depends_on = [google_project_service.secretmanager]
}

resource "google_secret_manager_secret_version" "r2_secret_key" {
  secret      = google_secret_manager_secret.r2_secret_key.id
  secret_data = var.r2_secret_access_key
}

# Grant Cloud Run access to secrets
resource "google_secret_manager_secret_iam_member" "r2_access_key" {
  secret_id = google_secret_manager_secret.r2_access_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloudrun.email}"
}

resource "google_secret_manager_secret_iam_member" "r2_secret_key" {
  secret_id = google_secret_manager_secret.r2_secret_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloudrun.email}"
}
