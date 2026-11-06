# ⚡ Omni-Graph Spec: Live Watcher Fidelity Parity & Reactive Concurrency Engine

> **Document ID**: `SPEC-OMNI-005`  
> **Status**: `implemented`  
> **Target Version**: `v0.7.0`  
> **Author**: Antigravity (Google DeepMind) & `@vadde`  
> **Date**: 2026-11-06  
> **Precedence**: Extends [`specs/catalog/omni-graph.md`](omni-graph.md), [`specs/catalog/omni-graph-two-pass-engine-hardening.md`](omni-graph-two-pass-engine-hardening.md), and [`specs/catalog/omni-graph-ai-native-galaxy-cosmology.md`](omni-graph-ai-native-galaxy-cosmology.md).  
> **Architecture Decision Record**: Governed by [`docs/architecture/adr-002-live-watcher-fidelity-concurrency-overhaul.md`](../../docs/architecture/adr-002-live-watcher-fidelity-concurrency-overhaul.md).

---

## 1. Executive Summary & Epistemic Motivation

### 1.1 The Core Imperative: Zero Stale or Corrupted Data for AI Coding Agents
Modern autonomous AI coding agents (such as Antigravity, Claude 3.7/3.8, Gemini 2.5/3.0 Pro, Cursor, Devin) execute complex multi-step reasoning, plan cross-module refactors, generate atomic diffs, and assess blast radius across distributed architectures. 

When an AI agent modifies files in a watched repository:
- **Serving Stale, Fragmented, or Corrupted Graph Data Compounds Failures**: If an agent queries `references?symbol=X` or `condense?symbol=X` and receives an edge set where reverse callers are broken, duplicate edges inflate caller counts, or newly exported symbols are unlinked, the agent's internal Chain of Thought is poisoned.
- **Correctness Outweighs Milliseconds**: As established in the Antigravity Constitution (Pillar V: Quality Over Time), **accuracy and topological ground truth are non-negotiable**. Waiting 200ms–2,000ms for deterministic graph convergence is exponentially superior to returning a 50ms answer derived from fragmented, stale, or duplicated graph state.

### 1.2 Architectural Decision Record (ADR-002) Traceability
To prevent theoretical contradictions and ensure future engineering continuity, every requirement in this specification is formally grounded in an accepted Architectural Decision in **ADR-002**:
- **ADR-002.1 (Transactional Atomic Batch Swap)** $\to$ Governs **R-101**, **R-102**, and **R-104**.
- **ADR-002.2 (Reactive Inverted Dependency Graph & Unresolved Retention)** $\to$ Governs **R-093**, **R-103**, and **R-106**.
- **ADR-002.3 (Zero-Overhead Concurrency & Coarsening Isolation)** $\to$ Governs **R-095**, **R-096**, and **R-100**.
- **ADR-002.4 (Agent-Centric Convergence Barrier & Self-Healing MCP)** $\to$ Governs **R-094**, **R-097**, **R-098**, **R-099**, and **R-105**.

---

## 2. The 14 Verified Forensic Gaps in v0.6.0 Live Watcher
A recursive divide-and-conquer code and database audit of `tools/omni-graph/src/` uncovered **14 verified structural vulnerabilities and race conditions** between cold-start ingestion and the live watcher:

1. **Gap 1: Multi-Roundtrip Torn-Read Blackout Window** ([`pipeline.rs:165-221`](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/tools/omni-graph/src/watcher/pipeline.rs#L165-L221)): Single-file sync executes 4 separate HTTP POST calls (`delete`, `store_nodes`, `store_edges`, `relink`). During the 50ms–250ms gap between delete and relink, the file's nodes and callers literally do not exist in SurrealDB.
2. **Gap 2: Read-Fencing Watcher Blind Spot** ([`api/mod.rs:329-355`](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/tools/omni-graph/src/api/mod.rs#L329-L355)): `check_read_fence` only checks cold-start batch ingestion, completely ignoring the live watcher while it is actively syncing.
3. **Gap 3: Asymmetric Temporal Edge Dropping** ([`symbol_table.rs:165-181`](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/tools/omni-graph/src/ingestion/symbol_table.rs#L165-L181)): Unresolved forward references are discarded upon failure. If an agent writes a caller before a callee, the call edge is dropped forever unless the caller file is re-touched.
4. **Gap 4: Asymmetric Foreign Reverse-Caller Staleness** ([`pipeline.rs:175-205`](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/tools/omni-graph/src/watcher/pipeline.rs#L175-L205)): The watcher resolves outbound edges from the modified file, but never re-evaluates foreign files that call into this file.
5. **Gap 5: Label-Only Inbound Relinking Collision** ([`db/mod.rs:1399-1412`](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/tools/omni-graph/src/db/mod.rs#L1399-L1412)): Inbound relinking indexes nodes purely by bare `label`, corrupting callers on sibling structs/classes (`User::new` vs `Session::new`).
6. **Gap 6: Directory Backbone Incompleteness** ([`pipeline.rs:126-142`](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/tools/omni-graph/src/watcher/pipeline.rs#L126-L142)): Creating a file in a new directory emits a `CONTAINS_FILE` edge from `{ws}:dir:parent`, but the directory node does not exist in SurrealDB.
7. **Gap 7: Watcher Reactor Starvation & Notify Buffer Overflow** ([`watcher/mod.rs:647-686`](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/tools/omni-graph/src/watcher/mod.rs#L647-L686)): Heavy graph fetch and 15-iteration Louvain run directly on the watcher loop thread, blocking event reception and overflowing the 512-event buffer during burst edits.
8. **Gap 8: Un-Deduplicated Multi-Edge Inflation** ([`db/mod.rs:1220`](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/tools/omni-graph/src/db/mod.rs#L1220)): Calling a function multiple times emits duplicate edge records in SurrealDB (empirically verified: 499 duplicate groups in PolyTuring!). Floods `/api/references` with repetitive caller rows and starves other distinct callers under `LIMIT`.
9. **Gap 9: Dead-Code Community Inference & New File Galaxy Orphanage** ([`analysis/mod.rs:685`](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/tools/omni-graph/src/analysis/mod.rs#L685), [`pipeline.rs:145`](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/tools/omni-graph/src/watcher/pipeline.rs#L145)): `infer_neighbor_community` was abandoned as dead code; new files are inserted with `community = NULL` until the 3.5s quiescent sweep.
10. **Gap 10: Sequential TEI Vector Batching** ([`embedder/mod.rs:56`](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/tools/omni-graph/src/embedder/mod.rs#L56)): TEI 32-text chunks execute sequentially over HTTP instead of concurrently via `tokio::task::JoinSet`.
11. **Gap 11: Language-Blind Candidate Selection in SAIR Tier 2 & Tier 4** ([`symbol_table.rs:142`](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/tools/omni-graph/src/ingestion/symbol_table.rs#L142)): Tier 2 blindly returns `candidates[0]` without checking language or import path, allowing cross-language symbol leakage in polyglot monorepos.
12. **Gap 12: $O(K \cdot |E|)$ Coarsening Loop Inefficiency** ([`analysis/mod.rs:417-430`](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/tools/omni-graph/src/analysis/mod.rs#L417-L430)): `coarsen_to_macro_galaxies` re-scans ALL graph links from scratch on every single merge iteration instead of maintaining an incremental affinity matrix.
13. **Gap 13: Full-Table Scan on Reference Lookups** ([`db/mod.rs:1220`](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/tools/omni-graph/src/db/mod.rs#L1220)): `find_references` searches `string::ends_with(label, ...)` on unindexed patterns without exact match fast-path, forcing full table scans on large codebases.
14. **Gap 14: Raw 423 Unhandled Exception in MCP Agent Bridge** ([`omni-mcp.py:323-334`](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/tools/omni-graph/scripts/omni-mcp.py#L323-L334)): `omni-mcp.py` catches HTTP 423 as a generic error rather than passing `wait_for_sync=true` or providing transparent retry handling.

---

## 2. Requirements & Traceability Matrix

### 2.1 Functional Requirements

| ID | Title | Description | Target Component |
| :--- | :--- | :--- | :--- |
| **R-093** | **Reactive Reverse Dependency Re-Resolution** | In-memory `ReverseDependencyIndex` (`symbol_callers`). When a file's exported symbols change, proactively re-resolve foreign caller edges in SurrealDB without touching foreign files on disk. | `symbol_table.rs`, `pipeline.rs` |
| **R-094** | **Autonomous Directory Backbone Synthesis** | `DirectoryTreeTracker` in pipeline automatically synthesizes ancestor `directory` nodes and hierarchical `CONTAINS_DIR` edges for any newly created directory paths. | `pipeline.rs`, `db/mod.rs` |
| **R-095** | **Dual-Tier Live Modularity Rebalancing** | Local fast-path modularity check in `reindex_file` using closed-form $\Delta Q$; dedicated background clustering worker with preemption via `tokio_util::sync::CancellationToken`. | `pipeline.rs`, `watcher/mod.rs` |
| **R-096** | **CPU-Bound Concurrency Isolation (`spawn_blocking`)** | Offload tree-sitter AST parsing, Leiden coarsening, and Uncle Bob metric calculations to dedicated blocking threads (`tokio::task::spawn_blocking`), preventing reactor starvation. | `watcher/`, `analysis/` |
| **R-097** | **Transactional Convergence Barrier (`wait_for_sync`)** | Add optional `wait_for_sync=true` query parameter to agent endpoints (`/api/symbol`, `/api/references`, `/api/condense`, `/api/query`, `/api/galaxies`). Awaits internal `tokio::sync::watch` convergence channel (up to 3,000ms timeout). | `api/mod.rs`, `watcher/mod.rs` |
| **R-098** | **Nanometer Fidelity Diagnostic Endpoint (`/api/watch/fidelity`)** | Expose a real-time fidelity diagnostic endpoint reporting `fidelity_score`, unresolved edges, dangling directory pointers, reverse invalidation counts, and convergence state. | `api/mod.rs`, `db/mod.rs` |
| **R-099** | **Lock-Free Generational Versioning (`graph_generation`)** | Increment an atomic 64-bit generation counter (`Arc<AtomicU64>`) on every graph mutation. Endpoints return `X-Graph-Generation` header. | `db/mod.rs`, `api/mod.rs` |
| **R-100** | **High-Throughput Parallel Batch Persistence (`JoinSet`)** | Parallelize independent SurrealDB and TEI writes using `tokio::task::JoinSet`, reducing single-file sync latency. | `pipeline.rs`, `embedder/mod.rs` |
| **R-101** | **Atomic Transactional Multi-Statement Swap** | Unify node deletion, node insertion, edge resolution, and inbound relinking into a single atomic `BEGIN TRANSACTION; ... COMMIT TRANSACTION;` SurrealQL payload, eliminating the 50ms–250ms torn-read blackout window. | `pipeline.rs`, `db/mod.rs` |
| **R-102** | **Disambiguated Inbound Edge Signature Relinking** | Relink inbound caller edges using composite signatures `(label, kind, parent_type)` instead of raw `label` alone, eliminating cross-type method collisions. | `db/mod.rs`, `pipeline.rs` |
| **R-103** | **Unresolved Edge Retention Pool** | Retain unresolved cross-file edges in an in-memory `UnresolvedReferencePool`. When a new symbol declaration is indexed, immediately satisfy and commit previously dangling callers without waiting for caller files to be edited. | `symbol_table.rs`, `pipeline.rs` |
| **R-104** | **Edge Relationship Deduplication & Call Weighting** | Store single unique directed edge between any source and target node in `linked_to` with an explicit `call_count: int` and `weight: float` field; deduplicate `/api/references` to return distinct caller entities. | `db/mod.rs`, `api/mod.rs` |
| **R-105** | **Zero-Delay Community Inference for New Files** | Activate `infer_neighbor_community` in `reindex_file` so newly created files immediately inherit the modal community of their connected neighbors upon creation instead of being inserted with `community = NULL`. | `analysis/mod.rs`, `pipeline.rs` |
| **R-106** | **Language-Aware SAIR Scoping & Exact Match Fast-Path** | Enhance `SymbolDef` with `language` field; enforce language compatibility in SAIR Tier 2 & Tier 4; add exact-match pre-query in `find_references` to utilize `idx_node_label` before substring scans. | `symbol_table.rs`, `db/mod.rs` |
| **R-107** | **MathWorks MATLAB (`.m`) AST & Classdef Extraction (Deferred)** | *Tail-End Research Phase*: Integrate Tree-sitter MATLAB grammar to parse `.m` functions, scripts, `classdef` OOP blocks, packages (`+pkg`), `methods`, and `properties` into first-class graph nodes. Requires deep research into proprietary MathWorks syntax and cross-file function resolution. | `src/parser/`, `Cargo.toml` |
| **R-108** | **MathWorks Simulink (`.slx` / `.mdl`) Block Diagram Ingestion (Deferred)** | *Tail-End Research Phase*: Model-Based Systems Engineering (MBSE) ingestion. Parse `.slx` Open Packaging XML (`simulink/blockdiagram.xml`) and `.mdl` hierarchical trees into graph nodes (`Block`, `Subsystem`) and edges (`SIGNAL_TO`, `REFERENCES_MODEL`). Demands dedicated investigation into proprietary MathWorks binary/archive formats. | `src/parser/simulink.rs` |
| **R-109** | **Filesystem Symlink Cycle-Guarded Graph Traversal** | Ensure the watcher and ingestion engine safely follow or resolve filesystem symlinks (`std::fs::symlink_metadata`) with cycle detection, preventing infinite loops while indexing symlinked submodules and shared packages. | `src/ingestion/`, `src/watcher/` |

### 2.2 Acceptance Criteria (AC)

- **AC-037**: Adding a new function `baz()` in `File A` immediately resolves a previously dangling `CALLS baz` reference in `File B` within 300ms, verified in SurrealDB without touching `File B`.
- **AC-038**: Renaming `foo()` to `bar()` in `File A` removes obsolete `CALLS foo` edges from foreign callers and creates valid `CALLS bar` edges if call sites were updated.
- **AC-039**: Creating a file in a new deep directory (e.g. `src/new_mod/sub/file.rs`) synthesizes `directory` nodes for `src/new_mod` and `src/new_mod/sub` with proper `CONTAINS_DIR` edges; zero dangling directory pointers in SurrealDB.
- **AC-040**: Shifting a file's primary connections triggers immediate local fast-path community reassignment without waiting for the 2.0s quiescent sweep.
- **AC-041**: `/api/watch/fidelity?workspace=X` returns `fidelity_score: 1.0` and `dangling_directory_edges: 0`.
- **AC-042**: Querying `/api/references?symbol=X&wait_for_sync=true` while saving a file blocks until sync completes and returns 100% accurate results.
- **AC-043**: Concurrent reads during file reindexing never experience torn reads or 0-reference drops; verified by transactional swap.
- **AC-044**: Files containing multiple sibling methods with identical names (e.g. `User::new` and `Session::new`) relink to their exact respective targets without cross-type contamination.
- **AC-045**: Callers calling a target multiple times (e.g. `make_http_request`) return exactly 1 reference entry per distinct caller entity in `/api/references`, with accurate `call_count`.
- **AC-046**: Newly created files are immediately assigned a non-null community upon insertion based on modal neighbor affinity.
- **AC-047**: SAIR resolution in polyglot repos strictly matches same-language candidates before cross-language fallbacks.
- **AC-048**: `find_references` utilizes `idx_node_label` index, completing exact-match queries in $\le 5\text{ms}$.
- **AC-049**: 100% of existing unit tests pass, and new tests validate all 14 requirements.
- **AC-050**: Zero clippy warnings, living docs updated in lockstep.
- **AC-051**: Traversing cyclic filesystem symlinks (e.g. `a/loop -> ..` or mutual circular links) terminates deterministically in $O(1)$ lookup per directory without recursion loop, process hang, or stack overflow.
- **AC-052**: A symlinked file within the workspace is indexed as a dedicated `symlink` node with a `LINKS_TO` edge pointing directly to the canonical target file, producing zero duplicate AST function/symbol declarations and zero duplicate community assignments.
- **AC-053**: The incremental watcher (`IncrementalPipeline`) safely resolves file or directory symlink modifications, deletions, or creations without panic.

### 2.3 Non-Functional Requirements (NF)

- **NF-019 (Sync Latency Bound)**: Single-file delta sync with reverse invalidation completes in $\le 350\text{ms}$ for repos up to 50,000 nodes.
- **NF-020 (Reactor Safety)**: Zero CPU-bound routines $> 5\text{ms}$ execute directly on Tokio worker threads.
- **NF-021 (Memory Efficiency)**: In-memory `ReverseDependencyIndex` and `UnresolvedReferencePool` consume $\le 20\text{MB}$ RAM for a 100,000-edge graph.
- **NF-022 (Graceful Degradation)**: If `wait_for_sync=true` times out (default 3,000ms), the endpoint gracefully falls back to returning the current graph snapshot with header `X-Graph-Converged: false`.

---

## 3. Mathematical & Algorithmic Foundation

### 3.1 Closed-Form Dynamic Modularity (Local Moving)
Instead of waiting 3.5 seconds for a global 15-iteration Louvain recomputation, v0.7.0 implements **instant local modularity rebalancing** in RAM during `reindex_file`:

$$\Delta Q(u: C_A \to C_B) = \left[\frac{k_{u, C_B}}{2m} - \frac{\Sigma_{\text{tot}}(C_B) \cdot k_u}{2m^2}\right] - \left[\frac{k_{u, C_A}}{2m} - \frac{\Sigma_{\text{tot}}(C_A) \cdot k_u}{2m^2}\right]$$

* **Execution Time**: $O(\text{deg}(u))$ — sub-microsecond in RAM ($< 0.05\text{ms}$).
* **Invariant**: If $\Delta Q > 0$, the file's community is updated **instantly** during the atomic file swap. Community assignment is never stale.

### 3.2 Unresolved Edge Retention Protocol
```rust
pub struct UnresolvedReferencePool {
    // (workspace, target_symbol_label) -> Vec<ExtractedEdge>
    pool: HashMap<(String, String), Vec<ExtractedEdge>>,
}

impl UnresolvedReferencePool {
    /// When a file is indexed and has unresolved cross-file calls:
    pub fn retain_unresolved(&mut self, edge: ExtractedEdge) {
        self.pool.entry((edge.workspace.clone(), edge.target_label.clone())).or_default().push(edge);
    }

    /// When a new symbol declaration is registered in any file:
    pub fn drain_satisfied(&mut self, workspace: &str, symbol_label: &str) -> Vec<ExtractedEdge> {
        self.pool.remove(&(workspace.to_string(), symbol_label.to_string())).unwrap_or_default()
    }
}
```

---

## 4. Verification & Testing Strategy

1. **Unit Test Suite**:
   - `test_r093_reactive_reverse_dependency_resolution`
   - `test_r094_autonomous_directory_backbone_synthesis`
   - `test_r095_dual_tier_local_modularity_fast_path`
   - `test_r097_convergence_barrier_blocks_until_sync`
   - `test_r098_fidelity_diagnostic_endpoint`
   - `test_r101_atomic_transactional_multi_statement_swap`
   - `test_r102_disambiguated_inbound_signature_relinking`
   - `test_r103_unresolved_edge_retention_pool`
   - `test_r104_edge_relationship_deduplication`
   - `test_r105_zero_delay_community_inference_for_new_files`
   - `test_r106_language_aware_sair_scoping`
2. **Empirical Verification on PolyTuring**:
   - Verify that 499 duplicate edge groups are cleanly deduplicated with `call_count`.
   - Verify that `/api/references` returns distinct caller rows with accurate call counts.
   - Verify that `/api/watch/fidelity` returns `fidelity_score: 1.0`.
   - Verify all 6 GitHub Actions workflows pass.

---

## 5. Implementation Roadmap & Execution Milestones

Execution proceeds across 8 deterministic milestones. At no point may code be implemented without test validation:

### Milestone 1: SymbolTable In-Memory Engine Hardening (`R-093`, `R-102`, `R-103`, `R-106`)
* Add `language: String` to `SymbolDef`.
* Enforce same-language preference in SAIR Tier 2 and Tier 4.
* Add `symbol_callers: HashMap<(String, String), HashSet<String>>` to track which files depend on a symbol.
* Add `UnresolvedReferencePool` to retain cross-file edges that cannot be resolved immediately, auto-draining them when target symbols appear.
* Update `unit_tests.rs` with `test_r093_reactive_reverse_dependency_resolution` and `test_r103_unresolved_edge_retention_pool`.

### Milestone 2: Autonomous Directory Backbone DAG Synthesis (`R-094`)
* Add `DirectoryTreeTracker` in `IncrementalPipeline` (`HashSet<String> known_dirs`).
* In `reindex_file`, inspect all directory ancestors. For any missing directory, synthesize `directory` node and `CONTAINS_DIR` edge up to root workspace.
* Update `unit_tests.rs` with `test_r094_autonomous_directory_backbone_synthesis`.

### Milestone 3: Atomic Transactional Multi-Statement Swap (`R-101`)
* Add `execute_atomic_swap` method in `DbClient` generating a single `BEGIN TRANSACTION; ... COMMIT TRANSACTION;` SurrealQL script combining deletion, node insertion, edge resolution, and inbound relinking.
* Eliminate the 4 sequential network round-trips in `reindex_file`.
* Update `unit_tests.rs` with `test_r101_atomic_transactional_multi_statement_swap`.

### Milestone 4: Edge Relationship Deduplication & Call Weighting (`R-104`)
* Deduplicate extracted edges in memory before database write, computing `call_count` and cumulative `weight`.
* Update `linked_to` schema with `call_count: int` and `weight: float`.
* Deduplicate callers in `/api/references` by `caller_id`.
* Update `unit_tests.rs` with `test_r104_edge_relationship_deduplication`.

### Milestone 5: Dual-Tier Modularity & Non-Blocking Louvain Actor (`R-095`, `R-096`, `R-105`)
* Implement local $\Delta Q$ modularity check in `reindex_file`.
* Activate `infer_neighbor_community` for new files upon creation (R-105).
* Decouple background clustering from the watcher event loop into an independent task with `CancellationToken` and `spawn_blocking`.
* Optimize `coarsen_to_macro_galaxies` to maintain an incremental affinity matrix ($O(|E|)$).
* Update `unit_tests.rs` with `test_r095_dual_tier_local_modularity_fast_path` and `test_r105_zero_delay_community_inference_for_new_files`.

### Milestone 6: Transactional Convergence Barrier (`wait_for_sync`) & Diagnostics (`R-097`, `R-098`, `R-099`, `R-100`)
* Add `wait_for_sync=true` query parameter support in `api/mod.rs` using a `tokio::sync::watch` convergence channel.
* Implement `GET /api/watch/fidelity` returning fidelity score, unresolved edges, and dangling pointer counts.
* Add `X-Graph-Generation` header via `Arc<AtomicU64>`.
* Update `omni-mcp.py` to pass `wait_for_sync=true` on all tool calls.
* Update `unit_tests.rs` with `test_r097_convergence_barrier_blocks_until_sync` and `test_r098_fidelity_diagnostic_endpoint`.

### Milestone 7: Full Test Suite Verification, Container Rebuild & Live Validation
* Verify 100% passing suite across all 100+ unit and integration tests.
* Verify 0 clippy warnings (`cargo clippy --all-targets -- -D warnings`).
* Rebuild Docker container `omni-rust-app` and test live on `PolyTuring`.
* Verify that 499 duplicate edge groups are resolved and `/api/watch/fidelity` returns `1.0`.

### Milestone 8 (Tail-End Research Spike): MathWorks MATLAB (`.m`) & Simulink (`.slx`) Ingestion (`R-107`, `R-108`, `R-109`)
* Research `tree-sitter-matlab` bindings and Simulink XML archive extraction.
* Author technical feasibility report for Model-Based Systems Engineering (MBSE) code graph ingestion.
