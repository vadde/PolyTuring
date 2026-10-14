// Tests for api.go
// Implements test coverage for R-006, R-008, R-009, R-011, R-012, R-013, R-017
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

