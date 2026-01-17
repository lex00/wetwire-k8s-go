package lint

import (
	"go/ast"
	"go/token"
)

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
