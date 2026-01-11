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

func TestGraphCommand_Help(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "graph", "--help"})
	assert.NoError(t, err)

	helpOutput := output.String()
	assert.Contains(t, helpOutput, "graph")
	assert.Contains(t, helpOutput, "--format")
	assert.Contains(t, helpOutput, "--output")
}

func TestGraphCommand_NoArgs(t *testing.T) {
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

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "graph"})
	assert.NoError(t, err)

	graphOutput := output.String()
	assert.Contains(t, graphOutput, "AppConfig")
	assert.Contains(t, graphOutput, "WebApp")
}

func TestGraphCommand_WithPath(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "deps.go")
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

	err = app.Run([]string{"wetwire-k8s", "graph", tempDir})
	assert.NoError(t, err)

	graphOutput := output.String()
	assert.Contains(t, graphOutput, "MyNamespace")
}

func TestGraphCommand_FormatASCII(t *testing.T) {
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

	err = app.Run([]string{"wetwire-k8s", "graph", "--format", "ascii", tempDir})
	assert.NoError(t, err)

	graphOutput := output.String()
	// ASCII format should show dependency arrows
	assert.Contains(t, graphOutput, "AppConfig")
	assert.Contains(t, graphOutput, "WebApp")
	// Should have a header and tree structure
	assert.Contains(t, graphOutput, "Resource Dependency Graph")
	assert.Contains(t, graphOutput, "===")
}

func TestGraphCommand_FormatDOT(t *testing.T) {
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

	err = app.Run([]string{"wetwire-k8s", "graph", "--format", "dot", tempDir})
	assert.NoError(t, err)

	graphOutput := output.String()
	// DOT format should have standard DOT syntax
	assert.Contains(t, graphOutput, "digraph")
	assert.Contains(t, graphOutput, "AppConfig")
	assert.Contains(t, graphOutput, "WebApp")
	assert.Contains(t, graphOutput, "->")
}

func TestGraphCommand_OutputToFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "deps.go")
	outputFile := filepath.Join(tempDir, "graph.dot")

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

	err = app.Run([]string{"wetwire-k8s", "graph", "--output", outputFile, "--format", "dot", tempDir})
	assert.NoError(t, err)

	// Check that output file was created
	data, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "digraph")
	assert.Contains(t, string(data), "MyNamespace")
}

func TestGraphCommand_InvalidPath(t *testing.T) {
	app := newApp()
	err := app.Run([]string{"wetwire-k8s", "graph", "/nonexistent/path"})
	assert.Error(t, err)
}

func TestGraphCommand_NoResources(t *testing.T) {
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

	err = app.Run([]string{"wetwire-k8s", "graph", tempDir})
	assert.NoError(t, err)

	graphOutput := output.String()
	assert.Contains(t, strings.ToLower(graphOutput), "no resources")
}

func TestGraphCommand_MultipleDepChain(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "chain.go")
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

var WebDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "web-deployment",
	},
	Spec: appsv1.DeploymentSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
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

var WebService = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name: "web-service",
	},
	Spec: corev1.ServiceSpec{
		Selector: WebDeployment.Spec.Template.Labels,
	},
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "graph", "--format", "dot", tempDir})
	assert.NoError(t, err)

	graphOutput := output.String()
	// Should have all three resources
	assert.Contains(t, graphOutput, "AppConfig")
	assert.Contains(t, graphOutput, "WebDeployment")
	assert.Contains(t, graphOutput, "WebService")
	// Should show the dependency chain
	assert.Contains(t, graphOutput, "->")
}

func TestGraphCommand_NoDependencies(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "independent.go")
	content := `package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Config1 = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "config-1",
	},
}

var Config2 = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "config-2",
	},
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "graph", tempDir})
	assert.NoError(t, err)

	graphOutput := output.String()
	// Should show both resources even without dependencies
	assert.Contains(t, graphOutput, "Config1")
	assert.Contains(t, graphOutput, "Config2")
}
