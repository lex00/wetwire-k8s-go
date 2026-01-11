package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatchCommand_Help(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "watch", "--help"})
	assert.NoError(t, err)

	helpOutput := output.String()
	assert.Contains(t, helpOutput, "watch")
	assert.Contains(t, helpOutput, "--output")
	assert.Contains(t, helpOutput, "--interval")
}

func TestWatchCommand_InvalidPath(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	errOutput := &bytes.Buffer{}
	app.Writer = output
	app.ErrWriter = errOutput

	err := app.Run([]string{"wetwire-k8s", "watch", "/nonexistent/path"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestWatchCommand_InitialBuild(t *testing.T) {
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
		Name: "my-config",
	},
}
`), 0644)
	require.NoError(t, err)

	outputFile := filepath.Join(tmpDir, "output.yaml")

	// Use a context with timeout to stop the watch
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	// Run watch in a goroutine
	var wg sync.WaitGroup
	var watchErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		watchErr = runWatchWithContext(ctx, tmpDir, outputFile, 100*time.Millisecond, output)
	}()

	// Wait for context to timeout
	wg.Wait()

	// Should have created output file from initial build
	if watchErr != nil && watchErr != context.DeadlineExceeded {
		t.Fatalf("unexpected error: %v", watchErr)
	}

	// Check if output file was created
	_, err = os.Stat(outputFile)
	assert.NoError(t, err, "output file should exist after initial build")
}

func TestWatchCommand_RebuildOnChange(t *testing.T) {
	tmpDir := t.TempDir()

	// Create initial Go source file
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

	outputFile := filepath.Join(tmpDir, "output.yaml")

	// Use a context with longer timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	output := &bytes.Buffer{}

	// Run watch in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runWatchWithContext(ctx, tmpDir, outputFile, 100*time.Millisecond, output)
	}()

	// Wait for initial build
	time.Sleep(500 * time.Millisecond)

	// Modify the source file - change variable name which changes the generated resource name
	// The build process uses the variable name (converted to k8s format) as the resource name
	err = os.WriteFile(goFile, []byte(`package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var UpdatedConfig = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "updated-config",
	},
}
`), 0644)
	require.NoError(t, err)

	// Wait for rebuild (debounce interval + build time)
	time.Sleep(800 * time.Millisecond)

	// Cancel and wait
	cancel()
	wg.Wait()

	// Check output contains updated content (variable name converted to k8s format)
	data, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "updated-config")
}

func TestWatchCommand_Debouncing(t *testing.T) {
	tmpDir := t.TempDir()

	// Create initial Go source file
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

	outputFile := filepath.Join(tmpDir, "output.yaml")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	output := &bytes.Buffer{}
	buildCount := 0
	buildCountMutex := sync.Mutex{}

	// Run watch with build counter
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runWatchWithContextAndCallback(ctx, tmpDir, outputFile, 200*time.Millisecond, output, func() {
			buildCountMutex.Lock()
			buildCount++
			buildCountMutex.Unlock()
		})
	}()

	// Wait for initial build
	time.Sleep(400 * time.Millisecond)

	// Make multiple rapid changes (should be debounced to single build)
	for i := 0; i < 5; i++ {
		err := os.WriteFile(goFile, []byte(`package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyConfigMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "rapid-change",
	},
}
`), 0644)
		require.NoError(t, err)
		time.Sleep(20 * time.Millisecond)
	}

	// Wait for debounced build
	time.Sleep(500 * time.Millisecond)

	cancel()
	wg.Wait()

	// Should have at most 2-3 builds (initial + debounced), not 5+
	buildCountMutex.Lock()
	count := buildCount
	buildCountMutex.Unlock()

	assert.LessOrEqual(t, count, 4, "builds should be debounced (got %d builds)", count)
}

func TestWatchCommand_OutputToStdout(t *testing.T) {
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
		Name: "my-config",
	},
}
`), 0644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	output := &bytes.Buffer{}

	// Run watch with stdout output (no output file)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runWatchWithContext(ctx, tmpDir, "-", 100*time.Millisecond, output)
	}()

	wg.Wait()

	// Output should contain generated YAML
	outputStr := output.String()
	assert.Contains(t, outputStr, "ConfigMap")
}

func TestWatchCommand_IntervalFlag(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	// Test that interval flag is accepted (help output check)
	err := app.Run([]string{"wetwire-k8s", "watch", "--help"})
	assert.NoError(t, err)

	helpOutput := output.String()
	assert.Contains(t, helpOutput, "--interval")
}

func TestWatchCommand_DefaultPath(t *testing.T) {
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
		Name: "my-config",
	},
}
`), 0644)
	require.NoError(t, err)

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	outputFile := filepath.Join(tmpDir, "output.yaml")

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	output := &bytes.Buffer{}

	// Run watch without path argument
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runWatchWithContext(ctx, ".", outputFile, 100*time.Millisecond, output)
	}()

	wg.Wait()

	// Should have created output file
	_, err = os.Stat(outputFile)
	assert.NoError(t, err, "output file should exist")
}

func TestWatchCommand_MultipleGoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple Go source files
	goFile1 := filepath.Join(tmpDir, "configmap.go")
	err := os.WriteFile(goFile1, []byte(`package main

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

	goFile2 := filepath.Join(tmpDir, "deployment.go")
	err = os.WriteFile(goFile2, []byte(`package main

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-deployment",
	},
}
`), 0644)
	require.NoError(t, err)

	outputFile := filepath.Join(tmpDir, "output.yaml")

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	output := &bytes.Buffer{}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runWatchWithContext(ctx, tmpDir, outputFile, 100*time.Millisecond, output)
	}()

	wg.Wait()

	// Check output contains both resources
	data, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "ConfigMap")
	assert.Contains(t, content, "Deployment")
}

func TestWatchCommand_NoResources(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Go source file with no k8s resources
	goFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(goFile, []byte(`package main

func main() {}
`), 0644)
	require.NoError(t, err)

	outputFile := filepath.Join(tmpDir, "output.yaml")

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	output := &bytes.Buffer{}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runWatchWithContext(ctx, tmpDir, outputFile, 100*time.Millisecond, output)
	}()

	wg.Wait()

	// Should handle gracefully - no error, but may not create output file
	// or create empty output file
}

func TestWatchCommand_NewFileAdded(t *testing.T) {
	tmpDir := t.TempDir()

	// Create initial Go source file
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

	outputFile := filepath.Join(tmpDir, "output.yaml")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	output := &bytes.Buffer{}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runWatchWithContext(ctx, tmpDir, outputFile, 100*time.Millisecond, output)
	}()

	// Wait for initial build
	time.Sleep(400 * time.Millisecond)

	// Add a new Go file
	newGoFile := filepath.Join(tmpDir, "service.go")
	err = os.WriteFile(newGoFile, []byte(`package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyService = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-service",
	},
}
`), 0644)
	require.NoError(t, err)

	// Wait for rebuild
	time.Sleep(500 * time.Millisecond)

	cancel()
	wg.Wait()

	// Check output contains the new resource
	data, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "Service")
}
