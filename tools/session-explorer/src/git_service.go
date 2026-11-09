// Package main provides Git and GitHub integration services for the Session Explorer.
//
// Implements: R-053 (Native Git Trackability Discovery), R-054 (Git Diff & Commits API),
// R-060 (Ambient Remote GitHub PR & Checks Integration)
// See: specs/catalog/session-explorer-github-review-pane.md

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// GitService provides deterministic Git and GitHub operations on local repositories.
type GitService struct {
	index *SessionIndex
}

// NewGitService creates a new GitService instance.
func NewGitService(idx *SessionIndex) *GitService {
	return &GitService{index: idx}
}

// CleanPath resolves and sanitizes a file path for git operations.
func (s *GitService) CleanPath(rawPath string) string {
	if strings.TrimSpace(rawPath) == "" {
		return ""
	}

	cleanPath := strings.TrimPrefix(rawPath, "file://")

	// Strip trailing line anchors (#L10-L20)
	if idx := strings.Index(cleanPath, "#"); idx != -1 {
		cleanPath = cleanPath[:idx]
	}
	// Strip trailing :line if file does not exist directly
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		if lastColon := strings.LastIndex(cleanPath, ":"); lastColon != -1 {
			candidate := cleanPath[:lastColon]
			if _, err2 := os.Stat(candidate); err2 == nil {
				cleanPath = candidate
			}
		}
	}

	cleanPath = filepath.Clean(cleanPath)

	// Expand ~ to user home directory
	if strings.HasPrefix(cleanPath, "~/") || cleanPath == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			cleanPath = filepath.Join(home, strings.TrimPrefix(cleanPath, "~"))
		}
	}

	// Resolve relative path against working directory or known project workspaces
	if !filepath.IsAbs(cleanPath) {
		if _, err := os.Stat(cleanPath); os.IsNotExist(err) && s.index != nil {
			s.index.mu.RLock()
			for _, projPath := range s.index.knownProjects {
				cand := filepath.Join(projPath, cleanPath)
				if _, err2 := os.Stat(cand); err2 == nil {
					cleanPath = cand
					break
				}
			}
			s.index.mu.RUnlock()
		}
		if !filepath.IsAbs(cleanPath) {
			if abs, err := filepath.Abs(cleanPath); err == nil {
				cleanPath = abs
			}
		}
	}

	return cleanPath
}

// runCommand runs a command with a strict timeout and returns stdout, stderr, error.
func runCommand(timeout time.Duration, dir string, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}

	// Inherit system environment variables
	cmd.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("command timed out after %s: %s %s", timeout, name, strings.Join(args, " "))
	}
	if err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr != "" {
			return nil, fmt.Errorf("%s (code %w)", errStr, err)
		}
		return nil, err
	}

	return stdout.Bytes(), nil
}

// GetFileStatus discovers repository context, branch, dirty state, and remote PR info.
// Implements: R-053, R-060, AC-117, AC-118, AC-119, AC-141
func (s *GitService) GetFileStatus(rawPath, sessionID string) (*GitFileStatus, error) {
	cleanPath := s.CleanPath(rawPath)
	if cleanPath == "" {
		return nil, fmt.Errorf("empty file path")
	}

	gitRoot := findGitRoot(cleanPath)
	if gitRoot == "" {
		// Not inside a git repository
		resp := &GitFileStatus{
			IsGit:        false,
			SessionSteps: s.countSessionSteps(cleanPath, sessionID),
		}
		return resp, nil
	}

	relPath, err := filepath.Rel(gitRoot, cleanPath)
	if err != nil {
		relPath = cleanPath
	}

	status := &GitFileStatus{
		IsGit:        true,
		GitRoot:      gitRoot,
		RelativePath: relPath,
		SessionSteps: s.countSessionSteps(cleanPath, sessionID),
	}

	// 1. Discover current branch
	branchOut, err := runCommand(3*time.Second, gitRoot, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err == nil {
		status.Branch = strings.TrimSpace(string(branchOut))
	}

	// 2. Discover remote origin URL
	remoteOut, err := runCommand(3*time.Second, gitRoot, "git", "config", "--get", "remote.origin.url")
	if err == nil {
		status.RemoteURL = strings.TrimSpace(string(remoteOut))
	}

	// 3. Inspect dirty state for this specific file
	statusOut, err := runCommand(3*time.Second, gitRoot, "git", "status", "--porcelain=v1", "--", relPath)
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(statusOut)), "\n")
		for _, line := range lines {
			line = strings.TrimRight(line, "\r")
			if len(line) >= 2 {
				indexStatus := line[0]
				workTreeStatus := line[1]

				if indexStatus != ' ' && indexStatus != '?' {
					status.HasStaged = true
				}
				if workTreeStatus != ' ' || indexStatus == '?' {
					status.HasUnstaged = true
				}
			}
		}
		status.IsDirty = status.HasStaged || status.HasUnstaged
	}

	// 4. Discover machine-level gh CLI and auth status
	if ghPath, err := exec.LookPath("gh"); err == nil && ghPath != "" {
		status.GHAvailable = true
		// Check auth status
		if _, authErr := runCommand(3*time.Second, gitRoot, "gh", "auth", "status"); authErr == nil {
			status.GHAuthenticated = true

			// 5. Query active PR info via gh pr view
			prJSON, prErr := runCommand(4*time.Second, gitRoot, "gh", "pr", "view", "--json",
				"number,title,url,state,headRefName,baseRefName,reviewDecision,statusCheckRollup")
			if prErr == nil && len(prJSON) > 0 {
				var prData struct {
					Number         int    `json:"number"`
					Title          string `json:"title"`
					URL            string `json:"url"`
					State          string `json:"state"`
					BaseRefName    string `json:"baseRefName"`
					HeadRefName    string `json:"headRefName"`
					ReviewDecision string `json:"reviewDecision"`
					StatusCheck    []struct {
						Conclusion string `json:"conclusion"`
						Status     string `json:"status"`
					} `json:"statusCheckRollup"`
				}
				if jsonErr := json.Unmarshal(prJSON, &prData); jsonErr == nil && prData.Number > 0 {
					totalChecks := len(prData.StatusCheck)
					passingChecks := 0
					for _, c := range prData.StatusCheck {
						if c.Conclusion == "SUCCESS" || c.Status == "COMPLETED" {
							passingChecks++
						}
					}
					statusSummary := ""
					if totalChecks > 0 {
						statusSummary = fmt.Sprintf("%d/%d checks passing", passingChecks, totalChecks)
					}

					status.ActivePR = &GitHubPR{
						Number:         prData.Number,
						Title:          prData.Title,
						URL:            prData.URL,
						State:          prData.State,
						BaseRef:        prData.BaseRefName,
						HeadRef:        prData.HeadRefName,
						ReviewDecision: prData.ReviewDecision,
						StatusSummary:  statusSummary,
					}
				}
			}
		}
	}

	return status, nil
}

// GetFileDiff returns formatted diff hunks and statistics for unstaged, staged, or commit diffs.
// Implements: R-054, AC-120, AC-121, AC-122, AC-124, AC-125
func (s *GitService) GetFileDiff(rawPath, mode, commitSHA string) (*GitDiffResponse, error) {
	cleanPath := s.CleanPath(rawPath)
	if cleanPath == "" {
		return nil, fmt.Errorf("empty file path")
	}

	gitRoot := findGitRoot(cleanPath)
	if gitRoot == "" {
		return nil, fmt.Errorf("path is not inside a git repository: %s", cleanPath)
	}

	relPath, err := filepath.Rel(gitRoot, cleanPath)
	if err != nil {
		relPath = cleanPath
	}

	var args []string
	switch mode {
	case "staged":
		args = []string{"diff", "--cached", "--no-ext-diff", "-U3", "--", relPath}
	case "commit":
		if strings.TrimSpace(commitSHA) == "" {
			return nil, fmt.Errorf("commit_sha parameter is required for mode=commit")
		}
		// Sanitize commit SHA to hex characters only
		cleanSHA := strings.TrimSpace(commitSHA)
		for _, r := range cleanSHA {
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
				return nil, fmt.Errorf("invalid commit_sha: %s", cleanSHA)
			}
		}
		args = []string{"show", "--no-ext-diff", "-U3", cleanSHA, "--", relPath}
	default:
		mode = "unstaged"
		args = []string{"diff", "--no-ext-diff", "-U3", "--", relPath}
	}

	const maxDiffBytes = 2 * 1024 * 1024 // 2MB ceiling (AC-125)
	rawOut, err := runCommand(3*time.Second, gitRoot, "git", args...)
	if err != nil {
		return nil, err
	}

	if len(rawOut) == 0 && mode == "unstaged" {
		// Fallback for untracked (new) files: diff against /dev/null
		// git diff --no-index exits with status 1 when differences exist
		if _, statErr := os.Stat(cleanPath); statErr == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "git", "diff", "--no-index", "--", "/dev/null", relPath)
			cmd.Dir = gitRoot
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			_ = cmd.Run() // exit status 1 is normal when diff exists
			if stdout.Len() > 0 {
				rawOut = stdout.Bytes()
			}
		}
	}

	truncated := false
	if len(rawOut) > maxDiffBytes {
		rawOut = rawOut[:maxDiffBytes]
		truncated = true
	}

	hunks, additions, deletions := ParseUnifiedDiff(string(rawOut))

	resp := &GitDiffResponse{
		Path:      cleanPath,
		Mode:      mode,
		RawDiff:   string(rawOut),
		Additions: additions,
		Deletions: deletions,
		Hunks:     hunks,
		Truncated: truncated,
	}

	return resp, nil
}

// GetFileCommits returns chronological commits touching this file.
// Implements: R-054, AC-123
func (s *GitService) GetFileCommits(rawPath string, limit int) ([]GitCommitEntry, error) {
	cleanPath := s.CleanPath(rawPath)
	if cleanPath == "" {
		return nil, fmt.Errorf("empty file path")
	}

	gitRoot := findGitRoot(cleanPath)
	if gitRoot == "" {
		return nil, fmt.Errorf("path is not inside a git repository: %s", cleanPath)
	}

	relPath, err := filepath.Rel(gitRoot, cleanPath)
	if err != nil {
		relPath = cleanPath
	}

	if limit <= 0 || limit > 100 {
		limit = 30
	}

	// %H (full SHA), %h (short SHA), %an (author name), %ae (author email), %at (epoch timestamp), %s (subject)
	formatStr := "%H%x1f%h%x1f%an%x1f%ae%x1f%at%x1f%s"
	out, err := runCommand(3*time.Second, gitRoot, "git", "log", "--follow", fmt.Sprintf("-n%d", limit),
		"--pretty=format:"+formatStr, "--", relPath)
	if err != nil {
		return nil, err
	}

	var commits []GitCommitEntry
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	now := time.Now().Unix()

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\x1f")
		if len(parts) >= 6 {
			ts, _ := strconv.ParseInt(parts[4], 10, 64)
			relTime := formatRelativeTime(time.Duration(now-ts) * time.Second)
			commits = append(commits, GitCommitEntry{
				SHA:          parts[0],
				ShortSHA:     parts[1],
				AuthorName:   parts[2],
				AuthorEmail:  parts[3],
				Timestamp:    ts,
				RelativeTime: relTime,
				Subject:      parts[5],
			})
		}
	}

	return commits, nil
}

// countSessionSteps counts how many steps in a given session touched the file.
func (s *GitService) countSessionSteps(cleanPath, sessionID string) int {
	if s.index == nil || strings.TrimSpace(sessionID) == "" {
		return 0
	}

	detail, err := s.index.GetSessionDetail(sessionID)
	if err != nil || detail == nil {
		return 0
	}

	normPath := strings.ToLower(cleanPath)
	baseName := strings.ToLower(filepath.Base(cleanPath))
	stepCount := 0

	for _, msg := range detail.Messages {
		matched := false
		for _, tc := range msg.ToolCalls {
			for _, v := range tc.Args {
				if strVal, ok := v.(string); ok {
					sClean := strings.ToLower(strings.TrimSpace(strVal))
					if strings.Contains(sClean, normPath) || (len(baseName) > 4 && strings.HasSuffix(sClean, baseName)) {
						matched = true
						break
					}
				}
			}
			if matched {
				break
			}
		}
		if matched {
			stepCount++
		}
	}

	return stepCount
}

// ParseUnifiedDiff parses unified diff text into structured DiffHunks and statistics.
func ParseUnifiedDiff(diffText string) ([]DiffHunk, int, int) {
	var hunks []DiffHunk
	var currentHunk *DiffHunk

	additions := 0
	deletions := 0

	lines := strings.Split(diffText, "\n")
	oldLine := 0
	newLine := 0

	for _, line := range lines {
		// Hunk header: @@ -oldStart,oldLines +newStart,newLines @@
		if strings.HasPrefix(line, "@@") {
			if currentHunk != nil {
				hunks = append(hunks, *currentHunk)
			}
			header := line
			oldStart, oldLines, newStart, newLines := parseHunkHeader(line)
			oldLine = oldStart
			newLine = newStart
			currentHunk = &DiffHunk{
				Header:   header,
				OldStart: oldStart,
				OldLines: oldLines,
				NewStart: newStart,
				NewLines: newLines,
				Lines:    make([]DiffLine, 0),
			}
			continue
		}

		if currentHunk == nil {
			// Skip metadata headers (diff --git, index, ---, +++)
			continue
		}

		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			additions++
			currentHunk.Lines = append(currentHunk.Lines, DiffLine{
				Type:       "add",
				Content:    line[1:],
				NewLineNum: newLine,
			})
			newLine++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			deletions++
			currentHunk.Lines = append(currentHunk.Lines, DiffLine{
				Type:       "delete",
				Content:    line[1:],
				OldLineNum: oldLine,
			})
			oldLine++
		} else if strings.HasPrefix(line, " ") {
			currentHunk.Lines = append(currentHunk.Lines, DiffLine{
				Type:       "context",
				Content:    line[1:],
				OldLineNum: oldLine,
				NewLineNum: newLine,
			})
			oldLine++
			newLine++
		}
	}

	if currentHunk != nil {
		hunks = append(hunks, *currentHunk)
	}

	return hunks, additions, deletions
}

// parseHunkHeader extracts line range bounds from @@ -l,s +l,s @@
func parseHunkHeader(header string) (int, int, int, int) {
	// e.g. @@ -42,6 +42,8 @@
	parts := strings.Split(header, "@@")
	if len(parts) < 2 {
		return 1, 1, 1, 1
	}
	rangeStr := strings.TrimSpace(parts[1])
	tokens := strings.Fields(rangeStr)
	oldStart, oldLines := 1, 1
	newStart, newLines := 1, 1

	for _, token := range tokens {
		if strings.HasPrefix(token, "-") {
			token = strings.TrimPrefix(token, "-")
			sub := strings.Split(token, ",")
			if len(sub) > 0 {
				oldStart, _ = strconv.Atoi(sub[0])
			}
			if len(sub) > 1 {
				oldLines, _ = strconv.Atoi(sub[1])
			}
		} else if strings.HasPrefix(token, "+") {
			token = strings.TrimPrefix(token, "+")
			sub := strings.Split(token, ",")
			if len(sub) > 0 {
				newStart, _ = strconv.Atoi(sub[0])
			}
			if len(sub) > 1 {
				newLines, _ = strconv.Atoi(sub[1])
			}
		}
	}

	return oldStart, oldLines, newStart, newLines
}

