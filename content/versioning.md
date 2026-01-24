---
title: "Versioning"
---


This document describes wetwire-k8s-go's versioning policy, Kubernetes compatibility, and upgrade procedures.

## Versioning Policy

wetwire-k8s-go follows [Semantic Versioning 2.0.0](https://semver.org/).

### Version Format

```
MAJOR.MINOR.PATCH
```

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality in backwards-compatible manner
- **PATCH**: Backwards-compatible bug fixes

### Pre-Release Versions

Development versions use pre-release identifiers:

```
0.1.0-alpha.1
0.1.0-beta.1
0.1.0-rc.1
1.0.0
```

- **alpha**: Early development, unstable API
- **beta**: Feature complete, API stabilizing
- **rc**: Release candidate, production-ready testing

## Kubernetes Version Compatibility

wetwire-k8s-go maintains compatibility with multiple Kubernetes versions simultaneously.

### Supported Kubernetes Versions

| wetwire-k8s-go Version | Kubernetes Versions | k8s.io/api Version | Go Version |
|------------------------|---------------------|-------------------|------------|
| 0.1.x | 1.28, 1.29, 1.30 | v0.28.x - v0.30.x | 1.21+ |
| 0.2.x | 1.29, 1.30, 1.31 | v0.29.x - v0.31.x | 1.21+ |
| 1.0.x | 1.30, 1.31, 1.32 | v0.30.x - v0.32.x | 1.22+ |

### Minimum Supported Version

wetwire-k8s-go requires **Kubernetes 1.28+**.

Older versions may work but are not tested or supported.

### Version Selection

Choose your k8s.io/api version based on your target cluster:

```bash
# For Kubernetes 1.28
go get k8s.io/api@v0.28.0
go get k8s.io/apimachinery@v0.28.0

# For Kubernetes 1.29
go get k8s.io/api@v0.29.0
go get k8s.io/apimachinery@v0.29.0

# For Kubernetes 1.30
go get k8s.io/api@v0.30.0
go get k8s.io/apimachinery@v0.30.0
```

### Compatibility Rules

1. **Forward Compatibility**: Code using older API versions works with newer Kubernetes clusters
2. **Backward Compatibility**: Code using newer API versions may not work with older clusters
3. **Field Additions**: New optional fields can be added in minor releases
4. **Field Removals**: Deprecated fields are removed only in major releases

## API Stability Guarantees

### Stable (v1) APIs

Stable Kubernetes APIs (`apps/v1`, `core/v1`, etc.) are guaranteed stable:

- Fields won't be removed without deprecation period
- Types won't change incompatibly
- Defaults won't change behavior

### Beta APIs

Beta APIs (`autoscaling/v2beta2`, etc.) may change:

- Fields can be added, removed, or changed
- Consider using stable versions when available

### Alpha APIs

Alpha APIs are experimental:

- No compatibility guarantees
- May be removed without notice
- Use only for testing

wetwire-k8s-go defaults to stable APIs. Beta/alpha APIs are opt-in.

## Breaking Change Policy

### What Constitutes a Breaking Change

Breaking changes require a major version bump:

1. **Removal of CLI commands** (e.g., removing `wetwire-k8s build`)
2. **Removal of CLI flags** (e.g., removing `--output`)
3. **Change in default behavior** (e.g., different output format by default)
4. **Removal of public Go APIs** (if exposing library functions)
5. **Change in generated output** that breaks existing workflows
6. **Incompatible changes to configuration file format**

### What Is NOT a Breaking Change

Minor version changes allow:

1. **Adding new CLI commands**
2. **Adding new CLI flags**
3. **Adding new lint rules** (with info/warning severity)
4. **Improving error messages**
5. **Performance improvements**
6. **Bug fixes** (even if output changes)
7. **Internal refactoring**

### Deprecation Policy

Before removing features:

1. **Deprecation Notice**: Mark as deprecated in release notes
2. **Deprecation Period**: Minimum 6 months or 2 minor versions
3. **Migration Guide**: Provide alternative approach
4. **Warnings**: CLI emits deprecation warnings
5. **Removal**: Only in next major version

## Version Compatibility Matrix

### CLI Command Compatibility

| Command | v0.1.x | v0.2.x | v1.0.x |
|---------|--------|--------|--------|
| build | Yes | Yes | Yes |
| lint | Yes | Yes | Yes |
| import | Yes | Yes | Yes |
| validate | Yes | Yes | Yes |
| list | Yes | Yes | Yes |
| graph | - | Yes | Yes |
| diff | - | - | Yes |
| watch | - | - | Yes |

### Output Format Compatibility

Generated YAML/JSON maintains backwards compatibility:

- **Field names**: Never change in stable APIs
- **Field order**: May change (not significant in YAML)
- **Zero values**: May be omitted or included (semantically equivalent)

## Upgrade Procedures

### Upgrading wetwire-k8s-go

#### Minor Version Upgrade (e.g., 0.1.0 to 0.2.0)

```bash
# Update CLI
go install github.com/lex00/wetwire-k8s-go/cmd/wetwire-k8s@v0.2.0

# Verify version
wetwire-k8s --version

# Test with existing code
wetwire-k8s build
wetwire-k8s lint
```

**Expected**: No breaking changes, new features available.

#### Major Version Upgrade (e.g., 0.x to 1.0)

```bash
# Read release notes
# https://github.com/lex00/wetwire-k8s-go/releases/tag/v1.0.0

# Update CLI
go install github.com/lex00/wetwire-k8s-go/cmd/wetwire-k8s@v1.0.0

# Check for deprecated features
wetwire-k8s build --help  # Review flag changes

# Test build
wetwire-k8s build -o test.yaml

# Compare output
diff <(wetwire-k8s@v0.x build) <(wetwire-k8s@v1.0 build)

# Update scripts/CI as needed
```

**Expected**: May have breaking changes. Review migration guide.

### Upgrading Kubernetes Dependencies

#### Update k8s.io/api

```bash
# Check current version
go list -m k8s.io/api

# Update to specific version
go get k8s.io/api@v0.30.0
go get k8s.io/apimachinery@v0.30.0

# Update go.mod
go mod tidy

# Test compatibility
go test ./...
wetwire-k8s build
```

#### Handling API Changes

When Kubernetes introduces API changes:

1. **Check release notes**: [Kubernetes Release Notes](https://kubernetes.io/releases/)
2. **Review deprecations**: [Deprecation Guide](https://kubernetes.io/docs/reference/using-api/deprecation-guide/)
3. **Update code**: Migrate from deprecated to new APIs
4. **Test thoroughly**: Run full test suite

Example migration:

```go
// Before (deprecated in Kubernetes 1.25)
import policyv1beta1 "k8s.io/api/policy/v1beta1"

var PDB = &policyv1beta1.PodDisruptionBudget{...}

// After (stable in Kubernetes 1.25+)
import policyv1 "k8s.io/api/policy/v1"

var PDB = &policyv1.PodDisruptionBudget{...}
```

## Release Channels

### Stable Releases

- **Version**: `v1.x.x`, `v2.x.x`
- **Frequency**: As needed (typically monthly)
- **Support**: Full support, bug fixes, security patches
- **Recommended for**: Production use

Install latest stable:

```bash
go install github.com/lex00/wetwire-k8s-go/cmd/wetwire-k8s@latest
```

### Pre-Release Versions

- **Version**: `v1.0.0-rc.1`, `v1.1.0-beta.1`
- **Frequency**: As needed during development
- **Support**: Community support only
- **Recommended for**: Testing, early feedback

Install specific pre-release:

```bash
go install github.com/lex00/wetwire-k8s-go/cmd/wetwire-k8s@v1.0.0-rc.1
```

### Development Builds

- **Version**: `main` branch commits
- **Frequency**: Continuous
- **Support**: None
- **Recommended for**: Development, contributing

Install from main:

```bash
go install github.com/lex00/wetwire-k8s-go/cmd/wetwire-k8s@main
```

## Long-Term Support (LTS)

Currently, wetwire-k8s-go does not have formal LTS releases.

Each major version receives:
- **Bug fixes**: For 12 months after next major release
- **Security patches**: For 18 months after next major release
- **Feature updates**: Only in current major version

## Compatibility Testing

wetwire-k8s-go is tested against:

- **Multiple Kubernetes versions**: 3 most recent minor versions
- **Multiple Go versions**: Current and previous major versions
- **Multiple platforms**: Linux, macOS, Windows
- **Multiple architectures**: amd64, arm64

CI matrix:

```yaml
strategy:
  matrix:
    go-version: [1.21, 1.22]
    k8s-version: [1.28, 1.29, 1.30]
    os: [ubuntu-latest, macos-latest, windows-latest]
```

## Version Detection

### Check CLI Version

```bash
wetwire-k8s --version
```

Output:

```
wetwire-k8s version 1.0.0 (commit abc123, built 2024-01-01)
```

### Environment Variable

Set target Kubernetes version:

```bash
export WETWIRE_K8S_VERSION=1.30
wetwire-k8s build
```

## Best Practices

1. **Pin versions in go.mod**: Don't use `@latest` in production
2. **Test upgrades in staging**: Never upgrade directly in production
3. **Read release notes**: Always review before upgrading
4. **Use stable APIs**: Avoid beta/alpha APIs in production
5. **Keep dependencies updated**: Update k8s.io/api regularly
6. **Monitor deprecations**: Watch for deprecation warnings
7. **Maintain compatibility matrix**: Document tested versions

## References

- [Semantic Versioning](https://semver.org/)
- [Kubernetes API Versioning](https://kubernetes.io/docs/reference/using-api/#api-versioning)
- [Kubernetes Deprecation Policy](https://kubernetes.io/docs/reference/using-api/deprecation-policy/)
- [Go Modules Version Selection](https://go.dev/ref/mod#version-queries)

## Next Steps

- Check [Codegen](/codegen/) for updating k8s.io/api dependencies
- Review [Developers](/developers/) for contributing
- See [CLI](/cli/) for version-specific command reference
