// Package mcpserver exposes the diagram-mcp tool surface.
package mcpserver

import (
	_ "embed"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed instructions.md
var instructions string

// NewServer builds the diagram-mcp MCP server with every tool registered.
func NewServer(version string) *mcp.Server {
	opts := &mcp.ServerOptions{
		Instructions: instructions,
		Capabilities: &mcp.ServerCapabilities{
			Tools: &mcp.ToolCapabilities{
				ListChanged: true,
			},
		},
	}
	s := mcp.NewServer(&mcp.Implementation{Name: "diagram-mcp", Title: "diagram-mcp", Version: version}, opts)
	registerTools(s)
	return s
}
