
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
