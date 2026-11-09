# 📋 Product Requirements Document (PRD) — Diff Sync, Dual Line Number Gutter & Spatial Change Intelligence

**Tool**: `session-explorer`  
**Feature**: Diff Spatial Sync, GitHub-Standard Dual Line Numbers, Gutter Annotations & Spatial Change Cockpit  
**Date**: 2026-11-09  
**Author**: Antigravity  
**Status**: `spec-review`  
**Constitutional Compliance**: Superpower 101 Constitution (Rule 101) & Spec-Driven Development (Rule 04)

---

## 1. Executive Summary & Problem Formulation

### 1.1 The User Problem
In software development and GenAI coding assistant environments, reading diffs in isolation causes severe **cognitive disorientation**. When an engineer inspects a code review pane on the right:
1. **Ambiguous Line Mapping**: A single line number column fails to communicate whether a number refers to the base revision before modification or the head revision after modification.
2. **Mental Disconnect from Full Codebase**: The user sees changed hunks in the review pane, but has no intuitive visual cues inside the left-hand code viewer showing where those changes exist in the context of the entire file.
3. **Scroll Blindness in Large Files**: In a 2,000-line file with 3 small modifications, an engineer must hunt and scroll blindly through the document to find the changed blocks.
4. **Tool Context Switching**: Engineers are frequently forced to leave their session intelligence cockpit and open full IDEs (VS Code, Cursor) or GitHub web revisions just to verify surrounding code relationships and symbols.

### 1.2 The 100-Year Visionary Solution
Transform the File Viewer and Review Pane into an **interconnected, bidirectional holographic change cockpit**:
- **GitHub-Standard Dual Line Numbers**: Pristine tabular dual-column numbering (`Old Line #` vs `New Line #`) in both unified and split views.
- **In-File Gutter Diff Stripes**: Real-time ambient green (`+`), blue/purple (`~`), and rose (`-`) diff markers inside the left code editor gutter, updating dynamically as review vectors switch.
- **Holographic Diff Minimap Rail**: A sleek, vertical spatial heatmap along the file viewport displaying glowing pips for every modification in the entire file.
- **Synchronized Hunk Stepping**: Dual-scrolling `⏮ Prev Hunk (⌥[)` and `⏭ Next Hunk (⌥])` controls that smoothly center both views on each modification with a pulsating cosmic aura.
- **AST Semantic Scope Capsules**: Automatically extracts enclosing structs, functions, or methods from hunk headers (`@@ ... @@`) to display semantic context badges.
- **Desktop IDE Deep-Linking & Hunk Patching**: 1-click surgical hunk copying and deep-links (`cursor://file/...` / `vscode://file/...`) to immediately open the exact line in local desktop IDEs.

---

## 2. User Personas & Lenses

### 2.1 The Junior / Fast-Prototyping GenAI Engineer (🌱 Kindergarten Lens)
- **Goal**: Needs instant visual clarity on what the AI agent did to their code without deciphering raw unified diff math.
- **Need**: Clear `Old` vs `New` line numbers, obvious color-coded gutter bars in the file viewer, and 1-click jumps so they never get lost.

### 2.2 The Senior Polyglot Systems Engineer (🛠️ Senior Pair Programmer Lens)
- **Goal**: High-speed, high-density code inspection with keyboard efficiency.
- **Need**: Keyboard shortcuts (`⌥[`, `⌥]`), rapid hunk-by-hunk navigation, surgical hunk patch copying, and tabular numeral alignment.

### 2.3 The Engineering Architect & Lead (🏛️ Systems Architect Lens)
- **Goal**: Zero layout shift, high frame-rate scrolling, zero DOM thrashing on files with 10,000+ lines, and semantic scope awareness.
- **Need**: Knowing which functions and blast-radius boundaries were touched by each revision or agent step.

### 2.4 The AI Research Scientist (🔬 Research Scientist Lens)
- **Goal**: Correlating LLM agent thinking with granular token mutations across prompt trajectories.
- **Need**: Tracing agent reasoning to specific line modifications and capturing fine-grained mutation velocity.

---

## 3. Product Requirements Breakdown

### Feature 1: GitHub-Standard Dual Line Numbers Gutter (R-061)
- Unified view MUST display two distinct gutter columns:
  - Column 1: `Old Line #` (pre-modification line number; blank on additions).
  - Column 2: `New Line #` (post-modification line number; blank on deletions).
- Context lines MUST display both old and new numbers with identical values.
- Monospaced typography with tabular numeral alignment (`font-variant-numeric: tabular-nums`).
- Split view MUST display dedicated line numbers in both the left (old) and right (new) columns.

### Feature 2: In-File Code Gutter Diff Decorations (R-062)
- When a diff is loaded in the review pane (Working Copy, Staged, Commit, or Agent Step), the left file viewer's gutter MUST dynamically reflect the active changes:
  - Added lines: Green border stripe + subtle glyph.
  - Modified lines: Blue/purple border stripe.
  - Deletion points: Rose chevron marker indicating the exact line where deletion occurred.
- Hovering over a gutter decoration reveals a mini tooltip with change details.
- Clicking a gutter marker smoothly scrolls the review pane directly to that hunk.

### Feature 3: Holographic Diff Minimap Rail / Heatmap (R-063)
- An ambient vertical rail mounted alongside the file viewer scrollbar.
- Renders glowing color-coded notches positioned at `(lineNum / totalLines) * 100%`:
  - Green notches for additions.
  - Rose notches for deletions.
  - Cyan/Amber for modifications.
- Clicking any notch on the minimap immediately scrolls the file viewer and review pane directly to that location.

### Feature 4: Synchronized Hunk Stepper & Navigation (R-064)
- Topbar and floating controls for `⏮ Prev Hunk (⌥[)` and `⏭ Next Hunk (⌥])`.
- Displays active position pill: `Hunk X of Y`.
- When stepped, both panes (source code viewer on the left, diff review pane on the right) smoothly scroll into view simultaneously, applying a glowing pulse animation (`jump-pulse`).

### Feature 5: AST Semantic Symbol Breadcrumbs (R-065)
- Extracts enclosing symbol definitions from hunk headers (e.g. `@@ -484,4 +484,92 @@ type SessionAgentContext struct {` $\to$ `type SessionAgentContext struct`).
- Renders an interactive symbol capsule (`📦 struct SessionAgentContext` or `⚡ func HandleGitFileDiff`) on each hunk header.
- Provides immediate mental mapping of the affected function or class.

### Feature 6: Surgical Hunk Patch Copy & Desktop IDE Deep-Linking (R-066)
- Individual `📋 Hunk` button on each hunk header to copy only that atomic hunk as a valid git patch.
- `↗ IDE` action button on the file viewer header that generates and triggers `vscode://file/<path>:<line>` or `cursor://file/<path>:<line>` deep-links.

---

## 4. Success Criteria & Metrics

1. **Zero Layout Shift ($\Delta y = 0$)**: Applying gutter decorations and dual numbers must not alter code line heights or horizontal layout flow.
2. **Sub-15ms Synchronization**: Clicking between review diff lines and file viewer lines must execute under 15ms.
3. **100% Theme Cohesion**: All components must support all 4 themes (`cosmic-glass`, `fluent-mica`, `holo-cyberpunk`, `matrix-obsidian`) with pristine contrast ratios exceeding WCAG AAA standards.
