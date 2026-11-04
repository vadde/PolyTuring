// Package main provides agent context discovery and behavioral analysis across global,
// workspace, and session levels.
//
// Implements: R-039 (Universal Agent Context & Behavioral Matrix)
// See: specs/catalog/session-explorer.md#R-039
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

var (
	reYAMLFrontmatter = regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---`)
	reNameField       = regexp.MustCompile(`(?m)^name:\s*["']?([^"'\r\n]+)["']?`)
	reDescField       = regexp.MustCompile(`(?m)^description:\s*(?:>-\s*|\s*["']?)([^"'\r\n]+)`)
	reCategoryField   = regexp.MustCompile(`(?m)^category:\s*["']?([^"'\r\n]+)["']?`)
	rePriorityBadge   = regexp.MustCompile(`(?i)>\s*\*\*Priority\*\*:\s*([^\n\r]+)`)
	rePrecedenceBadge = regexp.MustCompile(`(?i)>\s*\*\*Precedence\*\*:\s*([^\n\r]+)`)
	reTriggerBadge    = regexp.MustCompile(`(?i)>\s*\*\*Trigger\*\*:\s*([^\n\r]+)`)
	reH1Title         = regexp.MustCompile(`(?m)^#\s+(.+)`)
	reRuleTag         = regexp.MustCompile(`<RULE\[(.*?)\]>`)
	reSkillPromptTag  = regexp.MustCompile(`- ([a-zA-Z0-9_-]+) \((/[^)]+SKILL\.md)\)`)
)

// calculateTextMetrics computes line count, word count, character count, and estimated tokens for a string.
// Standard GenAI token estimation rule:
// - English words average ~1.33 tokens each.
// - Characters average ~4 characters per token for English text and code.
// - Formula: tokensEst = max(ceil(charCount / 4), ceil(wordCount * 1.33))
func calculateTextMetrics(content string) (lineCount int, wordCount int, charCount int, tokensEst int) {
	if len(content) == 0 {
		return 0, 0, 0, 0
	}
	lines := strings.Split(content, "\n")
	lineCount = len(lines)
	words := strings.Fields(content)
	wordCount = len(words)
	charCount = utf8.RuneCountInString(content)

	estByChar := (charCount + 3) / 4
	estByWord := int(float64(wordCount) * 1.33)
	tokensEst = estByChar
	if estByWord > tokensEst {
		tokensEst = estByWord
	}
	if tokensEst < 1 && charCount > 0 {
		tokensEst = 1
	}
	return lineCount, wordCount, charCount, tokensEst
}

// parseMarkdownHeader extracts title, priority, trigger, line count, word count, char count, estimated tokens, and description from a Markdown file.
func parseMarkdownHeader(filePath string) (title string, description string, priority string, trigger string, lineCount int, wordCount int, charCount int, tokensEst int, sizeBytes int64) {
	info, err := os.Stat(filePath)
	if err != nil {
		return filepath.Base(filePath), "", "Standard", "always_on", 0, 0, 0, 0, 0
	}
	sizeBytes = info.Size()

	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return filepath.Base(filePath), "", "Standard", "always_on", 0, 0, 0, 0, sizeBytes
	}
	content := string(contentBytes)
	lineCount, wordCount, charCount, tokensEst = calculateTextMetrics(content)

	// Extract Title
	if m := reH1Title.FindStringSubmatch(content); len(m) > 1 {
		title = strings.TrimSpace(m[1])
	} else {
		title = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	}

	// Extract Priority / Precedence
	if m := rePrecedenceBadge.FindStringSubmatch(content); len(m) > 1 {
		priority = cleanBadgeText(m[1])
	} else if m := rePriorityBadge.FindStringSubmatch(content); len(m) > 1 {
		priority = cleanBadgeText(m[1])
	} else if strings.Contains(strings.ToLower(filePath), "constitution") {
		priority = "🔴 Constitutional"
	} else {
		priority = "Standard"
	}

	// Extract Trigger
	if m := reTriggerBadge.FindStringSubmatch(content); len(m) > 1 {
		trigger = strings.TrimSpace(m[1])
	} else {
		trigger = "always_on"
	}

	// Extract Description (first paragraph following title or blockquote)
	var descLines []string
	inFrontmatter := false
	recordingDesc := false

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "---" {
			inFrontmatter = !inFrontmatter
			continue
		}
		if inFrontmatter {
			continue
		}
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ">") {
			continue
		}
		if line == "" {
			if recordingDesc && len(descLines) > 0 {
				break
			}
			continue
		}
		recordingDesc = true
		descLines = append(descLines, line)
		if len(descLines) >= 3 {
			break
		}
	}
	description = strings.Join(descLines, " ")
	return
}

func cleanBadgeText(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, ">")
	raw = strings.TrimSpace(raw)
	return raw
}

// parseSkillYAML extracts skill metadata from SKILL.md YAML frontmatter and computes text metrics.
func parseSkillYAML(filePath string) (name string, description string, category string, hasScripts bool, hasExamples bool, hasRefs bool, lineCount int, wordCount int, charCount int, tokensEst int, sizeBytes int64) {
	dir := filepath.Dir(filePath)
	name = filepath.Base(dir)

	if stat, err := os.Stat(filepath.Join(dir, "scripts")); err == nil && stat.IsDir() {
		hasScripts = true
	}
	if stat, err := os.Stat(filepath.Join(dir, "examples")); err == nil && stat.IsDir() {
		hasExamples = true
	}
	if stat, err := os.Stat(filepath.Join(dir, "references")); err == nil && stat.IsDir() {
		hasRefs = true
	}

	info, err := os.Stat(filePath)
	if err == nil {
		sizeBytes = info.Size()
	}

	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return name, "", "", hasScripts, hasExamples, hasRefs, 0, 0, 0, 0, sizeBytes
	}
	content := string(contentBytes)
	lineCount, wordCount, charCount, tokensEst = calculateTextMetrics(content)

	if fm := reYAMLFrontmatter.FindStringSubmatch(content); len(fm) > 1 {
		yamlBody := fm[1]
		if m := reNameField.FindStringSubmatch(yamlBody); len(m) > 1 {
			name = strings.TrimSpace(m[1])
		}
		if m := reDescField.FindStringSubmatch(yamlBody); len(m) > 1 {
			description = strings.TrimSpace(m[1])
		}
		if m := reCategoryField.FindStringSubmatch(yamlBody); len(m) > 1 {
			category = strings.TrimSpace(m[1])
		}
	}

	if description == "" {
		// Fallback: extract first paragraph after #
		lines := strings.Split(content, "\n")
		var p []string
		for _, l := range lines {
			t := strings.TrimSpace(l)
			if strings.HasPrefix(t, "#") || strings.HasPrefix(t, "---") || strings.HasPrefix(t, ">") || t == "" {
				if len(p) > 0 {
					break
				}
				continue
			}
			p = append(p, t)
			if len(p) >= 2 {
				break
			}
		}
		description = strings.Join(p, " ")
	}

	return name, description, category, hasScripts, hasExamples, hasRefs, lineCount, wordCount, charCount, tokensEst, sizeBytes
}

// parseMCPConfig parses an mcp_config.json file into MCPServerInfo structs.
func parseMCPConfig(filePath string, scope AgentContextScope) []MCPServerInfo {
	var servers []MCPServerInfo
	content, err := os.ReadFile(filePath)
	if err != nil {
		return servers
	}

	var data struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}

	if err := json.Unmarshal(content, &data); err != nil {
		return servers
	}

	for name, s := range data.MCPServers {
		servers = append(servers, MCPServerInfo{
			Name:    name,
			Command: s.Command,
			Args:    s.Args,
			Scope:   scope,
		})
	}
	return servers
}

// parseHooksConfig parses a hooks.json file into AgentHook structs.
func parseHooksConfig(filePath string, scope AgentContextScope) []AgentHook {
	var hooks []AgentHook
	content, err := os.ReadFile(filePath)
	if err != nil {
		return hooks
	}

	var data map[string]struct {
		Enabled       bool `json:"enabled"`
		PreToolUse    []hookEntry `json:"PreToolUse"`
		PreInvocation []hookEntry `json:"PreInvocation"`
		Stop          []hookEntry `json:"Stop"`
	}

	if err := json.Unmarshal(content, &data); err != nil {
		return hooks
	}

	for guardName, g := range data {
		if !g.Enabled {
			continue
		}
		for _, e := range g.PreToolUse {
			for _, h := range e.Hooks {
				hooks = append(hooks, AgentHook{
					Name:    guardName,
					Event:   "PreToolUse",
					Matcher: e.Matcher,
					Type:    h.Type,
					Command: h.Command,
					Timeout: h.Timeout,
					Scope:   scope,
				})
			}
		}
		for _, e := range g.PreInvocation {
			hooks = append(hooks, AgentHook{
				Name:    guardName,
				Event:   "PreInvocation",
				Type:    e.Type,
				Command: e.Command,
				Timeout: e.Timeout,
				Scope:   scope,
			})
		}
		for _, e := range g.Stop {
			hooks = append(hooks, AgentHook{
				Name:    guardName,
				Event:   "Stop",
				Type:    e.Type,
				Command: e.Command,
				Timeout: e.Timeout,
				Scope:   scope,
			})
		}
	}
	return hooks
}

type hookEntry struct {
	Matcher string      `json:"matcher,omitempty"`
	Type    string      `json:"type"`
	Command string      `json:"command"`
	Timeout int         `json:"timeout,omitempty"`
	Hooks   []hookEntry `json:"hooks,omitempty"`
}

// DiscoverGlobalAgentContext discovers all machine-wide rules, skills, plugins, MCP servers, and hooks.
// Implements: R-039, AC-040, AC-044
func (idx *SessionIndex) DiscoverGlobalAgentContext() GlobalAgentContext {
	ctx := GlobalAgentContext{
		Rules:      make([]AgentRule, 0),
		Skills:     make([]AgentSkill, 0),
		Plugins:    make([]AgentPlugin, 0),
		MCPServers: make([]MCPServerInfo, 0),
		Hooks:      make([]AgentHook, 0),
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ctx
	}

	seenRules := make(map[string]bool)
	seenSkills := make(map[string]bool)

	// 1. Antigravity Global Rules: ~/.gemini/config/rules/*.md
	globalRulesDir := filepath.Join(homeDir, ".gemini", "config", "rules")
	if entries, err := os.ReadDir(globalRulesDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				fullPath := filepath.Join(globalRulesDir, e.Name())
				if seenRules[fullPath] {
					continue
				}
				seenRules[fullPath] = true
				title, desc, priority, trigger, lines, words, chars, tokens, size := parseMarkdownHeader(fullPath)
				ctx.Rules = append(ctx.Rules, AgentRule{
					Name:        e.Name(),
					Title:       title,
					Scope:       ScopeGlobal,
					FilePath:    fullPath,
					Description: desc,
					Priority:    priority,
					Trigger:     trigger,
					LineCount:   lines,
					WordCount:   words,
					CharCount:   chars,
					TokensEst:   tokens,
					SizeBytes:   size,
					Provider:    "antigravity",
				})
			}
		}
	}

	// 2. Claude Code Global Rules: ~/.claude/rules/*.md
	claudeRulesDir := filepath.Join(homeDir, ".claude", "rules")
	if entries, err := os.ReadDir(claudeRulesDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				fullPath := filepath.Join(claudeRulesDir, e.Name())
				if seenRules[fullPath] {
					continue
				}
				seenRules[fullPath] = true
				title, desc, priority, trigger, lines, words, chars, tokens, size := parseMarkdownHeader(fullPath)
				ctx.Rules = append(ctx.Rules, AgentRule{
					Name:        e.Name(),
					Title:       title,
					Scope:       ScopeGlobal,
					FilePath:    fullPath,
					Description: desc,
					Priority:    priority,
					Trigger:     trigger,
					LineCount:   lines,
					WordCount:   words,
					CharCount:   chars,
					TokensEst:   tokens,
					SizeBytes:   size,
					Provider:    "claude",
				})
			}
		}
	}

	// 3. Cursor Global Rules: ~/.cursorrules
	cursorGlobal := filepath.Join(homeDir, ".cursorrules")
	if _, err := os.Stat(cursorGlobal); err == nil && !seenRules[cursorGlobal] {
		seenRules[cursorGlobal] = true
		title, desc, priority, trigger, lines, words, chars, tokens, size := parseMarkdownHeader(cursorGlobal)
		ctx.Rules = append(ctx.Rules, AgentRule{
			Name:        ".cursorrules",
			Title:       title,
			Scope:       ScopeGlobal,
			FilePath:    cursorGlobal,
			Description: desc,
			Priority:    priority,
			Trigger:     trigger,
			LineCount:   lines,
			WordCount:   words,
			CharCount:   chars,
			TokensEst:   tokens,
			SizeBytes:   size,
			Provider:    "cursor",
		})
	}

	// 4. Global Skills: ~/.gemini/config/skills/
	globalSkillsDir := filepath.Join(homeDir, ".gemini", "config", "skills")
	if entries, err := os.ReadDir(globalSkillsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				skillPath := filepath.Join(globalSkillsDir, e.Name(), "SKILL.md")
				if _, err := os.Stat(skillPath); err == nil && !seenSkills[skillPath] {
					seenSkills[skillPath] = true
					name, desc, cat, hasScr, hasEx, hasRef, lines, words, chars, tokens, size := parseSkillYAML(skillPath)
					ctx.Skills = append(ctx.Skills, AgentSkill{
						Name:          name,
						Scope:         ScopeGlobal,
						FilePath:      skillPath,
						Description:   desc,
						Category:      cat,
						Provider:      "antigravity",
						HasScripts:    hasScr,
						HasExamples:   hasEx,
						HasReferences: hasRef,
						LineCount:     lines,
						WordCount:     words,
						CharCount:     chars,
						TokensEst:     tokens,
						SizeBytes:     size,
					})
				}
			}
		}
	}

	// 5. Built-in Skills: ~/.gemini/antigravity-cli/builtin/skills/
	builtinSkillsDirs := []string{
		filepath.Join(homeDir, ".gemini", "antigravity-cli", "builtin", "skills"),
		filepath.Join(homeDir, ".gemini", "antigravity-ide", "builtin", "skills"),
	}
	for _, bDir := range builtinSkillsDirs {
		if entries, err := os.ReadDir(bDir); err == nil {
			for _, e := range entries {
				if e.IsDir() {
					skillPath := filepath.Join(bDir, e.Name(), "SKILL.md")
					if _, err := os.Stat(skillPath); err == nil && !seenSkills[skillPath] {
						seenSkills[skillPath] = true
						name, desc, cat, hasScr, hasEx, hasRef, lines, words, chars, tokens, size := parseSkillYAML(skillPath)
						ctx.Skills = append(ctx.Skills, AgentSkill{
							Name:          name,
							Scope:         ScopeBuiltin,
							FilePath:      skillPath,
							Description:   desc,
							Category:      cat,
							Provider:      "antigravity",
							HasScripts:    hasScr,
							HasExamples:   hasEx,
							HasReferences: hasRef,
							LineCount:     lines,
							WordCount:     words,
							CharCount:     chars,
							TokensEst:     tokens,
							SizeBytes:     size,
						})
					}
				}
			}
		}
	}

	// 6. Global Plugins: ~/.gemini/config/plugins/
	pluginsDir := filepath.Join(homeDir, ".gemini", "config", "plugins")
	if entries, err := os.ReadDir(pluginsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				pDir := filepath.Join(pluginsDir, e.Name())
				pluginJSON := filepath.Join(pDir, "plugin.json")
				var p AgentPlugin
				p.Name = e.Name()
				p.Path = pDir

				if pBytes, err := os.ReadFile(pluginJSON); err == nil {
					var raw struct {
						Name        string `json:"name"`
						Version     string `json:"version"`
						Description string `json:"description"`
					}
					if err := json.Unmarshal(pBytes, &raw); err == nil {
						if raw.Name != "" {
							p.Name = raw.Name
						}
						p.Version = raw.Version
						p.Description = raw.Description
					}
				}

				// Check plugin skills
				plugSkillsDir := filepath.Join(pDir, "skills")
				if sEntries, err := os.ReadDir(plugSkillsDir); err == nil {
					for _, se := range sEntries {
						if se.IsDir() {
							skillPath := filepath.Join(plugSkillsDir, se.Name(), "SKILL.md")
							if _, err := os.Stat(skillPath); err == nil && !seenSkills[skillPath] {
								seenSkills[skillPath] = true
								p.Skills = append(p.Skills, se.Name())
								name, desc, cat, hasScr, hasEx, hasRef, lines, words, chars, tokens, size := parseSkillYAML(skillPath)
								ctx.Skills = append(ctx.Skills, AgentSkill{
									Name:          name,
									Scope:         ScopePlugin,
									FilePath:      skillPath,
									Description:   desc,
									Category:      cat,
									Provider:      "antigravity",
									PluginName:    p.Name,
									HasScripts:    hasScr,
									HasExamples:   hasEx,
									HasReferences: hasRef,
									LineCount:     lines,
									WordCount:     words,
									CharCount:     chars,
									TokensEst:     tokens,
									SizeBytes:     size,
								})
							}
						}
					}
				}
				ctx.Plugins = append(ctx.Plugins, p)
			}
		}
	}

	// 7. Global MCP Config: ~/.gemini/config/mcp_config.json
	mcpPath := filepath.Join(homeDir, ".gemini", "config", "mcp_config.json")
	ctx.MCPServers = append(ctx.MCPServers, parseMCPConfig(mcpPath, ScopeGlobal)...)

	// 8. Global Hooks: ~/.gemini/config/hooks.json
	hooksPath := filepath.Join(homeDir, ".gemini", "config", "hooks.json")
	ctx.Hooks = append(ctx.Hooks, parseHooksConfig(hooksPath, ScopeGlobal)...)

	// Sort rules and skills for deterministic display
	sort.Slice(ctx.Rules, func(i, j int) bool {
		return ctx.Rules[i].Name < ctx.Rules[j].Name
	})
	sort.Slice(ctx.Skills, func(i, j int) bool {
		return ctx.Skills[i].Name < ctx.Skills[j].Name
	})

	ctx.TotalRules = len(ctx.Rules)
	ctx.TotalSkills = len(ctx.Skills)

	for _, r := range ctx.Rules {
		ctx.PreloadedTokensEst += r.TokensEst
		ctx.TotalWords += r.WordCount
	}
	for _, s := range ctx.Skills {
		ctx.OnDemandTokensEst += s.TokensEst
		ctx.TotalWords += s.WordCount
	}
	ctx.TotalTokensEst = ctx.PreloadedTokensEst + ctx.OnDemandTokensEst

	return ctx
}

// DiscoverWorkspaceAgentContext discovers workspace-specific rules, master agent docs, and skills,
// merging them with inherited global context with strict precedence tracking.
// Implements: R-039, AC-041, AC-044
func (idx *SessionIndex) DiscoverWorkspaceAgentContext(workspaceTarget string) WorkspaceAgentContext {
	wsClean := strings.TrimSpace(workspaceTarget)

	// Resolve target workspace path
	var wsPath string
	var projName string

	idx.mu.RLock()
	if p, ok := idx.knownProjects[wsClean]; ok {
		wsPath = p
		projName = wsClean
	} else {
		for name, p := range idx.knownProjects {
			if strings.EqualFold(name, wsClean) || strings.EqualFold(filepath.Base(p), wsClean) {
				wsPath = p
				projName = name
				break
			}
		}
	}
	idx.mu.RUnlock()

	if wsPath == "" {
		if strings.HasPrefix(wsClean, "/") || strings.HasPrefix(wsClean, "~") {
			homeDir, _ := os.UserHomeDir()
			if strings.HasPrefix(wsClean, "~/") {
				wsPath = filepath.Join(homeDir, wsClean[2:])
			} else {
				wsPath = wsClean
			}
			projName = filepath.Base(wsPath)
		} else {
			wsPath = wsClean
			projName = wsClean
		}
	}

	wctx := WorkspaceAgentContext{
		Workspace:   wsPath,
		ProjectName: projName,
		Rules:       make([]AgentRule, 0),
		Skills:      make([]AgentSkill, 0),
		MasterDocs:  make([]string, 0),
		MCPServers:  make([]MCPServerInfo, 0),
		Hooks:       make([]AgentHook, 0),
	}

	seenRules := make(map[string]bool)
	seenSkills := make(map[string]bool)

	// 1. Discover Workspace Project Rules (<workspace>/.agents/rules/*.md, .agent/rules/, etc.)
	agentDirs := []string{
		filepath.Join(wsPath, ".agents"),
		filepath.Join(wsPath, ".agent"),
		filepath.Join(wsPath, "_agents"),
	}

	for _, aDir := range agentDirs {
		rulesDir := filepath.Join(aDir, "rules")
		if entries, err := os.ReadDir(rulesDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
					fullPath := filepath.Join(rulesDir, e.Name())
					if seenRules[fullPath] {
						continue
					}
					seenRules[fullPath] = true
					title, desc, priority, trigger, lines, words, chars, tokens, size := parseMarkdownHeader(fullPath)
					wctx.Rules = append(wctx.Rules, AgentRule{
						Name:        e.Name(),
						Title:       title,
						Scope:       ScopeWorkspace,
						FilePath:    fullPath,
						Description: desc,
						Priority:    priority,
						Trigger:     trigger,
						LineCount:   lines,
						WordCount:   words,
						CharCount:   chars,
						TokensEst:   tokens,
						SizeBytes:   size,
						Provider:    "antigravity",
						Workspace:   projName,
					})
				}
			}
		}

		// Workspace Skills: <aDir>/skills/*/SKILL.md
		skillsDir := filepath.Join(aDir, "skills")
		if entries, err := os.ReadDir(skillsDir); err == nil {
			for _, e := range entries {
				if e.IsDir() {
					skillPath := filepath.Join(skillsDir, e.Name(), "SKILL.md")
					if _, err := os.Stat(skillPath); err == nil && !seenSkills[skillPath] {
						seenSkills[skillPath] = true
						name, desc, cat, hasScr, hasEx, hasRef, lines, words, chars, tokens, size := parseSkillYAML(skillPath)
						wctx.Skills = append(wctx.Skills, AgentSkill{
							Name:          name,
							Scope:         ScopeWorkspace,
							FilePath:      skillPath,
							Description:   desc,
							Category:      cat,
							Provider:      "antigravity",
							Workspace:     projName,
							HasScripts:    hasScr,
							HasExamples:   hasEx,
							HasReferences: hasRef,
							LineCount:     lines,
							WordCount:     words,
							CharCount:     chars,
							TokensEst:     tokens,
							SizeBytes:     size,
						})
					}
				}
			}
		}

		// Workspace Hooks: <aDir>/hooks.json
		hooksPath := filepath.Join(aDir, "hooks.json")
		wctx.Hooks = append(wctx.Hooks, parseHooksConfig(hooksPath, ScopeWorkspace)...)

		// Workspace MCP: <aDir>/mcp_config.json
		mcpPath := filepath.Join(aDir, "mcp_config.json")
		wctx.MCPServers = append(wctx.MCPServers, parseMCPConfig(mcpPath, ScopeWorkspace)...)
	}

	// 2. Discover Master Agent Instructions & Single Sources of Truth
	masterCandidatePaths := []struct {
		path     string
		provider string
		name     string
	}{
		{filepath.Join(wsPath, ".agents", "AGENTS.md"), "antigravity", ".agents/AGENTS.md"},
		{filepath.Join(wsPath, "AGENTS.md"), "antigravity", "AGENTS.md"},
		{filepath.Join(wsPath, "GEMINI.md"), "antigravity", "GEMINI.md"},
		{filepath.Join(wsPath, "CLAUDE.md"), "claude", "CLAUDE.md"},
		{filepath.Join(wsPath, ".claude", "CLAUDE.md"), "claude", ".claude/CLAUDE.md"},
		{filepath.Join(wsPath, ".cursorrules"), "cursor", ".cursorrules"},
		{filepath.Join(wsPath, ".github", "copilot-instructions.md"), "copilot", ".github/copilot-instructions.md"},
	}

	for _, cand := range masterCandidatePaths {
		if _, err := os.Stat(cand.path); err == nil && !seenRules[cand.path] {
			seenRules[cand.path] = true
			wctx.MasterDocs = append(wctx.MasterDocs, cand.name)
			title, desc, priority, trigger, lines, words, chars, tokens, size := parseMarkdownHeader(cand.path)
			wctx.Rules = append([]AgentRule{{
				Name:        cand.name,
				Title:       title,
				Scope:       ScopeWorkspace,
				FilePath:    cand.path,
				Description: desc,
				Priority:    priority,
				Trigger:     trigger,
				LineCount:   lines,
				WordCount:   words,
				CharCount:   chars,
				TokensEst:   tokens,
				SizeBytes:   size,
				Provider:    cand.provider,
				Workspace:   projName,
			}}, wctx.Rules...)
		}
	}

	// 3. Cursor MDC Rules: <wsPath>/.cursor/rules/*.mdc
	cursorMdcDir := filepath.Join(wsPath, ".cursor", "rules")
	if entries, err := os.ReadDir(cursorMdcDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".mdc") {
				fullPath := filepath.Join(cursorMdcDir, e.Name())
				if seenRules[fullPath] {
					continue
				}
				seenRules[fullPath] = true
				title, desc, priority, trigger, lines, words, chars, tokens, size := parseMarkdownHeader(fullPath)
				wctx.Rules = append(wctx.Rules, AgentRule{
					Name:        e.Name(),
					Title:       title,
					Scope:       ScopeWorkspace,
					FilePath:    fullPath,
					Description: desc,
					Priority:    priority,
					Trigger:     trigger,
					LineCount:   lines,
					WordCount:   words,
					CharCount:   chars,
					TokensEst:   tokens,
					SizeBytes:   size,
					Provider:    "cursor",
					Workspace:   projName,
				})
			}
		}
	}

	// 4. Inherit Global Context (Rules, Skills, MCP, Hooks) with Precedence Tagging
	global := idx.DiscoverGlobalAgentContext()

	// Append global rules that are not overridden by workspace rules
	for _, gr := range global.Rules {
		if !seenRules[gr.FilePath] {
			seenRules[gr.FilePath] = true
			wctx.Rules = append(wctx.Rules, gr)
		}
	}

	// Append global/builtin skills that are not overridden by workspace skills
	for _, gs := range global.Skills {
		if !seenSkills[gs.FilePath] {
			seenSkills[gs.FilePath] = true
			wctx.Skills = append(wctx.Skills, gs)
		}
	}

	// Append global MCP servers
	wctx.MCPServers = append(wctx.MCPServers, global.MCPServers...)

	// Append global hooks
	wctx.Hooks = append(wctx.Hooks, global.Hooks...)

	wctx.TotalRules = len(wctx.Rules)
	wctx.TotalSkills = len(wctx.Skills)

	for _, r := range wctx.Rules {
		wctx.PreloadedTokensEst += r.TokensEst
		wctx.TotalWords += r.WordCount
	}
	for _, s := range wctx.Skills {
		wctx.OnDemandTokensEst += s.TokensEst
		wctx.TotalWords += s.WordCount
	}
	wctx.TotalTokensEst = wctx.PreloadedTokensEst + wctx.OnDemandTokensEst

	return wctx
}

// cleanStringArg cleans up an argument interface removing outer quotes and excess whitespace.
func cleanStringArg(v interface{}) string {
	if v == nil {
		return ""
	}
	var s string
	switch val := v.(type) {
	case string:
		s = val
	case fmt.Stringer:
		s = val.String()
	default:
		s = fmt.Sprintf("%v", val)
	}
	s = strings.TrimSpace(s)
	for (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) || (strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
		if len(s) >= 2 {
			s = strings.TrimSpace(s[1 : len(s)-1])
		} else {
			break
		}
	}
	return s
}

// categorizeNativeTool maps a native tool name to an intuitive operational category.
func categorizeNativeTool(name string) string {
	switch name {
	case "read_file", "write_file", "edit_file", "view_file":
		return "Filesystem"
	case "run_command", "bash", "exec", "terminal":
		return "Terminal"
	case "invoke_subagent", "define_subagent", "manage_subagents", "send_message":
		return "Subagent"
	case "search_web", "ask_question", "list_resources", "read_resource":
		return "Search"
	case "schedule", "manage_task":
		return "Automation"
	default:
		if strings.HasPrefix(name, "browser_") {
			return "Browser"
		}
		return "Agentic"
	}
}

// extractToolArgSummary extracts a clean, human-readable summary of the most prominent argument.
func extractToolArgSummary(tc ToolCallDisplay) string {
	if tc.Name == "call_mcp_tool" {
		if rawArgs, ok := tc.Args["Arguments"]; ok {
			str := cleanStringArg(rawArgs)
			if str != "" && str != "{}" {
				var parsed map[string]interface{}
				if err := json.Unmarshal([]byte(str), &parsed); err == nil {
					for _, k := range []string{"query", "url", "symbol", "name", "path", "command", "function", "prompt"} {
						if v, exists := parsed[k]; exists {
							valStr := cleanStringArg(v)
							if len(valStr) > 60 {
								valStr = valStr[:60] + "..."
							}
							return fmt.Sprintf("%s=%s", k, valStr)
						}
					}
				}
				if len(str) > 60 {
					return str[:60] + "..."
				}
				return str
			}
		}
	}

	for _, k := range []string{"Path", "CommandLine", "query", "url", "symbol", "TypeName", "Role", "Prompt"} {
		if v, exists := tc.Args[k]; exists {
			s := cleanStringArg(v)
			if len(s) > 60 {
				s = s[:60] + "..."
			}
			return fmt.Sprintf("%s=%s", k, s)
		}
	}
	return ""
}

// parseCommandBinaryAndCategory decomposes a raw command line and tool summary into a binary name, category, and display summary.
// Implements: R-041
func parseCommandBinaryAndCategory(commandLine, toolSummary string) (binary string, category string, cleanSummary string) {
	cmd := strings.TrimSpace(commandLine)
	cleanSummary = strings.TrimSpace(toolSummary)
	if cleanSummary == "" {
		if len(cmd) > 60 {
			cleanSummary = cmd[:60] + "..."
		} else {
			cleanSummary = cmd
		}
	}

	// Split by chained operators (&&, ;, ||, |) to find the primary executable
	parts := strings.FieldsFunc(cmd, func(r rune) bool {
		return r == ';' || r == '|' || r == '&'
	})

	var primarySub string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		// Skip cd, export, set, env if another command follows
		fields := strings.Fields(trimmed)
		if len(fields) == 0 {
			continue
		}
		first := fields[0]
		if first == "cd" || first == "export" || first == "set" || first == "env" {
			continue
		}
		primarySub = trimmed
		break
	}

	if primarySub == "" && len(parts) > 0 {
		primarySub = strings.TrimSpace(parts[0])
	}
	if primarySub == "" {
		primarySub = cmd
	}

	fields := strings.Fields(primarySub)
	if len(fields) == 0 {
		return "unknown", "System", cleanSummary
	}

	rawBin := fields[0]
	// Handle environment variable assignments before command (e.g. PORT=9876 go run)
	for strings.Contains(rawBin, "=") && len(fields) > 1 {
		fields = fields[1:]
		rawBin = fields[0]
	}

	binBase := filepath.Base(rawBin)
	binBase = strings.ToLower(binBase)

	// Sub-command refinement
	secondArg := ""
	if len(fields) > 1 {
		secondArg = strings.ToLower(fields[1])
	}

	switch binBase {
	case "git", "gh", "svn", "hg":
		if secondArg != "" && !strings.HasPrefix(secondArg, "-") {
			return binBase + " " + secondArg, "VCS", cleanSummary
		}
		return binBase, "VCS", cleanSummary

	case "go":
		if secondArg == "test" {
			return "go test", "Test", cleanSummary
		} else if secondArg == "build" {
			return "go build", "Build", cleanSummary
		} else if secondArg == "run" {
			return "go run", "Runtime", cleanSummary
		}
		return "go", "Build", cleanSummary

	case "make":
		if secondArg != "" && !strings.HasPrefix(secondArg, "-") {
			return "make " + secondArg, "Build", cleanSummary
		}
		return "make", "Build", cleanSummary

	case "bun", "npm", "pnpm", "yarn":
		if secondArg == "test" {
			return binBase + " test", "Test", cleanSummary
		} else if secondArg == "run" && len(fields) > 2 {
			sub := strings.ToLower(fields[2])
			if sub == "build" {
				return binBase + " build", "Build", cleanSummary
			}
			return binBase + " " + sub, "Build", cleanSummary
		} else if secondArg == "build" {
			return binBase + " build", "Build", cleanSummary
		}
		return binBase, "Runtime", cleanSummary

	case "pytest", "jest", "vitest", "mocha":
		return binBase, "Test", cleanSummary

	case "cargo", "cmake", "gradle", "mvn", "webpack", "tsc":
		return binBase, "Build", cleanSummary

	case "lsof", "kill", "pkill", "ps", "top":
		return binBase, "Process", cleanSummary

	case "curl", "wget", "nc", "ping", "ssh":
		return binBase, "Network", cleanSummary

	case "grep", "cat", "jq", "find", "ls", "head", "tail", "wc", "diff", "file", "awk", "sed":
		return binBase, "Inspect", cleanSummary

	case "node", "python", "python3", "ruby", "bash", "zsh", "sh":
		return binBase, "Runtime", cleanSummary

	case "mkdir", "rm", "cp", "mv", "touch", "chmod", "chown":
		return binBase, "Filesystem", cleanSummary

	default:
		return binBase, "System", cleanSummary
	}
}

// parseFileExtensionAndCategory decomposes a file path into its extension and operational language/type.
// Implements: R-041
func parseFileExtensionAndCategory(filePath string) (ext string, category string, baseName string) {
	cleanPath := strings.TrimSpace(filePath)
	baseName = filepath.Base(cleanPath)
	ext = strings.ToLower(filepath.Ext(cleanPath))
	if ext == "" {
		if strings.EqualFold(baseName, "Makefile") {
			return "Makefile", "Build", baseName
		}
		if strings.EqualFold(baseName, "Dockerfile") {
			return "Dockerfile", "Container", baseName
		}
		return "no-ext", "Text", baseName
	}

	switch ext {
	case ".go":
		return ".go", "Go Source", baseName
	case ".js", ".mjs", ".cjs":
		return ".js", "JavaScript", baseName
	case ".ts", ".mts":
		return ".ts", "TypeScript", baseName
	case ".css":
		return ".css", "CSS Stylesheet", baseName
	case ".html", ".htm":
		return ".html", "HTML Document", baseName
	case ".json":
		return ".json", "JSON Data", baseName
	case ".md", ".markdown":
		return ".md", "Markdown Doc", baseName
	case ".sh", ".bash", ".zsh":
		return ".sh", "Shell Script", baseName
	case ".yaml", ".yml":
		return ".yaml", "YAML Config", baseName
	case ".py":
		return ".py", "Python", baseName
	case ".rs":
		return ".rs", "Rust", baseName
	case ".sql":
		return ".sql", "SQL", baseName
	default:
		return ext, "Source/Config", baseName
	}
}

// ExtractSessionAgentContext analyzes a specific conversation session's transcript and events
// to extract active rules, skills actually invoked vs available on-demand, MCP servers & tools,
// native tools, and analytical execution telemetry.
// Implements: R-039, R-040, R-041, AC-042, AC-046, AC-047, AC-048, AC-053, AC-054
func (idx *SessionIndex) ExtractSessionAgentContext(sessionID string) (SessionAgentContext, error) {
	detail, err := idx.GetSessionDetail(sessionID)
	if err != nil {
		return SessionAgentContext{}, err
	}

	sctx := SessionAgentContext{
		SessionID:       sessionID,
		Workspace:       detail.Session.Workspace,
		ProjectName:     detail.Session.ProjectName,
		ActiveRules:     make([]AgentRule, 0),
		ActiveSkills:    make([]AgentSkill, 0),
		AvailableSkills: make([]AgentSkill, 0),
		MCPServers:      make([]SessionMCPServer, 0),
		NativeTools:     make([]SessionNativeTool, 0),
		Subagents:       make([]string, 0),
		SlashCommands:   make([]string, 0),
	}

	// Load workspace context to resolve skill details
	wsCtx := idx.DiscoverWorkspaceAgentContext(detail.Session.ProjectName)
	if len(wsCtx.Rules) == 0 && detail.Session.Workspace != "" {
		wsCtx = idx.DiscoverWorkspaceAgentContext(detail.Session.Workspace)
	}

	// Track skill mentions and invocations across the session timeline (R-039, R-042)
	skillInvocationMap := make(map[string]int)       // skillName -> call count
	skillFirstStepMap := make(map[string]int)        // skillName -> first step_index
	skillLastStepMap := make(map[string]int)         // skillName -> last step_index (R-042)
	skillStepsMap := make(map[string][]int)          // skillName -> step indices slice (R-042)
	skillTriggersMap := make(map[string][]string)    // skillName -> trigger action descriptions (R-042)
	seenSubagents := make(map[string]bool)
	seenSlashCommands := make(map[string]bool)
	injectedRulePaths := make(map[string]bool)

	// Precedence helper for skill deduplication (workspace > plugin > global > builtin)
	scopePrecedence := func(s AgentContextScope) int {
		switch s {
		case ScopeWorkspace:
			return 4
		case ScopePlugin:
			return 3
		case ScopeGlobal:
			return 2
		case ScopeBuiltin:
			return 1
		default:
			return 0
		}
	}

	// Map of deduplicated available skills by name (R-042)
	availableSkillIndex := make(map[string]AgentSkill)
	for _, sk := range wsCtx.Skills {
		existing, exists := availableSkillIndex[sk.Name]
		if !exists || scopePrecedence(sk.Scope) > scopePrecedence(existing.Scope) {
			availableSkillIndex[sk.Name] = sk
		}
	}

	// MCP and Tool tracking
	type mcpKey struct {
		server string
		tool   string
	}
	mcpToolStats := make(map[mcpKey]*SessionMCPTool)
	serverInvokedCounts := make(map[string]int)
	nativeToolStats := make(map[string]*SessionNativeTool)
	uniqueToolsMap := make(map[string]bool)
	totalToolCalls := 0
	totalMCPCalls := 0
	totalNativeCalls := 0
	totalErrors := 0

	// Sub-breakdown tracking (R-041)
	type subBreakdownEntry struct {
		key         string
		label       string
		category    string
		count       int
		steps       []int
		recentItems []string
	}
	toolSubBreakdowns := make(map[string]map[string]*subBreakdownEntry)
	globalCommandBreakdown := make(map[string]*subBreakdownEntry)
	globalFileBreakdown := make(map[string]*subBreakdownEntry)
	totalCommandsCount := 0
	totalFilesCount := 0

	recordSubItem := func(toolName, key, label, category string, stepIdx int, summary string) {
		if toolSubBreakdowns[toolName] == nil {
			toolSubBreakdowns[toolName] = make(map[string]*subBreakdownEntry)
		}
		entry := toolSubBreakdowns[toolName][key]
		if entry == nil {
			entry = &subBreakdownEntry{
				key:         key,
				label:       label,
				category:    category,
				steps:       make([]int, 0),
				recentItems: make([]string, 0),
			}
			toolSubBreakdowns[toolName][key] = entry
		}
		entry.count++
		entry.steps = append(entry.steps, stepIdx)
		if summary != "" && len(entry.recentItems) < 3 {
			isDup := false
			for _, prev := range entry.recentItems {
				if prev == summary {
					isDup = true
					break
				}
			}
			if !isDup {
				entry.recentItems = append(entry.recentItems, summary)
			}
		}
	}

	// Scan transcript entries
	for _, m := range detail.Messages {
		contentLower := strings.ToLower(m.Content)
		isError := m.Status == "ERROR"

		// 1. Detect injected rule tags in prompt if present (<RULE[...]>)
		if strings.Contains(m.Content, "<RULE[") {
			matches := reRuleTag.FindAllStringSubmatch(m.Content, -1)
			for _, match := range matches {
				if len(match) > 1 {
					injectedRulePaths[match[1]] = true
				}
			}
		}

		// 2. Detect available skills tags in prompt if present
		if strings.Contains(m.Content, "<skills>") {
			matches := reSkillPromptTag.FindAllStringSubmatch(m.Content, -1)
			for _, match := range matches {
				if len(match) > 2 {
					skillName := match[1]
					skillPath := match[2]
					if _, exists := availableSkillIndex[skillName]; !exists {
						_, desc, cat, hasScr, hasEx, hasRef, lines, words, chars, tokens, size := parseSkillYAML(skillPath)
						if desc == "" {
							desc = "Mounted in system prompt"
						}
						availableSkillIndex[skillName] = AgentSkill{
							Name:          skillName,
							Scope:         ScopeBuiltin,
							FilePath:      skillPath,
							Provider:      "antigravity",
							Description:   desc,
							Category:      cat,
							HasScripts:    hasScr,
							HasExamples:   hasEx,
							HasReferences: hasRef,
							LineCount:     lines,
							WordCount:     words,
							CharCount:     chars,
							TokensEst:     tokens,
							SizeBytes:     size,
						}
					}
				}
			}
		}

		// 3. Detect slash commands mentioned (e.g. /goal, /plan, /teamwork)
		for _, cmd := range []string{"/goal", "/plan", "/browser", "/schedule", "/compact", "/teamwork", "/learn", "/boost", "/grill-me"} {
			if strings.Contains(contentLower, cmd) && !seenSlashCommands[cmd] {
				seenSlashCommands[cmd] = true
				sctx.SlashCommands = append(sctx.SlashCommands, cmd)
			}
		}

		// 4. Detect tool calls
		for _, tc := range m.ToolCalls {
			totalToolCalls++

			if tc.Name == "invoke_subagent" {
				if typeName, ok := tc.Args["TypeName"].(string); ok && typeName != "" {
					if !seenSubagents[typeName] {
						seenSubagents[typeName] = true
						sctx.Subagents = append(sctx.Subagents, typeName)
					}
				}
				if role, ok := tc.Args["Role"].(string); ok && role != "" {
					if !seenSubagents[role] {
						seenSubagents[role] = true
						sctx.Subagents = append(sctx.Subagents, role)
					}
				}
			}

			// Check tool calls referencing SKILL.md or skill names in args (R-042: Tier 1 Evidence)
			matchedSkillsInStep := make(map[string]bool)
			for _, val := range tc.Args {
				strVal := cleanStringArg(val)
				if strVal == "" {
					continue
				}
				strLower := strings.ToLower(strVal)
				for skName, sk := range availableSkillIndex {
					skLower := strings.ToLower(skName)
					isMatch := false
					triggerDesc := ""

					// Match by skill file path or skill directory
					if sk.FilePath != "" && strings.Contains(strLower, strings.ToLower(sk.FilePath)) {
						isMatch = true
						triggerDesc = fmt.Sprintf("%s: %s", tc.Name, filepath.Base(sk.FilePath))
					} else if strings.Contains(strLower, "/skills/"+skLower+"/") ||
						strings.Contains(strLower, "/skills/"+skLower+"/skill.md") ||
						strings.Contains(strLower, "/"+skLower+"/skill.md") ||
						strings.Contains(strLower, "skills/"+skLower+"/skill.md") {
						isMatch = true
						triggerDesc = fmt.Sprintf("%s: %s", tc.Name, "SKILL.md")
					} else if (tc.Name == "run_command" || tc.Name == "bash" || tc.Name == "exec") &&
						(strings.Contains(strLower, "skills/"+skLower) || strings.Contains(strLower, "setup-"+skLower)) {
						isMatch = true
						cmdPreview := strVal
						if len(cmdPreview) > 40 {
							cmdPreview = cmdPreview[:40] + "..."
						}
						triggerDesc = fmt.Sprintf("run_command: %s", cmdPreview)
					}

					if isMatch && !matchedSkillsInStep[skName] {
						matchedSkillsInStep[skName] = true
						skillInvocationMap[skName]++
						skillLastStepMap[skName] = m.StepIndex
						if _, exists := skillFirstStepMap[skName]; !exists {
							skillFirstStepMap[skName] = m.StepIndex
						}
						skillStepsMap[skName] = append(skillStepsMap[skName], m.StepIndex)

						if triggerDesc != "" {
							alreadyHave := false
							for _, prevTrig := range skillTriggersMap[skName] {
								if prevTrig == triggerDesc {
									alreadyHave = true
									break
								}
							}
							if !alreadyHave && len(skillTriggersMap[skName]) < 4 {
								skillTriggersMap[skName] = append(skillTriggersMap[skName], triggerDesc)
							}
						}
					}
				}
			}

			// Categorize tool: Lazy MCP vs Eager MCP vs Native Tool
			if tc.Name == "call_mcp_tool" {
				serverName := cleanStringArg(tc.Args["ServerName"])
				if serverName == "" {
					serverName = cleanStringArg(tc.Args["server_name"])
				}
				if serverName == "" {
					serverName = "unknown-mcp"
				}
				toolName := cleanStringArg(tc.Args["ToolName"])
				if toolName == "" {
					toolName = cleanStringArg(tc.Args["tool_name"])
				}
				if toolName == "" {
					toolName = "call_mcp_tool"
				}

				totalMCPCalls++
				uniqueToolsMap[toolName] = true
				serverInvokedCounts[serverName]++

				key := mcpKey{server: serverName, tool: toolName}
				entry, exists := mcpToolStats[key]
				if !exists {
					entry = &SessionMCPTool{
						Name:              toolName,
						ServerName:        serverName,
						FirstStepIdx:      m.StepIndex,
						LastStepIdx:       m.StepIndex,
						StepIndices:       make([]int, 0),
						RecentArgsSummary: extractToolArgSummary(tc),
					}
					mcpToolStats[key] = entry
				}
				entry.InvokedCount++
				entry.LastStepIdx = m.StepIndex
				entry.StepIndices = append(entry.StepIndices, m.StepIndex)
				if isError {
					entry.ErrorCount++
					totalErrors++
				}
			} else if strings.HasPrefix(tc.Name, "mcp_") {
				parts := strings.SplitN(tc.Name, "_", 3)
				serverName := "mcp"
				toolName := tc.Name
				if len(parts) >= 3 {
					serverName = parts[1]
					toolName = parts[2]
				}

				totalMCPCalls++
				uniqueToolsMap[toolName] = true
				serverInvokedCounts[serverName]++

				key := mcpKey{server: serverName, tool: toolName}
				entry, exists := mcpToolStats[key]
				if !exists {
					entry = &SessionMCPTool{
						Name:              toolName,
						ServerName:        serverName,
						FirstStepIdx:      m.StepIndex,
						LastStepIdx:       m.StepIndex,
						StepIndices:       make([]int, 0),
						RecentArgsSummary: extractToolArgSummary(tc),
					}
					mcpToolStats[key] = entry
				}
				entry.InvokedCount++
				entry.LastStepIdx = m.StepIndex
				entry.StepIndices = append(entry.StepIndices, m.StepIndex)
				if isError {
					entry.ErrorCount++
					totalErrors++
				}
			} else {
				// Native agent tool
				totalNativeCalls++
				uniqueToolsMap[tc.Name] = true

				entry, exists := nativeToolStats[tc.Name]
				if !exists {
					entry = &SessionNativeTool{
						Name:         tc.Name,
						Category:     categorizeNativeTool(tc.Name),
						FirstStepIdx: m.StepIndex,
						LastStepIdx:  m.StepIndex,
						StepIndices:  make([]int, 0),
					}
					nativeToolStats[tc.Name] = entry
				}
				entry.InvokedCount++
				entry.LastStepIdx = m.StepIndex
				entry.StepIndices = append(entry.StepIndices, m.StepIndex)
				if isError {
					entry.ErrorCount++
					totalErrors++
				}

				// Decompose generic commands and file operations (R-041)
				if tc.Name == "run_command" || tc.Name == "bash" || tc.Name == "exec" || tc.Name == "terminal" {
					cmdLine := cleanStringArg(tc.Args["CommandLine"])
					if cmdLine == "" {
						cmdLine = cleanStringArg(tc.Args["command"])
					}
					if cmdLine == "" {
						cmdLine = cleanStringArg(tc.Args["cmd"])
					}
					summary := cleanStringArg(tc.Args["toolSummary"])
					if summary == "" {
						summary = cleanStringArg(tc.Args["summary"])
					}
					bin, cat, cleanSumm := parseCommandBinaryAndCategory(cmdLine, summary)
					recordSubItem(tc.Name, bin, bin, cat, m.StepIndex, cleanSumm)

					totalCommandsCount++
					gEntry := globalCommandBreakdown[bin]
					if gEntry == nil {
						gEntry = &subBreakdownEntry{
							key:         bin,
							label:       bin,
							category:    cat,
							steps:       make([]int, 0),
							recentItems: make([]string, 0),
						}
						globalCommandBreakdown[bin] = gEntry
					}
					gEntry.count++
					gEntry.steps = append(gEntry.steps, m.StepIndex)
					if cleanSumm != "" && len(gEntry.recentItems) < 3 {
						isDup := false
						for _, prev := range gEntry.recentItems {
							if prev == cleanSumm {
								isDup = true
								break
							}
						}
						if !isDup {
							gEntry.recentItems = append(gEntry.recentItems, cleanSumm)
						}
					}
				} else if tc.Name == "read_file" || tc.Name == "write_file" || tc.Name == "edit_file" || tc.Name == "view_file" {
					filePath := cleanStringArg(tc.Args["Path"])
					if filePath == "" {
						filePath = cleanStringArg(tc.Args["path"])
					}
					if filePath == "" {
						filePath = cleanStringArg(tc.Args["TargetFile"])
					}
					if filePath == "" {
						filePath = cleanStringArg(tc.Args["target_file"])
					}
					ext, cat, base := parseFileExtensionAndCategory(filePath)
					lbl := fmt.Sprintf("%s (%s)", ext, cat)
					recordSubItem(tc.Name, ext, lbl, cat, m.StepIndex, base)

					totalFilesCount++
					gEntry := globalFileBreakdown[ext]
					if gEntry == nil {
						gEntry = &subBreakdownEntry{
							key:         ext,
							label:       lbl,
							category:    cat,
							steps:       make([]int, 0),
							recentItems: make([]string, 0),
						}
						globalFileBreakdown[ext] = gEntry
					}
					gEntry.count++
					gEntry.steps = append(gEntry.steps, m.StepIndex)
					if base != "" && len(gEntry.recentItems) < 3 {
						isDup := false
						for _, prev := range gEntry.recentItems {
							if prev == base {
								isDup = true
								break
							}
						}
						if !isDup {
							gEntry.recentItems = append(gEntry.recentItems, base)
						}
					}
				} else if tc.Name == "invoke_subagent" {
					tName := cleanStringArg(tc.Args["TypeName"])
					role := cleanStringArg(tc.Args["Role"])
					key := tName
					if key == "" {
						key = role
					}
					if key == "" {
						key = "subagent"
					}
					prompt := cleanStringArg(tc.Args["Prompt"])
					if len(prompt) > 60 {
						prompt = prompt[:60] + "..."
					}
					recordSubItem(tc.Name, key, key, role, m.StepIndex, prompt)
				}
			}
		}
	}

	// Populate ActiveRules: rules that apply to this session
	if len(injectedRulePaths) > 0 {
		for _, r := range wsCtx.Rules {
			if injectedRulePaths[r.FilePath] || injectedRulePaths[r.Name] || injectedRulePaths["user_global"] {
				sctx.ActiveRules = append(sctx.ActiveRules, r)
			}
		}
	}
	// If no explicit tags were logged in transcript, use all workspace + global rules as active
	if len(sctx.ActiveRules) == 0 {
		sctx.ActiveRules = wsCtx.Rules
	}

	// Populate ActiveSkills vs AvailableSkills (deduplicated by name, R-042)
	uniqueSkills := make([]AgentSkill, 0, len(availableSkillIndex))
	for _, sk := range availableSkillIndex {
		uniqueSkills = append(uniqueSkills, sk)
	}
	sort.Slice(uniqueSkills, func(i, j int) bool {
		return uniqueSkills[i].Name < uniqueSkills[j].Name
	})

	for _, sk := range uniqueSkills {
		invokedCount := skillInvocationMap[sk.Name]
		if invokedCount > 0 {
			sk.InvokedCount = invokedCount
			sk.FirstStepIdx = skillFirstStepMap[sk.Name]
			sk.LastStepIdx = skillLastStepMap[sk.Name]
			sk.StepIndices = skillStepsMap[sk.Name]
			sk.InvocationTriggers = skillTriggersMap[sk.Name]
			sctx.ActiveSkills = append(sctx.ActiveSkills, sk)
		} else {
			sctx.AvailableSkills = append(sctx.AvailableSkills, sk)
		}
	}

	// Sort active skills by invocation count descending, then by name
	sort.Slice(sctx.ActiveSkills, func(i, j int) bool {
		if sctx.ActiveSkills[i].InvokedCount != sctx.ActiveSkills[j].InvokedCount {
			return sctx.ActiveSkills[i].InvokedCount > sctx.ActiveSkills[j].InvokedCount
		}
		return sctx.ActiveSkills[i].Name < sctx.ActiveSkills[j].Name
	})

	// Synthesize SessionMCPServer items
	seenServers := make(map[string]bool)

	// 1. Process configured MCP servers from workspace context
	for _, srv := range wsCtx.MCPServers {
		seenServers[srv.Name] = true
		sessSrv := SessionMCPServer{
			Name:            srv.Name,
			Command:         srv.Command,
			Args:            srv.Args,
			Scope:           srv.Scope,
			ConfiguredTools: srv.Tools,
			ActiveTools:     make([]SessionMCPTool, 0),
			DormantTools:    make([]string, 0),
		}

		activeToolNames := make(map[string]bool)
		for key, tStat := range mcpToolStats {
			if key.server == srv.Name {
				sessSrv.ActiveTools = append(sessSrv.ActiveTools, *tStat)
				sessSrv.InvokedCount += tStat.InvokedCount
				activeToolNames[tStat.Name] = true
			}
		}

		// Identify dormant tools from configured list
		for _, cfgTool := range srv.Tools {
			if !activeToolNames[cfgTool] {
				sessSrv.DormantTools = append(sessSrv.DormantTools, cfgTool)
			}
		}

		// Sort active tools by invoked count descending
		sort.Slice(sessSrv.ActiveTools, func(i, j int) bool {
			return sessSrv.ActiveTools[i].InvokedCount > sessSrv.ActiveTools[j].InvokedCount
		})

		sctx.MCPServers = append(sctx.MCPServers, sessSrv)
	}

	// 2. Discover ad-hoc or dynamic MCP servers from transcript not in static config
	for srvName, count := range serverInvokedCounts {
		if !seenServers[srvName] {
			seenServers[srvName] = true
			sessSrv := SessionMCPServer{
				Name:         srvName,
				Command:      "runtime dynamic",
				Scope:        ScopeSession,
				InvokedCount: count,
				ActiveTools:  make([]SessionMCPTool, 0),
				DormantTools: make([]string, 0),
			}
			for key, tStat := range mcpToolStats {
				if key.server == srvName {
					sessSrv.ActiveTools = append(sessSrv.ActiveTools, *tStat)
				}
			}
			sort.Slice(sessSrv.ActiveTools, func(i, j int) bool {
				return sessSrv.ActiveTools[i].InvokedCount > sessSrv.ActiveTools[j].InvokedCount
			})
			sctx.MCPServers = append(sctx.MCPServers, sessSrv)
		}
	}

	// Sort MCP servers: active servers first (descending by invoked count), then by name
	sort.Slice(sctx.MCPServers, func(i, j int) bool {
		if sctx.MCPServers[i].InvokedCount != sctx.MCPServers[j].InvokedCount {
			return sctx.MCPServers[i].InvokedCount > sctx.MCPServers[j].InvokedCount
		}
		return sctx.MCPServers[i].Name < sctx.MCPServers[j].Name
	})

	// Synthesize NativeTools slice with sub-breakdown (R-041)
	for _, nt := range nativeToolStats {
		if subMap, ok := toolSubBreakdowns[nt.Name]; ok && len(subMap) > 0 {
			nt.SubBreakdown = make([]ToolSubBreakdown, 0, len(subMap))
			for _, entry := range subMap {
				pct := 0.0
				if nt.InvokedCount > 0 {
					pct = math.Round((float64(entry.count)/float64(nt.InvokedCount))*1000) / 10.0
				}
				nt.SubBreakdown = append(nt.SubBreakdown, ToolSubBreakdown{
					Key:         entry.key,
					Label:       entry.label,
					Category:    entry.category,
					Count:       entry.count,
					Percentage:  pct,
					StepIndices: entry.steps,
					RecentItems: entry.recentItems,
				})
			}
			sort.Slice(nt.SubBreakdown, func(i, j int) bool {
				return nt.SubBreakdown[i].Count > nt.SubBreakdown[j].Count
			})
		}
		sctx.NativeTools = append(sctx.NativeTools, *nt)
	}
	sort.Slice(sctx.NativeTools, func(i, j int) bool {
		return sctx.NativeTools[i].InvokedCount > sctx.NativeTools[j].InvokedCount
	})

	// Assemble global command & file breakdowns (R-041)
	cmdBreakdown := make([]ToolSubBreakdown, 0, len(globalCommandBreakdown))
	for _, entry := range globalCommandBreakdown {
		pct := 0.0
		if totalCommandsCount > 0 {
			pct = math.Round((float64(entry.count)/float64(totalCommandsCount))*1000) / 10.0
		}
		cmdBreakdown = append(cmdBreakdown, ToolSubBreakdown{
			Key:         entry.key,
			Label:       entry.label,
			Category:    entry.category,
			Count:       entry.count,
			Percentage:  pct,
			StepIndices: entry.steps,
			RecentItems: entry.recentItems,
		})
	}
	sort.Slice(cmdBreakdown, func(i, j int) bool {
		return cmdBreakdown[i].Count > cmdBreakdown[j].Count
	})

	fileBreakdown := make([]ToolSubBreakdown, 0, len(globalFileBreakdown))
	for _, entry := range globalFileBreakdown {
		pct := 0.0
		if totalFilesCount > 0 {
			pct = math.Round((float64(entry.count)/float64(totalFilesCount))*1000) / 10.0
		}
		fileBreakdown = append(fileBreakdown, ToolSubBreakdown{
			Key:         entry.key,
			Label:       entry.label,
			Category:    entry.category,
			Count:       entry.count,
			Percentage:  pct,
			StepIndices: entry.steps,
			RecentItems: entry.recentItems,
		})
	}
	sort.Slice(fileBreakdown, func(i, j int) bool {
		return fileBreakdown[i].Count > fileBreakdown[j].Count
	})

	// Assemble ToolAnalytics
	sctx.ToolAnalytics = SessionToolAnalytics{
		TotalToolCalls:   totalToolCalls,
		TotalMCPCalls:    totalMCPCalls,
		TotalNativeCalls: totalNativeCalls,
		UniqueToolsUsed:  len(uniqueToolsMap),
		TotalErrors:      totalErrors,
		CommandBreakdown: cmdBreakdown,
		FileBreakdown:    fileBreakdown,
	}
	if totalToolCalls > 0 {
		sctx.ToolAnalytics.ToolDiversityIndex = math.Round((float64(len(uniqueToolsMap))/float64(totalToolCalls))*1000) / 1000
	}

	// Compute session-level token bloat and word metrics
	for _, r := range sctx.ActiveRules {
		sctx.ActiveTokensEst += r.TokensEst
		sctx.TotalWordsInjected += r.WordCount
	}
	for _, s := range sctx.ActiveSkills {
		sctx.ActiveTokensEst += s.TokensEst
		sctx.TotalWordsInjected += s.WordCount
	}
	for _, s := range sctx.AvailableSkills {
		sctx.OnDemandTokensEst += s.TokensEst
	}

	return sctx, nil
}
