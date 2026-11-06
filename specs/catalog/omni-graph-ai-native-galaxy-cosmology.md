# 🌌 Omni-Graph Spec: AI-Native Galaxy Cosmology & Natural Modularity Architecture

> **Document ID**: `SPEC-OMNI-004`  
> **Status**: `spec-review`  
> **Target Version**: `v0.6.0`  
> **Author**: Antigravity (Google DeepMind) & `@vadde`  
> **Date**: 2026-11-05  
> **Precedence**: Extends [`specs/catalog/omni-graph.md`](omni-graph.md) and [`specs/catalog/omni-graph-two-pass-engine-hardening.md`](omni-graph-two-pass-engine-hardening.md).

---

## 1. Executive Summary & Epistemic Motivation

### 1.1 The Epistemic Flaw of Miller's Law in AI Systems
George A. Miller’s seminal 1956 paper, *"The Magical Number Seven, Plus or Minus Two"*, established that the biological human brain's working memory buffer can hold roughly $7 \pm 2$ items (5 to 9 chunks) simultaneously. While this physiological bound was appropriate for early graphical user interfaces, **its dogmatic enforcement as an architectural ceiling ($K \le 15$) in an AI-native code graph engine is fundamentally broken**.

Modern sovereign AI coding agents possess context windows spanning from 1,000,000 to 2,000,000+ tokens. They execute multi-hop graph traversals, evaluate call-tree reachability across tens of thousands of AST symbols, and require **uncompromising topological ground truth**.

When Omni-Graph forcibly coarsened repositories down to $K \le 15$—and utilized a toxic fallback (`merge with largest community`) when clusters lacked relational affinity:
1. **Domain Purity Collapsed**: In `tutor-intelligence`, Galaxy #1 ballooned to **1,434 nodes**, merging solver kinematics, documentation files, and test fixtures into a single monstrous mega-blob.
2. **Blast-Radius Calculations Were Invalidated**: Autonomous agents evaluating the blast radius of modifying a test file were falsely informed that pathfinding and kinematics shared the same tightly coupled architectural domain.
3. **Enterprise Monorepos (10M LOC) Became Impossible**: In a 10-million-line codebase with 100+ bona fide microservices and bounded contexts, clamping to 15 produces absurd 60,000-node landfills, defeating the very purpose of an architectural knowledge hub.

### 1.2 The Sovereign AI Core Principle
> **Constitutional Invariant**: Omni-Graph MUST reflect the **physical, topological ground truth** of the codebase. It shall never fabricate false couplings, merge decoupled domains, or truncate architectural entities solely to satisfy human biological working memory constraints.

---

## 2. Requirements & Traceability Matrix

### 2.1 Functional Requirements

| ID | Title | Description | Target |
| :--- | :--- | :--- | :--- |
| **R-086** | **Affinity-Floor Invariant (Zero Forced Merging)** | Two communities $C_A$ and $C_B$ may ONLY be merged if they possess positive topological edge coupling ($k_{i, j} > 0$) or shared directory lineage. If a community has zero coupling to other communities, it MUST remain an independent architectural domain. The fallback `sizes.last()` (dumping into largest cluster) is permanently purged. | Ingestion, Watcher, Analysis |
| **R-087** | **Modularity-Guarded Coarsening ($\Delta Q \ge -\epsilon$)** | Community coarsening evaluates Newman's modularity change $\Delta Q_{ij} = \frac{k_{ij}}{2m} - \frac{k_i k_j}{2m^2}$. Coarsening MUST terminate when further merges would degrade global graph modularity beyond tolerance ($\epsilon = 0.05$). | `analysis/mod.rs` |
| **R-088** | **Dynamic Scale-Adaptive Subsystem Sizing ($K_{\text{adaptive}}$)** | The target subsystem ceiling is computed dynamically as a sub-linear function of graph complexity: $K_{\text{target}}(N) = \text{clamp}\left(\lfloor 3.5 \cdot N^{0.35} \rfloor, \text{min}=5, \text{max}=150\right)$. Supports caller override via `max_galaxies`. | Ingestion, API |
| **R-089** | **Multi-Scale Hierarchical Zoom (Cosmos $\to$ Galaxy $\to$ Solar System)** | The engine persists dual resolution: macro architectural domains (`galaxy_id`) and fine-grained modular clusters (`community_id`). Exposes `mode=natural|macro|module` across endpoints. | DB, API, UI |
| **R-090** | **Clean Noise Floor Absorption (Designated Orphan Space)** | Isolated singletons ($< 3$ nodes) with zero edge affinity and zero directory lineage are absorbed into a designated `utilities / standalone` namespace, preventing pollution of core business domains. | `analysis/mod.rs` |
| **R-091** | **Uncle Bob Martin Package Metric Invariants** | Computes Afferent Coupling ($C_a$), Efferent Coupling ($C_e$), Instability ($I$), Abstractness ($A$), and Distance from Main Sequence ($D$) on true natural galaxies without artificial skew. | DB, API |
| **R-092** | **AI Agent Blast-Radius Ground Truth** | Eliminates false cluster-colocation in `/api/references` and `/api/condense`, ensuring blast radius reflects actual call and containment edges. | Condenser, API |

### 2.2 Acceptance Criteria (AC)

- **AC-030**: On `tutor-intelligence`, Galaxy #1 drops from 1,434 nodes to its true domain size; solver, pathfinding, specs, and docs emerge as distinct, unpolluted galaxies.
- **AC-031**: On `DesignPatterns`, 23 GoF patterns and support packages are partitioned into natural bounded contexts rather than being forced into 15.
- **AC-032**: On `django-sorcery`, distinct packages (`db`, `viewsets`, `views`, `alembic`, `forms`, `validators`) emerge as pure architectural domains.
- **AC-033**: Modularity $Q$ is strictly monotonically guarded; zero decoupled communities are force-merged.
- **AC-034**: `/api/galaxies` supports `mode=natural|macro|module` and optional `max_galaxies` override.
- **AC-035**: 100% of existing unit tests pass, and new cosmological unit tests verify dynamic scaling and affinity floor enforcement.
- **AC-036**: Zero clippy warnings, living documentation (`STATUS.md`, `CONTEXT.md`, `DEVLOG.md`, `CHANGELOG.md`) kept in lockstep.

### 2.3 Non-Functional Requirements (NF)

- **NF-016 (Computational Bound)**: Coarsening on a 10,000-node graph must execute in under 150ms in Rust RAM.
- **NF-017 (Memory Ceiling)**: Dual-resolution storage must add less than 8 bytes of RAM overhead per node.
- **NF-018 (Scale Invariance)**: Algorithm must function with zero crashes and zero infinite loops from $N = 1$ to $N = 1,000,000$.

---

## 3. Detailed Architectural Design

### 3.1 Mathematical Model of Modularity-Guarded Coarsening

Let $G = (V, E, W)$ be the weighted code entity graph with total edge weight $m = \sum_{e \in E} w(e)$.
For a candidate community $i$ and neighbor $j$:

$$k_{i, j} = \sum_{u \in C_i, v \in C_j} w(u, v)$$
$$k_i = \sum_{u \in C_i, v \in V} w(u, v)$$

The modularity differential of merging $C_i$ into $C_j$ is:
$$\Delta Q_{ij} = \frac{k_{ij}}{m} - \frac{k_i \cdot k_j}{2m^2}$$

```rust
// The Affinity-Floor Invariant:
if max_weight <= 0.0 && dir_affinity.is_none() {
    // Zero coupling to any other community.
    // MUST NOT MERGE with largest community.
    // Preserve as independent architectural domain.
    break;
}
```

### 3.2 The Scale-Adaptive Galaxy Sizing Formula

$$K_{\text{adaptive}}(N) = \text{clamp}\left(\left\lfloor 3.5 \cdot N^{0.35} \right\rfloor, \; 5, \; 150\right)$$

```
N = 37       (vadde.github.io):  3 - 5  galaxies
N = 1,871    (DesignPatterns):   15 - 35 galaxies (natural GoF domain breakdown)
N = 2,352    (django-sorcery):   15 - 30 galaxies (pure package boundaries)
N = 4,762    (tutor-intelligence): 25 - 50 galaxies (modular solver & simulator domains)
N = 50,000   (Enterprise Monorepo): 50 - 100 galaxies (microservice boundaries)
N = 1,000,000 (Linux Kernel Titan): 100 - 150 macro-subsystems
```

---

## 4. Migration & Compatibility

1. **Database Schema**: No breaking change. Table `galaxy` continues to store macro-architectural records; `DbNode.community` stores the assigned galaxy ID.
2. **REST API**:
   - `GET /api/galaxies?workspace=<ws>&mode=natural|macro&max_galaxies=<n>`: Fully backward compatible. Defaults to natural adaptive clustering.
   - `POST /api/cluster`: Accepts optional `max_galaxies` in JSON body.
3. **UI Force Graph**: The WebGL Cosmograph visualizer automatically renders the natural galaxy count with high-contrast color palette generation.
