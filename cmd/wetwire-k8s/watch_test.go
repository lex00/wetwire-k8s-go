package main

import (
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
	stdout, _, err := runTestCommand([]string{"watch", "--help"})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "monitors Go source files")
	assert.Contains(t, stdout.String(), "--output")
	assert.Contains(t, stdout.String(), "--interval")
}

func TestWatchCommand_NonExistentPath(t *testing.T) {
	_, _, err := runTestCommand([]string{"watch", "/nonexistent/path"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestRunWatchWithContext(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an empty Go file
	err := os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("package main\n"), 0644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	var buildCalled bool
	var mu sync.Mutex

	err = runWatchWithContextAndCallback(ctx, tmpDir, "-", 100*time.Millisecond, &discardWriter{}, func() {
		mu.Lock()
		buildCalled = true
		mu.Unlock()
	})

	// Context should have been canceled
	assert.Equal(t, context.DeadlineExceeded, err)

	// Initial build should have been called
	mu.Lock()
	assert.True(t, buildCalled)
	mu.Unlock()
}

// discardWriter implements io.Writer and discards all output
type discardWriter struct{}

func (d *discardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
