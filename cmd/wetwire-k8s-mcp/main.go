// Package main implements an MCP (Model Context Protocol) server for wetwire-k8s-go.
// This server enables Claude Code and other MCP clients to interact with wetwire-k8s
// tools for Kubernetes manifest generation and management.
package main

import (
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Version is set at build time.
var Version = "1.0.0"

// debugLog logs messages when WETWIRE_MCP_DEBUG is set.
func debugLog(format string, args ...interface{}) {
	if os.Getenv("WETWIRE_MCP_DEBUG") != "" {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func main() {
	debugLog("Starting wetwire-k8s MCP server version %s", Version)

	// Create MCP server
	s := server.NewMCPServer(
		"wetwire-k8s",
		Version,
		server.WithToolCapabilities(false),
	)

	// Register tools
	registerTools(s)

	debugLog("Registered all tools, starting stdio server")

	// Start stdio server
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

// registerTools adds all wetwire-k8s tools to the MCP server.
func registerTools(s *server.MCPServer) {
	// Build tool - generate Kubernetes YAML from Go code
	s.AddTool(
		mcp.NewTool("build",
			mcp.WithDescription("Generate Kubernetes YAML manifests from Go code. Discovers Kubernetes resource declarations in Go source files and outputs YAML or JSON manifests."),
			mcp.WithString("path",
				mcp.Description("Path to Go package or file to build. If not specified, uses the current directory."),
			),
			mcp.WithString("format",
				mcp.Description("Output format: 'yaml' (default) or 'json'."),
				mcp.Enum("yaml", "json"),
			),
		),
		buildHandler,
	)
	debugLog("Registered tool: build")

	// Lint tool - enforce patterns and best practices
	s.AddTool(
		mcp.NewTool("lint",
			mcp.WithDescription("Lint Go files for wetwire-k8s patterns and best practices. Enforces flat, declarative patterns and identifies potential issues."),
			mcp.WithString("path",
				mcp.Description("Path to Go file or directory to lint. If not specified, uses the current directory."),
			),
			mcp.WithString("format",
				mcp.Description("Output format: 'text' (default), 'json', or 'github'."),
				mcp.Enum("text", "json", "github"),
			),
		),
		lintHandler,
	)
	debugLog("Registered tool: lint")

	// Import tool - convert YAML to Go code
	s.AddTool(
		mcp.NewTool("import",
			mcp.WithDescription("Convert Kubernetes YAML manifests to Go code using the wetwire pattern. Supports multi-document YAML."),
			mcp.WithString("yaml",
				mcp.Description("YAML content to convert to Go code."),
				mcp.Required(),
			),
			mcp.WithString("package",
				mcp.Description("Go package name for generated code. Default: 'main'."),
			),
			mcp.WithString("var_prefix",
				mcp.Description("Prefix for generated variable names."),
			),
		),
		importHandler,
	)
	debugLog("Registered tool: import")

	// Validate tool - validate Kubernetes manifests
	s.AddTool(
		mcp.NewTool("validate",
			mcp.WithDescription("Validate Kubernetes manifests against the Kubernetes OpenAPI specification. Requires kubeconform to be installed."),
			mcp.WithString("yaml",
				mcp.Description("YAML content to validate."),
			),
			mcp.WithString("path",
				mcp.Description("Path to YAML file or directory to validate."),
			),
			mcp.WithBoolean("strict",
				mcp.Description("Enable strict validation (reject unknown fields)."),
			),
			mcp.WithString("kubernetes_version",
				mcp.Description("Kubernetes version to validate against (e.g., '1.29.0')."),
			),
		),
		validateHandler,
	)
	debugLog("Registered tool: validate")
}
