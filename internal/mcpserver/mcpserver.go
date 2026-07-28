package mcpserver

import (
	"context"
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	bpmn "github.com/xdung24/bpmn-mcp/pkg/mcpserver"
	drawio "github.com/xdung24/drawio-mcp/pkg/mcpserver"
	mermaid "github.com/xdung24/mermaid-mcp/pkg/mcpserver"
)

func StartMCPServer(mcpType string, httpAddr string, version string) {
	var server *mcp.Server
	switch mcpType {
	case "bpmn", "bpmn-mcp":
		server = bpmn.NewServer(version)
	case "mermaid", "mermaid-mcp":
		server = mermaid.NewServer(version)
	case "drawio", "drawio-mcp":
		server = drawio.NewServer(version)
	default:
		server = mermaid.NewServer(version)
	}

	if httpAddr != "" {
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
