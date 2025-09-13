output "chessh_domain" {
  description = "Full domain for chessh service"
  value       = "chessh.${var.domain}"
}

output "dns_zone_used" {
  description = "DNS zone being used"
  value       = data.google_dns_managed_zone.existing.name
}

output "ssh_service_ip" {
  description = "External IP address for SSH access"
  value       = google_compute_global_address.chessh.address
}

output "cloud_run_url" {
  description = "URL of the Cloud Run service"
  value       = google_cloud_run_v2_service.chessh.uri
}

output "host_key_secret_name" {
  description = "Full Secret Manager secret name for SSH host key"
  value       = google_secret_manager_secret.ssh_host_private_key.secret_id
}

output "service_account_email" {
  description = "Service account email for chessh"
  value       = google_service_account.chessh.email
}

output "ssh_proxy_service_account_email" {
  description = "Service account email for SSH proxy"
  value       = google_service_account.chessh.email
}

output "chessh_image_ref" {
  description = "Container image reference for chessh"
  value       = ko_build.chessh.image_ref
}

output "ssh_proxy_image_ref" {
  description = "Container image reference for SSH proxy"
  value       = ko_build.ssh_proxy.image_ref
}

output "ssh_endpoint" {
  description = "SSH endpoint for Git operations"
  value       = google_compute_global_address.chessh.address
}

output "websocket_url" {
  description = "WebSocket URL used by SSH proxy (internal)"
  value       = "${replace(google_cloud_run_v2_service.chessh.uri, "https://", "wss://")}/ssh"
}

output "global_ip_address" {
  description = "Global IP address for all services"
  value       = google_compute_global_address.chessh.address
}
