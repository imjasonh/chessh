terraform {
  required_version = ">= 1.0"

  required_providers {
    google = { source = "hashicorp/google" }
    ko     = { source = "ko-build/ko" }
    time   = { source = "hashicorp/time" }
    random = { source = "hashicorp/random" }
    apko   = { source = "chainguard-dev/apko" }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

provider "ko" {
  repo = "${var.region}-docker.pkg.dev/${var.project_id}/chessh"
}
