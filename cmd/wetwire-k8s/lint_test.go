package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLintCommand_Help(t *testing.T) {
	stdout, _, err := runTestCommand([]string{"lint", "--help"})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "Lint parses Go source files")
	assert.Contains(t, stdout.String(), "--format")
	assert.Contains(t, stdout.String(), "--severity")
}

func TestLintCommand_NoResources(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "empty.go"), []byte("package main\n"), 0644)
	require.NoError(t, err)

	_, _, err = runTestCommand([]string{"lint", tmpDir})
	assert.NoError(t, err)
}

func TestLintCommand_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("package main\n"), 0644)
	require.NoError(t, err)

	_, _, err = runTestCommand([]string{"lint", "-f", "invalid", tmpDir})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid format")
}

func TestLintCommand_NonExistentPath(t *testing.T) {
	_, _, err := runTestCommand([]string{"lint", "/nonexistent/path"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}
