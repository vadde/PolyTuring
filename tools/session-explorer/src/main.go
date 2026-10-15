// Package main is the entry point for Session Explorer.
//
// Session Explorer is a self-contained CLI tool that scans Antigravity IDE
// conversation sessions and serves a premium glassmorphic web UI for browsing,
// searching, and filtering agent session histories.
//
// Implements: R-006 (Serve web UI), R-007 (Auto-open browser), R-023 (--data-dir),
//             R-024 (--port), NF-005 (Zero dependencies), NF-009 (Fast startup)
// See: specs/catalog/session-explorer.md
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

const (
	appName    = "session-explorer"
	appVersion = "0.1.0"
)

func main() {
	homeDir, _ := os.UserHomeDir()
	defaultDataDir := filepath.Join(homeDir, ".gemini", "antigravity-ide", "brain")

	// Support 'session-explorer wip' subcommand — Implements: R-034
	if len(os.Args) > 1 && (os.Args[1] == "wip" || os.Args[1] == "--wip") {
		handleWIPSubcommand(os.Args[2:], defaultDataDir)
		return
	}

	// CLI flag parsing — Implements: R-023, R-024
	var cfg Config
	flag.StringVar(&cfg.DataDir, "data-dir", defaultDataDir, "Path to Antigravity brain directory")
	flag.StringVar(&cfg.DataDir, "d", defaultDataDir, "Path to Antigravity brain directory (shorthand)")
	flag.IntVar(&cfg.Port, "port", 9876, "Port to serve the web UI on")
	flag.IntVar(&cfg.Port, "p", 9876, "Port to serve the web UI on (shorthand)")
	flag.IntVar(&cfg.WatchInterval, "watch-interval", 10, "Seconds between background auto-rescans (0 = disabled)")
	flag.IntVar(&cfg.WatchInterval, "w", 10, "Seconds between background auto-rescans (shorthand)")
	flag.BoolVar(&cfg.NoOpen, "no-open", false, "Don't auto-open browser")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&cfg.Verbose, "v", false, "Enable verbose logging (shorthand)")

	showVersion := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s v%s\n", appName, appVersion)
		os.Exit(0)
	}

	// Validate data directory
	if cfg.DataDir == defaultDataDir {
		discovered := AutoDiscoverLocations()
		if len(discovered) == 0 {
			log.Fatalf("❌ No AI session directories found.\n\nChecked:\n- %s\n- %s\n\nUse --data-dir to specify a custom brain directory.",
				filepath.Join(homeDir, ".gemini", "antigravity-ide", "brain"),
				filepath.Join(homeDir, ".gemini", "antigravity-cli", "brain"))
		}
	} else if _, err := os.Stat(cfg.DataDir); os.IsNotExist(err) {
		log.Fatalf("❌ Data directory not found: %s\n\nUse --data-dir to specify the path to your brain directory.", cfg.DataDir)
	}

	// Banner
	fmt.Printf(`
  ╔══════════════════════════════════════════╗
  ║     🔍 Session Explorer v%s          ║
  ║  Universal Multi-Assistant Library (R-036)║
  ╚══════════════════════════════════════════╝
`, appVersion)

	// Create session index and scan — Implements: R-001
	index := NewSessionIndex(cfg.DataDir, cfg.Verbose)

	log.Println("📂 Scanning sessions...")
	if err := index.ScanAll(); err != nil {
		log.Fatalf("❌ Failed to scan sessions: %v", err)
	}

	stats := index.GetStats()
	log.Printf("✅ Found %d sessions, %d user messages, %d agent responses across %d workspaces",
		stats.TotalSessions, stats.TotalUserMessages, stats.TotalAgentResponses, len(stats.Workspaces))

	// Auto-open browser — Implements: R-007
	if !cfg.NoOpen {
		url := fmt.Sprintf("http://127.0.0.1:%d", cfg.Port)
		go func() {
			time.Sleep(300 * time.Millisecond) // Wait for server to start
			if err := openBrowser(url); err != nil {
				log.Printf("⚠️  Could not open browser: %v\n  → Navigate to %s manually", err, url)
			}
		}()
	}

	// Start server — Implements: R-006
	if err := StartServer(cfg, index); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}

// openBrowser opens the default browser on the current OS.
// Implements: R-007
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// handleWIPSubcommand processes headless CLI invocations of 'session-explorer wip'.
// Implements: R-034 (Headless CLI Subcommand)
func handleWIPSubcommand(args []string, defaultDataDir string) {
	wipFlags := flag.NewFlagSet("wip", flag.ExitOnError)
	workspace := wipFlags.String("workspace", "", "Workspace or repository name")
	wipFlags.StringVar(workspace, "w", "", "Workspace or repository name (shorthand)")
	sessionID := wipFlags.String("session", "", "Target exact session UUID")
	wipFlags.StringVar(sessionID, "s", "", "Target exact session UUID (shorthand)")
	dataDir := wipFlags.String("data-dir", defaultDataDir, "Path to brain directory")
	wipFlags.StringVar(dataDir, "d", defaultDataDir, "Path to brain directory (shorthand)")
	limit := wipFlags.Int("limit", 5, "Candidate limit")
	forceLatest := wipFlags.Bool("latest", false, "Force selection of newest session")
	wipFlags.BoolVar(forceLatest, "f", false, "Force selection of newest session (shorthand)")
	outputJSON := wipFlags.Bool("json", false, "Output structured JSON instead of Markdown")

	_ = wipFlags.Parse(args)

	// If workspace is empty, detect from current working directory
	targetWS := *workspace
	if targetWS == "" && *sessionID == "" {
		cwd, err := os.Getwd()
		if err == nil {
			targetWS = filepath.Base(cwd)
		}
	}

	index := NewSessionIndex(*dataDir, false)
	if err := index.ScanAll(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error scanning sessions: %v\n", err)
		os.Exit(1)
	}

	resp, err := index.GetWorkspaceWIP(targetWS, *sessionID, *forceLatest, *limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error retrieving WIP: %v\n", err)
		os.Exit(1)
	}

	if *outputJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(resp)
		return
	}

	// Markdown rendering
	if resp.Status == "not_found" {
		fmt.Printf("❌ No conversation sessions found for workspace %q.\n", targetWS)
		return
	}

	if resp.Multiple {
		fmt.Printf("🛑 MANDATORY AGENT HALT: Multiple sessions found for workspace %q.\n\n", targetWS)
		fmt.Println("You MUST stop immediately and ask the user which session they wish to resume:")
		fmt.Println()
		fmt.Println("| # | Session ID | Last Active | Steps | Prompts | Last Prompt Preview |")
		fmt.Println("|---|---|---|---|---|---|")
		for i, c := range resp.Candidates {
			fmt.Printf("| %d | `%s` | %s | %d | %d | %s |\n",
				i+1, c.SessionID, c.RelativeTime, c.StepCount, c.UserMessageCount, c.LastPromptPreview)
		}
		fmt.Println()
		fmt.Printf("Please ask the user: \"I found %d recent sessions for workspace %s. Which session should I resume?\"\n", len(resp.Candidates), targetWS)
		return
	}

	act := resp.ActiveSession
	if act == nil {
		fmt.Println("No active session data available.")
		return
	}

	fmt.Printf("### 🧭 Session WIP Summary: %s\n\n", act.ProjectName)
	fmt.Printf("- **Session ID**: `%s`\n", act.SessionID)
	fmt.Printf("- **Last Active**: %s (%s)\n", act.RelativeTime, act.LastMessageAt.Format(time.RFC3339))
	fmt.Printf("- **Steps**: %d | **Prompts**: %d | **Errors**: %d\n\n", act.StepCount, act.UserMessageCount, act.ErrorCount)

	if len(act.RecentPrompts) > 0 {
		fmt.Println("#### 📝 Recent User Prompts:")
		for i, p := range act.RecentPrompts {
			fmt.Printf("%d. %s\n", i+1, p)
		}
		fmt.Println()
	}

	if len(act.ModifiedFiles) > 0 {
		fmt.Println("#### 🛠️ Modified Files:")
		for _, f := range act.ModifiedFiles {
			fmt.Printf("- `%s`\n", f)
		}
		fmt.Println()
	}

	if len(act.ExecutedCommands) > 0 {
		fmt.Println("#### ⚡ Executed Shell Commands:")
		for _, cmd := range act.ExecutedCommands {
			fmt.Printf("- `%s`\n", cmd)
		}
		fmt.Println()
	}

	if len(act.UnresolvedErrors) > 0 {
		fmt.Println("#### ⚠️ Recent Errors / Blockers:")
		for _, e := range act.UnresolvedErrors {
			fmt.Printf("- %s\n", e)
		}
		fmt.Println()
	}

	if act.SuggestedContinuation != "" {
		fmt.Printf("#### 🎯 Suggested Continuation:\n%s\n", act.SuggestedContinuation)
	}
}

