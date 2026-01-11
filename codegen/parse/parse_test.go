package parse

import (
	"testing"

	"github.com/lex00/wetwire-k8s-go/codegen/fetch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseResourceTypes(t *testing.T) {
	schema := &fetch.Schema{
		Swagger: "2.0",
		Info: fetch.Info{
			Title:   "Kubernetes",
			Version: "v1.28.0",
		},
		Definitions: map[string]fetch.Definition{
			"io.k8s.api.core.v1.Pod": {
				Type:        "object",
				Description: "Pod is a collection of containers",
				Properties: map[string]fetch.Property{
					"apiVersion": {
						Type:        "string",
						Description: "APIVersion defines the versioned schema",
					},
					"kind": {
						Type:        "string",
						Description: "Kind is a string value representing the REST resource",
					},
					"metadata": {
						Ref: "#/definitions/io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta",
					},
					"spec": {
						Ref: "#/definitions/io.k8s.api.core.v1.PodSpec",
					},
				},
				XKubernetesGroupVersionKind: []fetch.GroupVersionKind{
					{Group: "", Kind: "Pod", Version: "v1"},
				},
			},
			"io.k8s.api.apps.v1.Deployment": {
				Type:        "object",
				Description: "Deployment enables declarative updates for Pods",
				Properties: map[string]fetch.Property{
					"apiVersion": {Type: "string"},
					"kind":       {Type: "string"},
					"metadata":   {Ref: "#/definitions/io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta"},
					"spec":       {Ref: "#/definitions/io.k8s.api.apps.v1.DeploymentSpec"},
				},
				XKubernetesGroupVersionKind: []fetch.GroupVersionKind{
					{Group: "apps", Kind: "Deployment", Version: "v1"},
				},
			},
		},
	}

	parser := NewParser()
	resources, err := parser.ParseResourceTypes(schema)

	require.NoError(t, err)
	assert.Len(t, resources, 2)

	// Check Pod resource
	pod := findResource(resources, "Pod", "", "v1")
	require.NotNil(t, pod, "Pod resource not found")
	assert.Equal(t, "Pod", pod.Kind)
	assert.Equal(t, "", pod.Group)
	assert.Equal(t, "v1", pod.Version)
	assert.Equal(t, "Pod is a collection of containers", pod.Description)
	assert.NotEmpty(t, pod.Properties)

	// Check Deployment resource
	deployment := findResource(resources, "Deployment", "apps", "v1")
	require.NotNil(t, deployment, "Deployment resource not found")
	assert.Equal(t, "Deployment", deployment.Kind)
	assert.Equal(t, "apps", deployment.Group)
	assert.Equal(t, "v1", deployment.Version)
}

func TestParseProperty(t *testing.T) {
	tests := []struct {
		name     string
		property fetch.Property
		expected PropertyInfo
	}{
		{
			name: "string property",
			property: fetch.Property{
				Type:        "string",
				Description: "A string field",
			},
			expected: PropertyInfo{
				Type:        "string",
				Description: "A string field",
				GoType:      "string",
			},
		},
		{
			name: "integer property",
			property: fetch.Property{
				Type:        "integer",
				Format:      "int32",
				Description: "An integer field",
			},
			expected: PropertyInfo{
				Type:        "integer",
				Format:      "int32",
				Description: "An integer field",
				GoType:      "int32",
			},
		},
		{
			name: "boolean property",
			property: fetch.Property{
				Type:        "boolean",
				Description: "A boolean field",
			},
			expected: PropertyInfo{
				Type:        "boolean",
				Description: "A boolean field",
				GoType:      "bool",
			},
		},
		{
			name: "array property",
			property: fetch.Property{
				Type: "array",
				Items: &fetch.Property{
					Type: "string",
				},
				Description: "An array of strings",
			},
			expected: PropertyInfo{
				Type:        "array",
				Description: "An array of strings",
				GoType:      "[]string",
				Items: &PropertyInfo{
					Type:   "string",
					GoType: "string",
				},
			},
		},
		{
			name: "reference property",
			property: fetch.Property{
				Ref:         "#/definitions/io.k8s.api.core.v1.PodSpec",
				Description: "A reference to PodSpec",
			},
			expected: PropertyInfo{
				Type:        "ref",
				Ref:         "io.k8s.api.core.v1.PodSpec",
				Description: "A reference to PodSpec",
				GoType:      "PodSpec",
			},
		},
	}

	parser := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ParseProperty(tt.property)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.GoType, result.GoType)
			assert.Equal(t, tt.expected.Description, result.Description)
			if tt.expected.Items != nil {
				require.NotNil(t, result.Items)
				assert.Equal(t, tt.expected.Items.GoType, result.Items.GoType)
			}
		})
	}
}

func TestParseResourceName(t *testing.T) {
	tests := []struct {
		name           string
		definitionName string
		expectedGroup  string
		expectedKind   string
		expectedVer    string
	}{
		{
			name:           "core v1 Pod",
			definitionName: "io.k8s.api.core.v1.Pod",
			expectedGroup:  "core",
			expectedKind:   "Pod",
			expectedVer:    "v1",
		},
		{
			name:           "apps v1 Deployment",
			definitionName: "io.k8s.api.apps.v1.Deployment",
			expectedGroup:  "apps",
			expectedKind:   "Deployment",
			expectedVer:    "v1",
		},
		{
			name:           "batch v1 Job",
			definitionName: "io.k8s.api.batch.v1.Job",
			expectedGroup:  "batch",
			expectedKind:   "Job",
			expectedVer:    "v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, kind, version := parseResourceName(tt.definitionName)
			assert.Equal(t, tt.expectedGroup, group)
			assert.Equal(t, tt.expectedKind, kind)
			assert.Equal(t, tt.expectedVer, version)
		})
	}
}

func TestFilterResourceTypes(t *testing.T) {
	schema := &fetch.Schema{
		Definitions: map[string]fetch.Definition{
			"io.k8s.api.core.v1.Pod": {
				Type: "object",
				XKubernetesGroupVersionKind: []fetch.GroupVersionKind{
					{Group: "", Kind: "Pod", Version: "v1"},
				},
			},
			"io.k8s.api.core.v1.PodSpec": {
				Type: "object",
				// No GVK - this is a supporting type, not a resource
			},
			"io.k8s.api.apps.v1.Deployment": {
				Type: "object",
				XKubernetesGroupVersionKind: []fetch.GroupVersionKind{
					{Group: "apps", Kind: "Deployment", Version: "v1"},
				},
			},
		},
	}

	parser := NewParser()
	resources, err := parser.ParseResourceTypes(schema)

	require.NoError(t, err)
	// Should only include types with GVK (actual resources)
	assert.Len(t, resources, 2)
}

// Helper function to find a resource in a slice
func findResource(resources []ResourceType, kind, group, version string) *ResourceType {
	for _, r := range resources {
		if r.Kind == kind && r.Group == group && r.Version == version {
			return &r
		}
	}
	return nil
}
