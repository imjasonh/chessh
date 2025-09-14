# Generate SSH key pair
resource "tls_private_key" "ssh_host_key" {
  algorithm = "ED25519"
}

# Store the private key in Secret Manager
resource "google_secret_manager_secret" "ssh_host_private_key" {
  secret_id = "${var.name}-ssh-proxy-host-private-key"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "ssh_host_private_key_version" {
  secret      = google_secret_manager_secret.ssh_host_private_key.id
  secret_data = tls_private_key.ssh_host_key.private_key_openssh
}

# Store the public key in Secret Manager (for reference/validation)
resource "google_secret_manager_secret" "ssh_host_public_key" {
  secret_id = "${var.name}-ssh-proxy-host-public-key"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "ssh_host_public_key_version" {
  secret      = google_secret_manager_secret.ssh_host_public_key.id
  secret_data = tls_private_key.ssh_host_key.public_key_openssh
}

# Output the secret name for use in environment variables
output "ssh_private_key_secret_name" {
  description = "The name of the secret containing the SSH private key"
  value       = google_secret_manager_secret.ssh_host_private_key.id
}

output "ssh_public_key_secret_name" {
  description = "The name of the secret containing the SSH public key"
  value       = google_secret_manager_secret.ssh_host_public_key.id
}

# Grant chessh service account access to the SSH private key secret
resource "google_secret_manager_secret_iam_member" "chessh_secret_access" {
  project   = var.project_id
  secret_id = google_secret_manager_secret.ssh_host_private_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.chessh.email}"
}
