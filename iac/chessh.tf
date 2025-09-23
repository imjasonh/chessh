# Container image build configuration
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
  repo   = "${var.region}-docker.pkg.dev/${var.project_id}/chessh/github.com/imjasonh/chessh"
  config = data.apko_config.config.config
}

resource "ko_build" "chessh" {
  importpath = "github.com/imjasonh/chessh"
  base_image = apko_build.base.image_ref
  depends_on = [google_artifact_registry_repository.artifact_repo]
}
