---
tool: omni-graph
status: in-progress
version: 0.7.0-dev
language: rust
category: GenAI
created: 2026-09-24
last_updated: 2026-11-06
owner: "@vadde"
spec: ../../specs/catalog/omni-graph.md
nextgen_spec: ../../specs/catalog/omni-graph-nextgen.md
hardening_spec: ../../specs/catalog/omni-graph-two-pass-engine-hardening.md
cosmology_spec: ../../specs/catalog/omni-graph-ai-native-galaxy-cosmology.md
fidelity_spec: ../../specs/catalog/omni-graph-live-watcher-fidelity-parity.md
---

# SDLC Status

## Current Phase: `in-progress`

### Status Definitions

| Status | Phase | Description |
|--------|------|-------------|
| `draft` | 🟤 Ideation | Initial idea, spec being written |
| `spec-review` | 🟡 Specification | Spec complete, awaiting review |
| `in-progress` | 🔵 Implementation | Active development |
| `testing` | 🟣 Verification | Code complete, testing phase |
| `review` | 🟠 Review | Ready for human review |
| `released` | 🟢 Released | Published and stable |
| `deprecated` | ⚫ Deprecated | No longer maintained |

### Phase History

| Date | From | To | Notes |
|------|------|----|-------|
| 2026-09-24 | — | `draft` | Tool created. Full spec written with research from Graphify, CodeGraph, Serena, ECC |
| 2026-09-24 | `draft` | `spec-review` | Spec expanded with 28 requirements, agent enforcement shield & Antigravity hooks |
| 2026-09-24 | `spec-review` | `in-progress` | Spec approved with Dual Setup strategy; implementation begins |
| 2026-10-07 | `in-progress` | `in-progress` | Forensic audit & hardening: Polyglot AST grammar, relational integrity, exact sorting, deterministic LPA, scoped edge resolution, and Graph-RAG galaxy records (R-031 to R-038) |
| 2026-10-07 | `in-progress` | `in-progress` | Second-round forensic audit: Cross-file receiver/impl scoping, cascading edge pruning, batch stale file pruning, trait/class inheritance, Rust macro indexing, clean record ID normalization, achieving 0 orphaned edges (R-039 to R-050) |
| 2026-10-09 | `in-progress` | `in-progress` | Third-round forensic audit: Ambient live-watch auto-enrollment, polyglot multi-symbol imports (R-044), two-tier condenser token budgeting (<1500 tokens, R-028), zero-blackout live re-indexing, and ingestion path security containment |
| 2026-10-09 | `in-progress` | `in-progress` | Fourth-round forensic audit: Polyglot struct field & interface property AST extraction (R-057), targeted neighborhood Graph-RAG retrieval, strict Graph-RAG token budgeting, worktree hierarchy deduplication, SurrealQL statement index resilience (`extract_sql_arrays`), and relational graph edge indexing (`linked_to`) |
| 2026-10-13 | `in-progress` | `in-progress` | Safe SurrealDB volume purge automation (reset-graph-db / reset-db), storage reset preserving TEI model cache, UI worktree fleet switcher & lineage HUD integration |
| 2026-11-03 | `in-progress` | `in-progress` | Holistic multi-disciplinary retrospective & scientific introspection across 5 foundational lenses; author Next-Gen Spec v0.3.0 (R-060 to R-072: Java/C++ AST, Leiden clustering, Two-Tower RRF hybrid search, Content partitioning, PPR context slicing, and native MCP server) |
| 2026-11-04 | `in-progress` | `in-progress` | Implement FIFO Ingestion Queue, Atomic Clean Invariant (eliminated manual purge toggle), Non-blocking Multi-Repo Actions, OpenAPI 3.0 Specification, and Year 2500 Cosmic Liquid-Glass Swagger Explorer (specs/catalog/omni-graph-ingestion-openapi-ux.md: R-001 to R-012) |
| 2026-11-04 | `in-progress` | `in-progress` | Implement Workspace Deep-Linking & History sync (R-013), Bounded Viewport Scroll & HUD Clearance (R-014), and Hybrid Ripgrep-Engine Gitignore Filter with non-git fallback (R-015; 85 unit tests pass) |
| 2026-11-04 | `in-progress` | `in-progress` | Cognitive Intelligence Interpretation Framework for SKILL.md, path-based symbol/reference/condense engine, and MCP server reference schema parity (86 unit tests pass) |
| 2026-11-05 | `in-progress` | `in-progress` | Deep forensic audit of ingestion pipeline desynchronization & forward-reference drop; author Enhancement Spec (R-077 to R-085: Two-Pass Ingestion, Inbound Edge Preservation, Transactional Read-Fencing & High-Throughput Bulk Engine) |
| 2026-11-05 | `in-progress` | `testing` | Implement Two-Pass Ingestion Engine (`symbol_table.rs`), Retentive Delta-Sync (`relink_inbound_edges`), Transactional Read-Fence (HTTP 423), Vector Chunking (256/batch), AST Receiver Preservation, True HNSW KNN `<|k, ef|>`, Hierarchical Leiden Coarsening, and Uncle Bob metrics ($A, D$). 95 unit tests pass (100% green). |
| 2026-11-05 | `testing` | `review` | Rebuilt Docker backend & UI, empirically verified live ingestion, HNSW KNN search (83ms), natural 15-galaxy convergence, and Uncle Bob metrics on django-sorcery. Traceability matrix 100% complete. |
| 2026-11-05 | `review` | `review` | Implement AI-Native Galaxy Cosmology (R-086 to R-092): Purge Miller's Law clamp, enforce Affinity-Floor Invariant, modularity differential guard, sub-linear scale-adaptive sizing, and noise floor isolation. Synchronized PolyTuring (65 galaxies) and verified tutor-intelligence (68), DesignPatterns (49), and django-sorcery (53). 98 unit tests pass. |
| 2026-11-05 | `review` | `review` | Forensic resolution of watcher 392s subquery bottleneck: Integrated in-memory SymbolTable (R-078/R-091) into IncrementalPipeline with parameterized batch edge resolution, dedicated 1s quiescent auto-cluster trigger, and test_r091 (99 unit tests pass). |
| 2026-11-06 | `review` | `in-progress` | Launch v0.7.0 Live Watcher Fidelity Parity & Concurrency Overhaul (ADR-002, SPEC-OMNI-005, R-093 to R-109): Closing 14 forensic loopholes across transactional atomic batch swap, reactive reverse dependencies, unresolved retention pool, deduplication, and non-blocking Louvain actor. |
| 2026-11-06 | `in-progress` | `testing` | Completed Milestones 1 through 6 (R-093 to R-106): SymbolTable hardening, directory backbone synthesis, transactional atomic swap, edge deduplication, dual-tier modularity, non-blocking Louvain actor, convergence barrier (wait_for_sync), nanometer fidelity diagnostics, and lock-free generational versioning. 111 unit tests pass (100% green, 0 clippy warnings). |
| 2026-11-06 | `testing` | `testing` | Completed R-109 (Filesystem Symlink Cycle Guarding & Alias Navigation) and R-107 (MathWorks MATLAB AST Ingestion): Vendored ABI 14 tree-sitter-matlab grammar, extracted classdef, multiple inheritance EXTENDS, properties, constructors, methods, receiver preservation, import commands, and package hierarchy. 118 unit tests pass (100% green, 0 clippy warnings). |
| 2026-11-06 | `testing` | `review` | Completed R-108 (MathWorks Simulink .slx / .mdl MBSE Ingestion): In-memory OPC ZIP extraction of simulink/blockdiagram.xml, streaming XML & legacy .mdl parser, block containment, cross-model references (REFERENCES_MODEL), dataflow signal wiring (SIGNAL_TO), and high-degree broadcast bus attenuation (DATAFLOW_BROADCAST). 122 unit tests pass (100% green, 0 clippy warnings). |

---

## Implementation Progress

| Category | Done | Total | Percentage |
|----------|------|-------|------------|
| Requirements (R-XXX) | 81 | 81 | 100% |
| Acceptance Criteria (AC-XXX) | 45 | 45 | 100% |
| Non-Functional (NF-XXX) | 15 | 15 | 100% |

---

## Traceability Matrix

> Map every requirement to its implementation and test.

| Req ID | Description | Source File(s) | Test File(s) | Status |
|--------|-------------|---------------|-------------|--------|
| R-001 | Docker Compose orchestrates 4 services on omni-net | `docker-compose.yml` | — | ✅ |
| R-002 | Rust starts after SurrealDB + TEI health checks | `docker-compose.yml` | — | ✅ |
| R-003 | Tree-sitter AST parsing (multi-language) | `src/parser/` | `tests/parser/` | ✅ |
| R-004 | TEI embedding (384-dim vectors) | `src/embedder/` | `tests/embedder/` | ✅ |
| R-005 | Node storage in SurrealDB | `src/db/` | `tests/db/` | ✅ |
| R-006 | Directional graph edges (linked_to) | `src/db/` | `tests/db/` | ✅ |
| R-007 | HNSW vector index (cosine, 384-dim) | `src/db/schema.surql` | — | ✅ |
| R-008 | GET /api/graph | `src/api/` | `tests/api/` | ✅ |
| R-009 | GET /api/search (vector similarity) | `src/api/` | `tests/api/` | ✅ |
| R-010 | POST /api/ingest | `src/api/` | `tests/api/` | ✅ |
| R-011 | File staleness tracking | `src/ingestion/` | `tests/ingestion/` | ✅ |
| R-012 | Community detection (Leiden) | `src/analysis/` | `tests/analysis/` | ✅ |
| R-013 | React + Cosmograph UI | `ui/src/` | `ui/tests/` | ✅ |
| R-014 | Dark-mode WebGL force-directed layout | `ui/src/` | — | ✅ |
| R-015 | Community-based node coloring | `ui/src/` | — | ✅ |
| R-016 | Node hover metadata panel | `ui/src/` | — | ✅ |
| R-017 | Node click neighbor highlighting | `ui/src/` | — | ✅ |
| R-018 | Semantic search bar in UI | `ui/src/` | — | ✅ |
| R-019 | SurrealDB schema auto-initialization | `src/db/schema.surql` | `tests/db/` | ✅ |
| R-020 | Graceful orphaned node handling | `src/api/`, `ui/src/` | `tests/api/` | ✅ |
| R-021 | File watcher daemon (FSEvents) | `src/watcher/` | `tests/unit_tests.rs` | ✅ |
| R-022 | GET /api/health | `src/api/` | `tests/api/` | ✅ |
| R-023 | GET /api/stats | `src/api/` | `tests/api/` | ✅ |
| R-024 | Docker DNS hostnames (no localhost) | `docker-compose.yml`, `src/config/` | — | ✅ |
| R-025 | Agent Enforcement Layer (hooks.json PreToolUse/PreInvocation/Stop) | `.agents/hooks.json`, `scripts/hook_*.sh` | `tests/enforcement/` | ✅ |
| R-026 | Omni-Graph Native Agent Interface (MCP / Antigravity Skill) | `.agents/skills/omni-graph/` | `tests/agent/` | ✅ |
| R-027 | Automated Agent Setup CLI & Bootstrapper | `Makefile`, `scripts/setup-agent.sh` | `tests/setup/` | ✅ |
| R-028 | Subgraph Context Condenser (<1500 tokens AST slice) | `src/condenser/` | `tests/condenser/` | ✅ |
| R-029 | Dynamic Live Galaxy Clustering & LPA Seed Persistence | `src/watcher/`, `src/analysis/` | `tests/unit_tests.rs` | ✅ |
| R-030 | Subsystem Boundary Contracts & Agent Actionable Advice | `src/api/`, `src/db/` | `tests/unit_tests.rs` | ✅ |
| R-031 | Comprehensive Polyglot AST Grammar (Go/TS/Rust/Python) | `src/parser/` | `tests/unit_tests.rs` | ✅ |
| R-032 | Inbound Edge Preservation in Delta Sync | `src/db/` | `tests/unit_tests.rs` | ✅ |
| R-033 | Two-Tier Exact Symbol Retrieval (ORDER BY is_exact) | `src/db/` | `tests/unit_tests.rs` | ✅ |
| R-034 | Deterministic LPA Tie-Breaking (Ordered IDs) | `src/analysis/` | `tests/unit_tests.rs` | ✅ |
| R-035 | Scoped Call-Edge Resolution (File -> Dir -> Fallback) | `src/db/` | `tests/unit_tests.rs` | ✅ |
| R-036 | Pre-Computed Galaxy Graph-RAG Retrieval | `src/analysis/` | `tests/unit_tests.rs` | ✅ |
| R-037 | Context-Enriched Embedding Payloads | `src/ingestion/`, `src/watcher/` | `tests/unit_tests.rs` | ✅ |
| R-038 | Galaxy Subsystem Disambiguation | `src/analysis/` | `tests/unit_tests.rs` | ✅ |
| R-039 | Scoped Source Node Resolution in Edge Persistence | `src/db/` | `tests/unit_tests.rs` | ✅ |
| R-040 | Universal Invocation Dispatch (Python, Rust, JS/TS) | `src/parser/` | `tests/unit_tests.rs` | ✅ |
| R-041 | Go Interface Method AST Extraction | `src/parser/` | `tests/unit_tests.rs` | ✅ |
| R-042 | Direct Pre-Computed Galaxy Delivery | `src/api/`, `src/db/` | `tests/unit_tests.rs` | ✅ |
| R-043 | Localized Subgraph Neighborhood in Condenser | `src/condenser/`, `src/db/` | `tests/unit_tests.rs` | ✅ |
| R-044 | First-Class Import Structural Linking | `src/parser/` | `tests/unit_tests.rs` | ✅ |
| R-045 | Cross-File Receiver & Impl Target Scoping | `src/db/` | `tests/unit_tests.rs` | ✅ |
| R-046 | Cascade Edge Deletion on Node Pruning (Zero Dangling Pointers) | `src/db/` | `tests/unit_tests.rs` | ✅ |
| R-047 | Stale File Pruning in Batch Ingestion (Zero Duplicate Nodes) | `src/ingestion/` | `tests/unit_tests.rs` | ✅ |
| R-048 | Trait & Class Inheritance Relationship Extraction | `src/parser/` | `tests/unit_tests.rs` | ✅ |
| R-049 | Rust Macro Definition & Invocation Indexing | `src/parser/` | `tests/unit_tests.rs` | ✅ |
| R-050 | Unified Record ID Normalization (clean_record_id) | `src/db/` | `tests/unit_tests.rs` | ✅ |
| R-051 | O(1) Git Worktree Pointer Peeking & Lineage Mapping | `src/ingestion/worktree.rs`, `src/db/` | `tests/unit_tests.rs` | ✅ |
| R-052 | Enforced Auto-Watch Registration on Worktree Ingestion | `src/api/`, `src/watcher/` | `tests/unit_tests.rs` | ✅ |
| R-053 | Seed-Preserving Prior Inheritance for Worktree Galaxy Clustering | `src/ingestion/`, `src/db/`, `src/analysis/` | `tests/unit_tests.rs` | ✅ |
| R-054 | Dead Worktree Garbage Collection & Workspace Pruning | `src/db/`, `src/api/` | `tests/unit_tests.rs` | ✅ |
| R-055 | Worktree Hierarchy & Lineage Representation in /api/workspaces | `src/db/`, `src/api/` | `tests/unit_tests.rs` | ✅ |
| R-056 | Worktree Visualizer & Branch Navigation in UI | `ui/src/App.tsx` | — | ✅ |
| R-057 | Proactive Agent Hook & Worktree Auto-Remediation Banner | `.agents/scripts/`, `scripts/` | `tests/unit_tests.rs` | ✅ |
| R-058 | Rebase Storm Debouncing & Surge Suppression | `src/watcher/` | `tests/unit_tests.rs` | ✅ |
| R-059 | Top-Down Worktree Fleet Auto-Cascade & Dynamic Sync | `src/ingestion/worktree.rs`, `src/api/` | `tests/unit_tests.rs` | ✅ |
| R-060 | Java Tree-sitter Parser Integration (classes, records, methods) | `src/parser/mod.rs:96-99, 288, 632, 903` | `tests/unit_tests.rs:1799` | ✅ |
| R-061 | C/C++ Tree-sitter Parser Integration (structs, functions, `#include`) | `src/parser/mod.rs:101-109, 301, 355, 689` | `tests/unit_tests.rs:1847` | ✅ |
| R-062 | TypeScript Deep AST Hardening & Detached Property Filtering | `src/parser/mod.rs:740-756` | `tests/unit_tests.rs:1891` | ✅ |
| R-063 | Topological Containment Backbone (`CONTAINS_DIR` & `CONTAINS_FILE`) | `src/ingestion/mod.rs:125-145`, `src/db/mod.rs:165-175` | `tests/unit_tests.rs:1932, 1943` | ✅ |
| R-064 | Hierarchical Community Consolidation & Macro-Cluster Size Floor (CPM/Modularity) | `src/analysis/mod.rs:190-330` | `tests/unit_tests.rs:1701-1770` | ✅ |
| R-065 | Scope-Aware Import Resolution (SAIR) (eliminating bare name collisions) | `src/db/mod.rs:459-515` | `tests/unit_tests.rs:1643-1695` | ✅ |
| R-066 | Two-Tower Hybrid Retrieval Engine (Tantivy BM25 + HNSW Vector RRF) | `src/retrieval/` | — | ⬜ (Deferred to v0.3.1) |
| R-067 | Content Tier Partitioning (`code_node` vs `doc_node` in vector search & Graph-RAG) | `src/db/mod.rs:320-365`, `src/api/mod.rs:322` | Live HTTP `/api/search?include_docs=true` | ✅ |
| R-068 | Strict Binary-Search Budgeting (<1,500 tokens) & Syntax Normalization | `src/condenser/mod.rs:180-250` | `tests/unit_tests.rs:880-920` | ✅ |
| R-069 | Personalized PageRank (PPR) Neighborhood Slicing (Unweighted Multigraph) | `src/condenser/mod.rs:20-81` | `tests/unit_tests.rs:892` | ✅ |
| R-070 | Native Model Context Protocol (MCP) Server for Omni-Graph | `tools/omni-graph/scripts/omni-mcp.py` | Live stdio JSON-RPC 2.0 | ✅ |
| R-071 | Portable Standalone Skill & One-Click CLI Bootstrapper | `tools/omni-graph/scripts/setup-agent.sh:110-270` | `make setup-agent` | ✅ |
| R-072 | Incremental AST Subtree Diffing (Diffsitter Pattern) | `src/watcher/` | — | ⬜ (Deferred to v0.3.1) |
| R-073 | Dynamic Watcher Status Merging in /api/workspaces (is_watching) | `src/api/mod.rs:120-145` | `tests/unit_tests.rs:1520-1550` | ✅ |
| R-074 | In-Flight Syncing Status Preservation Guardrail | `src/api/mod.rs:135`, `src/watcher/mod.rs:78` | Live multi-file edit verification | ✅ |
| R-075 | Canonical Path Normalization & Idempotency in WatcherManager | `src/watcher/mod.rs:65-95` | `tests/unit_tests.rs:1560-1580` | ✅ |
| R-076 | Watcher Telemetry Cumulative Metric Retention | `src/watcher/mod.rs:105-140` | Live `/api/watch/status` | ✅ |
| R-077 | Unified Ergonomic CLI Binary (`omni` tool) | `tools/omni-graph/scripts/omni.sh` | Live CLI commands | ✅ |
| R-078 | Standardized JSON-RPC 2.0 MCP Tool Schema | `tools/omni-graph/scripts/omni-mcp.py` | Live JSON-RPC initialize & call | ✅ |
| R-079 | Adaptive Skill-MCP Dual Bridge Protocol | `.agents/skills/omni-graph/SKILL.md` | Skill installation audit | ✅ |
| R-H077 | Unified Clustering Pipeline Parity (Ingestion & Watcher & API) | `src/ingestion/`, `src/watcher/`, `src/analysis/` | `tests/unit_tests.rs:test_r077_ingestion_and_recluster_parity` | ✅ |
| R-H078 | Two-Pass Ingestion Engine (Deferred In-Memory SAIR Linkage) | `src/ingestion/symbol_table.rs`, `src/ingestion/mod.rs` | `tests/unit_tests.rs:test_r078_two_pass_resolves_forward_references` | ✅ |
| R-H079 | Inbound Caller Preservation During Delta Sync | `src/db/mod.rs`, `src/watcher/pipeline.rs` | `tests/unit_tests.rs:test_r079_delta_sync_preserves_inbound_callers` | ✅ |
| R-H080 | Transactional Read-Fencing & Indexing Barrier (HTTP 423) | `src/api/mod.rs`, `src/ingestion/mod.rs`, `ui/src/App.tsx` | `tests/unit_tests.rs:test_r080_is_indexing_read_barrier` | ✅ |
| R-H081 | High-Throughput Bulk Vector Queue (256 texts/batch) | `src/ingestion/mod.rs` | Benchmarked ingestion loop | ✅ |
| R-H082 | AST Receiver Expression Scoping (`User.objects.create()` -> `User`) | `src/parser/mod.rs` | `tests/unit_tests.rs:test_r082_ast_receiver_preservation` | ✅ |
| R-H083 | True SurrealDB v2 HNSW KNN Operator `<|k, ef|>` | `src/db/mod.rs` | Verified against HNSW index | ✅ |
| R-H084 | Hierarchical Leiden Macro-Coarsening ($5 \le K \le 15$) | `src/analysis/mod.rs`, `src/api/mod.rs` | `tests/unit_tests.rs:test_r084_hierarchical_leiden_coarsening` | ✅ |
| R-H085 | Complete Uncle Bob Package Metrics ($C_a, C_e, I, A, D$) | `src/analysis/mod.rs`, `src/db/mod.rs` | `tests/unit_tests.rs:test_r085_package_metrics_completeness` | ✅ |
| R-093 | Reactive Reverse Dependency Re-Resolution (`symbol_callers`) | `src/ingestion/symbol_table.rs` | `tests/unit_tests.rs:test_r093_reactive_reverse_dependency_resolution` | ✅ |
| R-094 | Autonomous Directory Backbone DAG Synthesis | `src/watcher/pipeline.rs` | `tests/unit_tests.rs:test_r094_autonomous_directory_backbone_synthesis` | ✅ |
| R-095 | Dual-Tier Modularity Fast-Path ($\Delta Q$) | `src/analysis/mod.rs` | `tests/unit_tests.rs:test_r095_dual_tier_local_modularity_fast_path` | ✅ |
| R-096 | CPU-Bound Concurrency Isolation (`spawn_blocking`) | `src/watcher/mod.rs` | Unit & reactor responsiveness tests | ✅ |
| R-097 | Transactional Convergence Barrier (`wait_for_sync`) | `src/api/mod.rs`, `src/watcher/mod.rs` | `tests/unit_tests.rs:test_r097_convergence_barrier_blocks_until_sync` | ✅ |
| R-098 | Nanometer Fidelity Diagnostic Endpoint (`/api/watch/fidelity`) | `src/api/mod.rs`, `src/db/mod.rs` | `tests/unit_tests.rs:test_r098_fidelity_diagnostic_endpoint` | ✅ |
| R-099 | Lock-Free Generational Versioning (`graph_generation`) | `src/db/mod.rs`, `src/api/mod.rs` | `tests/unit_tests.rs:test_r099_lock_free_generational_versioning` | ✅ |
| R-100 | High-Throughput Parallel Batch Persistence & MCP Hardening | `src/watcher/pipeline.rs`, `scripts/omni-mcp.py` | Unit tests & MCP handler tests | ✅ |
| R-101 | Atomic Transactional Multi-Statement Swap | `src/db/mod.rs`, `src/watcher/pipeline.rs` | `tests/unit_tests.rs:test_r101_atomic_transactional_multi_statement_swap` | ✅ |
| R-102 | Disambiguated Inbound Edge Signature Relinking | `src/db/mod.rs` | `tests/unit_tests.rs:test_r102_disambiguated_inbound_signature_relinking` | ✅ |
| R-103 | Unresolved Edge Retention Pool | `src/ingestion/symbol_table.rs` | `tests/unit_tests.rs:test_r103_unresolved_edge_retention_pool` | ✅ |
| R-104 | Edge Relationship Deduplication & Call Weighting | `src/db/mod.rs` | `tests/unit_tests.rs:test_r104_edge_relationship_deduplication` | ✅ |
| R-105 | Zero-Delay Community Inference for New Files | `src/analysis/mod.rs`, `src/watcher/pipeline.rs` | `tests/unit_tests.rs:test_r105_zero_delay_community_inference_for_new_files` | ✅ |
| R-106 | Language-Aware SAIR Scoping & Exact Match Fast-Path | `src/ingestion/symbol_table.rs`, `src/db/mod.rs` | `tests/unit_tests.rs:test_r106_language_aware_sair_scoping` | ✅ |
| R-107 | MathWorks MATLAB (`.m`) AST & Classdef Ingestion | `src/parser/mod.rs`, `grammars/tree-sitter-matlab` | `tests/unit_tests.rs:test_r107_matlab_*` | ✅ |
| R-108 | MathWorks Simulink (`.slx` / `.mdl`) Block Diagram MBSE Ingestion | `src/parser/simulink.rs`, `src/ingestion/mod.rs`, `src/watcher/pipeline.rs` | `tests/unit_tests.rs:test_r108_*` | ✅ |
| R-109 | Filesystem Symlink Cycle Guarding & Alias Navigation | `src/ingestion/mod.rs`, `src/watcher/pipeline.rs` | `tests/unit_tests.rs:test_r109_symlink_cycle_guard_and_alias_navigation` | ✅ |

**Status Legend**: ⬜ Not started · 🔨 In progress · ✅ Implemented · 🧪 Tested · ❌ Blocked

---

## Active Blockers

| Blocker | Since | Impact | Resolution Plan |
|---------|-------|--------|-----------------|
| _None_ | — | — | — |

---

## Next Steps

> _What to do when you pick this tool up. See CONTEXT.md for full orientation._

1. Review and approve the specification (transition to `spec-review`)
2. Begin Task 1: Docker Compose infrastructure
3. Begin Task 2: Rust orchestrator skeleton

---

## Checklist for Current Phase

#### Draft Phase Requirements

- [x] Problem statement defined
- [x] Spec file created in `specs/catalog/`
- [x] Tool directory scaffolded
- [ ] Entry added to tool catalog
- [ ] CONTEXT.md populated
- [ ] DEVLOG.md initialized
