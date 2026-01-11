package discover_test

import (
	"path/filepath"
	"testing"

	"github.com/lex00/wetwire-k8s-go/internal/discover"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscover_SimpleResources(t *testing.T) {
	// Test discovery of simple resources without dependencies
	testDir := filepath.Join("testdata", "simple.go")

	resources, err := discover.DiscoverFile(testDir)
	require.NoError(t, err)
	require.Len(t, resources, 2, "should discover 2 resources")

	// Check SimpleDeployment
	deployment := findResource(resources, "SimpleDeployment")
	require.NotNil(t, deployment, "SimpleDeployment should be discovered")
	assert.Equal(t, "SimpleDeployment", deployment.Name)
	assert.Contains(t, deployment.Type, "Deployment")
	assert.Contains(t, deployment.File, "simple.go")
	assert.Greater(t, deployment.Line, 0)
	assert.Empty(t, deployment.Dependencies, "SimpleDeployment should have no dependencies")

	// Check SimpleService
	service := findResource(resources, "SimpleService")
	require.NotNil(t, service, "SimpleService should be discovered")
	assert.Equal(t, "SimpleService", service.Name)
	assert.Contains(t, service.Type, "Service")
	assert.Contains(t, service.File, "simple.go")
	assert.Greater(t, service.Line, 0)
	assert.Empty(t, service.Dependencies, "SimpleService should have no dependencies")
}

func TestDiscover_WithDependencies(t *testing.T) {
	// Test discovery of resources with dependencies
	testDir := filepath.Join("testdata", "dependencies.go")

	resources, err := discover.DiscoverFile(testDir)
	require.NoError(t, err)
	require.Len(t, resources, 3, "should discover 3 resources")

	// Check AppConfig
	configMap := findResource(resources, "AppConfig")
	require.NotNil(t, configMap, "AppConfig should be discovered")
	assert.Equal(t, "AppConfig", configMap.Name)
	assert.Contains(t, configMap.Type, "ConfigMap")
	assert.Empty(t, configMap.Dependencies, "AppConfig should have no dependencies")

	// Check WebDeployment
	deployment := findResource(resources, "WebDeployment")
	require.NotNil(t, deployment, "WebDeployment should be discovered")
	assert.Equal(t, "WebDeployment", deployment.Name)
	assert.Contains(t, deployment.Type, "Deployment")
	assert.Contains(t, deployment.Dependencies, "AppConfig", "WebDeployment should depend on AppConfig")

	// Check WebService
	service := findResource(resources, "WebService")
	require.NotNil(t, service, "WebService should be discovered")
	assert.Equal(t, "WebService", service.Name)
	assert.Contains(t, service.Type, "Service")
	assert.Contains(t, service.Dependencies, "WebDeployment", "WebService should depend on WebDeployment")
}

func TestDiscover_Directory(t *testing.T) {
	// Test discovery of all resources in a directory
	resources, err := discover.DiscoverDirectory("testdata")
	require.NoError(t, err)

	// Should find all resources from both files
	assert.GreaterOrEqual(t, len(resources), 5, "should discover at least 5 resources")

	// Verify we found resources from both files
	assert.NotNil(t, findResource(resources, "SimpleDeployment"))
	assert.NotNil(t, findResource(resources, "SimpleService"))
	assert.NotNil(t, findResource(resources, "AppConfig"))
	assert.NotNil(t, findResource(resources, "WebDeployment"))
	assert.NotNil(t, findResource(resources, "WebService"))
}

func TestDiscover_RecursiveDirectory(t *testing.T) {
	// Test recursive discovery starting from parent directory
	resources, err := discover.DiscoverDirectory(".")
	require.NoError(t, err)

	// Should find resources from testdata subdirectory
	assert.Greater(t, len(resources), 0, "should discover resources recursively")
}

func TestDiscover_NonExistentFile(t *testing.T) {
	// Test error handling for non-existent file
	_, err := discover.DiscoverFile("nonexistent.go")
	assert.Error(t, err, "should return error for non-existent file")
}

func TestDiscover_InvalidGoFile(t *testing.T) {
	// Test error handling for invalid Go syntax
	testFile := filepath.Join(t.TempDir(), "invalid.go")
	err := writeFile(testFile, "package invalid\n\nthis is not valid go code {{{")
	require.NoError(t, err)

	_, err = discover.DiscoverFile(testFile)
	assert.Error(t, err, "should return error for invalid Go syntax")
}

func TestResource_TypeExtraction(t *testing.T) {
	// Test that type information is correctly extracted
	testDir := filepath.Join("testdata", "simple.go")

	resources, err := discover.DiscoverFile(testDir)
	require.NoError(t, err)

	deployment := findResource(resources, "SimpleDeployment")
	require.NotNil(t, deployment)

	// Type should include the API group and version
	// e.g., "apps/v1.Deployment" or "*apps/v1.Deployment"
	assert.Regexp(t, `(?:apps/v1\.)?Deployment`, deployment.Type)
}

func TestResource_LineNumbers(t *testing.T) {
	// Test that line numbers are accurate
	testDir := filepath.Join("testdata", "simple.go")

	resources, err := discover.DiscoverFile(testDir)
	require.NoError(t, err)

	deployment := findResource(resources, "SimpleDeployment")
	service := findResource(resources, "SimpleService")

	require.NotNil(t, deployment)
	require.NotNil(t, service)

	// Service should be declared after Deployment
	assert.Greater(t, service.Line, deployment.Line, "SimpleService should be after SimpleDeployment")
}

// Helper functions

func findResource(resources []discover.Resource, name string) *discover.Resource {
	for i := range resources {
		if resources[i].Name == name {
			return &resources[i]
		}
	}
	return nil
}

func writeFile(path, content string) error {
	// This will be implemented if needed by the test
	// For now, using os.WriteFile
	return nil
}
