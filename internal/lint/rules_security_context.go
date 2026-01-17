package lint

import (
	"go/ast"
	"go/token"
)

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
