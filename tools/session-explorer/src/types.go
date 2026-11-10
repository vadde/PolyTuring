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

// DiscoveredLocation represents an assistant session store path and its runtime identity.
// Implements: R-036
type DiscoveredLocation struct {
	Path        string `json:"path"`
	Provider    string `json:"provider"`     // "antigravity", "claude", "copilot"
	RuntimeMode string `json:"runtime_mode"` // "ide", "cli", "desktop", "core", "custom"
}

// Session represents a discovered conversation session with computed metadata.
// Implements: R-001, R-008, R-036
type Session struct {
	ID                     string         `json:"id"`
	Provider               string         `json:"provider"`                 // e.g. "antigravity", "claude", "copilot"
	RuntimeMode            string         `json:"runtime_mode"`             // e.g. "ide", "cli", "desktop", "core"
	StorageDir             string         `json:"storage_dir,omitempty"`    // parent brain directory
	TranscriptPath         string         `json:"transcript_path,omitempty"`// absolute path to transcript.jsonl
	CreatedAt              time.Time      `json:"created_at"`
	LastMessageAt          time.Time      `json:"last_message_at"`
	Workspace              string         `json:"workspace"`
	ProjectName            string         `json:"project_name"`
	ProjectsTouched        []string       `json:"projects_touched,omitempty"`
	GitRepo                string         `json:"git_repo,omitempty"`
	GitBranch              string         `json:"git_branch,omitempty"`
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
	RuntimeBreakdown    map[string]int   `json:"runtime_breakdown,omitempty"`
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

// WatchIntervalRequest is the payload for POST /api/watch-interval.
type WatchIntervalRequest struct {
	IntervalSeconds int `json:"interval_seconds"`
}

// WatchIntervalResponse is the payload returned by GET and POST /api/watch-interval.
type WatchIntervalResponse struct {
	IntervalSeconds int    `json:"interval_seconds"`
	Status          string `json:"status"` // "active" or "paused"
	Message         string `json:"message"`
}

// FileContentResponse is the payload returned by GET /api/file.
// Implements: R-038 (Native In-Browser File & Artifact Viewer)
type FileContentResponse struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Size      int64  `json:"size"`
	ModTime   string `json:"mod_time"`
	Content   string `json:"content"`
	IsBinary  bool   `json:"is_binary"`
	MimeType  string `json:"mime_type"`
	LineCount int    `json:"line_count"`
}

// AgentContextScope defines whether a customization is global, workspace, session, builtin, or plugin.
// Implements: R-039 (Universal Agent Context & Behavioral Matrix)
type AgentContextScope string

const (
	ScopeGlobal    AgentContextScope = "global"
	ScopeWorkspace AgentContextScope = "workspace"
	ScopeSession   AgentContextScope = "session"
	ScopeBuiltin   AgentContextScope = "builtin"
	ScopePlugin    AgentContextScope = "plugin"
)

// AgentRule represents an instruction or behavioral guideline.
// Implements: R-039
type AgentRule struct {
	Name        string            `json:"name"`
	Title       string            `json:"title"`
	Scope       AgentContextScope `json:"scope"`                  // "global", "workspace", "builtin"
	FilePath    string            `json:"file_path"`
	Description string            `json:"description,omitempty"`
	Priority    string            `json:"priority,omitempty"`    // "Critical", "Important", "Constitutional", "Standard"
	Trigger     string            `json:"trigger,omitempty"`     // "always_on", "model_decision", "user_invoked"
	LineCount   int               `json:"line_count"`
	WordCount   int               `json:"word_count"`
	CharCount   int               `json:"char_count"`
	TokensEst   int               `json:"tokens_est"`
	SizeBytes   int64             `json:"size_bytes"`
	Provider    string            `json:"provider"`              // "antigravity", "claude", "copilot", "cursor"
	Workspace   string            `json:"workspace,omitempty"`
}

// AgentSkill represents an on-demand tool capability, runbook, or workflow.
// Implements: R-039
type AgentSkill struct {
	Name          string            `json:"name"`
	Scope         AgentContextScope `json:"scope"`               // "global", "workspace", "builtin", "plugin"
	FilePath      string            `json:"file_path"`
	Description   string            `json:"description"`
	Category      string            `json:"category,omitempty"`
	Provider      string            `json:"provider"`            // "antigravity", "claude", "copilot", "cursor"
	Workspace     string            `json:"workspace,omitempty"`
	PluginName    string            `json:"plugin_name,omitempty"`
	HasScripts    bool              `json:"has_scripts"`
	HasExamples   bool              `json:"has_examples"`
	HasReferences bool              `json:"has_references"`
	LineCount     int               `json:"line_count"`
	WordCount     int               `json:"word_count"`
	CharCount     int               `json:"char_count"`
	TokensEst     int               `json:"tokens_est"`
	SizeBytes     int64             `json:"size_bytes"`
	InvokedCount       int               `json:"invoked_count,omitempty"`       // Count of tool calls or reads in session
	FirstStepIdx       int               `json:"first_step_idx,omitempty"`      // Step index where first triggered
	LastStepIdx        int               `json:"last_step_idx,omitempty"`       // Step index where last triggered (R-042)
	StepIndices        []int             `json:"step_indices,omitempty"`        // Timeline step indices where invoked (R-042)
	InvocationTriggers []string          `json:"invocation_triggers,omitempty"` // Triggers: e.g. ["read_file: SKILL.md", "run_command: scripts/..."] (R-042)
}

// AgentPlugin represents a packaged bundle of skills, rules, policies, and configurations.
// Implements: R-039
type AgentPlugin struct {
	Name        string   `json:"name"`
	Version     string   `json:"version,omitempty"`
	Description string   `json:"description,omitempty"`
	Path        string   `json:"path"`
	Skills      []string `json:"skills,omitempty"`
	Policies    []string `json:"policies,omitempty"`
}

// MCPServerInfo represents a configured MCP tool provider.
// Implements: R-039
type MCPServerInfo struct {
	Name    string            `json:"name"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Scope   AgentContextScope `json:"scope"`
	Tools   []string          `json:"tools,omitempty"`
}

// AgentHook represents a lifecycle guardrail script or trigger hook.
// Implements: R-039
type AgentHook struct {
	Name    string            `json:"name"`
	Event   string            `json:"event"`   // "PreToolUse", "PreInvocation", "Stop"
	Matcher string            `json:"matcher,omitempty"`
	Type    string            `json:"type"`    // "command"
	Command string            `json:"command"`
	Scope   AgentContextScope `json:"scope"`
	Timeout int               `json:"timeout,omitempty"`
}

// GlobalAgentContext represents the machine-wide agent brain state.
// Implements: R-039
type GlobalAgentContext struct {
	Rules              []AgentRule     `json:"rules"`
	Skills             []AgentSkill    `json:"skills"`
	Plugins            []AgentPlugin   `json:"plugins"`
	MCPServers         []MCPServerInfo `json:"mcp_servers"`
	Hooks              []AgentHook     `json:"hooks"`
	TotalRules         int             `json:"total_rules"`
	TotalSkills        int             `json:"total_skills"`
	PreloadedTokensEst int             `json:"preloaded_tokens_est"` // Always-on / constitutional rules & master instructions
	OnDemandTokensEst  int             `json:"on_demand_tokens_est"` // Available skills & on-demand runbooks
	TotalTokensEst     int             `json:"total_tokens_est"`
	TotalWords         int             `json:"total_words"`
}

// WorkspaceAgentContext represents the merged contextual view for a specific repository.
// Implements: R-039
type WorkspaceAgentContext struct {
	Workspace          string          `json:"workspace"`
	ProjectName        string          `json:"project_name"`
	Rules              []AgentRule     `json:"rules"`       // Workspace rules + inherited global rules with scope tagging
	Skills             []AgentSkill    `json:"skills"`      // Workspace skills + inherited global/builtin skills
	MasterDocs         []string        `json:"master_docs"` // AGENTS.md, CLAUDE.md, .cursorrules
	MCPServers         []MCPServerInfo `json:"mcp_servers"`
	Hooks              []AgentHook     `json:"hooks"`
	TotalRules         int             `json:"total_rules"`
	TotalSkills        int             `json:"total_skills"`
	PreloadedTokensEst int             `json:"preloaded_tokens_est"`
	OnDemandTokensEst  int             `json:"on_demand_tokens_est"`
	TotalTokensEst     int             `json:"total_tokens_est"`
	TotalWords         int             `json:"total_words"`
}

// SessionMCPTool represents an empirical MCP tool execution in a session.
// Implements: R-040
type SessionMCPTool struct {
	Name              string `json:"name"`
	ServerName        string `json:"server_name"`
	InvokedCount      int    `json:"invoked_count"`
	StepIndices       []int  `json:"step_indices"`
	FirstStepIdx      int    `json:"first_step_idx"`
	LastStepIdx       int    `json:"last_step_idx"`
	ErrorCount        int    `json:"error_count"`
	RecentArgsSummary string `json:"recent_args_summary,omitempty"`
}

// SessionMCPServer represents an MCP server's presence and usage in a session.
// Implements: R-040
type SessionMCPServer struct {
	Name            string            `json:"name"`
	Command         string            `json:"command,omitempty"`
	Args            []string          `json:"args,omitempty"`
	Scope           AgentContextScope `json:"scope"`
	ConfiguredTools []string          `json:"configured_tools,omitempty"`
	InvokedCount    int               `json:"invoked_count"`
	ActiveTools     []SessionMCPTool  `json:"active_tools"`
	DormantTools    []string          `json:"dormant_tools,omitempty"`
}

// ToolSubBreakdown represents decomposed execution distribution for generic tools (e.g. run_command, read_file).
// Implements: R-041
type ToolSubBreakdown struct {
	Key         string   `json:"key"`                   // e.g. "git", "make", ".go", "research"
	Label       string   `json:"label"`                 // Clean human-readable display label
	Category    string   `json:"category,omitempty"`    // e.g. "VCS", "Build", "Test", "Process", "Inspect"
	Count       int      `json:"count"`                 // Number of executions
	Percentage  float64  `json:"percentage"`            // 0.0 to 100.0 of parent tool calls
	StepIndices []int    `json:"step_indices"`          // Timeline steps where this sub-command/file occurred
	RecentItems []string `json:"recent_items,omitempty"` // Top recent summaries or file paths
}

// SessionNativeTool represents a native agent tool execution (e.g. read_file, run_command).
// Implements: R-040, R-041
type SessionNativeTool struct {
	Name         string              `json:"name"`
	Category     string              `json:"category"` // Filesystem, Terminal, Subagent, Search, Browser, Orchestration
	InvokedCount int                 `json:"invoked_count"`
	StepIndices  []int               `json:"step_indices"`
	FirstStepIdx int                 `json:"first_step_idx"`
	LastStepIdx  int                 `json:"last_step_idx"`
	ErrorCount   int                 `json:"error_count"`
	SubBreakdown []ToolSubBreakdown  `json:"sub_breakdown,omitempty"` // Decomposed command/file breakdown
}

// SessionToolAnalytics provides analytical insights into agent tool usage.
// Implements: R-040, R-041
type SessionToolAnalytics struct {
	TotalToolCalls     int                `json:"total_tool_calls"`
	TotalMCPCalls      int                `json:"total_mcp_calls"`
	TotalNativeCalls   int                `json:"total_native_calls"`
	UniqueToolsUsed    int                `json:"unique_tools_used"`
	ToolDiversityIndex float64            `json:"tool_diversity_index"` // Unique tools / total calls
	TotalErrors        int                `json:"total_errors"`
	CommandBreakdown   []ToolSubBreakdown `json:"command_breakdown,omitempty"`
	FileBreakdown      []ToolSubBreakdown `json:"file_breakdown,omitempty"`
}

// SessionAgentContext represents active rules, skills, and tools in a conversation session.
// Implements: R-039, R-040
type SessionAgentContext struct {
	SessionID          string               `json:"session_id"`
	Workspace          string               `json:"workspace"`
	ProjectName        string               `json:"project_name"`
	ActiveRules        []AgentRule          `json:"active_rules"`
	ActiveSkills       []AgentSkill         `json:"active_skills"`    // Skills actually invoked during the session
	AvailableSkills    []AgentSkill         `json:"available_skills"` // Skills available via progressive disclosure
	MCPServers         []SessionMCPServer   `json:"mcp_servers"`
	NativeTools        []SessionNativeTool  `json:"native_tools"`
	ToolAnalytics      SessionToolAnalytics `json:"tool_analytics"`
	Subagents          []string             `json:"subagents,omitempty"`
	SlashCommands      []string             `json:"slash_commands,omitempty"`
	ActiveTokensEst    int                  `json:"active_tokens_est"`    // Tokens injected into session (rules + invoked skills)
	OnDemandTokensEst  int                  `json:"on_demand_tokens_est"`  // Tokens available if all on-demand skills were called
	TotalWordsInjected int                  `json:"total_words_injected"`
}

// GitFileStatus represents the git status and remote metadata of a file.
// Implements: R-053, R-060
type GitFileStatus struct {
	IsGit           bool      `json:"is_git"`
	GitRoot         string    `json:"git_root,omitempty"`
	RelativePath    string    `json:"relative_path,omitempty"`
	Branch          string    `json:"branch,omitempty"`
	IsDirty         bool      `json:"is_dirty"`
	HasStaged       bool      `json:"has_staged"`
	HasUnstaged     bool      `json:"has_unstaged"`
	RemoteURL       string    `json:"remote_url,omitempty"`
	GHAvailable     bool      `json:"gh_available"`
	GHAuthenticated bool      `json:"gh_authenticated"`
	ActivePR        *GitHubPR `json:"active_pr,omitempty"`
	SessionSteps    int       `json:"session_steps"`
}

// GitHubPR represents pull request metadata obtained from gh cli.
// Implements: R-060
type GitHubPR struct {
	Number         int    `json:"number"`
	Title          string `json:"title"`
	URL            string `json:"url"`
	State          string `json:"state"` // OPEN, CLOSED, MERGED
	BaseRef        string `json:"base_ref"`
	HeadRef        string `json:"head_ref"`
	ReviewDecision string `json:"review_decision,omitempty"` // APPROVED, CHANGES_REQUESTED
	StatusSummary  string `json:"status_summary,omitempty"`  // e.g. "12/12 passing"
}

// GitDiffResponse represents parsed diff hunks and statistics.
// Implements: R-054
type GitDiffResponse struct {
	Path      string     `json:"path"`
	Mode      string     `json:"mode"` // unstaged, staged, commit, pr, step
	RawDiff   string     `json:"raw_diff"`
	Additions int        `json:"additions"`
	Deletions int        `json:"deletions"`
	Hunks     []DiffHunk `json:"hunks"`
	Truncated bool       `json:"truncated"`
}

// DiffHunk represents an individual diff chunk with header and lines.
// Implements: R-054, R-058
type DiffHunk struct {
	Header   string     `json:"header"` // @@ -10,6 +10,8 @@
	OldStart int        `json:"old_start"`
	OldLines int        `json:"old_lines"`
	NewStart int        `json:"new_start"`
	NewLines int        `json:"new_lines"`
	Lines    []DiffLine `json:"lines"`
}

// DiffLine represents a single line in a diff.
// Implements: R-054, R-058
type DiffLine struct {
	Type       string `json:"type"` // "context", "add", "delete"
	Content    string `json:"content"`
	OldLineNum int    `json:"old_line_num,omitempty"`
	NewLineNum int    `json:"new_line_num,omitempty"`
}

// GitCommitEntry represents a single commit from git log.
// Implements: R-054
type GitCommitEntry struct {
	SHA          string `json:"sha"`
	ShortSHA     string `json:"short_sha"`
	AuthorName   string `json:"author_name"`
	AuthorEmail  string `json:"author_email"`
	Timestamp    int64  `json:"timestamp"`
	RelativeTime string `json:"relative_time"`
	Subject      string `json:"subject"`
}

// AgentStepDiff represents a synthetic diff hunk created from an LLM step.
// Implements: R-055
type AgentStepDiff struct {
	StepIndex    int        `json:"step_index"`
	ToolName     string     `json:"tool_name"`
	Timestamp    string     `json:"timestamp"`
	Thinking     string     `json:"thinking,omitempty"`
	UserPrompt   string     `json:"user_prompt,omitempty"`
	AgentSummary string     `json:"agent_summary,omitempty"`
	Hunks        []DiffHunk `json:"hunks"`
	Additions    int        `json:"additions"`
	Deletions    int        `json:"deletions"`
}

// OpenInIDERequest represents a request to launch a desktop IDE.
// Implements: R-077
type OpenInIDERequest struct {
	Editor string `json:"editor"` // "cursor", "vscode", "windsurf", "idea", "sublime", "system"
	Path   string `json:"path"`
	Line   int    `json:"line"`
}

// OpenInIDEResponse represents the result of an IDE launch attempt.
// Implements: R-077
type OpenInIDEResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	Editor      string `json:"editor,omitempty"`
	Command     string `json:"command,omitempty"`
	FallbackURI string `json:"fallback_uri,omitempty"`
}

// DetectedIDE represents an IDE discovered on the host system.
// Implements: R-077, R-079
type DetectedIDE struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	Installed bool   `json:"installed"`
	AppPath   string `json:"app_path,omitempty"`
	AppName   string `json:"app_name,omitempty"`
	CLIBinary string `json:"cli_binary,omitempty"`
	URIScheme string `json:"uri_scheme"`
	IsDefault bool   `json:"is_default"`
}

// DetectIDEsResponse represents the list of detected IDEs on the host.
// Implements: R-079
type DetectIDEsResponse struct {
	Detected  []DetectedIDE `json:"detected"`
	Preferred string        `json:"preferred"`
	OS        string        `json:"os"`
}




