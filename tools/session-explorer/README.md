# 🔍 session-explorer

> A high-performance, single-binary session intelligence explorer and browser for Antigravity IDE agent conversation histories.

![Status](https://img.shields.io/badge/status-released-brightgreen)
![Version](https://img.shields.io/badge/version-0.3.5-emerald)
![Language](https://img.shields.io/badge/language-Go%201.22%20%7C%20Bun-indigo)
![License](https://img.shields.io/badge/license-MIT-purple)

---

## Overview

**Session Explorer** is a high-performance, single-binary conversation intelligence browser and real-time observability cockpit for agentic coding transcripts (Antigravity CLI & IDE, Claude Code, GitHub Copilot). It concurrently ingests local JSONL logs, extracts workspaces and tool execution cycles, and serves a modern liquid-glass web interface embedded directly inside the compiled binary.

### Key Capabilities

- ⚡ **Concurrent Delta Scanner**: Indexes tens of thousands of JSONL transcript lines in under 200ms using Go goroutines with sub-50µs incremental delta rescans.
- 📦 **Zero-Dependency Binary**: Go embeds the entire web application, Highlight.js syntax engine, KaTeX LaTeX math, Mermaid diagram compiler, and Swagger UI via `//go:embed`.
- 🌐 **Universal Multi-Runtime Discovery**: Simultaneously discovers sessions across Antigravity IDE (`~/.gemini/antigravity-ide/brain/`) and CLI (`~/.gemini/antigravity-cli/brain/`) with runtime badges (`⚡ CLI`, `💻 IDE`).
- ⚡ **Real-Time Live Tailing**: Streams newly appended steps with sub-second latency via Server-Sent Events (`/api/sessions/{id}/tail`), Auto-Follow scroll mode, and zero-layout-shift staged buffering ($\Delta y = 0$).
- 💬 **Granular Filtering & Dialogue Mode**: 1-click `💬 Dialogue` mode isolates pure human-to-AI conversation, filtering out hundreds of autonomous tool execution cards.
- 📊 **Bounded Diagrams & Fullscreen Theater**: Inline Mermaid diagrams bounded to 440px to eliminate vertical scroll trapping; double-click or Theater button opens pan/drag 2D theater with 20%–500% zoom, Fit to View, and SVG/Source copy.
- 📁 **Native In-Browser File & Artifact Viewer**: Instant file inspection with Highlight.js syntax highlighting, line numbers gutter, anchor line highlighting (`#L10-L20`), full-width Markdown preview, pinned floating Find widget (`Cmd+F`), and dual-affordance 1-click copy path chips (`📋` → `✓`).
- 🧠 **Universal Session Agent Brain & MCP Behavioral Matrix (R-040)**: Complete multi-tier visibility into the constitutional rules, skills, plugins, MCP servers (`omni-graph`, `playwright`), and lifecycle hooks governing agent behavior. Auto-detects lazy MCP (`call_mcp_tool`) and eager MCP (`mcp_*`), tracks active vs dormant tools, step index spans, and GenAI Tool Diversity Index ($D = U/N$).
- 🔬 **Command Execution & Tool Sub-Breakdown Spectrum (R-041)**: Decomposes generic tools (e.g. `run_command` and filesystem operations) into executable binaries (`git`, `make`, `go test`, `bun`, `curl`, `lsof`), 8 operational categories (`VCS`, `Build`, `Test`, `Process`, `Inspect`, `Network`, `Runtime`, `Filesystem`), and file extensions (`.go`, `.js`, `.css`, `.json`, `.md`). Renders Year 2126 Holographic Command Spectrum bars, interactive sub-command matrix cards, and 1-click step jumping (`#38`, `#113`).
- ⚡ **Empirical Skill Invocation Telemetry & Standby Decomposition (R-042)**: Rigorously distinguishes between Tier 1 empirically executed skills (via `read_file`/`view_file` on `SKILL.md` or scripts) and Tier 2 standby skills mounted in prompts. Captures step markers (`#11`, `#944`), invocation triggers, first/last execution jumpers (`⏮ #11`, `⏭ #947`), and scope deduplication across workspace, global, and builtin tiers.
- 🧩 **EC2-Style Greedy Bin-Packing Spatial Density & Panoramic Masonry Matrix (R-043)**: Replaced static CSS Grid row constraints with greedy shortest-column bin packing ($\arg\min_k \sum_{j \in B_k} h_j$), stacking unequal height cards vertically directly beneath preceding elements without dead gaps. Expanded modal width to panoramic cockpit real estate (`min(1560px, 94vw)`), fitting 3 fluid columns across widescreen viewports, and upgrading session trail navigation counters into Year 2126 holographic action capsules (`.holo-header-btn`).
- 🛑 **Agent WIP Recovery & Disambiguation**: Sub-millisecond state restore (`/api/wip`, `make session-wip`) with mandatory agent halting when multiple parallel sessions exist in a workspace.
- 📖 **Embedded Offline OpenAPI 3.0 & Swagger UI**: Zero-CDN interactive API explorer at `http://localhost:9876/docs`.

---

## Installation & Quick Start

### Prerequisites

- **Go 1.22+** (to build the Go binary)
- **Bun 1.0+** (to bundle web assets)

### Quick Start (Makefile Commands)

From the **repository root**:

```bash
# Build and launch Session Explorer (auto-opens browser at http://127.0.0.1:9876)
make session-explorer

# Launch with custom arguments (e.g. port or verbose)
make run-tool T=session-explorer ARGS="--port 8080"

# Run tests or automated demo
make test-tool T=session-explorer
make demo-tool T=session-explorer
```

Or from inside `tools/session-explorer/`:

```bash
cd tools/session-explorer

# Interactive command directory
make help

# Build and launch web UI
make run

# Launch on custom port or in headless mode
make run PORT=8080
make run-headless

# Run automated mock dataset demo
make demo
```

---

## Makefile Command Directory

Every lifecycle action is managed through standard `make` targets:

| Command | Description |
|---------|-------------|
| `make run` | Build and launch Session Explorer (opens default browser) |
| `make run-headless` | Launch without opening browser (`--no-open`) |
| `make demo` | Run automated mock dataset demo script |
| `make build` | Complete build: bundle frontend with Bun + compile Go binary |
| `make build-web` | Bundle frontend web assets with Bun |
| `make build-go` | Compile standalone Go binary |
| `make test` | Run full unit test suite |
| `make test-race` | Run test suite with Go race detector enabled |
| `make test-cover` | Run tests with coverage summary |
| `make lint` | Run code linters (`go vet`) |
| `make status` | Display SDLC lifecycle phase and requirements matrix |
| `make spec` | Display specification summary |
| `make context` | Display current session context (`CONTEXT.md`) |
| `make devlog` | Display latest development log entries (`DEVLOG.md`) |
| `make setup` | Install Go and Bun frontend dependencies |
| `make clean` | Remove build artifacts (`bin/`, `src/web/dist/`) |

---

## Web UI Features & Shortcuts

| Feature | Description | Keyboard Shortcut |
|---------|-------------|-------------------|
| **Global Search** | Instant prompt & transcript search across all sessions | `Cmd + K` or `Ctrl + K` or `/` |
| **Find in File** | Search within active file in File Viewer with match count & cycling | `Cmd + F` or `Ctrl + F` |
| **Find Navigation** | Jump to Next / Previous match in file | `Enter` / `Shift + Enter` (or `▲` / `▼`) |
| **Close Dialogs** | Close Find widget, file viewer modal, search modal, or return to dashboard | `Esc` (hierarchical) |
| **Theme Toggle** | Switch between Liquid Dark and Frosted Light mode | Header button `☀️ / 🌙` |
| **Workspace Filter** | Filter sessions by detected workspace or project | Quick filter pills / Dropdown |
| **Date Range Filter** | Filter sessions by date range with quick presets (`Today`, `7d`, `30d`) | Toolbar date picker |
| **Dialogue Mode** | Filter out tool executions to read clean user-agent dialogue | Detail toolbar `💬 Dialogue` chip |
| **Theater Mode** | Expand Mermaid diagram into full-screen pan/zoom canvas (20%–500%) | Double-click canvas or `Theater` button |
| **Agent Brain Matrix** | Explore all global, workspace, plugin, and builtin rules & skills | Header button `🧠 Agent Brain` |
| **Session Context Drawer** | Inspect active rules, invoked skills (with call counts & step jump links) | Session header `🧠 Agent Context` |
| **Rule / Skill Inspection** | 1-click view of any markdown rule or skill in Native File Viewer | Card button `📖 Inspect` |
| **Export Session** | Download session as clean Markdown (`.md`) or structured JSON (`.json`) | Detail header buttons |

---

## Universal Agent Context & Behavioral Matrix (R-039)

Session Explorer provides full-spectrum transparency into the autonomous agent's cognitive boundaries, constitutional guardrails, skills, plugins, and MCP servers across multiple scopes:

### Multi-Scope Cascade Hierarchy
Rules and skills are discovered and prioritized according to strict constitutional precedence:
1. **Workspace / Codebase Scope** (`.agents/rules/`, `.agents/skills/`, `AGENTS.md`, `CLAUDE.md`, `.cursorrules`, `copilot-instructions.md`): Repository-specific guidelines override global defaults.
2. **Global System Scope** (`~/.gemini/config/rules/`, `~/.gemini/config/skills/`, `~/.claude/rules/`, `~/.cursorrules`): Machine-wide guardrails and constitutional invariants (e.g., Rule 101).
3. **Plugins Scope** (`~/.gemini/config/plugins/`): Composable community and engineering extension bundles.
4. **Built-in Scope** (`builtin/skills/`): Core agent runtime runbooks and fallback skills.

### Context Dual-Sourcing Architecture
1. **Static Filesystem Discovery**: Directly interrogates active filesystem locations to index the ground truth of all configured rules, skills, plugins, and MCP servers, even when transcripts omit initial system prompts to conserve context tokens.
2. **Dynamic Transcript Tool-Call Analysis**: Analyzes JSONL transcripts for tool executions (`read_file`, `view_file`, `invoke_subagent`, MCP calls) to compute empirical utilization metrics:
   - **⚡ Actively Invoked Skills**: Skills actually triggered during the session, complete with execution counts and first step index jump links.
   - **💤 Available On-Demand Skills**: Available skills that remained dormant in memory, ready for progressive disclosure.

### Deep Native File Viewer Integration (R-038 Re-use)
Every rule, constitution, and skill card features a 1-click `📖 Inspect` button that directly invokes the **Native In-Browser File Viewer** (`openFileViewer`), providing:
- Full-width fluid reading view (97.5vw $\times$ 96vh).
- Live **KaTeX math** formula typesetting and interactive **Mermaid diagrams**.
- Highlight.js syntax highlighting and unified line numbers gutter.
- Pinned floating in-page find HUD (`Cmd+F` / `Ctrl+F`).

### GenAI Token & Word Bloat Telemetry (AC-045)
To safeguard context windows from bloat and token exhaustion, Session Explorer analyzes the exact linguistic payload of every rule, guideline, prompt, and skill:
- **Linguistic Metrics**: Every rule and skill card displays exact line count (`lines`), word count (`words`), and byte size (`KB`).
- **Standard Token Estimation**: Estimates tokens using $T \approx \max(\lceil C / 4 \rceil, \lceil W \times 1.33 \rceil)$, providing a realistic token estimate across Gemini, Claude, and GPT tokenizers.
- **Pre-Loaded vs. On-Demand Analysis**:
  - **📌 Pre-Loaded Context Bloat**: Measures tokens permanently injected into turns by default via always-on constitutional rules and workspace guidelines (e.g., `~5,036 tokens`).
  - **💤 On-Demand Footprint**: Measures potential token capacity across available procedural skills and runbooks if loaded into context (e.g., `~34,198 tokens`).
  - **⚡ Active Session Ingestion**: Real-time measurement of tokens and words actually consumed by active rules and invoked skills in a given conversation (e.g., `~13,384 tokens` across 6,936 injected words).

---

## REST API Contract

The embedded HTTP server exposes a JSON REST API for integration:

| Endpoint | Method | Parameters | Description |
|----------|--------|------------|-------------|
| `/api/sessions` | `GET` | `workspace`, `runtime`, `sort_by`, `sort_order`, `from`, `to` | List all discovered sessions with summary stats |
| `/api/sessions/{id}` | `GET` | — | Retrieve full conversation messages and tool calls for a session |
| `/api/sessions/{id}/tail` | `GET` | `last_step` | Real-time Server-Sent Events (SSE) live tail stream |
| `/api/sessions/{id}/context` | `GET` | — | Session-specific active rules, actively invoked skills (with call counts & step index), and on-demand skills |
| `/api/agent/context/global` | `GET` | — | Discovered global rules, skills, plugins, MCP servers, and hooks across assistants |
| `/api/agent/context/workspace` | `GET` | `workspace` | Discovered codebase rules/skills + scope cascade (Workspace → Declared → Global → Built-in) |
| `/api/file` | `GET` | `path` | Secure local file retrieval with language detection & line anchors |
| `/api/search` | `GET` | `q`, `workspace`, `type`, `from`, `to` | Full-text search across transcripts |
| `/api/stats` | `GET` | — | Global statistics, workspace list, runtime breakdown, and top tools |
| `/api/wip` | `GET` | `workspace`, `session_id`, `limit`, `force_latest` | Recover agent WIP digest with multi-session disambiguation |
| `/api/watch-interval` | `GET` / `POST` | `interval_seconds` | Query or update background file watch polling interval |
| `/api/openapi.json` | `GET` | — | Raw OpenAPI 3.0 specification contract |
| `/docs` | `GET` | — | Offline Swagger UI documentation explorer |

---

## Development & Testing

```bash
# Run the complete test suite
make test

# Run Go linters (go vet)
make lint

# Run the automated demo with mock data
./examples/demo.sh

# Clean build artifacts
make clean
```

---

## Specification Traceability

This tool strictly adheres to Spec-Driven Development (SDD):

- **Specification**: [`specs/catalog/session-explorer.md`](../../specs/catalog/session-explorer.md)
- **Status & Traceability**: [`STATUS.md`](STATUS.md)
- **Session Continuity**: [`CONTEXT.md`](CONTEXT.md)
- **Development Journal**: [`DEVLOG.md`](DEVLOG.md)

---

## License

MIT © [vadde](https://github.com/vadde)
