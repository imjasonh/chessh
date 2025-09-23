output "external_ip" {
  description = "External IP address for SSH access"
  value       = google_compute_global_address.chessh_lb.address
}

output "host_key_secret_name" {
  description = "Full Secret Manager secret name for SSH host key"
  value       = google_secret_manager_secret.ssh_host_private_key.secret_id
}

output "ssh_public_key" {
  description = "The SSH public key"
  value       = tls_private_key.ssh_host_key.public_key_openssh
  sensitive   = false
}

output "vm_service_account_email" {
  description = "Service account email for chessh VM instances"
  value       = google_service_account.chessh_vm.email
}

output "chessh_image_ref" {
  description = "Container image reference for chessh"
  value       = ko_build.chessh.image_ref
}

