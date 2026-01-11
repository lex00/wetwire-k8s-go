package parse

import (
	"fmt"
	"strings"

	"github.com/lex00/wetwire-k8s-go/codegen/fetch"
)

// Parser handles parsing of Kubernetes OpenAPI schemas into resource types.
type Parser struct{}

// NewParser creates a new schema parser.
func NewParser() *Parser {
	return &Parser{}
}

// ResourceType represents a parsed Kubernetes resource type.
type ResourceType struct {
	// Kind is the resource kind (e.g., "Pod", "Deployment")
	Kind string

	// Group is the API group (e.g., "core", "apps", "batch")
	Group string

	// Version is the API version (e.g., "v1", "v1beta1")
	Version string

	// Description is the resource description
	Description string

	// Properties contains the resource properties
	Properties map[string]PropertyInfo

	// Required lists required property names
	Required []string

	// DefinitionName is the full definition name (e.g., "io.k8s.api.core.v1.Pod")
	DefinitionName string
}

// PropertyInfo represents information about a property.
type PropertyInfo struct {
	// Type is the schema type (string, integer, boolean, array, object, ref)
	Type string

	// Format is the schema format (e.g., "int32", "int64", "date-time")
	Format string

	// Description is the property description
	Description string

	// GoType is the Go type for this property
	GoType string

	// Ref is the reference name if this is a $ref property
	Ref string

	// Items contains information about array items
	Items *PropertyInfo

	// AdditionalProperties contains information about map values
	AdditionalProperties *PropertyInfo

	// Default is the default value
	Default interface{}
}

// ParseResourceTypes extracts all Kubernetes resource types from a schema.
func (p *Parser) ParseResourceTypes(schema *fetch.Schema) ([]ResourceType, error) {
	var resources []ResourceType

	for defName, def := range schema.Definitions {
		// Only process types that have x-kubernetes-group-version-kind
		// These are actual Kubernetes resources
		if len(def.XKubernetesGroupVersionKind) == 0 {
			continue
		}

		// Parse properties
		properties := make(map[string]PropertyInfo)
		for propName, prop := range def.Properties {
			properties[propName] = p.ParseProperty(prop)
		}

		// Use the first GVK (most resources only have one)
		gvk := def.XKubernetesGroupVersionKind[0]

		resource := ResourceType{
			Kind:           gvk.Kind,
			Group:          gvk.Group,
			Version:        gvk.Version,
			Description:    def.Description,
			Properties:     properties,
			Required:       def.Required,
			DefinitionName: defName,
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// ParseProperty parses a schema property into PropertyInfo.
func (p *Parser) ParseProperty(prop fetch.Property) PropertyInfo {
	info := PropertyInfo{
		Type:        prop.Type,
		Format:      prop.Format,
		Description: prop.Description,
		Default:     prop.Default,
	}

	// Handle $ref
	if prop.Ref != "" {
		info.Type = "ref"
		info.Ref = extractRefName(prop.Ref)
		info.GoType = extractTypeName(info.Ref)
		return info
	}

	// Map OpenAPI type to Go type
	info.GoType = mapTypeToGo(prop)

	// Handle array items
	if prop.Items != nil {
		items := p.ParseProperty(*prop.Items)
		info.Items = &items
	}

	// Handle map values
	if prop.AdditionalProperties != nil {
		additionalProps := p.ParseProperty(*prop.AdditionalProperties)
		info.AdditionalProperties = &additionalProps
	}

	return info
}

// mapTypeToGo maps OpenAPI types to Go types.
func mapTypeToGo(prop fetch.Property) string {
	switch prop.Type {
	case "string":
		return "string"
	case "integer":
		if prop.Format == "int64" {
			return "int64"
		}
		return "int32"
	case "number":
		if prop.Format == "double" {
			return "float64"
		}
		return "float32"
	case "boolean":
		return "bool"
	case "array":
		if prop.Items != nil {
			itemType := mapTypeToGo(*prop.Items)
			if prop.Items.Ref != "" {
				itemType = extractTypeName(extractRefName(prop.Items.Ref))
			}
			return "[]" + itemType
		}
		return "[]interface{}"
	case "object":
		if prop.AdditionalProperties != nil {
			valueType := mapTypeToGo(*prop.AdditionalProperties)
			if prop.AdditionalProperties.Ref != "" {
				valueType = extractTypeName(extractRefName(prop.AdditionalProperties.Ref))
			}
			return "map[string]" + valueType
		}
		return "map[string]interface{}"
	default:
		return "interface{}"
	}
}

// extractRefName extracts the type name from a $ref string.
// E.g., "#/definitions/io.k8s.api.core.v1.PodSpec" -> "io.k8s.api.core.v1.PodSpec"
func extractRefName(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ref
}

// extractTypeName extracts the short type name from a full definition name.
// E.g., "io.k8s.api.core.v1.PodSpec" -> "PodSpec"
func extractTypeName(fullName string) string {
	parts := strings.Split(fullName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullName
}

// parseResourceName parses a definition name into group, kind, and version.
// E.g., "io.k8s.api.core.v1.Pod" -> ("core", "Pod", "v1")
func parseResourceName(defName string) (group, kind, version string) {
	// Expected format: io.k8s.api.<group>.<version>.<Kind>
	// or for core: io.k8s.api.core.<version>.<Kind>
	parts := strings.Split(defName, ".")
	if len(parts) < 6 {
		return "", "", ""
	}

	// Extract version (e.g., "v1", "v1beta1")
	version = parts[len(parts)-2]

	// Extract kind (last part)
	kind = parts[len(parts)-1]

	// Extract group (between "api" and version)
	// Find "api" index
	apiIndex := -1
	for i, part := range parts {
		if part == "api" {
			apiIndex = i
			break
		}
	}

	if apiIndex >= 0 && apiIndex+1 < len(parts) {
		group = parts[apiIndex+1]
	}

	return group, kind, version
}

// GroupVersion returns the group/version string for a resource.
func (r *ResourceType) GroupVersion() string {
	if r.Group == "" || r.Group == "core" {
		return r.Version
	}
	return r.Group + "/" + r.Version
}

// Package returns the package name for this resource.
// E.g., "core/v1", "apps/v1", "batch/v1"
func (r *ResourceType) Package() string {
	if r.Group == "" || r.Group == "core" {
		return fmt.Sprintf("core/%s", r.Version)
	}
	return fmt.Sprintf("%s/%s", r.Group, r.Version)
}
