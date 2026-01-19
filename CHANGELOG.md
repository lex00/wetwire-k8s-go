# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Production Kubernetes manifests from CNCF projects** (#96)
  - Added `examples/imported/` directory with real-world manifests from Argo CD and kube-prometheus
  - Imported 6 manifests covering different resource types:
    - 2 Deployments (Argo CD Redis, Grafana)
    - 3 Services (Argo CD Server, Grafana, Alertmanager)
    - 1 ServiceAccount (Prometheus)
  - Created comprehensive README.md documenting source projects and usage
  - Demonstrates importer capabilities with production-grade configurations

- **Multi-tier app scenario example** (#97)
  - Added `examples/multitier_scenario/` with complete scenario configuration
  - Includes scenario.yaml, system_prompt.md, and three persona prompts (beginner, intermediate, expert)
  - Gold standard outputs demonstrate multi-tier e-commerce application with:
    - Namespace and ResourceQuota
    - Frontend Deployment and LoadBalancer Service
    - Backend Deployment and ClusterIP Service
    - ConfigMap and Secrets for configuration
    - NetworkPolicy for tier isolation
    - HorizontalPodAutoscaler for frontend scaling
  - Validates generation of at least 6 K8s resources across multiple files

### Changed

- **Split rules_workload.go for maintainability**
  - Extracted WK8302, WK8303, WK8304 (high availability rules) into new `rules_ha.go`
  - Reduced `rules_workload.go` from 703 lines to 299 lines
  - All lint rule files now under 550 lines for better maintainability

- **Adopt wetwire-core-go/ast for type extraction** (#90)
  - Replace local `getResourceType` with `coreast.ExtractTypeName`
  - Replace local `getResourceTypeFromExpr` with `coreast.InferTypeFromValue`
  - Keep K8s-specific `isKubernetesType` check local for domain logic

- **Adopt wetwire-core-go/lint Severity type** (#89)
  - Replace local Severity type with type alias from wetwire-core-go/lint
  - Update severity constants to reference core package
  - Upgrade wetwire-core-go to v1.16.0

- **MCP Migration** (#79)
  - Migrated to auto-generated MCP server using `domain.BuildMCPServer()`
  - Updated `wetwire-core-go` dependency to v1.13.0
  - Replaced manual MCP tool registration with automatic domain-based generation
  - Simplified `cmd/wetwire-k8s/mcp.go` from 381 lines to 47 lines
  - Now supports all standard domain tools: build, lint, validate, list, graph, init

- Updated `wetwire-core-go` dependency to v1.5.4 for Kiro provider cwd fix (#75)
  - Ensures MCP servers run in the correct working directory
  - Added test to verify cwd is set in Kiro agent configuration

### Added

- **LintOpts.Fix support in domain Linter** (#94)
  - Domain `k8sLinter.Lint()` now respects `opts.Fix` option
  - When Fix is true, auto-fixes fixable issues before reporting remaining issues
  - Improved error messages to indicate when auto-fix was attempted
  - Added comprehensive tests for Fix option behavior

### Added

#### Phase 1: Foundation (Issues #1, #2, #16)

- **Schema fetching and code generation** (`codegen/` package)
  - Fetch Kubernetes OpenAPI schemas from kubernetes/kubernetes GitHub
  - Parse schemas to extract resource types, fields, and metadata
  - Generate Go struct definitions organized by API group (resources/apps/v1/, etc.)
  - Support for Kubernetes 1.28+ schemas with local caching
  - Comprehensive test coverage

- **AST-based resource discovery** (`internal/discover/` package)
  - Discover Kubernetes resources from Go source files via AST parsing
  - Support for both single file and directory scanning
  - Track dependencies between resources via field references
  - No code execution required - pure static analysis

- **Required documentation**
  - `CLAUDE.md` - AI assistant context and repository guide
  - `docs/CLI.md` - Complete command reference for all 13 planned commands
  - `docs/FAQ.md` - Common questions and troubleshooting
  - `docs/LINT_RULES.md` - All 24 lint rules with examples
  - Updated `README.md` with badges and enhanced quick start guide

#### Phase 2: Core Pipeline (Issues #3, #4, #6)

- **Serialization** (`internal/serialize/` package)
  - Convert Go structs to Kubernetes-compliant YAML and JSON
  - Automatic camelCase field conversion for Kubernetes compatibility
  - Zero-value removal for cleaner output
  - Multi-resource serialization with YAML document separators
  - Support for all core Kubernetes resource types

- **Template builder and build pipeline** (`internal/build/` package)
  - 6-stage build pipeline: DISCOVER, VALIDATE, EXTRACT, ORDER, SERIALIZE, EMIT
  - Topological sorting using Kahn's algorithm for correct apply order
  - DFS-based circular dependency detection
  - Reference validation to catch undefined dependencies
  - Configurable output directory and serialization options

- **Lint rules and lint engine** (`internal/lint/` package)
  - Flexible AST-based lint engine for static analysis
  - 6 core rules implemented:
    - WK8001: Resources must be top-level variable declarations
    - WK8002: Avoid deeply nested inline structures (max depth 5)
    - WK8003: No duplicate resource names in same namespace
    - WK8004: Circular dependency detection
    - WK8005: Flag hardcoded secrets in environment variables
    - WK8006: Flag :latest image tags
  - Multiple output formats (text, JSON, GitHub Actions)
  - Configurable rule disabling and minimum severity filtering

#### Phase 4: Validation & Testing (Issues #8, #9, #15)

- **Validate CLI command** (`cmd/wetwire-k8s/validate.go`)
  - `wetwire-k8s validate [PATH]` - Validate Kubernetes manifests
  - Kubeconform integration for schema validation
  - Flags: `--schema-location`, `--strict`, `--output` (text/json/tap/junit), `--kubernetes-version`
  - `--from-build` flag to validate manifests from build pipeline
  - Comprehensive test coverage with 14 test cases

- **Utility CLI commands** (`cmd/wetwire-k8s/`)
  - `wetwire-k8s list [PATH]` - Discover and list K8s resources
    - Supports `--format` (table/json/yaml) and `--all` for dependencies
  - `wetwire-k8s init [PATH]` - Initialize new wetwire-k8s projects
    - Creates k8s/ directory with namespace.go template
    - Generates .wetwire.yaml configuration
    - `--example` flag for deployment/service templates
  - `wetwire-k8s graph [PATH]` - Visualize resource dependencies
    - ASCII tree format (default) and DOT format for Graphviz
    - `--output` flag to save to file

- **Round-trip testing** (`internal/roundtrip/` package)
  - YAML -> Go code -> YAML semantic equivalence testing
  - Multi-document YAML support
  - Normalization and comparison functions
  - Reference YAML files in `testdata/roundtrip/examples/`
  - Integration tests with kubernetes/examples patterns
  - Performance benchmarks for parsing and comparison

#### Phase 5: Developer Experience (Issues #10, #13, #17)

- **Diff CLI command** (`cmd/wetwire-k8s/diff.go`)
  - `wetwire-k8s diff [PATH] --against <manifest>` - Compare generated vs existing
  - Text diff mode (line-by-line) and semantic diff mode
  - Flags: `--semantic`, `--color`, `--output`
  - Uses internal/roundtrip for semantic YAML comparison

- **Watch CLI command** (`cmd/wetwire-k8s/watch.go`)
  - `wetwire-k8s watch [PATH]` - Monitor source files for changes
  - Auto-rebuild on change with debouncing via fsnotify
  - Flags: `--output`, `--interval`
  - Comprehensive test coverage

- **MCP server for Claude Code integration** (`cmd/wetwire-k8s-mcp/`)
  - Model Context Protocol server using github.com/mark3labs/mcp-go
  - Tools: build, lint, import, validate
  - JSON-RPC 2.0 communication via stdio
  - Full handler implementations for all tools

- **Recommended documentation** (`docs/`)
  - `docs/QUICK_START.md` - 5-minute getting started guide
  - `docs/INTERNALS.md` - Architecture deep-dive (build pipeline)
  - `docs/TROUBLESHOOTING.md` - Common issues and solutions
  - `docs/CONTRIBUTING.md` - Development guide

- **Working examples** (`examples/`)
  - `examples/guestbook/` - Multi-tier web app with Redis backend
  - `examples/web-service/` - Deployment + Service + Ingress with production features
  - `examples/configmap-secret/` - ConfigMap and Secret usage patterns
  - All examples compile, follow wetwire pattern, include README

#### Phase 6: Agent Integration (Issues #11, #12)

- **Design CLI command** (`cmd/wetwire-k8s/design.go`)
  - `wetwire-k8s design --prompt "description"` - AI-assisted K8s generation
  - Provider selection: `--provider anthropic|kiro`
  - K8sRunnerAgent with domain-specific tools:
    - `init_package`: Create package directories
    - `write_file`: Generate Go files with lint state tracking
    - `read_file`: Read existing files
    - `run_lint`: Run wetwire-k8s linter (JSON output)
    - `run_build`: Build Kubernetes YAML manifests
    - `ask_developer`: Clarifying questions
  - Lint enforcement per Wetwire Spec 6.3 (lint-after-write rule)
  - Completion gate checks (lint called, passed, no pending)
  - K8s-specific system prompt with wetwire patterns

- **Test CLI command** (`cmd/wetwire-k8s/test.go`)
  - `wetwire-k8s test --persona <name>` - Persona-based testing
  - 5 standard personas: beginner, intermediate, expert, terse, verbose
  - 5 scoring dimensions: Completeness, Lint Quality, Code Quality, Output Validity, Question Efficiency
  - Flags: `--persona`, `--provider`, `--all-personas`, `--scenario`, `--mock`, `--dry-run`, `--verbose`
  - Output files: RESULTS.md, session.json, score.json
  - Integration with wetwire-core-go personas/scoring/results packages

#### Phase 3: CLI Commands (Issues #5, #7, #14)

- **Build CLI command** (`cmd/wetwire-k8s/build.go`)
  - `wetwire-k8s build [PATH]` - Generate Kubernetes YAML/JSON from Go code
  - Flags: `--output/-o`, `--format/-f` (yaml/json), `--dry-run`
  - API version mapping for all core Kubernetes resource types
  - Multi-resource output with YAML document separators
  - Integration with build pipeline for dependency ordering
  - Built on urfave/cli/v2 framework

- **Import CLI command** (`cmd/wetwire-k8s/import.go`)
  - `wetwire-k8s import <file>` - Convert YAML manifests to Go code
  - Flags: `--output/-o`, `--package/-p`, `--var-prefix`
  - Supports multi-document YAML files
  - Generates idiomatic Go using k8s.io/api types
  - Proper variable naming from resource metadata
  - stdin support with `-` as file argument

- **Internal importer package** (`internal/importer/`)
  - Parse YAML manifests to structured resource info
  - Generate Go code with proper imports and type references
  - Map Kubernetes apiVersion/kind to Go package/type
  - Support for ConfigMap, Deployment, Service, Namespace, and more

- **Comprehensive unit test coverage** (Issue #14)
  - `codegen/fetch/`: Error handling, context cancellation tests
  - `codegen/parse/`: Property type tests, helper function tests
  - `internal/lint/`: Rule disabling, severity filtering tests
  - `internal/serialize/`: Resource serialization, edge case tests
  - Increased coverage across all packages

#### Phase 7: Spec Compliance (Issues #36, #37, #38, #39, #45, #46, #47)

- **Lint CLI command** (`cmd/wetwire-k8s/lint.go`) (Issue #36)
  - `wetwire-k8s lint [PATH]` - Lint Go files with K8s resources
  - `--fix` flag for auto-fixing violations (fully implemented)
  - `--format` flag supporting text, json, and github output formats
  - `--severity` flag to filter by error, warning, or info levels
  - `--disable` flag to disable specific rules (comma-separated)
  - Fixed severity filtering bug in internal/lint/lint.go

- **Recommended documentation** (Issue #38)
  - `docs/IMPORT_WORKFLOW.md` - Step-by-step import process
  - `docs/CODEGEN.md` - Code generation pipeline explanation
  - `docs/ADOPTION.md` - Migration guide from Helm/kustomize/kubectl
  - `docs/DEVELOPERS.md` - Development environment and contribution guide
  - `docs/EXAMPLES.md` - Detailed walkthrough of all examples
  - `docs/VERSIONING.md` - Version compatibility and upgrade paths

- **Additional lint rules** (Issues #37, #45)
  - WK8041: Detect hardcoded API keys/tokens in environment variables
  - WK8042: Detect private key headers in secret data
  - WK8101: Selector label mismatch between Deployment and Service
  - WK8102: Missing recommended metadata labels (app, version)
  - WK8103: Container name required (all containers must have a Name field)
  - WK8104: Port name recommended (Container and Service ports should be named)
  - WK8105: ImagePullPolicy explicit (should be explicitly set)
  - WK8201: Missing resource limits on containers
  - WK8202: Privileged containers detection
  - WK8203: ReadOnlyRootFilesystem (containers should use read-only root filesystem)
  - WK8204: RunAsNonRoot (containers should run as non-root)
  - WK8205: Drop capabilities (containers should drop unnecessary Linux capabilities)
  - WK8207: No host network (pods should not use HostNetwork: true)
  - WK8208: No host PID (pods should not use HostPID: true)
  - WK8209: No host IPC (pods should not use HostIPC: true)
  - WK8301: Missing health probes (liveness/readiness)
  - WK8302: Replicas minimum (deployments should have at least 2 replicas)
  - WK8303: PodDisruptionBudget (HA deployments should have a PDB)
  - WK8304: Anti-affinity recommended (HA deployments should use pod anti-affinity)
  - Total lint rules increased from 6 to 25

- **Test coverage improvements** (Issue #39)
  - Coverage increased from 65.1% to 92.1%
  - Fixed skipped tests in internal/lint/lint_test.go
  - Added comprehensive formatter tests (internal/lint/formatter_test.go)
  - Cross-platform path handling with filepath.Join
  - Updated docs/LINT_RULES.md with all new rules and examples

- **Auto-fix functionality for lint rules** (`internal/lint/fixer.go`) (Issue #46)
  - New Fixer type that handles automatic fixing of lint issues
  - AST manipulation using Go's go/ast, go/parser, and go/printer packages
  - Auto-fix for WK8105 (ImagePullPolicy): Automatically sets policy based on image tag
  - Auto-fix for WK8002 (deeply nested structures): Extracts nested structures to variables
  - Reports which files were modified and what fixes were applied
  - Comprehensive test coverage in fixer_test.go

- **Coverage badge** (Issue #47)
  - Added Codecov coverage badge to README.md

#### Phase 8: Real AI Provider Integration (Issue #51)

- **Real AI provider integration in design and test commands** (`cmd/wetwire-k8s/`)
  - `RunAgentLoop()` function for full agent execution with LLM
  - `NewAnthropicProvider()` helper using wetwire-core-go/providers/anthropic
  - `GetProviderTools()` method to convert K8sRunnerAgent tools to provider format
  - Design command now executes full agent loop when ANTHROPIC_API_KEY is set
  - Test command uses real LLM calls when provider is not "mock"
  - `PersonaDeveloper` type for AI-driven persona responses in test sessions
  - Comprehensive tests for real provider integration (skip when no API key)

- **Comprehensive test fixtures** (`testdata/comprehensive/`) (Issue #53)
  - 35 YAML fixtures covering all major Kubernetes resource types
  - Categories: pods (6), workloads (5), services (5), config (3), storage (3), networking (2), batch (2), rbac (5), autoscaling (4)
  - Round-trip test coverage for all fixture categories
  - SOURCES.md with attribution for test fixture patterns
  - New tests: TestComprehensiveYAMLFixtures, TestComprehensiveFixtureCategories, TestComprehensiveFixtureCount

- **File size lint rule** (`internal/lint/rules.go`) (Issue #57)
  - WK8401: Files should not exceed 20 resources
  - Encourages modular code organization per wetwire spec Section 9.2
  - Total lint rules increased from 25 to 26

### Fixed

- CI workflow now conditionally runs round-trip tests based on directory existence
- CI workflow excludes cmd packages from test coverage
- CI workflow excludes examples from test coverage

## [1.5.0] - 2026-01-19

### Changed

- **Claude CLI as default provider for design command**
  - No API key required - uses existing Claude authentication
  - Falls back to Anthropic API if Claude CLI not installed
  - Updated wetwire-core-go to v1.17.1
