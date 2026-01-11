package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffCommand_Help(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "diff", "--help"})
	assert.NoError(t, err)

	helpOutput := output.String()
	assert.Contains(t, helpOutput, "diff")
	assert.Contains(t, helpOutput, "--against")
	assert.Contains(t, helpOutput, "--semantic")
	assert.Contains(t, helpOutput, "--color")
}

func TestDiffCommand_MissingAgainstFlag(t *testing.T) {
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
	err = app.Run([]string{"wetwire-k8s", "diff", tmpDir})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "against")
}

func TestDiffCommand_NoDifference(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Go source file
	goFile := filepath.Join(tmpDir, "resources.go")
	err := os.WriteFile(goFile, []byte(`package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyConfigMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-config-map",
	},
}
`), 0644)
	require.NoError(t, err)

	// Create existing manifest that matches the generated output
	// Note: use same 4-space indentation as generated YAML
	manifestFile := filepath.Join(tmpDir, "manifest.yaml")
	err = os.WriteFile(manifestFile, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
    name: my-config-map
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "diff", "--against", manifestFile, tmpDir})
	assert.NoError(t, err)
	// No differences should be output or a message saying no differences
	out := output.String()
	// When there are no differences, the output should be empty or indicate no differences
	assert.True(t, out == "" || out == "No differences found\n", "Expected empty output or 'No differences found' message, got: %s", out)
}

func TestDiffCommand_WithDifference(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Go source file
	goFile := filepath.Join(tmpDir, "resources.go")
	err := os.WriteFile(goFile, []byte(`package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyConfigMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-config-map",
	},
}
`), 0644)
	require.NoError(t, err)

	// Create existing manifest with different name
	manifestFile := filepath.Join(tmpDir, "manifest.yaml")
	err = os.WriteFile(manifestFile, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: different-config
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "diff", "--against", manifestFile, tmpDir})
	assert.NoError(t, err)
	// Should show differences
	diffOutput := output.String()
	assert.NotEmpty(t, diffOutput)
}

func TestDiffCommand_SemanticMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Go source file
	goFile := filepath.Join(tmpDir, "resources.go")
	err := os.WriteFile(goFile, []byte(`package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyConfigMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-config-map",
	},
}
`), 0644)
	require.NoError(t, err)

	// Create existing manifest with same content but different formatting/ordering
	manifestFile := filepath.Join(tmpDir, "manifest.yaml")
	err = os.WriteFile(manifestFile, []byte(`kind: ConfigMap
apiVersion: v1
metadata:
  name: my-config-map
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	// Semantic mode should consider these equivalent
	err = app.Run([]string{"wetwire-k8s", "diff", "--against", manifestFile, "--semantic", tmpDir})
	assert.NoError(t, err)
}

func TestDiffCommand_TextMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Go source file
	goFile := filepath.Join(tmpDir, "resources.go")
	err := os.WriteFile(goFile, []byte(`package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyConfigMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-config-map",
	},
}
`), 0644)
	require.NoError(t, err)

	// Create existing manifest with different name
	manifestFile := filepath.Join(tmpDir, "manifest.yaml")
	err = os.WriteFile(manifestFile, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: other-config
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	// Default is text mode (line-by-line)
	err = app.Run([]string{"wetwire-k8s", "diff", "--against", manifestFile, tmpDir})
	assert.NoError(t, err)
	// Should show line-by-line diff with +/- markers
	diffOutput := output.String()
	assert.NotEmpty(t, diffOutput)
}

func TestDiffCommand_ColorMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Go source file
	goFile := filepath.Join(tmpDir, "resources.go")
	err := os.WriteFile(goFile, []byte(`package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyConfigMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-config-map",
	},
}
`), 0644)
	require.NoError(t, err)

	// Create existing manifest with different name
	manifestFile := filepath.Join(tmpDir, "manifest.yaml")
	err = os.WriteFile(manifestFile, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: other-config
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	// With color flag
	err = app.Run([]string{"wetwire-k8s", "diff", "--against", manifestFile, "--color", tmpDir})
	assert.NoError(t, err)
	// Color mode should include ANSI escape codes
	// Note: this test just verifies the flag is accepted; actual color output depends on implementation
}

func TestDiffCommand_InvalidSourcePath(t *testing.T) {
	tmpDir := t.TempDir()
	manifestFile := filepath.Join(tmpDir, "manifest.yaml")
	err := os.WriteFile(manifestFile, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
`), 0644)
	require.NoError(t, err)

	app := newApp()
	err = app.Run([]string{"wetwire-k8s", "diff", "--against", manifestFile, "/nonexistent/path"})
	assert.Error(t, err)
}

func TestDiffCommand_InvalidAgainstPath(t *testing.T) {
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
	err = app.Run([]string{"wetwire-k8s", "diff", "--against", "/nonexistent/manifest.yaml", tmpDir})
	assert.Error(t, err)
}

func TestDiffCommand_MultipleResources(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Go source file with multiple resources
	goFile := filepath.Join(tmpDir, "resources.go")
	err := os.WriteFile(goFile, []byte(`package main

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyConfigMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-config-map",
	},
}

var MyDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-deployment",
	},
}
`), 0644)
	require.NoError(t, err)

	// Create existing manifest with multiple documents
	manifestFile := filepath.Join(tmpDir, "manifest.yaml")
	err = os.WriteFile(manifestFile, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config-map
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "diff", "--against", manifestFile, tmpDir})
	assert.NoError(t, err)
}

func TestDiffCommand_NoResources(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Go source file with no k8s resources
	goFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(goFile, []byte(`package main

func main() {}
`), 0644)
	require.NoError(t, err)

	// Create existing manifest
	manifestFile := filepath.Join(tmpDir, "manifest.yaml")
	err = os.WriteFile(manifestFile, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "diff", "--against", manifestFile, tmpDir})
	// Should handle gracefully
	assert.NoError(t, err)
}

func TestDiffCommand_DefaultPath(t *testing.T) {
	// Create temp directory with Go source
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "resources.go")
	err := os.WriteFile(goFile, []byte(`package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyConfigMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-config-map",
	},
}
`), 0644)
	require.NoError(t, err)

	manifestFile := filepath.Join(tmpDir, "manifest.yaml")
	err = os.WriteFile(manifestFile, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config-map
`), 0644)
	require.NoError(t, err)

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	// No path argument should default to current directory
	err = app.Run([]string{"wetwire-k8s", "diff", "--against", manifestFile})
	assert.NoError(t, err)
}

func TestDiffCommand_SemanticWithDifferences(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Go source file
	goFile := filepath.Join(tmpDir, "resources.go")
	err := os.WriteFile(goFile, []byte(`package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyConfigMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-config-map",
	},
}
`), 0644)
	require.NoError(t, err)

	// Create manifest with semantically different content
	manifestFile := filepath.Join(tmpDir, "manifest.yaml")
	err = os.WriteFile(manifestFile, []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: different-name
`), 0644)
	require.NoError(t, err)

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err = app.Run([]string{"wetwire-k8s", "diff", "--against", manifestFile, "--semantic", tmpDir})
	assert.NoError(t, err)
	// Should show semantic differences
	diffOutput := output.String()
	assert.NotEmpty(t, diffOutput)
}
