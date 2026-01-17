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
