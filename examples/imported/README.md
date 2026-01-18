# Imported Manifests

This directory contains Kubernetes manifests imported from production CNCF projects, converted to Go code using the wetwire-k8s pattern. These examples demonstrate real-world resource definitions and serve as reference implementations.

## Overview

The manifests in this directory have been imported from open-source Kubernetes projects and converted to type-safe Go code. They showcase various resource types including Deployments, Services, and ServiceAccounts.

## Source Projects

### Argo CD

[Argo CD](https://github.com/argoproj/argo-cd) is a declarative, GitOps continuous delivery tool for Kubernetes.

**Source Repository**: https://github.com/argoproj/argo-cd

**Imported Resources**:
- `argo-cd/redis-deployment.go` - Redis Deployment for Argo CD
  - Source: https://raw.githubusercontent.com/argoproj/argo-cd/master/manifests/base/redis/argocd-redis-deployment.yaml
  - Type: `apps/v1.Deployment`
  - Description: Deployment configuration for Redis cache used by Argo CD

- `argo-cd/server-service.go` - Argo CD API Server Service
  - Source: https://raw.githubusercontent.com/argoproj/argo-cd/master/manifests/base/server/argocd-server-service.yaml
  - Type: `v1.Service`
  - Description: Service exposing the Argo CD API server on HTTP/HTTPS ports

### kube-prometheus

[kube-prometheus](https://github.com/prometheus-operator/kube-prometheus) provides easy to operate end-to-end Kubernetes cluster monitoring with Prometheus using the Prometheus Operator.

**Source Repository**: https://github.com/prometheus-operator/kube-prometheus

**Imported Resources**:
- `kube-prometheus/grafana-deployment.go` - Grafana Deployment
  - Source: https://raw.githubusercontent.com/prometheus-operator/kube-prometheus/main/manifests/grafana-deployment.yaml
  - Type: `apps/v1.Deployment`
  - Description: Grafana dashboard deployment for visualizing Prometheus metrics

- `kube-prometheus/grafana-service.go` - Grafana Service
  - Source: https://raw.githubusercontent.com/prometheus-operator/kube-prometheus/main/manifests/grafana-service.yaml
  - Type: `v1.Service`
  - Description: Service exposing Grafana dashboard on port 3000

- `kube-prometheus/alertmanager-service.go` - Alertmanager Service
  - Source: https://raw.githubusercontent.com/prometheus-operator/kube-prometheus/main/manifests/alertmanager-service.yaml
  - Type: `v1.Service`
  - Description: Service exposing Alertmanager web interface and API

- `kube-prometheus/prometheus-serviceaccount.go` - Prometheus ServiceAccount
  - Source: https://raw.githubusercontent.com/prometheus-operator/kube-prometheus/main/manifests/prometheus-serviceAccount.yaml
  - Type: `v1.ServiceAccount`
  - Description: ServiceAccount for Prometheus pods with cluster monitoring permissions

## Resource Types Covered

This collection includes the following Kubernetes resource types:

- **Deployment** (apps/v1): Standard workload deployments
  - Argo CD Redis
  - Grafana dashboard

- **Service** (v1): Kubernetes Services for network access
  - Argo CD API Server
  - Grafana dashboard
  - Alertmanager

- **ServiceAccount** (v1): Identity for pods
  - Prometheus monitoring service account

## How to Use

Each Go file in this directory can be:

1. **Built** to generate YAML manifests:
   ```bash
   wetwire-k8s build examples/imported/argo-cd/
   ```

2. **Linted** to check for best practices:
   ```bash
   wetwire-k8s lint examples/imported/
   ```

3. **Modified** and extended as templates for your own resources

4. **Referenced** as examples of how to define complex Kubernetes resources in Go

## Notes

- All manifests were imported using `wetwire-k8s import` command
- The import preserves the original structure and configuration from the source projects
- Some complex fields may require manual adjustment depending on your use case
- Only standard Kubernetes API resources are included (Deployments, Services, ServiceAccounts, etc.)
- Custom Resources (CRs) requiring CRDs are not included, as they need specialized Go client libraries

## Import Date

These manifests were imported on January 17, 2026.

## License

The original manifests are subject to their respective project licenses:
- Argo CD: Apache 2.0
- kube-prometheus: Apache 2.0

The Go conversions in this directory are provided as examples under the same Apache 2.0 license as wetwire-k8s-go.
