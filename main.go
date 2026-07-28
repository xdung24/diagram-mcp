// diagram-mcp runs an MCP server that generates, validates, and
// parses/describes diagrams, backed by the bpmn-mcp library,
// mermaid-mcp library, and drawio-mcp library.
// It can run over stdio (for use with MCP clients like VS Code) or
// over HTTP (for use with MCP clients like the MCP CLI).

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xdung24/diagram-mcp/internal/installer"
	"github.com/xdung24/diagram-mcp/internal/mcpserver"
	"github.com/xdung24/diagram-mcp/internal/render"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "-v", "--version":
			fmt.Printf("diagram-mcp_%s\n", version)
			return
		case "install":
			installer.RunInstall(os.Args[2:])
			return
		case "uninstall":
			installer.RunUninstall(os.Args[2:])
			return
		case "render":
			render.RenderDiagram(os.Args[2:])
			return
		}
	}

	var httpAddr string
	flag.StringVar(&httpAddr, "http", "", "serve MCP over streamable HTTP at this address (e.g. :8080) instead of stdio")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "diagram-mcp - MCP server for generating, validating, and parsing diagrams\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  diagram-mcp                     run as an MCP server (stdio)\n")
		fmt.Fprintf(os.Stderr, "  diagram-mcp -http :8080         run as an MCP server (HTTP)\n")
		fmt.Fprintf(os.Stderr, "  diagram-mcp install [flags]     install binary + register MCP client\n")
		fmt.Fprintf(os.Stderr, "  diagram-mcp uninstall [flags]   uninstall binary + unregister MCP client\n")
		fmt.Fprintf(os.Stderr, "  diagram-mcp render [flags]      render a diagram file\n")
		fmt.Fprintf(os.Stderr, "  diagram-mcp version\n\n")
	}

	flag.Parse()

	// Read the mcp type argument from the command line, if provided.
	// This allows the user to specify which MCP server to run (bpmn, mermaid, or drawio).
	mcpType := flag.Arg(0)
	mcpserver.StartMCPServer(mcpType, httpAddr, version)
}
