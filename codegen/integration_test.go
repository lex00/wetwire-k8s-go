package codegen_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/lex00/wetwire-k8s-go/codegen/fetch"
	"github.com/lex00/wetwire-k8s-go/codegen/generate"
	"github.com/lex00/wetwire-k8s-go/codegen/parse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullCodegenWorkflow demonstrates the complete codegen workflow:
// 1. Fetch Kubernetes OpenAPI schema
// 2. Parse resource types from schema
// 3. Generate Go code for resources
func TestFullCodegenWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	schemasDir := filepath.Join(tmpDir, "schemas")
	resourcesDir := filepath.Join(tmpDir, "resources")

	// Step 1: Fetch schema
	t.Log("Fetching Kubernetes v1.28.0 schema...")
	fetcher := fetch.NewFetcher(schemasDir)
	ctx := context.Background()

	schema, err := fetcher.FetchSchema(ctx, "v1.28.0")
	require.NoError(t, err, "Failed to fetch schema")
	assert.NotNil(t, schema)
	assert.Equal(t, "2.0", schema.Swagger)
	assert.NotEmpty(t, schema.Definitions, "Schema should have definitions")

	t.Logf("Fetched schema with %d definitions", len(schema.Definitions))

	// Step 2: Parse resource types
	t.Log("Parsing resource types...")
	parser := parse.NewParser()
	resources, err := parser.ParseResourceTypes(schema)
	require.NoError(t, err, "Failed to parse resources")
	assert.NotEmpty(t, resources, "Should have parsed some resources")

	t.Logf("Parsed %d resource types", len(resources))

	// Find some well-known resources to verify
	var foundPod, foundDeployment, foundService bool
	for _, r := range resources {
		switch r.Kind {
		case "Pod":
			foundPod = true
			assert.Equal(t, "v1", r.Version)
			assert.Contains(t, []string{"", "core"}, r.Group) // core group can be empty string
			assert.NotEmpty(t, r.Properties)
		case "Deployment":
			foundDeployment = true
			assert.Equal(t, "apps", r.Group)
			assert.Equal(t, "v1", r.Version)
		case "Service":
			foundService = true
			assert.Equal(t, "v1", r.Version)
		}
	}

	assert.True(t, foundPod, "Should have found Pod resource")
	assert.True(t, foundDeployment, "Should have found Deployment resource")
	assert.True(t, foundService, "Should have found Service resource")

	// Step 3: Generate Go code
	t.Log("Generating Go code...")
	generator := generate.NewGenerator(resourcesDir)

	// Generate a subset of resources for testing (to keep test fast)
	testResources := []parse.ResourceType{}
	for _, r := range resources {
		if r.Kind == "Pod" || r.Kind == "Deployment" || r.Kind == "Service" {
			testResources = append(testResources, r)
		}
	}

	err = generator.GenerateResources(testResources)
	require.NoError(t, err, "Failed to generate resources")

	// Verify generated files
	podFile := filepath.Join(resourcesDir, "core", "v1", "pod.go")
	assert.FileExists(t, podFile, "Pod file should be generated")

	deploymentFile := filepath.Join(resourcesDir, "apps", "v1", "deployment.go")
	assert.FileExists(t, deploymentFile, "Deployment file should be generated")

	serviceFile := filepath.Join(resourcesDir, "core", "v1", "service.go")
	assert.FileExists(t, serviceFile, "Service file should be generated")

	// Verify content of generated Pod file
	content, err := os.ReadFile(podFile)
	require.NoError(t, err)
	contentStr := string(content)

	assert.Contains(t, contentStr, "package v1")
	assert.Contains(t, contentStr, "type Pod struct")
	assert.Contains(t, contentStr, "json:", "Should have JSON tags")
	assert.Contains(t, contentStr, "yaml:", "Should have YAML tags")

	t.Log("Code generation workflow completed successfully!")
}

// TestWorkflowWithCachedSchema tests that the workflow works with cached schemas.
func TestWorkflowWithCachedSchema(t *testing.T) {
	tmpDir := t.TempDir()
	schemasDir := filepath.Join(tmpDir, "schemas")

	// Create a minimal cached schema
	cachedPath := filepath.Join(schemasDir, "v1.28.0.json")
	testSchema := `{
		"swagger": "2.0",
		"info": {
			"title": "Kubernetes",
			"version": "v1.28.0"
		},
		"definitions": {
			"io.k8s.api.core.v1.ConfigMap": {
				"type": "object",
				"description": "ConfigMap holds configuration data",
				"properties": {
					"apiVersion": {
						"type": "string",
						"description": "APIVersion defines the versioned schema"
					},
					"kind": {
						"type": "string",
						"description": "Kind is a string value"
					},
					"data": {
						"type": "object",
						"additionalProperties": {
							"type": "string"
						}
					}
				},
				"x-kubernetes-group-version-kind": [
					{
						"group": "",
						"kind": "ConfigMap",
						"version": "v1"
					}
				]
			}
		}
	}`

	err := os.MkdirAll(schemasDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(cachedPath, []byte(testSchema), 0644)
	require.NoError(t, err)

	// Fetch schema (should use cache)
	fetcher := fetch.NewFetcher(schemasDir)
	ctx := context.Background()
	schema, err := fetcher.FetchSchema(ctx, "v1.28.0")
	require.NoError(t, err)

	// Parse resources
	parser := parse.NewParser()
	resources, err := parser.ParseResourceTypes(schema)
	require.NoError(t, err)
	require.Len(t, resources, 1)

	configMap := resources[0]
	assert.Equal(t, "ConfigMap", configMap.Kind)
	assert.Contains(t, configMap.Properties, "data")
	assert.Equal(t, "map[string]string", configMap.Properties["data"].GoType)

	// Generate code
	resourcesDir := filepath.Join(tmpDir, "resources")
	generator := generate.NewGenerator(resourcesDir)
	err = generator.GenerateResourceFile(configMap)
	require.NoError(t, err)

	// Verify generated file
	configMapFile := filepath.Join(resourcesDir, "core", "v1", "configmap.go")
	assert.FileExists(t, configMapFile)

	content, err := os.ReadFile(configMapFile)
	require.NoError(t, err)
	contentStr := string(content)

	assert.Contains(t, contentStr, "type ConfigMap struct")
	assert.Contains(t, contentStr, "Data map[string]string")
}
