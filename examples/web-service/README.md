# Web Service Example

A production-ready web service with Deployment, Service, and Ingress demonstrating the wetwire pattern.

## Overview

This example deploys:

- **Deployment** - Web application with 3 replicas, probes, and resource limits
- **Service** - ClusterIP service for internal routing
- **Ingress** - External access with TLS termination

## Architecture

```
                    ┌─────────────────┐
                    │     Internet    │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │     Ingress     │
                    │  (TLS + Host)   │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │     Service     │
                    │   (ClusterIP)   │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
     ┌────────▼────────┐ ┌───▼───┐ ┌────────▼────────┐
     │     Pod 1       │ │ Pod 2 │ │     Pod 3       │
     │   (webapp)      │ │       │ │   (webapp)      │
     └─────────────────┘ └───────┘ └─────────────────┘
```

## Wetwire Pattern Highlights

### Resource References

The Service references the Deployment's selector labels directly:

```go
var WebAppService = &corev1.Service{
    Spec: corev1.ServiceSpec{
        Selector: WebAppDeployment.Spec.Selector.MatchLabels,
    },
}
```

The Ingress references the Service name:

```go
var WebAppIngress = &networkingv1.Ingress{
    Spec: networkingv1.IngressSpec{
        Rules: []networkingv1.IngressRule{
            {
                IngressRuleValue: networkingv1.IngressRuleValue{
                    HTTP: &networkingv1.HTTPIngressRuleValue{
                        Paths: []networkingv1.HTTPIngressPath{
                            {
                                Backend: networkingv1.IngressBackend{
                                    Service: &networkingv1.IngressServiceBackend{
                                        Name: WebAppService.Name,
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

### Shared Configuration

Labels and names are defined once and shared:

```go
const appName = "webapp"

var appLabels = map[string]string{
    "app":     appName,
    "version": "v1",
}
```

### Production-Ready Features

- **Health probes** - Liveness and readiness checks
- **Resource limits** - CPU and memory constraints
- **Security context** - Non-root user, read-only filesystem
- **Rolling updates** - Zero-downtime deployments
- **TLS** - Encrypted external traffic

## Build

Generate the Kubernetes manifests:

```bash
wetwire-k8s build ./examples/web-service -o webapp.yaml
```

## Prerequisites

Before deploying, ensure you have:

1. An ingress controller (e.g., nginx-ingress)
2. A TLS secret named `webapp-tls`
3. DNS pointing `webapp.example.com` to your ingress

Create the TLS secret (for testing):

```bash
kubectl create secret tls webapp-tls \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key
```

## Deploy

```bash
kubectl apply -f webapp.yaml
```

## Verify

Check deployment status:

```bash
kubectl get deployments webapp
kubectl get pods -l app=webapp
kubectl get service webapp
kubectl get ingress webapp
```

## Access

Once deployed and DNS is configured:

```bash
curl https://webapp.example.com
```

## Cleanup

```bash
kubectl delete -f webapp.yaml
```

## Resources

| Resource | Kind | Description |
|----------|------|-------------|
| webapp | Deployment | 3 replicas with probes and limits |
| webapp | Service | ClusterIP for internal routing |
| webapp | Ingress | TLS-enabled external access |

## Customization

### Change Replicas

Modify `WebAppDeployment.Spec.Replicas`:

```go
Replicas: ptr(int32(5)),
```

### Change Domain

Update the Ingress TLS and rules:

```go
TLS: []networkingv1.IngressTLS{
    {
        Hosts:      []string{"myapp.example.com"},
        SecretName: "myapp-tls",
    },
},
Rules: []networkingv1.IngressRule{
    {
        Host: "myapp.example.com",
        // ...
    },
},
```

### Add Environment Variables

Add to the container spec:

```go
Env: []corev1.EnvVar{
    {
        Name:  "LOG_LEVEL",
        Value: "info",
    },
},
```
