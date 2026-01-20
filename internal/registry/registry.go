// Package registry provides a central registry for Kubernetes resource types.
//
// The registry allows dynamic registration of resource types, enabling support
// for both standard Kubernetes resources and Custom Resource Definitions (CRDs).
//
// Usage:
//
//	// Check if a type is known
//	if registry.DefaultRegistry.IsKnownType("corev1.Pod") {
//	    // ...
//	}
//
//	// Register a custom type
//	registry.DefaultRegistry.Register(registry.TypeInfo{
//	    Package:    "computev1beta1",
//	    Group:      "compute.cnrm.cloud.google.com",
//	    Version:    "v1beta1",
//	    Kind:       "ComputeInstance",
//	    APIVersion: "compute.cnrm.cloud.google.com/v1beta1",
//	    Domain:     "cnrm",
//	})
package registry

import (
	"fmt"
	"strings"
	"sync"
)

// TypeInfo describes a registered Kubernetes resource type.
type TypeInfo struct {
	// Package is the Go package alias (e.g., "corev1", "appsv1", "computev1beta1")
	Package string

	// Group is the API group (e.g., "", "apps", "compute.cnrm.cloud.google.com")
	// Empty string means core API group.
	Group string

	// Version is the API version (e.g., "v1", "v1beta1")
	Version string

	// Kind is the resource kind (e.g., "Pod", "Deployment", "ComputeInstance")
	Kind string

	// APIVersion is the full apiVersion string (e.g., "v1", "apps/v1", "compute.cnrm.cloud.google.com/v1beta1")
	APIVersion string

	// Domain is an optional domain grouping (e.g., "cnrm" for Config Connector, "istio" for Istio)
	Domain string
}

// Registry holds registered type information.
type Registry struct {
	// types maps "pkg.Kind" to TypeInfo
	types map[string]TypeInfo

	// packages maps package names to their group/version
	packages map[string]PackageInfo

	// kinds maps unqualified Kind names to their TypeInfo (for backward compatibility)
	kinds map[string]TypeInfo

	mu sync.RWMutex
}

// PackageInfo holds information about a registered package.
type PackageInfo struct {
	Group      string
	Version    string
	APIVersion string
	Domain     string
}

// DefaultRegistry is the global registry instance.
var DefaultRegistry = NewRegistry()

// NewRegistry creates a new type registry.
func NewRegistry() *Registry {
	return &Registry{
		types:    make(map[string]TypeInfo),
		packages: make(map[string]PackageInfo),
		kinds:    make(map[string]TypeInfo),
	}
}

// Register adds a type to the registry.
func (r *Registry) Register(info TypeInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Register by full qualified name (pkg.Kind)
	key := fmt.Sprintf("%s.%s", info.Package, info.Kind)
	r.types[key] = info

	// Register package info
	r.packages[info.Package] = PackageInfo{
		Group:      info.Group,
		Version:    info.Version,
		APIVersion: info.APIVersion,
		Domain:     info.Domain,
	}

	// Register kind for backward compatibility (unqualified lookup)
	r.kinds[info.Kind] = info
}

// RegisterPackage registers a package without specific kind information.
// This is useful for registering a package that contains many resources.
func (r *Registry) RegisterPackage(pkg string, info PackageInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.packages[pkg] = info
}

// IsKnownType checks if a type name is a known Kubernetes resource type.
// Accepts both qualified (pkg.Kind) and unqualified (Kind) names.
func (r *Registry) IsKnownType(typeName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check for qualified name (pkg.Kind)
	if _, ok := r.types[typeName]; ok {
		return true
	}

	// Check if the package is known (for types not explicitly registered)
	parts := strings.Split(typeName, ".")
	if len(parts) == 2 {
		pkg := parts[0]
		if _, ok := r.packages[pkg]; ok {
			return true
		}
	}

	// Check for unqualified kind name (backward compatibility)
	if _, ok := r.kinds[typeName]; ok {
		return true
	}

	return false
}

// IsKnownPackage checks if a package name is registered.
func (r *Registry) IsKnownPackage(pkg string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.packages[pkg]
	return ok
}

// GetTypeInfo returns type information for a qualified type name.
func (r *Registry) GetTypeInfo(typeName string) (TypeInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if info, ok := r.types[typeName]; ok {
		return info, true
	}

	// Try unqualified kind
	if info, ok := r.kinds[typeName]; ok {
		return info, true
	}

	return TypeInfo{}, false
}

// GetPackageInfo returns package information for a package name.
func (r *Registry) GetPackageInfo(pkg string) (PackageInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.packages[pkg]
	return info, ok
}

// APIVersionForPackage returns the API version string for a package.
// Returns empty string if the package is not registered.
func (r *Registry) APIVersionForPackage(pkg string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if info, ok := r.packages[pkg]; ok {
		return info.APIVersion
	}
	return ""
}

// ListPackages returns all registered package names.
func (r *Registry) ListPackages() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pkgs := make([]string, 0, len(r.packages))
	for pkg := range r.packages {
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

// ListTypes returns all registered type names.
func (r *Registry) ListTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.types))
	for t := range r.types {
		types = append(types, t)
	}
	return types
}

// Clear removes all registered types and packages.
// Primarily useful for testing.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.types = make(map[string]TypeInfo)
	r.packages = make(map[string]PackageInfo)
	r.kinds = make(map[string]TypeInfo)
}

// RegisterBulk registers multiple types at once.
func (r *Registry) RegisterBulk(types []TypeInfo) {
	for _, t := range types {
		r.Register(t)
	}
}
