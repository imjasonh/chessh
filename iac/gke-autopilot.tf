# GKE Autopilot Cluster for SSH proxy
resource "google_container_cluster" "autopilot" {
  name                = "chessh-cluster"
  location            = var.region
  enable_autopilot    = true
  deletion_protection = var.deletion_protection
}

# Wait for cluster to be ready
resource "time_sleep" "wait_for_cluster" {
  depends_on      = [google_container_cluster.autopilot]
  create_duration = "30s"
}
