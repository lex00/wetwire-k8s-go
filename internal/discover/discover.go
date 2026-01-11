package discover

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// DiscoverFile discovers Kubernetes resources in a single Go source file.
func DiscoverFile(filePath string) ([]Resource, error) {
	// Create a new file set for position information
	fset := token.NewFileSet()

	// Parse the Go source file
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
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

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

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
	switch t := typeExpr.(type) {
	case *ast.StarExpr:
		// Pointer type, unwrap it
		return getResourceType(t.X)
	case *ast.SelectorExpr:
		// Qualified type like appsv1.Deployment
		if ident, ok := t.X.(*ast.Ident); ok {
			typeName := fmt.Sprintf("%s.%s", ident.Name, t.Sel.Name)
			if isKubernetesType(typeName) {
				return typeName
			}
		}
	case *ast.Ident:
		// Simple identifier
		if isKubernetesType(t.Name) {
			return t.Name
		}
	}
	return ""
}

// getResourceTypeFromExpr extracts the type from a value expression (e.g., composite literal).
func getResourceTypeFromExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.UnaryExpr:
		// Handle &Type{...}
		if e.Op == token.AND {
			return getResourceTypeFromExpr(e.X)
		}
	case *ast.CompositeLit:
		// Handle Type{...} or &Type{...}
		return getResourceType(e.Type)
	}
	return ""
}

// isKubernetesType checks if a type name is a known Kubernetes resource type.
func isKubernetesType(typeName string) bool {
	// Check if it's a K8s API package (these are the common patterns)
	k8sPackages := []string{
		"corev1", "appsv1", "batchv1", "networkingv1", "rbacv1",
		"storagev1", "policyv1", "autoscalingv1", "autoscalingv2",
	}

	// If it's a qualified name (package.Type), check the package
	parts := strings.Split(typeName, ".")
	if len(parts) == 2 {
		pkg := parts[0]
		for _, k8sPkg := range k8sPackages {
			if pkg == k8sPkg {
				return true
			}
		}
	}

	// Common Kubernetes resource types without package prefix
	k8sTypes := []string{
		"Pod", "Service", "Deployment", "ConfigMap", "Secret", "Ingress",
		"StatefulSet", "DaemonSet", "Job", "CronJob", "Namespace",
		"PersistentVolumeClaim", "ServiceAccount", "Role", "RoleBinding",
		"ClusterRole", "ClusterRoleBinding", "NetworkPolicy", "StorageClass",
		"HorizontalPodAutoscaler", "PodDisruptionBudget",
	}

	for _, k8sType := range k8sTypes {
		if typeName == k8sType {
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
