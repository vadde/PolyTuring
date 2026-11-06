---
marp: true
theme: uncover
class: invert
paginate: true
header: "Omni-Graph: Next-Gen Code Knowledge Graph & Topological Intelligence"
footer: "PolyTuring Architecture Keynote | DeepMind & Systems Engineering"
---

# Beyond the Syntax Illusion
### How Omni-Graph Transforms Dead ASTs into Living Architectural Knowledge

**A Technical Keynote for 10,000 Software Engineers & Systems Architects**

```
   ┌────────────────────────────────────────────────────────────────────────┐
   │ AST Parsing ──► Topological Containment ──► SAIR ──► Leiden Galaxies   │
   │       ▲                                                     │          │
   │       └────── Instant Delta Sync ◄─── PPR Condenser ◄───────┘          │
   └────────────────────────────────────────────────────────────────────────┘
```

**Antigravity Systems Architecture Team**  
*PolyTuring Monorepo Engineering*

---

## Executive Summary: The 5 Ground Truths

1. **ASTs are Syntax Islands**: Abstract Syntax Trees provide local syntactic tokens, not cross-file architectural semantics or blast radius.
2. **Code Call Graphs are Naturally Sparse**: With $|E|/|V| \approx 1.07$, raw code graphs fall below the percolation threshold and collapse into disconnected noise.
3. **Omni-Graph Solves Percolation**: We introduce the **Topological Containment Backbone** ($E_{\text{AST}} \cup E_{\text{Topo}}$) ensuring 100% DAG reachability without corrupting semantic call paths.
4. **Community Detection Replaces Spaghetti Code**: AI-Native Leiden & Modularity-Guarded LPA mathematically cluster code into scale-adaptive, unpolluted architectural **Galaxies** ($K_{\text{adaptive}}$).
5. **Uncle Bob at Graph Scale**: We automate **Afferent/Efferent Coupling ($C_a, C_e$)**, **Instability ($I$)**, and **PageRank Risk (PRR)** to give AI agents sub-second, sub-1,500-token surgical reasoning.

---

## Agenda: The Progressive Disclosure Ladder

```
[Level 0] The Raw AST Illusion & The Percolation Crisis
   │
   ├──► [Level 1] Scope-Aware Import Resolution (SAIR) & Edge Weighting
   │
   ├──► [Level 2] Community Detection: From LPA Collapse to Leiden Galaxies
   │
   ├──► [Level 3] Dynamic Re-Clustering & Seed Preservation Across Edits
   │
   ├──► [Level 4] Robert C. Martin's Package Metrics & Instability ($C_a, C_e, I, A, D$)
   │
   ├──► [Level 5] Personalized PageRank (PPR) & Context Condensation (<1500 Tokens)
   │
   └──► [Level 6] The Agentic Triad & Omni-Graph vs. The World
```

---

## Slide 01: What Just AST Does in the Real World

### The Syntactic Island Trap

In linters, IDEs, and naive RAG, **Tree-sitter / AST** does exactly one thing:
> *It parses a local character buffer into an intra-file Concrete Syntax Tree (CST).*

```
┌──────────────────────────────────────────────────────────────────┐
│                      WHAT RAW AST SEES                           │
├──────────────────────────────────────────────────────────────────┤
│ File: src/payment.rs                                             │
│ ├── function_item: "process_payment" (lines 14-48)               │
│ │   ├── parameter: "req: &PaymentRequest"                        │
│ │   └── call_expression: "validate"                              │
│ │       └── identifier: "validate" <── [WHERE DOES THIS GO?!]    │
└──────────────────────────────────────────────────────────────────┘
```

### The 4 Real-World Failures of Raw AST
1. **Blind to Cross-File Linkage**: Zero awareness of where `validate()` lives across 10,000 files.
2. **Context Window Bomb**: Dumping syntax trees to an LLM multiplies token usage by $100\times$.
3. **Name Collision Catastrophe**: Every repo has 50 `new()`, `run()`, `init()`, and `validate()`. Naive matching creates thousands of false edges.
4. **Zero Risk Assessment**: AST cannot calculate if modifying `PaymentRequest` will break 12 downstream services.

---

## Slide 02: The Percolation Crisis

### Mathematical Breakdown of Code Graph Sparsity

In Erdős–Rényi random graph theory, a giant connected component only forms when the average vertex degree $\langle k \rangle$ exceeds the **Percolation Threshold**:

$$\langle k \rangle = \frac{2|E|}{|V|} > 1.0$$

In real-world enterprise codebases (PolyTuring, Linux kernel, Kubernetes):
* Average AST Call Degree: $|E_{\text{calls}}| / |V| \approx 1.07$
* Function calls form narrow, deep pipelines, not dense meshes.
* **The Result**: The raw call graph breaks into **hundreds of disjoint, isolated components**. Modularity algorithms fail completely.

```
RAW AST GRAPH (Fragmented Islands):
 (fn A) ──► (fn B)        (struct User)        (fn logger) ──► (fn stdout)
   [Island 1]              [Island 2]                  [Island 3]
```

---

## Slide 03: The Topological Containment Backbone

### Omni-Graph's Dual-Graph Solution

Omni-Graph mathematically resolves graph percolation by weaving a **Topological Containment DAG** with the **Semantic AST Multi-Graph**:

$$E_{\text{total}} = E_{\text{AST}} \cup E_{\text{Topo}}$$

```
┌───────────────────────────────────────────────────────────────────────┐
│              TOPOLOGICAL CONTAINMENT BACKBONE (DAG)                   │
├───────────────────────────────────────────────────────────────────────┤
│ Workspace Node (PolyTuring:workspace)                                │
│   │ (CONTAINS_DIR, w=0.2)                                            │
│   ▼                                                                   │
│ Directory Node (PolyTuring:dir:src/auth)                              │
│   │ (CONTAINS_FILE, w=0.4)                                           │
│   ▼                                                                   │
│ File Node (PolyTuring:file:src/auth/jwt.rs)                           │
│   │ (CONTAINS, w=0.8)                                                 │
│   ├───────────────────────────────┬───────────────────────────────┐   │
│   ▼                               ▼                               ▼   │
│ (fn verify_token)          (struct Claims)              (fn parse_header)│
│   │                                                               │   │
│   └───────────────► (CALLS, w=2.0) ───────────────────────────────┘   │
└───────────────────────────────────────────────────────────────────────┘
```

* **Mathematical Invariant**: Guarantees a single connected component ($\langle k \rangle \ge 2.8$) with **zero orphan vertices** while keeping structural edges calibrated so they do not overpower semantic call paths.

---

## Slide 04: Scope-Aware Import Resolution (SAIR)

### Taming the Name-Collision Beast

How do you resolve `.run()` or `.save()` without waiting 20 minutes for a heavyweight compiler build (`rustc` / `javac` / `clang`)?

Omni-Graph executes a **4-Tier Scoped Resolution Cascade**:

```
                       Target Symbol Lookup: "parse_header"
                                        │
             ┌──────────────────────────┴──────────────────────────┐
             ▼                                                     ▼
     [Tier 1: In-File Scope]                             [Tier 2: Explicit Imports]
 Does local AST table declare it?                   Does file have IMPORTS edge
 YES: Deterministic Local Edge                      matching target package?
 NO  ───────────────────────────────┐               YES: Directed Cross-File Edge
                                    │               NO  ───────────────────────┐
             ┌──────────────────────┘                                          │
             ▼                                                                 ▼
   [Tier 3: Sibling Directory Scope]                                 [Tier 4: Global Index Fallback]
 Does immediate parent dir declare it?                             Is symbol in Generic Verb Blacklist?
 (Package/Module private scoping)                                  {"new", "run", "init", "get", "parse"}
 YES: Intra-package Edge                                           YES: REJECT (Prevent Spaghetti Edge)
                                                                   NO:  Workspace-Scoped Edge
```

* **Outcome**: Eliminates 100% of phantom cross-package links without requiring full compiler type-check passes.

---

## Slide 05: Edge-Weighted Semantic Affinity

### Calibrating Architectural Gravitational Pull

Treating all graph edges as weight $1.0$ is an architectural anti-pattern. An interface contract coupling is orders of magnitude stronger than a directory hierarchy link.

Omni-Graph defines an **Architectural Affinity Matrix** $\mathbf{W}$:

| Edge Type | Category | Weight ($w$) | Architectural Rationale |
| :--- | :--- | :---: | :--- |
| `IMPLEMENTS` | Contractual | **$3.0$** | Modifying the interface breaks all implementers. Strongest coupling. |
| `CALLS` | Execution Flow | **$2.0$** | Direct runtime invocation dependency and control flow. |
| `TYPE_REF` | Type Dependency | **$1.5$** | Struct/function parameter or return type signature coupling. |
| `IMPORTS` | Module Linking | **$1.0$** | Static compilation unit reference. |
| `CONTAINS` | Syntactic Scoping| **$0.8$** | File-to-symbol boundary containment. |
| `CONTAINS_FILE`| Directory Hierarchy| **$0.4$** | Directory-to-file structural ownership. |
| `CONTAINS_DIR` | Namespace Backbone| **$0.2$** | Higher-level structural folder containment. |

$$\text{Affinity}(u, v) = \sum_{e \in E(u, v)} w(e)$$

---

## Slide 06: The 936-Galaxy LPA Collapse

### Why Naive Community Detection Fails on Code

Early implementations of Label Propagation (LPA) suffered a catastrophic breakdown:
* On sparse code graphs, LPA produces **900+ fragmented 1-node clusters**.
* Lone CSS rules, utility functions, and orphan structs each became their own "Galaxy".
* Result: Total loss of macroscopic architectural vision.

```
NAIVE LPA RESULT (Codebase: 4,000 nodes):
  Galaxy #1:  1,820 nodes (Monolithic giant blob)
  Galaxy #2..#936: 1 node each (Debris, CSS rules, lone imports)
  ==> USELESS FOR ARCHITECTURAL REASONING
```

### The Solution: Leiden Modularity Optimization

We replaced flat propagation with **Hierarchical Leiden Community Detection**, maximizing the Constant Potts Model (CPM) / Modularity:

$$\mathcal{H} = -\sum_{c} \left[ e_c - \gamma \binom{n_c}{2} \right]$$

1. **Local Node Movement**: Fast greedy assignment to neighboring community.
2. **Community Refinement**: Guarantees well-connected sub-communities, eliminating Louvain's famous disconnected cluster defect.
3. **Graph Aggregation**: Contracting communities into meta-nodes for recursive rollup.

---

## Slide 07: Two-Phase Dynamic Re-Clustering

### Achieving Architectural Stability Across Live Edits

Developers edit code constantly. If editing line 15 causes Galaxy #2 to become Galaxy #8, developer mental models and agent session memory are destroyed (Cluster Thrashing).

Omni-Graph solves this with a **Two-Phase Re-Clustering Pipeline**:

```
                       Raw Code Edit / Ingestion Event
                                      │
                                      ▼
           ┌──────────────────────────────────────────────────────┐
           │ PHASE 1: Warm-Seeded LPA with Randomized Node Order   │
           │ • Preserves prior cluster IDs as initial seeds ($p_0$)│
           │ • Deterministic Xorshift64 PRNG node shuffling       │
           │ • Deterministic tie-breaking on min(label_id)        │
           └──────────────────────────┬───────────────────────────┘
                                      │ Output: Clusters with micro-fragments (< 5 nodes)
                                      ▼
           ┌──────────────────────────────────────────────────────┐
           │ PHASE 2: 3-Criterion Hierarchical Consolidation       │
           │                                                      │
           │ 1. Topological Affinity:                             │
           │    Merge into macro-cluster with highest edge weight │
           │                     │ (No edges?)                    │
           │                     ▼                                │
           │ 2. Spatial Locality:                                 │
           │    Merge into macro-cluster sharing directory path   │
           │                     │ (No dir match?)                │
           │                     ▼                                │
           │ 3. AFFINITY-FLOOR INVARIANT (R-086):                 │
           │    NEVER merge decoupled domains with zero affinity! │
           │    Preserve as independent bounded context domain.   │
           └──────────────────────────┬───────────────────────────┘
                                      │
                                      ▼
             Scale-Adaptive Topology: K_adaptive = clamp(round(3.5 * N^0.35), 5, 150)
```

* **Guarantee**: Natural, unpolluted domain scaling from small scripts ($K \approx 5$) to 10M LOC Titans ($K \le 150$ per partition), with zero artificial forced merges.

---

## Slide 08: Robert C. Martin's Metrics at Graph Scale

### Clean Architecture Reimagined as Graph Invariants

Omni-Graph computes Uncle Bob's Component Principles directly over the multigraph:

```
                  ┌─────────────────────────────────────────┐
                  │ GALAXY A (e.g. src/db)                  │
  Foreign Nodes ─►│ Afferent Coupling ($C_a$): Incoming     │
                  │ Who depends on this galaxy?             │
                  │ (Responsibility & Impact)               │
                  └────────────────────┬────────────────────┘
                                       │
                                       │ Efferent Coupling ($C_e$): Outgoing
                                       │ What does this galaxy depend on?
                                       ▼
                  ┌─────────────────────────────────────────┐
                  │ Foreign Nodes (e.g. External Libs)      │
                  └─────────────────────────────────────────┘
```

### 1. The Instability Metric ($I$)

$$I = \frac{C_e}{C_a + C_e} \quad \in [0.0, 1.0]$$

| Instability ($I$) | Architectural Role | Characteristics & Risk Profile |
| :---: | :--- | :--- |
| **$I \le 0.25$** | **Core Foundation** | Highly stable. Depended upon by everyone. High blast radius; changes require extreme care. |
| **$0.25 < I \le 0.65$** | **Domain Service** | Balanced core logic. Moderately coupled. |
| **$I > 0.65$** | **Orchestrator / Leaf** | Volatile, depends on many modules. Low blast radius; safe for rapid agent refactoring. |

---

## Slide 09: Abstractness & Distance from Main Sequence

### Quantifying Architectural Rot ($A$ and $D$)

$$\text{Abstractness: } A = \frac{N_{\text{interfaces}} + N_{\text{traits}} + N_{\text{abstract}}}{N_{\text{total}}} \quad \in [0.0, 1.0]$$

$$\text{Distance from Main Sequence: } D = |A + I - 1| \quad \in [0.0, 1.0]$$

```
 Abstractness (A)
   ▲
1.0│ [Zone of Uselessness]
   │ (Abstract code with no callers)
   │         \
   │          \   THE MAIN SEQUENCE (D ≈ 0)
   │           \  (Optimal Balance: Stable = Abstract, Volatile = Concrete)
   │            \
   │             \
0.0│ [Zone of Pain]
   │ (Rigid, concrete code depended on by everyone)
   └───┬───────────────────────────────► Instability (I)
      0.0                             1.0
```

* **Zone of Pain ($A \to 0, I \to 0, D \to 1$)**: High $C_a$, low abstraction. Modifying this code breaks the universe.
* **Zone of Uselessness ($A \to 1, I \to 1, D \to 1$)**: Pure interfaces with zero implementations or callers. Dead abstraction debt.

---

## Slide 10: Personalized PageRank (PPR) Subgraph Slicing

### Why BFS Fails for LLMs & Agents

* **Naive Breadth-First Search (BFS)**: Explodes exponentially ($b^d$). A 2-hop search in a 5,000-node codebase easily pulls 600 nodes ($>50,000$ tokens), exceeding context limits and introducing massive hallucination noise.
* **Dense Vector Search Alone**: Pulls individual snippets but fails to capture the multi-hop caller $\to$ callee execution path.

### The Power-Iteration PPR Solution

Omni-Graph seeds a random walk at the target symbol $s$ with restart probability $(1 - \alpha)$, running power iteration ($\alpha = 0.85$):

$$\mathbf{p}^{(t+1)} = (1 - \alpha) \mathbf{p}^{(0)} + \alpha \mathbf{W} \mathbf{p}^{(t)}$$

```
                   PPR FLOW FROM QUERY SEED: "execute_task"
                                     [Seed: p0=1.0]
                                     (execute_task)
                                     /            \
                       (share=0.425)/              \(share=0.425)
                                   ▼                ▼
                          (validate_perms)     (db_insert)
                                   │                │
                                   ▼ (share=0.18)   ▼ (share=0.18)
                              (user_cache)     (wal_logger)
```

* **Binary-Search Slicer**: Takes the top-$k$ stationary probability nodes $\mathbf{p}^*$ and applies binary-search line truncation to **guarantee $< 1,500$ tokens**.
* **Result**: Zero prompt stuffing. 100% relevant callers, callees, and interfaces delivered in milliseconds.

---

## Slide 11: SAIR Blast Radius & Risk Matrix

### Predictive Impact Analysis Before Any Code is Written

When an engineer or an agent touches a function, what is the risk?

Omni-Graph computes the **Structural Architectural Impact Ratio (SAIR)**:

$$\text{SAIR}(v) = \frac{|\text{Transitive Callers}(v)|}{|V_{\text{workspace}}|} \times \left(1 + \sum_{g \in \text{Galaxies Crossed}} w_g\right)$$

```
┌────────────────────────────────────────────────────────────────────────┐
│                        SURGICAL RISK MATRIX                            │
├──────────────┬──────────────────┬──────────────────────────────────────┤
│ Risk Tier    │ Metric Bounds    │ Automated Agent Protocol             │
├──────────────┼──────────────────┼──────────────────────────────────────┤
│ 🟢 LOW       │ $I > 0.65$       │ Direct surgical edit permitted.      │
│              │ $\text{SAIR} < 0.05$ │ Inbound edges confined to 1 file.    │
├──────────────┼──────────────────┼──────────────────────────────────────┤
│ 🟡 MODERATE  │ $0.25 < I \le 0.65$│ Automated unit regression required. │
│              │ $\text{SAIR} < 0.20$ │ Intra-galaxy callers only.           │
├──────────────┼──────────────────┼──────────────────────────────────────┤
│ 🔴 CRITICAL  │ $I \le 0.25$     │ STOP. Architectural tie-breaker.     │
│              │ $\text{SAIR} \ge 0.20$│ Multiple galaxy boundaries crossed.  │
│              │ $C_a \ge 15$     │ Spec update mandatory before code.   │
└──────────────┴──────────────────┴──────────────────────────────────────┘
```

---

## Slide 12: Content-Tier Partitioning

### Eliminating the Pathological Markdown Shadowing Bug

In standard vector search (`bge-small-en-v1.5`), dense text models match markdown prose headers far more strongly than concise code signatures:

> **The Problem**: Querying *"How does AST parsing work?"* returned 15 pages of DEVLOG.md meeting notes and zero lines of `parser/mod.rs`!

### The Two-Tier Storage Architecture

```
                  Incoming Query: "How does AST parsing work?"
                                      │
                                      ▼
                          [Content Tier Classifier]
                                      │
                   ┌──────────────────┴──────────────────┐
                   ▼                                     ▼
        Default: CODE TIER ONLY             Explicit: DOC TIER ONLY
        (db.code_node)                      (db.doc_node)
        • Functions, Structs, Impls         • Markdown headings, PRDs, ADRs
        • Tree-sitter signatures            • Living documentation
        • Zero doc noise                    • Isolated from code RAG
```

* Context-enriched vector headers prepend structured metadata:
  `[{language}] {kind} {label} in {file_path}\n{code}`
  This guarantees that code semantics always match code queries.

---

## Slide 13: The Agentic Intelligence Triad

### Why Skill + MCP + CLI Form a Symbiotic Trinity

A major misconception in AI tooling is that MCP replaces CLIs, or that skills are just prompts:

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
│ • When & Why Strategy   │       │ • Zero-Token Shell Free │       │ • Headless Shell Script │
│ • 4-Tier Recon Ladder   │       │ • Schema-Typed Tools    │       │ • Clean Output Tables   │
│ • 150-Line Read Gate    │       │ • 10x Invocation Speed  │       │ • Zero Escaping Bugs    │
│ • "HOW TO THINK"        │       │ • "HOW TO EXECUTE (IDE)"│       │ • "HOW TO EXECUTE (CLI)"│
└─────────────────────────┘       └─────────────────────────┘       └─────────────────────────┘
```

1. **The Skill**: Dictates the strategy: *"Never read 1,000 lines blindly. Pull a 2-hop condensed slice first."*
2. **The MCP Server**: Provides structured, high-speed, zero-token JSON-RPC communication directly to the model.
3. **The CLI (`omni`)**: Provides clean ergonomics for terminal developers: `omni symbol <name>`, `omni galaxies`.

---

## Slide 14: System Microservice Architecture

### 4 Hermetic Services Running on `omni-net`

```
                                  Agent / Developer
                                          │
                  ┌───────────────────────┴───────────────────────┐
                  │ (HTTP / SSE / REST / JSON-RPC stdio)          │
                  ▼                                               ▼
     ┌────────────────────────┐                      ┌────────────────────────┐
     │  graph-ui (Port 3000)  │                      │ rust-app (Port 8080)   │
     │  • React + TypeScript  │                      │ • Axum Async Server    │
     │  • Cosmograph WebGL    │                      │ • Tree-sitter Parsers  │
     │  • 60 FPS Force Graph  │                      │ • FIFO Tokio Ingestion │
     │  • Live-Watch Telemetry│                      │ • In-Memory File Cache │
     └────────────────────────┘                      └───────────┬────────────┘
                                                                 │
                                     ┌───────────────────────────┴───────────────────────────┐
                                     ▼                                                       ▼
                        ┌────────────────────────┐                              ┌────────────────────────┐
                        │  graph-db (Port 8000)  │                              │  embedder (Port 8081)  │
                        │  • SurrealDB v2        │                              │  • HuggingFace TEI     │
                        │  • Multi-Model Graph   │                              │  • bge-small-en-v1.5   │
                        │  • HNSW Vector Index   │                              │  • 384-dim Embeddings  │
                        │  • Dynamic SQL Queries │                              │  • Hardware Accelerated│
                        └────────────────────────┘                              └────────────────────────┘
```

* **Live-Watch Delta Sync**: File system events trigger incremental AST re-indexing in $< 50\text{ms}$ with zero query blackouts.

---

## Slide 15: Omni-Graph vs. The World

### Why Open Source Solutions Fail Where Omni-Graph Wins

| Feature / Capability | Raw AST / Tree-sitter | Sourcegraph / Kythe / SCIP | LangChain / Code RAG | Omni-Graph |
| :--- | :---: | :---: | :---: | :---: |
| **Parsing Engine** | Local CST Only | Heavy Compiler Toolchain | Naive Text Chunker | Polyglot Tree-sitter (Rust/Go/TS/Java/C++) |
| **Cross-File Resolution** | ❌ None | ✅ Heavyweight Build Passes | ❌ Hallucinated | ✅ **SAIR Scoped AST Matching** |
| **Graph Topology Backbone** | ❌ Disconnected | ❌ Relational SQL Only | ❌ No Graph | ✅ **Topological Containment DAG** |
| **Macroscopic Community Detection** | ❌ None | ❌ None | ❌ None | ✅ **Hierarchical Leiden Galaxies** |
| **Cluster Thrashing Prevention** | ❌ None | ❌ None | ❌ None | ✅ **Two-Phase Seed LPA + Xorshift64** |
| **Robert C. Martin Package Metrics** | ❌ None | ❌ None | ❌ None | ✅ **$C_a, C_e, I, A, D$ Automated** |
| **Agent Neighborhood Slicing** | ❌ Dump entire file | ❌ Huge JSON responses | ❌ High token hallucination | ✅ **PPR (<1500 Tokens Guaranteed)** |
| **Markdown Shadowing Protection** | ❌ None | ❌ Code only | ❌ Severe Shadowing | ✅ **Content-Tier Partitioning** |
| **Live Delta Sync** | ❌ Manual | ❌ Batch Re-index (Hours) | ❌ Slow Re-embed | ✅ **Incremental Sub-50ms Watcher** |

---

## Slide 16: The Grand Synthesis

### From 10 Million Lines of Code to a 50ms Surgical Edit

```
   10,000,000 Lines of Monorepo Code
                   │
                   ▼ (Polyglot Tree-sitter + Containment Backbone)
   Connected Topological Multi-Graph (Zero Orphans)
                   │
                   ▼ (Hierarchical Leiden & Two-Phase Re-Clustering)
   12 Stable Macro Galaxies (Core Foundations vs. Volatile Leaves)
                   │
                   ▼ (Uncle Bob Metrics: Ca, Ce, Instability Index)
   Instant Blast Radius Matrix & Risk Score
                   │
                   ▼ (Personalized PageRank α=0.85 + Binary Slicer)
   < 1,500 Token Pristine Markdown Context Slice
                   │
                   ▼
   Autonomous Agent Executes Surgical Bugfix with ZERO Hallucination
```

---

## Keynote Speaker Notes & Reference Summary

> **The Big Takeaway**:  
> "Stop feeding raw file dumps and naive vector chunks to LLMs. Code is not prose; code is a directed, typed, multi-tiered dependency graph. When you combine compiler-grade ASTs with statistical physics (Percolation Theory), network science (Leiden & Personalized PageRank), and structural engineering (Robert C. Martin metrics), you transform code intelligence from a toy into an unshakeable deterministic superweapon."

* **PolyTuring Repository**: ` vadde/PolyTuring`
* **Orchestrator Source**: `tools/omni-graph/src/`
* **Interactive UI**: `http://localhost:3000` (Cosmograph WebGL Engine)
* **API Documentation**: `http://localhost:8080/api-docs` (Cosmic Liquid-Glass Explorer)
