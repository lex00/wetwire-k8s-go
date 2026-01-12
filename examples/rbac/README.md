# RBAC Example

This example demonstrates Kubernetes Role-Based Access Control (RBAC) patterns using the wetwire pattern.

## Resources

| Resource | Name | Description |
|----------|------|-------------|
| ServiceAccount | `app-service-account` | Identity for application pods |
| Role | `app-role` | Namespace-scoped permissions |
| RoleBinding | `app-role-binding` | Binds Role to ServiceAccount |
| ClusterRole | `node-reader` | Cluster-wide read permissions |
| ClusterRoleBinding | `app-node-reader-binding` | Binds ClusterRole to ServiceAccount |

## Use Case

An application that needs to:
- Read ConfigMaps and Secrets in its namespace
- Watch pods and services in its namespace
- Read node information cluster-wide (for scheduling awareness)
- Read namespace information cluster-wide (for multi-tenant awareness)

## Permissions Granted

### Namespace-scoped (Role)
- `configmaps`: get, list, watch
- `secrets`: get, list, watch
- `pods`: get, list, watch
- `services`: get, list, watch, update, patch

### Cluster-wide (ClusterRole)
- `nodes`: get, list, watch
- `namespaces`: get, list, watch

## Usage

```bash
# Build YAML manifests
wetwire-k8s build ./examples/rbac

# Apply to cluster
wetwire-k8s build ./examples/rbac | kubectl apply -f -

# Use the ServiceAccount in a Deployment
# spec:
#   serviceAccountName: app-service-account
```

## Security Notes

- Follow the principle of least privilege
- Only grant permissions that are actually needed
- Use namespace-scoped Roles when possible
- Reserve ClusterRoles for truly cluster-wide needs
- Audit RBAC permissions regularly
