# Enable required APIs
resource "google_project_service" "compute" {
  service = "compute.googleapis.com"
}

resource "google_project_service" "container" {
  service = "container.googleapis.com"
}

resource "google_project_service" "secretmanager" {
  service = "secretmanager.googleapis.com"
}

# Service account for the GCE instances
resource "google_service_account" "chessh_vm" {
  account_id   = "${var.name}-vm"
  display_name = "ChessH VM Instance"
  description  = "Service account for ChessH GCE VM instances"
}

# Allow the VM service account to pull from Artifact Registry
resource "google_project_iam_member" "chessh_vm_artifact_registry" {
  project = var.project_id
  role    = "roles/artifactregistry.reader"
  member  = "serviceAccount:${google_service_account.chessh_vm.email}"
}

# Allow the VM service account to access secrets (for SSH host key)
resource "google_secret_manager_secret_iam_member" "ssh_host_key_accessor_vm" {
  secret_id = google_secret_manager_secret.ssh_host_private_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.chessh_vm.email}"
}

# Firewall rule to allow chess SSH traffic (port 2222)
resource "google_compute_firewall" "chessh_ssh" {
  name    = "${var.name}-ssh"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["2222"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["chessh-ssh"]
}

# Firewall rule to allow health check traffic (port 8080)
resource "google_compute_firewall" "chessh_health" {
  name    = "${var.name}-health"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["8080"]
  }

  # Google Cloud health check IP ranges
  source_ranges = ["35.191.0.0/16", "130.211.0.0/22"]
  target_tags   = ["chessh-health"]
}

# Instance template for ChessH
resource "google_compute_instance_template" "chessh" {
  name_prefix = "${var.name}-template-"
  description = "Template for ChessH game server instances"

  machine_type = "e2-micro"
  region       = var.region

  disk {
    source_image = "projects/cos-cloud/global/images/family/cos-stable"
    auto_delete  = true
    boot         = true
    disk_type    = "pd-standard"
    disk_size_gb = 10
  }

  network_interface {
    network = "default"
    access_config {
      # Ephemeral external IP
    }
  }

  service_account {
    email  = google_service_account.chessh_vm.email
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }

  tags = ["chessh-ssh", "chessh-health"]

  metadata = {
    "gce-container-declaration" = jsonencode({
      spec = {
        containers = [{
          name  = var.name
          image = ko_build.chessh.image_ref
          env = [
            {
              name  = "PORT"
              value = "8080"
            },
            {
              name  = "SSH_HOST_KEY_SECRET"
              value = google_secret_manager_secret_version.ssh_host_private_key_version.id
            },
            {
              name  = "LOG_LEVEL"
              value = "info"
            }
          ]
          args = ["-port", "2222"]
        }]
        restartPolicy = "Always"
      }
    })
  }

  depends_on = [
    google_project_service.compute,
    google_project_service.container,
    ko_build.chessh
  ]
  
  lifecycle {
    create_before_destroy = true
  }
}

# Instance group manager
resource "google_compute_region_instance_group_manager" "chessh" {
  name   = "${var.name}-group"
  region = var.region

  base_instance_name = "${var.name}-instance"
  target_size        = 1

  version {
    instance_template = google_compute_instance_template.chessh.id
  }

  # Rolling update policy for graceful deployments
  update_policy {
    type                         = "PROACTIVE"
    instance_redistribution_type = "PROACTIVE"
    minimal_action               = "REPLACE"
    most_disruptive_allowed_action = "REPLACE"
    max_surge_fixed             = 3  # Must be >= number of zones in region
    max_unavailable_fixed       = 0
  }

  named_port {
    name = "ssh"
    port = 2222
  }

  named_port {
    name = "health"
    port = 8080
  }

  auto_healing_policies {
    health_check      = google_compute_health_check.chessh.id
    initial_delay_sec = 60
  }

  depends_on = [google_compute_instance_template.chessh]
  
  lifecycle {
    create_before_destroy = true
  }
}

# Health check for the instances
resource "google_compute_health_check" "chessh" {
  name = "${var.name}-health-check"

  timeout_sec         = 5
  check_interval_sec  = 10
  healthy_threshold   = 2
  unhealthy_threshold = 3

  http_health_check {
    port         = 8080
    request_path = "/health"
  }
}
