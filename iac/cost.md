# Minimum Cost

Based on the configured resources, estimated minimum monthly cost with no traffic:

GKE Autopilot:

- 1 pod with 100m CPU / 128Mi RAM: ~$3-5/month (Autopilot charges for actual pod resources)

LoadBalancer Service:

- TCP LoadBalancer with forwarding rules: ~$18-22/month (Google Cloud Load Balancer pricing)

Cloud Run:

- Minimal idle cost for 1 CPU/2GB instance: ~$0-2/month (only charges when processing requests)

Other resources (minimal cost):

- DNS zone and records: ~$0.50/month
- Secret Manager: ~$0.18/month (6 secrets)
- Container Registry storage: ~$0.10/month

Total estimated minimum cost: ~$22-30/month

The LoadBalancer is the largest cost component at ~75% of the total. The GKE Autopilot pod and Cloud Run service are relatively efficient for small workloads, but the dedicated TCP LoadBalancer has a fixed monthly cost regardless of traffic volume.

# Usage Cost

Cloud Run per-second costs:

- CPU: 1 vCPU × $0.00002400/vCPU-second = $0.000024/second
- Memory: 2GB × $0.00000250/GB-second = $0.000005/second
- Total: ~$0.000029/second per Cloud Run instance

Per-80-user cost:

- Assuming each instance handles ~80 concurrent chess games
- Cost per user-second: $0.000029 ÷ 80 = ~$0.00000036/user-second
- Cost per user-minute: ~$0.000022/user-minute
- Cost per user-hour: ~$0.0013/user-hour

Example scenarios:
- 80 users playing for 10 minutes: ~$0.018
- 160 users (2 instances) for 30 minutes: ~$0.11
- 800 users (10 instances) for 1 hour: ~$1.04

The chess game's real-time nature means users stay connected for entire game durations (typically 10-60 minutes), making this quite cost-effective at ~$0.0013 per user per hour of gameplay.

On the other hand, a consistently active user would cost ~$0.94/user/month, so if this got popular and garnered 1M active users 24/7, it would cost $940k. So, that's something to think about...
