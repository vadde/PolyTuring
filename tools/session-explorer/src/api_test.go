// Tests for api.go
// Implements test coverage for R-006, R-008, R-009, R-011, R-012, R-013, R-017
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupTestIndex(t *testing.T) *SessionIndex {
	tmpDir := t.TempDir()

	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-18T10:00:00Z","content":"<USER_REQUEST>Test prompt for API</USER_REQUEST>\n<ADDITIONAL_METADATA>\nActive Document: /Users/test/projects/demo/main.go\n</ADDITIONAL_METADATA>","tool_calls":[]}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-09-18T10:01:00Z","content":"Response content from model","tool_calls":[{"name":"view_file","args":{"AbsolutePath":"/Users/test/projects/demo/main.go"}}]}`

	createMockSession(t, tmpDir, "session-api-1", content)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("failed to scan test index: %v", err)
	}
	return idx
}

func TestAPI_HandleSessions(t *testing.T) {
	idx := setupTestIndex(t)
	handler := NewAPIHandler(idx, false)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	rec := httptest.NewRecorder()

	handler.HandleSessions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp SessionsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if len(resp.Sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(resp.Sessions))
	}
	if resp.Sessions[0].ID != "session-api-1" {
		t.Errorf("expected session-api-1, got %s", resp.Sessions[0].ID)
	}
}

func TestAPI_HandleSessionDetail(t *testing.T) {
	idx := setupTestIndex(t)
	handler := NewAPIHandler(idx, false)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/session-api-1", nil)
	rec := httptest.NewRecorder()

	handler.HandleSessionDetail(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp SessionDetailResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if resp.Session.ID != "session-api-1" {
		t.Errorf("expected session-api-1, got %s", resp.Session.ID)
	}
	if len(resp.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(resp.Messages))
	}
}

func TestAPI_HandleSearch(t *testing.T) {
	idx := setupTestIndex(t)
	handler := NewAPIHandler(idx, false)

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=prompt", nil)
	rec := httptest.NewRecorder()

	handler.HandleSearch(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp SearchResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode search response: %v", err)
	}

	if resp.TotalCount != 1 {
		t.Errorf("expected 1 match, got %d", resp.TotalCount)
	}
}

func TestAPI_HandleStats(t *testing.T) {
	idx := setupTestIndex(t)
	handler := NewAPIHandler(idx, false)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rec := httptest.NewRecorder()

	handler.HandleStats(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var stats Stats
	if err := json.NewDecoder(rec.Body).Decode(&stats); err != nil {
		t.Fatalf("failed to decode stats: %v", err)
	}

	if stats.TotalSessions != 1 {
		t.Errorf("expected 1 session in stats, got %d", stats.TotalSessions)
	}
	if len(stats.Projects) == 0 {
		t.Errorf("expected stats.Projects to be populated, got empty")
	} else if stats.Projects[0].Name != "demo" {
		t.Errorf("expected project demo, got %s", stats.Projects[0].Name)
	}
}

func TestAPI_HandleSessions_DateFilterToday(t *testing.T) {
	tmpDir := t.TempDir()

	// Session created on 2026-09-18 at 14:30:00 UTC (after midnight UTC)
	contentToday := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-18T14:30:00Z","content":"<USER_REQUEST>Today prompt</USER_REQUEST>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-09-18T14:31:00Z","content":"Today response"}`
	createMockSession(t, tmpDir, "session-today", contentToday)

	// Session created on 2026-09-10
	contentPast := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-10T10:00:00Z","content":"<USER_REQUEST>Past prompt</USER_REQUEST>"}`
	createMockSession(t, tmpDir, "session-past", contentPast)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("failed to scan test index: %v", err)
	}

	handler := NewAPIHandler(idx, false)

	// Query from=2026-09-18&to=2026-09-18 (equivalent to "Today" preset)
	req := httptest.NewRequest(http.MethodGet, "/api/sessions?from=2026-09-18&to=2026-09-18", nil)
	rec := httptest.NewRecorder()
	handler.HandleSessions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp SessionsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Sessions) != 1 {
		t.Fatalf("expected 1 session for Today date filter, got %d", len(resp.Sessions))
	}
	if resp.Sessions[0].ID != "session-today" {
		t.Errorf("expected session-today, got %s", resp.Sessions[0].ID)
	}
}

func TestAPI_HandleSessions_SortDateAsc(t *testing.T) {
	tmpDir := t.TempDir()

	content1 := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-10T10:00:00Z","content":"<USER_REQUEST>First</USER_REQUEST>"}`
	createMockSession(t, tmpDir, "session-1", content1)

	content2 := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-18T10:00:00Z","content":"<USER_REQUEST>Second</USER_REQUEST>"}`
	createMockSession(t, tmpDir, "session-2", content2)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("failed to scan test index: %v", err)
	}

	handler := NewAPIHandler(idx, false)

	// Query sort_by=date_asc (oldest first)
	req := httptest.NewRequest(http.MethodGet, "/api/sessions?sort_by=date_asc", nil)
	rec := httptest.NewRecorder()
	handler.HandleSessions(rec, req)

	var resp SessionsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(resp.Sessions))
	}
	if resp.Sessions[0].ID != "session-1" {
		t.Errorf("expected oldest session-1 first with date_asc, got %s", resp.Sessions[0].ID)
	}
}

func TestAPI_HandleSessions_SortActive(t *testing.T) {
	tmpDir := t.TempDir()

	// session-old: created on Sept 1, but last message on Sept 20
	contentOld := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-01T10:00:00Z","content":"<USER_REQUEST>Old</USER_REQUEST>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-09-20T10:00:00Z","content":"Late response"}`
	createMockSession(t, tmpDir, "session-old", contentOld)

	// session-new: created on Sept 15, last message on Sept 15
	contentNew := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-15T10:00:00Z","content":"<USER_REQUEST>New</USER_REQUEST>"}`
	createMockSession(t, tmpDir, "session-new", contentNew)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("failed to scan test index: %v", err)
	}

	handler := NewAPIHandler(idx, false)

	// Query sort_by=active (most recently active first)
	req := httptest.NewRequest(http.MethodGet, "/api/sessions?sort_by=active", nil)
	rec := httptest.NewRecorder()
	handler.HandleSessions(rec, req)

	var resp SessionsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(resp.Sessions))
	}
	// session-old was active on Sept 20, whereas session-new was active on Sept 15.
	// Therefore session-old MUST be first when sorting by active recency!
	if resp.Sessions[0].ID != "session-old" {
		t.Errorf("expected session-old (active Sept 20) first with sort_by=active, got %s", resp.Sessions[0].ID)
	}
}

func TestAPI_HandleSessions_TriModalDateFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// session-old: created on Sept 1, active until Sept 20
	contentOld := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-01T10:00:00Z","content":"<USER_REQUEST>Old</USER_REQUEST>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-09-20T10:00:00Z","content":"Late response"}`
	createMockSession(t, tmpDir, "session-old", contentOld)

	// session-new: created on Sept 15, active until Sept 15
	contentNew := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-15T10:00:00Z","content":"<USER_REQUEST>New</USER_REQUEST>"}`
	createMockSession(t, tmpDir, "session-new", contentNew)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("failed to scan test index: %v", err)
	}

	handler := NewAPIHandler(idx, false)

	// 1. date_mode=active for 2026-09-20: session-old active on Sept 20 matches; session-new does not
	{
		req := httptest.NewRequest(http.MethodGet, "/api/sessions?from=2026-09-20&to=2026-09-20&date_mode=active", nil)
		rec := httptest.NewRecorder()
		handler.HandleSessions(rec, req)

		var resp SessionsResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp.Sessions) != 1 || resp.Sessions[0].ID != "session-old" {
			t.Fatalf("expected only session-old for date_mode=active on Sept 20, got %d sessions", len(resp.Sessions))
		}
	}

	// 2. date_mode=created for 2026-09-20: neither session was created on Sept 20 -> 0 sessions
	{
		req := httptest.NewRequest(http.MethodGet, "/api/sessions?from=2026-09-20&to=2026-09-20&date_mode=created", nil)
		rec := httptest.NewRecorder()
		handler.HandleSessions(rec, req)

		var resp SessionsResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp.Sessions) != 0 {
			t.Fatalf("expected 0 sessions for date_mode=created on Sept 20, got %d", len(resp.Sessions))
		}
	}

	// 3. date_mode=created for 2026-09-15: session-new was created on Sept 15 -> matches session-new
	{
		req := httptest.NewRequest(http.MethodGet, "/api/sessions?from=2026-09-15&to=2026-09-15&date_mode=created", nil)
		rec := httptest.NewRecorder()
		handler.HandleSessions(rec, req)

		var resp SessionsResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp.Sessions) != 1 || resp.Sessions[0].ID != "session-new" {
			t.Fatalf("expected session-new for date_mode=created on Sept 15, got %d sessions", len(resp.Sessions))
		}
	}

	// 4. date_mode=overlap for 2026-09-10: session-old (created Sept 1, active Sept 20) spans Sept 10
	{
		req := httptest.NewRequest(http.MethodGet, "/api/sessions?from=2026-09-10&to=2026-09-10&date_mode=overlap", nil)
		rec := httptest.NewRecorder()
		handler.HandleSessions(rec, req)

		var resp SessionsResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp.Sessions) != 1 || resp.Sessions[0].ID != "session-old" {
			t.Fatalf("expected session-old for date_mode=overlap on Sept 10, got %d sessions", len(resp.Sessions))
		}
	}
}

func TestAPI_HandleRefresh(t *testing.T) {
	tmpDir := t.TempDir()

	content1 := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-10T10:00:00Z","content":"<USER_REQUEST>Initial</USER_REQUEST>"}`
	createMockSession(t, tmpDir, "session-1", content1)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("failed to scan test index: %v", err)
	}

	handler := NewAPIHandler(idx, false)

	// Verify initial count is 1
	if len(idx.GetSessions("", "date", "desc", time.Time{}, time.Time{})) != 1 {
		t.Fatalf("expected 1 session initially")
	}

	// Add second session to disk while server is running
	content2 := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-18T10:00:00Z","content":"<USER_REQUEST>Added later</USER_REQUEST>"}`
	createMockSession(t, tmpDir, "session-2", content2)

	// Hit /api/refresh
	req := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	rec := httptest.NewRecorder()
	handler.HandleRefresh(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	// Verify index now contains both sessions
	sessions := idx.GetSessions("", "date", "desc", time.Time{}, time.Time{})
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions after refresh, got %d", len(sessions))
	}
}

func TestAPI_HandleWIP_SingleSession(t *testing.T) {
	idx := setupTestIndex(t)
	handler := NewAPIHandler(idx, false)

	req := httptest.NewRequest(http.MethodGet, "/api/wip?workspace=demo", nil)
	rec := httptest.NewRecorder()

	handler.HandleWIP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp WIPResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode WIP response: %v", err)
	}

	if resp.Status != "resolved" {
		t.Errorf("expected status 'resolved', got %q", resp.Status)
	}
	if resp.ActiveSession == nil || resp.ActiveSession.SessionID != "session-api-1" {
		t.Errorf("expected active session for session-api-1, got %v", resp.ActiveSession)
	}
}

func TestAPI_HandleWIP_MultiSession_AmbiguousAndForceLatest(t *testing.T) {
	tmpDir := t.TempDir()

	content1 := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-18T10:00:00Z","content":"<USER_REQUEST>Prompt 1</USER_REQUEST>\n<ADDITIONAL_METADATA>\nActive Document: /Users/test/projects/demo/main.go\n</ADDITIONAL_METADATA>"}`
	content2 := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-18T12:00:00Z","content":"<USER_REQUEST>Prompt 2</USER_REQUEST>\n<ADDITIONAL_METADATA>\nActive Document: /Users/test/projects/demo/main.go\n</ADDITIONAL_METADATA>"}`

	createMockSession(t, tmpDir, "session-api-1", content1)
	createMockSession(t, tmpDir, "session-api-2", content2)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("failed to scan test index: %v", err)
	}

	handler := NewAPIHandler(idx, false)

	// 1. Ambiguous multi-session halt
	req := httptest.NewRequest(http.MethodGet, "/api/wip?workspace=demo", nil)
	rec := httptest.NewRecorder()
	handler.HandleWIP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp WIPResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode WIP response: %v", err)
	}

	if resp.Status != "ambiguous_multi_session" {
		t.Fatalf("expected status 'ambiguous_multi_session', got %q", resp.Status)
	}
	if len(resp.Candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(resp.Candidates))
	}

	// 2. Resolve with latest=true
	reqLatest := httptest.NewRequest(http.MethodGet, "/api/wip?workspace=demo&latest=true", nil)
	recLatest := httptest.NewRecorder()
	handler.HandleWIP(recLatest, reqLatest)

	var respLatest WIPResponse
	if err := json.NewDecoder(recLatest.Body).Decode(&respLatest); err != nil {
		t.Fatalf("failed to decode WIP latest response: %v", err)
	}

	if respLatest.Status != "resolved" {
		t.Fatalf("expected status 'resolved' with latest=true, got %q", respLatest.Status)
	}
	if respLatest.ActiveSession.SessionID != "session-api-2" {
		t.Errorf("expected session-api-2, got %s", respLatest.ActiveSession.SessionID)
	}
}

func TestAPI_HandleInstallSkill(t *testing.T) {
	idx := setupTestIndex(t)
	handler := NewAPIHandler(idx, false)

	body := bytes.NewBufferString(`{"target":"workspace"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/skill/install", body)
	rec := httptest.NewRecorder()

	handler.HandleInstallSkill(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp InstallSkillResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
	if !strings.Contains(resp.InstalledPath, "session-wip") {
		t.Errorf("expected installed path to contain session-wip, got %s", resp.InstalledPath)
	}
}

func TestAPI_OpenAPI_And_SwaggerDocs(t *testing.T) {
	// 1. Verify openapiSpec embeds valid OpenAPI JSON
	var spec map[string]interface{}
	if err := json.Unmarshal(openapiSpec, &spec); err != nil {
		t.Fatalf("openapi.json is not valid JSON: %v", err)
	}

	if _, ok := spec["openapi"]; !ok {
		t.Errorf("missing 'openapi' version field in spec")
	}

	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing or invalid 'paths' object in spec")
	}

	expectedPaths := []string{
		"/api/sessions",
		"/api/sessions/{id}",
		"/api/sessions/{id}/tail",
		"/api/sessions/{id}/context",
		"/api/agent/context/global",
		"/api/agent/context/workspace",
		"/api/search",
		"/api/stats",
		"/api/refresh",
		"/api/watch-interval",
		"/api/wip",
		"/api/file",
		"/api/skill/install",
	}
	for _, ep := range expectedPaths {
		if _, exists := paths[ep]; !exists {
			t.Errorf("expected path %s in OpenAPI spec", ep)
		}
	}

	// 2. Verify handleSwaggerDocs renders HTML
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()

	handleSwaggerDocs(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 from /docs, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Session Explorer — OpenAPI & Swagger Documentation") {
		t.Errorf("expected Swagger / OpenAPI documentation title in /docs response")
	}
	if !strings.Contains(body, "executeGetWIP") {
		t.Errorf("expected interactive runner in /docs response")
	}
	if !strings.Contains(body, "executeGetAgentSession") {
		t.Errorf("expected session agent context runner in /docs response")
	}
	if !strings.Contains(body, "executeGetTail") {
		t.Errorf("expected live SSE tail runner in /docs response")
	}
}

func TestAPI_WatchInterval(t *testing.T) {
	idx := setupTestIndex(t)
	handler := NewAPIHandler(idx, false)

	// 1. GET default interval
	reqGet := httptest.NewRequest(http.MethodGet, "/api/watch-interval", nil)
	recGet := httptest.NewRecorder()
	handler.HandleWatchInterval(recGet, reqGet)

	if recGet.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recGet.Code)
	}

	var respGet WatchIntervalResponse
	if err := json.NewDecoder(recGet.Body).Decode(&respGet); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if respGet.IntervalSeconds != 10 {
		t.Errorf("expected default interval 10s, got %d", respGet.IntervalSeconds)
	}
	if respGet.Status != "active" {
		t.Errorf("expected status 'active', got %s", respGet.Status)
	}

	// 2. POST update to 60s
	bodyPost := bytes.NewBufferString(`{"interval_seconds": 60}`)
	reqPost := httptest.NewRequest(http.MethodPost, "/api/watch-interval", bodyPost)
	recPost := httptest.NewRecorder()
	handler.HandleWatchInterval(recPost, reqPost)

	if recPost.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recPost.Code)
	}

	var respPost WatchIntervalResponse
	if err := json.NewDecoder(recPost.Body).Decode(&respPost); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if respPost.IntervalSeconds != 60 {
		t.Errorf("expected updated interval 60s, got %d", respPost.IntervalSeconds)
	}
	if handler.GetWatchInterval() != 60 {
		t.Errorf("expected handler interval to be 60, got %d", handler.GetWatchInterval())
	}

	// 3. POST update to 0 (paused)
	bodyPause := bytes.NewBufferString(`{"interval_seconds": 0}`)
	reqPause := httptest.NewRequest(http.MethodPost, "/api/watch-interval", bodyPause)
	recPause := httptest.NewRecorder()
	handler.HandleWatchInterval(recPause, reqPause)

	var respPause WatchIntervalResponse
	if err := json.NewDecoder(recPause.Body).Decode(&respPause); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if respPause.Status != "paused" {
		t.Errorf("expected status 'paused', got %s", respPause.Status)
	}
}

func TestAPI_SessionTail_InitialReplay(t *testing.T) {
	idx := setupTestIndex(t)
	handler := NewAPIHandler(idx, false)

	// Create a context that cancels after 100ms so the SSE loop terminates cleanly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/session-api-1/tail?last_step=-1", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.HandleSessionTail(rec, req)

	output := rec.Body.String()
	if !strings.Contains(output, "event: connected") {
		t.Errorf("expected 'event: connected' in SSE output, got: %s", output)
	}
	if !strings.Contains(output, "event: step") {
		t.Errorf("expected at least one 'event: step' in SSE output, got: %s", output)
	}
	if !strings.Contains(output, "Test prompt for API") {
		t.Errorf("expected 'Test prompt for API' in SSE output, got: %s", output)
	}
}

func TestAPI_HandleFile(t *testing.T) {
	idx := setupTestIndex(t)
	handler := NewAPIHandler(idx, false)

	// 1. Create a temporary test file with markdown and code
	tmpDir := t.TempDir()
	sampleMd := filepath.Join(tmpDir, "sample.md")
	mdContent := "# Architecture Blueprint\n\n```mermaid\nflowchart TD\n  A --> B\n```\n\nDetailed specifications here."
	if err := os.WriteFile(sampleMd, []byte(mdContent), 0644); err != nil {
		t.Fatalf("failed to write sample md: %v", err)
	}

	// Test reading markdown file
	req := httptest.NewRequest(http.MethodGet, "/api/file?path="+sampleMd, nil)
	rec := httptest.NewRecorder()
	handler.HandleFile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp FileContentResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode FileContentResponse: %v", err)
	}

	if resp.Name != "sample.md" {
		t.Errorf("expected name 'sample.md', got %q", resp.Name)
	}
	if resp.Extension != ".md" {
		t.Errorf("expected extension '.md', got %q", resp.Extension)
	}
	if resp.IsBinary {
		t.Errorf("expected is_binary false for markdown, got true")
	}
	if resp.Content != mdContent {
		t.Errorf("expected content %q, got %q", mdContent, resp.Content)
	}
	if resp.LineCount != 8 {
		t.Errorf("expected line_count 8, got %d", resp.LineCount)
	}

	// 2. Test reading with anchor suffix: sample.md#L3-L5
	reqAnchor := httptest.NewRequest(http.MethodGet, "/api/file?path="+sampleMd+"%23L3-L5", nil)
	recAnchor := httptest.NewRecorder()
	handler.HandleFile(recAnchor, reqAnchor)
	if recAnchor.Code != http.StatusOK {
		t.Errorf("expected status 200 with #L anchor, got %d", recAnchor.Code)
	}

	// 3. Test reading with line number suffix: sample.md:3
	reqColon := httptest.NewRequest(http.MethodGet, "/api/file?path="+sampleMd+":3", nil)
	recColon := httptest.NewRecorder()
	handler.HandleFile(recColon, reqColon)
	if recColon.Code != http.StatusOK {
		t.Errorf("expected status 200 with :3 line suffix, got %d", recColon.Code)
	}

	// 4. Test missing path parameter
	reqNoPath := httptest.NewRequest(http.MethodGet, "/api/file", nil)
	recNoPath := httptest.NewRecorder()
	handler.HandleFile(recNoPath, reqNoPath)
	if recNoPath.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing path, got %d", recNoPath.Code)
	}

	// 5. Test non-existent file (404)
	reqMissing := httptest.NewRequest(http.MethodGet, "/api/file?path=/tmp/definitely_not_existent_file_12345.xyz", nil)
	recMissing := httptest.NewRecorder()
	handler.HandleFile(recMissing, reqMissing)
	if recMissing.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for missing file, got %d", recMissing.Code)
	}

	// 6. Test directory path (400)
	reqDir := httptest.NewRequest(http.MethodGet, "/api/file?path="+tmpDir, nil)
	recDir := httptest.NewRecorder()
	handler.HandleFile(recDir, reqDir)
	if recDir.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for directory path, got %d", recDir.Code)
	}

	// 7. Test restricted sensitive path (403)
	reqRestricted := httptest.NewRequest(http.MethodGet, "/api/file?path=/Users/test/.ssh/id_rsa", nil)
	recRestricted := httptest.NewRecorder()
	handler.HandleFile(recRestricted, reqRestricted)
	if recRestricted.Code != http.StatusForbidden {
		t.Errorf("expected status 403 for restricted .ssh path, got %d", recRestricted.Code)
	}
}

func TestAPI_HandleSessions_TimezoneOffsetFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Session created at 2026-11-07T00:15:00Z (which corresponds to 2026-11-06 19:15:00 EST / UTC-5)
	contentEve := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-11-07T00:15:00Z","content":"<USER_REQUEST>Evening session</USER_REQUEST>"}`
	createMockSession(t, tmpDir, "session-evening", contentEve)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("failed to scan test index: %v", err)
	}

	handler := NewAPIHandler(idx, false)

	// 1. Query with EST timezone offset (tz_offset=300 minutes) for 2026-11-06 to 2026-11-06
	reqEST := httptest.NewRequest(http.MethodGet, "/api/sessions?from=2026-11-06&to=2026-11-06&tz_offset=300", nil)
	recEST := httptest.NewRecorder()
	handler.HandleSessions(recEST, reqEST)

	if recEST.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recEST.Code)
	}

	var respEST SessionsResponse
	if err := json.NewDecoder(recEST.Body).Decode(&respEST); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(respEST.Sessions) != 1 {
		t.Fatalf("expected 1 session with EST tz_offset, got %d", len(respEST.Sessions))
	}
	if respEST.Sessions[0].ID != "session-evening" {
		t.Errorf("expected session-evening, got %s", respEST.Sessions[0].ID)
	}

	// 2. Query with UTC timezone offset (tz_offset=0) for 2026-11-06 to 2026-11-06
	// In UTC, 2026-11-07T00:15:00Z is on the next day, so it should NOT be returned.
	reqUTC := httptest.NewRequest(http.MethodGet, "/api/sessions?from=2026-11-06&to=2026-11-06&tz_offset=0", nil)
	recUTC := httptest.NewRecorder()
	handler.HandleSessions(recUTC, reqUTC)

	var respUTC SessionsResponse
	if err := json.NewDecoder(recUTC.Body).Decode(&respUTC); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(respUTC.Sessions) != 0 {
		t.Fatalf("expected 0 sessions with UTC tz_offset=0 for 2026-11-06, got %d", len(respUTC.Sessions))
	}
}

func TestAPI_HandleOpenInIDE(t *testing.T) {
	idx := setupTestIndex(t)
	handler := NewAPIHandler(idx, false)

	// 1. Method not allowed
	reqGet := httptest.NewRequest(http.MethodGet, "/api/ide/open", nil)
	recGet := httptest.NewRecorder()
	handler.HandleOpenInIDE(recGet, reqGet)
	if recGet.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", recGet.Code)
	}

	// 2. Invalid JSON
	reqBadJSON := httptest.NewRequest(http.MethodPost, "/api/ide/open", bytes.NewBufferString("not-json"))
	recBadJSON := httptest.NewRecorder()
	handler.HandleOpenInIDE(recBadJSON, reqBadJSON)
	if recBadJSON.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for invalid JSON, got %d", recBadJSON.Code)
	}

	// 3. Non-existent file
	reqNonExistent := httptest.NewRequest(http.MethodPost, "/api/ide/open", bytes.NewBufferString(`{"path":"/tmp/non_existent_file_xyz_123.rs","editor":"vscode","line":10}`))
	recNonExistent := httptest.NewRecorder()
	handler.HandleOpenInIDE(recNonExistent, reqNonExistent)
	if recNonExistent.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found for non-existent file, got %d", recNonExistent.Code)
	}

	// 4. Create real temporary file
	tmpFile, err := os.CreateTemp("", "test_ide_open_*.rs")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("// test content\nfn main() {}\n")
	tmpFile.Close()

	// 5. Test multiple editors on real file
	editors := []struct {
		editor      string
		expectedURI string
	}{
		{"cursor", fmt.Sprintf("cursor://file/%s:42", tmpFile.Name())},
		{"vscode", fmt.Sprintf("vscode://file/%s:42", tmpFile.Name())},
		{"windsurf", fmt.Sprintf("windsurf://file/%s:42", tmpFile.Name())},
		{"idea", fmt.Sprintf("idea://open?file=%s&line=42", tmpFile.Name())},
		{"sublime", fmt.Sprintf("subl://open?url=file://%s&line=42", tmpFile.Name())},
		{"system", fmt.Sprintf("file://%s", tmpFile.Name())},
	}

	for _, tc := range editors {
		payload := fmt.Sprintf(`{"path":"%s","editor":"%s","line":42}`, tmpFile.Name(), tc.editor)
		req := httptest.NewRequest(http.MethodPost, "/api/ide/open", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()
		handler.HandleOpenInIDE(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 OK for editor %s, got %d: %s", tc.editor, rec.Code, rec.Body.String())
			continue
		}

		var resp OpenInIDEResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Errorf("failed to decode response for editor %s: %v", tc.editor, err)
			continue
		}

		if resp.Editor != tc.editor && !(tc.editor == "idea" && resp.Editor == "intellij") {
			t.Errorf("expected editor %s, got %s", tc.editor, resp.Editor)
		}
		if resp.FallbackURI != tc.expectedURI {
			t.Errorf("expected fallback URI %s, got %s", tc.expectedURI, resp.FallbackURI)
		}
		if resp.Command == "" {
			t.Errorf("expected non-empty command for editor %s", tc.editor)
		}
	}
}

func TestAPI_HandleDetectIDEs(t *testing.T) {
	idx := setupTestIndex(t)
	handler := NewAPIHandler(idx, false)

	// 1. Method Not Allowed
	reqPost := httptest.NewRequest(http.MethodPost, "/api/ide/detected", nil)
	recPost := httptest.NewRecorder()
	handler.HandleDetectIDEs(recPost, reqPost)
	if recPost.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", recPost.Code)
	}

	// 2. Successful GET
	reqGet := httptest.NewRequest(http.MethodGet, "/api/ide/detected", nil)
	recGet := httptest.NewRecorder()
	handler.HandleDetectIDEs(recGet, reqGet)

	if recGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", recGet.Code, recGet.Body.String())
	}

	var resp DetectIDEsResponse
	if err := json.NewDecoder(recGet.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode DetectIDEsResponse: %v", err)
	}

	if len(resp.Detected) == 0 {
		t.Errorf("expected at least 1 known editor in detected list")
	}

	// Verify System Default is present and installed
	foundSystem := false
	for _, ide := range resp.Detected {
		if ide.ID == "system" {
			foundSystem = true
			if !ide.Installed {
				t.Errorf("expected system default editor to be installed")
			}
			break
		}
	}
	if !foundSystem {
		t.Errorf("expected 'system' editor in detected list")
	}

	if resp.Preferred == "" {
		t.Errorf("expected non-empty preferred IDE")
	}
}



