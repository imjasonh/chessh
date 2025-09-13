resource "google_project_service" "cloudrun" {
  service = "run.googleapis.com"
}

data "apko_config" "config" {
  config_contents = jsonencode({
    contents = {
      repositories = ["https://apk.cgr.dev/chainguard"]
      packages     = ["ca-certificates-bundle", "grype"]
    }
    archs = ["x86_64"]
  })
}

resource "apko_build" "base" {
  repo = "${var.region}-docker.pkg.dev/${var.project_id}/chessh/github.com/imjasonh/chessh"
  config = data.apko_config.config.config
}

resource "ko_build" "chessh" {
  importpath = "github.com/imjasonh/chessh"
  base_image = apko_build.base.image_ref
  depends_on = [google_artifact_registry_repository.artifact_repo]
}

# Service account for the SSH proxy
resource "google_service_account" "chessh" {
  account_id   = "${var.name}-chessh"
  display_name = "chessh"
  description  = "Service account for chessh Cloud Run service"
}

resource "google_cloud_run_v2_service" "chessh" {
  name     = var.name
  location = var.region

  deletion_protection = var.deletion_protection

  template {
    service_account       = google_service_account.chessh.email
    execution_environment = "EXECUTION_ENVIRONMENT_GEN2"
    timeout               = "${60 * 60}s" # 1 hour, effectively maximum session duration

    containers {
      image   = ko_build.chessh.image_ref

      # Cloud Run automatically sets PORT=8080
      env {
        name  = "LOG_LEVEL"
        value = "info"
      }

      ports {
        container_port = 8080
        name           = "http1"
      }

      resources {
        startup_cpu_boost = true

        limits = {
          cpu    = "4"
          memory = "4Gi"
        }
      }

      # startup_probe {
      #   initial_delay_seconds = 10
      #   timeout_seconds       = 3
      #   period_seconds        = 10
      #   failure_threshold     = 3

      #   http_get {
      #     path = "/health"
      #     port = 8080
      #   }
      # }
      # liveness_probe {
      #   initial_delay_seconds = 60
      #   timeout_seconds       = 3
      #   period_seconds        = 30
      #   failure_threshold     = 3

      #   http_get {
      #     path = "/health"
      #     port = 8080
      #   }
      # }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }

  depends_on = [google_project_service.cloudrun]
}

# Allow unauthenticated access to Cloud Run service (for WebSocket from proxy)
resource "google_cloud_run_service_iam_member" "chessh_invoker" {
  project  = var.project_id
  location = google_cloud_run_v2_service.chessh.location
  service  = google_cloud_run_v2_service.chessh.name
  role     = "roles/run.invoker"
  member   = "allUsers" # In production, restrict to proxy service account
}
