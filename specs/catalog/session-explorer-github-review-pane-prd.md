# 📋 PRD: Session Explorer — Holographic GitHub & Agent Step Review Pane

> **Document ID**: `PRD-SE-005`  
> **Target Subsystem**: `session-explorer` (Go REST API backend, Single-Page Application web UI, Native File Viewer Modal)  
> **Status**: `Draft / Ready for Review`  
> **Author**: Antigravity (Google DeepMind) & `@vadde`  
> **Target Release**: `v0.5.0`  
> **Date**: 2026-11-09  
> **Constitutional Precedence**: Strictly governed by [🏛️ Superpower 101 Constitution](file:///Users/aparv/.gemini/config/rules/00-superpower-101-agent-constitution.md) and [Rule 04 SDD Workflow](file:///Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/PolyTuring/.agents/rules/04-sdd-workflow.md).

---

## 1. Executive Summary & Vision

### 1.1 The Modern Developer Paradigm Shift
In the contemporary GenAI era of software engineering, code is no longer written solely line-by-line by humans. Autonomous coding agents (such as Antigravity, Claude Code, Cursor, and multi-agent swarms) execute dozens of tool mutations—creating files, applying surgical replacements, refactoring functions, and executing build scripts—in rapid succession.

Modern developers and software architects face a profound verification crisis: **static before-and-after snapshots (such as a standard single-turn Git diff or an uncontextualized code viewer) obscure the causal journey of *how* and *why* code changed**. 

When examining a file in Session Explorer—whether drilling down from a session trail message, browsing the Session Files & Artifacts Matrix (R-045), or inspecting agent context (R-039)—an engineer requires an immediate, forensic understanding of changes across **two orthogonal axes**:
1. **The Version Control Axis (Git & GitHub)**: What unstaged modifications exist right now? What is staged? How has this file evolved across recent commits? Is there an active GitHub Pull Request, and how does this file relate to the PR baseline?
2. **The Autonomous Agent Epistemic Axis (Session Step Chrono-Diff)**: How did the agent mutate this file during this exact session? What was deleted (`Target`) and what was added (`Replacement`) at Step 14? What was the agent's internal reasoning and chain-of-thought when it made that surgical edit?

### 1.2 The 100-Year Vision: The Quantum Holographic Review Pane
The **Holographic GitHub & Agent Step Review Pane** is an ultra-futuristic, multi-dimensional code diff and inspection workbench integrated directly into the far right of the Session Explorer File Viewer (`#fileViewerModal`). 

Crafted with the PolyTuring Year 2500 cosmic liquid-glass design system across 4 isomorphic themes (Liquid Glass, Fluent Aero Mica, Holo Cyberpunk, and Matrix Obsidian), the review pane offers:
- **Zero-Credential Friction**: Seamlessly leverages the user's local machine-level `git` installation, `gh` CLI (`/opt/homebrew/bin/gh`), and ambient environment tokens (`GITHUB_TOKEN`, `GH_TOKEN`, or standard macOS Keychain/system keyring)—requiring zero additional API keys, logins, or configurations.
- **The 4-Vector Forensic Review Matrix**:
  1. 📝 **Working Copy Diff** (Unstaged working tree changes vs. HEAD).
  2. 📦 **Staged Index Diff** (Changes staged for commit).
  3. 📜 **Commit History Time-Travel** (Interactive commit selector with author avatars, relative timestamps, and commit-specific diffs).
  4. ⚡ **Agent Step Chrono-Diff** (Scrubbable step-by-step diffs reconstructed directly from the session's JSONL transcript tool calls: `edit_file`, `write_file`, and `create_file`).
- **Remote GitHub Pull Request Awareness**: Auto-detects if the workspace branch has an associated GitHub PR via `gh pr view`, surfacing PR status badges (`Open`, `Draft`, `Merged`), reviewer approval states, and 1-click external navigation.
- **Dynamic Fluid Resizing & Splitter Control**: A silky-smooth, draggable split-resizer (`.file-viewer-resizer`) with double-click snap presets (50/50, 65/35, 35/65), keyboard toggle (`⌘\`), and persistent width memory via `localStorage`.
- **IDE-Grade Diff Ergonomics**: Toggle between **Side-by-Side (Split)** and **Inline (Unified)** diff views, with sub-line token/character-level change highlighting, syntax-colored code lines, collapsible hunk headers, and bidirectional scroll/line synchronization with the main file viewer.

---

## 2. Problem Statement & User Pain Points

| Pain Point | Root Cause | Impact on GenAI Engineers | PRD Solution |
| :--- | :--- | :--- | :--- |
| **Opaque Agent Mutations** | Files viewed in Session Explorer show current disk content only; the user cannot see what specific lines were touched by a specific turn. | Engineers cannot verify whether an agent made a correct surgical edit without switching to an external terminal or IDE. | **Vector 4: Agent Step Chrono-Diff**. Reconstructs exact `Target` vs `Replacement` per step with step scrubbing and agent rationale cards. |
| **Context Switching to Terminal / GitHub** | To see if local changes are staged or commit history of a file, users must leave Session Explorer and run `git log -p` or open github.com. | Destroys cognitive flow and slows down rapid multi-session debugging. | **Embedded Far-Right Review Pane**. Native git status, unstaged/staged diffs, commit selector, and GitHub PR telemetry in one view. |
| **Credential & Auth Fatigue** | Web tools often demand private GitHub Personal Access Tokens (PATs) or OAuth apps, creating security hazards and friction. | Users hesitate to enter tokens into web tools; enterprise security policies block third-party tokens. | **Zero-Config Machine Identity**. Binds directly to the local daemon's execution context, using the user's existing `gh` CLI and git credentials. |
| **Screen Real Estate Cannibalization** | File viewer modals are either too narrow or fixed-width, causing horizontal scrollbars and truncated diffs. | Reviewing wide code diffs in a cramped modal is mentally exhausting. | **97.5vw Fluid Shell + Draggable Resizer**. Flexible dynamic pane splitting, snap points, responsive collapse on mobile/tablets, and width persistence. |
| **Lack of Token-Level Precision** | Line-level diffs (`+` and `-`) make subtle 1-character typo fixes or argument renames hard to spot in a 100-character line. | High risk of missing silent logic bugs introduced by AI hallucinations. | **Sub-Line Token Diffing**. Computes intra-line character/word Levenshtein diff spans and highlights exact altered tokens. |

---

## 3. Personas & Core User Scenarios

### 3.1 Personas
1. **The Sovereign GenAI Fast-Prototyper**:
   - Runs autonomous multi-step agent coding loops.
   - Demands rapid verification of agent output at Step 12 vs Step 15 before letting the agent continue.
   - Expects instant visual feedback without leaving the session timeline.
2. **The Senior Systems Pair-Programmer & Reviewer**:
   - Audits code changes generated by junior developers or AI agents.
   - Needs to verify that local git changes conform to repo standards and that commits tell a coherent story.
   - Requires side-by-side split diffs with line gutters and syntax highlighting.
3. **The Engineering Lead / Release Architect**:
   - Manages PR merges and branch lifecycles.
   - Wants to see if the files touched in a session correspond to open GitHub PRs, check CI statuses, and confirm pull request alignment.

### 3.2 Key User Journeys

#### Journey 1: Forensic Agent Step Verification
- **Trigger**: User is reviewing a session trail where an agent claimed: *"I refactored `api.go` to handle streaming SSE connections and updated error parsing."*
- **Action**: User clicks the `.tool-file-pill` for `src/api.go` at Step 18.
- **Experience**: The Native File Viewer opens showing `src/api.go`. The review pane on the right automatically initializes to **⚡ Agent Step Mode**, locked to Step 18.
- **Result**: The user sees the exact hunk: lines 420-435 removed (red `-`), replaced by lines 420-455 (green `+`). Below the diff, an expandable capsule displays the Agent's Thinking from Step 18. The user scrubs to Step 22 to see the follow-up edit. Total time to verify: 4 seconds.

#### Journey 2: Pre-Commit Hygiene & Staged Verification
- **Trigger**: User completed a coding session and wants to know what files were modified on disk before committing.
- **Action**: User opens the Session Files & Artifacts Matrix (R-045), filters by `✏️ Codebase Mutations`, and clicks `Inspect` on `src/scanner.go`.
- **Experience**: The file viewer opens. User clicks the `📝 Working Copy` tab on the Review Pane.
- **Result**: The Review Pane shows the unstaged diff against `HEAD`. An isomorphic toggle allows switching to `📦 Staged` to verify what will go into the next commit. A `📋 Copy Patch` button copies a unified diff to the clipboard.

#### Journey 3: Commit History Exploration
- **Trigger**: User notices a function in `types.go` and wonders when a specific field was introduced.
- **Action**: In the Review Pane, user selects the `📜 Commits` tab.
- **Experience**: An interactive list of the last 30 commits touching `types.go` appears, showing commit messages, authors, relative timestamps, and short SHAs.
- **Result**: User clicks commit `a7b3c2d ("feat: add tool sub-breakdown telemetry")`. The diff view instantly updates to show the exact changes introduced in that commit.

#### Journey 4: GitHub PR Correlation
- **Trigger**: User is exploring a session in a feature branch (`feat/review-pane`).
- **Action**: User opens a file. The Review Pane's top ribbon displays `🐙 PR #42: Feature / GitHub Review Pane [Open]`.
- **Experience**: User clicks the `🐙 GitHub PR` tab. The pane displays the file diff comparing the PR head against the base branch (`main`), along with CI check status (`✓ 14 passing`). User can click `↗ Open on GitHub` to jump directly to the pull request in their browser.

---

## 4. Product Goals & Measurable Success Metrics

| Goal | Metric | Baseline | Target (v0.5.0) |
| :--- | :--- | :--- | :--- |
| **G-1: Zero-Config Discovery** | Frictionless machine git & `gh` CLI detection without entering any tokens | 0% | **100% Automatic** |
| **G-2: Verification Velocity** | Time required to verify an agent's code change in a session step | ~45s (open editor, git diff) | **< 3s (in-modal click)** |
| **G-3: UI Responsiveness** | Review pane split drag FPS and modal render latency | N/A | **60 FPS, < 50ms render** |
| **G-4: Layout Space Utilization** | Viewport efficiency on displays from 1200px to 3840px | Static single pane | **Adaptive 50/50, 65/35, fluid** |
| **G-5: Zero Layout Regression** | Compatibility with existing R-038 file viewer features (Find, Wrap, Copy) | 100% | **100% preserved** |
| **G-6: Test Automation Coverage** | Automated Go unit tests and Playwright headless tests | 36 suites | **100% passing across all new endpoints and UI interactions** |

---

## 5. Architectural & Technical Decomposition

### 5.1 System Architecture

```
┌────────────────────────────────────────────────────────────────────────┐
│                        SESSION EXPLORER WORKBENCH                      │
│                                                                        │
│   ┌────────────────────────────────────────────────────────────────┐   │
│   │                    Native File Viewer Modal                    │   │
│   │  ┌─────────────────────────┬───┬────────────────────────────┐  │   │
│   │  │   Active File Display   │   │  Holographic Review Pane   │  │   │
│   │  │   (Code / Markdown)     │ ║ │  (Git & Agent Step Diff)   │  │   │
│   │  │                         │ ║ │                            │  │   │
│   │  │  - Line Number Gutter   │ ║ │  - Quantum Ribbon          │  │   │
│   │  │  - Syntax Highlighting  │ ║ │  - Vector Tab Switcher     │  │   │
│   │  │  - Pinned Find Widget   │ ║ │  - Commit / Step Scrubber  │  │   │
│   │  │  - Anchor Navigation    │ ║ │  - Split / Unified Diff    │  │   │
│   │  │                         │ ║ │  - Token-Level Highlighting│  │   │
│   │  └─────────────────────────┴───┴────────────────────────────┘  │   │
│   │                            │ ║ │                               │   │
│   │                            │ ║ └── Draggable Split Resizer     │   │
│   └────────────────────────────┴─╫─────────────────────────────────┘   │
│                                  ║                                     │
│   REST API Layer (Go)            ║                                     │
│   ┌──────────────────────────────▼────────────────────────────────┐   │
│   │  /api/git/file/status   ──▶ GitRoot, Branch, PR, Dirty State   │   │
│   │  /api/git/file/diff     ──▶ Working Copy, Staged, Commit Diff │   │
│   │  /api/git/file/commits  ──▶ File Commit Chronology History    │   │
│   │  /api/git/pr/info       ──▶ GitHub PR Metadata & Checks        │   │
│   │  /api/sessions/:id/file/steps ──▶ Agent Tool Diff Hunks       │   │
│   └──────────────────────────────┬────────────────────────────────┘   │
│                                  │                                     │
│   Local Machine Layer            ▼                                     │
│   ┌───────────────────────────────────────────────────────────────┐   │
│   │  - git CLI (v2.56.0)     - Local Repo (.git)                  │   │
│   │  - gh CLI (v2.103.3)     - Machine GITHUB_TOKEN / Keyring     │   │
│   │  - Session JSONL Store   - transcript.jsonl (Tool Calls)      │   │
│   └───────────────────────────────────────────────────────────────┘   │
└────────────────────────────────────────────────────────────────────────┘
```

### 5.2 The 4 Review Vectors (Detailed Data Flow)

#### Vector 1: Working Copy Diff (Unstaged)
- **Backend Action**: Executes `git -C <repoRoot> diff --no-ext-diff -- <relPath>`
- **Safety**: Timeout bounded to 3 seconds; output capped at 2MB to prevent browser OOM.
- **Telemetry**: Computes lines added (`+`), lines removed (`-`), total hunks.

#### Vector 2: Staged Index Diff
- **Backend Action**: Executes `git -C <repoRoot> diff --cached --no-ext-diff -- <relPath>`
- **Output**: Formatted unified diff of changes staged in git index.

#### Vector 3: Commit History Time-Travel
- **Backend Action**:
  - List: `git -C <repoRoot> log --follow -n 30 --pretty=format:%H%x1f%h%x1f%an%x1f%at%x1f%s -- <relPath>`
  - Detail: `git -C <repoRoot> show --no-ext-diff <commitSha> -- <relPath>`
- **UI Interaction**: Interactive commit selector list with author initials, commit message, time ago, and diff viewer.

#### Vector 4: Agent Step Chrono-Diff (GenAI Innovation)
- **Backend/Frontend Action**:
  - Scans session transcript for messages where `step_index` touches the file.
  - For `edit_file` tool calls:
    - Extracts `Target` (exact lines removed) and `Replacement` (exact lines added).
    - Generates a synthetic unified diff hunk with accurate line context.
    - Captures the agent's prompt, response content, and `thinking` block from that step.
  - For `write_file` tool calls:
    - Analyzes whether file was created from scratch or overwritten.
    - Highlights full content or appended content.
- **UI Interaction**: A step timeline slider/scrubber (`Step 4` $\to$ `Step 9` $\to$ `Step 14` $\to$ `Step 22`) allowing the developer to watch code mutate turn-by-turn.

#### Vector 5: Remote GitHub PR Context (Bonus Vector)
- **Backend Action**:
  - Executes `gh pr view --json number,title,state,url,headRefName,baseRefName,reviewDecision,statusCheckRollup`
  - Fallback: If `gh` is unavailable or not authenticated, cleanly disables this tab without blocking local git features.

---

## 6. Functional Requirements & Acceptance Criteria

### 6.1 Requirements Overview (R-053 to R-060)

| Requirement ID | Module | Title & Description | Priority |
| :--- | :--- | :--- | :--- |
| **R-053** | Backend/Core | **Native Git Trackability & Ambient Identity Engine**: Detect repository boundaries using `findGitRoot`, extract branch, dirty status, remote origin, and ambient `gh` CLI status with zero credential configuration. | Must |
| **R-054** | Backend/API | **Unified Git File Diff & Commit History Endpoints**: Provide `/api/git/file/status`, `/api/git/file/diff`, and `/api/git/file/commits` with bounded timeouts, output size ceilings, and structured JSON responses. | Must |
| **R-055** | Core/Engine | **Agent Step Chrono-Diff & Hunk Extraction Engine**: Extract and synthesize step-by-step file modifications (`edit_file` Target vs Replacement, `write_file`) from session transcripts with agent thinking correlation. | Must |
| **R-056** | Frontend/UI | **Holographic Review Pane Container & Quantum Telemetry Ribbon**: Render `#fileViewerReviewPane` on the far right of `#fileViewerModal` with 4-theme styling, status badges (`+`/`-` churn, branch, PR), and tabbed vector switcher. | Must |
| **R-057** | Frontend/UX | **Interactive Draggable Splitter & Responsive Geometry**: Implement `.file-viewer-resizer` with mouse/touch drag, double-click snap presets (50/50, 65/35), persistence in `localStorage`, and `⌘\` keyboard toggle. | Must |
| **R-058** | Frontend/Diff | **Dual-Mode Diff Engine (Split & Unified) with Token Highlighting**: Render side-by-side or inline diffs with syntax highlighting, sub-line word/token diff highlights, and collapsible hunks. | Must |
| **R-059** | Frontend/Sync | **Bidirectional Viewport Sync & Anchor Navigation**: Clicking a diff line or hunk scrolls and highlights the corresponding line in the code viewer, and vice versa. | Should |
| **R-060** | Integration | **Ambient GitHub Pull Request & Remote Integration**: Query `gh pr view` for active PR status, review decisions, and CI checks, providing 1-click external navigation to GitHub. | Should |

---

## 7. UI/UX Design Specifications (Year 2500 Cosmic Aesthetic)

### 7.1 Visual Hierarchy & Layout Diagram

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│ .file-viewer-topbar                                                                    │
│ 📄 api.go  [GO] [774 lines] [28.4 KB]       [📖 Preview|📝 Raw] [🔍 Find] [↩ Wrap] ... │
│                                             [🔀 Review Pane: ⚡ Steps (3) | 🌿 git] ✕ │
├─────────────────────────────────────────────┬───┬──────────────────────────────────────┤
│ .file-viewer-code-container (55%)           │ ║ │ #fileViewerReviewPane (45%)          │
│                                             │ ║ │                                      │
│  1  package main                            │ ║ │  [⚡ Steps (3)] [📝 Work] [📜 Commits]│
│  2                                          │ ║ │  🌿 feat/review-pane  *Dirty  +42 -12│
│  3  import (                                │ ║ ├──────────────────────────────────────┤
│  4    "bytes"                               │ ║ │  Step 14 of 22: edit_file (2m ago)   │
│  5    "fmt"                                 │ ║ │  ◀ Prev Step        Next Step ▶     │
│  6  )                                       │ ║ │ ┌──────────────────────────────────┐ │
│  7                                          │ ║ │ │ 🧠 Agent Thought (Step 14):      │ │
│  8  func HandleGitStatus(w, r) {            │ ║ │ │ "Add git status endpoint to..."  │ │
│  9    // ...                                │ ║ │ └──────────────────────────────────┘ │
│ 10  }                                       │ ║ │ @@ -42,7 +42,12 @@                  │
│                                             │ ║ │ -  func oldMethod() {               │
│                                             │ ║ │ +  func newMethod() {               │
│                                             │ ║ │ +    checkSafety()                  │
│                                             │ ║ │ +  }                                │
│                                             │ ║ │                                      │
└─────────────────────────────────────────────┴───┴──────────────────────────────────────┘
```

### 7.2 Theme Palette Integration
- **Liquid Glass**: Deep space obsidian `#070a13`, neon cyan borders `rgba(6, 182, 212, 0.25)`, glowing diff additions `rgba(16, 185, 129, 0.15)` and deletions `rgba(244, 63, 94, 0.15)`.
- **Fluent Aero Mica**: Frosted acrylic backdrop blur (24px), subtle glass specular highlights, soft text contrast.
- **Holo Cyberpunk**: High-voltage magenta and cyan laser accents, monospace telemetry headers, phosphor glow.
- **Matrix Obsidian**: CRT terminal greens `#22c55e`, deep scanlines, monospaced ASCII glyphs.

---

## 8. Security, Performance & Scalability Guarantees

1. **Security & Sandbox Isolation**:
   - Zero command injection: All Git commands execute via `exec.Command` with discrete argument arrays (no `/bin/sh -c` shell evaluation).
   - Strict Path Traversal Protection: Target files must resolve within the verified Git repository root or session workspace.
   - Credentials Protection: System never prompts for, stores, or logs GitHub tokens; only standard environment variables and system keyrings are inherited by subprocesses.
2. **Performance Constraints**:
   - Max file diff size: 2MB (prevents DOM freezing).
   - Process execution timeout: 3.0 seconds per Git command.
   - Resizer frame rate: 60 FPS via `requestAnimationFrame` and CSS variable hardware acceleration.
3. **Graceful Fallback**:
   - If a file is NOT in a Git repository: The review pane cleanly notifies the user (`ℹ️ File is outside Git tracking`), and smoothly defaults to **⚡ Agent Step Chrono-Diff Mode**, which works 100% of the time based on session transcripts!

---

## 9. Conclusion & Next Steps
This PRD lays the foundation for transforming Session Explorer from a static viewer into an intelligent, multi-dimensional code review and AI verification workstation. Upon review and approval, we will proceed to the formal SDD Specification and test-driven implementation.
