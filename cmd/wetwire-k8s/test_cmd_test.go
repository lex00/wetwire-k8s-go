package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestCommand_Help(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "test", "--help"})
	assert.NoError(t, err)

	helpOutput := output.String()
	assert.Contains(t, helpOutput, "test")
	assert.Contains(t, helpOutput, "--persona")
	assert.Contains(t, helpOutput, "--provider")
	assert.Contains(t, helpOutput, "--all-personas")
	assert.Contains(t, helpOutput, "--scenario")
}

func TestTestCommand_ListPersonas(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "test", "--list-personas"})
	assert.NoError(t, err)

	listOutput := output.String()
	assert.Contains(t, listOutput, "beginner")
	assert.Contains(t, listOutput, "intermediate")
	assert.Contains(t, listOutput, "expert")
	assert.Contains(t, listOutput, "terse")
	assert.Contains(t, listOutput, "verbose")
}

func TestTestCommand_InvalidPersona(t *testing.T) {
	app := newApp()

	err := app.Run([]string{"wetwire-k8s", "test", "--persona", "nonexistent"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown persona")
}

func TestTestCommand_DryRun(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "test", "--persona", "beginner", "--dry-run"})
	assert.NoError(t, err)

	dryRunOutput := output.String()
	assert.Contains(t, dryRunOutput, "beginner")
	assert.Contains(t, dryRunOutput, "dry-run")
}

func TestTestCommand_OutputDir(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "results")

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "test", "--persona", "beginner", "--output", outputDir, "--dry-run"})
	assert.NoError(t, err)

	// In dry-run mode, output dir should be created but files not written
	dryRunOutput := output.String()
	assert.Contains(t, dryRunOutput, outputDir)
}

func TestTestCommand_AllPersonas(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "test", "--all-personas", "--dry-run"})
	assert.NoError(t, err)

	allOutput := output.String()
	// Should mention all personas
	assert.Contains(t, allOutput, "beginner")
	assert.Contains(t, allOutput, "intermediate")
	assert.Contains(t, allOutput, "expert")
	assert.Contains(t, allOutput, "terse")
	assert.Contains(t, allOutput, "verbose")
}

func TestTestCommand_Scenario(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "test", "--persona", "beginner", "--scenario", "deploy-nginx", "--dry-run"})
	assert.NoError(t, err)

	scenarioOutput := output.String()
	assert.Contains(t, scenarioOutput, "deploy-nginx")
}

func TestTestCommand_Provider(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "test", "--persona", "beginner", "--provider", "anthropic", "--dry-run"})
	assert.NoError(t, err)

	providerOutput := output.String()
	assert.Contains(t, providerOutput, "anthropic")
}

func TestTestCommand_InvalidProvider(t *testing.T) {
	app := newApp()

	err := app.Run([]string{"wetwire-k8s", "test", "--persona", "beginner", "--provider", "invalid-provider"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider")
}

func TestTestCommand_RequiresPersonaOrAllPersonas(t *testing.T) {
	app := newApp()

	// Without --persona or --all-personas, should fail
	err := app.Run([]string{"wetwire-k8s", "test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "persona")
}

func TestTestCommand_MockSession(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "results")

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	// Use --mock to run with mocked LLM responses
	err := app.Run([]string{"wetwire-k8s", "test", "--persona", "beginner", "--output", outputDir, "--mock"})
	assert.NoError(t, err)

	// Check that result files were created
	resultsPath := filepath.Join(outputDir, "beginner", "RESULTS.md")
	_, err = os.Stat(resultsPath)
	assert.NoError(t, err, "RESULTS.md should exist")

	sessionPath := filepath.Join(outputDir, "beginner", "session.json")
	_, err = os.Stat(sessionPath)
	assert.NoError(t, err, "session.json should exist")

	scorePath := filepath.Join(outputDir, "beginner", "score.json")
	_, err = os.Stat(scorePath)
	assert.NoError(t, err, "score.json should exist")

	// Verify score.json content
	scoreData, err := os.ReadFile(scorePath)
	require.NoError(t, err)

	var score map[string]interface{}
	err = json.Unmarshal(scoreData, &score)
	require.NoError(t, err)
	assert.Contains(t, score, "Persona")
}

func TestTestCommand_ScoreDimensions(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "results")

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "test", "--persona", "expert", "--output", outputDir, "--mock"})
	assert.NoError(t, err)

	// Read and verify score has all dimensions
	scorePath := filepath.Join(outputDir, "expert", "score.json")
	scoreData, err := os.ReadFile(scorePath)
	require.NoError(t, err)

	var score map[string]interface{}
	err = json.Unmarshal(scoreData, &score)
	require.NoError(t, err)

	// Verify dimensions exist (as nested objects)
	assert.Contains(t, score, "Completeness")
	assert.Contains(t, score, "LintQuality")
	assert.Contains(t, score, "CodeQuality")
	assert.Contains(t, score, "OutputValidity")
	assert.Contains(t, score, "QuestionEfficiency")
}

func TestTestCommand_ResultsMarkdown(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "results")

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "test", "--persona", "terse", "--output", outputDir, "--mock"})
	assert.NoError(t, err)

	// Read and verify RESULTS.md content
	resultsPath := filepath.Join(outputDir, "terse", "RESULTS.md")
	resultsData, err := os.ReadFile(resultsPath)
	require.NoError(t, err)

	resultsContent := string(resultsData)
	assert.Contains(t, resultsContent, "# Session Results")
	assert.Contains(t, resultsContent, "**Persona:** terse")
	assert.Contains(t, resultsContent, "## Score")
}

func TestTestCommand_VerboseOutput(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "test", "--persona", "beginner", "--verbose", "--dry-run"})
	assert.NoError(t, err)

	verboseOutput := output.String()
	// Verbose mode should show more details
	assert.Contains(t, verboseOutput, "Traits")
}
