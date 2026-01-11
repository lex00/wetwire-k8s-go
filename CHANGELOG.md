# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

### Fixed

- CI workflow now conditionally runs round-trip tests based on directory existence
- CI workflow excludes cmd packages from test coverage
