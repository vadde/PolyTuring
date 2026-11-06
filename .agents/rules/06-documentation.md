# Rule 06 — Documentation Standards

> **Priority**: 🔴 Critical — If it's not documented, it doesn't exist. If public docs are cluttered with internal sprint tickets, you fail.

---

## 🏛️ The Public Showcase vs. Internal Governance Law (The Two-World Invariant)

> **Constitutional Rule**: There is a strict, non-negotiable boundary between **Public Product Documentation (The Storefront)** and **Internal Engineering Governance (The Factory Floor)**. Never confuse the two.

### World 1: Public-Facing Product Documentation (The Storefront)
* **Scope**: `docs/site/index.html` (Living Portal / GitHub Pages), `vadde.github.io`, marketing presentations, and external landing pages.
* **Target Audience**: Users, Buyers, AI Engineers, OSS Contributors, and Tech Leads evaluating or adopting the product.
* **Core Objective**: Sell the capability, explain what it solves, demonstrate why existing tools (Tree-Sitter alone, raw grep, naive graph tools like Graphify) fail, and show how to run it in 60 seconds.
* **Tone & Vocabulary**:
  - Crisp, authoritative, crystal-clear, layperson-accessible language.
  - Tangible capability cards, problem-solution matrices, and intuitive mental models (e.g. "Airplane airlock engine swap", "browser tab worktrees").
* **⛔ ABSOLUTELY FORBIDDEN IN PUBLIC DOCS**:
  1. **Zero Internal Ticket IDs**: NEVER mention `R-101`, `R-093`, `R-H083`, or sprint requirement codes. A customer buying an engine does not care about your sprint board ticket numbers.
  2. **Zero Internal ADR References**: NEVER litter public pages with `ADR-002`, `ADR-003`, or internal architectural decision codes. ADRs belong in internal architecture archives.
  3. **Zero Internal Defect / Gap Numbers**: NEVER dump internal audit notes like `Gap 1`, `Gap 2`, `Loophole 14` into customer-facing text.
  4. **Zero Version Noise**: NEVER confuse users with competing version badges (e.g. `v0.4.0 Living Docs` vs `v0.7.0 Omni-Graph`). Public documentation always presents the **current live system** authoritatively without historical version chatter.
  5. **Zero Gratuitous Math Vanity**: Avoid dumping raw formulas, Greek matrices, or academic complexity equations into public marketing sections unless accompanied by an immediate, intuitive plain-English explanation.

### World 2: Internal Engineering Governance (The Factory Floor)
* **Scope**: `specs/catalog/`, `docs/architecture/` (ADRs), `tools/<name>/STATUS.md`, `CONTEXT.md`, `DEVLOG.md`, `tests/`.
* **Target Audience**: Core Maintainers, System Architects, Pair-Programming AI Agents.
* **Core Objective**: Complete traceability, formal contracts, mathematical correctness, blast-radius boundaries, and unyielding SDD discipline.
* **Mandated Contents**:
  - Exact numbered requirements (`R-001` through `R-109`).
  - Architecture Decision Records (`ADR-001`, `ADR-002`, `ADR-003`).
  - Formal mathematical formulations, graph complexity proofs ($O(|E|)$), and algorithmic steps.
  - Traceability matrices, test assertions, and append-only session devlogs.

---

## Documentation Hierarchy

```
Level 1: README.md (per tool)     — "What is this and how do I use it?"
Level 2: spec.md (per tool)       — "What exactly does this do?"
Level 3: Code comments/docstrings — "How does this work internally?"
Level 4: examples/ (per tool)     — "Show me it working"
Level 5: docs/book/ (repo-wide)   — "Teach me everything"
Level 6: docs/architecture/ (ADRs)— "Why did we decide this?"
```

---

## Tool README Template

Every tool's `README.md` MUST include these sections:

```markdown
# Tool Name

> One-line description of what this tool does.

[![Status](badge)](#) [![Version](badge)](#) [![Language](badge)](#)

## Overview

2-3 sentence description of the problem this tool solves and how it solves it.

## Installation

Step-by-step installation instructions. Include ALL dependencies.

## Usage

### CLI

\`\`\`bash
tool-name [OPTIONS] <input>
\`\`\`

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-o, --output` | Output format | `text` |
| `-v, --verbose` | Enable verbose logging | `false` |

### As a Library

\`\`\`python
from tool_name import process
result = process(input_data)
\`\`\`

## Examples

Link to and describe each example in `examples/`.

## Configuration

Document any configuration files or environment variables.

## Specification

See [spec.md](spec.md) for the full technical specification.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history.

## License

MIT — See [LICENSE](../../LICENSE) for details.
```

---

## Changelog Format (Keep-a-Changelog)

```markdown
# Changelog

All notable changes to this tool will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- New features

### Changed
- Changes in existing functionality

### Deprecated
- Features that will be removed

### Removed
- Features that were removed

### Fixed
- Bug fixes

### Security
- Vulnerability fixes

## [0.1.0] - 2026-XX-XX

### Added
- Initial implementation
```

---

## Code Documentation Standards

### Python — Docstrings (Google Style)

```python
def process(data: list[dict], config: Config | None = None) -> Result:
    """Process input data according to the specified configuration.

    This function applies transformation rules from the config to each
    item in the data list, producing an aggregated result.

    Implements: R-003 (Data processing pipeline)

    Args:
        data: List of dictionaries containing input records.
            Each dict must have 'id' and 'value' keys.
        config: Optional processing configuration. If None,
            uses default settings.

    Returns:
        A Result object containing processed records and metadata.

    Raises:
        ValueError: If data is empty or contains invalid records.
        ConfigError: If config has conflicting settings.

    Example:
        >>> result = process([{"id": 1, "value": 42}])
        >>> result.count
        1
    """
```

### Go — GoDoc Comments

```go
// Process applies transformation rules to the input data.
//
// It iterates through each record, applying the configuration rules,
// and returns an aggregated Result. If config is nil, default settings
// are used.
//
// Implements: R-003 (Data processing pipeline)
func Process(data []Record, config *Config) (*Result, error) {
```

### TypeScript — TSDoc Comments

```typescript
/**
 * Process input data according to the specified configuration.
 *
 * Applies transformation rules from the config to each item in the
 * data array, producing an aggregated result.
 *
 * @implements R-003 (Data processing pipeline)
 *
 * @param data - Array of input records to process.
 * @param config - Optional processing configuration.
 * @returns Processed result with metadata.
 * @throws {ValidationError} If data is empty or invalid.
 *
 * @example
 * ```ts
 * const result = process([{ id: 1, value: 42 }]);
 * console.log(result.count); // 1
 * ```
 */
```

---

## Architecture Decision Records (ADRs)

### When to Write an ADR

Write an ADR when:
- Choosing a technology, library, or framework
- Establishing a pattern or convention
- Making a trade-off that affects multiple tools
- Changing an existing architectural decision

### ADR Template

See `docs/architecture/adr-template.md` for the full template.

```markdown
# ADR-XXX: Title

## Status
Proposed | Accepted | Deprecated | Superseded by ADR-YYY

## Context
What is the issue that we're seeing that is motivating this decision?

## Decision
What is the change that we're proposing and/or doing?

## Consequences
What becomes easier or harder as a result of this change?
```

---

## 🏛️ The Full-Spectrum Parity Mandate (Unbreakable Correlation Invariant)

> **Mandatory Agent Law**: Never implement or enhance functionality locally while leaving external documentation, API specifications, or living GitHub Pages in a stale state. Stale documentation is a critical structural defect.

Every turn that implements, enhances, or modifies functionality (APIs, UI widgets, tools, or data schemas) MUST satisfy the **Quad-Parity Gate** in the **EXACT SAME TURN**:

### 1. API Contract & OpenAPI/Swagger Parity
- If an HTTP endpoint, parameter, or schema model is added/modified in the backend (`server.go`, `api.go`, etc.):
  - The corresponding OpenAPI 3.0 specification (`src/openapi.json`) MUST be updated with full path definitions, parameter descriptions, request/response bodies, and schema models in the **SAME TURN**.
  - The embedded interactive API console / Swagger UI (`/docs`) MUST be updated with interactive testing cards and handlers.
  - An endpoint without OpenAPI/Swagger documentation is considered broken and non-compliant.

### 2. Public Living Documentation Portal Parity (`docs/site/index.html`)
- The user-facing documentation portal / static GitHub pages MUST be updated in the **SAME TURN** with the new API contracts, architectural explanations, visual diagrams, and feature breakdowns.
- Run `./scripts/build-docs.sh` and `./.agents/skills/docs-portal/scripts/test-docs.sh` (`make docs`) to guarantee all 9 quality gates pass before concluding the turn.

### 3. Internal SDLC Quad Synchronization
- Every turn touching behavior MUST update:
  - `STATUS.md` — Traceability matrix (R-XXX, AC-XXX) & lifecycle phase.
  - `CONTEXT.md` — Single source of truth orientation & key decisions.
  - `DEVLOG.md` — Append-only session entry with forensic root causes & files changed.
  - `README.md` — Tool-level documentation, CLI flags, and API tables.

### 4. Zero-Assumption Verification
- Never claim completion or trust assertions without running deterministic tool queries (`curl`, headless Playwright, compiler runs) to verify that the documentation and APIs are reachable, syntactically valid, and returning expected status codes.

---

## Documentation Checklist

- [ ] README has all required sections
- [ ] All CLI flags/options documented
- [ ] All public API functions documented
- [ ] At least one runnable example exists
- [ ] CHANGELOG is up to date
- [ ] Living Documentation Portal (`docs/site/index.html`) updated with new endpoints/architectures
- [ ] Static documentation builds cleanly via `scripts/build-docs.sh` (`make docs`)
- [ ] SDLC Quad (`STATUS.md`, `CONTEXT.md`, `CHANGELOG.md`, `DEVLOG.md`) synchronized
- [ ] No broken links
- [ ] No placeholder text (TODO, TBD, etc.)
- [ ] Screenshots/diagrams for visual tools
- [ ] Installation instructions are complete and tested
