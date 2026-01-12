package main

import (
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

	stdout, _, err := runTestCommand([]string{"import", inputFile})
	assert.NoError(t, err)

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

	_, stderr, err := runTestCommand([]string{"import", "-o", outputFile, inputFile})
	assert.NoError(t, err)

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

	stdout, _, err := runTestCommand([]string{"import", "-p", "mypackage", inputFile})
	assert.NoError(t, err)
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

	stdout, _, err := runTestCommand([]string{"import", "--var-prefix", "Staging", inputFile})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "var StagingProdNamespace")
}

func TestImportCommand_MissingFile(t *testing.T) {
	_, _, err := runTestCommand([]string{"import"})
	assert.Error(t, err)
}

func TestImportCommand_NonExistentFile(t *testing.T) {
	_, _, err := runTestCommand([]string{"import", "nonexistent.yaml"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestImportCommand_Help(t *testing.T) {
	stdout, _, err := runTestCommand([]string{"import", "--help"})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "Import reads YAML manifests")
}

func TestHelpCommand(t *testing.T) {
	stdout, _, err := runTestCommand([]string{"--help"})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "wetwire-k8s")
	assert.Contains(t, stdout.String(), "import")
	assert.Contains(t, stdout.String(), "build")
}

func TestVersionCommand(t *testing.T) {
	stdout, _, err := runTestCommand([]string{"--version"})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "wetwire-k8s")
}
