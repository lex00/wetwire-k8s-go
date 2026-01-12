# StatefulSet Example

This example demonstrates StatefulSet patterns for stateful applications like databases.

## Resources

| Resource | Name | Description |
|----------|------|-------------|
| StatefulSet | `postgres` | PostgreSQL cluster with 3 replicas |
| Service | `postgres-headless` | Headless service for stable DNS |
| Service | `postgres` | ClusterIP service for clients |
| Secret | `postgres-credentials` | Database credentials |
| PodDisruptionBudget | `postgres-pdb` | Ensures minimum availability |

## Architecture

```
postgres-0.postgres-headless (Primary)
    ↓
postgres-1.postgres-headless (Replica)
    ↓
postgres-2.postgres-headless (Replica)
```

## Key Patterns

### Headless Service
- `ClusterIP: None` creates a headless service
- Each pod gets a stable DNS name: `<pod>.<service>.<namespace>.svc.cluster.local`
- Example: `postgres-0.postgres-headless.default.svc.cluster.local`

### Ordered Pod Management
- Pods are created sequentially: 0, 1, 2
- Pods are terminated in reverse order: 2, 1, 0
- Each pod waits for the previous to be Ready

### Persistent Storage
- `volumeClaimTemplates` creates a PVC per pod
- PVCs are named: `data-postgres-0`, `data-postgres-1`, `data-postgres-2`
- Data persists across pod restarts and rescheduling

### Pod Anti-Affinity
- Pods prefer to run on different nodes
- Improves availability during node failures

## Usage

```bash
# Build YAML manifests
wetwire-k8s build ./examples/stateful-app

# Apply to cluster
wetwire-k8s build ./examples/stateful-app | kubectl apply -f -

# Watch pods come up in order
kubectl get pods -l app.kubernetes.io/name=postgres -w

# Connect to the primary
kubectl exec -it postgres-0 -- psql -U postgres -d app

# Test DNS resolution
kubectl run -it --rm --image=busybox dns-test -- \
  nslookup postgres-headless.default.svc.cluster.local
```

## Configuration

### Scaling
```bash
# Scale up (new pods get their own PVC)
kubectl scale statefulset postgres --replicas=5

# Scale down (PVCs are retained!)
kubectl scale statefulset postgres --replicas=3

# Delete retained PVCs manually if needed
kubectl delete pvc data-postgres-3 data-postgres-4
```

### Rolling Updates
```bash
# Update is applied from highest ordinal to lowest
kubectl set image statefulset/postgres postgres=postgres:17-alpine

# Canary: only update pod-2 (partition=2 means update ordinal >= 2)
kubectl patch statefulset postgres -p '{"spec":{"updateStrategy":{"rollingUpdate":{"partition":2}}}}'

# Continue rollout
kubectl patch statefulset postgres -p '{"spec":{"updateStrategy":{"rollingUpdate":{"partition":0}}}}'
```

## Important Notes

1. **PVCs are NOT deleted** when scaling down or deleting StatefulSet
2. **Headless service is required** for StatefulSet DNS
3. **Init containers** run before main container on each pod
4. **PDB** prevents voluntary disruptions from reducing availability below threshold
5. **Credentials** should use external secrets management in production
