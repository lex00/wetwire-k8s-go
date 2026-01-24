---
title: "Scenarios"
---

Multi-cloud Kubernetes scenarios demonstrate wetwire's cross-domain capabilities, generating infrastructure across cloud providers and Kubernetes in a single workflow.

## Multi-Cloud Scenarios

| Scenario | Domains | Description |
|----------|---------|-------------|
| [multi-cloud-k8s-krm](https://github.com/lex00/wetwire/tree/main/scenarios/multi-cloud-k8s-krm) | AWS + GCP + Azure + K8s | Kubernetes Resource Model (KRM) approach using ACK, Config Connector, ASO |
| [multi-cloud-k8s-native](https://github.com/lex00/wetwire/tree/main/scenarios/multi-cloud-k8s-native) | AWS + GCP + Azure + K8s | Native cloud CLI approach with kubectl |

## KRM vs Native Approaches

### KRM (Kubernetes Resource Model)

The KRM approach manages cloud resources through Kubernetes CRDs:

| Cloud | Controller | CRD Example |
|-------|------------|-------------|
| AWS | [ACK](https://aws-controllers-k8s.github.io/community/) | `s3.services.k8s.aws/Bucket` |
| GCP | [Config Connector](https://cloud.google.com/config-connector/docs/overview) | `storage.cnrm.cloud.google.com/StorageBucket` |
| Azure | [ASO](https://azure.github.io/azure-service-operator/) | `storage.azure.com/StorageAccount` |

**Benefits:**
- GitOps-native: all resources in Git, applied via kubectl
- Unified API: same workflow for cloud and K8s resources
- Drift detection: controllers reconcile desired state

**Trade-offs:**
- Requires controller installation in cluster
- Some cloud features may lag behind native APIs

### Native Cloud CLIs

The native approach uses cloud-specific CLIs alongside kubectl:

```bash
# AWS resources via CloudFormation
wetwire-aws build ./aws | aws cloudformation deploy ...

# GCP resources via Deployment Manager
wetwire-gcp build ./gcp | gcloud deployment-manager ...

# K8s resources via kubectl
wetwire-k8s build ./k8s | kubectl apply -f -
```

**Benefits:**
- Full feature coverage from day one
- No controller dependencies
- Familiar tooling for cloud teams

**Trade-offs:**
- Multiple tools to orchestrate
- State management across systems

## Scenario Structure

```
multi-cloud-k8s-krm/
├── scenario.yaml       # Configuration and validation
├── system_prompt.md    # Domain knowledge for AI
├── prompts/
│   ├── beginner.md     # "I need a web app on multiple clouds"
│   ├── intermediate.md # "Deploy to EKS and GKE with shared config"
│   └── expert.md       # "Multi-region active-active with failover"
├── expected/
│   ├── aws/           # Expected ACK resources
│   ├── gcp/           # Expected Config Connector resources
│   ├── azure/         # Expected ASO resources
│   └── k8s/           # Expected K8s workloads
└── results/           # Generated outputs per persona
```

## Running Scenarios

```bash
# Run with wetwire-core
wetwire-core run ./scenarios/multi-cloud-k8s-krm beginner

# Run all personas
wetwire-core run ./scenarios/multi-cloud-k8s-krm --all

# Validate outputs
wetwire-core validate ./scenarios/multi-cloud-k8s-krm ./results
```

## Cloud-Specific Integration

Each cloud provider has unique Kubernetes integration patterns:

| Cloud | Guide | Key Topics |
|-------|-------|------------|
| AWS | [AWS Integration]({{< relref "/aws" >}}) | EKS, ACK, IAM Roles for Service Accounts |
| GCP | [GCP Integration]({{< relref "/gcp" >}}) | GKE, Config Connector, Workload Identity |
| Azure | [Azure Integration]({{< relref "/azure" >}}) | AKS, ASO, Azure AD Pod Identity |

## See Also

- [Examples]({{< relref "/examples" >}}) - Single-domain K8s examples
- [Quick Start]({{< relref "/quick-start" >}}) - Getting started with wetwire-k8s
- [Central Scenarios](https://lex00.github.io/wetwire/scenarios/) - All multi-domain scenarios
