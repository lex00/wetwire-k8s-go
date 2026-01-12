package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildToolHandler(t *testing.T) {
	t.Run("returns error for non-existent path", func(t *testing.T) {
		args := map[string]any{
			"package": "/non/existent/path",
		}

		result, err := buildToolHandler(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "path does not exist")
		assert.Empty(t, result)
	})

	t.Run("returns no resources for empty directory", func(t *testing.T) {
		// Create a temp directory
		tmpDir, err := os.MkdirTemp("", "wetwire-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		args := map[string]any{
			"package": tmpDir,
		}

		result, err := buildToolHandler(context.Background(), args)

		require.NoError(t, err)
		assert.Contains(t, result, "No Kubernetes resources found")
	})

	t.Run("uses default path when not specified", func(t *testing.T) {
		args := map[string]any{}

		result, err := buildToolHandler(context.Background(), args)

		require.NoError(t, err)
		// Should not error about missing path
		assert.NotContains(t, result, "path does not exist")
	})
}

func TestLintToolHandler(t *testing.T) {
	t.Run("returns error for non-existent path", func(t *testing.T) {
		args := map[string]any{
			"package": "/non/existent/path",
		}

		result, err := lintToolHandler(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "path does not exist")
		assert.Empty(t, result)
	})

	t.Run("lints empty directory successfully", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "wetwire-lint-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		args := map[string]any{
			"package": tmpDir,
		}

		result, err := lintToolHandler(context.Background(), args)

		require.NoError(t, err)
		assert.Contains(t, result, "No issues found")
	})

	t.Run("returns JSON format when requested", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "wetwire-lint-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		args := map[string]any{
			"package": tmpDir,
			"format":  "json",
		}

		result, err := lintToolHandler(context.Background(), args)

		require.NoError(t, err)

		// Verify it's valid JSON
		var jsonResult map[string]interface{}
		err = json.Unmarshal([]byte(result), &jsonResult)
		require.NoError(t, err)
		assert.Contains(t, jsonResult, "total_files")
		assert.Contains(t, jsonResult, "issues")
	})

	t.Run("lints Go file with issues", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "wetwire-lint-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create a Go file that might have lint issues
		goFile := filepath.Join(tmpDir, "test.go")
		content := `package test

import (
	appsv1 "k8s.io/api/apps/v1"
)

var MyDeployment = appsv1.Deployment{}
`
		err = os.WriteFile(goFile, []byte(content), 0644)
		require.NoError(t, err)

		args := map[string]any{
			"package": tmpDir,
		}

		result, err := lintToolHandler(context.Background(), args)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})
}

func TestImportToolHandler(t *testing.T) {
	t.Run("returns error when files/yaml is missing", func(t *testing.T) {
		args := map[string]any{}

		result, err := importToolHandler(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "files parameter is required")
		assert.Empty(t, result)
	})

	t.Run("imports simple deployment YAML via yaml param", func(t *testing.T) {
		yamlContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
`
		args := map[string]any{
			"yaml": yamlContent,
		}

		result, err := importToolHandler(context.Background(), args)

		require.NoError(t, err)
		assert.Contains(t, result, "package main")
		assert.Contains(t, result, "Deployment")
		assert.Contains(t, result, "nginx")
	})

	t.Run("imports from file via files param", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "wetwire-import-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		yamlContent := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
`
		yamlFile := filepath.Join(tmpDir, "config.yaml")
		err = os.WriteFile(yamlFile, []byte(yamlContent), 0644)
		require.NoError(t, err)

		args := map[string]any{
			"files":   []interface{}{yamlFile},
			"package": "k8s",
		}

		result, err := importToolHandler(context.Background(), args)

		require.NoError(t, err)
		assert.Contains(t, result, "package k8s")
	})

	t.Run("uses variable prefix", func(t *testing.T) {
		yamlContent := `apiVersion: v1
kind: Service
metadata:
  name: my-service
`
		args := map[string]any{
			"yaml":       yamlContent,
			"var_prefix": "App",
		}

		result, err := importToolHandler(context.Background(), args)

		require.NoError(t, err)
		assert.Contains(t, result, "App")
	})
}

func TestValidateToolHandler(t *testing.T) {
	t.Run("returns error when path is missing", func(t *testing.T) {
		args := map[string]any{}

		result, err := validateToolHandler(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "path parameter is required")
		assert.Empty(t, result)
	})

	t.Run("returns error for non-existent path", func(t *testing.T) {
		// Skip if kubeconform is not installed
		if _, lookupErr := exec.LookPath("kubeconform"); lookupErr != nil {
			t.Skip("kubeconform not installed")
		}

		args := map[string]any{
			"path": "/non/existent/path.yaml",
		}

		result, err := validateToolHandler(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "path does not exist")
		assert.Empty(t, result)
	})
}

func TestFormatLintText(t *testing.T) {
	t.Run("formats no issues", func(t *testing.T) {
		result := &lintResult{
			TotalFiles: 5,
			Issues:     nil,
		}

		output := formatLintTextFromResult(result)

		assert.Contains(t, output, "No issues found")
		assert.Contains(t, output, "5 files")
	})

	t.Run("formats issues", func(t *testing.T) {
		result := &lintResult{
			TotalFiles:      3,
			FilesWithIssues: 1,
			ErrorCount:      1,
			WarningCount:    1,
			InfoCount:       0,
			Issues: []lintIssue{
				{
					File:     "test.go",
					Line:     10,
					Column:   5,
					Severity: "error",
					Rule:     "WK8001",
					Message:  "test error",
				},
				{
					File:     "test.go",
					Line:     20,
					Column:   1,
					Severity: "warning",
					Rule:     "WK8002",
					Message:  "test warning",
				},
			},
		}

		output := formatLintTextFromResult(result)

		assert.Contains(t, output, "test.go:10:5")
		assert.Contains(t, output, "WK8001")
		assert.Contains(t, output, "1 errors")
		assert.Contains(t, output, "1 warnings")
	})
}

func TestFormatLintGitHub(t *testing.T) {
	t.Run("formats issues for GitHub Actions", func(t *testing.T) {
		result := &lintResult{
			Issues: []lintIssue{
				{
					File:     "test.go",
					Line:     10,
					Column:   5,
					Severity: "error",
					Rule:     "WK8001",
					Message:  "test error",
				},
			},
		}

		output := formatLintGitHubFromResult(result)

		assert.Contains(t, output, "::error")
		assert.Contains(t, output, "file=test.go")
		assert.Contains(t, output, "line=10")
		assert.Contains(t, output, "title=WK8001")
	})
}

// Helper types and functions for test formatting

type lintIssue struct {
	File     string
	Line     int
	Column   int
	Severity string
	Rule     string
	Message  string
}

type lintResult struct {
	TotalFiles      int
	FilesWithIssues int
	ErrorCount      int
	WarningCount    int
	InfoCount       int
	Issues          []lintIssue
}

func formatLintTextFromResult(result *lintResult) string {
	if len(result.Issues) == 0 {
		return fmt.Sprintf("No issues found. Scanned %d files.", result.TotalFiles)
	}

	var output strings.Builder
	for _, issue := range result.Issues {
		output.WriteString(fmt.Sprintf("%s:%d:%d: %s [%s] %s\n",
			issue.File, issue.Line, issue.Column, issue.Severity, issue.Rule, issue.Message))
	}

	output.WriteString(fmt.Sprintf("\nSummary: %d errors, %d warnings, %d info messages in %d files\n",
		result.ErrorCount, result.WarningCount, result.InfoCount, result.FilesWithIssues))

	return output.String()
}

func formatLintGitHubFromResult(result *lintResult) string {
	if len(result.Issues) == 0 {
		return ""
	}

	var output strings.Builder
	for _, issue := range result.Issues {
		output.WriteString(fmt.Sprintf("::%s file=%s,line=%d,col=%d,title=%s::%s\n",
			issue.Severity, issue.File, issue.Line, issue.Column, issue.Rule, issue.Message))
	}

	return output.String()
}
