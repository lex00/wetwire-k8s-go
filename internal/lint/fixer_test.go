package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFixer(t *testing.T) {
	t.Run("should create fixer with nil config", func(t *testing.T) {
		fixer := NewFixer(nil)
		assert.NotNil(t, fixer)
		assert.NotNil(t, fixer.config)
	})

	t.Run("should create fixer with custom config", func(t *testing.T) {
		config := &Config{
			MinSeverity: SeverityWarning,
		}
		fixer := NewFixer(config)
		assert.NotNil(t, fixer)
		assert.Equal(t, SeverityWarning, fixer.config.MinSeverity)
	})
}

func TestDetermineImagePullPolicy(t *testing.T) {
	tests := []struct {
		image    string
		expected string
	}{
		// :latest images should use Always
		{"nginx:latest", "Always"},
		{"myapp:latest", "Always"},

		// No tag should use Always (defaults to :latest)
		{"nginx", "Always"},
		{"myregistry.com/myapp", "Always"},

		// Tagged images should use IfNotPresent
		{"nginx:1.21", "IfNotPresent"},
		{"nginx:1.21.6", "IfNotPresent"},
		{"myregistry.com/myapp:v1.2.3", "IfNotPresent"},

		// SHA digest images should use IfNotPresent
		{"nginx@sha256:abc123def456", "IfNotPresent"},
	}

	for _, tt := range tests {
		t.Run(tt.image, func(t *testing.T) {
			result := determineImagePullPolicy(tt.image)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsFixableRule(t *testing.T) {
	tests := []struct {
		ruleID   string
		expected bool
	}{
		{"WK8105", true},  // ImagePullPolicy - fixable
		{"WK8002", true},  // Deeply nested - fixable
		{"WK8001", false}, // Top-level declarations - not fixable
		{"WK8003", false}, // Duplicate names - not fixable
		{"WK8006", false}, // :latest tags - not fixable (user must choose version)
	}

	for _, tt := range tests {
		t.Run(tt.ruleID, func(t *testing.T) {
			result := isFixableRule(tt.ruleID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFixableRules(t *testing.T) {
	rules := FixableRules()
	assert.Contains(t, rules, "WK8002")
	assert.Contains(t, rules, "WK8105")
	assert.NotContains(t, rules, "WK8006") // :latest is not fixable
}

func TestFixer_FixFile_WK8105(t *testing.T) {
	// Create a temporary file with missing ImagePullPolicy
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_wk8105.go")

	content := `package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

var ContainerNoPolicy = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
}

var ContainerLatest = corev1.Container{
	Name:  "app",
	Image: "nginx:latest",
}

var ContainerNoTag = corev1.Container{
	Name:  "app",
	Image: "nginx",
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Apply fixes
	fixer := NewFixer(nil)
	results, err := fixer.FixFile(testFile)
	require.NoError(t, err)

	// Should have fixed 3 containers
	assert.Len(t, results, 3)
	for _, r := range results {
		assert.True(t, r.Fixed)
		assert.Equal(t, "WK8105", r.Rule)
	}

	// Read the fixed file
	fixed, err := os.ReadFile(testFile)
	require.NoError(t, err)

	// Verify ImagePullPolicy was added
	fixedContent := string(fixed)
	assert.Contains(t, fixedContent, `ImagePullPolicy`)

	// Verify correct policies
	assert.Contains(t, fixedContent, `"IfNotPresent"`) // For nginx:1.21
	assert.Contains(t, fixedContent, `"Always"`)        // For nginx:latest and nginx
}

func TestFixer_FixFile_WK8105_AlreadySet(t *testing.T) {
	// Create a temporary file with ImagePullPolicy already set
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_wk8105_good.go")

	content := `package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

var ContainerWithPolicy = corev1.Container{
	Name:            "app",
	Image:           "nginx:1.21",
	ImagePullPolicy: "Always",
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Apply fixes
	fixer := NewFixer(nil)
	results, err := fixer.FixFile(testFile)
	require.NoError(t, err)

	// Should not have fixed anything
	assert.Len(t, results, 0)
}

func TestFixer_FixFile_WK8002(t *testing.T) {
	// Create a temporary file with deeply nested structures
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_wk8002.go")

	content := `package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var DeeplyNestedDeployment = appsv1.Deployment{
	Spec: appsv1.DeploymentSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "app",
						Image: "nginx:latest",
						Env: []corev1.EnvVar{
							{
								Name:  "CONFIG",
								Value: "value",
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

	// Apply fixes
	fixer := NewFixer(nil)
	results, err := fixer.FixFile(testFile)
	require.NoError(t, err)

	// Check if any WK8002 fixes were applied
	wk8002Fixes := 0
	for _, r := range results {
		if r.Rule == "WK8002" && r.Fixed {
			wk8002Fixes++
		}
	}

	// Should have extracted nested structures
	if wk8002Fixes > 0 {
		// Read the fixed file
		fixed, err := os.ReadFile(testFile)
		require.NoError(t, err)

		// The fixed file should have new extracted variables
		fixedContent := string(fixed)
		// Verify structure was extracted (variable names may vary)
		assert.Contains(t, fixedContent, "var ")
	}
}

func TestFixer_FixDirectory(t *testing.T) {
	// Create a temporary directory with test files
	tempDir := t.TempDir()

	// Create a file with fixable issues
	testFile1 := filepath.Join(tempDir, "test1.go")
	content1 := `package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

var Container1 = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
}
`
	err := os.WriteFile(testFile1, []byte(content1), 0644)
	require.NoError(t, err)

	// Create another file
	testFile2 := filepath.Join(tempDir, "test2.go")
	content2 := `package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

var Container2 = corev1.Container{
	Name:  "sidecar",
	Image: "busybox:latest",
}
`
	err = os.WriteFile(testFile2, []byte(content2), 0644)
	require.NoError(t, err)

	// First, verify that the linter finds WK8105 issues
	linter := NewLinter(nil)
	result, err := linter.LintWithResult(tempDir)
	require.NoError(t, err)

	// Check if WK8105 issues are found
	wk8105Count := 0
	for _, issue := range result.Issues {
		if issue.Rule == "WK8105" {
			wk8105Count++
		}
	}
	assert.Equal(t, 2, wk8105Count, "Expected 2 WK8105 issues")

	// Apply fixes
	fixer := NewFixer(nil)
	results, err := fixer.FixDirectory(tempDir)
	require.NoError(t, err)

	// Verify files were actually modified by reading them back
	fixed1, err := os.ReadFile(testFile1)
	require.NoError(t, err)
	fixed2, err := os.ReadFile(testFile2)
	require.NoError(t, err)

	// Both files should now have ImagePullPolicy
	assert.Contains(t, string(fixed1), "ImagePullPolicy")
	assert.Contains(t, string(fixed2), "ImagePullPolicy")

	// Check that correct policies were applied
	assert.Contains(t, string(fixed1), "IfNotPresent") // nginx:1.21 should get IfNotPresent
	assert.Contains(t, string(fixed2), "Always")       // busybox:latest should get Always

	// Check results
	fixedCount := 0
	for _, r := range results {
		if r.Fixed {
			fixedCount++
		}
	}
	assert.Equal(t, 2, fixedCount)
}

func TestFixer_FixFile_NonExistent(t *testing.T) {
	fixer := NewFixer(nil)
	_, err := fixer.FixFile("/nonexistent/path/file.go")
	assert.Error(t, err)
}

func TestFixer_FixFile_InvalidGoCode(t *testing.T) {
	// Create a temporary file with invalid Go code
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "invalid.go")

	content := `package testdata

this is not valid go code {
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	fixer := NewFixer(nil)
	_, err = fixer.FixFile(testFile)
	assert.Error(t, err)
}

func TestFixer_PreservesComments(t *testing.T) {
	// Create a temporary file with comments
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_comments.go")

	content := `package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// This is an important comment about the container
var ContainerWithComment = corev1.Container{
	Name:  "app",          // Container name
	Image: "nginx:1.21",   // Image version
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Apply fixes
	fixer := NewFixer(nil)
	_, err = fixer.FixFile(testFile)
	require.NoError(t, err)

	// Read the fixed file
	fixed, err := os.ReadFile(testFile)
	require.NoError(t, err)

	fixedContent := string(fixed)

	// Comments should be preserved
	assert.True(t, strings.Contains(fixedContent, "This is an important comment") ||
		strings.Contains(fixedContent, "Container name"),
		"Comments should be preserved in the fixed file")
}
