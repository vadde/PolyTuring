---
name: docs-portal
description: >-
  Authoritative engineering guide and automated testing protocol for the PolyTuring Living Documentation Portal. Covers cosmic liquid-glass design system, dynamic draggable sidebar resizer, Year 2500 mobile navigation drawer, strict touch-overflow containment, 1:1 bi-directional navigation sync, Mermaid diagram canvas invariants, KaTeX mathematical typesetting, interactive simulators, and headless verification.
---

# 📚 Living Documentation Portal Engineering Skill

This skill is the **single source of truth** for maintaining, styling, extending, and verifying the `vadde/PolyTuring` Living Documentation Portal ([https://vadde.github.io/PolyTuring/](https://vadde.github.io/PolyTuring/)).

Follow this guide whenever adding chapters, modifying CSS styles, rendering architecture diagrams, typesetting mathematical equations, building interactive simulator widgets, or supporting small commute gadgets.

---

## 🏛️ Portal Architecture & File Map

The documentation portal is built with **zero build toolchain dependencies** (pure vanilla HTML5, CSS3, ES2022 JavaScript, and CDN-loaded libraries):

```
PolyTuring/
├── docs/site/
│   ├── index.html        → Semantic HTML5 layout, chapter disclosure trees, widgets, mobile drawer
│   ├── style.css         → Liquid-glass design tokens, draggable resizer, responsive queries, canvas layout
│   └── app.js            → 1:1 ScrollSpy, sidebar resizer, mobile drawer, zoomable canvas pan/zoom, simulator engines
├── scripts/
│   └── build-docs.sh     → Static bundle packaging script (outputs to _site/)
├── .agents/skills/docs-portal/
│   ├── SKILL.md          → YOU ARE HERE — Authoritative design & engineering protocol
│   └── scripts/
│       └── test-docs.sh  → Automated 9-gate verification suite
└── .github/workflows/
    └── docs-publish.yml  → CI/CD workflow deploying main pushes to GitHub Pages
```

---

## 🎨 Design System & Color Tokens

The portal uses a **deep cosmic liquid-glass design system**. Always use CSS variables rather than hardcoding ad-hoc color values:

### Color Palette

| Token | CSS Variable | Hex / RGBA Value | Purpose |
|---|---|---|---|
| **Canvas Deep Base** | `--bg-deep` | `#0b0f19` | Deep space background canvas |
| **Card Surface** | `--bg-card` | `rgba(18, 26, 47, 0.55)` | Glassmorphism card background |
| **Card Border** | `--bg-card-border` | `rgba(255, 255, 255, 0.08)` | Subtle translucent glass card edge |
| **Hover Border** | `--bg-card-border-glow` | `rgba(0, 242, 254, 0.25)` | Interactive element border focus |
| **Cyan Accent** | `--accent-cyan` | `#00f2fe` | Primary interactive focus, metrics, active items |
| **Violet Accent** | `--accent-violet` | `#a855f7` | Diagram relational links, secondary badges |
| **Emerald Accent** | `--accent-emerald` | `#10b981` | Passing SLA badges, Rust indicators, health |
| **Amber Accent** | `--accent-amber` | `#f59e0b` | Cautionary callouts, architectural invariant warnings |
| **Rose Accent** | `--accent-rose` | `#f43f5e` | SLA breaches, error indicators, fatal alerts |
| **Primary Text** | `--text-main` | `#f1f5f9` | High-contrast body text and titles |
| **Muted Text** | `--text-muted` | `#94a3b8` | Subtitles, parameter descriptions |
| **Dim Text** | `--text-dim` | `#64748b` | Captions, footnotes, table metadata |

### Glassmorphism Recipe
```css
/* Canonical frosted glass surface */
background: rgba(15, 23, 42, 0.65);
backdrop-filter: blur(24px) saturate(190%);
-webkit-backdrop-filter: blur(24px) saturate(190%);
border: 1px solid var(--bg-card-border);
border-radius: var(--radius-md);
```

### Typography Hierarchy
- **Display Headings (`h1`, `h2`, `.chapter-title`)**: `'Outfit', sans-serif`, font-weight 700–800. Fluid sizing via `clamp()`.
- **Body & Controls**: `'Inter', sans-serif`, font-weight 400–600.
- **Code, Metrics & Data**: `'JetBrains Mono', monospace`, font-weight 500–700.

---

## ↔️ Interactive Draggable Sidebar Resizer Standard

Fixed-width sidebars degrade user experience by clipping titles on ultra-wide screens or consuming excessive reading space on medium displays.

### 1. Splitter DOM & Styling Invariants
The draggable splitter `.sidebar-resizer` must sit directly between `<aside class="sidebar">` and `<main class="content-area">`:

```html
<aside class="sidebar" id="main-sidebar">...</aside>
<div class="sidebar-resizer" id="sidebar-resizer" role="separator" aria-orientation="vertical" aria-label="Resize navigation sidebar" title="Drag to adjust sidebar width (Double-click to reset)"></div>
<main class="content-area">...</main>
```

```css
.sidebar-resizer {
  width: 10px;
  margin-left: -5px;
  margin-right: -5px;
  cursor: col-resize;
  position: sticky;
  top: var(--topbar-height);
  height: calc(100vh - var(--topbar-height));
  z-index: 35;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s ease;
  user-select: none;
  -webkit-user-select: none;
  touch-action: none;
}

.sidebar-resizer::after {
  content: '';
  width: 2px;
  height: 44px;
  border-radius: var(--radius-full);
  background: rgba(255, 255, 255, 0.2);
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.sidebar-resizer:hover::after,
.sidebar-resizer.is-resizing::after {
  height: 84px;
  width: 3px;
  background: var(--accent-cyan);
  box-shadow: 0 0 14px var(--accent-cyan);
}
```

### 2. Operational Invariants for Sidebar Resizing
1. **Bounds Clamping**: Resizing must be strictly clamped between `MIN_WIDTH = 240px` and `MAX_WIDTH = 560px`.
2. **Dragging Performance**: During drag, `body.resizing-sidebar` must set `cursor: col-resize !important; user-select: none !important;`, and `.sidebar` must set `transition: none !important;` to eliminate pointer lag.
3. **Persistence**: The chosen width must be persisted to `localStorage.setItem('polyturing_sidebar_width', width)` on pointer up, and restored on `DOMContentLoaded`.
4. **Double-Click Reset**: Double-clicking the splitter resets `--sidebar-width` to the default optimal width (`320px`).
5. **The Zero-Truncation Invariant**: Never clip titles with single-line ellipses (`white-space: nowrap; text-overflow: ellipsis;`). All navigation titles must use:
   ```css
   .tree-title-text {
     white-space: normal;
     word-break: break-word;
     line-height: 1.35;
     font-size: 0.85rem;
   }
   ```
   This ensures chapter names like `01. Omni-Graph Nervous System` wrap cleanly without ugly ellipsis truncation dots.

---

## 📱 Year 2500 Visionary Mobile Architecture (Commute Gadget Standard)

Modern software engineers and architects scout tools on mobile devices during commutes (smartphones, foldable devices, compact tablets) before testing on workstations. The portal must deliver a flawless, high-density, touch-optimized experience.

### 1. The Mobile Navigation Drawer Pattern
On viewports $\le 1024\text{px}$, the sidebar transforms into an off-canvas slide-out drawer:

```html
<header class="header-glass">
  <div class="header-left-cluster">
    <button class="mobile-nav-toggle" id="mobile-nav-toggle" aria-label="Open Navigation">
      <svg width="20" height="20" ...></svg>
    </button>
    ...
  </div>
</header>

<div class="layout-container">
  <!-- Backdrop MUST be inside layout container to share stacking context with sidebar -->
  <div class="mobile-drawer-backdrop" id="mobile-drawer-backdrop" aria-hidden="true"></div>
  
  <aside class="sidebar" id="main-sidebar">
    <div class="mobile-drawer-header">
      <div class="brand-section">...</div>
      <button class="mobile-drawer-close" id="mobile-drawer-close">✕</button>
    </div>
    <div class="mobile-drawer-tools">
      <!-- Token Simulator, API Playground, GitHub Repo links -->
    </div>
    ...
  </aside>
  ...
</div>
```

```css
@media (max-width: 1024px) {
  .mobile-nav-toggle { display: flex; }
  .sidebar-resizer { display: none !important; }

  .mobile-drawer-backdrop {
    display: block;
    position: fixed;
    inset: 0;
    background: rgba(4, 6, 12, 0.75);
    backdrop-filter: blur(12px);
    z-index: 1100;
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  }
  .mobile-drawer-backdrop.show {
    opacity: 1;
    pointer-events: auto;
  }

  .sidebar {
    position: fixed;
    top: 0; left: 0; bottom: 0;
    width: min(86vw, 360px);
    height: 100dvh;
    z-index: 1200;
    background: #090e1a;
    backdrop-filter: blur(32px) saturate(200%);
    box-shadow: 16px 0 60px rgba(0, 0, 0, 0.95);
    transform: translateX(-100%);
    transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }
  .sidebar.mobile-open {
    transform: translateX(0);
  }
}
```

### 2. The Stacking Context Invariant (CRITICAL)
> [!CAUTION]
> If `.layout-container` defines a CSS `z-index` and `position: relative`, placing `#mobile-drawer-backdrop` outside `.layout-container` traps the sidebar's stacking context below the backdrop.
> **Invariant**: Always place `#mobile-drawer-backdrop` inside `.layout-container` right before `<aside class="sidebar">`. This guarantees `.sidebar` (`z-index: 1200`) cleanly layers above the backdrop (`z-index: 1100`), which in turn dims `.content-area` (`z-index: 1`).

### 3. Strict Anti-Clobbering & Sticky-Safe Containment
1. **The Sticky Overflow Invariant**:
   - **NEVER use `overflow-x: hidden` on `html` or `body`** when sticky headers or sticky sidebars exist. In WebKit/Blink, `overflow-x: hidden` forces an implicit scroll container that destroys `position: sticky`.
   - **Always use `overflow-x: clip;`** on `html, body, .layout-container`. `clip` prevents horizontal scroll blowouts without creating a scroll container!
2. **The Table Containment Invariant**:
   - Every `<table>` must be wrapped in `<div class="table-wrapper">`.
   - CSS must enforce:
     ```css
     .table-wrapper {
       width: 100%;
       max-width: 100%;
       overflow-x: auto;
       -webkit-overflow-scrolling: touch;
     }
     table {
       min-width: 520px; /* Prevents columns from crushing into illegible slivers */
     }
     ```
3. **KaTeX & Code Blocks Containment**:
   - `.katex-display` must define `overflow-x: auto; overflow-y: hidden; -webkit-overflow-scrolling: touch; max-width: 100%;`.
   - `.code-container pre` must define `overflow-x: auto; -webkit-overflow-scrolling: touch; max-width: 100%;`.
4. **Fluid Typography (`clamp`)**:
   - Hero Headings: `font-size: clamp(1.6rem, 5.5vw, 2.2rem);`
   - Chapter Titles: `font-size: clamp(1.25rem, 4.2vw, 1.75rem);`
   - Section H2: `font-size: clamp(1.1rem, 3vw, 1.4rem);`
   - Mode Buttons: `flex: 1; min-width: 80px; justify-content: center;`

---

## 🧭 Navigation & The 1:1 Synchrony Invariant

### 1. The Structure of Navigation
The left sidebar is composed of `.nav-tree-group` sections containing:
1. `.nav-tree-header`: Chapter group bar with disclosure chevron and language pill.
2. `.nav-tree-children`: List of `.nav-tree-link` anchors (`href="#sec-id"`).

### 2. The Bi-Directional Synchrony Contract
* **Header Click**: Clicking `.nav-tree-header` toggles `.expanded` on the sidebar group **AND** simultaneously toggles `.chapter-expanded` on the corresponding `<section class="chapter-section">` in the content pane.
* **Child Link Click**: Clicking a `.nav-tree-link`:
  1. Ensures containing parent chapter is expanded.
  2. Applies Focus Mode rules (collapses sibling chapters if Focus Mode is active).
  3. Smooth-scrolls to the target element with sticky-header offset (`160px`).
  4. Applies temporary `.flash-focus` neon glow (1.8s) to the target element.
  5. Sets `isNavClicking = true` for `850ms` to prevent smooth-scroll events from overriding the active highlight.
  6. On mobile ($\le 1024\text{px}$), automatically dismisses the mobile navigation drawer.
* **ScrollSpy Tracking (CRITICAL)**:
  - ScrollSpy tracks **EVERY TARGET ELEMENT LINKED IN THE NAVIGATION** individually, NOT just `.chapter-section`!
  - Targets include: `#sec-omni-arch`, `#sec-omni-galaxies`, `#sec-omni-coupling`, `#sec-omni-ebf`, `#sec-omni-ast`, etc.
  - **INVARIANT**: ScrollSpy must **NEVER** mutate tree open/close state (`classList.add('expanded')`) while scrolling! It ONLY updates `.active` on the link and `.active-branch` on the group header.
  - **INVARIANT**: When "Collapse All" is clicked, scrolling must **NEVER** forcefully pop open closed chapters.

---

## 🔬 Theoretical & Mathematical Foundations Standard

When documenting tools in PolyTuring, avoid superficial descriptions of CLI flags. Present the **rigorous computer science and graph theory foundations** that empower the architecture:

### 1. Omni-Graph: Galaxy Subsystems & Louvain/Leiden Modularity Clustering
Document how the polyglot AST call graph is partitioned into architectural galaxy subsystems using Louvain / Leiden modularity optimization ($Q$):
$$Q = \frac{1}{2m} \sum_{ij} \left( A_{ij} - \gamma \frac{k_i k_j}{2m} \right) \delta(c_i, c_j)$$

### 2. Robert C. Martin Coupling Metrics & Dynamic Blast Radius Contracts
Document how cross-boundary dependency vectors are quantified before writing refactored code:
- **Afferent Coupling ($C_a$)**: Incoming external symbols depending on this subsystem.
- **Efferent Coupling ($C_e$)**: Outgoing dependencies this subsystem calls.
- **Instability Metric ($I$)**:
  $$I = \frac{C_e}{C_a + C_e}, \quad I \in [0, 1]$$
- **Architectural SLA Thresholds**:
  - **Core Foundation**: $I \le 0.25$ (High stability, maximum caution).
  - **Domain Service**: $0.25 < I \le 0.65$ (Balanced volatility).
  - **Orchestrator / Leaf**: $I > 0.65$ (High volatility, safe to patch).

### 3. Session Explorer: Trajectory Observability & Memory Recovery
Document how standalone Go + Bun binaries index JSONL agent trajectory logs into SQLite3 WAL, enabling instant WIP memory recovery for agent context crashes.

### 4. Universal Assistant Matrix
Position every tool as natively integrated with Google Antigravity today, while providing an architectural blueprint for Model Context Protocol (MCP) server integration with GitHub Copilot, Claude Code, and Cursor.

---

## 📊 Mermaid Architecture Diagrams & Zoomable Canvas

### 1. Canvas DOM Anatomy
Every architecture diagram and complex table must be wrapped in a zoomable canvas:
```html
<div class="canvas-container">
  <div class="canvas-header-bar">
    <span class="canvas-type-tag">Architecture Blueprint · 4-Tier Container Stack</span>
    <div class="canvas-controls-group">
      <button class="canvas-tool-btn zoom-out-btn" title="Zoom Out">−</button>
      <span class="zoom-level-badge">100%</span>
      <button class="canvas-tool-btn zoom-in-btn" title="Zoom In">+</button>
      <button class="canvas-tool-btn zoom-reset-btn" title="Reset Zoom">⟲</button>
      <button class="canvas-tool-btn fullscreen-btn" title="View Fullscreen Theater">
        <svg width="13" height="13" ...></svg>
        <span>Theater</span>
      </button>
    </div>
  </div>
  <div class="canvas-viewport">
    <div class="canvas-content-surface">
      <pre class="mermaid">
flowchart TB
    ...
      </pre>
    </div>
  </div>
  <div class="canvas-caption">Figure 1.1: Caption text here</div>
</div>
```

### 2. The Critical SVG Layout Invariant (DO NOT VIOLATE)
> [!CAUTION]
> **NEVER apply `width: auto !important` or `height: auto !important` to `.mermaid svg` inside nested flex containers!**
> Overriding SVG width/height with `auto` inside flex containers collapses the SVG's flex basis to **0px by 0px**, rendering diagrams invisible.

**The Golden Layout Rules for Mermaid:**
```css
.canvas-viewport {
  overflow: auto;
  padding: 28px 24px;
  display: block;
  text-align: center;
  min-height: 220px;
  max-height: 600px;
  cursor: grab;
  user-select: none;
}

.canvas-content-surface {
  transform-origin: center top;
  transition: transform 0.12s ease-out;
  display: inline-block;
  margin: 0 auto;
  text-align: center;
  max-width: 100%;
}

.mermaid {
  display: inline-block !important;
  margin: 0 auto !important;
  text-align: center !important;
  overflow: visible !important;
}

.mermaid svg {
  display: inline-block !important;
  max-width: 100% !important;
  height: auto !important;
}
```

### 3. Mermaid Node Label Escaping Rule
> [!WARNING]
> Unescaped `<` or `>` characters inside Mermaid node labels (e.g. `[Condenser (<1500 Tokens)]`) parse as HTML tags and **crash the Mermaid parser**!
> **Always write:** `≤ 1,500 Tokens` or `&lt;1500 Tokens`.

---

## 🎛️ Interactive Simulators & Playgrounds

Interactive modules (such as the **Context Condenser Token Simulator** and **API Playground**) are first-class citizens of the portal:

### 1. Parameter Control Card Anatomy (The 2-Line Rule)
Never squeeze parameter titles and values into a single row with `justify-content: space-between`. On narrow cards or multi-word titles, this forces ugly multi-line wrapping.

**Always use the 2-Line Stack inside isolated frosted cards:**
```html
<div class="control-item">
  <!-- Line 1: Parameter Title -->
  <div class="control-title">Neighbor Signatures Cap</div>
  <!-- Line 2: Monospace Quantitative Value -->
  <div class="control-value" id="val-neighbor-chars">2400 chars</div>
  <!-- Line 3: Range Slider Track -->
  <input type="range" id="sim-neighbor-chars" min="500" max="4500" step="100" value="2400">
</div>
```

---

## 🧪 Automated Testing Protocol (9-Gate Audit)

Never push documentation portal changes without running the automated test suite!

```bash
# Run the all-in-one headless audit:
bash .agents/skills/docs-portal/scripts/test-docs.sh

# Or via the root Makefile target:
make docs-test
```

### The 9 Automated Verification Gates
1. **Core Asset Presence**: Verifies `index.html`, `style.css`, and `app.js`.
2. **Mermaid Syntax Auditing**: Detects unescaped `<` or `>` characters in node labels.
3. **SVG Sizing Invariants**: Validates that CSS does not contain collapsing `width: auto !important` on `.mermaid svg`.
4. **KaTeX Integration**: Verifies stylesheet, scripts, and auto-render calls.
5. **Draggable Sidebar Resizer Invariants**: Audits `#sidebar-resizer`, `col-resize`, `initSidebarResizer`, `localStorage` key, and non-clipping text wrapping.
6. **Year 2500 Mobile Navigation Drawer & Anti-Clobbering**: Audits `#mobile-nav-toggle`, `#mobile-drawer-backdrop`, `#mobile-drawer-close`, `initMobileDrawer`, `.sidebar.mobile-open`, and sticky-safe `overflow-x: clip`.
7. **Table & Math Touch Overflow Containment**: Ensures all `<table>` tags are inside `.table-wrapper`, min-width is specified, and `.katex-display` has `overflow-x: auto`.
8. **Static Bundle Build**: Compiles to `_site/` and checks size/file counts.
9. **Built Bundle Integrity**: Verifies compiled distribution files.

---

## 🚀 Playbook: Adding a New Chapter or Tool

When documenting a new tool (e.g. `Chapter 05 · New Tool`):

1. **Update Left Navigation (`docs/site/index.html`)**:
   Add a new `.nav-tree-group`:
   ```html
   <div class="nav-tree-group">
     <div class="nav-tree-header" data-target="#ch5-new-tool" title="05. New Tool Name">
       <div class="tree-title-group">
         <span class="tree-chevron">▶</span>
         <span class="tree-icon">🚀</span>
         <span class="tree-title-text" title="05. New Tool Name">05. New Tool Name</span>
       </div>
       <span class="nav-pill pill-rust">Rust</span>
     </div>
     <div class="nav-tree-children">
       <a href="#ch5-new-tool" class="nav-tree-link">System Overview</a>
       <a href="#sec-new-arch" class="nav-tree-link">Architecture Topology</a>
       <a href="#sec-new-theory" class="nav-tree-link">Theoretical Foundations</a>
       <a href="#sec-new-usage" class="nav-tree-link">Usage & Quickstart</a>
     </div>
   </div>
   ```

2. **Add Chapter Section (`docs/site/index.html`)**:
   Enclose all comparison tables in `<div class="table-wrapper">`, KaTeX in `$$ ... $$`, and diagrams in `.canvas-container`.

3. **Verify and Deploy**:
   ```bash
   make docs-test
   git commit -am "feat(docs): add Chapter 05 documentation"
   git push origin main
   ```
