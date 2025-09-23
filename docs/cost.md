# Cost Analysis: GCE VM + TCP Load Balancer Architecture

Based on the current GCE VM with TCP Load Balancer architecture, estimated costs:

## Minimum Monthly Cost (No Traffic)

**Compute Engine VM:**
- 1 × e2-micro instance (1 vCPU, 1GB RAM): ~$6-8/month
- Persistent disk (10GB standard): ~$0.40/month

**TCP Load Balancer:**
- Global Network Load Balancer with forwarding rules: ~$18-22/month
- Health check service: ~$0.50/month

**Other Resources (minimal cost):**
- DNS zone and records: ~$0.50/month  
- Secret Manager (2 secrets): ~$0.06/month
- Container Registry storage: ~$0.10/month

**Total estimated minimum cost: ~$25-31/month**

The TCP Load Balancer remains the largest cost component (~70% of total). However, this architecture is more cost-effective than the previous Cloud Run + GKE setup by eliminating the GKE Autopilot cluster overhead.

## Usage-Based Costs

**VM Instance Scaling:**
- Base e2-micro handles ~50-100 concurrent SSH sessions efficiently
- For higher load, can scale up to larger machine types:
  - e2-small (2 vCPU, 2GB): ~$12-16/month
  - e2-medium (2 vCPU, 4GB): ~$24-32/month

**Per-Connection Costs:**
Since the VM runs 24/7, marginal cost per additional SSH connection is minimal until CPU/memory limits are reached.

**Example Scaling Scenarios:**
- 50 concurrent users: Base e2-micro (~$25-31/month total)
- 150 concurrent users: e2-small (~$30-38/month total)  
- 500+ concurrent users: e2-medium + load balancer (~$42-54/month total)

## Cost Advantages vs Previous Architecture

**Eliminated Costs:**
- GKE Autopilot cluster management fees
- Multiple pod overhead
- WebSocket proxy layer complexity

**Simplified Pricing:**
- Predictable monthly VM cost regardless of session count
- Direct SSH connections without proxy overhead
- Single-instance architecture easier to monitor and optimize

## High-Scale Considerations

For very high usage (1000+ concurrent users), consider:
- Auto-scaling instance groups with multiple VMs
- Regional distribution for latency optimization
- Cost would scale roughly linearly: ~$50-100 per 1000 concurrent users

The current architecture provides excellent cost efficiency for small-to-medium scale deployments while maintaining the flexibility to scale up cost-effectively.