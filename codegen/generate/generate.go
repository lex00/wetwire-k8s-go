package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lex00/wetwire-k8s-go/codegen/parse"
)

// Generator handles code generation for Kubernetes resource types.
type Generator struct {
	outputDir string
}

// NewGenerator creates a new code generator with the given output directory.
func NewGenerator(outputDir string) *Generator {
	return &Generator{
		outputDir: outputDir,
	}
}

// GenerateResources generates Go source files for all resource types.
func (g *Generator) GenerateResources(resources []parse.ResourceType) error {
	for _, resource := range resources {
		if err := g.GenerateResourceFile(resource); err != nil {
			return fmt.Errorf("failed to generate %s: %w", resource.Kind, err)
		}
	}
	return nil
}

// GenerateResourceFile generates a Go source file for a single resource type.
func (g *Generator) GenerateResourceFile(resource parse.ResourceType) error {
	// Determine output path
	pkgPath := resource.Package()
	outputPath := filepath.Join(g.outputDir, pkgPath, toSnakeCase(resource.Kind)+".go")

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate code
	code := g.generateCode(resource)

	// Write file
	if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// generateCode generates the Go source code for a resource type.
func (g *Generator) generateCode(resource parse.ResourceType) string {
	var b strings.Builder

	// Package declaration
	b.WriteString("package ")
	b.WriteString(resource.Version)
	b.WriteString("\n\n")

	// Struct comment
	if resource.Description != "" {
		writeComment(&b, resource.Description, "")
	}

	// Struct definition
	b.WriteString("type ")
	b.WriteString(resource.Kind)
	b.WriteString(" struct {\n")

	// Sort property names for consistent output
	propNames := make([]string, 0, len(resource.Properties))
	for name := range resource.Properties {
		propNames = append(propNames, name)
	}
	sort.Strings(propNames)

	// Generate fields
	for _, propName := range propNames {
		prop := resource.Properties[propName]

		// Field comment
		if prop.Description != "" {
			writeComment(&b, prop.Description, "\t")
		}

		// Field declaration
		b.WriteString("\t")
		b.WriteString(formatFieldName(propName))
		b.WriteString(" ")
		b.WriteString(prop.GoType)

		// Struct tag
		isRequired := contains(resource.Required, propName)
		b.WriteString(" ")
		b.WriteString(generateStructTag(propName, isRequired))
		b.WriteString("\n")
	}

	b.WriteString("}\n")

	return b.String()
}

// writeComment writes a comment to the builder, wrapping long lines.
func writeComment(b *strings.Builder, comment string, indent string) {
	// Clean up the comment
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return
	}

	// Simple comment writing (could be enhanced to wrap long lines)
	lines := strings.Split(comment, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			b.WriteString(indent)
			b.WriteString("// ")
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
}

// formatFieldName converts a JSON field name to a Go field name.
// E.g., "apiVersion" -> "APIVersion", "metadata" -> "Metadata"
func formatFieldName(name string) string {
	// Handle special cases
	switch name {
	case "apiVersion":
		return "APIVersion"
	case "podIP":
		return "PodIP"
	case "hostIP":
		return "HostIP"
	case "clusterIP":
		return "ClusterIP"
	case "targetCPUUtilizationPercentage":
		return "TargetCPUUtilizationPercentage"
	}

	// Convert to PascalCase
	parts := splitCamelCase(name)
	for i, part := range parts {
		// Capitalize known acronyms
		upper := strings.ToUpper(part)
		if isAcronym(upper) {
			parts[i] = upper
		} else {
			// Capitalize first letter
			if len(part) > 0 {
				parts[i] = strings.ToUpper(part[:1]) + part[1:]
			}
		}
	}

	return strings.Join(parts, "")
}

// splitCamelCase splits a camelCase string into parts.
func splitCamelCase(s string) []string {
	var parts []string
	var current strings.Builder

	for i, r := range s {
		if i > 0 && isUpper(r) && !isUpper(rune(s[i-1])) {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		}
		current.WriteRune(r)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// isUpper checks if a rune is uppercase.
func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// isAcronym checks if a string is a known acronym.
func isAcronym(s string) bool {
	acronyms := map[string]bool{
		"API": true,
		"IP":  true,
		"CPU": true,
		"DNS": true,
		"HTTP": true,
		"HTTPS": true,
		"TCP": true,
		"UDP": true,
		"UID": true,
		"GID": true,
		"URL": true,
		"URI": true,
		"ID": true,
	}
	return acronyms[s]
}

// generateStructTag generates a struct tag for a field.
func generateStructTag(fieldName string, isRequired bool) string {
	omitempty := ""
	if !isRequired {
		omitempty = ",omitempty"
	}

	return fmt.Sprintf("`json:\"%s%s\" yaml:\"%s%s\"`",
		fieldName, omitempty, fieldName, omitempty)
}

// toSnakeCase converts a PascalCase string to snake_case.
// E.g., "Pod" -> "pod", "StatefulSet" -> "statefulset"
func toSnakeCase(s string) string {
	// For simplicity, just lowercase for single-word resource types
	// In a real implementation, you'd handle multi-word properly
	return strings.ToLower(s)
}

// contains checks if a slice contains a string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
