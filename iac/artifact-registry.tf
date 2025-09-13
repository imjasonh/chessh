# Enable Artifact Registry API
resource "google_project_service" "artifact_registry" {
  service = "artifactregistry.googleapis.com"
}

# Create Artifact Registry repository for Docker images
resource "google_artifact_registry_repository" "artifact_repo" {
  location      = var.region
  repository_id = var.name
  description   = "Docker repository for ${var.name} service"
  format        = "DOCKER"

  depends_on = [google_project_service.artifact_registry]
}
