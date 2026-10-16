#!/usr/bin/env bash
# ==============================================================================
# Omni-Graph PreInvocation Hook (High-Density / Zero-Fluff)
# Dynamically resolves active workspace from CLI payload (workspacePaths) or CWD,
# debounces duplicate hook executions (<1.5s per conversation), and injects an
# ultra-compact (<50 token) AST endpoint cheat-sheet into the agent prompt.
# ==============================================================================

set -e

PAYLOAD=$(cat)

python3 - "$PAYLOAD" << 'EOF'
import sys, json, os, subprocess, urllib.request, time

try:
    # 0. Parse Payload from stdin
    raw_input = sys.argv[1] if len(sys.argv) > 1 else sys.stdin.read()
    data = json.loads(raw_input) if raw_input.strip() else {}

    # Idempotency Guard: Debounce duplicate hook triggers (<1.5s) if both
    # local (.agents) and global (~/.gemini/config) hooks execute on the same turn
    conv_id = data.get("conversationId", "default")
    marker = f"/tmp/omni_pre_inv_{conv_id}.last"
    now = time.time()
    try:
        if os.path.exists(marker):
            with open(marker, "r") as f:
                last_ts = float(f.read().strip())
            if (now - last_ts) < 1.5:
                print(json.dumps({"injectSteps": []}))
                sys.exit(0)
        with open(marker, "w") as f:
            f.write(str(now))
    except Exception:
        pass

    # 1. Daemon Liveness Check (fast 500ms timeout)
    req = urllib.request.Request("http://localhost:8080/api/health", headers={"User-Agent": "OmniHook/1.0"})
    with urllib.request.urlopen(req, timeout=0.5) as resp:
        if resp.status != 200:
            print(json.dumps({"injectSteps": []}))
            sys.exit(0)

    # 2. Dynamically determine active workspace root directory
    active_dir = ""
    ws_paths = data.get("workspacePaths", [])
    if ws_paths and isinstance(ws_paths, list) and ws_paths[0]:
        active_dir = ws_paths[0]

    if not active_dir:
        active_dir = os.environ.get("WORKSPACE_CWD") or os.environ.get("WORKSPACE") or ""

    if not active_dir:
        active_dir = os.getcwd()

    git_root = ""
    try:
        git_root = subprocess.check_output(
            ["git", "-C", active_dir, "rev-parse", "--show-toplevel"],
            stderr=subprocess.DEVNULL,
            encoding="utf-8"
        ).strip()
        ws_candidate = os.path.basename(git_root)
    except Exception:
        ws_candidate = os.path.basename(active_dir)

    target_dir = git_root or active_dir

    # Detect if working inside a Git Worktree
    wt_info = None
    git_file = os.path.join(git_root, ".git") if git_root else os.path.join(target_dir, ".git")
    if os.path.isfile(git_file):
        try:
            with open(git_file, "r") as gf:
                first_line = gf.readline().strip()
            if first_line.startswith("gitdir:"):
                gitdir_path = first_line[7:].strip()
                if not os.path.isabs(gitdir_path):
                    gitdir_path = os.path.normpath(os.path.join(os.path.dirname(git_file), gitdir_path))
                head_file = os.path.join(gitdir_path, "HEAD")
                branch = "detached"
                if os.path.isfile(head_file):
                    with open(head_file, "r") as hf:
                        head_line = hf.readline().strip()
                        if head_line.startswith("ref: refs/heads/"):
                            branch = head_line[16:]
                parts = gitdir_path.replace("\\", "/").split("/")
                parent_ws = "parent"
                if ".git" in parts:
                    idx = parts.index(".git")
                    if idx > 0:
                        parent_ws = parts[idx - 1]
                wt_name = os.path.basename(git_root) or os.path.basename(target_dir)
                wt_info = {
                    "is_worktree": True,
                    "worktree_name": wt_name,
                    "parent_workspace": parent_ws,
                    "branch": branch
                }
        except Exception:
            pass

    # 3. Query workspaces from Omni-Graph
    req_ws = urllib.request.Request("http://localhost:8080/api/workspaces", headers={"User-Agent": "OmniHook/1.0"})
    workspaces = []
    with urllib.request.urlopen(req_ws, timeout=0.8) as resp_ws:
        workspaces = json.loads(resp_ws.read().decode("utf-8"))

    # Cache list of indexed workspace names for fast PreToolUse lookup
    try:
        indexed_names = [w.get("workspace") for w in workspaces if w.get("total_nodes", 0) > 0]
        for w in workspaces:
            for wt in w.get("worktrees", []):
                if wt.get("total_nodes", 0) > 0:
                    indexed_names.append(wt.get("workspace"))
        with open("/tmp/omni_indexed_workspaces.json", "w") as f:
            json.dump(indexed_names, f)
    except Exception:
        pass

    matched_ws = None
    node_count = 0
    matched_entry = None

    norm_active = os.path.realpath(target_dir).lower()
    for w in workspaces:
        root_path = w.get("root_path", "")
        norm_root = os.path.realpath(root_path).lower() if root_path else ""
        if norm_root and (norm_active == norm_root or norm_active.startswith(norm_root + "/")):
            matched_ws = w.get("workspace")
            node_count = w.get("total_nodes", 0)
            matched_entry = w
            break
        for wt in w.get("worktrees", []):
            wt_root = wt.get("root_path", "")
            norm_wt_root = os.path.realpath(wt_root).lower() if wt_root else ""
            if norm_wt_root and (norm_active == norm_wt_root or norm_active.startswith(norm_wt_root + "/")):
                matched_ws = wt.get("workspace")
                node_count = wt.get("total_nodes", 0)
                matched_entry = wt
                break
        if matched_ws:
            break

    if not matched_ws:
        for w in workspaces:
            if w.get("workspace", "").lower() == ws_candidate.lower():
                matched_ws = w.get("workspace")
                node_count = w.get("total_nodes", 0)
                matched_entry = w
                break
            for wt in w.get("worktrees", []):
                if wt.get("workspace", "").lower() == ws_candidate.lower():
                    matched_ws = wt.get("workspace")
                    node_count = wt.get("total_nodes", 0)
                    matched_entry = wt
                    break
            if matched_ws:
                break

    if not matched_ws:
        matched_ws = ws_candidate

    if node_count > 0:
        try:
            with open(f"/tmp/omni_indexed_{matched_ws}.flag", "w") as f:
                f.write(str(node_count))
        except Exception:
            pass

        is_watching = matched_entry.get("is_watching", False) if matched_entry else False
        if not is_watching:
            try:
                watch_payload = json.dumps({"workspace": matched_ws, "path": target_dir}).encode("utf-8")
                watch_req = urllib.request.Request(
                    "http://localhost:8080/api/watch/start",
                    data=watch_payload,
                    headers={"Content-Type": "application/json", "User-Agent": "OmniHook/1.0"}
                )
                urllib.request.urlopen(watch_req, timeout=0.8)
                is_watching = True
            except Exception:
                pass

        wt_suffix = f" (worktree: {wt_info['branch']} -> {wt_info['parent_workspace']})" if wt_info else ""
        watch_str = " | live-watch: ON" if is_watching else ""
        msg = (
            f"🧭 Omni-Graph | ws: {matched_ws}{wt_suffix} ({node_count:,} nodes{watch_str})\n"
            f"Endpoints (:8080/api): /query (Graph-RAG) | /galaxies | /condense?symbol=X | /references?symbol=X | /symbol?name=X | /search?q=X\n"
            f"[Gate: direct source reads >150 lines blocked — query AST endpoints first]"
        )
    else:
        try:
            flag = f"/tmp/omni_indexed_{matched_ws}.flag"
            if os.path.exists(flag):
                os.remove(flag)
        except Exception:
            pass

        if wt_info:
            msg = (
                f"🧭 Omni-Graph | unindexed worktree: {matched_ws} ({wt_info['branch']} -> {wt_info['parent_workspace']})\n"
                f"• Ingest: curl -s -X POST http://localhost:8080/api/ingest -H 'Content-Type: application/json' -d '{{\"path\": \"{target_dir}\", \"watch\": true}}'\n"
                f"[Auto-inherits galaxy seeds from parent: {wt_info['parent_workspace']}]"
            )
        else:
            msg = (
                f"🧭 Omni-Graph | ws: {matched_ws} (unindexed)\n"
                f"• Ingest: curl -s -X POST http://localhost:8080/api/ingest -H 'Content-Type: application/json' -d '{{\"path\": \"{target_dir}\", \"watch\": true}}'\n"
                f"[Note: Index once; live-watch auto-syncs. Never re-ingest an active workspace.]"
            )

    print(json.dumps({
        "injectSteps": [
            {
                "ephemeralMessage": msg
            }
        ]
    }))

except Exception:
    print(json.dumps({"injectSteps": []}))
EOF
