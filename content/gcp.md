---
title: "GCP Integration"
slug: "gcp"
---

Kubernetes integration patterns for Google Cloud Platform, covering GKE cluster management and GCP resource provisioning via Config Connector.

## Overview

| Approach | Tool | Use Case |
|----------|------|----------|
| **GKE Manifests** | wetwire-k8s | K8s workloads deployed to GKE |
| **Config Connector** | wetwire-k8s | GCP resources managed via kubectl |
| **Deployment Manager** | wetwire-gcp | GKE cluster and network infrastructure |

## GKE Workload Patterns

### Workload Identity

Associate Kubernetes service accounts with GCP service accounts:

```go
// Service account with Workload Identity annotation
var GCSAccessSA = corev1.ServiceAccount{
    Metadata: metav1.ObjectMeta{
        Name:      "gcs-access",
        Namespace: "default",
        Annotations: map[string]string{
            "iam.gke.io/gcp-service-account": "gcs-reader@my-project.iam.gserviceaccount.com",
        },
    },
}

// Pod using the service account
var DataProcessor = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                ServiceAccountName: "gcs-access",
                Containers: []corev1.Container{
                    {
                        Name:  "processor",
                        Image: "gcr.io/my-project/myapp:latest",
                        // GCP client libraries automatically use Workload Identity
                    },
                },
            },
        },
    },
}
```

### GKE Ingress

Annotations for Google Cloud Load Balancer:

```go
var WebIngress = networkingv1.Ingress{
    Metadata: metav1.ObjectMeta{
        Name: "web",
        Annotations: map[string]string{
            "kubernetes.io/ingress.class":                 "gce",
            "kubernetes.io/ingress.global-static-ip-name": "web-ip",
            "networking.gke.io/managed-certificates":      "web-cert",
        },
    },
    Spec: networkingv1.IngressSpec{
        Rules: []networkingv1.IngressRule{
            {
                Host: "app.example.com",
                HTTP: &networkingv1.HTTPIngressRuleValue{
                    Paths: []networkingv1.HTTPIngressPath{
                        {
                            Path:     "/",
                            PathType: networkingv1.PathTypePrefix,
                            Backend: networkingv1.IngressBackend{
                                Service: &networkingv1.IngressServiceBackend{
                                    Name: "web",
                                    Port: networkingv1.ServiceBackendPort{Number: 80},
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

## Config Connector

Manage GCP resources as Kubernetes custom resources.

### Installation

```bash
# Install Config Connector
gcloud container clusters update CLUSTER_NAME \
  --update-addons ConfigConnector=ENABLED

# Or via Helm
helm install configconnector \
  oci://gcr.io/config-connector/charts/config-connector
```

### Cloud Storage Bucket

```go
import (
    storagev1beta1 "github.com/lex00/wetwire-gcp-go/resources/cnrm/storage/v1beta1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DataBucket = storagev1beta1.StorageBucket{
    TypeMeta: metav1.TypeMeta{
        APIVersion: "storage.cnrm.cloud.google.com/v1beta1",
        Kind:       "StorageBucket",
    },
    ObjectMeta: metav1.ObjectMeta{
        Name:      "my-data-bucket",
        Namespace: "config-connector",
        Annotations: map[string]string{
            "cnrm.cloud.google.com/project-id": "my-project",
        },
    },
    Spec: storagev1beta1.StorageBucketSpec{
        Location: "US",
        Versioning: &storagev1beta1.BucketVersioning{
            Enabled: true,
        },
    },
}
```

### Cloud SQL Database

```go
import (
    sqlv1beta1 "github.com/lex00/wetwire-gcp-go/resources/cnrm/sql/v1beta1"
)

var AppDatabase = sqlv1beta1.SQLInstance{
    TypeMeta: metav1.TypeMeta{
        APIVersion: "sql.cnrm.cloud.google.com/v1beta1",
        Kind:       "SQLInstance",
    },
    ObjectMeta: metav1.ObjectMeta{
        Name:      "app-db",
        Namespace: "config-connector",
    },
    Spec: sqlv1beta1.SQLInstanceSpec{
        DatabaseVersion: "POSTGRES_15",
        Region:          "us-central1",
        Settings: sqlv1beta1.InstanceSettings{
            Tier: "db-f1-micro",
        },
    },
}

var AppDatabaseDB = sqlv1beta1.SQLDatabase{
    TypeMeta: metav1.TypeMeta{
        APIVersion: "sql.cnrm.cloud.google.com/v1beta1",
        Kind:       "SQLDatabase",
    },
    ObjectMeta: metav1.ObjectMeta{
        Name:      "app",
        Namespace: "config-connector",
    },
    Spec: sqlv1beta1.SQLDatabaseSpec{
        InstanceRef: sqlv1beta1.InstanceRef{
            Name: "app-db",
        },
    },
}
```

### Pub/Sub Topic and Subscription

```go
import (
    pubsubv1beta1 "github.com/lex00/wetwire-gcp-go/resources/cnrm/pubsub/v1beta1"
)

var EventsTopic = pubsubv1beta1.PubSubTopic{
    TypeMeta: metav1.TypeMeta{
        APIVersion: "pubsub.cnrm.cloud.google.com/v1beta1",
        Kind:       "PubSubTopic",
    },
    ObjectMeta: metav1.ObjectMeta{
        Name:      "events",
        Namespace: "config-connector",
    },
}

var EventsSubscription = pubsubv1beta1.PubSubSubscription{
    TypeMeta: metav1.TypeMeta{
        APIVersion: "pubsub.cnrm.cloud.google.com/v1beta1",
        Kind:       "PubSubSubscription",
    },
    ObjectMeta: metav1.ObjectMeta{
        Name:      "events-sub",
        Namespace: "config-connector",
    },
    Spec: pubsubv1beta1.PubSubSubscriptionSpec{
        TopicRef: pubsubv1beta1.TopicRef{
            Name: "events",
        },
        AckDeadlineSeconds: 20,
    },
}
```

## GKE Cluster via Deployment Manager

For the GKE cluster itself, use wetwire-gcp:

```go
// In wetwire-gcp project
import (
    "github.com/lex00/wetwire-gcp-go/resources/container"
)

var Cluster = container.Cluster{
    Name:             "production",
    Location:         "us-central1",
    InitialNodeCount: 3,
    NodeConfig: container.NodeConfig{
        MachineType: "e2-medium",
        OauthScopes: []string{
            "https://www.googleapis.com/auth/cloud-platform",
        },
    },
}
```

## Multi-Cloud Scenario: GCP + K8s

The [multi-cloud-k8s-krm](https://github.com/lex00/wetwire/tree/main/scenarios/multi-cloud-k8s-krm) scenario demonstrates:

1. GKE cluster provisioning via Deployment Manager
2. GCP resources (GCS, Cloud SQL, Pub/Sub) via Config Connector
3. Application workloads via standard K8s manifests
4. Cross-resource references using Config Connector's status fields

## See Also

- [Scenarios]({{< relref "/scenarios" >}}) - Multi-cloud scenario overview
- [wetwire-gcp-go](https://lex00.github.io/wetwire-gcp-go/) - GCP Deployment Manager synthesis
- [Config Connector Documentation](https://cloud.google.com/config-connector/docs/overview)
