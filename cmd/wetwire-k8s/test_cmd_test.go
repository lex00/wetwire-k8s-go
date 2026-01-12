package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestCommand_Personas(t *testing.T) {
	stdout, _, err := runTestCommand([]string{"test", "--help"})
	assert.NoError(t, err)

	// Check that persona options are documented
	helpText := stdout.String()
	assert.Contains(t, helpText, "beginner")
	assert.Contains(t, helpText, "intermediate")
	assert.Contains(t, helpText, "expert")
	assert.Contains(t, helpText, "terse")
	assert.Contains(t, helpText, "verbose")
}
