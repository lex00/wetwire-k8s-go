package lint

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// AllRules returns all available lint rules.
func AllRules() []Rule {
	return []Rule{
		RuleWK8001(),
		RuleWK8002(),
		RuleWK8003(),
		RuleWK8004(),
		RuleWK8005(),
		RuleWK8006(),
	}
}

// RuleWK8001 checks that resources are top-level variable declarations.
func RuleWK8001() Rule {
	return Rule{
		ID:          "WK8001",
		Name:        "Top-level resource declarations",
		Description: "Resources must be top-level variable declarations",
		Severity:    SeverityError,
		Check:       checkWK8001,
		Fix:         nil, // No auto-fix available
	}
}

func checkWK8001(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	// Walk through all declarations
	for _, decl := range file.Decls {
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

			// Check each variable
			for i, name := range valueSpec.Names {
				if name.Name == "_" {
					continue
				}

				// Check if variable is initialized with a function call
				if i < len(valueSpec.Values) {
					if isResourceFromFunctionCall(valueSpec.Values[i]) {
						pos := fset.Position(name.Pos())
						issues = append(issues, Issue{
							Rule:     "WK8001",
							Message:  fmt.Sprintf("Resource %s is assigned from function call, should be a direct composite literal", name.Name),
							File:     pos.Filename,
							Line:     pos.Line,
							Column:   pos.Column,
							Severity: SeverityError,
						})
					}
				}
			}
		}
	}

	return issues
}

// isResourceFromFunctionCall checks if an expression is a function call that might return a resource.
func isResourceFromFunctionCall(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.CallExpr:
		// Direct function call
		return true
	case *ast.UnaryExpr:
		// Handle &funcCall()
		if e.Op == token.AND {
			return isResourceFromFunctionCall(e.X)
		}
	}
	return false
}

// RuleWK8002 checks for deeply nested inline structures.
func RuleWK8002() Rule {
	return Rule{
		ID:          "WK8002",
		Name:        "Avoid deeply nested structures",
		Description: "Avoid deeply nested inline structures (max depth 5)",
		Severity:    SeverityError,
		Check:       checkWK8002,
		Fix:         nil, // Auto-fix could be implemented later
	}
}

func checkWK8002(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	// Walk through all declarations
	for _, decl := range file.Decls {
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

			// Check each variable
			for i, name := range valueSpec.Names {
				if name.Name == "_" {
					continue
				}

				// Check nesting depth in the initializer
				if i < len(valueSpec.Values) {
					depth := calculateNestingDepth(valueSpec.Values[i])
					if depth > 5 {
						pos := fset.Position(name.Pos())
						issues = append(issues, Issue{
							Rule:     "WK8002",
							Message:  fmt.Sprintf("Resource %s has nesting depth %d (max 5), consider extracting to variables", name.Name, depth),
							File:     pos.Filename,
							Line:     pos.Line,
							Column:   pos.Column,
							Severity: SeverityError,
						})
					}
				}
			}
		}
	}

	return issues
}

// calculateNestingDepth calculates the maximum nesting depth of composite literals.
func calculateNestingDepth(expr ast.Expr) int {
	switch e := expr.(type) {
	case *ast.CompositeLit:
		maxDepth := 0
		for _, elt := range e.Elts {
			depth := calculateNestingDepth(elt)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
		return maxDepth + 1
	case *ast.KeyValueExpr:
		return calculateNestingDepth(e.Value)
	case *ast.UnaryExpr:
		return calculateNestingDepth(e.X)
	default:
		return 0
	}
}

// RuleWK8003 checks for duplicate resource names in the same namespace.
func RuleWK8003() Rule {
	return Rule{
		ID:          "WK8003",
		Name:        "No duplicate resource names",
		Description: "No duplicate resource names in same namespace",
		Severity:    SeverityError,
		Check:       checkWK8003,
		Fix:         nil, // No auto-fix available
	}
}

func checkWK8003(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	// Map to track resource names by namespace
	// Key: "namespace:name", Value: variable name
	resourceMap := make(map[string]string)
	resourcePositions := make(map[string]token.Position)

	// Walk through all declarations
	for _, decl := range file.Decls {
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

			// Check each variable
			for i, name := range valueSpec.Names {
				if name.Name == "_" {
					continue
				}

				// Extract metadata from the resource
				if i < len(valueSpec.Values) {
					resourceName, namespace := extractMetadata(valueSpec.Values[i])
					if resourceName != "" {
						// Default namespace
						if namespace == "" {
							namespace = "default"
						}

						key := fmt.Sprintf("%s:%s", namespace, resourceName)
						if existingVar, exists := resourceMap[key]; exists {
							pos := fset.Position(name.Pos())
							prevPos := resourcePositions[key]
							issues = append(issues, Issue{
								Rule:     "WK8003",
								Message:  fmt.Sprintf("Duplicate resource name %q in namespace %q (first defined as %s at line %d)", resourceName, namespace, existingVar, prevPos.Line),
								File:     pos.Filename,
								Line:     pos.Line,
								Column:   pos.Column,
								Severity: SeverityError,
							})
						} else {
							resourceMap[key] = name.Name
							resourcePositions[key] = fset.Position(name.Pos())
						}
					}
				}
			}
		}
	}

	return issues
}

// extractMetadata extracts the resource name and namespace from a composite literal.
func extractMetadata(expr ast.Expr) (name, namespace string) {
	compLit := unwrapCompositeLit(expr)
	if compLit == nil {
		return "", ""
	}

	// Look for Metadata field
	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Metadata" {
			continue
		}

		// Extract from Metadata composite literal
		metadataLit := unwrapCompositeLit(kv.Value)
		if metadataLit == nil {
			continue
		}

		for _, metaElt := range metadataLit.Elts {
			metaKV, ok := metaElt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			metaKey, ok := metaKV.Key.(*ast.Ident)
			if !ok {
				continue
			}

			switch metaKey.Name {
			case "Name":
				if lit, ok := metaKV.Value.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					name = strings.Trim(lit.Value, `"`)
				}
			case "Namespace":
				if lit, ok := metaKV.Value.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					namespace = strings.Trim(lit.Value, `"`)
				}
			}
		}
	}

	return name, namespace
}

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

// RuleWK8004 checks for circular dependencies.
func RuleWK8004() Rule {
	return Rule{
		ID:          "WK8004",
		Name:        "Circular dependency detection",
		Description: "Circular dependency detection",
		Severity:    SeverityError,
		Check:       checkWK8004,
		Fix:         nil, // No auto-fix available
	}
}

func checkWK8004(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	// Build dependency graph
	dependencies := make(map[string][]string)
	varPositions := make(map[string]token.Position)

	// Walk through all declarations
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

			for i, name := range valueSpec.Names {
				if name.Name == "_" {
					continue
				}

				varPositions[name.Name] = fset.Position(name.Pos())

				// Find dependencies in the initializer
				if i < len(valueSpec.Values) {
					deps := findVariableReferences(valueSpec.Values[i], file)
					dependencies[name.Name] = deps
				}
			}
		}
	}

	// Detect cycles using DFS
	for varName := range dependencies {
		if cycle := detectCycle(varName, dependencies, make(map[string]bool), []string{}); cycle != nil {
			pos := varPositions[varName]
			issues = append(issues, Issue{
				Rule:     "WK8004",
				Message:  fmt.Sprintf("Circular dependency detected: %s", strings.Join(cycle, " -> ")),
				File:     pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Severity: SeverityError,
			})
		}
	}

	return issues
}

// detectCycle detects circular dependencies using DFS.
func detectCycle(current string, deps map[string][]string, visited map[string]bool, path []string) []string {
	// Check if we've already seen this node in the current path
	for _, node := range path {
		if node == current {
			// Found a cycle
			return append(path, current)
		}
	}

	// Check if we've already visited this node in a previous path
	if visited[current] {
		return nil
	}

	visited[current] = true
	newPath := append(path, current)

	// Visit all dependencies
	for _, dep := range deps[current] {
		if cycle := detectCycle(dep, deps, visited, newPath); cycle != nil {
			return cycle
		}
	}

	return nil
}

// findVariableReferences finds all variable references in an expression.
func findVariableReferences(expr ast.Expr, file *ast.File) []string {
	refs := make(map[string]bool)

	ast.Inspect(expr, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.Ident:
			if isTopLevelVar(node.Name, file) {
				refs[node.Name] = true
			}
		case *ast.SelectorExpr:
			if ident, ok := node.X.(*ast.Ident); ok {
				if isTopLevelVar(ident.Name, file) {
					refs[ident.Name] = true
				}
			}
		}
		return true
	})

	// Convert map to slice
	var result []string
	for ref := range refs {
		result = append(result, ref)
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

// RuleWK8005 checks for hardcoded secrets in environment variables.
func RuleWK8005() Rule {
	return Rule{
		ID:          "WK8005",
		Name:        "Flag hardcoded secrets",
		Description: "Flag hardcoded secrets in env vars",
		Severity:    SeverityError,
		Check:       checkWK8005,
		Fix:         nil, // No auto-fix available
	}
}

func checkWK8005(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	// Sensitive environment variable patterns
	sensitivePatterns := []string{
		"password", "passwd", "pwd",
		"secret", "token", "key",
		"api_key", "apikey", "auth",
		"credential", "private",
	}

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Check if this is an EnvVar struct
		if !isEnvVarType(compLit) {
			return true
		}

		// Check for hardcoded values
		var envName, envValue string
		var valueLine int

		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok {
				continue
			}

			if key.Name == "Name" {
				if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					envName = strings.Trim(lit.Value, `"`)
				}
			} else if key.Name == "Value" {
				if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					envValue = strings.Trim(lit.Value, `"`)
					valueLine = fset.Position(lit.Pos()).Line
				}
			}
		}

		// Check if the env var name matches sensitive patterns and has a hardcoded value
		if envName != "" && envValue != "" {
			envNameLower := strings.ToLower(envName)
			for _, pattern := range sensitivePatterns {
				if strings.Contains(envNameLower, pattern) {
					pos := fset.Position(compLit.Pos())
					if valueLine > 0 {
						pos.Line = valueLine
					}
					issues = append(issues, Issue{
						Rule:     "WK8005",
						Message:  fmt.Sprintf("Hardcoded secret detected in environment variable %q, use SecretKeyRef instead", envName),
						File:     pos.Filename,
						Line:     pos.Line,
						Column:   pos.Column,
						Severity: SeverityError,
					})
					break
				}
			}
		}

		return true
	})

	return issues
}

// isEnvVarType checks if a composite literal is an EnvVar type.
func isEnvVarType(compLit *ast.CompositeLit) bool {
	if compLit.Type == nil {
		return false
	}

	switch t := compLit.Type.(type) {
	case *ast.SelectorExpr:
		if _, ok := t.X.(*ast.Ident); ok {
			// Check for corev1.EnvVar or similar
			return t.Sel.Name == "EnvVar"
		}
	case *ast.Ident:
		return t.Name == "EnvVar"
	}

	return false
}

// RuleWK8006 checks for :latest image tags.
func RuleWK8006() Rule {
	return Rule{
		ID:          "WK8006",
		Name:        "Flag :latest image tags",
		Description: "Flag :latest image tags",
		Severity:    SeverityError,
		Check:       checkWK8006,
		Fix:         nil, // No auto-fix available
	}
}

func checkWK8006(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Check if this is a Container struct
		if !isContainerType(compLit) {
			return true
		}

		// Check for Image field
		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "Image" {
				continue
			}

			// Check if the image value uses :latest
			if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				imageValue := strings.Trim(lit.Value, `"`)
				if strings.HasSuffix(imageValue, ":latest") || !strings.Contains(imageValue, ":") && !strings.Contains(imageValue, "@") {
					pos := fset.Position(lit.Pos())
					issues = append(issues, Issue{
						Rule:     "WK8006",
						Message:  fmt.Sprintf("Image %q uses :latest tag or no tag (defaults to :latest), specify a version tag", imageValue),
						File:     pos.Filename,
						Line:     pos.Line,
						Column:   pos.Column,
						Severity: SeverityError,
					})
				}
			}
		}

		return true
	})

	return issues
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
