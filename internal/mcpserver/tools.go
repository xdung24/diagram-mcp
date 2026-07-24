package mcpserver

import (
	"context"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xdung24/mermaid-mcp/pkg/describe"
	"github.com/xdung24/mermaid-mcp/pkg/parser"
	"github.com/xdung24/mermaid-mcp/pkg/suggest"
	"github.com/xdung24/mermaid-mcp/pkg/utils"
	"github.com/xdung24/mermaid-mcp/pkg/utils/basediagram"
)

// applyPieDefaultTheme re-parses pie chart content and, unless it already
// declares an explicit theme, applies a vibrant default "base" theme (see
// pie.EnsureColorfulTheme) so generated pie charts are colorful by default.
// Returns the (possibly unchanged) content re-serialized via the diagram's
// own String() method.
func applyPieDefaultTheme(content string) (string, error) {
	d, err := parser.ParsePie(content)
	if err != nil {
		return "", err
	}
	d.EnsureColorfulTheme()
	return d.String(), nil
}

// applyGanttDefaults re-parses gantt chart content and, unless it already
// declares its own axisFormat/excludes, applies sensible defaults (see
// gantt.EnsureDefaults: axisFormat "%m-%d" and excludes weekends) so
// generated gantt charts render with a compact axis and skip weekends by
// default. Returns the (possibly unchanged) content re-serialized via the
// diagram's own String() method.
func applyGanttDefaults(content string) (string, error) {
	d, err := parser.ParseGantt(content)
	if err != nil {
		return "", err
	}
	d.EnsureDefaults()
	return d.String(), nil
}

// applyArchitectureTitleQuoting re-parses architecture-beta content and
// re-serializes it via the diagram's own String() method, which auto-quotes
// any group/service title containing characters outside Mermaid's unquoted
// ARCH_TITLE charset (`/[\w ]+/`) — e.g. hyphens or parentheses — via
// architecture.FormatTitle. This repo's own parser is lenient about raw
// bracket contents, but real Mermaid rejects such titles unless quoted, so
// callers authoring e.g. `[AI-Powered Pipeline]` or `[Worker Pool (Python)]`
// get working output without needing to know/apply the quoting rule
// themselves.
//
// Additionally, ensures the config block always includes the standard defaults:
// theme: default, maxTextSize: 50000, maxEdges: 500, fontSize: 16.
func applyArchitectureTitleQuoting(content string) (string, error) {
	d, err := parser.ParseArchitecture(content)
	if err != nil {
		return "", err
	}
	// Ensure standard config defaults are always present
	d.Config.SetMaxTextSize(50000)
	d.Config.SetMaxEdges(500)
	d.Config.SetFontSize(16)
	// Ensure theme is set to default (should already be, but be explicit)
	d.Config.Theme.Name = basediagram.ThemeDefault
	return d.String(), nil
}

// applyKanbanNormalization re-parses kanban content and re-serializes it via
// the diagram's own String() method, which always emits each column/task as
// an `id[Title]` node (see kanban.Column.String / kanban.Task.String) —
// matching the format real Mermaid expects. This repo's own parser
// tolerates bare, bracket-less, multi-word column/task titles (e.g. "In
// Progress"), but real Mermaid's kanban grammar requires bracketed node
// syntax for such titles, so callers authoring plain text get working,
// consistently-formatted output without needing to manually add ids/brackets
// themselves.
func applyKanbanNormalization(content string) (string, error) {
	d, err := parser.ParseKanban(content)
	if err != nil {
		return "", err
	}
	return d.String(), nil
}

// applyGitGraphNormalization re-parses gitGraph content and re-serializes it
// via the diagram's own String() method. This repo's own parser tolerates
// id:/type:/tag: commit attributes split onto their own line (merging them
// into the preceding commit/merge statement), but real Mermaid requires
// those attributes on the same line as the commit/merge keyword, so callers
// authoring the lenient form get working, spec-compliant output without
// needing to manually combine the lines themselves.
func applyGitGraphNormalization(content string) (string, error) {
	d, err := parser.ParseGitGraph(content)
	if err != nil {
		return "", err
	}
	return d.String(), nil
}

// applyPacketNormalization re-parses packet content and re-serializes it via
// the diagram's own String() method, which always builds field lines through
// the Diagram/Field model (NewDiagram + AddRange/AddBits, matching the
// programmatic builder flow in examples/packet_diagram) and always emits
// quoted labels. This repo's own parser tolerates unquoted field labels
// (e.g. `0-7: Version`), but real Mermaid requires labels to be quoted, so
// callers authoring the lenient form get working, spec-compliant output
// without needing to add quotes themselves.
func applyPacketNormalization(content string) (string, error) {
	d, err := parser.ParsePacket(content)
	if err != nil {
		return "", err
	}
	return d.String(), nil
}

// sourceInput is the common shape for tools that accept either inline
// Mermaid text or a path to an existing file, but not both.
type sourceInput struct {
	Content string `json:"content,omitempty" jsonschema:"inline Mermaid diagram text; mutually exclusive with path"`
	Path    string `json:"path,omitempty" jsonschema:"path to an existing Mermaid/markdown file to read; mutually exclusive with content"`
}

// resolveContent returns the Mermaid text to operate on: either in.Content
// directly, or the contents of the file at in.Path. Exactly one of the two
// must be set.
func resolveContent(in sourceInput) (string, error) {
	if in.Content != "" && in.Path != "" {
		return "", fmt.Errorf("provide either content or path, not both")
	}
	if in.Path != "" {
		data, err := os.ReadFile(in.Path)
		if err != nil {
			return "", fmt.Errorf("reading %q: %w", in.Path, err)
		}
		return string(data), nil
	}
	if in.Content == "" {
		return "", fmt.Errorf("provide either content or path")
	}
	return in.Content, nil
}

// GenerateDiagramResult is returned by generate_diagram.
type GenerateDiagramResult struct {
	DiagramType      string `json:"diagramType"`
	FilePath         string `json:"filePath"`
	SVGPath          string `json:"svgPath,omitempty"`
	PNGPath          string `json:"pngPath,omitempty"`
	RenderError      string `json:"renderError,omitempty"`
	ImageSizeWarning string `json:"imageSizeWarning,omitempty"`
	Message          string `json:"message"`
}

// renderImages best-effort renders content to SVG/PNG files alongside
// outputPath (same directory/basename, .svg/.png extensions) via
// utils.RenderImagesToFile. It never fails the caller: if the diagram type
// isn't supported by the image renderer, svgPath/pngPath are
// empty with no warning; if rendering unexpectedly fails for a supported
// type, svgPath/pngPath are empty and warning describes the failure — the
// primary Mermaid text file write is unaffected either way.
//
// When a PNG was written, its pixel dimensions are checked via
// utils.PNGDimensionsFromFile/utils.ImageSizeAdvice; sizeWarning is non-empty
// whenever the image is too large to safely/usefully view (see
// utils.ImageSizeAdvice for the exact thresholds/wording) — callers should
// surface this to the calling agent so it does not attempt to view/upload
// an oversized image or waste tokens reviewing one that's merely too large
// to be worth it.
func renderImages(outputPath, content string) (svgPath, pngPath, warning, sizeWarning string) {
	svgPath, pngPath, ok, err := utils.RenderImagesToFile(outputPath, content)
	if err != nil {
		return "", "", fmt.Sprintf("could not render svg/png preview: %v", err), ""
	}
	if !ok {
		return "", "", "", ""
	}
	if w, h, err := utils.PNGDimensionsFromFile(pngPath); err == nil {
		sizeWarning = utils.ImageSizeAdvice(w, h)
	}
	return svgPath, pngPath, "", sizeWarning
}

// ValidateDiagramResult is returned by validate_diagram.
type ValidateDiagramResult struct {
	Valid       bool   `json:"valid"`
	DiagramType string `json:"diagramType,omitempty"`
	Error       string `json:"error,omitempty"`
	FilePath    string `json:"filePath,omitempty"`
}

// ParseDiagramResult is returned by parse_diagram.
type ParseDiagramResult struct {
	DiagramType string           `json:"diagramType"`
	Summary     describe.Summary `json:"summary"`
	Narrative   string           `json:"narrative"`
	FilePath    string           `json:"filePath,omitempty"`
}

// SuggestDiagramTypeResult is returned by suggest_diagram_type.
type SuggestDiagramTypeResult = suggest.Result

func registerTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "generate_diagram",
		Description: "Validate Mermaid diagram text and write it to a file. Prefer a .mmd/.mermaid " +
			"outputPath (raw Mermaid syntax, no fence) unless the user specifically wants the diagram " +
			"embedded in a Markdown document; a .md/.markdown path instead wraps the content in a " +
			"```mermaid fence automatically. Fails with a clear error if the content does not parse as " +
			"one of the 23 supported diagram types. Also renders an SVG and a PNG preview image " +
			"alongside the written file (same directory/basename, .svg/.png extensions) using a pure-Go " +
			"renderer. If the PNG is unusually large, the result's imageSizeWarning field (and the trailing " +
			"part of message) tells you not to view/upload it (or to skip reviewing it) — see that field " +
			"before opening the PNG yourself.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in struct {
		Content    string `json:"content" jsonschema:"the Mermaid diagram text to write"`
		OutputPath string `json:"outputPath" jsonschema:"file path to write to; prefer 'diagram.mmd' (raw Mermaid, no fence) unless the user wants it embedded in a Markdown doc, in which case use 'diagram.md'"`
	}) (*mcp.CallToolResult, GenerateDiagramResult, error) {
		result, err := describe.Describe(in.Content)
		if err != nil {
			return nil, GenerateDiagramResult{}, fmt.Errorf("invalid Mermaid content: %w", err)
		}

		content := in.Content
		switch result.DiagramType {
		case parser.DiagramTypePie:
			themed, err := applyPieDefaultTheme(content)
			if err != nil {
				return nil, GenerateDiagramResult{}, fmt.Errorf("invalid Mermaid content: %w", err)
			}
			content = themed
		case parser.DiagramTypeGantt:
			defaulted, err := applyGanttDefaults(content)
			if err != nil {
				return nil, GenerateDiagramResult{}, fmt.Errorf("invalid Mermaid content: %w", err)
			}
			content = defaulted
		case parser.DiagramTypeArchitecture:
			quoted, err := applyArchitectureTitleQuoting(content)
			if err != nil {
				return nil, GenerateDiagramResult{}, fmt.Errorf("invalid Mermaid content: %w", err)
			}
			content = quoted
		case parser.DiagramTypeKanban:
			normalized, err := applyKanbanNormalization(content)
			if err != nil {
				return nil, GenerateDiagramResult{}, fmt.Errorf("invalid Mermaid content: %w", err)
			}
			content = normalized
		case parser.DiagramTypeGitGraph:
			normalized, err := applyGitGraphNormalization(content)
			if err != nil {
				return nil, GenerateDiagramResult{}, fmt.Errorf("invalid Mermaid content: %w", err)
			}
			content = normalized
		case parser.DiagramTypePacket:
			normalized, err := applyPacketNormalization(content)
			if err != nil {
				return nil, GenerateDiagramResult{}, fmt.Errorf("invalid Mermaid content: %w", err)
			}
			content = normalized
		}

		if err := utils.RenderToFile(in.OutputPath, content); err != nil {
			return nil, GenerateDiagramResult{}, fmt.Errorf("writing %q: %w", in.OutputPath, err)
		}

		svgPath, pngPath, warning, sizeWarning := renderImages(in.OutputPath, content)
		msg := fmt.Sprintf("wrote %s diagram to %s", result.DiagramType, in.OutputPath)
		switch {
		case svgPath != "":
			msg += fmt.Sprintf("; rendered SVG to %s and PNG to %s", svgPath, pngPath)
		case warning != "":
			msg += "; " + warning
		}
		if sizeWarning != "" {
			msg += "; " + sizeWarning
		}

		return nil, GenerateDiagramResult{
			DiagramType:      string(result.DiagramType),
			FilePath:         in.OutputPath,
			SVGPath:          svgPath,
			PNGPath:          pngPath,
			RenderError:      warning,
			ImageSizeWarning: sizeWarning,
			Message:          msg,
		}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "validate_diagram",
		Description: "Check whether Mermaid diagram text (inline, or read from a file) is " +
			"syntactically valid, and if so, which of the 23 supported diagram types it is. Returns " +
			"a clear error message when parsing fails instead of throwing.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in sourceInput) (*mcp.CallToolResult, ValidateDiagramResult, error) {
		content, err := resolveContent(in)
		if err != nil {
			return nil, ValidateDiagramResult{}, err
		}

		result, err := describe.Describe(content)
		if err != nil {
			return nil, ValidateDiagramResult{
				Valid:    false,
				Error:    err.Error(),
				FilePath: in.Path,
			}, nil
		}

		return nil, ValidateDiagramResult{
			Valid:       true,
			DiagramType: string(result.DiagramType),
			FilePath:    in.Path,
		}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "parse_diagram",
		Description: "Parse Mermaid diagram text (inline, or read from a file) and return a " +
			"structured summary (element counts) plus a Markdown narrative explaining what the " +
			"diagram contains and how its pieces relate to each other. Use this to understand the " +
			"type and meaning of a diagram you didn't write yourself.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in sourceInput) (*mcp.CallToolResult, ParseDiagramResult, error) {
		content, err := resolveContent(in)
		if err != nil {
			return nil, ParseDiagramResult{}, err
		}

		result, err := describe.Describe(content)
		if err != nil {
			return nil, ParseDiagramResult{}, err
		}

		return nil, ParseDiagramResult{
			DiagramType: string(result.DiagramType),
			Summary:     result.Summary,
			Narrative:   result.Narrative,
			FilePath:    in.Path,
		}, nil
	})

	registerPerTypeGenerateTools(s)

	mcp.AddTool(s, &mcp.Tool{
		Name: "suggest_diagram_type",
		Description: "Given a plain-language description of what the user wants to diagram " +
			"(NOT Mermaid syntax), suggest which of the 23 supported diagram types (flowchart, " +
			"classDiagram, sequenceDiagram, stateDiagram, erDiagram, timeline, journey, block, gantt, " +
			"pie, quadrantChart, requirementDiagram, architecture, kanban, mindmap, packet, xychart, " +
			"treemap, sankey, venn, gitGraph, c4, cynefin) best " +
			"fits, based on keywords/context in the description. Use this before composing Mermaid " +
			"text when the user hasn't said which diagram type they want. If the result has " +
			"confident:true, proceed using diagramType. If confident:false, ask the user the " +
			"returned clarifyingQuestion instead of guessing.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in struct {
		Requirement string `json:"requirement" jsonschema:"plain-language description of what the user wants to diagram, e.g. the user's original request text"`
	}) (*mcp.CallToolResult, SuggestDiagramTypeResult, error) {
		if in.Requirement == "" {
			return nil, SuggestDiagramTypeResult{}, fmt.Errorf("provide a requirement description")
		}
		return nil, suggest.Suggest(in.Requirement), nil
	})
}

// registerPerTypeGenerateTools registers one generate_<type>_diagram tool per
// supported diagram type (see diagramTypes in diagram_types.go). Each behaves
// like generate_diagram but additionally rejects content that doesn't match
// its specific diagram type, and its description explains when/how to use
// that particular diagram type — so a caller can pick the right tool
// directly instead of always going through the generic generate_diagram.
func registerPerTypeGenerateTools(s *mcp.Server) {
	for _, info := range diagramTypes {
		mcp.AddTool(s, &mcp.Tool{
			Name: info.ToolName,
			Description: fmt.Sprintf(
				"Validate Mermaid %s diagram text and write it to a file. %s Example:\n%s\n\n"+
					"Prefer a .mmd/.mermaid outputPath (raw Mermaid syntax, no fence) unless the user "+
					"specifically wants the diagram embedded in a Markdown document; a .md/.markdown path "+
					"instead wraps the content in a ```mermaid fence automatically. Fails with a clear error "+
					"if the content is not valid %s syntax. Also renders an SVG and a PNG preview image "+
					"alongside the written file (same directory/basename, .svg/.png extensions) using a "+
					"pure-Go renderer. If the PNG is unusually large, the result's imageSizeWarning field "+
					"(and the trailing part of message) tells you not to view/upload it (or to skip reviewing "+
					"it) — see that field before opening the PNG yourself.",
				info.DiagramType, info.WhenToUse, info.Example, info.DiagramType,
			),
		}, func(_ context.Context, _ *mcp.CallToolRequest, in struct {
			Content    string `json:"content" jsonschema:"the Mermaid diagram text to write"`
			OutputPath string `json:"outputPath" jsonschema:"file path to write to; prefer 'diagram.mmd' (raw Mermaid, no fence) unless the user wants it embedded in a Markdown doc, in which case use 'diagram.md'"`
		}) (*mcp.CallToolResult, GenerateDiagramResult, error) {
			result, err := describe.Describe(in.Content)
			if err != nil {
				return nil, GenerateDiagramResult{}, fmt.Errorf("invalid Mermaid content: %w", err)
			}
			if result.DiagramType != info.DiagramType {
				return nil, GenerateDiagramResult{}, fmt.Errorf(
					"content is a %s diagram, not %s; use the matching generate_<type>_diagram tool "+
						"or the generic generate_diagram tool instead",
					result.DiagramType, info.DiagramType)
			}

			content := in.Content
			switch info.DiagramType {
			case parser.DiagramTypePie:
				themed, err := applyPieDefaultTheme(content)
				if err != nil {
					return nil, GenerateDiagramResult{}, fmt.Errorf("invalid Mermaid content: %w", err)
				}
				content = themed
			case parser.DiagramTypeGantt:
				defaulted, err := applyGanttDefaults(content)
				if err != nil {
					return nil, GenerateDiagramResult{}, fmt.Errorf("invalid Mermaid content: %w", err)
				}
				content = defaulted
			case parser.DiagramTypeArchitecture:
				quoted, err := applyArchitectureTitleQuoting(content)
				if err != nil {
					return nil, GenerateDiagramResult{}, fmt.Errorf("invalid Mermaid content: %w", err)
				}
				content = quoted
			case parser.DiagramTypeKanban:
				normalized, err := applyKanbanNormalization(content)
				if err != nil {
					return nil, GenerateDiagramResult{}, fmt.Errorf("invalid Mermaid content: %w", err)
				}
				content = normalized
			case parser.DiagramTypeGitGraph:
				normalized, err := applyGitGraphNormalization(content)
				if err != nil {
					return nil, GenerateDiagramResult{}, fmt.Errorf("invalid Mermaid content: %w", err)
				}
				content = normalized
			case parser.DiagramTypePacket:
				normalized, err := applyPacketNormalization(content)
				if err != nil {
					return nil, GenerateDiagramResult{}, fmt.Errorf("invalid Mermaid content: %w", err)
				}
				content = normalized
			}

			if err := utils.RenderToFile(in.OutputPath, content); err != nil {
				return nil, GenerateDiagramResult{}, fmt.Errorf("writing %q: %w", in.OutputPath, err)
			}

			svgPath, pngPath, warning, sizeWarning := renderImages(in.OutputPath, content)
			msg := fmt.Sprintf("wrote %s diagram to %s", result.DiagramType, in.OutputPath)
			switch {
			case svgPath != "":
				msg += fmt.Sprintf("; rendered SVG to %s and PNG to %s", svgPath, pngPath)
			case warning != "":
				msg += "; " + warning
			}
			if sizeWarning != "" {
				msg += "; " + sizeWarning
			}

			return nil, GenerateDiagramResult{
				DiagramType:      string(result.DiagramType),
				FilePath:         in.OutputPath,
				SVGPath:          svgPath,
				PNGPath:          pngPath,
				RenderError:      warning,
				ImageSizeWarning: sizeWarning,
				Message:          msg,
			}, nil
		})
	}
}
