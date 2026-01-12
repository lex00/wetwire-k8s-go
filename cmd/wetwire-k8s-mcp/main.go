// Package main implements an MCP (Model Context Protocol) server for wetwire-k8s-go.
// This server enables Claude Code and other MCP clients to interact with wetwire-k8s
// tools for Kubernetes manifest generation and management.
package main

import (
	"context"
	"log"
	"os"

	"github.com/lex00/wetwire-core-go/mcp"
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

	// Create MCP server using wetwire-core-go infrastructure
	server := mcp.NewServer(mcp.Config{
		Name:    "wetwire-k8s",
		Version: Version,
		Debug:   os.Getenv("WETWIRE_MCP_DEBUG") != "",
	})

	// Register standard wetwire tools using core infrastructure
	mcp.RegisterStandardTools(server, "k8s", mcp.StandardToolHandlers{
		Init:     nil, // Not yet implemented
		Build:    buildToolHandler,
		Lint:     lintToolHandler,
		Validate: validateToolHandler,
		Import:   importToolHandler,
		List:     nil, // Not yet implemented
		Graph:    nil, // Not yet implemented
	})

	debugLog("Registered all tools, starting stdio server")

	// Start stdio server
	if err := server.Start(context.Background()); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}
