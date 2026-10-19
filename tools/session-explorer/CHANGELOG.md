# Changelog

All notable changes to this tool will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- **Code Viewer Flex Layout Clipping & Line Gutter Height Drift**:
  - Replaced constrained flex layout on `.file-viewer-body` with natural block container layout (`display: block; overflow: auto;`) and `.code-content-pre { overflow: visible; }`, resolving empty blank rendering below viewport height on long code files.
  - Standardized `.code-line-numbers` and `code` font metrics (`font-size: 0.84rem; line-height: 22px;`) to eliminate vertical alignment drift over multi-hundred-line files.
  - Converted preformatted line highlight marks (`mark.line-highlight`) from `display: block` to `display: inline` with styled background and shadow, eliminating duplicate phantom line-breaks caused by block elements before trailing `\n` characters in `<pre>` tags.
  - Fixed floating Find bar to stay anchored at `position: fixed` within viewer viewport during deep multi-page scrolling.

### Added
- **Bounded Diagram Viewports & Fullscreen Theater Mode (R-037)**:
  - Constrained inline `.canvas-viewport` containers to `max-height: 440px` with slim internal scrollbars, eliminating vertical scroll trapping in long conversational timelines.
  - Added Fullscreen Theater mode modal (`#diagramTheaterModal`) with 20% to 500% zoom controls, click-and-drag 2D panning, Fit to View, Copy SVG, and Copy Diagram Source.
  - Implemented double-click to expand on diagram canvas viewports, seamlessly launching Theater mode.
- **Native In-Browser File & Artifact Viewer (R-038)**:
  - Added secure Go endpoint `GET /api/file?path=<path>` supporting absolute and relative paths with workspace auto-resolution, credential path protection (`.ssh`, `.gnupg`, `.aws`, `/etc/shadow`), 10MB file ceiling, and content-aware metadata detection.
  - Built frosted glass `#fileViewerModal` featuring detected language badge, line count, human-readable size, and breadcrumb path copy.
  - Added segmented toggle for Markdown files switching between rich HTML Preview (with KaTeX LaTeX math and client-side Mermaid rendering) and syntax-highlighted Raw mode, with 100% full-width real estate utilization and strict view separation.
  - Added `.code-line-numbers` gutter and automatic line anchor scrolling and highlighting for paths specifying line ranges (`#L10-L20` or `:10`).
  - Added embedded preview for image files (`.png`, `.jpg`, `.svg`, `.webp`).
  - Unified all tool call file parameters (`AbsolutePath`, `TargetFile`, `path`) and markdown file links (`[text](file://...)`) into interactive clickable `.tool-file-pill` chips with dedicated 1-click copy path buttons (`📋` → `✓`).
  - Added IDE-grade In-File Find widget (`#fileViewerFindWidget`) with `Cmd+F` / `Ctrl+F` shortcut, match counter (`X of Y`), case-sensitivity toggle (`Aa`), next/previous navigation (`Enter`/`Shift+Enter`, `▲`/`▼`), smooth scrolling to matches, and non-destructive syntax-preserving tree highlights.
  - Enforced single unified scrollbar architecture on File Viewer, eliminating nested dual vertical scrollbars.
  - Added instant micro-feedback with color glow transitions to all viewer actions (`📋 Copy Content`, `🔗 Copy Path`, `↩ Wrap`).
  - Added top action button `📁 Files & Artifacts (count)` and comprehensive session files aggregator drawer (`#sessionFilesModal`) categorizing all session files into `✨ Artifact`, `✏️ Modified`, and `👁️ Read`.
- **Universal Multi-Runtime Discovery & Provider Architecture (R-036)**:
  - Decoupled scanner from a single hardcoded IDE data directory; auto-discovers Antigravity IDE (`~/.gemini/antigravity-ide/brain`), Antigravity CLI (`~/.gemini/antigravity-cli/brain`), and Core runtime stores concurrently.
  - Dynamically routes session detail retrieval, transcript search, and real-time SSE live tailing to each session's originating brain path (`session.TranscriptPath`, `session.StorageDir`).
  - Added CLI metadata parsing for `Active Workspaces:` and `Command Working Directory:`, instantly extracting proper repository context (`tool-scripts`) for CLI sessions.
  - Added visual runtime badges (`⚡ CLI`, `💻 IDE`, `⚙️ Core`, `🤖 Claude`, `🐙 Copilot`) on session cards and sticky detail navigation headers.
  - Added Runtime segmented control filter in the dashboard toolbar (`All`, `💻 IDE`, `⚡ CLI`) with instant backend query parameter support (`/api/sessions?runtime=cli`).
  - Future-proof provider architecture ready for seamless expansion to Claude Code and GitHub Copilot.
- **Granular Message Type & Dialogue Filtering (R-035)**: Separated pure conversational agent replies (`AGENT_RESPONSES`) from mechanical tool execution cycles (`TOOL_CALLS`). Introduced a dedicated **`💬 Dialogue`** mode (User Prompts + Agent prose) for a clean, noise-free two-way conversation read, backed by symmetric Go search filtering and real-time SSE live tailing sync.
- **Real-Time Live Tailing via Server-Sent Events (SSE)**: Added `/api/sessions/{id}/tail` endpoint featuring byte-offset delta seeking on JSONL logs, 15s keepalive heartbeats, and sub-second push updates for active IDE coding sessions.
- **Dynamic Watch Interval Coordination**: Added `GET` and `POST` `/api/watch-interval` to synchronize UI refresh settings with the Go background ticker, supporting active live intervals or full pauses.
- **Keyed DOM Reconciliation Engine**: Replaced full-DOM redraws with an in-memory `renderedStepMap` cache, enabling fine-grained $O(\Delta)$ appends without tearing down past DOM nodes or resetting interactive Mermaid diagrams and KaTeX formulas.
- **Client-Side Hash Router & Deep-Link Route Persistence**: Added hash-based routing (`#/session/:id`, `#/project/:name`, `#/`) to guarantee persistent navigation states across browser reloads (`Cmd+R` / `F5`) and native history navigation.
- **De Facto "Latest First" Sorting Default**: Defaulted in-session step sorting to `step_desc` (Latest First) for both session opening and live tailing, prepending new steps to the top of the viewport.
- **Interactive Live Stream Play/Pause Toggle**: Converted the session stream pill into an interactive toggle button (`🟢 Live Stream` ⇄ `⏸️ Stream Paused`) with instant connection control.
- **Smart Auto-Follow Mode**: Preserved pinned scroll during live tailing while automatically unpinning when the user scrolls away to inspect earlier steps, displaying a floating direction-aware indicator (`↑ New steps arrived`).
- **Scanner Delta Caching**: Implemented `mtime` and file-size delta tracking in `scanner.go`, slashing incremental rescan times from ~950µs to ~50µs (~19x speedup).
- **Multi-Lens Markdown & LaTeX Math Rendering**: Integrated `katex@0.19.0` with offline embedded fonts and custom Marked tokenizer extensions (`mathBlockExt`, `inlineMathExt`), rendering display math (`$$...$$`, `\[...\]`) and inline math (`$...$`, `\(...\)`) with pre-tokenizer protection against underscore-to-italic corruption.
- **`isLikelyMath` Heuristic**: Intelligent discrimination between real mathematical expressions and currency or price ranges (`$100`, `$25.50`, `$10 - $20`).
- **Interactive Client-Side Mermaid Diagrams**: Integrated `mermaid@12.2.0` with dark acrylic glass theme, interactive "Source" toggle, "Copy Code" action, and graceful error recovery.
- **Interactive File Links**: Converted `file:///` and absolute filesystem paths into interactive pills with document badges (`📄`), copying normalized paths with line numbers (e.g. `file:2950`) to clipboard for immediate editor/terminal use.
- **Responsive Table Containers**: Wrapped Markdown tables in `.table-responsive` with custom slim glass scrollbars to prevent overflow.
- **GitHub Alert Callouts**: Styled `> [!NOTE]`, `> [!TIP]`, `> [!IMPORTANT]`, `> [!WARNING]`, `> [!CAUTION]` into colored acrylic callout banners.
- **Enhanced Code Blocks**: Added language header badges (`GO`, `PYTHON`, `JSON`, `TELEMETRY`) and one-click copy buttons to all code blocks.

### Fixed
- Fixed critical frontend initialization failure on initial landing where `initLiveWatch` was nested inside `initEventListeners()`, causing an unhandled `ReferenceError: initLiveWatch is not defined` that blocked `loadSessions()` from running.
- Defaulted initial date filter to Last 7 Days (`7d`) with pre-filled inputs and active chip highlight, eliminating the stuck "Loading..." badge on landing and allowing smooth toggling to "All Time" and custom dates.
- Initialized `groupBy` in app state from `localStorage` to ensure consistent active state on the segmented control.
- Re-synchronized date preset chip active states dynamically upon programmatic preset switches and manual input clearance.
- Fixed critical workspace accumulation bug where `idx.stats` was not reinitialized during background rescans, causing workspaces list to balloon to 3,000+ entries with duplicates.
- Fixed workspace path leakage where filenames (`all_mermaids.txt`), usernames (`aparv`), and internal subfolders (`_templates`, `docs`, `scripts`, `rules`, `analytics`) were mistakenly recognized as workspaces.
- Added strict `isValidWorkspaceName` filter rejecting non-workspace paths, files with dots, and system directories (`library`, `applications`, `system`, `volumes`).
- Added client-side Set deduplication and change detection in `renderWorkspaceOptions` to eliminate UI glitches and duplicate options.
- Added unit tests `TestRescanStatsDeduplication` and `TestExtractProjectName_Sanitization`.

### Added
- Initial tool scaffolding
