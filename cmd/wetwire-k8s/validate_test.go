package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to check if kubeconform is installed
func kubeconformInstalled() bool {
	_, err := exec.LookPath("kubeconform")
	return err == nil
}

func TestValidateCommand_Help(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "validate", "--help"})
	assert.NoError(t, err)

	helpOutput := output.String()
	assert.Contains(t, helpOutput, "validate")
	assert.Contains(t, helpOutput, "--schema-location")
	assert.Contains(t, helpOutput, "--strict")
	assert.Contains(t, helpOutput, "--output")
}

func TestValidateCommand_ValidYAML(t *testing.T) {
	if !kubeconformInstalled() {
		t.Skip("kubeconform not installed")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "valid.yaml")
	err := os.WriteFile(inputFile, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key: value
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	errOutput := &bytes.Buffer{}
	app.Writer = output
	app.ErrWriter = errOutput

	err = app.Run([]string{"wetwire-k8s", "validate", inputFile})
	assert.NoError(t, err)
}

func TestValidateCommand_InvalidYAML(t *testing.T) {
	if !kubeconformInstalled() {
		t.Skip("kubeconform not installed")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "invalid.yaml")
	// Invalid: Deployment missing required selector and template fields
	err := os.WriteFile(inputFile, []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 1
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	errOutput := &bytes.Buffer{}
	app.Writer = output
	app.ErrWriter = errOutput

	err = app.Run([]string{"wetwire-k8s", "validate", inputFile})
	// Should return error for invalid manifest
	assert.Error(t, err)
}

func TestValidateCommand_MultipleFiles(t *testing.T) {
	if !kubeconformInstalled() {
		t.Skip("kubeconform not installed")
	}

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "configmap.yaml")
	file2 := filepath.Join(tmpDir, "service.yaml")

	err := os.WriteFile(file1, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value
`), 0644)
	require.NoError(t, err)

	err = os.WriteFile(file2, []byte(`apiVersion: v1
kind: Service
metadata:
  name: test-service
spec:
  selector:
    app: test
  ports:
    - port: 80
      targetPort: 8080
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "validate", file1, file2})
	assert.NoError(t, err)
}

func TestValidateCommand_JSONOutput(t *testing.T) {
	if !kubeconformInstalled() {
		t.Skip("kubeconform not installed")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "valid.yaml")
	err := os.WriteFile(inputFile, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "validate", "--output", "json", inputFile})
	assert.NoError(t, err)
	// JSON output should contain JSON structure
	assert.Contains(t, output.String(), "{")
}

func TestValidateCommand_StdinInput(t *testing.T) {
	if !kubeconformInstalled() {
		t.Skip("kubeconform not installed")
	}

	// Create a pipe to simulate stdin
	yamlContent := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value
`
	// Save original stdin
	oldStdin := os.Stdin

	// Create a temp file to simulate stdin
	tmpFile, err := os.CreateTemp("", "stdin")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	_, err = tmpFile.Seek(0, 0)
	require.NoError(t, err)

	os.Stdin = tmpFile
	defer func() { os.Stdin = oldStdin }()

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "validate", "-"})
	assert.NoError(t, err)
}

func TestValidateCommand_NoArgs(t *testing.T) {
	app := newApp()
	err := app.Run([]string{"wetwire-k8s", "validate"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file")
}

func TestValidateCommand_NonExistentFile(t *testing.T) {
	app := newApp()
	err := app.Run([]string{"wetwire-k8s", "validate", "nonexistent.yaml"})
	assert.Error(t, err)
}

func TestValidateCommand_KubeconformNotInstalled(t *testing.T) {
	// Save PATH
	oldPath := os.Getenv("PATH")
	// Set empty PATH so kubeconform cannot be found
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", oldPath)

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.yaml")
	err := os.WriteFile(inputFile, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
`), 0644)
	require.NoError(t, err)

	app := newApp()
	errOutput := &bytes.Buffer{}
	app.ErrWriter = errOutput

	err = app.Run([]string{"wetwire-k8s", "validate", inputFile})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kubeconform")
}

func TestValidateCommand_StrictMode(t *testing.T) {
	if !kubeconformInstalled() {
		t.Skip("kubeconform not installed")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "strict.yaml")
	// This manifest has an unknown field which strict mode should catch
	err := os.WriteFile(inputFile, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  unknownField: invalid
data:
  key: value
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "validate", "--strict", inputFile})
	// Strict mode should fail on unknown fields
	assert.Error(t, err)
}

func TestValidateCommand_SchemaLocation(t *testing.T) {
	if !kubeconformInstalled() {
		t.Skip("kubeconform not installed")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "valid.yaml")
	err := os.WriteFile(inputFile, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	// Test with default schema location (should work)
	err = app.Run([]string{"wetwire-k8s", "validate", "--schema-location", "default", inputFile})
	assert.NoError(t, err)
}

func TestValidateCommand_MultiDocumentYAML(t *testing.T) {
	if !kubeconformInstalled() {
		t.Skip("kubeconform not installed")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "multi.yaml")
	err := os.WriteFile(inputFile, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: config-1
data:
  key: value
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-2
data:
  key: value2
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "validate", inputFile})
	assert.NoError(t, err)
}

func TestValidateCommand_BuildIntegration(t *testing.T) {
	if !kubeconformInstalled() {
		t.Skip("kubeconform not installed")
	}

	// Create a Go source file that generates valid K8s manifests
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "resources.go")
	err := os.WriteFile(goFile, []byte(`package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyConfigMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-config",
	},
}
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	// Test validating build output using --from-build flag
	err = app.Run([]string{"wetwire-k8s", "validate", "--from-build", tmpDir})
	assert.NoError(t, err)
}

func TestValidateCommand_Directory(t *testing.T) {
	if !kubeconformInstalled() {
		t.Skip("kubeconform not installed")
	}

	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "configmap.yaml"), []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value
`), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "service.yaml"), []byte(`apiVersion: v1
kind: Service
metadata:
  name: test-service
spec:
  selector:
    app: test
  ports:
    - port: 80
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	// Validate all files in directory
	err = app.Run([]string{"wetwire-k8s", "validate", tmpDir})
	assert.NoError(t, err)
}
