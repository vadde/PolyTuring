#!/usr/bin/env bash
# ==============================================================================
# Omni-Graph PreToolUse Guardrail Hook (Zero-Fluff / High-Density)
# Intercepts brute-force grep, un-reconnoitered source access, and unbounded
# source file reads (>150 lines). Enforces Session-Scoped Recon Gate.
# ==============================================================================

set -e

PAYLOAD=$(cat)

python3 - "$PAYLOAD" << 'EOF'
import sys, json, os, urllib.request, time, subprocess, re

try:
    # 0. Parse Tool Call Payload
    raw_input = sys.argv[1] if len(sys.argv) > 1 else sys.stdin.read()
    data = json.loads(raw_input) if raw_input.strip() else {}
    tool_call = data.get("toolCall", {})
    tool_name = tool_call.get("name", "")
    args = tool_call.get("args", {})

    # Hard Block: browser_subagent causes language server IPC deadlocks/crashes
    if tool_name == "browser_subagent":
        print(json.dumps({
            "decision": "deny",
            "reason": "BLOCKED: browser_subagent is disabled (causes LS crashes). Use Playwright MCP or run_command."
        }))
        sys.exit(0)

    # 1. Daemon Liveness Check (Fail-Open Safeguard)
    try:
        req = urllib.request.Request("http://localhost:8080/api/health", headers={"User-Agent": "OmniHook/1.0"})
        with urllib.request.urlopen(req, timeout=0.5) as resp:
            if resp.status != 200:
                print(json.dumps({"decision": "allow"}))
                sys.exit(0)
    except Exception:
        print(json.dumps({"decision": "allow"}))
        sys.exit(0)

    RECON_TTL = 7200  # 2 hours

    def is_workspace_indexed(workspace):
        if not workspace:
            return False
        if os.path.exists(f"/tmp/omni_indexed_{workspace}.flag"):
            return True
        try:
            with open("/tmp/omni_indexed_workspaces.json", "r") as f:
                indexed = json.load(f)
                return workspace in indexed
        except Exception:
            return workspace in {"PolyTuring", "tool-scripts", "session-explorer", "k8s-eks", "python", "tutor-intelligence", "DSA"}

    def get_workspace(target_path=""):
        ws = os.environ.get("WORKSPACE")
        if ws:
            return ws

        if target_path:
            norm = target_path.replace("\\", "/")
            for part in reversed(norm.split("/")):
                if part and is_workspace_indexed(part):
                    return part

        for k in ("Cwd", "SearchPath", "TargetFile", "AbsolutePath"):
            val = args.get(k, "")
            if val:
                norm = val.replace("\\", "/")
                for part in reversed(norm.split("/")):
                    if part and is_workspace_indexed(part):
                        return part

        ws_paths = data.get("workspacePaths", [])
        if ws_paths:
            for wp in ws_paths:
                norm = wp.replace("\\", "/")
                for part in reversed(norm.split("/")):
                    if part and is_workspace_indexed(part):
                        return part
                base = os.path.basename(wp)
                if base:
                    return base

        base_dir = ws_paths[0] if ws_paths else (args.get("Cwd") or "")
        try:
            git_cmd = ["git", "-C", base_dir, "rev-parse", "--show-toplevel"] if base_dir else ["git", "rev-parse", "--show-toplevel"]
            git_root = subprocess.check_output(
                git_cmd,
                stderr=subprocess.DEVNULL,
                encoding="utf-8"
            ).strip()
            return os.path.basename(git_root)
        except Exception:
            if base_dir:
                return os.path.basename(base_dir)
            return os.path.basename(os.getcwd()) or "default"

    def is_recon_done(workspace):
        keys = [workspace] if workspace and workspace != "default" else ["default"]
        for ws_key in keys:
            marker = f"/tmp/omni_recon_{ws_key}.marker"
            try:
                with open(marker, "r") as f:
                    ts = int(f.read().strip())
                if (int(time.time()) - ts) < RECON_TTL:
                    return True
            except Exception:
                continue
        return False

    def record_recon(workspace):
        now = str(int(time.time()))
        for ws_key in filter(None, [workspace]):
            try:
                with open(f"/tmp/omni_recon_{ws_key}.marker", "w") as f:
                    f.write(now)
            except Exception:
                pass

    NON_CODE_EXTS = {
        ".md", ".markdown", ".mdown", ".mkd", ".txt", ".text",
        ".json", ".json5", ".yaml", ".yml", ".toml", ".ini", ".cfg", ".conf",
        ".csv", ".tsv", ".svg", ".png", ".jpg", ".jpeg", ".gif", ".ico", ".webp",
        ".lock", ".sum", ".env", ".example", ".sample", ".pdf", ".mp4", ".mov",
        ".log", ".tmp", ".bak", ".patch", ".diff", ".pid", ".socket", ".wasm",
        ".bin", ".exe", ".o", ".a", ".so", ".dylib", ".dll", ".class", ".jar"
    }

    NON_CODE_FILENAMES = {
        "makefile", "dockerfile", "docker-compose.yml", "docker-compose.yaml",
        "license", "licence", "readme", "notice", "authors", "contributors",
        "cargo.lock", "package-lock.json", "pnpm-lock.yaml", "yarn.lock",
        "go.sum", "gemfile.lock", "poetry.lock", "mix.lock"
    }

    def is_source_code(file_path):
        if not file_path:
            return False
        normalized = file_path.replace("\\", "/")
        if "/.agents/" in normalized or "/.gemini/" in normalized or "/.git/" in normalized:
            return False
        base = os.path.basename(file_path).lower()
        if base in NON_CODE_FILENAMES:
            return False
        _, ext = os.path.splitext(base)
        if ext in NON_CODE_EXTS:
            return False
        return True

    # 3. Intercept run_command
    if tool_name == "run_command":
        cmd = args.get("CommandLine", "")

        omni_patterns = [
            "localhost:8080/api/symbol", "127.0.0.1:8080/api/symbol",
            "localhost:8080/api/condense", "127.0.0.1:8080/api/condense",
            "localhost:8080/api/references", "127.0.0.1:8080/api/references",
            "localhost:8080/api/search", "127.0.0.1:8080/api/search",
            "localhost:8080/api/galaxy", "127.0.0.1:8080/api/galaxy",
            "localhost:8080/api/ast", "127.0.0.1:8080/api/ast",
            "localhost:8080/api/workspaces", "127.0.0.1:8080/api/workspaces",
            "localhost:8080/api/ingest", "127.0.0.1:8080/api/ingest",
            "make graph-", "make search-graph", "make query-graph",
            "omni.sh", "make workspaces"
        ]
        if any(p in cmd for p in omni_patterns):
            ws = get_workspace()
            record_recon(ws)
            ws_match = re.search(r'(?:workspace|PROJECT)=([a-zA-Z0-9_-]+)', cmd)
            if ws_match:
                record_recon(ws_match.group(1))

        blocked_patterns = [
            "grep -r", "grep -rn", "grep -ri", "grep -rin", "grep -rIn",
            "find . -name", "cat $(find", "ag -l", "rg -l"
        ]
        if any(p in cmd for p in blocked_patterns):
            print(json.dumps({
                "decision": "deny",
                "reason": "BLOCKED: Blind recursive search. Query Omni-Graph: :8080/api/search?q=X or :8080/api/symbol?name=X."
            }))
            sys.exit(0)

    # 4. Intercept view_file
    elif tool_name == "view_file":
        target_path = args.get("AbsolutePath", "")
        ws = get_workspace(target_path)
        if is_source_code(target_path):
            if is_workspace_indexed(ws) and not is_recon_done(ws):
                print(json.dumps({
                    "decision": "deny",
                    "reason": f"🧭 Omni-Graph recon required for {os.path.basename(target_path)} (ws: {ws}). Query :8080/api/symbol?name=X, /condense?symbol=X, or /search?q=X first."
                }))
                sys.exit(0)

            start_line = args.get("StartLine")
            end_line = args.get("EndLine")

            if end_line is None:
                print(json.dumps({
                    "decision": "deny",
                    "reason": "BLOCKED: Direct reading >150 lines of source code exhausts context. Query :8080/api/symbol?name=X or /condense?symbol=X, or specify EndLine <= 150."
                }))
                sys.exit(0)

            try:
                start = int(start_line) if start_line is not None else 1
                end = int(end_line)
                span = end - start + 1
                if span > 150:
                    print(json.dumps({
                        "decision": "deny",
                        "reason": f"BLOCKED: Slice of {span} lines exceeds the 150-line gate. Query :8080/api/symbol?name=X or narrow slice to <= 150 lines."
                    }))
                    sys.exit(0)
            except (ValueError, TypeError):
                pass

        elif target_path and any(target_path.lower().endswith(ext) for ext in [".md", ".markdown", ".mdown", ".mkd"]):
            start_line = args.get("StartLine")
            end_line = args.get("EndLine")

            if end_line is None:
                try:
                    with open(target_path, "r", encoding="utf-8", errors="ignore") as f:
                        line_count = sum(1 for _ in f)
                    if line_count > 450:
                        print(json.dumps({
                            "decision": "deny",
                            "reason": f"BLOCKED: Markdown '{os.path.basename(target_path)}' has {line_count} lines (>450 ceiling). Query :8080/api/search?q=X&workspace={ws} or narrow slice to <= 450 lines."
                        }))
                        sys.exit(0)
                except Exception:
                    pass
            else:
                try:
                    start = int(start_line) if start_line is not None else 1
                    end = int(end_line)
                    span = end - start + 1
                    if span > 450:
                        print(json.dumps({
                            "decision": "deny",
                            "reason": f"BLOCKED: Slice of {span} lines exceeds 450-line documentation limit. Narrow slice to <= 450 lines or query :8080/api/search."
                        }))
                        sys.exit(0)
                except (ValueError, TypeError):
                    pass

    # 5. Intercept grep_search
    elif tool_name == "grep_search":
        search_path = args.get("SearchPath", "")
        ws = get_workspace(search_path)
        query = args.get("Query", "").strip()
        includes = args.get("Includes", [])

        is_markdown_grep = any(inc.lower().endswith(ext) for inc in includes for ext in [".md", ".markdown"]) if includes else False
        if is_markdown_grep and is_workspace_indexed(ws) and not is_recon_done(ws):
            print(json.dumps({
                "decision": "deny",
                "reason": f"🧭 Omni-Graph recon recommended before searching docs in ws: {ws}. Query :8080/api/search?q={query}&workspace={ws}."
            }))
            sys.exit(0)

        is_non_code_search = False
        normalized_path = search_path.replace("\\", "/")
        if "/.agents" in normalized_path or "/.gemini" in normalized_path or "/.git" in normalized_path:
            is_non_code_search = True
        elif search_path and os.path.isfile(search_path) and not is_source_code(search_path):
            is_non_code_search = True
        elif includes and all(any(inc.lower().endswith(ext) for ext in NON_CODE_EXTS) for inc in includes):
            is_non_code_search = True

        if not is_non_code_search and is_workspace_indexed(ws) and not is_recon_done(ws):
            print(json.dumps({
                "decision": "deny",
                "reason": f"🧭 Omni-Graph recon required for ws: {ws}. Query :8080/api/symbol?name=X, /condense?symbol=X, or /search?q={query}&workspace={ws}."
            }))
            sys.exit(0)

        if len(query) < 2:
            print(json.dumps({
                "decision": "deny",
                "reason": "BLOCKED: Empty/broad grep query. Query :8080/api/symbol?name=X or :8080/api/search?q=X."
            }))
            sys.exit(0)

    print(json.dumps({"decision": "allow"}))

except Exception:
    print(json.dumps({"decision": "allow"}))
EOF
