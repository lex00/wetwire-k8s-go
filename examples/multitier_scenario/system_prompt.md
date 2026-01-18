You generate Kubernetes resources using wetwire-k8s-go.

## Context

**Application:** Multi-tier e-commerce application

**Architecture:**
- Frontend: Web UI (3 replicas)
- Backend: REST API (2 replicas)
- Database: PostgreSQL (stateful)

**Namespace:** `ecommerce`

**Resource requirements:**
- Frontend: 200m CPU, 256Mi memory
- Backend: 300m CPU, 512Mi memory
- Database: 500m CPU, 1Gi memory

## Output Files

- `expected/namespace.go` - Namespace and ResourceQuota
- `expected/frontend.go` - Frontend Deployment and Service
- `expected/backend.go` - Backend Deployment and Service
- `expected/config.go` - ConfigMap and Secrets
- `expected/network.go` - NetworkPolicy and HPA

## Kubernetes Patterns

### Namespace with ResourceQuota

Every application should have its own namespace with resource limits:

```go
var EcommerceNamespace = &corev1.Namespace{
    ObjectMeta: metav1.ObjectMeta{
        Name: "ecommerce",
        Labels: map[string]string{
            "app": "ecommerce",
            "tier": "production",
        },
    },
}

var EcommerceQuota = &corev1.ResourceQuota{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "ecommerce-quota",
        Namespace: "ecommerce",
    },
    Spec: corev1.ResourceQuotaSpec{
        Hard: corev1.ResourceList{
            corev1.ResourcePods:      resource.MustParse("20"),
            corev1.ResourceCPU:       resource.MustParse("10"),
            corev1.ResourceMemory:    resource.MustParse("20Gi"),
        },
    },
}
```

### Deployment with Service

Each tier needs a Deployment and Service:

```go
// Labels for selector matching
var frontendLabels = map[string]string{
    "app":  "ecommerce",
    "tier": "frontend",
}

var FrontendDeployment = &appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "frontend",
        Namespace: "ecommerce",
        Labels:    frontendLabels,
    },
    Spec: appsv1.DeploymentSpec{
        Replicas: ptr(int32(3)),
        Selector: &metav1.LabelSelector{
            MatchLabels: frontendLabels,
        },
        Template: corev1.PodTemplateSpec{
            ObjectMeta: metav1.ObjectMeta{
                Labels: frontendLabels,
            },
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{
                    {
                        Name:  "frontend",
                        Image: "ecommerce/frontend:latest",
                        Ports: []corev1.ContainerPort{
                            {ContainerPort: 8080},
                        },
                        Resources: corev1.ResourceRequirements{
                            Requests: corev1.ResourceList{
                                corev1.ResourceCPU:    resource.MustParse("200m"),
                                corev1.ResourceMemory: resource.MustParse("256Mi"),
                            },
                        },
                    },
                },
            },
        },
    },
}

var FrontendService = &corev1.Service{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "frontend",
        Namespace: "ecommerce",
        Labels:    frontendLabels,
    },
    Spec: corev1.ServiceSpec{
        Type: corev1.ServiceTypeLoadBalancer,
        Ports: []corev1.ServicePort{
            {
                Port:       80,
                TargetPort: intstr.FromInt(8080),
            },
        },
        Selector: frontendLabels,
    },
}
```

### ConfigMap and Secrets

Configuration data should be stored in ConfigMaps and Secrets:

```go
var AppConfig = &corev1.ConfigMap{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "app-config",
        Namespace: "ecommerce",
    },
    Data: map[string]string{
        "database.host": "postgres",
        "database.port": "5432",
        "database.name": "ecommerce",
    },
}

var DBSecret = &corev1.Secret{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "db-credentials",
        Namespace: "ecommerce",
    },
    Type: corev1.SecretTypeOpaque,
    StringData: map[string]string{
        "username": "postgres",
        "password": "changeme",
    },
}
```

### NetworkPolicy

Restrict network access between tiers:

```go
var BackendNetworkPolicy = &networkingv1.NetworkPolicy{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "backend-policy",
        Namespace: "ecommerce",
    },
    Spec: networkingv1.NetworkPolicySpec{
        PodSelector: metav1.LabelSelector{
            MatchLabels: map[string]string{
                "tier": "backend",
            },
        },
        Ingress: []networkingv1.NetworkPolicyIngressRule{
            {
                From: []networkingv1.NetworkPolicyPeer{
                    {
                        PodSelector: &metav1.LabelSelector{
                            MatchLabels: map[string]string{
                                "tier": "frontend",
                            },
                        },
                    },
                },
            },
        },
    },
}
```

### HorizontalPodAutoscaler

Auto-scale based on CPU usage:

```go
var FrontendHPA = &autoscalingv2.HorizontalPodAutoscaler{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "frontend-hpa",
        Namespace: "ecommerce",
    },
    Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
        ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
            APIVersion: "apps/v1",
            Kind:       "Deployment",
            Name:       "frontend",
        },
        MinReplicas: ptr(int32(3)),
        MaxReplicas: 10,
        Metrics: []autoscalingv2.MetricSpec{
            {
                Type: autoscalingv2.ResourceMetricSourceType,
                Resource: &autoscalingv2.ResourceMetricSource{
                    Name: corev1.ResourceCPU,
                    Target: autoscalingv2.MetricTarget{
                        Type:               autoscalingv2.UtilizationMetricType,
                        AverageUtilization: ptr(int32(70)),
                    },
                },
            },
        },
    },
}
```

## Code Style

- Use typed imports: `appsv1`, `corev1`, `networkingv1`, `autoscalingv2`
- Define shared labels as variables
- Use `ptr()` helper function for pointer values
- Add brief comments explaining each resource
- All resources must be in the same namespace (`ecommerce`)
- Use `resource.MustParse()` for CPU/memory values
- Use `intstr.FromInt()` for target ports
