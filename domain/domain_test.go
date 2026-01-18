package domain

import (
	"os"
	"path/filepath"
	"testing"

	coredomain "github.com/lex00/wetwire-core-go/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time interface checks
var (
	_ coredomain.Domain        = (*K8sDomain)(nil)
	_ coredomain.ListerDomain  = (*K8sDomain)(nil)
	_ coredomain.GrapherDomain = (*K8sDomain)(nil)
)

func TestK8sLinter_Lint_WithFixOption(t *testing.T) {
	// Create a temporary directory with a file that has fixable issues
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_fix.go")

	// This file has a WK8105 issue (missing ImagePullPolicy)
	content := `package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

var ContainerMissingPolicy = corev1.Container{
	Name:  "app",
	Image: "nginx:1.21",
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	domain := &K8sDomain{}
	linter := domain.Linter()
	ctx := &Context{}

	t.Run("should accept Fix option without error", func(t *testing.T) {
		opts := LintOpts{
			Fix: true,
		}
		result, err := linter.Lint(ctx, tempDir, opts)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("should apply fixes when Fix is true", func(t *testing.T) {
		// Write fresh content again
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		opts := LintOpts{
			Fix: true,
		}
		_, err = linter.Lint(ctx, tempDir, opts)
		require.NoError(t, err)

		// Read the fixed file
		fixed, err := os.ReadFile(testFile)
		require.NoError(t, err)

		// Verify ImagePullPolicy was added by the fixer
		fixedContent := string(fixed)
		assert.Contains(t, fixedContent, "ImagePullPolicy", "Fix option should have triggered auto-fix")
	})

	t.Run("should not modify file when Fix is false", func(t *testing.T) {
		// Write fresh content again
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		opts := LintOpts{
			Fix: false,
		}
		_, err = linter.Lint(ctx, tempDir, opts)
		require.NoError(t, err)

		// Read the file
		unchanged, err := os.ReadFile(testFile)
		require.NoError(t, err)

		// Verify file was not modified
		assert.Equal(t, content, string(unchanged), "File should not be modified when Fix is false")
	})
}

func TestK8sLinter_Lint_FixWithNoFixableIssues(t *testing.T) {
	// Create a temporary directory with a file that has only unfixable issues
	// (WK8105/WK8002 are fixable, but this file doesn't trigger those)
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_good.go")

	// This file has ImagePullPolicy set (so WK8105 won't fire)
	// but may have other unfixable security warnings
	content := `package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

var ContainerWithPolicy = corev1.Container{
	Name:            "app",
	Image:           "nginx:1.21",
	ImagePullPolicy: "IfNotPresent",
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	domain := &K8sDomain{}
	linter := domain.Linter()
	ctx := &Context{}

	opts := LintOpts{
		Fix: true,
	}
	result, err := linter.Lint(ctx, tempDir, opts)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// File should not be modified (no fixable issues)
	// But result may contain unfixable security warnings
	unchanged, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, content, string(unchanged), "File should not be modified when there are no fixable issues")
}

func TestK8sLinter_Lint_FixWithUnfixableIssues(t *testing.T) {
	// Create a temporary directory with a file that has unfixable issues
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_unfixable.go")

	// This file has WK8006 issue (:latest tag) which is not auto-fixable
	content := `package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

var ContainerLatest = corev1.Container{
	Name:            "app",
	Image:           "nginx:latest",
	ImagePullPolicy: "Always",
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	domain := &K8sDomain{}
	linter := domain.Linter()
	ctx := &Context{}

	opts := LintOpts{
		Fix: true,
	}
	result, err := linter.Lint(ctx, tempDir, opts)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should still report the unfixable issue
	assert.Contains(t, result.Message, "lint issues found")
}
