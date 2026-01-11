package importer_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lex00/wetwire-k8s-go/internal/importer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportFile_SingleDeployment(t *testing.T) {
	testFile := filepath.Join("testdata", "deployment.yaml")
	result, err := importer.ImportFile(testFile, importer.DefaultOptions())
	require.NoError(t, err)
	assert.Equal(t, 1, result.ResourceCount)
	assert.Contains(t, result.GoCode, "package main")
	assert.Contains(t, result.GoCode, `"k8s.io/api/apps/v1"`)
	assert.Contains(t, result.GoCode, "var NginxDeployment")
	assert.Contains(t, result.GoCode, `Name: "nginx-deployment"`)
}

func TestImportFile_MultiDocument(t *testing.T) {
	testFile := filepath.Join("testdata", "multi-document.yaml")
	result, err := importer.ImportFile(testFile, importer.DefaultOptions())
	require.NoError(t, err)
	assert.Equal(t, 4, result.ResourceCount)
	assert.Contains(t, result.GoCode, "var MyAppNamespace")
	assert.Contains(t, result.GoCode, "var AppConfigConfigMap")
	assert.Contains(t, result.GoCode, "var WebAppDeployment")
	assert.Contains(t, result.GoCode, "var WebServiceService")
}

func TestImportFile_Service(t *testing.T) {
	testFile := filepath.Join("testdata", "service.yaml")
	result, err := importer.ImportFile(testFile, importer.DefaultOptions())
	require.NoError(t, err)
	assert.Equal(t, 1, result.ResourceCount)
	assert.Contains(t, result.GoCode, "corev1.Service")
	assert.Contains(t, result.GoCode, "ServiceTypeLoadBalancer")
}

func TestImportFile_ConfigMap(t *testing.T) {
	testFile := filepath.Join("testdata", "configmap.yaml")
	result, err := importer.ImportFile(testFile, importer.DefaultOptions())
	require.NoError(t, err)
	assert.Equal(t, 1, result.ResourceCount)
	assert.Contains(t, result.GoCode, "corev1.ConfigMap")
	assert.Contains(t, result.GoCode, "Data:")
}

func TestImportFile_EmptyFile(t *testing.T) {
	testFile := filepath.Join("testdata", "empty.yaml")
	result, err := importer.ImportFile(testFile, importer.DefaultOptions())
	require.NoError(t, err)
	assert.Equal(t, 0, result.ResourceCount)
	assert.Contains(t, result.GoCode, "package main")
}

func TestImportFile_CustomPackageName(t *testing.T) {
	testFile := filepath.Join("testdata", "deployment.yaml")
	opts := importer.Options{PackageName: "myapp"}
	result, err := importer.ImportFile(testFile, opts)
	require.NoError(t, err)
	assert.Contains(t, result.GoCode, "package myapp")
}

func TestImportFile_VarPrefix(t *testing.T) {
	testFile := filepath.Join("testdata", "deployment.yaml")
	opts := importer.Options{PackageName: "main", VarPrefix: "Prod"}
	result, err := importer.ImportFile(testFile, opts)
	require.NoError(t, err)
	assert.Contains(t, result.GoCode, "var ProdNginxDeployment")
}

func TestImportFile_NonExistentFile(t *testing.T) {
	_, err := importer.ImportFile("nonexistent.yaml", importer.DefaultOptions())
	assert.Error(t, err)
}

func TestImportFile_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(tmpFile, []byte("{{invalid yaml content"), 0644)
	require.NoError(t, err)
	_, err = importer.ImportFile(tmpFile, importer.DefaultOptions())
	assert.Error(t, err)
}

func TestImportBytes_SimpleDeployment(t *testing.T) {
	yamlContent := []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
        - name: test
          image: test:latest
`)
	result, err := importer.ImportBytes(yamlContent, importer.DefaultOptions())
	require.NoError(t, err)
	assert.Equal(t, 1, result.ResourceCount)
	assert.Contains(t, result.GoCode, "var TestDeployDeployment")
}

func TestParseYAML_MultiDocument(t *testing.T) {
	yamlContent := []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: config1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config2
`)
	resources, err := importer.ParseYAML(yamlContent)
	require.NoError(t, err)
	assert.Len(t, resources, 2)
	assert.Equal(t, "config1", resources[0].Name)
	assert.Equal(t, "config2", resources[1].Name)
}

func TestGenerateVarName(t *testing.T) {
	tests := []struct {
		name, kind, prefix, expected string
	}{
		{"nginx-deployment", "Deployment", "", "NginxDeploymentDeployment"},
		{"my-service", "Service", "", "MyServiceService"},
		{"app-config", "ConfigMap", "Prod", "ProdAppConfigConfigMap"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := importer.GenerateVarName(tt.name, tt.kind, tt.prefix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateGoCode_ValidSyntax(t *testing.T) {
	testFile := filepath.Join("testdata", "deployment.yaml")
	result, err := importer.ImportFile(testFile, importer.DefaultOptions())
	require.NoError(t, err)
	openBraces := strings.Count(result.GoCode, "{")
	closeBraces := strings.Count(result.GoCode, "}")
	assert.Equal(t, openBraces, closeBraces, "Braces should be balanced")
}

func TestImportFile_PointerFields(t *testing.T) {
	testFile := filepath.Join("testdata", "deployment.yaml")
	result, err := importer.ImportFile(testFile, importer.DefaultOptions())
	require.NoError(t, err)
	assert.Contains(t, result.GoCode, "ptr.To[int32](3)")
}

func TestAPIVersionToImport(t *testing.T) {
	tests := []struct {
		apiVersion, expectedPath, expectedAlias string
	}{
		{"v1", "k8s.io/api/core/v1", "corev1"},
		{"apps/v1", "k8s.io/api/apps/v1", "appsv1"},
		{"batch/v1", "k8s.io/api/batch/v1", "batchv1"},
	}
	for _, tt := range tests {
		t.Run(tt.apiVersion, func(t *testing.T) {
			importPath, alias := importer.APIVersionToImport(tt.apiVersion)
			assert.Equal(t, tt.expectedPath, importPath)
			assert.Equal(t, tt.expectedAlias, alias)
		})
	}
}
