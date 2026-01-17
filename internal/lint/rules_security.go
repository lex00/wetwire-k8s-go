package lint

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

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
