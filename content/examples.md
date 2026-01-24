---
title: "Examples"
---


This guide walks through all the example projects included with wetwire-k8s-go. Each example demonstrates different patterns and use cases.

## Overview

The examples demonstrate:

1. **guestbook** - Multi-tier application with Redis backend
2. **web-service** - Production-ready web service with Ingress
3. **configmap-secret** - Configuration and secret management patterns

All examples follow wetwire best practices and can be used as templates for your own projects.

## Example 1: Guestbook

**Location:** `/examples/guestbook/`

**What it demonstrates:**
- Multi-tier application architecture
- Redis leader-follower pattern
- Service-to-service communication
- Resource requests and limits
- Label organization across tiers

### Architecture

```
                        Frontend
                        Service    (3 replicas, LoadBalancer)
                           |
                           v
                        Frontend
                       Deployment
                           |
          +----------------+----------------+
          |                                 |
          v                                 v
   Redis Follower                    Redis Leader
      Service                           Service
          |                                 |
          v                                 v
   Redis Follower                    Redis Leader
     Deployment                        Deployment
     (2 replicas)                      (1 replica)
```

### Key Components

#### Shared Labels

```go
var redisLeaderLabels = map[string]string{
    "app":  "redis",
    "role": "leader",
    "tier": "backend",
}

var redisFollowerLabels = map[string]string{
    "app":  "redis",
    "role": "follower",
    "tier": "backend",
}

var frontendLabels = map[string]string{
    "app":  "guestbook",
    "tier": "frontend",
}
```

**Pattern:** Extract shared labels to variables for consistency and reuse.

#### Redis Leader (Master)

```go
var RedisLeaderDeployment = &appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{
        Name:   "redis-leader",
        Labels: redisLeaderLabels,
    },
    Spec: appsv1.DeploymentSpec{
        Replicas: ptr(int32(1)),  // Single leader instance
        Selector: &metav1.LabelSelector{
            MatchLabels: redisLeaderLabels,
        },
        // ... pod template
    },
}

var RedisLeaderService = &corev1.Service{
    ObjectMeta: metav1.ObjectMeta{
        Name:   "redis-leader",
        Labels: redisLeaderLabels,
    },
    Spec: corev1.ServiceSpec{
        Selector: redisLeaderLabels,  // Direct reference to labels
        // ... ports
    },
}
```

**Pattern:** Use the same label variable for Deployment labels, selector, and Service selector.

#### Resource Requests

```go
Resources: corev1.ResourceRequirements{
    Requests: corev1.ResourceList{
        corev1.ResourceCPU:    resource.MustParse("100m"),
        corev1.ResourceMemory: resource.MustParse("100Mi"),
    },
}
```

**Pattern:** Always specify resource requests for production workloads.

### Running the Example

```bash
cd examples/guestbook

# Build manifests
wetwire-k8s build -o guestbook.yaml

# Verify
cat guestbook.yaml

# Apply to cluster
kubectl apply -f guestbook.yaml

# Check deployment
kubectl get all -l tier=backend
kubectl get all -l tier=frontend

# Access the application
kubectl get service frontend
# Use the EXTERNAL-IP to access the guestbook

# Cleanup
kubectl delete -f guestbook.yaml
```

### When to Use This Pattern

Use the guestbook pattern for:
- Multi-tier applications
- Applications with separate frontend and backend
- Leader-follower architectures
- Microservices that need service discovery

### Variations

**Horizontal scaling:**
```go
// Scale frontend based on traffic
var frontendReplicas = int32(5)

var FrontendDeployment = &appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Replicas: ptr(frontendReplicas),
        // ...
    },
}
```

**Environment-specific configuration:**
```go
// Different replica counts per environment
const (
    devReplicas  = 1
    prodReplicas = 5
)

var replicas = prodReplicas // or use build tags
```

## Example 2: Web Service

**Location:** `/examples/web-service/`

**What it demonstrates:**
- Production-ready deployment configuration
- Health checks (liveness and readiness probes)
- Resource limits and requests
- Security context settings
- Ingress configuration with TLS
- Rolling update strategy
- Prometheus annotations

### Architecture

```
      Ingress    (webapp.example.com)
      (HTTPS)
          |
          v
      Service    (ClusterIP)
      (Port 80)
          |
          v
    Deployment   (3 replicas)
      (nginx)
```

### Key Components

#### Application Configuration

```go
const appName = "webapp"

var appLabels = map[string]string{
    "app":     appName,
    "version": "v1",
}
```

**Pattern:** Use constants for values referenced in multiple places.

#### Production-Ready Deployment

```go
var WebAppDeployment = &appsv1.Deployment{
    // ... metadata
    Spec: appsv1.DeploymentSpec{
        Replicas: ptr(int32(3)),
        Strategy: appsv1.DeploymentStrategy{
            Type: appsv1.RollingUpdateDeploymentStrategyType,
            RollingUpdate: &appsv1.RollingUpdateDeployment{
                MaxUnavailable: ptr(intstr.FromString("25%")),
                MaxSurge:       ptr(intstr.FromString("25%")),
            },
        },
        // ... template
    },
}
```

**Pattern:** Configure rolling updates for zero-downtime deployments.

#### Health Probes

```go
LivenessProbe: &corev1.Probe{
    ProbeHandler: corev1.ProbeHandler{
        HTTPGet: &corev1.HTTPGetAction{
            Path: "/healthz",
            Port: intstr.FromString("http"),
        },
    },
    InitialDelaySeconds: 10,
    PeriodSeconds:       10,
    TimeoutSeconds:      5,
    FailureThreshold:    3,
},
ReadinessProbe: &corev1.Probe{
    ProbeHandler: corev1.ProbeHandler{
        HTTPGet: &corev1.HTTPGetAction{
            Path: "/ready",
            Port: intstr.FromString("http"),
        },
    },
    InitialDelaySeconds: 5,
    PeriodSeconds:       5,
    TimeoutSeconds:      3,
    FailureThreshold:    3,
}
```

**Pattern:**
- Liveness probe restarts unhealthy pods
- Readiness probe removes unready pods from service endpoints
- Use different paths for different probe types

#### Security Context

```go
SecurityContext: &corev1.SecurityContext{
    ReadOnlyRootFilesystem:   ptr(true),
    RunAsNonRoot:             ptr(true),
    RunAsUser:                ptr(int64(1000)),
    AllowPrivilegeEscalation: ptr(false),
}
```

**Pattern:** Follow security best practices by default.

#### Resource Limits

```go
Resources: corev1.ResourceRequirements{
    Requests: corev1.ResourceList{
        corev1.ResourceCPU:    resource.MustParse("100m"),
        corev1.ResourceMemory: resource.MustParse("128Mi"),
    },
    Limits: corev1.ResourceList{
        corev1.ResourceCPU:    resource.MustParse("500m"),
        corev1.ResourceMemory: resource.MustParse("256Mi"),
    },
}
```

**Pattern:** Set both requests (guaranteed) and limits (maximum) for resources.

#### Service with Direct Reference

```go
var WebAppService = &corev1.Service{
    // ... metadata
    Spec: corev1.ServiceSpec{
        Type:     corev1.ServiceTypeClusterIP,
        Selector: WebAppDeployment.Spec.Selector.MatchLabels,  // Direct reference
        // ... ports
    },
}
```

**Pattern:** Reference deployment's selector directly instead of duplicating labels.

#### Ingress with TLS

```go
var ingressClassName = "nginx"

var WebAppIngress = &networkingv1.Ingress{
    ObjectMeta: metav1.ObjectMeta{
        Name:   appName,
        Labels: appLabels,
        Annotations: map[string]string{
            "nginx.ingress.kubernetes.io/ssl-redirect": "true",
            "nginx.ingress.kubernetes.io/use-regex":    "false",
        },
    },
    Spec: networkingv1.IngressSpec{
        IngressClassName: &ingressClassName,
        TLS: []networkingv1.IngressTLS{
            {
                Hosts:      []string{"webapp.example.com"},
                SecretName: "webapp-tls",
            },
        },
        Rules: []networkingv1.IngressRule{
            {
                Host: "webapp.example.com",
                IngressRuleValue: networkingv1.IngressRuleValue{
                    HTTP: &networkingv1.HTTPIngressRuleValue{
                        Paths: []networkingv1.HTTPIngressPath{
                            {
                                Path:     "/",
                                PathType: ptr(networkingv1.PathTypePrefix),
                                Backend: networkingv1.IngressBackend{
                                    Service: &networkingv1.IngressServiceBackend{
                                        Name: WebAppService.Name,  // Direct reference
                                        Port: networkingv1.ServiceBackendPort{
                                            Name: "http",
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    },
}
```

**Pattern:** Reference service name directly from service definition.

### Running the Example

```bash
cd examples/web-service

# Build manifests
wetwire-k8s build -o web-service.yaml

# Review generated YAML
cat web-service.yaml

# Apply to cluster
kubectl apply -f web-service.yaml

# Watch rollout
kubectl rollout status deployment/webapp

# Check health
kubectl get pods -l app=webapp
kubectl describe pod -l app=webapp

# Test service
kubectl port-forward service/webapp 8080:80
curl http://localhost:8080/

# Cleanup
kubectl delete -f web-service.yaml
```

### When to Use This Pattern

Use the web-service pattern for:
- Production web applications
- Services exposed via Ingress
- Applications requiring health checks
- Workloads with strict security requirements
- Services needing zero-downtime updates

### Variations

**Multiple backends:**
```go
var APIIngress = &networkingv1.Ingress{
    Spec: networkingv1.IngressSpec{
        Rules: []networkingv1.IngressRule{
            {
                Host: "api.example.com",
                IngressRuleValue: networkingv1.IngressRuleValue{
                    HTTP: &networkingv1.HTTPIngressRuleValue{
                        Paths: []networkingv1.HTTPIngressPath{
                            {
                                Path:     "/v1",
                                PathType: ptr(networkingv1.PathTypePrefix),
                                Backend:  /* backend for v1 API */,
                            },
                            {
                                Path:     "/v2",
                                PathType: ptr(networkingv1.PathTypePrefix),
                                Backend:  /* backend for v2 API */,
                            },
                        },
                    },
                },
            },
        },
    },
}
```

## Example 3: ConfigMap and Secret

**Location:** `/examples/configmap-secret/`

**What it demonstrates:**
- ConfigMap for application configuration
- ConfigMap for file content (nginx.conf, HTML)
- Secret for sensitive data
- Environment variables from ConfigMap/Secret
- Volume mounts for ConfigMap/Secret
- Different mounting strategies (directory, single file, subPath)

### Key Components

#### ConfigMap with Key-Value Data

```go
var AppConfig = &corev1.ConfigMap{
    ObjectMeta: metav1.ObjectMeta{
        Name:   "app-config",
        Labels: appLabels,
    },
    Data: map[string]string{
        "LOG_LEVEL":    "info",
        "ENABLE_DEBUG": "false",
        "MAX_WORKERS":  "10",
    },
}
```

**Pattern:** Simple key-value configuration.

#### ConfigMap with File Content

```go
var NginxConfig = &corev1.ConfigMap{
    Data: map[string]string{
        "nginx.conf": `events {
    worker_connections 1024;
}

http {
    server {
        listen 80;
        // ... config
    }
}`,
        "index.html": `<!DOCTYPE html>
<html>
<body>
    <h1>Configuration Demo</h1>
</body>
</html>`,
    },
}
```

**Pattern:** Store entire configuration files in ConfigMap.

#### Secret for Sensitive Data

```go
var AppSecrets = &corev1.Secret{
    Type: corev1.SecretTypeOpaque,
    StringData: map[string]string{
        "DATABASE_URL":      "postgres://user:password@db:5432/mydb",
        "API_KEY":           "sk-example-api-key-12345",
        "credentials.json": `{"client_id": "...", "client_secret": "..."}`,
    },
}
```

**Pattern:** Use `StringData` for automatic base64 encoding.

#### Environment Variables from ConfigMap

```go
Env: []corev1.EnvVar{
    {
        Name: "LOG_LEVEL",
        ValueFrom: &corev1.EnvVarSource{
            ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
                LocalObjectReference: corev1.LocalObjectReference{
                    Name: AppConfig.Name,  // Direct reference
                },
                Key: "LOG_LEVEL",
            },
        },
    },
}
```

**Pattern:** Individual environment variables from specific keys.

#### All ConfigMap Keys as Environment Variables

```go
EnvFrom: []corev1.EnvFromSource{
    {
        ConfigMapRef: &corev1.ConfigMapEnvSource{
            LocalObjectReference: corev1.LocalObjectReference{
                Name: AppConfig.Name,
            },
        },
    },
}
```

**Pattern:** Import all ConfigMap keys as environment variables.

#### Volume Mount for Entire ConfigMap

```go
Volumes: []corev1.Volume{
    {
        Name: "config-volume",
        VolumeSource: corev1.VolumeSource{
            ConfigMap: &corev1.ConfigMapVolumeSource{
                LocalObjectReference: corev1.LocalObjectReference{
                    Name: AppConfig.Name,
                },
            },
        },
    },
}

VolumeMounts: []corev1.VolumeMount{
    {
        Name:      "config-volume",
        MountPath: "/etc/app",
        ReadOnly:  true,
    },
}
```

**Pattern:** Mount entire ConfigMap as directory. Each key becomes a file.

#### Volume Mount for Single File (SubPath)

```go
VolumeMounts: []corev1.VolumeMount{
    {
        Name:      "nginx-config",
        MountPath: "/etc/nginx/nginx.conf",
        SubPath:   "nginx.conf",  // Mount single key as file
        ReadOnly:  true,
    },
}
```

**Pattern:** Use `SubPath` to mount individual ConfigMap keys as specific files.

#### Secret Volume with Permissions

```go
Volumes: []corev1.Volume{
    {
        Name: "secrets-volume",
        VolumeSource: corev1.VolumeSource{
            Secret: &corev1.SecretVolumeSource{
                SecretName:  AppSecrets.Name,
                DefaultMode: ptr(int32(0400)),  // Read-only by owner
            },
        },
    },
}
```

**Pattern:** Set restrictive permissions on secret files.

### Running the Example

```bash
cd examples/configmap-secret

# Build manifests
wetwire-k8s build -o config-demo.yaml

# Review secrets (note: not recommended for production)
cat config-demo.yaml

# Apply to cluster
kubectl apply -f config-demo.yaml

# Verify ConfigMaps
kubectl get configmap app-config -o yaml
kubectl get configmap nginx-config -o yaml

# Verify Secrets (base64 encoded)
kubectl get secret app-secrets -o yaml

# Check mounted files in pod
kubectl exec deployment/config-demo -- ls -la /etc/app
kubectl exec deployment/config-demo -- cat /etc/app/config.yaml
kubectl exec deployment/config-demo -- cat /etc/nginx/nginx.conf

# Check environment variables
kubectl exec deployment/config-demo -- env | grep -E 'LOG_LEVEL|DATABASE_URL'

# Cleanup
kubectl delete -f config-demo.yaml
```

### When to Use This Pattern

Use ConfigMaps for:
- Application configuration (non-sensitive)
- Configuration files (nginx.conf, application.yaml)
- Static content (HTML, JSON schemas)
- Feature flags and settings

Use Secrets for:
- Database credentials
- API keys and tokens
- TLS certificates
- OAuth credentials

### Variations

**ConfigMap from file:**
```bash
kubectl create configmap nginx-config --from-file=nginx.conf
```

Then reference in Go code:
```go
var NginxConfig = &corev1.ConfigMap{
    ObjectMeta: metav1.ObjectMeta{
        Name: "nginx-config",
    },
    // Data populated by kubectl create
}
```

**Immutable ConfigMap:**
```go
var AppConfig = &corev1.ConfigMap{
    Immutable: ptr(true),  // Prevents updates
    Data: map[string]string{
        "VERSION": "1.0.0",
    },
}
```

## Common Patterns Across Examples

### 1. Pointer Helper Function

All examples use:
```go
func ptr[T any](v T) *T { return &v }
```

**Why:** Many Kubernetes fields require pointers to distinguish "not set" from zero value.

### 2. Direct References

```go
var Service = &corev1.Service{
    Spec: corev1.ServiceSpec{
        Selector: Deployment.Spec.Selector.MatchLabels,  // Not duplicated
    },
}
```

**Why:** Creates dependencies that wetwire-k8s tracks for correct ordering.

### 3. Shared Label Variables

```go
var appLabels = map[string]string{"app": "myapp"}

var Deployment = &appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{
        Labels: appLabels,
    },
    Spec: appsv1.DeploymentSpec{
        Selector: &metav1.LabelSelector{
            MatchLabels: appLabels,
        },
        // ...
    },
}
```

**Why:** Ensures label consistency and makes changes easier.

### 4. Resource Organization

All examples group related resources with comments:
```go
// =============================================================================
// Application
// =============================================================================

var Deployment = ...
var Service = ...

// =============================================================================
// Configuration
// =============================================================================

var ConfigMap = ...
```

**Why:** Improves code readability and organization.

## Next Steps

- Try modifying the examples for your use case
- Run `wetwire-k8s lint` on the examples to see best practices
- Combine patterns from different examples
- Read [Quick Start](/quick-start/) for creating your own project
- See [Adoption](/adoption/) for migrating existing resources
- Check [Developers](/developers/) for contributing new examples
