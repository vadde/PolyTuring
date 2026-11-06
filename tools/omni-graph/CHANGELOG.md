# Changelog

All notable changes to this tool will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.6.0-rc1] - 2026-11-05

### Added
- **AI-Native Galaxy Cosmology (R-086, R-087)**: Purged the 1956 human Miller's Law working memory clamp ($K \le 15$) and toxic fallback `sizes.last()` cheat in favor of true natural bounded contexts.
- **Affinity-Floor Invariant (R-086)**: Enforced strict prohibition against merging decoupled systems; communities sharing zero topological coupling ($k_{i, j} = 0$) and zero directory lineage are strictly preserved as independent domains.
- **Newman Modularity Differential Guard (R-087)**: Coarsening maximizes $\Delta Q_{ij} = \frac{k_{ij}}{2m} - \frac{k_i k_j}{2m^2}$ and halts when merges would cause modularity degradation ($\Delta Q < -0.05$).
- **Sub-Linear Scale-Adaptive Subsystem Sizing (R-088)**: Dynamic scale-adaptive ceiling $K_{\text{adaptive}}(N) = \text{clamp}(\lfloor 3.5 \cdot N^{0.35} \rfloor, 5, 150)$ scaling logarithmically from 5 to 150 subsystems.
- **Noise Floor Isolation (R-090)**: Isolated singletons ($< 3$ nodes) gathered into designated utilities namespace, preventing pollution of core business domains.
- **Multi-Scale API & Parameter Controls (R-089)**: Supported `max_galaxies` override on `/api/cluster` and `/api/galaxies`.

### Fixed
- **PolyTuring Database Desynchronization**: Clean atomic purge resolved 13 vs 185 desync across `node.community` and `galaxy` tables; synchronized both views to 65 natural bounded contexts.

## [0.5.0-rc1] - 2026-11-05

### Added
- **Two-Pass Ingestion Engine (R-078)**: Decoupled symbol declaration (Pass 1) from in-memory linkage resolution (Pass 2) via `SymbolTable`, completely eliminating forward-reference edge loss and recovering >84% of missing cross-file call traces.
- **Inbound Caller Preservation in Delta-Sync (R-079)**: Retentive re-indexing preserving existing incoming call edges from foreign files during file updates and re-linking them to matching symbol signatures when line numbers shift.
- **Transactional Read-Fencing & Staging Barrier (R-080)**: Enforced HTTP 423 (Locked) barrier on `/api/cluster`, `/api/galaxies`, and `/api/galaxy/topology` while workspaces undergo active background ingestion, preventing premature dirty reads and cluster thrashing.
- **High-Throughput Cross-File Vector Batching (R-081)**: Buffered vector embeddings across files in batches of up to 256 symbols before dispatching to TEI, reducing HTTP overhead by 25x and accelerating ingestion throughput to sub-60-second bounds.
- **AST Method Receiver Scoping (R-082)**: Parser now preserves root receiver expressions in chained method calls (e.g. `User.objects.create()` emits a `REFERENCES` edge to `User`), restoring object dependency visibility.
- **Native SurrealDB v2 HNSW KNN Operator (R-083)**: Activated native vector KNN syntax `<|k, ef|>` in `search_vector`, enabling true logarithmic $O(\log N)$ vector search over the 384-dimensional cosine HNSW index.
- **Hierarchical Leiden Galaxy Macro-Coarsening (R-084)**: Multi-level meta-node coarsening contracting fine-grained communities into cohesive macro subsystems naturally bounded in $[5, 15]$ without relying on artificial API truncation.
- **Complete Uncle Bob Martin Package Metrics (R-085)**: Automated calculation and persistence of Abstractness ($A = N_{\text{abstract}} / N_{\text{total}}$) and Distance from Main Sequence ($D = |A + I - 1|$) in `GalaxyRecord` and `/api/galaxy/topology`.

### Fixed
- **Pipeline Modularity Desynchronization (R-077)**: Synchronized auto-clustering in `ingestion/mod.rs` and `watcher/mod.rs` to run identical consolidation and coarsening phases, guaranteeing 100% galaxy count parity between automated ingestion and manual UI reclustering.

## [0.4.0] - 2026-11-04

### Added
- **Tokio FIFO Ingestion Task Queue (R-081)**: Bounded in-memory FIFO queue (`tokio::sync::mpsc`) and worker loop serializing background ingestion tasks to eliminate SurrealDB v2 write-conflict aborts (`1001: Commit transaction resulted in a conflict`). Supports atomic workspace purge (`DELETE /ws:<name>`) for 100% clean graphs with zero zombie nodes.
- **3-Layer Hybrid Ripgrep Filter Engine (R-082)**: Dynamic hot-reloading `.gitignore` parsing via `ignore::gitignore::GitignoreBuilder` paired with non-git directory fallback and static Layer 1 invariants (`.git`, `target`, `node_modules`, `dist`, `.gemini`), completely suppressing live-watcher event storms and indexing bloat.
- **Interactive OpenAPI 3.0.3 Specification & Swagger Explorer (R-083)**: Exposed OpenAPI 3.0.3 specification at `/api/openapi.json` and interactive liquid-glass explorer at `/api-docs` with schema definitions for all 9 REST endpoints, latency telemetry, and live execution playground.
- **Path-Based AST Resolution & Condensation (R-084)**: First-class file path and directory lookup in `find_symbols`, `find_references`, and `ContextCondenser::condense` (`file_path = $target OR file_path CONTAINS $target`). Querying a file path (e.g. `src/watcher/mod.rs`) resolves the file node, its contained symbols, and generates a prompt-safe (<1,490 token) neighborhood context slice.
- **Cognitive Intelligence Interpretation Framework (R-085)**: Added 5-step Agent Reconnaissance Ladder, Reference Edge Categorization Matrix, and Token Budget SLA verification protocols to `SKILL.md` across repository, tool, and global locations.

### Fixed
- **MCP Server References Desynchronization**: Resolved schema mismatch in `omni-mcp.py` (`omni_references` tool) by parsing `res.get("references", [])` into Direct Callers, Containment, Inheritance, and Document nodes instead of reading obsolete keys.

## [0.3.0] - 2026-11-03

### Added
- **Topological Containment Backbone (R-063)**: Hierarchical containment nodes (`ws:workspace`, `ws:dir:<path>`, `ws:file:<path>`) and weighted structural edges (`CONTAINS_DIR` $w=0.2$, `CONTAINS_FILE` $w=0.4$, `CONTAINS` $w=0.8$) guaranteeing 100% DAG reachability from root with zero orphan components.
- **Personalized PageRank (PPR) Neighborhood Condenser (R-068, R-069)**: 25-step power-iteration Personalized PageRank ($\alpha = 0.85$) centering extraction around root symbol vector $\mathbf{p}_0$ and strictly bounding context to $\le 1,490$ tokens with clean record ID formatting.
- **Scope-Aware Import Resolution (SAIR) (R-065)**: 3-tier lexical resolution ladder (in-file $\to$ directory/module $\to$ generic verb blacklist) eliminating false-positive call edges for common verbs (`new`, `run`, `init`, `parse`, `handle`).
- **Dual-Tier Content Partitioning (R-067)**: Dedicated storage and vector index partitioning into Code and Documentation tiers, curing Markdown Semantic Shadowing during dense vector retrieval.
- **Hierarchical Galaxy Subsystem Consolidation (R-064)**: 3-criterion hierarchical community rollup enforcing cluster noise floor $\ge 5$ and limiting output to top 15 macro galaxies (reducing payload from 292 KB to 8.6 KB).
- **Polyglot Tree-Sitter Parsers (Java, C, C++) (R-060, R-061)**: Added Tree-sitter parsers extracting Java classes, interfaces, records, methods, and C/C++ structs, classes, functions, namespaces, and `#include` dependencies.
- **Detached Property Filtering (R-062)**: Filtered detached TS/JS object literal properties (`cursor`, `color`, `borderRadius`), eliminating spurious single-node galaxies.
- **Native Model Context Protocol (MCP) Server (R-070, R-078)**: Stdio JSON-RPC 2.0 MCP server (`tools/omni-graph/scripts/omni-mcp.py`) executing hermetically via `uv run --script` with PEP 723 metadata.
- **Ergonomic Subcommand CLI (`scripts/omni`, R-077)**: Lightweight POSIX subcommand CLI with auto-detected workspace resolution and formatted ASCII tables.
- **The Agentic Triad Cognitive Skill (R-079)**: Synchronized `SKILL.md` with Deterministic Decision Flowchart, Master Rosetta Stone, and 9-Scenario Degraded Recovery Playbook.

### Fixed
- **Live Watch Telemetry Synchronization & Guardrail (R-074, R-076)**: Included `status == "syncing"` in watcher status checks, eliminating active sync task cancellation mid-flight and preserving cumulative metrics across re-watches.

## [0.2.2] - 2026-10-09

### Added
- **Targeted SurrealDB Volume Reset Automation (`make reset-graph-db` / `make reset-db`)**: One-command complete storage wipe of `omni_surreal_data` allowing clean re-indexing without deleting the Hugging Face TEI embedding model cache.
- **UI Git Worktree Fleet Switcher & Lineage HUD (`WorktreeFleetModal.tsx`, `App.tsx`)**:
  - Interactive glassmorphic Git Worktree Fleet modal showing active worktree strategies, branches, commit SHAs, file count, and live watch daemon status.
  - Top bar Lineage Breadcrumb HUD: `[Parent Trunk] ──▶ [🌿 Worktree Branch]` with 1-click strategy switching.
  - `Worktrees ▾` indicator pill in top bar displaying strategy count and quick switcher.
- **Polyglot Struct Field & Interface Property Extraction (R-057)**: Extracted `field_declaration` (Rust, Go), `property_signature` (TypeScript), and `public_field_definition`/`field_definition` (TS/JS) into `kind: "field"` nodes linked via `CONTAINS` edges from parent structs/interfaces and `REFERENCES` edges to concrete types.
- **Targeted Neighborhood Subgraph Retrieval (`DbClient::get_neighborhood_subgraph`)**: Localized 1-hop BFS in SurrealDB replaces full-graph HTTP downloads in `GraphRagEngine::query`.
- **Strict Context Budgeting on Graph-RAG Synthesis**: Two-tier budgeting capping symbol definitions at 3,800 chars and reserving $\ge 1,800$ chars for call/import relationships, enforcing the <1500 token SLA.
- **Relational Graph Edge Indexes (`schema.surql`)**: Defined `idx_edge_in`, `idx_edge_out`, `idx_edge_out_ws`, and `idx_edge_in_ws` on `linked_to` table for sub-millisecond edge traversals.
- **Worktree Hierarchy Deduplication**: Pruned adopted worktrees from top-level listing in `/api/workspaces` while adding `?flat=true` parameter for flat backwards compatibility.
- **Robust Multi-Statement SurrealQL Parsing (`extract_sql_arrays`)**: Neutralized `LET` statement index offset fragility by dynamically filtering non-null result arrays.

## [0.2.1] - 2026-10-09

### Added
- **Ambient Live-Watch Auto-Enrollment**: Pre-invocation hooks dynamically detect active indexed workspaces and enroll them in background live-watch without user intervention.
- **Multi-Symbol Polyglot Imports (R-044)**: Discrete `IMPORTS` edges extracted across TS/JS named imports (`import { A, B }`), Python multi-imports (`from m import A, B`), and Rust grouped use statements (`use m::{A, B}`).
- **Two-Tier Subgraph Condenser Token Budgeting (R-028)**: Bounded symbol rendering guaranteeing $\ge 35\%$ of character budget is preserved for structural call/import traces (<1500 token SLA).
- **Ingestion Security Path Containment**: Added canonical path resolution and `allowed_browse_roots()` verification on `POST /api/ingest`, rejecting unauthorized paths with `403 Forbidden`.
- **Zero-Blackout Live Re-indexing**: Reordered live sync pipeline to generate vector embeddings before pruning old nodes, eliminating the 404 deletion window.

## [0.2.0] - 2026-10-09

### Added
- **Ephemeral Branch Fabric (EBF) for Git Worktrees (R-051 - R-058)**:
  - **O(1) Worktree Pointer Peeking (R-051)**: Detects `.git` worktree pointers in $O(1)$ without spawning subshells or executing `git` CLI, resolving parent repository path, branch name, and lineage.
  - **Enforced Auto-Watch Enrollment (R-052)**: `/api/ingest` automatically starts the live watch daemon (`watch: true` default), ensuring worktrees edited by coding agents are immediately kept fresh without manual watcher setup.
  - **Seed-Preserving Prior Inheritance (R-053)**: Worktree LPA galaxy clustering inherits community seeds from the parent workspace, preserving 99% cluster stability across branches.
  - **Dead Worktree Garbage Collection (R-054)**: `/api/workspaces` automatically sweeps and purges deleted worktrees from SurrealDB, preventing phantom zombie workspaces.
  - **Worktree Hierarchy & Lineage (R-055)**: `/api/workspaces` nests active worktrees under parent repositories (`worktrees: [...]`) while preserving backward-compatible flat workspace listing.
  - **Worktree Visualizer & Branch Navigation (R-056)**: Cosmograph React UI displays worktree indicators, branch tags, and nested worktree switchers in the workspace selector.
  - **Proactive Agent Hook & Remediation (R-057)**: Pre-invocation and pre-tool hooks detect unindexed worktrees and emit one-command self-healing ingest banners.
  - **Rebase Storm Surge Suppression (R-058)**: Idempotent watch registration and event debounce windows prevent CPU thrashing during multi-file git rebases.
  - **Top-Down Worktree Fleet Auto-Cascade & Dynamic Sync (R-059)**: Ingesting a parent trunk automatically discovers all child worktrees in `.git/worktrees/*`, resolves sibling roots, cascades background ingestion, inherits parent LPA seeds, and enrolls concurrent watchers. Added `POST /api/worktrees/sync`, `make sync-worktrees`, and `omni.sh worktrees sync`.

## [0.1.1] - 2026-10-07

### Added
- **Full Polyglot AST Grammar (R-031)**:
  - Go: method declarations (`method_declaration`) with receiver `DECLARES` edges, structs and interfaces (`type_spec`), constants (`const_spec`), and imports (`import_spec`).
  - TypeScript/JavaScript: lexical arrow functions (`const x = () => ...`) and function expressions under `variable_declarator`.
  - TypeScript: interfaces (`interface_declaration`), type aliases (`type_alias_declaration`), and enums (`enum_declaration`).
  - Rust: `impl_item` blocks emitting `DECLARES` edges from target structs to methods, `enum_item`, and `trait_item`.
  - Python: coroutines (`async_function_definition`).
- **Relational Delta Sync Integrity (R-032)**: Single-file delta re-indexing only deletes outbound edges (`in IN $nodes`), preserving inbound caller edges to eliminate graph island dissolution.
- **Two-Tier Exact Symbol Retrieval (R-033)**: SurrealQL `ORDER BY (label = $name) DESC, label ASC LIMIT 20` prioritizes exact symbol matches over substrings.
- **Deterministic LPA Tie-Breaking (R-034)**: Tie-break equal community edge weights using minimum label ID for 100% deterministic galaxy assignments across runs.
- **Scoped Call-Edge Resolution (R-035)**: AST call edges resolve against local source file and directory before falling back to workspace, eliminating cross-package collisions.
- **Pre-Computed Galaxy Graph-RAG Retrieval (R-036)**: `GraphRagEngine::query` reads pre-computed `galaxy` table records directly instead of re-summarizing entire workspace in-memory.
- **Context-Enriched Embedding Payloads (R-037)**: Prepend structured header `[{language}] {kind} {label} in {file_path}\n{text}` before generating TEI vector embeddings.
- **Unified Record ID Normalization (R-050)**: Standardized SurrealDB ID sanitization via `clean_record_id`, stripping `node:`, `⟨...⟩`, and backticks across all storage/query paths, eliminating all 114 orphaned edges in SurrealDB.
- **Cross-File Receiver & Impl Target Scoping (R-045)**: Resolved 3-part source IDs (`ws:file:Struct`) locally in file, then workspace scope; falls back to empty relation array `[]` instead of creating phantom node records.
- **Cascade Edge Deletion on Node Pruning (R-046)**: Cascaded edge deletion across both inbound (`out IN $nodes`) and outbound (`in IN $nodes`) directions during file re-indexing/deletion, ensuring zero dangling graph pointers.
- **Stale File Pruning in Batch Ingestion (R-047)**: Pre-prune file records before batch insertion in `ingest_directory` to prevent duplicate zombie nodes when file contents shift.
- **Trait & Class Inheritance Relationship Extraction (R-048)**:
  - Rust: `impl Trait for Struct` emits `IMPLEMENTS` edges.
  - Python: class inheritance from `superclasses` emits `EXTENDS` edges.
  - TypeScript/JavaScript: `class_heritage` clauses emit `EXTENDS` and `IMPLEMENTS` edges.
- **Rust Macro Definition & Invocation Indexing (R-049)**: Tree-sitter `macro_definition` emits `kind: "macro"`; `macro_invocation` emits `CALLS` edge.
- **Universal Invocation Dispatch (R-040)**: Extracted Python `call`, Rust `method_call_expression`, and TS/JS `new_expression` as `CALLS` edges.
- **Go Interface Method AST Extraction (R-041)**: Extracted `method_spec` inside Go `interface_type` as `kind: "method"` with `DECLARES` edge from parent interface.
- **Direct Pre-Computed Galaxy Delivery (R-042)**: Served `/api/galaxies` directly from `galaxy` table with zero runtime graph deserialization overhead.
- **Localized Subgraph Neighborhood in Condenser (R-043)**: `/api/condense` queries localized 1-2 hop neighborhood in SurrealDB instead of whole-graph deserialization.
- **First-Class Import Structural Linking (R-044)**: Emitted `CONTAINS` edge from parent file to `import` node and `IMPORTS` edge to imported module/symbol.
- **Galaxy Subsystem Disambiguation (R-038)**: Automatically append primary cohesive symbol on directory name collisions to eliminate duplicate galaxy titles.

## [0.1.0] - 2026-09-24

### Fixed
- Fixed ingestion bottleneck on large datasets/dumps by pruning non-code directories (`solutions`, `data`, `logs`), applying 512KB file limit, and capping fallback blocks.
- Fixed SurrealDB v2 HNSW delete timeout during workspace re-indexing by temporarily detaching the vector index during bulk record purges.
- Fixed Axum synchronous ingestion cancellation by running ingestion pipelines in detached Tokio tasks.
- Fixed duplicate and uncanonical workspace listings in `/api/workspaces` by coalescing legacy `workspace` into `tool-scripts` and merging AST node/file counts.
- Filtered out empty, `default`, and `global` pseudo-workspaces from API output.
- Hardened React UI `loadWorkspaces` with client-side Set deduplication and filter checks.

### Added
- Initial tool scaffolding with full SDD specification
- Project genesis: 24 functional requirements, 14 acceptance criteria, 11 NFRs
- Research integration from Graphify, CodeGraph, Serena, ECC reference architectures
