# diagram MCP Server

This server helps you generate, validate, and understand diagram diagrams. It
supports 23 diagram types: flowchart, classDiagram, sequenceDiagram,
stateDiagram(-v2), erDiagram, timeline, journey (user journey), block-beta,
gantt, pie, quadrantChart, requirementDiagram, architecture-beta, kanban,
mindmap, packet, xychart, treemap-beta, sankey-beta, venn-beta, gitGraph, C4
(C4Context/C4Container/C4Component/C4Dynamic/C4Deployment), and cynefin-beta.

You already know diagram syntax — write the diagram text yourself, then use
these tools to validate it and save it to disk, or to inspect/understand an
existing diagram.

## Tools

- **suggest_diagram_type** — Given a plain-language description of what the
  user wants to diagram (not diagram syntax), suggests which of the 23
  supported diagram types best fits. Use this first whenever the user hasn't
  said which diagram type they want. If the result has `confident: true`,
  proceed using its `diagramType`. If `confident: false`, ask the user the
  returned `clarifyingQuestion` instead of guessing.
- **generate_<type>_diagram** (one per diagram type, e.g.
  `generate_flowchart_diagram`, `generate_sequenceDiagram_diagram`,
  `generate_c4_diagram`) — Prefer these once you know the diagram type: each
  tool's own description explains exactly when/how to use that diagram type
  (with an example snippet) and rejects content that doesn't match its type.
  Use this instead of the generic `generate_diagram` whenever the target type
  is already known/decided, so the tool choice itself documents intent.
- **generate_diagram** — Generic fallback: validate diagram diagram text and
  write it to a file, auto-detecting which of the 23 types it is. Prefer a
  ".mmd"/".diagram" outputPath (raw diagram syntax, no fence) unless the user
  specifically wants the diagram embedded in a Markdown document, in which
  case use a ".md"/".markdown" path — that wraps the content in a
  ```diagram fence automatically. Content may itself already include a
  ```diagram fence and/or a "---" front-matter block (title/config) — both
  are handled correctly either way. Also renders an SVG and a PNG preview
  image alongside the written file (same directory/basename, .svg/.png
  extensions) via a pure-Go renderer.
- **validate_diagram** — Check whether diagram text (or an existing file) is
  syntactically valid and, if so, which of the 23 diagram types it is. Returns
  a clear error message when parsing fails.
- **parse_diagram** — Parse diagram text (or an existing file) and return a
  structured summary (element counts) plus a Markdown narrative explaining
  what the diagram contains and how its pieces relate — use this to
  understand the type and meaning of a diagram you didn't write yourself.

Each of validate_diagram and parse_diagram accepts either inline "content" or
a "path" to an existing file on disk (exactly one of the two); they work
across all 23 types automatically since they auto-detect the type.

## If a Tool You Expect Isn't in Your Available Tools List

This document describes every tool this server can offer, but some chat
hosts don't always send you every tool directly — especially when many MCP
servers/extensions are active at once, a host may show only a subset and
expose the rest only through some OTHER tool that is already sitting in your
own current tool list (for example, one whose own description says calling
it grants access to more tools). This document cannot know whether your
particular host does this, so follow this rule instead of guessing:

1. If the specific `generate_<type>_diagram` tool you want isn't visible,
   check whether the generic **generate_diagram** tool is visible instead —
   it covers all 23 types and behaves the same way.
2. If that's ALSO not visible, look through your OWN current tool list
   (not this document) for any other tool whose name/description indicates
   it can grant access to more tools, and call that one.
3. Only conclude this MCP server is disabled/unavailable if NONE of the
   above helps AND literally no diagram-mcp tool at all (not
   `suggest_diagram_type`, not `validate_diagram`, not `parse_diagram`, not
   any `generate_*` tool) appears anywhere in your tool list. If even one
   diagram-mcp tool is visible, the server is active — say so, and use
   whatever diagram-mcp tool IS available rather than telling the user the
   server is disabled.
4. Never call, mention, or invent a tool name that is not literally present
   in your own tool list — only act on tools you can actually see.

## Typical workflow

1. If the user's request doesn't specify a diagram type, call
   **suggest_diagram_type** with a plain-language description of the
   requirement. Only proceed automatically when it returns
   `confident: true`; otherwise ask the user its `clarifyingQuestion` and
   wait for their answer before continuing.
2. Compose diagram diagram text for the (chosen or specified) diagram type
   yourself.
3. Call the matching **generate_<type>_diagram** tool for that type (or
   **validate_diagram** first if you just want to check syntax without
   writing a file yet) to catch syntax mistakes early and write the file in
   one step.
4. To understand an existing diagram (e.g. one found in the repo or supplied
   by the user), call **parse_diagram** on its content or file path.

## Local CLI preference for read-only file tasks

When operating in a local workspace that has the `diagram-mcp` CLI available,
prefer CLI commands for read-only file operations:

- Use `diagram-mcp render -i <input.mmd|input.diagram> [-o <output.png>] [-f svg|png|jpg]`
  when the user only wants image rendering from an existing file.
- Use `diagram-mcp describe -i <input.mmd|input.diagram> [-o <output.md>]`
  when the user only wants a Markdown description of an existing file.

Use MCP generation/parsing tools when the user asks for generation,
normalization, or tool-driven transformation/analysis beyond these direct
read-only CLI file operations.

## Reviewing Generated PNG Previews (size limits)

Every `generate_diagram`/`generate_<type>_diagram` call renders a PNG preview
alongside the written file and checks its pixel width/height. The tool
result's `imageSizeWarning` field (also appended to `message`) tells you what
to do — **always check it before deciding to view/upload the PNG yourself**:

- **Width or height > 8000px**: the image will be **REJECTED** by the
  server if you try to view/upload it. Do NOT attempt to open or upload
  this PNG at all — just report the file path and ask the user to check it
  manually.
- **Width or height > 4000px** (but ≤ 8000px): the image is technically
  uploadable, but reviewing it yourself wastes a large number of tokens for
  little benefit. Skip automatic review — report the file path and ask the
  user to check it manually instead of viewing it.
- Otherwise (`imageSizeWarning` is empty): the PNG is small enough to view
  normally if you want to double-check the rendered result.

This check exists because very large/complex diagrams (many nodes, wide
gantt charts, etc.) can render to PNGs whose pixel dimensions exceed what an
image-upload backend accepts — always trust `imageSizeWarning` over assuming
"it rendered successfully" means "it's safe/worth viewing".

Diagram-type-specific syntax rules and quoting/authoring gotchas (Treemap,
Packet, Requirement Diagram, etc.) live in each type's own
`generate_<type>_diagram` tool description (see
`internal/mcpserver/diagram_types.go`), not here — check that tool's
`WhenToUse` text before authoring a diagram of an unfamiliar type.

## architecture-beta: System Architecture & Infrastructure Diagrams

**Purpose**: Draw system/cloud architecture with boxes (services/groups) and
labeled connection lines showing how data/requests flow between them.

**Key Concepts**:
- `group ID(icon)[Title]` — Container/cluster (e.g., AWS Cloud, microservice, database tier). Groups can nest. Icon pool: `cloud`, `database`, `disk`, `internet`, `server`, or a custom URI like `logos:aws-lambda`.
- `service ID(icon)[Title] in ParentGroup` — A component/service (e.g., Lambda, PostgreSQL, S3 bucket). Must be inside a group (or omit to place at top level).
- `junction ID` — Tiny diamond; used as a relay point when you need to fork/join edges without adding a visible component.
- `ID1:Side1 -- Side2:ID2` — Connection from ID1's Side1 edge to ID2's Side2 edge. Sides are compass points: `T` (top), `B` (bottom), `L` (left), `R` (right). Arrow: add `>` or `<` around the `--` (e.g., `ID1:R --> L:ID2`).

**Edges Determine Layout** (Critical):
The renderer is NOT just drawing arbitrary boxes and lines — it builds a
**2D grid from your edges**. Each edge's T/B/L/R sides implicitly say "one
box sits to the left/right/above/below the other." Mixed topologies (e.g.,
`db:L--R:server` PLUS `disk:T--B:server`) create genuine 2D grids, not forced
single-row/column layouts.

**How Sides Work**:
- `A:L -- R:B` means "A exits via its LEFT edge; B is approached via its RIGHT edge." B sits ONE COLUMN to the LEFT of A (lower column index).
- `A:R -- L:B` means B sits ONE COLUMN to the RIGHT of A.
- `A:T -- B:B` means B sits ONE ROW ABOVE A.
- `A:B -- T:B` means B sits ONE ROW BELOW A.
- If you connect `A:L--R:B` and `A:T--B:C`, then B is to A's left AND C is below A — a 2D arrangement, not a line.

**Detoured Edges** (Automatic):
When a same-axis edge (both T/B, or both L/R) connects two boxes that ended
up far apart on the OTHER axis, the renderer automatically routes it with 2
bends instead of one long diagonal: down→across→up (or left→across→right).
This keeps connection lines from cutting through unrelated content in between.
The entry point automatically snaps to whichever target edge actually faces
the approach direction (e.g., approaching from below → enter via target's
BOTTOM edge, not the declared side if it conflicts).

**Common Mistakes**:
1. **Forgetting edges** — If you describe a system but don't add edge lines,
   all boxes end up in a single row. Add AT LEAST ONE edge to define layout.
2. **Conflicting topology** — Declaring `A:L--R:B` (B is left of A) AND
   `A:R--L:B` (B is right of A) in the same scope causes the grid solver to
   pick one; result is undefined. Keep topology consistent per scope.
3. **Assuming declared side is sacred** — If you write `aimodel:B-->T:postgres`
   but the grid places aimodel far below postgres, the routed path approaches
   from BELOW → entry snaps to postgres's BOTTOM edge (the one facing the
   approach), not the top. This is correct behavior, not a bug.
4. **Icon typos** — Mismatch like `service foo(cloud)[...] in bar` when bar
   is a service, not a group. Only groups can contain others.
5. **Cross-group alignment** — `align row a b c` only works if a, b, c are
   siblings (same parent). Cross-group alignment is not supported.

**Authoring Tips**:
- Start by drawing a mental map: "Box A talks to B (left-right), B talks to C (up-down)."
- Declare boxes in a logical order (e.g., layers: input, compute, storage).
- Add edges that express the real data/control flow, using T/B/L/R to match physical arrangement.
- If the layout doesn't match your mental map, re-examine your edges — they drive it, not the declaration order.
- Test early with `generate_architecture_diagram` to see renders and catch topology issues before they compound.
- For complex systems (10+ boxes), sketch out your edge list on paper first — it saves iteration.
- **Group by layer/tier** — Don't scatter services randomly across top-level scope. Always use groups to represent architectural layers (Client, API, Compute, Storage, etc.). Groups make the structure clear and guide layout.
- **Minimize junctions** — Junctions are for relay points only (one-to-many fan-out, or multi-way junctions). Don't add a junction to "break up" a simple linear chain.

**Reference Example: AI-Powered Image Processing Pipeline**

This is the correct pattern for generating architecture diagrams. It demonstrates:
- **Three distinct groups** (tiers): Worker Compute, AWS Cloud, AWS Infra — services never scattered at top level
- **Proper layering**: Web/storage tier (AWS) → Processing tier (SQS) → Compute tier (Workers) → AI Model → Database
- **Clear data flow edges**: web→s3→sqs→workers; s3→workers; workers↔model; workers→postgres; workers→web
- **Minimal junctions**: None used (only actual services and groups)
- **Bidirectional feedback**: workers↔model shows iterative AI processing
- **Complete topology**: 8 edges form a coherent cycle, not a scattered layout

```
---
config:
    theme: default
    maxTextSize: 50000
    maxEdges: 500
    fontSize: 16
---
architecture-beta
    group compute(server)[Worker Compute]
    group aws(cloud)[AWS Cloud]
    group aws_db(server)[AWS Infra]

    service web(internet)[Web Application]  in aws
    service s3(database)[S3 Bucket] in aws
    service sqs(internet)[SQS Queue] in aws
    service workers(server)["Worker Pool Python"] in compute
    service model(server)["AI Model Service ANN"] in compute
    service postgres(database)[PostgreSQL] in aws_db

    web:B --> T:s3
    s3:B --> T:sqs
    sqs:R --> L:workers
    s3:R --> L:workers
    workers:B --> T:model
    model:T --> B:workers
    workers:R --> L:postgres
    workers:L --> R:web
```

**Why this pattern works**:
- Groups define clear system boundaries (compute infrastructure vs cloud vs database infrastructure)
- Services are logically organized within groups, not randomly scattered
- Edges form a complete pipeline: user requests → storage → queue → workers → AI model (with feedback) → persistent storage → response
- The layout naturally emerges from the edge topology without manual positioning
- All 6 services connect meaningfully; nothing is orphaned
