# DNS onfiguration for chessh.${var.domain}

# Use existing DNS zone specified by dns_zone variable
data "google_dns_managed_zone" "existing" { name = var.dns_zone }

# DNS A record for chessh.${var.domain}
resource "google_dns_record_set" "chessh" {
  name         = "chessh.${var.domain}."
  type         = "A"
  ttl          = 300
  managed_zone = data.google_dns_managed_zone.existing.name
  rrdatas      = [data.kubernetes_service.ssh_proxy.status.0.load_balancer.0.ingress.0.ip]
}
