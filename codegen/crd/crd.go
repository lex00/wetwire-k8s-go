package crd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-k8s-go/codegen/parse"
	"gopkg.in/yaml.v3"
)

// Parser handles parsing of CRD YAML files into resource types.
type Parser struct {
	// Domain is an optional domain prefix for organizing generated types
	// (e.g., "cnrm" for Config Connector, "istio" for Istio)
	Domain string
}

// NewParser creates a new CRD parser.
func NewParser(domain string) *Parser {
	return &Parser{Domain: domain}
}

// ParseFile parses a single CRD YAML file.
func (p *Parser) ParseFile(path string) ([]parse.ResourceType, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.ParseBytes(data)
}

// ParseBytes parses CRD YAML data (may contain multiple documents).
func (p *Parser) ParseBytes(data []byte) ([]parse.ResourceType, error) {
	var resources []parse.ResourceType

	// Split multi-document YAML
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))

	for {
		var crd CRD
		err := decoder.Decode(&crd)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("failed to decode YAML: %w", err)
		}

		// Skip if not a CRD
		if crd.Kind != "CustomResourceDefinition" {
			continue
		}

		// Parse each version
		for _, version := range crd.Spec.Versions {
			if !version.Served {
				continue // Skip versions that aren't served
			}

			resource, err := p.parseCRDVersion(crd, version)
			if err != nil {
				return nil, fmt.Errorf("failed to parse version %s: %w", version.Name, err)
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// ParseDirectory parses all CRD YAML files in a directory.
func (p *Parser) ParseDirectory(dir string) ([]parse.ResourceType, error) {
	var resources []parse.ResourceType

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process YAML files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		parsed, err := p.ParseFile(path)
		if err != nil {
			// Log warning but continue with other files
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", path, err)
			return nil
		}

		resources = append(resources, parsed...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return resources, nil
}

// parseCRDVersion converts a CRD version to a ResourceType.
func (p *Parser) parseCRDVersion(crd CRD, version CRDVersion) (parse.ResourceType, error) {
	resource := parse.ResourceType{
		Kind:        crd.Spec.Names.Kind,
		Group:       extractShortGroup(crd.Spec.Group),
		Version:     version.Name,
		Description: "",
		Properties:  make(map[string]parse.PropertyInfo),
		Required:    []string{},
	}

	// Set definition name for reference
	resource.DefinitionName = fmt.Sprintf("%s.%s.%s",
		crd.Spec.Group, version.Name, crd.Spec.Names.Kind)

	// Parse schema if present
	if version.Schema != nil && version.Schema.OpenAPIV3Schema != nil {
		schema := version.Schema.OpenAPIV3Schema

		// Get description from schema
		if schema.Description != "" {
			resource.Description = schema.Description
		}

		// Parse top-level properties
		if schema.Properties != nil {
			// Look for spec and status properties specifically
			for propName, propSchema := range schema.Properties {
				propInfo := p.parseProperty(propSchema)
				resource.Properties[propName] = propInfo
			}

			// Also parse nested spec properties for direct access
			if specSchema, ok := schema.Properties["spec"]; ok && specSchema != nil {
				for propName, propSchema := range specSchema.Properties {
					propInfo := p.parseProperty(propSchema)
					// Prefix with "spec." for nested properties
					resource.Properties["spec."+propName] = propInfo
				}
				resource.Required = append(resource.Required, specSchema.Required...)
			}
		}

		// Collect required fields
		resource.Required = append(resource.Required, schema.Required...)
	}

	return resource, nil
}

// parseProperty converts an OpenAPI v3 schema to PropertyInfo.
func (p *Parser) parseProperty(schema *OpenAPIV3Schema) parse.PropertyInfo {
	if schema == nil {
		return parse.PropertyInfo{Type: "interface{}", GoType: "interface{}"}
	}

	info := parse.PropertyInfo{
		Type:        schema.Type,
		Format:      schema.Format,
		Description: schema.Description,
		Default:     schema.Default,
	}

	// Map OpenAPI type to Go type
	info.GoType = p.mapTypeToGo(schema)

	// Handle array items
	if schema.Items != nil {
		items := p.parseProperty(schema.Items)
		info.Items = &items
	}

	// Handle map values
	if schema.AdditionalProperties != nil {
		additionalProps := p.parseProperty(schema.AdditionalProperties)
		info.AdditionalProperties = &additionalProps
	}

	return info
}

// mapTypeToGo maps OpenAPI v3 types to Go types.
func (p *Parser) mapTypeToGo(schema *OpenAPIV3Schema) string {
	if schema == nil {
		return "interface{}"
	}

	// Handle x-kubernetes-int-or-string
	if schema.XKubernetesIntOrString {
		return "intstr.IntOrString"
	}

	// Handle x-kubernetes-embedded-resource
	if schema.XKubernetesEmbeddedResource {
		return "runtime.RawExtension"
	}

	switch schema.Type {
	case "string":
		return "string"

	case "integer":
		switch schema.Format {
		case "int64":
			return "int64"
		case "int32", "":
			return "int32"
		default:
			return "int32"
		}

	case "number":
		switch schema.Format {
		case "double":
			return "float64"
		case "float":
			return "float32"
		default:
			return "float64"
		}

	case "boolean":
		return "bool"

	case "array":
		if schema.Items != nil {
			itemType := p.mapTypeToGo(schema.Items)
			return "[]" + itemType
		}
		return "[]interface{}"

	case "object":
		// Check for additionalProperties (map type)
		if schema.AdditionalProperties != nil {
			valueType := p.mapTypeToGo(schema.AdditionalProperties)
			return "map[string]" + valueType
		}

		// Check for nested properties (struct type)
		if len(schema.Properties) > 0 {
			// This would need to generate a nested struct
			// For now, return a generic type
			return "map[string]interface{}"
		}

		// Generic object
		return "map[string]interface{}"

	default:
		return "interface{}"
	}
}

// extractShortGroup extracts a short group name from a full group.
// E.g., "compute.cnrm.cloud.google.com" -> "compute"
func extractShortGroup(group string) string {
	parts := strings.Split(group, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return group
}

// ResourceTypeExtended extends ResourceType with CRD-specific information.
type ResourceTypeExtended struct {
	parse.ResourceType

	// FullGroup is the full API group (e.g., "compute.cnrm.cloud.google.com")
	FullGroup string

	// Domain is the domain prefix for organizing types
	Domain string

	// Plural is the plural resource name
	Plural string

	// Scope is "Namespaced" or "Cluster"
	Scope string
}

// ParseFileExtended parses a CRD file with extended information.
func (p *Parser) ParseFileExtended(path string) ([]ResourceTypeExtended, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.ParseBytesExtended(data)
}

// ParseBytesExtended parses CRD YAML with extended information.
func (p *Parser) ParseBytesExtended(data []byte) ([]ResourceTypeExtended, error) {
	var resources []ResourceTypeExtended

	decoder := yaml.NewDecoder(strings.NewReader(string(data)))

	for {
		var crd CRD
		err := decoder.Decode(&crd)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("failed to decode YAML: %w", err)
		}

		if crd.Kind != "CustomResourceDefinition" {
			continue
		}

		for _, version := range crd.Spec.Versions {
			if !version.Served {
				continue
			}

			baseResource, err := p.parseCRDVersion(crd, version)
			if err != nil {
				return nil, fmt.Errorf("failed to parse version %s: %w", version.Name, err)
			}

			extended := ResourceTypeExtended{
				ResourceType: baseResource,
				FullGroup:    crd.Spec.Group,
				Domain:       p.Domain,
				Plural:       crd.Spec.Names.Plural,
				Scope:        crd.Spec.Scope,
			}

			resources = append(resources, extended)
		}
	}

	return resources, nil
}

// Package returns the package path for a CRD resource type.
// For CRDs, this includes the domain if set.
// E.g., domain="cnrm" group="compute" version="v1beta1" -> "cnrm/compute/v1beta1"
func (r *ResourceTypeExtended) Package() string {
	if r.Domain != "" {
		return fmt.Sprintf("%s/%s/%s", r.Domain, r.Group, r.Version)
	}
	return fmt.Sprintf("%s/%s", r.Group, r.Version)
}

// APIVersion returns the full API version string.
// E.g., "compute.cnrm.cloud.google.com/v1beta1"
func (r *ResourceTypeExtended) APIVersion() string {
	return fmt.Sprintf("%s/%s", r.FullGroup, r.Version)
}
