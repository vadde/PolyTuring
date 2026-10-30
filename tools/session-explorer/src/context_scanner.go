// Package main provides agent context discovery and behavioral analysis across global,
// workspace, and session levels.
//
// Implements: R-039 (Universal Agent Context & Behavioral Matrix)
// See: specs/catalog/session-explorer.md#R-039
package main

import (
	"bufio"
	"encoding/json"
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

// ExtractSessionAgentContext analyzes a specific conversation session's transcript and events
// to extract active rules, skills actually invoked vs available on-demand, and tools used.
// Implements: R-039, AC-042, AC-043
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
		MCPServers:      make([]MCPServerInfo, 0),
		Subagents:       make([]string, 0),
		SlashCommands:   make([]string, 0),
	}

	// Load workspace context to resolve skill details
	wsCtx := idx.DiscoverWorkspaceAgentContext(detail.Session.ProjectName)
	if len(wsCtx.Rules) == 0 && detail.Session.Workspace != "" {
		wsCtx = idx.DiscoverWorkspaceAgentContext(detail.Session.Workspace)
	}

	// Track skill mentions and invocations across the session timeline
	skillInvocationMap := make(map[string]int)  // skillName -> call count
	skillFirstStepMap := make(map[string]int)   // skillName -> first step_index
	seenSubagents := make(map[string]bool)
	seenSlashCommands := make(map[string]bool)
	injectedRulePaths := make(map[string]bool)

	// Map of available skills by name
	availableSkillIndex := make(map[string]AgentSkill)
	for _, sk := range wsCtx.Skills {
		availableSkillIndex[sk.Name] = sk
	}

	// Scan transcript entries
	for _, m := range detail.Messages {
		contentLower := strings.ToLower(m.Content)

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

		// 4. Detect tool calls that invoke or read skills
		for _, tc := range m.ToolCalls {
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

			// Check tool calls referencing SKILL.md or skill names in args
			for _, val := range tc.Args {
				if strVal, ok := val.(string); ok {
					strLower := strings.ToLower(strVal)
					for skName, sk := range availableSkillIndex {
						// Matches skill file path or skill directory
						if strings.Contains(strLower, strings.ToLower(sk.FilePath)) ||
							strings.Contains(strLower, "/skills/"+strings.ToLower(skName)+"/") ||
							strings.Contains(strLower, "/"+strings.ToLower(skName)+"/skill.md") {
							skillInvocationMap[skName]++
							if _, exists := skillFirstStepMap[skName]; !exists {
								skillFirstStepMap[skName] = m.StepIndex
							}
						}
					}
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

	// Populate ActiveSkills vs AvailableSkills
	for _, sk := range wsCtx.Skills {
		invokedCount := skillInvocationMap[sk.Name]
		if invokedCount > 0 {
			sk.InvokedCount = invokedCount
			sk.FirstStepIdx = skillFirstStepMap[sk.Name]
			sctx.ActiveSkills = append(sctx.ActiveSkills, sk)
		} else {
			sctx.AvailableSkills = append(sctx.AvailableSkills, sk)
		}
	}

	// Sort active skills by invocation count descending
	sort.Slice(sctx.ActiveSkills, func(i, j int) bool {
		return sctx.ActiveSkills[i].InvokedCount > sctx.ActiveSkills[j].InvokedCount
	})

	sctx.MCPServers = wsCtx.MCPServers

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
