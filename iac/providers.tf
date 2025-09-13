terraform {
  required_version = ">= 1.0"

  required_providers {
    google     = { source = "hashicorp/google" }
    kubernetes = { source = "hashicorp/kubernetes" }
    ko         = { source = "ko-build/ko" }
    time       = { source = "hashicorp/time" }
    random     = { source = "hashicorp/random" }
    apko       = { source = "chainguard-dev/apko" }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

data "google_client_config" "default" {}

# Use the GKE Autopilot cluster created in gke-autopilot.tf
# The try() function prevents errors when the cluster doesn't exist yet
provider "kubernetes" {
  host                   = try("https://${google_container_cluster.autopilot.endpoint}", "")
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = try(base64decode(google_container_cluster.autopilot.master_auth[0].cluster_ca_certificate), "")
}

provider "ko" {
  repo = "${var.region}-docker.pkg.dev/${var.project_id}/chessh"
}
