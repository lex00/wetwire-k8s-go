# Quick Start Guide

Get your e-commerce application running in 5 minutes!

## Prerequisites

- Kubernetes cluster (minikube, kind, GKE, EKS, AKS, etc.)
- `kubectl` configured to talk to your cluster
- `wetwire-k8s` CLI installed

## Step 1: Build the Manifests

```bash
make build
```

This generates `manifests.yaml` from the Go code.

## Step 2: Deploy Everything

```bash
make deploy
```

This creates:
- ✓ Namespace: `ecommerce`
- ✓ Database: PostgreSQL with persistent storage
- ✓ Backend: 2 API replicas
- ✓ Frontend: 3 web UI replicas (auto-scaling)
- ✓ Services: LoadBalancer for frontend, ClusterIP for backend/database
- ✓ Config: ConfigMap and Secret for configuration
- ✓ Security: NetworkPolicies for isolation
- ✓ Scaling: HPA for frontend auto-scaling

## Step 3: Check Status

```bash
make status
```

Wait for all pods to show `Running` status.

## Step 4: Get the URL

```bash
make url
```

This shows the LoadBalancer IP for accessing your frontend.

Visit `http://<EXTERNAL-IP>` in your browser!

## Common Commands

| Command | Description |
|---------|-------------|
| `make status` | Check if everything is running |
| `make logs-frontend` | View frontend logs |
| `make logs-backend` | View backend logs |
| `make logs-db` | View database logs |
| `make url` | Get the frontend URL |
| `make delete` | Remove everything from Kubernetes |

## Architecture Overview

```
┌─────────────────────────────────────────────┐
│                  Internet                    │
└────────────────┬────────────────────────────┘
                 │
                 ▼
         ┌───────────────┐
         │ LoadBalancer  │ (Port 80)
         └───────┬───────┘
                 │
                 ▼
         ┌───────────────┐
         │   Frontend    │ (3-10 replicas)
         │   Pods        │ Auto-scaling
         └───────┬───────┘
                 │
                 ▼
         ┌───────────────┐
         │  Backend API  │ (2 replicas)
         │   Pods        │
         └───────┬───────┘
                 │
                 ▼
         ┌───────────────┐
         │  PostgreSQL   │ (1 replica)
         │  Stateful     │ Persistent storage
         └───────────────┘
```

## What Each Component Does

### Frontend (Web UI)
- **What**: The user-facing web interface
- **Access**: Public (via LoadBalancer)
- **Replicas**: 3-10 (auto-scales with load)
- **Talks to**: Backend API

### Backend (REST API)
- **What**: Business logic and API endpoints
- **Access**: Internal only (ClusterIP)
- **Replicas**: 2
- **Talks to**: PostgreSQL database

### Database (PostgreSQL)
- **What**: Persistent data storage
- **Access**: Internal only (ClusterIP)
- **Replicas**: 1 (StatefulSet)
- **Storage**: 10Gi persistent volume

### Configuration
- **ConfigMap**: Non-sensitive settings (URLs, ports)
- **Secret**: Sensitive data (database password)

### Security
- **NetworkPolicies**: Control traffic between tiers
- **Namespace**: Isolated environment
- **Secrets**: Encrypted storage for credentials

## Resource Usage

| Component | CPU | Memory | Storage |
|-----------|-----|--------|---------|
| Frontend | 200m | 256Mi | - |
| Backend | 300m | 512Mi | - |
| Database | 500m | 1Gi | 10Gi |

## Troubleshooting

### Pods stuck in Pending
```bash
kubectl describe pod <pod-name> -n ecommerce
```
Usually means insufficient cluster resources.

### Can't access frontend
```bash
make url
```
LoadBalancer might take a few minutes to provision.

### Backend can't connect to database
```bash
make logs-backend
```
Check for connection errors. Database might still be initializing.

### Need to change something?
Edit `ecommerce.go` then run:
```bash
make build deploy
```

## Next Steps

1. **Update container images** in `ecommerce.go`:
   - Change `ecommerce/frontend:latest` to your frontend image
   - Change `ecommerce/backend:latest` to your backend image

2. **Change database password** (IMPORTANT):
   - Edit `DatabaseSecret.StringData["POSTGRES_PASSWORD"]`
   - Use a strong password for production!

3. **Test auto-scaling**:
   ```bash
   # Generate load on frontend
   kubectl run -i --tty load-generator --rm --image=busybox --restart=Never -- /bin/sh
   # Inside pod:
   while true; do wget -q -O- http://frontend-service.ecommerce; done

   # Watch pods scale up
   kubectl get hpa -n ecommerce -w
   ```

4. **Add health checks**:
   - Add `livenessProbe` and `readinessProbe` to containers
   - Ensures Kubernetes knows when pods are healthy

5. **Set up monitoring**:
   - Install Prometheus and Grafana
   - Monitor CPU, memory, request rates

## Questions?

Check the full [README.md](README.md) for detailed documentation.

## Clean Up

When you're done:
```bash
make delete
```

This removes everything from your cluster.
