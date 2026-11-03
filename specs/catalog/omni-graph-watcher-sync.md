# 📐 Specification: Omni-Graph Watcher Telemetry Synchronization & Idempotency Hardening

> **Status**: `in-progress`  
> **Author**: Antigravity (Google DeepMind) — Pair Programming Architect  
> **Parent Spec**: [`specs/catalog/omni-graph.md`](omni-graph.md) & [`specs/catalog/omni-graph-nextgen.md`](omni-graph-nextgen.md)  
> **Feature Branch**: `feat/omni-graph-nextgen-hardening-v0.3.0`  
> **Tool Path**: `tools/omni-graph/`  
> **Date**: 2026-11-03  

---

## 1. Problem Statement

Forensic analysis of the live multi-container Omni-Graph stack revealed that `omni-rust-app` is bombarded with continuous `POST /api/watch/start` requests on **every single agent and user interaction turn**. 

In the container logs (`docker logs omni-rust-app`), telemetry shows:
```
INFO omni_graph::db: Recorded agent_api_call for endpoint /api/watch/start
INFO omni_graph::watcher: Watcher for 'PolyTuring' received cancellation signal
INFO omni_graph::watcher: Started watching workspace 'PolyTuring' at path '...'
```
This loop occurs dozens of times per session, generating redundant `agent_api_call` rows in SurrealDB, repeatedly cancelling and re-spawning FSEvents watcher tasks, and causing unnecessary filesystem scanning.

---

## 2. Root Cause Analysis (Forensic Trace)

The problem is caused by **telemetry desynchronization between persistent storage and in-memory runtime state**:

1. **Storage-Runtime Disconnect in `GET /api/workspaces`**:
   - In `tools/omni-graph/src/api/mod.rs:262`, `workspaces_handler` delegates directly to `state.db.get_workspaces().await`.
   - `state.db` queries SurrealDB (`workspace_meta` and `node` tables). SurrealDB stores persistent file lists and node counts, but **has no access to the in-memory `WatchManager` state** (`state.watcher`).
   - Consequently, `GET /api/workspaces` returns workspace records with `total_nodes`, `files`, `languages`, and `root_path`, but **omits `is_watching`** (or evaluates to `None`).
2. **Hook Script Assumption in `hook_pre_invocation.sh`**:
   - On every agent prompt invocation, `hook_pre_invocation.sh` inspects the active workspace from `GET /api/workspaces`.
   - It checks:
     ```python
     is_watching = matched_entry.get("is_watching", False) if matched_entry else False
     if not is_watching:
         # Trigger POST /api/watch/start
     ```
   - Because `matched_entry.get("is_watching")` is always `None` / `False`, the hook believes live-watching is offline and fires `POST /api/watch/start` on every turn.
3. **Watcher Path Normalization Mismatch in `start_watch`**:
   - In `tools/omni-graph/src/watcher/mod.rs:126`:
     ```rust
     if let Some(active) = watchers.get(workspace) {
         let s = active.status.read().await;
         if s.path == path && s.status == "watching" {
             return Ok(s.clone());
         }
     }
     ```
   - If the path sent by the hook differs even slightly (e.g., trailing slash, symlink resolution, canonical path vs container mounted path), `s.path == path` evaluates to `false`.
   - Instead of recognizing that the workspace is already actively monitored, `start_watch` replaces the watcher entry, drops the old cancellation receiver, and spawns a new `tokio::spawn(run_watcher)` loop.

---

## 3. Requirements & Technical Specifications

| Req ID | Requirement | Priority | Verification Criteria |
| :--- | :--- | :---: | :--- |
| **R-073** | **Dynamic Watcher Status Merging in `workspaces_handler`**: `workspaces_handler` in `api/mod.rs` must fetch live watcher statuses from `state.watcher.get_status().await` and dynamically inject `"is_watching": true/false` and `"watcher_status": {...}` into each workspace record before returning JSON. | P0 | `GET /api/workspaces` returns `"is_watching": true` for actively watched codebases. |
| **R-074** | **Canonical Path Normalization in `WatcherManager::start_watch`**: `WatcherManager` must canonicalize paths (resolving symlinks and stripping trailing slashes) before path comparison, ensuring that path string formatting variants do not cause false cache misses. | P0 | Multiple `start_watch` calls with varying trailing slashes return the existing watcher without restart. |
| **R-075** | **Idempotent No-Op on Active Watcher**: If a watcher is already running for a workspace, `start_watch` must immediately return the existing status and **MUST NOT** re-spawn tasks, cancel existing loops, or record duplicate telemetry entries. | P0 | Zero `Watcher received cancellation signal` logs during repetitive `start_watch` calls for the same workspace. |
| **R-076** | **Worktree Fleet Watcher Status Parity**: Nested worktrees inside parent workspace records must also receive accurate `"is_watching": true/false` fields matching their active watcher entries. | P1 | `GET /api/workspaces` reflects accurate watching status for both root workspaces and nested worktrees. |

---

## 4. Blast Radius & Multi-User Development Considerations

Because other developers and coding assistants use this repository and rely on Omni-Graph:
- **Zero Breaking Changes to API Contracts**:
  - `GET /api/workspaces` retains all existing fields (`workspace`, `total_nodes`, `root_path`, `languages`, `files`, `worktrees`). The added fields (`is_watching`, `watcher_status`) are purely additive.
  - `POST /api/watch/start` retains its exact JSON request/response schema.
- **Safety Guardrail**:
  - If `state.watcher.get_status().await` encounters any lock delay or timeout, it falls back to the database response with `"is_watching": false` without crashing the endpoint.
- **Agent Skill Stability**:
  - `.agents/skills/omni-graph/SKILL.md` already instructs agents to check `"is_watching": true`. Once `GET /api/workspaces` provides this field accurately, agents will naturally respect the invariant.

---

## 5. Step-by-Step Test & Verification Plan

1. **Unit Tests (`cargo test`)**:
   - Test canonical path normalization in `WatcherManager::start_watch`.
   - Test that calling `start_watch` twice with the same workspace returns the exact same watcher instance without cancellation.
2. **Integration Verification via Docker / Curl**:
   - Rebuild or update the binary and verify `GET /api/workspaces`:
     ```bash
     curl -s http://localhost:8080/api/workspaces | jq '.[].is_watching'
     ```
     Must return `true` for `PolyTuring`.
   - Verify `hook_pre_invocation.sh` behavior:
     - Run `hook_pre_invocation.sh` 5 times in a row.
     - Inspect `docker logs omni-rust-app`: Confirm ZERO `Recorded agent_api_call for endpoint /api/watch/start` and ZERO watcher cancellations.
