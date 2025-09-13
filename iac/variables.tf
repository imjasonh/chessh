
variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "us-central1"
}

variable "name" {
  description = "Name prefix for resources"
  type        = string
}

variable "deletion_protection" {
  description = "Enable deletion protection for GKE Autopilot cluster"
  type        = bool
  default     = false
}

variable "domain" {
  description = "Domain name to use for DNS records"
  type        = string
}

variable "dns_zone" {
  description = "DNS zone name to use for managing DNS records"
  type        = string
}
