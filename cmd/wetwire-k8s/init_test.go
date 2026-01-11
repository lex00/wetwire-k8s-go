package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCommand_Help(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "init", "--help"})
	assert.NoError(t, err)

	helpOutput := output.String()
	assert.Contains(t, helpOutput, "init")
	assert.Contains(t, helpOutput, "--example")
}

func TestInitCommand_DefaultDirectory(t *testing.T) {
	tempDir := t.TempDir()

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "init"})
	assert.NoError(t, err)

	// Check that k8s directory was created
	k8sDir := filepath.Join(tempDir, "k8s")
	info, err := os.Stat(k8sDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// Check that namespace.go was created
	nsFile := filepath.Join(k8sDir, "namespace.go")
	_, err = os.Stat(nsFile)
	assert.NoError(t, err)

	// Check that .wetwire.yaml was created
	configFile := filepath.Join(tempDir, ".wetwire.yaml")
	_, err = os.Stat(configFile)
	assert.NoError(t, err)
}

func TestInitCommand_SpecificDirectory(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, "myproject")

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "init", targetDir})
	assert.NoError(t, err)

	// Check that k8s directory was created
	k8sDir := filepath.Join(targetDir, "k8s")
	info, err := os.Stat(k8sDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// Check that namespace.go was created
	nsFile := filepath.Join(k8sDir, "namespace.go")
	_, err = os.Stat(nsFile)
	assert.NoError(t, err)
}

func TestInitCommand_WithExamples(t *testing.T) {
	tempDir := t.TempDir()

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "init", "--example", tempDir})
	assert.NoError(t, err)

	k8sDir := filepath.Join(tempDir, "k8s")

	// Check that namespace.go was created
	nsFile := filepath.Join(k8sDir, "namespace.go")
	nsContent, err := os.ReadFile(nsFile)
	require.NoError(t, err)
	assert.Contains(t, string(nsContent), "corev1.Namespace")

	// Check that deployment.go was created with examples
	deployFile := filepath.Join(k8sDir, "deployment.go")
	_, err = os.Stat(deployFile)
	assert.NoError(t, err)

	deployContent, err := os.ReadFile(deployFile)
	require.NoError(t, err)
	assert.Contains(t, string(deployContent), "appsv1.Deployment")

	// Check that service.go was created with examples
	svcFile := filepath.Join(k8sDir, "service.go")
	_, err = os.Stat(svcFile)
	assert.NoError(t, err)

	svcContent, err := os.ReadFile(svcFile)
	require.NoError(t, err)
	assert.Contains(t, string(svcContent), "corev1.Service")
}

func TestInitCommand_ExistingDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Pre-create the k8s directory with a file
	k8sDir := filepath.Join(tempDir, "k8s")
	err := os.MkdirAll(k8sDir, 0755)
	require.NoError(t, err)

	existingFile := filepath.Join(k8sDir, "existing.go")
	err = os.WriteFile(existingFile, []byte("package k8s\n"), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	// Init should succeed but not overwrite existing files
	err = app.Run([]string{"wetwire-k8s", "init", tempDir})
	assert.NoError(t, err)

	// Check existing file is preserved
	content, err := os.ReadFile(existingFile)
	require.NoError(t, err)
	assert.Equal(t, "package k8s\n", string(content))
}

func TestInitCommand_ConfigFileContent(t *testing.T) {
	tempDir := t.TempDir()

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "init", tempDir})
	assert.NoError(t, err)

	configFile := filepath.Join(tempDir, ".wetwire.yaml")
	content, err := os.ReadFile(configFile)
	require.NoError(t, err)

	configContent := string(content)
	assert.Contains(t, configContent, "source")
	assert.Contains(t, configContent, "k8s")
}

func TestInitCommand_NamespaceFileContent(t *testing.T) {
	tempDir := t.TempDir()

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "init", tempDir})
	assert.NoError(t, err)

	nsFile := filepath.Join(tempDir, "k8s", "namespace.go")
	content, err := os.ReadFile(nsFile)
	require.NoError(t, err)

	nsContent := string(content)
	assert.Contains(t, nsContent, "package k8s")
	assert.Contains(t, nsContent, "corev1")
	assert.Contains(t, nsContent, "Namespace")
}

func TestInitCommand_OutputMessage(t *testing.T) {
	tempDir := t.TempDir()

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "init", tempDir})
	assert.NoError(t, err)

	outputStr := output.String()
	assert.True(t, strings.Contains(outputStr, "Created") || strings.Contains(outputStr, "Initialized"),
		"Output should indicate successful initialization")
}

func TestInitCommand_NestedDirectory(t *testing.T) {
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "path", "to", "project")

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "init", nestedDir})
	assert.NoError(t, err)

	// Check that nested directories were created
	k8sDir := filepath.Join(nestedDir, "k8s")
	info, err := os.Stat(k8sDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}
