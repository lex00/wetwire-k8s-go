package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestCommand_Help(t *testing.T) {
	stdout, _, err := runTestCommand([]string{"test", "--help"})
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "persona")
	assert.Contains(t, stdout.String(), "--prompt")
	assert.Contains(t, stdout.String(), "--all-personas")
}

func TestTestCommand_MissingPrompt(t *testing.T) {
	_, _, err := runTestCommand([]string{"test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt")
}
