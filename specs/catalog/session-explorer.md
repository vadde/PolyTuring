# Spec: Session Explorer

> **Status**: Review
> **Author**: @vadde
> **Created**: 2026-09-18
> **Last Updated**: 2026-09-18
> **Tool Path**: `tools/session-explorer/`

---

## 1. Overview

### 1.1 Problem Statement

Engineers using the Antigravity IDE accumulate rich conversational histories with AI agents — prompts, responses, tool invocations, reasoning traces, and artifacts — across multiple workspaces and projects. This data is invaluable for:

- **Knowledge retrieval**: "What did I ask the agent about Kubernetes last week?"
- **Prompt reuse**: "Find the prompt I used to scaffold that Python project"
- **Audit and reflection**: "What tools did the agent use for debugging?"
- **Learning**: "Show me all GenAI conversations to study agent reasoning patterns"

However, this data is buried in raw JSONL transcript files scattered across `~/.gemini/antigravity-ide/brain/<conversation-id>/` directories with no UI, no search, and no organization. Users of all technical levels — from junior engineers to principal architects to executives — need a fast, beautiful, and intuitive way to explore, search, and filter their agent session history.

### 1.2 Proposed Solution

**Session Explorer** is a self-contained CLI tool (single Go binary) that:

1. Scans the Antigravity IDE's `brain/` directory to discover all conversation sessions
2. Parses JSONL transcript files to extract user prompts, agent responses, tool calls, timestamps, and workspace context
3. Serves a stunning, premium web UI (embedded in the binary) on a local port
4. Automatically opens the user's default browser to the dashboard
5. Provides rich filtering, search, and visualization of session data

The tool delivers a **glassmorphic, liquid-glass, acrylic-inspired** UI that feels premium and modern — not a basic CRUD interface, but a polished product worthy of executive demos.

### 1.3 Target Audience

| Persona | Needs |
|---------|-------|
| **Junior Engineer** | Search past prompts, learn from agent responses, reuse effective prompts |
| **Senior Engineer** | Audit agent tool usage, compare approaches across sessions, filter by workspace |
| **Tech Lead / Architect** | Review team prompts (shared machines), analyze agent reasoning quality |
| **Engineering Manager** | Understand tool adoption, session frequency, conversation complexity |
| **Executive** | High-level dashboard of AI-assisted development activity |

### 1.4 Success Criteria

- Single command (`session-explorer`) launches a browser with all sessions visible
- Users can find any past prompt within 10 seconds using search/filter
- UI receives "wow" reactions — glassmorphism, smooth animations, responsive layout
- Works on macOS, Linux, and Windows without dependencies
- Handles 1000+ sessions and 100K+ transcript lines without lag

---

## 2. Requirements

### 2.1 Functional Requirements

| ID | Requirement | Priority | Impl | Tested | Notes |
|----|------------|----------|------|--------|-------|
| R-001 | Discover and list all conversation sessions from `~/.gemini/antigravity-ide/brain/` | Must | ✅ | 🧪 | scanner.go ScanAll(), scanner_test.go |
| R-002 | Parse JSONL transcript files (`transcript.jsonl` and `transcript_full.jsonl`) to extract structured data | Must | ✅ | 🧪 | scanner.go scanSession(), scanner_test.go |
| R-003 | Extract user prompts (type: `USER_INPUT`, source: `USER_EXPLICIT`) with timestamps | Must | ✅ | 🧪 | scanner.go extractUserPrompt(), scanner_test.go |
| R-004 | Extract agent responses (type: `PLANNER_RESPONSE`) with content and tool calls | Must | ✅ | 🧪 | scanner.go scanSession(), scanner_test.go |
| R-005 | Extract workspace/project context from `ADDITIONAL_METADATA` in user messages | Must | ✅ | 🧪 | scanner.go extractWorkspace(), scanner_test.go |
| R-006 | Serve an embedded web UI on a local HTTP port (default: 9876) | Must | ✅ | 🧪 | server.go StartServer(), verified with embedded assets |
| R-007 | Auto-open the user's default browser on launch | Must | ✅ | 🧪 | main.go openBrowser() |
| R-008 | Dashboard view: show all sessions as cards with metadata (date, workspace, step count, size) | Must | ✅ | 🧪 | src/web/src/app.js renderSessionGrid() |
| R-009 | Session detail view: show full conversation timeline (user ↔ agent) with rendered content | Must | ✅ | 🧪 | src/web/src/app.js renderTimelineMessages() |
| R-010 | Filter sessions by date range (from/to date picker) | Must | ✅ | 🧪 | scanner.go GetSessions(), scanner_test.go |
| R-011 | Filter sessions by workspace/project | Must | ✅ | 🧪 | scanner.go GetSessions(), api_test.go |
| R-012 | Full-text search across all user prompts | Must | ✅ | 🧪 | scanner.go Search(), scanner_test.go |
| R-013 | Filter by message type (user prompts, agent responses, tool calls, errors) | Should | ✅ | 🧪 | src/web/src/app.js, api.go |
| R-014 | Sort sessions by date (newest/oldest), size, or step count | Should | ✅ | 🧪 | scanner.go GetSessions(), app.js |
| R-015 | Show tool call details (name, arguments) in a collapsible section | Should | ✅ | 🧪 | src/web/src/app.js details accordion |
| R-016 | Render markdown content in agent responses beautifully | Should | ✅ | 🧪 | marked.js + highlight.js integration |
| R-017 | Show session statistics: total sessions, total prompts, most active workspace, date distribution | Should | ✅ | 🧪 | scanner.go computeStats(), scanner_test.go |
| R-018 | Glassmorphic/acrylic/liquid-glass UI design with smooth animations | Must | ✅ | 🧪 | src/web/src/styles.css |
| R-019 | Responsive design: work on desktop, tablet, and mobile viewports | Should | ✅ | 🧪 | src/web/src/styles.css media queries |
| R-020 | Dark mode by default with optional light mode toggle | Should | ✅ | 🧪 | src/web/src/styles.css, app.js theme toggle |
| R-021 | Export a session's conversation to markdown or JSON | Could | ✅ | 🧪 | src/web/src/app.js exportCurrentSession() |
| R-022 | Keyboard shortcuts for navigation (Ctrl+K for search, Esc to close) | Could | ✅ | 🧪 | src/web/src/app.js keydown listener |
| R-023 | Configurable data directory path via CLI flag (`--data-dir`) | Must | ✅ | 🧪 | main.go flags, tested in demo.sh |
| R-024 | Configurable port via CLI flag (`--port`) | Must | ✅ | 🧪 | main.go flags, tested in demo.sh |
| R-025 | Show "most frequently used prompts" or prompt frequency analysis | Could | ✅ | 🧪 | scanner.go computeStats(), app.js Insights modal |
| R-026 | Copy prompt text to clipboard with one click | Should | ✅ | 🧪 | src/web/src/app.js copy buttons + toast |
| R-027 | Lazy loading / virtual scrolling for large conversation timelines | Should | ✅ | 🧪 | src/web/src/app.js DOM rendering |
| R-028 | Multi-tier repository and workspace clustering: authoritative DB metadata, weighted empirical tool touches, command touch sanitization, Genesis Workspace Invariant (home/non-git protection), and graceful non-git folder fallback | Must | ✅ | 🧪 | scanner.go discoverKnownProjects(), scanSession(), scanner_test.go |
| R-029 | Flexible user-driven grouping modes: Project Hub, Date Timeline, or Flat List | Must | ✅ | 🧪 | src/web/src/app.js renderView(), index.html segmented control |
| R-030 | Agent Work-In-Progress (WIP) Recovery API (`/api/wip`) | Must | ✅ | 🧪 | scanner.go GetWorkspaceWIP(), api.go HandleWIP |
| R-031 | Multi-Session Disambiguation Invariant & Mandatory Agent Halting | Must | ✅ | 🧪 | api.go, scripts/wip.sh, session-wip skill |
| R-032 | Embedded OpenAPI 3.0 Contract & Offline Swagger UI (`/docs`, `/api/openapi.json`) | Must | ✅ | 🧪 | openapi.json, server.go, api.go |
| R-033 | One-Click GUI Skill Installer Modal & CLI Command (`make setup-session-wip`) | Should | ✅ | 🧪 | src/web/src/app.js, setup-skill.sh, api.go HandleInstallSkill |
| R-034 | Headless CLI Subcommand (`session-explorer wip`) & Fallback Companion Script | Should | ✅ | 🧪 | main.go, scripts/wip.sh |
| R-035 | Granular Message Type Filtering: Separate Conversational Agent Responses from Tool Execution Steps, with Dialogue Mode | Should | ✅ | 🧪 | src/web/src/app.js, src/web/index.html, src/scanner.go |
| R-036 | Universal Multi-Runtime Discovery & Provider Architecture (Antigravity IDE + CLI, Claude Code, GitHub Copilot) | Must | ✅ | 🧪 | src/types.go, src/scanner.go, src/main.go, src/api.go, src/web/ |
| R-037 | Bounded Diagram Viewports & Fullscreen Theater Mode (constrained 440px scrollable canvas, distraction-free pan/zoom theater modal, double-click to expand) | Must | ✅ | 🧪 | src/web/src/styles.css, src/web/src/app.js, src/web/index.html |
| R-038 | Native In-Browser File & Artifact Viewer (read files securely via GET /api/file, line number gutter, target line anchor scrolling/highlighting, markdown preview with KaTeX/Mermaid, session files aggregation drawer) | Must | ✅ | 🧪 | src/api.go HandleFile, src/web/src/app.js openFileViewer, src/web/index.html |
| R-039 | Universal Agent Context & Behavioral Matrix: discover and inspect global, workspace, and session rules, skills, plugins, MCP servers, and hooks with precedence cascade and native file viewer inspection | Must | ✅ | 🧪 | context_scanner.go, api.go, context_test.go, src/web/ |
| R-040 | Universal Session Agent Brain & MCP Behavioral Matrix: Empirical MCP and Native tool telemetry, step timelines, tool diversity index, and holographic mica matrix UI/UX | Must | ✅ | 🧪 | context_scanner.go, src/types.go, api.go, src/web/ |
| R-041 | Command Execution & Tool Sub-Breakdown Distribution Engine: Decompose run_command and filesystem tools into command spectrum, binary distribution, category weights, step timelines, and holographic telemetry | Must | ✅ | 🧪 | context_scanner.go, src/types.go, context_test.go, src/web/ |
| R-042 | Empirical Skill Invocation Telemetry & Standby Decomposition: 3-tier epistemic taxonomy, step indices, invocation triggers, scope deduplication, and holographic skill card UI/UX | Must | ✅ | 🧪 | context_scanner.go, src/types.go, context_test.go, src/web/ |
| R-043 | Bin-Packing Spatial Density & Panoramic Masonry Matrix: EC2-style greedy shortest-column container placement, 3-column panoramic widescreen expansion (width: min(1560px, 94vw)), and zero-vertical-gap flow | Must | ✅ | 🧪 | src/web/src/app.js, src/web/src/styles.css, src/web/index.html |
| R-044 | Live Reactive Telemetry Synchronization & Session Trail Refresh: Real-time dynamic updates of Files & Artifacts and Agent Context during SSE live following, tactile counter pulse animations, and dedicated non-destructive Session Trail refresh button (`#refreshTrailBtn`) | Must | ✅ | 🧪 | src/web/src/app.js, src/web/src/styles.css, src/web/index.html |
| R-045 | Universal Session Files & Artifacts Holographic Matrix: Year 2500 panoramic widescreen dialog shell (`width: min(1560px, 94vw)`), 4 forensic lenses (Universal Catalog, Deliverables & Artifacts, Codebase Mutations, Churn & Blast Radius Telemetry), 5-capsule telemetry ribbon, multi-spectrum filter chips, EC2-style greedy bin-packing, timeline step jump navigation, and native file inspection | Must | ✅ | 🧪 | src/web/src/app.js, src/web/src/styles.css, src/web/index.html |
| R-046 | Universal Multi-Dimensional Forensic Sorting Engine: Interactive holographic sort capsules across Files & Artifacts, In-Session Agent Brain, and Global Machine matrix modals supporting churn intensity, deliverable priority, step recency, chronological flow, read volume, token bloat, and touch span telemetry | Must | ✅ | 🧪 | src/web/src/app.js, src/web/src/styles.css, src/web/index.html |
| R-047 | 1-Click Interactive Session UUID Copy Capsule: Top-left session trail navigation header renders an isomorphic aero-mica capsule with 1-click clipboard copy button, kinetic icon transition (`📋` → `✓`), and tactile toast notification | Must | ✅ | 🧪 | src/web/src/app.js, src/web/src/styles.css, src/web/index.html |
| R-048 | Dual-Temporal Telemetry, Active Recency Engine & Multi-Day Lifespan Tracking: Session tiles render both ⚡ Last Active recency (with live neon pulse) and 🌱 Inception Date; multi-day lifespan badges (span: $t_{\text{init}} \rightarrow t_{\text{active}}$); Date Grouped View partitions by effective recency ($\max(t_{\text{active}}, t_{\text{init}})$); sorting engine supports `active_desc` (Last Active Recent First) across backend and frontend | Must | ✅ | 🧪 | src/scanner.go, src/api.go, src/web/src/app.js, src/web/src/styles.css, src/web/index.html |
| R-049 | Inclusive Lifetime Date Filtering Engine: Date range filtering (`from`/`to` and presets `Today`, `7d`, `30d`) naturally evaluates inclusive lifetime overlap (`lastActive >= from && createdAt <= to`), surfacing all sessions touched or initiated in the window without cluttering toolbar elements | Must | ✅ | 🧪 | src/scanner.go, src/api.go, src/web/src/app.js, src/web/src/styles.css, src/web/index.html |
| R-050 | Session Trail Dual-Temporal Quantum Capsule & Live Recency Telemetry: Top-left Session Trail header modernizes `#detailTime` into a liquid-glass dual-temporal capsule rendering `⚡ Active <relativeTime>` with live neon pulse dot, `🌱 Initiated: <formattedDate>`, and multi-day span pill | Must | ✅ | 🧪 | src/web/src/app.js, src/web/src/styles.css, src/web/index.html |
| R-051 | Session Card 1-Click UUID Copy & Layout Harmonization: Main dashboard session tiles render a subtle 1-click UUID copy button on card header; standardized metadata rows eliminate masonry baseline jitter; line-height padding fixes prompt snippet glyph slicing | Must | ✅ | 🧪 | src/web/src/app.js, src/web/src/styles.css |
| R-052 | Step Counter & Filter State Integrity: Toggling message filter chips (`💬 Dialogue`, `Prompts`, etc.) and clicking `Refresh` preserves filter context and accurately displays `Showing X of Y steps` without losing step subset telemetry | Must | ✅ | 🧪 | src/web/src/app.js |
| R-053 | Native Git Trackability Discovery & Ambient Identity Engine: Detect repository boundaries using `findGitRoot`, extract branch, dirty status, remote origin, and ambient `gh` CLI status with zero credential configuration | Must | ✅ | 🧪 | src/git_service.go, src/api.go |
| R-054 | Multi-Vector Git File Diff & History REST API: Implement `/api/git/file/status`, `/api/git/file/diff`, and `/api/git/file/commits` supporting Working Copy (unstaged), Staged Index, Commit Diffs, and file commit history with bounded timeouts (3s) and size ceilings (2MB) | Must | ✅ | 🧪 | src/api.go, src/types.go, src/git_service.go |
| R-055 | Agent Step Chrono-Diff & Hunk Extraction Engine: Extract and synthesize step-by-step file modifications (`edit_file` Target vs Replacement, `write_file` creations/overwrites) from session transcripts with agent thinking correlation | Must | ✅ | 🧪 | src/api.go, src/web/src/app.js |
| R-056 | Year 2500 Cosmic Holographic Review Pane UI: Render `#fileViewerReviewPane` on far right of `#fileViewerModal` featuring quantum telemetry ribbon (branch, churn counts, PR pill), isomorphic 4-theme styling, and vector mode selector | Must | ✅ | 🧪 | src/web/index.html, src/web/src/styles.css |
| R-057 | Interactive Draggable Splitter & Responsive Geometry: Draggable splitter (`.file-viewer-resizer`) with visual grip, mouse/touch drag, double-click snap points (50/50, 65/35, 35/65), local storage width persistence, and `⌘\` toggle shortcut | Must | ✅ | 🧪 | src/web/src/app.js, src/web/src/styles.css |
| R-058 | Dual-Mode Diff Engine (Split & Unified) with Token Highlighting: Render side-by-side or unified diff views with syntax highlighting, intra-line word/token change spans, collapsible hunks, and line number gutters | Must | ✅ | 🧪 | src/web/src/app.js, src/web/src/styles.css |
| R-059 | Bidirectional Viewport Synchronization & Anchor Navigation: Clicking diff hunk or line highlights and smoothly scrolls the code viewer to that exact line; navigating code anchors cross-highlights in review pane | Should | ✅ | 🧪 | src/web/src/app.js |
| R-060 | Ambient Remote GitHub PR & Checks Integration: Query `gh pr view` to surface active PR number, title, state, review decisions, and CI checks with 1-click external navigation to GitHub | Should | ✅ | 🧪 | src/git_service.go, src/api.go, src/web/src/app.js |
| R-061 | GitHub-Standard Dual Line Number Gutter: Render dual-column line numbers (`diff-gutter-old` vs `diff-gutter-new`) in unified view with tabular numeral alignment and blank states for additions/deletions | Must | ✅ | 🧪 | src/web/src/app.js, src/web/src/styles.css |
| R-062 | In-File Code Gutter Diff Decorations: Dynamically decorate left file viewer line gutter with high-contrast colored bars (`.gutter-diff-add`, `.gutter-diff-mod`, `.gutter-diff-del`) reflecting active diff hunks | Must | ✅ | 🧪 | src/web/src/app.js, src/web/src/styles.css |
| R-063 | Holographic Diff Minimap Rail: Mount vertical 4px heatmap rail alongside file viewer scrollbar visualizing spatial distribution of additions, modifications, and deletions across the entire scrollable file | Should | ✅ | 🧪 | src/web/src/app.js, src/web/src/styles.css |
| R-064 | Synchronized Hunk Stepper Navigation: Provide `⏮ Prev Hunk (⌥[)` and `⏭ Next Hunk (⌥])` controls that simultaneously scroll and highlight corresponding blocks in both the review pane and the source file viewer | Must | ✅ | 🧪 | src/web/src/app.js, src/web/index.html |
| R-065 | AST Semantic Symbol Breadcrumbs: Parse trailing hunk header contexts to display enclosing function, struct, or class capsules (`📦 struct Name`, `⚡ func Name`) on each hunk header | Should | ✅ | 🧪 | src/web/src/app.js, src/web/src/styles.css |
| R-066 | Surgical Hunk Patch Copy & Desktop IDE Deep-Linking: Provide 1-click `📋 Hunk` patch copy button on each hunk header and `↗ IDE` deep links (`vscode://`, `cursor://`) to open exact lines in local editors | Should | ✅ | 🧪 | src/web/src/app.js, src/web/index.html |
| R-067 | Sub-Pixel Code & Gutter Synchronous Alignment Engine: Eliminate newline joiners between block code lines, normalize Highlight.js padding to zero, lock code and gutter line heights to 22px, and guarantee 0.00px vertical drift across all file lines | Must | 🔨 | 🧪 | src/web/src/app.js, src/web/src/styles.css |
| R-068 | Forensic File Buffer Telemetry Capsule: Ambient header capsule displaying exact file line count, file size, and explicit buffer integrity (`● 100% On-Disk / Full Buffer`); documents boundary conditions (10MB HTTP ceiling, 2MB git diff ceiling) | Should | 🔨 | 🧪 | src/web/src/app.js, src/web/index.html |
| R-069 | Multi-Hunk Navigation Radar & Spatial Quick-Jump Dropdown: Interactive dropdown selector for jumping directly to any hunk among 13+ hunks with AST symbols, line offsets, and change signs | Must | 🔨 | 🧪 | src/web/src/app.js, src/web/index.html |
| R-070 | Minimap Holographic Tooltip Previews & Spatial Hover Radar: Floating acrylic tooltips on minimap hover and synchronized hover aura between gutter and code lines | Should | 🔨 | 🧪 | src/web/src/app.js, src/web/src/styles.css |
| R-071 | Active Hunk Spotlight & Ambient Framing Rail: Dynamic illuminated framing rail highlighting the active hunk's line range in the left file viewer | Should | 🔨 | 🧪 | src/web/src/app.js, src/web/src/styles.css |
| R-072 | High-Scale Viewport Virtualization & Safe Buffer Handling: Smooth 60fps scrolling and instant search across files up to 5,000 lines | Should | 🔨 | 🧪 | src/web/src/app.js |

**Priority levels**: Must (required for MVP), Should (important), Could (nice-to-have)

**Status legend**: ⬜ Not started · 🔨 In progress · ✅ Implemented · 🧪 Tested & verified · ❌ Blocked

### 2.2 Non-Functional Requirements

| ID | Requirement | Metric | Impl | Tested |
|----|------------|--------|------|--------|
| NF-001 | Performance — Initial scan | Index 1000 sessions in < 3 seconds | ✅ | 🧪 |
| NF-002 | Performance — Search | Full-text search returns results in < 200ms | ✅ | 🧪 |
| NF-003 | Performance — UI render | Dashboard loads in < 1 second after server start | ✅ | 🧪 |
| NF-004 | Portability | Single binary for macOS (arm64/amd64), Linux (amd64/arm64), Windows (amd64) | ✅ | 🧪 |
| NF-005 | Zero dependencies | Users need no runtime (no Node, no Python, no Docker) | ✅ | 🧪 |
| NF-006 | Memory | Handle 100K transcript lines with < 512MB RAM | ✅ | 🧪 |
| NF-007 | Reliability | Gracefully handle malformed JSONL, missing files, permission errors | ✅ | 🧪 |
| NF-008 | Security | Bind only to localhost (127.0.0.1), no external network exposure | ✅ | 🧪 |
| NF-009 | Startup time | Server ready in < 500ms before data scan completes | ✅ | 🧪 |

---

## 3. Interface Contract

### 3.1 CLI Interface

```
Usage: session-explorer [OPTIONS]

Options:
  -d, --data-dir <PATH>    Path to Antigravity brain directory
                            [default: ~/.gemini/antigravity-ide/brain]
  -p, --port <PORT>        Port to serve the web UI on
                            [default: 9876]
  --no-open                Don't auto-open browser
  -v, --verbose            Enable verbose logging
  -h, --help               Show help message
  --version                Show version
```

### 3.2 REST API (Internal, served by the Go backend)

#### `GET /api/sessions`

Returns list of all sessions with metadata.

**Response:**
```json
{
  "sessions": [
    {
      "id": "952f3203-4fe9-4e6c-9edb-a93984b1f3f0",
      "created_at": "2026-09-18T15:47:39Z",
      "last_message_at": "2026-09-18T16:31:30Z",
      "workspace": "/path/to/workspace",
      "project_name": "PolyTuring",
      "step_count": 169,
      "user_message_count": 6,
      "agent_response_count": 50,
      "tool_call_count": 120,
      "transcript_size_bytes": 524000,
      "first_user_prompt_preview": "I want to build a tool..."
    }
  ],
  "stats": {
    "total_sessions": 12,
    "total_user_messages": 150,
    "total_agent_responses": 400,
    "date_range": { "from": "2026-06-17", "to": "2026-09-18" },
    "workspaces": ["PolyTuring", "vadde.github.io", "DSA"]
  }
}
```

#### `GET /api/sessions/:id`

Returns full conversation timeline for a session.

**Response:**
```json
{
  "session": { "id": "...", "created_at": "...", "workspace": "..." },
  "messages": [
    {
      "step_index": 0,
      "source": "USER_EXPLICIT",
      "type": "USER_INPUT",
      "status": "DONE",
      "created_at": "2026-09-18T15:47:39Z",
      "content": "I want to build a tool...",
      "has_thinking": false,
      "tool_calls": null
    },
    {
      "step_index": 3,
      "source": "MODEL",
      "type": "PLANNER_RESPONSE",
      "status": "DONE",
      "created_at": "2026-09-18T15:48:00Z",
      "content": "Let me analyze...",
      "has_thinking": true,
      "tool_calls": [
        { "name": "list_dir", "args_preview": "{\"DirectoryPath\": \"/path...\"}" }
      ]
    }
  ]
}
```

#### `GET /api/search?q=<query>&workspace=<ws>&from=<date>&to=<date>&type=<type>`

Full-text search across all sessions.

**Response:**
```json
{
  "results": [
    {
      "session_id": "...",
      "step_index": 0,
      "type": "USER_INPUT",
      "content_preview": "...matching text...",
      "created_at": "2026-09-18T15:47:39Z",
      "workspace": "PolyTuring",
      "highlight_ranges": [[10, 25]]
    }
  ],
  "total_count": 42,
  "query": "kubernetes deployment"
}
```

#### `GET /api/wip`

Extracts recent work-in-progress (WIP) session status for a workspace or exact session ID. Implements R-030 and R-031.

**Query Parameters:**
- `workspace` (string): Workspace name, repository name, or path.
- `session_id` (string, optional): Specific session ID.
- `limit` (int, optional, default: 5): Number of candidates or steps to retrieve.
- `force_latest` (bool, optional, default: false): If true, skips multi-session halt and selects newest session.

**Response (Single Resolved Session):**
```json
{
  "status": "resolved",
  "multiple": false,
  "workspace": "PolyTuring",
  "active_session": {
    "session_id": "952f3203-4fe9-4e6c-9edb-a93984b1f3f0",
    "workspace": "PolyTuring",
    "project_name": "PolyTuring",
    "created_at": "2026-09-18T15:47:39Z",
    "last_message_at": "2026-09-18T16:31:30Z",
    "relative_time": "12 minutes ago",
    "step_count": 42,
    "user_message_count": 5,
    "error_count": 0,
    "recent_prompts": ["Add OpenAPI swagger documentation to session explorer"],
    "recent_responses": ["I have added openapi.json and embedded the swagger viewer."],
    "modified_files": ["tools/session-explorer/src/server.go", "tools/session-explorer/src/api.go"],
    "executed_commands": ["make test-tool T=session-explorer"],
    "last_tool_calls": [{"name": "replace_file_content", "args_preview": "..."}],
    "unresolved_errors": []
  }
}
```

**Response (Multiple Ambiguous Sessions — Agent Halt Invariant):**
```json
{
  "status": "ambiguous_multi_session",
  "multiple": true,
  "workspace": "PolyTuring",
  "candidate_count": 3,
  "candidates": [
    {
      "session_id": "952f3203-4fe9-4e6c-9edb-a93984b1f3f0",
      "created_at": "2026-09-18T15:47:39Z",
      "last_message_at": "2026-09-18T16:31:30Z",
      "relative_time": "12 minutes ago",
      "step_count": 42,
      "user_message_count": 5,
      "error_count": 0,
      "first_prompt_preview": "Refactor session explorer API...",
      "last_prompt_preview": "Add OpenAPI swagger documentation..."
    }
  ],
  "instruction": "MANDATORY: Multiple sessions found. STOP and prompt the user with candidates."
}
```

#### `POST /api/skill/install`

Installs the `session-wip` recovery skill and lifecycle hooks to the workspace or user global directory.

**Request Body:**
```json
{
  "target": "workspace" // or "global"
}
```

**Response:**
```json
{
  "success": true,
  "target": "workspace",
  "installed_path": "/Users/aparv/Library/.../PolyTuring/.agents/skills/session-wip",
  "message": "Skill installed successfully."
}
```

#### `GET /api/openapi.json` & `GET /docs`

Serves the offline OpenAPI 3.0 specification JSON and the embedded interactive Swagger API explorer.

#### `GET /api/file?path=<clean_path>`

Retrieves file content safely from local disk for the native in-browser file and artifact viewer (R-038).

**Query Parameters:**
- `path` (string, required): Absolute or relative filesystem path. Relative paths resolve against scanned workspace roots.

**Response:**
```json
{
  "path": "/path/to/README.md",
  "name": "README.md",
  "extension": "md",
  "language": "markdown",
  "size_bytes": 5214,
  "size_formatted": "5.1 KB",
  "line_count": 86,
  "content": "# Tutor Intelligence Challenge...",
  "is_binary": false,
  "is_image": false
}
```

**Security & Error Handling:**
- Blocks access to sensitive files (`.ssh`, `.gnupg`, `.aws`, `/etc/shadow`) with HTTP 403 Forbidden.
- Limits file size to 10MB ceiling with HTTP 413 Payload Too Large.
- Returns HTTP 400 for directories or missing path, and HTTP 404 for nonexistent files.

#### `GET /api/agent/context/global`

Discovers and returns machine-wide agent context (R-039), including global rules (`~/.gemini/config/rules/`), global skills (`~/.gemini/config/skills/`), builtin runbooks, plugins, MCP tool configurations, and lifecycle hooks.

#### `GET /api/agent/context/workspace?workspace=<name|path>`

Discovers workspace-specific agent guidelines (`.agents/rules/`, `AGENTS.md`, `CLAUDE.md`, `.cursorrules`), repo-local skills, hooks, and merges them with inherited global context with explicit scope tagging (`workspace` vs `global` vs `builtin`).

#### `GET /api/sessions/{id}/context`

Analyzes and extracts the complete empirical behavioral matrix of a specific session (R-040), tracking:
- **Active vs Available Skills**: Execution counts, step markers, injected tokens, and progressive disclosure footprint.
- **Empirical MCP Telemetry**: Discovered and ad-hoc MCP servers, active vs dormant tools, invocation frequencies, step index spans, and arguments insights.
- **Native Agentic Tools**: Breakdown of filesystem, terminal, subagent, and search executions.
- **Analytical Telemetry**: Tool Diversity Index ($D = U / N$), total calls, and error counts.
- **Active Rules & Governance**: Constitutional invariants and workspace rules active in this session.


### 3.3 Data Source Specification

| Field | Type | Source |
|-------|------|--------|
| Session ID | UUID string | Directory name under `brain/` |
| Transcript | JSONL file | `brain/<id>/.system_generated/logs/transcript.jsonl` |
| Full Transcript | JSONL file | `brain/<id>/.system_generated/logs/transcript_full.jsonl` |
| Trajectory Metadata | SQLite DB | `conversations/<id>.db` (`trajectory_metadata_blob` table: git repo, branch, workspace URI) |
| Workspace / Project | string (path / name) | Multi-tier clustering engine: Authoritative DB → Genesis Workspace Invariant → Empirical Tool Touches (Mutations > Reads > Exec Cwd) → Folder Name → "Default" |
| Git Repo & Branch | string | Extracted from SQLite `trajectory_metadata_blob` or filesystem `.git` traversal |

**Clustering & Identity Hierarchy (R-028 Invariant):**
1. **Tier 0 (Authoritative Trajectory / Genesis Metadata)**: Probe SQLite `trajectory_metadata_blob` and user info for `WorkspaceURI` and `GitRepo`.
2. **Genesis Non-Git Invariant**: When a session is initiated in the user's home directory (`~`), root, or a generic parent container without an authoritative Git repository, the session **MUST** remain clustered under `"Default"` unless the agent performed substantive empirical file mutations (`write_file`, `edit_file`) in a specific repository during that session.
3. **Command Touch Sanitization**: `run_command` only constitutes an empirical tool touch if explicit directory execution (`Cwd`, `Directory`) or directory navigation (`cd <path>`, `git -C <path>`) is present. Blanket regex scraping of arbitrary command strings (especially inside inline `-c` / `-e` interpreter code or `echo` statements) is strictly prohibited to prevent mock test scripts from hijacking session identity.
4. **Runtime Artifact Exclusion Boundary**: Skip internal runtime stores (`/brain/`, `/conversations/`, `/logs/`, `/presence/`, `/annotations/`, `.system_generated`, `.user_uploaded`), while allowing user code and configuration files (e.g. `~/.gemini/antigravity-cli/statusline.py`) to be recognized.
5. **Project Hub "Default" Aggregation**: When sessions exist with `ProjectName == "Default"`, aggregate them into a `ProjectSummary` with `is_git: false`, path `~`, and render a dedicated project card on Project Hub with non-git badge styling (`folder` / `ad-hoc`).

**JSONL Entry Schema (per line):**

| Field | Type | Description |
|-------|------|-------------|
| `step_index` | int | Sequential index of the step |
| `source` | string | `USER_EXPLICIT`, `MODEL`, `SYSTEM` |
| `type` | string | `USER_INPUT`, `PLANNER_RESPONSE`, `VIEW_FILE`, `RUN_COMMAND`, `GREP_SEARCH`, `LIST_DIRECTORY`, `CODE_ACTION`, `SEARCH_WEB`, `READ_URL_CONTENT`, `BROWSER_SUBAGENT`, `ASK_QUESTION`, `CONVERSATION_HISTORY`, `KNOWLEDGE_ARTIFACTS`, `CHECKPOINT`, `EPHEMERAL_MESSAGE`, `ERROR_MESSAGE`, `GENERIC`, `SYSTEM_MESSAGE` |
| `status` | string | `DONE`, `ERROR` |
| `created_at` | ISO 8601 | Timestamp |
| `content` | string | Text content (may contain XML tags for USER_INPUT) |
| `thinking` | string | Agent's reasoning (PLANNER_RESPONSE only) |
| `tool_calls` | array | Tool invocations (PLANNER_RESPONSE only) |
| `is_truncated` | boolean | Whether content was truncated in compact transcript |

### 3.4 Error Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | Data directory not found or not readable |
| 4 | Port already in use |

---

## 4. Constraints

### 4.1 Technical Constraints

- Must compile to a single static binary via Go (v1.22+)
- Frontend assets must be embedded via Go's `embed` package — no external file dependencies
- Must use only the Go standard library + minimal well-maintained dependencies
- HTML/CSS/JS frontend — vanilla JavaScript (no React/Vue/Angular build step)

### 4.2 Security Constraints

- HTTP server MUST bind to `127.0.0.1` only (localhost), never `0.0.0.0`
- Must not write to or modify any transcript files (read-only access)
- Must not transmit any data over the network
- Must sanitize rendered content to prevent XSS

### 4.3 Performance Constraints

- Initial page load: < 1 second
- Session scan: < 3 seconds for 1000 sessions
- Search latency: < 200ms for full-text search across all transcripts
- Memory: < 512MB for 100K transcript lines
- Binary size: < 20MB

---

## 5. Dependencies

### 5.1 External Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| Go standard library | 1.22+ | HTTP server, JSON parsing, embed, filepath, os |
| `github.com/pkg/browser` | latest | Cross-platform browser opening (optional, can use `exec.Command`) |

### 5.2 Internal Dependencies

| Tool | Purpose |
|------|---------|
| None | This is a standalone tool |

---

## 6. Acceptance Criteria

Tests that MUST pass for this spec to be considered satisfied:

| AC ID | Criteria | Traces To |
|-------|----------|-----------|
| AC-001 | Given a valid brain directory, when session-explorer runs, then all sessions are discovered and listed | R-001 |
| AC-002 | Given transcript.jsonl files, when parsed, then user messages and agent responses are correctly extracted with timestamps | R-002, R-003, R-004 |
| AC-003 | Given USER_INPUT entries with ADDITIONAL_METADATA, when parsed, then workspace/project paths are extracted | R-005 |
| AC-004 | Given the server starts, then it binds to localhost on the configured port and serves the web UI | R-006, R-008, NF-008 |
| AC-005 | Given the `--port` flag, when specified, then the server uses the custom port | R-024 |
| AC-006 | Given the `--data-dir` flag, when specified, then the custom path is used for session discovery | R-023 |
| AC-007 | Given the dashboard loads, then all sessions are displayed as cards with date, workspace, step count | R-008 |
| AC-008 | Given a session card is clicked, then the full conversation timeline is rendered with user/agent messages | R-009 |
| AC-009 | Given a date range filter is applied, then only sessions within that range are shown | R-010 |
| AC-010 | Given a workspace filter is selected, then only sessions from that workspace are shown | R-011 |
| AC-011 | Given a search query, then matching prompts across all sessions are returned with highlights | R-012 |
| AC-012 | Given the UI renders, then glassmorphic/acrylic styling is applied with smooth animations | R-018 |
| AC-013 | Given malformed JSONL lines, when parsing, then they are skipped gracefully without crashing | NF-007 |
| AC-014 | Given the server starts, then the default browser opens automatically (unless `--no-open`) | R-007 |
| AC-015 | Given the UI renders on mobile viewport, then layout adapts responsively | R-019 |
| AC-016 | Given user triggers export, conversation can be saved as Markdown or JSON | R-021 |
| AC-017 | Given user groups sessions by project, sessions are clustered by repository/working directory | R-028, R-029 |
| AC-018 | Given a workspace with a single recent session, WIP digest returns active document and recent context | R-030 |
| AC-019 | Given a workspace with multiple recent sessions, WIP endpoint halts with candidate list (Invariant R-031) | R-031 |
| AC-020 | Given OpenAPI endpoint /api/openapi.json and /docs, Swagger UI renders completely offline | R-032 |
| AC-021 | Given install skill modal or make setup-session-wip, session-wip skill is installed to target | R-033 |
| AC-022 | Given headless CLI session-explorer wip, JSON or human summary is emitted | R-034 |
| AC-023 | Given the Responses filter is clicked, only agent turns without tools (conversational prose) are displayed | R-035 |
| AC-024 | Given the Dialogue mode is clicked, only user prompts and agent responses without tools are shown | R-035 |
| AC-025 | Given real-time SSE live tailing, incoming steps respect active message filters and stage seamlessly | R-035 |
| AC-026 | Given multiple Antigravity brain directories (~/.gemini/antigravity-ide/brain, ~/.gemini/antigravity-cli/brain), sessions from all runtimes are auto-discovered concurrently | R-036 |
| AC-027 | Given discovered sessions, each session is attributed with provider and runtime_mode (ide, cli, desktop, core) and rendered with distinct visual badges | R-036 |
| AC-028 | Given a session detail request or live SSE stream, engine dynamically resolves transcript path from the session's originating brain path | R-036 |
| AC-029 | Given extensible SessionProvider interface, architecture supports plugging in Claude Code and GitHub Copilot providers without core scanner rewrites | R-036 |
| AC-030 | Given an inline Mermaid diagram or chart, canvas viewport is bounded to max-height 440px with internal scrolling, preventing vertical scrolling trapping in conversation timeline | R-037 |
| AC-031 | Given a diagram card, clicking the "Theater" button or double-clicking the canvas opens fullscreen distraction-free Theater mode with pan, drag, zoom (20%-500%), and SVG/source copy controls | R-037 |
| AC-032 | Given file paths, markdown links (`[text](file://...)`), or tool parameters (`AbsolutePath`, `TargetFile`), user clicking the element opens the file natively inside the browser | R-038 |
| AC-033 | Given a file loaded in File Viewer, viewer renders language badge, total line count, human-readable size, line number gutter, and highlights anchor lines (`#L10-L20` or `:10`) | R-038 |
| AC-034 | Given a Markdown file in File Viewer, viewer provides segmented toggle between rich rendered Preview (supporting KaTeX and Mermaid) and syntax-highlighted Raw mode | R-038 |
| AC-035 | Given a session with tool calls and markdown references, detail view aggregates all files into an interactive Files & Artifacts drawer categorized by Artifact, Modified, and Read | R-038 |
| AC-036 | Given File Viewer is active, Cmd+F / Ctrl+F opens the in-file Find widget pinned to the outer shell (remaining 100% visible across all vertical scroll positions) with match counter (X of Y), case toggle (Aa), next/previous navigation (Enter/Shift+Enter, arrows), and syntax-preserving tree highlights | R-038 |
| AC-037 | Given a Markdown file in File Viewer, the preview container dynamically expands to 100% pane width, and mode toggling guarantees strict visual separation without raw content leakage | R-038 |
| AC-038 | Given a syntax-highlighted code block with line number gutter, vertical scrolling is unified into a single scrollbar on the viewer body, eliminating nested double scrollbars | R-038 |
| AC-039 | Given file references on the session trail, the UI renders a dual-action chip with an immediate 1-click copy path button (`📋` → `✓`) alongside the open-file pill, and viewer action buttons provide instant micro-feedback (`✓ Copied!`, `✓ Path Copied!`, `✓ Wrapped`) | R-038 |
| AC-040 | Given the top header bar, user clicking `🧠 Agent Brain` opens the Agent Context Matrix modal displaying Global and Workspace rules, skills, plugins, MCP servers, and hooks with instant search and precedence badges | R-039 |
| AC-041 | Given a workspace or repository filter, the engine computes the hierarchical cascade (Workspace rules overriding or extending Global rules, workspace skills augmenting global/builtin skills) | R-039 |
| AC-042 | Given a session detail view, the sticky header renders an Agent Context button (`🧠 Agent Context (<count>)`); clicking opens a drawer showing active rules, invoked skills (with step counts), and on-demand skills | R-039 |
| AC-043 | Given any rule or skill item, clicking `Inspect` opens its full content directly in the embedded Native File Viewer (R-038) with syntax highlighting, Markdown preview, and line numbers | R-039 |
| AC-044 | Given multiple assistant ecosystems, the backend provider architecture abstracts rule and skill discovery across Antigravity, Claude Code, GitHub Copilot, and Cursor | R-039 |
| AC-045 | Given rules and skills in the Agent Brain modal or session context drawer, the system computes and renders GenAI bloat telemetry including line count, word count, character count, and estimated tokens ($T \approx \max(\lceil C/4 \rceil, \lceil W \times 1.33 \rceil)$), distinguishing pre-loaded context bloat from on-demand footprint | R-039 |
| AC-046 | Given a session with tool calls, requesting `GET /api/sessions/:id/context` returns structured `mcp_servers` with `invoked_count`, `active_tools`, `dormant_tools`, and `native_tools` breakdown | R-040 |
| AC-047 | Given a transcript containing lazy MCP tool calls (`call_mcp_tool`) or eager calls (`mcp_*`), the extraction engine accurately identifies server names, tool names, invocation counts, first/last step indices, and recent arguments | R-040 |
| AC-048 | Given native tool executions (`read_file`, `write_file`, `run_command`, `invoke_subagent`, etc.), the engine categorizes each tool, computes execution counts, step spans, and detects error states | R-040 |
| AC-049 | Given the session detail view, the sticky header renders isomorphic Holographic Aero-Mica action capsules for `Agent Context` and `Files & Artifacts` featuring illuminated color-coded digital counter badges matching behavioral matrix palettes; clicking opens the Holographic Session Matrix or Files Drawer | R-040 |
| AC-050 | Given the Session Agent Matrix modal, the user can search rules, skills, MCP servers, and tools via real-time search input or filter chips (`All`, `⚡ Skills`, `🔌 MCP Tools`, `🛠️ Native Tools`, `📜 Rules`, `🤖 Subagents`) | R-040 |
| AC-051 | Given the Session Agent Matrix modal, the Telemetry Ribbon displays `📌 Active Injected Tokens`, `🔌 MCP Calls`, `🛠️ Tool Diversity Index` ($D = U/N$), and `💤 Dormant Bloat` | R-040 |
| AC-052 | Given any tool step pill (`#X`), clicking navigates smoothly to that step in the main conversation timeline with an animated neon pulse highlight; clicking `📖 Inspect` on any rule/skill opens the file in the embedded Native File Viewer | R-040 |
| AC-053 | Given a session with `run_command` executions, the engine decomposes them by binary (`git`, `make`, `go`, `bun`, `curl`, `kill`) and category (`VCS`, `Build`, `Test`, `Process`, `Inspect`, `Network`, `Runtime`, `System`) | R-041 |
| AC-054 | Given a session with filesystem tool calls (`read_file`, `write_file`, `edit_file`), the engine decomposes operations by target file extension (`.go`, `.js`, `.css`, `.md`, `.json`) and tracks top files touched with step indices | R-041 |
| AC-055 | Given the `run_command` card in the Session Agent Matrix, the UI renders a Year 2126 Holographic Command Spectrum bar showing color-coded proportions and an interactive sub-command breakdown drawer with percentages and counts | R-041 |
| AC-056 | Given any command or file sub-breakdown item, clicking a step pill (`#38`) smoothly jumps to that exact step in the transcript timeline and highlights it | R-041 |
| AC-057 | Given the filter chips bar in `#sessionContextModal`, chips wrap dynamically into a non-scrollable flex matrix with illuminated color-coded counter badges (`.matrix-chip-counter`) | R-041 |
| AC-058 | Given the Holographic Telemetry & Audit tab (Lens 3), the UI renders a dedicated Command Execution Spectrum and Sub-Tool Distribution breakdown | R-041 |
| AC-059 | Given a session transcript, the engine detects Tier 1 skill invocations across all tool calls (`read_file`, `view_file`, `edit_file`, `run_command`) and populates `step_indices`, `last_step_idx`, and `invocation_triggers` | R-042 |
| AC-060 | Given skills discovered across multiple scopes (workspace, plugin, global, builtin), skills are cleanly deduplicated by Name prioritizing workspace over global/builtin | R-042 |
| AC-061 | Given the `#sessionContextModal`, the filter chip dynamically labels `Invoked Skills (${activeCount})` in Active lens and `Skills (${activeCount} active / ${totalCount} total)` in Cascade lens | R-042 |
| AC-062 | Given the Session Agent Matrix modal, `skill_active` and `skill_dormant` render full-parity Holographic Skill Cards matching MCP and Native cards | R-042 |
| AC-063 | Given an active skill card, the UI renders an illuminated `⚡ Nx` count badge, clickable `#X` step pills with timeline navigation, and trigger breakdown badges | R-042 |
| AC-064 | Given a standby/dormant skill card in Cascade lens, the UI renders a `💤 Standby` badge, token bloat metric, line count, and 1-click `📖 Inspect` in embedded File Viewer | R-042 |
| AC-065 | Given the enhanced skill models, all existing unit tests pass and new test `TestR042_ExtractSessionAgentContext_SkillBreakdownAndDeduplication` passes with 100% assertions | R-042 |
| AC-066 | Given the `#sessionContextModal` and `#agentMatrixModal`, the modal viewport expands to widescreen panoramic dimensions (`width: min(1560px, 94vw); max-width: 1560px; max-height: 90vh;`), utilizing available desktop real estate | R-043 |
| AC-067 | Given $N$ variable-height agent cards, the frontend executes an EC2-style greedy bin-packing algorithm ($\min_k \sum h$) across $K$ column bins (3 columns on desktop >1200px, 2 columns on tablet, 1 on mobile) | R-043 |
| AC-068 | Given the bin-packed layout, cards flow continuously down each column track with zero empty vertical gaps beneath shorter cards, eliminating dead grid row space | R-043 |
| AC-069 | Given an expanding card (e.g. `run_command` drawer toggle), only its column track stretches smoothly without displacing or stretching adjacent column cards | R-043 |
| AC-070 | Given window resizing or filter changes, the bin-pack layout dynamically re-computes column assignments and updates the DOM in <16ms (60 FPS) | R-043 |
| AC-071 | Given live SSE streaming (`/api/sessions/:id/tail`), incoming steps dynamically re-evaluate and uptick the Files & Artifacts counter badge (`#detailFilesCountBadge`) and re-render the open files modal in real-time | R-044 |
| AC-072 | Given live SSE streaming, incoming steps trigger debounced background sync of `/api/sessions/:id/context`, dynamically upticking the Agent Context badge (`#detailAgentContextCount`) and live-updating the open Session Matrix modal | R-044 |
| AC-073 | Given any count update on the Agent Context or Files & Artifacts capsules, the badge triggers a tactile holographic pulse animation (`.holo-btn-counter.updated`) | R-044 |
| AC-074 | Given the Session Trail toolbar (`.tail-stream-control`), a dedicated Refresh button (`#refreshTrailBtn`) allows on-demand non-destructive reload of all steps, agent context, and files/artifacts without full browser reload | R-044 |
| AC-075 | Given the user resumes live streaming or toggles Follow Live, the engine triggers an automatic catch-up refresh of session trail messages, agent context, and files/artifacts | R-044 |
| AC-076 | Given active turn filters (e.g. Dialogue, Prompts, Search query), live SSE incoming turns and session trail refreshes strictly preserve the filter slice and maintain the 'Showing X of Y steps' counter summary without wiping filter telemetry | R-044 |
| AC-077 | Given the Files & Artifacts button (`#detailFilesBtn`) is clicked, the modal opens in widescreen panoramic dimensions (`width: min(1560px, 94vw); max-width: 1560px; max-height: 90vh;`) with liquid glass backdrop blur, glowing icon pulse, and ESC/close controls | R-045 |
| AC-078 | Given `#sessionFilesModal` is open, user can switch between 4 forensic lenses: `🗂️ Universal Catalog`, `✨ Artifacts & Deliverables`, `✏️ Codebase Mutations`, and `📊 Churn & Blast Radius Telemetry` | R-045 |
| AC-079 | Given session file data, the ribbon renders 5 color-coded metric capsules: `✨ Deliverables`, `✏️ Code Mutations`, `👁️ Inspected Files`, `🎯 Blast Radius`, and `📈 Reconnaissance Ratio ($R_{\text{RW}}$)` | R-045 |
| AC-080 | Given the active lens, filter chips allow 1-click filtering: `🌐 All Files`, `✨ Artifacts`, `✏️ Modified Code`, `👁️ Read Files`, `🔗 Markdown References`, and dynamic language chips (`.go`, `.js`, etc.) with counters | R-045 |
| AC-081 | Given $N$ variable-height forensic file cards, cards are distributed across 3 columns (desktop), 2 columns (tablet), or 1 column (mobile) using greedy bin-packing with zero vertical voids | R-045 |
| AC-082 | Given any file card with step markers (`#X`), clicking a step pill closes the files modal, scrolls smoothly to that exact turn in the conversation timeline, and triggers an animated neon pulse highlight | R-045 |
| AC-083 | Given any file card, clicking `Open ↗` opens the target file inside the embedded Native File Viewer (R-038) with syntax highlighting, line gutter, and anchor scrolling; clicking `📋` copies the path with micro-feedback | R-045 |
| AC-084 | Given generated artifacts, the card renders artifact summaries, user-facing delivery badges, and file preview anchors | R-045 |
| AC-085 | Given real-time SSE streaming or manual Session Trail refresh (`#refreshTrailBtn`), incoming steps dynamically re-aggregate forensic file data, update the `#detailFilesCountBadge` counter with pulse animations, and re-render the open files matrix in-place with zero layout shift | R-045 |
| AC-086 | Given `#sessionFilesModal` is open, the toolbar renders `.agent-matrix-sort-box` with select dropdown `#sessionFilesSortSelect` offering 7 sorting dimensions | R-046 |
| AC-087 | Given user selects `🔥 Churn Intensity (High → Low)`, files are sorted by mutation count descending ($C(f)$), placing heavily edited files at the top of the bin-packed columns | R-046 |
| AC-088 | Given user selects `⏮️ First Touched (Chronological)`, files are sorted by their first appearance turn index ($s_{\text{first}}$) from earliest to latest | R-046 |
| AC-089 | Given user selects `🏷️ File Extension`, files are clustered alphabetically by file extension (.go, .js, .md, .py) and then by name | R-046 |
| AC-090 | Given `#sessionContextModal` is open, the toolbar renders `#sessionMatrixSortSelect` offering execution frequency, step recency, token bloat, precedence, and alphabetical sorting | R-046 |
| AC-091 | Given `#agentMatrixModal` is open, the toolbar renders `#agentMatrixSortSelect` offering token bloat, line count, scope, and name sorting | R-046 |
| AC-092 | Given any active filter chip or search query, changing the sort selection preserves active filter slices, re-computes the bin-pack layout in < 16ms (60 FPS), and maintains zero vertical gaps | R-046 |
| AC-093 | Given file cards in `#sessionFilesModal`, each card displays a subtle monospace touch span badge (`Span: #X ➔ #Y` or turn count) indicating the step range across which the file was accessed | R-046 |
| AC-094 | Given `#detailView` navigation header, clicking the session ID capsule or copy button copies the full session UUID to clipboard and displays instant visual confirmation (`✓ Copied!`) and toast notification | R-047 |
| AC-095 | Given `#copySessionIdBtn`, the button renders a Year 2500 aero-mica glass capsule with smooth hover elevation and tactile click pulse | R-047 |
| AC-096 | Given any session card on the main dashboard or in project drilldown, the card renders both **⚡ Last Active** recency and **🌱 Inception Date** | R-048 |
| AC-097 | Given a session active within the last 3 hours, the card displays an illuminated neon green/amber pulse dot indicating live or ongoing recency | R-048 |
| AC-098 | Given a session spanning $\ge 24\text{ hours}$ between inception and last message, the card displays a monospace Multi-Day Lifespan badge indicating duration and boundary dates | R-048 |
| AC-099 | Given Date Grouped View (`renderDateGroupedView`), sessions are bucketed by effective recency ($\max(t_{\text{active}}, t_{\text{init}})$), ensuring sessions active today appear under `🌟 Today` | R-048 |
| AC-100 | Given `#sortBySelect` and `#projectSortSelect`, the dropdowns offer `⚡ Last Active (Recent First)` as the primary sort option alongside Inception Date options | R-048 |
| AC-101 | Given `GET /api/sessions?sort_by=active`, the Go backend sorts sessions by `LastMessageAt` descending | R-048 |
| AC-102 | Given date range filters (`from`/`to` or presets `today`, `7d`, `30d`), multi-day sessions touched within the window are correctly surfaced | R-048 |
| AC-103 | Given all automated tests and Playwright headless runs, verification passes with 100% assertions and zero visual regressions | R-047, R-048 |
| AC-104 | Given `/api/sessions`, the endpoint accepts `date_mode` query parameter supporting `active` (default), `created`, and `overlap` | R-049 |
| AC-105 | Given `date_mode=active`, sessions match if `last_message_at >= from && last_message_at <= to` (falling back to `created_at` for single-turn sessions) | R-049 |
| AC-106 | Given `date_mode=created`, sessions match if `created_at >= from && created_at <= to` | R-049 |
| AC-107 | Given `date_mode=overlap`, sessions match if `last_message_at >= from && created_at <= to` | R-049 |
| AC-108 | Given global date filter toolbar, the UI renders a clean, uncluttered layout `Date: [From] to [To] [Presets]` without intrusive sandwiched buttons, seamlessly delegating sort order to `#sortBySelect` | R-049 |
| AC-109 | Given in-project drilldown toolbar (`getFilteredAndSortedProjSessions`), temporal date filtering mirrors tri-modal date logic seamlessly | R-049 |
| AC-110 | Given Session Trail navigation header, `#detailTime` is upgraded to `.detail-temporal-capsule` rendering `⚡ Active <relativeTime>` with neon pulse dot, `🌱 Initiated: <formattedDate>`, and multi-day span pill | R-050 |
| AC-111 | Given any element in the detail temporal capsule, hovering surfaces a high-precision ISO-8601 tooltip | R-050 |
| AC-112 | Given any session card on the main dashboard grid, the card header renders a subtle 1-click UUID copy button (`.card-copy-btn`) | R-051 |
| AC-113 | Given user clicks `.card-copy-btn`, the full UUID is copied to clipboard with `✓` feedback without triggering card drilldown navigation | R-051 |
| AC-114 | Given session cards grid, multi-day span pills are placed in a standardized metadata row to ensure aligned prompt snippet baselines across columns | R-051 |
| AC-115 | Given prompt snippets in session cards, line-height padding prevents vertical glyph clipping on multi-line text | R-051 |
| AC-116 | Given Session Trail step filters, toggling message types and refreshing the trail maintains `Showing X of Y steps` state integrity | R-052 |
| AC-117 | When `GET /api/git/file/status?path=<path>` is called for a git-tracked file, it returns `is_git: true`, `git_root`, `branch`, `is_dirty`, and `remote_url` | R-053 |
| AC-118 | When called for a non-git file, it returns `is_git: false` without error, allowing the UI to gracefully switch to Agent Step Chrono-Diff mode | R-053 |
| AC-119 | The backend discovers machine-level `gh` CLI binary and checks authentication status (`gh auth status`) with zero user credentials required | R-053 |
| AC-120 | `GET /api/git/file/diff?path=<path>&mode=unstaged` returns the unstaged working copy diff with lines added, deleted, and structured hunks | R-054 |
| AC-121 | `GET /api/git/file/diff?path=<path>&mode=staged` returns the staged git index diff | R-054 |
| AC-122 | `GET /api/git/file/diff?path=<path>&mode=commit&commit_sha=<sha>` returns the exact diff introduced by that commit for the file | R-054 |
| AC-123 | `GET /api/git/file/commits?path=<path>&limit=30` returns up to 30 chronological commits touching the file, each with SHA, short SHA, author, timestamp, and message | R-054 |
| AC-124 | If a git command takes longer than 3.0 seconds, the context cancels and returns HTTP 504 Gateway Timeout without hanging the server | R-054 |
| AC-125 | Diff outputs exceeding 2MB are cleanly truncated with a notice capsule to preserve browser performance | R-054 |
| AC-126 | For any session touching a file via `edit_file`, the system synthesizes a diff hunk where `Target` represents deleted lines (`-`) and `Replacement` represents added lines (`+`) | R-055 |
| AC-127 | For `write_file` tool calls, the system detects whether it was an overwrite or append and creates corresponding addition hunks | R-055 |
| AC-128 | Each agent step diff displays the step index, timestamp, tool call name, and an expandable capsule containing the agent's internal thought (`thinking`) | R-055 |
| AC-129 | The UI provides a Step Timeline Scrubber allowing the user to step forward and backward across all agent turns touching that file | R-055 |
| AC-130 | `#fileViewerReviewPane` mounts inside `.file-viewer-shell` to the right of `.file-viewer-body` | R-056 |
| AC-131 | The Quantum Telemetry Ribbon displays current branch (`🌿 <branch>`), dirty status dot, net churn (`+A / -B`), and active vector mode pill | R-056 |
| AC-132 | The Review Pane adheres to all 4 PolyTuring themes (Liquid Glass, Fluent Aero Mica, Holo Cyberpunk, Matrix Obsidian) with zero CSS contrast defects | R-056 |
| AC-133 | The draggable resizer (`.file-viewer-resizer`) allows dynamic horizontal resizing between 25% and 75% of viewport width with smooth 60fps tracking | R-057 |
| AC-134 | Double-clicking the resizer cycles through snap points: 50% $\to$ 65% $\to$ 35% | R-057 |
| AC-135 | Resizer position persists in `localStorage` across page reloads and modal re-openings | R-057 |
| AC-136 | Pressing `⌘\` (Mac) or `Ctrl+\` (Windows/Linux) or clicking the topbar review button toggles review pane collapse/expand | R-057 |
| AC-137 | User can switch between Side-by-Side (Split) and Unified (Inline) diff views via segmented control | R-058 |
| AC-138 | Code lines in diff blocks inherit Highlight.js syntax highlighting matching the file extension | R-058 |
| AC-139 | Modified lines highlight sub-line changed tokens/words using intra-line Levenshtein/character diffing | R-058 |
| AC-140 | Clicking a line number in the diff view scrolls and highlights the corresponding line in the left code viewer | R-059 |
| AC-141 | When an active PR exists for the branch, the Review Pane renders a GitHub PR capsule showing PR number, title, review decision, and CI checks | R-060 |
| AC-142 | Clicking `↗ GitHub` opens the commit or PR directly on github.com in a new tab | R-060 |
| AC-143 | Unified diff rows render two distinct gutter columns (`diff-gutter-old` and `diff-gutter-new`) | R-061 |
| AC-144 | On additions, `diff-gutter-old` is blank and `diff-gutter-new` renders the line number | R-061 |
| AC-145 | On deletions, `diff-gutter-new` is blank and `diff-gutter-old` renders the line number | R-061 |
| AC-146 | On context lines, both old and new columns render their respective numbers | R-061 |
| AC-147 | Numbers in gutters use tabular numeric alignment without horizontal jitter | R-061 |
| AC-148 | When a diff is rendered, the left file viewer gutter receives `.gutter-diff-add` for added lines | R-062 |
| AC-149 | When a diff is rendered, modified lines receive `.gutter-diff-mod` | R-062 |
| AC-150 | Deletion points render a rose marker `.gutter-diff-del` indicating insertion coordinates | R-062 |
| AC-151 | Switching vectors (`⚡ Agent Steps` $\to$ `📝 Working Copy`) re-calculates file gutter decorations | R-062 |
| AC-152 | Closing review pane or viewing a clean file removes all in-file diff decorations | R-062 |
| AC-153 | Clicking a gutter decoration in the file viewer scrolls the review pane to that hunk | R-062 |
| AC-154 | The vertical `.file-diff-minimap-rail` renders color-coded ticks corresponding to changed lines | R-063 |
| AC-155 | Minimap ticks are calculated via percentage coordinates based on total file line count | R-063 |
| AC-156 | Clicking a minimap tick smoothly scrolls the file viewer to that line | R-063 |
| AC-157 | Topbar renders `⏮ Prev Hunk` and `⏭ Next Hunk` buttons and `Hunk X of Y` indicator | R-064 |
| AC-158 | Keyboard shortcuts `⌥[` and `⌥]` trigger previous and next hunk stepping | R-064 |
| AC-159 | Stepping to a hunk simultaneously centers both the review hunk and the file viewer line | R-064 |
| AC-160 | Both target elements display a temporary animated cosmic pulse aura (`jump-pulse`) | R-064 |
| AC-161 | Trailing symbol contexts in hunk headers are parsed into AST Scope Capsules | R-065 |
| AC-162 | AST capsules display appropriate glyphs (`📦` struct, `⚡` func) | R-065 |
| AC-163 | Each hunk header renders an individual `📋 Hunk` button | R-066 |
| AC-164 | Clicking `📋 Hunk` copies only that specific hunk as a valid unified diff patch | R-066 |
| AC-165 | `↗ IDE` action generates functional `vscode://` and `cursor://` deep links | R-066 |
| AC-166 | Rendered code lines in `symbol_table.rs` (399 lines) MUST NOT contain empty blank lines between code lines | R-067 |
| AC-167 | Line 1 of code and Line 1 of gutter MUST have identical vertical bounding client top: $\|y_{code}(1) - y_{gutter}(1)\| = 0.00\text{px}$ | R-067 |
| AC-168 | Line 50 of code and Line 50 of gutter MUST have identical vertical bounding client top: $\|y_{code}(50) - y_{gutter}(50)\| = 0.00\text{px}$ | R-067 |
| AC-169 | Line 200 of code and Line 200 of gutter MUST have identical vertical bounding client top: $\|y_{code}(200) - y_{gutter}(200)\| = 0.00\text{px}$ | R-067 |
| AC-170 | Line 399 of code and Line 399 of gutter MUST have identical vertical bounding client top: $\|y_{code}(399) - y_{gutter}(399)\| = 0.00\text{px}$ | R-067 |
| AC-171 | `code.hljs` MUST NOT have default Highlight.js padding pushing line 1 down | R-067 |
| AC-172 | Toggling line wrapping on (`wrap: true`) MUST invoke `syncWrappedLineHeights()` and adapt gutter span heights | R-067 |
| AC-173 | Toggling line wrapping off (`wrap: false`) MUST reset gutter span heights back to standard 22px | R-067 |
| AC-174 | Header displays `#fileViewerBufferBadge` with `● 100% On-Disk` for files read from disk | R-068 |
| AC-175 | Tooltip on buffer badge displays exact line count, file size in KB/MB, and confirms no truncation occurred | R-068 |
| AC-176 | In `extractAgentStepDiffsForFile`, `write_file` agent step previews are NOT artificially limited to 300 lines | R-068 |
| AC-177 | Review pane header displays `#hunkRadarSelect` dropdown alongside `<` and `>` buttons | R-069 |
| AC-178 | For `symbol_table.rs` commit diff (13 hunks), `#hunkRadarSelect` lists all 13 hunks with their AST symbols and lines | R-069 |
| AC-179 | Selecting an item from `#hunkRadarSelect` immediately steps to that hunk and smoothly scrolls the left file viewer | R-069 |
| AC-180 | Clicking `<` or `>` or pressing `⌥[` / `⌥]` updates `#hunkRadarSelect.value` to reflect active hunk index | R-069 |
| AC-181 | Hovering over a tick in `#fileDiffMinimapRail` displays an informative tooltip preview | R-070 |
| AC-182 | Hovering over a gutter line number adds `.line-hover-sync` to highlight the corresponding code line | R-070 |
| AC-183 | When a hunk is selected, lines in the left file viewer belonging to that hunk receive `.hunk-active-spotlight` | R-071 |
| AC-184 | When stepping to the next hunk, previous spotlight classes are cleanly removed | R-071 |
| AC-185 | Files up to 5,000 lines render in `< 50ms` and scroll smoothly at 60fps | R-072 |

---

## 7. Design Notes

### 7.1 Architecture

```
┌──────────────────────────────────────────────────────┐
│                    session-explorer                    │
│                                                        │
│  ┌──────────────┐    ┌──────────────┐                │
│  │   CLI Layer   │───▶│  HTTP Server │                │
│  │  (flags/args) │    │ (net/http)   │                │
│  └──────────────┘    └──────┬───────┘                │
│                             │                          │
│           ┌─────────────────┼─────────────────┐       │
│           │                 │                 │       │
│    ┌──────▼──────┐  ┌──────▼──────┐  ┌──────▼────┐  │
│    │  REST API   │  │  Static     │  │ WebSocket  │  │
│    │  /api/*     │  │  Assets     │  │ (optional) │  │
│    └──────┬──────┘  │  (embed)    │  └────────────┘  │
│           │         └─────────────┘                   │
│    ┌──────▼──────────────────────┐                    │
│    │     Session Index Engine    │                    │
│    │  ┌────────────────────────┐ │                    │
│    │  │ JSONL Parser (scanner) │ │                    │
│    │  │ In-Memory Search Index │ │                    │
│    │  │ Workspace Extractor    │ │                    │
│    │  └────────────────────────┘ │                    │
│    └──────┬──────────────────────┘                    │
│           │                                            │
│    ┌──────▼──────────────────────┐                    │
│    │   ~/.gemini/antigravity-ide │                    │
│    │   /brain/<id>/transcript.*  │                    │
│    └─────────────────────────────┘                    │
│                                                        │
└──────────────────────────────────────────────────────┘
```

### 7.2 UI Design Philosophy

- **Dark mode default** with deep navy/slate backgrounds (#0a0f1e, #111827)
- **Glassmorphism**: Semi-transparent cards with `backdrop-filter: blur()`, subtle borders
- **Liquid glass**: Gradient overlays with animated mesh backgrounds
- **Acrylic/Mica**: Frosted-glass effects on panels, modals, and navigation
- **Micro-animations**: Hover effects, card transitions, search result fade-ins
- **Typography**: Google Fonts (Inter or Outfit) for clean, modern readability
- **Color palette**: Indigo (#6366f1) → Violet (#8b5cf6) → Cyan (#06b6d4) gradient accents
- **Conversation bubbles**: User messages (right-aligned, accent color), Agent messages (left-aligned, glass)

### 7.3 Frontend Rendering Strategy

- Parse markdown in agent responses client-side using a lightweight renderer (e.g., `marked.js` via CDN or embedded)
- Syntax highlighting for code blocks (e.g., `highlight.js` or `prism.js` embedded)
- Virtual scrolling for large conversations (custom implementation or intersection observer)
- Responsive grid layout with CSS Grid/Flexbox

---

## 8. Open Questions

- [x] Tech stack: Go backend + vanilla JS frontend (decided)
- [ ] Should we support VSCode/Cursor/Claude Code transcript formats in a future version?
- [ ] Should we persist a search index to disk for faster subsequent launches?
- [ ] Should the tool support a "watch mode" that updates when new sessions appear?

---

## Revision History

| Date | Author | Changes |
|------|--------|---------|
| 2026-09-18 | @vadde | Initial draft |
| 2026-09-18 | @vadde | Complete specification with all requirements |
| 2026-10-19 | @vadde | Add R-037 (Bounded Diagram Viewports & Fullscreen Theater Mode) and R-038 (Native In-Browser File & Artifact Viewer) |
| 2026-11-04 | @vadde | Add R-040 (Session Agent Brain & MCP Behavioral Matrix), R-041 (Command Execution & Tool Breakdown), and R-042 (Empirical Skill Invocation Telemetry & Standby Decomposition) |
| 2026-11-05 | @vadde | Add R-043 (Bin-Packing Spatial Density & Panoramic Masonry Matrix), R-044 (Live Reactive Telemetry Sync), and R-045 (Session Files & Artifacts Holographic Matrix) |
| 2026-11-06 | @vadde | Add R-046 (Universal Multi-Dimensional Forensic Sorting Engine & Touch Span Telemetry) |
| 2026-11-09 | @vadde | Add R-053..R-066 (Holographic GitHub Review Diff Pane & Spatial Sync Engine) |
| 2026-11-09 | @vadde | Add R-067..R-072 (Sub-Pixel Code Gutter Alignment, Multi-Hunk Navigation Radar & File Buffer Intelligence) |
