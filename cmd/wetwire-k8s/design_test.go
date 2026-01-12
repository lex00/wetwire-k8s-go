package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDesignCommand_Help(t *testing.T) {
	rootCmd := newDesignCmd()
	output := &bytes.Buffer{}
	rootCmd.SetOut(output)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	assert.NoError(t, err)

	helpOutput := output.String()
	assert.Contains(t, helpOutput, "design")
	assert.Contains(t, helpOutput, "--prompt")
	assert.Contains(t, helpOutput, "--output-dir")
}

func TestDesignCommand_MissingPrompt(t *testing.T) {
	rootCmd := newDesignCmd()
	output := &bytes.Buffer{}
	errOutput := &bytes.Buffer{}
	rootCmd.SetOut(output)
	rootCmd.SetErr(errOutput)
	rootCmd.SetArgs([]string{})

	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt")
}

func TestK8sDomain(t *testing.T) {
	domain := K8sDomain()

	assert.Equal(t, "k8s", domain.Name)
	assert.Equal(t, "wetwire-k8s", domain.CLICommand)
	assert.Equal(t, "Kubernetes YAML", domain.OutputFormat)
	assert.Contains(t, domain.SystemPrompt, "Kubernetes")
	assert.Contains(t, domain.SystemPrompt, "wetwire-k8s")
}
