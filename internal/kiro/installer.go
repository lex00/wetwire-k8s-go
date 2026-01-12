package kiro

import (
	"encoding/json"
	"fmt"
	"os"

	corekiro "github.com/lex00/wetwire-core-go/kiro"
)

// AgentConfig represents a Kiro agent configuration.
type AgentConfig struct {
	Name       string     `json:"name"`
	Prompt     string     `json:"prompt"`
	MCPServers []MCPEntry `json:"mcpServers"`
}

// MCPEntry represents an MCP server configuration.
type MCPEntry struct {
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Cwd     string   `json:"cwd,omitempty"`
}

// EnsureInstalled installs the Kiro agent configuration if not already present.
func EnsureInstalled() error {
	return EnsureInstalledWithForce(false)
}

// EnsureInstalledWithForce installs the Kiro agent configuration.
// If force is true, overwrites any existing configuration.
func EnsureInstalledWithForce(force bool) error {
	// Use core kiro Install function
	config := NewConfig()
	return corekiro.Install(config)
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// installAgentConfig installs the wetwire-k8s agent configuration.
// This is kept for backwards compatibility but uses core kiro.Install.
func installAgentConfig(path string) error {
	// Determine MCP server command (prefer binary, fallback to go run)
	binaryPath, err := os.Executable()
	if err != nil {
		binaryPath = "go"
	}

	var mcpEntry MCPEntry
	if binaryPath == "go" {
		mcpEntry = MCPEntry{
			Name:    "wetwire",
			Command: "go",
			Args:    []string{"run", ".", "design", "--mcp-server"},
		}
	} else {
		mcpEntry = MCPEntry{
			Name:    "wetwire",
			Command: binaryPath,
			Args:    []string{"design", "--mcp-server"},
		}
	}

	// Build config using constants from config.go
	config := AgentConfig{
		Name:       AgentName,
		Prompt:     AgentPrompt,
		MCPServers: []MCPEntry{mcpEntry},
	}

	// Write config
	output, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, output, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// installMCPConfig installs the project-level MCP configuration.
// This is kept for backwards compatibility but uses core kiro.Install.
func installMCPConfig(path string) error {
	cwd, _ := os.Getwd()

	// Determine command (prefer binary, fallback to go run)
	binaryPath, err := os.Executable()
	if err != nil {
		binaryPath = "go"
	}

	var mcpConfig map[string]MCPEntry
	if binaryPath == "go" {
		mcpConfig = map[string]MCPEntry{
			"wetwire": {
				Command: "go",
				Args:    []string{"run", ".", "design", "--mcp-server"},
				Cwd:     cwd,
			},
		}
	} else {
		mcpConfig = map[string]MCPEntry{
			"wetwire": {
				Command: binaryPath,
				Args:    []string{"design", "--mcp-server"},
				Cwd:     cwd,
			},
		}
	}

	output, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal mcp config: %w", err)
	}

	if err := os.WriteFile(path, output, 0644); err != nil {
		return fmt.Errorf("write mcp config: %w", err)
	}

	return nil
}
