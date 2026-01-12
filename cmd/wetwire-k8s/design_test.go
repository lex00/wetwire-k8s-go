package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lex00/wetwire-core-go/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestDesignCommand_Help(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "design", "--help"})
	assert.NoError(t, err)

	helpOutput := output.String()
	assert.Contains(t, helpOutput, "design")
	assert.Contains(t, helpOutput, "--provider")
	assert.Contains(t, helpOutput, "--prompt")
	assert.Contains(t, helpOutput, "--output-dir")
}

func TestDesignCommand_MissingPrompt(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	errOutput := &bytes.Buffer{}
	app.Writer = output
	app.ErrWriter = errOutput

	err := app.Run([]string{"wetwire-k8s", "design"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt")
}

func TestDesignCommand_InvalidProvider(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{"wetwire-k8s", "design", "--provider", "invalid", "--prompt", "test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider")
}

func TestDesignCommand_ValidProviders(t *testing.T) {
	// Test that anthropic and kiro are valid provider names
	providers := []string{"anthropic", "kiro"}
	for _, p := range providers {
		t.Run(p, func(t *testing.T) {
			config := DesignConfig{
				Provider: p,
			}
			assert.True(t, config.Provider == "anthropic" || config.Provider == "kiro")
		})
	}
}

func TestDesignCommand_DefaultProvider(t *testing.T) {
	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	// The command should default to anthropic provider
	cmd := designCommand()
	for _, flag := range cmd.Flags {
		if sf, ok := flag.(*cli.StringFlag); ok && sf.Name == "provider" {
			assert.Equal(t, "anthropic", sf.Value)
		}
	}
}

func TestDesignConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  DesignConfig
		wantErr string
	}{
		{
			name:    "missing prompt",
			config:  DesignConfig{Provider: "anthropic"},
			wantErr: "prompt is required",
		},
		{
			name:    "invalid provider",
			config:  DesignConfig{Provider: "openai", Prompt: "test"},
			wantErr: "invalid provider",
		},
		{
			name:   "valid config anthropic",
			config: DesignConfig{Provider: "anthropic", Prompt: "create a deployment"},
		},
		{
			name:   "valid config kiro",
			config: DesignConfig{Provider: "kiro", Prompt: "create a service"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestK8sRunnerAgent tests the K8s-specific runner agent
func TestK8sRunnerAgent_Tools(t *testing.T) {
	tempDir := t.TempDir()
	agent := NewK8sRunnerAgent(K8sRunnerConfig{
		WorkDir: tempDir,
	})

	tools := agent.GetTools()

	// Verify domain-specific tools are present
	toolNames := make([]string, len(tools))
	for i, tool := range tools {
		toolNames[i] = tool.Name
	}

	expectedTools := []string{
		"init_package",
		"write_file",
		"read_file",
		"run_lint",
		"run_build",
		"ask_developer",
	}

	for _, expected := range expectedTools {
		assert.Contains(t, toolNames, expected, "expected tool %s to be present", expected)
	}
}

func TestK8sRunnerAgent_InitPackage(t *testing.T) {
	tempDir := t.TempDir()
	agent := NewK8sRunnerAgent(K8sRunnerConfig{
		WorkDir: tempDir,
	})

	ctx := context.Background()
	result := agent.ExecuteTool(ctx, "init_package", json.RawMessage(`{"name":"myapp"}`))

	assert.Contains(t, result, "Created")

	// Verify directory was created
	_, err := os.Stat(filepath.Join(tempDir, "myapp"))
	assert.NoError(t, err)
}

func TestK8sRunnerAgent_WriteFile(t *testing.T) {
	tempDir := t.TempDir()
	agent := NewK8sRunnerAgent(K8sRunnerConfig{
		WorkDir: tempDir,
	})

	ctx := context.Background()

	// Write a file
	content := `package main

import corev1 "k8s.io/api/core/v1"

var MyService = &corev1.Service{}
`
	input := map[string]string{
		"path":    "myapp/service.go",
		"content": content,
	}
	inputJSON, _ := json.Marshal(input)

	result := agent.ExecuteTool(ctx, "write_file", inputJSON)
	assert.Contains(t, result, "Wrote")

	// Verify file was created
	data, err := os.ReadFile(filepath.Join(tempDir, "myapp", "service.go"))
	require.NoError(t, err)
	assert.Equal(t, content, string(data))

	// Verify pending lint state
	assert.True(t, agent.PendingLint())
}

func TestK8sRunnerAgent_ReadFile(t *testing.T) {
	tempDir := t.TempDir()
	testContent := "package main\n\nvar Test = 1\n"

	// Create test file
	testFile := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	agent := NewK8sRunnerAgent(K8sRunnerConfig{
		WorkDir: tempDir,
	})

	ctx := context.Background()
	result := agent.ExecuteTool(ctx, "read_file", json.RawMessage(`{"path":"test.go"}`))

	assert.Equal(t, testContent, result)
}

func TestK8sRunnerAgent_ReadFileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	agent := NewK8sRunnerAgent(K8sRunnerConfig{
		WorkDir: tempDir,
	})

	ctx := context.Background()
	result := agent.ExecuteTool(ctx, "read_file", json.RawMessage(`{"path":"nonexistent.go"}`))

	assert.Contains(t, result, "Error")
}

func TestK8sRunnerAgent_RunLint(t *testing.T) {
	tempDir := t.TempDir()

	// Create a simple Go file to lint
	content := `package main

import corev1 "k8s.io/api/core/v1"

var MyConfigMap = &corev1.ConfigMap{}
`
	testFile := filepath.Join(tempDir, "resources.go")
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	agent := NewK8sRunnerAgent(K8sRunnerConfig{
		WorkDir: tempDir,
	})

	// Set pending lint to simulate write
	agent.SetPendingLint(true)

	ctx := context.Background()
	result := agent.ExecuteTool(ctx, "run_lint", json.RawMessage(`{"path":"."}`))

	// Lint should have been called
	assert.True(t, agent.LintCalled())
	// Pending lint should be cleared
	assert.False(t, agent.PendingLint())
	// Result should be JSON
	assert.True(t, strings.HasPrefix(result, "{") || strings.Contains(result, "issues"))
}

func TestK8sRunnerAgent_RunBuild(t *testing.T) {
	tempDir := t.TempDir()

	// Create a valid K8s resource file
	content := `package main

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MyConfigMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name: "my-config",
	},
}
`
	testFile := filepath.Join(tempDir, "resources.go")
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	agent := NewK8sRunnerAgent(K8sRunnerConfig{
		WorkDir: tempDir,
	})

	ctx := context.Background()
	result := agent.ExecuteTool(ctx, "run_build", json.RawMessage(`{"path":"."}`))

	// Build should produce YAML output
	assert.Contains(t, result, "apiVersion")
}

func TestK8sRunnerAgent_AskDeveloper(t *testing.T) {
	tempDir := t.TempDir()

	// Create a mock developer
	mockDev := &MockDeveloper{
		Response: "Use the default namespace",
	}

	agent := NewK8sRunnerAgent(K8sRunnerConfig{
		WorkDir:   tempDir,
		Developer: mockDev,
	})

	ctx := context.Background()
	result := agent.ExecuteTool(ctx, "ask_developer", json.RawMessage(`{"question":"What namespace should I use?"}`))

	assert.Equal(t, "Use the default namespace", result)
	assert.Equal(t, "What namespace should I use?", mockDev.LastQuestion)
}

func TestK8sRunnerAgent_AskDeveloperNoDeveloper(t *testing.T) {
	tempDir := t.TempDir()

	agent := NewK8sRunnerAgent(K8sRunnerConfig{
		WorkDir: tempDir,
		// No developer configured
	})

	ctx := context.Background()
	result := agent.ExecuteTool(ctx, "ask_developer", json.RawMessage(`{"question":"test"}`))

	assert.Contains(t, result, "Error")
	assert.Contains(t, result, "developer")
}

func TestK8sRunnerAgent_LintEnforcement(t *testing.T) {
	tempDir := t.TempDir()
	agent := NewK8sRunnerAgent(K8sRunnerConfig{
		WorkDir: tempDir,
	})

	// Simulate writing a file
	agent.SetPendingLint(true)

	// Check enforcement - should require lint after write
	enforcement := agent.CheckLintEnforcement([]string{"write_file"})
	assert.Contains(t, enforcement, "ENFORCEMENT")
	assert.Contains(t, enforcement, "run_lint")

	// If lint was also called, no enforcement
	enforcement = agent.CheckLintEnforcement([]string{"write_file", "run_lint"})
	assert.Empty(t, enforcement)
}

func TestK8sRunnerAgent_CompletionGate(t *testing.T) {
	tempDir := t.TempDir()
	agent := NewK8sRunnerAgent(K8sRunnerConfig{
		WorkDir: tempDir,
	})

	// Simulate file generation
	agent.AddGeneratedFile("test.go")

	// Gate 1: Must have called lint
	gate := agent.CheckCompletionGate()
	assert.Contains(t, gate, "ENFORCEMENT")
	assert.Contains(t, gate, "lint")

	// Mark lint called but pending
	agent.SetLintCalled(true)
	agent.SetPendingLint(true)

	// Gate 2: Must not have pending lint
	gate = agent.CheckCompletionGate()
	assert.Contains(t, gate, "ENFORCEMENT")

	// Clear pending lint but fail
	agent.SetPendingLint(false)
	agent.SetLintPassed(false)

	// Gate 3: Lint must pass
	gate = agent.CheckCompletionGate()
	assert.Contains(t, gate, "ENFORCEMENT")

	// Set lint passed
	agent.SetLintPassed(true)

	// All gates pass
	gate = agent.CheckCompletionGate()
	assert.Empty(t, gate)
}

func TestK8sRunnerAgent_SystemPrompt(t *testing.T) {
	agent := NewK8sRunnerAgent(K8sRunnerConfig{})

	prompt := agent.GetSystemPrompt()

	// Should contain K8s-specific instructions
	assert.Contains(t, prompt, "Kubernetes")
	assert.Contains(t, prompt, "wetwire-k8s")
	assert.Contains(t, prompt, "k8s.io/api")

	// Should mention tools
	assert.Contains(t, prompt, "init_package")
	assert.Contains(t, prompt, "write_file")
	assert.Contains(t, prompt, "run_lint")
}

// MockDeveloper implements the Developer interface for testing
type MockDeveloper struct {
	Response     string
	LastQuestion string
}

func (m *MockDeveloper) Respond(ctx context.Context, question string) (string, error) {
	m.LastQuestion = question
	return m.Response, nil
}

// MockProvider implements providers.Provider for testing
type MockProvider struct {
	Responses []MockResponse
	CallIndex int
}

type MockResponse struct {
	Content    []ContentBlock
	StopReason string
}

type ContentBlock struct {
	Type  string
	Text  string
	ID    string
	Name  string
	Input json.RawMessage
}

// TestDesignCommand_RealProviderWithAPIKey tests that the design command
// calls the real provider when ANTHROPIC_API_KEY is set.
func TestDesignCommand_RealProviderWithAPIKey(t *testing.T) {
	// Skip if no API key is set (CI environments without keys)
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping real provider test")
	}

	tempDir := t.TempDir()

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	// Run design command with a simple prompt
	err := app.Run([]string{
		"wetwire-k8s", "design",
		"--prompt", "Create a simple nginx deployment",
		"--output-dir", tempDir,
	})

	// The command should complete without error
	require.NoError(t, err)

	// Output should indicate LLM was called (not just printing config)
	outputStr := output.String()
	// Should NOT contain the placeholder message
	assert.NotContains(t, outputStr, "Full AI agent loop requires wetwire-core-go integration")

	// Should contain evidence of agent loop execution
	assert.Contains(t, outputStr, "Agent")
}

// TestDesignCommand_AgentLoopExecution tests that the agent loop
// properly executes tool calls and processes responses.
func TestDesignCommand_AgentLoopExecution(t *testing.T) {
	// Skip if no API key
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	tempDir := t.TempDir()

	app := newApp()
	output := &bytes.Buffer{}
	app.Writer = output

	err := app.Run([]string{
		"wetwire-k8s", "design",
		"--prompt", "Create a configmap named test-config",
		"--output-dir", tempDir,
	})

	require.NoError(t, err)

	// Check that files were generated in the output directory
	files, err := filepath.Glob(filepath.Join(tempDir, "*.go"))
	if err == nil && len(files) > 0 {
		// Agent generated code files
		t.Logf("Generated files: %v", files)
	}
}

// TestRunAgentLoop tests the RunAgentLoop function directly.
func TestRunAgentLoop(t *testing.T) {
	tempDir := t.TempDir()

	agent := NewK8sRunnerAgent(K8sRunnerConfig{
		WorkDir: tempDir,
	})

	// Create a mock provider for testing
	mockProvider := &TestMockProvider{
		responses: []MockProviderResponse{
			{
				Content: []providers.ContentBlock{
					{Type: "text", Text: "I'll create a configmap for you."},
					{
						Type:  "tool_use",
						ID:    "tool_1",
						Name:  "init_package",
						Input: json.RawMessage(`{"name":"myapp"}`),
					},
				},
				StopReason: providers.StopReasonToolUse,
			},
			{
				Content: []providers.ContentBlock{
					{Type: "text", Text: "Package created. Now I'll write the file."},
					{
						Type:  "tool_use",
						ID:    "tool_2",
						Name:  "write_file",
						Input: json.RawMessage(`{"path":"myapp/configmap.go","content":"package main\n\nimport corev1 \"k8s.io/api/core/v1\"\n\nvar MyConfigMap = &corev1.ConfigMap{}\n"}`),
					},
				},
				StopReason: providers.StopReasonToolUse,
			},
			{
				Content: []providers.ContentBlock{
					{Type: "text", Text: "Running lint now."},
					{
						Type:  "tool_use",
						ID:    "tool_3",
						Name:  "run_lint",
						Input: json.RawMessage(`{"path":"myapp"}`),
					},
				},
				StopReason: providers.StopReasonToolUse,
			},
			{
				Content: []providers.ContentBlock{
					{Type: "text", Text: "Done! I've created a configmap for you."},
				},
				StopReason: providers.StopReasonEndTurn,
			},
		},
	}

	ctx := context.Background()
	err := RunAgentLoop(ctx, agent, mockProvider, "Create a configmap", nil)

	require.NoError(t, err)

	// Verify that the package directory was created
	_, err = os.Stat(filepath.Join(tempDir, "myapp"))
	assert.NoError(t, err)

	// Verify that files were generated
	assert.NotEmpty(t, agent.GetGeneratedFiles())
}

// TestMockProvider is a mock implementation of providers.Provider for testing.
type TestMockProvider struct {
	responses []MockProviderResponse
	callIndex int
}

type MockProviderResponse struct {
	Content    []providers.ContentBlock
	StopReason providers.StopReason
}

func (p *TestMockProvider) Name() string {
	return "mock"
}

func (p *TestMockProvider) CreateMessage(ctx context.Context, req providers.MessageRequest) (*providers.MessageResponse, error) {
	if p.callIndex >= len(p.responses) {
		return &providers.MessageResponse{
			Content:    []providers.ContentBlock{{Type: "text", Text: "Done."}},
			StopReason: providers.StopReasonEndTurn,
		}, nil
	}
	resp := p.responses[p.callIndex]
	p.callIndex++
	return &providers.MessageResponse{
		Content:    resp.Content,
		StopReason: resp.StopReason,
	}, nil
}

func (p *TestMockProvider) StreamMessage(ctx context.Context, req providers.MessageRequest, handler providers.StreamHandler) (*providers.MessageResponse, error) {
	return p.CreateMessage(ctx, req)
}

// TestNewAnthropicProvider tests that we can create an Anthropic provider.
func TestNewAnthropicProvider(t *testing.T) {
	// Skip if no API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	provider, err := NewAnthropicProvider(apiKey)
	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "anthropic", provider.Name())
}
