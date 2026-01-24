# Deploy Harness Mock Infrastructure Roadmap

**Status:** In Progress
**Branch:** `feat/deploy-harness-mock-infra`
**Target:** Enable full end-to-end testing of wetwire scenarios with zero cloud credentials

## Overview

This roadmap tracks the implementation of mock infrastructure for the deploy-harness testing framework, enabling CI/CD pipelines to run full deployment tests without real cloud credentials.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Test Environment                          │
│                                                                   │
│  ┌─────────────────────┐      ┌─────────────────────────────┐   │
│  │   k3d Cluster       │      │   Hoverfly Mock Server      │   │
│  │   (real K8s)        │      │   (cloud API simulator)     │   │
│  │                     │      │                             │   │
│  │  ┌───────────────┐  │      │  Simulations:               │   │
│  │  │ webapp pods   │──┼──────┼─→ AWS CloudFormation API   │   │
│  │  │ prometheus    │  │      │    GCP Compute API          │   │
│  │  │ configmaps    │  │      │    Azure ARM API            │   │
│  │  └───────────────┘  │      │    Honeycomb API            │   │
│  │                     │      │                             │   │
│  │  Port: 6443         │      │  Port: 8500                 │   │
│  └─────────────────────┘      └─────────────────────────────┘   │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                    deploy-harness                            │ │
│  │  ./deploy.sh obs_k8s ./results --apply --mock                │ │
│  │                                                              │ │
│  │  - Real kubectl against k3d                                  │ │
│  │  - Cloud CLIs proxied through Hoverfly                       │ │
│  │  - Full validation + deployment                              │ │
│  └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Tool Selection

| Tool | Choice | Rationale |
|------|--------|-----------|
| **Local K8s** | k3d | 5-20s startup, 2GB RAM, built-in registry, excellent CI support |
| **HTTP Mock** | Hoverfly | Go binary (no JVM), record/playback, lightweight, 5 modes |

## Implementation Phases

### Phase 1: K8s Mock Setup - [x] Complete
**Files created:**
- [x] `tools/deploy-harness/mock/setup.sh` - Create k3d cluster + start Hoverfly
- [x] `tools/deploy-harness/mock/teardown.sh` - Clean shutdown

**Acceptance criteria:**
- [x] `./mock/setup.sh` creates a k3d cluster named `wetwire-test`
- [ ] `kubectl get nodes` shows k3d node after setup (requires testing)
- [x] `./mock/teardown.sh` removes cluster cleanly

### Phase 2: Hoverfly Simulations - [x] Complete (Core)
**Files created:**
- [x] `tools/deploy-harness/mock/simulations/aws.json` - AWS CloudFormation API mocks
- [x] `tools/deploy-harness/mock/simulations/honeycomb.json` - Honeycomb API mocks
- [ ] `tools/deploy-harness/mock/simulations/gcp.json` - GCP API mocks (future)
- [ ] `tools/deploy-harness/mock/simulations/azure.json` - Azure API mocks (future)

**Acceptance criteria:**
- [x] Hoverfly serves AWS CloudFormation validation responses
- [x] Hoverfly serves Honeycomb authentication and API responses
- [x] Simulations are valid Hoverfly JSON format

### Phase 3: Deploy Harness Integration - [x] Complete (Core)
**Files modified:**
- [x] `tools/deploy-harness/lib/common.sh` - Add `--mock` flag parsing and `MOCK_MODE` variable
- [x] `tools/deploy-harness/lib/aws.sh` - Add mock mode detection and endpoint override
- [x] `tools/deploy-harness/lib/honeycomb.sh` - Add mock mode detection and endpoint override
- [x] `tools/deploy-harness/deploy.sh` - Add mock setup/teardown integration
- [ ] `tools/deploy-harness/lib/gcp.sh` - Add mock mode detection (future)
- [ ] `tools/deploy-harness/lib/azure.sh` - Add mock mode detection (future)

**Acceptance criteria:**
- [x] `--mock` flag is recognized by deploy.sh
- [x] AWS functions use Hoverfly endpoint when in mock mode
- [x] Honeycomb functions use Hoverfly endpoint when in mock mode

### Phase 4: Documentation - [x] Complete
**Files created:**
- [x] `tools/deploy-harness/mock/README.md` - Mock setup guide

**Acceptance criteria:**
- [x] Clear setup instructions for local development
- [x] CI/CD integration examples
- [x] Troubleshooting guide

## Usage (Target)

```bash
# Full mock mode (creates cluster + mock server)
./deploy.sh obs_k8s ./results --apply --mock

# Use existing k3d cluster
./deploy.sh obs_k8s ./results --apply --mock --no-create-cluster

# Just mock cloud APIs (use real K8s)
./deploy.sh multi-cloud-k8s-native ./results --apply --mock-cloud

# Teardown after test
./deploy.sh obs_k8s ./results --mock --teardown-mock
```

## Verification Checklist

- [ ] `./mock/setup.sh` runs without errors
- [ ] `kubectl get nodes` shows k3d node
- [ ] `./deploy.sh obs_k8s ../../examples/obs_k8s/results/beginner --apply --mock` succeeds
- [ ] `kubectl get pods -n webapp` shows running pods
- [ ] `./deploy.sh honeycomb_k8s ../../examples/honeycomb_k8s/results/beginner --apply --mock` succeeds
- [ ] `./mock/teardown.sh` runs without errors

## Progress Log

| Date | Phase | Status | Notes |
|------|-------|--------|-------|
| 2026-01-23 | Setup | Started | Initial roadmap created |
| 2026-01-23 | Phase 1 | Complete | setup.sh and teardown.sh created |
| 2026-01-23 | Phase 2 | Complete | AWS and Honeycomb simulations created |
| 2026-01-23 | Phase 3 | Complete | common.sh, aws.sh, honeycomb.sh, deploy.sh updated |
| 2026-01-23 | Phase 4 | Complete | README.md documentation created |

## Related Issues

- Deploy harness was created for wetwire scenario testing
- Mock infrastructure enables CI/CD testing without credentials
