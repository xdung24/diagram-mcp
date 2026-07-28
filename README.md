# diagram-tests

A test/scratch repository for exercising three MCP (Model Context Protocol)
diagram servers end-to-end from plain-text/Markdown process descriptions.

## Online demo

You can view online demo [Here](https://diagram-mcp.netlify.app/) 

## Purpose

This repo has no application code of its own — it exists purely to drive and
validate the following MCP servers against real-world sample inputs:

### 1. `bpmn-mcp`

Generates BPMN 2.0 diagrams from process descriptions and exports them to SVG
and PNG.

- Supported diagram types:
  - **Business Process Diagram (BPD)** — a single-pool process graph (tasks,
    gateways, events, sequence flows).
  - **Collaboration Diagram** — multiple pools/lanes with message flows
    between participants.
- Input samples live in [bpmn/](bpmn/) as plain-text process descriptions
  (e.g. [bpmn/1-dispatch-of-goods.txt](bpmn/1-dispatch-of-goods.txt),
  [bpmn/9-insurance-claim-handling.txt](bpmn/9-insurance-claim-handling.txt)).
- Workflow: `start_session` → build the model (pools/lanes/tasks/gateways/
  events/connections) → `review_outline` → `review_structure` → `run_layout`
  → `review_layout` → `finish` (writes the `.bpmn` file), then `render_image`
  to export SVG/PNG.

### 2. `mermaid-mcp`

Generates Mermaid diagram files from descriptions and exports them to SVG and
PNG. Currently supports all **21** diagram types: flowchart, class,
sequence, state, entity relationship, timeline, user journey, block,
gantt, pie, quadrant, requirement, architecture, mindmap, C4, kanban,
gitGraph, packet, treemap, xy chart, and sankey diagrams.

- Input samples live in [mermaid/](mermaid/), one file per diagram type
  (e.g. [mermaid/01-flowchart-morning-routine.txt](mermaid/01-flowchart-morning-routine.txt),
  [mermaid/21-sankey-diagram-annual-revenue-flow.txt](mermaid/21-sankey-diagram-annual-revenue-flow.txt)).
- Workflow: compose Mermaid syntax from the source description, then call the  matching `generate_<type>_diagram` tool with an `outputPath` next to the  source file — this validates the syntax, writes the `.mmd`/`.md` file, and renders SVG/PNG previews alongside it.

### 3. `drawio-mcp`

Generates draw.io diagrams and exports them to SVG and PNG. Unlike the two servers above, it has no fixed set of supported diagram types — it can generate a draw.io diagram from any plain-text description. For this repo, testing is scoped to two specific conversion paths:

- **BPMN → draw.io**: generate a draw.io diagram from an existing `.bpmn` file (see [bpmn/](bpmn/) outputs).
- **Mermaid → draw.io**: generate a draw.io diagram from an existing Mermaid file (see [mermaid/](mermaid/) outputs).
- Editing existing drawio file from plain text instruction.

Generated draw.io files and their SVG/PNG exports are stored in [drawio/](drawio/).

## Repository layout

```
bpmn/       Plain-text process descriptions used as input for bpmn-mcp
mermaid/    Plain-text diagram descriptions used as input for mermaid-mcp,
            one file per supported diagram type
drawio/     Output draw.io diagrams (converted from bpmn/ and mermaid/
            outputs) and their SVG/PNG exports
```
