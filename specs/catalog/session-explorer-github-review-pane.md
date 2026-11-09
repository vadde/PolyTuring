# 📐 Specification: Holographic GitHub & Agent Step Review Pane for Session Explorer

> **Document ID**: `SPEC-SE-005`  
> **Target Subsystem**: `session-explorer` (`src/api.go`, `src/types.go`, `src/git_service.go`, `src/web/index.html`, `src/web/src/styles.css`, `src/web/src/app.js`)  
> **Status**: `Draft / Ready for Review`  
> **Author**: Antigravity (Google DeepMind) & `@vadde`  
> **Target Release**: `v0.5.0`  
> **Date**: 2026-11-09  
> **Constitutional Precedence**: Governed by [🏛️ Superpower 101 Constitution](file:///Users/aparv/.gemini/config/rules/00-superpower-101-agent-constitution.md) and [Rule 04 SDD Workflow](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/.agents/rules/04-sdd-workflow.md).

---

## 1. Overview & Architectural Scope

This specification formalizes the requirements, interface contracts, data models, UI components, and testing strategies for the **Holographic GitHub & Agent Step Review Pane** inside Session Explorer.

The feature extends the Native In-Browser File & Artifact Viewer (R-038) by embedding an interactive, dynamically resizable review pane on the far right of `#fileViewerModal`. It unifies local Git version control, GitHub remote metadata (via machine-level `gh` CLI), and autonomous Agent Session Step Diffs into a cohesive developer verification cockpit.

---

## 2. Requirements & Acceptance Criteria (R-053 to R-060, AC-117 to AC-142)

### 2.1 Requirements Table

| Req ID | Title | Priority | Description | Target Module |
| :--- | :--- | :--- | :--- | :--- |
| **R-053** | Native Git Trackability Discovery & Ambient Identity Engine | Must | Determine if a file resides inside a Git repository using `findGitRoot`. Discover repository root, active branch, dirty state, remote origin URL, and ambient machine-level `gh` CLI availability and auth status without requiring extra credentials. | `src/git_service.go`, `src/api.go` |
| **R-054** | Multi-Vector Git File Diff & History REST API | Must | Implement `/api/git/file/status`, `/api/git/file/diff`, and `/api/git/file/commits` supporting Working Copy (unstaged), Staged Index, Commit Diffs, and file commit history with bounded timeouts (3s) and size ceilings (2MB). | `src/api.go`, `src/types.go` |
| **R-055** | Agent Step Chrono-Diff & Hunk Extraction Engine | Must | Reconstruct per-step code mutations from session transcripts (`edit_file` Target vs Replacement, `write_file` creations and overwrites) into structured diff hunks correlated with agent thoughts. | `src/api.go`, `src/scanner.go`, `src/web/src/app.js` |
| **R-056** | Year 2500 Cosmic Holographic Review Pane UI | Must | Render `#fileViewerReviewPane` on the far right of `#fileViewerModal` featuring quantum telemetry ribbon (branch, churn counts, PR pill), isomorphic 4-theme styling, and vector mode selector. | `src/web/index.html`, `src/web/src/styles.css` |
| **R-057** | Interactive Draggable Splitter & Responsive Geometry | Must | Draggable splitter (`.file-viewer-resizer`) with visual grip, mouse/touch drag, double-click snap points (50/50, 65/35, 35/65), local storage width persistence, and `⌘\` toggle shortcut. | `src/web/src/app.js`, `src/web/src/styles.css` |
| **R-058** | Dual-Mode Diff Engine (Split & Unified) with Token Highlighting | Must | Render side-by-side or unified diff views with syntax highlighting, intra-line word/token change spans, collapsible hunks, and line number gutters. | `src/web/src/app.js`, `src/web/src/styles.css` |
| **R-059** | Bidirectional Viewport Synchronization & Anchor Navigation | Should | Clicking a diff hunk or line highlights and smoothly scrolls the code viewer to that exact line; navigating code anchors updates the review pane view. | `src/web/src/app.js` |
| **R-060** | Ambient Remote GitHub PR & Checks Integration | Should | Query `gh pr view` to surface active PR number, title, state, review decisions, and CI checks with 1-click external navigation to GitHub. | `src/git_service.go`, `src/api.go`, `src/web/src/app.js` |

### 2.2 Acceptance Criteria Matrix

| AC ID | Target Req | Criteria Description | Verification Method |
| :--- | :--- | :--- | :--- |
| **AC-117** | R-053 | When `GET /api/git/file/status?path=<path>` is called for a git-tracked file, it returns `is_git: true`, `git_root`, `branch`, `is_dirty`, and `remote_url`. | Go unit test (`TestGitService_FileStatus`) |
| **AC-118** | R-053 | When called for a non-git file, it returns `is_git: false` without error, allowing the UI to gracefully switch to Agent Step Chrono-Diff mode. | Go unit test (`TestGitService_NonGitFile`) |
| **AC-119** | R-053 | The backend discovers machine-level `gh` CLI binary and checks authentication status (`gh auth status`) with zero user credentials required. | Go unit test (`TestGitService_GHDiscovery`) |
| **AC-120** | R-054 | `GET /api/git/file/diff?path=<path>&mode=unstaged` returns the unstaged working copy diff with lines added, deleted, and structured hunks. | Go unit test (`TestGitService_UnstagedDiff`) |
| **AC-121** | R-054 | `GET /api/git/file/diff?path=<path>&mode=staged` returns the staged git index diff. | Go unit test (`TestGitService_StagedDiff`) |
| **AC-122** | R-054 | `GET /api/git/file/diff?path=<path>&mode=commit&commit_sha=<sha>` returns the exact diff introduced by that commit for the file. | Go unit test (`TestGitService_CommitDiff`) |
| **AC-123** | R-054 | `GET /api/git/file/commits?path=<path>&limit=30` returns up to 30 chronological commits touching the file, each with SHA, short SHA, author, timestamp, and message. | Go unit test (`TestGitService_FileCommits`) |
| **AC-124** | R-054 | If a git command takes longer than 3.0 seconds, the context cancels and returns HTTP 504 Gateway Timeout without hanging the server. | Go unit test (`TestGitService_Timeout`) |
| **AC-125** | R-054 | Diff outputs exceeding 2MB are cleanly truncated with a notice capsule to preserve browser performance. | Go unit test (`TestGitService_SizeCeiling`) |
| **AC-126** | R-055 | For any session touching a file via `edit_file`, the system synthesizes a diff hunk where `Target` represents deleted lines (`-`) and `Replacement` represents added lines (`+`). | Go unit test / JS Unit test |
| **AC-127** | R-055 | For `write_file` tool calls, the system detects whether it was an overwrite or append and creates corresponding addition hunks. | Go unit test / JS Unit test |
| **AC-128** | R-055 | Each agent step diff displays the step index, timestamp, tool call name, and an expandable capsule containing the agent's internal thought (`thinking`). | Playwright browser audit |
| **AC-129** | R-055 | The UI provides a Step Timeline Scrubber allowing the user to step forward and backward across all agent turns touching that file. | Playwright browser audit |
| **AC-130** | R-056 | `#fileViewerReviewPane` mounts inside `.file-viewer-shell` to the right of `.file-viewer-body`. | Playwright browser audit |
| **AC-131** | R-056 | The Quantum Telemetry Ribbon displays current branch (`🌿 <branch>`), dirty status dot, net churn (`+A / -B`), and active vector mode pill. | Playwright browser audit |
| **AC-132** | R-056 | The Review Pane adheres to all 4 PolyTuring themes (Liquid Glass, Fluent Aero Mica, Holo Cyberpunk, Matrix Obsidian) with zero CSS contrast defects. | Playwright visual check |
| **AC-133** | R-057 | The draggable resizer (`.file-viewer-resizer`) allows dynamic horizontal resizing between 25% and 75% of viewport width with smooth 60fps tracking. | Playwright drag test |
| **AC-134** | R-057 | Double-clicking the resizer cycles through snap points: 50% $\to$ 65% $\to$ 35%. | Playwright click test |
| **AC-135** | R-057 | Resizer position persists in `localStorage` across page reloads and modal re-openings. | Playwright persistence test |
| **AC-136** | R-057 | Pressing `⌘\` (Mac) or `Ctrl+\` (Windows/Linux) or clicking the topbar review button toggles review pane collapse/expand. | Playwright keyboard test |
| **AC-137** | R-058 | User can switch between Side-by-Side (Split) and Unified (Inline) diff views via segmented control. | Playwright interaction test |
| **AC-138** | R-058 | Code lines in diff blocks inherit Highlight.js syntax highlighting matching the file extension. | Playwright syntax test |
| **AC-139** | R-058 | Modified lines highlight sub-line changed tokens/words using intra-line Levenshtein/character diffing. | Playwright inspection |
| **AC-140** | R-059 | Clicking a line number in the diff view scrolls and highlights the corresponding line in the left code viewer. | Playwright navigation test |
| **AC-141** | R-060 | When an active PR exists for the branch, the Review Pane renders a GitHub PR capsule showing PR number, title, review decision, and CI checks. | Playwright / API test |
| **AC-142** | R-060 | Clicking `↗ GitHub` opens the commit or PR directly on github.com in a new tab. | Playwright link test |

---

## 3. Data Models & API Contracts

### 3.1 Go Structs (`src/types.go`)

```go
// GitFileStatus represents the git status and remote metadata of a file.
type GitFileStatus struct {
	IsGit           bool        `json:"is_git"`
	GitRoot         string      `json:"git_root,omitempty"`
	RelativePath    string      `json:"relative_path,omitempty"`
	Branch          string      `json:"branch,omitempty"`
	IsDirty         bool        `json:"is_dirty"`
	HasStaged       bool        `json:"has_staged"`
	HasUnstaged     bool        `json:"has_unstaged"`
	RemoteURL       string      `json:"remote_url,omitempty"`
	GHAvailable     bool        `json:"gh_available"`
	GHAuthenticated bool        `json:"gh_authenticated"`
	ActivePR        *GitHubPR   `json:"active_pr,omitempty"`
	SessionSteps    int         `json:"session_steps"`
}

// GitHubPR represents pull request metadata obtained from gh cli.
type GitHubPR struct {
	Number         int    `json:"number"`
	Title          string `json:"title"`
	URL            string `json:"url"`
	State          string `json:"state"` // OPEN, CLOSED, MERGED
	BaseRef        string `json:"base_ref"`
	HeadRef        string `json:"head_ref"`
	ReviewDecision string `json:"review_decision,omitempty"` // APPROVED, CHANGES_REQUESTED
	StatusSummary  string `json:"status_summary,omitempty"`  // e.g. "12/12 passing"
}

// GitDiffResponse represents parsed diff hunks and statistics.
type GitDiffResponse struct {
	Path      string        `json:"path"`
	Mode      string        `json:"mode"` // unstaged, staged, commit, pr, step
	RawDiff   string        `json:"raw_diff"`
	Additions int           `json:"additions"`
	Deletions int           `json:"deletions"`
	Hunks     []DiffHunk    `json:"hunks"`
	Truncated bool          `json:"truncated"`
}

// DiffHunk represents an individual diff chunk with header and lines.
type DiffHunk struct {
	Header   string     `json:"header"` // @@ -10,6 +10,8 @@
	OldStart int        `json:"old_start"`
	OldLines int        `json:"old_lines"`
	NewStart int        `json:"new_start"`
	NewLines int        `json:"new_lines"`
	Lines    []DiffLine `json:"lines"`
}

// DiffLine represents a single line in a diff.
type DiffLine struct {
	Type       string `json:"type"` // "context", "add", "delete"
	Content    string `json:"content"`
	OldLineNum int    `json:"old_line_num,omitempty"`
	NewLineNum int    `json:"new_line_num,omitempty"`
}

// GitCommitEntry represents a single commit from git log.
type GitCommitEntry struct {
	SHA          string `json:"sha"`
	ShortSHA     string `json:"short_sha"`
	AuthorName   string `json:"author_name"`
	AuthorEmail  string `json:"author_email"`
	Timestamp    int64  `json:"timestamp"`
	RelativeTime string `json:"relative_time"`
	Subject      string `json:"subject"`
}

// AgentStepDiff represents a synthetic diff hunk created from an LLM step.
type AgentStepDiff struct {
	StepIndex    int        `json:"step_index"`
	ToolName     string     `json:"tool_name"`
	Timestamp    string     `json:"timestamp"`
	Thinking     string     `json:"thinking,omitempty"`
	UserPrompt   string     `json:"user_prompt,omitempty"`
	AgentSummary string     `json:"agent_summary,omitempty"`
	Hunks        []DiffHunk `json:"hunks"`
	Additions    int        `json:"additions"`
	Deletions    int        `json:"deletions"`
}
```

### 3.2 REST API Endpoints

#### 1. `GET /api/git/file/status?path=<path>&session_id=<id>`
- **Response**: `GitFileStatus` JSON object.
- **Errors**: `400 Bad Request` if path missing, `403 Forbidden` if path restricted.

#### 2. `GET /api/git/file/diff?path=<path>&mode=<unstaged|staged|commit|pr>&commit_sha=<sha>`
- **Response**: `GitDiffResponse` JSON object.
- **Parameters**:
  - `path`: Absolute or relative path to file.
  - `mode`: `unstaged` (default), `staged`, `commit`, or `pr`.
  - `commit_sha`: Required if `mode=commit`.

#### 3. `GET /api/git/file/commits?path=<path>&limit=30`
- **Response**: Array of `GitCommitEntry` objects.

#### 4. `GET /api/sessions/:id/file/steps?path=<path>`
- **Response**: Array of `AgentStepDiff` objects chronologically ordered by `step_index`.

---

## 4. UI/UX Component Specifications

### 4.1 DOM Tree Layout

```html
<!-- Inside .file-viewer-shell -->
<div class="file-viewer-topbar">
  <!-- Left: File info -->
  <!-- Middle: Meta badges -->
  <!-- Right: Actions + Review Pane Toggle Button -->
  <button class="btn-ghost-sm active" id="toggleReviewPaneBtn" title="Toggle Review Pane (⌘\)">
    <span class="btn-icon-glyph">🔀</span> Review Pane
    <span class="badge-mini" id="reviewPaneBadge">Git</span>
  </button>
</div>

<!-- Main Split Container -->
<div class="file-viewer-split-container" id="fileViewerSplitContainer">
  <!-- Left: Original File Viewer Body -->
  <div class="file-viewer-body" id="fileViewerBody">
    <!-- Code block, markdown, images, find widget -->
  </div>

  <!-- Draggable Resizer Bar -->
  <div class="file-viewer-resizer" id="fileViewerResizer" title="Drag to resize, double-click to snap">
    <div class="resizer-handle"></div>
  </div>

  <!-- Far Right: Holographic Review Pane -->
  <aside class="file-viewer-review-pane" id="fileViewerReviewPane">
    <!-- Quantum Telemetry Ribbon -->
    <div class="review-telemetry-ribbon">
      <div class="review-branch-capsule" id="reviewBranchCapsule">
        <span class="branch-icon">🌿</span>
        <span class="branch-name" id="reviewBranchName">main</span>
        <span class="dirty-indicator" id="reviewDirtyIndicator" title="Uncommitted changes">•</span>
      </div>
      <div class="review-churn-capsule" id="reviewChurnCapsule">
        <span class="churn-add" id="reviewChurnAdd">+0</span>
        <span class="churn-del" id="reviewChurnDel">-0</span>
      </div>
      <div class="review-pr-capsule" id="reviewPrCapsule" style="display: none;">
        <span class="pr-icon">🐙</span>
        <a class="pr-link" id="reviewPrLink" target="_blank" rel="noopener">PR #0</a>
      </div>
    </div>

    <!-- Review Vector Selector Tabs -->
    <div class="review-vector-tabs" role="tablist">
      <button class="review-tab-btn active" data-vector="steps" id="tabReviewSteps">
        ⚡ Agent Steps <span class="tab-count" id="stepsCountBadge">0</span>
      </button>
      <button class="review-tab-btn" data-vector="working" id="tabReviewWorking">
        📝 Working Copy
      </button>
      <button class="review-tab-btn" data-vector="staged" id="tabReviewStaged">
        📦 Staged
      </button>
      <button class="review-tab-btn" data-vector="commits" id="tabReviewCommits">
        📜 Commits
      </button>
    </div>

    <!-- Review Sub-Header & Controls -->
    <div class="review-subbar">
      <!-- Step Scrubber Controls (when vector == steps) -->
      <div class="review-step-scrubber" id="reviewStepScrubber">
        <button class="scrubber-btn" id="scrubberPrevBtn" title="Previous Step">◀</button>
        <span class="scrubber-label" id="scrubberLabel">Step 1 of 1</span>
        <button class="scrubber-btn" id="scrubberNextBtn" title="Next Step">▶</button>
      </div>

      <!-- Diff Mode Segmented Control: Unified vs Split -->
      <div class="segmented-control segmented-control-sm" id="diffViewModeControl">
        <button class="segment-btn active" data-diff-mode="unified" title="Unified Inline View">☰ Inline</button>
        <button class="segment-btn" data-diff-mode="split" title="Side-by-Side Split View">☵ Split</button>
      </div>

      <!-- Copy Diff Button -->
      <button class="btn-ghost-xs" id="copyDiffPatchBtn" title="Copy Diff as Patch">📋 Patch</button>
    </div>

    <!-- Agent Thinking Card (shown in steps vector) -->
    <div class="review-agent-thought-card" id="reviewAgentThoughtCard" style="display: none;">
      <div class="thought-card-header">
        <span class="thought-brain-icon">🧠</span>
        <span class="thought-title">Agent Rationale &amp; Goal</span>
        <button class="thought-toggle-btn" id="toggleThoughtBtn">▾</button>
      </div>
      <div class="thought-card-body" id="reviewThoughtBody"></div>
    </div>

    <!-- Commit Picker Dropdown (shown in commits vector) -->
    <div class="review-commit-picker" id="reviewCommitPicker" style="display: none;">
      <select id="reviewCommitSelect" class="holo-matrix-select">
        <option value="">Select a commit to view changes...</option>
      </select>
    </div>

    <!-- Diff Content Container -->
    <div class="review-diff-container" id="reviewDiffContainer">
      <!-- Loading State -->
      <div class="review-loading" id="reviewLoading">
        <div class="loading-spinner-sm"></div>
        <span>Analyzing repository diffs...</span>
      </div>

      <!-- Empty / Clean State -->
      <div class="review-empty-state" id="reviewEmptyState" style="display: none;">
        <span class="empty-icon">✓</span>
        <span class="empty-title">Working tree clean</span>
        <span class="empty-desc">No uncommitted modifications found in this file.</span>
      </div>

      <!-- Diff Hunk Surface -->
      <div class="review-diff-surface" id="reviewDiffSurface"></div>
    </div>
  </aside>
</div>
```

---

## 5. Resizer & Responsive Geometry Engine

### 5.1 Split Calculation Physics
- When resizing is active (`mousedown` on `.file-viewer-resizer`):
  $$\text{ContainerWidth} = \text{splitContainer.getBoundingClientRect().width}$$
  $$\text{PointerX} = \text{event.clientX} - \text{splitContainer.getBoundingClientRect().left}$$
  $$\text{LeftPercentage} = \text{clamp}\left(\frac{\text{PointerX}}{\text{ContainerWidth}} \times 100, 25\%, 75\%\right)$$
  $$\text{RightPercentage} = 100\% - \text{LeftPercentage}$$
- CSS Variables dynamically updated on the container:
  ```css
  --code-pane-width: 55%;
  --review-pane-width: 45%;
  ```
- **Snap Presets**:
  - Double clicking resizer cycles: $50\% \to 65\% \to 35\% \to 50\%$.
- **Storage Persistence**:
  `localStorage.setItem('session_explorer_review_pane_split', leftPercentage)`

---

## 6. Implementation & Verification Plan

### 6.1 Phased Execution Ladder (SDD)
1. **Phase 1: PRD & Spec Signoff** (Current Phase) — Align on contracts and design with user.
2. **Phase 2: Go Backend Service & Endpoints** — Create `src/git_service.go`, add endpoints in `src/api.go`, write Go unit tests in `src/git_test.go`.
3. **Phase 3: Frontend DOM & CSS Styling** — Implement `.file-viewer-split-container`, `.file-viewer-resizer`, `#fileViewerReviewPane`, 4-theme styling in `src/web/src/styles.css`.
4. **Phase 4: Diff Parser & Step Chrono-Diff Logic** — Implement unified/split diff parsing, token-level diffing, transcript step extraction in `src/web/src/app.js`.
5. **Phase 5: Resizer Physics & Event Wiring** — Resizer drag, double-click snap, `⌘\` shortcut, local storage persistence.
6. **Phase 6: Automated Verification** — Go test suites (`make test-tool T=session-explorer`) + Headless Playwright browser audit.
7. **Phase 7: Living Documentation & Handoff** — Update `CONTEXT.md`, `STATUS.md`, `DEVLOG.md`, and `README.md`.
