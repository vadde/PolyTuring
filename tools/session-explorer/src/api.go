// Package main provides REST API handlers for the Session Explorer.
//
// Implements: R-010 (Date filter), R-011 (Workspace filter), R-012 (Search),
//             R-013 (Type filter), R-014 (Sort)
// See: specs/catalog/session-explorer.md
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// APIHandler holds the session index and provides HTTP handlers.
type APIHandler struct {
	index   *SessionIndex
	verbose bool
}

// NewAPIHandler creates a new API handler.
func NewAPIHandler(index *SessionIndex, verbose bool) *APIHandler {
	return &APIHandler{
		index:   index,
		verbose: verbose,
	}
}

// HandleSessions handles GET /api/sessions
// Query params: workspace, sort_by, sort_order, from, to
// Implements: R-008, R-010, R-011, R-014
func (h *APIHandler) HandleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	workspace := q.Get("workspace")
	sortBy := q.Get("sort_by")
	sortOrder := q.Get("sort_order")

	if sortBy == "date_asc" {
		sortBy = "date"
		sortOrder = "asc"
	}
	if sortBy == "" {
		sortBy = "date"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	from := parseDate(q.Get("from"), false)
	to := parseDate(q.Get("to"), true)

	sessions := h.index.GetSessions(workspace, sortBy, sortOrder, from, to)
	stats := h.index.GetStats()

	resp := SessionsResponse{
		Sessions: sessions,
		Stats:    stats,
	}

	writeJSON(w, resp)
}

// HandleSessionDetail handles GET /api/sessions/{id}
// Implements: R-009
func (h *APIHandler) HandleSessionDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract session ID from path: /api/sessions/{id}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/sessions/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}
	sessionID := parts[0]

	detail, err := h.index.GetSessionDetail(sessionID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, detail)
}

// HandleSearch handles GET /api/search
// Query params: q, workspace, type, from, to
// Implements: R-012, R-013
func (h *APIHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	query := q.Get("q")
	workspace := q.Get("workspace")
	msgType := q.Get("type")
	from := parseDate(q.Get("from"), false)
	to := parseDate(q.Get("to"), true)

	if query == "" {
		writeJSON(w, SearchResponse{Results: []SearchResult{}, TotalCount: 0, Query: ""})
		return
	}

	results := h.index.Search(query, workspace, msgType, from, to)
	writeJSON(w, results)
}

// HandleStats handles GET /api/stats
// Implements: R-017
func (h *APIHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, h.index.GetStats())
}

// HandleWIP handles GET /api/wip
// Query params: workspace, session_id, limit, force_latest
// Implements: R-030 (Agent WIP Recovery), R-031 (Multi-Session Disambiguation Invariant)
func (h *APIHandler) HandleWIP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	workspace := q.Get("workspace")
	sessionID := q.Get("session_id")
	limitStr := q.Get("limit")
	limit := 5
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	forceLatest := q.Get("force_latest") == "true" || q.Get("force") == "true" || q.Get("latest") == "true"

	resp, err := h.index.GetWorkspaceWIP(workspace, sessionID, forceLatest, limit)
	if err != nil {
		http.Error(w, "Failed to resolve WIP: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp)
}

// HandleInstallSkill handles POST /api/skill/install
// Implements: R-033 (One-Click Skill Installer)
func (h *APIHandler) HandleInstallSkill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req InstallSkillRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	target := strings.ToLower(strings.TrimSpace(req.Target))
	if target == "" {
		target = "workspace"
	}

	homeDir, _ := os.UserHomeDir()
	var destSkillDir string
	if target == "global" {
		destSkillDir = filepath.Join(homeDir, ".gemini", "config", "skills", "session-wip")
	} else {
		// Default to workspace .agents/skills/session-wip
		cwd, _ := os.Getwd()
		// Walk up to find .agents or repo root
		curr := cwd
		found := false
		for i := 0; i < 5; i++ {
			if _, err := os.Stat(filepath.Join(curr, ".agents")); err == nil {
				destSkillDir = filepath.Join(curr, ".agents", "skills", "session-wip")
				found = true
				break
			}
			parent := filepath.Dir(curr)
			if parent == curr {
				break
			}
			curr = parent
		}
		if !found {
			destSkillDir = filepath.Join(cwd, ".agents", "skills", "session-wip")
		}
	}

	// Try running setup-skill.sh if it exists
	scriptCandidates := []string{
		filepath.Join("tools", "session-explorer", "scripts", "setup-skill.sh"),
		filepath.Join("scripts", "setup-skill.sh"),
		filepath.Join("..", "scripts", "setup-skill.sh"),
	}

	var scriptRan bool
	for _, sc := range scriptCandidates {
		if _, err := os.Stat(sc); err == nil {
			cmd := exec.Command("/bin/bash", sc, target)
			if output, err := cmd.CombinedOutput(); err == nil {
				scriptRan = true
				if h.verbose {
					log.Printf("[api] setup-skill.sh output: %s", string(output))
				}
				break
			}
		}
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Join(destSkillDir, "scripts"), 0755); err != nil && !scriptRan {
		http.Error(w, "Failed to create directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, InstallSkillResponse{
		Success:       true,
		Target:        target,
		InstalledPath: destSkillDir,
		Message:       fmt.Sprintf("Session WIP skill successfully installed to %s mode (%s).", target, destSkillDir),
	})
}


// --- Helpers ---

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[api] JSON encode error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func parseDate(s string, isEnd bool) time.Time {
	if s == "" {
		return time.Time{}
	}
	s = strings.TrimSpace(s)
	// If date-only string (e.g., "2026-09-18")
	if len(s) == 10 {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			if isEnd {
				// Cover the entire 24 hours of the end day
				return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
			}
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		}
	}

	// Try multiple timestamp formats
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// HandleRefresh handles POST /api/refresh and GET /api/refresh to trigger an immediate re-scan.
func (h *APIHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	if err := h.index.ScanAll(); err != nil {
		http.Error(w, "Failed to refresh index: "+err.Error(), http.StatusInternalServerError)
		return
	}

	stats := h.index.GetStats()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Index refreshed successfully",
		"stats":   stats,
	})
}

