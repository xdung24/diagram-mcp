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
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	bpmnmcp "github.com/xdung24/bpmn-mcp/pkg/mcpserver"
	bpmn "github.com/xdung24/bpmn-mcp/pkg/renderer"
	installer "github.com/xdung24/diagram-mcp/internal/installer"
	drawiomcp "github.com/xdung24/drawio-mcp/pkg/mcpserver"
	drawio "github.com/xdung24/drawio-mcp/pkg/render"
	mermaidmcp "github.com/xdung24/mermaid-mcp/pkg/mcpserver"
	mermaid "github.com/xdung24/mermaid-mcp/pkg/render"
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
			RenderDiagram(os.Args[2:])
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
	startMCPServer(mcpType, httpAddr)
}

func RenderDiagram(args []string) {
	// Read the command-line flags for input file, output format, and output file path
	fs := flag.NewFlagSet("render", flag.ExitOnError)
	input := fs.String("i", "", "path to the input diagram file(bpmn, dmn, mermaid, mmd, drawio, xml, json)")
	format := fs.String("f", "svg", "specify the output format (png, svg, jpeg, jpg)")
	output := fs.String("o", "", "path to the output file")
	if err := fs.Parse(args); err != nil {
		log.Fatalf("imagerender: %v", err)
	}
	inputPath := strings.ToLower(strings.TrimSpace(*input))
	if inputPath == "" {
		log.Fatal("imagerender: -i is required")
	}
	selectedFormat := strings.ToLower(strings.TrimSpace(*format))
	if selectedFormat != "svg" && selectedFormat != "png" && selectedFormat != "jpg" {
		log.Fatalf("imagerender: invalid -f value %q (allowed: svg|png|jpg)", *format)
	}
	baseName := strings.TrimSuffix(inputPath, filepath.Ext(inputPath))
	targetPath := baseName + "." + selectedFormat
	outputPath := strings.ToLower(strings.TrimSpace(*output))
	if outputPath != "" {
		targetPath = outputPath
		if filepath.Ext(targetPath) == "" {
			targetPath += "." + selectedFormat
		}
	}
	dir := filepath.Dir(targetPath)
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("imagerender: failed to create directory: %v", err)
		}
	}

	// Call the render function from the appropriate package based on the input file extension
	switch filepath.Ext(inputPath) {
	case ".bpmn":
		fileContent, err := os.ReadFile(inputPath)
		if err != nil {
			log.Fatalf("bpmn-to-image render: failed to read input file: %v", err)
		}
		switch selectedFormat {
		case "svg":
			img, err := bpmn.BpmnRenderSvg(string(fileContent), 30)
			if err != nil {
				log.Fatalf("bpmn-to-image render: failed to render image: %v", err)
			}
			if err := os.WriteFile(targetPath, img, 0644); err != nil {
				log.Fatalf("bpmn-to-image render: failed to write %s file: %v", selectedFormat, err)
			}
		case "png":
			img, err := bpmn.BpmnRenderPng(string(fileContent), 30)
			if err != nil {
				log.Fatalf("bpmn-to-image render: failed to render image: %v", err)
			}
			if err := os.WriteFile(targetPath, img, 0644); err != nil {
				log.Fatalf("bpmn-to-image render: failed to write %s file: %v", selectedFormat, err)
			}
		}
	case ".dmn":
		fileContent, err := os.ReadFile(inputPath)
		if err != nil {
			log.Fatalf("dmn-to-image render: failed to read input file: %v", err)
		}
		switch selectedFormat {
		case "svg":
			img, err := bpmn.DmnRenderSvg(string(fileContent), 30)
			if err != nil {
				log.Fatalf("dmn-to-image render: failed to render image: %v", err)
			}
			if err := os.WriteFile(targetPath, img, 0644); err != nil {
				log.Fatalf("dmn-to-image render: failed to write %s file: %v", selectedFormat, err)
			}
		case "png":
			img, err := bpmn.DmnRenderPng(string(fileContent), 30)
			if err != nil {
				log.Fatalf("dmn-to-image render: failed to render image: %v", err)
			}
			if err := os.WriteFile(targetPath, img, 0644); err != nil {
				log.Fatalf("dmn-to-image render: failed to write %s file: %v", selectedFormat, err)
			}
		}
	case ".mmd", ".mermaid":
		fileContent, err := os.ReadFile(inputPath)
		if err != nil {
			log.Fatalf("mermaid-to-image render: failed to read input file: %v", err)
		}
		svg, png, err := mermaid.RenderImages(string(fileContent))
		if err != nil {
			log.Fatalf("mermaid-to-image render: failed to render images: %v", err)
		}

		switch selectedFormat {
		case "svg":
			if err := os.WriteFile(targetPath, svg, 0644); err != nil {
				log.Fatalf("mermaid-to-image render: failed to write svg file: %v", err)
			}
		case "png":
			if err := os.WriteFile(targetPath, png, 0644); err != nil {
				log.Fatalf("mermaid-to-image render: failed to write png file: %v", err)
			}
		}
	case ".drawio":
		fileContent, err := os.ReadFile(inputPath)
		if err != nil {
			log.Fatalf("drawio-to-image render: failed to read input file: %v", err)
		}
		switch selectedFormat {
		case "svg":
			img, err := drawio.RenderSvg(string(fileContent))
			if err != nil {
				log.Fatalf("drawio-to-image render: failed to render SVG: %v", err)
			}
			if err := os.WriteFile(targetPath, img, 0644); err != nil {
				log.Fatalf("drawio-to-image render: failed to write SVG file: %v", err)
			}
		case "png":
			img, err := drawio.RenderPng(string(fileContent))
			if err != nil {
				log.Fatalf("drawio-to-image render: failed to render PNG: %v", err)
			}
			if err := os.WriteFile(targetPath, img, 0644); err != nil {
				log.Fatalf("drawio-to-image render: failed to write PNG file: %v", err)
			}
		}
	default:
		log.Fatalf("unsupported input file format %q", filepath.Ext(inputPath))
	}
}

func startMCPServer(mcpType string, httpAddr string) {
	var server *mcp.Server
	switch mcpType {
	case "bpmn-mcp":
		server = bpmnmcp.NewServer(version)
	case "mermaid-mcp":
		server = mermaidmcp.NewServer(version)
	case "drawio-mcp":
		server = drawiomcp.NewServer(version)
	default:
		server = mermaidmcp.NewServer(version)
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
