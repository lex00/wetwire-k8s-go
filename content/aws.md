---
title: "AWS Integration"
slug: "aws"
---

Kubernetes integration patterns for Amazon Web Services, covering EKS cluster management and AWS resource provisioning via ACK.

## Overview

| Approach | Tool | Use Case |
|----------|------|----------|
| **EKS Manifests** | wetwire-k8s | K8s workloads deployed to EKS |
| **ACK CRDs** | wetwire-k8s | AWS resources managed via kubectl |
| **CloudFormation** | wetwire-aws | EKS cluster and VPC infrastructure |

## EKS Workload Patterns

### IAM Roles for Service Accounts (IRSA)

Associate Kubernetes service accounts with IAM roles for secure AWS API access:

```go
// Service account with IAM role annotation
var S3AccessSA = corev1.ServiceAccount{
    Metadata: metav1.ObjectMeta{
        Name:      "s3-access",
        Namespace: "default",
        Annotations: map[string]string{
            "eks.amazonaws.com/role-arn": "arn:aws:iam::123456789:role/S3AccessRole",
        },
    },
}

// Pod using the service account
var DataProcessor = appsv1.Deployment{
    Spec: appsv1.DeploymentSpec{
        Template: corev1.PodTemplateSpec{
            Spec: corev1.PodSpec{
                ServiceAccountName: "s3-access",
                Containers: []corev1.Container{
                    {
                        Name:  "processor",
                        Image: "myapp:latest",
                        // AWS SDK automatically uses IRSA credentials
                    },
                },
            },
        },
    },
}
```

### AWS Load Balancer Controller

Annotations for ALB/NLB integration:

```go
var WebIngress = networkingv1.Ingress{
    Metadata: metav1.ObjectMeta{
        Name: "web",
        Annotations: map[string]string{
            "kubernetes.io/ingress.class":               "alb",
            "alb.ingress.kubernetes.io/scheme":          "internet-facing",
            "alb.ingress.kubernetes.io/target-type":     "ip",
            "alb.ingress.kubernetes.io/certificate-arn": "arn:aws:acm:...",
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

## ACK (AWS Controllers for Kubernetes)

Manage AWS resources as Kubernetes custom resources.

### Installation

```bash
# Install ACK S3 controller
helm install ack-s3-controller \
  oci://public.ecr.aws/aws-controllers-k8s/s3-chart \
  --namespace ack-system
```

### S3 Bucket via ACK

```go
import (
    s3v1alpha1 "github.com/lex00/wetwire-aws-go/resources/k8s/s3/v1alpha1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DataBucket = s3v1alpha1.Bucket{
    TypeMeta: metav1.TypeMeta{
        APIVersion: "s3.services.k8s.aws/v1alpha1",
        Kind:       "Bucket",
    },
    ObjectMeta: metav1.ObjectMeta{
        Name:      "my-data-bucket",
        Namespace: "ack-system",
    },
    Spec: s3v1alpha1.BucketSpec{
        Name: "my-company-data-bucket",
        Versioning: &s3v1alpha1.VersioningConfiguration{
            Status: "Enabled",
        },
    },
}
```

### RDS Database via ACK

```go
import (
    rdsv1alpha1 "github.com/lex00/wetwire-aws-go/resources/k8s/rds/v1alpha1"
)

var AppDatabase = rdsv1alpha1.DBInstance{
    TypeMeta: metav1.TypeMeta{
        APIVersion: "rds.services.k8s.aws/v1alpha1",
        Kind:       "DBInstance",
    },
    ObjectMeta: metav1.ObjectMeta{
        Name:      "app-db",
        Namespace: "ack-system",
    },
    Spec: rdsv1alpha1.DBInstanceSpec{
        DBInstanceIdentifier: "app-database",
        DBInstanceClass:      "db.t3.micro",
        Engine:               "postgres",
        EngineVersion:        "15",
        MasterUsername:       "admin",
        MasterUserPassword: &rdsv1alpha1.SecretKeyReference{
            SecretKeyRef: &corev1.SecretKeySelector{
                LocalObjectReference: corev1.LocalObjectReference{
                    Name: "db-credentials",
                },
                Key: "password",
            },
        },
    },
}
```

## EKS Cluster via CloudFormation

For the EKS cluster itself, use wetwire-aws:

```go
// In wetwire-aws project
import (
    "github.com/lex00/wetwire-aws-go/resources/eks"
)

var Cluster = eks.Cluster{
    Name:    "production",
    Version: "1.28",
    RoleArn: ClusterRole.Arn,
    ResourcesVpcConfig: eks.ResourcesVpcConfig{
        SubnetIds:        []string{PrivateSubnet1.Ref, PrivateSubnet2.Ref},
        SecurityGroupIds: []string{ClusterSG.Ref},
    },
}
```

## Multi-Cloud Scenario: AWS + K8s

The [multi-cloud-k8s-krm](https://github.com/lex00/wetwire/tree/main/scenarios/multi-cloud-k8s-krm) scenario demonstrates:

1. EKS cluster provisioning via CloudFormation
2. AWS resources (S3, RDS) via ACK CRDs
3. Application workloads via standard K8s manifests
4. Cross-resource references (app → database connection string)

## See Also

- [Scenarios]({{< relref "/scenarios" >}}) - Multi-cloud scenario overview
- [wetwire-aws-go](https://lex00.github.io/wetwire-aws-go/) - AWS CloudFormation synthesis
- [ACK Documentation](https://aws-controllers-k8s.github.io/community/)
