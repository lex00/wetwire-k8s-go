# Namespace Setup Example

This example demonstrates a production-ready namespace setup for multi-tenant Kubernetes clusters.

## Resources

| Resource | Name | Description |
|----------|------|-------------|
| Namespace | `team-alpha` | Isolated namespace for the team |
| ResourceQuota | `compute-quota` | Limits total resources in namespace |
| LimitRange | `default-limits` | Default/max limits for containers |
| NetworkPolicy | `default-deny-all` | Deny all traffic by default |
| NetworkPolicy | `allow-dns` | Allow DNS resolution |
| NetworkPolicy | `allow-same-namespace` | Allow intra-namespace traffic |
| ServiceAccount | `team-workload` | Identity for team workloads |
| Role | `namespace-developer` | Developer permissions |
| RoleBinding | `team-developers` | Binds role to team group |

## Configuration

### ResourceQuota Limits
| Resource | Request | Limit |
|----------|---------|-------|
| CPU | 10 cores | 20 cores |
| Memory | 20Gi | 40Gi |
| Pods | 50 | - |
| Services | 20 | - |
| Secrets | 100 | - |
| ConfigMaps | 100 | - |
| PVCs | 10 | - |
| Storage | 100Gi | - |

### LimitRange Defaults
| Resource | Default Request | Default Limit | Min | Max |
|----------|-----------------|---------------|-----|-----|
| CPU | 100m | 500m | 50m | 4 |
| Memory | 128Mi | 512Mi | 64Mi | 8Gi |
| PVC Storage | - | - | 1Gi | 20Gi |

## Usage

```bash
# Build YAML manifests
wetwire-k8s build ./examples/namespace-setup

# Apply to cluster
wetwire-k8s build ./examples/namespace-setup | kubectl apply -f -

# Verify resources
kubectl get namespace team-alpha
kubectl get resourcequota -n team-alpha
kubectl get limitrange -n team-alpha
kubectl get networkpolicy -n team-alpha
kubectl get role,rolebinding -n team-alpha
```

## Multi-Tenant Use Case

This example is designed for multi-tenant clusters where:

1. **Resource Isolation** - Each team has quota limits preventing resource starvation
2. **Network Isolation** - Default deny with explicit allow rules
3. **Access Control** - RBAC limits what team members can do
4. **Cost Tracking** - Labels and annotations for cost allocation

## Customization

To create namespaces for other teams:

1. Change `namespaceName` constant
2. Update `teamLabels` map
3. Adjust quota values as needed
4. Configure RoleBinding subjects for your identity provider

## Important Notes

- ResourceQuota requires resource requests on all pods
- LimitRange applies defaults when requests/limits are omitted
- NetworkPolicies require a CNI with policy support
- RBAC Group subjects require identity provider configuration
