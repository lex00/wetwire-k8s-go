package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeRequest creates a CallToolRequest with the given arguments.
func makeRequest(args map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

func TestBuildHandler(t *testing.T) {
	t.Run("returns error for non-existent path", func(t *testing.T) {
		req := makeRequest(map[string]interface{}{
			"path": "/non/existent/path",
		})

		result, err := buildHandler(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "path does not exist")
	})

	t.Run("returns no resources for empty directory", func(t *testing.T) {
		// Create a temp directory
		tmpDir, err := os.MkdirTemp("", "wetwire-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		req := makeRequest(map[string]interface{}{
			"path": tmpDir,
		})

		result, err := buildHandler(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.Contains(t, getTextContent(result), "No Kubernetes resources found")
	})

	t.Run("uses default path when not specified", func(t *testing.T) {
		req := makeRequest(map[string]interface{}{})

		result, err := buildHandler(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, result)
		// Should not error about missing path
		assert.False(t, strings.Contains(getTextContent(result), "path does not exist"))
	})
}

func TestLintHandler(t *testing.T) {
	t.Run("returns error for non-existent path", func(t *testing.T) {
		req := makeRequest(map[string]interface{}{
			"path": "/non/existent/path",
		})

		result, err := lintHandler(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "path does not exist")
	})

	t.Run("lints empty directory successfully", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "wetwire-lint-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		req := makeRequest(map[string]interface{}{
			"path": tmpDir,
		})

		result, err := lintHandler(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.Contains(t, getTextContent(result), "No issues found")
	})

	t.Run("returns JSON format when requested", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "wetwire-lint-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		req := makeRequest(map[string]interface{}{
			"path":   tmpDir,
			"format": "json",
		})

		result, err := lintHandler(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)

		// Verify it's valid JSON
		var jsonResult map[string]interface{}
		err = json.Unmarshal([]byte(getTextContent(result)), &jsonResult)
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

		req := makeRequest(map[string]interface{}{
			"path": tmpDir,
		})

		result, err := lintHandler(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
	})
}

func TestImportHandler(t *testing.T) {
	t.Run("returns error when yaml is missing", func(t *testing.T) {
		req := makeRequest(map[string]interface{}{})

		result, err := importHandler(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "yaml parameter is required")
	})

	t.Run("imports simple deployment YAML", func(t *testing.T) {
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
		req := makeRequest(map[string]interface{}{
			"yaml": yamlContent,
		})

		result, err := importHandler(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)

		output := getTextContent(result)
		assert.Contains(t, output, "package main")
		assert.Contains(t, output, "Deployment")
		assert.Contains(t, output, "nginx")
	})

	t.Run("uses custom package name", func(t *testing.T) {
		yamlContent := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
`
		req := makeRequest(map[string]interface{}{
			"yaml":    yamlContent,
			"package": "k8s",
		})

		result, err := importHandler(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.Contains(t, getTextContent(result), "package k8s")
	})

	t.Run("uses variable prefix", func(t *testing.T) {
		yamlContent := `apiVersion: v1
kind: Service
metadata:
  name: my-service
`
		req := makeRequest(map[string]interface{}{
			"yaml":       yamlContent,
			"var_prefix": "App",
		})

		result, err := importHandler(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.Contains(t, getTextContent(result), "App")
	})
}

func TestValidateHandler(t *testing.T) {
	t.Run("returns error when both yaml and path are missing", func(t *testing.T) {
		req := makeRequest(map[string]interface{}{})

		result, err := validateHandler(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, getTextContent(result), "either 'yaml' or 'path' parameter is required")
	})

	t.Run("returns error for non-existent path", func(t *testing.T) {
		// Skip if kubeconform is not installed
		if _, err := os.Stat("/usr/local/bin/kubeconform"); os.IsNotExist(err) {
			t.Skip("kubeconform not installed")
		}

		req := makeRequest(map[string]interface{}{
			"path": "/non/existent/path.yaml",
		})

		result, err := validateHandler(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, result)
		// Either kubeconform not installed or path doesn't exist
		assert.True(t, result.IsError || strings.Contains(getTextContent(result), "not exist"))
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

// getTextContent extracts text content from a CallToolResult.
func getTextContent(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}

	// Try to extract text content from the result
	for _, content := range result.Content {
		if textContent, ok := mcp.AsTextContent(content); ok {
			return textContent.Text
		}
	}

	return ""
}
