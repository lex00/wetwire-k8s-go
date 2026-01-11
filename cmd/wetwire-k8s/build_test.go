package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCommand_Help(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "build", "--help"})
	assert.NoError(t, err)

	helpOutput := output.String()
	assert.Contains(t, helpOutput, "build")
	assert.Contains(t, helpOutput, "--output")
	assert.Contains(t, helpOutput, "--format")
	assert.Contains(t, helpOutput, "--dry-run")
}

func TestBuildCommand_NoArgs(t *testing.T) {
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

	err = app.Run([]string{"wetwire-k8s", "build"})
	assert.NoError(t, err)

	yamlOutput := output.String()
	assert.Contains(t, yamlOutput, "apiVersion")
	assert.Contains(t, yamlOutput, "kind")
}

func TestBuildCommand_WithPath(t *testing.T) {
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

	err = app.Run([]string{"wetwire-k8s", "build", tempDir})
	assert.NoError(t, err)

	yamlOutput := output.String()
	assert.Contains(t, yamlOutput, "ConfigMap")
	assert.Contains(t, yamlOutput, "my-config-map")
}

func TestBuildCommand_OutputFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "resources.go")
	outputFile := filepath.Join(tempDir, "output.yaml")

	content := `package main

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var TestDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-app",
	},
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	app := newApp()
	err = app.Run([]string{"wetwire-k8s", "build", "--output", outputFile, tempDir})
	assert.NoError(t, err)

	_, err = os.Stat(outputFile)
	assert.NoError(t, err, "output file should exist")

	data, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "Deployment")
	assert.Contains(t, string(data), "test-deployment")
}

func TestBuildCommand_FormatJSON(t *testing.T) {
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

	err = app.Run([]string{"wetwire-k8s", "build", "--format", "json", tempDir})
	assert.NoError(t, err)

	jsonOutput := output.String()
	assert.Contains(t, jsonOutput, "{")
	assert.Contains(t, jsonOutput, "\"kind\"")
	assert.Contains(t, jsonOutput, "\"Namespace\"")
}

func TestBuildCommand_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "resources.go")
	outputFile := filepath.Join(tempDir, "output.yaml")

	content := `package main

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DryRunDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "dry-run-app",
	},
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "build", "--dry-run", "--output", outputFile, tempDir})
	assert.NoError(t, err)

	_, err = os.Stat(outputFile)
	assert.True(t, os.IsNotExist(err), "output file should not exist in dry-run mode")

	dryRunOutput := output.String()
	assert.Contains(t, dryRunOutput, "dry-run-deployment")
}

func TestBuildCommand_MultipleResources(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "resources.go")

	content := `package main

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MultiDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "multi-app",
	},
}

var MultiService = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name: "multi-svc",
	},
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "build", tempDir})
	assert.NoError(t, err)

	yamlOutput := output.String()
	assert.Contains(t, yamlOutput, "---")
	assert.Contains(t, yamlOutput, "multi-deployment")
	assert.Contains(t, yamlOutput, "multi-service")
}

func TestBuildCommand_InvalidPath(t *testing.T) {
	app := newApp()
	err := app.Run([]string{"wetwire-k8s", "build", "/nonexistent/path"})
	assert.Error(t, err)
}

func TestBuildCommand_InvalidFormat(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "resources.go")

	content := `package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var SimpleResource = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "simple",
	},
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	app := newApp()
	err = app.Run([]string{"wetwire-k8s", "build", "--format", "xml", tempDir})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "format")
}

func TestBuildCommand_WithDependencies(t *testing.T) {
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

	err = app.Run([]string{"wetwire-k8s", "build", tempDir})
	assert.NoError(t, err)

	yamlOutput := output.String()
	configIdx := strings.Index(yamlOutput, "ConfigMap")
	deployIdx := strings.Index(yamlOutput, "Deployment")

	assert.Greater(t, configIdx, -1, "ConfigMap should be in output")
	assert.Greater(t, deployIdx, -1, "Deployment should be in output")
	assert.Less(t, configIdx, deployIdx, "ConfigMap should appear before Deployment")
}

func TestBuildCommand_APIVersionMapping(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "resources.go")

	content := `package main

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var TestDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{Name: "test-deployment"},
}

var TestService = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{Name: "test-service"},
}

var TestJob = &batchv1.Job{
	ObjectMeta: metav1.ObjectMeta{Name: "test-job"},
}

var TestIngress = &networkingv1.Ingress{
	ObjectMeta: metav1.ObjectMeta{Name: "test-ingress"},
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "build", tempDir})
	assert.NoError(t, err)

	yamlOutput := output.String()
	assert.Contains(t, yamlOutput, "apps/v1")
	assert.Contains(t, yamlOutput, "batch/v1")
	assert.Contains(t, yamlOutput, "networking.k8s.io/v1")
	assert.Contains(t, yamlOutput, "apiVersion: v1")
}
