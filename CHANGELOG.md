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

### Fixed

- CI workflow now conditionally runs round-trip tests based on directory existence
