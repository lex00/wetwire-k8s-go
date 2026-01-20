package crd

import (
	"testing"
)

// Sample Config Connector CRD for testing
const sampleComputeInstanceCRD = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: computeinstances.compute.cnrm.cloud.google.com
  labels:
    cnrm.cloud.google.com/managed-by-kcc: "true"
spec:
  group: compute.cnrm.cloud.google.com
  names:
    kind: ComputeInstance
    plural: computeinstances
    singular: computeinstance
    shortNames:
    - gcpcomputeinstance
    categories:
    - gcp
  scope: Namespaced
  versions:
  - name: v1beta1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        description: A Google Compute Engine VM instance.
        properties:
          apiVersion:
            type: string
            description: 'APIVersion defines the versioned schema'
          kind:
            type: string
            description: 'Kind is a string value representing the REST resource'
          metadata:
            type: object
          spec:
            type: object
            description: ComputeInstanceSpec defines the desired state
            required:
            - zone
            properties:
              machineType:
                type: string
                description: The machine type to create
              zone:
                type: string
                description: The zone that the machine should be created in
              bootDisk:
                type: object
                description: The boot disk for the instance
                properties:
                  autoDelete:
                    type: boolean
                    description: Whether the disk will be auto-deleted
                  initializeParams:
                    type: object
                    properties:
                      size:
                        type: integer
                        format: int64
                        description: The size of the image in gigabytes
                      sourceImage:
                        type: string
                        description: The image from which to initialize this disk
              networkInterfaces:
                type: array
                description: Networks to attach to the instance
                items:
                  type: object
                  properties:
                    networkRef:
                      type: object
                      description: Reference to the network
                      properties:
                        name:
                          type: string
                        namespace:
                          type: string
                        external:
                          type: string
              labels:
                type: object
                additionalProperties:
                  type: string
                description: Labels to apply to the instance
          status:
            type: object
            properties:
              instanceId:
                type: string
              selfLink:
                type: string
              conditions:
                type: array
                items:
                  type: object
                  properties:
                    type:
                      type: string
                    status:
                      type: string
                    reason:
                      type: string
                    message:
                      type: string
`

func TestParseCRD(t *testing.T) {
	parser := NewParser("cnrm")

	resources, err := parser.ParseBytes([]byte(sampleComputeInstanceCRD))
	if err != nil {
		t.Fatalf("ParseBytes failed: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("Expected 1 resource, got %d", len(resources))
	}

	resource := resources[0]

	// Check basic info
	if resource.Kind != "ComputeInstance" {
		t.Errorf("Expected Kind 'ComputeInstance', got '%s'", resource.Kind)
	}

	if resource.Group != "compute" {
		t.Errorf("Expected Group 'compute', got '%s'", resource.Group)
	}

	if resource.Version != "v1beta1" {
		t.Errorf("Expected Version 'v1beta1', got '%s'", resource.Version)
	}

	// Check that we have some properties
	if len(resource.Properties) == 0 {
		t.Error("Expected properties to be parsed")
	}

	// Check for spec property
	if _, ok := resource.Properties["spec"]; !ok {
		t.Error("Expected 'spec' property")
	}
}

func TestParseCRDExtended(t *testing.T) {
	parser := NewParser("cnrm")

	resources, err := parser.ParseBytesExtended([]byte(sampleComputeInstanceCRD))
	if err != nil {
		t.Fatalf("ParseBytesExtended failed: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("Expected 1 resource, got %d", len(resources))
	}

	resource := resources[0]

	// Check extended info
	if resource.FullGroup != "compute.cnrm.cloud.google.com" {
		t.Errorf("Expected FullGroup 'compute.cnrm.cloud.google.com', got '%s'", resource.FullGroup)
	}

	if resource.Domain != "cnrm" {
		t.Errorf("Expected Domain 'cnrm', got '%s'", resource.Domain)
	}

	if resource.Plural != "computeinstances" {
		t.Errorf("Expected Plural 'computeinstances', got '%s'", resource.Plural)
	}

	if resource.Scope != "Namespaced" {
		t.Errorf("Expected Scope 'Namespaced', got '%s'", resource.Scope)
	}

	// Check package path
	expectedPkg := "cnrm/compute/v1beta1"
	if resource.Package() != expectedPkg {
		t.Errorf("Expected Package '%s', got '%s'", expectedPkg, resource.Package())
	}

	// Check API version
	expectedAPIVersion := "compute.cnrm.cloud.google.com/v1beta1"
	if resource.APIVersion() != expectedAPIVersion {
		t.Errorf("Expected APIVersion '%s', got '%s'", expectedAPIVersion, resource.APIVersion())
	}
}

func TestMapTypeToGo(t *testing.T) {
	parser := NewParser("")

	tests := []struct {
		name     string
		schema   *OpenAPIV3Schema
		expected string
	}{
		{
			name:     "string",
			schema:   &OpenAPIV3Schema{Type: "string"},
			expected: "string",
		},
		{
			name:     "integer default",
			schema:   &OpenAPIV3Schema{Type: "integer"},
			expected: "int32",
		},
		{
			name:     "integer int64",
			schema:   &OpenAPIV3Schema{Type: "integer", Format: "int64"},
			expected: "int64",
		},
		{
			name:     "boolean",
			schema:   &OpenAPIV3Schema{Type: "boolean"},
			expected: "bool",
		},
		{
			name:     "array of strings",
			schema:   &OpenAPIV3Schema{Type: "array", Items: &OpenAPIV3Schema{Type: "string"}},
			expected: "[]string",
		},
		{
			name:     "map of strings",
			schema:   &OpenAPIV3Schema{Type: "object", AdditionalProperties: &OpenAPIV3Schema{Type: "string"}},
			expected: "map[string]string",
		},
		{
			name:     "int-or-string",
			schema:   &OpenAPIV3Schema{XKubernetesIntOrString: true},
			expected: "intstr.IntOrString",
		},
		{
			name:     "embedded resource",
			schema:   &OpenAPIV3Schema{XKubernetesEmbeddedResource: true},
			expected: "runtime.RawExtension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.mapTypeToGo(tt.schema)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestExtractShortGroup(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"compute.cnrm.cloud.google.com", "compute"},
		{"storage.cnrm.cloud.google.com", "storage"},
		{"sql.cnrm.cloud.google.com", "sql"},
		{"apps", "apps"},
		{"", ""},
	}

	for _, tt := range tests {
		result := extractShortGroup(tt.input)
		if result != tt.expected {
			t.Errorf("extractShortGroup(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseCRDSource(t *testing.T) {
	tests := []struct {
		input        string
		expectedType string
		expectedPath string
		expectedURL  string
	}{
		{"./crds", "directory", "./crds", ""},
		{"/path/to/crds", "directory", "/path/to/crds", ""},
		{"config-connector", "github", "", ConfigConnectorSource},
		{"https://example.com/crd.yaml", "url", "", "https://example.com/crd.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			source := ParseCRDSource(tt.input)

			if source.Type != tt.expectedType {
				t.Errorf("Type: got %q, want %q", source.Type, tt.expectedType)
			}

			if tt.expectedPath != "" && source.Path != tt.expectedPath {
				t.Errorf("Path: got %q, want %q", source.Path, tt.expectedPath)
			}

			if tt.expectedURL != "" && source.URL != tt.expectedURL {
				t.Errorf("URL: got %q, want %q", source.URL, tt.expectedURL)
			}
		})
	}
}

// TestMultiVersionCRD tests parsing a CRD with multiple versions
const multiVersionCRD = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: examples.test.example.com
spec:
  group: test.example.com
  names:
    kind: Example
    plural: examples
  scope: Namespaced
  versions:
  - name: v1alpha1
    served: true
    storage: false
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              name:
                type: string
  - name: v1beta1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              name:
                type: string
              description:
                type: string
  - name: v1
    served: false
    storage: false
`

func TestMultiVersionCRD(t *testing.T) {
	parser := NewParser("")

	resources, err := parser.ParseBytes([]byte(multiVersionCRD))
	if err != nil {
		t.Fatalf("ParseBytes failed: %v", err)
	}

	// Should have 2 resources (v1alpha1 and v1beta1 are served, v1 is not)
	if len(resources) != 2 {
		t.Fatalf("Expected 2 resources (served versions only), got %d", len(resources))
	}

	// Check versions
	versions := make(map[string]bool)
	for _, r := range resources {
		versions[r.Version] = true
	}

	if !versions["v1alpha1"] {
		t.Error("Expected v1alpha1 version")
	}
	if !versions["v1beta1"] {
		t.Error("Expected v1beta1 version")
	}
	if versions["v1"] {
		t.Error("Did not expect v1 version (served=false)")
	}
}
