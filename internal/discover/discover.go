package discover

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	coreast "github.com/lex00/wetwire-core-go/ast"
	"github.com/lex00/wetwire-k8s-go/internal/registry"
)

// DiscoverFile discovers Kubernetes resources in a single Go source file.
func DiscoverFile(filePath string) ([]Resource, error) {
	// Parse the Go source file using shared utility
	file, fset, err := coreast.ParseFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	// Get absolute path for the file
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	var resources []Resource

	// Walk through all declarations in the file
	for _, decl := range file.Decls {
		// We're only interested in variable declarations
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		// Check each spec in the declaration
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			// Extract variable name
			for i, name := range valueSpec.Names {
				if name.Name == "_" {
					continue
				}

				// Get the type - either from explicit type or from initializer
				var resourceType string
				if valueSpec.Type != nil {
					resourceType = getResourceType(valueSpec.Type)
				} else if i < len(valueSpec.Values) {
					// Type is inferred from initializer
					resourceType = getResourceTypeFromExpr(valueSpec.Values[i])
				}

				// Skip if not a Kubernetes resource
				if resourceType == "" {
					continue
				}

				// Find dependencies in the initializer
				var deps []string
				if i < len(valueSpec.Values) {
					deps = findDependencies(valueSpec.Values[i], file)
				}

				resource := Resource{
					Name:         name.Name,
					Type:         resourceType,
					File:         absPath,
					Line:         fset.Position(name.Pos()).Line,
					Dependencies: deps,
				}
				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// DiscoverDirectory discovers Kubernetes resources in all Go files within a directory recursively.
func DiscoverDirectory(dir string) ([]Resource, error) {
	var allResources []Resource

	opts := coreast.ParseOptions{
		SkipTests:  true,
		SkipVendor: true,
		SkipHidden: true,
	}

	err := coreast.WalkGoFiles(dir, opts, func(path string) error {
		// Discover resources in this file
		resources, err := DiscoverFile(path)
		if err != nil {
			// Log error but continue processing other files
			return nil
		}

		allResources = append(allResources, resources...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}

	return allResources, nil
}

// getResourceType extracts the Kubernetes resource type from an AST type expression.
// Returns empty string if not a recognized K8s resource type.
func getResourceType(typeExpr ast.Expr) string {
	typeName, pkgName := coreast.ExtractTypeName(typeExpr)
	if typeName == "" {
		return ""
	}

	// Build full type name if we have a package
	fullTypeName := typeName
	if pkgName != "" {
		fullTypeName = fmt.Sprintf("%s.%s", pkgName, typeName)
	}

	if isKubernetesType(fullTypeName) {
		return fullTypeName
	}
	return ""
}

// getResourceTypeFromExpr extracts the type from a value expression (e.g., composite literal).
func getResourceTypeFromExpr(expr ast.Expr) string {
	typeName, pkgName := coreast.InferTypeFromValue(expr)
	if typeName == "" {
		return ""
	}

	// Build full type name if we have a package
	fullTypeName := typeName
	if pkgName != "" {
		fullTypeName = fmt.Sprintf("%s.%s", pkgName, typeName)
	}

	if isKubernetesType(fullTypeName) {
		return fullTypeName
	}
	return ""
}

// isKubernetesType checks if a type name is a known Kubernetes resource type.
// Uses the registry for type lookup, supporting both standard K8s types and CRDs.
func isKubernetesType(typeName string) bool {
	// Use the global registry for type lookup
	// This supports both qualified (pkg.Kind) and unqualified (Kind) names
	if registry.DefaultRegistry.IsKnownType(typeName) {
		return true
	}

	// If it's a qualified name (package.Type), also check if the package is known
	// This allows any type from a known package to be discovered
	parts := strings.Split(typeName, ".")
	if len(parts) == 2 {
		pkg := parts[0]
		if registry.DefaultRegistry.IsKnownPackage(pkg) {
			return true
		}
	}

	return false
}

// findDependencies finds references to other top-level variables in an expression.
// This identifies dependencies between resources.
func findDependencies(expr ast.Expr, file *ast.File) []string {
	deps := make(map[string]bool)

	// Walk the expression tree
	ast.Inspect(expr, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.Ident:
			// Check if this identifier is a top-level variable in the file
			if isTopLevelVar(node.Name, file) {
				deps[node.Name] = true
			}
		case *ast.SelectorExpr:
			// Handle cases like AppConfig.Name
			if ident, ok := node.X.(*ast.Ident); ok {
				if isTopLevelVar(ident.Name, file) {
					deps[ident.Name] = true
				}
			}
		}
		return true
	})

	// Convert map to slice
	var result []string
	for dep := range deps {
		result = append(result, dep)
	}

	return result
}

// isTopLevelVar checks if a name is a top-level variable in the file.
func isTopLevelVar(name string, file *ast.File) bool {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for _, varName := range valueSpec.Names {
				if varName.Name == name {
					return true
				}
			}
		}
	}

	return false
}
