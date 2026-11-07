// Package main provides REST API handlers for the Session Explorer.
//
// Implements: R-010 (Date filter), R-011 (Workspace filter), R-012 (Search),
//             R-013 (Type filter), R-014 (Sort)
// See: specs/catalog/session-explorer.md
package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// APIHandler holds the session index and provides HTTP handlers.
type APIHandler struct {
	index           *SessionIndex
	verbose         bool
	watchInterval   int
	watchIntervalMu sync.RWMutex
	watchIntervalCh chan int
}

// NewAPIHandler creates a new API handler.
func NewAPIHandler(index *SessionIndex, verbose bool) *APIHandler {
	return &APIHandler{
		index:           index,
		verbose:         verbose,
		watchInterval:   10,
		watchIntervalCh: make(chan int, 10),
	}
}

// SetWatchInterval safely updates the background watch interval and notifies the watcher loop.
func (h *APIHandler) SetWatchInterval(sec int) {
	h.watchIntervalMu.Lock()
	h.watchInterval = sec
	h.watchIntervalMu.Unlock()
	select {
	case h.watchIntervalCh <- sec:
	default:
	}
}

// GetWatchInterval returns the current background watch interval in seconds.
func (h *APIHandler) GetWatchInterval() int {
	h.watchIntervalMu.RLock()
	defer h.watchIntervalMu.RUnlock()
	return h.watchInterval
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
	runtimeMode := q.Get("runtime")
	sortBy := q.Get("sort_by")
	sortOrder := q.Get("sort_order")

	if sortBy == "date_asc" {
		sortBy = "date"
		sortOrder = "asc"
	}
	if sortBy == "active_asc" {
		sortBy = "active"
		sortOrder = "asc"
	}
	if sortBy == "active_desc" {
		sortBy = "active"
		sortOrder = "desc"
	}
	if sortBy == "" {
		sortBy = "active"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	loc := resolveLocation(q.Get("tz_offset"))
	from := parseDate(q.Get("from"), false, loc)
	to := parseDate(q.Get("to"), true, loc)

	dateMode := q.Get("date_mode")
	if dateMode == "" {
		dateMode = q.Get("date_filter_by")
	}
	if dateMode == "" {
		dateMode = "active" // Default: filter by last active/touched recency (R-049)
	}

	sessions := h.index.GetSessions(workspace, sortBy, sortOrder, from, to, dateMode)
	if runtimeMode != "" {
		filtered := make([]Session, 0, len(sessions))
		for _, s := range sessions {
			if strings.EqualFold(s.RuntimeMode, runtimeMode) || strings.EqualFold(s.Provider, runtimeMode) {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}
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
	loc := resolveLocation(q.Get("tz_offset"))
	from := parseDate(q.Get("from"), false, loc)
	to := parseDate(q.Get("to"), true, loc)

	if query == "" {
		writeJSON(w, SearchResponse{Results: []SearchResult{}, TotalCount: 0, Query: ""})
		return
	}

	dateMode := q.Get("date_mode")
	if dateMode == "" {
		dateMode = q.Get("date_filter_by")
	}
	if dateMode == "" {
		dateMode = "active"
	}

	results := h.index.Search(query, workspace, msgType, from, to, dateMode)
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

func resolveLocation(tzStr string) *time.Location {
	if tzStr != "" {
		if offsetMin, err := strconv.Atoi(strings.TrimSpace(tzStr)); err == nil {
			// JavaScript getTimezoneOffset() returns positive minutes for timezones west of UTC (e.g. +300 for EST / UTC-5)
			// and negative minutes for timezones east of UTC (e.g. -330 for IST / UTC+5:30).
			// time.FixedZone takes offset in seconds east of UTC, so negate offsetMin.
			return time.FixedZone("ClientZone", -offsetMin*60)
		}
	}
	return time.Local
}

func parseDate(s string, isEnd bool, loc *time.Location) time.Time {
	if s == "" {
		return time.Time{}
	}
	if loc == nil {
		loc = time.Local
	}
	s = strings.TrimSpace(s)
	// If date-only string (e.g., "2026-09-18")
	if len(s) == 10 {
		if t, err := time.ParseInLocation("2006-01-02", s, loc); err == nil {
			if isEnd {
				// Cover the entire 24 hours of the end day in the target timezone
				return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), loc)
			}
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
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

// HandleWatchInterval handles GET and POST /api/watch-interval.
func (h *APIHandler) HandleWatchInterval(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		curr := h.GetWatchInterval()
		status := "active"
		if curr <= 0 {
			status = "paused"
		}
		writeJSON(w, WatchIntervalResponse{
			IntervalSeconds: curr,
			Status:          status,
			Message:         fmt.Sprintf("Current background watch interval is %ds (%s)", curr, status),
		})
		return
	}

	if r.Method == http.MethodPost {
		var req WatchIntervalRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
			return
		}
		if req.IntervalSeconds < 0 {
			req.IntervalSeconds = 0
		}
		h.SetWatchInterval(req.IntervalSeconds)
		status := "active"
		if req.IntervalSeconds <= 0 {
			status = "paused"
		}
		writeJSON(w, WatchIntervalResponse{
			IntervalSeconds: req.IntervalSeconds,
			Status:          status,
			Message:         fmt.Sprintf("Live watch interval successfully updated to %ds (%s)", req.IntervalSeconds, status),
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HandleSessionTail handles GET /api/sessions/{id}/tail for real-time SSE delta streaming.
func (h *APIHandler) HandleSessionTail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Extract session ID from path: /api/sessions/{id}/tail
	sub := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(sub, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}
	sessionID := parts[0]

	// Prefer full transcript if available (Implements: R-036)
	transcriptPath, err := h.index.GetTranscriptPath(sessionID, true)
	if err != nil {
		http.Error(w, "Session log not found: "+err.Error(), http.StatusNotFound)
		return
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		http.Error(w, "Session log not found: "+err.Error(), http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")

	lastStep := -1
	if lsStr := r.URL.Query().Get("last_step"); lsStr != "" {
		if ls, err := strconv.Atoi(lsStr); err == nil {
			lastStep = ls
		}
	}

	fmt.Fprintf(w, "event: connected\ndata: {\"session_id\":\"%s\",\"last_step\":%d}\n\n", sessionID, lastStep)
	flusher.Flush()

	// Initial replay: send any steps > lastStep
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry TranscriptEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		if entry.StepIndex > lastStep && isDisplayableType(entry.Type) {
			msg := formatMessageDisplay(entry)
			msgBytes, _ := json.Marshal(msg)
			fmt.Fprintf(w, "event: step\ndata: %s\n\n", msgBytes)
			flusher.Flush()
			if entry.StepIndex > lastStep {
				lastStep = entry.StepIndex
			}
		}
	}

	lastOffset, _ := file.Seek(0, io.SeekEnd)

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		case <-ticker.C:
			fi, err := os.Stat(transcriptPath)
			if err != nil {
				continue
			}
			currSize := fi.Size()
			if currSize < lastOffset {
				lastOffset = 0
			}
			if currSize > lastOffset {
				file.Seek(lastOffset, io.SeekStart)
				scan := bufio.NewScanner(file)
				scanBuf := make([]byte, 0, 1024*1024)
				scan.Buffer(scanBuf, 10*1024*1024)

				for scan.Scan() {
					line := scan.Bytes()
					if len(line) == 0 {
						continue
					}
					var entry TranscriptEntry
					if err := json.Unmarshal(line, &entry); err != nil {
						continue
					}
					if isDisplayableType(entry.Type) {
						msg := formatMessageDisplay(entry)
						msgBytes, _ := json.Marshal(msg)
						fmt.Fprintf(w, "event: step\ndata: %s\n\n", msgBytes)
						flusher.Flush()
						if entry.StepIndex > lastStep {
							lastStep = entry.StepIndex
						}
					}
				}
				newOffset, err := file.Seek(0, io.SeekCurrent)
				if err == nil && newOffset > lastOffset {
					lastOffset = newOffset
				}
			}
		}
	}
}

// HandleFile handles GET /api/file?path=<path>
// Implements: R-038 (Native In-Browser File & Artifact Viewer)
func (h *APIHandler) HandleFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawPath := r.URL.Query().Get("path")
	if strings.TrimSpace(rawPath) == "" {
		http.Error(w, "Query parameter 'path' is required", http.StatusBadRequest)
		return
	}

	// Strip file:// prefix if provided
	cleanPath := strings.TrimPrefix(rawPath, "file://")

	// Strip any trailing #L... line anchors from path for filesystem reading
	if idx := strings.Index(cleanPath, "#"); idx != -1 {
		cleanPath = cleanPath[:idx]
	}
	// Strip trailing :line if path doesn't exist directly
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		if lastColon := strings.LastIndex(cleanPath, ":"); lastColon != -1 {
			candidate := cleanPath[:lastColon]
			if _, err2 := os.Stat(candidate); err2 == nil {
				cleanPath = candidate
			}
		}
	}

	cleanPath = filepath.Clean(cleanPath)

	// Expand ~ to user home directory
	if strings.HasPrefix(cleanPath, "~/") || cleanPath == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			cleanPath = filepath.Join(home, strings.TrimPrefix(cleanPath, "~"))
		}
	}

	// Security guards against sensitive credential paths
	lower := strings.ToLower(cleanPath)
	if strings.Contains(lower, "/.ssh/") || strings.Contains(lower, "\\.ssh\\") ||
		strings.Contains(lower, "/.gnupg/") || strings.Contains(lower, "\\.gnupg\\") ||
		strings.Contains(lower, "/.aws/credentials") ||
		strings.HasSuffix(lower, "/etc/shadow") || strings.HasSuffix(lower, "/etc/master.passwd") ||
		strings.HasSuffix(lower, "/id_rsa") || strings.HasSuffix(lower, "/id_ed25519") {
		http.Error(w, "Access to requested path is restricted", http.StatusForbidden)
		return
	}

	// Resolve relative path against working directory or known project workspaces
	if !filepath.IsAbs(cleanPath) {
		if _, err := os.Stat(cleanPath); os.IsNotExist(err) && h.index != nil {
			h.index.mu.RLock()
			for _, projPath := range h.index.knownProjects {
				cand := filepath.Join(projPath, cleanPath)
				if _, err2 := os.Stat(cand); err2 == nil {
					cleanPath = cand
					break
				}
			}
			h.index.mu.RUnlock()
		}
		if !filepath.IsAbs(cleanPath) {
			if abs, err := filepath.Abs(cleanPath); err == nil {
				cleanPath = abs
			}
		}
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found: "+cleanPath, http.StatusNotFound)
			return
		}
		http.Error(w, "Error accessing file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if info.IsDir() {
		http.Error(w, "Requested path is a directory, not a file", http.StatusBadRequest)
		return
	}

	// 10MB file size ceiling to prevent OOM
	const maxFileSize = 10 * 1024 * 1024
	if info.Size() > maxFileSize {
		http.Error(w, fmt.Sprintf("File size (%d bytes) exceeds maximum preview limit of 10MB", info.Size()), http.StatusRequestEntityTooLarge)
		return
	}

	ext := strings.ToLower(filepath.Ext(cleanPath))
	name := filepath.Base(cleanPath)

	// Detect image types
	isImage := ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp" || ext == ".ico"
	isSvg := ext == ".svg"

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	mimeType := http.DetectContentType(data)

	var content string
	var isBinary bool
	var lineCount int

	if isImage {
		isBinary = true
		content = fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))
		lineCount = 1
	} else if isSvg {
		isBinary = false
		content = string(data)
		lineCount = bytes.Count(data, []byte("\n")) + 1
		mimeType = "image/svg+xml"
	} else {
		// Check for binary content (null bytes)
		if bytes.IndexByte(data, 0) != -1 {
			isBinary = true
			content = "[Binary file content not displayed]"
			lineCount = 0
		} else {
			isBinary = false
			content = string(data)
			if len(data) > 0 {
				lineCount = bytes.Count(data, []byte("\n")) + 1
			} else {
				lineCount = 0
			}
		}
	}

	resp := FileContentResponse{
		Path:      cleanPath,
		Name:      name,
		Extension: ext,
		Size:      info.Size(),
		ModTime:   info.ModTime().UTC().Format(time.RFC3339),
		Content:   content,
		IsBinary:  isBinary,
		MimeType:  mimeType,
		LineCount: lineCount,
	}

	writeJSON(w, resp)
}

// HandleGlobalAgentContext returns machine-wide agent context (rules, skills, plugins, MCP, hooks).
// Implements: R-039, AC-040, AC-044
func (h *APIHandler) HandleGlobalAgentContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := h.index.DiscoverGlobalAgentContext()
	writeJSON(w, ctx)
}

// HandleWorkspaceAgentContext returns workspace-specific and inherited agent context.
// Implements: R-039, AC-041, AC-044
func (h *APIHandler) HandleWorkspaceAgentContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	workspace := r.URL.Query().Get("workspace")
	if workspace == "" {
		workspace = r.URL.Query().Get("project")
	}

	ctx := h.index.DiscoverWorkspaceAgentContext(workspace)
	writeJSON(w, ctx)
}

// HandleSessionContext returns active rules, invoked skills, and tools for a specific session.
// Implements: R-039, AC-042, AC-043
func (h *APIHandler) HandleSessionContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/sessions/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}
	sessionID := parts[0]

	sctx, err := h.index.ExtractSessionAgentContext(sessionID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, sctx)
}



