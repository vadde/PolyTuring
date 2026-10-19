// Tests for api.go
// Implements test coverage for R-006, R-008, R-009, R-011, R-012, R-013, R-017
package main

import (
	"bytes"
	"context"
	"encoding/json"
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

	expectedPaths := []string{"/api/sessions", "/api/stats", "/api/wip", "/api/skill/install"}
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


