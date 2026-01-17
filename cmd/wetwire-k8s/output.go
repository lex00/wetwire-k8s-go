package main

import (
	"strings"

	"github.com/lex00/wetwire-k8s-go/internal/discover"
	"github.com/lex00/wetwire-k8s-go/internal/serialize"
)

// generateOutput converts discovered resources to YAML or JSON output.
func generateOutput(resources []discover.Resource, format string) ([]byte, error) {
	var manifests []interface{}
	for _, r := range resources {
		manifest := createManifestFromResource(r)
		manifests = append(manifests, manifest)
	}

	if len(manifests) == 0 {
		return []byte{}, nil
	}

	if format == "json" {
		return serializeResourcesJSON(manifests)
	}
	return serializeResourcesYAML(manifests)
}

// createManifestFromResource creates a basic manifest map from discovered resource.
func createManifestFromResource(r discover.Resource) map[string]interface{} {
	apiVersion, kind := parseResourceType(r.Type)

	manifest := map[string]interface{}{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name": toKubernetesName(r.Name),
		},
	}

	return manifest
}

// parseResourceType extracts apiVersion and kind from a Go type string.
func parseResourceType(typeStr string) (string, string) {
	apiVersion := "v1"
	kind := "Unknown"

	parts := strings.Split(typeStr, ".")
	if len(parts) == 2 {
		pkg := parts[0]
		kind = parts[1]
		apiVersion = mapPackageToAPIVersion(pkg)
	} else if len(parts) == 1 {
		kind = parts[0]
	}

	return apiVersion, kind
}

// mapPackageToAPIVersion maps Go package aliases to Kubernetes API versions.
func mapPackageToAPIVersion(pkg string) string {
	packageMap := map[string]string{
		"corev1":         "v1",
		"appsv1":         "apps/v1",
		"batchv1":        "batch/v1",
		"networkingv1":   "networking.k8s.io/v1",
		"rbacv1":         "rbac.authorization.k8s.io/v1",
		"storagev1":      "storage.k8s.io/v1",
		"policyv1":       "policy/v1",
		"autoscalingv1":  "autoscaling/v1",
		"autoscalingv2":  "autoscaling/v2",
		"admissionv1":    "admissionregistration.k8s.io/v1",
		"certificatesv1": "certificates.k8s.io/v1",
	}

	if version, ok := packageMap[pkg]; ok {
		return version
	}
	return "v1"
}

// toKubernetesName converts a Go variable name to a Kubernetes resource name.
func toKubernetesName(name string) string {
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// serializeResourcesYAML converts resources to multi-document YAML.
func serializeResourcesYAML(resources []interface{}) ([]byte, error) {
	return serialize.ToMultiYAML(resources)
}

// serializeResourcesJSON converts resources to JSON array.
func serializeResourcesJSON(resources []interface{}) ([]byte, error) {
	if len(resources) == 1 {
		return serialize.ToJSON(resources[0])
	}

	var result []byte
	result = append(result, '[')
	for i, r := range resources {
		if i > 0 {
			result = append(result, ',', '\n')
		}
		jsonBytes, err := serialize.ToJSON(r)
		if err != nil {
			return nil, err
		}
		result = append(result, jsonBytes...)
	}
	result = append(result, ']')
	return result, nil
}
