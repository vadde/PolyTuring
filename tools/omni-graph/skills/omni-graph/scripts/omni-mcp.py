#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.10"
# dependencies = []
# ///
"""
omni-mcp.py — Native Model Context Protocol (MCP) stdio server for Omni-Graph.
Implements: R-070, R-078 (JSON-RPC 2.0 stdio protocol 2024-11-05).

Provides structured zero-token shell-free AST queries for AI coding assistants:
- omni_definition: Lookup precise AST definition and metadata
- omni_references: Find all callers and incoming dependency edges
- omni_condense: Extract multi-hop call trace slice (<1500 tokens)
- omni_search: Dense vector semantic search across code symbols
- omni_galaxies: Inspect architectural subsystems and coupling metrics
- omni_query: Execute hybrid Graph-RAG retrieval (seeds + AST + community)
"""

import sys
import os
import json
import urllib.request
import urllib.parse
import subprocess

API_URL = os.environ.get("OMNI_API_URL", "http://localhost:8080")

def get_default_workspace() -> str:
    env_ws = os.environ.get("OMNI_WORKSPACE")
    if env_ws:
        return env_ws.rstrip("/").split("/")[-1]
    try:
        git_root = subprocess.check_output(
            ["git", "rev-parse", "--show-toplevel"],
            stderr=subprocess.DEVNULL,
            text=True
        ).strip()
        if git_root:
            return os.path.basename(git_root)
    except Exception:
        pass
    return os.path.basename(os.getcwd())

TOOLS_SPEC = [
    {
        "name": "omni_definition",
        "description": "Lookup precise AST definition, signatures, and file line ranges for a code symbol. Essential before editing code.",
        "inputSchema": {
            "type": "object",
            "properties": {
                "symbol": {
                    "type": "string",
                    "description": "The exact identifier name of the function, struct, class, method, or interface."
                },
                "workspace": {
                    "type": "string",
                    "description": "Target workspace name (defaults to current project root)."
                }
            },
            "required": ["symbol"]
        }
    },
    {
        "name": "omni_references",
        "description": "Find all callers, references, and incoming dependency edges for a symbol across the entire codebase.",
        "inputSchema": {
            "type": "object",
            "properties": {
                "symbol": {
                    "type": "string",
                    "description": "The target symbol name to find references for."
                },
                "workspace": {
                    "type": "string",
                    "description": "Target workspace name (defaults to current project root)."
                }
            },
            "required": ["symbol"]
        }
    },
    {
        "name": "omni_condense",
        "description": "Extract a strictly budgeted (<1500 tokens) multi-hop AST call slice around a symbol. Contains full function signatures, callers, callees, and dependency edges without reading thousands of file lines.",
        "inputSchema": {
            "type": "object",
            "properties": {
                "symbol": {
                    "type": "string",
                    "description": "Root symbol to center the call slice around."
                },
                "hops": {
                    "type": "integer",
                    "description": "Call graph traversal radius (1 or 2 hops, default 2)."
                },
                "workspace": {
                    "type": "string",
                    "description": "Target workspace name (defaults to current project root)."
                }
            },
            "required": ["symbol"]
        }
    },
    {
        "name": "omni_search",
        "description": "Perform dense vector semantic search across indexed code symbols to locate relevant files and implementations.",
        "inputSchema": {
            "type": "object",
            "properties": {
                "query": {
                    "type": "string",
                    "description": "Natural language or technical question/search query."
                },
                "k": {
                    "type": "integer",
                    "description": "Number of top results to return (default 10)."
                },
                "workspace": {
                    "type": "string",
                    "description": "Target workspace name (defaults to current project root)."
                },
                "include_docs": {
                    "type": "boolean",
                    "description": "Whether to include markdown documentation nodes (default false: code-only)."
                }
            },
            "required": ["query"]
        }
    },
    {
        "name": "omni_galaxies",
        "description": "Inspect macro architectural subsystems (galaxies), modular boundaries, instability scores, and cross-subsystem coupling metrics.",
        "inputSchema": {
            "type": "object",
            "properties": {
                "workspace": {
                    "type": "string",
                    "description": "Target workspace name (defaults to current project root)."
                },
                "limit": {
                    "type": "integer",
                    "description": "Maximum number of subsystems to return (default 15)."
                }
            }
        }
    },
    {
        "name": "omni_query",
        "description": "Execute hybrid Graph-RAG retrieval combining vector search, community summaries, and multi-hop AST traces into a grounded Markdown brief.",
        "inputSchema": {
            "type": "object",
            "properties": {
                "prompt": {
                    "type": "string",
                    "description": "Architectural or implementation question to synthesize."
                },
                "workspace": {
                    "type": "string",
                    "description": "Target workspace name (defaults to current project root)."
                },
                "top_k": {
                    "type": "integer",
                    "description": "Number of top AST seed hits (default 5)."
                },
                "include_docs": {
                    "type": "boolean",
                    "description": "Whether to include documentation in search seeds (default false: code-only)."
                }
            },
            "required": ["prompt"]
        }
    }
]

def make_http_request(url: str, data: dict = None, timeout: int = 15):
    headers = {"Content-Type": "application/json"}
    body = json.dumps(data).encode("utf-8") if data is not None else None
    req = urllib.request.Request(url, data=body, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        if e.code == 423:
            try:
                err_data = json.loads(e.read().decode("utf-8"))
                return {
                    "error": "HTTP 423 Locked",
                    "message": err_data.get("message", "Workspace is actively synchronizing graph topology."),
                    "locked": True,
                }
            except Exception:
                return {
                    "error": "HTTP 423 Locked",
                    "message": "Workspace is actively synchronizing graph topology. Please retry in a moment.",
                    "locked": True,
                }
        raise

def handle_tool_call(name: str, args: dict) -> dict:
    workspace = args.get("workspace") or get_default_workspace()
    
    try:
        if name == "omni_definition":
            symbol = args["symbol"]
            params = urllib.parse.urlencode({"name": symbol, "workspace": workspace, "wait_for_sync": "true"})
            res = make_http_request(f"{API_URL}/api/symbol?{params}")
            if res.get("locked"):
                return {"content": [{"type": "text", "text": f"⏳ Workspace '{workspace}' is synchronizing graph topology. {res.get('message', '')}"}]}
            matches = res.get("matches", [])
            if not matches:
                return {"content": [{"type": "text", "text": f"Symbol '{symbol}' not found in workspace '{workspace}'."}]}
            
            lines = [f"Found {len(matches)} match(es) for '{symbol}' in workspace '{workspace}':\n"]
            for m in matches:
                lines.append(f"• [{m.get('kind', 'symbol')}] `{m.get('label')}` in `{m.get('file_path')}:{m.get('line_start')}-{m.get('line_end')}` ({m.get('language')})")
                text_preview = m.get('text', '').strip()
                if text_preview:
                    lines.append(f"  ```{m.get('language', '')}\n  {text_preview[:400]}\n  ```")
            return {"content": [{"type": "text", "text": "\n".join(lines)}]}

        elif name == "omni_references":
            symbol = args["symbol"]
            params = urllib.parse.urlencode({"symbol": symbol, "workspace": workspace, "wait_for_sync": "true"})
            res = make_http_request(f"{API_URL}/api/references?{params}")
            if res.get("locked"):
                return {"content": [{"type": "text", "text": f"⏳ Workspace '{workspace}' is synchronizing graph topology. {res.get('message', '')}"}]}
            raw_refs = res.get("references", [])
            # Also check legacy callers/importers as fallback if present
            if not raw_refs and ("callers" in res or "importers" in res):
                raw_refs = res.get("callers", []) + res.get("importers", [])
            
            total = len(raw_refs)
            if total == 0:
                return {"content": [{"type": "text", "text": f"No incoming references found for '{symbol}' in workspace '{workspace}'."}]}
            
            lines = [f"Total References for '{symbol}' ({total}):\n"]
            
            callers = []
            containment = []
            doc_refs = []
            inheritance = []
            others = []
            
            for ref in raw_refs:
                edge_type = ref.get("type", "CALLS")
                label = ref.get("label", ref.get("source_symbol", "unknown"))
                file_path = ref.get("file_path", ref.get("source_file", ""))
                line_start = ref.get("line_start", 0)
                kind = ref.get("kind", "")
                loc = f"`{file_path}:{line_start}`" if line_start else f"`{file_path}`"
                entry = f"  ← `{label}` ({kind}) in {loc} [{edge_type}]" if kind else f"  ← `{label}` in {loc} [{edge_type}]"
                
                if edge_type == "CALLS":
                    callers.append(entry)
                elif edge_type in ("CONTAINS", "CONTAINS_FILE", "CONTAINS_DIR"):
                    containment.append(entry)
                elif edge_type == "REFERENCES" or ref.get("language") == "markdown":
                    doc_refs.append(entry)
                elif edge_type in ("IMPLEMENTS", "EXTENDS"):
                    inheritance.append(entry)
                else:
                    others.append(entry)
            
            if callers:
                lines.append(f"Direct Callers ({len(callers)}):")
                lines.extend(callers[:20])
            if inheritance:
                lines.append(f"\nInheritance & Interfaces ({len(inheritance)}):")
                lines.extend(inheritance[:15])
            if containment:
                lines.append(f"\nContainment & Hierarchy ({len(containment)}):")
                lines.extend(containment[:15])
            if doc_refs:
                lines.append(f"\nDocument References ({len(doc_refs)}):")
                lines.extend(doc_refs[:10])
            if others:
                lines.append(f"\nOther References ({len(others)}):")
                lines.extend(others[:15])
                
            return {"content": [{"type": "text", "text": "\n".join(lines)}]}

        elif name == "omni_condense":
            symbol = args["symbol"]
            hops = args.get("hops", 2)
            params = urllib.parse.urlencode({"symbol": symbol, "workspace": workspace, "hops": hops, "wait_for_sync": "true"})
            res = make_http_request(f"{API_URL}/api/condense?{params}")
            if res.get("locked"):
                return {"content": [{"type": "text", "text": f"⏳ Workspace '{workspace}' is synchronizing graph topology. {res.get('message', '')}"}]}
            formatted_md = res.get("formatted_markdown", "")
            tokens = res.get("token_estimate", 0)
            text = f"{formatted_md}\n\n*(Token estimate: ~{tokens} tokens)*"
            return {"content": [{"type": "text", "text": text}]}

        elif name == "omni_search":
            query = args["query"]
            k = args.get("k", 10)
            include_docs = "true" if args.get("include_docs", False) else "false"
            params = urllib.parse.urlencode({"q": query, "workspace": workspace, "k": k, "include_docs": include_docs, "wait_for_sync": "true"})
            res = make_http_request(f"{API_URL}/api/search?{params}")
            if isinstance(res, dict) and res.get("locked"):
                return {"content": [{"type": "text", "text": f"⏳ Workspace '{workspace}' is synchronizing graph topology. {res.get('message', '')}"}]}
            hits = res.get("results", []) if isinstance(res, dict) else res
            if not hits:
                return {"content": [{"type": "text", "text": f"No matches found for query: '{query}'"}]}
            
            lines = [f"Semantic search results for '{query}' (top {len(hits)}):\n"]
            for hit in hits:
                sim = hit.get("similarity", 0.0)
                lines.append(f"• `{hit.get('label')}` ({hit.get('kind')}) in `{hit.get('file_path')}:{hit.get('line_start')}` [cos: {sim:.3f}]")
                sig = hit.get("text", "").splitlines()
                first_line = sig[0].strip() if sig else ""
                if first_line:
                    lines.append(f"  Signature: `{first_line[:120]}`")
            return {"content": [{"type": "text", "text": "\n".join(lines)}]}

        elif name == "omni_galaxies":
            limit = args.get("limit", 15)
            params = urllib.parse.urlencode({"workspace": workspace, "limit": limit, "wait_for_sync": "true"})
            res = make_http_request(f"{API_URL}/api/galaxies?{params}")
            if res.get("locked"):
                return {"content": [{"type": "text", "text": f"⏳ Workspace '{workspace}' is synchronizing graph topology. {res.get('message', '')}"}]}
            galaxies = res.get("galaxies", [])
            total = res.get("total_galaxies", len(galaxies))
            
            lines = [f"Architectural Galaxies for '{workspace}' (showing {len(galaxies)} of {total}):\n"]
            for g in galaxies:
                role = g.get("role", "Service")
                nodes = g.get("node_count", 0)
                ca = g.get("afferent_coupling", 0)
                ce = g.get("efferent_coupling", 0)
                inst = g.get("instability", 0.0)
                lines.append(f"🌌 **{g.get('name')}** [{role}] — {nodes} nodes | Ca: {ca}, Ce: {ce}, Instability: {inst:.2f}")
                syms = g.get("sample_symbols", [])
                if syms:
                    lines.append(f"   Key symbols: {', '.join(syms[:6])}")
            return {"content": [{"type": "text", "text": "\n".join(lines)}]}

        elif name == "omni_query":
            prompt = args["prompt"]
            include_docs = args.get("include_docs", False)
            top_k = args.get("top_k", 5)
            payload = {"prompt": prompt, "workspace": workspace, "include_docs": include_docs, "top_k": top_k, "wait_for_sync": True}
            res = make_http_request(f"{API_URL}/api/query", data=payload)
            if res.get("locked"):
                return {"content": [{"type": "text", "text": f"⏳ Workspace '{workspace}' is synchronizing graph topology. {res.get('message', '')}"}]}
            macro = res.get("macroscopic_summary", "")
            subgraph = res.get("expanded_subgraph", "")
            return {"content": [{"type": "text", "text": f"{macro}\n\n{subgraph}"}]}

        else:
            return {"isError": True, "content": [{"type": "text", "text": f"Unknown tool: '{name}'"}]}

    except urllib.error.URLError as e:
        return {
            "isError": True,
            "content": [
                {
                    "type": "text",
                    "text": f"Omni-Graph API connection error ({API_URL}): {e.reason}. Ensure Omni-Graph is running (`make omni-graph`)."
                }
            ]
        }
    except Exception as e:
        return {"isError": True, "content": [{"type": "text", "text": f"Tool execution failed: {str(e)}"}]}

def main():
    while True:
        line = sys.stdin.readline()
        if not line:
            break
        line = line.strip()
        if not line:
            continue

        try:
            req = json.loads(line)
        except Exception:
            continue

        msg_id = req.get("id")
        method = req.get("method")
        params = req.get("params", {})

        if method == "initialize":
            res = {
                "jsonrpc": "2.0",
                "id": msg_id,
                "result": {
                    "protocolVersion": "2024-11-05",
                    "capabilities": {
                        "tools": {}
                    },
                    "serverInfo": {
                        "name": "omni-graph-mcp",
                        "version": "0.3.0"
                    }
                }
            }
            sys.stdout.write(json.dumps(res) + "\n")
            sys.stdout.flush()

        elif method == "notifications/initialized":
            # No-op acknowledgement
            pass

        elif method == "tools/list":
            res = {
                "jsonrpc": "2.0",
                "id": msg_id,
                "result": {
                    "tools": TOOLS_SPEC
                }
            }
            sys.stdout.write(json.dumps(res) + "\n")
            sys.stdout.flush()

        elif method == "tools/call":
            tool_name = params.get("name", "")
            tool_args = params.get("arguments", {})
            call_res = handle_tool_call(tool_name, tool_args)
            res = {
                "jsonrpc": "2.0",
                "id": msg_id,
                "result": call_res
            }
            sys.stdout.write(json.dumps(res) + "\n")
            sys.stdout.flush()

        elif method == "ping":
            res = {"jsonrpc": "2.0", "id": msg_id, "result": {}}
            sys.stdout.write(json.dumps(res) + "\n")
            sys.stdout.flush()

        else:
            if msg_id is not None:
                res = {
                    "jsonrpc": "2.0",
                    "id": msg_id,
                    "error": {
                        "code": -32601,
                        "message": f"Method not found: {method}"
                    }
                }
                sys.stdout.write(json.dumps(res) + "\n")
                sys.stdout.flush()

if __name__ == "__main__":
    main()
