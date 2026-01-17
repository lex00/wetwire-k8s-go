// MCP server implementation for embedded design mode.
//
// When mcp subcommand is called, this runs the MCP protocol over stdio,
// providing wetwire_build, wetwire_lint, wetwire_validate, wetwire_list, wetwire_graph, and wetwire_init tools.
package main

import (
	"context"

	"github.com/lex00/wetwire-core-go/domain"
	k8sdomain "github.com/lex00/wetwire-k8s-go/domain"
	"github.com/spf13/cobra"
)

// newMCPCmd creates the mcp subcommand.
func newMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server for Claude Code integration",
		Long: `Start an MCP (Model Context Protocol) server for integration with Claude Code.

This command runs an MCP server over stdio, providing tools for:
  - wetwire_build: Generate Kubernetes manifests from Go code
  - wetwire_lint: Lint Go code for wetwire-k8s patterns
  - wetwire_validate: Validate manifests using kubeconform
  - wetwire_list: List discovered Kubernetes resources
  - wetwire_graph: Generate dependency graphs
  - wetwire_init: Initialize new wetwire-k8s projects

This is typically called by Claude Code or other MCP clients, not directly by users.`,
		RunE: runMCPServer,
	}

	return cmd
}

// runMCPServer starts the MCP server on stdio transport.
func runMCPServer(cmd *cobra.Command, args []string) error {
	// Create K8s domain instance
	k8sDomain := &k8sdomain.K8sDomain{}

	// Build MCP server using auto-generation from domain
	server := domain.BuildMCPServer(k8sDomain)

	// Start stdio server
	return server.Start(context.Background())
}
