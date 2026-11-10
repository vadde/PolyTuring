// Package main provides automated IDE detection and process execution.
//
// Implements: R-076, R-077, R-079
// See: specs/catalog/session-explorer-spatial-autolock-and-ide-cockpit.md
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type editorDef struct {
	ID        string
	Name      string
	Icon      string
	CLIName   string
	AppNames  []string
	URIScheme string
}

var knownEditors = []editorDef{
	{
		ID:        "vscode",
		Name:      "Visual Studio Code",
		Icon:      "💙",
		CLIName:   "code",
		AppNames:  []string{"Visual Studio Code.app"},
		URIScheme: "vscode://file/",
	},
	{
		ID:        "cursor",
		Name:      "Cursor",
		Icon:      "🚀",
		CLIName:   "cursor",
		AppNames:  []string{"Cursor.app"},
		URIScheme: "cursor://file/",
	},
	{
		ID:        "windsurf",
		Name:      "Windsurf",
		Icon:      "🌊",
		CLIName:   "windsurf",
		AppNames:  []string{"Windsurf.app"},
		URIScheme: "windsurf://file/",
	},
	{
		ID:        "intellij",
		Name:      "IntelliJ IDEA",
		Icon:      "🧠",
		CLIName:   "idea",
		AppNames:  []string{"IntelliJ IDEA.app", "IntelliJ IDEA CE.app", "IntelliJ IDEA Ultimate.app"},
		URIScheme: "idea://open?file=",
	},
	{
		ID:        "goland",
		Name:      "GoLand",
		Icon:      "🐹",
		CLIName:   "goland",
		AppNames:  []string{"GoLand.app"},
		URIScheme: "goland://open?file=",
	},
	{
		ID:        "pycharm",
		Name:      "PyCharm",
		Icon:      "🐍",
		CLIName:   "pycharm",
		AppNames:  []string{"PyCharm.app", "PyCharm CE.app"},
		URIScheme: "pycharm://open?file=",
	},
	{
		ID:        "webstorm",
		Name:      "WebStorm",
		Icon:      "⚡",
		CLIName:   "webstorm",
		AppNames:  []string{"WebStorm.app"},
		URIScheme: "webstorm://open?file=",
	},
	{
		ID:        "sublime",
		Name:      "Sublime Text",
		Icon:      "📜",
		CLIName:   "subl",
		AppNames:  []string{"Sublime Text.app"},
		URIScheme: "subl://open?url=file://",
	},
	{
		ID:        "vscode-insiders",
		Name:      "VS Code Insiders",
		Icon:      "⚡",
		CLIName:   "code-insiders",
		AppNames:  []string{"Visual Studio Code - Insiders.app"},
		URIScheme: "vscode-insiders://file/",
	},
	{
		ID:        "antigravity",
		Name:      "Antigravity IDE",
		Icon:      "🛸",
		CLIName:   "agy",
		AppNames:  []string{"Antigravity IDE.app", "Antigravity.app"},
		URIScheme: "antigravity://file/",
	},
	{
		ID:        "xcode",
		Name:      "Xcode",
		Icon:      "🛠️",
		CLIName:   "xed",
		AppNames:  []string{"Xcode.app"},
		URIScheme: "xcode://file/",
	},
	{
		ID:        "zed",
		Name:      "Zed",
		Icon:      "⚡",
		CLIName:   "zed",
		AppNames:  []string{"Zed.app"},
		URIScheme: "zed://file/",
	},
	{
		ID:        "system",
		Name:      "System Default",
		Icon:      "🍎",
		CLIName:   "open",
		AppNames:  nil,
		URIScheme: "file://",
	},
}

// DetectInstalledIDEs scans the host environment to discover installed IDEs.
// Implements: R-077, R-079, AC-208, AC-209
func DetectInstalledIDEs() []DetectedIDE {
	var results []DetectedIDE
	homeDir, _ := os.UserHomeDir()

	for _, ed := range knownEditors {
		detected := DetectedIDE{
			ID:        ed.ID,
			Name:      ed.Name,
			Icon:      ed.Icon,
			URIScheme: ed.URIScheme,
			Installed: false,
		}

		if ed.ID == "system" {
			detected.Installed = true
			if runtime.GOOS != "darwin" {
				detected.Icon = "💻"
				if _, err := exec.LookPath("xdg-open"); err == nil {
					detected.CLIBinary = "xdg-open"
				}
			} else {
				detected.CLIBinary = "/usr/bin/open"
			}
			results = append(results, detected)
			continue
		}

		// 1. Check CLI binary in PATH
		if p, err := exec.LookPath(ed.CLIName); err == nil {
			detected.Installed = true
			detected.CLIBinary = p
		}

		// 2. On macOS, inspect /Applications and ~/Applications
		if runtime.GOOS == "darwin" && len(ed.AppNames) > 0 {
			for _, app := range ed.AppNames {
				sysApp := filepath.Join("/Applications", app)
				if _, err := os.Stat(sysApp); err == nil {
					detected.Installed = true
					detected.AppPath = sysApp
					detected.AppName = strings.TrimSuffix(app, ".app")
					break
				}
				if homeDir != "" {
					userApp := filepath.Join(homeDir, "Applications", app)
					if _, err := os.Stat(userApp); err == nil {
						detected.Installed = true
						detected.AppPath = userApp
						detected.AppName = strings.TrimSuffix(app, ".app")
						break
					}
				}
			}
		}

		results = append(results, detected)
	}

	return results
}

// LaunchIDE launches the specified editor at path and line using native process execution.
// Implements: R-077, R-079
func LaunchIDE(editor, cleanPath string, line int) OpenInIDEResponse {
	detectedList := DetectInstalledIDEs()
	var target *DetectedIDE
	for i := range detectedList {
		if strings.EqualFold(detectedList[i].ID, editor) {
			target = &detectedList[i]
			break
		}
	}

	// Alias fallback (e.g. idea vs intellij)
	if target == nil {
		if editor == "idea" {
			for i := range detectedList {
				if detectedList[i].ID == "intellij" {
					target = &detectedList[i]
					break
				}
			}
		}
	}

	if target == nil {
		target = &DetectedIDE{
			ID:        editor,
			Name:      strings.Title(editor),
			Icon:      "🚀",
			CLIBinary: editor,
			URIScheme: fmt.Sprintf("%s://file/", editor),
		}
	}

	var fallbackURI string
	switch target.ID {
	case "intellij", "idea":
		fallbackURI = fmt.Sprintf("idea://open?file=%s&line=%d", cleanPath, line)
	case "goland":
		fallbackURI = fmt.Sprintf("goland://open?file=%s&line=%d", cleanPath, line)
	case "pycharm":
		fallbackURI = fmt.Sprintf("pycharm://open?file=%s&line=%d", cleanPath, line)
	case "webstorm":
		fallbackURI = fmt.Sprintf("webstorm://open?file=%s&line=%d", cleanPath, line)
	case "sublime":
		fallbackURI = fmt.Sprintf("subl://open?url=file://%s&line=%d", cleanPath, line)
	case "system":
		fallbackURI = fmt.Sprintf("file://%s", cleanPath)
	default:
		fallbackURI = fmt.Sprintf("%s%s:%d", target.URIScheme, cleanPath, line)
	}

	// 1. Execute via CLI binary if available
	if target.CLIBinary != "" {
		var args []string
		switch target.ID {
		case "vscode", "vscode-insiders", "cursor", "windsurf", "zed":
			args = []string{"-g", fmt.Sprintf("%s:%d", cleanPath, line)}
		case "intellij", "idea", "goland", "pycharm", "webstorm":
			args = []string{"--line", strconv.Itoa(line), cleanPath}
		case "sublime":
			args = []string{fmt.Sprintf("%s:%d", cleanPath, line)}
		case "xcode":
			args = []string{"-l", strconv.Itoa(line), cleanPath}
		case "system":
			args = []string{cleanPath}
		default:
			args = []string{fmt.Sprintf("%s:%d", cleanPath, line)}
		}

		cmd := exec.Command(target.CLIBinary, args...)
		if err := cmd.Start(); err == nil {
			return OpenInIDEResponse{
				Success:     true,
				Message:     fmt.Sprintf("Opened in %s via CLI at line %d", target.Name, line),
				Editor:      target.ID,
				Command:     fmt.Sprintf("%s %s", target.CLIBinary, strings.Join(args, " ")),
				FallbackURI: fallbackURI,
			}
		}
	}

	// 2. On macOS, execute via open -a <AppName>
	if runtime.GOOS == "darwin" && target.AppName != "" {
		var openArgs []string
		switch target.ID {
		case "vscode", "vscode-insiders", "cursor", "windsurf", "zed":
			openArgs = []string{"-a", target.AppName, "--args", "-g", fmt.Sprintf("%s:%d", cleanPath, line)}
		case "intellij", "idea", "goland", "pycharm", "webstorm":
			openArgs = []string{"-a", target.AppName, "--args", "--line", strconv.Itoa(line), cleanPath}
		case "sublime":
			openArgs = []string{"-a", target.AppName, fmt.Sprintf("%s:%d", cleanPath, line)}
		case "xcode":
			openArgs = []string{"-a", target.AppName, cleanPath}
		default:
			openArgs = []string{"-a", target.AppName, cleanPath}
		}

		cmd := exec.Command("/usr/bin/open", openArgs...)
		if err := cmd.Start(); err == nil {
			return OpenInIDEResponse{
				Success:     true,
				Message:     fmt.Sprintf("Opened in %s at line %d", target.Name, line),
				Editor:      target.ID,
				Command:     fmt.Sprintf("open %s", strings.Join(openArgs, " ")),
				FallbackURI: fallbackURI,
			}
		}
	}

	// 3. Fallback to protocol URI
	return OpenInIDEResponse{
		Success:     false,
		Message:     fmt.Sprintf("Native launch unhandled for %s; use protocol URI", target.Name),
		Editor:      target.ID,
		Command:     fmt.Sprintf("open %s", fallbackURI),
		FallbackURI: fallbackURI,
	}
}
