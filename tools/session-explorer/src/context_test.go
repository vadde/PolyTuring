package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestR039_DiscoverGlobalAgentContext tests discovery of machine-level rules and skills.
// Implements: R-039, AC-040
func TestR039_DiscoverGlobalAgentContext(t *testing.T) {
	tmpDir := t.TempDir()
	idx := NewSessionIndex(tmpDir, false)

	ctx := idx.DiscoverGlobalAgentContext()

	// On any real machine with ~/.gemini/config, rules or skills should be discovered
	t.Logf("Discovered %d global rules, %d global skills, %d plugins, %d MCP servers, %d hooks",
		len(ctx.Rules), len(ctx.Skills), len(ctx.Plugins), len(ctx.MCPServers), len(ctx.Hooks))

	if ctx.TotalRules != len(ctx.Rules) {
		t.Errorf("expected TotalRules == len(Rules), got %d vs %d", ctx.TotalRules, len(ctx.Rules))
	}
	if ctx.TotalSkills != len(ctx.Skills) {
		t.Errorf("expected TotalSkills == len(Skills), got %d vs %d", ctx.TotalSkills, len(ctx.Skills))
	}
}

// TestR039_DiscoverWorkspaceAgentContext_Cascade tests workspace rule and skill discovery,
// precedence cascade, and master documentation detection.
// Implements: R-039, AC-041
func TestR039_DiscoverWorkspaceAgentContext_Cascade(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Setup mock workspace with .agents/rules, .agents/skills, and AGENTS.md
	wsDir := filepath.Join(tmpDir, "mock-repo")
	agentsDir := filepath.Join(wsDir, ".agents")
	rulesDir := filepath.Join(agentsDir, "rules")
	skillsDir := filepath.Join(agentsDir, "skills", "test-skill")

	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rulesDir: %v", err)
	}
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("failed to create skillsDir: %v", err)
	}

	// Create AGENTS.md
	masterDoc := filepath.Join(wsDir, "AGENTS.md")
	if err := os.WriteFile(masterDoc, []byte("# 🤖 AGENTS.md — Master Agent Instructions\n\n> **Priority**: 🔴 Critical\n\nSingle source of truth."), 0644); err != nil {
		t.Fatalf("failed to create AGENTS.md: %v", err)
	}

	// Create workspace rule
	ruleFile := filepath.Join(rulesDir, "01-custom-rule.md")
	if err := os.WriteFile(ruleFile, []byte("# Custom Rule 01\n\n> **Priority**: 🔴 Critical\n\nAlways write unit tests first."), 0644); err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

	// Create workspace skill
	skillFile := filepath.Join(skillsDir, "SKILL.md")
	skillContent := `---
name: test-skill
description: Comprehensive testing skill for mock repo.
---

# Test Skill Runbook

Run the tests with care.`
	if err := os.WriteFile(skillFile, []byte(skillContent), 0644); err != nil {
		t.Fatalf("failed to create skill: %v", err)
	}

	idx := NewSessionIndex(tmpDir, false)
	idx.knownProjects["mock-repo"] = wsDir

	wctx := idx.DiscoverWorkspaceAgentContext("mock-repo")

	if wctx.ProjectName != "mock-repo" {
		t.Errorf("expected ProjectName 'mock-repo', got %q", wctx.ProjectName)
	}

	// Verify master document was discovered
	foundMaster := false
	for _, doc := range wctx.MasterDocs {
		if doc == "AGENTS.md" {
			foundMaster = true
			break
		}
	}
	if !foundMaster {
		t.Errorf("expected AGENTS.md in MasterDocs, got %v", wctx.MasterDocs)
	}

	// Verify workspace rule was discovered with ScopeWorkspace
	foundWorkspaceRule := false
	for _, r := range wctx.Rules {
		if r.Name == "01-custom-rule.md" {
			foundWorkspaceRule = true
			if r.Scope != ScopeWorkspace {
				t.Errorf("expected ScopeWorkspace, got %q", r.Scope)
			}
			if r.Workspace != "mock-repo" {
				t.Errorf("expected Workspace 'mock-repo', got %q", r.Workspace)
			}
		}
	}
	if !foundWorkspaceRule {
		t.Errorf("expected 01-custom-rule.md to be discovered in workspace rules")
	}

	// Verify workspace skill was discovered with ScopeWorkspace
	foundWorkspaceSkill := false
	for _, s := range wctx.Skills {
		if s.Name == "test-skill" {
			foundWorkspaceSkill = true
			if s.Scope != ScopeWorkspace {
				t.Errorf("expected ScopeWorkspace, got %q", s.Scope)
			}
		}
	}
	if !foundWorkspaceSkill {
		t.Errorf("expected test-skill to be discovered in workspace skills")
	}

	// Verify Token & Word metrics calculation (R-039 GenAI Bloat Telemetry)
	if wctx.PreloadedTokensEst <= 0 {
		t.Errorf("expected PreloadedTokensEst > 0, got %d", wctx.PreloadedTokensEst)
	}
	if wctx.OnDemandTokensEst <= 0 {
		t.Errorf("expected OnDemandTokensEst > 0, got %d", wctx.OnDemandTokensEst)
	}
	if wctx.TotalWords <= 0 {
		t.Errorf("expected TotalWords > 0, got %d", wctx.TotalWords)
	}
	if wctx.TotalTokensEst != wctx.PreloadedTokensEst+wctx.OnDemandTokensEst {
		t.Errorf("expected TotalTokensEst == Preloaded + OnDemand, got %d vs %d+%d",
			wctx.TotalTokensEst, wctx.PreloadedTokensEst, wctx.OnDemandTokensEst)
	}
}

// TestR039_ExtractSessionAgentContext_SkillInvocations tests that tool calls and reads
// referencing skills are tracked as ActiveSkills with InvokedCount and FirstStepIdx.
// Implements: R-039, AC-042
func TestR039_ExtractSessionAgentContext_SkillInvocations(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup workspace with skill
	wsDir := filepath.Join(tmpDir, "repo-with-skills")
	skillDir := filepath.Join(wsDir, ".agents", "skills", "deploy-tool")
	os.MkdirAll(skillDir, 0755)
	skillPath := filepath.Join(skillDir, "SKILL.md")
	os.WriteFile(skillPath, []byte("---\nname: deploy-tool\ndescription: Deploys the service.\n---\n# Deploy Tool"), 0644)

	// Create session that invokes deploy-tool at step 1
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-10-27T10:00:00Z","content":"<USER_REQUEST>Deploy the tool</USER_REQUEST>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-10-27T10:01:00Z","content":"Reading deploy skill","tool_calls":[{"name":"read_file","args":{"Path":"` + skillPath + `"}}]}
{"step_index":2,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-10-27T10:02:00Z","content":"Deploying again","tool_calls":[{"name":"read_file","args":{"Path":"` + skillPath + `"}}]}
{"step_index":3,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-10-27T10:03:00Z","content":"<USER_REQUEST>/plan next step</USER_REQUEST>"}`

	createMockSession(t, tmpDir, "sess-skill-test", content)

	idx := NewSessionIndex(tmpDir, false)
	idx.knownProjects["repo-with-skills"] = wsDir

	if err := idx.ScanAll(); err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	sctx, err := idx.ExtractSessionAgentContext("sess-skill-test")
	if err != nil {
		t.Fatalf("ExtractSessionAgentContext failed: %v", err)
	}

	// Verify deploy-tool is in ActiveSkills with InvokedCount == 2
	foundActive := false
	for _, sk := range sctx.ActiveSkills {
		if sk.Name == "deploy-tool" {
			foundActive = true
			if sk.InvokedCount != 2 {
				t.Errorf("expected InvokedCount 2, got %d", sk.InvokedCount)
			}
			if sk.FirstStepIdx != 1 {
				t.Errorf("expected FirstStepIdx 1, got %d", sk.FirstStepIdx)
			}
		}
	}
	if !foundActive {
		t.Errorf("expected deploy-tool to be in ActiveSkills, got: %v", sctx.ActiveSkills)
	}

	// Verify slash command /plan was detected
	foundPlan := false
	for _, cmd := range sctx.SlashCommands {
		if cmd == "/plan" {
			foundPlan = true
			break
		}
	}
	if !foundPlan {
		t.Errorf("expected /plan in SlashCommands, got %v", sctx.SlashCommands)
	}

	// Verify Session-level Token & Word Metrics
	if sctx.ActiveTokensEst <= 0 {
		t.Errorf("expected ActiveTokensEst > 0, got %d", sctx.ActiveTokensEst)
	}
	if sctx.TotalWordsInjected <= 0 {
		t.Errorf("expected TotalWordsInjected > 0, got %d", sctx.TotalWordsInjected)
	}
}

// TestR039_API_AgentContextEndpoints tests the REST API endpoints.
// Implements: R-039, AC-040, AC-041, AC-042
func TestR039_API_AgentContextEndpoints(t *testing.T) {
	idx := setupTestIndex(t)
	handler := NewAPIHandler(idx, false)

	// 1. GET /api/agent/context/global
	reqGlobal := httptest.NewRequest(http.MethodGet, "/api/agent/context/global", nil)
	recGlobal := httptest.NewRecorder()
	handler.HandleGlobalAgentContext(recGlobal, reqGlobal)

	if recGlobal.Code != http.StatusOK {
		t.Fatalf("expected status 200 for global context, got %d", recGlobal.Code)
	}
	var gResp GlobalAgentContext
	if err := json.NewDecoder(recGlobal.Body).Decode(&gResp); err != nil {
		t.Fatalf("failed to decode GlobalAgentContext: %v", err)
	}

	// 2. GET /api/agent/context/workspace?workspace=demo
	reqWs := httptest.NewRequest(http.MethodGet, "/api/agent/context/workspace?workspace=demo", nil)
	recWs := httptest.NewRecorder()
	handler.HandleWorkspaceAgentContext(recWs, reqWs)

	if recWs.Code != http.StatusOK {
		t.Fatalf("expected status 200 for workspace context, got %d", recWs.Code)
	}
	var wsResp WorkspaceAgentContext
	if err := json.NewDecoder(recWs.Body).Decode(&wsResp); err != nil {
		t.Fatalf("failed to decode WorkspaceAgentContext: %v", err)
	}
	if wsResp.ProjectName != "demo" {
		t.Errorf("expected ProjectName 'demo', got %q", wsResp.ProjectName)
	}

	// 3. GET /api/sessions/session-api-1/context
	reqSess := httptest.NewRequest(http.MethodGet, "/api/sessions/session-api-1/context", nil)
	recSess := httptest.NewRecorder()
	handler.HandleSessionContext(recSess, reqSess)

	if recSess.Code != http.StatusOK {
		t.Fatalf("expected status 200 for session context, got %d", recSess.Code)
	}
	var sessResp SessionAgentContext
	if err := json.NewDecoder(recSess.Body).Decode(&sessResp); err != nil {
		t.Fatalf("failed to decode SessionAgentContext: %v", err)
	}
	if sessResp.SessionID != "session-api-1" {
		t.Errorf("expected SessionID 'session-api-1', got %q", sessResp.SessionID)
	}
}

// TestR040_ExtractSessionAgentContext_MCPAndToolTelemetry tests the empirical extraction
// of MCP servers, active tools, dormant tools, native agent tools, and analytical diversity metrics.
// Implements: R-040, AC-046, AC-047, AC-048
func TestR040_ExtractSessionAgentContext_MCPAndToolTelemetry(t *testing.T) {
	tmpDir := t.TempDir()

	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-11-04T10:00:00Z","content":"<USER_REQUEST>Search and navigate</USER_REQUEST>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T10:01:00Z","content":"Searching AST","tool_calls":[{"name":"call_mcp_tool","args":{"ServerName":"\"omni-graph\"","ToolName":"\"omni_search\"","Arguments":"{\"query\":\"ast query\"}"}}]}
{"step_index":2,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T10:02:00Z","content":"Searching again","tool_calls":[{"name":"call_mcp_tool","args":{"ServerName":"omni-graph","ToolName":"omni_search","Arguments":"{\"query\":\"second query\"}"}}]}
{"step_index":3,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T10:03:00Z","content":"Looking up definition","tool_calls":[{"name":"call_mcp_tool","args":{"ServerName":"omni-graph","ToolName":"omni_definition","Arguments":"{\"symbol\":\"MyClass\"}"}}]}
{"step_index":4,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T10:04:00Z","content":"Navigating web page","tool_calls":[{"name":"call_mcp_tool","args":{"ServerName":"playwright","ToolName":"browser_navigate","Arguments":"{\"url\":\"http://localhost:3000\"}"}}]}
{"step_index":5,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T10:05:00Z","content":"Reading local file","tool_calls":[{"name":"read_file","args":{"Path":"/repo/main.go"}}]}
{"step_index":6,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T10:06:00Z","content":"Running test suite","tool_calls":[{"name":"run_command","args":{"CommandLine":"go test ./..."}}]}`

	createMockSession(t, tmpDir, "sess-mcp-telemetry", content)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	sctx, err := idx.ExtractSessionAgentContext("sess-mcp-telemetry")
	if err != nil {
		t.Fatalf("ExtractSessionAgentContext failed: %v", err)
	}

	// 1. Verify Overall Tool Analytics
	if sctx.ToolAnalytics.TotalToolCalls != 6 {
		t.Errorf("expected TotalToolCalls == 6, got %d", sctx.ToolAnalytics.TotalToolCalls)
	}
	if sctx.ToolAnalytics.TotalMCPCalls != 4 {
		t.Errorf("expected TotalMCPCalls == 4, got %d", sctx.ToolAnalytics.TotalMCPCalls)
	}
	if sctx.ToolAnalytics.TotalNativeCalls != 2 {
		t.Errorf("expected TotalNativeCalls == 2, got %d", sctx.ToolAnalytics.TotalNativeCalls)
	}
	if sctx.ToolAnalytics.UniqueToolsUsed != 5 {
		t.Errorf("expected UniqueToolsUsed == 5, got %d", sctx.ToolAnalytics.UniqueToolsUsed)
	}
	if sctx.ToolAnalytics.ToolDiversityIndex <= 0.0 {
		t.Errorf("expected ToolDiversityIndex > 0, got %f", sctx.ToolAnalytics.ToolDiversityIndex)
	}

	// 2. Verify MCP Servers and Tools
	foundOmni := false
	foundPlaywright := false
	for _, srv := range sctx.MCPServers {
		if srv.Name == "omni-graph" {
			foundOmni = true
			if srv.InvokedCount != 3 {
				t.Errorf("expected omni-graph InvokedCount == 3, got %d", srv.InvokedCount)
			}
			if len(srv.ActiveTools) != 2 {
				t.Errorf("expected 2 active tools on omni-graph, got %d", len(srv.ActiveTools))
			}
			for _, tool := range srv.ActiveTools {
				if tool.Name == "omni_search" {
					if tool.InvokedCount != 2 {
						t.Errorf("expected omni_search InvokedCount == 2, got %d", tool.InvokedCount)
					}
					if tool.FirstStepIdx != 1 || tool.LastStepIdx != 2 {
						t.Errorf("expected omni_search steps 1..2, got %d..%d", tool.FirstStepIdx, tool.LastStepIdx)
					}
				}
			}
		}
		if srv.Name == "playwright" {
			foundPlaywright = true
			if srv.InvokedCount != 1 {
				t.Errorf("expected playwright InvokedCount == 1, got %d", srv.InvokedCount)
			}
		}
	}
	if !foundOmni {
		t.Errorf("expected omni-graph MCP server to be discovered")
	}
	if !foundPlaywright {
		t.Errorf("expected playwright MCP server to be discovered")
	}

	// 3. Verify Native Tools
	foundReadFile := false
	foundRunCommand := false
	for _, nt := range sctx.NativeTools {
		if nt.Name == "read_file" {
			foundReadFile = true
			if nt.Category != "Filesystem" {
				t.Errorf("expected read_file Category 'Filesystem', got %q", nt.Category)
			}
			if nt.InvokedCount != 1 || nt.FirstStepIdx != 5 {
				t.Errorf("expected read_file 1 call at step 5, got %d at step %d", nt.InvokedCount, nt.FirstStepIdx)
			}
		}
		if nt.Name == "run_command" {
			foundRunCommand = true
			if nt.Category != "Terminal" {
				t.Errorf("expected run_command Category 'Terminal', got %q", nt.Category)
			}
			if nt.InvokedCount != 1 || nt.FirstStepIdx != 6 {
				t.Errorf("expected run_command 1 call at step 6, got %d at step %d", nt.InvokedCount, nt.FirstStepIdx)
			}
		}
	}
	if !foundReadFile {
		t.Errorf("expected read_file in NativeTools")
	}
	if !foundRunCommand {
		t.Errorf("expected run_command in NativeTools")
	}
}

// Implements: R-041, AC-053, AC-054, AC-055
func TestR041_ExtractSessionAgentContext_CommandAndToolBreakdown(t *testing.T) {
	tmpDir := t.TempDir()

	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-11-04T11:00:00Z","content":"<USER_REQUEST>Build, test, and commit</USER_REQUEST>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T11:01:00Z","content":"Checking git","tool_calls":[{"name":"run_command","args":{"CommandLine":"git status --porcelain","toolSummary":"Git status check"}}]}
{"step_index":2,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T11:02:00Z","content":"Building web bundle","tool_calls":[{"name":"run_command","args":{"CommandLine":"cd tools/session-explorer && make build-web","toolSummary":"Web asset bundling"}}]}
{"step_index":3,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T11:03:00Z","content":"Running unit tests","tool_calls":[{"name":"run_command","args":{"CommandLine":"go test -v ./...","toolSummary":"Go test execution"}}]}
{"step_index":4,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T11:04:00Z","content":"Running bun test","tool_calls":[{"name":"run_command","args":{"CommandLine":"bun test","toolSummary":"Bun test suite"}}]}
{"step_index":5,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T11:05:00Z","content":"Reading Go file","tool_calls":[{"name":"read_file","args":{"Path":"/repo/src/server.go"}}]}
{"step_index":6,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T11:06:00Z","content":"Reading JS file","tool_calls":[{"name":"read_file","args":{"Path":"/repo/src/web/app.js"}}]}
{"step_index":7,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T11:07:00Z","content":"Editing CSS file","tool_calls":[{"name":"edit_file","args":{"Path":"/repo/src/web/styles.css"}}]}
{"step_index":8,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T11:08:00Z","content":"Checking port","tool_calls":[{"name":"run_command","args":{"CommandLine":"lsof -i :9876","toolSummary":"Port check"}}]}
{"step_index":9,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T11:09:00Z","content":"Running curl","tool_calls":[{"name":"run_command","args":{"CommandLine":"curl -s http://localhost:9876/api/version","toolSummary":"Health probe"}}]}`

	createMockSession(t, tmpDir, "sess-command-breakdown", content)

	idx := NewSessionIndex(tmpDir, false)
	if err := idx.ScanAll(); err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	sctx, err := idx.ExtractSessionAgentContext("sess-command-breakdown")
	if err != nil {
		t.Fatalf("ExtractSessionAgentContext failed: %v", err)
	}

	// 1. Check NativeTools run_command sub-breakdown
	var runCmdTool *SessionNativeTool
	for i := range sctx.NativeTools {
		if sctx.NativeTools[i].Name == "run_command" {
			runCmdTool = &sctx.NativeTools[i]
			break
		}
	}
	if runCmdTool == nil {
		t.Fatalf("expected run_command in NativeTools")
	}
	if runCmdTool.InvokedCount != 6 {
		t.Errorf("expected run_command InvokedCount == 6, got %d", runCmdTool.InvokedCount)
	}
	if len(runCmdTool.SubBreakdown) == 0 {
		t.Fatalf("expected run_command SubBreakdown to be populated")
	}

	// Map sub-breakdowns by key
	cmdMap := make(map[string]ToolSubBreakdown)
	for _, sb := range runCmdTool.SubBreakdown {
		cmdMap[sb.Key] = sb
	}

	// Verify git status decomposed
	if gitSub, ok := cmdMap["git status"]; !ok {
		t.Errorf("expected 'git status' in run_command sub_breakdown, keys were: %+v", runCmdTool.SubBreakdown)
	} else {
		if gitSub.Category != "VCS" {
			t.Errorf("expected git status category 'VCS', got %q", gitSub.Category)
		}
		if len(gitSub.StepIndices) != 1 || gitSub.StepIndices[0] != 1 {
			t.Errorf("expected git status step indices [1], got %v", gitSub.StepIndices)
		}
	}

	// Verify make decomposed from cd ... && make
	if makeSub, ok := cmdMap["make build-web"]; !ok {
		t.Errorf("expected 'make build-web' in run_command sub_breakdown")
	} else {
		if makeSub.Category != "Build" {
			t.Errorf("expected make category 'Build', got %q", makeSub.Category)
		}
	}

	// Verify go test decomposed
	if goSub, ok := cmdMap["go test"]; !ok {
		t.Errorf("expected 'go test' in run_command sub_breakdown")
	} else {
		if goSub.Category != "Test" {
			t.Errorf("expected go test category 'Test', got %q", goSub.Category)
		}
	}

	// Verify lsof decomposed to Process
	if lsofSub, ok := cmdMap["lsof"]; !ok {
		t.Errorf("expected 'lsof' in run_command sub_breakdown")
	} else {
		if lsofSub.Category != "Process" {
			t.Errorf("expected lsof category 'Process', got %q", lsofSub.Category)
		}
	}

	// Verify curl decomposed to Network
	if curlSub, ok := cmdMap["curl"]; !ok {
		t.Errorf("expected 'curl' in run_command sub_breakdown")
	} else {
		if curlSub.Category != "Network" {
			t.Errorf("expected curl category 'Network', got %q", curlSub.Category)
		}
	}

	// 2. Check Global CommandBreakdown in ToolAnalytics
	if len(sctx.ToolAnalytics.CommandBreakdown) == 0 {
		t.Fatalf("expected sctx.ToolAnalytics.CommandBreakdown to be populated")
	}
	totalCmdPercentage := 0.0
	for _, cb := range sctx.ToolAnalytics.CommandBreakdown {
		totalCmdPercentage += cb.Percentage
	}
	if totalCmdPercentage < 95.0 || totalCmdPercentage > 105.0 {
		t.Errorf("expected total command percentage ~100%%, got %f", totalCmdPercentage)
	}

	// 3. Check FileBreakdown in ToolAnalytics
	if len(sctx.ToolAnalytics.FileBreakdown) == 0 {
		t.Fatalf("expected sctx.ToolAnalytics.FileBreakdown to be populated")
	}
	fileExtMap := make(map[string]ToolSubBreakdown)
	for _, fb := range sctx.ToolAnalytics.FileBreakdown {
		fileExtMap[fb.Key] = fb
	}
	if _, ok := fileExtMap[".go"]; !ok {
		t.Errorf("expected .go in FileBreakdown")
	}
	if _, ok := fileExtMap[".js"]; !ok {
		t.Errorf("expected .js in FileBreakdown")
	}
	if _, ok := fileExtMap[".css"]; !ok {
		t.Errorf("expected .css in FileBreakdown")
	}
}

// TestR042_ExtractSessionAgentContext_SkillBreakdownAndDeduplication tests that skills
// are deduplicated across scopes, step indices and last step indices are captured,
// and standby skills are properly segregated with complete telemetry.
// Implements: R-042, AC-059, AC-060, AC-065
func TestR042_ExtractSessionAgentContext_SkillBreakdownAndDeduplication(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup workspace with two skills
	wsDir := filepath.Join(tmpDir, "repo-skills-test")
	skillDir1 := filepath.Join(wsDir, ".agents", "skills", "omni-graph")
	os.MkdirAll(skillDir1, 0755)
	skillPath1 := filepath.Join(skillDir1, "SKILL.md")
	os.WriteFile(skillPath1, []byte("---\nname: omni-graph\ndescription: Knowledge graph and AST navigation.\n---\n# Omni Graph"), 0644)

	skillDir2 := filepath.Join(wsDir, ".agents", "skills", "dormant-skill")
	os.MkdirAll(skillDir2, 0755)
	skillPath2 := filepath.Join(skillDir2, "SKILL.md")
	os.WriteFile(skillPath2, []byte("---\nname: dormant-skill\ndescription: Dormant on-demand capability.\n---\n# Dormant"), 0644)

	// Create session that invokes omni-graph at steps 1, 3, and 5 via read_file and run_command
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-11-04T10:00:00Z","content":"<USER_REQUEST>Search the codebase</USER_REQUEST>"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T10:01:00Z","content":"Reading skill","tool_calls":[{"name":"read_file","args":{"Path":"` + skillPath1 + `"}}]}
{"step_index":3,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T10:03:00Z","content":"Inspecting skill again","tool_calls":[{"name":"view_file","args":{"AbsolutePath":"` + skillPath1 + `"}}]}
{"step_index":5,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T10:05:00Z","content":"Running helper script","tool_calls":[{"name":"run_command","args":{"CommandLine":"skills/omni-graph/scripts/check.sh"}}]}
`

	createMockSession(t, tmpDir, "sess-skill-r042", content)

	idx := NewSessionIndex(tmpDir, false)
	idx.knownProjects["repo-skills-test"] = wsDir

	if err := idx.ScanAll(); err != nil {
		t.Fatalf("ScanAll failed: %v", err)
	}

	sctx, err := idx.ExtractSessionAgentContext("sess-skill-r042")
	if err != nil {
		t.Fatalf("ExtractSessionAgentContext failed: %v", err)
	}

	// 1. Verify omni-graph is in ActiveSkills with InvokedCount == 3
	if len(sctx.ActiveSkills) != 1 {
		t.Fatalf("expected 1 active skill, got %d", len(sctx.ActiveSkills))
	}
	active := sctx.ActiveSkills[0]
	if active.Name != "omni-graph" {
		t.Errorf("expected omni-graph, got %s", active.Name)
	}
	if active.InvokedCount != 3 {
		t.Errorf("expected InvokedCount 3, got %d", active.InvokedCount)
	}
	if active.FirstStepIdx != 1 {
		t.Errorf("expected FirstStepIdx 1, got %d", active.FirstStepIdx)
	}
	if active.LastStepIdx != 5 {
		t.Errorf("expected LastStepIdx 5, got %d", active.LastStepIdx)
	}
	if len(active.StepIndices) != 3 {
		t.Errorf("expected 3 step indices, got %v", active.StepIndices)
	}
	if len(active.InvocationTriggers) == 0 {
		t.Errorf("expected invocation triggers to be populated, got %v", active.InvocationTriggers)
	}

	// 2. Verify dormant-skill is in AvailableSkills and NOT duplicated
	foundDormant := false
	for _, sk := range sctx.AvailableSkills {
		if sk.Name == "dormant-skill" {
			foundDormant = true
			if sk.InvokedCount != 0 {
				t.Errorf("expected dormant skill InvokedCount 0, got %d", sk.InvokedCount)
			}
		}
		if sk.Name == "omni-graph" {
			t.Errorf("active skill omni-graph should not appear in AvailableSkills")
		}
	}
	if !foundDormant {
		t.Errorf("expected dormant-skill in AvailableSkills")
	}
}
