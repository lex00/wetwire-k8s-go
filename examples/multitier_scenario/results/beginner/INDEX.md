# E-Commerce Kubernetes Setup - File Index

This directory contains everything you need to deploy a complete multi-tier e-commerce application on Kubernetes.

## 📁 Files Overview

### Core Files

| File | Description | Start Here? |
|------|-------------|-------------|
| **[ecommerce.go](ecommerce.go)** | Main Go source file defining all Kubernetes resources | No |
| **[go.mod](go.mod)** | Go module definition | No |
| **[Makefile](Makefile)** | Convenient commands for building and deploying | ⭐ Yes |

### Documentation

| File | Description | Start Here? |
|------|-------------|-------------|
| **[QUICKSTART.md](QUICKSTART.md)** | Get running in 5 minutes | ⭐⭐⭐ Start Here! |
| **[README.md](README.md)** | Complete setup guide with detailed instructions | ⭐⭐ Read Second |
| **[ARCHITECTURE.md](ARCHITECTURE.md)** | System architecture and design decisions | Read Third |
| **[TROUBLESHOOTING.md](TROUBLESHOOTING.md)** | Common problems and solutions | When Issues Occur |
| **[INDEX.md](INDEX.md)** | This file - overview of all files | You Are Here |

### Generated Files

| File | Description | How to Create |
|------|-------------|---------------|
| **manifests.yaml** | Generated Kubernetes YAML (not included) | Run `make build` |
| **graph.md** | Dependency graph visualization (optional) | Run `make graph` |

## 🚀 Quick Start Path

Follow these steps in order:

1. **Read**: [QUICKSTART.md](QUICKSTART.md) - 5-minute deployment guide
2. **Build**: Run `make build` to generate manifests
3. **Deploy**: Run `make deploy` to deploy to Kubernetes
4. **Verify**: Run `make status` to check if everything is running
5. **Access**: Run `make url` to get the frontend URL

## 📚 Learning Path

If you want to understand the system deeply:

1. **[QUICKSTART.md](QUICKSTART.md)** - Get it running first
2. **[README.md](README.md)** - Understand what each component does
3. **[ARCHITECTURE.md](ARCHITECTURE.md)** - Learn the architecture and design
4. **[ecommerce.go](ecommerce.go)** - Study the actual resource definitions
5. **[TROUBLESHOOTING.md](TROUBLESHOOTING.md)** - Bookmark for when issues arise

## 🎯 File Purposes

### ecommerce.go
The heart of the system. Contains Go code that defines:
- Namespace (`EcommerceNamespace`)
- Frontend deployment and service (3-10 replicas, auto-scaling)
- Backend deployment and service (2 replicas)
- PostgreSQL StatefulSet and service (1 replica, persistent storage)
- ConfigMap for app settings
- Secret for database credentials
- HPA for auto-scaling
- NetworkPolicies for security

### QUICKSTART.md
Your first stop. Contains:
- Prerequisites checklist
- 4-step deployment process
- Common commands reference
- Quick architecture diagram
- Troubleshooting basics

### README.md
Complete reference guide with:
- Detailed component descriptions
- Step-by-step deployment instructions
- Configuration details
- Security features explained
- Production readiness checklist

### ARCHITECTURE.md
Technical deep dive including:
- System architecture diagram (Mermaid)
- Network flow visualization
- Component specifications
- Service communication patterns
- Security architecture
- Auto-scaling behavior
- High availability considerations
- Monitoring recommendations

### TROUBLESHOOTING.md
Problem-solving guide covering:
- Deployment issues (Pending pods, ImagePullBackOff, CrashLoopBackOff)
- Networking issues (LoadBalancer, service communication)
- Database issues (startup, connection, storage)
- Scaling issues (HPA not working)
- Configuration issues (ConfigMap/Secret updates)
- Performance issues (slow response, high memory)
- Emergency procedures

### Makefile
Automation helper with commands:
- `make build` - Generate Kubernetes manifests
- `make deploy` - Deploy to Kubernetes
- `make status` - Check deployment status
- `make url` - Get frontend URL
- `make logs-*` - View component logs
- `make delete` - Remove everything
- `make validate` - Validate code
- `make lint` - Check for issues

## 🏗️ What Gets Created

When you run `make deploy`, Kubernetes creates:

### Namespace
- `ecommerce` - Isolated namespace for all resources

### Frontend (Web UI)
- Deployment: `frontend` (3-10 replicas)
- Service: `frontend-service` (LoadBalancer, port 80)
- HPA: `frontend-hpa` (auto-scales 3-10 based on CPU)

### Backend (REST API)
- Deployment: `backend` (2 replicas)
- Service: `backend-service` (ClusterIP, port 8080)

### Database (PostgreSQL)
- StatefulSet: `postgres` (1 replica)
- Service: `postgres-service` (ClusterIP, port 5432)
- PVC: `postgres-storage-postgres-0` (10Gi)

### Configuration
- ConfigMap: `app-config` (service URLs, ports)
- Secret: `postgres-secret` (database credentials)

### Security
- NetworkPolicy: `frontend-to-backend`
- NetworkPolicy: `backend-to-database`

## 🔧 Common Commands

```bash
# Build manifests
make build

# Deploy everything
make deploy

# Check status
make status

# View logs
make logs-frontend
make logs-backend
make logs-db

# Get frontend URL
make url

# Clean up
make delete
```

## 📊 Resource Requirements

| Component | CPU | Memory | Storage |
|-----------|-----|--------|---------|
| Frontend (min 3 pods) | 600m | 768Mi | - |
| Backend (2 pods) | 600m | 1Gi | - |
| Database (1 pod) | 500m | 1Gi | 10Gi |
| **Total (min)** | **1.7 cores** | **2.7Gi** | **10Gi** |

With auto-scaling maxed out: **3.7 cores**, **4.5Gi**, **10Gi**

## 🔒 Security Features

- ✅ Database credentials in Secrets (encrypted)
- ✅ NetworkPolicies (restrict pod-to-pod communication)
- ✅ ClusterIP services (backend/database not exposed externally)
- ✅ Namespace isolation (all resources in `ecommerce` namespace)
- ❌ TLS/HTTPS not configured (add in production)
- ❌ Pod Security Policies not set (add in production)

## ⚠️ Before Production

Update these in `ecommerce.go`:

1. **Change database password** (line 36):
   ```go
   "POSTGRES_PASSWORD": "your-strong-password-here",
   ```

2. **Update container images** (lines 189, 287, 390):
   ```go
   Image: "your-registry/frontend:v1.0.0",
   Image: "your-registry/backend:v1.0.0",
   Image: "postgres:15-alpine",  // or managed DB service
   ```

3. **Add resource limits**
4. **Add health checks** (liveness/readiness probes)
5. **Configure TLS/HTTPS**
6. **Set up monitoring**
7. **Configure backups**

## 🆘 Need Help?

1. **Can't deploy?** → Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md) Deployment Issues
2. **Pods not running?** → Run `make status` and `kubectl describe pod <pod-name> -n ecommerce`
3. **Can't access frontend?** → Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md) Networking Issues
4. **Database issues?** → Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md) Database Issues
5. **General questions?** → Read [README.md](README.md)

## 📖 Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Wetwire K8s Go Repository](https://github.com/lex00/wetwire-k8s-go)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

## 🎓 Learning Kubernetes?

This project demonstrates:
- ✅ Multi-tier application architecture
- ✅ StatefulSets for databases
- ✅ Deployments for stateless apps
- ✅ Service types (LoadBalancer, ClusterIP)
- ✅ ConfigMaps and Secrets
- ✅ Horizontal Pod Autoscaling
- ✅ NetworkPolicies for security
- ✅ Persistent volumes
- ✅ Label selectors and pod communication
- ✅ Namespace isolation

Perfect for learning production-ready Kubernetes patterns!

## 📝 License

This example is part of the wetwire-k8s-go project.

---

**Next Step**: Read [QUICKSTART.md](QUICKSTART.md) to deploy in 5 minutes! 🚀
