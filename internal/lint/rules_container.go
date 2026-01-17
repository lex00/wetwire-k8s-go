package lint

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

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
