# Enhancement Spec: Omni-Graph Two-Pass Engine, Transactional Read-Fencing, Inbound Edge Preservation & High-Throughput Hardening

> **Status**: Approved  
> **Author**: Antigravity (Google DeepMind) & @vadde  
> **Created**: 2026-11-05  
> **Target Tool**: `tools/omni-graph/`  
> **Parent Spec**: [`specs/catalog/omni-graph.md`](omni-graph.md) & [`specs/catalog/omni-graph-nextgen.md`](omni-graph-nextgen.md)  
> **Version Target**: `v0.5.0`

---

## 1. Executive Summary

A forensic audit of Omni-Graph during real-world repository ingestion (`django-sorcery`, 252 files, 2,349 nodes) exposed multiple fundamental defects:
1. **Pipeline Desynchronization**: Ingestion auto-clustering omitted micro-cluster consolidation, creating 429 fragmented clusters, whereas manual UI re-clustering collapsed them to 141 clusters.
2. **Forward-Reference Edge Loss**: A single-pass directory walk executed immediate database lookups for call/import targets; targets in files visited later in walk order were absent from SurrealDB, silently dropping >84% of cross-file AST edges (e.g., 5,623 Python calls produced only 878 stored edges).
3. **Delta-Sync Inbound Edge Annihilation**: Live file re-indexing executed `DELETE linked_to WHERE in IN $nodes OR out IN $nodes`, destroying all incoming references from unaffected files across the codebase whenever a file was touched.
4. **Ingestion Latency & Read Contamination**: Ingestion took 1,456.7 seconds (~24.3 minutes) due to serial per-file HTTP embedding and unbatched dynamic SQL, during which un-isolated partial database states were exposed to queries, causing false galaxy counts (71 vs 141).
5. **Receiver Truncation in AST Parser**: In `parser/mod.rs`, `User.objects.create()` was stripped down to `"create"`, blacklisted as a generic verb, dropping references to the `User` class entirely.
6. **Bypassed HNSW Vector Indexing**: Similarity search queries ran full-table scans with `ORDER BY vector::similarity::cosine(...)` instead of SurrealDB v2's native KNN `<|k|>` operator.
7. **Ghost Metrics & Truncation Mirage**: Abstractness ($A$) and Distance from Main Sequence ($D$) were never computed, and the "15 galaxies" claim was an artifact of API truncation (`limit=15`) masking 141–521 true clusters.

This specification establishes the **v0.5.0 Production-Grade Engine**, delivering deterministic two-pass graph construction, retentive delta-sync, true SAIR Tier 2 resolution, sub-60-second throughput, and recursive Leiden macro-subsystems.

---

## 2. Forensic Motivation & Root-Cause Matrix

```
┌───────────────────────────┬──────────────────────────────────┬───────────────────────────────────────────┐
│ Observed Failure          │ Root Cause in Code               │ Impact on AI Agents & Graph Analytics     │
├───────────────────────────┼──────────────────────────────────┼───────────────────────────────────────────┤
│ 71 vs 141 vs 429 clusters │ Missing consolidation in         │ Severe UI cluster thrashing; agent session│
│ on django-sorcery         │ ingestion/watcher + dirty read   │ memory destroyed; unstable boundaries.    │
├───────────────────────────┼──────────────────────────────────┼───────────────────────────────────────────┤
│ >84% of cross-file calls  │ Single-pass linear file walk     │ False "ISOLATED" blast radius scores;     │
│ missing in SurrealDB      │ queries DB before targets exist  │ agents break caller contracts blindly.    │
├───────────────────────────┼──────────────────────────────────┼───────────────────────────────────────────┤
│ Callers vanish during     │ delete_file deletes WHERE out    │ Active graph degradation over editing     │
│ live watch file edits     │ IN $nodes without re-linking     │ sessions; live watch rots graph topology. │
├───────────────────────────┼──────────────────────────────────┼───────────────────────────────────────────┤
│ 24.3 min ingestion for    │ Per-file synchronous HTTP calls  │ Severe violation of NF-001 (<60s/1000f);  │
│ 252 files (5.8s/file)     │ to TEI + dynamic SQL subqueries  │ unusable on enterprise repositories.      │
├───────────────────────────┼──────────────────────────────────┼───────────────────────────────────────────┤
│ Receiver types dropped    │ rsplit('.').next() strips        │ Type references and object receiver       │
│ in method call chains     │ callers down to terminal verb    │ dependencies are completely invisible.    │
├───────────────────────────┼──────────────────────────────────┼───────────────────────────────────────────┤
│ Vector search table scan  │ Missing KNN `<|k|>` operator     │ O(N) linear search latency scaling with   │
│ in SurrealDB              │ in search_vector query           │ codebase size rather than O(log N) HNSW.  │
└───────────────────────────┴──────────────────────────────────┴───────────────────────────────────────────┘
```

---

## 3. Detailed Requirements & Traceability Matrix

| Req ID | Domain | Requirement Statement | Priority | Status | Verification Target |
|:---|:---|:---|:---:|:---:|:---|
| **R-077** | Clustering | **Unified Clustering Pipeline Invariant**: Both `ingestion/mod.rs` and `watcher/mod.rs` MUST execute identical two-phase clustering (`CommunityDetector::detect_with_seeds` followed by `CommunityDetector::consolidate_communities`) with identical resolution and `min_size` parameters, guaranteeing that automated and manual reclustering yield 100% identical galaxy counts. | P0 | 🔨 Ready | Verify `job.clusters_computed == total_communities` between ingestion completion and subsequent `/api/cluster` call. |
| **R-078** | Ingestion | **Two-Pass Ingestion Engine**: Ingestion MUST execute in two decoupled phases:<br>• **Pass 1 (Symbol Declaration Pass)**: Walk directory, parse all file ASTs, build in-memory Symbol Table, embed vectors in bulk batches (up to 256 texts per request), and insert all `node` and topological containment edges (`CONTAINS`, `CONTAINS_FILE`, `CONTAINS_DIR`) into SurrealDB.<br>• **Pass 2 (Deferred Linkage Pass)**: Traverse all extracted cross-file edges (`CALLS`, `TYPE_REF`, `IMPORTS`, `IMPLEMENTS`) in memory against the complete Symbol Table, resolving targets with zero forward-reference dropouts. | P0 | 🔨 Ready | Prove that cross-file calls (e.g. `test_profiler.py` to `SQLAlchemyProfiler`) resolve with 100% success regardless of directory walk order. |
| **R-079** | Watcher | **Inbound Edge Preservation in Delta-Sync**: `IncrementalPipeline::reindex_file` and `db.delete_file` MUST NOT execute unconditional deletion of incoming edges (`out IN $nodes`). Instead, existing incoming edges targeting surviving symbols MUST be preserved or re-linked to matching symbol signatures in the re-indexed file. | P0 | 🔨 Ready | Edit a target file under live watch and verify callers in other files retain active `CALLS` edges in SurrealDB. |
| **R-080** | API / Safety | **Transactional Read-Fencing & Staging Barrier**: Workspaces undergoing active ingestion (`phase == "scanning"` or `"indexing"`) MUST enforce a read barrier. Endpoints `/api/cluster`, `/api/galaxies`, and `/api/galaxy/topology` MUST return HTTP 423 (Locked) or a structured staging status `{ "status": "indexing", "progress_pct": N }`, preventing premature clustering and corrupt dirty reads. | P1 | 🔨 Ready | Concurrent requests to `/api/cluster` during bulk ingestion are cleanly locked until `phase == "completed"`. |
| **R-081** | Performance | **High-Throughput Parallel Ingestion Engine**: AST parsing MUST be parallelized across available CPU cores (via `rayon` or tokio threadpool), TEI embeddings MUST be batched across files (256 texts/batch), and SurrealDB inserts MUST use parameterized bulk array operations. Ingestion of 1,000 polyglot files MUST complete in < 60 seconds (NF-001). | P1 | 🔨 Ready | Ingest `django-sorcery` (252 files) in < 45 seconds (down from 1,456 seconds). |
| **R-082** | AST / Parser | **Receiver & Explicit Import Scoping (SAIR Tier 2)**: Parser MUST preserve receiver expressions (`User.objects.create()` $\to$ `User`), and edge resolution MUST resolve cross-file targets via explicit file `import`/`use` declarations before falling back to package-directory or global indices. | P1 | 🔨 Ready | Verify `User.objects.create()` emits a `REFERENCES` edge to `User`, and explicit imports resolve to the exact imported module. |
| **R-083** | Database | **SurrealDB True HNSW KNN Query Activation**: `search_vector` MUST use SurrealDB v2's native KNN `<|k, ef|>` vector operator rather than `ORDER BY vector::similarity::cosine(...)`, enabling true $O(\log N)$ index traversal. | P1 | 🔨 Ready | Verify vector search query execution plan activates `idx_node_embedding`. |
| **R-084** | Topology | **Hierarchical Leiden Galaxy Coarsening**: The community detection engine MUST implement recursive multi-level aggregation (contracting communities into meta-nodes) to guarantee clean logarithmic scaling ($5 \le N_{\text{galaxies}} \le 15$) for macro architectural vision, eliminating reliance on API truncation (`limit=15`). | P2 | 🔨 Ready | `PolyTuring` and `django-sorcery` generate between 5 and 15 cohesive macro-galaxies without API truncation. |
| **R-085** | Analytics | **Package Metrics Completeness ($C_a, C_e, I, A, D$)**: Automate Uncle Bob's Abstractness ($A = N_{\text{abstract}} / N_{\text{total}}$) and Distance from Main Sequence ($D = |A + I - 1|$), returning them in `GalaxyRecord` and `/api/galaxy/topology`. | P2 | 🔨 Ready | Galaxy records contain valid $A \in [0, 1]$ and $D \in [0, 1]$ values matching theoretical bounds. |

---

## 4. Technical Architecture & Data Flow

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                      TWO-PASS COMPILATION & RESOLUTION PIPELINE                        │
├────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                        │
│   [PASS 1: SYMBOL DECLARATION & STRUCTURAL BACKBONE]                                   │
│                                                                                        │
│   Source Files (Rayon Parallel)                                                        │
│         │                                                                              │
│         ├──► Tree-sitter AST Extraction ──► In-Memory Symbol Table                     │
│         │                                   • (ws, symbol) ──► NodeMetadata            │
│         │                                   • (ws, file)   ──► ImportsTrie             │
│         │                                   • Unresolved ExtractedEdges Queue          │
│         │                                                                              │
│         ├──► Cross-File Chunked Embedder Queue (256 symbols/batch) ──► TEI Engine      │
│         │                                                                              │
│         └──► Bulk Parameterized Inserts ──► SurrealDB (node & CONTAINS edges)          │
│                                                                                        │
│ ────────────────────────────────────────────────────────────────────────────────────── │
│                                                                                        │
│   [PASS 2: DEFERRED SAIR LINKAGE PASS]                                                 │
│                                                                                        │
│   Unresolved Edges Queue                                                               │
│         │                                                                              │
│         ├──► Tier 1: Local File Scope? (In-memory lookup)                              │
│         ├──► Tier 2: Explicit Import Scope? (Match against file ImportsTrie)           │
│         ├──► Tier 3: Sibling Directory Scope? (Prefix match in Symbol Table)           │
│         └──► Tier 4: Global Index Fallback (Excluding generic verb blacklist)          │
│                                                                                        │
│         ▼ (100% In-Memory Resolution — Zero DB Query Overhead)                         │
│   Fully Resolved Edges [source_id ──► target_id]                                       │
│         │                                                                              │
│         └──► Parameterized Bulk RELATE ──► SurrealDB (linked_to table)                 │
│                                                                                        │
│ ────────────────────────────────────────────────────────────────────────────────────── │
│                                                                                        │
│   [POST-INGESTION: UNIFIED TWO-PHASE MODULARITY]                                       │
│                                                                                        │
│   SurrealDB Full Graph                                                                 │
│         │                                                                              │
│         ├──► Phase 1: Warm-Seeded LPA (detect_with_seeds)                              │
│         │                                                                              │
│         └──► Phase 2: Hierarchical Consolidation (consolidate_communities, min_size=5) │
│                     • Topological Edge Affinity                                        │
│                     • Spatial Directory Locality                                       │
│                     • Dominant Cluster Rollup                                          │
│                                                                                        │
│         ▼                                                                              │
│   Deterministic, Stable Subsystem Galaxy Topology (Persisted to SurrealDB)             │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 5. Interface & Schema Contracts

### 5.1 SurrealQL Schema Additions (`schema.surql`)

```sql
-- Galaxy table schema updates for full Uncle Bob metrics (R-085)
DEFINE TABLE galaxy SCHEMAFULL;
DEFINE FIELD workspace          ON TABLE galaxy TYPE string;
DEFINE FIELD galaxy_id          ON TABLE galaxy TYPE int;
DEFINE FIELD name               ON TABLE galaxy TYPE string;
DEFINE FIELD dominant_path      ON TABLE galaxy TYPE string;
DEFINE FIELD node_count         ON TABLE galaxy TYPE int;
DEFINE FIELD internal_edges     ON TABLE galaxy TYPE int;
DEFINE FIELD external_edges     ON TABLE galaxy TYPE int;
DEFINE FIELD afferent_coupling  ON TABLE galaxy TYPE int;
DEFINE FIELD efferent_coupling  ON TABLE galaxy TYPE int;
DEFINE FIELD instability        ON TABLE galaxy TYPE float;
DEFINE FIELD abstractness       ON TABLE galaxy TYPE float;    -- R-085: A = N_abstract / N_total
DEFINE FIELD distance           ON TABLE galaxy TYPE float;    -- R-085: D = |A + I - 1|
DEFINE FIELD role               ON TABLE galaxy TYPE string;
DEFINE FIELD key_symbols        ON TABLE galaxy TYPE array<string>;
DEFINE FIELD languages          ON TABLE galaxy TYPE array<string>;
DEFINE FIELD updated_at         ON TABLE galaxy TYPE datetime DEFAULT time::now();
```

### 5.2 API Interface Contracts

#### REST API: `POST /api/cluster` and `GET /api/galaxies` Read-Fencing
When an ingestion job is active for the target workspace, the API returns a structured HTTP 423 status:

```diff
- HTTP 200 OK
- { "status": "success", "total_communities": 76, "galaxies": [...] }
+ HTTP 423 Locked
+ {
+   "error": "Workspace 'django-sorcery' is currently indexing",
+   "phase": "indexing",
+   "files_indexed": 133,
+   "total_files": 252,
+   "retry_after_ms": 2000
+ }
```

#### Ingestion Response Telemetry
`IngestResult` and `/api/ingest/status` expose edge resolution completeness metrics:

```diff
  pub struct IngestResult {
      pub workspace: String,
      pub files_scanned: usize,
      pub files_indexed: usize,
      pub files_skipped: usize,
      pub nodes_created: usize,
      pub edges_created: usize,
+     pub cross_file_edges_resolved: usize,
+     pub unresolvable_external_edges: usize,
      pub clusters_computed: usize,
      pub duration_ms: u64,
  }
```

---

## 6. Acceptance Criteria

| AC ID | Criteria | Traces To |
|:---|:---|:---|
| **AC-021** | Given a repository ingestion, when auto-clustering completes, then `job.clusters_computed` matches the exact count returned by an immediate subsequent call to `/api/cluster`. | R-077 |
| **AC-022** | Given two files $A$ and $B$ where $A$ precedes $B$ in directory traversal and $A$ invokes a symbol defined in $B$, when Two-Pass Ingestion completes, then a directional `CALLS` edge exists in SurrealDB from $A$ to $B$. | R-078 |
| **AC-023** | Given a file $B$ with incoming call edges from $A$, when $B$ is modified and re-indexed under live-watch, then the incoming edge from $A$ to $B$ remains intact in the graph. | R-079 |
| **AC-024** | Given an active ingestion in progress (`phase == "indexing"`), when `/api/cluster` or `/api/galaxy/topology` is requested for that workspace, then the API returns HTTP 423 with `retry_after_ms`. | R-080 |
| **AC-025** | Given a repository of 250 files, when ingested with bulk vectorization and parallel AST parsing, then ingestion completes in $< 45$ seconds. | R-081, NF-001 |
| **AC-026** | Given a method invocation on an object `User.objects.create()`, when parsed, then an edge referencing `User` is emitted. | R-082 |
| **AC-027** | Given vector similarity search, when `/api/search` executes, then SurrealDB uses HNSW index rather than full table scan. | R-083 |
| **AC-028** | Given a codebase of $>2,000$ nodes, when community detection runs, then the total number of macro-galaxies generated is between 5 and 15 without API truncation. | R-084 |
| **AC-029** | Given `/api/galaxy/topology`, when galaxy metrics return, then each galaxy includes `abstractness` and `distance` from the main sequence. | R-085 |

---

## 7. Revision History

| Date | Author | Changes |
|:---|:---|:---|
| 2026-11-05 | Antigravity (Google DeepMind) & @vadde | Final comprehensive specification for Omni-Graph v0.5.0 Hardening (R-077 through R-085). |
