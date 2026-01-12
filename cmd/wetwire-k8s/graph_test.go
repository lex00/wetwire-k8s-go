package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphCommand_Help(t *testing.T) {
	stdout, _, err := runTestCommand([]string{"graph", "--help"})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "dependency relationships")
	assert.Contains(t, stdout.String(), "--format")
}

func TestGraphCommand_NoResources(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "empty.go"), []byte("package main\n"), 0644)
	require.NoError(t, err)

	stdout, _, err := runTestCommand([]string{"graph", tmpDir})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "No resources found")
}

func TestGraphCommand_WithResource(t *testing.T) {
	tmpDir := t.TempDir()

	content := `package main

import corev1 "k8s.io/api/core/v1"

var MyService = &corev1.Service{}
`
	err := os.WriteFile(filepath.Join(tmpDir, "service.go"), []byte(content), 0644)
	require.NoError(t, err)

	stdout, _, err := runTestCommand([]string{"graph", tmpDir})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "MyService")
}

func TestGraphCommand_DOTFormat(t *testing.T) {
	tmpDir := t.TempDir()

	content := `package main

import corev1 "k8s.io/api/core/v1"

var MyPod = &corev1.Pod{}
`
	err := os.WriteFile(filepath.Join(tmpDir, "pod.go"), []byte(content), 0644)
	require.NoError(t, err)

	stdout, _, err := runTestCommand([]string{"graph", "-f", "dot", tmpDir})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "digraph")
}

func TestGraphCommand_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("package main\n"), 0644)
	require.NoError(t, err)

	_, _, err = runTestCommand([]string{"graph", "-f", "invalid", tmpDir})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid format")
}

func TestGraphCommand_NonExistentPath(t *testing.T) {
	_, _, err := runTestCommand([]string{"graph", "/nonexistent/path"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}
