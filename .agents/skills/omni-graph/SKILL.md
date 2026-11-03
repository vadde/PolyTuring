---
name: omni-graph
description: >-
  Use this skill to navigate, semantically search, trace call chains, and condense AST subgraphs for codebases using the local Omni-Graph knowledge hub. Always use this instead of running brute-force grep or multi-file cat searches.
---

# 🧭 Omni-Graph Codebase Intelligence Skill

This skill equips AI coding assistants with **deterministic compiler AST analysis, multi-workspace exploration, personalized PageRank call graph condensation, and dense vector semantic search** across indexed polyglot codebases (Rust, Go, Python, TypeScript/JavaScript, Java, C/C++, Shell, and Markdown docs).

Always use this skill instead of running blind recursive `grep -rn` or reading whole files sequentially.

---

## 🏛️ The Agentic Triad Architecture (3 Offerings Explained)

Omni-Graph is architected around a unified high-performance daemon (Rust on `:8080`, SurrealDB v2 on `:8000`, TEI Embeddings on `:8081`) exposed through **three distinct interface tiers**:

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                                 THE AGENTIC TRIAD                                      │
├─────────────────────────┬─────────────────────────────┬────────────────────────────────┤
│   1. NATIVE MCP SERVER  │      2. ERGONOMIC CLI       │       3. REST / CURL API       │
│  (Structured Cognitive) │      (Terminal Hands)       │     (Indestructible Floor)     │
├─────────────────────────┼─────────────────────────────┼────────────────────────────────┤
│ • Zero-token execution  │ • Human/Agent friendly      │ • Zero runtime dependencies    │
│ • Native IDE function   │ • Clean subcommands         │ • Universal across any OS/env  │
│   calling (JSON-RPC)    │ • Formatted ASCII summaries │ • High-performance direct      │
│ • Schema validated args │ • Auto workspace detection  │   JSON streaming               │
│ • No shell escaping     │ • Fast terminal debugging   │ • Bulletproof fallback         │
└─────────────────────────┴─────────────────────────────┴────────────────────────────────┘
                                          │
                                          ▼
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                          OMNI-GRAPH UNIFIED RUST CORE (:8080)                          │
│     AST Extraction • SAIR • Topological DAG • Leiden Clustering • PPR Neighborhood     │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

### Why 3 Offerings?
Underneath, all three tiers hit the identical Rust HTTP daemon endpoints. However, coding assistants operate in diverse host environments with differing execution permissions:
1. **Tier 1 (Native MCP)**: Ideal for modern IDE agents (Claude Desktop, Cursor, Windsurf, Roo Code, Antigravity IDE) where function calling happens directly in the model context without consuming shell execution tokens.
2. **Tier 2 (Ergonomic CLI `./scripts/omni`)**: Ideal for terminal-driven coding assistants (Aider, CLI agents, or agents with bash execution tools). It automatically resolves workspace context, colors output, formats ASCII tables, and handles command-line flags cleanly.
3. **Tier 3 (Direct Curl / REST API)**: The indestructible bedrock. When running in minimal subshells, sandboxes without `python` or `uv`, or when shell wrapper scripts fail, direct `curl` commands to `http://localhost:8080` guarantee 100% operational continuity.

---

## 🚦 Deterministic Execution Decision Tree

When executing an Omni-Graph task, follow this strict decision hierarchy:

```
                  ┌────────────────────────────────────────┐
                  │ Does your environment expose native   │
                  │ 'omni_*' tools in its active toolset?  │
                  └───────────────────┬────────────────────┘
                                      │
                         ┌────────────┴────────────┐
                         │ YES                     │ NO
                         ▼                         ▼
            ┌─────────────────────────┐   ┌────────────────────────────────┐
            │   TIER 1: NATIVE MCP    │   │ Do you have terminal / bash    │
            │ Call 'omni_*' tool call │   │ tool execution access?         │
            └────────────┬────────────┘   └───────────────┬────────────────┘
                         │                                │
                 Fails / │ Error             ┌────────────┴────────────┐
                         ▼                   │ YES                     │ NO
            ┌─────────────────────────┐      │                         ▼
            │   TIER 2: ERGONOMIC CLI │ ◀────┘            ┌─────────────────────────┐
            │  Run './scripts/omni'   │                   │   TIER 3: REST / CURL   │
            └────────────┬────────────┘                   │  Run 'curl :8080/api/'  │
                         │                                └────────────┬────────────┘
                 Fails / │ Error                                       │
                         ▼                                             │ Fails / Refused
            ┌──────────────────────────────────────────────────────────┴────────────┐
            │                   DEGRADED SELF-HEALING PROTOCOL                      │
            │   Check daemon health ──▶ Spin up stack ──▶ Scoped file read (<150L)  │
            └───────────────────────────────────────────────────────────────────────┘
```

---

## 🗺️ Master Rosetta Stone: 1-to-1 Equivalence Across All 3 Tiers

Use this table to translate any codebase reconnaissance intent across MCP, CLI, and REST:

| Capability / Intent | Tier 1: Native MCP Tool | Tier 2: Ergonomic CLI (`./scripts/omni`) | Tier 3: Low-Level REST / Curl |
| :--- | :--- | :--- | :--- |
| **Symbol Definition** | `omni_definition(symbol="X")` | `./scripts/omni symbol X` | `curl -s "http://localhost:8080/api/symbol?name=X&workspace=PolyTuring"` |
| **Callers & References** | `omni_references(symbol="X")` | `./scripts/omni references X` | `curl -s "http://localhost:8080/api/references?symbol=X&workspace=PolyTuring"` |
| **Condensed Call Slice** | `omni_condense(symbol="X", hops=2)` | `./scripts/omni condense X PolyTuring 2` | `curl -s "http://localhost:8080/api/condense?symbol=X&workspace=PolyTuring&hops=2"` |
| **Semantic Code Search** | `omni_search(query="...", k=10)` | `./scripts/omni search "..." PolyTuring 10` | `curl -s "http://localhost:8080/api/search?q=...&workspace=PolyTuring&k=10"` |
| **Doc Search (Partitioned)**| `omni_search(query="...", include_docs=True)` | `curl` with `include_docs=true` | `curl -s "http://localhost:8080/api/search?q=...&workspace=PolyTuring&include_docs=true"` |
| **Architectural Galaxies** | `omni_galaxies(limit=15)` | `./scripts/omni galaxies` | `curl -s "http://localhost:8080/api/galaxies?workspace=PolyTuring&limit=15"` |
| **Hybrid Graph-RAG Query**| `omni_query(prompt="...")` | `./scripts/omni query "..."` | `curl -s -X POST http://localhost:8080/api/query -H "Content-Type: application/json" -d '{"prompt": "...", "workspace": "PolyTuring"}'` |
| **Augment Runtime Edge** | *(Via CLI or REST)* | `./scripts/omni relate Src Tgt CALLS` | `curl -s -X POST http://localhost:8080/api/relationships -H "Content-Type: application/json" -d '{"source_symbol":"Src","target_symbol":"Tgt","type":"CALLS"}'` |
| **List Workspaces & Stats**| *(Via CLI or REST)* | `./scripts/omni workspaces` | `curl -s "http://localhost:8080/api/workspaces"` |
| **Git Worktrees List/Sync** | *(Via CLI or REST)* | `./scripts/omni worktrees sync` | `curl -s -X POST http://localhost:8080/api/worktrees/sync -H "Content-Type: application/json" -d '{}'` |
| **Watcher Telemetry** | *(Via CLI or REST)* | `./scripts/omni watch status` | `curl -s "http://localhost:8080/api/watch/status"` |
| **System Health Check** | *(Via CLI or REST)* | `./scripts/omni health` | `curl -s "http://localhost:8080/api/health"` |

---

## ⚙️ Detailed Interface Specifications

### 1. Tier 1: Native MCP Tools Reference

The stdio MCP server runs hermetically via `tools/omni-graph/scripts/omni-mcp.py` using `uv run --script`:

```json
{
  "tools": [
    {
      "name": "omni_definition",
      "description": "Lookup precise AST definition, signatures, and file line ranges for a code symbol.",
      "parameters": {
        "symbol": "string (required) — Exact identifier name (function, struct, class, interface)",
        "workspace": "string (optional) — Workspace name (defaults to auto-detected Git root)"
      }
    },
    {
      "name": "omni_references",
      "description": "Find all callers, references, and incoming dependency edges across the entire codebase.",
      "parameters": {
        "symbol": "string (required) — Target symbol name",
        "workspace": "string (optional) — Target workspace name"
      }
    },
    {
      "name": "omni_condense",
      "description": "Extract a strictly budgeted (<1500 tokens) multi-hop AST call slice around a symbol using Personalized PageRank.",
      "parameters": {
        "symbol": "string (required) — Root symbol to center call slice around",
        "hops": "integer (optional, default 2) — Call traversal radius (1 or 2)",
        "workspace": "string (optional) — Target workspace name"
      }
    },
    {
      "name": "omni_search",
      "description": "Perform dense vector semantic search across indexed symbols to locate relevant implementations.",
      "parameters": {
        "query": "string (required) — Natural language or technical concept",
        "k": "integer (optional, default 10) — Number of top results",
        "workspace": "string (optional) — Target workspace name",
        "include_docs": "boolean (optional, default false) — Include markdown documentation nodes"
      }
    },
    {
      "name": "omni_galaxies",
      "description": "Inspect macro architectural subsystems (galaxies), modular boundaries, instability scores, and coupling metrics.",
      "parameters": {
        "workspace": "string (optional) — Target workspace name",
        "limit": "integer (optional, default 15) — Max subsystems to return"
      }
    },
    {
      "name": "omni_query",
      "description": "Execute hybrid Graph-RAG retrieval combining vector search, community summaries, and multi-hop AST traces into a grounded Markdown brief.",
      "parameters": {
        "prompt": "string (required) — Architectural question to synthesize",
        "workspace": "string (optional) — Target workspace name",
        "include_docs": "boolean (optional, default false) — Include documentation in seeds"
      }
    }
  ]
}
```

### 2. Tier 2: Ergonomic CLI Command Reference

The command-line wrapper lives at `./scripts/omni` (symlinked to `tools/omni-graph/scripts/omni.sh`):

```bash
# Discovery & Health
./scripts/omni health                    # Verify Rust core, SurrealDB, and TEI container health
./scripts/omni workspaces                # Discover all indexed codebases, languages, and node counts
./scripts/omni worktrees [sync]          # Scan & synchronize ephemeral Git worktrees

# AST Code Intelligence
./scripts/omni symbol <name> [ws]        # Locate symbol definition, line bounds, and code snippet
./scripts/omni references <name> [ws]    # Locate all callers, importing files, and test suites
./scripts/omni condense <name> [ws] [h]  # Extract <1500-token PPR-condensed call graph slice (hops: 1 or 2)
./scripts/omni search <query> [ws] [k]   # Dense vector semantic code search (top k)
./scripts/omni galaxies [ws]             # ASCII table of modular architectural subsystems & coupling metrics
./scripts/omni query <prompt> [ws] [k]   # Hybrid Graph-RAG architectural brief
./scripts/omni relate <src> <tgt> [t]    # Inject dynamic/inferred runtime relationship edge (CALLS, ROUTES_TO)

# Ingestion & Live Watch
./scripts/omni ingest <path> [ws]        # Initial one-time indexing of an untracked repository
./scripts/omni watch status              # Inspect real-time file watcher sync metrics & event counters
./scripts/omni watch start [path]        # Start live file watcher daemon for repository
./scripts/omni watch stop                # Stop live file watcher daemon
```

### 3. Tier 3: Direct REST API Endpoints Reference

All endpoints serve JSON directly from the Rust core on `http://localhost:8080`:

- `GET  /api/health` — System status (`{"status":"ok","surrealdb":true,"embeddings":true}`)
- `GET  /api/workspaces` — Array of indexed workspaces, node counts, languages, and live watch states
- `GET  /api/symbol?name=<sym>&workspace=<ws>` — Precise symbol definition records
- `GET  /api/references?symbol=<sym>&workspace=<ws>` — Callers, importers, and edge categories
- `GET  /api/condense?symbol=<sym>&workspace=<ws>&hops=2` — PPR-condensed Markdown call slice
- `GET  /api/search?q=<query>&workspace=<ws>&k=10&include_docs=false` — Vector similarity search
- `GET  /api/galaxies?workspace=<ws>&limit=15` — Louvain/Leiden modular clusters and instability metrics
- `POST /api/query` — Hybrid Graph-RAG synthesis (`{"prompt":"...", "workspace":"...", "include_docs":false}`)
- `POST /api/relationships` — Edge augmentation (`{"source_symbol":"...", "target_symbol":"...", "type":"..."}`)
- `GET  /api/watch/status` — Live watcher metrics, event counts, reindexed files, and debouncing status
- `POST /api/watch/start` — Enroll directory in live watch (`{"path":"/path/to/repo"}`)
- `POST /api/worktrees/sync` — Rescan `.git/worktrees` and enroll child branches

---

## 🧩 Resolving Permutations, Combinations & Scoping

### 1. Workspace Scoping Permutations
- **Auto-Detection (Default)**: In both MCP and CLI, omitting `workspace` automatically resolves the current Git repository name (`git rev-parse --show-toplevel`).
- **Explicit Scoping**: When working in a multi-repo monorepo or inspecting dependencies in an indexed sibling workspace, pass `workspace="TargetRepo"`:
  - MCP: `omni_definition(symbol="UserHandler", workspace="session-explorer")`
  - CLI: `./scripts/omni symbol UserHandler session-explorer`
  - REST: `curl -s "http://localhost:8080/api/symbol?name=UserHandler&workspace=session-explorer"`
- **Unscoped Monorepo Scan (Divergence Check)**: Omit `workspace` on REST `/api/symbol?name=X` to view matches across ALL indexed repositories side-by-side.

### 2. Content Partitioning: Code vs Documentation
Omni-Graph v0.3.0 enforces strict Content Partitioning to prevent Markdown documentation from shadowing source code:
- By default, all vector queries (`omni_search`, `omni_query`, `/api/search`) search **code symbols only** (`include_docs: false`).
- To search architectural guides, specifications, and living documentation, explicitly enable documentation search:
  - MCP: `omni_search(query="how to deploy", include_docs=True)`
  - CLI/REST: `curl -s "http://localhost:8080/api/search?q=how+to+deploy&workspace=PolyTuring&include_docs=true"`

### 3. Hop Radius & Context Budgeting
- Default traversal is `hops=2`, covering caller $\to$ root $\to$ callee.
- If investigating an ultra-dense central hub function with dozens of dependencies, pass `hops=1` for a tighter, atomic call view.
- Subgraph condensation automatically executes 25-step **Personalized PageRank (PPR)** ($\alpha=0.85$) and strictly truncates output at 1,490 tokens to guarantee prompt safety.

---

## 🛠️ Degraded & Out-of-Order Self-Healing Playbook

When operating in real environments, things may not always be in order. Follow these deterministic recovery protocols:

### Scenario 1: MCP Tool Not Present in Model Context
- **Symptom**: Model environment does not list `omni_*` tools in its tool catalog.
- **Root Cause**: IDE does not support MCP, or MCP server is not registered in client configuration.
- **Recovery Protocol**:
  1. **DO NOT** attempt to guess or hallucinate MCP tool names.
  2. Switch immediately to **Tier 2 Ergonomic CLI**: execute commands via `./scripts/omni <subcommand>`.
  3. If your agent framework has no shell tool, use **Tier 3 Curl** via whatever HTTP client is available.

### Scenario 2: CLI Permission Denied or File Not Found
- **Symptom**: Shell execution returns `./scripts/omni: Permission denied` or `No such file or directory`.
- **Recovery Protocol**:
  1. Verify execution path: run `ls -la scripts/omni` from repository root.
  2. If missing executable bit: run `chmod +x ./scripts/omni tools/omni-graph/scripts/*.sh`.
  3. If current working directory is a subfolder, use absolute path: `${REPO_ROOT}/scripts/omni` or fallback to **Tier 3 Curl**.

### Scenario 3: Daemon Down or Connection Refused (`localhost:8080`)
- **Symptom**: `curl` or CLI returns `Failed to connect to localhost port 8080: Connection refused`.
- **Recovery Protocol**:
  1. Check Docker container status:
     ```bash
     docker ps --filter "name=omni" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
     ```
  2. If containers are stopped, spin up the stack from repo root:
     ```bash
     make up
     # or: docker compose -f tools/omni-graph/docker-compose.yml up -d
     ```
  3. Wait 5 seconds for health status: `curl -s http://localhost:8080/api/health`.
  4. **Cold Fallback (If Docker is completely unavailable)**:
     - DO NOT crash or enter an infinite loop.
     - Switch to **Cold Graceful Degradation**: locate files using `git ls-files` and inspect focused file slices using `view_file` strictly respecting the **150-line gate** (`StartLine: X, EndLine: X+140`).

### Scenario 4: Workspace Not Indexed or Total Nodes is Zero
- **Symptom**: Query returns `total_results: 0` or workspace name is not found.
- **Recovery Protocol**:
  1. Check registered workspaces:
     ```bash
     ./scripts/omni workspaces
     ```
  2. Verify if the repository is registered under a slightly different name (e.g., casing or subfolder name).
  3. If the repository is completely unindexed, run one-time ingestion:
     ```bash
     ./scripts/omni ingest . [workspace_name]
     ```
  4. Wait for ingestion confirmation (`ingestion_time_ms`), then resume querying.

### Scenario 5: Symbol Lookup Returns 0 Matches
- **Symptom**: `omni_definition(symbol="processOrder")` returns empty matches.
- **Root Causes**:
  - Spelling, case conventions (snake_case vs camelCase vs PascalCase), or language differences.
  - Target symbol is an unexported local variable or anonymous closure.
  - Symbol is part of an external third-party library or vendor crate.
- **Recovery Protocol**:
  1. Fall back to Dense Vector Semantic Search:
     ```bash
     ./scripts/omni search "process order" [workspace] 5
     ```
  2. Query architectural galaxies to locate the owning module:
     ```bash
     ./scripts/omni galaxies [workspace]
     ```
  3. If external library, query the caller function that imports it to view the usage pattern.

### Scenario 6: Symbol Ambiguity (Multiple Matches Across Files)
- **Symptom**: `omni_definition(symbol="validate")` returns 6 matches across multiple files.
- **Recovery Protocol**:
  1. Inspect the returned JSON array: each match includes `file_path`, `language`, and `kind`.
  2. Filter by the specific file path or language you are targeting.
  3. For callers/references, query with workspace scoping or inspect the caller's `source_file` to isolate the exact dependency graph.

### Scenario 7: Condensed Subgraph Too Large / Budget Exceeded
- **Symptom**: Target function is a massive utility hub; 2-hop trace is dense.
- **Recovery Protocol**:
  1. Omni-Graph v0.3.0 automatically applies Personalized PageRank (PPR) power iteration ($\alpha=0.85$, 25 iterations) and caps output at 1,490 tokens.
  2. If a more focused trace is desired, dial down hops to `hops=1`:
     - MCP: `omni_condense(symbol="X", hops=1)`
     - CLI: `./scripts/omni condense X PolyTuring 1`

### Scenario 8: Live Watch Telemetry Desynchronized
- **Symptom**: You edited a file, but AST queries still report old line numbers after 1 second.
- **Recovery Protocol**:
  1. Check live watch status:
     ```bash
     ./scripts/omni watch status
     ```
  2. If status is `stopped` or `inactive`, restart the watcher:
     ```bash
     ./scripts/omni watch start .
     ```
  3. The watcher automatically debounces (500ms) and re-indexes modified files with zero full-table resets.

### Scenario 9: Operating in Isolated Git Worktrees
- **Symptom**: Agent is running inside a Git worktree (`git worktree add ../wt-feature`).
- **Recovery Protocol**:
  1. Sync worktrees:
     ```bash
     ./scripts/omni worktrees sync
     ```
  2. Omni-Graph resolves the `.git` pointer, inherits cluster seeds from the parent workspace, and starts live watching the worktree.
  3. Scope all queries using `workspace=<worktree_name>` to prevent cross-branch AST edge bleeding.

---

## 🔌 Client Installation & Setup Guide

To install and register the Omni-Graph skill and MCP server in your coding environment:

### 1. Antigravity CLI / IDE
Run the canonical automated installer:
```bash
# Workspace installation (.agents/)
make setup-agent

# Or global system installation (~/.gemini/config/)
make setup-agent global
```

### 2. Claude Desktop
Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):
```json
{
  "mcpServers": {
    "omni-graph": {
      "command": "uv",
      "args": [
        "run",
        "--script",
        "/absolute/path/to/PolyTuring/tools/omni-graph/scripts/omni-mcp.py"
      ],
      "env": {
        "OMNI_API_URL": "http://localhost:8080"
      }
    }
  }
}
```

### 3. Cursor
Add to your project's `.cursor/mcp.json`:
```json
{
  "mcpServers": {
    "omni-graph": {
      "command": "uv",
      "args": [
        "run",
        "--script",
        "${workspaceFolder}/tools/omni-graph/scripts/omni-mcp.py"
      ]
    }
  }
}
```

### 4. Windsurf / Codeium
Add to `~/.codeium/windsurf/mcp_config.json`:
```json
{
  "mcpServers": {
    "omni-graph": {
      "command": "uv",
      "args": [
        "run",
        "--script",
        "/absolute/path/to/PolyTuring/tools/omni-graph/scripts/omni-mcp.py"
      ]
    }
  }
}
```

### 5. Aider / Standalone Terminal Coding Assistants
Aider and terminal agents execute commands directly via the `./scripts/omni` CLI:
```bash
# Add to .aider.conf.yml or run in chat:
/run ./scripts/omni symbol <function_name>
/run ./scripts/omni condense <function_name>
```

---

## 📋 Recommended Agent Workflow Recipes

### Recipe A: Investigating a Bug in an Unfamiliar Module
1. Discover the owning subsystem without reading files:
   - MCP: `omni_galaxies()`
   - CLI: `./scripts/omni galaxies`
2. Search for the error message or feature concept:
   - MCP: `omni_search(query="rate limiter token refill", k=5)`
   - CLI: `./scripts/omni search "rate limiter token refill"`
3. Pull a condensed AST call graph slice (<1500 tokens) around the entrypoint:
   - MCP: `omni_condense(symbol="refill_tokens", hops=2)`
   - CLI: `./scripts/omni condense refill_tokens PolyTuring 2`
4. Only now open the specific target file via `view_file` at the exact line range to write your fix.

### Recipe B: Safe Refactoring or Renaming a Public Symbol
1. Inspect current definition and signature:
   - MCP: `omni_definition(symbol="calculate_tax")`
   - CLI: `./scripts/omni symbol calculate_tax`
2. **Mandatory Blast-Radius Check**: Locate all callers and importing files across the repository:
   - MCP: `omni_references(symbol="calculate_tax")`
   - CLI: `./scripts/omni references calculate_tax`
3. Edit the definition and all callers with surgical precision, ensuring zero broken dependencies.

### Recipe C: Augmenting Dynamic Cross-Service Dependencies
When inspecting Dockerfiles, HTTP routing tables, or message brokers, Tree-sitter cannot infer runtime connections:
1. Identify source and target components (e.g. `FrontendClient` $\to$ `OrderServiceAPI`).
2. Inject an inferred relationship edge:
   - CLI: `./scripts/omni relate FrontendClient OrderServiceAPI ROUTES_TO`
   - REST: `POST /api/relationships` with `{"source_symbol":"FrontendClient", "target_symbol":"OrderServiceAPI", "type":"ROUTES_TO", "category":"INFERRED"}`
3. All future blast-radius queries (`omni_references`) will immediately track this cross-service dependency.
