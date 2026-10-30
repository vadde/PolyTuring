---
name: docs-portal
description: >-
  Authoritative engineering guide and automated testing protocol for the PolyTuring Living Documentation Portal. Covers Year 2500 cosmic liquid-glass design system, 4-theme engine (Liquid Glass, Fluent Aero Mica, Holo Cyberpunk, Matrix Obsidian), Haptic Web Audio feedback micro-transients, interactive draggable sidebar resizer and collapsible rail (⌘B / [), live sidebar search filter, commute gadget mobile drawer, zero-truncation titles, touch-overflow containment, 1:1 ScrollSpy and right-side TOC reading progress HUD, Mermaid SVG layout rules, fullscreen visual theater workbench with dynamic viewport fit and SVG text selection, KaTeX math typesetting, interactive simulators, and the 14-gate automated headless audit suite.
---

# 📚 Living Documentation Portal Engineering Skill

This skill is the **authoritative engineering protocol and single source of truth** for maintaining, styling, extending, and verifying the `vadde/PolyTuring` Living Documentation Portal ([https://vadde.github.io/PolyTuring/](https://vadde.github.io/PolyTuring/)).

Follow this guide whenever adding chapters, modifying CSS styles, rendering architecture diagrams, typesetting mathematical equations, building interactive simulator widgets, or supporting small commute gadgets.

---

## 🏛️ Portal Architecture & File Map

The documentation portal is built with **zero build toolchain dependencies** (pure vanilla HTML5, CSS3, ES2022 JavaScript, and standard CDN libraries):

```
PolyTuring/
├── Makefile                               # Canonical DevEx commands (docs-test, docs-serve, docs-build)
├── docs/site/
│   ├── index.html                         # Semantic HTML5 layout, chapter disclosure trees, widgets, mobile drawer
│   ├── style.css                          # Liquid glass tokens, 4-theme engine, resizer, collapsible rail, TOC HUD
│   └── app.js                             # Haptic audio, ScrollSpy, sidebar filter, collapsible rail, theater mode
├── scripts/
│   └── build-docs.sh                      # Static bundle packaging script (outputs to _site/)
├── .agents/skills/docs-portal/
│   ├── SKILL.md                           # YOU ARE HERE — Authoritative design & engineering protocol
│   └── scripts/
│       └── test-docs.sh                   # 14-Gate automated headless verification suite
└── .github/workflows/
    └── docs-publish.yml                   # CI/CD workflow deploying main pushes to GitHub Pages
```

---

## 🎨 Design System & Multi-Theme Engine

The portal uses a **deep cosmic liquid-glass design system** supporting 4 distinct themes:

### Color Palette Tokens

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

### Theme Invariants:
1. **Liquid Glass (`[data-theme="liquid"]`)**: Default deep cosmic glass with vibrant cyan/violet edge glows.
2. **Fluent Aero Mica (`[data-theme="mica"]`)**:
   - **High-Contrast Code Blocks**: Code blocks MUST use dark text on bright frosted surface (`background: #ffffff; color: #0f172a !important;`). Never leave white-on-white text!
   - **Crisp Paper Surfaces**: Frosted glass with subtle slate borders and deep navy accents (`--accent-cyan: #0284c7;`).
3. **Holo Cyberpunk Neon (`[data-theme="holo"]`)**: High-saturation hot-pink (`#ff007f`) and electric cyan (`#00f0ff`) contrast grid.
4. **Matrix Obsidian (`[data-theme="matrix"]`)**: Matte pure-black stealth palette with emerald accents (`#10b981`).

---

## 🔊 Curated Web Audio Haptic System (`HapticAudio`)

A tactile acoustic feedback engine built directly on the Web Audio API without external audio file dependencies. Synthesizes directional micro-transients for every user action:

| Action / Interaction | Synthesis Curve | Frequency / Duration | Aesthetic Sensation |
|---|---|---|---|
| **Expand (`expand`)** | Ascending exponential sine ramp | `540Hz → 880Hz` (`38ms`) | Mechanical switch unfolding / opening |
| **Collapse (`collapse`)** | Descending exponential sine ramp | `780Hz → 420Hz` (`38ms`) | Snappy mechanical tuck / folding shut |
| **Sidebar Open (`sidebarOpen`)** | Resonant ascending triangle glide | `440Hz → 760Hz` (`55ms`) | Smooth sliding panel reveal |
| **Sidebar Close (`sidebarClose`)** | Solid mechanical sine latch | `680Hz → 320Hz` (`45ms`) | Secure latch click |
| **Tactile Tap (`tap`)** | High-transient triangle pop | `920Hz → 560Hz` (`22ms`) | iOS-style segmented chip toggle |
| **Copy Code (`copySuccess`)** | Dual-tone arpeggiated chime | `C6 (1046Hz) + E6 (1318Hz)` (`90ms`) | Celebratory clipboard success chime |
| **Search Trigger (`searchOpen`)** | Holographic digital ping | `1180Hz → 1480Hz` (`40ms`) | Futuristic HUD activation |
| **Theme Change (`themeChange`)** | Cosmic frequency sweep | `600Hz → 1200Hz` (`60ms`) | Dimension / palette shift |
| **Theater Launch (`theaterOpen`)** | Low-frequency harmonic sine | `440Hz → 880Hz` (`70ms`) | Cinematic workbench focus |

### Acoustic Invariants:
1. **Directional Consistency**: Main content chapters, problem cards, nested accordions, and sidebar tree headers share the exact same `HapticAudio.expand()` and `HapticAudio.collapse()` curves.
2. **Batch Expansion Guard**: During "Expand All" and "Collapse All", `STATE.isBatchExpanding` strictly mutes individual accordion blips to eliminate audio distortion and DOM thread starvation.
3. **Audio Persistence**: State saved to `localStorage.getItem('polyturing_audio_enabled')`, defaulting to ON for tactile mobile-like delight. Toggled via HUD `🔊 / 🔇`.

---

## ↔️ Interactive Draggable Sidebar Resizer & Collapsible Rail Standard

Fixed-width sidebars degrade user experience by clipping titles on ultra-wide screens or consuming excessive reading space on medium displays. PolyTuring implements both continuous draggable resizing and full collapsible rail distraction-free mode:

```html
<header class="header-glass">
  <button class="sidebar-dock-toggle" id="sidebar-dock-toggle" aria-label="Toggle Navigation Sidebar" title="Toggle Sidebar (⌘B / [)">...</button>
</header>
<div class="layout-container">
  <button class="floating-sidebar-toggle" id="floating-sidebar-toggle" aria-label="Open Navigation Sidebar" title="Open Sidebar (⌘B / [)">...</button>
  <aside class="sidebar" id="main-sidebar">
    <div class="sidebar-search-row">
      <div class="sidebar-search-box">
        <span class="sidebar-search-icon">🔍</span>
        <input type="text" class="sidebar-search-input" id="sidebar-filter-input" placeholder="Filter architecture & chapters..." autocomplete="off">
        <button class="sidebar-filter-clear" id="sidebar-filter-clear" title="Clear filter (Esc)">✕</button>
      </div>
      <button class="sidebar-collapse-pill sidebar-close-btn" id="sidebar-close-btn" aria-label="Collapse navigation sidebar" title="Collapse sidebar (⌘B / [)">...</button>
    </div>
    ...
  </aside>
  <div class="sidebar-resizer" id="sidebar-resizer" role="separator" aria-orientation="vertical" aria-label="Resize navigation sidebar" title="Drag to adjust sidebar width (Double-click to reset)"></div>
  <main class="content-area">...</main>
</div>
```

### Operational Invariants for Sidebar:
1. **Collapsible Distraction-Free Reading Mode**:
   - Closed via `#sidebar-close-btn` inside sidebar or `#sidebar-dock-toggle` in header.
   - When collapsed, `body.sidebar-collapsed` shifts `.sidebar` by `margin-left: calc(-1 * var(--sidebar-width)); opacity: 0; visibility: hidden;` and hides `.sidebar-resizer`.
   - Floating reopen tab `#floating-sidebar-toggle` appears on the left edge (`top: calc(var(--topbar-height) + 14px); left: 14px;`).
   - Global keyboard shortcuts: `⌘B`, `Ctrl+B`, or `[` (when not focused in inputs) toggle the rail instantly.
   - State persisted in `localStorage.setItem('polyturing_sidebar_collapsed', 'true' / 'false')`.
2. **Real-Time Sidebar Filtering**:
   - Filter input `#sidebar-filter-input` filters all navigation groups and links instantly as the user types.
   - Automatically expands matching groups (`.expanded`) and hides non-matching items.
   - Pressing `Escape` or clicking `✕` clears the filter and restores previous tree state.
3. **Bounds Clamping**: Resizing strictly clamped between `MIN_WIDTH = 240px` and `MAX_WIDTH = 560px`.
4. **Dragging Performance**: During drag, `body.resizing-sidebar` sets `cursor: col-resize !important; user-select: none !important;`.
5. **Persistence**: The chosen width is persisted to `localStorage.setItem('polyturing_sidebar_width', width)` on pointer up.
6. **Double-Click Reset**: Double-clicking resets `--sidebar-width` to the default optimal width (`320px`).
7. **The Zero-Truncation Invariant**: Never clip titles with single-line ellipses (`white-space: nowrap; text-overflow: ellipsis;`). All navigation titles wrap cleanly:
   ```css
   .tree-title-text {
     white-space: normal;
     word-break: break-word;
     line-height: 1.35;
     font-size: 0.85rem;
   }
   ```

---

## 📑 Right-Side "On This Page" TOC HUD & 1:1 Synchrony Invariant

The right-side Table of Contents provides an interactive, live-updating outline that strictly mirrors the main content:

```html
<aside class="toc-sidebar" id="toc-sidebar" aria-label="Table of Contents">
  <div class="toc-header-hud">
    <div class="toc-header-title-wrap">
      <span class="toc-header-icon">📑</span>
      <span class="toc-title">On This Page</span>
    </div>
    <div class="toc-header-actions">
      <span class="toc-reading-progress" id="toc-reading-progress">0%</span>
      <button class="toc-jump-top-btn" id="toc-jump-top" title="Back to Top (Home)">⌃</button>
    </div>
  </div>
  <div class="toc-progress-track">
    <div class="toc-progress-indicator" id="toc-progress-indicator"></div>
  </div>
  <nav class="toc-nav" id="toc-nav">
    <div class="toc-chapter-group" data-chapter="ch0">
      <div class="toc-chapter-heading">00 · Manifesto</div>
      <ul class="toc-list">
        <li class="toc-item"><a href="#ch0-manifesto" class="toc-link active"><span class="toc-bullet"></span>Platform Manifesto</a></li>
        ...
      </ul>
    </div>
    ...
  </nav>
</aside>
```

### Operational Invariants for TOC:
1. **1:1 Exact Heading Alignment**: Every TOC label must 100% align with the corresponding section header in `<main>` (e.g. `sec-sdd-governance` must never be omitted or mislabeled).
2. **Dynamic Reading Progress**: HUD `#toc-reading-progress` and `#toc-progress-indicator` track document scroll depth in real time (`0%` to `100%`).
3. **Bi-Directional Synchrony**:
   - ScrollSpy tracks **ALL** unique target anchors referenced across Left Nav and Right TOC.
   - Highlights the active section with `.active` and lights up the parent chapter with `.active-chapter`.
   - Never mutates open/collapsed state during passive scrolling.
4. **Smooth Anchor Interception**:
   - Clicking any TOC or Left Nav link intercepts native jump, automatically uncollapses parent chapter in `<main>`, scrolls with an 80px cushion below the sticky header, applies a 1.8s `.flash-focus` neon glow, and triggers `HapticAudio.tap()`.
5. **Back to Top Action**: `#toc-jump-top` smoothly scrolls to top and triggers `HapticAudio.collapse()`.

---

## 📱 Year 2500 Visionary Mobile Architecture (Commute Gadget Standard)

On viewports $\le 1024\text{px}$, the sidebar transforms into an off-canvas slide-out drawer:

```html
<header class="header-glass">
  <button class="mobile-nav-toggle" id="mobile-nav-toggle" aria-label="Open Navigation">...</button>
</header>
<div class="layout-container">
  <div class="mobile-drawer-backdrop" id="mobile-drawer-backdrop" aria-hidden="true"></div>
  <aside class="sidebar" id="main-sidebar">
    <div class="mobile-drawer-header">
      <div class="brand-section">...</div>
      <button class="mobile-drawer-close" id="mobile-drawer-close">✕</button>
    </div>
    ...
  </aside>
</div>
```

### Mobile Invariants:
1. **Stacking Context Hierarchy**: Always place `#mobile-drawer-backdrop` inside `.layout-container` right before `<aside class="sidebar">` to avoid trapped z-index bugs.
2. **Sticky Overflow Invariant**:
   - **NEVER use `overflow-x: hidden` on `html` or `body`** when sticky headers or sticky sidebars exist. In WebKit/Blink, `overflow-x: hidden` creates an implicit scroll container that breaks `position: sticky`.
   - **Always use `overflow-x: clip;`** on `html, body, .layout-container`.
3. **Table & Formula Containment**:
   - Every `<table>` must be wrapped in `<div class="table-wrapper">` with `min-width: 520px` and `-webkit-overflow-scrolling: touch`.
   - KaTeX displays (`.katex-display`) must use `overflow-x: auto; max-width: 100%;`.

---

## 📊 Mermaid Architecture Diagrams & Fullscreen Visual Theater Workbench

### 1. Canvas DOM Anatomy
Every architecture diagram and complex table is wrapped in a zoomable canvas:

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

### 2. Fullscreen Theater Workbench (`<dialog id="theater-modal">`):
1. **Top-Layer `<dialog>` Isolation**: Uses HTML5 dialog `showModal()` to guarantee the visual workbench layers over all content with an authentic backdrop blur.
2. **Sub-Millisecond Zoom/Pan SLA**:
   - `.theater-surface` MUST specify `transition: none !important;` to eliminate pointer lag.
   - Panning uses `translate(${panX}px, ${panY}px) scale(${zoom})` for 1:1 pixel accuracy.
3. **Clear UX Distinction: `100%` vs `Fit`**:
   - `100%` (`#theater-zoom-reset`): Resets scale to **`1.0`** (true 1:1 native design scale).
   - `Fit` (`#theater-zoom-fit`): Dynamically calculates exact scale based on live viewport boundaries and SVG `viewBox` (`calculateFitScale`).
4. **SVG Text Selection**:
   - `.theater-surface svg text, foreignObject, span` have `user-select: text !important; cursor: text;` so engineers can highlight and copy symbol names directly.
5. **Safari WebKit Flex-SVG Sizing Invariant**:
   ```css
   .mermaid svg {
     display: block !important;
     width: 100% !important;
     max-width: 100% !important;
     height: auto !important;
     max-height: 580px;
     object-fit: contain;
     margin: 0 auto;
   }
   ```
6. **Node Label Escaping Rule**:
   - Escape `<` and `>` in labels as `≤` or `&lt;` / `&gt;` to avoid crashing Mermaid's HTML parser.

---

## 💻 Code Display Enhancements Standard

Interactive code containers feature frosted glass headers, semantic language pills, copy chimes, and theme adaptation:

```html
<div class="code-container">
  <div class="code-header">
    <div class="code-header-left">
      <span class="code-lang-tag">SURREALQL</span>
      <span class="code-file-path">src/db/schema.surql</span>
    </div>
    <button class="copy-btn" data-target="surql-indexes" title="Copy code snippet">
      <svg width="13" height="13" ...></svg>
      <span>Copy</span>
    </button>
  </div>
  <pre><code id="surql-indexes">...</code></pre>
</div>
```

- Clicking Copy copies code text, animates the button to `✓ Copied!`, and triggers `HapticAudio.copySuccess()`.
- Aero Mica theme enforces high-contrast dark text (`#0f172a`) on bright frosted card surfaces.

---

## 🧪 Automated Testing Protocol (14-Gate Audit)

Never push documentation portal changes without running the automated test suite!

```bash
# Run the all-in-one headless audit:
bash .agents/skills/docs-portal/scripts/test-docs.sh

# Or via the root Makefile target:
make docs-test
```

### The 14 Automated Verification Gates:
1. **Core Asset Presence**: Verifies `index.html`, `style.css`, and `app.js`.
2. **Mermaid Syntax Auditing**: Detects unescaped `<` or `>` characters in node labels.
3. **SVG Sizing Invariants**: Validates that CSS does not contain collapsing `width: auto !important` on `.mermaid svg`.
4. **KaTeX Integration**: Verifies stylesheet, scripts, and auto-render calls.
5. **Draggable Sidebar Resizer Invariants**: Audits `#sidebar-resizer`, `col-resize`, `initSidebarResizer`, and `localStorage` key.
6. **Sidebar Collapsible Rail**: Audits `#floating-sidebar-toggle`, `#sidebar-close-btn`, `body.sidebar-collapsed`, and `⌘B` shortcuts.
7. **Sidebar Real-Time Search Filter**: Audits `#sidebar-filter-input` and `initSidebarFilters`.
8. **Haptic Web Audio Feedback System**: Audits `HapticAudio`, `#audio-toggle-btn`, and `copySuccess` celebration chime.
9. **Multi-Theme Engine**: Audits `data-theme` support across `liquid`, `mica`, `holo`, `matrix`.
10. **Fullscreen Visual Theater Workbench**: Audits `#theater-modal`, `#theater-zoom-fit`, and `calculateFitScale`.
11. **Right-Side TOC Alignment & HUD**: Audits `#toc-sidebar`, `#toc-reading-progress`, `#toc-progress-indicator`, and `sec-sdd-governance`.
12. **Year 2500 Mobile Navigation Drawer & Anti-Clobbering**: Audits `#mobile-nav-toggle`, `#mobile-drawer-backdrop`, and sticky-safe `overflow-x: clip`.
13. **Table & Math Touch Overflow Containment**: Ensures all `<table>` tags are inside `.table-wrapper`, min-width is specified, and `.katex-display` has `overflow-x: auto`.
14. **Static Bundle Packaging & Output Integrity**: Compiles to `_site/` and checks size/file counts.

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

2. **Update Right TOC (`docs/site/index.html`)**:
   Add a matching `.toc-chapter-group`:
   ```html
   <div class="toc-chapter-group" data-chapter="ch5">
     <div class="toc-chapter-heading">05 · New Tool</div>
     <ul class="toc-list">
       <li class="toc-item"><a href="#ch5-new-tool" class="toc-link"><span class="toc-bullet"></span>System Overview</a></li>
       <li class="toc-item"><a href="#sec-new-arch" class="toc-link"><span class="toc-bullet"></span>Architecture Topology</a></li>
       <li class="toc-item"><a href="#sec-new-theory" class="toc-link"><span class="toc-bullet"></span>Theoretical Foundations</a></li>
       <li class="toc-item"><a href="#sec-new-usage" class="toc-link"><span class="toc-bullet"></span>Usage &amp; Quickstart</a></li>
     </ul>
   </div>
   ```

3. **Add Chapter Section (`docs/site/index.html`)**:
   Enclose all comparison tables in `<div class="table-wrapper">`, KaTeX in `$$ ... $$`, and diagrams in `.canvas-container`.

4. **Verify and Deploy**:
   ```bash
   make docs-test
   git commit -am "feat(docs): add Chapter 05 documentation"
   git push origin main
   ```
