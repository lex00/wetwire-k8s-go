# E-Commerce Kubernetes Setup

This directory contains a complete Kubernetes setup for a multi-tier e-commerce application.

## What's Included

### 1. **Namespace** (`EcommerceNamespace`)
   - Creates an isolated namespace called `ecommerce` for all your resources

### 2. **Database Layer** (PostgreSQL)
   - **StatefulSet** (`PostgresStatefulSet`): Runs PostgreSQL with persistent storage
   - **Service** (`PostgresService`): Internal service for database connections
   - **Secret** (`DatabaseSecret`): Safely stores database credentials
   - Resources: 500m CPU, 1Gi memory, 10Gi storage

### 3. **Backend API Layer**
   - **Deployment** (`BackendDeployment`): Runs 2 replicas of your REST API
   - **Service** (`BackendService`): Internal service for API connections
   - Resources: 300m CPU, 512Mi memory
   - Connects to database using credentials from the Secret

### 4. **Frontend Web Layer**
   - **Deployment** (`FrontendDeployment`): Runs 3 replicas of your web UI
   - **Service** (`FrontendService`): LoadBalancer type for internet access
   - **HPA** (`FrontendHPA`): Auto-scales from 3 to 10 replicas based on CPU (70% threshold)
   - Resources: 200m CPU, 256Mi memory

### 5. **Configuration**
   - **ConfigMap** (`AppConfig`): Stores non-sensitive configuration like service URLs
   - **Secret** (`DatabaseSecret`): Stores sensitive database credentials

### 6. **Network Security**
   - **NetworkPolicy** (`FrontendToBackendPolicy`): Allows frontend to talk to backend
   - **NetworkPolicy** (`BackendToDatabasePolicy`): Allows backend to talk to database

## Architecture

```
Internet
   │
   ▼
[LoadBalancer] ← Frontend Service
   │
   ▼
[Frontend Pods] (3-10 replicas, auto-scaling)
   │
   ▼
[Backend Service]
   │
   ▼
[Backend Pods] (2 replicas)
   │
   ▼
[Postgres Service]
   │
   ▼
[PostgreSQL StatefulSet] (1 replica with persistent storage)
```

## How to Use

### Step 1: Build the Kubernetes manifests

```bash
cd /Users/alex/Documents/checkouts/wetwire-k8s-go/examples/multitier_scenario/results/beginner
wetwire-k8s build -o manifests.yaml
```

### Step 2: Deploy to Kubernetes

```bash
kubectl apply -f manifests.yaml
```

### Step 3: Check the deployment

```bash
# See all resources in the ecommerce namespace
kubectl get all -n ecommerce

# Check if pods are running
kubectl get pods -n ecommerce

# Get the frontend LoadBalancer IP
kubectl get service frontend-service -n ecommerce
```

### Step 4: Access your application

Once the LoadBalancer is ready, you'll see an EXTERNAL-IP:

```bash
kubectl get service frontend-service -n ecommerce -w
```

Visit `http://<EXTERNAL-IP>` in your browser to access the frontend.

## Configuration Details

### Database Connection

The backend connects to PostgreSQL using:
- **Host**: `postgres-service` (from ConfigMap)
- **Port**: `5432` (from ConfigMap)
- **User**: `ecommerce_user` (from Secret)
- **Password**: `changeme123` (from Secret - **CHANGE THIS!**)
- **Database**: `ecommerce_db` (from Secret)

### Backend API URL

The frontend connects to the backend using:
- **URL**: `http://backend-service:8080` (from ConfigMap)

### Scaling Behavior

The frontend will automatically scale:
- **Minimum**: 3 replicas (always running)
- **Maximum**: 10 replicas (under high load)
- **Trigger**: When CPU usage exceeds 70%

## Security Features

1. **Secrets**: Database credentials are stored in a Secret (not plaintext)
2. **Network Policies**:
   - Frontend can only talk to backend
   - Backend can only talk to database
   - Database is isolated from direct external access
3. **Internal Services**: Backend and database use ClusterIP (not exposed to internet)
4. **Namespace Isolation**: All resources in dedicated `ecommerce` namespace

## Important Notes

### Before Production Deployment

1. **Change the database password** in `DatabaseSecret.StringData["POSTGRES_PASSWORD"]`
2. **Update container images** to point to your actual Docker images:
   - `ecommerce/frontend:latest`
   - `ecommerce/backend:latest`
3. **Configure storage class** for persistent volumes if needed
4. **Set up TLS/HTTPS** for the frontend LoadBalancer
5. **Add resource limits** (currently only requests are set)
6. **Configure health checks** (readiness and liveness probes)

### Container Images

You'll need to build and push your Docker images:

```bash
# Frontend
docker build -t ecommerce/frontend:latest ./frontend
docker push ecommerce/frontend:latest

# Backend
docker build -t ecommerce/backend:latest ./backend
docker push ecommerce/backend:latest
```

### Database Backup

The PostgreSQL data is stored in a PersistentVolume. Make sure to:
- Set up regular backups
- Test restore procedures
- Consider using a managed database service for production

## Troubleshooting

### Pods not starting

```bash
kubectl describe pod <pod-name> -n ecommerce
kubectl logs <pod-name> -n ecommerce
```

### Service not accessible

```bash
# Check service endpoints
kubectl get endpoints -n ecommerce

# Check network policies
kubectl get networkpolicies -n ecommerce
```

### Database connection issues

```bash
# Check if database is ready
kubectl get statefulset postgres -n ecommerce

# Check database logs
kubectl logs postgres-0 -n ecommerce

# Test connection from backend pod
kubectl exec -it <backend-pod> -n ecommerce -- sh
# Inside pod:
# nc -zv postgres-service 5432
```

## Next Steps

1. **Monitoring**: Add Prometheus and Grafana for observability
2. **Logging**: Set up centralized logging (ELK stack or similar)
3. **Ingress**: Replace LoadBalancer with an Ingress controller for better routing
4. **CI/CD**: Automate deployments with GitOps (ArgoCD, Flux)
5. **Backup**: Implement automated database backups
6. **Health Checks**: Add readiness and liveness probes to all deployments
