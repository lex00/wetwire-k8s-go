<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../../docs/wetwire-dark.svg">
  <img src="../../docs/wetwire-light.svg" width="100" height="67">
</picture>

This example demonstrates Kubernetes autoscaling patterns using the wetwire pattern.

## Resources

| Resource | Name | Description |
|----------|------|-------------|
| Deployment | `web-app` | Target deployment for autoscaling |
| HPA | `web-app-hpa` | CPU and memory-based autoscaling |
| HPA | `web-app-hpa-custom` | Custom metrics autoscaling |

## Scaling Configuration

### Basic HPA (`web-app-hpa`)
- **Min replicas**: 2
- **Max replicas**: 10
- **CPU target**: 70% utilization
- **Memory target**: 80% utilization
- **Scale up**: Aggressive (100% increase or +4 pods per 15s)
- **Scale down**: Conservative (10% decrease per 60s, 5min stabilization)

### Custom Metrics HPA (`web-app-hpa-custom`)
- **Min replicas**: 2
- **Max replicas**: 20
- **CPU target**: 50% utilization
- **Custom metric**: 1000 requests/second per pod

## Prerequisites

1. **Metrics Server** - Required for CPU/memory metrics
2. **Prometheus Adapter** - Required for custom metrics (optional)

## Usage

```bash
# Build YAML manifests
wetwire-k8s build ./examples/hpa

# Apply to cluster
wetwire-k8s build ./examples/hpa | kubectl apply -f -

# Watch HPA status
kubectl get hpa -w

# Generate load to trigger scaling
kubectl run load-generator --image=busybox --restart=Never -- \
  /bin/sh -c "while true; do wget -q -O- http://web-app; done"
```

## Important Notes

1. **Resource requests are REQUIRED** - HPA needs resource requests to calculate utilization
2. **Metrics Server must be installed** - `kubectl top pods` should work
3. **Stabilization windows** - Prevent thrashing during load spikes
4. **Multiple metrics** - HPA scales to satisfy ALL metrics (most aggressive wins)
