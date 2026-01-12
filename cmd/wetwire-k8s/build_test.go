package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCommand_Help(t *testing.T) {
	stdout, _, err := runTestCommand([]string{"build", "--help"})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "Build parses Go source files")
	assert.Contains(t, stdout.String(), "--output")
	assert.Contains(t, stdout.String(), "--format")
}

func TestBuildCommand_NoResources(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "empty.go"), []byte("package main\n"), 0644)
	require.NoError(t, err)

	stdout, _, err := runTestCommand([]string{"build", tmpDir})
	assert.NoError(t, err)
	assert.Empty(t, stdout.String())
}

func TestBuildCommand_WithResource(t *testing.T) {
	tmpDir := t.TempDir()

	content := `package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyConfigMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-config",
	},
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "resources.go"), []byte(content), 0644)
	require.NoError(t, err)

	stdout, _, err := runTestCommand([]string{"build", tmpDir})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "apiVersion")
}

func TestBuildCommand_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("package main\n"), 0644)
	require.NoError(t, err)

	_, _, err = runTestCommand([]string{"build", "-f", "invalid", tmpDir})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid format")
}

func TestBuildCommand_NonExistentPath(t *testing.T) {
	_, _, err := runTestCommand([]string{"build", "/nonexistent/path"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}
