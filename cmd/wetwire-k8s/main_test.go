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

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "import", inputFile})
	assert.NoError(t, err)

	assert.Contains(t, output.String(), "package main")
	assert.Contains(t, output.String(), "var TestAppDeployment")
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

	app := newApp()
	errOutput := &bytes.Buffer{}
	app.ErrWriter = errOutput

	err = app.Run([]string{"wetwire-k8s", "import", "-o", outputFile, inputFile})
	assert.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "package main")
	assert.Contains(t, string(content), "corev1.Service")
	assert.Contains(t, errOutput.String(), "Imported 1 resources")
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

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "import", "-p", "mypackage", inputFile})
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "package mypackage")
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

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "import", "--var-prefix", "Staging", inputFile})
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "var StagingProdNamespace")
}

func TestImportCommand_MissingFile(t *testing.T) {
	app := newApp()
	err := app.Run([]string{"wetwire-k8s", "import"})
	assert.Error(t, err)
}

func TestImportCommand_NonExistentFile(t *testing.T) {
	app := newApp()
	err := app.Run([]string{"wetwire-k8s", "import", "nonexistent.yaml"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestImportCommand_Help(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "import", "--help"})
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Convert Kubernetes YAML manifests")
}

func TestHelpCommand(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "--help"})
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "wetwire-k8s")
	assert.Contains(t, output.String(), "import")
	assert.Contains(t, output.String(), "build")
}

func TestVersionCommand(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "--version"})
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "wetwire-k8s")
}
