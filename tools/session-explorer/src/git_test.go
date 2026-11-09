// Tests for git_service.go and Git API endpoints
// Implements: R-053, R-054, R-055, R-060, AC-117..AC-125

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestGitService_CleanPath(t *testing.T) {
	svc := NewGitService(nil)

	// Strip file:// prefix and #L anchor
	p := svc.CleanPath("file:///Users/aparv/test.go#L10-L20")
	if p != "/Users/aparv/test.go" {
		t.Errorf("Expected /Users/aparv/test.go, got %s", p)
	}

	// Empty path
	if svc.CleanPath("") != "" {
		t.Errorf("Expected empty path for empty input")
	}
}

func TestGitService_FileStatus(t *testing.T) {
	svc := NewGitService(nil)

	// Test on this repository's src/api.go
	apiPath, err := filepath.Abs("api.go")
	if err != nil {
		t.Fatalf("Failed to resolve api.go: %v", err)
	}

	status, err := svc.GetFileStatus(apiPath, "")
	if err != nil {
		t.Fatalf("GetFileStatus failed: %v", err)
	}

	if !status.IsGit {
		t.Errorf("Expected api.go to be in a git repo, got is_git=false")
	}
	if status.GitRoot == "" {
		t.Errorf("Expected non-empty git_root")
	}
	if status.Branch == "" {
		t.Errorf("Expected non-empty branch")
	}
	if status.RelativePath == "" {
		t.Errorf("Expected non-empty relative_path")
	}
}

func TestGitService_NonGitFile(t *testing.T) {
	svc := NewGitService(nil)

	// Create a temp file outside any git repo
	tmpDir := t.TempDir()
	// Even if /tmp might be somehow inside, a random nested path without .git
	nonGitFile := filepath.Join(tmpDir, "sample.txt")
	_ = os.WriteFile(nonGitFile, []byte("hello"), 0644)

	// Check if tempdir is in git (in mac /var/folders it usually isn't)
	if findGitRoot(nonGitFile) == "" {
		status, err := svc.GetFileStatus(nonGitFile, "")
		if err != nil {
			t.Fatalf("GetFileStatus on non-git file returned error: %v", err)
		}
		if status.IsGit {
			t.Errorf("Expected is_git=false for non-git file")
		}
	}
}

func TestGitService_FileCommits(t *testing.T) {
	svc := NewGitService(nil)

	apiPath, _ := filepath.Abs("scanner.go")
	commits, err := svc.GetFileCommits(apiPath, 10)
	if err != nil {
		t.Fatalf("GetFileCommits failed: %v", err)
	}

	if len(commits) == 0 {
		t.Errorf("Expected commits for scanner.go, got 0")
	} else {
		first := commits[0]
		if first.SHA == "" || first.ShortSHA == "" {
			t.Errorf("Expected commit SHA, got empty")
		}
		if first.Subject == "" {
			t.Errorf("Expected commit subject, got empty")
		}
		if first.Timestamp == 0 {
			t.Errorf("Expected commit timestamp > 0")
		}
	}
}

func TestGitService_ParseUnifiedDiff(t *testing.T) {
	sampleDiff := `diff --git a/file.go b/file.go
index 1234567..89abcdef 100644
--- a/file.go
+++ b/file.go
@@ -10,4 +10,5 @@
 func main() {
-    oldCall()
+    newCall()
+    extraLine()
 }
`

	hunks, adds, dels := ParseUnifiedDiff(sampleDiff)
	if len(hunks) != 1 {
		t.Fatalf("Expected 1 hunk, got %d", len(hunks))
	}
	if adds != 2 {
		t.Errorf("Expected 2 additions, got %d", adds)
	}
	if dels != 1 {
		t.Errorf("Expected 1 deletion, got %d", dels)
	}

	hunk := hunks[0]
	if hunk.OldStart != 10 || hunk.OldLines != 4 {
		t.Errorf("Expected old range 10,4, got %d,%d", hunk.OldStart, hunk.OldLines)
	}
	if hunk.NewStart != 10 || hunk.NewLines != 5 {
		t.Errorf("Expected new range 10,5, got %d,%d", hunk.NewStart, hunk.NewLines)
	}
	if len(hunk.Lines) != 5 {
		t.Errorf("Expected 5 diff lines, got %d", len(hunk.Lines))
	}
}

func TestAPI_HandleGitEndpoints(t *testing.T) {
	idx := &SessionIndex{
		sessions:      make([]Session, 0),
		fileCache:     make(map[string]cachedSessionInfo),
		knownProjects: make(map[string]string),
	}
	handler := NewAPIHandler(idx, false)

	apiPath, _ := filepath.Abs("api.go")

	// 1. GET /api/git/file/status
	req := httptest.NewRequest(http.MethodGet, "/api/git/file/status?path="+apiPath, nil)
	w := httptest.NewRecorder()
	handler.HandleGitFileStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 from HandleGitFileStatus, got %d: %s", w.Code, w.Body.String())
	}

	var status GitFileStatus
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Fatalf("Failed to parse status response: %v", err)
	}
	if !status.IsGit {
		t.Errorf("Expected is_git=true for api.go")
	}

	// 2. GET /api/git/file/commits
	reqCommits := httptest.NewRequest(http.MethodGet, "/api/git/file/commits?path="+apiPath+"&limit=5", nil)
	wCommits := httptest.NewRecorder()
	handler.HandleGitFileCommits(wCommits, reqCommits)

	if wCommits.Code != http.StatusOK {
		t.Errorf("Expected 200 from HandleGitFileCommits, got %d: %s", wCommits.Code, wCommits.Body.String())
	}

	var commits []GitCommitEntry
	if err := json.Unmarshal(wCommits.Body.Bytes(), &commits); err != nil {
		t.Fatalf("Failed to parse commits response: %v", err)
	}
	if len(commits) == 0 {
		t.Errorf("Expected >= 1 commit for api.go")
	}

	// 3. GET /api/git/file/diff
	reqDiff := httptest.NewRequest(http.MethodGet, "/api/git/file/diff?path="+apiPath+"&mode=unstaged", nil)
	wDiff := httptest.NewRecorder()
	handler.HandleGitFileDiff(wDiff, reqDiff)

	if wDiff.Code != http.StatusOK {
		t.Errorf("Expected 200 from HandleGitFileDiff, got %d: %s", wDiff.Code, wDiff.Body.String())
	}

	var diffResp GitDiffResponse
	if err := json.Unmarshal(wDiff.Body.Bytes(), &diffResp); err != nil {
		t.Fatalf("Failed to parse diff response: %v", err)
	}
	if diffResp.Mode != "unstaged" {
		t.Errorf("Expected mode=unstaged, got %s", diffResp.Mode)
	}
}

func TestAPI_HandleSessionFileSteps(t *testing.T) {
	tmpDir := t.TempDir()
	mockSessionID := "test-session-123"

	transcriptContent := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-11-09T10:00:00Z","content":"<USER_REQUEST>Refactor api.go</USER_REQUEST>","tool_calls":[]}
{"step_index":14,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-09T10:05:00Z","content":"Refactored function","thinking":"Refactoring api.go to support git diffs","tool_calls":[{"name":"edit_file","args":{"Path":"/Users/aparv/src/api.go","Target":"func oldFunc() {}","Replacement":"func newFunc() { log.Println(\"new\") }"}}]}`

	createMockSession(t, tmpDir, mockSessionID, transcriptContent)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("Failed to scan test index: %v", err)
	}

	handler := NewAPIHandler(idx, false)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+mockSessionID+"/file/steps?path=/Users/aparv/src/api.go", nil)
	w := httptest.NewRecorder()
	handler.HandleSessionFileSteps(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var steps []AgentStepDiff
	if err := json.Unmarshal(w.Body.Bytes(), &steps); err != nil {
		t.Fatalf("Failed to unmarshal steps: %v", err)
	}
	if len(steps) != 1 {
		t.Fatalf("Expected 1 step diff, got %d", len(steps))
	}
	if steps[0].StepIndex != 14 {
		t.Errorf("Expected step index 14, got %d", steps[0].StepIndex)
	}
	if steps[0].ToolName != "edit_file" {
		t.Errorf("Expected tool edit_file, got %s", steps[0].ToolName)
	}
	if steps[0].Additions != 1 || steps[0].Deletions != 1 {
		t.Errorf("Expected 1 add and 1 del, got +%d/-%d", steps[0].Additions, steps[0].Deletions)
	}
	if steps[0].Thinking != "Refactoring api.go to support git diffs" {
		t.Errorf("Expected thinking preserved, got %s", steps[0].Thinking)
	}
}
