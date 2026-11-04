---
tool: omni-graph
status: in-progress
last_session: 2026-11-03
last_agent: "@antigravity"
blockers: []
---

# 🧭 Tool Context — omni-graph

> **Read this file FIRST before any action on this tool.**
> This is the single source of truth for the current state of this tool.
> It is the contract between development sessions.

---

## Current State

- **SDLC Phase**: `in-progress`
- **Version**: 0.3.0
- **Language**: Rust (backend orchestrator) + React/TypeScript (frontend UI)
- **Category**: GenAI
- **Architecture**: Multi-container Docker microservices (4 services on `omni-net`)

### Progress

| Metric | Done | Total | Status |
|--------|------|-------|--------|
| FIFO Ingestion Task Queue & Concurrency Protection | Complete | — | 🟢 Tokio mpsc FIFO queue serializing SurrealDB writes and eliminating write conflicts |
| Atomic Clean Invariant & Purge Integration | Complete | — | 🟢 purge_workspace automatically executed before re-indexing; removed manual toggle |
| OpenAPI 3.0 Specification & Route (/api/openapi.json) | Complete | — | 🟢 Complete 27-endpoint schema covering all Omni-Graph endpoints |
| Cosmic Liquid-Glass Swagger Explorer (/api-docs) | Complete | — | 🟢 Year 2500 frosted glass UI, Try-It-Live console, latency timer, cURL copy |
| Ingestion Palette Non-Blocking Multi-Repo Actions | Complete | — | 🟢 Directory row badges (Ingesting, Queued, Indexed), quick Queue/Ingest actions, no button lockout |
| Spec Requirements (R-XXX) | 78 | 80 | 🟢 97.5% Complete (2 Deferred to v0.3.1: R-066, R-072) |
| Acceptance Criteria (AC-XXX) | 35 | 37 | 🟢 94.6% Complete (2 Deferred: AC for R-066, R-072) |
| Non-Functional (NF-XXX) | 15 | 15 | 🟢 100% Complete |
| Test Coverage | 82/82 Unit & Integration Tests | 82 | 🟢 100% Verified in Docker |
| Topological Containment Backbone (R-063) | Complete | — | 🟢 Workspace, Directory & File DAG backbone with zero disconnected vertices |
| Watcher Syncing Guardrail & Telemetry Parity (R-074, R-076) | Complete | — | 🟢 Included `status == "syncing"` in is_watching, eliminating cancellation in flight |
| PPR Neighborhood Condenser & Clean Record IDs (R-068, R-069) | Complete | — | 🟢 Power iteration Personalized PageRank (alpha=0.85) & <1500 strict token ceiling |
| Watcher Telemetry Synchronization (R-073 to R-076) | Complete | — | 🟢 Dynamic in-memory status injection & idempotent watch |
| Polyglot AST Parsers (Java, C, C++, R-060, R-061) | Complete | — | 🟢 Tree-sitter Java/C/C++ classes, functions, includes & inheritance |
| Galaxy Micro-Cluster Consolidation (R-064) | Complete | — | 🟢 Topological & spatial rollup, size floor >=5, 15 macro galaxies |
| Scope-Aware Import Resolution (SAIR, R-065) | Complete | — | 🟢 In-file -> dir/module -> generic verb protection avoiding cross-package contamination |
| Content-Tier Partitioning (Code vs Docs, R-067) | Complete | — | 🟢 Code-first search curing Markdown Semantic Shadowing |
| Native Model Context Protocol (MCP) Server (R-070, R-078) | Complete | — | 🟢 Zero-token shell-free stdio JSON-RPC 2.0 MCP server (`omni-mcp.py`) |
| Unified Ergonomic CLI (`scripts/omni`, R-077) | Complete | — | 🟢 Subcommand CLI auto-detecting workspace and formatting clean outputs |
| Agentic Triad Architecture & Cognitive Bridge (R-079) | Complete | — | 🟢 Integrated SKILL.md (Brain) + MCP (Nervous) + CLI (Hands) |
| Negative Ingestion Guardrail & Zero Redundant Ingest | Complete | — | 🟢 Suppressed /api/ingest in banners for indexed repos |
| Accurate Stdin Workspace Resolution in Lifecycle Hooks | Complete | — | 🟢 CLI payload workspacePaths prioritization |
| Hook Invocation Idempotency (Zero Duplicate Banners) | Complete | — | 🟢 Token marker eliminating dual local/global banner injection |
| Top-Down Worktree Auto-Cascade & Dynamic Sync (R-059) | Complete | — | 🟢 Discover .git/worktrees/*, auto-cascade ingest & watchers |
| Safe SurrealDB Volume Reset & Automation | Complete | — | 🟢 make reset-graph-db / reset-db preserving TEI cache |
| Git Worktree Fleet Switcher & Lineage HUD | Complete | — | 🟢 Interactive Strategy Modal, Branch HUD & Top-Bar Breadcrumbs |
| Polyglot Struct Field & Interface Properties (R-057) | Complete | — | 🟢 Rust, Go, TS, JS Field Nodes with CONTAINS & REFERENCES Edges |
| Targeted Graph-RAG Neighborhood Subgraph | Complete | — | 🟢 Localized 1-Hop BFS in SurrealDB Eliminating Full-Graph Fetch |
| Graph-RAG Strict Context Budgeting (<1500 tokens) | Complete | — | 🟢 Two-Tier Budgeting with Guaranteed Structural Edge Capacity |
| Relational Graph Edge Indexes (linked_to) | Complete | — | 🟢 idx_edge_in/out/ws Built on SurrealDB |
| Worktree Hierarchy Deduplication | Complete | — | 🟢 Pruned Top-Level Worktrees with ?flat=true Fallback |
| Multi-Statement SurrealQL Parsing Resilience | Complete | — | 🟢 extract_sql_arrays Eliminates LET Statement Offset Fragility |
| Ambient Live-Watch Auto-Enrollment | Complete | — | 🟢 Zero-Friction Pre-Invocation Hook & Live Watch Active |
| Multi-Symbol Polyglot Imports (R-044) | Complete | — | 🟢 Discrete IMPORTS Edges across TS/JS, Python, Rust, Go |
| Subgraph Condenser Token Budget (<1500 tokens, R-028) | Complete | — | 🟢 Two-Tier Budgeting with Guaranteed Call Traces |
| Ingest Security Path Containment | Complete | — | 🟢 Canonical Path Checking & 403 Forbidden Guard on /api/ingest |
| Zero-Blackout Live Re-indexing | Complete | — | 🟢 Embedding Pre-computation Eliminating 404 Window |
| Ephemeral Branch Fabric (R-051 - R-058) | Complete | — | 🟢 O(1) Worktree Peeking, Seed LPA, Auto-Watch, Auto-Purge |
| Zero Orphaned Edges (R-045, R-046, R-050) | Complete | — | 🟢 Verified in SurrealDB (`in.id IS NONE OR out.id IS NONE` = 0) |
| Polyglot AST Grammar (R-031, R-040, R-041) | Complete | — | 🟢 Go methods/structs/consts/imports, TS arrow fns/types, Rust impl/enum/trait, Python async |
| Trait & Class Inheritance (R-048) | Complete | — | 🟢 Rust `impl Trait for Struct`, Python inheritance, TS `extends`/`implements` |
| Rust Macros (R-049) | Complete | — | 🟢 `macro_definition` and `macro_invocation` call graph indexing |
| Cascading Node Pruning (R-046) | Complete | — | 🟢 Symmetric inbound and outbound edge pruning eliminating dangling pointers |
| Stale File Batch Ingestion (R-047) | Complete | — | 🟢 Pre-delete stale file records before re-inserting to prevent ghost nodes |
| Relational Delta Sync Integrity (R-032) | Complete | — | 🟢 Inbound Call Edges Preserved Across Single-File Edits |
| Two-Tier Exact Symbol Retrieval (R-033) | Complete | — | 🟢 SurrealQL `ORDER BY is_exact DESC, label ASC` Pre-Limit |
| Deterministic LPA Tie-Breaking (R-034) | Complete | — | 🟢 Cluster Stability via Minimum Label ID Tie-Breaks |
| Scoped Call-Edge Resolution (R-035) | Complete | — | 🟢 File -> Dir -> Workspace Fallback Resolution |
| Pre-Computed Galaxy Graph-RAG Retrieval (R-036) | Complete | — | 🟢 Direct DB Table Query without Full-Graph Re-Summarization |
| Context-Enriched Vector Embeddings (R-037) | Complete | — | 🟢 `[lang] kind label in path\ntext` Header Payload |
| Galaxy Subsystem Disambiguation (R-038) | Complete | — | 🟢 Primary Member Symbol Appended on Dir Collisions |
| First-Class Markdown Retrieval & Hyperlinks | Complete | — | 🟢 Line-Precise Vector Search, LINKS_TO & REFERENCES Edges |
| Calibrated Documentation Guardrail (450 lines) | Complete | — | 🟢 Live Across Machine & .gemini/config/scripts |
| Dynamic Live Galaxy Clustering (R-029) | Complete | — | 🟢 Seed-Preserving LPA & Quiescent Re-Clustering |
| Agent Boundary Contracts (R-030) | Complete | — | 🟢 /api/galaxy/boundary & /api/galaxy/topology Live |
| Agent Relationship Augmentation (POST /api/relationships) | Complete | — | 🟢 Inferred Runtime Linking & Virtual Component Synthesis |
| Refresh Ingestion & Non-Blocking Ingest UX | Complete | — | 🟢 Atomic Purge, Clean Re-ingest & Progress Toast |
| 3-Tier Omniverse Reconnaissance Ladder | Complete | — | 🟢 Live Across Machine & Ephemeral Prompt Banners |
| UI/UX Macro Topology Hub & Boundary Inspector | Complete | — | 🟢 Live on Port 3000 (Topology Hub & Inspector) |
| Agent Analytics & LSP Telemetry | Complete | — | 🟢 Live on Port 3000 & 8080 (Dual-Source Telemetry) |
| Dynamic Live Delta Sync & Watch HUD | Complete | — | 🟢 Live on Port 3000 & 8080 |
| Active Ingestion HUD & Multi-Repo Tracker | Complete | — | 🟢 Live on Port 3000 & 8080 (Survives Hard Browser Refresh) |
| Fail-Safe Directory Browser & Escape Hatch | Complete | — | 🟢 Non-blocking UI, 3.5s Timeout, AbortController & Recovery Actions |
| Workspace Meta & Path Auto-Resolution | Complete | — | 🟢 Verified & Live in Docker |
| Workspace Deduplication & Sanitization | Complete | — | 🟢 Live on Port 3000 & 8080 |
| Session-Scoped Recon Gate & Sequence Enforcer | Complete | — | 🟢 Live Machine-Wide (Workspace & Global) |
| Unbypassable 150-Line Gate & Rule 00 | Complete | — | 🟢 Live Machine-Wide (Polyglot) |
| Documentation | Complete | — | ✅ Complete |

---

## Quick Links

| Resource | Path | Notes |
|----------|------|-------|
| Specification | [`specs/catalog/omni-graph.md`](../../specs/catalog/omni-graph.md) | Full requirements (28 R-XXX, 18 AC-XXX, 11 NF-XXX) |
| Hardening Spec | [`specs/catalog/omni-graph-hardening.md`](../../specs/catalog/omni-graph-hardening.md) | Forensic fixes & worktree fleet sync (R-031 to R-059) |
| Next-Gen Spec (v0.3.0) | [`specs/catalog/omni-graph-nextgen.md`](../../specs/catalog/omni-graph-nextgen.md) | Scientific architecture, Leiden Graph-RAG, Two-Tower RRF & Polyglot (R-060 to R-072) |
| Status | [`STATUS.md`](STATUS.md) | SDLC phase + traceability matrix |
| Dev Log | [`DEVLOG.md`](DEVLOG.md) | Session-by-session journal |
| Changelog | [`CHANGELOG.md`](CHANGELOG.md) | Version history |
| Source | [`src/`](src/) | Rust orchestrator source |
| UI | [`ui/`](ui/) | React + TypeScript + Cosmograph frontend |
| Tests | [`tests/`](tests/) | Test suite |
| Examples | [`examples/`](examples/) | Runnable demos |

---

## Architecture Summary

```
4 Docker containers on omni-net:
├── rust-app (Axum + tree-sitter)     → Port 8080 — Graph API + telemetry ingestion
├── graph-db (SurrealDB v2)           → Port 8000 — Multi-model DB + HNSW vector + agent telemetry
├── embedding-engine (HuggingFace TEI) → Port 8081 — bge-small-en-v1.5 (384-dim)
└── graph-ui (React + Cosmograph)     → Port 3000 — WebGL force-directed visualization + Analytics
```

### Key Technical Decisions

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-10-07 | Unified Record ID Normalization (R-050) | Centralized `clean_record_id` stripping `node:`, `⟨...⟩`, and backticks across all storage/edge queries; standardizes fallback/config parser node IDs to `{}:{}:{}` format, eliminating all 114 orphaned edges in SurrealDB |
| 2026-10-07 | Scoped Cross-File Receiver & Impl Target Scoping (R-045) | 3-part source IDs (`ws:file:Struct`) resolve locally in file, then workspace scope; falls back to empty relation array `[]` instead of creating phantom node IDs |
| 2026-10-07 | Cascading Edge Pruning on Node Pruning (R-046) | In `delete_file`, cascade delete both outbound (`in IN $nodes`) and inbound (`out IN $nodes`) edges, eliminating dangling pointers and WebGL UI crashes |
| 2026-10-07 | Stale File Pruning in Batch Ingestion (R-047) | Call `delete_file` prior to re-indexing stale files in `ingest_directory` to prevent duplicate symbol nodes and shifted line numbers |
| 2026-10-07 | Trait & Class Inheritance Graph Extraction (R-048) | Rust `impl Trait for Struct` emits `IMPLEMENTS`; Python `class A(B)` emits `EXTENDS`; TS/JS `extends` and `implements` emit `EXTENDS` and `IMPLEMENTS` |
| 2026-10-07 | Rust Macro Definition & Invocation (R-049) | Tree-sitter `macro_definition` emits `kind: "macro"`; `macro_invocation` emits `CALLS` edge |
| 2026-10-07 | Polyglot Tree-sitter AST Hardening (R-031) | Added Go struct methods, structs, interfaces, consts, TS arrow functions, type aliases, interfaces, Rust `impl_item` (`DECLARES` edge), `enum_item`, `trait_item`, and Python async coroutines |
| 2026-10-07 | Relational Delta Sync Outbound Only (R-032) | Reindexing a single file prunes only `in IN $nodes` (outbound edges), preserving inbound caller edges to prevent graph fragmentation |
| 2026-10-07 | SurrealQL Pre-Limit Exact Ordering (R-033) | Injected `(label = $name) AS is_exact` with `ORDER BY is_exact DESC, label ASC LIMIT 20` directly in query to ensure exact matches are never masked by substring hits |
| 2026-10-07 | Scoped Call-Edge Resolution (R-035) | Resolved AST call links locally within the source file first, preventing common function names from cross-linking into foreign packages |
| 2026-10-07 | Pre-Computed Galaxy Graph-RAG Retrieval (R-036) | Switched `GraphRagEngine::query` to read pre-computed `galaxy` table records instead of pulling the entire workspace graph over HTTP |
| 2026-10-07 | Context-Enriched Vector Embeddings (R-037) | Prepended structured header `[{language}] {kind} {label} in {file_path}\n{text}` before passing to TEI, boosting semantic relevance |
| 2026-10-02 | Nginx 600s Proxy Timeout & 300s SurrealDB Client | Configured 600s proxy timeouts in `nginx.conf` and 300s in `DbClient` with indexed `linked_to` deletion, resolving the "Unknown error" 504 drops on large repos (DSA) |
| 2026-10-02 | Atomic Workspace Purge & Non-Blocking Ingest UX | Implemented 4-step atomic purge in SurrealDB (`DELETE linked_to -> DELETE node -> DELETE galaxy`), cleared in-memory staleness hashes, replaced blocking modal with instant dismiss into non-blocking progress toast with 5-minute timeout window |
| 2026-09-30 | Node Inspector Breadcrumbs & Caller Stack | Added `nodeHistory` stack, interactive breadcrumbs, `Back to [prev]` button, and `Return to caller` action in BoundaryContractCard for one-click return from foreign callers |
| 2026-09-30 | Progressive Windowing & Flexbox Isolation | SubsystemTopologyHub uses `visibleCount` progressive slicing and `flexShrink: 0` to prevent flex collapse on 250+ cluster workspaces (QuarkDock) |
| 2026-09-30 | Live Watch HUD Consolidation | Unified live file watching and background modularity under single `LIVE WATCH` HUD; removed redundant top-bar chip |
| 2026-09-30 | Dynamic Live Galaxy Clustering (Dual-Phase) | Instant single-file community inheritance in reindex_file + quiescent (3.5s) background re-clustering using seed-preserving weighted LPA |
| 2026-09-30 | Agent Boundary Contracts (/api/galaxy/boundary) | Exposes macro subsystem containment, cross-galaxy callers, Robert C. Martin coupling metrics (Ca, Ce, Instability I), and actionable risk advice |
| 2026-09-26 | Dual-Source Telemetry (SurrealDB + Transcripts) | Blends real-time SurrealDB `agent_api_call` events with IDE session logs using `max(t, db)` per capability/workspace |
| 2026-09-24 | HNSW index (not MTREE) for vector similarity | SurrealDB v2 production-ready ANN index; MTREE deprecated/experimental |
| 2026-09-24 | TEI CPU-only in Docker (not Metal) | macOS Docker doesn't support GPU passthrough; Metal only via native Homebrew |
| 2026-09-24 | `BAAI/bge-small-en-v1.5` (384-dim) | Lightweight, high-quality BERT-based embeddings, ARM64 SIMD optimized |
| 2026-09-24 | Cosmograph for WebGL graph viz | GPU-accelerated force-directed layout; handles 100K+ nodes at 60fps |
| 2026-09-24 | Edge taxonomy: EXTRACTED vs INFERRED | Following Graphify's proven categorization pattern |

---

## Next Steps

> _What to do when you pick this tool up. Updated at the end of every session._

1. Review Next-Gen Specification [`specs/catalog/omni-graph-nextgen.md`](../../specs/catalog/omni-graph-nextgen.md) (R-060 to R-072) and multi-disciplinary retrospective.
2. Execute Phase 1: Integrate `tree-sitter-java` and `tree-sitter-c`/`tree-sitter-cpp` into `Cargo.toml` and implement topological containment edges (`CONTAINS_DIR`, `CONTAINS_FILE`).
3. Execute Phase 2: Replace flat LPA with Hierarchical Leiden Community Detection to eliminate 936 single-node galaxies and enforce 5–15 macro subsystems.
4. Execute Phase 3: Scope-Aware Import Resolution (SAIR) and Content Tier Partitioning (`code_node` vs `doc_node`).
5. Execute Phase 4: Two-Tower Hybrid Search (Tantivy BM25 + HNSW RRF) and Strict Budget Context Condenser.

---

## Active Blockers

> _Anything preventing progress. Remove when resolved._

| Blocker | Since | Impact | Resolution Plan |
|---------|-------|--------|-----------------|
| _None_ | — | — | — |

---

## Reference Projects Studied

| Project | Key Insight |
|---------|-------------|
| [Graphify](https://github.com/Graphify-Labs/graphify) | Tree-sitter AST → knowledge graph; Leiden community detection; `EXTRACTED` vs `INFERRED` edges |
| [CodeGraph](https://github.com/colbymchenry/codegraph) | Rust parser kernel; FSEvents file watcher with debounce; content-hash staleness |
| [Serena](https://github.com/oraios/serena) | LSP-wrapped symbolic operations; agent-first tool design |
| [ECC](https://github.com/affaan-m/ECC) | PreToolUse hooks; content-hash caching; token budget monitoring |
| [Microsoft GraphRAG](https://github.com/microsoft/graphrag) | Hierarchical Leiden community detection & macroscopic summary aggregation |
| [RAGFlow](https://github.com/infiniflow/ragflow) | Deep structural chunking over raw lexical chunking |
| [Sourcegraph SCIP](https://github.com/sourcegraph/scip) | Standardized language-agnostic code intelligence protocol; separates definitions from occurrences |
| [Google Kythe](https://github.com/kythe/kythe) | Compiler-level semantic graph schema (anchor, defines, childof, ref) for multi-million line monorepos |
| [DeusData codebase-memory-mcp](https://github.com/DeusData/codebase-memory-mcp) | Embedded SQLite C knowledge graph exposing 15 MCP tools for agentic call graph traversal |
| [ast-grep](https://github.com/ast-grep/ast-grep) | Blazing-fast Tree-sitter structural pattern matcher (`$A.foo($B)`) for syntactic search & rewrites |
| [RepoGraph (ICLR 2024)](https://github.com/microsoft/RepoGraph) | Fine-grained line-level code graph navigation plugin boosting LLM agent repository task accuracy |

---

## Traceability Summary

> _Quick reference: which requirements are done, which aren't._
> _Detailed traceability lives in STATUS.md and the spec._

| Req ID | Description | Impl | Tested |
|--------|-------------|------|--------|
| R-001 | Docker Compose (4 services) | ✅ | ✅ |
| R-002 | Health-check dependencies | ✅ | ✅ |
| R-003 | Tree-sitter multi-lang parsing | ✅ | ✅ |
| R-004 | TEI 384-dim embeddings | ✅ | ✅ |
| R-005 | Node storage (SurrealDB) | ✅ | ✅ |
| R-006 | Graph edges (linked_to) | ✅ | ✅ |
| R-007 | HNSW vector index | ✅ | ✅ |
| R-008–R-010 | REST API (graph, search, ingest) | ✅ | ✅ |
| R-011 | File staleness tracking | ✅ | ✅ |
| R-012 | Community detection | ✅ | ✅ |
| R-013–R-018 | React + Cosmograph UI | ✅ | ✅ |
| R-019 | Schema auto-init | ✅ | ✅ |
| R-020–R-024 | Health, stats, orphans, DNS | ✅ | ✅ |
| R-025 | Agent Enforcement Layer (hooks.json) | ✅ | ✅ |
| R-026 | Omni-Graph Native Agent Interface (MCP / Skill) | ✅ | ✅ |
| R-027 | Automated Agent Setup CLI & Bootstrapper | ✅ | ✅ |
| R-028 | Subgraph Context Condenser (<1500 tokens) | ✅ | ✅ |
| R-029 | Dynamic Live Galaxy Clustering & LPA Seed Persistence | ✅ | ✅ |
| R-030 | Subsystem Boundary Contracts & Agent Actionable Advice | ✅ | ✅ |

---

_Last updated: 2026-10-13_
