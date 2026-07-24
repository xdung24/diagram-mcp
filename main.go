// diagram-mcp runs an MCP server that generates, validates, and
// parses/describes diagrams, backed by the bpmn-mcp library,
// mermaid-mcp library, and drawio-mcp library.
// It can run over stdio (for use with MCP clients like VS Code) or
// over HTTP (for use with MCP clients like the MCP CLI).

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xdung24/diagram-mcp/internal/installer"
	"github.com/xdung24/diagram-mcp/internal/mcpserver"
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
		fmt.Fprintf(os.Stderr, "  diagram-mcp version\n\n")
	}

	flag.Parse()

	server := mcpserver.NewServer(version)

	if httpAddr != "" {
		// JSONResponse avoids the text/event-stream (SSE) response mode: some
		// MCP clients' fetch implementations (e.g. VS Code's, depending on the
		// Node/Electron version) don't fully support streamed/duplex fetch and
		// fail with errors like "fetch failed: not implemented... yet...".
		// Plain JSON responses sidestep that, at the cost of no server-push
		// notifications, which this server doesn't use anyway.
		opts := &mcp.StreamableHTTPOptions{JSONResponse: true}
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, opts)
		log.Printf("diagram-mcp: serving MCP over HTTP at %s", httpAddr)
		if err := http.ListenAndServe(httpAddr, handler); err != nil {
			log.Fatalf("diagram-mcp http server error: %v", err)
		}
		return
	}

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("diagram-mcp server error: %v", err)
	}
}
