package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCommand_Help(t *testing.T) {
	stdout, _, err := runTestCommand([]string{"init", "--help"})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "creates a new wetwire-k8s project")
	assert.Contains(t, stdout.String(), "--example")
}

func TestInitCommand_Default(t *testing.T) {
	tmpDir := t.TempDir()

	stdout, _, err := runTestCommand([]string{"init", tmpDir})
	assert.NoError(t, err)

	// Check that k8s directory was created
	k8sDir := filepath.Join(tmpDir, "k8s")
	_, err = os.Stat(k8sDir)
	require.NoError(t, err)

	// Check that namespace.go was created
	nsFile := filepath.Join(k8sDir, "namespace.go")
	_, err = os.Stat(nsFile)
	require.NoError(t, err)

	// Check that .wetwire.yaml was created
	configFile := filepath.Join(tmpDir, ".wetwire.yaml")
	_, err = os.Stat(configFile)
	require.NoError(t, err)

	assert.Contains(t, stdout.String(), "Created")
	assert.Contains(t, stdout.String(), "Initialized")
}

func TestInitCommand_WithExample(t *testing.T) {
	tmpDir := t.TempDir()

	stdout, _, err := runTestCommand([]string{"init", "--example", tmpDir})
	assert.NoError(t, err)

	// Check that example files were created
	k8sDir := filepath.Join(tmpDir, "k8s")

	deployFile := filepath.Join(k8sDir, "deployment.go")
	_, err = os.Stat(deployFile)
	require.NoError(t, err)

	svcFile := filepath.Join(k8sDir, "service.go")
	_, err = os.Stat(svcFile)
	require.NoError(t, err)

	assert.Contains(t, stdout.String(), "deployment.go")
	assert.Contains(t, stdout.String(), "service.go")
}
