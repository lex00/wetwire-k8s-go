# Guestbook Example

A multi-tier web application demonstrating the wetwire pattern with Redis and a PHP frontend.

## Overview

This example deploys:

- **Redis Leader** - Primary Redis instance for writes
- **Redis Followers** - Redis replicas for read scaling
- **Frontend** - PHP web application that stores guestbook entries in Redis

## Architecture

```
                    ┌─────────────────┐
                    │   LoadBalancer  │
                    │   (frontend)    │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │    Frontend     │
                    │   (3 replicas)  │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
     ┌────────▼────────┐    │    ┌─────────▼────────┐
     │  Redis Follower │    │    │  Redis Follower  │
     │   (replica 1)   │    │    │   (replica 2)    │
     └────────┬────────┘    │    └─────────┬────────┘
              │             │              │
              └─────────────┼──────────────┘
                            │
                   ┌────────▼────────┐
                   │  Redis Leader   │
                   │    (master)     │
                   └─────────────────┘
```

## Wetwire Pattern Highlights

### Shared Labels

Labels are defined once and reused across Deployments and Services:

```go
var redisLeaderLabels = map[string]string{
    "app":  "redis",
    "role": "leader",
    "tier": "backend",
}

var RedisLeaderDeployment = &appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Selector: &metav1.LabelSelector{
            MatchLabels: redisLeaderLabels,
        },
        Template: corev1.PodTemplateSpec{
            ObjectMeta: metav1.ObjectMeta{
                Labels: redisLeaderLabels,
            },
        },
    },
}

var RedisLeaderService = &corev1.Service{
    Spec: corev1.ServiceSpec{
        Selector: redisLeaderLabels,
    },
}
```

### Resource Tiers

Resources are organized by tier (frontend/backend) for clarity and management.

## Build

Generate the Kubernetes manifests:

```bash
wetwire-k8s build ./examples/guestbook -o guestbook.yaml
```

## Deploy

Apply to your cluster:

```bash
kubectl apply -f guestbook.yaml
```

## Access

Get the external IP:

```bash
kubectl get service frontend
```

Open the external IP in your browser to use the guestbook.

## Cleanup

```bash
kubectl delete -f guestbook.yaml
```

## Resources

| Resource | Kind | Replicas | Purpose |
|----------|------|----------|---------|
| redis-leader | Deployment | 1 | Redis primary for writes |
| redis-leader | Service | - | Internal access to leader |
| redis-follower | Deployment | 2 | Redis replicas for reads |
| redis-follower | Service | - | Internal access to followers |
| frontend | Deployment | 3 | Web application |
| frontend | Service | - | External LoadBalancer |

## Notes

- The frontend connects to Redis using DNS service discovery
- Redis followers replicate from the leader automatically
- This is based on the official Kubernetes Guestbook tutorial
