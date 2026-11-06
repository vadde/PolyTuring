# ADR-002: Omni-Graph v0.7.0 Live Watcher Fidelity Parity & AI-Native Reactive Concurrency Architecture

## Status

Accepted (Feature Branch: `feat/omni-graph-live-watcher-fidelity-parity`)

## Date

2026-11-06

## Context

Omni-Graph was engineered as an autonomous code graph knowledge engine providing multi-language AST extraction, vector similarity search, hierarchical community clustering, and sub-graph neighborhood slicing (<1500 tokens).

While cold-start batch ingestion (`execute_ingest`) achieved 100% relational fidelity through a two-pass architecture, a deep recursive forensic audit revealed that the **live file watcher delta engine** (`IncrementalPipeline::reindex_file` and `watcher/mod.rs`) suffered from 14 critical structural vulnerabilities, race conditions, and algorithmic bottlenecks:

1. **Non-Atomic Torn Reads**: File re-indexing executed 4 separate HTTP requests to SurrealDB (`delete`, `store_nodes`, `store_edges`, `relink`), creating a 50–250ms blackout window where nodes and inbound callers were missing from the database.
2. **Read-Fencing Blind Spot**: The API's read-fence (`check_read_fence`) only checked cold-start batch ingestion, allowing agent queries to hit SurrealDB directly in the middle of the mutation window.
3. **Asymmetric Temporal Edge Dropping**: When a caller was written before a callee (common in real-world incremental coding), unresolvable forward references were discarded into `/dev/null`.
4. **Foreign Reverse-Caller Staleness**: Export changes in a modified file never triggered re-evaluation of foreign callers in SurrealDB.
5. **Label-Only Relinking Collision**: Inbound caller relinking matched on raw `label`, causing cross-type collisions on sibling structs (e.g. `User::new` vs `Session::new`).
6. **Directory Backbone Incompleteness**: Files created in new directories lacked synthesized `directory` nodes in SurrealDB, leaving dangling pointers.
7. **Watcher Reactor Starvation**: Quiescent clustering executed heavy graph fetches and 15-iteration Louvain loops directly on the watcher async loop, risking inotify channel overflow during burst edits.
8. **Multi-Edge Inflation**: Multiple function calls between the same two nodes created duplicate edge records in SurrealDB (499 duplicate groups verified live in PolyTuring), bloating `/api/references` and starving distinct callers under `LIMIT`.
9. **Dead-Code Community Inference**: Newly created files were blindly inserted with `community = NULL` until the 3.5s quiescent sweep.
10. **Sequential Embedding Batching**: TEI 32-text chunks executed sequentially over HTTP.
11. **Language-Blind SAIR Scoping**: SAIR Tier 2 returned `candidates[0]` without checking programming language, allowing cross-language symbol leakage.
12. **$O(K \cdot |E|)$ Coarsening Loop**: Macro-galaxy coarsening re-scanned all graph edges on every single merge iteration inside the `while` loop.
13. **Full Table Scans on Reference Lookups**: Unindexed `string::ends_with` forced table scans on symbol queries.
14. **Raw 423 Unhandled Exception in MCP Bridge**: `omni-mcp.py` treated HTTP 423 as a crash rather than providing convergence wait.

---

## Architectural Decisions & Options Considered

### Decision 1: Atomic Transactional Multi-Statement Batch Swap (R-101)
* **Problem**: 4 sequential HTTP calls created a 50–250ms blackout window in SurrealDB.
* **Option A (SurrealDB Mutex Lock)**: Lock database writes in memory using an async mutex.  
  *Cons*: Only guards requests within the same process; does not prevent torn reads from external queries or readers.
* **Option B (SurrealQL Multi-Statement Transaction - ACCEPTED)**: Wrap deletion, node insertion, edge resolution, and inbound relinking into a single `BEGIN TRANSACTION; ... COMMIT TRANSACTION;` SurrealQL script.
  *Pros*: 100% database-enforced atomicity, zero torn reads, single network round-trip (latency reduced from ~180ms to ~35ms), automatic rollback on failure.

### Decision 2: Inverted Dependency Graph & Unresolved Edge Retention (R-093, R-103)
* **Problem**: Callers written before callees had their edges dropped; modifying exports left foreign callers stale.
* **Option A (Full Cold Re-Ingestion on Delta - Rejected)**: Rerun two-pass ingestion across the whole workspace on every file save.  
  *Cons*: Completely unscalable on codebases $> 100,000$ lines.
* **Option B (In-Memory Reactive Index & Unresolved Pool - ACCEPTED)**:
  - Maintain `symbol_callers: HashMap<(Workspace, SymbolLabel), HashSet<FilePath>>`.
  - Maintain `unresolved_pool: HashMap<(Workspace, SymbolLabel), Vec<ExtractedEdge>>`.
  - When a file adds an export: immediately drain matching edges from `unresolved_pool` and commit to SurrealDB.
  - When a file removes/renames an export: immediately invalidate foreign caller edges in SurrealDB and return them to the unresolved pool.
  *Pros*: True $O(k)$ incremental maintenance where $k$ is the number of dependent files; zero dropped edges regardless of edit ordering.

### Decision 3: Language-Aware Scope-Aware Import Resolution (SAIR) (R-106)
* **Problem**: In polyglot codebases, SAIR Tier 2 returned `candidates[0]`, cross-linking different languages (e.g. Rust to Python).
* **Option A (Monoglot Workspace Partitioning)**: Forbid multiple languages in the same workspace.  
  *Cons*: Monorepos are intrinsically polyglot.
* **Option B (Language-Weighted Composite Scoring - ACCEPTED)**:
  - Add `language` field to `SymbolDef`.
  - Enforce language compatibility in SAIR: candidates sharing the source file's language extension receive highest priority before cross-language fallbacks.

### Decision 4: Edge Deduplication & Call Weighting (R-104)
* **Problem**: Calling a target function $N$ times created $N$ duplicate edges in SurrealDB and returned repetitive rows in `/api/references`.
* **Option A (Client-Side Filtering Only)**: Filter duplicates only in REST response formatting.  
  *Cons*: Leaves database bloated with duplicate edge records; distorts graph algorithms (PPR, Louvain modularity).
* **Option B (Database Unique Edge with `call_count` - ACCEPTED)**:
  - Deduplicate edges in memory before SurrealDB insertion.
  - Store a single directed edge with `call_count: int` and `weight: float`.
  - `/api/references` returns distinct caller entities with call count metadata.

### Decision 5: Non-Blocking Quiescent Clustering & Concurrency Isolation (R-095, R-096)
* **Problem**: Running 15-iteration Louvain inside the watcher event loop caused reactor starvation and notify buffer overflow.
* **Option A (Synchronous Loop Yields)**: Insert `tokio::task::yield_now().await` inside Louvain loops.  
  *Cons*: Ineffective on heavy CPU loops; still blocks event processing.
* **Option B (Decoupled Background Actor with Preemption - ACCEPTED)**:
  - Move clustering into a dedicated background worker task governed by `tokio_util::sync::CancellationToken`.
  - Heavy graph algorithms execute inside `tokio::task::spawn_blocking`.
  - If new filesystem events arrive while clustering is running, the cancellation token immediately preempts the background task and resets the debounce.
  - In `coarsen_to_macro_galaxies`, maintain an incremental affinity matrix to reduce complexity from $O(K \cdot |E|)$ to $O(|E|)$.

### Decision 6: Transactional Convergence Barrier (`wait_for_sync`) (R-097)
* **Problem**: Agents querying the API mid-sync received torn or stale data; MCP client crashed on HTTP 423 Locked.
* **Option A (Polling with HTTP 423 Retry-After)**: Return HTTP 423 and force agents to implement backoff loops.  
  *Cons*: Flaky across different agent harnesses; wastes context window.
* **Option B (Server-Side Convergence Barrier via `tokio::sync::watch` - ACCEPTED)**:
  - Read endpoints support `wait_for_sync=true`.
  - If watcher is syncing, endpoint awaits internal convergence signal (up to 3,000ms timeout) and returns guaranteed fresh ground truth.
  - `omni-mcp.py` passes `wait_for_sync=true` by default on all tool calls.

---

## Consequences

### Positive
- **100% Relational Parity**: The live watcher maintains identical graph topology to a cold-start two-pass ingestion.
- **Zero Blackout Window**: Atomic SurrealQL multi-query transactions eliminate torn reads completely.
- **Zero Dropped Call Edges**: Callers written before callees are preserved and resolved dynamically.
- **Clean Token Economy**: Edge deduplication eliminates 499+ duplicate edge groups in PolyTuring, delivering compact, distinct caller lists to AI agents.
- **Sub-Millisecond Read Latency**: Exact-match index fast-paths eliminate full-table scans.
- **Zero Reactor Starvation**: Watcher event loop remains 100% responsive during background clustering.

### Negative
- Slight memory increase ($\le 20\text{MB}$) in `omni-rust-app` to maintain the in-memory `ReverseDependencyIndex` and `UnresolvedReferencePool`.
- Additional complexity in `pipeline.rs` to construct the unified multi-statement transactional payload.

### Neutral
- No breaking changes to existing REST endpoints: `wait_for_sync` defaults to `false`.
- WebGL Cosmograph UI continues to function with dynamic SSE updates.

---

## References
- [`specs/catalog/omni-graph-live-watcher-fidelity-parity.md`](../specs/catalog/omni-graph-live-watcher-fidelity-parity.md)
- [`specs/catalog/omni-graph-two-pass-engine-hardening.md`](../specs/catalog/omni-graph-two-pass-engine-hardening.md)
- [`specs/catalog/omni-graph-ai-native-galaxy-cosmology.md`](../specs/catalog/omni-graph-ai-native-galaxy-cosmology.md)
- Newman, M. E. J. (2004). *Analysis of weighted networks*. Physical Review E.
- Blondel, V. D., et al. (2008). *Fast unfolding of communities in large networks (Louvain algorithm)*.
- Traag, V. A., et al. (2019). *From Louvain to Leiden: guaranteeing well-connected communities*. Nature Scientific Reports.
