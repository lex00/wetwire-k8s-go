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
		RuleWK8041(),
		RuleWK8042(),
		RuleWK8101(),
		RuleWK8102(),
		RuleWK8103(),
		RuleWK8104(),
		RuleWK8105(),
		RuleWK8201(),
		RuleWK8202(),
		RuleWK8203(),
		RuleWK8204(),
		RuleWK8205(),
		RuleWK8207(),
		RuleWK8208(),
		RuleWK8209(),
		RuleWK8301(),
		RuleWK8302(),
		RuleWK8303(),
		RuleWK8304(),
		RuleWK8401(),
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
		Fix:         nil, // No auto-fix available - user must specify version
	}
}

// RuleWK8105 checks for missing ImagePullPolicy on containers.
func RuleWK8105() Rule {
	return Rule{
		ID:          "WK8105",
		Name:        "ImagePullPolicy explicit",
		Description: "ImagePullPolicy should be explicitly set",
		Severity:    SeverityWarning,
		Check:       checkWK8105,
		Fix:         nil, // Fix implemented in fixer.go
	}
}

func checkWK8105(file *ast.File, fset *token.FileSet) []Issue {
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

		// Check if ImagePullPolicy is set
		hasImagePullPolicy := false
		var imageValue string
		var containerPos token.Position

		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok {
				continue
			}

			if key.Name == "ImagePullPolicy" {
				hasImagePullPolicy = true
			} else if key.Name == "Image" {
				if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					imageValue = strings.Trim(lit.Value, `"`)
				}
			}
		}

		containerPos = fset.Position(compLit.Pos())

		if !hasImagePullPolicy && imageValue != "" {
			issues = append(issues, Issue{
				Rule:     "WK8105",
				Message:  fmt.Sprintf("Container with image %q should have explicit ImagePullPolicy", imageValue),
				File:     containerPos.Filename,
				Line:     containerPos.Line,
				Column:   containerPos.Column,
				Severity: SeverityWarning,
			})
		}

		return true
	})

	return issues
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

// RuleWK8041 checks for hardcoded API keys/tokens.
func RuleWK8041() Rule {
	return Rule{
		ID:          "WK8041",
		Name:        "Hardcoded API keys/tokens",
		Description: "Hardcoded API keys/tokens detected",
		Severity:    SeverityError,
		Check:       checkWK8041,
		Fix:         nil,
	}
}

func checkWK8041(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	// Patterns to detect API keys and tokens
	tokenPatterns := []string{
		"Bearer ",
		"api_key=",
		"apikey=",
		"token:",
		"ghp_",      // GitHub personal token
		"gho_",      // GitHub OAuth token
		"ghs_",      // GitHub server token
		"AKIA",      // AWS access key
		"sk_live_",  // Stripe live key
		"sk_test_",  // Stripe test key
		"rk_live_",  // Stripe restricted key
		"pk_live_",  // Stripe publishable key
	}

	ast.Inspect(file, func(n ast.Node) bool {
		// Check string literals in the AST
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}

		value := strings.Trim(lit.Value, `"`)
		valueLower := strings.ToLower(value)

		for _, pattern := range tokenPatterns {
			if strings.Contains(value, pattern) || strings.Contains(valueLower, strings.ToLower(pattern)) {
				pos := fset.Position(lit.Pos())
				issues = append(issues, Issue{
					Rule:     "WK8041",
					Message:  fmt.Sprintf("Hardcoded API key/token pattern detected: %q, use SecretKeyRef instead", pattern),
					File:     pos.Filename,
					Line:     pos.Line,
					Column:   pos.Column,
					Severity: SeverityError,
				})
				break
			}
		}

		return true
	})

	return issues
}

// RuleWK8042 checks for private key headers in ConfigMaps.
func RuleWK8042() Rule {
	return Rule{
		ID:          "WK8042",
		Name:        "Private key headers",
		Description: "Private key headers detected in ConfigMap",
		Severity:    SeverityError,
		Check:       checkWK8042,
		Fix:         nil,
	}
}

func checkWK8042(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	// Private key header patterns
	privateKeyPatterns := []string{
		"-----BEGIN RSA PRIVATE KEY-----",
		"-----BEGIN PRIVATE KEY-----",
		"-----BEGIN EC PRIVATE KEY-----",
		"-----BEGIN OPENSSH PRIVATE KEY-----",
		"-----BEGIN DSA PRIVATE KEY-----",
		"-----BEGIN ENCRYPTED PRIVATE KEY-----",
	}

	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}

		value := strings.Trim(lit.Value, "`\"")

		for _, pattern := range privateKeyPatterns {
			if strings.Contains(value, pattern) {
				pos := fset.Position(lit.Pos())
				issues = append(issues, Issue{
					Rule:     "WK8042",
					Message:  "Private key detected in configuration, use Secret instead of ConfigMap",
					File:     pos.Filename,
					Line:     pos.Line,
					Column:   pos.Column,
					Severity: SeverityError,
				})
				break
			}
		}

		return true
	})

	return issues
}

// RuleWK8101 checks for selector label mismatch in Deployments/StatefulSets/DaemonSets.
func RuleWK8101() Rule {
	return Rule{
		ID:          "WK8101",
		Name:        "Selector label mismatch",
		Description: "Deployment selector labels must match template labels",
		Severity:    SeverityError,
		Check:       checkWK8101,
		Fix:         nil,
	}
}

func checkWK8101(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Check if this is a Deployment, StatefulSet, or DaemonSet
		resourceType := getResourceType(compLit)
		if resourceType != "Deployment" && resourceType != "StatefulSet" && resourceType != "DaemonSet" {
			return true
		}

		// Extract selector and template labels
		selectorLabels := extractSelectorLabels(compLit)
		templateLabels := extractTemplateLabels(compLit)

		if len(selectorLabels) == 0 || len(templateLabels) == 0 {
			return true
		}

		// Check if all selector labels are present in template labels
		for key, selectorValue := range selectorLabels {
			templateValue, exists := templateLabels[key]
			if !exists {
				pos := fset.Position(compLit.Pos())
				issues = append(issues, Issue{
					Rule:     "WK8101",
					Message:  fmt.Sprintf("Selector label %q not found in template labels", key),
					File:     pos.Filename,
					Line:     pos.Line,
					Column:   pos.Column,
					Severity: SeverityError,
				})
			} else if selectorValue != templateValue && selectorValue != "" && templateValue != "" {
				pos := fset.Position(compLit.Pos())
				issues = append(issues, Issue{
					Rule:     "WK8101",
					Message:  fmt.Sprintf("Selector label %q has value %q but template has %q", key, selectorValue, templateValue),
					File:     pos.Filename,
					Line:     pos.Line,
					Column:   pos.Column,
					Severity: SeverityError,
				})
			}
		}

		return true
	})

	return issues
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

// extractSelectorLabels extracts selector labels from a resource spec.
func extractSelectorLabels(compLit *ast.CompositeLit) map[string]string {
	labels := make(map[string]string)

	// Look for Spec.Selector.MatchLabels
	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Spec" {
			continue
		}

		specLit := unwrapCompositeLit(kv.Value)
		if specLit == nil {
			continue
		}

		for _, specElt := range specLit.Elts {
			specKV, ok := specElt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			specKey, ok := specKV.Key.(*ast.Ident)
			if !ok || specKey.Name != "Selector" {
				continue
			}

			selectorLit := unwrapCompositeLit(specKV.Value)
			if selectorLit == nil {
				continue
			}

			for _, selectorElt := range selectorLit.Elts {
				selectorKV, ok := selectorElt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				selectorKey, ok := selectorKV.Key.(*ast.Ident)
				if !ok || selectorKey.Name != "MatchLabels" {
					continue
				}

				matchLabels := extractMapLiteral(selectorKV.Value)
				return matchLabels
			}
		}
	}

	return labels
}

// extractTemplateLabels extracts template labels from a resource spec.
func extractTemplateLabels(compLit *ast.CompositeLit) map[string]string {
	labels := make(map[string]string)

	// Look for Spec.Template.Metadata.Labels or Spec.Template.ObjectMeta.Labels
	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Spec" {
			continue
		}

		specLit := unwrapCompositeLit(kv.Value)
		if specLit == nil {
			continue
		}

		for _, specElt := range specLit.Elts {
			specKV, ok := specElt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			specKey, ok := specKV.Key.(*ast.Ident)
			if !ok || specKey.Name != "Template" {
				continue
			}

			templateLit := unwrapCompositeLit(specKV.Value)
			if templateLit == nil {
				continue
			}

			for _, templateElt := range templateLit.Elts {
				templateKV, ok := templateElt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				templateKey, ok := templateKV.Key.(*ast.Ident)
				if !ok || (templateKey.Name != "Metadata" && templateKey.Name != "ObjectMeta") {
					continue
				}

				metadataLit := unwrapCompositeLit(templateKV.Value)
				if metadataLit == nil {
					continue
				}

				for _, metadataElt := range metadataLit.Elts {
					metadataKV, ok := metadataElt.(*ast.KeyValueExpr)
					if !ok {
						continue
					}

					metadataKey, ok := metadataKV.Key.(*ast.Ident)
					if !ok || metadataKey.Name != "Labels" {
						continue
					}

					return extractMapLiteral(metadataKV.Value)
				}
			}
		}
	}

	return labels
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

// RuleWK8102 checks for missing labels on resources.
func RuleWK8102() Rule {
	return Rule{
		ID:          "WK8102",
		Name:        "Missing labels",
		Description: "Resources should have metadata labels",
		Severity:    SeverityWarning,
		Check:       checkWK8102,
		Fix:         nil,
	}
}

func checkWK8102(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	// K8s resource types that should have labels
	resourceTypes := map[string]bool{
		"Deployment":  true,
		"Service":     true,
		"Pod":         true,
		"ConfigMap":   true,
		"Secret":      true,
		"StatefulSet": true,
		"DaemonSet":   true,
		"Ingress":     true,
		"Job":         true,
		"CronJob":     true,
	}

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		resourceType := getResourceType(compLit)
		if !resourceTypes[resourceType] {
			return true
		}

		// Check if ObjectMeta/Metadata has Labels field
		hasLabels := false
		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok || (key.Name != "Metadata" && key.Name != "ObjectMeta") {
				continue
			}

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
				if !ok || metaKey.Name != "Labels" {
					continue
				}

				// Check if labels map is not empty
				labelsMap := extractMapLiteral(metaKV.Value)
				if len(labelsMap) > 0 {
					hasLabels = true
				}
			}
		}

		if !hasLabels {
			pos := fset.Position(compLit.Pos())
			issues = append(issues, Issue{
				Rule:     "WK8102",
				Message:  fmt.Sprintf("%s should have metadata labels for better organization", resourceType),
				File:     pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Severity: SeverityWarning,
			})
		}

		return true
	})

	return issues
}

// RuleWK8201 checks for missing resource limits on containers.
func RuleWK8201() Rule {
	return Rule{
		ID:          "WK8201",
		Name:        "Missing resource limits",
		Description: "Containers should have resource limits",
		Severity:    SeverityWarning,
		Check:       checkWK8201,
		Fix:         nil,
	}
}

func checkWK8201(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		if !isContainerType(compLit) {
			return true
		}

		// Check if container has Resources field with Limits
		hasLimits := false
		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "Resources" {
				continue
			}

			resourcesLit := unwrapCompositeLit(kv.Value)
			if resourcesLit == nil {
				continue
			}

			for _, resElt := range resourcesLit.Elts {
				resKV, ok := resElt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				resKey, ok := resKV.Key.(*ast.Ident)
				if !ok || resKey.Name != "Limits" {
					continue
				}

				// Check if Limits is not empty
				limitsLit := unwrapCompositeLit(resKV.Value)
				if limitsLit != nil && len(limitsLit.Elts) > 0 {
					hasLimits = true
				}
			}
		}

		if !hasLimits {
			pos := fset.Position(compLit.Pos())
			issues = append(issues, Issue{
				Rule:     "WK8201",
				Message:  "Container should have resource limits (cpu, memory) to prevent resource exhaustion",
				File:     pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Severity: SeverityWarning,
			})
		}

		return true
	})

	return issues
}

// RuleWK8202 checks for privileged containers.
func RuleWK8202() Rule {
	return Rule{
		ID:          "WK8202",
		Name:        "Privileged containers",
		Description: "Containers should not run in privileged mode",
		Severity:    SeverityError,
		Check:       checkWK8202,
		Fix:         nil,
	}
}

func checkWK8202(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		if !isContainerType(compLit) {
			return true
		}

		// Check if container has SecurityContext.Privileged = true
		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "SecurityContext" {
				continue
			}

			securityLit := unwrapCompositeLit(kv.Value)
			if securityLit == nil {
				continue
			}

			for _, secElt := range securityLit.Elts {
				secKV, ok := secElt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				secKey, ok := secKV.Key.(*ast.Ident)
				if !ok || secKey.Name != "Privileged" {
					continue
				}

				// Check if value is true (through function call or literal)
				if isTrue(secKV.Value) {
					pos := fset.Position(compLit.Pos())
					issues = append(issues, Issue{
						Rule:     "WK8202",
						Message:  "Container should not run in privileged mode, it has full access to the host",
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

// RuleWK8301 checks for missing health probes on containers.
func RuleWK8301() Rule {
	return Rule{
		ID:          "WK8301",
		Name:        "Missing health probes",
		Description: "Containers should have liveness and readiness probes",
		Severity:    SeverityWarning,
		Check:       checkWK8301,
		Fix:         nil,
	}
}

func checkWK8301(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		if !isContainerType(compLit) {
			return true
		}

		hasLiveness := false
		hasReadiness := false

		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok {
				continue
			}

			if key.Name == "LivenessProbe" {
				hasLiveness = true
			} else if key.Name == "ReadinessProbe" {
				hasReadiness = true
			}
		}

		if !hasLiveness || !hasReadiness {
			pos := fset.Position(compLit.Pos())
			missing := []string{}
			if !hasLiveness {
				missing = append(missing, "liveness")
			}
			if !hasReadiness {
				missing = append(missing, "readiness")
			}
			issues = append(issues, Issue{
				Rule:     "WK8301",
				Message:  fmt.Sprintf("Container should have %s probe(s) for automatic failure detection", strings.Join(missing, " and ")),
				File:     pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Severity: SeverityWarning,
			})
		}

		return true
	})

	return issues
}

// RuleWK8103 checks that containers have a Name field.
func RuleWK8103() Rule {
	return Rule{
		ID:          "WK8103",
		Name:        "Container name required",
		Description: "All containers must have a Name field",
		Severity:    SeverityError,
		Check:       checkWK8103,
		Fix:         nil,
	}
}

func checkWK8103(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		if !isContainerType(compLit) {
			return true
		}

		// Check if Name field is set
		hasName := false
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
				// Check if it has a non-empty value
				if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					value := strings.Trim(lit.Value, `"`)
					if value != "" {
						hasName = true
					}
				} else {
					// Name is set via variable or other expression
					hasName = true
				}
			}
		}

		if !hasName {
			pos := fset.Position(compLit.Pos())
			issues = append(issues, Issue{
				Rule:     "WK8103",
				Message:  "Container must have a Name field",
				File:     pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Severity: SeverityError,
			})
		}

		return true
	})

	return issues
}

// RuleWK8104 checks that container and service ports are named.
func RuleWK8104() Rule {
	return Rule{
		ID:          "WK8104",
		Name:        "Port name recommended",
		Description: "Container and Service ports should be named",
		Severity:    SeverityWarning,
		Check:       checkWK8104,
		Fix:         nil,
	}
}

func checkWK8104(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Check for ContainerPort or ServicePort types
		portType := getPortType(compLit)
		if portType == "" {
			return true
		}

		// Check if Name field is set
		hasName := false
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
					value := strings.Trim(lit.Value, `"`)
					if value != "" {
						hasName = true
					}
				} else {
					hasName = true
				}
			}
		}

		if !hasName {
			pos := fset.Position(compLit.Pos())
			issues = append(issues, Issue{
				Rule:     "WK8104",
				Message:  fmt.Sprintf("%s should have a Name for better documentation and service mesh support", portType),
				File:     pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Severity: SeverityWarning,
			})
		}

		return true
	})

	return issues
}

// getPortType returns the port type name if it's a port type.
func getPortType(compLit *ast.CompositeLit) string {
	if compLit.Type == nil {
		return ""
	}

	switch t := compLit.Type.(type) {
	case *ast.SelectorExpr:
		if _, ok := t.X.(*ast.Ident); ok {
			if t.Sel.Name == "ContainerPort" || t.Sel.Name == "ServicePort" {
				return t.Sel.Name
			}
		}
	case *ast.Ident:
		if t.Name == "ContainerPort" || t.Name == "ServicePort" {
			return t.Name
		}
	}

	return ""
}

// RuleWK8203 checks for ReadOnlyRootFilesystem in SecurityContext.
func RuleWK8203() Rule {
	return Rule{
		ID:          "WK8203",
		Name:        "ReadOnlyRootFilesystem",
		Description: "Containers should set ReadOnlyRootFilesystem: true",
		Severity:    SeverityWarning,
		Check:       checkWK8203,
		Fix:         nil,
	}
}

func checkWK8203(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		if !isContainerType(compLit) {
			return true
		}

		// Check for SecurityContext.ReadOnlyRootFilesystem = true
		hasReadOnlyFS := false
		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "SecurityContext" {
				continue
			}

			securityLit := unwrapCompositeLit(kv.Value)
			if securityLit == nil {
				continue
			}

			for _, secElt := range securityLit.Elts {
				secKV, ok := secElt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				secKey, ok := secKV.Key.(*ast.Ident)
				if !ok || secKey.Name != "ReadOnlyRootFilesystem" {
					continue
				}

				if isTrue(secKV.Value) {
					hasReadOnlyFS = true
				}
			}
		}

		if !hasReadOnlyFS {
			pos := fset.Position(compLit.Pos())
			issues = append(issues, Issue{
				Rule:     "WK8203",
				Message:  "Container should set ReadOnlyRootFilesystem: true to reduce attack surface",
				File:     pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Severity: SeverityWarning,
			})
		}

		return true
	})

	return issues
}

// RuleWK8204 checks for RunAsNonRoot in SecurityContext.
func RuleWK8204() Rule {
	return Rule{
		ID:          "WK8204",
		Name:        "RunAsNonRoot",
		Description: "Containers should set RunAsNonRoot: true",
		Severity:    SeverityWarning,
		Check:       checkWK8204,
		Fix:         nil,
	}
}

func checkWK8204(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		if !isContainerType(compLit) {
			return true
		}

		// Check for SecurityContext.RunAsNonRoot = true
		hasRunAsNonRoot := false
		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "SecurityContext" {
				continue
			}

			securityLit := unwrapCompositeLit(kv.Value)
			if securityLit == nil {
				continue
			}

			for _, secElt := range securityLit.Elts {
				secKV, ok := secElt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				secKey, ok := secKV.Key.(*ast.Ident)
				if !ok || secKey.Name != "RunAsNonRoot" {
					continue
				}

				if isTrue(secKV.Value) {
					hasRunAsNonRoot = true
				}
			}
		}

		if !hasRunAsNonRoot {
			pos := fset.Position(compLit.Pos())
			issues = append(issues, Issue{
				Rule:     "WK8204",
				Message:  "Container should set RunAsNonRoot: true to limit security risks",
				File:     pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Severity: SeverityWarning,
			})
		}

		return true
	})

	return issues
}

// RuleWK8205 checks for dropping capabilities in SecurityContext.
func RuleWK8205() Rule {
	return Rule{
		ID:          "WK8205",
		Name:        "Drop capabilities",
		Description: "Containers should drop unnecessary Linux capabilities",
		Severity:    SeverityWarning,
		Check:       checkWK8205,
		Fix:         nil,
	}
}

func checkWK8205(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		if !isContainerType(compLit) {
			return true
		}

		// Check for SecurityContext.Capabilities.Drop with non-empty list
		hasDropCapabilities := false
		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "SecurityContext" {
				continue
			}

			securityLit := unwrapCompositeLit(kv.Value)
			if securityLit == nil {
				continue
			}

			for _, secElt := range securityLit.Elts {
				secKV, ok := secElt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				secKey, ok := secKV.Key.(*ast.Ident)
				if !ok || secKey.Name != "Capabilities" {
					continue
				}

				capLit := unwrapCompositeLit(secKV.Value)
				if capLit == nil {
					continue
				}

				for _, capElt := range capLit.Elts {
					capKV, ok := capElt.(*ast.KeyValueExpr)
					if !ok {
						continue
					}

					capKey, ok := capKV.Key.(*ast.Ident)
					if !ok || capKey.Name != "Drop" {
						continue
					}

					// Check if Drop has non-empty slice
					if dropLit, ok := capKV.Value.(*ast.CompositeLit); ok {
						if len(dropLit.Elts) > 0 {
							hasDropCapabilities = true
						}
					}
				}
			}
		}

		if !hasDropCapabilities {
			pos := fset.Position(compLit.Pos())
			issues = append(issues, Issue{
				Rule:     "WK8205",
				Message:  "Container should drop unnecessary capabilities (e.g., Drop: []string{\"ALL\"})",
				File:     pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Severity: SeverityWarning,
			})
		}

		return true
	})

	return issues
}

// RuleWK8207 checks for HostNetwork usage in PodSpec.
func RuleWK8207() Rule {
	return Rule{
		ID:          "WK8207",
		Name:        "No host network",
		Description: "Pods should not use HostNetwork: true",
		Severity:    SeverityWarning,
		Check:       checkWK8207,
		Fix:         nil,
	}
}

func checkWK8207(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		if !isPodSpecType(compLit) {
			return true
		}

		// Check for HostNetwork = true
		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "HostNetwork" {
				continue
			}

			if isTrue(kv.Value) {
				pos := fset.Position(kv.Pos())
				issues = append(issues, Issue{
					Rule:     "WK8207",
					Message:  "Pod should not use HostNetwork: true, it bypasses network policies",
					File:     pos.Filename,
					Line:     pos.Line,
					Column:   pos.Column,
					Severity: SeverityWarning,
				})
			}
		}

		return true
	})

	return issues
}

// RuleWK8208 checks for HostPID usage in PodSpec.
func RuleWK8208() Rule {
	return Rule{
		ID:          "WK8208",
		Name:        "No host PID",
		Description: "Pods should not use HostPID: true",
		Severity:    SeverityWarning,
		Check:       checkWK8208,
		Fix:         nil,
	}
}

func checkWK8208(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		if !isPodSpecType(compLit) {
			return true
		}

		// Check for HostPID = true
		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "HostPID" {
				continue
			}

			if isTrue(kv.Value) {
				pos := fset.Position(kv.Pos())
				issues = append(issues, Issue{
					Rule:     "WK8208",
					Message:  "Pod should not use HostPID: true, it allows viewing host processes",
					File:     pos.Filename,
					Line:     pos.Line,
					Column:   pos.Column,
					Severity: SeverityWarning,
				})
			}
		}

		return true
	})

	return issues
}

// RuleWK8209 checks for HostIPC usage in PodSpec.
func RuleWK8209() Rule {
	return Rule{
		ID:          "WK8209",
		Name:        "No host IPC",
		Description: "Pods should not use HostIPC: true",
		Severity:    SeverityWarning,
		Check:       checkWK8209,
		Fix:         nil,
	}
}

func checkWK8209(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		if !isPodSpecType(compLit) {
			return true
		}

		// Check for HostIPC = true
		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "HostIPC" {
				continue
			}

			if isTrue(kv.Value) {
				pos := fset.Position(kv.Pos())
				issues = append(issues, Issue{
					Rule:     "WK8209",
					Message:  "Pod should not use HostIPC: true, it enables IPC with host processes",
					File:     pos.Filename,
					Line:     pos.Line,
					Column:   pos.Column,
					Severity: SeverityWarning,
				})
			}
		}

		return true
	})

	return issues
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

// RuleWK8302 checks for minimum replicas in Deployments.
func RuleWK8302() Rule {
	return Rule{
		ID:          "WK8302",
		Name:        "Replicas minimum",
		Description: "Deployments should have at least 2 replicas for high availability",
		Severity:    SeverityInfo,
		Check:       checkWK8302,
		Fix:         nil,
	}
}

func checkWK8302(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		resourceType := getResourceType(compLit)
		if resourceType != "Deployment" {
			return true
		}

		// Check for Spec.Replicas
		replicaCount := int64(-1) // -1 means not set (defaults to 1)
		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "Spec" {
				continue
			}

			specLit := unwrapCompositeLit(kv.Value)
			if specLit == nil {
				continue
			}

			for _, specElt := range specLit.Elts {
				specKV, ok := specElt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				specKey, ok := specKV.Key.(*ast.Ident)
				if !ok || specKey.Name != "Replicas" {
					continue
				}

				// Extract replica count
				replicaCount = extractIntValue(specKV.Value)
			}
		}

		// Check if replicas < 2 (including not set, which defaults to 1)
		if replicaCount < 2 {
			pos := fset.Position(compLit.Pos())
			msg := "Deployment should have at least 2 replicas for high availability"
			if replicaCount == -1 {
				msg = "Deployment should explicitly set replicas >= 2 for high availability"
			}
			issues = append(issues, Issue{
				Rule:     "WK8302",
				Message:  msg,
				File:     pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Severity: SeverityInfo,
			})
		}

		return true
	})

	return issues
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

// RuleWK8303 checks for PodDisruptionBudget for HA deployments.
func RuleWK8303() Rule {
	return Rule{
		ID:          "WK8303",
		Name:        "PodDisruptionBudget",
		Description: "HA deployments should have a PodDisruptionBudget",
		Severity:    SeverityInfo,
		Check:       checkWK8303,
		Fix:         nil,
	}
}

func checkWK8303(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	// First, collect all PDB selectors in the file
	pdbSelectors := make(map[string]bool)
	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		if getResourceType(compLit) != "PodDisruptionBudget" {
			return true
		}

		// Extract selector labels from PDB
		labels := extractPDBSelectorLabels(compLit)
		for key, value := range labels {
			pdbSelectors[key+"="+value] = true
		}
		return true
	})

	// Then check deployments
	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		resourceType := getResourceType(compLit)
		if resourceType != "Deployment" {
			return true
		}

		// Get replica count
		replicaCount := extractDeploymentReplicas(compLit)
		if replicaCount < 2 {
			// Not an HA deployment
			return true
		}

		// Get deployment labels (selector match labels)
		deploymentLabels := extractSelectorLabels(compLit)
		if len(deploymentLabels) == 0 {
			return true
		}

		// Check if any PDB matches these labels
		hasPDB := false
		for key, value := range deploymentLabels {
			if pdbSelectors[key+"="+value] {
				hasPDB = true
				break
			}
		}

		if !hasPDB {
			pos := fset.Position(compLit.Pos())
			issues = append(issues, Issue{
				Rule:     "WK8303",
				Message:  "HA deployment (replicas >= 2) should have a PodDisruptionBudget",
				File:     pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Severity: SeverityInfo,
			})
		}

		return true
	})

	return issues
}

// extractPDBSelectorLabels extracts selector labels from a PodDisruptionBudget.
func extractPDBSelectorLabels(compLit *ast.CompositeLit) map[string]string {
	labels := make(map[string]string)

	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Spec" {
			continue
		}

		specLit := unwrapCompositeLit(kv.Value)
		if specLit == nil {
			continue
		}

		for _, specElt := range specLit.Elts {
			specKV, ok := specElt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			specKey, ok := specKV.Key.(*ast.Ident)
			if !ok || specKey.Name != "Selector" {
				continue
			}

			selectorLit := unwrapCompositeLit(specKV.Value)
			if selectorLit == nil {
				continue
			}

			for _, selectorElt := range selectorLit.Elts {
				selectorKV, ok := selectorElt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				selectorKey, ok := selectorKV.Key.(*ast.Ident)
				if !ok || selectorKey.Name != "MatchLabels" {
					continue
				}

				return extractMapLiteral(selectorKV.Value)
			}
		}
	}

	return labels
}

// extractDeploymentReplicas extracts the replica count from a Deployment.
func extractDeploymentReplicas(compLit *ast.CompositeLit) int64 {
	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Spec" {
			continue
		}

		specLit := unwrapCompositeLit(kv.Value)
		if specLit == nil {
			continue
		}

		for _, specElt := range specLit.Elts {
			specKV, ok := specElt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			specKey, ok := specKV.Key.(*ast.Ident)
			if !ok || specKey.Name != "Replicas" {
				continue
			}

			return extractIntValue(specKV.Value)
		}
	}

	return 1 // Default replicas
}

// RuleWK8304 checks for anti-affinity in HA deployments.
func RuleWK8304() Rule {
	return Rule{
		ID:          "WK8304",
		Name:        "Anti-affinity recommended",
		Description: "HA deployments should use pod anti-affinity",
		Severity:    SeverityInfo,
		Check:       checkWK8304,
		Fix:         nil,
	}
}

func checkWK8304(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	ast.Inspect(file, func(n ast.Node) bool {
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		resourceType := getResourceType(compLit)
		if resourceType != "Deployment" {
			return true
		}

		// Get replica count
		replicaCount := extractDeploymentReplicas(compLit)
		if replicaCount < 2 {
			// Not an HA deployment
			return true
		}

		// Check for PodAntiAffinity in template spec
		hasAntiAffinity := checkForPodAntiAffinity(compLit)

		if !hasAntiAffinity {
			pos := fset.Position(compLit.Pos())
			issues = append(issues, Issue{
				Rule:     "WK8304",
				Message:  "HA deployment (replicas >= 2) should use pod anti-affinity to spread across nodes",
				File:     pos.Filename,
				Line:     pos.Line,
				Column:   pos.Column,
				Severity: SeverityInfo,
			})
		}

		return true
	})

	return issues
}

// checkForPodAntiAffinity checks if a Deployment has PodAntiAffinity configured.
func checkForPodAntiAffinity(compLit *ast.CompositeLit) bool {
	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Spec" {
			continue
		}

		specLit := unwrapCompositeLit(kv.Value)
		if specLit == nil {
			continue
		}

		for _, specElt := range specLit.Elts {
			specKV, ok := specElt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			specKey, ok := specKV.Key.(*ast.Ident)
			if !ok || specKey.Name != "Template" {
				continue
			}

			templateLit := unwrapCompositeLit(specKV.Value)
			if templateLit == nil {
				continue
			}

			for _, templateElt := range templateLit.Elts {
				templateKV, ok := templateElt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				templateKey, ok := templateKV.Key.(*ast.Ident)
				if !ok || templateKey.Name != "Spec" {
					continue
				}

				podSpecLit := unwrapCompositeLit(templateKV.Value)
				if podSpecLit == nil {
					continue
				}

				for _, podSpecElt := range podSpecLit.Elts {
					podSpecKV, ok := podSpecElt.(*ast.KeyValueExpr)
					if !ok {
						continue
					}

					podSpecKey, ok := podSpecKV.Key.(*ast.Ident)
					if !ok || podSpecKey.Name != "Affinity" {
						continue
					}

					affinityLit := unwrapCompositeLit(podSpecKV.Value)
					if affinityLit == nil {
						continue
					}

					for _, affinityElt := range affinityLit.Elts {
						affinityKV, ok := affinityElt.(*ast.KeyValueExpr)
						if !ok {
							continue
						}

						affinityKey, ok := affinityKV.Key.(*ast.Ident)
						if !ok {
							continue
						}

						if affinityKey.Name == "PodAntiAffinity" {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// RuleWK8401 checks for file size limits (resources per file).
func RuleWK8401() Rule {
	return Rule{
		ID:          "WK8401",
		Name:        "File size limits",
		Description: "Files should not exceed 20 resources",
		Severity:    SeverityWarning,
		Check:       checkWK8401,
		Fix:         nil, // No auto-fix available
	}
}

// MaxResourcesPerFile is the maximum recommended resources per file.
const MaxResourcesPerFile = 20

func checkWK8401(file *ast.File, fset *token.FileSet) []Issue {
	var issues []Issue

	// Count K8s resources in the file
	resourceCount := 0

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

			// Check if any value is a K8s resource type
			for _, value := range valueSpec.Values {
				if isK8sResourceType(value) {
					resourceCount++
				}
			}
		}
	}

	if resourceCount > MaxResourcesPerFile {
		pos := fset.Position(file.Package)
		issues = append(issues, Issue{
			Rule:     "WK8401",
			Message:  fmt.Sprintf("File contains %d resources (max %d), consider splitting into smaller files", resourceCount, MaxResourcesPerFile),
			File:     pos.Filename,
			Line:     1,
			Column:   1,
			Severity: SeverityWarning,
		})
	}

	return issues
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
