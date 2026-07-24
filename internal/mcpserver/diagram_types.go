package mcpserver

import (
	"github.com/xdung24/mermaid-mcp/pkg/parser"
)

// diagramTypeInfo describes one of the 23 supported Mermaid diagram types
// for the purpose of registering a dedicated generate_<type>_diagram tool:
// its ToolName, and a WhenToUse description explaining to a caller (an LLM)
// how to recognize that a user's request calls for this specific diagram
// type, plus a minimal Example snippet showing the expected syntax shape.
type diagramTypeInfo struct {
	DiagramType parser.DiagramType
	ToolName    string
	WhenToUse   string
	Example     string
}

// diagramTypes lists metadata for all 23 supported diagram types, in the
// same fixed order used elsewhere in this repo (internal/suggest).
var diagramTypes = []diagramTypeInfo{
	{
		DiagramType: parser.DiagramTypeFlowchart,
		ToolName:    "generate_flowchart_diagram",
		WhenToUse: "Use for process flows, decision trees, step-by-step procedures, or algorithms " +
			"with branching logic (yes/no, if/else paths) — e.g. \"show the steps to reset a " +
			"password\" or \"diagram this approval workflow with decision points\".",
		Example: "flowchart TD\n    A[Start] --> B{Decision?}\n    B -->|Yes| C[Do X]\n    B -->|No| D[Do Y]",
	},
	{
		DiagramType: parser.DiagramTypeClass,
		ToolName:    "generate_classDiagram_diagram",
		WhenToUse: "Use for object-oriented designs: classes with attributes/methods, inheritance, " +
			"interfaces, and relationships between them — e.g. \"show the class hierarchy for this " +
			"UML model\" or \"diagram these Go structs and their fields\".",
		Example: "classDiagram\n    class Animal {\n      +String name\n      +makeSound()\n    }\n    Animal <|-- Dog",
	},
	{
		DiagramType: parser.DiagramTypeSequence,
		ToolName:    "generate_sequenceDiagram_diagram",
		WhenToUse: "Use for interactions between actors/participants over time: API calls, message " +
			"exchanges, request/response flows — e.g. \"show the sequence of calls between the " +
			"client, API, and database for this login flow\". Authoring conventions: declare human " +
			"entities (users, admins, customers, etc.) with the `actor` keyword instead of " +
			"`participant`; reserve `participant` for non-human entities (services, APIs, databases, " +
			"queues). Wrap server-side processing or API request handling in explicit " +
			"`activate`/`deactivate` blocks (or the `+`/`-` shorthand on the arrow, e.g. " +
			"`Client->>+Server: request` ... `Server-->>-Client: response`) to show how long that " +
			"work takes.",
		Example: "sequenceDiagram\n    actor User\n    participant Server\n    User->>+Server: Request\n    Server-->>-User: Response",
	},
	{
		DiagramType: parser.DiagramTypeState,
		ToolName:    "generate_stateDiagram_diagram",
		WhenToUse: "Use for state machines: the states of an entity and the transitions/events " +
			"between them — e.g. \"diagram the lifecycle of an order from placed to delivered\" or " +
			"\"show the states of a smart light bulb\".",
		Example: "stateDiagram-v2\n    [*] --> Idle\n    Idle --> Running : start\n    Running --> [*] : stop",
	},
	{
		DiagramType: parser.DiagramTypeEntityRelationship,
		ToolName:    "generate_erDiagram_diagram",
		WhenToUse: "Use for database schemas: entities/tables, their attributes, and " +
			"foreign/primary-key relationships — e.g. \"diagram this e-commerce database schema\" or " +
			"\"show how these tables relate\". Syntax rules (see " +
			"https://mermaid.js.org/syntax/entityRelationshipDiagram.html): " +
			"Header is exactly \"erDiagram\". Each relationship statement is " +
			"\"<entity1> <cardinality> <entity2> : <label>\" where <cardinality> is a crow's-foot " +
			"token made of a left part + a line + a right part: left parts are |o (zero-or-one), " +
			"|| (exactly-one), }o (zero-or-more), }| (one-or-more); right parts are o| (zero-or-one), " +
			"|| (exactly-one), o{ (zero-or-more), |{ (one-or-more) — e.g. \"CUSTOMER ||--o{ ORDER : " +
			"places\" or \"CUSTOMER }|..|{ DELIVERY-ADDRESS : uses\". The line is \"--\" for an " +
			"identifying relationship (solid line, the child cannot exist without the parent) or " +
			"\"..\" for non-identifying (dashed line, both entities can exist independently). " +
			"An alternative English-word form is also accepted when parsing/normalizing an existing " +
			"diagram (this tool's own generation always uses the crow's-foot symbol form above, which " +
			"is more compact and unambiguous): \"<entity1> <cardinality-phrase> [optionally] to " +
			"<cardinality-phrase> <entity2> : <label>\", e.g. \"CAR 1 to zero or more NAMED-DRIVER : " +
			"allows\" or \"PERSON many(0) optionally to 0+ NAMED-DRIVER : is\" — \"to\" means " +
			"identifying, \"optionally to\" means non-identifying; the cardinality-phrase aliases are " +
			"\"one or zero\"/\"zero or one\" (zero-or-one), \"one or more\"/\"one or many\"/\"many(1)\"/" +
			"\"1+\" (one-or-more), \"zero or more\"/\"zero or many\"/\"many(0)\"/\"0+\" (zero-or-more), " +
			"and \"only one\"/\"1\" (exactly-one). " +
			"Attributes go inside a \"<entity> { <type> <name> [PK|FK|UK[, PK|FK|UK...]] " +
			"[\\\"comment\\\"] ... }\" block, e.g. \"CUSTOMER { string name string custNumber PK }\"; " +
			"a type ending in \"?\" (e.g. \"string?\") marks it optional/nullable (v11.16.0+); array " +
			"(\"string[]\") and sized (\"string(99)\") types are written directly as the type token; " +
			"multiple keys are comma-separated (\"PK, FK\"); a trailing quoted string is a comment " +
			"(\"string driversLicense PK \\\"The license #\\\"\"). An entity with no attributes is " +
			"written bare (no \"{ }\"), e.g. just \"HOUSE\" on its own line — it does not need to " +
			"appear in any relationship either. Entity names support unicode/spaces/markdown when " +
			"double-quoted (e.g. \"\\\"This **is** _Markdown_\\\"\"); an alias for display purposes " +
			"is \"id[Alias]\" or \"id[\\\"Alias With Spaces\\\"]\" immediately after the id, e.g. " +
			"\"p[Person] { string firstName }\" — the alias, not the id, is what's shown in the " +
			"diagram, but relationships still reference the id (\"p\"), not the alias. \"direction " +
			"TB|BT|RL|LR\" sets the overall orientation. Styling: \"style <id>[,<id>...] " +
			"<fill:#f9f,stroke:#333,...>\" applies a direct style; \"classDef <name>[,<name>...] " +
			"<styleList>\" defines a reusable named style, applied via \"id:::className\" appended " +
			"directly to an entity's own declaration or relationship reference (e.g. \"CAR:::someclass " +
			"{ ... }\" or \"PERSON:::foo ||--|| CAR : owns\"); a classDef named \"default\" applies to " +
			"every entity without its own class.",
		Example: "erDiagram\n    direction LR\n    CUSTOMER ||--o{ ORDER : places\n    CUSTOMER {\n        string name\n        string custNumber PK\n    }\n    ORDER ||--|{ LINE-ITEM : contains\n    ORDER {\n        int orderNumber PK\n        string? deliveryAddress\n    }\n    LINE-ITEM {\n        string productCode PK, FK\n        int quantity\n        float pricePerUnit \"unit price in USD\"\n    }\n    style CUSTOMER fill:#f9f,stroke:#333",
	},
	{
		DiagramType: parser.DiagramTypeTimeline,
		ToolName:    "generate_timeline_diagram",
		WhenToUse: "Use for chronological sequences of events — e.g. \"show the history of web " +
			"development\" or \"diagram the major milestones over the last decade\".",
		Example: "timeline\n    title History\n    2020 : Event A\n    2021 : Event B",
	},
	{
		DiagramType: parser.DiagramTypeUserJourney,
		ToolName:    "generate_journey_diagram",
		WhenToUse: "Use for user/customer journey maps: steps a persona takes plus a satisfaction " +
			"score at each step — e.g. \"map the user journey for booking a flight\".",
		Example: "journey\n    title Booking a Flight\n    section Search\n      Find flight: 5: Customer",
	},
	{
		DiagramType: parser.DiagramTypeBlock,
		ToolName:    "generate_block_diagram",
		WhenToUse: "Use for simple block/system-component diagrams (grid of labeled blocks, not " +
			"cloud-service-specific) — e.g. \"show the high-level modules of this system and how " +
			"they connect\". Header must be exactly `block-beta` (not `block`). Every block is a " +
			"bare-word id immediately followed by a shape wrapper with the label INSIDE double " +
			"quotes — never write unwrapped/plain text as a block: rectangle `id[\"Label\"]`, " +
			"rounded `id(\"Label\")`, stadium `id([\"Label\"])`, subroutine `id[[\"Label\"]]`, " +
			"cylindrical `id[(\"Label\")]`, circle `id((\"Label\"))`, double circle " +
			"`id(((\"Label\")))`, asymmetric `id>\"Label\"]`, rhombus `id{\"Label\"}`, hexagon " +
			"`id{{\"Label\"}}`, parallelogram `id[/\"Label\"/]` (or mirrored `id[\\\"Label\"\\]`), " +
			"trapezoid `id[/\"Label\"\\]` (or alt `id[\\\"Label\"/]`). Directional arrow blocks are " +
			"NOT plain text with an arrow glyph — use the dedicated syntax " +
			"`id<[\"Label\"]>(direction[, direction2])` where direction is right/left/up/down/x/y, " +
			"e.g. `t1<[\"Transform\"]>(right)` or `agg<[\"Aggregate\"]>(up, down)`; writing " +
			"`id[\"→ Transform\"]` is wrong. `columns N` sets the grid width (top-level or inside a " +
			"group); append `:N` after a shape to span N columns; `space`/`space:N` skips cells. " +
			"Group/cluster related blocks with a real named composite — `block:groupId:N` ... " +
			"`end` — not a comment or blank-line trick; the group id is itself stylable/" +
			"connectable. Styling is ALWAYS a separate `style <id> <css-props>` line, never inline " +
			"on the block's declaration. Reusable styles: `classDef name <css-props>` then " +
			"`class id1,id2 name` (space-separated statement, not flowchart's `:::name` shorthand). " +
			"Links always reference block ids, never label text: `A --> B` (directed), `A --- B` " +
			"(undirected), or with a label `A -- \"Data\" --> B` / `A -- \"Data\" --- B`.",
		Example: "block-beta\n    columns 3\n    dash[\"Dashboard\"]:3\n    style dash fill:#0066CC,color:#fff\n    " +
			"block:sources:3\n        db[(\"Database\")]\n        api{{\"API\"}}\n        files[/\"Files\"/]\n    end\n    " +
			"style sources fill:#9933CC\n    etl[[\"ETL Pipeline\"]]\n    t1<[\"Transform\"]>(right)\n    ml{\"ML Model\"}\n    " +
			"dash --> sources\n    db -- \"extract\" --> etl\n    etl --> t1\n    t1 --> ml",
	},
	{
		DiagramType: parser.DiagramTypeGantt,
		ToolName:    "generate_gantt_diagram",
		WhenToUse: "Use for project schedules: tasks with start/end dates or durations, " +
			"dependencies, and milestones — e.g. \"create a project timeline for this release plan\". " +
			"Always give the chart a `title` (this tool cannot invent one for you, so always author " +
			"a short descriptive title line yourself). If the content doesn't already set its own " +
			"axisFormat or excludes, this tool auto-applies sensible defaults before writing the " +
			"file: `axisFormat %m-%d` and `excludes weekends` — you don't need to add these yourself " +
			"unless you want different values, in which case set your own and they'll be left as-is. " +
			"Authoring conventions (enforced - invalid content is rejected): give every task a short, " +
			"logical, UNIQUE id (e.g. d1, d2, dev1, dev2) and chain scheduling with the `after` " +
			"keyword (e.g. `after d1`) rather than hardcoding static calendar dates, unless the user " +
			"explicitly asked for fixed dates. Define milestones strictly as `:milestone, after " +
			"<task_id>, 0d` — never combine `milestone` with another modifier tag (`crit`, `active`, " +
			"`done`) in the same comma-separated list; that mixture breaks Mermaid's parser and this " +
			"tool will reject it (duplicate task ids are rejected too). When presenting the result to " +
			"the user, output only the raw Mermaid code in a ```mermaid fence — no conversational " +
			"filler before or after it.",
		Example: "gantt\n    title Project\n    dateFormat YYYY-MM-DD\n    section Phase 1\n    " +
			"Design :d1, 2024-01-01, 5d\n    Development :d2, after d1, 10d\n    " +
			"Launch :milestone, after d2, 0d",
	},
	{
		DiagramType: parser.DiagramTypePie,
		ToolName:    "generate_pie_diagram",
		WhenToUse: "Use for proportional/percentage breakdowns of a small set of values — e.g. " +
			"\"show the market share breakdown as a pie chart\". This tool automatically applies a " +
			"colorful \"base\" theme with a vibrant per-slice palette (pie1-pie12 theme variables) " +
			"if the content doesn't already declare an explicit theme via a YAML front-matter " +
			"`config: theme:` block — don't author theme colors yourself unless the user asks for " +
			"specific ones, and don't add a `%%{init: ...}%%` comment directive (it's ignored; use " +
			"front-matter instead).",
		Example: "pie title Share\n    \"A\" : 40\n    \"B\" : 60",
	},
	{
		DiagramType: parser.DiagramTypeQuadrant,
		ToolName:    "generate_quadrantChart_diagram",
		WhenToUse: "Use for plotting items across two axes split into four quadrants (e.g. " +
			"impact-vs-effort, reach-vs-engagement prioritization matrices) — e.g. \"plot these " +
			"features on an impact vs effort quadrant chart\". Per " +
			"https://mermaid.js.org/syntax/quadrantChart.html:\n" +
			"CRITICAL: the x-axis and y-axis text must explain WHAT EACH AXIS IS MEASURING, not be " +
			"generic/empty labels — e.g. `x-axis Low Reach --> High Reach` tells the reader the " +
			"horizontal position represents \"Reach\", `y-axis Low Engagement --> High Engagement` " +
			"tells them the vertical position represents \"Engagement\". Never author a bare " +
			"`x-axis Low --> High` / `y-axis Low --> High` with no subject — always name the actual " +
			"metric/quantity being compared on each end (e.g. \"Low Cost\"/\"High Cost\", not just " +
			"\"Low\"/\"High\"). Syntax rules:\n" +
			"1. `title <text>` — short chart description, rendered above the plot.\n" +
			"2. `x-axis <left text> [--> <right text>]` — the statement must start with `x-axis`; " +
			"the right side after `-->` is optional (only the left/low label renders if omitted), " +
			"but the left label is always required and must describe the low end of the measured " +
			"quantity.\n" +
			"3. `y-axis <bottom text> [--> <top text>]` — same rule, bottom label is required and " +
			"describes the low end; top after `-->` is optional.\n" +
			"4. `quadrant-1 <text>` = top-right, `quadrant-2 <text>` = top-left, `quadrant-3 <text>` " +
			"= bottom-left, `quadrant-4 <text>` = bottom-right — each names the action/category for " +
			"items landing in that quadrant (e.g. quadrant-1 \"We should expand\"). If no data " +
			"points are declared at all, both the axis text and quadrant text render centered in " +
			"each quadrant; once points exist, axis labels shift to the plot edges and quadrant text " +
			"moves to the top of its quadrant.\n" +
			"5. Points: `<label>: [x, y]` where x and y are each in the 0-1 range (0 = low end of " +
			"that axis, 1 = high end) — e.g. `Campaign A: [0.3, 0.6]`. Every point must fit inside " +
			"the 0-1 unit square.\n" +
			"6. Point styling: append space-separated `key: value` pairs after the coordinates, " +
			"comma-joined — `color` (fill), `radius` (dot size), `stroke-width`, `stroke-color` — " +
			"e.g. `Campaign B: [0.8, 0.1] color: #ff3300, radius: 10`.\n" +
			"7. Shared styling via classes: define `classDef <name> <key>: <value>, ...` then assign " +
			"it to a point with `<label>:::<name>: [x, y]` (a point's own inline styles override the " +
			"classDef's).\n" +
			"8. Optional nested `config.quadrantChart.*` front-matter block (all numeric px values " +
			"unless noted) tunes layout without touching the data: chartWidth/chartHeight (plot " +
			"size, default 500), titlePadding/titleFontSize, quadrantPadding (gap outside all " +
			"quadrants), quadrantTextTopPadding, quadrantLabelFontSize, " +
			"quadrantInternalBorderStrokeWidth/quadrantExternalBorderStrokeWidth, " +
			"xAxisLabelPadding/xAxisLabelFontSize/xAxisPosition (\"top\"|\"bottom\", forced to " +
			"\"bottom\" once points exist), yAxisLabelPadding/yAxisLabelFontSize/yAxisPosition " +
			"(\"left\"|\"right\"), pointTextPadding/pointLabelFontSize/pointRadius.",
		Example: "---\n" +
			"config:\n" +
			"  quadrantChart:\n" +
			"    chartWidth: 400\n" +
			"    chartHeight: 400\n" +
			"---\n" +
			"quadrantChart\n" +
			"    title Reach and engagement of campaigns\n" +
			"    x-axis Low Reach --> High Reach\n" +
			"    y-axis Low Engagement --> High Engagement\n" +
			"    quadrant-1 We should expand\n" +
			"    quadrant-2 Need to promote\n" +
			"    quadrant-3 Re-evaluate\n" +
			"    quadrant-4 May be improved\n" +
			"    Campaign A: [0.3, 0.6]\n" +
			"    Campaign B:::highPriority: [0.8, 0.1] radius: 10\n" +
			"    classDef highPriority color: #ff3300",
	},
	{
		DiagramType: parser.DiagramTypeRequirement,
		ToolName:    "generate_requirementDiagram_diagram",
		WhenToUse: "Use for requirements traceability / SysML-style requirement diagrams: " +
			"requirements, their risk/verification method, and relationships (satisfies, traces, " +
			"etc.) to design elements — e.g. \"diagram how these system requirements trace to " +
			"design elements\". IMPORTANT: a requirement/element block has TWO distinct " +
			"identifiers that must never be conflated — the token right after the block keyword " +
			"(e.g. `requirement CollisionAvoidance {`) is the diagram's internal node name, used " +
			"to reference this block in relationship lines; the `id:` field inside the block body " +
			"is a separate free-form label such as a real-world requirement ID (e.g. \"REQ-001\"). " +
			"When the source material gives both a name/title and a separate ID, put the " +
			"name/title (as an identifier-safe token) in the block-header position and the given " +
			"ID in the `id:` field — never put the real ID in the header position, and never " +
			"invent a fake numeric id when the source didn't provide one. Block types: " +
			"`requirement`, `functionalRequirement`, `interfaceRequirement`, " +
			"`performanceRequirement`, `physicalRequirement`, `designConstraint` (fields: `id`, " +
			"`text`, `risk` [low/medium/high], `verifymethod` " +
			"[analysis/inspection/test/demonstration]) and `element` (fields: `type`, `docRef`). " +
			"Model physical/software components, documents, or test cases that satisfy/verify " +
			"requirements as `element { type: \"...\" }` blocks, not as fabricated requirement " +
			"subtypes with made-up risk/verifymethod values. Always quote `text`/`id`/`type`/" +
			"`docRef` values — an unquoted value containing spaces or a reserved keyword (e.g. " +
			"\"test\", \"high\") fails to parse. A relationship may be written forward " +
			"(`{source} - <type> -> {destination}`) or backward (`{destination} <- <type> - " +
			"{source}`) — both forms are fully supported and equivalent. `<type>` is one of " +
			"`contains`, `copies`, `derives`, `satisfies`, `verifies`, `refines`, `traces`. An " +
			"optional `direction TB|BT|LR|RL` statement controls layout (default TB). Styling is " +
			"supported via `style <id> fill:#..,stroke:#..`, reusable `classDef <name> " +
			"fill:#..,stroke:#..` applied via `class <id1>,<id2> <name>` or the `<id>:::<name>` " +
			"shorthand (a classDef named \"default\" applies to every node unless overridden by " +
			"a more specific class/style).",
		Example: "requirementDiagram\n    direction LR\n\n    requirement CollisionAvoidance {\n      id: \"REQ-001\"\n      text: \"must avoid collision with obstacles\"\n      risk: high\n      verifymethod: test\n    }\n    element SensorFusionModule {\n      type: \"interface\"\n    }\n\n    SensorFusionModule - satisfies -> CollisionAvoidance\n    CollisionAvoidance <- derives - SensorFusionModule\n\n    classDef important fill:#f96,stroke:#333\n    class CollisionAvoidance important",
	},
	{
		DiagramType: parser.DiagramTypeArchitecture,
		ToolName:    "generate_architecture_diagram",
		WhenToUse: "Use for cloud/infrastructure architecture diagrams with grouped services, " +
			"icons, and connections between them — e.g. \"diagram this AWS deployment with its " +
			"services and how they connect\". Write group/service titles naturally (spaces, " +
			"hyphens, parentheses, etc. are all fine, e.g. `[Worker Pool (Python)]` or " +
			"`[AI-Powered Pipeline]`) — this tool automatically quotes any title containing " +
			"characters outside real Mermaid's unquoted title charset (letters/digits/" +
			"underscore/spaces only) before writing the file, so you don't need to manually " +
			"quote or simplify titles yourself.\n" +
			"Icons: the only 5 icon names that render as a real drawn glyph (cloud/database/" +
			"disk/internet/server shape) are `cloud`, `database`, `disk`, `internet`, `server` " +
			"— e.g. `service db(database)[Database]`. Any other icon name is treated as a " +
			"registered iconify.design icon pack reference in the form `pack:icon-name` (e.g. " +
			"`logos:aws-lambda`); this renderer draws those as a small text fallback (it can't " +
			"fetch/embed arbitrary external icon artwork) so prefer one of the 5 built-in icons " +
			"unless the user specifically wants a named cloud-provider icon in the source text.\n" +
			"Groups: `group {id}({icon})[{title}] (in {parentId})?` — groups nest via `in`, and " +
			"services/junctions join a group the same way (`service db(database)[DB] in api`).\n" +
			"Edges are the most common source of syntax errors: every edge MUST specify an " +
			"explicit side (T/B/L/R) on BOTH ends, in the form `sourceId:SIDE -- SIDE:targetId` " +
			"(e.g. `db:L -- R:server`) — do NOT write plain flowchart-style arrows like `a --> b` " +
			"with no sides. Add `<` before/`>` after the `--` for arrowheads, e.g. " +
			"`web:R --> L:api` or `db:L <--> R:cache`. Edges do NOT support text labels at all — " +
			"never write `a --> b: some label`; if you need to explain what a connection does, " +
			"put that detail in the connected services' `[Title]` text instead, not on the edge. " +
			"To connect to a service's ENCLOSING GROUP rather than the service itself, append " +
			"`{group}` right after the service id on either end, e.g. " +
			"`server{group}:B --> T:subnet{group}` draws the edge out of `server`'s own group " +
			"and into `subnet`'s own group — `{group}` only works on a service inside a group, " +
			"never directly on a group id.\n" +
			"Alignment (v11.16.0+, use when several services independently feeding/fed-by one " +
			"downstream/upstream node would otherwise collapse onto the same layout position): " +
			"`align row {id1} {id2} ...` forces those already-declared services/junctions (>= 2, " +
			"same parent group or all top-level) to share one horizontal row, in that listed " +
			"order; `align column {id1} {id2} ...` stacks them in one vertical column instead. " +
			"Use `row` when the members feed a shared node via top/bottom ports (e.g. all " +
			"`B --> T:proc`); use `column` when they feed it via left/right ports (e.g. all " +
			"`R --> L:mcp`). Each align directive lives on its own line and only affects members " +
			"that all share the same parent (or are all top-level) — this tool's renderer does " +
			"not support aligning members split across different groups.\n" +
			"A nested `config.architecture` front-matter block tunes the underlying fcose force " +
			"layout: `randomize` (bool), `nodeSeparation` (px between siblings), " +
			"`idealEdgeLengthMultiplier`, `edgeElasticity` (0-1), `numIter`, `seed` (deterministic " +
			"layout variant) — these are recorded and round-tripped but this tool's own " +
			"deterministic row/column renderer doesn't visually react to them the way real " +
			"Mermaid's fcose layout does, so don't expect them to change the generated preview " +
			"image noticeably.",
		Example: "architecture-beta\n    group api(cloud)[API]\n    service web(server)[Web App]\n    service db(database)[Database] in api\n    web:R --> L:db",
	},
	{
		DiagramType: parser.DiagramTypeKanban,
		ToolName:    "generate_kanban_diagram",
		WhenToUse: "Use for kanban/task boards: columns (e.g. To Do / In Progress / Done) each " +
			"containing tasks — e.g. \"show this sprint's kanban board\".",
		Example: "kanban\n    Todo\n      [Task A]\n    Done\n      [Task B]",
	},
	{
		DiagramType: parser.DiagramTypeMindmap,
		ToolName:    "generate_mindmap_diagram",
		WhenToUse: "Use for brainstorming/idea maps: a central concept with branching sub-ideas " +
			"in a tree — e.g. \"create a mindmap of ideas for this project\".",
		Example: "mindmap\n  root((Idea))\n    Branch A\n    Branch B",
	},
	{
		DiagramType: parser.DiagramTypePacket,
		ToolName:    "generate_packet_diagram",
		WhenToUse: "Use for network packet/protocol header layouts showing bit fields — e.g. " +
			"\"diagram the bit layout of this TCP packet header\". Each field line after the " +
			"header is either an explicit bit range or the \"+N\" bit-count shorthand (v11.7.0+) " +
			"— it's fine to mix both forms in the same diagram: (1) explicit range: " +
			"`start-end: \"Label\"` (e.g. `0-15: \"Source Port\"`), or `start: \"Label\"` for a " +
			"single bit (e.g. `106: \"URG\"`); (2) `+N` shorthand: `+N: \"Label\"` declares a " +
			"field N bits wide starting immediately after the previous field's last bit (0 if " +
			"it's the first field) — much easier than manually recalculating every start/end " +
			"when a field's width changes. Field labels must always be quoted (`\"...\"`). An " +
			"inline `title <text>` line may appear right after the header (in addition to/" +
			"instead of the front-matter `title:` form). A nested `config.packet` front-matter " +
			"block controls rendering: `rowHeight` (default 32), `bitWidth` (default 32), " +
			"`bitsPerRow` (default 32), `showBits` (boolean, default true — toggles the " +
			"bit-index ruler), `paddingX`/`paddingY` (spacing between blocks/rows).",
		Example: "packet\ntitle UDP Packet\n+16: \"Source Port\"\n+16: \"Destination Port\"\n32-47: \"Length\"\n48-63: \"Checksum\"\n64-95: \"Data (variable length)\"",
	},
	{
		DiagramType: parser.DiagramTypeXYChart,
		ToolName:    "generate_xychart_diagram",
		WhenToUse: "Use for bar/line charts plotting numeric series across categories or a numeric " +
			"axis — e.g. \"chart monthly revenue as a bar chart\".",
		Example: "xychart-beta\n    x-axis [Jan, Feb, Mar]\n    y-axis \"Revenue\" 0 --> 100\n    bar [10, 40, 70]",
	},
	{
		DiagramType: parser.DiagramTypeTreemap,
		ToolName:    "generate_treemap_diagram",
		WhenToUse: "Use for hierarchical proportional data shown as nested rectangles — e.g. " +
			"\"show disk space usage by folder as a treemap\" or \"diagram budget allocation by " +
			"department\". CRITICAL QUOTING RULES (required to avoid parser errors): (1) EVERY " +
			"section name and leaf name MUST be wrapped in double quotes — no exceptions. " +
			"(2) Leaf names with values use the syntax `\"Name\": value` (colon + value after " +
			"closing quote). (3) Section names (parents, no direct value) use `\"Name\"` alone " +
			"on a line, indented. (4) INVALID: unquoted names, `\"Name[value]\"` syntax, or " +
			"`\"Name\"[value]` — these WILL fail parsing. (5) Nesting: indent children 4 spaces " +
			"deeper than their parent. Do NOT add a config.treemap.valueFormat setting " +
			"(e.g. \"$0,0\", \".1%\") unless the user explicitly asks for a specific number " +
			"format for the leaf values — leave it unset by default so values render as plain " +
			"numbers.",
		Example: "treemap-beta\n    \"Global Budget\"\n      \"Engineering\"\n        \"Developers\": 5000000\n        \"QA\": 1000000\n      \"Sales\": 3000000",
	},
	{
		DiagramType: parser.DiagramTypeSankey,
		ToolName:    "generate_sankey_diagram",
		WhenToUse: "Use for flow/quantity distribution between nodes where band width is " +
			"proportional to the amount flowing — e.g. \"show how energy moves from sources to " +
			"end uses\" or \"visualize the budget flowing from revenue streams to expense " +
			"categories\". Syntax is CSV: one `source,target,value` flow per line after the " +
			"`sankey-beta` header. A node name containing a comma must be wrapped in double " +
			"quotes.",
		Example: "sankey-beta\n    Revenue,Salaries,50\n    Revenue,Marketing,20\n    Revenue,Savings,30",
	},
	{
		DiagramType: parser.DiagramTypeVenn,
		ToolName:    "generate_venn_diagram",
		WhenToUse: "Use for showing how sets/categories overlap or share members, drawn as overlapping " +
			"circles — e.g. \"show how these two teams' responsibilities overlap\" or \"diagram the " +
			"overlap between desirable, feasible, and viable ideas\". Authoring conventions: declare " +
			"each set with `set <id>[\"label\"]`, then declare each overlap with `union " +
			"<id1>,<id2>,...[\"label\"]` referencing set ids already declared earlier (2 or more member " +
			"ids per union; 3+ is a valid \"higher-arity\" union). Only 2-3 sets render as a true " +
			"geometric Venn diagram; avoid more than 3 sets unless the user explicitly asks for it, " +
			"since real overlapping-circle layouts beyond 3 sets are visually cluttered. Optional: " +
			"nest `text <id>[\"label\"]` lines (indented, with an id distinct from any set/union id) " +
			"under a set/union to place extra text inside that region, and `style <id1>,<id2> " +
			"fill:#hex,color:#hex` lines to color specific sets/unions/text nodes. IMPORTANT: a `text` " +
			"line attaches to whichever `set`/`union` was declared MOST RECENTLY above it, purely by " +
			"line order — NOT by matching any id in the text line itself. So each `text` line must " +
			"come immediately after its own set's/union's declaration line, before any other `set`/" +
			"`union` line; writing all `set`/`union` lines first and appending every `text` line " +
			"afterward will silently attach ALL of them to the last-declared union instead of their " +
			"intended sets.",
		Example: "venn-beta\n    title Team overlap\n    set Frontend\n        text F1[\"UI components\"]\n    set Backend\n        text B1[\"Database\"]\n    union Frontend,Backend[\"APIs\"]",
	},
	{
		DiagramType: parser.DiagramTypeGitGraph,
		ToolName:    "generate_gitGraph_diagram",
		WhenToUse: "Use for git commit/branch history: commits, branches, checkouts, merges, and " +
			"cherry-picks — e.g. \"diagram the git branching strategy for this release\".",
		Example: "gitGraph\n    commit\n    branch develop\n    checkout develop\n    commit",
	},
	{
		DiagramType: parser.DiagramTypeC4,
		ToolName:    "generate_c4_diagram",
		WhenToUse: "Use for C4 model software architecture views (context, container, component, " +
			"dynamic, or deployment) — e.g. \"create a C4 container diagram for this system\" or " +
			"\"show the system context diagram with external actors\".",
		Example: "C4Context\n    Person(user, \"User\")\n    System(sys, \"System\")\n    Rel(user, sys, \"Uses\")",
	},
	{
		DiagramType: parser.DiagramTypeCynefin,
		ToolName:    "generate_cynefin_diagram",
		WhenToUse: "Use for categorizing items/situations into the Cynefin sense-making framework's five " +
			"complexity domains to match a response approach to how well the cause-and-effect of a " +
			"situation is understood — e.g. \"categorize these incident types by Cynefin domain\" or " +
			"\"show which of these decisions are clear/complicated/complex/chaotic\". Authoring " +
			"conventions: only five domain keywords are recognized — `complex`, `complicated`, `clear`, " +
			"`chaotic`, `confusion` — declared as a bare line (any order; layout position is always fixed: " +
			"Complex top-left, Complicated top-right, Chaotic bottom-left, Clear bottom-right, Confusion " +
			"center). Nest quoted item labels on their own indented lines under a domain to place them in " +
			"that region; keep the `confusion` domain's item list very short (3 or fewer) since it renders " +
			"in a small center region — its purpose is to hold items whose domain isn't yet known, not to " +
			"collect many items. Optional: top-level `<domain> --> <domain> : \"label\"` lines show movement " +
			"of items between domains over time (e.g. complex --> complicated once a pattern is understood).",
		Example: "cynefin-beta\n    title Incident Response\n    complex\n      \"Investigate root cause\"\n    complicated\n      \"Expert review needed\"\n    complex --> complicated : \"Pattern identified\"",
	},
}
