// Tests for scanner.go
// Implements test coverage for R-001, R-002, R-003, R-004, R-005, R-010, R-011, R-012, R-014, R-017, NF-007
package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Helper to create a mock brain session directory
func createMockSession(t *testing.T, baseDir, sessionID, content string) {
	t.Helper()
	logDir := filepath.Join(baseDir, sessionID, ".system_generated", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("failed to create mock log dir: %v", err)
	}

	filePath := filepath.Join(logDir, "transcript.jsonl")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write mock transcript: %v", err)
	}
}

// TestR001_SessionDiscovery tests session directory discovery
func TestR001_SessionDiscovery(t *testing.T) {
	tmpDir := t.TempDir()

	mockJSONL := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-18T10:00:00Z","content":"<USER_REQUEST>\nHello world test\n</USER_REQUEST>\n<ADDITIONAL_METADATA>\nActive Document: /Users/test/workspace/main.go\n</ADDITIONAL_METADATA>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-09-18T10:01:00Z","content":"Here is the response","tool_calls":[{"name":"view_file","args":{"AbsolutePath":"/Users/test/workspace/main.go"}}]}`

	createMockSession(t, tmpDir, "session-abc-123", mockJSONL)
	createMockSession(t, tmpDir, "session-def-456", mockJSONL)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	if len(idx.sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(idx.sessions))
	}
}

// TestR002_ParseJSONL tests JSONL parsing and NF-007 malformed handling
func TestR002_ParseJSONL(t *testing.T) {
	tmpDir := t.TempDir()

	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-18T12:00:00Z","content":"<USER_REQUEST>First prompt</USER_REQUEST>"}
not-a-valid-json-line
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-09-18T12:01:00Z","content":"Response 1"}
{"step_index":2,"source":"SYSTEM","type":"ERROR_MESSAGE","status":"ERROR","created_at":"2026-09-18T12:02:00Z","content":"Something went wrong"}`

	createMockSession(t, tmpDir, "session-parse-test", content)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	if len(idx.sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(idx.sessions))
	}

	s := idx.sessions[0]
	if s.StepCount != 3 {
		t.Errorf("expected 3 valid steps, got %d", s.StepCount)
	}
	if s.UserMessageCount != 1 {
		t.Errorf("expected 1 user message, got %d", s.UserMessageCount)
	}
	if s.AgentResponseCount != 1 {
		t.Errorf("expected 1 agent response, got %d", s.AgentResponseCount)
	}
	if s.ErrorCount != 1 {
		t.Errorf("expected 1 error count, got %d", s.ErrorCount)
	}
}

// TestR003_ExtractUserPrompt tests stripping metadata from prompts
func TestR003_ExtractUserPrompt(t *testing.T) {
	raw := `<USER_REQUEST>
Build a distributed cache in Go
</USER_REQUEST>
<ADDITIONAL_METADATA>
The current local time is: 2026-09-18T10:00:00.
Active Document: /foo/bar.go
</ADDITIONAL_METADATA>`

	extracted := extractUserPrompt(raw)
	expected := "Build a distributed cache in Go"
	if extracted != expected {
		t.Errorf("expected %q, got %q", expected, extracted)
	}
}

// TestR005_ExtractWorkspace tests workspace extraction patterns
func TestR005_ExtractWorkspace(t *testing.T) {
	contentWithDoc := `Active Document: /Users/test/projects/my-cool-app/src/main.go`
	ws := extractWorkspace(contentWithDoc)
	if ws != "/Users/test/projects/my-cool-app/src" {
		t.Errorf("expected /Users/test/projects/my-cool-app/src, got %q", ws)
	}

	proj := extractProjectName(ws)
	if proj != "my-cool-app" {
		t.Errorf("expected my-cool-app, got %q", proj)
	}
}

// TestR010_FilterDateRange tests filtering sessions by date
func TestR010_FilterDateRange(t *testing.T) {
	tmpDir := t.TempDir()

	s1 := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-10T10:00:00Z","content":"<USER_REQUEST>Prompt 1</USER_REQUEST>"}`
	s2 := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-15T10:00:00Z","content":"<USER_REQUEST>Prompt 2</USER_REQUEST>"}`
	s3 := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-20T10:00:00Z","content":"<USER_REQUEST>Prompt 3</USER_REQUEST>"}`

	createMockSession(t, tmpDir, "sess-1", s1)
	createMockSession(t, tmpDir, "sess-2", s2)
	createMockSession(t, tmpDir, "sess-3", s3)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	from, _ := time.Parse(time.RFC3339, "2026-09-12T00:00:00Z")
	to, _ := time.Parse(time.RFC3339, "2026-09-18T00:00:00Z")

	filtered := idx.GetSessions("", "date", "desc", from, to)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 session in date range, got %d", len(filtered))
	}
	if filtered[0].ID != "sess-2" {
		t.Errorf("expected sess-2, got %s", filtered[0].ID)
	}
}

// TestR012_Search tests full-text search across transcripts
func TestR012_Search(t *testing.T) {
	tmpDir := t.TempDir()

	s1 := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-18T10:00:00Z","content":"<USER_REQUEST>Implement Kubernetes ingress controller</USER_REQUEST>"}`
	s2 := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-18T11:00:00Z","content":"<USER_REQUEST>Write terraform modules for AWS VPC</USER_REQUEST>"}`

	createMockSession(t, tmpDir, "sess-k8s", s1)
	createMockSession(t, tmpDir, "sess-tf", s2)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	res := idx.Search("kubernetes", "", "", time.Time{}, time.Time{})
	if res.TotalCount != 1 {
		t.Fatalf("expected 1 search result, got %d", res.TotalCount)
	}
	if res.Results[0].SessionID != "sess-k8s" {
		t.Errorf("expected sess-k8s, got %s", res.Results[0].SessionID)
	}
}

// TestR017_Stats tests aggregate stats computation
func TestR017_Stats(t *testing.T) {
	tmpDir := t.TempDir()

	s1 := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-18T10:00:00Z","content":"<USER_REQUEST>Fix bug</USER_REQUEST>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-09-18T10:01:00Z","content":"Fixed","tool_calls":[{"name":"replace_file_content","args":{}},{"name":"view_file","args":{}}]}`

	createMockSession(t, tmpDir, "sess-stats", s1)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	stats := idx.GetStats()
	if stats.TotalSessions != 1 {
		t.Errorf("expected 1 total session, got %d", stats.TotalSessions)
	}
	if stats.TotalUserMessages != 1 {
		t.Errorf("expected 1 user message, got %d", stats.TotalUserMessages)
	}
	if stats.TotalAgentResponses != 1 {
		t.Errorf("expected 1 agent response, got %d", stats.TotalAgentResponses)
	}
	if stats.TotalToolCalls != 2 {
		t.Errorf("expected 2 tool calls, got %d", stats.TotalToolCalls)
	}
	if len(stats.TopToolCalls) != 2 {
		t.Errorf("expected 2 top tools, got %d", len(stats.TopToolCalls))
	}
}

// TestR028_ProjectClustering tests repository and project clustering from tool calls and metadata
func TestR028_ProjectClustering(t *testing.T) {
	tmpDir := t.TempDir()

	// Mock session where user has unrelated active doc, but agent performs tool calls inside tutor-intelligence
	sTutor := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-18T10:00:00Z","content":"<USER_REQUEST>\nLet's implement the robot puzzle\n</USER_REQUEST>\n<ADDITIONAL_METADATA>\nActive Document: /Users/test/knowledge/vadde.github.io/README.md\n</ADDITIONAL_METADATA>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-09-18T10:01:00Z","content":"Inspecting files","tool_calls":[{"name":"view_file","args":{"AbsolutePath":"/Users/test/knowledge/tutor-intelligence/src/robot.py"}},{"name":"run_command","args":{"Cwd":"/Users/test/knowledge/tutor-intelligence","CommandLine":"pytest"}}]}`

	// Mock session for tool-scripts
	sTools := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-18T11:00:00Z","content":"<USER_REQUEST>\nBuild the session explorer tool\n</USER_REQUEST>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-09-18T11:01:00Z","content":"Writing code","tool_calls":[{"name":"write_to_file","args":{"TargetFile":"/Users/test/knowledge/tool-scripts/tools/session-explorer/src/main.go"}}]}`

	createMockSession(t, tmpDir, "sess-tutor", sTutor)
	createMockSession(t, tmpDir, "sess-tools", sTools)

	idx := NewSessionIndex(tmpDir, false)
	// Register test known projects
	idx.knownProjects["tutor-intelligence"] = "/Users/test/knowledge/tutor-intelligence"
	idx.knownProjects["tool-scripts"] = "/Users/test/knowledge/tool-scripts"
	idx.knownProjects["vadde.github.io"] = "/Users/test/knowledge/vadde.github.io"

	if err := idx.ScanAll(); err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	if len(idx.sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(idx.sessions))
	}

	// Verify tutor-intelligence session was accurately assigned
	var tutorSess, toolsSess *Session
	for i := range idx.sessions {
		if idx.sessions[i].ID == "sess-tutor" {
			tutorSess = &idx.sessions[i]
		} else if idx.sessions[i].ID == "sess-tools" {
			toolsSess = &idx.sessions[i]
		}
	}

	if tutorSess == nil {
		t.Fatal("sess-tutor not found")
	}
	if tutorSess.ProjectName != "tutor-intelligence" {
		t.Errorf("expected ProjectName tutor-intelligence, got %q", tutorSess.ProjectName)
	}
	if len(tutorSess.ProjectsTouched) == 0 || tutorSess.ProjectsTouched[0] != "tutor-intelligence" {
		t.Errorf("expected ProjectsTouched to have tutor-intelligence first, got %v", tutorSess.ProjectsTouched)
	}

	if toolsSess == nil {
		t.Fatal("sess-tools not found")
	}
	if toolsSess.ProjectName != "tool-scripts" {
		t.Errorf("expected ProjectName tool-scripts, got %q", toolsSess.ProjectName)
	}

	// Verify stats.Projects contains tutor-intelligence and tool-scripts
	stats := idx.GetStats()
	if len(stats.Projects) == 0 {
		t.Fatal("expected stats.Projects to be populated")
	}

	foundTutor := false
	for _, p := range stats.Projects {
		if p.Name == "tutor-intelligence" {
			foundTutor = true
			if p.SessionCount != 1 {
				t.Errorf("expected 1 session for tutor-intelligence, got %d", p.SessionCount)
			}
			if len(p.RecentPrompts) != 1 || p.RecentPrompts[0] != "Let's implement the robot puzzle" {
				t.Errorf("unexpected recent prompts for tutor-intelligence: %v", p.RecentPrompts)
			}
		}
	}
	if !foundTutor {
		t.Error("tutor-intelligence project not found in stats.Projects")
	}
}

// TestR028_ProjectClustering_PortableFallback verifies project clustering in headless CI environments
// where knownProjects is empty and paths are arbitrary.
func TestR028_ProjectClustering_PortableFallback(t *testing.T) {
	tmpDir := t.TempDir()

	sLinuxCI := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-18T10:00:00Z","content":"<USER_REQUEST>Run tests</USER_REQUEST>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-09-18T10:01:00Z","content":"Running","tool_calls":[{"name":"run_command","args":{"Cwd":"/home/runner/work/awesome-service/src"}}]}`

	createMockSession(t, tmpDir, "sess-ci", sLinuxCI)

	idx := NewSessionIndex(tmpDir, false)
	// Clear any discovered knownProjects to simulate headless container
	idx.knownProjects = make(map[string]string)

	if err := idx.ScanAll(); err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	if len(idx.sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(idx.sessions))
	}

	if idx.sessions[0].ProjectName != "awesome-service" {
		t.Errorf("expected ProjectName awesome-service, got %q", idx.sessions[0].ProjectName)
	}

	stats := idx.GetStats()
	if len(stats.Projects) != 1 || stats.Projects[0].Name != "awesome-service" {
		t.Errorf("expected stats.Projects to contain awesome-service, got %+v", stats.Projects)
	}
}

// TestRescanStatsDeduplication verifies that repeated scans do NOT accumulate workspaces or stats
func TestRescanStatsDeduplication(t *testing.T) {
	tmpDir := t.TempDir()

	s1 := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-18T10:00:00Z","content":"<USER_REQUEST>Build tool</USER_REQUEST>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-09-18T10:01:00Z","content":"ok","tool_calls":[{"name":"view_file","args":{"AbsolutePath":"/Users/test/knowledge/QuarkDock/docs/sizing.md"}}]}`

	s2 := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-09-18T11:00:00Z","content":"<USER_REQUEST>Do tests</USER_REQUEST>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-09-18T11:01:00Z","content":"done","tool_calls":[{"name":"view_file","args":{"AbsolutePath":"/Users/test/knowledge/DSA/problem_1.py"}}]}`

	createMockSession(t, tmpDir, "sess-1", s1)
	createMockSession(t, tmpDir, "sess-2", s2)

	idx := NewSessionIndex(tmpDir, false)
	idx.knownProjects = map[string]string{
		"QuarkDock": "/Users/test/knowledge/QuarkDock",
		"DSA":       "/Users/test/knowledge/DSA",
	}

	// Scan 5 times in a row to simulate live watch rescans
	for i := 0; i < 5; i++ {
		if err := idx.ScanAll(); err != nil {
			t.Fatalf("ScanAll failed on iteration %d: %v", i, err)
		}
	}

	stats := idx.GetStats()

	// 1. TotalSessions must remain exactly 2, NOT 10!
	if stats.TotalSessions != 2 {
		t.Fatalf("expected TotalSessions to be 2 after 5 rescans, got %d", stats.TotalSessions)
	}

	// 2. Workspaces slice must not have duplicates
	seen := make(map[string]bool)
	for _, ws := range stats.Workspaces {
		if seen[ws] {
			t.Fatalf("duplicate workspace %q detected in stats.Workspaces: %v", ws, stats.Workspaces)
		}
		seen[ws] = true
	}

	if len(stats.Workspaces) != 2 {
		t.Fatalf("expected exactly 2 unique workspaces, got %d: %v", len(stats.Workspaces), stats.Workspaces)
	}
}

// TestExtractProjectName_Sanitization verifies that files, usernames, and internal subdirs are rejected
func TestExtractProjectName_Sanitization(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"/tmp/all_mermaids.txt", "Default"},
		{"/Users/aparv/something.py", "Default"},
		{"/workspace/specs/_templates/tool-spec.md", "PolyTuring"},
		{"/workspace/tools/session-explorer/src/scanner.go", "PolyTuring"},
		{"/Users/aparv/knowledge/tool-scripts/tools/session-explorer/src/scanner.go", "PolyTuring"},
		{"/Users/aparv/knowledge/PolyTuring/tools/session-explorer/src/scanner.go", "PolyTuring"},
		{"/Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/QuarkDock/docs/sizing.md", "QuarkDock"},
		{"/Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/DSA/binary_search.py", "DSA"},
		{"", "Default"},
	}

	for _, c := range cases {
		actual := extractProjectName(c.input)
		if actual != c.expected {
			t.Errorf("extractProjectName(%q) = %q, expected %q", c.input, actual, c.expected)
		}
	}

	// Test isValidWorkspaceName directly
	invalidNames := []string{"all_mermaids.txt", "aparv", "_templates", "docs", "scripts", "rules", "analytics", "Default", "workspace", "src", ""}
	for _, name := range invalidNames {
		if isValidWorkspaceName(name) {
			t.Errorf("expected isValidWorkspaceName(%q) to be false, got true", name)
		}
	}

	validNames := []string{"QuarkDock", "PolyTuring", "tool-scripts", "tutor-intelligence", "DSA", "GoLang", "GenAI"}
	for _, name := range validNames {
		if !isValidWorkspaceName(name) {
			t.Errorf("expected isValidWorkspaceName(%q) to be true, got false", name)
		}
	}
}

// TestR030_GetWorkspaceWIP_SingleSession tests WIP recovery when exactly 1 session exists
func TestR030_GetWorkspaceWIP_SingleSession(t *testing.T) {
	tmpDir := t.TempDir()

	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-10-14T10:00:00Z","content":"<USER_REQUEST>Implement feature X</USER_REQUEST>\n<ADDITIONAL_METADATA>\nActive Document: /Users/test/workspace/app/main.go\n</ADDITIONAL_METADATA>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-10-14T10:01:00Z","content":"I am editing the file now","tool_calls":[{"name":"write_to_file","args":{"TargetFile":"/Users/test/workspace/app/main.go","CodeContent":"package main\n"}},{"name":"run_command","args":{"CommandLine":"go test ./..."}}]}
{"step_index":2,"source":"SYSTEM","type":"ERROR_MESSAGE","status":"ERROR","created_at":"2026-10-14T10:02:00Z","content":"compile error: undefined foo"}`

	createMockSession(t, tmpDir, "session-wip-1", content)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	wip, err := idx.GetWorkspaceWIP("app", "", false, 5)
	if err != nil {
		t.Fatalf("GetWorkspaceWIP failed: %v", err)
	}

	if wip.Status != "resolved" {
		t.Errorf("expected status 'resolved', got %q", wip.Status)
	}
	if wip.ActiveSession == nil {
		t.Fatalf("expected non-nil ActiveSession")
	}
	if wip.ActiveSession.SessionID != "session-wip-1" {
		t.Errorf("expected session-wip-1, got %s", wip.ActiveSession.SessionID)
	}
	if len(wip.ActiveSession.RecentPrompts) != 1 {
		t.Errorf("expected 1 recent prompt, got %d", len(wip.ActiveSession.RecentPrompts))
	}
	if len(wip.ActiveSession.ModifiedFiles) != 1 || wip.ActiveSession.ModifiedFiles[0] != "/Users/test/workspace/app/main.go" {
		t.Errorf("expected modified file /Users/test/workspace/app/main.go, got %v", wip.ActiveSession.ModifiedFiles)
	}
	if len(wip.ActiveSession.ExecutedCommands) != 1 || wip.ActiveSession.ExecutedCommands[0] != "go test ./..." {
		t.Errorf("expected executed command 'go test ./...', got %v", wip.ActiveSession.ExecutedCommands)
	}
	if len(wip.ActiveSession.UnresolvedErrors) != 1 {
		t.Errorf("expected 1 unresolved error, got %d", len(wip.ActiveSession.UnresolvedErrors))
	}
}

// TestR031_GetWorkspaceWIP_MultiSession_AmbiguousHalt tests mandatory halt when multiple sessions exist
func TestR031_GetWorkspaceWIP_MultiSession_AmbiguousHalt(t *testing.T) {
	tmpDir := t.TempDir()

	contentOld := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-10-10T10:00:00Z","content":"<USER_REQUEST>Old session prompt</USER_REQUEST>\n<ADDITIONAL_METADATA>\nActive Document: /Users/test/workspace/app/old.go\n</ADDITIONAL_METADATA>"}`
	contentNew := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-10-14T10:00:00Z","content":"<USER_REQUEST>New session prompt</USER_REQUEST>\n<ADDITIONAL_METADATA>\nActive Document: /Users/test/workspace/app/new.go\n</ADDITIONAL_METADATA>"}`

	createMockSession(t, tmpDir, "session-old-100", contentOld)
	createMockSession(t, tmpDir, "session-new-200", contentNew)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	// 1. Without forceLatest and without sessionID -> MUST return ambiguous_multi_session and candidates
	wip, err := idx.GetWorkspaceWIP("app", "", false, 5)
	if err != nil {
		t.Fatalf("GetWorkspaceWIP failed: %v", err)
	}

	if wip.Status != "ambiguous_multi_session" {
		t.Fatalf("expected status 'ambiguous_multi_session', got %q", wip.Status)
	}
	if wip.ActiveSession != nil {
		t.Errorf("expected nil ActiveSession on ambiguous halt")
	}
	if len(wip.Candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(wip.Candidates))
	}
	// Verify candidates are sorted newest first
	if wip.Candidates[0].SessionID != "session-new-200" {
		t.Errorf("expected newest candidate first (session-new-200), got %s", wip.Candidates[0].SessionID)
	}
	if wip.Candidates[1].SessionID != "session-old-100" {
		t.Errorf("expected older candidate second (session-old-100), got %s", wip.Candidates[1].SessionID)
	}
	if wip.Instruction == "" {
		t.Errorf("expected instruction to halt and prompt user")
	}

	// 2. With forceLatest=true -> Must resolve directly to session-new-200
	wipLatest, err := idx.GetWorkspaceWIP("app", "", true, 5)
	if err != nil {
		t.Fatalf("GetWorkspaceWIP with forceLatest failed: %v", err)
	}
	if wipLatest.Status != "resolved" {
		t.Fatalf("expected status 'resolved' with forceLatest, got %q", wipLatest.Status)
	}
	if wipLatest.ActiveSession.SessionID != "session-new-200" {
		t.Errorf("expected session-new-200, got %s", wipLatest.ActiveSession.SessionID)
	}

	// 3. With direct sessionID -> Must resolve directly to selected session
	wipExplicit, err := idx.GetWorkspaceWIP("app", "session-old-100", false, 5)
	if err != nil {
		t.Fatalf("GetWorkspaceWIP with sessionID failed: %v", err)
	}
	if wipExplicit.Status != "resolved" {
		t.Fatalf("expected status 'resolved' with explicit sessionID, got %q", wipExplicit.Status)
	}
	if wipExplicit.ActiveSession.SessionID != "session-old-100" {
		t.Errorf("expected session-old-100, got %s", wipExplicit.ActiveSession.SessionID)
	}
}

// TestR030_GetWorkspaceWIP_NotFound tests behavior when workspace has no sessions
func TestR030_GetWorkspaceWIP_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	wip, err := idx.GetWorkspaceWIP("nonexistent-project", "", false, 5)
	if err != nil {
		t.Fatalf("GetWorkspaceWIP failed: %v", err)
	}
	if wip.Status != "not_found" {
		t.Errorf("expected status 'not_found', got %q", wip.Status)
	}
}

// TestR035_GranularMessageTypeSearch tests searching with granular message filters (Implements: R-035, AC-023, AC-024)
func TestR035_GranularMessageTypeSearch(t *testing.T) {
	tmpDir := t.TempDir()

	s1 := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-10-15T10:00:00Z","content":"<USER_REQUEST>How do I fix this database connection bug?</USER_REQUEST>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-10-15T10:01:00Z","content":"Let me inspect the connection pool config to fix it.","tool_calls":[{"name":"view_file","args":{"path":"db.go"}}]}
{"step_index":2,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-10-15T10:02:00Z","content":"Here is the verified fix: increase the max open connections in database pool."}`

	createMockSession(t, tmpDir, "sess-filter", s1)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	// 1. Search with no type filter -> Should find all 3 steps matching "fix"
	resAll := idx.Search("fix", "", "", time.Time{}, time.Time{})
	if resAll.TotalCount != 3 {
		t.Fatalf("expected 3 results for all types, got %d", resAll.TotalCount)
	}

	// 2. Search with AGENT_RESPONSES -> Should only match Step 2 (prose response without tools)
	resResponses := idx.Search("fix", "", "AGENT_RESPONSES", time.Time{}, time.Time{})
	if resResponses.TotalCount != 1 {
		t.Fatalf("expected 1 result for AGENT_RESPONSES, got %d", resResponses.TotalCount)
	}
	if resResponses.Results[0].StepIndex != 2 {
		t.Errorf("expected step_index 2 for AGENT_RESPONSES, got %d", resResponses.Results[0].StepIndex)
	}

	// 3. Search with TOOL_CALLS -> Should only match Step 1 (has tool_calls)
	resTools := idx.Search("fix", "", "TOOL_CALLS", time.Time{}, time.Time{})
	if resTools.TotalCount != 1 {
		t.Fatalf("expected 1 result for TOOL_CALLS, got %d", resTools.TotalCount)
	}
	if resTools.Results[0].StepIndex != 1 {
		t.Errorf("expected step_index 1 for TOOL_CALLS, got %d", resTools.Results[0].StepIndex)
	}

	// 4. Search with DIALOGUE -> Should match Step 0 (USER_INPUT) and Step 2 (pure response), excluding Step 1
	resDialogue := idx.Search("fix", "", "DIALOGUE", time.Time{}, time.Time{})
	if resDialogue.TotalCount != 2 {
		t.Fatalf("expected 2 results for DIALOGUE, got %d", resDialogue.TotalCount)
	}
	for _, r := range resDialogue.Results {
		if r.StepIndex == 1 {
			t.Errorf("DIALOGUE filter must exclude tool call Step 1, but it was returned")
		}
	}
}

// TestR036_MultiRuntimeDiscovery verifies universal discovery across multi-runtime locations (IDE, CLI)
// Implements: R-036, AC-026, AC-027, AC-028, AC-029
func TestR036_MultiRuntimeDiscovery(t *testing.T) {
	mockIdeDir := t.TempDir()
	mockCliDir := t.TempDir()

	ideJSONL := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-10-15T10:00:00Z","content":"<USER_REQUEST>IDE prompt</USER_REQUEST>\n<ADDITIONAL_METADATA>\nActive Document: /Users/test/knowledge/my-web/index.html\n</ADDITIONAL_METADATA>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-10-15T10:01:00Z","content":"IDE response"}`

	cliJSONL := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-10-15T11:00:00Z","content":"<user_information>\nActive Workspaces:\n- /Users/test/knowledge/my-cli-tool\nCommand Working Directory: /Users/test/knowledge/my-cli-tool\n</user_information>\n<USER_REQUEST>CLI prompt</USER_REQUEST>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-10-15T11:01:00Z","content":"CLI response"}`

	createMockSession(t, mockIdeDir, "session-ide-01", ideJSONL)
	createMockSession(t, mockCliDir, "session-cli-01", cliJSONL)

	idx := NewSessionIndex(mockIdeDir, false)
	// Explicitly configure multi-runtime locations
	idx.locations = []DiscoveredLocation{
		{
			Path:        mockIdeDir,
			Provider:    "antigravity",
			RuntimeMode: "ide",
		},
		{
			Path:        mockCliDir,
			Provider:    "antigravity",
			RuntimeMode: "cli",
		},
	}

	if err := idx.ScanAll(); err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	if len(idx.sessions) != 2 {
		t.Fatalf("expected 2 sessions discovered across IDE and CLI locations, got %d", len(idx.sessions))
	}

	sessionMap := make(map[string]Session)
	for _, s := range idx.sessions {
		sessionMap[s.ID] = s
	}

	ideSess, ok := sessionMap["session-ide-01"]
	if !ok {
		t.Fatalf("missing session-ide-01 in index")
	}
	if ideSess.RuntimeMode != "ide" {
		t.Errorf("expected runtime_mode 'ide', got %q", ideSess.RuntimeMode)
	}
	if ideSess.Provider != "antigravity" {
		t.Errorf("expected provider 'antigravity', got %q", ideSess.Provider)
	}
	if ideSess.StorageDir != mockIdeDir {
		t.Errorf("expected storage_dir %q, got %q", mockIdeDir, ideSess.StorageDir)
	}

	cliSess, ok := sessionMap["session-cli-01"]
	if !ok {
		t.Fatalf("missing session-cli-01 in index")
	}
	if cliSess.RuntimeMode != "cli" {
		t.Errorf("expected runtime_mode 'cli', got %q", cliSess.RuntimeMode)
	}
	if cliSess.Provider != "antigravity" {
		t.Errorf("expected provider 'antigravity', got %q", cliSess.Provider)
	}
	if cliSess.StorageDir != mockCliDir {
		t.Errorf("expected storage_dir %q, got %q", mockCliDir, cliSess.StorageDir)
	}
	if cliSess.ProjectName != "my-cli-tool" {
		t.Errorf("expected project_name 'my-cli-tool', got %q", cliSess.ProjectName)
	}

	// Verify GetTranscriptPath correctly routes to respective directory
	idePath, err := idx.GetTranscriptPath("session-ide-01", false)
	if err != nil {
		t.Fatalf("GetTranscriptPath failed for IDE session: %v", err)
	}
	if !filepath.HasPrefix(idePath, mockIdeDir) {
		t.Errorf("expected IDE path inside %s, got %s", mockIdeDir, idePath)
	}

	cliPath, err := idx.GetTranscriptPath("session-cli-01", false)
	if err != nil {
		t.Fatalf("GetTranscriptPath failed for CLI session: %v", err)
	}
	if !filepath.HasPrefix(cliPath, mockCliDir) {
		t.Errorf("expected CLI path inside %s, got %s", mockCliDir, cliPath)
	}

	// Verify stats runtime breakdown
	stats := idx.GetStats()
	if stats.RuntimeBreakdown["ide"] != 1 {
		t.Errorf("expected 1 IDE session in runtime breakdown, got %d", stats.RuntimeBreakdown["ide"])
	}
	if stats.RuntimeBreakdown["cli"] != 1 {
		t.Errorf("expected 1 CLI session in runtime breakdown, got %d", stats.RuntimeBreakdown["cli"])
	}

	// Verify GetSessionDetail opens CLI session with zero path errors
	cliDetail, err := idx.GetSessionDetail("session-cli-01")
	if err != nil {
		t.Fatalf("GetSessionDetail failed for CLI session: %v", err)
	}
	if len(cliDetail.Messages) != 2 {
		t.Errorf("expected 2 messages in CLI session detail, got %d", len(cliDetail.Messages))
	}
}

