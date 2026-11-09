# Changelog

All notable changes to this tool will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.2] - 2026-11-09

### Fixed
- **Sub-Pixel Zero-Drift Alignment Engine (R-067, AC-166..AC-169)**:
  - Fixed spacing anomaly where code lines in `<pre>` blocks had double-line spacing (43.6px per code line vs 22px per gutter number) caused by `<span class="code-line">` block elements combined with trailing `\n` in `.join('\n')`.
  - Joined code line spans with `''` (`.join('')`), normalized Highlight.js container styles (`padding: 0 !important; margin: 0 !important; background: transparent !important;`), and locked line heights to 22px.
  - Added dynamic line-wrap reconciliation (`syncWrappedLineHeights()`) that syncs line heights dynamically when soft-wrap is enabled.
  - Empirically verified across lines 1, 50, 100, 200, 399: $|y_{code} - y_{gutter}| = 0.00\text{px}$ across all viewports.
- **Agent Step Preview Expansion (R-068)**:
  - Expanded `write_file` agent step previews from 300 to 2,500 lines to support large file generation reviews without artificial truncation.

### Added
- **Forensic File Buffer Telemetry Capsule (R-068, AC-170..AC-173)**:
  - Mounted `#fileViewerBufferBadge` in File Viewer header displaying `● 100% On-Disk` alongside exact physical line count and formatted byte size.
  - Interactive tooltip confirms 10MB preview safety ceiling and provides transparent verification that no truncation occurred.
- **Multi-Hunk Navigation Radar & Spatial Dropdown (R-069, AC-174..AC-177)**:
  - Mounted `#hunkRadarSelect` quick-jump dropdown in the review subbar, listing all hunks with 1-indexed numbers, AST function names (`impl SymbolTable`, `func NewScanner`, etc.), and file line offsets.
  - Provides instantaneous 1-click navigation across files with 13+ hunks, bidirectionally synchronized with `<` `>` buttons and `⌥[` / `⌥]` keyboard shortcuts.
- **Holographic Minimap Tooltips & Gutter Hover Sync (R-070, AC-178..AC-181)**:
  - Hovering over minimap heatmap rail ticks displays `#minimapFloatingTooltip` with hunk number, line number, diff sign, and enclosing AST symbol.
  - Hovering over gutter line numbers highlights corresponding code lines via `.line-hover-sync`.
- **Active Hunk Spotlight Framing Rail (R-071, AC-182..AC-183)**:
  - Left file viewer frames lines belonging to the currently active hunk with an illuminated accent border rail (`.hunk-active-spotlight`).
- **One-Click Unified Patch Export & Clipboard Bridge (R-072, AC-184..AC-185)**:
  - Added 1-click patch export button to review header and individual hunks, copying Git-compatible unified diff blocks to the clipboard with tactile visual feedback.

## [0.5.1] - 2026-11-09

### Added
- **Diff Spatial Sync & Gutter Intelligence (R-061..R-066, AC-143..AC-165)**:
  - **GitHub-Standard Dual Line Number Gutter (R-061)**: Unified diff rows render dedicated `diff-gutter-old` (base line #) and `diff-gutter-new` (head line #) columns with tabular numeral alignment and blank states for additions/deletions.
  - **In-File Code Gutter Diff Decorations (R-062)**: Dynamically annotates line numbers in the left file viewer with emerald borders (`.gutter-diff-add`), cyan borders (`.gutter-diff-mod`), and rose insertion markers (`.gutter-diff-del`). Clicking any decorated line scrolls the review pane directly to that hunk.
  - **Holographic Diff Minimap Rail / Heatmap (R-063)**: 4px-10px vertical heatmap rail mounted alongside the file viewer scrollbar, rendering glowing percentage-based ticks across the entire file height with 1-click jump scrolling.
  - **Synchronized Hunk Stepper Navigation (R-064)**: Dual-pane stepper (`⏮ Prev Hunk` / `⏭ Next Hunk`, `⌥[` / `⌥]`, `F7` / `F8`) that simultaneously centers both the review hunk and the left file viewer line with an animated cosmic pulse (`jump-pulse`).
  - **AST Semantic Scope Breadcrumbs (R-065)**: Extracts trailing symbol signatures from hunk headers (`@@ ... @@ [symbol]`), rendering interactive capsules (`📦 struct Name`, `⚡ func Name`) on each hunk card.
  - **Surgical Hunk Patch Copy & Desktop IDE Deep-Linking (R-066)**: Individual `📋 Hunk` patch copy button on each hunk header and `↗ IDE` launcher generating `cursor://file/<path>:<line>` deep-links.

## [0.5.0] - 2026-11-09

### Added
- **Holographic Git & GitHub Review Pane (R-053..R-060, AC-117..AC-142)**:
  - **Native Git Trackability Discovery (R-053)**: Automatic repository root discovery (`findGitRoot`), branch detection, dirty file detection (`git status --porcelain`), and ambient GitHub CLI (`gh`) integration with zero custom credentials required.
  - **Multi-Vector Git Diff & History API (R-054)**: Backend REST endpoints `/api/git/file/status`, `/api/git/file/diff`, `/api/git/file/commits`, and `/api/sessions/{id}/file/steps`. Includes automatic `/dev/null` fallback for newly created/untracked files so all added lines render cleanly.
  - **Agent Step Chrono-Diff Engine (R-055)**: Reconstructs exact file mutations from transcript `edit_file` (Target vs Replacement) and `write_file` calls per step, paired with an Agent Thought Card displaying raw LLM reasoning context (`thinking`).
  - **Year 2500 Cosmic Holographic Review Pane UI (R-056)**: Integrated side-by-side review pane inside `#fileViewerModal` featuring Quantum Telemetry Ribbon (branch name, uncommitted indicator, churn counter `+A / -D`), vector tab switcher, and keyboard toggle shortcut (`⌘\`).
  - **Interactive Draggable Splitter & Physics-Based Geometry (R-057)**: Smooth drag resizer (`.file-viewer-resizer`) with 25%-80% bounds clamping, double-click snap points (50% -> 70% -> 35%), and persistent column width via `localStorage` (`se_review_split`).
  - **Dual-Mode Diff Engine with Sub-Line Token Highlighting (R-058)**: Unified inline and side-by-side split diff views with sub-line word/token diffing (`.diff-token-add`, `.diff-token-del`) calculated using common prefix/suffix token boundaries.
  - **Bidirectional Viewport Synchronization (R-059)**: Click any diff line in the review pane to automatically jump and highlight that exact line in the code viewer pane.
  - **Ambient Remote GitHub PR & Checks Integration (R-060)**: Leverages machine-level `gh pr view` to surface open pull requests, review decisions, and CI status checks directly in the review header.

## [0.4.0] - 2026-11-06

### Fixed
- **Timezone Boundary Skew & Local Time Offset Alignment (R-010, R-049)**:
  - Resolved timezone skew bug where evening sessions created after 00:00:00 UTC (e.g. 7:00 PM EST) fell onto the next UTC calendar day, causing them to be prematurely dropped by naive UTC midnight boundary checks in `Today` and `7 Days` filters.
  - Implemented dynamic client timezone offset transmission (`tz_offset`) via API query parameters and `resolveLocation(tzStr)` in Go backend (`time.FixedZone` with fallback to `time.Local`), ensuring date ranges evaluate full 24-hour windows in the user's local timezone.
  - Fixed client-side in-project date filtering in `getFilteredAndSortedProjSessions` by removing hardcoded UTC `Z` specifiers, aligning with local browser day boundaries.
  - Added `active` (`⚡ Last Touched`) as primary sort option to in-project toolbar dropdown and synchronized Date Grouped view partitioning.

### Added
- **Inclusive Lifetime Date Filtering Engine (R-049)**:

  - Date filtering naturally evaluates inclusive lifetime overlap (`lastActive >= from && createdAt <= to`), surfacing all sessions active or initiated in the window without cluttering the toolbar.
- **Session Trail Dual-Temporal Quantum Capsule (R-050)**:
  - Upgraded Session Trail detail header to `.detail-temporal-capsule` rendering `⚡ Active <relativeTime>` with neon pulse dot, `🌱 Initiated: <formattedDate>`, and dynamic `⏳ Multi-Day Span: Xd (<start> ➔ <end>)` pill.
- **Card-Level 1-Click UUID Copy & Layout Harmonization (R-047, R-051)**:
  - Added `.detail-session-capsule` to top-left of Session Trail header with 1-click UUID copy button (`📋 Copy`), visual `✓ Copied!` state, and toast feedback.
  - Added `.card-copy-btn` (`[📋 <uuid:8>]`) directly on each session card header on the main grid for frictionless 1-click UUID copying.
  - Compact single-line lifespan pill `⏳ Xd Span (Nov 4 ➔ Nov 6)` with CSS `white-space: nowrap; overflow: hidden; text-overflow: ellipsis;` preventing multi-line wrapping in 5-column layouts.
  - Standardized `.session-lifespan-row` with uniform minimum height (`1.6rem`), eliminating grid baseline jitter.
  - Enhanced `.prompt-snippet` with `line-height: 1.5`, `min-height: 4.8rem`, and glyph padding to prevent character cutoff.
- **Dual-Temporal Telemetry & Multi-Day Lifespan Tracking (R-048)**:
  - Added dual temporal badges to session cards (`⚡ Active <time>` + `🌱 Inception: <date>`).
  - Added sorting by `active` (`sort_by=active` / `sort_by=last_active`) as primary default sort across backend and frontend.
  - Enhanced Date Grouped view to partition by effective recency with in-view lens toggle.
- **Step Counter & Filter State Integrity (R-052)**:
  - Preserved `Showing X of Y steps` step subset telemetry when switching message type chips or refreshing Session Trail.
- **Universal Session Files & Artifacts Holographic Matrix (R-045)**:
  - Year 2500 panoramic widescreen command bridge (`width: min(1560px, 94vw); max-width: 1560px; max-height: 90vh;`) with liquid glass backdrop blur and glowing icon pulse.
  - 4 Forensic Lenses: `🗂️ Universal Catalog`, `✨ Artifacts & Deliverables`, `✏️ Codebase Mutations`, `📊 Churn & Blast Radius Telemetry`.
  - 5-Capsule Holographic Telemetry Ribbon: Deliverables count, Code Mutations, Inspected Files, Blast Radius, and Reconnaissance Ratio ($R_{\text{RW}}$).
  - Dynamic language and file extension spectrum breakdown bar with color-coded segments and telemetry cards.
  - EC2-style greedy shortest-column bin-packing spatial placement with zero vertical voids.
  - 1-click step navigation markers (`#X`) with smooth scroll and animated neon pulse highlight.
  - 1-click Open in Native File Viewer (`Open ↗`) and Quick Path Copy (`📋 Full` / `📋 Rel`).

- **Universal Multi-Dimensional Forensic Sorting Engine (R-046)**:
  - Interactive Year 2500 `.agent-matrix-sort-box` with holographic select dropdowns styled in aero-mica liquid glass.
  - **Files & Artifacts Sorting (7 Dimensions)**: Deliverable Priority (`type_priority`), Churn Intensity (`churn_desc`), Last Touched Recency (`last_touch`), Chronological Flow (`first_touch`), Read Volume (`reads_desc`), Alphabetical (`name_asc`), and File Extension (`ext_asc`).
  - **In-Session Agent Brain Matrix Sorting (5 Dimensions)**: Execution Frequency (`calls_desc`), Step Recency (`last_step`), Token Bloat (`bloat_desc`), Precedence Cascade (`precedence`), and Alphabetical (`name_asc`).
  - **Global Machine Agent Matrix Sorting (4 Dimensions)**: Token Bloat (`bloat_desc`), Line Count (`lines_desc`), Alphabetical (`name_asc`), and Scope Precedence (`scope`).
  - **Touch Span Telemetry Capsule**: Monospace capsule pill (`Span: #X ➔ #Y` or `Step #X`) on file cards for instantaneous forensic visibility of a file's active lifetime span across turns.
  - Zero-layout-shift bin-pack re-computation executed in < 12ms.

## [0.3.6] - 2026-11-05

### Added
- **Live Reactive Telemetry Synchronization & Session Trail Refresh (R-044)**:
  - Dynamic real-time synchronization of **Files & Artifacts** (`#detailFilesCountBadge`) and **Agent Context** (`#detailAgentContextCount`) during SSE live tailing (`/api/sessions/{id}/tail`) without requiring full browser reload.
  - In-place reactive DOM updates for open Files & Artifacts modal and Session Agent Brain Matrix modal upon arrival of new steps.
  - Debounced 350ms background synchronization of `/api/sessions/{id}/context` to prevent backend disk thrashing.
  - Tactile holographic counter badge pulse animations (`.holo-btn-counter.updated` and `.holo-btn-counter.files.updated`) with scale and cyan/purple acrylic glow transitions on count upticks.
  - Dedicated interactive Session Trail refresh button (`#refreshTrailBtn`) in `.tail-stream-control` with spinning animation (`@keyframes spinRefresh`) and toast notification feedback.
  - Keyboard shortcut <kbd>R</kbd> (in session detail view) for non-destructive dual-source trail and context reload.
  - Automatic catch-up telemetry refresh when toggling *Follow Live* or resuming paused live streams.
  - Verified OpenAPI 3.0 schema parity across all agent context and telemetry endpoints.

## [0.3.5] - 2026-11-04

### Added
- **Bin-Packing Spatial Density & Panoramic Masonry Matrix (R-043)**:
  - Replaced static CSS Grid horizontal row constraints with an EC2-style greedy shortest-column bin packing engine (`binPackCards`).
  - Greedily assigns each pallet card to $\arg\min_k \sum_{j \in B_k} h_j$, stacking cards vertically directly under previous cards with zero dead whitespace.
  - Expanded `#sessionContextModal` and `#agentMatrixModal` viewport width to panoramic widescreen real estate (`min(1560px, 94vw)`), seamlessly displaying 3 fluid columns across desktop viewports.
  - Built responsive multi-breakpoint column adapter: 3 columns for $\ge 1200\text{px}$, 2 columns for $768\text{px} \dots 1199\text{px}$, and 1 column for mobile viewports $< 768\text{px}$.
  - Added debounced window `resize` event listener dynamically recalculating column bins and re-laying out cards in real time.
  - Facelifted session trail navigation header counters next to "Agent Context" and "Files & Artifacts" into Year 2126 holographic glass action capsules (`.holo-header-btn`) with illuminated digital badges (`.holo-btn-counter.context` and `.holo-btn-counter.files`).
  - Fixed file aggregator in `updateSessionFilesBadge` to support uppercase `args.Path` and `args.FilePath`, correctly discovering all 30 touched files and artifacts in session `20dadec4`.

### Added
- **Empirical Skill Invocation Telemetry & Standby Decomposition (R-042)**:
  - Implemented 3-Tier Epistemic Skill Evidence Taxonomy distinguishing Tier 1 (Provable Execution via `read_file`/`view_file` on `SKILL.md` or `skills/.../scripts`), Tier 2 (Standby On-Demand via system prompt mount), and Tier 3 (Cognitive / Semantic Reference).
  - Enriched Go backend model `AgentSkill` with `last_step_idx`, `step_indices`, and `invocation_triggers`.
  - Added scope-based skill deduplication (`workspace` > `plugin` > `global` > `builtin`) preventing duplicate skill cards across inherited layers.
  - Upgraded `#sessionContextModal` with first-class Holographic Skill Cards rendering illuminated `⚡ Nx Active` badges, clickable `#X` step markers, first/last execution jumpers, and invocation trigger tags.
  - Dynamically labeled filter chip between `Invoked Skills (${activeCount})` in Active lens and `Skills (${activeCount} active / ${totalCount} total)` in Full Context Cascade lens.
  - Connected 1-click `📖 Inspect` to embedded Native File Viewer and `📋 Copy Path` to clipboard.

- **Command Execution & Tool Sub-Breakdown Spectrum (R-041)**:
  - Deconstructed generic execution tools (especially `run_command` and filesystem operations `read_file`/`write_file`/`edit_file`) into discrete operational components.
  - Implemented `parseCommandBinaryAndCategory`: splits chained pipelines (`&&`, `;`, `|`), strips environment assignments, and classifies binaries into 8 operational categories (`VCS`, `Build`, `Test`, `Process`, `Inspect`, `Network`, `Runtime`, `Filesystem`).
  - Added Year 2126 Holographic Command Spectrum bar (`.command-spectrum-bar`) inside `run_command` cards and Holographic Telemetry & Audit lens.
  - Built interactive Sub-Command Grid with category badges, execution counts, percentages, and clickable timeline step markers (`#38`, `#113`).
  - Implemented `parseFileExtensionAndCategory`: decomposes file operations by extension (`.go`, `.js`, `.css`, `.json`, `.md`, `.html`) and tracks top touched files.
  - Facelifted filter chips (`.session-matrix-filter-chips`) into a fluid, adaptive, non-scrollable flex matrix with color-coded illuminated count pill badges (`.matrix-chip-counter`).
  - Replaced all legacy `St ` step notation with universal `#X` format (`#38`, `#113`) across cards, pills, buttons, and toasts.

- **Universal Session Agent Brain & MCP Behavioral Matrix (R-040)**:
  - Extracted empirical MCP tool telemetry from session transcripts (lazy MCP `call_mcp_tool` and eager MCP `mcp_*`).
  - Auto-discovered ad-hoc and configured MCP servers (`omni-graph`, `playwright`), tracking invocation counts, active vs dormant tools, first/last step indices, and error rates.
  - Computed GenAI Tool Diversity Index ($D = U/N$) and Reliability Quotient in Session Tool Analytics.
  - Upgraded Session Context Modal (`#sessionContextModal`) to mirror the Global Agent Matrix with 3 holographic lenses (`Active Runtime Matrix`, `Full Context Cascade`, `Holographic Telemetry & Audit`), live multi-field search, and GenAI bloat telemetry ribbon.
  - Implemented 1-click step jumping with cyan neon pulse highlights in the conversation timeline.

### Fixed
- **Genesis Workspace Invariant & Command Touch Sanitization (R-028)**:
  - Fixed false-positive repository classification where a session started in the user home directory `~` (`/Users/aparv`) was claimed by `DSA` due to inline command string literals (`python3 -c "repos = ['.../DSA']"`).
  - Enforced the **Genesis Workspace Invariant**: sessions initiated in generic parent directories (`~`, `/Untitled`) without an authoritative Git repository strictly belong to `"Default"` unless substantive file mutations (`write_file`/`edit_file`) occur in an external repository.
  - Sanitized command execution path extraction (`extractCommandPaths`) to parse only explicit directory targets (`cd`, `git -C`, direct script paths) while skipping inline evaluation scripts (`-c`, `-e`) and print commands.
  - Refined internal runtime filtering (`isInternalRuntimeArtifact`) to target engine storage artifacts (`/brain/`, `/conversations/`, `/logs/`, `.system_generated`), removing blanket exclusions on user dotfiles (`statusline.py`).
  - Added visual distinction with dedicated `.git-badge.non-git` (`folder`) badges for unversioned workspaces and aggregated `"Default"` sessions into project metrics.

### Added
- **Universal Agent Context & Behavioral Matrix (R-039)**:
  - Added holistic observability across global instructions (`~/.gemini/config/rules/`, `~/.claude/rules/`), global skills (`~/.gemini/config/skills/`, builtin runbooks, plugins), MCP servers, and lifecycle hooks (`hooks.json`).
  - Added workspace-level discovery for repository guidelines (`.agents/rules/`, `AGENTS.md`, `CLAUDE.md`, `.cursorrules`, `copilot-instructions.md`), skills, and computed precedence cascade (Workspace > Declared > Global > Built-in).
  - Added in-session behavioral context analysis at `GET /api/sessions/:id/context` distinguishing between **⚡ Actively Invoked Skills** (with invocation counts and step markers) and **💤 On-Demand Skills** (progressive disclosure).
  - Built liquid-glass **Agent Brain & Context Matrix Modal** (`#agentMatrixModal`) accessible via top header action `🧠 Agent Brain` and Project Hub (`#projectAgentContextBtn`).
  - Built **In-Session Agent Context Drawer** (`#sessionContextModal`) launched via sticky detail header pill `🧠 Agent Context (<count>)`.
  - Added **GenAI Token & Word Bloat Telemetry (AC-045)**: computes line counts, word counts, character counts, and estimated tokens ($T \approx \max(\lceil C/4 \rceil, \lceil W \times 1.33 \rceil)$) for every rule and skill; renders dynamic bloat telemetry ribbons distinguishing pre-loaded context from on-demand footprint.
  - Seamlessly reuses embedded Native File Viewer (R-038) for 1-click high-fidelity Markdown and code inspection with syntax highlighting and line numbers.

### Fixed
- **Multi-Tier Intelligent Workspace & Codebase Clustering Engine (R-028, R-005)**:
  - Eliminated false-positive workspace clustering collisions where macOS OneDrive path segments (e.g. `OneDrive-Personal`) caused sessions to be erroneously classified under "Personal".
  - Replaced naive keyword counting with a 5-tier resolution hierarchy:
    1. **Tier 0 (Authoritative DB Metadata)**: Probes `<storage_dir>/../conversations/<id>.db` table `trajectory_metadata_blob` (`id='main'`) to extract authoritative `WorkspaceURI`, `GitRepo`, `GitURL`, and `GitBranch`.
    2. **Tier 1 (Empirical Tool Touch Matrix)**: Parses active tool parameters (`Path`, `CommandLine`, `Arguments`, `TargetFile`, `AbsolutePath`) with weighted scoring (Mutations `write_file`/`edit_file` = 50 pts, Execution `run_command` = 25 pts, Inspections `read_file` = 15 pts) and Git repository boundary root-finding.
    3. **Tier 2 (System Prompt Metadata)**: Extracts explicit workspace paths and working directories from system prompts (`Active Workspaces`, `Command Working Directory`).
    4. **Tier 3 (Non-Git Folder Fallback)**: Gracefully falls back to the enclosing directory name for ad-hoc scripts or unversioned folders, skipping generic system path segments (`Users`, `home`, `workspace`, `src`).
    5. **Tier 4 (Default Graceful Degradation)**: Falls back safely to `"Default"`.
  - Added strict workspace path sanitization (`sanitizeWorkspacePath`) to prevent global filesystem substring collisions across all filtering and WIP recovery operations.
- **Unified Markdown Document Hierarchy & Reading Lane**:
  - Replaced isolated timeline-only `.message-body` markdown CSS rules with a shared typography system covering `.markdown-body`, `.message-body`, and `.file-viewer-markdown`.
  - Implemented an editorial-grade typographic scale (`h1` at 1.85rem with glass divider, `h2` at 1.45rem, `h3` at 1.22rem in cyan, `h4` at 1.05rem, `h5/h6` uppercase tracked at 0.88rem) with proportional spacing.
  - Standardized paragraph margins (`1.05rem`), nested list spacing (`1.75rem`), frosted acrylic blockquotes, and table border matrices.
  - Applied comfortable reading lane constraints to `.file-viewer-markdown` (`max-width: 960px; margin: 0 auto;`) with responsive mobile padding (`1.5rem 1rem`), preventing line stretching on ultrawide monitors and delivering 100% visual and structural parity between the session trail and native file viewer.
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
