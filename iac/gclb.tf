# TCP Proxy Load Balancer configuration
# Uses a single global IP address with TCP proxy for different ports:
# - Port 22 -> GKE SSH proxy pods (via NEG)

# Reserve a global static IP address (TCP Proxy requires global IP)
resource "google_compute_global_address" "chessh" {
  name         = "chessh-global-ip"
  description  = "Global IP for chessh hosting service"
  address_type = "EXTERNAL"
  ip_version   = "IPV4"
}

# Health check for SSH proxy
resource "google_compute_health_check" "ssh_proxy" {
  name                = "chessh-ssh-health"
  check_interval_sec  = 10
  timeout_sec         = 5
  healthy_threshold   = 2
  unhealthy_threshold = 3

  tcp_health_check { port = 22 }
}

# Parse NEG name and zones from managed service annotation with null safety
locals {
  service_annotations = try(kubernetes_service.ssh_proxy.metadata[0].annotations, {})
  neg_status_raw      = try(local.service_annotations["cloud.google.com/neg-status"], null)
  neg_status          = local.neg_status_raw != null ? jsondecode(local.neg_status_raw) : null
  neg_name            = local.neg_status != null ? local.neg_status.network_endpoint_groups["22"] : null
  neg_zones           = local.neg_status != null ? local.neg_status.zones : []
}

# Data source to discover the NEGs created by Kubernetes using the dynamic name
data "google_compute_network_endpoint_group" "ssh_negs" {
  for_each = toset(local.neg_zones)

  name = local.neg_name
  zone = each.key

  depends_on = [
    kubernetes_service.ssh_proxy,
    time_sleep.wait_for_neg_creation
  ]
}

# Wait for NEG creation after service deployment
resource "time_sleep" "wait_for_neg_creation" {
  depends_on = [
    kubernetes_service.ssh_proxy,
    kubernetes_deployment.ssh_proxy
  ]

  create_duration = "60s"
}

# Backend service for SSH using TCP Proxy (global)
resource "google_compute_backend_service" "ssh_tcp_proxy" {
  name                  = "${var.name}-ssh-tcp-backend"
  protocol              = "TCP"
  port_name             = "ssh"
  timeout_sec           = 30
  load_balancing_scheme = "EXTERNAL"
  health_checks         = [google_compute_health_check.ssh_proxy.id]

  # Add backends for each zone's NEG
  dynamic "backend" {
    for_each = data.google_compute_network_endpoint_group.ssh_negs
    content {
      group           = backend.value.id
      balancing_mode  = "CONNECTION"
      max_connections = 1000
      capacity_scaler = 1.0
    }
  }

  connection_draining_timeout_sec = 30

  # Enable access logging for TCP proxy
  log_config {
    enable      = true
    sample_rate = 1.0
  }
}

# TCP proxy for SSH traffic
resource "google_compute_target_tcp_proxy" "ssh" {
  name            = "${var.name}-ssh-tcp-proxy"
  backend_service = google_compute_backend_service.ssh_tcp_proxy.id
}

# Global forwarding rule for SSH (port 22)
resource "google_compute_global_forwarding_rule" "ssh" {
  name                  = "${var.name}-ssh-forwarding"
  ip_protocol           = "TCP"
  port_range            = "22"
  load_balancing_scheme = "EXTERNAL"
  target                = google_compute_target_tcp_proxy.ssh.id
  ip_address            = google_compute_global_address.chessh.address
}

# Data source for available zones
data "google_compute_zones" "available" {
  region = var.region
}
