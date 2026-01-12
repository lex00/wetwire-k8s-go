package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffCommand_Help(t *testing.T) {
	stdout, _, err := runTestCommand([]string{"diff", "--help"})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "compares the generated Kubernetes manifests")
	assert.Contains(t, stdout.String(), "--against")
	assert.Contains(t, stdout.String(), "--semantic")
}

func TestDiffCommand_MissingAgainst(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("package main\n"), 0644)
	require.NoError(t, err)

	_, _, err = runTestCommand([]string{"diff", tmpDir})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "against")
}

func TestDiffCommand_NonExistentPath(t *testing.T) {
	_, _, err := runTestCommand([]string{"diff", "--against", "manifest.yaml", "/nonexistent/path"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestDiffCommand_NonExistentAgainst(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("package main\n"), 0644)
	require.NoError(t, err)

	_, _, err = runTestCommand([]string{"diff", "--against", "nonexistent.yaml", tmpDir})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read")
}
