# Migration: Cloud Run → GCE VM with TCP Load Balancer

## Overview
This migration moves ChessSH from Cloud Run + GKE SSH proxy to a direct GCE VM deployment with TCP load balancing.

## Architecture Changes

### Before (Cloud Run + SSH Proxy)
- Cloud Run service hosting the chess app with WebSocket endpoint
- GKE Autopilot cluster with SSH proxy deployment
- Kubernetes LoadBalancer service for SSH traffic
- SSH proxy converts WebSocket ↔ SSH traffic

### After (GCE VM + TCP Load Balancer)  
- GCE VM instances running chess containers directly
- TCP Load Balancer with direct SSH access (port 22)
- HTTP health checks on port 8080
- No proxy layer - direct SSH connections

## Benefits
- **Simpler architecture**: No WebSocket proxy layer
- **Better performance**: Direct TCP connections
- **Native SSH features**: Full SSH compatibility
- **Cost optimization**: Single layer instead of Cloud Run + GKE

## New Resources
- `gce-vm.tf`: VM instances, instance groups, health checks
- `tcp-load-balancer.tf`: Global TCP load balancer
- Updated `main.go`: Simple `/health` endpoint instead of WebSocket proxy
- Updated `dns.tf`: Points to load balancer IP

## Migration Steps
1. Deploy new infrastructure alongside existing
2. Test health checks and SSH connectivity
3. Update DNS to point to new load balancer
4. Clean up old Cloud Run and GKE resources

## Connection Method
- **Before**: `ssh chessh.domain.com -p 22` (via proxy)
- **After**: `ssh chessh.domain.com -p 22` (direct)

The user experience remains the same, but the backend is significantly simplified.