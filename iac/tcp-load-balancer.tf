# Global static IP address for the load balancer
resource "google_compute_global_address" "chessh_lb" {
  name = "${var.name}-lb-ip"
}

# Backend service for the instance group
resource "google_compute_backend_service" "chessh" {
  name        = "${var.name}-backend"
  protocol    = "TCP"
  port_name   = "ssh"
  timeout_sec = 5 * 60 # 5 minutes timeout for ~long-lived SSH sessions

  health_checks = [google_compute_health_check.chessh.id]

  backend {
    group                        = google_compute_region_instance_group_manager.chessh.instance_group
    balancing_mode               = "CONNECTION"
    max_connections_per_instance = 1000
  }

  # Session affinity to keep existing connections on same backend
  session_affinity = "CLIENT_IP"

  depends_on = [google_compute_region_instance_group_manager.chessh]
}

# Target TCP proxy
resource "google_compute_target_tcp_proxy" "chessh" {
  name            = "${var.name}-tcp-proxy"
  backend_service = google_compute_backend_service.chessh.id
}

# Global forwarding rule for TCP traffic
resource "google_compute_global_forwarding_rule" "chessh_tcp" {
  name       = "${var.name}-tcp-forwarding"
  target     = google_compute_target_tcp_proxy.chessh.id
  port_range = "22"
  ip_address = google_compute_global_address.chessh_lb.address
}

# Output the load balancer IP
output "load_balancer_ip" {
  description = "IP address of the TCP load balancer"
  value       = google_compute_global_address.chessh_lb.address
}

output "ssh_connection_command" {
  description = "Command to connect via SSH"
  value       = "ssh ${google_compute_global_address.chessh_lb.address}"
}
