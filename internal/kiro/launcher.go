package kiro

import (
	"context"
	"fmt"

	corekiro "github.com/lex00/wetwire-core-go/kiro"
)

// LaunchChat launches an interactive Kiro CLI chat session with the wetwire-k8s agent.
// It connects stdin/stdout directly to the terminal for interactive use.
func LaunchChat(agentName, initialPrompt string) error {
	// Ensure latest config is installed
	if err := EnsureInstalledWithForce(true); err != nil {
		return fmt.Errorf("installing kiro config: %w", err)
	}

	// Use core kiro package to launch
	config := NewConfig()
	return corekiro.Launch(context.Background(), config, initialPrompt)
}

// RunTest runs the Kiro agent in non-interactive test mode.
// Returns the output from the test session.
func RunTest(ctx context.Context, prompt string) (string, error) {
	// Ensure config is installed
	if err := EnsureInstalledWithForce(true); err != nil {
		return "", fmt.Errorf("installing kiro config: %w", err)
	}

	config := NewConfig()
	result, err := corekiro.RunTest(ctx, config, prompt)
	if err != nil {
		return "", err
	}
	return result.Output, nil
}
