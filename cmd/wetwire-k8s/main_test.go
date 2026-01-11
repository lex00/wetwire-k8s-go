package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportCommand_SingleDeployment(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "deployment.yaml")
	err := os.WriteFile(inputFile, []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
        - name: app
          image: nginx:latest
`), 0644)
	require.NoError(t, err)

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"import", inputFile}, &stdout, &stderr)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "package main")
	assert.Contains(t, stdout.String(), "var TestAppDeployment")
}

func TestImportCommand_OutputToFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "service.yaml")
	outputFile := filepath.Join(tmpDir, "output.go")

	err := os.WriteFile(inputFile, []byte(`
apiVersion: v1
kind: Service
metadata:
  name: my-svc
spec:
  selector:
    app: test
  ports:
    - port: 80
      targetPort: 8080
`), 0644)
	require.NoError(t, err)

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"import", "-o", outputFile, inputFile}, &stdout, &stderr)

	assert.Equal(t, 0, exitCode)

	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "package main")
	assert.Contains(t, string(content), "corev1.Service")
	assert.Contains(t, stderr.String(), "Imported 1 resources")
}

func TestImportCommand_CustomPackage(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "configmap.yaml")

	err := os.WriteFile(inputFile, []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  key: value
`), 0644)
	require.NoError(t, err)

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"import", "-p", "mypackage", inputFile}, &stdout, &stderr)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "package mypackage")
}

func TestImportCommand_VarPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "namespace.yaml")

	err := os.WriteFile(inputFile, []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: prod
`), 0644)
	require.NoError(t, err)

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"import", "--var-prefix", "Staging", inputFile}, &stdout, &stderr)

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "var StagingProdNamespace")
}

func TestImportCommand_MissingFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"import"}, &stdout, &stderr)
	assert.Equal(t, 2, exitCode)
}

func TestImportCommand_NonExistentFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"import", "nonexistent.yaml"}, &stdout, &stderr)
	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr.String(), "error reading file")
}

func TestImportCommand_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"import", "-h"}, &stdout, &stderr)
	assert.Equal(t, 2, exitCode) // flag.ErrHelp
	assert.Contains(t, stderr.String(), "Convert Kubernetes YAML manifests")
}

func TestHelpCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"help"}, &stdout, &stderr)
	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "wetwire-k8s")
	assert.Contains(t, stdout.String(), "import")
}

func TestVersionCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"version"}, &stdout, &stderr)
	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "wetwire-k8s version")
}

func TestUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"unknown"}, &stdout, &stderr)
	assert.Equal(t, 2, exitCode)
	assert.Contains(t, stderr.String(), "unknown command")
}
