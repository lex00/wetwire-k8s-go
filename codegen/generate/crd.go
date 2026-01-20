package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lex00/wetwire-k8s-go/codegen/crd"
	"github.com/lex00/wetwire-k8s-go/codegen/parse"
)

// CRDGenerator generates Go code from CRD definitions.
type CRDGenerator struct {
	outputDir string
	domain    string
}

// NewCRDGenerator creates a new CRD code generator.
func NewCRDGenerator(outputDir, domain string) *CRDGenerator {
	return &CRDGenerator{
		outputDir: outputDir,
		domain:    domain,
	}
}

// GenerateFromCRDDirectory generates Go code from all CRDs in a directory.
func (g *CRDGenerator) GenerateFromCRDDirectory(crdDir string) error {
	parser := crd.NewParser(g.domain)

	resources, err := parser.ParseBytesExtended(nil)
	if err != nil {
		return fmt.Errorf("parse CRDs: %w", err)
	}

	// Parse all CRD files
	files, err := crd.ListCRDFiles(crdDir)
	if err != nil {
		return fmt.Errorf("list CRD files: %w", err)
	}

	var allResources []crd.ResourceTypeExtended
	for _, file := range files {
		parsed, err := parser.ParseFileExtended(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", file, err)
			continue
		}
		allResources = append(allResources, parsed...)
	}

	if len(allResources) == 0 {
		return fmt.Errorf("no CRD resources found in %s", crdDir)
	}

	// Generate code for each resource
	for _, resource := range allResources {
		if err := g.generateCRDResourceFile(resource); err != nil {
			return fmt.Errorf("generate %s: %w", resource.Kind, err)
		}
	}

	// Generate package init files
	if err := g.generatePackageFiles(allResources); err != nil {
		return fmt.Errorf("generate package files: %w", err)
	}

	fmt.Printf("Generated %d resource types from CRDs\n", len(allResources))
	_ = resources // unused, but shows pattern compatibility

	return nil
}

// GenerateFromResources generates Go code from parsed CRD resources.
func (g *CRDGenerator) GenerateFromResources(resources []crd.ResourceTypeExtended) error {
	for _, resource := range resources {
		if err := g.generateCRDResourceFile(resource); err != nil {
			return fmt.Errorf("generate %s: %w", resource.Kind, err)
		}
	}

	// Generate package init files
	if err := g.generatePackageFiles(resources); err != nil {
		return fmt.Errorf("generate package files: %w", err)
	}

	return nil
}

// generateCRDResourceFile generates a Go source file for a single CRD resource.
func (g *CRDGenerator) generateCRDResourceFile(resource crd.ResourceTypeExtended) error {
	// Determine output path: {outputDir}/{domain}/{group}/{version}/{kind}.go
	pkgPath := resource.Package()
	outputPath := filepath.Join(g.outputDir, pkgPath, toSnakeCase(resource.Kind)+".go")

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Generate code
	code := g.generateCRDCode(resource)

	// Write file
	if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// generateCRDCode generates the Go source code for a CRD resource type.
func (g *CRDGenerator) generateCRDCode(resource crd.ResourceTypeExtended) string {
	var b strings.Builder

	// Package declaration
	b.WriteString("package ")
	b.WriteString(resource.Version)
	b.WriteString("\n\n")

	// Imports
	b.WriteString("import (\n")
	b.WriteString("\tmetav1 \"k8s.io/apimachinery/pkg/apis/meta/v1\"\n")
	b.WriteString(")\n\n")

	// Main resource struct comment
	if resource.Description != "" {
		writeComment(&b, resource.Description, "")
	} else {
		b.WriteString("// ")
		b.WriteString(resource.Kind)
		b.WriteString(" represents a ")
		b.WriteString(resource.FullGroup)
		b.WriteString(" ")
		b.WriteString(resource.Kind)
		b.WriteString(" resource.\n")
	}

	// Main resource struct
	b.WriteString("type ")
	b.WriteString(resource.Kind)
	b.WriteString(" struct {\n")
	b.WriteString("\tmetav1.TypeMeta   `json:\",inline\"`\n")
	b.WriteString("\tmetav1.ObjectMeta `json:\"metadata,omitempty\"`\n")
	b.WriteString("\n")
	b.WriteString("\tSpec   ")
	b.WriteString(resource.Kind)
	b.WriteString("Spec   `json:\"spec,omitempty\"`\n")
	b.WriteString("\tStatus ")
	b.WriteString(resource.Kind)
	b.WriteString("Status `json:\"status,omitempty\"`\n")
	b.WriteString("}\n\n")

	// Generate Spec struct
	g.generateSpecStruct(&b, resource)

	// Generate Status struct (placeholder)
	b.WriteString("// ")
	b.WriteString(resource.Kind)
	b.WriteString("Status defines the observed state of ")
	b.WriteString(resource.Kind)
	b.WriteString(".\n")
	b.WriteString("type ")
	b.WriteString(resource.Kind)
	b.WriteString("Status struct {\n")
	b.WriteString("\t// Conditions represent the latest available observations of the resource's state.\n")
	b.WriteString("\tConditions []metav1.Condition `json:\"conditions,omitempty\"`\n")
	b.WriteString("}\n")

	return b.String()
}

// generateSpecStruct generates the Spec struct for a CRD resource.
func (g *CRDGenerator) generateSpecStruct(b *strings.Builder, resource crd.ResourceTypeExtended) {
	b.WriteString("// ")
	b.WriteString(resource.Kind)
	b.WriteString("Spec defines the desired state of ")
	b.WriteString(resource.Kind)
	b.WriteString(".\n")
	b.WriteString("type ")
	b.WriteString(resource.Kind)
	b.WriteString("Spec struct {\n")

	// Get spec properties from the base ResourceType
	specProps := make(map[string]parse.PropertyInfo)
	for propName, prop := range resource.Properties {
		// Look for spec.* properties
		if strings.HasPrefix(propName, "spec.") {
			shortName := strings.TrimPrefix(propName, "spec.")
			specProps[shortName] = prop
		}
	}

	// Sort property names for consistent output
	propNames := make([]string, 0, len(specProps))
	for name := range specProps {
		propNames = append(propNames, name)
	}
	sort.Strings(propNames)

	// Generate fields
	for _, propName := range propNames {
		prop := specProps[propName]

		// Field comment
		if prop.Description != "" {
			writeComment(b, prop.Description, "\t")
		}

		// Field declaration
		b.WriteString("\t")
		b.WriteString(formatFieldName(propName))
		b.WriteString(" ")
		b.WriteString(g.mapCRDTypeToGo(prop))

		// Struct tag
		isRequired := false // TODO: check required fields
		b.WriteString(" ")
		b.WriteString(generateStructTag(propName, isRequired))
		b.WriteString("\n")
	}

	// If no spec properties found, add a placeholder
	if len(specProps) == 0 {
		b.WriteString("\t// TODO: Add spec fields from CRD schema\n")
	}

	b.WriteString("}\n\n")
}

// mapCRDTypeToGo maps a CRD property type to a Go type.
func (g *CRDGenerator) mapCRDTypeToGo(prop parse.PropertyInfo) string {
	// Use the GoType if already computed by the CRD parser
	if prop.GoType != "" {
		return prop.GoType
	}

	// Fallback mapping
	switch prop.Type {
	case "string":
		return "string"
	case "integer":
		if prop.Format == "int64" {
			return "int64"
		}
		return "int32"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	case "array":
		if prop.Items != nil {
			itemType := g.mapCRDTypeToGo(*prop.Items)
			return "[]" + itemType
		}
		return "[]interface{}"
	case "object":
		if prop.AdditionalProperties != nil {
			valueType := g.mapCRDTypeToGo(*prop.AdditionalProperties)
			return "map[string]" + valueType
		}
		return "map[string]interface{}"
	default:
		return "interface{}"
	}
}

// generatePackageFiles generates doc.go and register.go files for each package.
func (g *CRDGenerator) generatePackageFiles(resources []crd.ResourceTypeExtended) error {
	// Group resources by package
	packages := make(map[string][]crd.ResourceTypeExtended)
	for _, r := range resources {
		pkg := r.Package()
		packages[pkg] = append(packages[pkg], r)
	}

	// Generate files for each package
	for pkgPath, pkgResources := range packages {
		pkgDir := filepath.Join(g.outputDir, pkgPath)

		// Generate doc.go
		if err := g.generateDocFile(pkgDir, pkgResources); err != nil {
			return fmt.Errorf("generate doc.go for %s: %w", pkgPath, err)
		}

		// Generate register.go
		if err := g.generateRegisterFile(pkgDir, pkgResources); err != nil {
			return fmt.Errorf("generate register.go for %s: %w", pkgPath, err)
		}
	}

	return nil
}

// generateDocFile generates a doc.go file for a package.
func (g *CRDGenerator) generateDocFile(pkgDir string, resources []crd.ResourceTypeExtended) error {
	if len(resources) == 0 {
		return nil
	}

	first := resources[0]
	version := first.Version

	var b strings.Builder
	b.WriteString("// Package ")
	b.WriteString(version)
	b.WriteString(" contains ")
	b.WriteString(first.FullGroup)
	b.WriteString(" API types.\n")
	b.WriteString("//\n")
	b.WriteString("// This package is auto-generated from CRD schemas.\n")
	b.WriteString("// Do not edit manually.\n")
	b.WriteString("package ")
	b.WriteString(version)
	b.WriteString("\n")

	outputPath := filepath.Join(pkgDir, "doc.go")
	return os.WriteFile(outputPath, []byte(b.String()), 0644)
}

// generateRegisterFile generates a register.go file that registers types with the registry.
func (g *CRDGenerator) generateRegisterFile(pkgDir string, resources []crd.ResourceTypeExtended) error {
	if len(resources) == 0 {
		return nil
	}

	first := resources[0]
	version := first.Version
	pkgAlias := g.packageAlias(first)

	var b strings.Builder
	b.WriteString("package ")
	b.WriteString(version)
	b.WriteString("\n\n")

	b.WriteString("import (\n")
	b.WriteString("\t\"github.com/lex00/wetwire-k8s-go/internal/registry\"\n")
	b.WriteString(")\n\n")

	b.WriteString("func init() {\n")
	b.WriteString("\tregistry.DefaultRegistry.RegisterCRDTypes(\"")
	b.WriteString(g.domain)
	b.WriteString("\", []registry.CRDTypeInfo{\n")

	for _, r := range resources {
		b.WriteString("\t\t{\n")
		b.WriteString("\t\t\tPackage:    \"")
		b.WriteString(pkgAlias)
		b.WriteString("\",\n")
		b.WriteString("\t\t\tGroup:      \"")
		b.WriteString(r.FullGroup)
		b.WriteString("\",\n")
		b.WriteString("\t\t\tVersion:    \"")
		b.WriteString(r.Version)
		b.WriteString("\",\n")
		b.WriteString("\t\t\tKind:       \"")
		b.WriteString(r.Kind)
		b.WriteString("\",\n")
		b.WriteString("\t\t\tAPIVersion: \"")
		b.WriteString(r.APIVersion())
		b.WriteString("\",\n")
		b.WriteString("\t\t},\n")
	}

	b.WriteString("\t})\n")
	b.WriteString("}\n")

	outputPath := filepath.Join(pkgDir, "register.go")
	return os.WriteFile(outputPath, []byte(b.String()), 0644)
}

// packageAlias returns the package alias for a resource (e.g., "computev1beta1").
func (g *CRDGenerator) packageAlias(r crd.ResourceTypeExtended) string {
	return r.Group + r.Version
}
