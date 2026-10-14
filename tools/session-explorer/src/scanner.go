// Package main provides the session discovery and JSONL parsing engine.
//
// Implements: R-001 (Discover sessions), R-002 (Parse JSONL), R-003 (Extract user prompts),
//             R-004 (Extract agent responses), R-005 (Extract workspace context)
// See: specs/catalog/session-explorer.md
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// SessionIndex holds all discovered sessions and provides search capabilities.
type SessionIndex struct {
	mu            sync.RWMutex
	sessions      []Session
	// entryCache stores parsed transcript entries per session ID for detail views
	entryCache    map[string][]TranscriptEntry
	stats         Stats
	dataDir       string
	verbose       bool
	knownProjects map[string]string // project name -> absolute directory path
}

// NewSessionIndex creates a new session index for the given data directory.
func NewSessionIndex(dataDir string, verbose bool) *SessionIndex {
	idx := &SessionIndex{
		sessions:      make([]Session, 0),
		entryCache:    make(map[string][]TranscriptEntry),
		dataDir:       dataDir,
		verbose:       verbose,
		knownProjects: make(map[string]string),
	}
	idx.discoverKnownProjects()
	return idx
}

// discoverKnownProjects dynamically discovers repositories and projects across parent/known paths.
func (idx *SessionIndex) discoverKnownProjects() {
	searchDirs := []string{
		"/Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge",
		"/Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews",
	}

	if cwd, err := os.Getwd(); err == nil {
		parent := filepath.Dir(cwd)
		searchDirs = append(searchDirs, parent, filepath.Dir(parent))
	}

	skipDirs := map[string]bool{
		"brain": true, "Library": true, "CloudStorage": true, "OneDrive-Personal": true,
		"src": true, "tools": true, "specs": true, "docs": true, "scripts": true,
		"sdlc": true, "bin": true, "dist": true, "node_modules": true, "build": true,
		"_templates": true, "_internal": true, "knowledge": true, "Interviews": true,
	}

	for _, dir := range searchDirs {
		if strings.Contains(dir, "/tool-scripts/") || strings.HasSuffix(dir, "/tool-scripts") {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || skipDirs[name] || !isValidWorkspaceName(name) {
				continue
			}
			fullPath := filepath.Join(dir, name)
			if _, exists := idx.knownProjects[name]; !exists {
				idx.knownProjects[name] = fullPath
			}
		}
	}
}

// ScanAll discovers and indexes all sessions concurrently.
// Implements: R-001
func (idx *SessionIndex) ScanAll() error {
	start := time.Now()

	entries, err := os.ReadDir(idx.dataDir)
	if err != nil {
		return fmt.Errorf("failed to read data directory %s: %w", idx.dataDir, err)
	}

	var wg sync.WaitGroup
	sessionChan := make(chan Session, len(entries))
	errChan := make(chan error, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sessionID := entry.Name()
		transcriptPath := filepath.Join(idx.dataDir, sessionID, ".system_generated", "logs", "transcript.jsonl")

		// Check if transcript file exists
		if _, err := os.Stat(transcriptPath); os.IsNotExist(err) {
			if idx.verbose {
				log.Printf("[scanner] skipping %s: no transcript.jsonl", sessionID)
			}
			continue
		}

		wg.Add(1)
		go func(id, path string) {
			defer wg.Done()
			session, err := idx.scanSession(id, path)
			if err != nil {
				errChan <- fmt.Errorf("session %s: %w", id, err)
				return
			}
			sessionChan <- session
		}(sessionID, transcriptPath)
	}

	// Wait for all goroutines, then close channels
	go func() {
		wg.Wait()
		close(sessionChan)
		close(errChan)
	}()

	// Collect results into fresh slice for atomic swap
	newSessions := make([]Session, 0, len(sessionChan))
	for session := range sessionChan {
		newSessions = append(newSessions, session)
	}

	// Log errors but don't fail
	for err := range errChan {
		if idx.verbose {
			log.Printf("[scanner] warning: %v", err)
		}
	}

	// Sort sessions by creation date (newest first)
	sort.Slice(newSessions, func(i, j int) bool {
		return newSessions[i].CreatedAt.After(newSessions[j].CreatedAt)
	})

	// Atomically update index
	idx.mu.Lock()
	idx.sessions = newSessions
	idx.entryCache = make(map[string][]TranscriptEntry) // clear detail cache on rescan
	idx.computeStats()
	idx.stats.LastScannedAt = time.Now()
	idx.mu.Unlock()

	elapsed := time.Since(start)
	log.Printf("[scanner] indexed %d sessions in %v", len(newSessions), elapsed)

	return nil
}

// scanSession parses a single session's transcript file to extract metadata.
// Implements: R-002, R-003, R-004, R-005
func (idx *SessionIndex) scanSession(sessionID, transcriptPath string) (Session, error) {
	file, err := os.Open(transcriptPath)
	if err != nil {
		return Session{}, fmt.Errorf("failed to open %s: %w", transcriptPath, err)
	}
	defer file.Close()

	// Get file size
	info, err := file.Stat()
	if err != nil {
		return Session{}, fmt.Errorf("failed to stat %s: %w", transcriptPath, err)
	}

	session := Session{
		ID:                  sessionID,
		TranscriptSizeBytes: info.Size(),
	}

	// Check for full transcript
	fullPath := filepath.Join(filepath.Dir(transcriptPath), "transcript_full.jsonl")
	if _, err := os.Stat(fullPath); err == nil {
		session.HasFullTranscript = true
		if fi, err := os.Stat(fullPath); err == nil {
			session.TranscriptSizeBytes = fi.Size()
		}
	}

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large lines (some transcript lines can be very long)
	buf := make([]byte, 0, 1024*1024) // 1MB initial
	scanner.Buffer(buf, 10*1024*1024)  // 10MB max

	var firstTimestamp, lastTimestamp time.Time
	firstUserPromptFound := false
	toolCounts := make(map[string]int)
	projectScores := make(map[string]int)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry TranscriptEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			// Implements: NF-007 — gracefully handle malformed JSONL
			if idx.verbose {
				log.Printf("[scanner] malformed JSONL in %s line: %v", sessionID, err)
			}
			continue
		}

		session.StepCount++

		// Track timestamps
		if !entry.CreatedAt.IsZero() {
			if firstTimestamp.IsZero() || entry.CreatedAt.Before(firstTimestamp) {
				firstTimestamp = entry.CreatedAt
			}
			if entry.CreatedAt.After(lastTimestamp) {
				lastTimestamp = entry.CreatedAt
			}
		}

		// Count by type
		switch entry.Type {
		case "USER_INPUT":
			if entry.Source == "USER_EXPLICIT" {
				session.UserMessageCount++

				// Extract first user prompt preview
				if !firstUserPromptFound {
					session.FirstUserPromptPreview = extractUserPrompt(entry.Content)
					firstUserPromptFound = true
				}

				// Score repo/project mentions in prompt (filtering out OneDrive-Personal path noise)
				contentLower := strings.ToLower(entry.Content)
				contentForScoring := strings.ReplaceAll(contentLower, "onedrive-personal", "")
				for name := range idx.knownProjects {
					if strings.Contains(contentForScoring, strings.ToLower(name)) {
						projectScores[name] += 10
					}
				}

				// Extract workspace from metadata
				if session.Workspace == "" || strings.HasPrefix(session.Workspace, "/Untitled") {
					extracted := extractWorkspace(entry.Content)
					if extracted != "" {
						session.Workspace = extracted
						pName := extractProjectName(extracted)
						if pName != "" && pName != "Default" {
							projectScores[pName] += 15
						}
					}
				}
			}

		case "PLANNER_RESPONSE":
			session.AgentResponseCount++
			if entry.ToolCalls != nil {
				session.ToolCallCount += len(entry.ToolCalls)
				for _, tc := range entry.ToolCalls {
					toolCounts[tc.Name]++

					for _, key := range []string{"Cwd", "DirectoryPath", "SearchPath", "TargetFile", "AbsolutePath", "TargetDirectory"} {
						if rawVal, exists := tc.Args[key]; exists {
							if valStr, ok := rawVal.(string); ok && valStr != "" {
								valStr = strings.Trim(valStr, `"'`)
								if (strings.HasPrefix(valStr, "/") || strings.HasPrefix(valStr, "~")) &&
									!strings.Contains(valStr, ".gemini/antigravity-ide/brain") &&
									!strings.Contains(valStr, "/.gemini/") &&
									!strings.HasPrefix(valStr, "/Untitled") {

									matched := false
									// Check against known projects (ignoring OneDrive-Personal path segment for Personal project)
									for name, path := range idx.knownProjects {
										if !isValidWorkspaceName(name) {
											continue
										}
										if name == "Personal" && strings.Contains(valStr, "OneDrive-Personal") && !strings.Contains(valStr, "/knowledge/Personal") {
											continue
										}
										if strings.Contains(valStr, "/"+name+"/") || strings.HasSuffix(valStr, "/"+name) || strings.Contains(valStr, "/knowledge/"+name) {
											projectScores[name] += 20
											matched = true
										} else if path != "" && strings.Contains(valStr, path) {
											projectScores[name] += 20
											matched = true
										}
									}

									// Extract repo name directly from /knowledge/<repo> or /Interviews/<repo>
									if !matched {
										if kIdx := strings.Index(valStr, "/knowledge/"); kIdx != -1 {
											sub := valStr[kIdx+len("/knowledge/"):]
											parts := strings.Split(sub, "/")
											if len(parts) > 0 && isValidWorkspaceName(parts[0]) {
												projectScores[parts[0]] += 25
												matched = true
											}
										} else if iIdx := strings.Index(valStr, "/Interviews/"); iIdx != -1 {
											sub := valStr[iIdx+len("/Interviews/"):]
											parts := strings.Split(sub, "/")
											if len(parts) > 0 && isValidWorkspaceName(parts[0]) {
												projectScores[parts[0]] += 25
												matched = true
											}
										}
									}

									// Fallback: derive project name directly from path if no known project matched
									if !matched {
										derived := extractProjectName(valStr)
										if isValidWorkspaceName(derived) {
											projectScores[derived] += 15
											if session.Workspace == "" || strings.HasPrefix(session.Workspace, "/Untitled") {
												session.Workspace = filepath.Dir(valStr)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}

		// Count errors
		if entry.Status == "ERROR" || entry.Type == "ERROR_MESSAGE" {
			session.ErrorCount++
		}
	}

	if err := scanner.Err(); err != nil {
		return Session{}, fmt.Errorf("scanner error for %s: %w", sessionID, err)
	}

	session.CreatedAt = firstTimestamp
	session.LastMessageAt = lastTimestamp
	session.ToolCallsMap = toolCounts

	// Determine best project from projectScores
	var bestProject string
	maxScore := 0
	type pScore struct {
		name  string
		score int
	}
	var scoredList []pScore
	for name, score := range projectScores {
		if score > 0 && isValidWorkspaceName(name) {
			scoredList = append(scoredList, pScore{name: name, score: score})
			if score > maxScore {
				maxScore = score
				bestProject = name
			}
		}
	}

	sort.Slice(scoredList, func(i, j int) bool {
		return scoredList[i].score > scoredList[j].score
	})

	session.ProjectsTouched = make([]string, 0, len(scoredList))
	seenTouched := make(map[string]bool)
	for _, ps := range scoredList {
		if isValidWorkspaceName(ps.name) && !seenTouched[ps.name] {
			seenTouched[ps.name] = true
			session.ProjectsTouched = append(session.ProjectsTouched, ps.name)
		}
	}

	if bestProject != "" && isValidWorkspaceName(bestProject) {
		session.ProjectName = bestProject
		if p, ok := idx.knownProjects[bestProject]; ok {
			session.Workspace = p
		} else if session.Workspace == "" || strings.HasPrefix(session.Workspace, "/Untitled") {
			knowledgePath := filepath.Join("/Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge", bestProject)
			if _, err := os.Stat(knowledgePath); err == nil {
				session.Workspace = knowledgePath
			} else {
				session.Workspace = "/" + bestProject
			}
		}
	} else if session.Workspace != "" {
		derived := extractProjectName(session.Workspace)
		if isValidWorkspaceName(derived) {
			session.ProjectName = derived
			session.ProjectsTouched = []string{derived}
		} else {
			session.ProjectName = "Default"
		}
	} else {
		session.ProjectName = "Default"
	}

	return session, nil
}

// GetSessions returns all sessions, optionally filtered.
// Implements: R-010, R-011, R-014
func (idx *SessionIndex) GetSessions(workspace, sortBy, sortOrder string, from, to time.Time) []Session {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	filtered := make([]Session, 0, len(idx.sessions))
	for _, s := range idx.sessions {
		// Filter by workspace
		if workspace != "" {
			wsLower := strings.ToLower(workspace)
			matched := strings.Contains(strings.ToLower(s.Workspace), wsLower) ||
				strings.Contains(strings.ToLower(s.ProjectName), wsLower)
			if !matched {
				for _, pt := range s.ProjectsTouched {
					if strings.Contains(strings.ToLower(pt), wsLower) {
						matched = true
						break
					}
				}
			}
			if !matched {
				continue
			}
		}

		// Filter by date range (inclusive session lifetime overlap)
		lastActive := s.LastMessageAt
		if lastActive.IsZero() {
			lastActive = s.CreatedAt
		}
		if !from.IsZero() && lastActive.Before(from) {
			continue
		}
		if !to.IsZero() && s.CreatedAt.After(to) {
			continue
		}

		filtered = append(filtered, s)
	}

	// Sort
	switch sortBy {
	case "date":
		if sortOrder == "asc" {
			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
			})
		} else {
			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
			})
		}
	case "date_asc":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
		})
	case "steps":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].StepCount > filtered[j].StepCount
		})
	case "size":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].TranscriptSizeBytes > filtered[j].TranscriptSizeBytes
		})
	case "messages":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].UserMessageCount > filtered[j].UserMessageCount
		})
	}

	return filtered
}

// GetSessionDetail loads the full conversation for a session.
// Implements: R-009
func (idx *SessionIndex) GetSessionDetail(sessionID string) (*SessionDetailResponse, error) {
	// Find the session
	idx.mu.RLock()
	var session *Session
	for _, s := range idx.sessions {
		if s.ID == sessionID {
			sCopy := s
			session = &sCopy
			break
		}
	}
	idx.mu.RUnlock()

	if session == nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Prefer full transcript if available
	transcriptPath := filepath.Join(idx.dataDir, sessionID, ".system_generated", "logs", "transcript.jsonl")
	if session.HasFullTranscript {
		fullPath := filepath.Join(idx.dataDir, sessionID, ".system_generated", "logs", "transcript_full.jsonl")
		if _, err := os.Stat(fullPath); err == nil {
			transcriptPath = fullPath
		}
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open transcript: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	messages := make([]MessageDisplay, 0)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry TranscriptEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		// Only include meaningful message types in the UI
		switch entry.Type {
		case "USER_INPUT", "PLANNER_RESPONSE", "ERROR_MESSAGE":
			// Always include these
		case "VIEW_FILE", "RUN_COMMAND", "GREP_SEARCH", "LIST_DIRECTORY",
			"CODE_ACTION", "SEARCH_WEB", "READ_URL_CONTENT", "BROWSER_SUBAGENT",
			"ASK_QUESTION":
			// Include tool results
		default:
			// Skip system messages, checkpoints, ephemeral, etc.
			continue
		}

		msg := MessageDisplay{
			StepIndex:   entry.StepIndex,
			Source:      entry.Source,
			Type:        entry.Type,
			Status:      entry.Status,
			CreatedAt:   entry.CreatedAt,
			HasThinking: entry.Thinking != "",
			Thinking:    entry.Thinking,
		}

		// Clean user input content (strip XML tags)
		if entry.Type == "USER_INPUT" && entry.Source == "USER_EXPLICIT" {
			msg.Content = extractUserPrompt(entry.Content)
		} else {
			msg.Content = entry.Content
		}

		// Convert tool calls
		if entry.ToolCalls != nil {
			for _, tc := range entry.ToolCalls {
				argsJSON, _ := json.Marshal(tc.Args)
				preview := string(argsJSON)
				if len(preview) > 500 {
					preview = preview[:500] + "..."
				}
				msg.ToolCalls = append(msg.ToolCalls, ToolCallDisplay{
					Name:        tc.Name,
					ArgsPreview: preview,
					Args:        tc.Args,
				})
			}
		}

		messages = append(messages, msg)
	}

	return &SessionDetailResponse{
		Session:  *session,
		Messages: messages,
	}, nil
}

// Search performs full-text search across all sessions.
// Implements: R-012
func (idx *SessionIndex) Search(query, workspace, msgType string, from, to time.Time) *SearchResponse {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if query == "" {
		return &SearchResponse{Results: []SearchResult{}, TotalCount: 0, Query: query}
	}

	queryLower := strings.ToLower(query)
	results := make([]SearchResult, 0)

	for _, session := range idx.sessions {
		// Apply workspace filter
		if workspace != "" {
			wsLower := strings.ToLower(workspace)
			matched := strings.Contains(strings.ToLower(session.Workspace), wsLower) ||
				strings.Contains(strings.ToLower(session.ProjectName), wsLower)
			if !matched {
				for _, pt := range session.ProjectsTouched {
					if strings.Contains(strings.ToLower(pt), wsLower) {
						matched = true
						break
					}
				}
			}
			if !matched {
				continue
			}
		}

		// Apply date filter (inclusive lifetime overlap)
		lastActive := session.LastMessageAt
		if lastActive.IsZero() {
			lastActive = session.CreatedAt
		}
		if !from.IsZero() && lastActive.Before(from) {
			continue
		}
		if !to.IsZero() && session.CreatedAt.After(to) {
			continue
		}

		// Scan transcript for matches
		transcriptPath := filepath.Join(idx.dataDir, session.ID, ".system_generated", "logs", "transcript.jsonl")
		file, err := os.Open(transcriptPath)
		if err != nil {
			continue
		}

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

			// Apply type filter
			if msgType != "" && entry.Type != msgType {
				continue
			}

			// Search in content
			content := entry.Content
			if entry.Type == "USER_INPUT" && entry.Source == "USER_EXPLICIT" {
				content = extractUserPrompt(content)
			}

			if content == "" {
				continue
			}

			if strings.Contains(strings.ToLower(content), queryLower) {
				// Create preview around the match
				preview := createSearchPreview(content, queryLower, 200)

				results = append(results, SearchResult{
					SessionID:      session.ID,
					StepIndex:      entry.StepIndex,
					Type:           entry.Type,
					Source:         entry.Source,
					ContentPreview: preview,
					CreatedAt:      entry.CreatedAt,
					Workspace:      session.Workspace,
				})
			}
		}

		file.Close()
	}

	// Sort results by date (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	// Limit results
	maxResults := 100
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return &SearchResponse{
		Results:    results,
		TotalCount: len(results),
		Query:      query,
	}
}

// GetStats returns aggregate statistics.
// Implements: R-017
func (idx *SessionIndex) GetStats() Stats {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.stats
}

// computeStats calculates aggregate statistics from all sessions.
func (idx *SessionIndex) computeStats() {
	// CRITICAL: Reinitialize stats on every computation to prevent cumulative leakage across rescans
	idx.stats = Stats{
		Workspaces:     make([]string, 0),
		Projects:       make([]ProjectSummary, 0),
		TopToolCalls:   make([]ToolFreq, 0),
		TopPrompts:     make([]PromptFreq, 0),
		MessagesByType: make(map[string]int),
	}

	workspaceSet := make(map[string]bool)
	toolCallFreq := make(map[string]int)
	promptFreqMap := make(map[string]int)
	projectStatsMap := make(map[string]*ProjectSummary)

	var minDate, maxDate time.Time

	for _, s := range idx.sessions {
		idx.stats.TotalSessions++
		idx.stats.TotalUserMessages += s.UserMessageCount
		idx.stats.TotalAgentResponses += s.AgentResponseCount
		idx.stats.TotalToolCalls += s.ToolCallCount
		idx.stats.TotalErrors += s.ErrorCount

		proj := s.ProjectName
		if isValidWorkspaceName(proj) {
			workspaceSet[proj] = true

			ps, exists := projectStatsMap[proj]
			if !exists {
				ps = &ProjectSummary{
					Name:          proj,
					Path:          s.Workspace,
					RecentPrompts: make([]string, 0),
				}
				if _, err := os.Stat(filepath.Join(s.Workspace, ".git")); err == nil {
					ps.IsGit = true
				}
				projectStatsMap[proj] = ps
			}
			ps.SessionCount++
			ps.TotalSteps += s.StepCount
			ps.TotalUserMessages += s.UserMessageCount
			ps.TotalAgentResponses += s.AgentResponseCount
			ps.TotalToolCalls += s.ToolCallCount
			ps.TotalErrors += s.ErrorCount

			if s.LastMessageAt.After(ps.LastActiveAt) {
				ps.LastActiveAt = s.LastMessageAt
			} else if ps.LastActiveAt.IsZero() || s.CreatedAt.After(ps.LastActiveAt) {
				ps.LastActiveAt = s.CreatedAt
			}

			if s.FirstUserPromptPreview != "" && len(ps.RecentPrompts) < 3 {
				dup := false
				for _, rp := range ps.RecentPrompts {
					if rp == s.FirstUserPromptPreview {
						dup = true
						break
					}
				}
				if !dup {
					ps.RecentPrompts = append(ps.RecentPrompts, s.FirstUserPromptPreview)
				}
			}
		}

		// Record any touched projects in workspaceSet
		for _, pt := range s.ProjectsTouched {
			if isValidWorkspaceName(pt) {
				workspaceSet[pt] = true
			}
		}

		// Aggregate tool call counts from session
		for toolName, count := range s.ToolCallsMap {
			toolCallFreq[toolName] += count
		}

		// Aggregate prompt frequency
		p := strings.TrimSpace(s.FirstUserPromptPreview)
		if p != "" && len(p) > 3 {
			if len(p) > 120 {
				p = p[:120] + "..."
			}
			promptFreqMap[p]++
		}

		if !s.CreatedAt.IsZero() {
			if minDate.IsZero() || s.CreatedAt.Before(minDate) {
				minDate = s.CreatedAt
			}
			if s.CreatedAt.After(maxDate) {
				maxDate = s.CreatedAt
			}
		}
	}

	// Also add any known projects from knowledge directory even if 0 sessions
	for name, path := range idx.knownProjects {
		if isValidWorkspaceName(name) {
			if _, exists := projectStatsMap[name]; !exists {
				isGit := false
				if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
					isGit = true
				}
				projectStatsMap[name] = &ProjectSummary{
					Name:          name,
					Path:          path,
					IsGit:         isGit,
					RecentPrompts: []string{},
				}
			}
			workspaceSet[name] = true
		}
	}

	// Collect projects into sorted slice
	var projects []ProjectSummary
	for _, ps := range projectStatsMap {
		projects = append(projects, *ps)
	}
	sort.Slice(projects, func(i, j int) bool {
		if (projects[i].SessionCount > 0) != (projects[j].SessionCount > 0) {
			return projects[i].SessionCount > projects[j].SessionCount
		}
		if !projects[i].LastActiveAt.Equal(projects[j].LastActiveAt) {
			return projects[i].LastActiveAt.After(projects[j].LastActiveAt)
		}
		return projects[i].Name < projects[j].Name
	})
	idx.stats.Projects = projects

	// Collect strictly unique, valid workspaces
	for ws := range workspaceSet {
		if isValidWorkspaceName(ws) {
			idx.stats.Workspaces = append(idx.stats.Workspaces, ws)
		}
	}
	sort.Strings(idx.stats.Workspaces)

	idx.stats.DateRange = DateRange{From: minDate, To: maxDate}

	// Sort tool calls by frequency
	for name, count := range toolCallFreq {
		idx.stats.TopToolCalls = append(idx.stats.TopToolCalls, ToolFreq{Name: name, Count: count})
	}
	sort.Slice(idx.stats.TopToolCalls, func(i, j int) bool {
		return idx.stats.TopToolCalls[i].Count > idx.stats.TopToolCalls[j].Count
	})
	if len(idx.stats.TopToolCalls) > 12 {
		idx.stats.TopToolCalls = idx.stats.TopToolCalls[:12]
	}

	// Sort prompts by frequency
	for prompt, count := range promptFreqMap {
		idx.stats.TopPrompts = append(idx.stats.TopPrompts, PromptFreq{Prompt: prompt, Count: count})
	}
	sort.Slice(idx.stats.TopPrompts, func(i, j int) bool {
		return idx.stats.TopPrompts[i].Count > idx.stats.TopPrompts[j].Count
	})
	if len(idx.stats.TopPrompts) > 10 {
		idx.stats.TopPrompts = idx.stats.TopPrompts[:10]
	}
}

// --- Helper functions ---

// extractUserPrompt extracts clean user text from the raw content (strips XML metadata).
// Implements: R-003
var userRequestRegex = regexp.MustCompile(`(?s)<USER_REQUEST>\s*(.*?)\s*</USER_REQUEST>`)

func extractUserPrompt(content string) string {
	matches := userRequestRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		prompt := strings.TrimSpace(matches[1])
		return prompt
	}
	// If no XML tags, return the content as-is (truncated)
	content = strings.TrimSpace(content)
	if len(content) > 2000 {
		content = content[:2000] + "..."
	}
	return content
}

// extractWorkspace extracts the workspace/project path from ADDITIONAL_METADATA.
// Implements: R-005
var activeDocRegex = regexp.MustCompile(`Active Document:\s*(/[^\s\n\r]+)`)
var workspaceMappingRegex = regexp.MustCompile(`(?m)(/[^\s\n\r]+)\s*->\s*([^\s\n\r]+)`)

func extractWorkspace(content string) string {
	// 1. Check workspace mapping: /path/to/project -> corpus
	matches := workspaceMappingRegex.FindStringSubmatch(content)
	if len(matches) > 1 && !strings.Contains(matches[1], ".gemini/antigravity-ide/brain") {
		return matches[1]
	}

	// 2. Check active document
	matches = activeDocRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		doc := matches[1]
		if !strings.HasPrefix(doc, "/Untitled") {
			return filepath.Dir(doc)
		}
	}
	return ""
}

// isValidWorkspaceName checks if a candidate string is a genuine repository or workspace name.
func isValidWorkspaceName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" || name == "Default" || name == "workspace" || name == "aparv" {
		return false
	}
	// No hidden files/folders, templates, or private paths
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return false
	}
	// Workspaces cannot be files (e.g. all_mermaids.txt) or paths containing slashes
	if strings.Contains(name, ".") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return false
	}
	// Lowercase check against system folders, repository subdirectories, and common keywords
	lower := strings.ToLower(name)
	blocked := map[string]bool{
		"src": true, "tools": true, "specs": true, "sdlc": true, "docs": true,
		"scripts": true, "rules": true, "analytics": true, "tests": true,
		"examples": true, "internal": true, "pkg": true, "node_modules": true,
		"dist": true, "build": true, "target": true, "config": true,
		"builtin": true, "skills": true, "plugins": true, "hooks": true,
		"book": true, "guides": true, "architecture": true, "scratch": true,
		"logs": true, "brain": true, "users": true, "home": true, "tmp": true,
		"var": true, "private": true, "knowledge": true, "interviews": true,
		"cloudstorage": true, "onedrive-personal": true, "g-drive": true,
		"default": true, "workspace": true, "aparv": true, "runner": true,
		"work": true, "bin": true, "cmd": true, "lib": true,
		"library": true, "applications": true, "system": true, "volumes": true,
	}
	if blocked[lower] {
		return false
	}
	return len(name) >= 2
}

// extractProjectName derives a clean, human-readable project or repository name from a workspace or file path.
func extractProjectName(workspace string) string {
	if workspace == "" || workspace == "/Untitled-1" || strings.HasPrefix(workspace, "/Untitled") {
		return "Default"
	}

	cleaned := filepath.Clean(strings.Trim(workspace, `"'`))

	// 1. Direct match for knowledge or Interviews parent paths: /knowledge/<repo> or /Interviews/<repo>
	if kIdx := strings.Index(cleaned, "/knowledge/"); kIdx != -1 {
		sub := cleaned[kIdx+len("/knowledge/"):]
		parts := strings.Split(sub, "/")
		if len(parts) > 0 && isValidWorkspaceName(parts[0]) {
			return parts[0]
		}
	}
	if iIdx := strings.Index(cleaned, "/Interviews/"); iIdx != -1 {
		sub := cleaned[iIdx+len("/Interviews/"):]
		parts := strings.Split(sub, "/")
		if len(parts) > 0 && isValidWorkspaceName(parts[0]) {
			return parts[0]
		}
	}

	// 2. If it contains tool-scripts or starts with /workspace, return tool-scripts
	if strings.Contains(cleaned, "tool-scripts") || strings.HasPrefix(cleaned, "/workspace") {
		return "tool-scripts"
	}

	// 3. Scan path components from right to left, skipping files and blocked directories
	parts := strings.Split(cleaned, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if part == "" {
			continue
		}
		// If last segment has an extension or dot, it's a file - skip it
		if i == len(parts)-1 && strings.Contains(part, ".") {
			continue
		}
		if isValidWorkspaceName(part) {
			return part
		}
	}

	return "Default"
}

// createSearchPreview creates a preview snippet around a search match.
func createSearchPreview(content, queryLower string, maxLen int) string {
	contentLower := strings.ToLower(content)
	idx := strings.Index(contentLower, queryLower)
	if idx == -1 {
		if len(content) > maxLen {
			return content[:maxLen] + "..."
		}
		return content
	}

	// Center the preview around the match
	start := idx - maxLen/2
	if start < 0 {
		start = 0
	}
	end := start + maxLen
	if end > len(content) {
		end = len(content)
	}

	preview := content[start:end]
	if start > 0 {
		preview = "..." + preview
	}
	if end < len(content) {
		preview = preview + "..."
	}

	return preview
}

// formatRelativeTime converts a duration into a human-friendly string.
func formatRelativeTime(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	secs := int(d.Seconds())
	if secs < 60 {
		return "just now"
	}
	mins := secs / 60
	if mins < 60 {
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	}
	hours := mins / 60
	if hours < 24 {
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := hours / 24
	if days == 1 {
		return "yesterday"
	}
	return fmt.Sprintf("%d days ago", days)
}

// GetWorkspaceWIP resolves the recent work-in-progress state for a workspace or session ID.
// Implements: R-030 (Agent WIP Recovery API), R-031 (Multi-Session Disambiguation Invariant)
func (idx *SessionIndex) GetWorkspaceWIP(workspace, sessionID string, forceLatest bool, limit int) (*WIPResponse, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if limit <= 0 {
		limit = 5
	}
	if limit > 50 {
		limit = 50
	}

	// 1. Direct Session ID lookup
	if sessionID != "" {
		for _, s := range idx.sessions {
			if s.ID == sessionID {
				digest, err := idx.buildWIPDigest(&s)
				if err != nil {
					return nil, err
				}
				return &WIPResponse{
					Status:         "resolved",
					Multiple:       false,
					Workspace:      s.Workspace,
					CandidateCount: 1,
					ActiveSession:  digest,
				}, nil
			}
		}
		return &WIPResponse{
			Status:         "not_found",
			Multiple:       false,
			Workspace:      workspace,
			CandidateCount: 0,
			Instruction:    fmt.Sprintf("Session ID %q was not found.", sessionID),
		}, nil
	}

	// 2. Filter matching sessions by workspace
	matches := make([]Session, 0)
	wsTarget := strings.TrimSpace(workspace)
	wsTargetClean := extractProjectName(wsTarget)

	for _, s := range idx.sessions {
		if wsTarget == "" {
			matches = append(matches, s)
			continue
		}

		wsLower := strings.ToLower(wsTarget)
		cleanLower := strings.ToLower(wsTargetClean)

		matched := strings.EqualFold(s.ProjectName, wsTarget) ||
			strings.EqualFold(s.ProjectName, wsTargetClean) ||
			strings.Contains(strings.ToLower(s.Workspace), wsLower) ||
			strings.Contains(strings.ToLower(s.ProjectName), wsLower) ||
			(cleanLower != "default" && strings.EqualFold(s.ProjectName, cleanLower))

		if !matched {
			for _, pt := range s.ProjectsTouched {
				if strings.EqualFold(pt, wsTarget) || strings.EqualFold(pt, wsTargetClean) ||
					strings.Contains(strings.ToLower(pt), wsLower) {
					matched = true
					break
				}
			}
		}

		if matched {
			matches = append(matches, s)
		}
	}

	if len(matches) == 0 {
		return &WIPResponse{
			Status:         "not_found",
			Multiple:       false,
			Workspace:      workspace,
			CandidateCount: 0,
			Instruction:    fmt.Sprintf("No active sessions found for workspace %q.", workspace),
		}, nil
	}

	// Sort matching sessions by LastMessageAt (descending)
	sort.Slice(matches, func(i, j int) bool {
		ti := matches[i].LastMessageAt
		if ti.IsZero() {
			ti = matches[i].CreatedAt
		}
		tj := matches[j].LastMessageAt
		if tj.IsZero() {
			tj = matches[j].CreatedAt
		}
		return ti.After(tj)
	})

	now := time.Now()

	// 3. Multi-Session Disambiguation Invariant (R-031):
	// If multiple sessions exist and caller did not explicitly force the latest,
	// agent MUST STOP and ask the user to disambiguate.
	if len(matches) > 1 && !forceLatest {
		candCount := len(matches)
		if candCount > limit {
			candCount = limit
		}

		candidates := make([]SessionCandidate, 0, candCount)
		for i := 0; i < candCount; i++ {
			s := matches[i]
			lastAt := s.LastMessageAt
			if lastAt.IsZero() {
				lastAt = s.CreatedAt
			}

			// Get last prompt preview
			lastPrompt := s.FirstUserPromptPreview
			entries, ok := idx.entryCache[s.ID]
			if ok {
				for j := len(entries) - 1; j >= 0; j-- {
					if entries[j].Type == "USER_INPUT" && entries[j].Source == "USER_EXPLICIT" {
						clean := extractUserPrompt(entries[j].Content)
						if clean != "" {
							if len(clean) > 80 {
								lastPrompt = clean[:80] + "..."
							} else {
								lastPrompt = clean
							}
							break
						}
					}
				}
			}

			candidates = append(candidates, SessionCandidate{
				SessionID:          s.ID,
				CreatedAt:          s.CreatedAt,
				LastMessageAt:      lastAt,
				RelativeTime:       formatRelativeTime(now.Sub(lastAt)),
				Workspace:          s.Workspace,
				ProjectName:        s.ProjectName,
				StepCount:          s.StepCount,
				UserMessageCount:   s.UserMessageCount,
				ErrorCount:         s.ErrorCount,
				FirstPromptPreview: s.FirstUserPromptPreview,
				LastPromptPreview:  lastPrompt,
			})
		}

		return &WIPResponse{
			Status:         "ambiguous_multi_session",
			Multiple:       true,
			Workspace:      workspace,
			CandidateCount: len(matches),
			Candidates:     candidates,
			Instruction:    "MANDATORY: Multiple sessions found for this workspace. You MUST STOP immediately and ask the user which session they wish to resume. Present the options with their timestamps and prompt previews.",
		}, nil
	}

	// 4. Single session resolved (or forced latest)
	target := matches[0]
	digest, err := idx.buildWIPDigest(&target)
	if err != nil {
		return nil, err
	}

	return &WIPResponse{
		Status:         "resolved",
		Multiple:       false,
		Workspace:      target.Workspace,
		CandidateCount: len(matches),
		ActiveSession:  digest,
	}, nil
}

// buildWIPDigest constructs the detailed work-in-progress summary for a single session.
func (idx *SessionIndex) buildWIPDigest(session *Session) (*WIPDigest, error) {
	// Prefer full transcript if available
	transcriptPath := filepath.Join(idx.dataDir, session.ID, ".system_generated", "logs", "transcript.jsonl")
	if session.HasFullTranscript {
		fullPath := filepath.Join(idx.dataDir, session.ID, ".system_generated", "logs", "transcript_full.jsonl")
		if _, err := os.Stat(fullPath); err == nil {
			transcriptPath = fullPath
		}
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open transcript: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	entries := make([]TranscriptEntry, 0)
	for scanner.Scan() {
		var entry TranscriptEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	lastAt := session.LastMessageAt
	if lastAt.IsZero() {
		lastAt = session.CreatedAt
	}

	digest := &WIPDigest{
		SessionID:        session.ID,
		Workspace:        session.Workspace,
		ProjectName:      session.ProjectName,
		CreatedAt:        session.CreatedAt,
		LastMessageAt:    lastAt,
		RelativeTime:     formatRelativeTime(time.Since(lastAt)),
		StepCount:        session.StepCount,
		UserMessageCount: session.UserMessageCount,
		ErrorCount:       session.ErrorCount,
		RecentPrompts:    make([]string, 0),
		RecentResponses:  make([]string, 0),
		ModifiedFiles:    make([]string, 0),
		ExecutedCommands: make([]string, 0),
		LastToolCalls:    make([]ToolCallDisplay, 0),
		UnresolvedErrors: make([]string, 0),
	}

	modifiedFilesSet := make(map[string]bool)
	commandsSet := make(map[string]bool)

	// Collect from entries in reverse chronological order
	promptCount := 0
	responseCount := 0

	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]

		// 1. Recent user prompts
		if entry.Type == "USER_INPUT" && entry.Source == "USER_EXPLICIT" && promptCount < 3 {
			clean := extractUserPrompt(entry.Content)
			if clean != "" {
				digest.RecentPrompts = append(digest.RecentPrompts, clean)
				promptCount++
			}
		}

		// 2. Recent planner responses
		if entry.Type == "PLANNER_RESPONSE" && responseCount < 3 {
			cleanResp := strings.TrimSpace(entry.Content)
			if cleanResp != "" {
				// Strip overly verbose outputs if needed, keeping up to 1000 characters
				if len(cleanResp) > 1000 {
					cleanResp = cleanResp[:1000] + "..."
				}
				digest.RecentResponses = append(digest.RecentResponses, cleanResp)
				responseCount++
			}
		}

		// 3. Tool calls inspection (modified files and shell commands)
		for _, tc := range entry.ToolCalls {
			if len(digest.LastToolCalls) < 5 {
				argsJSON, _ := json.Marshal(tc.Args)
				preview := string(argsJSON)
				if len(preview) > 120 {
					preview = preview[:120] + "..."
				}
				digest.LastToolCalls = append(digest.LastToolCalls, ToolCallDisplay{
					Name:        tc.Name,
					ArgsPreview: preview,
					Args:        tc.Args,
				})
			}

			// File modifications
			if tc.Name == "write_to_file" || tc.Name == "replace_file_content" || tc.Name == "multi_replace_file_content" {
				if tf, ok := tc.Args["TargetFile"].(string); ok && tf != "" {
					if !modifiedFilesSet[tf] {
						modifiedFilesSet[tf] = true
						digest.ModifiedFiles = append(digest.ModifiedFiles, tf)
					}
				}
			}

			// Executed commands
			if tc.Name == "run_command" {
				if cmd, ok := tc.Args["CommandLine"].(string); ok && cmd != "" {
					if !commandsSet[cmd] {
						commandsSet[cmd] = true
						digest.ExecutedCommands = append(digest.ExecutedCommands, cmd)
					}
				}
			}
		}

		// 4. Unresolved errors in recent steps
		if (entry.Status == "ERROR" || entry.Type == "ERROR_MESSAGE") && len(digest.UnresolvedErrors) < 3 {
			errMsg := entry.Content
			if len(errMsg) > 250 {
				errMsg = errMsg[:250] + "..."
			}
			digest.UnresolvedErrors = append(digest.UnresolvedErrors, errMsg)
		}
	}

	// Reverse recent prompts and responses so they are in forward chronological order
	for i, j := 0, len(digest.RecentPrompts)-1; i < j; i, j = i+1, j-1 {
		digest.RecentPrompts[i], digest.RecentPrompts[j] = digest.RecentPrompts[j], digest.RecentPrompts[i]
	}
	for i, j := 0, len(digest.RecentResponses)-1; i < j; i, j = i+1, j-1 {
		digest.RecentResponses[i], digest.RecentResponses[j] = digest.RecentResponses[j], digest.RecentResponses[i]
	}

	// Construct concise suggested continuation
	if len(digest.RecentPrompts) > 0 {
		lastPrompt := digest.RecentPrompts[len(digest.RecentPrompts)-1]
		if len(lastPrompt) > 100 {
			lastPrompt = lastPrompt[:100] + "..."
		}
		digest.SuggestedContinuation = fmt.Sprintf("Last prompt was: %q. %d files modified.", lastPrompt, len(digest.ModifiedFiles))
	} else {
		digest.SuggestedContinuation = "Ready to resume work on active workspace."
	}

	return digest, nil
}

