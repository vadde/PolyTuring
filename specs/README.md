# 📋 Spec-Driven Development (SDD)

> **No code without a spec. No spec without a purpose.**

This directory contains all specifications for tools in this repository.
Specs are the single source of truth for what a tool should do.

---

## Directory Structure

```
specs/
├── README.md           ← You are here
├── _templates/         ← Spec templates (copy, don't modify)
│   ├── tool-spec.md    ← Template for new tool specs
│   └── enhancement-spec.md  ← Template for enhancements
└── catalog/            ← Active specs (one per tool)
    ├── tool-a.md
    ├── tool-b.md
    └── ...
```

## How to Use

### Creating a New Spec

1. Copy the appropriate template:
   ```bash
   cp specs/_templates/tool-spec.md specs/catalog/<tool-name>.md
   ```
2. Fill in ALL sections (see template for guidance)
3. Number all requirements: `R-001`, `R-002`, ...
4. Create the tool's `STATUS.md` with `status: draft`

### Spec Lifecycle

```
Draft → Review → Approved → Implemented → Verified
```

| Phase | Who | Action |
|-------|-----|--------|
| Draft | Author | Write the spec |
| Review | Peer/Agent | Review for completeness and clarity |
| Approved | Maintainer | Approve for implementation |
| Implemented | Developer/Agent | Code written against spec |
| Verified | Tester/CI | Tests prove spec is satisfied |

### Spec Requirements

All specs MUST have:
- [ ] Numbered requirements (R-001 format)
- [ ] Clear interface contract (inputs/outputs)
- [ ] Testable acceptance criteria
- [ ] Defined constraints (performance, security)
- [ ] Dependency list

---

## Active Specifications Catalog

| Spec File | Tool / Domain | Target Version | Status | Governing ADR |
|:---|:---|:---|:---|:---|
| [`catalog/omni-graph.md`](catalog/omni-graph.md) | Omni-Graph | `v0.1.0` | `implemented` | — |
| [`catalog/omni-graph-nextgen.md`](catalog/omni-graph-nextgen.md) | Omni-Graph | `v0.3.0` | `implemented` | — |
| [`catalog/omni-graph-two-pass-engine-hardening.md`](catalog/omni-graph-two-pass-engine-hardening.md) | Omni-Graph | `v0.5.0` | `implemented` | — |
| [`catalog/omni-graph-ai-native-galaxy-cosmology.md`](catalog/omni-graph-ai-native-galaxy-cosmology.md) | Omni-Graph | `v0.6.0` | `implemented` | — |
| [`catalog/omni-graph-live-watcher-fidelity-parity.md`](catalog/omni-graph-live-watcher-fidelity-parity.md) | Omni-Graph | `v0.7.0` | `implemented` | [`ADR-002`](../docs/architecture/adr-002-live-watcher-fidelity-concurrency-overhaul.md), [`ADR-003`](../docs/architecture/adr-003-matlab-simulink-mbse-graph-ingestion.md) |
| [`catalog/omni-graph-ingestion-openapi-ux.md`](catalog/omni-graph-ingestion-openapi-ux.md) | Omni-Graph | `v0.4.0` | `implemented` | — |
| [`catalog/session-explorer.md`](catalog/session-explorer.md) | Session Explorer | `v0.1.0` | `implemented` | — |

---

## Quick Links

- [Tool Spec Template](_templates/tool-spec.md)
- [Enhancement Spec Template](_templates/enhancement-spec.md)
- [SDD Workflow Rules](../.agents/rules/04-sdd-workflow.md)
- [Tool Catalog](../tools/README.md)
