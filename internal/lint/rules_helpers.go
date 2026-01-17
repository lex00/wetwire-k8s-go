package lint

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// unwrapCompositeLit unwraps a composite literal from potential unary expressions.
func unwrapCompositeLit(expr ast.Expr) *ast.CompositeLit {
	switch e := expr.(type) {
	case *ast.CompositeLit:
		return e
	case *ast.UnaryExpr:
		if e.Op == token.AND {
			return unwrapCompositeLit(e.X)
		}
	}
	return nil
}

// isContainerType checks if a composite literal is a Container type.
func isContainerType(compLit *ast.CompositeLit) bool {
	if compLit.Type == nil {
		return false
	}

	switch t := compLit.Type.(type) {
	case *ast.SelectorExpr:
		if _, ok := t.X.(*ast.Ident); ok {
			// Check for corev1.Container or similar
			return t.Sel.Name == "Container"
		}
	case *ast.Ident:
		return t.Name == "Container"
	}

	return false
}

// getResourceType returns the type name of a composite literal.
func getResourceType(compLit *ast.CompositeLit) string {
	if compLit.Type == nil {
		return ""
	}

	switch t := compLit.Type.(type) {
	case *ast.SelectorExpr:
		return t.Sel.Name
	case *ast.Ident:
		return t.Name
	}

	return ""
}

// extractMapLiteral extracts string key-value pairs from a map literal.
func extractMapLiteral(expr ast.Expr) map[string]string {
	result := make(map[string]string)

	compLit := unwrapCompositeLit(expr)
	if compLit == nil {
		return result
	}

	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		keyLit, ok := kv.Key.(*ast.BasicLit)
		if !ok || keyLit.Kind != token.STRING {
			continue
		}

		valueLit, ok := kv.Value.(*ast.BasicLit)
		if !ok || valueLit.Kind != token.STRING {
			continue
		}

		key := strings.Trim(keyLit.Value, `"`)
		value := strings.Trim(valueLit.Value, `"`)
		result[key] = value
	}

	return result
}

// isTrue checks if an expression evaluates to true.
func isTrue(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name == "true"
	case *ast.CallExpr:
		// Check for ptrBool(true) pattern
		if len(e.Args) == 1 {
			if ident, ok := e.Args[0].(*ast.Ident); ok {
				return ident.Name == "true"
			}
		}
	case *ast.UnaryExpr:
		// Handle &true or similar
		return isTrue(e.X)
	}
	return false
}

// isPodSpecType checks if a composite literal is a PodSpec type.
func isPodSpecType(compLit *ast.CompositeLit) bool {
	if compLit.Type == nil {
		return false
	}

	switch t := compLit.Type.(type) {
	case *ast.SelectorExpr:
		if _, ok := t.X.(*ast.Ident); ok {
			return t.Sel.Name == "PodSpec"
		}
	case *ast.Ident:
		return t.Name == "PodSpec"
	}

	return false
}

// extractIntValue extracts an integer value from an expression.
func extractIntValue(expr ast.Expr) int64 {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.INT {
			var val int64
			fmt.Sscanf(e.Value, "%d", &val)
			return val
		}
	case *ast.CallExpr:
		// Handle ptrInt32(N) pattern
		if len(e.Args) == 1 {
			return extractIntValue(e.Args[0])
		}
	case *ast.UnaryExpr:
		return extractIntValue(e.X)
	}
	return -1
}

// isK8sResourceType checks if an expression is a K8s resource type.
func isK8sResourceType(expr ast.Expr) bool {
	compLit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return false
	}

	// Check if the type is from k8s.io/api or similar
	switch t := compLit.Type.(type) {
	case *ast.SelectorExpr:
		// Check for appsv1.Deployment, corev1.Pod, etc.
		if ident, ok := t.X.(*ast.Ident); ok {
			// Check if it's a K8s API group package alias
			switch ident.Name {
			case "appsv1", "corev1", "batchv1", "networkingv1", "rbacv1", "storagev1", "autoscalingv1", "autoscalingv2", "policyv1":
				return true
			}
		}
	}

	return false
}
