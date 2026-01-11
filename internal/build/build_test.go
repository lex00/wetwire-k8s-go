package build_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lex00/wetwire-k8s-go/internal/build"
	"github.com/lex00/wetwire-k8s-go/internal/discover"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuild_SimpleResources(t *testing.T) {
	// Test building simple resources without dependencies
	testDir := filepath.Join("..", "discover", "testdata", "simple.go")

	result, err := build.Build(testDir, build.Options{
		OutputMode: build.SingleFile,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should discover 2 resources
	assert.Len(t, result.Resources, 2)

	// Resources should be in stable order (no dependencies)
	assert.NotEmpty(t, result.OrderedResources)
	assert.Len(t, result.OrderedResources, 2)
}

func TestBuild_WithDependencies(t *testing.T) {
	// Test building resources with dependencies
	testDir := filepath.Join("..", "discover", "testdata", "dependencies.go")

	result, err := build.Build(testDir, build.Options{
		OutputMode: build.SingleFile,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should discover 3 resources
	assert.Len(t, result.Resources, 3)

	// Resources should be ordered by dependencies
	assert.Len(t, result.OrderedResources, 3)

	// AppConfig should come before WebDeployment
	// WebDeployment should come before WebService
	configIdx := findResourceIndex(result.OrderedResources, "AppConfig")
	deployIdx := findResourceIndex(result.OrderedResources, "WebDeployment")
	svcIdx := findResourceIndex(result.OrderedResources, "WebService")

	require.NotEqual(t, -1, configIdx, "AppConfig should be in ordered resources")
	require.NotEqual(t, -1, deployIdx, "WebDeployment should be in ordered resources")
	require.NotEqual(t, -1, svcIdx, "WebService should be in ordered resources")

	assert.Less(t, configIdx, deployIdx, "AppConfig should come before WebDeployment")
	assert.Less(t, deployIdx, svcIdx, "WebDeployment should come before WebService")
}

func TestBuild_Directory(t *testing.T) {
	// Test building all resources in a directory
	testDir := filepath.Join("..", "discover", "testdata")

	result, err := build.Build(testDir, build.Options{
		OutputMode: build.SingleFile,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should discover all resources from both files
	assert.GreaterOrEqual(t, len(result.Resources), 5)
	assert.GreaterOrEqual(t, len(result.OrderedResources), 5)
}

func TestBuild_SingleFileOutput(t *testing.T) {
	// Test single file output mode
	testDir := filepath.Join("..", "discover", "testdata")
	outputDir := t.TempDir()
	outputFile := filepath.Join(outputDir, "output.yaml")

	result, err := build.Build(testDir, build.Options{
		OutputMode: build.SingleFile,
		OutputPath: outputFile,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Output file should be set
	assert.Equal(t, outputFile, result.OutputPath)
}

func TestBuild_SeparateFilesOutput(t *testing.T) {
	// Test separate files output mode
	testDir := filepath.Join("..", "discover", "testdata")
	outputDir := t.TempDir()

	result, err := build.Build(testDir, build.Options{
		OutputMode: build.SeparateFiles,
		OutputPath: outputDir,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Output paths should contain multiple files
	assert.Greater(t, len(result.OutputPaths), 0)

	// Each resource should have its own output path
	assert.Len(t, result.OutputPaths, len(result.OrderedResources))
}

func TestValidateReferences_ValidReferences(t *testing.T) {
	// Test validation of valid resource references
	resources := []discover.Resource{
		{Name: "ConfigMap1", Type: "corev1.ConfigMap", Dependencies: []string{}},
		{Name: "Deployment1", Type: "appsv1.Deployment", Dependencies: []string{"ConfigMap1"}},
		{Name: "Service1", Type: "corev1.Service", Dependencies: []string{"Deployment1"}},
	}

	err := build.ValidateReferences(resources)
	assert.NoError(t, err, "valid references should pass validation")
}

func TestValidateReferences_InvalidReferences(t *testing.T) {
	// Test validation of invalid resource references
	resources := []discover.Resource{
		{Name: "Deployment1", Type: "appsv1.Deployment", Dependencies: []string{"NonExistent"}},
	}

	err := build.ValidateReferences(resources)
	assert.Error(t, err, "invalid reference should fail validation")
	assert.Contains(t, err.Error(), "NonExistent", "error should mention the missing resource")
}

func TestValidateReferences_SelfReference(t *testing.T) {
	// Test validation of self-referencing resource
	resources := []discover.Resource{
		{Name: "Resource1", Type: "corev1.ConfigMap", Dependencies: []string{"Resource1"}},
	}

	err := build.ValidateReferences(resources)
	assert.Error(t, err, "self-reference should fail validation")
	assert.Contains(t, err.Error(), "Resource1", "error should mention the self-referencing resource")
}

func TestDetectCycles_NoCycles(t *testing.T) {
	// Test detection of cycles with no cycles present
	resources := []discover.Resource{
		{Name: "A", Type: "corev1.ConfigMap", Dependencies: []string{}},
		{Name: "B", Type: "appsv1.Deployment", Dependencies: []string{"A"}},
		{Name: "C", Type: "corev1.Service", Dependencies: []string{"B"}},
	}

	err := build.DetectCycles(resources)
	assert.NoError(t, err, "no cycles should pass detection")
}

func TestDetectCycles_SimpleCycle(t *testing.T) {
	// Test detection of simple two-resource cycle
	resources := []discover.Resource{
		{Name: "A", Type: "corev1.ConfigMap", Dependencies: []string{"B"}},
		{Name: "B", Type: "appsv1.Deployment", Dependencies: []string{"A"}},
	}

	err := build.DetectCycles(resources)
	assert.Error(t, err, "simple cycle should be detected")
	assert.Contains(t, err.Error(), "cycle", "error should mention cycle")
}

func TestDetectCycles_ComplexCycle(t *testing.T) {
	// Test detection of complex multi-resource cycle
	resources := []discover.Resource{
		{Name: "A", Type: "corev1.ConfigMap", Dependencies: []string{"B"}},
		{Name: "B", Type: "appsv1.Deployment", Dependencies: []string{"C"}},
		{Name: "C", Type: "corev1.Service", Dependencies: []string{"D"}},
		{Name: "D", Type: "corev1.Secret", Dependencies: []string{"A"}},
	}

	err := build.DetectCycles(resources)
	assert.Error(t, err, "complex cycle should be detected")
	assert.Contains(t, err.Error(), "cycle", "error should mention cycle")
}

func TestDetectCycles_MultipleCycles(t *testing.T) {
	// Test detection of multiple independent cycles
	resources := []discover.Resource{
		{Name: "A", Type: "corev1.ConfigMap", Dependencies: []string{"B"}},
		{Name: "B", Type: "appsv1.Deployment", Dependencies: []string{"A"}},
		{Name: "C", Type: "corev1.Service", Dependencies: []string{"D"}},
		{Name: "D", Type: "corev1.Secret", Dependencies: []string{"C"}},
	}

	err := build.DetectCycles(resources)
	assert.Error(t, err, "multiple cycles should be detected")
}

func TestTopologicalSort_SimpleChain(t *testing.T) {
	// Test topological sort with simple linear dependency chain
	resources := []discover.Resource{
		{Name: "C", Type: "corev1.Service", Dependencies: []string{"B"}},
		{Name: "B", Type: "appsv1.Deployment", Dependencies: []string{"A"}},
		{Name: "A", Type: "corev1.ConfigMap", Dependencies: []string{}},
	}

	sorted, err := build.TopologicalSort(resources)
	require.NoError(t, err)
	require.Len(t, sorted, 3)

	// Should be sorted in dependency order: A -> B -> C
	assert.Equal(t, "A", sorted[0].Name)
	assert.Equal(t, "B", sorted[1].Name)
	assert.Equal(t, "C", sorted[2].Name)
}

func TestTopologicalSort_DAG(t *testing.T) {
	// Test topological sort with diamond-shaped DAG
	//     A
	//    / \
	//   B   C
	//    \ /
	//     D
	resources := []discover.Resource{
		{Name: "D", Type: "corev1.Service", Dependencies: []string{"B", "C"}},
		{Name: "C", Type: "corev1.Secret", Dependencies: []string{"A"}},
		{Name: "B", Type: "appsv1.Deployment", Dependencies: []string{"A"}},
		{Name: "A", Type: "corev1.ConfigMap", Dependencies: []string{}},
	}

	sorted, err := build.TopologicalSort(resources)
	require.NoError(t, err)
	require.Len(t, sorted, 4)

	// A should come first
	assert.Equal(t, "A", sorted[0].Name)

	// D should come last
	assert.Equal(t, "D", sorted[3].Name)

	// B and C can be in any order, but both should come after A and before D
	bIdx := findResourceIndex(sorted, "B")
	cIdx := findResourceIndex(sorted, "C")
	assert.Equal(t, 1, min(bIdx, cIdx), "B or C should be at index 1")
	assert.Equal(t, 2, max(bIdx, cIdx), "B or C should be at index 2")
}

func TestTopologicalSort_NoDependencies(t *testing.T) {
	// Test topological sort with resources that have no dependencies
	resources := []discover.Resource{
		{Name: "A", Type: "corev1.ConfigMap", Dependencies: []string{}},
		{Name: "B", Type: "appsv1.Deployment", Dependencies: []string{}},
		{Name: "C", Type: "corev1.Service", Dependencies: []string{}},
	}

	sorted, err := build.TopologicalSort(resources)
	require.NoError(t, err)
	require.Len(t, sorted, 3)

	// All resources should be present (order doesn't matter)
	names := []string{sorted[0].Name, sorted[1].Name, sorted[2].Name}
	assert.Contains(t, names, "A")
	assert.Contains(t, names, "B")
	assert.Contains(t, names, "C")
}

func TestTopologicalSort_WithCycle(t *testing.T) {
	// Test that topological sort fails when there's a cycle
	resources := []discover.Resource{
		{Name: "A", Type: "corev1.ConfigMap", Dependencies: []string{"B"}},
		{Name: "B", Type: "appsv1.Deployment", Dependencies: []string{"A"}},
	}

	sorted, err := build.TopologicalSort(resources)
	assert.Error(t, err, "topological sort should fail with cycle")
	assert.Nil(t, sorted, "should return nil when there's a cycle")
	assert.Contains(t, err.Error(), "cycle", "error should mention cycle")
}

func TestTopologicalSort_EmptyInput(t *testing.T) {
	// Test topological sort with empty input
	resources := []discover.Resource{}

	sorted, err := build.TopologicalSort(resources)
	require.NoError(t, err)
	assert.Empty(t, sorted, "empty input should produce empty output")
}

func TestBuild_InvalidReferences(t *testing.T) {
	// Test that build fails with invalid references
	// Create a temporary file that references a non-existent resource
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "invalid.go")
	content := `package testdata

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NonExistentResource is referenced but not a K8s resource
var NonExistentResource = "not-a-resource"

// This resource references NonExistentResource
var Invalid = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: NonExistentResource,
	},
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Build should fail due to invalid reference
	_, err = build.Build(tempDir, build.Options{
		OutputMode: build.SingleFile,
	})
	assert.Error(t, err, "build should fail with invalid references")
}

func TestBuild_CircularDependencies(t *testing.T) {
	// Test that build fails with circular dependencies
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "circular.go")
	content := `package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var A = &corev1.ConfigMap{
	Data: map[string]string{
		"ref": B.Name,
	},
}

var B = &appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						EnvFrom: []corev1.EnvFromSource{
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: A.Name,
									},
								},
							},
						},
					},
				},
			},
		},
	},
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Build should fail due to circular dependency
	_, err = build.Build(tempDir, build.Options{
		OutputMode: build.SingleFile,
	})
	assert.Error(t, err, "build should fail with circular dependencies")
	assert.Contains(t, err.Error(), "cycle", "error should mention cycle")
}

func TestBuild_OutputDirectory(t *testing.T) {
	// Test that output directory is created if it doesn't exist
	testDir := filepath.Join("..", "discover", "testdata", "simple.go")
	outputDir := filepath.Join(t.TempDir(), "nested", "output")

	_, err := build.Build(testDir, build.Options{
		OutputMode: build.SeparateFiles,
		OutputPath: outputDir,
	})
	require.NoError(t, err)

	// Output directory should be created
	info, err := os.Stat(outputDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir(), "output path should be a directory")
}

// Helper functions

func findResourceIndex(resources []discover.Resource, name string) int {
	for i, r := range resources {
		if r.Name == name {
			return i
		}
	}
	return -1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
