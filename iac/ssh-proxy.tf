

# Build the SSH proxy image
resource "ko_build" "ssh_proxy" {
  importpath = "github.com/imjasonh/ssh-proxy/cmd/ssh-proxy"
}

# Service account for the SSH proxy
resource "google_service_account" "ssh_proxy" {
  account_id   = "${var.name}-ssh-proxy"
  display_name = "SSH Proxy"
  description  = "Service account for SSH proxy that connects to Cloud Run"
}

# IAM member for SSH proxy is defined in cloudrun.tf

# Kubernetes service account for SSH proxy
resource "kubernetes_service_account" "ssh_proxy" {
  metadata {
    name      = "${var.name}-ssh-proxy"
    namespace = "default"
    annotations = {
      "iam.gke.io/gcp-service-account" = google_service_account.ssh_proxy.email
    }
  }

  depends_on = [
    google_container_cluster.autopilot,
    time_sleep.wait_for_cluster
  ]
}

# Workload Identity binding for SSH proxy
resource "google_service_account_iam_member" "ssh_proxy_workload_identity" {
  service_account_id = google_service_account.ssh_proxy.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.project_id}.svc.id.goog[default/${kubernetes_service_account.ssh_proxy.metadata[0].name}]"
}

# Allow SSH proxy service account to invoke Cloud Run
resource "google_cloud_run_service_iam_member" "ssh_proxy_invoker" {
  project  = var.project_id
  location = google_cloud_run_v2_service.chessh.location
  service  = google_cloud_run_v2_service.chessh.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.ssh_proxy.email}"
}

# Deploy SSH proxy as a Kubernetes Deployment
resource "kubernetes_deployment" "ssh_proxy" {
  metadata {
    name      = "${var.name}-ssh-proxy"
    namespace = "default"
    labels = {
      app = "${var.name}-ssh-proxy"
    }
  }

  # Ignore GKE Autopilot annotations that get added automatically
  lifecycle {
    ignore_changes = [
      metadata[0].annotations["autopilot.gke.io/resource-adjustment"],
      metadata[0].annotations["autopilot.gke.io/warden-version"],
    ]
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "${var.name}-ssh-proxy"
      }
    }

    template {
      metadata {
        labels = {
          app = "${var.name}-ssh-proxy"
        }
        annotations = {
          "iam.gke.io/gcp-service-account" = google_service_account.ssh_proxy.email
        }
      }

      spec {
        service_account_name = kubernetes_service_account.ssh_proxy.metadata[0].name

        # GKE Autopilot adds these automatically
        security_context {
          run_as_non_root = true
          seccomp_profile { type = "RuntimeDefault" }
        }

        toleration {
          effect   = "NoSchedule"
          key      = "kubernetes.io/arch"
          operator = "Equal"
          value    = "amd64"
        }

        container {
          image = ko_build.ssh_proxy.image_ref
          name  = "ssh-proxy"

          resources {
            requests = {
              cpu               = "100m"
              memory            = "128Mi"
              ephemeral-storage = "1Gi"
            }
            limits = {
              cpu               = "500m"
              memory            = "512Mi"
              ephemeral-storage = "1Gi"
            }
          }

          security_context {
            allow_privilege_escalation = false
            privileged                 = false
            read_only_root_filesystem  = true
            run_as_non_root            = true
            capabilities { drop = ["NET_RAW"] }
          }

          env {
            name  = "SSH_ADDR"
            value = ":22"
          }
          env {
            name  = "WEBSOCKET_URL"
            value = "${replace(google_cloud_run_v2_service.chessh.uri, "https://", "wss://")}/ssh"
          }

          port {
            container_port = 22
            name           = "ssh"
          }

          liveness_probe {
            tcp_socket {
              port = 22
            }
            initial_delay_seconds = 10
            period_seconds        = 30
          }

          readiness_probe {
            tcp_socket {
              port = 22
            }
            initial_delay_seconds = 5
            period_seconds        = 10
          }
        }
      }
    }
  }

  depends_on = [
    google_container_cluster.autopilot,
    time_sleep.wait_for_cluster,
    ko_build.ssh_proxy,
    google_cloud_run_service_iam_member.ssh_proxy_invoker,
    google_service_account_iam_member.ssh_proxy_workload_identity,
  ]
}

# Service for SSH proxy with NEG for Network Load Balancer
resource "kubernetes_service" "ssh_proxy" {
  metadata {
    name      = "${var.name}-ssh"
    namespace = "default"
    annotations = {
      # Enable NEG creation for this service
      "cloud.google.com/neg" = jsonencode({
        exposed_ports = {
          "22" = {}
        }
      })
    }
  }

  spec {
    type = "ClusterIP" # Changed from LoadBalancer to ClusterIP for NEG
    selector = {
      app = "${var.name}-ssh-proxy"
    }
    port {
      name        = "ssh"
      port        = 22
      target_port = 22
      protocol    = "TCP"
    }
  }

  # Ignore changes to annotations that GCP manages
  lifecycle {
    ignore_changes = [
      metadata[0].annotations["cloud.google.com/neg-status"],
    ]
  }

  depends_on = [
    google_container_cluster.autopilot,
    time_sleep.wait_for_cluster
  ]
}
