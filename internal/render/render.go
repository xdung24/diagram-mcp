package render

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	bpmn "github.com/xdung24/bpmn-mcp/pkg/renderer"
	drawio "github.com/xdung24/drawio-mcp/pkg/render"
	mermaid "github.com/xdung24/mermaid-mcp/pkg/render"
)

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
	if selectedFormat != "svg" && selectedFormat != "png" && selectedFormat != "jpg" && selectedFormat != "json" {
		log.Fatalf("imagerender: invalid -f value %q (allowed: svg|png|jpg|json)", *format)
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

		if selectedFormat == "json" {
			jsonData, err := mermaid.RenderDataJSON(string(fileContent))
			if err != nil {
				log.Fatalf("mermaid-to-image render: failed to render JSON: %v", err)
			}
			if err := os.WriteFile(targetPath, jsonData, 0644); err != nil {
				log.Fatalf("mermaid-to-image render: failed to write json file: %v", err)
			}
		} else {
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
