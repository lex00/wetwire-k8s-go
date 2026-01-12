package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCommand_Help(t *testing.T) {
	stdout, _, err := runTestCommand([]string{"list", "--help"})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "displays discovered Kubernetes resources")
	assert.Contains(t, stdout.String(), "--format")
	assert.Contains(t, stdout.String(), "--all")
}

func TestListCommand_NoResources(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "empty.go"), []byte("package main\n"), 0644)
	require.NoError(t, err)

	stdout, _, err := runTestCommand([]string{"list", tmpDir})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "No resources found")
}

func TestListCommand_WithResource(t *testing.T) {
	tmpDir := t.TempDir()

	content := `package main

import corev1 "k8s.io/api/core/v1"

var MyService = &corev1.Service{}
`
	err := os.WriteFile(filepath.Join(tmpDir, "service.go"), []byte(content), 0644)
	require.NoError(t, err)

	stdout, _, err := runTestCommand([]string{"list", tmpDir})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "MyService")
}

func TestListCommand_JSONFormat(t *testing.T) {
	tmpDir := t.TempDir()

	content := `package main

import corev1 "k8s.io/api/core/v1"

var MyPod = &corev1.Pod{}
`
	err := os.WriteFile(filepath.Join(tmpDir, "pod.go"), []byte(content), 0644)
	require.NoError(t, err)

	stdout, _, err := runTestCommand([]string{"list", "-f", "json", tmpDir})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), `"name"`)
}

func TestListCommand_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("package main\n"), 0644)
	require.NoError(t, err)

	_, _, err = runTestCommand([]string{"list", "-f", "invalid", tmpDir})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid format")
}

func TestListCommand_NonExistentPath(t *testing.T) {
	_, _, err := runTestCommand([]string{"list", "/nonexistent/path"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}
