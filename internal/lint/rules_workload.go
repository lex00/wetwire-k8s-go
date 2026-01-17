package lint

import (
	"fmt"
	"go/ast"
	"go/token"
)

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
