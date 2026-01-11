package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestListCommand_Help(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "list", "--help"})
	assert.NoError(t, err)

	helpOutput := output.String()
	assert.Contains(t, helpOutput, "list")
	assert.Contains(t, helpOutput, "--format")
	assert.Contains(t, helpOutput, "--all")
}

func TestListCommand_NoArgs(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "resources.go")
	content := `package main

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-app",
	},
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "list"})
	assert.NoError(t, err)

	listOutput := output.String()
	assert.Contains(t, listOutput, "MyDeployment")
	assert.Contains(t, listOutput, "appsv1.Deployment")
}

func TestListCommand_WithPath(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "app.go")
	content := `package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyConfigMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "my-config",
		Namespace: "default",
	},
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "list", tempDir})
	assert.NoError(t, err)

	listOutput := output.String()
	assert.Contains(t, listOutput, "MyConfigMap")
	assert.Contains(t, listOutput, "corev1.ConfigMap")
}

func TestListCommand_FormatTable(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "resources.go")
	content := `package main

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var TestDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{Name: "test-deployment"},
}

var TestService = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{Name: "test-service"},
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "list", "--format", "table", tempDir})
	assert.NoError(t, err)

	tableOutput := output.String()
	// Table should have headers
	assert.Contains(t, tableOutput, "NAME")
	assert.Contains(t, tableOutput, "TYPE")
	assert.Contains(t, tableOutput, "FILE")
	// And data
	assert.Contains(t, tableOutput, "TestDeployment")
	assert.Contains(t, tableOutput, "TestService")
}

func TestListCommand_FormatJSON(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "resources.go")
	content := `package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyNamespace = &corev1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-namespace",
	},
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "list", "--format", "json", tempDir})
	assert.NoError(t, err)

	jsonOutput := output.String()
	var result []map[string]interface{}
	err = json.Unmarshal([]byte(jsonOutput), &result)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "MyNamespace", result[0]["name"])
	assert.Equal(t, "corev1.Namespace", result[0]["type"])
}

func TestListCommand_FormatYAML(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "resources.go")
	content := `package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MySecret = &corev1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-secret",
	},
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "list", "--format", "yaml", tempDir})
	assert.NoError(t, err)

	yamlOutput := output.String()
	var result []map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlOutput), &result)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "MySecret", result[0]["name"])
}

func TestListCommand_WithDependencies(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "deps.go")
	content := `package main

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var AppConfig = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "app-config",
	},
}

var WebApp = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "web-app",
	},
	Spec: appsv1.DeploymentSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "web",
						EnvFrom: []corev1.EnvFromSource{
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: AppConfig.Name,
									},
								},
							},
						},
					},
				},
			},
		},
	},
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "list", "--all", tempDir})
	assert.NoError(t, err)

	listOutput := output.String()
	assert.Contains(t, listOutput, "AppConfig")
	assert.Contains(t, listOutput, "WebApp")
	// With --all flag, should show dependencies
	assert.Contains(t, listOutput, "DEPENDENCIES")
}

func TestListCommand_InvalidPath(t *testing.T) {
	app := newApp()
	err := app.Run([]string{"wetwire-k8s", "list", "/nonexistent/path"})
	assert.Error(t, err)
}

func TestListCommand_NoResources(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "empty.go")
	content := `package main

func main() {
	// No Kubernetes resources here
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "list", tempDir})
	assert.NoError(t, err)

	listOutput := output.String()
	assert.Contains(t, strings.ToLower(listOutput), "no resources")
}

func TestListCommand_MultipleFiles(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "namespace.go")
	file2 := filepath.Join(tempDir, "deployment.go")

	content1 := `package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var AppNamespace = &corev1.Namespace{
	ObjectMeta: metav1.ObjectMeta{Name: "app-ns"},
}
`
	content2 := `package main

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var AppDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{Name: "app"},
}
`
	err := os.WriteFile(file1, []byte(content1), 0644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte(content2), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "list", tempDir})
	assert.NoError(t, err)

	listOutput := output.String()
	assert.Contains(t, listOutput, "AppNamespace")
	assert.Contains(t, listOutput, "AppDeployment")
}
