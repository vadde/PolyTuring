// Package main provides the HTTP server with embedded static assets.
//
// Implements: R-006 (Serve embedded web UI), NF-008 (Bind to localhost only)
// See: specs/catalog/session-explorer.md
package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"
)

//go:embed all:web/dist
var staticFiles embed.FS

//go:embed openapi.json
var openapiSpec []byte

// StartServer creates and starts the HTTP server.
// Implements: R-006, NF-008, R-030, R-032, R-033
func StartServer(cfg Config, index *SessionIndex) error {
	api := NewAPIHandler(index, cfg.Verbose)
	api.SetWatchInterval(cfg.WatchInterval)

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		// Route to detail if path has an ID segment
		path := r.URL.Path
		if path != "/api/sessions" && path != "/api/sessions/" {
			if strings.HasSuffix(path, "/tail") {
				api.HandleSessionTail(w, r)
				return
			}
			api.HandleSessionDetail(w, r)
			return
		}
		api.HandleSessions(w, r)
	})
	mux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/tail") {
			api.HandleSessionTail(w, r)
			return
		}
		api.HandleSessionDetail(w, r)
	})
	mux.HandleFunc("/api/search", api.HandleSearch)
	mux.HandleFunc("/api/stats", api.HandleStats)
	mux.HandleFunc("/api/refresh", api.HandleRefresh)
	mux.HandleFunc("/api/watch-interval", api.HandleWatchInterval)
	mux.HandleFunc("/api/wip", api.HandleWIP)
	mux.HandleFunc("/api/skill/install", api.HandleInstallSkill)

	// OpenAPI Spec & Interactive Swagger Docs
	mux.HandleFunc("/api/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(openapiSpec)
	})
	mux.HandleFunc("/docs", handleSwaggerDocs)
	mux.HandleFunc("/swagger", handleSwaggerDocs)

	// Background Live Watcher (dynamically reacts to interval changes)
	go func() {
		interval := cfg.WatchInterval
		var ticker *time.Ticker
		var tickerC <-chan time.Time

		resetTicker := func(sec int) {
			if ticker != nil {
				ticker.Stop()
				ticker = nil
				tickerC = nil
			}
			if sec > 0 {
				ticker = time.NewTicker(time.Duration(sec) * time.Second)
				tickerC = ticker.C
			}
		}

		resetTicker(interval)

		for {
			select {
			case newSec := <-api.watchIntervalCh:
				interval = newSec
				resetTicker(interval)
				if cfg.Verbose {
					log.Printf("[watcher] background rescan interval updated to %ds", interval)
				}
			case <-tickerC:
				if err := index.ScanAll(); err != nil && cfg.Verbose {
					log.Printf("[watcher] background rescan error: %v", err)
				}
			}
		}
	}()

	// Static files (embedded frontend with no-cache headers to avoid browser caching stale JS)
	staticFS, err := fs.Sub(staticFiles, "web/dist")
	if err != nil {
		return fmt.Errorf("failed to create sub-filesystem: %w", err)
	}
	fileServer := http.FileServer(http.FS(staticFS))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		fileServer.ServeHTTP(w, r)
	}))

	addr := fmt.Sprintf("127.0.0.1:%d", cfg.Port)
	log.Printf("🚀 Session Explorer serving at http://%s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return server.ListenAndServe()
}

// handleSwaggerDocs serves a self-contained, liquid-glass interactive Swagger / OpenAPI documentation UI.
// Implements: R-032 (Embedded OpenAPI 3.0 & Swagger UI)
func handleSwaggerDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Session Explorer — OpenAPI & Swagger Documentation</title>
  <style>
    :root {
      --bg: #070a13;
      --card-bg: rgba(15, 23, 42, 0.75);
      --card-border: rgba(255, 255, 255, 0.08);
      --cyan: #00f2fe;
      --emerald: #10b981;
      --violet: #a855f7;
      --amber: #f59e0b;
      --text: #f8fafc;
      --text-muted: #94a3b8;
      --mono: 'JetBrains Mono', monospace;
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      background: var(--bg);
      color: var(--text);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      padding: 2rem 1.5rem;
      min-height: 100vh;
    }
    .container { max-width: 1000px; margin: 0 auto; }
    .header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 2rem;
      padding-bottom: 1.5rem;
      border-bottom: 1px solid var(--card-border);
      flex-wrap: wrap;
      gap: 1rem;
    }
    .header h1 { font-size: 1.8rem; font-weight: 700; color: var(--cyan); }
    .header .actions { display: flex; gap: 0.75rem; }
    .btn {
      background: rgba(255,255,255,0.06);
      border: 1px solid var(--card-border);
      color: var(--text);
      padding: 0.5rem 1rem;
      border-radius: 8px;
      font-size: 0.85rem;
      text-decoration: none;
      cursor: pointer;
      display: inline-flex;
      align-items: center;
      gap: 0.4rem;
      transition: all 0.2s;
    }
    .btn:hover { background: rgba(0, 242, 254, 0.15); border-color: var(--cyan); color: #fff; }
    .endpoint-card {
      background: var(--card-bg);
      border: 1px solid var(--card-border);
      border-radius: 12px;
      margin-bottom: 1.5rem;
      overflow: hidden;
      backdrop-filter: blur(16px);
      transition: border-color 0.2s;
    }
    .endpoint-card:hover { border-color: rgba(0, 242, 254, 0.3); }
    .ep-header {
      padding: 1rem 1.25rem;
      display: flex;
      align-items: center;
      gap: 1rem;
      cursor: pointer;
      user-select: none;
    }
    .verb {
      font-size: 0.75rem;
      font-weight: 800;
      padding: 0.25rem 0.65rem;
      border-radius: 6px;
      letter-spacing: 0.5px;
      font-family: var(--mono);
    }
    .verb.get { background: rgba(16, 185, 129, 0.15); color: var(--emerald); border: 1px solid rgba(16, 185, 129, 0.3); }
    .verb.post { background: rgba(168, 85, 247, 0.15); color: var(--violet); border: 1px solid rgba(168, 85, 247, 0.3); }
    .ep-path { font-family: var(--mono); font-size: 0.95rem; font-weight: 600; color: #fff; }
    .ep-summary { color: var(--text-muted); font-size: 0.85rem; margin-left: auto; }
    .ep-body { padding: 1.25rem; border-top: 1px solid var(--card-border); background: rgba(0,0,0,0.25); }
    .ep-desc { font-size: 0.9rem; line-height: 1.5; color: var(--text-muted); margin-bottom: 1rem; }
    .params-table { width: 100%; border-collapse: collapse; margin-bottom: 1rem; font-size: 0.85rem; }
    .params-table th { text-align: left; padding: 0.5rem; color: var(--text-muted); border-bottom: 1px solid var(--card-border); }
    .params-table td { padding: 0.5rem; border-bottom: 1px solid rgba(255,255,255,0.04); }
    .input-field {
      width: 100%;
      background: rgba(15, 23, 42, 0.9);
      border: 1px solid var(--card-border);
      border-radius: 6px;
      color: #fff;
      padding: 0.4rem 0.65rem;
      font-family: var(--mono);
      font-size: 0.85rem;
    }
    .input-field:focus { outline: none; border-color: var(--cyan); }
    .try-btn {
      background: linear-gradient(135deg, rgba(0, 242, 254, 0.2), rgba(168, 85, 247, 0.2));
      border: 1px solid var(--cyan);
      color: #fff;
      padding: 0.5rem 1.2rem;
      border-radius: 8px;
      font-size: 0.85rem;
      font-weight: 600;
      cursor: pointer;
      margin-top: 0.5rem;
      transition: all 0.2s;
    }
    .try-btn:hover { background: var(--cyan); color: #000; }
    .response-box {
      margin-top: 1rem;
      background: #02040a;
      border: 1px solid var(--card-border);
      border-radius: 8px;
      padding: 1rem;
      font-family: var(--mono);
      font-size: 0.82rem;
      max-height: 350px;
      overflow: auto;
      white-space: pre-wrap;
      word-break: break-all;
      display: none;
    }
  </style>
</head>
<body>
  <div class="container">
    <header class="header">
      <div>
        <h1>🔍 Session Explorer API Explorer</h1>
        <div style="font-size: 0.85rem; color: var(--text-muted); margin-top: 0.25rem;">
          OpenAPI 3.0.3 Specification & Interactive Testing Console
        </div>
      </div>
      <div class="actions">
        <a href="/" class="btn">← Back to Dashboard</a>
        <a href="/api/openapi.json" target="_blank" class="btn">📄 Raw openapi.json</a>
      </div>
    </header>

    <!-- GET /api/sessions -->
    <div class="endpoint-card">
      <div class="ep-header" onclick="toggleDetails(this)">
        <span class="verb get">GET</span>
        <span class="ep-path">/api/sessions</span>
        <span class="ep-summary">List conversation sessions with filters</span>
      </div>
      <div class="ep-body">
        <div class="ep-desc">Retrieve indexed conversation sessions, sorted and filtered by workspace, date range, and sort order.</div>
        <table class="params-table">
          <tr><th>Parameter</th><th>Type</th><th>Value</th></tr>
          <tr><td>workspace</td><td>query (string)</td><td><input type="text" id="p_sessions_workspace" class="input-field" placeholder="e.g. tool-scripts"></td></tr>
          <tr><td>sort_by</td><td>query (string)</td><td><input type="text" id="p_sessions_sort_by" class="input-field" placeholder="date, steps, messages, size"></td></tr>
          <tr><td>from</td><td>query (string)</td><td><input type="text" id="p_sessions_from" class="input-field" placeholder="YYYY-MM-DD"></td></tr>
          <tr><td>to</td><td>query (string)</td><td><input type="text" id="p_sessions_to" class="input-field" placeholder="YYYY-MM-DD"></td></tr>
        </table>
        <button class="try-btn" onclick="executeGetSessions()">Try It Out</button>
        <div id="resp_sessions" class="response-box"></div>
      </div>
    </div>

    <!-- GET /api/wip -->
    <div class="endpoint-card">
      <div class="ep-header" onclick="toggleDetails(this)">
        <span class="verb get">GET</span>
        <span class="ep-path">/api/wip</span>
        <span class="ep-summary">Work-in-Progress (WIP) recovery & multi-session check</span>
      </div>
      <div class="ep-body">
        <div class="ep-desc">Inspect recent session state. If multiple sessions match, returns candidates and halts according to the Multi-Session Disambiguation Invariant.</div>
        <table class="params-table">
          <tr><th>Parameter</th><th>Type</th><th>Value</th></tr>
          <tr><td>workspace</td><td>query (string)</td><td><input type="text" id="p_wip_workspace" class="input-field" placeholder="e.g. tool-scripts" value="tool-scripts"></td></tr>
          <tr><td>session_id</td><td>query (string)</td><td><input type="text" id="p_wip_session_id" class="input-field" placeholder="Optional exact session UUID"></td></tr>
          <tr><td>limit</td><td>query (integer)</td><td><input type="text" id="p_wip_limit" class="input-field" placeholder="5" value="5"></td></tr>
          <tr><td>force_latest</td><td>query (boolean)</td><td><input type="text" id="p_wip_force" class="input-field" placeholder="true or false" value="false"></td></tr>
        </table>
        <button class="try-btn" onclick="executeGetWIP()">Try It Out</button>
        <div id="resp_wip" class="response-box"></div>
      </div>
    </div>

    <!-- GET /api/sessions/{id} -->
    <div class="endpoint-card">
      <div class="ep-header" onclick="toggleDetails(this)">
        <span class="verb get">GET</span>
        <span class="ep-path">/api/sessions/{id}</span>
        <span class="ep-summary">Get session conversation timeline</span>
      </div>
      <div class="ep-body">
        <div class="ep-desc">Retrieve the complete message trajectory, tool calls, and error traces for a session UUID.</div>
        <table class="params-table">
          <tr><th>Parameter</th><th>Type</th><th>Value</th></tr>
          <tr><td>id</td><td>path (string, required)</td><td><input type="text" id="p_detail_id" class="input-field" placeholder="Session UUID"></td></tr>
        </table>
        <button class="try-btn" onclick="executeGetDetail()">Try It Out</button>
        <div id="resp_detail" class="response-box"></div>
      </div>
    </div>

    <!-- GET /api/search -->
    <div class="endpoint-card">
      <div class="ep-header" onclick="toggleDetails(this)">
        <span class="verb get">GET</span>
        <span class="ep-path">/api/search</span>
        <span class="ep-summary">Full-text search prompt transcripts</span>
      </div>
      <div class="ep-body">
        <div class="ep-desc">Full-text keyword search across all prompts and transcript entries.</div>
        <table class="params-table">
          <tr><th>Parameter</th><th>Type</th><th>Value</th></tr>
          <tr><td>q</td><td>query (string, required)</td><td><input type="text" id="p_search_q" class="input-field" placeholder="search query..."></td></tr>
          <tr><td>workspace</td><td>query (string)</td><td><input type="text" id="p_search_ws" class="input-field" placeholder="Optional workspace filter"></td></tr>
        </table>
        <button class="try-btn" onclick="executeSearch()">Try It Out</button>
        <div id="resp_search" class="response-box"></div>
      </div>
    </div>

    <!-- GET /api/stats -->
    <div class="endpoint-card">
      <div class="ep-header" onclick="toggleDetails(this)">
        <span class="verb get">GET</span>
        <span class="ep-path">/api/stats</span>
        <span class="ep-summary">Aggregate session and tool statistics</span>
      </div>
      <div class="ep-body">
        <div class="ep-desc">Returns aggregate counts, prompt frequencies, top tools, and project clusters.</div>
        <button class="try-btn" onclick="executeGetStats()">Try It Out</button>
        <div id="resp_stats" class="response-box"></div>
      </div>
    </div>

    <!-- POST /api/skill/install -->
    <div class="endpoint-card">
      <div class="ep-header" onclick="toggleDetails(this)">
        <span class="verb post">POST</span>
        <span class="ep-path">/api/skill/install</span>
        <span class="ep-summary">Install Session WIP skill & hooks</span>
      </div>
      <div class="ep-body">
        <div class="ep-desc">Deploys the session-wip agent skill and scripts to the active workspace (.agents) or global user directory.</div>
        <table class="params-table">
          <tr><th>Parameter</th><th>Type</th><th>Value</th></tr>
          <tr><td>target</td><td>body (string)</td><td>
            <select id="p_install_target" class="input-field">
              <option value="workspace">workspace (.agents)</option>
              <option value="global">global (~/.gemini/config)</option>
            </select>
          </td></tr>
        </table>
        <button class="try-btn" onclick="executeInstallSkill()">Install Skill</button>
        <div id="resp_install" class="response-box"></div>
      </div>
    </div>

    <!-- POST /api/refresh -->
    <div class="endpoint-card">
      <div class="ep-header" onclick="toggleDetails(this)">
        <span class="verb post">POST</span>
        <span class="ep-path">/api/refresh</span>
        <span class="ep-summary">Trigger immediate brain re-scan</span>
      </div>
      <div class="ep-body">
        <div class="ep-desc">Forces an immediate scan of the brain directory without waiting for background watcher interval.</div>
        <button class="try-btn" onclick="executeRefresh()">Trigger Refresh</button>
        <div id="resp_refresh" class="response-box"></div>
      </div>
    </div>

  </div>

  <script>
    function toggleDetails(header) {
      const body = header.nextElementSibling;
      body.style.display = body.style.display === 'none' ? 'block' : 'none';
    }

    async function runFetch(url, options, respElId) {
      const respEl = document.getElementById(respElId);
      respEl.style.display = 'block';
      respEl.textContent = '⏳ Executing request...';
      const start = performance.now();
      try {
        const res = await fetch(url, options);
        const data = await res.json();
        const duration = Math.round(performance.now() - start);
        respEl.textContent = '// Status: ' + res.status + ' ' + res.statusText + ' (' + duration + 'ms)\n\n' + JSON.stringify(data, null, 2);
      } catch (err) {
        respEl.textContent = '❌ Request failed: ' + err.message;
      }
    }

    function executeGetSessions() {
      const params = new URLSearchParams();
      const ws = document.getElementById('p_sessions_workspace').value.trim();
      const sort = document.getElementById('p_sessions_sort_by').value.trim();
      const from = document.getElementById('p_sessions_from').value.trim();
      const to = document.getElementById('p_sessions_to').value.trim();
      if (ws) params.set('workspace', ws);
      if (sort) params.set('sort_by', sort);
      if (from) params.set('from', from);
      if (to) params.set('to', to);
      runFetch('/api/sessions?' + params.toString(), { method: 'GET' }, 'resp_sessions');
    }

    function executeGetWIP() {
      const params = new URLSearchParams();
      const ws = document.getElementById('p_wip_workspace').value.trim();
      const sid = document.getElementById('p_wip_session_id').value.trim();
      const lim = document.getElementById('p_wip_limit').value.trim();
      const force = document.getElementById('p_wip_force').value.trim();
      if (ws) params.set('workspace', ws);
      if (sid) params.set('session_id', sid);
      if (lim) params.set('limit', lim);
      if (force) params.set('force_latest', force);
      runFetch('/api/wip?' + params.toString(), { method: 'GET' }, 'resp_wip');
    }

    function executeGetDetail() {
      const id = document.getElementById('p_detail_id').value.trim();
      if (!id) { alert('Please enter a session UUID'); return; }
      runFetch('/api/sessions/' + encodeURIComponent(id), { method: 'GET' }, 'resp_detail');
    }

    function executeSearch() {
      const q = document.getElementById('p_search_q').value.trim();
      const ws = document.getElementById('p_search_ws').value.trim();
      if (!q) { alert('Please enter search query'); return; }
      const params = new URLSearchParams({ q });
      if (ws) params.set('workspace', ws);
      runFetch('/api/search?' + params.toString(), { method: 'GET' }, 'resp_search');
    }

    function executeGetStats() {
      runFetch('/api/stats', { method: 'GET' }, 'resp_stats');
    }

    function executeInstallSkill() {
      const target = document.getElementById('p_install_target').value;
      runFetch('/api/skill/install', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ target })
      }, 'resp_install');
    }

    function executeRefresh() {
      runFetch('/api/refresh', { method: 'POST' }, 'resp_refresh');
    }
  </script>
</body>
</html>`

	w.Write([]byte(html))
}

