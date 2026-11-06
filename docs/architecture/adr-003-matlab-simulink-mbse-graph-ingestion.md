# ADR-003: MathWorks MATLAB, Simulink MBSE, and Symlink-Cycle-Guarded Graph Ingestion

## Status
**Proposed / Accepted (Research Spike Complete)** — Ready for phased implementation in v0.8.0.

## Context
Omni-Graph currently provides semantic AST and call-graph indexing across 8 core languages (Rust, Python, JavaScript, TypeScript, Go, Java, C, C++), along with structural extractors for Shell, SQL, YAML, TOML, JSON, and Markdown.

Enterprise engineering and aerospace/defense systems heavily rely on **Model-Based Systems Engineering (MBSE)** using the MathWorks toolchain:
1. **MathWorks MATLAB (`.m`)**: Proprietary functional, procedural, and object-oriented (`classdef`) programming language with package namespaces (`+pkg`) and class method folders (`@class`).
2. **MathWorks Simulink (`.slx` / `.mdl`)**: Graphical dataflow block diagram environment for simulation and automated embedded code generation.
3. **Filesystem Symlinks**: Monorepos frequently share code, configuration, or model libraries via symlinks, posing circular recursion hazards if not cycle-guarded.

The user explicitly requested:
> *"what aer the langs we supported so far in the omni graph kotlin java golang javscript typescript react blah blah ? can we see if we can add matlab, symlink stuff as well as a req 107 and so forth. we an doit at tail end once we solidify the fundamental system is fail proof... browser internet for matlab and symlink since it's very proprietary mathworks stuff. just capture and put it out there with clear notes that this needs thorough analysis and research when taken up."*

This Architecture Decision Record documents the epistemic research spike, technical feasibility, file container formats, schema mappings, and execution strategy for Requirements `R-107`, `R-108`, and `R-109`.

---

## 1. Supported Languages Inventory

| Language | Extractor Engine | AST Grammar Crate | Symbol Coverage | Edge Types |
| :--- | :--- | :--- | :--- | :--- |
| **Rust** | Tree-sitter | `tree-sitter-rust` | fn, struct, enum, trait, impl, macro, field | `CALLS`, `DECLARES`, `IMPORTS`, `IMPLEMENTS` |
| **Python** | Tree-sitter | `tree-sitter-python` | def, async def, class, method | `CALLS`, `DECLARES`, `IMPORTS`, `EXTENDS` |
| **JavaScript** | Tree-sitter | `tree-sitter-javascript` | function, class, method, arrow fn | `CALLS`, `DECLARES`, `IMPORTS` |
| **TypeScript / TSX** | Tree-sitter | `tree-sitter-typescript` | function, class, interface, type, method | `CALLS`, `DECLARES`, `IMPORTS`, `IMPLEMENTS` |
| **Go** | Tree-sitter | `tree-sitter-go` | func, type struct, type interface, method | `CALLS`, `DECLARES`, `IMPORTS`, `IMPLEMENTS` |
| **Java** | Tree-sitter | `tree-sitter-java` | class, interface, method, enum | `CALLS`, `DECLARES`, `IMPORTS`, `EXTENDS`, `IMPLEMENTS` |
| **C** | Tree-sitter | `tree-sitter-c` | function, struct, typedef | `CALLS`, `DECLARES`, `IMPORTS` |
| **C++** | Tree-sitter | `tree-sitter-cpp` | class, struct, function, method | `CALLS`, `DECLARES`, `IMPORTS`, `EXTENDS` |
| **Shell / Bash** | Regex / Fallback | Internal Regex Chunker | functions, variables | `CALLS`, `DECLARES` |
| **SQL** | Regex / Fallback | Internal SQL Chunker | CREATE TABLE, VIEW, PROCEDURE | `DECLARES`, `REFERENCES` |
| **YAML / TOML / JSON**| Structural | Chunker | Top-level configurations, manifest keys | `CONTAINS` |
| **Markdown** | Markdown parser | `src/parser/markdown.rs` | Headings, sections, code snippets, links | `CONTAINS`, `REFERENCES` |

---

## 2. Research Spike: MathWorks MATLAB (`.m`) — R-107

### 2.1 Grammar Engine: `tree-sitter-matlab`
- **Crate**: `tree-sitter-matlab` (available on crates.io, derived from Álan Crístoffer's grammar).
- **Feasibility**: High. Exposes the standard Tree-sitter C ABI compatible with `tree-sitter = "0.22"`.

### 2.2 AST Node Mapping
MATLAB code is organized into three distinct file structures:
1. **Function M-Files**:
   ```matlab
   function [out1, out2] = calculateTrajectory(speed, angle, drag)
       out1 = speed * cos(angle);
       out2 = speed * sin(angle) - 0.5 * 9.81 * drag;
   end
   ```
   - **AST Grammar Rule**: `function_definition` -> `function_output`, `name`, `parameters`, `body`.
   - **Graph Node**: `kind = "function"`, `label = "calculateTrajectory"`.
   - **Edges**: In-body invocations emit `CALLS` edges.

2. **OOP Classdef Files**:
   ```matlab
   classdef KalmanFilter < BaseFilter
       properties
           CovarianceMatrix
           StateEstimate
       end
       methods
           function obj = KalmanFilter(initialCov)
               obj.CovarianceMatrix = initialCov;
           end
           function update(obj, measurement)
               % ...
           end
       end
   end
   ```
   - **AST Grammar Rule**: `class_definition` -> `superclasses`, `properties`, `methods`.
   - **Graph Nodes**:
     - `kind = "class"`, `label = "KalmanFilter"`
     - `kind = "property"`, `label = "CovarianceMatrix"`
     - `kind = "method"`, `label = "update"`
   - **Edges**:
     - `KalmanFilter -> EXTENDS -> BaseFilter`
     - `KalmanFilter -> CONTAINS -> CovarianceMatrix`
     - `KalmanFilter -> DECLARES -> update`

3. **Namespaces & Packages**:
   - MathWorks packages use directory prefix `+` (e.g., `+control/+guidance/NavSystem.m`).
   - The fully qualified symbol is `control.guidance.NavSystem`.
   - Method folders use prefix `@` (e.g., `@FlightController/stabilize.m`).
   - **Resolution Invariant**: Ingestion must synthesize package hierarchy: `package:control -> CONTAINS -> package:control.guidance -> CONTAINS -> class:NavSystem`.

---

## 3. Research Spike: MathWorks Simulink (`.slx` / `.mdl`) — R-108

### 3.1 Container Architecture: Open Packaging Conventions (OPC)
Unlike legacy `.mdl` files (which are proprietary ASCII text trees), modern `.slx` files are **standard ZIP archives conforming to ISO/IEC 29500-2 (Open Packaging Conventions)**.

When unpacked, an `.slx` container consists of:
```
model.slx (ZIP archive)
├── [Content_Types].xml          # OPC package types
├── _rels/.rels                  # Global relationships
├── metadata/
│   └── mwcoreProperties.xml     # Author, version, creation timestamp
└── simulink/
    ├── blockdiagram.xml         # ⭐ Primary computational block diagram DAG
    ├── stateflow.xml            # Statecharts and finite state machines
    ├── graphicalInterface.xml   # Inports, Outports, Root bus interfaces
    └── _rels/
        └── blockdiagram.xml.rels # External model references and libraries
```

### 3.2 Computational Graph Extraction from `simulink/blockdiagram.xml`
MathWorks represents the visual dataflow topology using structured XML:
```xml
<System RefName="FlightControlSystem">
  <!-- Blocks (Nodes) -->
  <Block RefClass="SubSystem" Name="SensorFusion" SID="1">
    <P Name="Position">[100, 50, 200, 150]</P>
  </Block>

  <Block RefClass="ModelReference" Name="ActuatorDynamics" SID="2">
    <P Name="ModelName">ActuatorModel</P>
  </Block>

  <!-- Signal Lines (Edges) -->
  <Line>
    <P Name="Source">1#out:1</P>
    <P Name="Destination">2#in:1</P>
  </Line>
</System>
```

### 3.3 Topological Schema Mapping for Simulink MBSE

| Simulink XML Element | Omni-Graph Node Kind | Node Identifier | Inferred Properties |
| :--- | :--- | :--- | :--- |
| `<Block RefClass="SubSystem">` | `subsystem` | `{ws}:{file_path}:{SID}:{Name}` | Subsystem hierarchy, contained blocks |
| `<Block RefClass="ModelReference">`| `model_reference` | `{ws}:{file_path}:{SID}:{Name}` | Target referenced `.slx` file name |
| `<Block RefClass="Chart">` | `stateflow_chart` | `{ws}:{file_path}:{SID}:{Name}` | State machine logic |
| `<Block RefClass="Inport/Outport">`| `port` | `{ws}:{file_path}:{SID}:{Name}` | Interface bus definition |

| Simulink Connectivity | Omni-Graph Edge Type | Edge Category | Rationale |
| :--- | :--- | :--- | :--- |
| `<Line>` (Port to Port) | `SIGNAL_TO` | `DATAFLOW` | Direct data propagation along signal wires |
| Subsystem containment | `CONTAINS` | `CONTAINMENT` | Hierarchical block decomposition |
| `<P Name="ModelName">X</P>` | `REFERENCES_MODEL` | `EXTRACTED` | Inter-model dependencies across separate `.slx` files |

### 3.4 Ingestion Pipeline Strategy for `.slx`
1. Read `.slx` file into memory (or streaming unzipper `zip = "2.2"`).
2. Extract `simulink/blockdiagram.xml` and `simulink/stateflow.xml`.
3. Use fast streaming XML parser (`quick-xml = "0.36"`) to extract `<Block>` and `<Line>` elements without holding full DOM trees in RAM.
4. Pass 1: Register all blocks and subsystems in `SymbolTable`.
5. Pass 2: Connect `SIGNAL_TO` and `REFERENCES_MODEL` cross-model edges.

---

## 4. Filesystem Symlink Cycle Guarding — R-109

### 4.1 Hazard Analysis
When traversing deep directories in monorepos or enterprise projects with shared submodules (e.g. `ln -s ../../shared/ ./shared_lib`), naive recursive traversals encounter:
- Infinite recursion loops causing memory exhaustion and stack overflow.
- Duplicate node generation where symlinked files are indexed multiple times under different aliases.

### 4.2 Architectural Decision: Inode & Canonical Path Guard
- Use `std::fs::symlink_metadata` during directory discovery.
- Maintain an active traversal set: `visited_canonical_paths: HashSet<PathBuf>`.
- For Unix systems, record `(dev_t, ino_t)` via `std::os::unix::fs::MetadataExt`.
- If a symlink resolves to an already-visited canonical path or ancestor:
  - Halt traversal on that branch.
  - Emit a single graph edge: `node:symlink_path -> LINKS_TO -> node:canonical_target_path`.
  - Zero duplicate symbol declarations. Zero infinite loops.

---

## 5. Execution Roadmap

- [x] **Phase 1 (v0.7.0)**: Research spike completed, formats verified, XML schemas mapped, ADR-003 approved.
- [ ] **Phase 2 (v0.8.0 - MATLAB Parser)**: Add `tree-sitter-matlab` dependency, author `src/parser/matlab.rs`, support function/classdef/package symbols, and add unit test suite.
- [ ] **Phase 3 (v0.8.0 - Symlink Guard)**: Integrate cycle-guarded crawler in `src/ingestion/mod.rs` and `src/watcher/mod.rs`.
- [ ] **Phase 4 (v0.8.1 - Simulink OPC Ingestion)**: Add `zip` and `quick-xml` dependencies, author `src/parser/simulink.rs`, extract blocks/signals from `.slx`, and support cross-model Graph-RAG retrieval.

---

## References
- MathWorks Simulink Open Packaging Conventions (OPC) Specification (ISO/IEC 29500-2).
- Álan Crístoffer, `tree-sitter-matlab` Grammar Specification.
- [`specs/catalog/omni-graph-live-watcher-fidelity-parity.md`](../../specs/catalog/omni-graph-live-watcher-fidelity-parity.md)
- [`docs/architecture/adr-002-live-watcher-fidelity-concurrency-overhaul.md`](adr-002-live-watcher-fidelity-concurrency-overhaul.md)
