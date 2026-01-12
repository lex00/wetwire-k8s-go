package main

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper to check if kubeconform is installed
func kubeconformInstalled() bool {
	_, err := exec.LookPath("kubeconform")
	return err == nil
}

func TestValidateCommand_Help(t *testing.T) {
	stdout, _, err := runTestCommand([]string{"validate", "--help"})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "Validate parses Kubernetes YAML manifests")
	assert.Contains(t, stdout.String(), "--strict")
	assert.Contains(t, stdout.String(), "kubeconform")
}

func TestValidateCommand_MissingKubeconform(t *testing.T) {
	if kubeconformInstalled() {
		t.Skip("kubeconform is installed, skipping missing kubeconform test")
	}

	_, _, err := runTestCommand([]string{"validate", "manifest.yaml"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kubeconform")
}

func TestValidateCommand_MissingInput(t *testing.T) {
	_, _, err := runTestCommand([]string{"validate"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing input")
}
