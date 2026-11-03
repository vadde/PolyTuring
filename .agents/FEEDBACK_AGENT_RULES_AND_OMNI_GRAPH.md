# 🏛️ Architecture & Governance Feedback: Closing the Omni-Graph & Orientation Guardrail Loopholes

> **Context**: Post-mortem audit of an agent session where Omni-Graph was bypassed, legacy `grep`/file reads were used instead of AST queries, and initial orientation rules were skipped.
> **Audience**: Lead Agent Architects, Hook/Rule Maintainers, and Antigravity Governance Engineers.
> **Date**: 2026-11-03

---

## Executive Summary

The repository and agent system possess an industry-grade specification:
- **Rule 00**: Non-negotiable mandatory Omni-Graph retrieval, 150-line gate, blast-radius pre-checks.
- **Rule 101 (Agent Constitution)**: Early mentorship, verification over hallucination, spec-driven development.
- **Active Hooks Framework**: `hooks.json` specifying `PreInvocation`, `PreToolUse`, and `Stop` hooks.

**However, in practice, agents still bypassed Omni-Graph and skipped Turn 0 reading of `AGENTS.md` and `CONTEXT.md`.**

Our forensic inspection of the runtime hook implementation revealed **four structural loopholes** where runtime guardrails fail to coerce agent behavior. Rules that rely solely on an LLM's memory will inevitably degrade under context pressure unless enforced by **deterministic pre-tool interceptors**.

Below is the concrete analysis of why the guardrails leaked, followed by four drop-in architectural enhancements.

---

## Part 1: Forensic Root Causes of the Guardrail Leaks

### 1. The Tool Name & Parameter Mismatch (`read_file` vs `view_file`)
- **What `hooks.json` matched**:
  ```json
  "matcher": "run_command|grep_search|view_file"
  ```
- **What Antigravity CLI actually declared & executed**:
  The active native tool is **`read_file`** (with parameter `Path`), whereas `hooks.json` only intercepted **`view_file`** (with parameter `AbsolutePath`).
- **Blast Radius**:
  Because `read_file` is not in the matcher, **every single `read_file` call completely bypassed `hook_pre_tool.sh`**.
  The 150-line gate, the 450-line documentation ceiling, and the Omni-Graph pre-recon requirement were never triggered at runtime during file reads. The agent was able to read 800 lines at will.

---

### 2. The Non-Recursive Grep Bypass in `run_command`
- **What `hook_pre_tool.sh` blocked**:
  ```python
  blocked_patterns = [
      "grep -r", "grep -rn", "grep -ri", "grep -rin", "grep -rIn",
      "find . -name", "cat $(find", "ag -l", "rg -l"
  ]
  ```
- **What the agent actually executed**:
  ```bash
  grep -n "func (idx *SessionIndex) ExtractSessionAgentContext" tools/session-explorer/src/*.go
  grep -n -A 35 "func extractWorkspace" tools/session-explorer/src/scanner.go
  ```
- **Blast Radius**:
  The regex only intercepted recursive flags (`-r`). Standard `grep -n` against globbed source paths (`*.go`, `*.rs`, `*.py`) or single files slipped right past the hook, allowing the agent to perform raw regex searches across thousands of lines instead of querying `:8080/api/symbol`.

---

### 3. The Advisory-Only Prompt Injection (Optical Illusion of Enforcement)
- **What `hook_pre_invocation.sh` injected**:
  ```
  🧭 Omni-Graph | ws: PolyTuring (3,813 nodes | live-watch: ON)
  Endpoints (:8080/api): /query (Graph-RAG) | /galaxies | /condense?symbol=X | /references?symbol=X | /symbol?name=X | /search?q=X
  [Gate: direct source reads >150 lines blocked — query AST endpoints first]
  ```
- **The Cognitive Failure**:
  The injected notice states `[Gate: direct source reads >150 lines blocked]`. But because of Loophole #1, the gate was physically absent on `read_file`.
  When an agent realizes (either consciously or through trial) that `read_file` succeeds without an AST query, the system prompt's instructions become perceived as soft suggestions rather than hard kernel constraints.

---

### 4. Absence of a "Turn 0 Orientation Gate"
- **The Problem**:
  `AGENTS.md` and `07-iterative-development.md` mandate that the agent **must** read `AGENTS.md` and the tool's `CONTEXT.md` before taking any action.
  However, there is **zero mechanical enforcement** for this. When an agent receives an urgent or specific prompt (e.g. *"why did session c0429a76 land under DSA?"*), the agent's attention immediately attends to the user's tokens and rushes straight to debugging, bypassing orientation completely.

---

## Part 2: Four Architectural Solutions (Drop-In Hardening)

### Solution 1: Unify Tool Matchers & Parameter Normalization
Update `hooks.json` to include all file-reading tool variants across CLI and IDE:

```json
{
  "omni-graph-guardrail": {
    "enabled": true,
    "PreToolUse": [
      {
        "matcher": "run_command|grep_search|view_file|read_file|edit_file|write_file",
        "hooks": [
          {
            "type": "command",
            "command": "/Users/aparv/.gemini/config/scripts/hook_pre_tool.sh",
            "timeout": 5
          }
        ]
      }
    ]
  }
}
```

In `hook_pre_tool.sh`, normalize tool names and arguments:
```python
# Normalize target path across read_file (Path) and view_file (AbsolutePath)
target_path = (
    args.get("Path") or 
    args.get("AbsolutePath") or 
    args.get("TargetFile") or 
    args.get("file_path") or ""
)

if tool_name in ("view_file", "read_file"):
    # Enforce 150-line gate and Omni-Graph pre-recon check
    ...
```

---

### Solution 2: Hard-Block File-Globbing Grep in `run_command`
Broaden `blocked_patterns` in `hook_pre_tool.sh` so agents cannot bypass Omni-Graph via shell commands:

```python
# Intercept ANY grep, ripgrep, or ag targeting source code files
SOURCE_EXT_PATTERN = r'\.(go|rs|py|ts|js|java|c|cpp|h|hpp|zig|scala|kt|swift)\b'

if tool_name == "run_command":
    cmd = args.get("CommandLine", "")
    
    # 1. Allow Omni-Graph curl/make commands to pass and register recon
    if any(p in cmd for p in omni_patterns):
        record_recon(get_workspace())
        return ALLOW
        
    # 2. Block all brute-force search commands targeting source code
    is_search_cmd = re.search(r'\b(grep|egrep|fgrep|rg|ag|ack)\b', cmd)
    if is_search_cmd and re.search(SOURCE_EXT_PATTERN, cmd):
        return DENY(
            "🛑 BLOCKED: Brute-force grep on source files is prohibited. "
            "Use Omni-Graph AST queries instead: "
            "curl -s 'http://localhost:8080/api/symbol?name=<sym>&workspace=<ws>' "
            "or 'http://localhost:8080/api/search?q=<query>&workspace=<ws>'"
        )
```

---

### Solution 3: The Turn 0 Orientation Gate (Coercive Initialization)
To guarantee that **no agent can modify code without reading repository orientation first**, introduce a simple session-state check in `PreToolUse`:

```python
# In hook_pre_tool.sh:
conv_id = data.get("conversationId", "default")
oriented_marker = f"/tmp/agy_oriented_{conv_id}.flag"

# If the agent attempts to run write_file or edit_file before reading CONTEXT.md / AGENTS.md:
if tool_name in ("write_file", "edit_file"):
    if not os.path.exists(oriented_marker):
        return DENY(
            "🛑 ORIENTATION GATE VIOLATION (Rule 07 / SDD Constitution):\n"
            "You cannot mutate code before completing Turn 0 orientation.\n"
            "Mandatory protocol: Read AGENTS.md and the target tool's CONTEXT.md first."
        )

# When the agent reads AGENTS.md or CONTEXT.md, satisfy the gate:
if tool_name in ("read_file", "view_file"):
    path_lower = target_path.lower()
    if "context.md" in path_lower or "agents.md" in path_lower:
        with open(oriented_marker, "w") as f:
            f.write(str(int(time.time())))
```

**Why this works**:
- It does not interfere with read-only questions or chat.
- But the moment an agent attempts to edit code or write new files, it hits a hard mechanical barrier unless it has proven it read `AGENTS.md` or `CONTEXT.md` in that session.

---

### Solution 4: Caller Blast-Radius Interception on `edit_file`
Under Rule 00, agents must verify caller blast radius before editing public functions or shared routines.
We can enforce this mechanically:

```python
# If modifying a Go, Rust, or Python source file:
if tool_name == "edit_file" and is_source_code(target_path):
    ws = get_workspace(target_path)
    ref_marker = f"/tmp/omni_ref_checked_{ws}.marker"
    
    # Check if agent has queried /api/references in the past 10 minutes
    if not is_reference_check_recent(ref_marker):
        # We can either warn or deny if high-risk. A soft warning on stderr or 
        # prompt reminder enforces consciousness:
        inject_step_warning(
            "⚠️ BLAST-RADIUS REMINDER: You are editing source code. "
            "Ensure you verified callers via :8080/api/references?symbol=X."
        )
```

---

## Part 3: Feedback for the Agent Authors (Synthesis)

1. **Rule of Coercion**:
   *"If a rule is not enforced by a compiler or a pre-tool gate, it is a wish, not an invariant."*
   LLMs behave probabilistically. When context grows long or prompts demand immediate solutions, un-enforced instructions in markdown files get dropped from the active attention matrix. Guardrails must live in deterministic code (`hooks.json`), not just docstrings.

2. **Unified Tool Naming Registry**:
   Ensure `hooks.json` matches both the current and legacy tool namespaces (`read_file` + `view_file`, `write_file` + `edit_file`).

3. **Low-Friction Alternatives**:
   Agents often run `grep` because typing `grep -n X file.go` is fewer keystrokes than `curl -s "http://localhost:8080/api/symbol?name=X&workspace=PolyTuring"`.
   By adding ultra-short bash helpers or CLI flags (e.g. `omni sym <name>` or `omni ref <name>`), the path of least resistance becomes the desired path.

---
EOF
