package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lex00/wetwire-k8s-go/codegen/parse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateResourceFile(t *testing.T) {
	resource := parse.ResourceType{
		Kind:        "Pod",
		Group:       "core",
		Version:     "v1",
		Description: "Pod is a collection of containers that can run on a host.",
		Properties: map[string]parse.PropertyInfo{
			"apiVersion": {
				Type:        "string",
				GoType:      "string",
				Description: "APIVersion defines the versioned schema",
			},
			"kind": {
				Type:        "string",
				GoType:      "string",
				Description: "Kind is a string value representing the REST resource",
			},
			"metadata": {
				Type:        "ref",
				Ref:         "io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta",
				GoType:      "ObjectMeta",
				Description: "Standard object's metadata",
			},
			"spec": {
				Type:        "ref",
				Ref:         "io.k8s.api.core.v1.PodSpec",
				GoType:      "PodSpec",
				Description: "Specification of the desired behavior of the pod",
			},
		},
		Required: []string{"spec"},
	}

	tmpDir := t.TempDir()
	generator := NewGenerator(tmpDir)

	err := generator.GenerateResourceFile(resource)
	require.NoError(t, err)

	// Check that file was created
	expectedPath := filepath.Join(tmpDir, "core", "v1", "pod.go")
	assert.FileExists(t, expectedPath)

	// Read and verify content
	content, err := os.ReadFile(expectedPath)
	require.NoError(t, err)

	contentStr := string(content)
	// Check package declaration
	assert.Contains(t, contentStr, "package v1")

	// Check struct definition
	assert.Contains(t, contentStr, "type Pod struct")
	assert.Contains(t, contentStr, "APIVersion string")
	assert.Contains(t, contentStr, "Kind string")
	assert.Contains(t, contentStr, "Metadata ObjectMeta")
	assert.Contains(t, contentStr, "Spec PodSpec")

	// Check JSON tags
	assert.Contains(t, contentStr, "`json:\"apiVersion,omitempty\" yaml:\"apiVersion,omitempty\"`")
	assert.Contains(t, contentStr, "`json:\"kind,omitempty\" yaml:\"kind,omitempty\"`")
	assert.Contains(t, contentStr, "`json:\"metadata,omitempty\" yaml:\"metadata,omitempty\"`")
	assert.Contains(t, contentStr, "`json:\"spec\" yaml:\"spec\"`") // required field, no omitempty
}

func TestGenerateMultipleResources(t *testing.T) {
	resources := []parse.ResourceType{
		{
			Kind:        "Pod",
			Group:       "core",
			Version:     "v1",
			Description: "Pod is a collection of containers",
			Properties: map[string]parse.PropertyInfo{
				"apiVersion": {Type: "string", GoType: "string"},
				"kind":       {Type: "string", GoType: "string"},
			},
		},
		{
			Kind:        "Deployment",
			Group:       "apps",
			Version:     "v1",
			Description: "Deployment enables declarative updates",
			Properties: map[string]parse.PropertyInfo{
				"apiVersion": {Type: "string", GoType: "string"},
				"kind":       {Type: "string", GoType: "string"},
			},
		},
		{
			Kind:        "Service",
			Group:       "core",
			Version:     "v1",
			Description: "Service is a named abstraction of software service",
			Properties: map[string]parse.PropertyInfo{
				"apiVersion": {Type: "string", GoType: "string"},
				"kind":       {Type: "string", GoType: "string"},
			},
		},
	}

	tmpDir := t.TempDir()
	generator := NewGenerator(tmpDir)

	err := generator.GenerateResources(resources)
	require.NoError(t, err)

	// Check that files were created
	assert.FileExists(t, filepath.Join(tmpDir, "core", "v1", "pod.go"))
	assert.FileExists(t, filepath.Join(tmpDir, "apps", "v1", "deployment.go"))
	assert.FileExists(t, filepath.Join(tmpDir, "core", "v1", "service.go"))
}

func TestFormatFieldName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"apiVersion", "APIVersion"},
		{"kind", "Kind"},
		{"metadata", "Metadata"},
		{"spec", "Spec"},
		{"status", "Status"},
		{"replicas", "Replicas"},
		{"containerPort", "ContainerPort"},
		{"nodeSelector", "NodeSelector"},
		{"podIP", "PodIP"},
		{"serviceAccount", "ServiceAccount"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatFieldName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateStructTag(t *testing.T) {
	tests := []struct {
		name       string
		fieldName  string
		isRequired bool
		expected   string
	}{
		{
			name:       "required field",
			fieldName:  "spec",
			isRequired: true,
			expected:   "`json:\"spec\" yaml:\"spec\"`",
		},
		{
			name:       "optional field",
			fieldName:  "metadata",
			isRequired: false,
			expected:   "`json:\"metadata,omitempty\" yaml:\"metadata,omitempty\"`",
		},
		{
			name:       "apiVersion",
			fieldName:  "apiVersion",
			isRequired: false,
			expected:   "`json:\"apiVersion,omitempty\" yaml:\"apiVersion,omitempty\"`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateStructTag(tt.fieldName, tt.isRequired)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateResourceWithArrayAndMap(t *testing.T) {
	resource := parse.ResourceType{
		Kind:        "ConfigMap",
		Group:       "core",
		Version:     "v1",
		Description: "ConfigMap holds configuration data",
		Properties: map[string]parse.PropertyInfo{
			"apiVersion": {Type: "string", GoType: "string"},
			"kind":       {Type: "string", GoType: "string"},
			"data": {
				Type:   "object",
				GoType: "map[string]string",
			},
			"binaryData": {
				Type:   "object",
				GoType: "map[string][]byte",
			},
		},
	}

	tmpDir := t.TempDir()
	generator := NewGenerator(tmpDir)

	err := generator.GenerateResourceFile(resource)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "core", "v1", "configmap.go"))
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "Data map[string]string")
	assert.Contains(t, contentStr, "BinaryData map[string][]byte")
}

func TestGeneratePackageForVersion(t *testing.T) {
	resources := []parse.ResourceType{
		{Kind: "Pod", Group: "core", Version: "v1"},
		{Kind: "Service", Group: "core", Version: "v1"},
		{Kind: "Deployment", Group: "apps", Version: "v1"},
	}

	tmpDir := t.TempDir()
	generator := NewGenerator(tmpDir)

	err := generator.GenerateResources(resources)
	require.NoError(t, err)

	// Check that package directories were created correctly
	assert.DirExists(t, filepath.Join(tmpDir, "core", "v1"))
	assert.DirExists(t, filepath.Join(tmpDir, "apps", "v1"))
}

func TestGenerateWithComments(t *testing.T) {
	resource := parse.ResourceType{
		Kind:        "Pod",
		Group:       "core",
		Version:     "v1",
		Description: "Pod is a collection of containers that can run on a host. This comment has multiple lines.",
		Properties: map[string]parse.PropertyInfo{
			"apiVersion": {
				Type:        "string",
				GoType:      "string",
				Description: "APIVersion defines the versioned schema of this representation",
			},
		},
	}

	tmpDir := t.TempDir()
	generator := NewGenerator(tmpDir)

	err := generator.GenerateResourceFile(resource)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "core", "v1", "pod.go"))
	require.NoError(t, err)

	contentStr := string(content)
	// Check for struct comment
	assert.Contains(t, contentStr, "// Pod is a collection of containers")

	// Check for field comment
	assert.Contains(t, contentStr, "// APIVersion defines the versioned schema")
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Pod", "pod"},
		{"Deployment", "deployment"},
		{"StatefulSet", "statefulset"},
		{"DaemonSet", "daemonset"},
		{"ConfigMap", "configmap"},
		{"ServiceAccount", "serviceaccount"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGeneratedCodeCompiles(t *testing.T) {
	// This test would require actually compiling the generated code
	// For now, we just check that the generated code looks reasonable
	resource := parse.ResourceType{
		Kind:    "Pod",
		Group:   "core",
		Version: "v1",
		Properties: map[string]parse.PropertyInfo{
			"apiVersion": {Type: "string", GoType: "string"},
			"kind":       {Type: "string", GoType: "string"},
		},
	}

	tmpDir := t.TempDir()
	generator := NewGenerator(tmpDir)

	err := generator.GenerateResourceFile(resource)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "core", "v1", "pod.go"))
	require.NoError(t, err)

	contentStr := string(content)
	// Basic syntax checks
	assert.Equal(t, 1, strings.Count(contentStr, "package v1"))
	assert.True(t, strings.Count(contentStr, "{") == strings.Count(contentStr, "}"), "Braces should be balanced")
}
