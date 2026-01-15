package kiro

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEmbeddedConfigs_ValidJSON(t *testing.T) {
	// Test that NewConfig returns a valid configuration
	config := NewConfig()

	// AgentName should be set
	if config.AgentName == "" {
		t.Error("NewConfig should return non-empty AgentName")
	}

	if config.AgentName != AgentName {
		t.Errorf("AgentName = %q, want %q", config.AgentName, AgentName)
	}

	// AgentPrompt should not be empty
	if config.AgentPrompt == "" {
		t.Error("NewConfig should return non-empty AgentPrompt")
	}

	// MCPCommand should not be empty
	if config.MCPCommand == "" {
		t.Error("NewConfig should return non-empty MCPCommand")
	}
}

func TestInstall_HasToolsArray(t *testing.T) {
	// Test that the generated config includes tools array
	// Required for kiro to enable MCP tool usage
	// See: https://github.com/aws/amazon-q-developer-cli/issues/2640

	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	homeDir := filepath.Join(tmpDir, "home")

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Override home directory
	t.Setenv("HOME", homeDir)

	// Override working directory for the install
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	// Run install
	if err := EnsureInstalledWithForce(true); err != nil {
		t.Fatalf("EnsureInstalledWithForce failed: %v", err)
	}

	// Read the agent config
	agentPath := filepath.Join(homeDir, ".kiro", "agents", AgentName+".json")
	data, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("failed to read agent config: %v", err)
	}

	var agent map[string]any
	if err := json.Unmarshal(data, &agent); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Must have tools array
	tools, ok := agent["tools"].([]any)
	if !ok {
		t.Fatal("agent config must have 'tools' array - required for kiro MCP tool usage")
	}

	// Must have at least one tool reference
	if len(tools) == 0 {
		t.Error("tools array must not be empty")
	}

	// First tool should be @server_name format
	if len(tools) > 0 {
		tool, ok := tools[0].(string)
		if !ok || len(tool) == 0 || tool[0] != '@' {
			t.Errorf("tools should use @server_name format, got: %v", tools)
		}
	}
}

func TestInstall_HasCwd(t *testing.T) {
	// Test that the generated config includes cwd for MCP server
	// This ensures the provider runs in the correct directory
	// See: wetwire-core-go v1.5.4 fix

	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	homeDir := filepath.Join(tmpDir, "home")

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Override home directory
	t.Setenv("HOME", homeDir)

	// Override working directory for the install
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	// Run install
	if err := EnsureInstalledWithForce(true); err != nil {
		t.Fatalf("EnsureInstalledWithForce failed: %v", err)
	}

	// Read the agent config
	agentPath := filepath.Join(homeDir, ".kiro", "agents", AgentName+".json")
	data, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("failed to read agent config: %v", err)
	}

	var agent map[string]any
	if err := json.Unmarshal(data, &agent); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Check mcpServers map
	mcpServers, ok := agent["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("agent config must have 'mcpServers' map")
	}

	if len(mcpServers) == 0 {
		t.Fatal("mcpServers map must not be empty")
	}

	// Resolve symlinks in project dir for comparison (macOS uses /private/var -> /var)
	expectedCwd, err := filepath.EvalSymlinks(projectDir)
	if err != nil {
		expectedCwd = projectDir
	}

	// Check MCP server has cwd
	// The server should be named MCPCommand from config
	var foundCwd bool
	for serverName, serverConfigAny := range mcpServers {
		serverConfig, ok := serverConfigAny.(map[string]any)
		if !ok {
			continue
		}

		cwd, ok := serverConfig["cwd"].(string)
		if ok && cwd != "" {
			foundCwd = true
			// Verify cwd is the project directory (resolve symlinks for comparison)
			resolvedCwd, err := filepath.EvalSymlinks(cwd)
			if err != nil {
				resolvedCwd = cwd
			}
			if resolvedCwd != expectedCwd {
				t.Errorf("server %q: cwd = %q, want %q", serverName, resolvedCwd, expectedCwd)
			}
		}
	}

	if !foundCwd {
		t.Fatal("MCP server config must have 'cwd' field set - required for provider to run in correct directory")
	}
}
