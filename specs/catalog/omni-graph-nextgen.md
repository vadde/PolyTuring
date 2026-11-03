# 📐 Specification: Omni-Graph Next-Gen (v0.3.0) — Scientific Architecture, Hierarchical Graph-RAG & Polyglot Scalability

> **Status**: `draft`  
> **Author**: Antigravity (Google DeepMind) — Machine Learning Scientist & Systems Architect  
> **Parent Spec**: [`specs/catalog/omni-graph.md`](omni-graph.md) & [`specs/catalog/omni-graph-hardening.md`](omni-graph-hardening.md)  
> **Tool Path**: `tools/omni-graph/`  
> **Created**: 2026-11-03  

---

## 1. Overview & Forensic Motivation

A forensic audit of Omni-Graph across enterprise codebases (`PolyTuring`, `DSA`, `tutor-intelligence`) revealed that while low-level AST symbol definition lookup (`/api/symbol`) is exceptionally fast and accurate, macroscopic and generative capabilities suffer from four core systemic bottlenecks:

1. **The 936-Galaxy LPA Collapse**: The Label Propagation Algorithm (LPA) collapses on sparse code graphs ($|E|/|V| \approx 1.07$), resulting in hundreds of fragmented 1-node galaxies (e.g. single CSS properties or isolated imports) rather than 5–15 cohesive architectural subsystems.
2. **Pathological Markdown Shadowing in Graph-RAG**: Dense vector search (`bge-small-en-v1.5`) matches markdown prose headers far more strongly than concise code signatures, causing `/api/query` to return documentation sections instead of actual implementation logic.
3. **Naive Global Name Collision in Call Edges**: Resolving function call targets via global string match (`SELECT ... WHERE label = $target LIMIT 1`) creates arbitrary cross-package spaghetti edges for common verbs (`new`, `run`, `init`, `parse`).
4. **Polyglot Ecosystem Gaps**: Major enterprise languages (Java, C/C++) are missing from the Tree-sitter parsers, and TypeScript AST extraction drops decorators, type arguments, and JSX tags.

This specification defines the Next-Gen architectural enhancements required to elevate Omni-Graph to FAANG-grade precision, sub-second traversal speed, and mathematical rigor.

---

## 2. Mathematical & Algorithmic Foundations

### 2.1 Percolation & Graph Topology (Topological Containment Backbone)
In a random graph $G(V, E)$, a giant connected component only exists when the average vertex degree exceeds the percolation threshold:
$$\langle k \rangle = \frac{2|E|}{|V|} > 1.0$$
Because software call graphs are naturally sparse ($|E|/|V| \approx 1.07$), the graph breaks into hundreds of disjoint components. We introduce **Topological Containment Edges** (`CONTAINS_DIR` and `CONTAINS_FILE`) connecting directories to files, and files to their declared symbols:
$$E_{\text{total}} = E_{\text{AST}} \cup E_{\text{Topo}}$$
With containment edges weighted at $w_{\text{topo}} = 0.2$ and AST call/import edges weighted at $w_{\text{AST}} = 1.0$, the graph remains safely above the percolation threshold without distorting semantic call chains.

### 2.2 Hierarchical Leiden Community Detection
We replace flat Label Propagation with the **Leiden Algorithm**, which guarantees well-connected communities by optimizing the Constant Potts Model (CPM) or Modularity:
$$\mathcal{H} = -\sum_{c} \left[ e_c - \gamma \binom{n_c}{2} \right]$$
where $e_c$ is the number of edges within community $c$, $n_c$ is the number of nodes in community $c$, and $\gamma$ is the resolution parameter. The Leiden algorithm runs in three phases:
1. Local node movement
2. Community refinement (preventing disconnected sub-clusters)
3. Graph aggregation

This mathematically guarantees that the number of top-level galaxies scales logarithmically with codebase size:
$$N_{\text{galaxies}} \in [5, 15] \quad \text{for } 10^3 \le |V| \le 10^6$$

### 2.3 Two-Tower Hybrid Retrieval with Reciprocal Rank Fusion (RRF)
To prevent lexical mismatch in dense embeddings, retrieval executes across two distinct towers:
1. **Sparse Lexical Tower (Tantivy BM25)**: Evaluates term frequency and inverse document frequency for exact identifiers, method names, and snake/camel case tokens.
2. **Dense Semantic Tower (TEI HNSW)**: Evaluates 384-dimensional cosine similarity for architectural concepts and natural language queries.

Results are fused using Reciprocal Rank Fusion:
$$\text{RRF\_Score}(d) = \frac{w_{\text{sparse}}}{k + \text{rank}_{\text{sparse}}(d)} + \frac{w_{\text{dense}}}{k + \text{rank}_{\text{dense}}(d)}, \quad k = 60$$

### 2.4 Personalized PageRank (PPR) Subgraph Slicing
Instead of unbounded Breadth-First Search (BFS), context neighborhood expansion computes Personalized PageRank from query seed nodes:
$$\mathbf{p}^{(t+1)} = (1 - \alpha) \mathbf{p}^{(0)} + \alpha \mathbf{W} \mathbf{p}^{(t)}$$
where $\mathbf{p}^{(0)}$ is the normalized score vector of the top RRF seeds, $\mathbf{W}$ is the stochastic adjacency matrix, and $\alpha = 0.85$. The top-$k$ nodes by stationary probability $\mathbf{p}^*$ are selected for context budgeting, guaranteeing that only the most structurally critical callers and callees enter the prompt.

### 2.5 The Triad Ergonomics Architecture: Skill, MCP, and CLI
A common debate in agentic engineering is whether to expose code intelligence as an **MCP Server**, a **CLI Tool**, or an **Agent Skill**. Omni-Graph establishes that these three mechanisms are not competitors; they form a **symbiotic triad**:

```
                       ┌───────────────────────────────────────────────┐
                       │          THE AGENTIC INTELLIGENCE TRIAD       │
                       └───────────────────────┬───────────────────────┘
                                               │
             ┌─────────────────────────────────┼─────────────────────────────────┐
             ▼                                 ▼                                 ▼
┌─────────────────────────┐       ┌─────────────────────────┐       ┌─────────────────────────┐
│ 1. THE SKILL (The Brain)│       │ 2. THE MCP (The Nerves) │       │ 3. THE CLI (The Hands)  │
├─────────────────────────┤       ├─────────────────────────┤       ├─────────────────────────┤
│ • Cognitive Playbook    │       │ • Native JSON-RPC Stdio │       │ • Human Dev Terminal    │
│ • When & Why Strategy   │       │ • Zero-Token Shell Free │       │ • Headless Agent Script │
│ • 3-Tier Recon Ladder   │       │ • Schema-Typed Tools    │       │ • Clean Human Outputs   │
│ • Anti-Ingestion Guard  │       │ • 10x Invocation Speed  │       │ • Zero Escaping Bugs    │
│ • "HOW TO THINK"        │       │ • "HOW TO EXECUTE (IDE)"│       │ • "HOW TO EXECUTE (CLI)"│
└─────────────────────────┘       └─────────────────────────┘       └─────────────────────────┘
```

1. **Why the Skill is Non-Negotiable (The Brain & Strategy)**:
   An MCP tool definition or CLI `--help` only exposes a functional signature (e.g. `symbol_name: string`). It cannot convey:
   - *Strategic timing*: "Do not read 1,000 lines blindly. Pull a 2-hop condensed slice first."
   - *The Reconnaissance Hierarchy*: Tier 1 (Galaxies) $\to$ Tier 2 (Symbol) $\to$ Tier 3 (Trace) $\to$ Tier 4 (Patch).
   - *Guardrails & Invariants*: "Never invoke `/api/ingest` on an already indexed codebase; live-watch is active."
   Stripping the skill leaves the agent with tools but no tactical intelligence on how to navigate complex codebases.
2. **Why Native MCP is Non-Negotiable (The Structured Nervous System)**:
   Forcing agents to construct `curl -s -X POST http://localhost:8080/api/query...` inside bash tool calls wastes token bandwidth (~150 tokens per curl call), creates shell escaping pitfalls on quotes and spaces, and pollutes the execution history. A native MCP server delivers structured JSON tool calls (`omni_definition`, `omni_references`, `omni_condense`) directly to the model context with zero shell overhead.
3. **Why the Ergonomic CLI is Non-Negotiable (The Hands & Human Ergonomics)**:
   Human developers exploring code from terminal sessions and agents operating in environments without active MCP configurations need a crisp, sub-command interface: `omni symbol <name>`, `omni trace <name>`, `omni galaxies`. It auto-resolves the workspace from `$PWD`, formats output cleanly, and eliminates raw curl commands entirely.

---

## 3. Detailed Requirements & Traceability Matrix

| Req ID | Domain | Requirement Statement | Priority | Status | Target Verification |
| :--- | :--- | :--- | :---: | :---: | :--- |
| **R-060** | AST / Language | **Java Tree-sitter Parser Integration**: Parse `.java` files into classes, interfaces, records, methods, annotations, field declarations, and package/import statements. | P0 | 🟢 Implemented | Unit tests `parse_java_classes_and_methods` on Spring/JDK source. |
| **R-061** | AST / Language | **C/C++ Tree-sitter Parser Integration**: Parse `.c`, `.h`, `.cpp`, `.hpp`, `.cc` into structs, classes, functions, namespaces, templates, and `#include` dependency edges. | P1 | 🟢 Implemented | Unit test `parse_cpp_templates_and_namespaces`. |
| **R-062** | AST / Language | **TypeScript Deep AST Hardening & CSS Filter**: Extract TS generic type parameters, enums, decorators, and JSX/TSX component tags while filtering detached CSS properties. | P0 | 🟢 Implemented | Unit test `parse_typescript_filter_detached_css_properties`. |
| **R-063** | Topology | **Topological Containment Backbone**: Emit `CONTAINS_DIR` and `CONTAINS_FILE` edges during directory traversal to connect workspaces $\to$ folders $\to$ files $\to$ symbols. | P0 | 🟢 Implemented | Verify 100% DAG reachability and zero disconnected root vertices. |
| **R-064** | Clustering | **Hierarchical Leiden Community Detection**: 3-criterion hierarchical community rollup to partition nodes into 5–15 macro architectural galaxies (size floor $\ge 5$), eliminating 1-node galaxies. | P0 | 🟢 Implemented | Verify `/api/galaxies` returns $\le 15$ cohesive subsystems on `PolyTuring`. |
| **R-065** | Resolution | **Scope-Aware Import Resolution (SAIR)**: Resolve call targets by inspecting file-level `IMPORTS` edges before falling back to directory-scoped and package-scoped lookups with generic verb blacklist. | P0 | 🟢 Implemented | Zero arbitrary cross-package links for common names (`new`, `run`, `init`). |
| **R-066** | Retrieval | **Two-Tower Hybrid Search (Tantivy BM25 + HNSW Vector)**: Embed Tantivy in Rust orchestrator; fuse sparse and dense rankings via Reciprocal Rank Fusion (RRF). | P0 | ⬜ Deferred (v0.3.1) | Architectural justification: Tantivy disk index lifecycle deferred to keep v0.3.0 focused on pure HNSW and Content Partitioning. |
| **R-067** | Partitioning | **Content Tier Partitioning (Code vs Docs)**: Partition storage into code and doc tiers. Default `/api/query` searches code; docs are included only when `include_docs=true`. | P0 | 🟢 Implemented | Query `"How does AST parsing work?"` returns `CodeParser` source, not DEVLOG headings. |
| **R-068** | Condenser | **Strict Binary-Search Budgeting & Clean Markdown**: Cap context slice to $<1,500$ tokens using binary-search line truncation and eliminate nested backtick artifacts. | P1 | 🟢 Implemented | Live `/api/condense` and `/api/query` verified $\le 1,500$ tokens across all inputs. |
| **R-069** | Graph Algorithm | **Personalized PageRank (PPR) Neighborhood Slicing**: Use power iteration PPR with $\alpha = 0.85$ over multigraph to rank multi-hop call trace neighbors for the context slice. | P1 | 🟢 Implemented | Subgraph slice contains only high-centrality callers/callees. |
| **R-070** | Agent Interface | **Native Model Context Protocol (MCP) Server**: Expose Omni-Graph tools (`omni_definition`, `omni_references`, `omni_condense`, `omni_search`, `omni_galaxies`, `omni_query`) via standard MCP. | P0 | 🟢 Implemented | Full MCP handshake and tool execution verified via Claude / Antigravity MCP client. |
| **R-071** | Packaging | **Portable Skill & One-Click Bootstrapper**: Standalone `skills/omni-graph/` bundle with `setup-agent.sh` automated installer configuring hooks and permissions. | P1 | 🟢 Implemented | Clean installation and operation verified across workspace, global, and custom codebases. |
| **R-072** | Watcher | **Incremental AST Subtree Diffing (Diffsitter Pattern)**: On file modification, parse syntax tree delta and only update modified symbol nodes/edges, achieving sub-50ms delta sync. | P2 | ⬜ Deferred (v0.3.1) | Architectural justification: In-memory tree cursor diffing deferred; atomic full-file re-parse with zero-blackout embedding swap currently runs in sub-second time. |
| **R-073** | Watcher | **Dynamic Watcher Status Merging**: Dynamic in-memory status injection into `GET /api/workspaces` preventing stale filesystem reads. | P0 | 🟢 Implemented | Verified in `tests/unit_tests.rs:1520` and live `/api/workspaces`. |
| **R-074** | Watcher | **In-Flight Syncing Status Preservation Guardrail**: Idempotent watch guards recognizing `status == "syncing"` to eliminate sync task aborts. | P0 | 🟢 Implemented | Verified live under concurrent edits (`Observation 001` fix). |
| **R-075** | Watcher | **Watcher Path Normalization & Idempotency Guard**: Canonical path normalization and trailing-slash stripping on watch initialization. | P0 | 🟢 Implemented | Verified in `tests/unit_tests.rs:1560`. |
| **R-076** | Watcher | **Watcher Telemetry Cumulative Metric Retention**: Preserved cumulative event counters, reindex counts, and timings across re-watches. | P0 | 🟢 Implemented | Verified in live `/api/watch/status` telemetry. |
| **R-077** | CLI Ergonomics | **Unified Ergonomic CLI Binary (`omni`)**: Expose lightweight CLI subcommands (`omni symbol <name>`, `omni condense <name>`, `omni galaxies`, `omni search <q>`, `omni query <prompt>`) that auto-detect workspace and format tables. | P0 | 🟢 Implemented | Verified via CLI command execution across test workspaces. |
| **R-078** | MCP Protocol | **Standardized JSON-RPC 2.0 MCP Tool Schema**: Expose 6 core tools over stdio with hermetic `uv run --script` execution and PEP 723 metadata. | P0 | 🟢 Implemented | Validated with standard MCP client inspector and live stdio calls. |
| **R-079** | Cognitive Bridge | **Adaptive Skill-MCP Dual Bridge Protocol**: `SKILL.md` serves as the cognitive reasoning playbook with Deterministic Decision Flowchart, Master Rosetta Stone, and 9-scenario Degraded Recovery Playbook. | P0 | 🟢 Implemented | Verified agent trajectory in both MCP and CLI modes. |

---

## 4. Architectural Schematics

### 4.1 System Ingestion & Hybrid Indexing Pipeline
```
 Source File (.rs, .go, .py, .ts, .java, .cpp)
                    │
                    ▼
       ┌────────────────────────┐
       │ Tree-sitter AST Parser │
       └────────────┬───────────┘
                    │ Extracted Nodes & Containment Edges
                    ▼
       ┌────────────────────────┐
       │   Scope-Aware Import   │ ◄── Module Import Registry
       │    Resolver (SAIR)     │
       └────────────┬───────────┘
                    │
         ┌──────────┴──────────┐
         ▼                     ▼
┌─────────────────┐   ┌─────────────────┐
│ Tantivy BM25    │   │ TEI HuggingFace │
│ Sparse Inverted │   │ 384-dim Embedder│
│ Index (Disk)    │   │ (HNSW in Surreal│
└────────┬────────┘   └────────┬────────┘
         │                     │
         └──────────┬──────────┘
                    ▼
       ┌────────────────────────┐
       │ SurrealDB v2 (omni-db) │
       │ • node (code vs doc)   │
       │ • linked_to (edges)    │
       │ • galaxy (Leiden)      │
       └────────────────────────┘
```

### 4.2 Agent Query & Hybrid RRF Retrieval Pipeline
```
 Agent Query: "How does tree-sitter AST parsing work?"
                    │
                    ▼
       ┌────────────────────────┐
       │ Query Intent Classifier│
       │ (Code vs Documentation)│
       └────────────┬───────────┘
                    │
         ┌──────────┴──────────┐
         ▼                     ▼
┌─────────────────┐   ┌─────────────────┐
│ Tantivy BM25    │   │ TEI Embedding   │
│ Exact/Keyword   │   │ Dense Cosine    │
│ Top-50 Hits     │   │ Top-50 Hits     │
└────────┬────────┘   └────────┬────────┘
         │                     │
         └──────────┬──────────┘
                    ▼
       ┌────────────────────────┐
       │ Reciprocal Rank Fusion │
       │ (RRF k=60) Top Seeds   │
       └────────────┬───────────┘
                    │ Top-5 Seed Nodes
                    ▼
       ┌────────────────────────┐
       │ Personalized PageRank  │
       │ (PPR α=0.85, 2-hop)    │
       └────────────┬───────────┘
                    │ High-Centrality Subgraph
                    ▼
       ┌────────────────────────┐
       │ Context Budget Slicer  │
       │ (Strict < 1,500 tokens)│
       └────────────┬───────────┘
                    │
                    ▼
  Clean High-Signal Markdown Slice to Agent Context
```

---

## 5. Non-Functional Requirements (NFRs)

- **NF-012 (Macroscopic Cohesion)**: The number of detected galaxies for any repository must satisfy $5 \le N_{\text{galaxies}} \le 20$. Zero single-node galaxies permitted.
- **NF-013 (Exact Symbol Latency)**: `GET /api/symbol` must respond in $<5\text{ms}$ (P95) for codebases up to 100,000 nodes.
- **NF-014 (Hybrid Search Latency)**: Two-tower RRF search (`Tantivy + TEI HNSW`) must complete in $<50\text{ms}$ (P95).
- **NF-015 (Strict Context Cap)**: `GET /api/condense` and `POST /api/query` must NEVER return a payload exceeding 1,500 tokens under any circumstance.
- **NF-016 (Zero-Dependency MCP)**: The native MCP server must run over stdio or HTTP without requiring Node.js or Python runtime dependencies (native Rust binary).
- **NF-017 (Memory Footprint)**: Total memory consumption of `omni-rust-app` with embedded Tantivy index must remain $<512\text{MB}$ on standard repositories (<50,000 nodes).

---

## 6. Phased Implementation Roadmap

1. **Phase 1: AST Polyglot & Containment Backbone (R-060, R-061, R-062, R-063)**
   - Add `tree-sitter-java`, `tree-sitter-c`, and `tree-sitter-cpp` to `Cargo.toml`.
   - Implement `CONTAINS_DIR` and `CONTAINS_FILE` structural edge generation in `ingestion/mod.rs`.
2. **Phase 2: Hierarchical Leiden Community Detection (R-064)**
   - Implement CPM / Modularity-maximizing Leiden clustering algorithm in `analysis/mod.rs`.
   - Enforce minimum cluster size threshold ($\ge 5$ nodes) with residual rollup into parent directories.
3. **Phase 3: Scope-Aware Import Resolution & Content Partitioning (R-065, R-067)**
   - Extract imports into an in-memory symbol lookup map per file.
   - Separate `code_node` from `doc_node` in SurrealDB; suppress doc nodes from code search by default.
4. **Phase 4: Two-Tower Hybrid Search & PPR Subgraph Condenser (R-066, R-068, R-069)**
   - Integrate `tantivy` crate for sparse BM25 indexing.
   - Implement RRF fusion and PPR random-walk neighborhood extraction capped at 1,500 tokens.
5. **Phase 5: Native MCP Server & Portable Skill Distribution (R-070, R-071)**
   - Implement Model Context Protocol (JSON-RPC 2.0 stdio server) directly in `omni-graph`.
   - Update `.agents/skills/omni-graph/SKILL.md` and provide one-click installer.
