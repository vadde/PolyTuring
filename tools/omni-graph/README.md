# 🧭 Omni-Graph

> A 100% local, multi-container semantic knowledge hub that models codebases as rich AST graphs augmented with vector embeddings and community clustering — purpose-built for Apple Silicon ARM64 unified memory.

![Status](https://img.shields.io/badge/status-in--progress-blue)
![Version](https://img.shields.io/badge/version-0.2.2-emerald)
![Language](https://img.shields.io/badge/language-rust%20%2B%20react-orange)
![Engine](https://img.shields.io/badge/database-SurrealDB%20v2-green)
![Embeddings](https://img.shields.io/badge/embeddings-HF%20TEI%20(384--dim)-purple)

---

## 1. Overview & The 60-Second Mental Model

### 💡 Why Does Omni-Graph Exist? (In Plain English)

Imagine hiring a brilliant senior software engineer, but forcing them to read your **entire 100,000-line codebase from top to bottom** every single time you ask: *"Where is payment processing handled?"*. They would waste 30 minutes reading, blow thousands of dollars in billing fees, and lose their train of thought due to cognitive exhaustion.

That is exactly how standard AI coding agents behave when relying on blind `cat` and `grep` file-walking operations.

| ❌ The Old Way (Blind Agent) | ✅ The Omni-Graph Way (Instant X-Ray) |
| :--- | :--- |
| Runs blind `grep` or `find` across the whole repo | Pre-indexes syntax trees & caller relationships locally |
| Dumps 10,000 lines of noise, minified assets, and test logs | Pinpoints symbol declarations, signatures, and callers in 5ms |
| Burns 40,000+ LLM tokens per simple question | Strict mathematical guarantee: **≤ 1,500 tokens per prompt slice** |
| Blows past model context limits, triggering hallucinations | Agent gets exact context instantly and writes working code |

**Omni-Graph replaces brute-force file walking with deterministic AST graphs and local vector search.**

```
┌────────────────────────────────────────────────────────────────────────┐
│                        DOCKER BRIDGE: omni-net                         │
│                                                                        │
│  ┌──────────────────────┐     ┌───────────────────────────────────┐    │
│  │   REACT UI           │     │   RUST ORCHESTRATOR               │    │
│  │   (graph-ui:3000)    │────▶│   (rust-app:8080)                 │    │
│  │   Cosmograph WebGL   │     │   Axum + Tree-Sitter + Graph Engine│   │
│  └──────────────────────┘     └───────┬───────────────────────────┘    │
│                                       │                                │
│                      ┌────────────────┼────────────────┐               │
│                      │                                 │               │
│                      ▼                                 ▼               │
│        ┌───────────────────────────┐    ┌─────────────────────────────┐│
│        │   SURREALDB v2            │    │   HUGGINGFACE TEI           ││
│        │   (graph-db:8000)         │    │   (embedding-engine:80)     ││
│        │   384-dim HNSW Vector ANN │    │   BAAI/bge-small-en-v1.5    ││
│        │   Typed linked_to Edges   │    │   ARM64 SIMD Accelerated    ││
│        └───────────────────────────┘    └─────────────────────────────┘│
└────────────────────────────────────────────────────────────────────────┘
```

---

## 2. The 4-Tier Agent Enforcement Shield

To prevent LLM agents from reverting to primitive string searches when under heavy cognitive load, Omni-Graph enforces a **4-tier deterministic guardrail shield**:

```
┌────────────────────────────────────────────────────────────────────────┐
│                        AGENT RUNTIME LOOP                              │
│                                                                        │
│  [Tier 1: Cognitive Steering Rule]                                     │
│  .agents/rules/00-omni-graph-mandatory-retrieval.md                    │
│  Mandates semantic AST lookup before inspecting raw files              │
│                                                                        │
│  [Tier 2: Pre-Invocation Injection]                                    │
│  .agents/hooks.json: PreInvocation Hook                                │
│  Injects dynamic active graph endpoints before model output            │
│                                                                        │
│  [Tier 3: Pre-Tool Deterministic Interceptor]                          │
│  .agents/hooks.json: PreToolUse Hook (matcher: run_command|grep_search|view_file) │
│  Hard-blocks blind grep/cat; rewrites command with Omni-Graph guidance  │
│                                                                        │
│  [Tier 4: Native Symbolic Tooling & Skill]                             │
│  .agents/skills/omni-graph/SKILL.md + omni.sh CLI                      │
│  Provides structured symbolic operations:                              │
│  - symbol: exact definition & line range                               │
│  - references: cross-file caller lineage                               │
│  - condense: sub-1500 token AST subgraphs                              │
│  - query: hybrid Graph-RAG retrieval                                   │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Symbolic LSP Tools (Serena Model)

Inspired by Serena's language server architecture, Omni-Graph extracts deterministic Abstract Syntax Trees via Tree-Sitter across polyglot codebases (Rust, Python, Go, TypeScript, JavaScript):

| LSP Capability | Omni-Graph API | CLI Command | Description |
| :--- | :--- | :--- | :--- |
| `textDocument/definition` | `GET /api/symbol?name=<sym>` | `./scripts/omni.sh symbol <name>` | Locates exact declaration, file path, line range, and code signature |
| `textDocument/references` | `GET /api/references?symbol=<sym>` | `./scripts/omni.sh references <name>` | Discovers all functions and structs calling or referencing the symbol |
| `subgraph/condense` | `GET /api/condense?symbol=<sym>&hops=2` | `./scripts/omni.sh condense <name> 2` | Slices a 2-hop neighborhood into a rich Markdown block (<1500 tokens) |

---

## 4. Multi-Workspace Isolation & Database Schema Pattern

Omni-Graph is architected for multi-repo and multi-workspace development. Every codebase ingested into SurrealDB preserves its unique semantic RAG properties without collision:

### Record Namespace Pattern
- **Node ID Format**: `type::thing('node', '{workspace}:{file_path}:{symbol_name}:{line}')`
- **Node Record**:
  ```surql
  {
    workspace: "ecommerce-platform",
    label: "ingest_directory",
    kind: "function",
    file_path: "src/ingestion/mod.rs",
    language: "rust",
    line_start: 66,
    line_end: 180,
    text: "pub async fn ingest_directory(...)",
    embedding: [0.034, -0.082, ...], -- 384-dim HNSW indexed
    community: 3                    -- Galaxy Cluster ID
  }
  ```
- **Relationship Edge Record**:
  ```surql
  RELATE node:`ecommerce-platform:src/main.rs:main:27`->linked_to->node:`ecommerce-platform:src/ingestion/mod.rs:ingest_directory:66`
  SET workspace = 'ecommerce-platform', type = 'CALLS', category = 'EXTRACTED';
  ```

### Partitioned Retrieval
- All vector searches, symbolic lookups, and Graph-RAG queries can optionally filter by `workspace` (e.g. `WHERE workspace = $workspace`) or query across all workspaces simultaneously.
- Run `make workspaces` to view all ingested codebases and their node/file counts.

---

## 5. Community Clustering Computation (Graph RAG Macro-Retrieval)

### What is Community Clustering?
Standard vector embeddings capture **microscopic similarity** (finding a snippet that matches a keyword). But they cannot explain the **macroscopic architecture** (e.g., "What does this codebase do?", "How are services partitioned?").

Community Clustering solves this by applying **modularity-based community detection (Louvain / Leiden algorithm)** directly onto the Abstract Syntax Tree relation graph:
1. **Adjacency Graph Formation**: Nodes (`function`, `struct`, `class`) and their call/import edges are converted into an undirected structural graph.
2. **Iterative Label Propagation**: Nodes iteratively propagate module labels to maximize modularity density ($Q$-score), grouping tightly coupled functions into cohesive clusters.
3. **Galaxy ID Assignment**: Each cluster is assigned a unique `galaxy_id` (0, 1, 2, ...), stored directly on SurrealDB node records.
4. **Macro-Summaries**: The top symbols and primary directories of each community cluster are aggregated into human-readable architectural summaries (e.g., *Cluster #1: DB Repository*, *Cluster #2: Axum REST API*).
5. **Celestial WebGL Rendering**: In the Cosmograph WebGL frontend, each cluster is rendered as a distinct, color-coded celestial galaxy nebula.
6. **Graph RAG Context Injection**: When an agent asks an architectural question (`make query Q="..."`), Omni-Graph retrieves both the macro community summaries and the micro code snippets, keeping the entire prompt slice under **1500 tokens**.

---

## 6. Git Worktrees: The "Browser Tabs" Mental Model

### 💡 What is a Git Worktree? (In Plain English)

**Think of Git Worktrees like having multiple browser tabs open for the exact same website at the same time:**

* In normal Git, if you want to switch from developing a new feature to fix an urgent bug, you have to run `git stash`, run `git checkout bugfix`, and restart your compiler or local dev server.
* With **Git Worktrees**, you have multiple physical folders on disk sharing the same underlying Git history. You (or 4 parallel AI coding agents) can work on `feature/stripe-v2` and `refactor/auth-jwt` *simultaneously* in separate folders without branch switching conflicts.

### The Problem With Traditional Indexers:
Traditional tools get confused, treat each folder as an unrelated duplicate repository, or force you to re-index the entire project from scratch every time a branch changes.

### The Omni-Graph Solution (Ephemeral Branch Fabric):
Point Omni-Graph at the main project folder once (`make ingest PATH=/path/to/ecommerce-platform`). It automatically discovers all active worktree folders (e.g. `.git/worktrees/*`), links their lineage, indexes them in parallel, and live-updates whichever folder you edit in real time!

---

## 7. The 30-Second Quick Start

Users and agents can invoke all operations with simple, intuitive Make commands:

### Step 1: Launch the Local Stack (Docker)
```bash
make up
```
Starts all 4 containers in the background (`omni-rust-app`, `omni-graph-db`, `omni-embedding-engine`, `omni-graph-ui`). On macOS, host `/Users` is automatically mounted into the container.

### Step 2: Point Omni-Graph at ANY Codebase on Your Machine
```bash
# Ingest any project directory on your machine (e.g. ecommerce-platform)
make ingest PATH=/path/to/ecommerce-platform

# Or optionally assign a custom project namespace
make ingest PATH=/path/to/ecommerce-platform PROJECT=ecommerce-core
```
Omni-Graph extracts AST symbols, computes 384-dim embeddings, detects child worktrees, and enrolls real-time live file watchers automatically.

### Step 3: Ask Architectural Questions or Trace Symbols
```bash
# Hybrid Graph-RAG architectural question (<1500 tokens context slice):
make query Q="Where is payment validation handled?" PROJECT=ecommerce-core

# Or open the 3D visual graph in your browser:
open http://localhost:3000
```

---

## 8. Daily Cheat Sheet for Developers & Coding Agents

| What You Want to Do | Command to Run | What It Does |
| :--- | :--- | :--- |
| **Start everything** | `make up` | Starts the 4 Docker containers in background |
| **Index a codebase** | `make ingest PATH=/path/to/repo` | Parses AST, generates embeddings, starts live watcher |
| **List indexed repos** | `make workspaces` | Prints table of all indexed workspaces and child worktrees |
| **Sync parallel worktrees** | `make sync-worktrees` | Reconciles new Git branches and aligns live watchers |
| **Find symbol definition** | `make graph-symbol SYM=OrderService` | Locates exact file, line range, and code signature in 5ms |
| **Find who calls a symbol** | `make graph-references SYM=OrderService` | Lists all callers and dependent functions across files |
| **Get <1500 token slice** | `make graph-condense SYM=OrderService` | Extracts 2-hop neighborhood formatted for LLM prompts |
| **Ask architectural question** | `make query-graph Q="How does auth work?"` | Hybrid Graph-RAG synthesis using local vector + graph search |
| **Wipe DB for clean slate** | `make reset-graph-db` | Nukes SurrealDB storage and auto-restarts with clean schema |

---

## 9. CLI & API Reference

### CLI Usage (`omni.sh`)

```bash
# Ingest
./scripts/omni.sh ingest "/path/to/codebase" [project]

# Cluster
./scripts/omni.sh cluster [project]

# Workspaces
./scripts/omni.sh workspaces

# Semantic search
./scripts/omni.sh search "query terms" [workspace] [k]

# Symbolic definition lookup (LSP)
./scripts/omni.sh symbol "init_pool" [workspace]

# Symbolic caller references (LSP)
./scripts/omni.sh references "init_pool" [workspace]

# Extract condensed AST subgraph slice (<1500 tokens)
./scripts/omni.sh condense "init_pool" [workspace] [hops]

# Hybrid Graph-RAG prompt query
./scripts/omni.sh query "How is authentication handled?" [workspace] [k]
```

### Agent Enforcement Shield Setup

Install Omni-Graph rules (`00-omni-graph-mandatory-retrieval.md`), skill (`skills/omni-graph/`), and lifecycle hooks (`hooks.json`):

```bash
# 1. Install to ANY user-specified folder or external codebase:
make setup-agent /path/to/any/codebase
# or
make setup-agent DIR=/path/to/any/codebase

# 2. Install GLOBALLY across all workspaces on machine (~/.gemini/config):
make setup-agent global

# 3. Install to current repository (.agents):
make setup-agent
```

### HTTP Endpoints (Port 8080)

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/api/health` | Aggregated health check for Rust, SurrealDB, and TEI |
| `GET` | `/api/stats` | Ingestion statistics (total nodes, edges, languages, workspaces) |
| `GET` | `/api/workspaces` | Lists all indexed codebases/workspaces with nested worktree hierarchies (`?flat=true` supported) |
| `GET` | `/api/worktrees` | Discovers all active git worktrees across indexed repositories |
| `GET` | `/api/graph?workspace=...` | Full graph topology (`{ nodes: [...], links: [...] }`) |
| `GET` | `/api/search?q=...&workspace=...&k=N` | Vector similarity search on HNSW 384-dim index |
| `GET` | `/api/symbol?name=...&workspace=...` | Symbolic definition lookup (LSP definition, supports struct fields) |
| `GET` | `/api/references?symbol=...&workspace=...`| Call and reference lineage (LSP references) |
| `GET` | `/api/condense?symbol=...&workspace=...&hops=2` | Bounded AST subgraph slice with guaranteed call traces (&lt;1500 tokens) |
| `POST` | `/api/query` | Targeted Graph-RAG retrieval via 1-hop BFS with two-tier token budgeting (&lt;1500 tokens) |
| `POST` | `/api/cluster` | Recomputes Louvain/Leiden community assignments |
| `POST` | `/api/ingest` | Triggers directory AST parsing & indexing (with path security containment) |
| `POST` | `/api/watch/start` | Enrolls directory into real-time notify live sync |

---

## 10. WebGL Visualization Dashboard

Open [http://localhost:3000](http://localhost:3000) to access the interactive Cosmograph WebGL visualization:
- **GPU-accelerated force-directed layout** rendering 100K+ nodes at 60fps
- **Community Galaxies**: Nodes color-coded by functional community clusters
- **Metadata Drawer**: Code preview, symbol signatures, and one-click **"Condense for Agent Prompt (<1500t)"** button.

---

## 11. Specification & Verification

- **Full Specification**: [specs/catalog/omni-graph.md](../../specs/catalog/omni-graph.md)
- **Lifecycle Status**: [STATUS.md](STATUS.md) (`in-progress`)
- **Development Log**: [DEVLOG.md](DEVLOG.md)
- **Validation**: Run `make validate-specs` to verify compliance with repository SDD rules.
