#!/usr/bin/env bash
# ==============================================================================
# wip.sh — Agent Work-in-Progress (WIP) Recovery Script
#
# Enables coding assistants to inspect session history, pick up recent changes,
# and recover context following context window truncation, crashes, or reboots.
#
# Enforces the Multi-Session Disambiguation Invariant (R-031):
#   If multiple sessions match the workspace, it HALTS and outputs a candidate
#   selection table, directing the assistant to stop and ask the user.
#
# Dual-Mode Operation:
#   1. Queries running Session Explorer daemon (http://127.0.0.1:9876/api/wip)
#   2. Falls back to headless CLI binary (session-explorer wip) if daemon is offline
# ==============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TOOL_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${TOOL_DIR}/../.." && pwd)"
BIN_PATH="${TOOL_DIR}/bin/session-explorer"

# Default arguments
WORKSPACE=""
SESSION_ID=""
FORCE_LATEST="false"
OUTPUT_JSON="false"
LIMIT="5"
PORT="${SESSION_EXPLORER_PORT:-9876}"

# Parse CLI arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    -w|--workspace)
      WORKSPACE="$2"
      shift 2
      ;;
    -s|--session)
      SESSION_ID="$2"
      shift 2
      ;;
    -l|--latest|-f|--force)
      FORCE_LATEST="true"
      shift
      ;;
    --limit)
      LIMIT="$2"
      shift 2
      ;;
    --json)
      OUTPUT_JSON="true"
      shift
      ;;
    -h|--help)
      echo "Usage: ./wip.sh [workspace] [--session <uuid>] [--latest] [--limit <n>] [--json]"
      echo ""
      echo "Options:"
      echo "  -w, --workspace <NAME>   Target repository or workspace name"
      echo "  -s, --session <UUID>     Target exact session UUID"
      echo "  -l, --latest             Force selection of newest session (skip multi-session halt)"
      echo "      --limit <N>          Maximum candidates to return (default: 5)"
      echo "      --json               Output structured JSON instead of Markdown"
      exit 0
      ;;
    *)
      if [[ -z "${WORKSPACE}" ]]; then
        WORKSPACE="$1"
      fi
      shift
      ;;
  esac
done

# If workspace is still empty, derive from current working directory
if [[ -z "${WORKSPACE}" && -z "${SESSION_ID}" ]]; then
  WORKSPACE="$(basename "$(pwd)")"
fi

# 1. Try querying local daemon if alive
DAEMON_URL="http://127.0.0.1:${PORT}/api/wip"
QUERY="workspace=$(python3 -c "import urllib.parse, sys; print(urllib.parse.quote(sys.argv[1]))" "${WORKSPACE}")"
if [[ -n "${SESSION_ID}" ]]; then
  QUERY="${QUERY}&session_id=${SESSION_ID}"
fi
if [[ "${FORCE_LATEST}" == "true" ]]; then
  QUERY="${QUERY}&force_latest=true"
fi
if [[ -n "${LIMIT}" ]]; then
  QUERY="${QUERY}&limit=${LIMIT}"
fi

if curl -s -f -m 1 "http://127.0.0.1:${PORT}/api/wip?workspace=${WORKSPACE}&limit=1" >/dev/null 2>&1; then
  # Daemon is online and supports /api/wip!
  RESP=$(curl -s "${DAEMON_URL}?${QUERY}")
  if [[ "${OUTPUT_JSON}" == "true" ]]; then
    echo "${RESP}"
    exit 0
  fi

  # Render with python formatting for clean CLI output
  python3 -c '
import json, sys

data = json.loads(sys.argv[1])
status = data.get("status")
workspace = data.get("workspace", "")

if status == "not_found":
    print("❌ No active sessions found for workspace \"" + str(workspace) + "\".")
    sys.exit(0)

if data.get("multiple"):
    cands = data.get("candidates", [])
    print("🛑 MANDATORY AGENT HALT: Multiple sessions found for workspace \"" + str(workspace) + "\".\n")
    print("You MUST stop immediately and ask the user which session they wish to resume:\n")
    print("| # | Session ID | Last Active | Steps | Prompts | Last Prompt Preview |")
    print("|---|---|---|---|---|---|")
    for i, c in enumerate(cands, 1):
        prev = (c.get("last_prompt_preview") or "").replace("\n", " ")
        if len(prev) > 80: prev = prev[:80] + "..."
        sid = c.get("session_id")
        rel = c.get("relative_time")
        steps = c.get("step_count")
        prompts = c.get("user_message_count")
        print(f"| {i} | `{sid}` | {rel} | {steps} | {prompts} | {prev} |")
    print(f"\nPlease ask the user: \"I found {len(cands)} recent sessions for workspace {workspace}. Which session should I resume?\"")
    sys.exit(0)

act = data.get("active_session")
if not act:
    print("No active session data.")
    sys.exit(0)

pname = act.get("project_name", "")
sid = act.get("session_id", "")
rel = act.get("relative_time", "")
last_msg = act.get("last_message_at", "")
steps = act.get("step_count", 0)
prompts = act.get("user_message_count", 0)
errors = act.get("error_count", 0)

print(f"### 🧭 Session WIP Summary: {pname}\n")
print(f"- **Session ID**: `{sid}`")
print(f"- **Last Active**: {rel} ({last_msg})")
print(f"- **Steps**: {steps} | **Prompts**: {prompts} | **Errors**: {errors}\n")

prompt_list = act.get("recent_prompts", [])
if prompt_list:
    print("#### 📝 Recent User Prompts:")
    for i, p in enumerate(prompt_list, 1):
        clean_p = p.replace("\n", " ").strip()
        if len(clean_p) > 140: clean_p = clean_p[:140] + "..."
        print(f"{i}. {clean_p}")
    print()

mods = act.get("modified_files", [])
if mods:
    print("#### 🛠️ Modified Files:")
    for m in mods:
        print(f"- `{m}`")
    print()

cmds = act.get("executed_commands", [])
if cmds:
    print("#### ⚡ Executed Shell Commands:")
    for cmd in cmds:
        print(f"- `{cmd}`")
    print()

errs = act.get("unresolved_errors", [])
if errs:
    print("#### ⚠️ Recent Errors / Blockers:")
    for e in errs:
        clean_e = e.replace("\n", " ").strip()
        if len(clean_e) > 120: clean_e = clean_e[:120] + "..."
        print(f"- {clean_e}")
    print()

sugg = act.get("suggested_continuation")
if sugg:
    print(f"#### 🎯 Suggested Continuation:\n{sugg}")
' "${RESP}"
  exit 0
fi

# 2. Daemon is offline -> fallback to headless CLI binary
if [[ ! -x "${BIN_PATH}" ]]; then
  echo "⚙️  Building session-explorer CLI binary..." >&2
  (cd "${TOOL_DIR}" && make build-go >/dev/null 2>&1)
fi

CLI_ARGS=("-workspace" "${WORKSPACE}")
if [[ -n "${SESSION_ID}" ]]; then
  CLI_ARGS+=("-session" "${SESSION_ID}")
fi
if [[ "${FORCE_LATEST}" == "true" ]]; then
  CLI_ARGS+=("-latest")
fi
if [[ "${OUTPUT_JSON}" == "true" ]]; then
  CLI_ARGS+=("-json")
fi
CLI_ARGS+=("-limit" "${LIMIT}")

exec "${BIN_PATH}" wip "${CLI_ARGS[@]}"
