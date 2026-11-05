---
name: docs-portal
description: >-
  Authoritative engineering guide and automated testing protocol for the PolyTuring Living Documentation Portal. Covers Year 2500 cosmic liquid-glass design system, 4-theme engine (Liquid Glass, Fluent Aero Mica, Holo Cyberpunk, Matrix Obsidian), Haptic Web Audio feedback micro-transients, interactive draggable sidebar resizer and collapsible rail (⌘B / [), live sidebar search filter, clean-canvas single-source-of-truth navigation architecture, ambient top reading progress rail, floating glassmorphic Back-to-Top FAB, commute gadget mobile drawer, zero-truncation titles, touch-overflow containment, Mermaid SVG layout rules, fullscreen visual theater workbench with dynamic viewport fit and SVG text selection, KaTeX math typesetting, interactive simulators, and the 14-gate automated headless audit suite.
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
│   ├── style.css                          # Liquid glass tokens, 4-theme engine, resizer, collapsible rail, clean canvas
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

### Color Palette Tokens & 4-Theme Token Matrix

| Token | CSS Variable | Liquid Glass | Fluent Aero Mica | Holo Cyberpunk | Matrix Obsidian | Purpose |
|---|---|---|---|---|---|---|
| **Canvas Space** | `--bg-space` | `#050813` | `#f6f4ee` | `#03020a` | `#000000` | Deep root page backdrop |
| **Canvas Deep Base** | `--bg-deep` | `#050813` | `#f6f4ee` | `#03020a` | `#000000` | Deep canvas fill |
| **Subtle Container** | `--bg-subtle` | `#0c1224` | `#eae6db` | `#0c0822` | `#031407` | Topbars, table headers, widget items |
| **Surface Translucent**| `--bg-surface` | `rgba(15, 23, 42, 0.65)` | `rgba(255, 255, 255, 0.88)` | `rgba(15, 10, 35, 0.8)` | `rgba(4, 20, 10, 0.8)` | Floating panels, sidebars, header glass |
| **Surface Hover** | `--bg-surface-hover` | `rgba(30, 41, 59, 0.75)` | `rgba(255, 255, 255, 0.98)` | `rgba(24, 16, 52, 0.92)`| `rgba(6, 30, 15, 0.92)` | Interactive item hover state |
| **Card Surface** | `--bg-card` | `rgba(13, 20, 39, 0.65)` | `rgba(255, 255, 255, 0.82)` | `rgba(15, 10, 35, 0.78)`| `rgba(4, 20, 10, 0.78)` | Chapter sections, cards, widgets |
| **Card Border** | `--bg-card-border` | `rgba(255, 255, 255, 0.08)` | `rgba(30, 41, 59, 0.12)` | `rgba(0, 240, 255, 0.25)` | `rgba(16, 185, 129, 0.26)`| Subtle glass card edge divider |
| **Hover Border Glow** | `--bg-card-border-glow`| `rgba(0, 242, 254, 0.35)` | `rgba(2, 132, 199, 0.45)` | `#ff007f` | `#10b981` | Interactive element border focus |
| **Cyan Accent** | `--accent-cyan` | `#00f2fe` | `#0284c7` | `#00f0ff` | `#34d399` | Primary focus, metrics, active indicators |
| **Violet Accent** | `--accent-violet` | `#a855f7` | `#7c3aed` | `#ff007f` | `#059669` | Relational links, secondary badges |
| **Emerald Accent** | `--accent-emerald` | `#10b981` | `#059669` | `#00ff88` | `#10b981` | Passing SLA badges, healthy statuses |
| **Amber Accent** | `--accent-amber` | `#f59e0b` | `#d97706` | `#ffaa00` | `#f59e0b` | Cautionary callouts, invariant warnings |
| **Rose Accent** | `--accent-rose` | `#f43f5e` | `#e11d48` | `#ff0055` | `#ef4444` | SLA breaches, fatal alerts |
| **Primary Text** | `--text-main` | `#f1f5f9` | `#0f172a` | `#d8f4ff` | `#e6fced` | High-contrast body text and titles |
| **Muted Text** | `--text-muted` | `#94a3b8` | `#475569` | `#9ac8dd` | `#86efac` | Subtitles, descriptions, captions |
| **Dim Text** | `--text-dim` | `#64748b` | `#64748b` | `#628ea4` | `#4ade80` | Footnotes, metadata |
| **Code Surface** | `--code-bg` | `rgba(6, 10, 20, 0.9)` | `#090e1a` | `rgba(7, 4, 22, 0.95)` | `rgba(2, 10, 4, 0.96)` | Sleek terminal background for syntax |
| **Code Text** | `--code-text` | `#f1f5f9` | `#f1f5f9` | `#d8f4ff` | `#e6fced` | Syntax token foreground |
| **Badge Background** | `--badge-bg` | `rgba(0, 242, 254, 0.12)` | `rgba(2, 132, 199, 0.1)` | `rgba(255, 170, 0, 0.14)`| `rgba(16, 185, 129, 0.15)`| Pill tags, category chips |
| **Badge Text** | `--badge-text` | `#00f2fe` | `#0369a1` | `#ffaa00` | `#34d399` | Chip foreground label |
| **Modal Glass** | `--modal-bg` | `rgba(8, 12, 22, 0.95)` | `rgba(255, 255, 255, 0.98)` | `rgba(10, 7, 24, 0.98)` | `rgba(2, 8, 4, 0.98)` | Fullscreen theater and search dialogs |
| **Grid Lines** | `--grid-line` | `rgba(255, 255, 255, 0.02)`| `rgba(0, 0, 0, 0.04)` | `rgba(0, 240, 255, 0.03)`| `rgba(16, 185, 129, 0.03)`| Ambient background perspective mesh |

### Theme Invariants:
1. **Liquid Glass (`[data-theme="liquid"]`)**: Default deep cosmic glass with vibrant cyan/violet edge glows.
2. **Fluent Aero Mica (`[data-theme="mica"]`)**:
   - **Zero Hardcoded Dark Styles**: NEVER hardcode `rgba(15, 23, 42, ...)` or `#fff` into component rules. Always inherit from `var(--...)`.
   - **High-Contrast Terminal Code Blocks**: Code blocks maintain a sleek deep slate terminal surface (`#090e1a`) with crisp `#f1f5f9` text for razor-sharp syntax highlighting contrast.
   - **Crisp Frosted Paper Surfaces**: Cards, sidebar, and headers use `rgba(255, 255, 255, 0.82)` with dark slate text (`#0f172a`), muted text (`#475569`), and subtle borders (`rgba(30, 41, 59, 0.12)`).
3. **Holo Cyberpunk Neon (`[data-theme="holo"]`)**: High-saturation hot-pink (`#ff007f`) and electric cyan (`#00f0ff`) contrast grid.
4. **Matrix Obsidian (`[data-theme="matrix"]`)**: Matte pure-black stealth palette (`#000000`) with emerald accents (`#10b981`).

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

## 📐 Clean Canvas Architecture, Ambient Reading Progress & Back-to-Top FAB

PolyTuring enforces a **Single Source of Navigation Truth** in the Left Navigation tree, deliberately eliminating redundant right-side TOC sidebars to maximize the reading stage:

```html
<!-- Ambient Top Reading Progress Bar (Pinned to Sticky Header) -->
<header class="header-glass">
  ...
  <div class="global-reading-progress-track">
    <div class="global-reading-progress" id="global-reading-progress"></div>
  </div>
</header>

<!-- Floating Glassmorphic Back-to-Top FAB (Pinned to Bottom-Right) -->
<button class="back-to-top-fab" id="back-to-top-fab" aria-label="Back to Top" title="Back to Top (Home)">
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
    <polyline points="18 15 12 9 6 15"></polyline>
  </svg>
</button>
```

### Operational Invariants for Canvas & Reading HUD:
1. **Zero Redundant Sidebars**: Never add secondary right-side TOC panes that duplicate the Left Navigation. In a dense single-page architecture, dual sidebars compress the central reading canvas on laptops from `1120px` down to `840px`, squeezing Mermaid diagrams and comparison tables.
2. **Ambient Header Reading Progress Bar**:
   - Pinned directly to the bottom edge of `.header-glass` (`position: absolute; bottom: 0; left: 0; width: 100%; height: 2.5px;`).
   - Linear gradient fill from `--accent-cyan` to `--accent-violet` tracking scroll progress `(window.scrollY / totalScrollable) * 100%`.
   - Consumes **0 pixels** of lateral reading width.
3. **Floating Glassmorphic Back-to-Top FAB**:
   - Pinned at `bottom: 32px; right: 32px;` with frosted glass backdrop (`rgba(13, 20, 39, 0.75)`, blur `16px`).
   - Automatically transitions to `.visible` (`opacity: 1; transform: translateY(0);`) when `window.scrollY > 300px`.
   - Smoothly scrolls to top on click and triggers `HapticAudio.play('tap')`.
4. **Dynamic Left Nav ScrollSpy**:
   - ScrollSpy monitors all anchor targets referenced in `.nav-tree-link`.
   - As the user reads through the document, the corresponding tree group automatically expands (`.expanded`), the active chapter header glows with `.active-branch`, and the active link receives `.active`.
   - The active tree group gently scrolls into view within the sidebar (`scrollIntoView({ block: 'nearest', behavior: 'smooth' })`).
5. **Fluid Content Stage**:
   - `.content-area` expands to `max-width: 1400px; margin: 0 auto; width: 100%;`.
   - When the sidebar is collapsed (`body.sidebar-collapsed`), `.content-area` receives `padding-left: 68px; padding-right: 68px;` providing an expansive, distraction-free reading sanctuary.

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
3. **Content Surface Scoping & Unconstrained Dimensioning (SVG, Image, Table)**:
   - **Scope Invariant**: When extracting content to display in `#theater-surface`, NEVER query `sourceCanvas.querySelector('svg')` at the container root, as it will inadvertently match the 13px `<svg>` icon inside the toolbar's `.fullscreen-btn`. Always isolate extraction strictly inside `sourceCanvas.querySelector('.canvas-content-surface')`.
   - **SVG ViewBox Dimensioning**: When cloning SVG into `#theater-surface`, extract `viewBox` width and height (`[minX, minY, vbW, vbH]`). Assign explicit unconstrained pixel dimensions:
     ```javascript
     cleanSvg.style.width = vbW + 'px';
     cleanSvg.style.height = vbH + 'px';
     cleanSvg.style.maxWidth = 'none';
     cleanSvg.style.maxHeight = 'none';
     ```
   - **Image Natural Dimensioning**: When cloning `<img>` elements into `#theater-surface`, force `cleanImg.loading = 'eager'`, extract `naturalWidth` and `naturalHeight`, assign explicit pixel dimensions (`cleanImg.style.width = natW + 'px'; cleanImg.style.maxWidth = 'none';`), and attach an `onload` recalibration handler if still loading.
   - **Table Glass Card Preservation**: When displaying `<table>` elements in theater mode, wrap in high-contrast frosted glass (`.theater-surface .table-wrapper`).
   - **Why this is mandatory**: Global CSS rules apply `.mermaid svg { width: 100% !important; max-width: 100% !important; }` and `.canvas-viewport img { max-width: 100%; }` to fit items within card containers. Without setting explicit pixel width and unconstrained `max-width: none;` on the cloned content, CSS forces the content to shrink to modal bounds, causing `scale(1.0)` and `Fit` to look identical! Explicit pixel dimensions unlock **true 1:1 crisp physical resolution** at 100%.
4. **Clear UX Distinction: `100%` vs `Fit`**:
   - `100%` (`#theater-zoom-reset`): Sets zoom scale to **`1.0`** (true native physical resolution). Panning allows exploring detailed clusters without pixel blur.
   - `Fit` (`#theater-zoom-fit`): Dynamically calculates optimal fit scale `Math.min(availW / targetW, availH / targetH)` across SVGs, images, and tables so the entire target fits cleanly inside viewport margins with zero clipping.
5. **Fluid Multiplicative Wheel Zoom**:
   - Trackpad and mouse wheel zooming uses multiplicative scaling (`factor = e.deltaY < 0 ? 1.15 : 0.87`) clamped between `0.15` and `5.0`.
6. **Scroll Position Grounding & Flash-Focus Animation**:
   - Opening theater preserves `savedScrollY = window.scrollY || window.pageYOffset;`.
   - Closing theater immediately restores `window.scrollTo({ top: savedScrollY, behavior: 'instant' })` and triggers `.flash-focus` on the source container to eliminate disorientation ("treasure hunt" bugs).
7. **SVG Text Selection**:
   - `.theater-surface svg text, foreignObject, span` have `user-select: text !important; cursor: text;` so engineers can highlight and copy symbol names directly.
8. **Safari WebKit Flex-SVG Sizing Invariant**:
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
9. **Node Label Escaping Rule**:
   - Escape `<` and `>` in labels as `≤` or `&lt;` / `&gt;` to avoid crashing Mermaid's HTML parser.
10. **Dual-Scope In-Card & Theater Mode Styling (Zero Disparity Invariant)**:
   - When an SVG is cloned into `#theater-surface` during Fullscreen Theater Mode, it is NO LONGER nested inside `.mermaid`.
   - Every CSS rule modifying node shapes, text, edges, markers, or labels MUST explicitly target both `.mermaid` and `.theater-surface` to guarantee zero visual disparity between in-card and theater workbench modes.
11. **The Dual Fill/Color SVG & HTML ForeignObject Contrast Invariant**:
   - Mermaid flowcharts with HTML labels render text inside `<foreignObject><div xmlns="..."><span class="nodeLabel">...</span></div></foreignObject>`.
   - SVG `fill: ...` only affects pure SVG `<text>` elements and is completely ignored by HTML `<div>` / `<span>` elements (which default to Mermaid's internal generated `color: #f1f5f9`).
   - To guarantee crystal-clear contrast and eliminate white-on-white bugs in light themes (Mica), CSS MUST enforce both `fill` and `color` simultaneously across SVG and foreignObject tags:
     ```css
     [data-theme="mica"] .mermaid svg text,
     [data-theme="mica"] .mermaid svg foreignObject div,
     [data-theme="mica"] .mermaid svg foreignObject span,
     [data-theme="mica"] .theater-surface svg text,
     [data-theme="mica"] .theater-surface svg foreignObject div,
     [data-theme="mica"] .theater-surface svg foreignObject span {
       fill: #0f172a !important;
       color: #0f172a !important;
     }
     ```

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

- Clicking Copy copies code text, animates the button to `✓ Copied!`, and triggers `HapticAudio.play('copySuccess')`.
- Aero Mica theme enforces high-contrast dark text (`#0f172a`) on bright frosted card surfaces.

---

## 🛡️ Future Static Page Design & Anti-Conflict Architectural Protocol

To avoid layout regressions, broken CSS inheritance, or conflicting sidebars in future edits, follow these non-negotiable architectural invariants:

### 1. The Single Navigation Hierarchy Invariant
- **Rule**: Never construct secondary right-side navigation sidebars. The Left Navigation tree is the single source of truth for the entire page.
- **Why**: Multi-page frameworks (Docusaurus) need in-page right TOCs because the left sidebar navigates between separate HTML documents. In PolyTuring, the portal is a monolithic single-page reference deck; duplicate TOCs halve the available reading width.

### 2. The DOM Layering & Stacking Invariant
All elements MUST follow this deterministic DOM order:
```html
<body>
  <!-- Layer 1: Global Sticky Glass Header (z-index: 1000) -->
  <header class="header-glass">
    ...
    <div class="global-reading-progress-track">
      <div class="global-reading-progress" id="global-reading-progress"></div>
    </div>
  </header>

  <!-- Layer 2: Modals & Fullscreen Workbenches (Top Layer Dialog / z-index: 2000) -->
  <div class="search-modal" id="search-modal">...</div>
  <dialog id="theater-modal" class="theater-modal">...</dialog>

  <!-- Layer 3: Main Layout Container (z-index: 10) -->
  <div class="layout-container">
    <button class="floating-sidebar-toggle" id="floating-sidebar-toggle">...</button>
    <div class="mobile-drawer-backdrop" id="mobile-drawer-backdrop"></div>
    <aside class="sidebar" id="main-sidebar">...</aside>
    <div class="sidebar-resizer" id="sidebar-resizer"></div>
    <main class="content-area">...</main>
  </div>

  <!-- Layer 4: Floating Action Controls (z-index: 999) -->
  <button class="back-to-top-fab" id="back-to-top-fab">...</button>

  <script src="app.js"></script>
</body>
```

### 3. The Responsive Containment Invariants
- **Overflow Safe Guard**: Always use `overflow-x: clip;` on `html, body, .layout-container`. Never use `overflow-x: hidden;` (breaks `position: sticky` in WebKit).
- **Table Isolation**: Never render a `<table>` directly in `.content-area`. Always wrap in `<div class="table-wrapper"><table>...</table></div>`.
- **Diagram Isolation**: Never render raw `<pre class="mermaid">` directly in content flow without `.canvas-container > .canvas-viewport > .canvas-content-surface`.
- **Math Overflow**: Always verify KaTeX block formulas render within `.katex-display` which has `overflow-x: auto; max-width: 100%;`.

### 4. The Audio Synthesis Invariant
- Never load `.mp3` or `.wav` sound files. All acoustic feedback must use `HapticAudio` synthesized via Web Audio oscillators.
- Always check `STATE.isBatchExpanding` during bulk operations to prevent acoustic clipping.

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
3. **SVG Sizing Invariants**: Validates responsive `width: 100% !important;` and no 0px collapsing.
4. **KaTeX Integration**: Verifies stylesheet, scripts, and auto-render calls.
5. **Draggable Sidebar Resizer Invariants**: Audits `#sidebar-resizer`, `col-resize`, `initSidebarResizer`, and `localStorage` key.
6. **Sidebar Collapsible Rail**: Audits `#floating-sidebar-toggle`, `#sidebar-close-btn`, `body.sidebar-collapsed`, and `⌘B` shortcuts.
7. **Sidebar Real-Time Search Filter**: Audits `#sidebar-filter-input` and `initSidebarFilters`.
8. **Haptic Web Audio Feedback System**: Audits `HapticAudio`, `#audio-toggle-btn`, and `copySuccess` celebration chime.
9. **Multi-Theme Engine**: Audits `data-theme` support across `liquid`, `mica`, `holo`, `matrix`.
10. **Fullscreen Visual Theater Workbench**: Audits `#theater-modal`, `#theater-zoom-fit`, and `calculateFitScale`.
11. **Clean Canvas Architecture, Ambient Reading Progress & Back-to-Top FAB**: Audits `#global-reading-progress`, `#back-to-top-fab`, ensures `#toc-sidebar` is absent, and verifies `sec-sdd-governance` in Left Nav.
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

2. **Add Chapter Section in `<main class="content-area">` (`docs/site/index.html`)**:
   Enclose all comparison tables in `<div class="table-wrapper">`, KaTeX in `$$ ... $$`, and diagrams in `.canvas-container`.

3. **Verify and Deploy**:
   ```bash
   make docs-test
   git commit -am "feat(docs): add Chapter 05 documentation"
   git push origin main
   ```
