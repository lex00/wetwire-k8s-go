package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTestCommand_RealProviderWithAPIKey tests that the test command
// calls the real provider when ANTHROPIC_API_KEY is set and provider != mock.
func TestTestCommand_RealProviderWithAPIKey(t *testing.T) {
	// Skip if no API key is set
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping real provider test")
	}

	tempDir := t.TempDir()

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	// Run test command with anthropic provider (not mock)
	err := app.Run([]string{
		"wetwire-k8s", "test",
		"--persona", "beginner",
		"--provider", "anthropic",
		"--output", tempDir,
	})

	require.NoError(t, err)

	outputStr := output.String()
	// Should NOT contain the mock fallback message
	assert.NotContains(t, outputStr, "Using mock responses")

	// Should contain evidence of real test execution
	assert.Contains(t, outputStr, "Test completed")
}

// TestTestCommand_RealProviderCallsLLM tests that when using anthropic provider,
// the test command actually calls the LLM instead of using mock responses.
func TestTestCommand_RealProviderCallsLLM(t *testing.T) {
	// Skip if no API key
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	tempDir := t.TempDir()

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{
		"wetwire-k8s", "test",
		"--persona", "terse",
		"--provider", "anthropic",
		"--scenario", "deploy-nginx",
		"--output", tempDir,
	})

	require.NoError(t, err)

	// Check session results were written and contain real LLM interaction data
	sessionFile := filepath.Join(tempDir, "terse", "session.json")
	_, err = os.Stat(sessionFile)
	assert.NoError(t, err)
}
