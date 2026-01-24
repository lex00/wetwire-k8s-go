---
title: "Azure Integration"
slug: "azure"
---

Kubernetes integration patterns for Microsoft Azure, covering AKS cluster management and Azure resource provisioning via ASO.

## Overview

| Approach | Tool | Use Case |
|----------|------|----------|
| **AKS Manifests** | wetwire-k8s | K8s workloads deployed to AKS |
| **ASO CRDs** | wetwire-k8s | Azure resources managed via kubectl |
| **ARM Templates** | wetwire-azure | AKS cluster and network infrastructure |

## AKS Workload Patterns

### Azure AD Workload Identity

Associate Kubernetes service accounts with Azure managed identities:

```go
// Service account with Azure Workload Identity annotations
var BlobAccessSA = corev1.ServiceAccount{
    Metadata: metav1.ObjectMeta{
        Name:      "blob-access",
        Namespace: "default",
        Annotations: map[string]string{
            "azure.workload.identity/client-id": "00000000-0000-0000-0000-000000000000",
        },
        Labels: map[string]string{
            "azure.workload.identity/use": "true",
        },
    },
}

// Pod using the service account
var DataProcessor = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Template: corev1.PodTemplateSpec{
            Metadata: metav1.ObjectMeta{
                Labels: map[string]string{
                    "azure.workload.identity/use": "true",
                },
            },
            Spec: corev1.PodSpec{
                ServiceAccountName: "blob-access",
                Containers: []corev1.Container{
                    {
                        Name:  "processor",
                        Image: "myregistry.azurecr.io/myapp:latest",
                        // Azure SDK automatically uses Workload Identity
                    },
                },
            },
        },
    },
}
```

### Azure Application Gateway Ingress

Annotations for Application Gateway integration:

```go
var WebIngress = networkingv1.Ingress{
    Metadata: metav1.ObjectMeta{
        Name: "web",
        Annotations: map[string]string{
            "kubernetes.io/ingress.class":                       "azure/application-gateway",
            "appgw.ingress.kubernetes.io/ssl-redirect":          "true",
            "appgw.ingress.kubernetes.io/backend-protocol":      "http",
            "appgw.ingress.kubernetes.io/request-timeout":       "30",
        },
    },
    Spec: networkingv1.IngressSpec{
        TLS: []networkingv1.IngressTLS{
            {
                Hosts:      []string{"app.example.com"},
                SecretName: "tls-secret",
            },
        },
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

## ASO (Azure Service Operator)

Manage Azure resources as Kubernetes custom resources.

### Installation

```bash
# Install ASO via Helm
helm repo add aso2 https://raw.githubusercontent.com/Azure/azure-service-operator/main/v2/charts
helm install aso2 aso2/azure-service-operator \
  --namespace azureserviceoperator-system \
  --create-namespace
```

### Storage Account

```go
import (
    storagev1 "github.com/lex00/wetwire-azure-go/resources/k8s/storage/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DataStorage = storagev1.StorageAccount{
    TypeMeta: metav1.TypeMeta{
        APIVersion: "storage.azure.com/v1",
        Kind:       "StorageAccount",
    },
    ObjectMeta: metav1.ObjectMeta{
        Name:      "mydatastorage",
        Namespace: "aso-system",
    },
    Spec: storagev1.StorageAccountSpec{
        Location:          "eastus",
        ResourceGroupName: "my-resource-group",
        Kind:              "StorageV2",
        Sku: storagev1.Sku{
            Name: "Standard_LRS",
        },
    },
}

var DataContainer = storagev1.StorageAccountsBlobServicesContainer{
    TypeMeta: metav1.TypeMeta{
        APIVersion: "storage.azure.com/v1",
        Kind:       "StorageAccountsBlobServicesContainer",
    },
    ObjectMeta: metav1.ObjectMeta{
        Name:      "data",
        Namespace: "aso-system",
    },
    Spec: storagev1.StorageAccountsBlobServicesContainerSpec{
        Owner: storagev1.Owner{
            Name: "mydatastorage",
        },
    },
}
```

### Azure SQL Database

```go
import (
    sqlv1 "github.com/lex00/wetwire-azure-go/resources/k8s/sql/v1"
)

var SQLServer = sqlv1.Server{
    TypeMeta: metav1.TypeMeta{
        APIVersion: "sql.azure.com/v1",
        Kind:       "Server",
    },
    ObjectMeta: metav1.ObjectMeta{
        Name:      "app-sql-server",
        Namespace: "aso-system",
    },
    Spec: sqlv1.ServerSpec{
        Location:                    "eastus",
        ResourceGroupName:           "my-resource-group",
        AdministratorLogin:          "sqladmin",
        AdministratorLoginPassword: &sqlv1.SecretReference{
            Name: "sql-admin-password",
            Key:  "password",
        },
    },
}

var AppDatabase = sqlv1.ServersDatabase{
    TypeMeta: metav1.TypeMeta{
        APIVersion: "sql.azure.com/v1",
        Kind:       "ServersDatabase",
    },
    ObjectMeta: metav1.ObjectMeta{
        Name:      "appdb",
        Namespace: "aso-system",
    },
    Spec: sqlv1.ServersDatabaseSpec{
        Owner: sqlv1.Owner{
            Name: "app-sql-server",
        },
        Sku: sqlv1.DatabaseSku{
            Name: "Basic",
        },
    },
}
```

### Service Bus

```go
import (
    servicebusv1 "github.com/lex00/wetwire-azure-go/resources/k8s/servicebus/v1"
)

var MessageBus = servicebusv1.Namespace{
    TypeMeta: metav1.TypeMeta{
        APIVersion: "servicebus.azure.com/v1",
        Kind:       "Namespace",
    },
    ObjectMeta: metav1.ObjectMeta{
        Name:      "app-servicebus",
        Namespace: "aso-system",
    },
    Spec: servicebusv1.NamespaceSpec{
        Location:          "eastus",
        ResourceGroupName: "my-resource-group",
        Sku: servicebusv1.SBSku{
            Name: "Standard",
        },
    },
}

var EventsQueue = servicebusv1.NamespacesQueue{
    TypeMeta: metav1.TypeMeta{
        APIVersion: "servicebus.azure.com/v1",
        Kind:       "NamespacesQueue",
    },
    ObjectMeta: metav1.ObjectMeta{
        Name:      "events",
        Namespace: "aso-system",
    },
    Spec: servicebusv1.NamespacesQueueSpec{
        Owner: servicebusv1.Owner{
            Name: "app-servicebus",
        },
    },
}
```

## AKS Cluster via ARM Templates

For the AKS cluster itself, use wetwire-azure:

```go
// In wetwire-azure project
import (
    "github.com/lex00/wetwire-azure-go/resources/containerservice"
)

var Cluster = containerservice.ManagedCluster{
    Name:     "production",
    Location: "eastus",
    Properties: containerservice.ManagedClusterProperties{
        KubernetesVersion: "1.28",
        DNSPrefix:         "prod-aks",
        AgentPoolProfiles: []containerservice.AgentPoolProfile{
            {
                Name:   "default",
                Count:  3,
                VMSize: "Standard_D2s_v3",
                Mode:   "System",
            },
        },
        Identity: containerservice.ManagedClusterIdentity{
            Type: "SystemAssigned",
        },
    },
}
```

## Multi-Cloud Scenario: Azure + K8s

The [multi-cloud-k8s-krm](https://github.com/lex00/wetwire/tree/main/scenarios/multi-cloud-k8s-krm) scenario demonstrates:

1. AKS cluster provisioning via ARM templates
2. Azure resources (Storage, SQL, Service Bus) via ASO
3. Application workloads via standard K8s manifests
4. Cross-resource references using ASO's status fields

## See Also

- [Scenarios]({{< relref "/scenarios" >}}) - Multi-cloud scenario overview
- [wetwire-azure-go](https://lex00.github.io/wetwire-azure-go/) - Azure ARM synthesis
- [ASO Documentation](https://azure.github.io/azure-service-operator/)
