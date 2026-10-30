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
