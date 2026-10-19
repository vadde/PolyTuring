---
name: session-wip
description: >-
  Recover work-in-progress (WIP) state, recent prompt trajectories, modified files, and execution errors for coding assistants in the event of context window truncation, memory loss, or session wipeout. Enforces mandatory agent halting when multiple workspace sessions exist.
---

# 🧭 Session WIP & Trajectory Recovery Skill

This skill allows coding assistants to **inspect recent conversation trajectories, recover lost memory, and resume unfinished tasks** seamlessly using Session Explorer's telemetry engine.

---

## 🎯 When to Use

1. **Memory Loss / Session Truncation**: When your context window was reset, wiped out, or you are resuming work on an active repository.
2. **User Triggers**: When the user asks `/most-recent-wip`, `/wip`, "pick up where we left off", "what was the last thing we did?", or "recover recent session".
3. **Forensic Audit**: To discover which files were recently modified, which terminal commands were run, or what errors were encountered in prior steps.

---

## 🛑 MANDATORY INVARIANT: Multi-Session Disambiguation (R-031)

> [!CAUTION]
> **NEVER ASSUME OR SILENTLY PICK A SESSION WHEN MULTIPLE CANDIDATES EXIST.**
> If the workspace contains more than one active session, the recovery tool will return `multiple: true` with a candidate table.
>
> In this scenario, you **MUST STOP IMMEDIATELY** and:
> 1. Render the candidate table to the user (showing Session ID, Last Active time, Step Count, and Prompt Previews).
> 2. Ask the user: *"Multiple active sessions were found for this workspace. Which session would you like me to resume?"*
> 3. **Wait for the user's explicit response** before loading a session. Do NOT proceed autonomously!

---

## ⚡ Execution Protocol

### Step 1: Run the Recovery Tool
Run the companion script for your active workspace:

```bash
# Auto-detects workspace from current directory
./tools/session-explorer/scripts/wip.sh

# Or target a specific workspace
./tools/session-explorer/scripts/wip.sh -w PolyTuring

# Or query via the local REST API if daemon is running
curl -s "http://127.0.0.1:9876/api/wip?workspace=PolyTuring"
```

### Step 2: Handle the Response

#### Case A: Multiple Sessions Returned (`multiple: true`)
Stop immediately and present the formatted candidate table:

```markdown
🛑 Multiple sessions found for workspace "PolyTuring":

| # | Session ID | Last Active | Steps | Prompts | Last Prompt Preview |
|---|---|---|---|---|---|
| 1 | `70af9b9e-...` | just now | 380 | 2 | Expand session explorer with WIP recovery... |
| 2 | `32e5e03b-...` | 4 days ago | 14,107 | 39 | Min-Max mTSP warehouse router optimization... |

Please select which session you would like me to resume.
```

#### Case B: Single Session Resolved (or user specified `-s <id>` or `-l`)
The tool outputs the deep WIP snapshot:
- **Last Active Timestamp**: Exact date/time and relative age.
- **Recent Prompts**: The last 3 user prompts defining the task goals.
- **Modified Files**: Exact paths touched via `write_to_file` or `replace_file_content`.
- **Executed Commands**: Shell commands and test runs.
- **Recent Errors**: Any compilation, test, or tool failures from final steps.
- **Suggested Continuation**: Recommended next step to pick up the slack.

---

## 🛠️ Command Reference

| Flag | Description | Example |
|---|---|---|
| `-w, --workspace <NAME>` | Filter by workspace or repo name | `./scripts/wip.sh -w PolyTuring` |
| `-s, --session <UUID>` | Force inspection of an exact session ID | `./scripts/wip.sh -s 70af9b9e-...` |
| `-l, --latest` | Override disambiguation and force latest | `./scripts/wip.sh -w PolyTuring -l` |
| `--limit <N>` | Maximum candidate sessions to display | `./scripts/wip.sh --limit 10` |
| `--json` | Return structured JSON output | `./scripts/wip.sh -w PolyTuring --json` |

---

## 🌐 HTTP REST API Equivalent

When communicating programmatically:
- `GET http://127.0.0.1:9876/api/wip?workspace=<name>&limit=5`
- `GET http://127.0.0.1:9876/api/wip?session_id=<uuid>`
- `GET http://127.0.0.1:9876/api/wip?workspace=<name>&force_latest=true`
- `POST http://127.0.0.1:9876/api/skill/install` (Payload: `{"target": "workspace"|"global"}`)

Interactive Swagger documentation is available at **`http://127.0.0.1:9876/docs`**.
