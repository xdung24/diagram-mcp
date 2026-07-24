package mcpserver

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// connectTestClient wires an in-memory client to a freshly built server and
// returns a session ready for CallTool. The session is closed automatically
// when the test ends.
func connectTestClient(t *testing.T) *mcp.ClientSession {
	t.Helper()

	server := NewServer("test")
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	ctx := context.Background()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect() error = %v", err)
	}
	t.Cleanup(func() { serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect() error = %v", err)
	}
	t.Cleanup(func() { clientSession.Close() })

	return clientSession
}

// callTool invokes a tool and decodes its structured content into out.
func callTool(t *testing.T, session *mcp.ClientSession, name string, args any, out any) *mcp.CallToolResult {
	t.Helper()

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("CallTool(%s) error = %v", name, err)
	}

	if out != nil && res.StructuredContent != nil {
		data, err := json.Marshal(res.StructuredContent)
		if err != nil {
			t.Fatalf("marshal structured content: %v", err)
		}
		if err := json.Unmarshal(data, out); err != nil {
			t.Fatalf("unmarshal structured content: %v", err)
		}
	}

	return res
}

func TestMCPTools_GenerateValidateParse(t *testing.T) {
	session := connectTestClient(t)

	const flowchartText = "flowchart TB\n    A --> B\n"

	t.Run("generate_diagram writes a .md file with a mermaid fence", func(t *testing.T) {
		outPath := filepath.Join(t.TempDir(), "diagram.md")

		var genResult GenerateDiagramResult
		res := callTool(t, session, "generate_diagram", map[string]any{
			"content":    flowchartText,
			"outputPath": outPath,
		}, &genResult)
		if res.IsError {
			t.Fatalf("generate_diagram returned an error result: %+v", res)
		}
		if genResult.DiagramType != "flowchart" {
			t.Errorf("DiagramType = %q, want flowchart", genResult.DiagramType)
		}

		writtenBytes, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("reading generated file: %v", err)
		}
		written := string(writtenBytes)
		if !strings.HasPrefix(written, "```mermaid\n") || !strings.HasSuffix(written, "\n```\n") {
			t.Errorf("expected .md output to be fenced, got:\n%s", written)
		}

		if genResult.SVGPath == "" || genResult.PNGPath == "" {
			t.Fatalf("expected svgPath/pngPath to be populated for a supported diagram type, got %+v", genResult)
		}
		if _, err := os.Stat(genResult.SVGPath); err != nil {
			t.Errorf("svg file not written: %v", err)
		}
		if _, err := os.Stat(genResult.PNGPath); err != nil {
			t.Errorf("png file not written: %v", err)
		}
	})

	t.Run("generate_diagram rejects invalid content", func(t *testing.T) {
		outPath := filepath.Join(t.TempDir(), "diagram.md")

		res := callTool(t, session, "generate_diagram", map[string]any{
			"content":    "this is not mermaid",
			"outputPath": outPath,
		}, nil)
		if !res.IsError {
			t.Error("expected generate_diagram to report an error for invalid content")
		}
	})

	t.Run("validate_diagram reports valid content and type", func(t *testing.T) {
		var validateResult ValidateDiagramResult
		res := callTool(t, session, "validate_diagram", map[string]any{
			"content": flowchartText,
		}, &validateResult)
		if res.IsError {
			t.Fatalf("validate_diagram returned an error result: %+v", res)
		}
		if !validateResult.Valid || validateResult.DiagramType != "flowchart" {
			t.Errorf("validateResult = %+v", validateResult)
		}
	})

	t.Run("validate_diagram reports invalid content without erroring", func(t *testing.T) {
		var validateResult ValidateDiagramResult
		res := callTool(t, session, "validate_diagram", map[string]any{
			"content": "definitely not mermaid syntax",
		}, &validateResult)
		if res.IsError {
			t.Fatalf("validate_diagram should report invalid content in the result, not as a tool error: %+v", res)
		}
		if validateResult.Valid || validateResult.Error == "" {
			t.Errorf("validateResult = %+v, want Valid=false with an Error message", validateResult)
		}
	})

	t.Run("parse_diagram returns a summary and narrative", func(t *testing.T) {
		var parseResult ParseDiagramResult
		res := callTool(t, session, "parse_diagram", map[string]any{
			"content": flowchartText,
		}, &parseResult)
		if res.IsError {
			t.Fatalf("parse_diagram returned an error result: %+v", res)
		}
		if parseResult.DiagramType != "flowchart" {
			t.Errorf("DiagramType = %q, want flowchart", parseResult.DiagramType)
		}
		if !strings.Contains(parseResult.Narrative, "Flowchart") {
			t.Errorf("Narrative missing expected content: %s", parseResult.Narrative)
		}
		if parseResult.Summary.Counts["nodes"] != 2 {
			t.Errorf("Summary.Counts[nodes] = %d, want 2", parseResult.Summary.Counts["nodes"])
		}
	})

	t.Run("parse_diagram from a file path", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "diagram.mermaid")
		if err := os.WriteFile(filePath, []byte(flowchartText), 0644); err != nil {
			t.Fatalf("writing test file: %v", err)
		}

		var parseResult ParseDiagramResult
		res := callTool(t, session, "parse_diagram", map[string]any{
			"path": filePath,
		}, &parseResult)
		if res.IsError {
			t.Fatalf("parse_diagram returned an error result: %+v", res)
		}
		if parseResult.DiagramType != "flowchart" {
			t.Errorf("DiagramType = %q, want flowchart", parseResult.DiagramType)
		}
	})
}

func TestMCPTools_SuggestDiagramType(t *testing.T) {
	session := connectTestClient(t)

	t.Run("confidently suggests a diagram type", func(t *testing.T) {
		var result SuggestDiagramTypeResult
		res := callTool(t, session, "suggest_diagram_type", map[string]any{
			"requirement": "Create a flowchart for the steps to make coffee, with a decision for whether I'm tired",
		}, &result)
		if res.IsError {
			t.Fatalf("suggest_diagram_type returned an error result: %+v", res)
		}
		if !result.Confident || result.DiagramType != "flowchart" {
			t.Errorf("result = %+v, want confident flowchart", result)
		}
	})

	t.Run("asks for clarification when ambiguous", func(t *testing.T) {
		var result SuggestDiagramTypeResult
		res := callTool(t, session, "suggest_diagram_type", map[string]any{
			"requirement": "draw a picture of my day",
		}, &result)
		if res.IsError {
			t.Fatalf("suggest_diagram_type returned an error result: %+v", res)
		}
		if result.Confident || result.ClarifyingQuestion == "" {
			t.Errorf("result = %+v, want unconfident with a clarifying question", result)
		}
	})

	t.Run("rejects empty requirement", func(t *testing.T) {
		res := callTool(t, session, "suggest_diagram_type", map[string]any{
			"requirement": "",
		}, nil)
		if !res.IsError {
			t.Error("expected suggest_diagram_type to report an error for an empty requirement")
		}
	})
}
