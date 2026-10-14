// Package main provides data types for the Session Explorer tool.
//
// Implements: R-002 (Parse JSONL transcript files)
// See: specs/catalog/session-explorer.md#R-002
package main

import (
	"time"
)

// TranscriptEntry represents a single line from a JSONL transcript file.
// Implements: R-002
type TranscriptEntry struct {
	StepIndex   int        `json:"step_index"`
	Source      string     `json:"source"`       // USER_EXPLICIT, MODEL, SYSTEM
	Type        string     `json:"type"`         // USER_INPUT, PLANNER_RESPONSE, etc.
	Status      string     `json:"status"`       // DONE, ERROR
	CreatedAt   time.Time  `json:"created_at"`
	Content     string     `json:"content,omitempty"`
	Thinking    string     `json:"thinking,omitempty"`
	ToolCalls   []ToolCall `json:"tool_calls,omitempty"`
	IsTruncated bool       `json:"is_truncated,omitempty"`
}

// ToolCall represents a tool invocation within a PLANNER_RESPONSE.
type ToolCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args,omitempty"`
}

// Session represents a discovered conversation session with computed metadata.
// Implements: R-001, R-008
type Session struct {
	ID                     string    `json:"id"`
	CreatedAt              time.Time `json:"created_at"`
	LastMessageAt          time.Time `json:"last_message_at"`
	Workspace              string    `json:"workspace"`
	ProjectName            string         `json:"project_name"`
	ProjectsTouched        []string       `json:"projects_touched,omitempty"`
	StepCount              int            `json:"step_count"`
	UserMessageCount       int            `json:"user_message_count"`
	AgentResponseCount     int            `json:"agent_response_count"`
	ToolCallCount          int            `json:"tool_call_count"`
	ToolCallsMap           map[string]int `json:"tool_calls_map,omitempty"`
	ErrorCount             int            `json:"error_count"`
	TranscriptSizeBytes    int64          `json:"transcript_size_bytes"`
	FirstUserPromptPreview string         `json:"first_user_prompt_preview"`
	HasFullTranscript      bool           `json:"has_full_transcript"`
}

// ProjectSummary provides clustered aggregate metrics for a repository or workspace.
// Implements: R-028 (Project & Repository Clustering)
type ProjectSummary struct {
	Name                string    `json:"name"`
	Path                string    `json:"path"`
	IsGit               bool      `json:"is_git"`
	SessionCount        int       `json:"session_count"`
	TotalSteps          int       `json:"total_steps"`
	TotalUserMessages   int       `json:"total_user_messages"`
	TotalAgentResponses int       `json:"total_agent_responses"`
	TotalToolCalls      int       `json:"total_tool_calls"`
	TotalErrors         int       `json:"total_errors"`
	LastActiveAt        time.Time `json:"last_active_at"`
	RecentPrompts       []string  `json:"recent_prompts"`
}

// SessionDetail is the full session with its conversation messages.
// Implements: R-009
type SessionDetail struct {
	Session  Session          `json:"session"`
	Messages []MessageDisplay `json:"messages"`
}

// MessageDisplay is a UI-friendly message representation.
type MessageDisplay struct {
	StepIndex   int               `json:"step_index"`
	Source      string            `json:"source"`
	Type        string            `json:"type"`
	Status      string            `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	Content     string            `json:"content,omitempty"`
	HasThinking bool              `json:"has_thinking"`
	Thinking    string            `json:"thinking,omitempty"`
	ToolCalls   []ToolCallDisplay `json:"tool_calls,omitempty"`
}

// ToolCallDisplay is a UI-friendly tool call representation.
type ToolCallDisplay struct {
	Name        string                 `json:"name"`
	ArgsPreview string                 `json:"args_preview,omitempty"`
	Args        map[string]interface{} `json:"args,omitempty"`
}

// SearchResult represents a matching entry from full-text search.
// Implements: R-012
type SearchResult struct {
	SessionID      string    `json:"session_id"`
	StepIndex      int       `json:"step_index"`
	Type           string    `json:"type"`
	Source         string    `json:"source"`
	ContentPreview string    `json:"content_preview"`
	CreatedAt      time.Time `json:"created_at"`
	Workspace      string    `json:"workspace"`
}

// Stats represents aggregate statistics across all sessions.
// Implements: R-017, R-025
type Stats struct {
	TotalSessions       int            `json:"total_sessions"`
	TotalUserMessages   int            `json:"total_user_messages"`
	TotalAgentResponses int            `json:"total_agent_responses"`
	TotalToolCalls      int            `json:"total_tool_calls"`
	TotalErrors         int            `json:"total_errors"`
	DateRange           DateRange      `json:"date_range"`
	Workspaces          []string         `json:"workspaces"`
	Projects            []ProjectSummary `json:"projects"`
	TopToolCalls        []ToolFreq       `json:"top_tool_calls"`
	TopPrompts          []PromptFreq     `json:"top_prompts"`
	MessagesByType      map[string]int   `json:"messages_by_type"`
	LastScannedAt       time.Time        `json:"last_scanned_at"`
}

// DateRange represents a from/to date pair.
type DateRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// ToolFreq tracks the frequency of tool calls.
type ToolFreq struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// PromptFreq tracks the frequency of common prompt patterns.
// Implements: R-025
type PromptFreq struct {
	Prompt string `json:"prompt"`
	Count  int    `json:"count"`
}

// SessionsResponse is the JSON response for GET /api/sessions.
type SessionsResponse struct {
	Sessions []Session `json:"sessions"`
	Stats    Stats     `json:"stats"`
}

// SessionDetailResponse is the JSON response for GET /api/sessions/:id.
type SessionDetailResponse struct {
	Session  Session          `json:"session"`
	Messages []MessageDisplay `json:"messages"`
}

// SearchResponse is the JSON response for GET /api/search.
type SearchResponse struct {
	Results    []SearchResult `json:"results"`
	TotalCount int            `json:"total_count"`
	Query      string         `json:"query"`
}

// Config holds the application configuration from CLI flags.
type Config struct {
	DataDir       string
	Port          int
	NoOpen        bool
	Verbose       bool
	WatchInterval int // seconds between background auto-rescans (0 = disabled)
	Command       string
	WorkspaceArg  string
	SessionIDArg  string
	ForceLatest   bool
	OutputJSON    bool
}

// SessionCandidate represents a summary of a session when resolving ambiguous multi-session targets.
// Implements: R-031 (Multi-Session Disambiguation Invariant)
type SessionCandidate struct {
	SessionID          string    `json:"session_id"`
	CreatedAt          time.Time `json:"created_at"`
	LastMessageAt      time.Time `json:"last_message_at"`
	RelativeTime       string    `json:"relative_time"`
	Workspace          string    `json:"workspace"`
	ProjectName        string    `json:"project_name"`
	StepCount          int       `json:"step_count"`
	UserMessageCount   int       `json:"user_message_count"`
	ErrorCount         int       `json:"error_count"`
	FirstPromptPreview string    `json:"first_prompt_preview"`
	LastPromptPreview  string    `json:"last_prompt_preview"`
}

// WIPDigest represents the synthesized work-in-progress state for an active session.
// Implements: R-030 (Agent WIP Recovery API)
type WIPDigest struct {
	SessionID             string            `json:"session_id"`
	Workspace             string            `json:"workspace"`
	ProjectName           string            `json:"project_name"`
	CreatedAt             time.Time         `json:"created_at"`
	LastMessageAt         time.Time         `json:"last_message_at"`
	RelativeTime          string            `json:"relative_time"`
	StepCount             int               `json:"step_count"`
	UserMessageCount      int               `json:"user_message_count"`
	ErrorCount            int               `json:"error_count"`
	RecentPrompts         []string          `json:"recent_prompts"`
	RecentResponses       []string          `json:"recent_responses"`
	ModifiedFiles         []string          `json:"modified_files"`
	ExecutedCommands      []string          `json:"executed_commands"`
	LastToolCalls         []ToolCallDisplay `json:"last_tool_calls"`
	UnresolvedErrors      []string          `json:"unresolved_errors"`
	SuggestedContinuation string            `json:"suggested_continuation,omitempty"`
}

// WIPResponse is the JSON response for GET /api/wip.
// Implements: R-030, R-031
type WIPResponse struct {
	Status         string             `json:"status"` // "resolved", "ambiguous_multi_session", "not_found"
	Multiple       bool               `json:"multiple"`
	Workspace      string             `json:"workspace"`
	CandidateCount int                `json:"candidate_count"`
	Candidates     []SessionCandidate `json:"candidates,omitempty"`
	ActiveSession  *WIPDigest         `json:"active_session,omitempty"`
	Instruction    string             `json:"instruction,omitempty"`
}

// InstallSkillRequest is the payload for POST /api/skill/install.
// Implements: R-033
type InstallSkillRequest struct {
	Target string `json:"target"` // "workspace" or "global"
}

// InstallSkillResponse is the response for POST /api/skill/install.
// Implements: R-033
type InstallSkillResponse struct {
	Success       bool   `json:"success"`
	Target        string `json:"target"`
	InstalledPath string `json:"installed_path"`
	Message       string `json:"message"`
}

