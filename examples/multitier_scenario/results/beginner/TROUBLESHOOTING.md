# Troubleshooting Guide

Common issues and how to fix them.

## Table of Contents

- [Deployment Issues](#deployment-issues)
- [Networking Issues](#networking-issues)
- [Database Issues](#database-issues)
- [Scaling Issues](#scaling-issues)
- [Configuration Issues](#configuration-issues)
- [Performance Issues](#performance-issues)

---

## Deployment Issues

### Pods Stuck in "Pending" State

**Symptom:**
```bash
$ kubectl get pods -n ecommerce
NAME                        READY   STATUS    RESTARTS   AGE
frontend-7d8f9-x2k1m        0/1     Pending   0          5m
```

**Diagnosis:**
```bash
kubectl describe pod frontend-7d8f9-x2k1m -n ecommerce
```

**Common Causes & Solutions:**

#### 1. Insufficient Cluster Resources
```
Error: 0/3 nodes are available: 3 Insufficient cpu
```

**Solution:**
- Scale up your cluster nodes
- Reduce resource requests in `ecommerce.go`
- Delete unused pods/deployments

#### 2. Persistent Volume Not Available
```
Error: pod has unbound immediate PersistentVolumeClaims
```

**Solution:**
```bash
# Check PVC status
kubectl get pvc -n ecommerce

# If no storage class available, install one (for local testing):
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml
```

#### 3. Image Pull Errors
```
Error: Failed to pull image "ecommerce/frontend:latest"
```

**Solution:**
- Verify image exists: `docker pull ecommerce/frontend:latest`
- Check image registry authentication
- Use a public image for testing: `nginx:latest`

### Pods Stuck in "ImagePullBackOff"

**Symptom:**
```bash
NAME                        READY   STATUS             RESTARTS   AGE
backend-5d7f8-abc123        0/1     ImagePullBackOff   0          2m
```

**Solution:**

1. **Check image name:**
```bash
kubectl describe pod backend-5d7f8-abc123 -n ecommerce
```

2. **Fix in code:**
Edit `ecommerce.go` and change the image:
```go
// From:
Image: "ecommerce/backend:latest",

// To (for testing):
Image: "nginx:latest",
```

3. **Rebuild and redeploy:**
```bash
make build deploy
```

### Pods CrashLoopBackOff

**Symptom:**
```bash
NAME                        READY   STATUS              RESTARTS   AGE
backend-5d7f8-abc123        0/1     CrashLoopBackOff    5          5m
```

**Diagnosis:**
```bash
# Check logs
kubectl logs backend-5d7f8-abc123 -n ecommerce

# Check previous logs if pod restarted
kubectl logs backend-5d7f8-abc123 -n ecommerce --previous
```

**Common Causes:**

#### 1. Application Error
```
Error: Cannot connect to database
```

**Solution:**
- Check database is running: `kubectl get pods -n ecommerce | grep postgres`
- Verify environment variables are correct
- Check Secret values

#### 2. Missing Environment Variables
```
Error: DATABASE_HOST is not set
```

**Solution:**
```bash
# Verify ConfigMap exists
kubectl get configmap app-config -n ecommerce -o yaml

# Verify Secret exists
kubectl get secret postgres-secret -n ecommerce
```

#### 3. Port Already in Use
```
Error: listen tcp :8080: bind: address already in use
```

**Solution:**
- Check if multiple containers are trying to use same port
- Verify `containerPort` in deployment spec

---

## Networking Issues

### Cannot Access Frontend via LoadBalancer

**Symptom:**
```bash
$ make url
Frontend LoadBalancer:
Waiting for LoadBalancer IP...
```

**Diagnosis:**
```bash
kubectl get service frontend-service -n ecommerce
```

**Common Causes & Solutions:**

#### 1. LoadBalancer Not Supported (Local Clusters)

**For Minikube:**
```bash
# In a separate terminal, run:
minikube tunnel
```

**For Kind:**
```bash
# Install MetalLB
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.13.7/config/manifests/metallb-native.yaml

# Configure IP range
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      - 172.18.255.1-172.18.255.250
EOF
```

**For Cloud Providers:**
- Wait 2-5 minutes for LoadBalancer provisioning
- Check cloud provider console for LB creation errors

#### 2. Service Selector Mismatch

**Diagnosis:**
```bash
# Check endpoints
kubectl get endpoints frontend-service -n ecommerce
```

If no endpoints, selector doesn't match pod labels.

**Solution:**
Verify labels in `ecommerce.go`:
```go
// Service selector MUST match Pod labels
Service.Spec.Selector: map[string]string{
    "app":  "ecommerce",
    "tier": "frontend",  // Must match deployment labels
}
```

### Frontend Cannot Reach Backend

**Symptom:**
Frontend logs show:
```
Error: connect ECONNREFUSED 10.96.0.1:8080
```

**Diagnosis:**
```bash
# Test from frontend pod
kubectl exec -it -n ecommerce <frontend-pod> -- wget -O- http://backend-service:8080
```

**Solutions:**

#### 1. Check Backend Service
```bash
# Verify service exists
kubectl get service backend-service -n ecommerce

# Check endpoints
kubectl get endpoints backend-service -n ecommerce
```

#### 2. Check Network Policy
```bash
# List network policies
kubectl get networkpolicies -n ecommerce

# Describe the policy
kubectl describe networkpolicy frontend-to-backend -n ecommerce
```

If network policy is blocking traffic, verify labels match:
```go
// In ecommerce.go
FrontendToBackendPolicy.Spec.Ingress[0].From[0].PodSelector.MatchLabels
// Should match frontend pod labels
```

#### 3. Check Backend Pod Status
```bash
kubectl get pods -n ecommerce -l tier=backend
```

If no pods are ready, check backend logs.

### Backend Cannot Reach Database

**Symptom:**
Backend logs show:
```
Error: could not connect to server: Connection refused
```

**Diagnosis:**
```bash
# Test from backend pod
kubectl exec -it -n ecommerce <backend-pod> -- nc -zv postgres-service 5432
```

**Solutions:**

#### 1. Check Database Status
```bash
kubectl get statefulset postgres -n ecommerce
kubectl get pods -n ecommerce -l tier=database
```

#### 2. Check Database Service
```bash
kubectl get service postgres-service -n ecommerce
kubectl get endpoints postgres-service -n ecommerce
```

#### 3. Verify Connection String
The backend should use:
- Host: `postgres-service`
- Port: `5432`
- User: from `DATABASE_USER` env var (from Secret)
- Password: from `DATABASE_PASSWORD` env var (from Secret)

---

## Database Issues

### Database Pod Won't Start

**Symptom:**
```bash
postgres-0   0/1     Error    0          1m
```

**Diagnosis:**
```bash
kubectl logs postgres-0 -n ecommerce
```

**Common Causes:**

#### 1. PVC Already Exists with Incompatible Data
```
Error: database files are incompatible with server
```

**Solution:**
```bash
# Delete PVC (WARNING: This deletes all data!)
kubectl delete pvc postgres-storage-postgres-0 -n ecommerce

# Delete StatefulSet
kubectl delete statefulset postgres -n ecommerce

# Redeploy
make deploy
```

#### 2. Insufficient Permissions
```
Error: permission denied
```

**Solution:**
Add securityContext to StatefulSet in `ecommerce.go`:
```go
Spec: corev1.PodSpec{
    SecurityContext: &corev1.PodSecurityContext{
        FsGroup: ptrInt64(999), // postgres user group
    },
    // ... containers ...
}
```

### Cannot Connect to Database

**Symptom:**
```bash
$ make shell-db
Error from server: error dialing backend: dial tcp 10.1.2.3:5432: connect: connection refused
```

**Diagnosis:**
```bash
# Check if database is ready
kubectl get pods postgres-0 -n ecommerce

# Check logs
kubectl logs postgres-0 -n ecommerce
```

**Solution:**
Wait for database initialization to complete (30-60 seconds on first start).

### Database Running Out of Space

**Symptom:**
```
Error: no space left on device
```

**Diagnosis:**
```bash
# Check PVC usage
kubectl exec postgres-0 -n ecommerce -- df -h /var/lib/postgresql/data
```

**Solution:**

1. **Expand PVC (if supported by storage class):**
```bash
kubectl patch pvc postgres-storage-postgres-0 -n ecommerce -p '{"spec":{"resources":{"requests":{"storage":"20Gi"}}}}'
```

2. **Or create new larger PVC:**
```bash
# Backup first!
kubectl exec postgres-0 -n ecommerce -- pg_dump -U ecommerce_user ecommerce_db > backup.sql

# Delete old resources
kubectl delete statefulset postgres -n ecommerce
kubectl delete pvc postgres-storage-postgres-0 -n ecommerce

# Update storage size in ecommerce.go
VolumeClaimTemplates[0].Spec.Resources.Requests["storage"] = "20Gi"

# Redeploy
make build deploy

# Restore data
kubectl exec -i postgres-0 -n ecommerce -- psql -U ecommerce_user ecommerce_db < backup.sql
```

---

## Scaling Issues

### HPA Not Scaling

**Symptom:**
```bash
$ kubectl get hpa -n ecommerce
NAME           REFERENCE              TARGETS         MINPODS   MAXPODS   REPLICAS
frontend-hpa   Deployment/frontend    <unknown>/70%   3         10        3
```

**Diagnosis:**
```bash
kubectl describe hpa frontend-hpa -n ecommerce
```

**Common Causes:**

#### 1. Metrics Server Not Installed
```
Error: unable to get metrics for resource cpu
```

**Solution:**
```bash
# Install metrics-server
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# For local clusters (minikube, kind), add --kubelet-insecure-tls flag:
kubectl patch deployment metrics-server -n kube-system --type='json' -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-insecure-tls"}]'
```

#### 2. No Resource Requests Defined
HPA needs CPU requests to calculate percentage.

**Solution:**
Verify in `ecommerce.go`:
```go
Resources: corev1.ResourceRequirements{
    Requests: map[string]string{
        "cpu": "200m",  // Required for HPA
    },
}
```

#### 3. Not Enough Load
CPU usage is below 70% threshold.

**Solution - Generate Load:**
```bash
# Start load generator
kubectl run -i --tty load-generator --rm --image=busybox --restart=Never -- /bin/sh

# Inside pod, generate requests:
while true; do wget -q -O- http://frontend-service.ecommerce; done

# Watch HPA in another terminal:
kubectl get hpa -n ecommerce -w
```

### HPA Scaling Too Aggressively

**Symptom:**
Pods scaling up and down rapidly.

**Solution:**
Add stabilization window in `ecommerce.go`:
```go
Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
    // ... existing config ...
    Behavior: &autoscalingv2.HorizontalPodAutoscalerBehavior{
        ScaleDown: &autoscalingv2.HPAScalingRules{
            StabilizationWindowSeconds: ptrInt32(300), // 5 minutes
        },
        ScaleUp: &autoscalingv2.HPAScalingRules{
            StabilizationWindowSeconds: ptrInt32(60), // 1 minute
        },
    },
}
```

---

## Configuration Issues

### Pods Not Seeing ConfigMap/Secret Changes

**Symptom:**
Updated ConfigMap but pods still use old values.

**Explanation:**
Pods don't automatically reload environment variables from ConfigMaps/Secrets.

**Solution:**

1. **Rolling restart:**
```bash
kubectl rollout restart deployment frontend -n ecommerce
kubectl rollout restart deployment backend -n ecommerce
```

2. **Or use ConfigMap/Secret as volume mount (auto-reloads):**
```go
// In ecommerce.go, change from envFrom to volumeMount
VolumeMounts: []corev1.VolumeMount{
    {
        Name:      "config",
        MountPath: "/etc/config",
    },
}
Volumes: []corev1.Volume{
    {
        Name: "config",
        VolumeSource: corev1.VolumeSource{
            ConfigMap: &corev1.ConfigMapVolumeSource{
                Name: AppConfig.Metadata.Name,
            },
        },
    },
}
```

### Secret Values Not Working

**Symptom:**
Backend can't authenticate to database despite correct credentials.

**Diagnosis:**
```bash
# Check Secret values
kubectl get secret postgres-secret -n ecommerce -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d
```

**Common Issues:**

#### 1. Secret Created Before StringData Support
Use `StringData` (plain text) instead of `Data` (base64):
```go
// Good (plain text)
StringData: map[string]string{
    "POSTGRES_PASSWORD": "changeme123",
}

// Avoid (requires base64 encoding)
Data: map[string]string{
    "POSTGRES_PASSWORD": "Y2hhbmdlbWUxMjM=",
}
```

#### 2. Environment Variable Not Mounted
```bash
# Check env vars in pod
kubectl exec backend-xxx -n ecommerce -- env | grep DATABASE
```

If missing, verify `envFrom` in deployment spec.

---

## Performance Issues

### Slow Response Times

**Diagnosis:**
```bash
# Check pod CPU/Memory usage
kubectl top pods -n ecommerce

# Check for throttling
kubectl describe pod <pod-name> -n ecommerce | grep -i throttl
```

**Solutions:**

#### 1. Insufficient Resources
Increase resource requests:
```go
Resources: corev1.ResourceRequirements{
    Requests: map[string]string{
        "cpu":    "500m",  // Increased from 300m
        "memory": "1Gi",   // Increased from 512Mi
    },
}
```

#### 2. Database Connection Pool Exhausted
Check backend logs for:
```
Error: remaining connection slots are reserved
```

Increase PostgreSQL connections or optimize queries.

#### 3. Network Latency
Check if pods are on same node:
```bash
kubectl get pods -n ecommerce -o wide
```

Consider pod affinity/anti-affinity rules.

### High Memory Usage

**Diagnosis:**
```bash
kubectl top pods -n ecommerce
```

**Solutions:**

#### 1. Memory Leak
Check application logs and fix code.

#### 2. Set Resource Limits
```go
Resources: corev1.ResourceRequirements{
    Requests: map[string]string{
        "memory": "512Mi",
    },
    Limits: map[string]string{
        "memory": "1Gi",  // Kill if exceeds
    },
}
```

---

## Emergency Procedures

### Complete Reset

If everything is broken:

```bash
# Delete everything
make delete

# Or force delete namespace
kubectl delete namespace ecommerce --force --grace-period=0

# Wait for cleanup
kubectl get namespace ecommerce

# Redeploy fresh
make build deploy
```

### Backup Before Changes

```bash
# Export current state
kubectl get all -n ecommerce -o yaml > backup.yaml

# Backup database
kubectl exec postgres-0 -n ecommerce -- pg_dump -U ecommerce_user ecommerce_db > db-backup.sql
```

### View All Events

```bash
# Recent events in namespace
kubectl get events -n ecommerce --sort-by='.lastTimestamp'

# Watch events live
kubectl get events -n ecommerce --watch
```

---

## Getting Help

If you're still stuck:

1. **Check logs:**
   ```bash
   make logs-frontend
   make logs-backend
   make logs-db
   ```

2. **Check resource status:**
   ```bash
   make status
   ```

3. **Describe problem resources:**
   ```bash
   kubectl describe pod <pod-name> -n ecommerce
   ```

4. **Check events:**
   ```bash
   kubectl get events -n ecommerce
   ```

5. **Export for debugging:**
   ```bash
   kubectl get all -n ecommerce -o yaml > debug.yaml
   ```
