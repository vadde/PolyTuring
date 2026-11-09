# 📋 Product Requirements Document (PRD) — Sub-Pixel Code Gutter Alignment, Multi-Hunk Radar & File Buffer Intelligence

**Tool**: `session-explorer`  
**Feature**: Sub-Pixel Code & Gutter Synchronous Alignment Engine, Forensic File Buffer Telemetry, and Multi-Hunk Navigation Radar (13+ Hunks Scalability)  
**Date**: 2026-11-09  
**Author**: Antigravity  
**Status**: `spec-review`  
**Constitutional Compliance**: Superpower 101 Constitution (Rule 101) & Spec-Driven Development (Rule 04)

---

## 1. Executive Summary & Problem Formulation

### 1.1 The User Problem
In session explorer file review, two critical friction points have surfaced from real-world usage on complex codebase files (such as `tools/omni-graph/src/ingestion/symbol_table.rs` in session `3bff045b-cc16-4535-92aa-1f55dbac041b`):
1. **Vertical Line & Spacing Desynchronization**: Code lines and line numbers diverged severely in vertical spacing. Because `<span class="code-line">` was formatted as a block and joined with `\n` inside `<pre>`, each line of code occupied 43.6px while each gutter line occupied 22px (~2x expansion). Code lines appeared double-spaced with blank gaps, and by line 50 or 100, the code had drifted dozens of lines away from its true gutter number. Furthermore, Highlight.js default padding pushed line 1 down by 13.44px.
2. **File Size & Line Count Ambiguity (The "400 Lines" Mystery)**: Users observed files showing ~400 lines (e.g. `symbol_table.rs` has 399 lines on disk) and were uncertain whether the file viewer has a hardcoded 400-line ceiling, whether truncation occurred, or what edge-case boundary conditions govern large-file rendering across the application.
3. **Multi-Hunk Cognitive Overload (13+ Hunks Scalability)**: In heavily modified files with 13 or more hunks, sequential stepping (`<` `>`) requires excessive clicking, and engineers lose their spatial orientation across the file's overall architecture.

### 1.2 The Visionary Solution (Year 2500 Cosmic Liquid-Glass Engine)
1. **Sub-Pixel Synchronous Alignment Engine (Zero-Drift Invariant)**:
   - Normalize Highlight.js styling: remove container padding/margins (`padding: 0 !important; margin: 0 !important;`).
   - Lock `.code-line` and `.code-line-numbers span` to identical `22px` height, `22px` line-height, and `box-sizing: border-box`.
   - Eliminate `\n` joiner between block spans.
   - Mathematically guarantee $\Delta(line_i) = |y_{code}(i) - y_{gutter}(i)| = 0.00\text{px}$ across all lines $i \in [1, N]$.
   - Dynamic wrap reconciliation: when `Wrap` is toggled on, an automated height-synchronization observer adjusts gutter span heights to match wrapped code line client heights in real time.
2. **Forensic File Buffer Telemetry Capsule**:
   - Provide an ambient holographic telemetry capsule in the viewer header indicating exact line count, file size, encoding, and explicit buffer integrity: `● 100% On-Disk (Full Buffer)`.
   - Clear edge-case documentation and visual boundary badges for files approaching scale thresholds.
3. **Multi-Hunk Navigation Radar & Quick-Jump Dropdown**:
   - Introduce an interactive **Hunk Spatial Radar Dropdown** in the review pane header: listing all hunks with 1-indexed numbers, AST function/struct names, line ranges, and change signs (`Hunk 4: @@ -120,8 @@ fn parse_symbol (+5/-2)`).
   - Allows instant 1-click jumps to any hunk across 13+ hunk files without sequential click-fatigue.
4. **Minimap Holographic Tooltip Previews & Active Hunk Spotlight**:
   - Hovering over ticks on the 6px minimap rail displays a floating acrylic glass tooltip showing the hunk summary and AST scope.
   - When a hunk is active, the corresponding line range in the left file viewer is framed with an illuminated ambient spotlight rail (`.hunk-active-spotlight`).

---

## 2. User Personas & Lenses

### 2.1 The Junior / Fast-Prototyping GenAI Engineer (🌱 Kindergarten Lens)
- **Need**: Pristine, clean typography without weird double-spacing or ghost blank lines. Absolute certainty that line 42 on the left is line 42 of the code.

### 2.2 The Senior Polyglot Systems Engineer (🛠️ Senior Pair Programmer Lens)
- **Need**: When reviewing a file with 13 or 30 hunks, wants a quick-jump radar menu to jump directly to `fn resolve_scope` without clicking "Next Hunk" 12 times.

### 2.3 The Systems Architect & Performance Lead (🏛️ Systems Architect Lens)
- **Need**: Explicit understanding of system limits: 10MB HTTP buffer ceiling, 2MB git diff ceiling, memory bounds, and zero layout shift.

### 2.4 The AI Research Scientist (🔬 Research Scientist Lens)
- **Need**: High-density code visualization where every token and line position corresponds exactly to LLM agent step coordinates.

---

## 3. Detailed Requirements Matrix

| ID | Title | Description | Priority |
| :--- | :--- | :--- | :--- |
| **R-067** | Sub-Pixel Code & Gutter Synchronous Alignment Engine | Eliminates `\n` joiner, normalizes `.hljs` padding to 0, locks line heights to 22px, guarantees 0.00px drift across all lines, and supports dynamic wrap height sync. | 🔴 Critical |
| **R-068** | Forensic File Buffer Telemetry Capsule | Ambient header badge displaying line count, size, and explicit `● 100% On-Disk (Full Buffer)` status; documents system boundary conditions. | 🟡 High |
| **R-069** | Multi-Hunk Navigation Radar & Spatial Quick-Jump Dropdown | Interactive dropdown selector for jumping directly to any hunk among 13+ hunks with AST symbols. | 🟡 High |
| **R-070** | Minimap Holographic Tooltip Previews | Floating glassmorphic tooltip upon hovering over minimap ticks showing hunk number, symbol, and line offset. | 🟢 Medium |
| **R-071** | Active Hunk Spotlight & Ambient Framing Rail | Framing line container highlighting the currently active hunk's line range in the left file viewer. | 🟢 Medium |
| **R-072** | High-Scale Viewport Virtualization & Safe Buffer Handling | Graceful handling and rendering of files with 5,000+ lines with 60fps scrolling. | 🟢 Medium |

---

## 4. Edge Cases & Boundary Conditions

1. **Large Files (> 10MB)**: Backend returns HTTP 413 (`exceeds maximum preview limit of 10MB`); frontend displays a helpful, non-blocking error banner with command-line instructions (`head -n 500`).
2. **Files with 399 vs 400 Lines (Trailing Newline)**: Posix trailing newlines are preserved accurately; line count reflects exact physical lines.
3. **Line Wrapping Toggled On**: Long lines wrap onto multiple visual rows; `syncWrappedLineHeights()` measures code line height and expands the gutter number span proportionally so line numbers never drift.
4. **Files with 0 Hunks vs 20+ Hunks**: Hunk stepper and radar dropdown smoothly adapt; if 0 hunks, stepper hides; if > 1 hunk, radar dropdown provides quick random access.
5. **Minified Single-Line Files**: A 500KB single-line JSON or minified JS file renders as 1 line without crashing.

---

## 5. Success Metrics & Verification Gate
- **Zero-Drift Invariant**: Measured in Playwright across lines 1, 50, 100, 200, 399: $|y_{code} - y_{gutter}| = 0\text{px}$.
- **Zero Blank Gaps**: Line height ratio $\frac{H_{code}}{H_{gutter}} = 1.000$.
- **Test Suite**: 100% pass on all Go unit tests (`go test -v ./src/...`).
- **Playwright Headless Audit**: Visual verification on `symbol_table.rs` (13 hunks, 399 lines).
