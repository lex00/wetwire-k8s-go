# NetworkPolicy Example

This example demonstrates Kubernetes NetworkPolicy patterns for securing pod-to-pod communication.

## Architecture

```
Internet → Frontend (tier=frontend)
              ↓
          Backend (tier=backend)
              ↓
          Database (tier=database)
```

## Policies

| Policy | Target | Ingress | Egress |
|--------|--------|---------|--------|
| `default-deny-all` | All pods | Deny all | Deny all |
| `frontend-policy` | Frontend | External:80,443 | Backend:8080, DNS |
| `backend-policy` | Backend | Frontend:8080 | Database:5432, DNS |
| `database-policy` | Database | Backend:5432 | DNS only |
| `allow-monitoring` | All pods | Prometheus:9090 | - |

## Security Model

1. **Default Deny** - Start with deny-all baseline
2. **Least Privilege** - Only allow required connections
3. **Tier Isolation** - Each tier only talks to adjacent tiers
4. **DNS Allowed** - All pods can resolve DNS
5. **Monitoring** - Cross-namespace access for Prometheus

## Usage

```bash
# Build YAML manifests
wetwire-k8s build ./examples/network-policy

# Apply to cluster
wetwire-k8s build ./examples/network-policy | kubectl apply -f -

# Verify policies
kubectl get networkpolicies

# Test connectivity (should work)
kubectl exec frontend-pod -- curl backend:8080

# Test connectivity (should fail)
kubectl exec frontend-pod -- curl database:5432
```

## Prerequisites

- CNI plugin with NetworkPolicy support (Calico, Cilium, Weave, etc.)
- Label your pods with `tier=frontend`, `tier=backend`, `tier=database`

## Important Notes

- NetworkPolicies are **additive** - multiple policies combine
- Empty `podSelector` matches **all pods** in namespace
- Empty `from`/`to` allows traffic from/to **anywhere**
- Always include DNS egress rules (port 53)
