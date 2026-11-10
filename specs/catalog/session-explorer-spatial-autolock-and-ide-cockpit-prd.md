# 📋 PRD: Session Explorer — Spatial Auto-Lock, Sticky Zero-Drift Gutter, & Holographic Multi-IDE Cockpit

> **Document ID**: `PRD-SE-008`  
> **Target Subsystem**: `tools/session-explorer`  
> **Lifecycle Phase**: `draft` → `spec-review`  
> **Author**: Antigravity Autonomous Agent & Systems Architect  
> **Date**: 2026-11-10  
> **Prerequisites**: `PRD-SE-005`, `PRD-SE-006`, `PRD-SE-007` (v0.5.2)

---

## 1. Executive Summary & Problem Formulation

In `v0.5.2`, Session Explorer introduced the **Holographic Git & Review Pane**, GitHub-standard dual line gutters, diff minimap heatmap, and sub-pixel alignment engine. However, real-world forensic evaluation across multi-thousand-line source codebases revealed three critical ergonomics and developer-experience friction points:

1. **Hunk Desynchronization on Step Scrubber Navigation**:
   When stepping through agent mutation turns using `◀` / `▶` (or arrow keys), the review diff pane updates its inline/split view, but the left file viewer pane remains parked at its previous scroll offset. The user is forced to manually click the Hunk header / pill to locate and highlight where the change occurred on the file.
2. **Gutter Loss on Horizontal Scroll**:
   When source code contains long lines exceeding the viewport width, scrolling horizontally to the right causes `.code-line-numbers` to scroll off to the left and disappear. The user loses line number orientation completely and must scroll all the way back to the left.
3. **Hardcoded Monolithic IDE Launcher**:
   The "↗ IDE" button in the file viewer topbar hardcodes `cursor://file/<path>:<line>`. Developers using Visual Studio Code, Windsurf, JetBrains (IntelliJ, GoLand, PyCharm, WebStorm), Sublime Text, or terminal editors cannot launch their editor of choice, suffer browser URI scheme failures, and have no way to configure their preferred editor.

This PRD establishes the requirements to eliminate these friction points, transforming Session Explorer into a fluid, Year 2500 **GenAI Mission Control Cockpit**.

---

## 2. User Personas & The 4 Constitutional Lenses

| Lens | Role & Hat | Core Need in this Milestone |
| :--- | :--- | :--- |
| **🌱 Kindergarten Lens** | Junior Dev / First-Time Auditor | "When I click next turn, show me immediately what changed in the file. Don't hide line numbers when I scroll sideways. Let me click an editor I actually have installed." |
| **🛠️ Senior Pair Programmer Lens** | Full-Stack / GenAI Engineer | "Dynamic link between active hunk and IDE button. The moment I click 'Open in VS Code', land my cursor directly on the mutated AST symbol. Save my preference in `localStorage`." |
| **🏛️ Systems Architect Lens** | Tech Lead & Core Maintainer | "Dual-vector launch strategy: Go backend (`POST /api/ide/open`) launches local CLI tools directly with zero browser confirmation dialogs; browser URI protocol acts as graceful remote fallback." |
| **🔬 Research Scientist Lens** | AI Agent & Reasoning Auditor | "Maintain a zero-latency, context-preserving bridge between the agent's monologue/trajectory and the developer's runtime workspace. Zero cognitive friction." |

---

## 3. High-Level Requirements

### R-073: Automatic Hunk 1 Spatial Auto-Lock & Framing on Scrubber Navigation
- When navigating steps via the step scrubber (`◀` / `▶`), keyboard arrow keys (`Left`/`Right`), or review vector tab switching, the system must automatically invoke `stepToHunk(0)`.
- The left file viewer pane must smoothly scroll to the exact start line of the primary hunk.
- The ambient active hunk spotlight (`.hunk-active-spotlight`) must frame all lines of the hunk.
- The radiant beam pulse animation must trigger on the target line in the file viewer.
- The IDE launcher target line must dynamically sync to the active hunk's start line.

### R-074: Sub-Pixel Sticky Zero-Drift Line Number Gutter
- `.code-line-numbers` must be anchored with `position: sticky; left: 0; z-index: 10;`.
- The sticky gutter must use an opaque, frosted glassmorphic backdrop token across all 4 themes (**Liquid Glass**, **Fluent Aero Mica**, **Holo Cyberpunk**, **Matrix Obsidian**).
- Horizontal code scrolling behind the gutter must never bleed text or glyphs through the line number column.
- Gutter diff decorations (`.gutter-diff-add`, `.gutter-diff-mod`, `.gutter-diff-del`) must remain pinned to the left edge along with the line numbers.
- Maintain zero line-height drift (`line-height: 22px`) across 10,000+ lines.

### R-075: Sticky Viewport Docking for Diff Minimap Heatmap Rail
- `.file-diff-minimap-rail` must be anchored with `position: sticky; right: 0; z-index: 12;` within the scrolling code container.
- Horizontal code scrolling must not push the minimap heatmap rail offscreen; it must remain docked to the visible right boundary of `.file-viewer-body`.

### R-076: Holographic Multi-IDE Selector Split-Button & Glassmorphic Palette
- Replace the static "↗ IDE" button with an interactive holographic split-button:
  - **Left Action**: Displays the active editor icon and name (e.g. `[ 🚀 Cursor ]` or `[ 💙 VS Code ]` or `[ 🌊 Windsurf ]`). Clicking it triggers immediate launch at the active hunk line.
  - **Right Chevron (`▾`)**: Triggers the **Holographic IDE Selector Palette** dropdown.
- Palette options:
  - 🚀 **Cursor** (`cursor://file/{path}:{line}`)
  - 💙 **Visual Studio Code** (`vscode://file/{path}:{line}`)
  - 🌊 **Windsurf** (`windsurf://file/{path}:{line}`)
  - 🧠 **JetBrains** (IntelliJ IDEA / GoLand / PyCharm / WebStorm) (`idea://open?file={path}&line={line}`)
  - ⚡ **VS Code Insiders** (`vscode-insiders://file/{path}:{line}`)
  - 📜 **Sublime Text** (`subl://open?url=file://{path}&line={line}`)
  - 🍎 **System Default** (`open {path}`)
  - 📋 **Copy CLI Launch Command** (e.g. `cursor -g /path:123`) to clipboard with toast notification.
- Persistence: Default editor preference stored in `localStorage('se_preferred_ide')`.

### R-077: Backend Native Process Launch API (`POST /api/ide/open`)
- Implement a secure Go REST endpoint `POST /api/ide/open`.
- Request Payload:
  ```json
  {
    "editor": "cursor",
    "path": "/absolute/path/to/file.rs",
    "line": 42
  }
  ```
- Backend execution via `exec.LookPath`:
  - `cursor`: `cursor -g <path>:<line>`
  - `code`: `code -g <path>:<line>`
  - `windsurf`: `windsurf -g <path>:<line>`
  - `idea`: `idea --line <line> <path>`
  - `subl`: `subl <path>:<line>`
  - `system`: `open <path>` (macOS) / `xdg-open` (Linux)
- Path validation & safety: reject directory traversal attacks or paths outside permitted roots.
- Fallback response: if binary is not installed in `$PATH`, returns `{ "success": false, "fallback_uri": "...", "command": "..." }`, prompting frontend to execute protocol URI or copy CLI command.

### R-078: Automated End-to-End Verification Harness
- Go unit test suite covering `/api/ide/open` payload parsing, security guards, and command resolution.
- Playwright automated browser verification confirming:
  1. Scrubber click auto-scrolls and highlights Hunk 1 in file viewer.
  2. Horizontal scroll retains sticky line numbers at `left: 0` without text bleed.
  3. Sticky minimap rail stays pinned to visible `right: 0`.
  4. IDE split-button opens palette, remembers preference, and executes dual-vector dispatch.
